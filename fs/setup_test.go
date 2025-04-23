package fs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/jstaf/onedriver/fs/graph"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	mountLoc     = "mount"
	testDBLoc    = "tmp"
	TestDir      = mountLoc + "/onedriver_tests"
	DeltaDir     = TestDir + "/delta"
	retrySeconds = 60 * time.Second //lint:ignore ST1011 a
)

var (
	auth *graph.Auth
	fs   *Filesystem
)

// Tests are done in the main project directory with a mounted filesystem to
// avoid having to repeatedly recreate auth_tokens.json and juggle multiple auth
// sessions.
func TestMain(m *testing.M) {
	// We used to skip paging test setup for single tests, but that caused issues
	// when running TestListChildrenPaging individually

	// Check if we're already in the project root directory
	cwd, _ := os.Getwd()
	if strings.HasSuffix(cwd, "/fs") {
		// If we're in the fs directory, change to the project root
		os.Chdir("..")
	} else if !strings.HasSuffix(cwd, "/onedriver") {
		// If we're not in the project root, try to find it
		// This handles the case where tests are run from GoLand with a different working directory
		if strings.Contains(cwd, "/onedriver") {
			// Extract the path up to and including "onedriver"
			index := strings.Index(cwd, "/onedriver")
			projectRoot := cwd[:index+len("/onedriver")]
			os.Chdir(projectRoot)
		}
	}

	// attempt to unmount regardless of what happens (in case previous tests
	// failed and didn't clean themselves up)
	exec.Command("fusermount3", "-uz", mountLoc).Run()
	os.Mkdir(mountLoc, 0755)
	// wipe all cached data from previous tests
	os.RemoveAll(testDBLoc)
	os.Mkdir(testDBLoc, 0755)

	f, _ := os.OpenFile("fusefs_tests.log", os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: f, TimeFormat: "15:04:05"})
	defer f.Close()

	var err error
	auth, err = graph.Authenticate(context.Background(), graph.AuthConfig{}, ".auth_tokens.json", false)
	if err != nil {
		fmt.Println("Authentication failed:", err)
		os.Exit(1)
	}
	var fsErr error
	fs, fsErr = NewFilesystem(auth, filepath.Join(testDBLoc, "test"), 30)
	if fsErr != nil {
		log.Error().Err(fsErr).Msg("Failed to initialize filesystem")
		os.Exit(1)
	}

	server, err := fuse.NewServer(
		fs,
		mountLoc,
		&fuse.MountOptions{
			Name:          "onedriver",
			FsName:        "onedriver",
			DisableXAttrs: true,
			MaxBackground: 1024,
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create FUSE server")
		os.Exit(1)
	}

	// setup sigint handler for graceful unmount on interrupt/terminate
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	go UnmountHandler(sigChan, server)

	// mount fs in background thread
	go server.Serve()

	// cleanup from last run
	log.Info().Msg("Setup test environment ---------------------------------")
	if err := os.RemoveAll(TestDir); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Mkdir(TestDir, 0755)
	os.Mkdir(DeltaDir, 0755)

	// create paging test files before the delta thread is created
	os.Mkdir(filepath.Join(TestDir, "paging"), 0755)
	createPagingTestFiles()
	go fs.DeltaLoop(5 * time.Second)

	// not created by default on onedrive for business
	os.Mkdir(mountLoc+"/Documents", 0755)

	// we do not cd into the mounted directory or it will hang indefinitely on
	// unmount with "device or resource busy"
	log.Info().Msg("Test session start ---------------------------------")

	// run tests
	code := m.Run()

	log.Info().Msg("Test session end -----------------------------------")
	fmt.Printf("Waiting 5 seconds for any remaining uploads to complete")
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		fmt.Printf(".")
	}
	fmt.Printf("\n")

	// unmount
	if server.Unmount() != nil {
		log.Error().Msg("Failed to unmount test fuse server, attempting lazy unmount")
		exec.Command("fusermount3", "-zu", "mount").Run()
	}
	fmt.Println("Successfully unmounted fuse server!")
	os.Exit(code)
}

// Apparently 200 reqests is the default paging limit.
// Upload at least this many for a later test before the delta thread is created.
func createPagingTestFiles() {
	fmt.Println("Setting up paging test files.")
	var group sync.WaitGroup
	var errCounter int64
	for i := 0; i < 250; i++ {
		group.Add(1)
		go func(n int, wg *sync.WaitGroup) {
			_, err := graph.Put(
				graph.ResourcePath(fmt.Sprintf("/onedriver_tests/paging/%d.txt", n))+":/content",
				auth,
				strings.NewReader("test\n"),
			)
			if err != nil {
				log.Error().Err(err).Msg("Paging upload fail.")
				atomic.AddInt64(&errCounter, 1)
			}
			wg.Done()
		}(i, &group)
	}
	group.Wait()
	log.Info().Msgf("%d failed paging uploads.\n", errCounter)
	fmt.Println("Finished with paging test setup.")
}

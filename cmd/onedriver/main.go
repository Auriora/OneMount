package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/v22/unit"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/jstaf/onedriver/cmd/common"
	"github.com/jstaf/onedriver/fs"
	"github.com/jstaf/onedriver/fs/graph"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
)

func usage() {
	fmt.Printf(`onedriver - A Linux client for Microsoft OneDrive.

This program will mount your OneDrive account as a Linux filesystem at the
specified mountpoint. Note that this is not a sync client - files are only
fetched on-demand and cached locally. Only files you actually use will be
downloaded. While offline, the filesystem will be read-only until
connectivity is re-established.

Usage: onedriver [options] <mountpoint>

Valid options:
`)
	flag.PrintDefaults()
}

// setupFlags initializes and parses command-line flags, returning the configuration and other flag values
func setupFlags() (config *common.Config, authOnly, headless, debugOn bool, mountpoint string) {
	// setup cli parsing
	authOnlyFlag := flag.BoolP("auth-only", "a", false,
		"Authenticate to OneDrive and then exit.")
	headlessFlag := flag.BoolP("no-browser", "n", false,
		"This disables launching the built-in web browser during authentication. "+
			"Follow the instructions in the terminal to authenticate to OneDrive.")
	configPath := flag.StringP("config-file", "f", common.DefaultConfigPath(),
		"A YAML-formatted configuration file used by onedriver.")
	logLevel := flag.StringP("log", "l", "",
		"Set logging level/verbosity for the filesystem. "+
			"Can be one of: fatal, error, warn, info, debug, trace")
	cacheDir := flag.StringP("cache-dir", "c", "",
		"Change the default cache directory used by onedriver. "+
			"Will be created if the path does not already exist.")
	wipeCache := flag.BoolP("wipe-cache", "w", false,
		"Delete the existing onedriver cache directory and then exit. "+
			"This is equivalent to resetting the program.")
	versionFlag := flag.BoolP("version", "v", false, "Display program version.")
	debugOnFlag := flag.BoolP("debug", "d", false, "Enable FUSE debug logging. "+
		"This logs communication between onedriver and the kernel.")
	syncTree := flag.BoolP("sync-tree", "s", false,
		"Sync the full directory tree to the local metadata store in the background. "+
			"This improves performance by pre-caching directory structure without blocking startup.")
	deltaInterval := flag.IntP("delta-interval", "i", 0,
		"Set the interval in seconds between delta query checks. "+
			"Default is 1 seconds. Set to 0 to use the default.")
	cacheExpiration := flag.IntP("cache-expiration", "e", 0,
		"Set the number of days after which files will be removed from the content cache. "+
			"Default is 30 days. Set to 0 to use the default.")
	help := flag.BoolP("help", "h", false, "Displays this help message.")
	flag.Usage = usage
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}
	if *versionFlag {
		fmt.Println("onedriver", common.Version())
		os.Exit(0)
	}

	if *wipeCache {
		config = common.LoadConfig(*configPath)
		if *cacheDir != "" {
			config.CacheDir = *cacheDir
		}
		log.Info().Str("path", config.CacheDir).Msg("Removing cache.")
		os.RemoveAll(config.CacheDir)
		os.Exit(0)
	}

	// determine and validate mountpoint
	if len(flag.Args()) == 0 {
		flag.Usage()
		fmt.Fprintf(os.Stderr, "\nNo mountpoint provided, exiting.\n")
		os.Exit(1)
	}
	mountpoint = flag.Arg(0)

	config = common.LoadConfig(*configPath)
	// command line options override config options
	if *cacheDir != "" {
		config.CacheDir = *cacheDir
	}
	if *logLevel != "" {
		config.LogLevel = *logLevel
	}
	if *syncTree {
		config.SyncTree = true
	}
	if *deltaInterval > 0 {
		config.DeltaInterval = *deltaInterval
	}
	if *cacheExpiration > 0 {
		config.CacheExpiration = *cacheExpiration
	}

	zerolog.SetGlobalLevel(common.StringToLevel(config.LogLevel))

	return config, *authOnlyFlag, *headlessFlag, *debugOnFlag, mountpoint
}

// initializeFilesystem sets up the filesystem and returns the filesystem, auth, server, and paths
func initializeFilesystem(config *common.Config, mountpoint string, authOnly, headless, debugOn bool) (*fs.Filesystem, *graph.Auth, *fuse.Server, string, string, error) {
	// compute cache name as systemd would
	absMountPath, _ := filepath.Abs(mountpoint)
	cachePath := filepath.Join(config.CacheDir, unit.UnitNamePathEscape(absMountPath))

	// authenticate/re-authenticate if necessary
	os.MkdirAll(cachePath, 0700)
	authPath := filepath.Join(cachePath, "auth_tokens.json")
	if authOnly {
		os.Remove(authPath)
		_, err := graph.Authenticate(context.Background(), config.AuthConfig, authPath, headless)
		if err != nil {
			log.Error().Err(err).Msg("Authentication failed")
			return nil, nil, nil, "", "", fmt.Errorf("authentication failed: %w", err)
		}
		os.Exit(0)
	}

	// create the filesystem
	log.Info().Msgf("onedriver %s", common.Version())
	auth, err := graph.Authenticate(context.Background(), config.AuthConfig, authPath, headless)
	if err != nil {
		log.Error().Err(err).Msg("Authentication failed")
		return nil, nil, nil, "", "", fmt.Errorf("authentication failed: %w", err)
	}

	filesystem, err := fs.NewFilesystem(auth, cachePath, config.CacheExpiration)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize filesystem")
		return nil, nil, nil, "", "", fmt.Errorf("failed to initialize filesystem: %w", err)
	}

	log.Info().Msgf("Setting delta query interval to %d second(s)", config.DeltaInterval)
	go filesystem.DeltaLoop(time.Duration(config.DeltaInterval) * time.Second)

	// Start the content cache cleanup routine
	if config.CacheExpiration > 0 {
		log.Info().Msgf("Setting content cache expiration to %d day(s)", config.CacheExpiration)
		filesystem.StartCacheCleanup()
	}

	common.CreateXDGVolumeInfo(filesystem, auth)

	// Sync the full directory tree if requested
	if config.SyncTree {
		log.Info().Msg("Starting full directory tree synchronization in background...")
		go func() {
			if err := filesystem.SyncDirectoryTree(auth); err != nil {
				log.Error().Err(err).Msg("Error syncing directory tree")
			} else {
				log.Info().Msg("Directory tree sync completed successfully")
			}
		}()
	}

	server, err := fuse.NewServer(filesystem, mountpoint, &fuse.MountOptions{
		Name:          "onedriver",
		FsName:        "onedriver",
		DisableXAttrs: true,
		MaxBackground: 1024,
		Debug:         debugOn,
	})
	if err != nil {
		log.Error().Err(err).Msgf("Mount failed. Is the mountpoint already in use? "+
			"(Try running \"fusermount3 -uz %s\")\n", mountpoint)
		return nil, nil, nil, "", "", fmt.Errorf("mount failed (is the mountpoint already in use?): %w", err)
	}

	return filesystem, auth, server, cachePath, absMountPath, nil
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	config, authOnly, headless, debugOn, mountpoint := setupFlags()

	st, err := os.Stat(mountpoint)
	if err != nil || !st.IsDir() {
		log.Fatal().
			Str("mountpoint", mountpoint).
			Msg("Mountpoint did not exist or was not a directory.")
	}
	if res, _ := os.ReadDir(mountpoint); len(res) > 0 {
		log.Fatal().Str("mountpoint", mountpoint).Msg("Mountpoint must be empty.")
	}

	// Initialize the filesystem
	filesystem, _, server, cachePath, absMountPath, err := initializeFilesystem(config, mountpoint, authOnly, headless, debugOn)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize filesystem")
	}

	// setup signal handler for graceful unmount on signals like sigint
	setupSignalHandler(filesystem, server)

	// serve filesystem
	log.Info().
		Str("cachePath", cachePath).
		Str("mountpoint", absMountPath).
		Msg("Serving filesystem.")
	server.Serve()
}

// setupSignalHandler sets up a handler for SIGINT and SIGTERM signals to gracefully unmount the filesystem
func setupSignalHandler(filesystem *fs.Filesystem, server *fuse.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a custom signal handler that stops background processes before unmounting
	go func() {
		sig := <-sigChan // block until signal
		log.Info().Str("signal", strings.ToUpper(sig.String())).
			Msg("Signal received, cleaning up and unmounting filesystem.")

		// Stop the cache cleanup routine
		filesystem.StopCacheCleanup()

		// Stop the delta loop
		filesystem.StopDeltaLoop()

		// Stop the download manager
		filesystem.StopDownloadManager()

		// Stop the upload manager
		filesystem.StopUploadManager()

		// Give the system a moment to release all resources
		log.Info().Msg("Waiting for all resources to be released before unmounting...")
		time.Sleep(500 * time.Millisecond)

		// Unmount the filesystem
		err := server.Unmount()
		if err != nil {
			log.Error().Err(err).Msg("Failed to unmount filesystem cleanly! " +
				"Run \"fusermount3 -uz /MOUNTPOINT/GOES/HERE\" to unmount.")
		}

		os.Exit(128)
	}()
}

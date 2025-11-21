package fs

import (
	"path/filepath"
	"testing"

	"github.com/auriora/onemount/internal/testutil"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

type sandboxSnapshot struct {
	dir           string
	tmpDir        string
	authPath      string
	logPath       string
	graphDir      string
	mountPoint    string
	testDir       string
	systemMount   string
	systemDataDir string
	systemLog     string
}

func captureSandboxSnapshot() sandboxSnapshot {
	return sandboxSnapshot{
		dir:           testutil.TestSandboxDir,
		tmpDir:        testutil.TestSandboxTmpDir,
		authPath:      testutil.AuthTokensPath,
		logPath:       testutil.TestLogPath,
		graphDir:      testutil.GraphTestDir,
		mountPoint:    testutil.TestMountPoint,
		testDir:       testutil.TestDir,
		systemMount:   testutil.SystemTestMountPoint,
		systemDataDir: testutil.SystemTestDataDir,
		systemLog:     testutil.SystemTestLogPath,
	}
}

func restoreSandboxSnapshot(s sandboxSnapshot) {
	testutil.TestSandboxDir = s.dir
	testutil.TestSandboxTmpDir = s.tmpDir
	testutil.AuthTokensPath = s.authPath
	testutil.TestLogPath = s.logPath
	testutil.GraphTestDir = s.graphDir
	testutil.TestMountPoint = s.mountPoint
	testutil.TestDir = s.testDir
	testutil.SystemTestMountPoint = s.systemMount
	testutil.SystemTestDataDir = s.systemDataDir
	testutil.SystemTestLogPath = s.systemLog
}

func withTempSandbox(t *testing.T, fn func()) {
	t.Helper()
	original := captureSandboxSnapshot()

	tempRoot := t.TempDir()
	tempSandbox := filepath.Join(tempRoot, "sandbox")
	testutil.TestSandboxDir = tempSandbox
	testutil.TestSandboxTmpDir = filepath.Join(tempSandbox, "tmp")
	testutil.AuthTokensPath = filepath.Join(tempSandbox, ".auth_tokens.json")
	testutil.TestLogPath = filepath.Join(tempSandbox, "logs", "fusefs_tests.log")
	testutil.GraphTestDir = filepath.Join(tempSandbox, "graph_test_dir")
	testutil.TestMountPoint = filepath.Join(testutil.TestSandboxTmpDir, "mount")
	testutil.TestDir = filepath.Join(testutil.TestMountPoint, "onemount_tests")
	testutil.SystemTestMountPoint = filepath.Join(testutil.TestSandboxTmpDir, "system-test-mount")
	testutil.SystemTestDataDir = filepath.Join(tempSandbox, "system-test-data")
	testutil.SystemTestLogPath = filepath.Join(tempSandbox, "logs", "system_tests.log")

	t.Cleanup(func() {
		restoreSandboxSnapshot(original)
	})

	if err := helpers.EnsureTestDirectories(); err != nil {
		t.Fatalf("Failed to prepare test directories: %v", err)
	}

	fn()
}

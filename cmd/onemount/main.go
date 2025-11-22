package main

// TODO: Refactor main.go into discrete services (Issue #54)
// This file has grown quite large (~677 lines) and contains multiple responsibilities:
// - Command-line argument parsing and validation
// - Configuration management
// - Authentication handling
// - Filesystem initialization and mounting
// - Statistics display
// - Logging setup
// - Daemon mode handling
// - Signal handling and cleanup
//
// Proposed refactoring for v1.1:
// 1. Extract CLI handling into cmd/onemount/cli/
// 2. Extract filesystem service into cmd/onemount/service/
// 3. Extract statistics service into cmd/onemount/stats/
// 4. Extract daemon handling into cmd/onemount/daemon/
// 5. Keep main.go as a thin coordinator
//
// Target: v1.1 release
// Priority: Medium (architectural improvement, not blocking core functionality)
// Dependencies: None (can be done incrementally)

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/auriora/onemount/cmd/common"
	"github.com/auriora/onemount/internal/errors"
	"github.com/auriora/onemount/internal/fs"
	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
	"github.com/auriora/onemount/internal/metadata"
	"github.com/coreos/go-systemd/v22/unit"
	"github.com/hanwen/go-fuse/v2/fuse"
	flag "github.com/spf13/pflag"
)

func usage() {
	fmt.Printf(`onemount - A Linux client for Microsoft OneDrive.

This program will mount your OneDrive account as a Linux filesystem at the
specified mountpoint. Note that this is not a sync client - files are only
fetched on-demand and cached locally. Only files you actually use will be
downloaded. While offline, the filesystem will be read-only until
connectivity is re-established.

Usage: onemount [options] <mountpoint>

Valid options:
`)
	flag.PrintDefaults()
}

// setupFlags initializes and parses command-line flags, returning the configuration and other flag values
func setupFlags() (config *common.Config, authOnly, headless, debugOn, stats, daemon bool, mountpoint string) {
	// setup cli parsing
	authOnlyFlag := flag.BoolP("auth-only", "a", false,
		"Authenticate to OneDrive and then exit.")
	headlessFlag := flag.BoolP("no-browser", "n", false,
		"This disables launching the built-in web browser during authentication. "+
			"Follow the instructions in the terminal to authenticate to OneDrive.")
	configPath := flag.StringP("config-file", "f", common.DefaultConfigPath(),
		"A YAML-formatted configuration file used by onemount.")
	logLevel := flag.StringP("log", "l", "",
		"Set logging level/verbosity for the filesystem. "+
			"Can be one of: fatal, error, warn, info, debug, trace")
	logOutput := flag.StringP("log-output", "o", "",
		"Set the output location for logs. "+
			"Can be STDOUT, STDERR, or a file path. Default is STDOUT.")
	cacheDir := flag.StringP("cache-dir", "c", "",
		"Change the default cache directory used by onemount. "+
			"Will be created if the path does not already exist.")
	wipeCache := flag.BoolP("wipe-cache", "w", false,
		"Delete the existing onemount cache directory and then exit. "+
			"This is equivalent to resetting the program.")
	versionFlag := flag.BoolP("version", "v", false, "Display program version.")
	debugOnFlag := flag.BoolP("debug", "d", false, "Enable FUSE debug logging. "+
		"This logs communication between onemount and the kernel.")
	syncTree := flag.BoolP("sync-tree", "s", false,
		"Sync the full directory tree to the local metadata store in the background. "+
			"This improves performance by pre-caching directory structure without blocking startup. "+
			"(Enabled by default, use --no-sync-tree to disable)")
	noSyncTree := flag.Bool("no-sync-tree", false,
		"Disable automatic full directory tree synchronization. "+
			"This reduces startup performance but uses less bandwidth and memory.")
	deltaInterval := flag.IntP("delta-interval", "i", 0,
		"Set the interval in seconds between delta query checks. "+
			"Default is 300 seconds (5 minutes) when no realtime subscription is active. "+
			"Set to 0 to use the default.")
	cacheExpiration := flag.IntP("cache-expiration", "e", 0,
		"Set the number of days after which files will be removed from the content cache. "+
			"Default is 30 days. Set to 0 to use the default.")
	cacheCleanupInterval := flag.IntP("cache-cleanup-interval", "", 0,
		"Set the interval in hours between cache cleanup runs. "+
			"Default is 24 hours. Valid range: 1-720 hours (1 hour to 30 days). Set to 0 to use the default.")
	mountTimeout := flag.IntP("mount-timeout", "t", 60,
		"Set the timeout in seconds for mount operations. "+
			"Default is 60 seconds. Increase this if mounting fails due to slow network.")
	hydrationWorkers := flag.Int("hydration-workers", 0, "Number of concurrent hydration/download workers (default 4).")
	hydrationQueueSize := flag.Int("hydration-queue-size", 0, "Maximum queued hydration requests (default 500).")
	metadataWorkers := flag.Int("metadata-workers", 0, "Number of metadata request workers (default 3).")
	metadataHighQueue := flag.Int("metadata-high-queue-size", 0, "High-priority metadata queue size (default 100).")
	metadataLowQueue := flag.Int("metadata-low-queue-size", 0, "Low-priority metadata queue size (default 1000).")
	realtimeFallback := flag.Int("realtime-fallback-seconds", 0, "Override realtime fallback polling interval in seconds (default 1800).")
	overlayPolicy := flag.String("overlay-policy", "", "Default overlay policy (REMOTE_WINS, LOCAL_WINS, MERGED).")
	statsFlag := flag.BoolP("stats", "", false, "Display statistics about the metadata, content caches, "+
		"outstanding changes for upload, etc. Does not start a mount point.")
	pollingOnlyFlag := flag.Bool("polling-only", false, "Force delta polling even if realtime subscriptions are configured (disables the Socket.IO transport).")
	daemonFlag := flag.BoolP("daemon", "", false, "Run onemount in daemon mode (detached from terminal).")
	help := flag.BoolP("help", "h", false, "Displays this help message.")
	flag.Usage = usage
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}
	if *versionFlag {
		fmt.Println("onemount", common.Version())
		os.Exit(0)
	}

	if *wipeCache {
		config = common.LoadConfig(*configPath)
		if *cacheDir != "" {
			config.CacheDir = *cacheDir
		}
		logging.Info().Str("path", config.CacheDir).Msg("Removing cache.")
		if err := os.RemoveAll(config.CacheDir); err != nil {
			logging.Error().Err(err).Msg("Failed to remove cache directory")
		}
		os.Exit(0)
	}

	// determine and validate mountpoint
	if len(flag.Args()) == 0 {
		flag.Usage()
		if _, err := fmt.Fprintf(os.Stderr, "\nNo mountpoint provided, exiting.\n"); err != nil {
			logging.Error().Err(err).Msg("Failed to write to stderr")
		}
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
	if *logOutput != "" {
		config.LogOutput = *logOutput
	}
	// Handle sync tree flags - explicit flags override defaults
	if *syncTree {
		config.SyncTree = true
	}
	if *noSyncTree {
		config.SyncTree = false
	}
	if *deltaInterval > 0 {
		config.DeltaInterval = *deltaInterval
	}
	if *cacheExpiration > 0 {
		config.CacheExpiration = *cacheExpiration
	}
	if *cacheCleanupInterval > 0 {
		config.CacheCleanupInterval = *cacheCleanupInterval
	}
	if *mountTimeout > 0 {
		config.MountTimeout = *mountTimeout
	}
	if *hydrationWorkers > 0 {
		config.Hydration.Workers = *hydrationWorkers
	}
	if *hydrationQueueSize > 0 {
		config.Hydration.QueueSize = *hydrationQueueSize
	}
	if *metadataWorkers > 0 {
		config.MetadataQueue.Workers = *metadataWorkers
	}
	if *metadataHighQueue > 0 {
		config.MetadataQueue.HighPrioritySize = *metadataHighQueue
	}
	if *metadataLowQueue > 0 {
		config.MetadataQueue.LowPrioritySize = *metadataLowQueue
	}
	if *realtimeFallback > 0 {
		config.Realtime.FallbackInterval = *realtimeFallback
	}
	if *overlayPolicy != "" {
		config.Overlay.DefaultPolicy = *overlayPolicy
	}
	if *pollingOnlyFlag {
		config.Realtime.PollingOnly = true
	}

	logging.SetGlobalLevel(common.StringToLevel(config.LogLevel))

	return config, *authOnlyFlag, *headlessFlag, *debugOnFlag, *statsFlag, *daemonFlag, mountpoint
}

// checkConnectivity performs a pre-mount connectivity check to ensure network access
func checkConnectivity(ctx context.Context, timeout time.Duration) error {
	logging.Info().Msg("Performing pre-mount connectivity check...")

	// Create a context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create a simple HTTP client with timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Try to reach the Graph API endpoint
	req, err := http.NewRequestWithContext(checkCtx, "GET", "https://graph.microsoft.com/v1.0/", nil)
	if err != nil {
		return errors.Wrap(err, "failed to create connectivity check request")
	}

	resp, err := client.Do(req)
	if err != nil {
		// Check if it's a timeout or network error
		if checkCtx.Err() == context.DeadlineExceeded {
			return errors.New("connectivity check timed out - network may be slow or unavailable")
		}
		return errors.Wrap(err, "connectivity check failed - cannot reach Microsoft Graph API")
	}
	defer resp.Body.Close()

	// Any response (even 401) means we can reach the API
	logging.Info().
		Int("statusCode", resp.StatusCode).
		Msg("Connectivity check successful")

	return nil
}

// initializeFilesystem sets up the filesystem and returns the filesystem, auth, server, and paths
func initializeFilesystem(ctx context.Context, config *common.Config, mountpoint string, authOnly, headless, debugOn bool) (*fs.Filesystem, *graph.Auth, *fuse.Server, string, string, error) {
	// compute cache name as systemd would
	absMountPath, err := filepath.Abs(mountpoint)
	if err != nil {
		return nil, nil, nil, "", "", errors.Wrap(err, "failed to get absolute path for mountpoint")
	}
	cachePath := filepath.Join(config.CacheDir, unit.UnitNamePathEscape(absMountPath))

	// Configure D-Bus service name deterministically for this mountpoint before the filesystem starts
	fs.SetDBusServiceNameForMount(absMountPath)

	// Perform connectivity check before attempting mount
	connectivityTimeout := time.Duration(config.MountTimeout/2) * time.Second
	if connectivityTimeout < 10*time.Second {
		connectivityTimeout = 10 * time.Second
	}

	if err := checkConnectivity(ctx, connectivityTimeout); err != nil {
		logging.Warn().Err(err).Msg("Connectivity check failed, but continuing with mount attempt")
		// Don't fail here - just warn. The mount may still succeed if it's a transient issue.
	}

	// authenticate/re-authenticate if necessary
	if err := os.MkdirAll(cachePath, 0700); err != nil {
		return nil, nil, nil, "", "", errors.Wrap(err, "failed to create cache directory")
	}
	authPath := graph.GetAuthTokensPathFromCacheDir(cachePath)

	// Apply runtime tunables before filesystem construction
	fs.SetHydrationDefaults(config.Hydration.Workers, config.Hydration.QueueSize)
	fs.SetMetadataQueueDefaults(config.MetadataQueue.Workers, config.MetadataQueue.HighPrioritySize, config.MetadataQueue.LowPrioritySize)

	if authOnly {
		if err := os.Remove(authPath); err != nil && !os.IsNotExist(err) {
			logging.LogError(err, "Failed to remove auth tokens file",
				logging.FieldOperation, "initializeFilesystem",
				logging.FieldPath, authPath)
		}
		_, err := graph.Authenticate(context.Background(), config.AuthConfig, authPath, headless)
		if err != nil {
			logging.LogError(err, "Authentication failed",
				logging.FieldOperation, "initializeFilesystem",
				logging.FieldPath, authPath)
			return nil, nil, nil, "", "", errors.Wrap(err, "authentication failed")
		}
		os.Exit(0)
	}

	// create the filesystem
	logging.Info().Msgf("onemount %s", common.Version())
	auth, err := graph.Authenticate(context.Background(), config.AuthConfig, authPath, headless)
	if err != nil {
		logging.LogError(err, "Authentication failed",
			logging.FieldOperation, "initializeFilesystem",
			logging.FieldPath, authPath)
		return nil, nil, nil, "", "", errors.Wrap(err, "authentication failed")
	}

	filesystem, err := fs.NewFilesystemWithContext(ctx, auth, cachePath, config.CacheExpiration, config.CacheCleanupInterval, config.MaxCacheSize)
	if err != nil {
		logging.LogError(err, "Failed to initialize filesystem",
			logging.FieldOperation, "initializeFilesystem",
			logging.FieldPath, cachePath)
		return nil, nil, nil, "", "", errors.Wrap(err, "failed to initialize filesystem")
	}

	realtimeOpts := toRealtimeOptions(config.Realtime)
	if realtimeOpts.Enabled {
		filesystem.ConfigureRealtime(realtimeOpts)
	}
	filesystem.SetDefaultOverlayPolicy(metadata.OverlayPolicy(strings.ToUpper(config.Overlay.DefaultPolicy)))

	filesystem.ConfigureDeltaTuning(fs.DeltaTuning{
		ActiveInterval: time.Duration(config.ActiveDeltaInterval) * time.Second,
		ActiveWindow:   time.Duration(config.ActiveDeltaWindow) * time.Second,
	})

	logging.Info().Msgf("Setting base delta query interval to %d second(s)", config.DeltaInterval)
	go filesystem.DeltaLoop(time.Duration(config.DeltaInterval) * time.Second)

	// Start the content cache cleanup routine
	if config.CacheExpiration > 0 {
		logging.Info().Msgf("Setting content cache expiration to %d day(s)", config.CacheExpiration)
		filesystem.StartCacheCleanup()
	}

	// Start the status cache cleanup routine
	filesystem.StartStatusCacheCleanup()

	common.CreateXDGVolumeInfo(filesystem, auth)

	// Sync the full directory tree if requested
	if config.SyncTree {
		logging.Info().Msg("Starting full directory tree synchronization in background...")
		filesystem.Wg.Add(1)
		go func(ctx context.Context) {
			defer filesystem.Wg.Done()

			// Check if context is already cancelled
			select {
			case <-ctx.Done():
				logging.Debug().Msg("Directory tree synchronization cancelled due to context cancellation")
				return
			default:
				// Continue with normal operation
			}

			if err := filesystem.SyncDirectoryTreeWithContext(ctx, auth); err != nil {
				// Check if the error is due to context cancellation
				if ctx.Err() != nil {
					logging.Debug().Msg("Directory tree synchronization cancelled due to context cancellation")
					return
				}
				logging.LogError(err, "Error syncing directory tree",
					logging.FieldOperation, "SyncDirectoryTreeWithContext")
			} else {
				logging.Info().Msg("Directory tree sync completed successfully")
			}
		}(ctx)
	}

	// Create mount options
	mountOptions := &fuse.MountOptions{
		Name:          "onemount",
		FsName:        "onemount",
		DisableXAttrs: false,
		MaxBackground: 1024,
		Debug:         debugOn,
	}

	// Only set AllowOther if user_allow_other is enabled in /etc/fuse.conf
	if common.IsUserAllowOtherEnabled() {
		logging.Info().Msg("Setting AllowOther mount option (user_allow_other is enabled in /etc/fuse.conf)")
		mountOptions.AllowOther = true
	} else {
		logging.Info().Msg("Not setting AllowOther mount option (user_allow_other is not enabled in /etc/fuse.conf)")
	}

	// Create the FUSE server
	server, err := fuse.NewServer(filesystem, mountpoint, mountOptions)
	if err != nil {
		logging.LogError(err, fmt.Sprintf("Mount failed. Is the mountpoint already in use? (Try running \"fusermount3 -uz %s\")", mountpoint),
			logging.FieldOperation, "NewServer",
			logging.FieldPath, mountpoint)
		return nil, nil, nil, "", "", errors.Wrap(err, "mount failed (is the mountpoint already in use?)")
	}

	return filesystem, auth, server, cachePath, absMountPath, nil
}

func toRealtimeOptions(cfg common.RealtimeConfig) fs.RealtimeOptions {
	return fs.RealtimeOptions{
		Enabled:          cfg.Enabled,
		PollingOnly:      cfg.PollingOnly,
		ClientState:      cfg.ClientState,
		Resource:         cfg.Resource,
		FallbackInterval: time.Duration(cfg.FallbackInterval) * time.Second,
	}
}

// displayStats gathers and displays statistics about the filesystem
func displayStats(ctx context.Context, config *common.Config, mountpoint string) {
	// Determine the cache directory
	if mountpoint == "" {
		logging.Fatal().Msg("No mountpoint specified. Please provide a mountpoint.")
	}
	absMountPath, _ := filepath.Abs(mountpoint)
	cachePath := filepath.Join(config.CacheDir, unit.UnitNamePathEscape(absMountPath))

	// Ensure deterministic D-Bus name for stats mode as well (even though the service won't be awaited).
	fs.SetDBusServiceNameForMount(absMountPath)

	// Authenticate to get access to the filesystem
	authPath := graph.GetAuthTokensPathFromCacheDir(cachePath)
	auth, err := graph.Authenticate(ctx, config.AuthConfig, authPath, true)
	if err != nil {
		logging.Error().Err(err).Msg("Authentication failed")
		os.Exit(1)
	}

	// Initialize the filesystem without mounting
	filesystem, err := fs.NewFilesystemWithContext(ctx, auth, cachePath, config.CacheExpiration, config.CacheCleanupInterval, config.MaxCacheSize)
	if err != nil {
		logging.Error().Err(err).Msg("Failed to initialize filesystem")
		os.Exit(1)
	}

	// Get statistics
	stats, err := filesystem.GetStats()
	if err != nil {
		logging.Error().Err(err).Msg("Failed to get statistics")
		os.Exit(1)
	}

	// Display statistics header
	fmt.Println("onemount Statistics")
	fmt.Println("===================")

	// Metadata statistics
	fmt.Printf("\nMetadata Cache:\n")
	fmt.Printf("  Items in memory: %d\n", stats.MetadataCount)

	// Content cache statistics
	fmt.Printf("\nContent Cache:\n")
	fmt.Printf("  Files: %d\n", stats.ContentCount)
	fmt.Printf("  Total size: %s\n", fs.FormatSize(stats.ContentSize))
	fmt.Printf("  Cache directory: %s\n", stats.ContentDir)
	fmt.Printf("  Expiration: %d days\n", stats.Expiration)

	// Upload queue statistics
	fmt.Printf("\nUpload Queue:\n")
	fmt.Printf("  Total uploads: %d\n", stats.UploadCount)
	fmt.Printf("  Not started: %d\n", stats.UploadsNotStarted)
	fmt.Printf("  In progress: %d\n", stats.UploadsInProgress)
	fmt.Printf("  Completed: %d\n", stats.UploadsCompleted)
	fmt.Printf("  Errors: %d\n", stats.UploadsErrored)

	// Hydration/download queue statistics
	fmt.Printf("\nHydration Queue:\n")
	fmt.Printf("  Queue depth: %d\n", stats.HydrationQueueDepth)
	fmt.Printf("  Active downloads: %d\n", stats.HydrationActiveDownloads)
	fmt.Printf("  Hydrated items: %d\n", stats.HydrationHydrated)
	fmt.Printf("  Hydrating items: %d\n", stats.HydrationHydrating)
	fmt.Printf("  Ghost items: %d\n", stats.HydrationGhost)
	fmt.Printf("  Dirty local items: %d\n", stats.HydrationDirtyLocal)
	fmt.Printf("  Errored items: %d\n", stats.HydrationErrored)

	// Metadata request queue statistics
	fmt.Printf("\nMetadata Request Queue:\n")
	fmt.Printf("  High-priority depth: %d\n", stats.MetadataQueueHighDepth)
	fmt.Printf("  Low-priority depth: %d\n", stats.MetadataQueueLowDepth)
	fmt.Printf("  Avg wait (ms): %.2f\n", stats.MetadataQueueAvgWaitMs)

	// File status statistics
	fmt.Printf("\nFile Statuses:\n")
	fmt.Printf("  Cloud: %d\n", stats.StatusCloud)
	fmt.Printf("  Local: %d\n", stats.StatusLocal)
	fmt.Printf("  LocalModified: %d\n", stats.StatusLocalModified)
	fmt.Printf("  Syncing: %d\n", stats.StatusSyncing)
	fmt.Printf("  Downloading: %d\n", stats.StatusDownloading)
	fmt.Printf("  OutofSync: %d\n", stats.StatusOutofSync)
	fmt.Printf("  Error: %d\n", stats.StatusError)
	fmt.Printf("  Conflict: %d\n", stats.StatusConflict)

	// Delta link information
	fmt.Printf("\nDelta Link:\n")
	fmt.Printf("  %s\n", stats.DeltaLink)

	// Offline status
	fmt.Printf("\nOffline Status: %v\n", stats.IsOffline)

	// Realtime transport status
	fmt.Printf("\nRealtime Notifications:\n")
	fmt.Printf("  Mode: %s\n", stats.RealtimeMode)
	fmt.Printf("  Status: %s\n", string(stats.RealtimeStatus))
	fmt.Printf("  Missed heartbeats: %d\n", stats.RealtimeMissedHeartbeats)
	fmt.Printf("  Consecutive failures: %d\n", stats.RealtimeConsecutiveFailures)
	fmt.Printf("  Reconnect count: %d\n", stats.RealtimeReconnectCount)
	if !stats.RealtimeLastHeartbeat.IsZero() {
		fmt.Printf("  Last heartbeat: %s\n", stats.RealtimeLastHeartbeat.Format(time.RFC3339))
	}
	if stats.RealtimeLastError != "" {
		fmt.Printf("  Last error: %s\n", stats.RealtimeLastError)
	}

	// BBolt database statistics
	fmt.Printf("\nBBolt Database:\n")
	fmt.Printf("  Database path: %s\n", stats.DBPath)
	fmt.Printf("  Database size: %s\n", fs.FormatSize(stats.DBSize))
	fmt.Printf("  Page count: %d\n", stats.DBPageCount)
	fmt.Printf("  Page size: %s\n", fs.FormatSize(int64(stats.DBPageSize)))
	fmt.Printf("  Metadata items: %d\n", stats.DBMetadataCount)
	fmt.Printf("  Delta items: %d\n", stats.DBDeltaCount)
	fmt.Printf("  Offline changes: %d\n", stats.DBOfflineCount)
	fmt.Printf("  Upload records: %d\n", stats.DBUploadsCount)

	// Directory statistics derived from metadata
	fmt.Printf("\nDirectory Statistics:\n")
	fmt.Printf("  Total directories: %d\n", stats.DirCount)
	fmt.Printf("  Empty directories: %d\n", stats.EmptyDirCount)
	fmt.Printf("  Maximum directory depth: %d\n", stats.MaxDirDepth)
	fmt.Printf("  Average directory depth: %.2f\n", stats.AvgDirDepth)
	fmt.Printf("  Average files per directory: %.2f\n", stats.AvgFilesPerDir)
	fmt.Printf("  Maximum files in a directory: %d\n", stats.MaxFilesInDir)

	// File type statistics
	if len(stats.FileExtensions) > 0 {
		fmt.Printf("\nFile Type Distribution:\n")
		// Sort extensions for consistent display
		extensions := make([]string, 0, len(stats.FileExtensions))
		for ext := range stats.FileExtensions {
			extensions = append(extensions, ext)
		}
		sort.Strings(extensions)

		for _, ext := range extensions {
			fmt.Printf("  %s: %d\n", ext, stats.FileExtensions[ext])
		}
	}

	// File size statistics
	if len(stats.FileSizeRanges) > 0 {
		fmt.Printf("\nFile Size Distribution:\n")
		// Define order for size ranges
		sizeRangeOrder := []string{
			"Empty (0 bytes)",
			"< 1 KB",
			"1 KB - 1 MB",
			"1 MB - 10 MB",
			"10 MB - 100 MB",
			"100 MB - 1 GB",
			"> 1 GB",
		}

		for _, sizeRange := range sizeRangeOrder {
			if count, exists := stats.FileSizeRanges[sizeRange]; exists {
				fmt.Printf("  %s: %d\n", sizeRange, count)
			}
		}
	}

	// File age statistics
	if len(stats.FileAgeRanges) > 0 {
		fmt.Printf("\nFile Age Distribution:\n")
		// Define order for age ranges
		ageRangeOrder := []string{
			"Today",
			"This week",
			"This month",
			"Last 3 months",
			"This year",
			"Older than a year",
		}

		for _, ageRange := range ageRangeOrder {
			if count, exists := stats.FileAgeRanges[ageRange]; exists {
				fmt.Printf("  %s: %d\n", ageRange, count)
			}
		}
	}

	// Clean up
	filesystem.StopCacheCleanup()
	filesystem.StopDeltaLoop()
	filesystem.StopDownloadManager()
	filesystem.StopUploadManager()
}

// setupLogging configures the logger based on the configuration
func setupLogging(config *common.Config, daemon bool) error {
	// Set the global log level
	logging.SetGlobalLevel(common.StringToLevel(config.LogLevel))

	// Configure the log output
	var output io.Writer

	// If running in daemon mode and no specific log file is set, use a default log file
	if daemon && (config.LogOutput == "STDOUT" || config.LogOutput == "STDERR") {
		// Use a default log file in the cache directory
		logFile := filepath.Join(config.CacheDir, "onemount.log")
		logging.Info().Str("logFile", logFile).Msg("Daemon mode: redirecting logs to file")

		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			logging.Error().Err(err).Str("path", logFile).Msg("Failed to open log file, falling back to STDOUT")
			output = os.Stdout
		} else {
			output = file
			// Update the config to reflect the actual log output
			config.LogOutput = logFile
		}
	} else {
		// Normal logging setup
		switch config.LogOutput {
		case "STDOUT":
			output = os.Stdout
		case "STDERR":
			output = os.Stderr
		default:
			// Open the log file
			file, err := os.OpenFile(config.LogOutput, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				logging.Error().Err(err).Str("path", config.LogOutput).Msg("Failed to open log file, falling back to STDOUT")
				output = os.Stdout
			} else {
				output = file
			}
		}
	}

	// Set up the logger with console formatting
	logging.DefaultLogger = logging.New(logging.NewConsoleWriterWithOptions(output, logging.HumanReadableTimeFormat))
	return nil
}

func main() {
	// Initialize with a basic logger that outputs to stderr
	// This will be replaced after loading the configuration
	logging.DefaultLogger = logging.New(logging.NewConsoleWriterWithOptions(os.Stderr, logging.HumanReadableTimeFormat))

	// Create a root context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, authOnly, headless, debugOn, stats, daemon, mountpoint := setupFlags()

	// Configure logging based on the configuration
	if err := setupLogging(config, daemon); err != nil {
		logging.Error().Err(err).Msg("Failed to set up logging")
	}

	// If daemon flag is set, daemonize the process
	if daemon {
		logging.Info().Msg("Starting onemount in daemon mode...")
		daemonize()
	}

	// If stats flag is set, display statistics and exit
	if stats {
		displayStats(ctx, config, mountpoint)
		os.Exit(0)
	}

	// Check if the mountpoint might be a mistyped flag
	if len(mountpoint) == 1 && strings.Contains("acdefhilnsvw", mountpoint) {
		logging.Fatal().
			Str("mountpoint", mountpoint).
			Msg("Mountpoint looks like a flag without the hyphen prefix. Did you mean '-" + mountpoint + "'? Use '--help' for usage information.")
	}

	st, err := os.Stat(mountpoint)
	if err != nil || !st.IsDir() {
		common.HandleErrorAndExit(
			fmt.Errorf("mountpoint '%s' did not exist or was not a directory", mountpoint),
			1)
	}
	if res, _ := os.ReadDir(mountpoint); len(res) > 0 {
		common.HandleErrorAndExit(
			fmt.Errorf("mountpoint '%s' must be empty", mountpoint),
			1)
	}

	// Check if the mountpoint is already mounted
	if isMounted := checkIfMounted(mountpoint); isMounted {
		common.HandleErrorAndExit(
			fmt.Errorf("mountpoint '%s' is already mounted. Unmount it first or choose a different mountpoint", mountpoint),
			1)
	}

	// Initialize the filesystem
	filesystem, _, server, cachePath, absMountPath, err := initializeFilesystem(ctx, config, mountpoint, authOnly, headless, debugOn)
	if err != nil {
		common.HandleErrorAndExit(err, 1)
	}

	// setup signal handler for graceful unmount on signals like sigint
	setupSignalHandler(filesystem, server, absMountPath, cancel)

	// serve filesystem
	logging.Info().
		Str("cachePath", cachePath).
		Str("mountpoint", absMountPath).
		Msg("Serving filesystem.")
	server.Serve()
}

// isMountpointMounted checks if a filesystem is mounted at the given mountpoint
func isMountpointMounted(mountpoint string) bool {
	if mountpoint == "" {
		return false
	}

	// Check if it's a mount point using findmnt
	cmd := exec.Command("findmnt", "--noheadings", "--output", "TARGET", mountpoint)
	if output, err := cmd.Output(); err == nil && len(output) > 0 {
		return true
	}

	return false
}

// checkIfMounted checks if a filesystem is already mounted at the given mountpoint
func checkIfMounted(mountpoint string) bool {
	// Check if it's a mount point using findmnt
	cmd := exec.Command("findmnt", "--noheadings", "--output", "TARGET", mountpoint)
	if output, err := cmd.Output(); err == nil && len(output) > 0 {
		logging.Warn().Str("mountpoint", mountpoint).Msg("Mount point is already mounted")
		return true
	}

	// Additional check: try to create and remove a test file
	// If the mountpoint is already mounted but empty, the previous check might not catch it
	testFile := filepath.Join(mountpoint, ".onemount-mount-test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		logging.Warn().Err(err).Str("mountpoint", mountpoint).Msg("Failed to write test file, mountpoint might be mounted or inaccessible")
		return true
	}

	// Clean up the test file
	if err := os.Remove(testFile); err != nil {
		logging.Warn().Err(err).Str("mountpoint", mountpoint).Msg("Failed to remove test file, but mountpoint appears accessible")
	}

	return false
}

// daemonize detaches the process from the terminal and runs it in the background
func daemonize() {
	// Fork the process
	args := os.Args[:]

	// Remove the daemon flag to prevent infinite forking
	for i, arg := range args {
		if arg == "--daemon" {
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	// Prepare the command to run in the background
	cmd := exec.Command(args[0])
	if len(args) > 1 {
		cmd.Args = args
	}

	// Detach from terminal
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Create a new process group
		Setpgid: true,
		// Detach from parent
		Setsid: true,
	}

	// Start the process in the background
	if err := cmd.Start(); err != nil {
		logging.Fatal().Err(err).Msg("Failed to start daemon process")
	}

	logging.Info().Msg("Daemon process started successfully")
	os.Exit(0)
}

// setupSignalHandler sets up a handler for SIGINT and SIGTERM signals to gracefully unmount the filesystem
func setupSignalHandler(filesystem *fs.Filesystem, server *fuse.Server, mountpoint string, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a custom signal handler that stops background processes before unmounting
	go func() {
		sig := <-sigChan // block until signal
		logging.Info().Str("signal", strings.ToUpper(sig.String())).
			Msg("Signal received, cleaning up and unmounting filesystem.")

		// Cancel the context to notify all goroutines to stop
		logging.Info().Msg("Canceling context to notify all goroutines to stop...")
		cancel()

		// Stop the cache cleanup routine
		filesystem.StopCacheCleanup()

		// Stop the delta loop
		filesystem.StopDeltaLoop()

		// Stop the download manager
		filesystem.StopDownloadManager()

		// Stop the upload manager
		filesystem.StopUploadManager()

		// Stop the metadata request manager
		filesystem.StopMetadataRequestManager()

		// Give the system a moment to release all resources
		logging.Info().Msg("Waiting for all resources to be released before unmounting...")
		time.Sleep(500 * time.Millisecond)

		// Unmount the filesystem with retries
		maxRetries := 3
		retryDelay := 500 * time.Millisecond
		var err error

		// Check if the filesystem is actually mounted before attempting to unmount
		if !isMountpointMounted(mountpoint) {
			logging.Warn().Str("mountpoint", mountpoint).Msg("Filesystem does not appear to be mounted, skipping unmount operation")
		} else {
			for i := 0; i < maxRetries; i++ {
				err = server.Unmount()
				if err == nil {
					break
				}

				if i < maxRetries-1 {
					logging.Warn().Err(err).
						Int("retry", i+1).
						Dur("delay", retryDelay).
						Msg("Failed to unmount filesystem, retrying after delay...")
					time.Sleep(retryDelay)
					retryDelay *= 2 // Exponential backoff
				}
			}
		}

		if err != nil {
			logging.Error().Err(err).Msg("Failed to unmount filesystem cleanly after multiple attempts! " +
				"Run \"fusermount3 -uz /MOUNTPOINT/GOES/HERE\" to unmount.")
			os.Exit(1) // Exit with error code 1 to indicate failure
		} else {
			logging.Info().Msg("Filesystem unmounted successfully.")
			os.Exit(0) // Exit with success code 0
		}
	}()
}

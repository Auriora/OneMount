package fs

import (
	"fmt"
	"os"
	"sync"

	"github.com/auriora/onemount/internal/logging"
	"github.com/coreos/go-systemd/v22/unit"
	dbus "github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

const (
	// DBusInterface is the D-Bus interface name for onemount
	DBusInterface = "org.onemount.FileStatus"
	// DBusObjectPath is the D-Bus object path for onemount
	DBusObjectPath = "/org/onemount/FileStatus"
	// DBusServiceNameBase is the base D-Bus service name for onemount
	DBusServiceNameBase = "org.onemount.FileStatus"
	// DBusServiceNameFile is the file where the D-Bus service name is written for discovery
	DBusServiceNameFile = "/tmp/onemount-dbus-service-name"
)

// DBusServiceName returns the D-Bus service name, which may be unique in test environments
var DBusServiceName string

// SetDBusServiceNamePrefix sets DBusServiceName with the given suffix component (no randomness).
// Tests can supply custom prefixes, and production code can use this helper via higher-level functions.
func SetDBusServiceNamePrefix(prefix string) {
	if prefix == "" {
		prefix = "instance"
	}
	DBusServiceName = fmt.Sprintf("%s.%s", DBusServiceNameBase, prefix)
	logging.Debug().Str("dbusName", DBusServiceName).Msg("Using deterministic D-Bus service name")
}

// SetDBusServiceNameForMount derives a deterministic D-Bus service name from the mount path.
func SetDBusServiceNameForMount(mountPath string) {
	escaped := unit.UnitNamePathEscape(mountPath)
	if escaped == "" {
		escaped = "root"
	}
	SetDBusServiceNamePrefix("mnt_" + escaped)
}

func init() {
	// Initialize the DBusServiceName variable with the default prefix
	SetDBusServiceNamePrefix("instance")
}

// FileStatusDBusServer implements a D-Bus server for file status updates
type FileStatusDBusServer struct {
	fs       FilesystemInterface
	conn     *dbus.Conn
	mutex    sync.RWMutex
	started  bool
	stopChan chan struct{}
}

// NewFileStatusDBusServer creates a new D-Bus server for file status updates
func NewFileStatusDBusServer(fs FilesystemInterface) *FileStatusDBusServer {
	return &FileStatusDBusServer{
		fs:       fs,
		stopChan: make(chan struct{}),
	}
}

// StartForTesting starts the D-Bus server in test mode
// This method is used for testing purposes only and doesn't try to register a service name
func (s *FileStatusDBusServer) StartForTesting() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.started {
		return nil
	}

	conn, err := dbus.SessionBus()
	if err != nil {
		logging.Error().Err(err).Msg("Failed to connect to D-Bus session bus")
		return err
	}
	s.conn = conn

	// Export the FileStatusDBusServer object
	err = conn.Export(s, DBusObjectPath, DBusInterface)
	if err != nil {
		logging.Error().Err(err).Msg("Failed to export D-Bus object")
		s.conn = nil
		return err
	}

	// Export the introspection data
	node := &introspect.Node{
		Name: DBusObjectPath,
		Interfaces: []introspect.Interface{
			{
				Name: DBusInterface,
				Methods: []introspect.Method{
					{
						Name: "GetFileStatus",
						Args: []introspect.Arg{
							{Name: "path", Type: "s", Direction: "in"},
							{Name: "status", Type: "s", Direction: "out"},
						},
					},
				},
				Signals: []introspect.Signal{
					{
						Name: "FileStatusChanged",
						Args: []introspect.Arg{
							{Name: "path", Type: "s"},
							{Name: "status", Type: "s"},
						},
					},
				},
			},
		},
	}
	err = conn.Export(introspect.NewIntrospectable(node), DBusObjectPath, "org.freedesktop.DBus.Introspectable")
	if err != nil {
		logging.Error().Err(err).Msg("Failed to export introspection data")
		s.conn = nil
		return err
	}

	s.started = true
	logging.Info().Msg("D-Bus server started in test mode")
	return nil
}

// Start starts the D-Bus server
func (s *FileStatusDBusServer) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.started {
		return nil
	}

	conn, err := dbus.SessionBus()
	if err != nil {
		logging.Error().Err(err).Msg("Failed to connect to D-Bus session bus")
		return err
	}
	s.conn = conn

	// Request a name on the bus with flags to allow replacement and not queue
	// This ensures we can always get a name, even if there are conflicts
	reply, err := conn.RequestName(DBusServiceName, dbus.NameFlagAllowReplacement|dbus.NameFlagReplaceExisting|dbus.NameFlagDoNotQueue)
	if err != nil {
		logging.Error().Err(err).Msg("Failed to request D-Bus name")
		s.conn = nil
		return err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		// Since we're using a unique name and NameFlagReplaceExisting, this should rarely happen
		// But if it does, we'll log it and continue
		logging.Warn().Msgf("Not primary owner of D-Bus name (reply: %v), but continuing with unique name: %s", reply, DBusServiceName)
	} else {
		logging.Debug().Str("dbusName", DBusServiceName).Msg("Successfully acquired D-Bus name")
	}

	// Export the FileStatusDBusServer object
	err = conn.Export(s, DBusObjectPath, DBusInterface)
	if err != nil {
		logging.Error().Err(err).Msg("Failed to export D-Bus object")
		s.conn = nil
		return err
	}

	// Export the introspection data
	node := &introspect.Node{
		Name: DBusObjectPath,
		Interfaces: []introspect.Interface{
			{
				Name: DBusInterface,
				Methods: []introspect.Method{
					{
						Name: "GetFileStatus",
						Args: []introspect.Arg{
							{Name: "path", Type: "s", Direction: "in"},
							{Name: "status", Type: "s", Direction: "out"},
						},
					},
				},
				Signals: []introspect.Signal{
					{
						Name: "FileStatusChanged",
						Args: []introspect.Arg{
							{Name: "path", Type: "s"},
							{Name: "status", Type: "s"},
						},
					},
				},
			},
		},
	}
	err = conn.Export(introspect.NewIntrospectable(node), DBusObjectPath, "org.freedesktop.DBus.Introspectable")
	if err != nil {
		logging.Error().Err(err).Msg("Failed to export introspection data")
		s.conn = nil
		return err
	}

	// Write the service name to a file for discovery by clients (e.g., Nemo extension)
	// This allows clients to discover the actual service name even when it includes a unique suffix
	if err := s.writeServiceNameFile(); err != nil {
		// Log warning but don't fail - clients can still use extended attributes as fallback
		logging.Warn().Err(err).Msg("Failed to write D-Bus service name file")
	}

	s.started = true
	logging.Info().Msg("D-Bus server started")
	return nil
}

// Stop stops the D-Bus server and cleans up all resources
func (s *FileStatusDBusServer) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.started {
		return
	}

	// Don't close the channel again if it's already closed
	select {
	case <-s.stopChan:
		// Channel is already closed
	default:
		close(s.stopChan)
	}

	if s.conn != nil {
		// Release the D-Bus name before closing the connection
		// This helps prevent name conflicts in subsequent runs
		logging.Debug().Str("dbusName", DBusServiceName).Msg("Releasing D-Bus name")
		if _, err := s.conn.ReleaseName(DBusServiceName); err != nil {
			logging.Warn().Err(err).Msg("Failed to release D-Bus name")
		}

		// Unexport the objects to clean up resources
		if err := s.conn.Export(nil, DBusObjectPath, DBusInterface); err != nil {
			logging.Warn().Err(err).Msg("Failed to unexport D-Bus object")
		}
		if err := s.conn.Export(nil, DBusObjectPath, "org.freedesktop.DBus.Introspectable"); err != nil {
			logging.Warn().Err(err).Msg("Failed to unexport introspection data")
		}

		// Close the connection
		if err := s.conn.Close(); err != nil {
			logging.Error().Err(err).Msg("Failed to close D-Bus connection")
		}
		s.conn = nil
	}

	// Remove the service name file
	if err := s.removeServiceNameFile(); err != nil {
		logging.Warn().Err(err).Msg("Failed to remove D-Bus service name file")
	}

	s.started = false
	logging.Info().Msg("D-Bus server stopped and resources cleaned up")
}

// GetFileStatus returns the status of a file by path
func (s *FileStatusDBusServer) GetFileStatus(path string) (string, *dbus.Error) {
	// Find the inode for this path by traversing the filesystem tree
	inode := s.findInodeByPath(path)
	if inode == nil {
		logging.Debug().Str("path", path).Msg("File not found in filesystem")
		return "Unknown", nil
	}

	// Get the file status for this inode
	status := s.fs.GetFileStatus(inode.ID())

	// Convert FileStatusInfo to string representation
	statusStr := status.Status.String()

	logging.Debug().
		Str("path", path).
		Str("id", inode.ID()).
		Str("status", statusStr).
		Msg("Retrieved file status via D-Bus method")

	return statusStr, nil
}

// findInodeByPath traverses the filesystem tree to find an inode by its path
// The path should be relative to the mount point (e.g., "/Documents/file.txt")
func (s *FileStatusDBusServer) findInodeByPath(path string) *Inode {
	// Get the root inode - need to cast to *Filesystem to access the root field
	fs, ok := s.fs.(*Filesystem)
	if !ok {
		logging.Warn().Msg("Filesystem is not of type *Filesystem")
		return nil
	}

	if path == "" || path == "/" {
		// Root path - get the root inode using the filesystem's root ID
		return fs.GetID(fs.root)
	}

	// Split the path into components
	// Remove leading slash if present
	if path[0] == '/' {
		path = path[1:]
	}

	// If path is now empty, it was just "/"
	if path == "" {
		return fs.GetID(fs.root)
	}

	// Split by '/' to get path components
	components := splitPath(path)
	if len(components) == 0 {
		return fs.GetID(fs.root)
	}

	// Start from root and traverse down
	current := fs.GetID(fs.root)
	if current == nil {
		logging.Warn().Msg("Root inode not found")
		return nil
	}

	// Traverse each component
	for i, component := range components {
		// Look for this component in the current directory's children
		found := false
		children := current.GetChildren()
		for _, childID := range children {
			child := s.fs.GetID(childID)
			if child != nil && child.Name() == component {
				current = child
				found = true
				break
			}
		}

		if !found {
			logging.Debug().
				Str("path", path).
				Str("component", component).
				Int("depth", i).
				Msg("Path component not found in filesystem")
			return nil
		}
	}

	return current
}

// splitPath splits a path into components, handling empty strings
func splitPath(path string) []string {
	if path == "" {
		return []string{}
	}

	components := []string{}
	start := 0

	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				components = append(components, path[start:i])
			}
			start = i + 1
		}
	}

	// Add the last component if there is one
	if start < len(path) {
		components = append(components, path[start:])
	}

	return components
}

// SendFileStatusUpdate sends a D-Bus signal with the updated file status
func (s *FileStatusDBusServer) SendFileStatusUpdate(path string, status string) {
	if !s.started || s.conn == nil {
		return
	}

	err := s.conn.Emit(
		DBusObjectPath,
		DBusInterface+".FileStatusChanged",
		path,
		status,
	)
	if err != nil {
		logging.Error().Err(err).Str("path", path).Str("status", status).Msg("Failed to emit D-Bus signal")
	}
}

// writeServiceNameFile writes the D-Bus service name to a file for discovery by clients
func (s *FileStatusDBusServer) writeServiceNameFile() error {
	// Write the service name to a temporary file first, then rename atomically
	tempFile := DBusServiceNameFile + ".tmp"

	// Create the file with restricted permissions (only owner can read/write)
	f, err := os.OpenFile(tempFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create service name file: %w", err)
	}
	defer f.Close()

	// Write the service name
	if _, err := f.WriteString(DBusServiceName + "\n"); err != nil {
		os.Remove(tempFile) // Clean up temp file on error
		return fmt.Errorf("failed to write service name: %w", err)
	}

	// Sync to ensure data is written to disk
	if err := f.Sync(); err != nil {
		os.Remove(tempFile) // Clean up temp file on error
		return fmt.Errorf("failed to sync service name file: %w", err)
	}

	// Close the file before renaming
	if err := f.Close(); err != nil {
		os.Remove(tempFile) // Clean up temp file on error
		return fmt.Errorf("failed to close service name file: %w", err)
	}

	// Atomically rename the temp file to the final location
	if err := os.Rename(tempFile, DBusServiceNameFile); err != nil {
		os.Remove(tempFile) // Clean up temp file on error
		return fmt.Errorf("failed to rename service name file: %w", err)
	}

	logging.Debug().
		Str("file", DBusServiceNameFile).
		Str("serviceName", DBusServiceName).
		Msg("Wrote D-Bus service name to file for client discovery")

	return nil
}

// removeServiceNameFile removes the D-Bus service name file
func (s *FileStatusDBusServer) removeServiceNameFile() error {
	// Only remove the file if it contains our service name
	// This prevents removing a file written by another instance
	data, err := os.ReadFile(DBusServiceNameFile)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, nothing to do
			return nil
		}
		return fmt.Errorf("failed to read service name file: %w", err)
	}

	// Check if the file contains our service name
	storedName := string(data)
	// Trim whitespace and newlines
	storedName = storedName[:len(storedName)-1] // Remove trailing newline
	if storedName != DBusServiceName {
		// File contains a different service name, don't remove it
		logging.Debug().
			Str("file", DBusServiceNameFile).
			Str("storedName", storedName).
			Str("ourName", DBusServiceName).
			Msg("Service name file contains different name, not removing")
		return nil
	}

	// Remove the file
	if err := os.Remove(DBusServiceNameFile); err != nil {
		if os.IsNotExist(err) {
			// File was already removed, nothing to do
			return nil
		}
		return fmt.Errorf("failed to remove service name file: %w", err)
	}

	logging.Debug().
		Str("file", DBusServiceNameFile).
		Msg("Removed D-Bus service name file")

	return nil
}

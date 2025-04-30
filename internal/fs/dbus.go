package fs

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/rs/zerolog/log"
)

const (
	// DBusInterface is the D-Bus interface name for onemount
	DBusInterface = "org.onemount.FileStatus"
	// DBusObjectPath is the D-Bus object path for onemount
	DBusObjectPath = "/org/onemount/FileStatus"
	// DBusServiceNameBase is the base D-Bus service name for onemount
	DBusServiceNameBase = "org.onemount.FileStatus"
)

// DBusServiceName returns the D-Bus service name, which may be unique in test environments
var DBusServiceName string

// SetDBusServiceNamePrefix sets the DBusServiceName with the given prefix
// This allows tests to set a custom prefix without relying on environment variables
func SetDBusServiceNamePrefix(prefix string) {
	// Always generate a unique name to avoid conflicts in tests and parallel mounts
	// Generate a unique suffix based on process ID and a random number
	uniqueSuffix := fmt.Sprintf("%d_%d", os.Getpid(), time.Now().UnixNano()%10000)

	// Use the provided prefix or default to "instance"
	if prefix == "" {
		prefix = "instance"
	}

	DBusServiceName = fmt.Sprintf("%s.%s_%s", DBusServiceNameBase, prefix, uniqueSuffix)
	log.Debug().Str("dbusName", DBusServiceName).Msg("Using unique D-Bus service name")
}

func init() {
	// Initialize the DBusServiceName variable with the default prefix
	SetDBusServiceNamePrefix("instance")
}

// FileStatusDBusServer implements a D-Bus server for file status updates
type FileStatusDBusServer struct {
	fs       *Filesystem
	conn     *dbus.Conn
	mutex    sync.RWMutex
	started  bool
	stopChan chan struct{}
}

// NewFileStatusDBusServer creates a new D-Bus server for file status updates
func NewFileStatusDBusServer(fs *Filesystem) *FileStatusDBusServer {
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
		log.Error().Err(err).Msg("Failed to connect to D-Bus session bus")
		return err
	}
	s.conn = conn

	// Export the FileStatusDBusServer object
	err = conn.Export(s, DBusObjectPath, DBusInterface)
	if err != nil {
		log.Error().Err(err).Msg("Failed to export D-Bus object")
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
		log.Error().Err(err).Msg("Failed to export introspection data")
		s.conn = nil
		return err
	}

	s.started = true
	log.Info().Msg("D-Bus server started in test mode")
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
		log.Error().Err(err).Msg("Failed to connect to D-Bus session bus")
		return err
	}
	s.conn = conn

	// Request a name on the bus with flags to allow replacement and not queue
	// This ensures we can always get a name, even if there are conflicts
	reply, err := conn.RequestName(DBusServiceName, dbus.NameFlagAllowReplacement|dbus.NameFlagReplaceExisting|dbus.NameFlagDoNotQueue)
	if err != nil {
		log.Error().Err(err).Msg("Failed to request D-Bus name")
		s.conn = nil
		return err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		// Since we're using a unique name and NameFlagReplaceExisting, this should rarely happen
		// But if it does, we'll log it and continue
		log.Warn().Msgf("Not primary owner of D-Bus name (reply: %v), but continuing with unique name: %s", reply, DBusServiceName)
	} else {
		log.Debug().Str("dbusName", DBusServiceName).Msg("Successfully acquired D-Bus name")
	}

	// Export the FileStatusDBusServer object
	err = conn.Export(s, DBusObjectPath, DBusInterface)
	if err != nil {
		log.Error().Err(err).Msg("Failed to export D-Bus object")
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
		log.Error().Err(err).Msg("Failed to export introspection data")
		s.conn = nil
		return err
	}

	s.started = true
	log.Info().Msg("D-Bus server started")
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
		log.Debug().Str("dbusName", DBusServiceName).Msg("Releasing D-Bus name")
		if _, err := s.conn.ReleaseName(DBusServiceName); err != nil {
			log.Warn().Err(err).Msg("Failed to release D-Bus name")
		}

		// Unexport the objects to clean up resources
		if err := s.conn.Export(nil, DBusObjectPath, DBusInterface); err != nil {
			log.Warn().Err(err).Msg("Failed to unexport D-Bus object")
		}
		if err := s.conn.Export(nil, DBusObjectPath, "org.freedesktop.DBus.Introspectable"); err != nil {
			log.Warn().Err(err).Msg("Failed to unexport introspection data")
		}

		// Close the connection
		if err := s.conn.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close D-Bus connection")
		}
		s.conn = nil
	}
	s.started = false
	log.Info().Msg("D-Bus server stopped and resources cleaned up")
}

// GetFileStatus returns the status of a file
func (s *FileStatusDBusServer) GetFileStatus(path string) (string, *dbus.Error) {
	inode, err := s.fs.GetPath(path, nil)
	if err != nil || inode == nil {
		log.Error().Err(err).Str("path", path).Msg("Failed to get inode for path")
		return "Unknown", nil
	}

	status := s.fs.GetFileStatus(inode.ID())
	return status.Status.String(), nil
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
		log.Error().Err(err).Str("path", path).Str("status", status).Msg("Failed to emit D-Bus signal")
	}
}

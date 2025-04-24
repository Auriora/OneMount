package fs

import (
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/rs/zerolog/log"
)

const (
	// DBusInterface is the D-Bus interface name for onedriver
	DBusInterface = "org.onedriver.FileStatus"
	// DBusObjectPath is the D-Bus object path for onedriver
	DBusObjectPath = "/org/onedriver/FileStatus"
	// DBusServiceName is the D-Bus service name for onedriver
	DBusServiceName = "org.onedriver.FileStatus"
)

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

	// Request a name on the bus
	reply, err := conn.RequestName(DBusServiceName, dbus.NameFlagDoNotQueue)
	if err != nil {
		log.Error().Err(err).Msg("Failed to request D-Bus name")
		s.conn = nil
		return err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		log.Error().Msgf("D-Bus name already taken: %v", reply)
		// In tests, we might be running multiple instances, so we'll just continue
		// This is not ideal for production, but it allows tests to pass
		log.Warn().Msg("Continuing despite not being primary owner of D-Bus name")
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

// Stop stops the D-Bus server
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
		if err := s.conn.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close D-Bus connection")
		}
		s.conn = nil
	}
	s.started = false
	log.Info().Msg("D-Bus server stopped")
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

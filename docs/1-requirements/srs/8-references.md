# 8. References

This section provides references to the libraries, library documentation, Microsoft Graph API, and other external resources used in the OneMount project.

## 8.1 Go Libraries

| Library | Description | Documentation URL |
|---------|-------------|------------------|
| go-fuse/v2 | FUSE library for implementing filesystems | https://github.com/hanwen/go-fuse |
| gotk3 | GTK3 bindings for Go, used for GUI components | https://github.com/gotk3/gotk3 |
| bbolt | Embedded key/value database for Go, used for caching | https://pkg.go.dev/go.etcd.io/bbolt |
| zerolog | Structured logging library for Go | https://github.com/rs/zerolog |
| testify | Testing toolkit for Go | https://github.com/stretchr/testify |
| go-systemd | Go bindings for systemd | https://github.com/coreos/go-systemd |
| dbus | Go bindings for D-Bus | https://github.com/godbus/dbus |
| mergo | Helper for merging structs and maps in Go | https://github.com/imdario/mergo |
| pflag | Drop-in replacement for Go's flag package | https://github.com/spf13/pflag |
| socketio-go | Socket.IO implementation in Go | https://github.com/yousong/socketio-go |
| yaml | YAML support for Go | https://github.com/go-yaml/yaml |

## 8.2 Microsoft Graph API

The OneMount project uses the Microsoft Graph API to interact with OneDrive. The following resources are relevant:

| Resource | Description | URL |
|----------|-------------|-----|
| Graph API Overview | General overview of the Microsoft Graph API | https://docs.microsoft.com/en-us/graph/overview |
| OneDrive API | Documentation for the OneDrive API in Microsoft Graph | https://docs.microsoft.com/en-us/graph/api/resources/onedrive |
| Drive Resource | Documentation for the Drive resource | https://docs.microsoft.com/en-us/onedrive/developer/rest-api/resources/drive |
| DriveItem Resource | Documentation for the DriveItem resource | https://docs.microsoft.com/en-us/onedrive/developer/rest-api/resources/driveitem |
| User Resource | Documentation for the User resource | https://docs.microsoft.com/en-ca/graph/api/user-get |
| Quota Resource | Documentation for the Quota resource | https://docs.microsoft.com/en-us/onedrive/developer/rest-api/resources/quota |
| Authentication | OAuth 2.0 authentication for Microsoft Graph | https://docs.microsoft.com/en-us/graph/auth/auth-concepts |

## 8.3 FUSE (Filesystem in Userspace)

| Resource | Description | URL |
|----------|-------------|-----|
| FUSE Overview | General overview of FUSE | https://www.kernel.org/doc/html/latest/filesystems/fuse.html |
| libfuse | C library for implementing FUSE filesystems | https://github.com/libfuse/libfuse |

## 8.4 GTK3

| Resource | Description | URL |
|----------|-------------|-----|
| GTK3 Documentation | Official documentation for GTK3 | https://docs.gtk.org/gtk3/ |
| GTK3 API Reference | API reference for GTK3 | https://docs.gtk.org/gtk3/index.html |

## 8.5 Systemd

| Resource | Description | URL |
|----------|-------------|-----|
| Systemd Documentation | Official documentation for systemd | https://www.freedesktop.org/software/systemd/man/ |
| D-Bus Specification | D-Bus specification used for systemd integration | https://dbus.freedesktop.org/doc/dbus-specification.html |
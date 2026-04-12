# OneMount Code Architecture Review

OneMount is structured as a robust, concurrent Go application utilizing FUSE to map Microsoft OneDrive into a native Linux filesystem.

## High-Level Architecture
1. **Application Entry Point (`cmd/onemount/main.go`)**
   - Handles CLI flag parsing, configuration loading (`common.Config`), and environment checks.
   - Sets up logging, caching directory structure, and authentication credentials.
   - Instantiates the FUSE server and mounts the `internal/fs.Filesystem` instance at the provided mountpoint.
   - Manages background tasks upon launch, such as early delta synchronization and content cache cleanup.

2. **Filesystem Core (`internal/fs/`)**
   - **`fs.go` & `raw_filesystem.go`:** Houses the core `Filesystem` structured interface that routes FUSE operations.
   - **Entity Operations:** `file_operations.go`, `dir_operations.go`, `xattr_operations.go` are responsible for translating standard POSIX filesystem calls (Read, Write, Open, Mkdir) into internal caching logic and OneDrive API commands.
   - **Inodes:** `inode.go` handles the metadata abstractions matching OneDrive items to FUSE inode numbers.
   - **Data Transfer Management:** `download_manager.go` and `upload_manager.go` represent concurrent worker pools designed to smoothly hydrate files and dispatch offline/local modifications. 
   - **IPC & UI:** `dbus.go` implements a D-Bus interface, enabling file managers (like Nemo, via the `internal/nemo` package) to show file statuses (e.g., syncing, downloading, cloud-only) via badges inside the native UI.

3. **Cloud Integrations (`internal/graph/`)**
   - **API Client:** Wraps the Microsoft Graph API. Functions in `graph.go` and `drive_item.go` interface directly with OneDrive for fetching drive trees, executing delta queries, and managing metadata.
   - **Authentication:** `oauth2*.go` handles browser-based GTK token generation, headless setups, and secure token caching on disk based on the system instance path. 
   - **Realtime Push:** The `socketio/` library alongside `socket_subscription.go` uses WebSockets and the Microsoft Graph subscriptions API for push notifications of remotely modified files, dropping the dependence on continuous polling.

4. **Cache & Offline Storage**
   - **bbolt Database ($XDG_CACHE_HOME/onemount/.db):** Uses a local optimized key/value store mapping OneDrive paths to ETag metadata (`metadata_store.go`), allowing instant reads of folder contents without initial API checks.
   - **Sparse Files & Conflict Resolution:** Implements granular conflict resolution and delta tree caching (`delta.go`), enabling robust local write operations even when disconnected (`offline.go`, `conflict_resolution.go`).
   
## Noteworthy Strengths
- **Concurrency:** Extensive use of worker queues, timeouts, and `sync.RWMutex`, which protects against deadlock across parallel I/O loads.
- **Testing Approach:** Contains extensive integration tests (`comprehensive_integration_test.go`, `delta_sync_integration_test.go`) utilizing an advanced mock graph (`internal/graph/mock`). Performance characteristics are tested aggressively alongside properties of concurrent data structure behavior.

## Further Exploration Points
- Let me know if you would like me to deep dive into how **Offline Conflict Resolution** operates and how it recovers data.
- Would you like a deeper analysis of the **Socket.IO Real-time Synchronization** mechanisms bounding bandwidth usage?
- Perhaps a detailed breakdown of the **FUSE caching strategy (Eviction & ETag caching)**?

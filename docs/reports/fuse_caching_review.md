# OneMount FUSE Caching & ETag Strategy Review

OneMount goes to great lengths to provide snappy, near-native filesystem performance without monopolizing network bandwidth or local disk space. It achieves this via a multi-layered caching system and strict ETag verification.

## 1. Directory & Metadata Caching
OneMount utilizes a local optimized DB (`bbolt`) stored in `~/.cache/onemount/.db`.
- This database caches folder structures, file metadata (names, sizes, last modified times), and remote ETags.
- Because FUSE constantly queries file attributes (e.g., when you run `ls` or a file manager generates thumbnails), querying the network for every `getattr` would render the filesystem unusable. The local metadata store serves these FUSE calls instantly.

## 2. File Content Caching (Loopback Cache)
When an application opens a file on the OneMount FUSE drive, the filesystem downloads the file, placing the raw bytes into a local "loopback" directory (`~/.cache/onemount/content`). Subsequent reads of that file pull directly from this local disk cache. This enables functions like video scrubbing or compiling code to run at NVMe/SSD speeds rather than network speeds.

### Cache Eviction (LRU & Time-based)
`LoopbackCache` (`internal/fs/content_cache.go`) manages disk footprint so it doesn't grow indefinitely:
- **Maximum Size Enforcement:** When `maxCacheSize` is defined, the system performs an LRU (Least Recently Used) eviction. Every time a file is accessed via the mount, its cache entry's `lastAccessed` timestamp is updated. If a new download exceeds the maximum configured size, the oldest un-open files are purged until enough space is reclaimed.
- **Time-based Cleanup:** A background `CleanupCache` routine runs periodically, sweeping the content directory. By default, any file that hasn't been accessed in `X` days (configurable via `--cache-expiration`) is deleted from disk. Open and locked files are protected by an `evictionGuard`.

## 3. ETag-based Cache Invalidation
Whenever data is fetched from the Graph API or a realtime Socket.IO push notification is received, the filesystem checks the remote item's `ETag` (an opaque string representing the current state/version of the remote file).
- If the remote ETag has advanced past the local ETag stored in the `bbolt` metadata database, the current FUSE data is deemed stale.
- OneMount immediately evicts the corresponding local file from the `LoopbackCache` (`fs.content.Delete(id)`).
- The next time the user (or system) attempts to read that file, a fresh copy is downloaded transparently from OneDrive, ensuring data consistency without the user ever noticing the cache rotation.

# Addressing FUSE Memory Mapping & Large File Support

Based on a review of `internal/fs/file_operations.go` and `internal/fs/download_manager.go`, OneMount currently handles file access via a "download-first" approach. When an application attempts to `Read()` a file on the FUSE mount, the `Open()` syscall blocks until `DownloadManager` pulls the entire file from the OneDrive API into a local `LoopbackCache` (`~/.cache/onemount/content/{id}`). FUSE then fulfills the read by passing a file descriptor to this local cache back to the kernel.

## The Problem with Large Files
1. **Blocking Opens:** Standard file browsers and applications expecting instantaneous open syscalls will hang or timeout (FUSE `openDownloadTimeout` is currently hardcoded to 60 seconds) if they attempt to open a massive file (e.g., a 10GB video) that takes longer than 60 seconds to download.
2. **Local Storage Thrashing:** Opening a 10GB file requires 10GB of free space on the local cache drive. FUSE memory mapping loads accessed portions into memory, which makes local caching fast but causes high I/O and memory pressure for sequential reads of massive files that will never be cached efficiently anyway.
3. **Bandwidth Waste:** If an application only needs to read the header of a large video file to get its duration, the current architecture still downloads the entire multi-gigabyte file before fulfilling that 1KB read request.

## Recommendations for Improvement

### 1. Implement Partial/Range Downloads
Currently, Microsoft Graph API supports the standard `Range: bytes={start}-{end}` HTTP header. OneMount should be updated to recognize partial read requirements.
- **How:** Modify `internal/fs/file_operations.go:Read()` to detect if the local cache is empty but the requested `Offset` and `Size` are a small subset of the file. 
- Instead of queueing a full file download via `DownloadManager`, `Read()` could execute a synchronous Range request to the Microsoft Graph API, returning just the requested bytes directly to FUSE without persisting them to disk.
- **Benefit:** Applications probing file headers or streaming video metadata would respond instantly without triggering massive background downloads.

### 2. Stream-Through File Descriptors
For sequential reads of large files, wait-for-full-download is highly inefficient.
- **How:** Introduce a streaming cache mode. When `Open()` is called on a file larger than a specific threshold (e.g., >500MB), OneMount should return a successful handle immediately and spin up a background download thread.
- As the application calls `Read()`, if the requested offset has already been downloaded, serve it from the disk cache. If the requested offset hasn't been downloaded yet, block *just that Read call* until the downloader thread reaches that offset.
- **Benefit:** Users can begin watching standard media files or parsing large datasets immediately, rather than waiting minutes for the download to complete before the file even "opens".

### 3. FUSE `direct_io` Flag Optimization
When files are opened, the kernel tries to cache data aggressively.
- **How:** For files exceeding a certain size limit, OneMount should set the `FOPEN_DIRECT_IO` flag in its `OpenOut` response in `file_operations.go`. 
- **Benefit:** This tells the Linux kernel *not* to cache the file contents in the page cache, bypassing the memory-mapping bottleneck entirely. The FUSE daemon will receive read requests directly and can serve them straight from the local loopback cache (or HTTP stream) without trashing system memory.

### 4. Background Prefetching Hints
Many video editors and large-file processors read files in chunks.
- **How:** Track sequential read patterns in the FUSE daemon. Add logic to the `DownloadManager` to aggressively pre-fetch the *next* block of a large file before the application explicitly asks for it, caching it locally just in time.
- **Benefit:** Hides network latency behind local application processing times.

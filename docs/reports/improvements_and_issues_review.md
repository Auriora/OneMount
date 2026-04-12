# OneMount Notable Improvements & Issues

Based on a review of the codebase and the project's internal tracking documents, OneMount is a highly capable and well-documented project. However, there are several notable technical debts, missing features, and architectural bottlenecks that should be considered for future improvement.

## 🔴 Critical Issues (Blocking v1.0 Release)
The project's own internal audit from today flags the following as critical limiters for a stable 1.0 release:
1. **Test Naming Convention Violations:** ~30% of tests (226 out of 752) lack proper categorization labels, making it difficult to run specific test types (e.g., unit vs. integration) in isolation.
2. **QuickXORHash Testing Missing:** There is a critical lack of comprehensive testing for the `QuickXORHash` implementation against Microsoft's official test vectors. Because OneDrive file integrity and conflict resolution depend on these hashes matching exactly, this is a significant risk.

## 🟡 Moderate Technical Debt & Architecture Improvements
These issues are currently tracked in the codebase as `TODO`s targeted for a `v1.1` release or later:

1. **`main.go` Monolith:**
   - The application entry point (`cmd/onemount/main.go`) has grown to over 1,000 lines. 
   - It handles CLI parsing, daemonization, filesystem initialization, configuration management, and statistics gathering all in one file. 
   - **Improvement:** It should be refactored into discrete service packages (e.g., `cmd/onemount/cli/`, `cmd/onemount/service/`).
2. **50+ Unimplemented Test Cases:**
   - Across the project (especially in `internal/fs` and UI integration), there are over 50 placeholder test functions with `TODO` comments.
3. **Security - Plaintext Token Storage:**
   - OAuth2 tokens are cached on disk to maintain the mount across reboots. Currently, these tokens are stored in plaintext. 
   - **Improvement:** Implement encrypted token storage leveraging the OS native keychain (e.g., Secret Service API via D-Bus on Linux).
4. **Performance - Statistics Collection on Large Drives:**
   - The `--stats` command performs a full filesystem traversal. For users with >100k files, this is highly inefficient.
   - **Improvement:** Implement incremental tracking or background aggregation for FUSE statistics.
5. **Advanced Error Monitoring:**
   - While the app currently logs errors well via `zerolog`, there is no aggregation or pattern detection to help users understand *why* they might be experiencing degraded performance (e.g., persistent Microsoft Graph API rate limits).

## 🔵 Known Functional Limitations
As outlined in the `README`:
- **Large Files:** FUSE memory mapping loads accessed files into memory. This makes small files fast but chokes on multi-gigabyte files (e.g., 4K video editing directly off the mount).
- **Symlinks:** Microsoft Graph API does not support symbolic links. Attempts to create them on the mount return `ENOSYS`.
- **Thumbnail Crawlers:** File browsers like GNOME Nautilus will attempt to download every file in a directory to generate thumbnails. Standard behavior, but it will initially trigger massive network spikes on first access.

# Cache Deletion Policy Follow-Up

**Date**: 2025-11-14  
**Type**: Investigation Note  
**Component**: Filesystem Cache / Sync Pipeline  
**Status**: Pending Investigation

## Summary

- Earlier guidance required immediately removing cached file and folder metadata whenever users delete items.
- Product direction has shifted: cache eviction for deletions should occur **only after** the deletion is synchronized successfully to OneDrive, preventing accidental local data loss when a delete is later rolled back or fails to upload.
- We need to audit the current delete path (`Filesystem.DeleteID`, sync manager, conflict resolution) to ensure it retains cache entries until the remote confirmation arrives.

## Next Steps / Questions

1. Confirm which components currently trigger cache removal (local delete, delta processing, conflict resolution) and map the exact timing relative to OneDrive confirmation.
2. Determine whether “sync complete” means Graph delete delta observed, successful REST response, or background retry exhaustion.
3. Design tests covering offline deletes, retries, and rollback scenarios to validate the revised requirement.

Document created to track the requirement change; implementation is out of scope for this task.

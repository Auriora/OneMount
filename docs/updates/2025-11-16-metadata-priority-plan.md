# Metadata Request Prioritization Plan

**Date**: 2025-11-16  
**Type**: Design exploration  
**Component**: Filesystem / Metadata Fetcher  
**Status**: Proposed

## Summary

Interactive commands (`ls`, `cd`, `cat`, etc.) still block whenever a directory’s metadata has to be fetched, especially while `syncTree` and the delta loop run in the background. Foreground `GetChildrenID` requests compete with the tree walker inside `MetadataRequestManager`, and today’s cache invalidations (typos, `.xdg-volume-info` refresh) make the root directory refetch constantly. This document captures the design goals and targeted fixes so we can implement them incrementally.

## Findings

- The metadata manager still launches only three workers (`internal/fs/cache.go:267`). Background (`PriorityBackground`) and foreground (`PriorityForeground`) jobs share those workers. When all three workers are busy with background jobs, new foreground requests sit in the queue until a worker finishes—leading to multi-minute hangs after a cache invalidation.
- Foreground `GetChildrenID` (`internal/fs/cache.go:1070-1278`) wipes the cached child list whenever a lookup fails (`GetChild` calls `parent.ClearChildren()` if a name isn’t found). A simple typo like `cd Pictires` forces the root to be re-downloaded immediately afterward.
- The `.xdg-volume-info` helper (`cmd/common/common.go:90-137`) also clears the root children when it recreates the virtual file at startup, creating an unavoidable second refetch.
- The background tree sync (`internal/fs/sync.go:162-220`) never populates `inode.children`, so it doesn’t actually warm the directories it walks. Every directory still hits Graph the first time the user opens it.
- The delta loop was running every second (`deltaInterval: 1`), constantly scheduling metadata work and keeping the queues saturated even when nothing changed.

## Immediate Mitigation

- **Increase delta interval (temporary deviation)**: Update `~/.config/onemount/config.yml` to set `deltaInterval: 10`. This reduces the steady-state load on the metadata queues so foreground calls are less likely to starve while we implement deeper fixes.  
  _Note_: Requirements 5.5/5.7/17.5 (delta loop should back off to ~5 minutes when no webhook is available) are temporarily violated. This change is strictly diagnostic; we must either restore the longer interval after the queue fixes or bring webhook subscriptions online so we meet the requirement without saturating metadata workers.

## Proposed Direction

1. **Deduplicate in-flight metadata fetches.** Track active requests inside `MetadataRequestManager` keyed by directory ID/path. Foreground callers attach callbacks to the existing job instead of issuing duplicate Graph requests.
2. **Adjust worker accounting instead of requeueing/dropping.** When a worker pulls a low-priority job but a high-priority request arrives, pause the background job locally, service the high-priority queue immediately, then resume the paused job. Also consider dedicating at least one worker to high-priority work or making the worker count configurable.
3. **Serve cached children while refresh runs.** Once a directory has been fetched at least once, keep its child slice intact while a refresh happens in the background. Failed lookups (typos, missing files) should no longer call `parent.ClearChildren()`; they can trigger an async refresh without blocking the current command.
4. **Have background sync populate the parent’s child slice.** When `SyncDirectoryTreeWithContext` fetches a directory, write the results into the parent inode so the cache is warm before the user visits it.
5. **Avoid clearing root for `.xdg-volume-info`.** Update the helper to refresh just that entry rather than calling `root.ClearChildren()`.

## Next Steps

1. **Config change (done)**: `deltaInterval` set to 10 seconds in `~/.config/onemount/config.yml` to lower baseline load.
2. **Short-term hotfix**: Patch `GetChild` so failed lookups don’t clear the entire child list, and update `.xdg-volume-info` creation to leave other entries untouched.
3. **Design & implement in-flight dedup + improved worker scheduling**: define the shared request map, add unit tests in `internal/fs/metadata_priority_test.go`, and validate with the stress harness.
4. **Teach tree sync to warm `inode.children`** so the background walk actually caches directories.
5. **Measure** cold `ls` latency before/after each change to ensure the queue stalls disappear.

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)  
- AGENT-GUIDE-Operational-Best-Practices (priority 40)  
- AGENT-RULE-Documentation-Conventions (priority 20)

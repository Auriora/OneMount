# TODO Comments Summary

This document summarizes all the TODO comments that have been added to mark incomplete features with clear implementation guidance.

## Overview

As part of the release action plan, incomplete features have been marked with comprehensive TODO comments that include:
- Specific implementation details
- Target release version
- Priority level
- Dependencies and considerations
- Reference to related GitHub issues where applicable

## Added TODO Comments

### 1. Documentation TODOs

#### README.md (lines 138-141)
- **Feature**: Ubuntu/Debian installation instructions
- **Target**: v1.1 release
- **Priority**: Medium
- **Details**: Need to add PPA setup or direct package download instructions

#### docs/installation-guide.md (lines 67-71)
- **Feature**: Fix PPA removal instructions
- **Target**: v1.1 release
- **Priority**: Low
- **Details**: Current PPA removal command may be invalid, need verification

### 2. Development Tooling TODOs

#### scripts/implement_github_issue.py (lines 59-68)
- **Feature**: JetBrains task opening functionality
- **Target**: Deferred to post-v1.0
- **Priority**: Low (development tooling, not core functionality)
- **Details**: Need proper IDE integration for task management

### 3. Test Implementation TODOs

#### pkg/graph/path_test.go
- **TestUT_GR_26_01_IDPath_VariousItemIDs_FormatsCorrectly** (lines 24-34)
  - Target: v1.1 release
  - Priority: Medium (testing infrastructure)
  - Details: Test IDPath function with various OneDrive item IDs

- **TestUT_GR_27_01_ChildrenPath_VariousPaths_FormatsCorrectly** (lines 54-65)
  - Target: v1.1 release
  - Priority: Medium (testing infrastructure)
  - Details: Test childrenPath function with various filesystem paths

- **TestUT_GR_28_01_ChildrenPathID_VariousItemIDs_FormatsCorrectly** (lines 85-95)
  - Target: v1.1 release
  - Priority: Medium (testing infrastructure)
  - Details: Test childrenPathID function with various item IDs

#### pkg/graph/hash_functions_test.go
- **TestUT_GR_11_01_SHA256Hash_VariousInputs_ReturnsCorrectHash** (lines 25-35)
  - Target: v1.1 release
  - Priority: High (cryptographic functions need thorough testing)
  - Details: Test SHA256Hash with various inputs including edge cases

- **TestUT_GR_12_01_SHA256HashStream_VariousInputs_ReturnsCorrectHash** (lines 56-66)
  - Target: v1.1 release
  - Priority: High (used for file integrity verification during uploads)
  - Details: Test SHA256HashStream with various reader types

- **TestUT_GR_15_01_QuickXORHash_VariousInputs_ReturnsCorrectHash** (lines 137-148)
  - Target: v1.1 release
  - Priority: Critical (OneDrive file integrity depends on this)
  - Details: Test QuickXORHash with Microsoft's test vectors

### 4. Architecture Refactoring TODOs

#### cmd/onemount/main.go (lines 3-23)
- **Feature**: Refactor main.go into discrete services (Issue #54)
- **Target**: v1.1 release
- **Priority**: Medium (architectural improvement)
- **Details**: Break down large main.go into separate service modules

### 5. Performance Optimization TODOs

#### internal/fs/stats.go (lines 79-91)
- **Feature**: Optimize statistics collection for large filesystems (Issues #11, #10, #9, #8, #7)
- **Target**: v1.1 release
- **Priority**: Medium (acceptable performance for typical use cases)
- **Details**: Implement incremental updates, caching, and background processing

### 6. Advanced Feature TODOs

#### pkg/errors/error_monitoring.go (lines 108-120)
- **Feature**: Enhance error monitoring with advanced features (Issues #75, #74, #73, #72, #71, #65)
- **Target**: v1.2 release
- **Priority**: Low (basic monitoring is sufficient for initial release)
- **Details**: Add error aggregation, pattern detection, external monitoring integration

## TODO Comment Standards

All TODO comments follow this format:
```
// TODO: [Brief description] ([Issue references if applicable])
// [Detailed explanation of what needs to be implemented]
// [Specific implementation suggestions or options]
// Target: [Release version]
// Priority: [High/Medium/Low] ([Justification])
// [Additional context or dependencies]
```

## Next Steps

1. **For v1.0 Release**: All critical TODOs have been deferred to post-release
2. **For v1.1 Release**: Focus on test coverage, documentation, and performance optimizations
3. **For v1.2 Release**: Implement advanced monitoring and analytics features

## Maintenance

- Review TODO comments quarterly to reassess priorities
- Update target versions based on actual development progress
- Remove TODO comments when features are implemented
- Add new TODO comments for newly identified incomplete features

This document should be updated whenever new TODO comments are added or existing ones are resolved.

## Recent Updates

### June 2025
- **Fixed**: Database persistence hanging issue in `internal/fs/upload_signal_basic_test.go:162`
  - Issue was resolved and test now passes with proper persistence verification
  - TODO comment removed and test enhanced with data verification

# TODO Comments Summary

This document summarizes all the TODO comments that have been added to mark incomplete features with clear implementation guidance.

## Overview

As part of the release action plan, incomplete features have been marked with comprehensive TODO comments that include:
- Specific implementation details
- Target release version
- Priority level
- Dependencies and considerations
- Reference to related GitHub issues where applicable

## Current TODO Comments (Updated: January 2025)

### 1. Documentation TODOs

#### README.md (line 147)
- **Feature**: Ubuntu/Debian installation instructions
- **Target**: v1.1 release
- **Priority**: Medium
- **Details**: Need to add PPA setup or direct package download instructions

### 2. Architecture Refactoring TODOs

#### cmd/onemount/main.go (lines 3-23)
- **Feature**: Refactor main.go into discrete services (Issue #54)
- **Target**: v1.1 release
- **Priority**: Medium (architectural improvement)
- **Details**: Break down large main.go (~677 lines) into separate service modules:
  - Extract CLI handling into cmd/onemount/cli/
  - Extract filesystem service into cmd/onemount/service/
  - Extract statistics service into cmd/onemount/stats/
  - Extract daemon handling into cmd/onemount/daemon/
  - Keep main.go as a thin coordinator

### 3. Performance Optimization TODOs

#### internal/fs/stats.go (lines 79-91)
- **Feature**: Optimize statistics collection for large filesystems (Issues #11, #10, #9, #8, #7)
- **Target**: v1.1 release
- **Priority**: Medium (acceptable performance for typical use cases)
- **Details**: Current implementation performs full traversal which can be slow for large filesystems (>100k files):
  - Implement incremental statistics updates instead of full recalculation
  - Cache frequently accessed statistics with TTL
  - Use background goroutines for expensive calculations
  - Implement sampling for very large datasets
  - Add pagination support for statistics display
  - Optimize database queries with better indexing
  - Consider using separate statistics database/table

### 4. Advanced Feature TODOs

#### pkg/errors/error_monitoring.go (lines 108-120)
- **Feature**: Enhance error monitoring with advanced features (Issues #75, #74, #73, #72, #71, #65)
- **Target**: v1.2 release
- **Priority**: Low (basic monitoring is sufficient for initial release)
- **Details**: Current implementation provides basic error rate monitoring. Advanced features to implement:
  - Error aggregation across multiple time windows (hourly, daily, weekly)
  - Error pattern detection and alerting
  - Integration with external monitoring systems (Prometheus, Grafana)
  - Error correlation analysis (e.g., network errors leading to auth errors)
  - Automatic error recovery suggestions
  - Error trend analysis and prediction
  - Centralized error monitoring dashboard
  - Error severity classification and escalation

### 5. Test Implementation TODOs

#### Hash Functions Testing (pkg/graph/hash_functions_test.go)
- **TestUT_GR_15_01_QuickXORHash_VariousInputs_ReturnsCorrectHash** (lines 185-196)
  - Target: v1.1 release
  - Priority: Critical (OneDrive file integrity depends on this)
  - Details: Test QuickXORHash with Microsoft's test vectors, empty arrays, small/large content
  - Reference: https://docs.microsoft.com/en-us/onedrive/developer/code-snippets/quickxorhash

- **Other hash function tests** (lines 135, 160, 217)
  - Target: v1.1 release
  - Priority: High (cryptographic functions need thorough testing)
  - Details: Test SHA256Hash and SHA256HashStream with various inputs including edge cases

#### Comprehensive Test Coverage (Multiple Files)
The following test files contain numerous TODO comments for unimplemented test cases:

**Filesystem Tests:**
- `internal/fs/dbus_test.go` (2 TODOs)
- `internal/fs/delta_test.go` (7 TODOs)
- `internal/fs/fs_integration_test.go` (13 TODOs)
- `internal/fs/inode_test.go` (4 TODOs)
- `internal/fs/thumbnail_test.go` (3 TODOs)
- `internal/fs/upload_manager_test.go` (3 TODOs)
- `internal/fs/upload_session_test.go` (1 TODO)
- `internal/fs/xattr_operations_test.go` (3 TODOs)

**UI Tests:**
- `internal/ui/onemount_test.go` (3 TODOs)
- `internal/ui/systemd/systemd_test.go` (2 TODOs)

**Graph/API Tests:**
- `pkg/graph/oauth2_gtk_test.go` (1 TODO)
- `pkg/graph/offline_test.go` (3 TODOs)

**Common Tests:**
- `cmd/common/common_test.go` (1 TODO)
- `cmd/common/config_test.go` (4 TODOs)

**Total Test TODOs**: 50+ unimplemented test cases
- **Target**: v1.1 release (test coverage improvement)
- **Priority**: Medium to High (depending on component criticality)
- **Details**: Most are placeholder test cases that need implementation to improve test coverage

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

## Summary Statistics

- **Total TODO Comments**: 55+ across the codebase
- **Documentation TODOs**: 1 (README.md installation instructions)
- **Architecture TODOs**: 1 (main.go refactoring)
- **Performance TODOs**: 1 (statistics optimization)
- **Advanced Feature TODOs**: 1 (error monitoring enhancement)
- **Test Implementation TODOs**: 50+ (comprehensive test coverage)

## Priority Breakdown

- **Critical**: 1 (QuickXORHash testing - OneDrive file integrity)
- **High**: 4+ (cryptographic function testing)
- **Medium**: 45+ (general test coverage, architecture, performance)
- **Low**: 1 (advanced error monitoring)

## Target Release Distribution

- **v1.1 Release**: 53+ TODOs (test coverage, documentation, performance, architecture)
- **v1.2 Release**: 1 TODO (advanced error monitoring features)

## Next Steps

1. **For v1.1 Release**:
   - Priority focus on critical QuickXORHash testing
   - Implement comprehensive test coverage for filesystem components
   - Add Ubuntu/Debian installation documentation
   - Consider main.go refactoring and statistics optimization

2. **For v1.2 Release**:
   - Implement advanced error monitoring and analytics features

## Maintenance

- Review TODO comments quarterly to reassess priorities
- Update target versions based on actual development progress
- Remove TODO comments when features are implemented
- Add new TODO comments for newly identified incomplete features

This document should be updated whenever new TODO comments are added or existing ones are resolved.

## Recent Updates

### January 2025
- **Updated**: Complete audit of all TODO comments in codebase
- **Identified**: 55+ TODO comments across multiple categories
- **Prioritized**: Critical QuickXORHash testing for OneDrive file integrity
- **Categorized**: Test implementation TODOs represent majority of outstanding work

### Previous Updates (June 2024)
- **Fixed**: Database persistence hanging issue in `internal/fs/upload_signal_basic_test.go:162`
  - Issue was resolved and test now passes with proper persistence verification
  - TODO comment removed and test enhanced with data verification
- **Implemented**: Path function tests in `pkg/graph/path_test.go` (Issue #117)
  - `TestUT_GR_26_01_IDPath_VariousItemIDs_FormatsCorrectly` - ✅ COMPLETE
  - `TestUT_GR_27_01_ChildrenPath_VariousPaths_FormatsCorrectly` - ✅ COMPLETE
  - `TestUT_GR_28_01_ChildrenPathID_VariousItemIDs_FormatsCorrectly` - ✅ COMPLETE
  - All tests include comprehensive coverage of edge cases, special characters, and URL encoding

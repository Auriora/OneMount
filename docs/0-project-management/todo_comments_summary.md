# TODO Comments Summary

This document summarizes all the TODO comments remaining in the codebase.

## Overview

As part of the release action plan, incomplete features have been marked with comprehensive TODO comments that include:
- Specific implementation details
- Target release version
- Priority level
- Dependencies and considerations
- Reference to related GitHub issues where applicable

## Current TODO Comments (Updated: April 2026)

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

### 5. Security Enhancement TODOs

#### internal/graph/security_property_test.go (line 144)
- **Feature**: Encrypt auth tokens at rest using AES-256-GCM with OS keyring
- **Target**: v1.2 release
- **Priority**: Medium
- **Details**: Future implementation note for token storage hardening

### ~~6. Test Implementation TODOs~~ — ✅ RESOLVED (April 2026)

All 46 previously-unimplemented test cases have been implemented and verified passing in Docker. See `docs/reports/2026-04-12-test-coverage-backlog-analysis.md` for the pre-implementation audit and `docs/updates/2026-04-12-191500-test-coverage-backlog-implementation.md` for the implementation summary.

## Summary Statistics

- **Total remaining TODO Comments**: ~5 across the codebase
- **Documentation TODOs**: 1 (README.md installation instructions)
- **Architecture TODOs**: 1 (main.go refactoring)
- **Performance TODOs**: 1 (statistics optimization)
- **Advanced Feature TODOs**: 1 (error monitoring enhancement)
- **Security TODOs**: 1 (token encryption at rest)
- **Test Implementation TODOs**: 0 (all 46 implemented April 2026)

## Priority Breakdown

- **Critical**: 0
- **High**: 0
- **Medium**: 4 (documentation, architecture, performance, security)
- **Low**: 1 (advanced error monitoring)

## Target Release Distribution

- **v1.1 Release**: 3 TODOs (documentation, performance, architecture)
- **v1.2 Release**: 2 TODOs (advanced error monitoring, token encryption)

## Maintenance

- Review TODO comments quarterly to reassess priorities
- Update target versions based on actual development progress
- Remove TODO comments when features are implemented
- Add new TODO comments for newly identified incomplete features

This document should be updated whenever new TODO comments are added or existing ones are resolved.

## Recent Updates

### April 2026
- **Implemented**: All 46 test coverage TODOs — every previously-stubbed test case now has a working implementation
- **Categories resolved**: Hash functions (4), FS integration (12), delta sync (7), inode (4), xattr (3), thumbnail (3), upload manager/session (3), offline (3), OAuth2 GTK (1), config (4), upload manager partial (2)
- **Verified**: All tests pass in Docker (`docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests`)

### January 2025
- **Updated**: Complete audit of all TODO comments in codebase
- **Identified**: 55+ TODO comments across multiple categories
- **Prioritized**: Critical QuickXORHash testing for OneDrive file integrity
- **Categorized**: Test implementation TODOs represent majority of outstanding work

### Previous Updates (June 2024)
- **Fixed**: Database persistence hanging issue in `internal/fs/upload_signal_basic_test.go:162`
- **Implemented**: Path function tests in `internal/graph/path_test.go` (Issue #117)

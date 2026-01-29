# Tasks: Virtual File Management

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/tasks.md`. Sections below are verbatim.

## Phase 15: XDG Compliance Verification ✅ COMPLETE

- [x] 26. Verify XDG Base Directory compliance ✅ COMPLETE
- [x] 26.1 Review XDG implementation ✅ COMPLETE
- [x] 26.2 Test XDG_CONFIG_HOME environment variable ✅ COMPLETE
- [x] 26.3 Test XDG_CACHE_HOME environment variable ✅ COMPLETE
- [x] 26.4 Test default XDG paths ✅ COMPLETE
- [x] 26.5 Test command-line override ✅ COMPLETE
- [x] 26.6 Test directory permissions ✅ COMPLETE
- [x] 26.7 Document XDG compliance verification results ✅ COMPLETE

- [x] 26.8 Implement XDG compliance property-based tests
- [x] 26.8.1 Implement Property 37: XDG Configuration Directory Usage
  - **Property 37: XDG Configuration Directory Usage**
  - **Validates: Requirements 15.1**
  - Create `internal/config/xdg_property_test.go`
  - Generate random system configuration scenarios
  - Verify os.UserConfigDir() is used for configuration
  - Test with various XDG environment settings
  - _Requirements: 15.1_

- [x] 26.8.2 Implement Property 38: Token Storage Location
  - **Property 38: Token Storage Location**
  - **Validates: Requirements 15.7**
  - Generate random authentication token storage scenarios
  - Verify tokens are stored in configuration directory
  - Test storage location consistency
  - _Requirements: 15.7_

- [x] 26.8.3 Implement Property 39: Cache Storage Location
  - **Property 39: Cache Storage Location**
  - **Validates: Requirements 15.8**
  - Generate random file content caching scenarios
  - Verify cache is stored in cache directory
  - Test cache location consistency
  - _Requirements: 15.8_

**Status**: ✅ **COMPLETE** (2025-11-13)  
**Requirements**: 15.1-15.10 all verified  
**Results**: OneMount correctly implements XDG Base Directory Specification with only minor deviation (auth token location) that has minimal impact.

---

## Related Tasks (References Only)

- Validate Requirement 15.11–15.13 in `.kiro/specs/filesystem-mounting/tasks.md` (Phase 15: XDG Compliance Verification)


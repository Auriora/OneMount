# Requirements Document: Virtual File Management

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/requirements.md`. Sections below are verbatim.

## Requirements

### Requirement 2B: Virtual File Management

**User Story:** As a Linux desktop user, I want virtual files like `.xdg-volume-info` to work correctly so that my file manager displays proper volume information.

#### Acceptance Criteria

1. WHEN the path `.xdg-volume-info` is requested, THE OneMount System SHALL bypass Graph and cached metadata lookups and serve the virtual file immediately so that it is always available, even on first mount
2. WHEN representing filesystem entries that exist only locally (e.g., `.xdg-volume-info`, policy folders, or pinned views), THE OneMount System SHALL persist them as metadata records with `local-*` identifiers and overlay policies describing precedence so that the virtual view is resolved inside the metadata database without a separate wrapper layer

## Related Requirements (References Only)

- Requirement 15.11–15.13 (XDG Base Directory Compliance) in `.kiro/specs/filesystem-mounting/requirements.md`
  - `.xdg-volume-info` local-only virtual file behavior


---

### 1. Adopt Standard Go Project Layout
**Labels:** `architecture`, `refactor`

**Description**  
Reorganize the repository to follow the conventional Go layout by introducing an `internal/` directory for non-public packages and a `pkg/` directory for shared libraries.

**Rationale**  
Enforcing package encapsulation reduces accidental API exposure and aligns the project with community best practices.

**Impact**  
Low – Code imports will need updating, but overall behavior remains unchanged.

**Relevant Documentation**
- Requirements: ARCH-STD-001
- Architecture: Go Project Layout Guidelines
- Design: N/A
- Implementation: N/A

**Implementation Notes**
- Create `internal/fs`, `internal/ui`, etc., and move all non-public packages there.
- Create `pkg/utils` for any packages intended for external reuse.
- Update all import paths accordingly.
- Add a note in `README.md` about the new layout.

---

### 2. Refactor `main.go` into Discrete Services
**Labels:** `refactor`, `code-quality`

**Description**  
Extract large, multi-responsibility functions from `cmd/onedriver/main.go` (e.g., filesystem initialization, stats display) into dedicated service types (`AuthService`, `FilesystemService`, `StatsService`).

**Rationale**  
Reducing function size and adhering to Single Responsibility Principle improves readability, testability, and future maintainability.

**Impact**  
Medium – Significant internal restructuring, but public CLI behavior should be preserved.

**Relevant Documentation**
- Requirements: SRP-001
- Architecture: Clean/Hexagonal Architecture Proposal
- Design: N/A
- Implementation: N/A

**Implementation Notes**
- Define service interfaces in `internal/services`.
- Move setup and teardown logic into methods on those services.
- Wire them together in `main.go` via constructor injection.
- Add unit tests against each new service.

---

### 3. Introduce Dependency Injection for External Clients
**Labels:** `testing`, `refactor`

**Description**  
Define interfaces for all external dependencies (e.g., Graph API client, D-Bus adapter) and pass concrete implementations in at startup rather than instantiating them directly.

**Rationale**  
This decoupling makes it possible to mock external services in tests and swap implementations more easily.

**Impact**  
Medium – Changes to constructors and service signatures; improves test coverage.

**Relevant Documentation**
- Requirements: DI-TEST-002
- Architecture: Service Interface Definitions
- Design: N/A
- Implementation: N/A

**Implementation Notes**
- Create interfaces (e.g., `type GraphClient interface { … }`) in `internal/api`.
- Refactor existing code to depend on interfaces.
- Provide default implementations in `pkg/graph`.
- Update CI to include interface-based mocks in unit tests.

---

### 4. Enhance Project Documentation
**Labels:** `documentation`, `good first issue`

**Description**  
Improve `README.md` and `docs/DEVELOPMENT.md` by adding a table of contents, contribution guidelines, code-of-conduct, and a high-level architecture diagram.

**Rationale**  
Better onboarding for new contributors and clearer governance helps grow the community around the project.

**Impact**  
Low – Documentation only.

**Relevant Documentation**
- Requirements: DOCS-ONB-001
- Architecture: N/A
- Design: N/A
- Implementation: N/A

**Implementation Notes**
- Generate a TOC via markdown-anchor links.
- Write `CONTRIBUTING.md` and `CODE_OF_CONDUCT.md`.
- Sketch a simple UML/service-diagram and embed in `docs/`.
- Add badges (build status, coverage) to `README.md`.

---

### 5. Increase Test Coverage to ≥ 80%
**Labels:** `testing`, `enhancement`

**Description**  
Add new table-driven unit tests and integration tests to raise overall coverage to at least 80%, focusing especially on error paths, boundary conditions, and concurrency scenarios.

**Rationale**  
Higher coverage provides confidence in code correctness and reduces regressions.

**Impact**  
Medium – Requires writing many new tests but does not change production code.

**Relevant Documentation**
- Requirements: TEST-COV-003
- Architecture: N/A
- Design: N/A
- Implementation: N/A

**Implementation Notes**
- Use Go’s `testing` package with sub-tests (`t.Run`).
- Add mocks for injected interfaces to simulate failures.
- Include tests for `DeltaLoop` cancellation once context support is added.
- Add Python extension tests if applicable.

---

### 6. Implement Context-Based Concurrency Cancellation
**Labels:** `enhancement`, `concurrency`

**Description**  
Replace raw goroutine launches (e.g., in `DeltaLoop`) with context-aware `go func(ctx)` patterns and use `sync.WaitGroup` to manage shutdown.

**Rationale**  
Graceful shutdown prevents orphaned goroutines and resource leaks, improving reliability on exit or reload.

**Impact**  
Medium – Changes to concurrency model; tests will need updating.

**Relevant Documentation**
- Requirements: CONC-CTL-004
- Architecture: N/A
- Design: N/A
- Implementation: N/A

**Implementation Notes**
- Pass `context.Context` from `main` into all long-running routines.
- Use `select { case <-ctx.Done(): return }` in loops.
- Signal cancellation on SIGINT/SIGTERM.
- Update integration tests to verify shutdown behavior.

---

### 7. Standardize Error Handling Across Modules
**Labels:** `refactor`, `logging`

**Description**  
Adopt a consistent error-wrapping and logging strategy using Go’s standard `errors` package (e.g., `fmt.Errorf("…: %w", err)`) or a chosen wrapper library.

**Rationale**  
Uniform patterns make it easier to trace errors and produce structured logs.

**Impact**  
Low – Changes are largely mechanical but improve observability.

**Relevant Documentation**
- Requirements: ERR-HNDL-005
- Architecture: N/A
- Design: N/A
- Implementation: N/A

**Implementation Notes**
- Audit all `if err != nil` blocks and apply a standard wrap.
- Centralize log formatting (timestamp, level, module).
- Remove any inconsistent or duplicate logging calls.
- Add tests to verify error chains where appropriate.

---

# Kiro Specs and Steering Reference (OneMount)

**Last updated**: 2026-01-28

## Purpose

This document captures how Kiro defines specs and steering files and how OneMount stores and maintains them. Use it as a quick reference when updating `.kiro/specs/` or `.kiro/steering/`.

## Kiro Specs (Official Model)

Kiro specs are a three-file bundle that bridge requirements and implementation. Kiro defines the following files for each spec:

- `requirements.md`: User stories and acceptance criteria, written in structured EARS notation.
- `design.md`: Technical design details, including architecture and sequence diagrams where needed.
- `tasks.md`: A discrete, trackable implementation plan.

## OneMount Spec Layout

**Root index**: `.kiro/specs/README.md`
- Tracks active specs, status, and dependencies.
- Defines the standard spec structure and workflow.

**Per-spec structure**:

```
.kiro/specs/<spec-name>/
├── README.md
├── requirements.md
├── design.md
└── tasks.md
```

**Per-spec README fields (common pattern)**:
- **Status**: Current status, Created date, Last Updated date
- **Contents**: Links to requirements/design/tasks
- **Scope**: Feature coverage
- **Dependencies**: Upstream/downstream dependencies
- **Related docs/specs**: SRS and architecture references

**Lifecycle notes (from `.kiro/specs/README.md`)**:
- **Create**: add folder, copy template, update README index
- **Update**: modify the relevant file, update `Last Updated`, adjust status
- **Complete**: mark tasks complete, set status to “Completed”, document deviations, update index
- **Archive**: move legacy specs to `.kiro/specs/archive/` when superseded

## Kiro Steering (Official Model)

Steering files provide project-specific context and guidelines for Kiro. Steering lives in `.kiro/steering/` and uses frontmatter to control inclusion behavior:

- `inclusion: always` — Always included in context.
- `inclusion: fileMatch` + `fileMatchPattern` — Included when matching files are in scope.
- `inclusion: manual` — Added only when invoked (e.g., via a slash command).

## OneMount Steering Files

| File | Inclusion | Scope | Priority | Purpose |
| --- | --- | --- | --- | --- |
| `coding-standards.md` | `always` | `src/**` | 100 | Coding standards and mandatory design principles. |
| `general-preferences.md` | `always` | `.*` | 50 | Rule discovery, prioritization, and documentation expectations. |
| `operational-best-practices.md` | `always` | `.*` | 40 | Tool usage, process transparency, SRS alignment, and test protocol reminders. |
| `planning-protocol.md` | `manual` | `planning-flow` | 30 | Structured planning flow for complex, multi-file changes. |
| `testing-conventions.md` | `always` | `tests/**`, `internal/**/*_test.go` | 25 | Docker-only testing requirements and test conventions. |
| `documentation-conventions.md` | `fileMatch` | `docs/**` | 20 | Docs structure, placement, and update-log requirements. |
| `git-conventions.md` | `always` | `git-*` | 15 | Commit message and branching conventions. |

## Practical Workflow Reminders

- **Docs work**: If you touch `docs/**`, the documentation conventions apply and you must add an update entry in `docs/updates/`.
- **Tests**: OneMount tests must run in Docker and follow the timeout wrapper rules in `testing-conventions.md`.
- **Planning**: For complex, multi-file work, invoke the planning protocol manually.
- **Specs**: Keep spec README dates current and update `.kiro/specs/README.md` when status changes.

## Local References

- [`.kiro/specs/README.md`](../../../.kiro/specs/README.md)
- [`.kiro/specs/<spec-name>/README.md`](../../../.kiro/specs/authentication-and-accounts/README.md)
- [`.kiro/steering/`](../../../.kiro/steering/)
- [`docs/guides/ai-agent/`](../ai-agent/)

## External References

- [Kiro specs concepts (requirements/design/tasks structure)](https://kiro.dev/docs/specs/concepts/)
- [Kiro steering inclusion modes](https://kiro.dev/docs/steering/)

# AGENTS

This repository uses centralized agent instructions located in `docs/guides/ai-agent/`.

- The rule files in `docs/guides/ai-agent/` are the single source of truth and take precedence over any guidance elsewhere in the repo.
- To avoid duplication or conflicts, this file intentionally does not restate operational commands, workflows, or protocols.

For general project context and developer documentation, refer to:
- `README.md` (project overview, commands, architecture)
- `docs/` (AI model configuration, Batch API, GPT-5 parameter notes, plan/execute workflow, migration guide)
- `CLAUDE.md` (editor-specific tips, if applicable)

If you are implementing or running agents/tools:
- Load `docs/guides/ai-agent/` at task start and follow the highest-priority instructions found there.
- Log task-scoped notes and updates in `docs/updates/` using the repository template.

## 🚨 CRITICAL: Testing Protocol (MANDATORY)

When executing tests in the OneMount project, you MUST follow these rules:

1. **ALWAYS use Docker** - Never run `go test` directly on host
2. **ALWAYS use timeout wrapper** - For integration/system tests: `./scripts/timeout-test-wrapper.sh "TestPattern" 60`
3. **NEVER use `cd` command** - The workspace root is already correct; use `cwd` parameter if needed
4. **ALWAYS include auth override** - For tests requiring authentication: `-f docker/compose/docker-compose.auth.yml`

**Violation of these rules will result in test failures and environment corruption.**

See `.kiro/steering/testing-conventions.md` for complete details.

---

This document is intentionally minimal to prevent divergence from `docs/guides/ai-agent/`. Consult those rule files first, and prefer updating them over this file when behavior or priorities change.

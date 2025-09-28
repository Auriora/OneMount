# Debugging Guide (Developers)

Purpose: Practical procedures to diagnose issues quickly without changing dependencies or system state. For end-user fixes, see the Troubleshooting Guide.

Related docs: [Logging Guidelines](logging-guidelines.md), [Logging Examples](logging-examples.md), [Development CLI Guide](../../scripts/README.md), [Docker Dev Workflow](../docker-development-workflow.md), [Test Architecture](../2-architecture-and-design/test-architecture-design.md), [Troubleshooting Guide](troubleshooting-guide.md).

## Quick checklist
- Reproduce with minimal steps and note exact command/args.
- Capture logs (see below) and environment details (distro, kernel, OneMount version).
- Validate with latest main or the tagged build you are testing.
- Avoid using production auth tokens for tests (see token safety below).

## Logging and verbosity
- One-off verbose run (foreground):
  - GUI launcher: run the CLI directly to capture logs
    - `ONEMOUNT_DEBUG=1 onemount /path/to/mount`
  - Help and args: `onemount --help`
- Systemd user logs (packaged service):
  - Show recent logs for all OneMount user units:
    - `journalctl --user -u onemount@* -S today -o short-iso`
  - Follow live logs:
    - `journalctl --user -u onemount@* -f`
- Increase detail via environment:
  - `ONEMOUNT_DEBUG=1` enables verbose output in many paths.
  - For targeted areas, prefer component-specific flags or config if available.

## Common diagnostics
- Filesystem mount check:
  - Verify mount exists and is accessible: `mount | grep -i onemount` and `ls -l /path/to/mount`
- D-Bus integration
  - Monitor signals: `dbus-monitor --session "type='signal',sender='org.onemount'"`
  - List names: `gdbus list --session`
- Graph API boundaries (no real tokens in logs):
  - Confirm environment-based proxies are off if testing local failures.
  - Simulate offline: disable network for the process/sandbox and observe retry behavior.

## Using the development CLI (scripts/dev.py)
- Show environment status: `./scripts/dev info`
- Run tests with coverage (fast iteration):
  - `./scripts/dev test coverage --threshold-line 80`
  - Smallest scope first (package/file) when possible.
- System tests (controlled): `./scripts/dev test system --category smoke`
- Quality analysis: `./scripts/dev analyze test-suite --mode resolve`
- Build packages (no push): `./scripts/dev build deb --docker`
- Get help: `./scripts/dev --help` and subcommand `--help`.

## Docker test shell (manual repro)
- Launch a disposable test shell:
  - `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
- Inside shell:
  - Run CLI: `onemount --help` or targeted commands.
  - Mount to a temp path under `/workspace` and validate basics.

## Nemo extension and D-Bus tips
- Python Nemo extension path: `internal/nemo/src/nemo-onemount.py`.
- Focus on D-Bus signals and method calls rather than UI when isolating issues.
- Use `dbus-monitor` to confirm events arrive; compare with Nemo log output if available.
- Example: monitor OneMount FileManager signals only:

  ```bash
  dbus-monitor --session "type='signal',sender='org.onemount',interface='org.onemount.FileManager'"
  ```


## Token safety (testing)
- Do NOT use production tokens in tests.
  - Production location: `~/.cache/onemount/auth_tokens.json` — keep protected.
- OneMount stores test-time tokens in cache subdirectories named after the mount path; prefer isolated mounts like `/tmp/onemount-test-*`.
- Use scripts under `scripts/` (e.g., `setup-test-auth.sh`) to prepare safe test credentials.

## Capturing a useful bug report (developers)
- Commands executed, exact output (trim secrets), environment (distro, kernel, desktop, versions).
- Log excerpts around the failure with timestamps.
- If concurrency-related, note whether the issue reproduces with a single file op.
- For offline/online transitions, include the timing of network loss/restore.

## Next steps
- If a defect is confirmed, add a minimal failing test first (unit or integration) before code changes.
- Link logs and reproduction notes in the PR/issue; keep sensitive data out of the repo.


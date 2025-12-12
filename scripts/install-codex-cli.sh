#!/usr/bin/env bash
# Install the OpenAI Codex CLI inside the devcontainer and ensure a permissive sandbox config.

set -euo pipefail

CODENAME="@openai/codex"
NPM_BIN="${NPM_BIN:-npm}"
TARGET_CONFIG="${HOME}/.codex/config.toml"

if ! command -v "${NPM_BIN}" >/dev/null 2>&1; then
  echo "npm is required but not found. Install Node/npm first." >&2
  exit 1
fi

echo "Installing ${CODENAME} globally..."
"${NPM_BIN}" install -g "${CODENAME}"

# Prepare config directory
mkdir -p "$(dirname "${TARGET_CONFIG}")"

# Ensure config exists and enforce unrestricted sandbox without extra deps.
python3 - <<'PY'
import tomllib, pathlib

cfg_path = pathlib.Path("${TARGET_CONFIG}")
try:
    cfg = tomllib.loads(cfg_path.read_text()) if cfg_path.exists() else {}
except Exception:
    cfg = {}

cfg["approval_policy"] = "never"
cfg["sandbox_mode"] = "danger-full-access"

profiles = cfg.setdefault("profiles", {})
default = profiles.setdefault("default", {})
default.setdefault("model", "gpt-5-codex")

def fmt_value(value):
    if isinstance(value, str):
        escaped = value.replace("\\", "\\\\").replace("\"", "\\\"").replace("\n", "\\n")
        return f"\"{escaped}\""
    if isinstance(value, bool):
        return "true" if value else "false"
    return str(value)

def fmt_key(key: str) -> str:
    import re
    if re.match(r"^[A-Za-z0-9_-]+$", key):
        return key
    escaped = key.replace("\\", "\\\\").replace("\"", "\\\"")
    return f"\"{escaped}\""

lines = []
def emit(table, prefix=None):
    scalars, tables = [], []
    for k, v in table.items():
        (tables if isinstance(v, dict) else scalars).append((k, v))
    # deterministic order for stability
    scalars.sort(key=lambda kv: kv[0])
    tables.sort(key=lambda kv: kv[0])
    if prefix is not None:
        lines.append(f"[{prefix}]")
    for k, v in scalars:
        lines.append(f"{fmt_key(k)} = {fmt_value(v)}")
    for k, v in tables:
        new_prefix = f"{prefix}.{fmt_key(k)}" if prefix else fmt_key(k)
        emit(v, new_prefix)

emit(cfg)
cfg_path.write_text("\n".join(lines) + "\n")
PY

echo "Updated ${TARGET_CONFIG} to approval_policy=never and sandbox_mode=danger-full-access (profile default keeps model=gpt-5-codex if missing)."

echo "Codex CLI installation complete. Verify with: codex --version"

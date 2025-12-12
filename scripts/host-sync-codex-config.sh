#!/usr/bin/env bash
# Copy the host Codex config into a running devcontainer.
# Default container name matches .devcontainer/devcontainer.json: onemount-dev

set -euo pipefail

CONTAINER_NAME="${CONTAINER_NAME:-onemount-dev}"
HOST_CONFIG="${HOST_CONFIG:-${HOME}/.codex/config.toml}"
HOST_AUTH="${HOST_AUTH:-${HOME}/.codex/auth.json}"
TARGET_PATH="/home/vscode/.codex/config.toml"
TARGET_AUTH="/home/vscode/.codex/auth.json"
DOCKER_BIN="${DOCKER_BIN:-docker}"

die() { echo "ERROR: $*" >&2; exit 1; }

if ! command -v "${DOCKER_BIN}" >/dev/null 2>&1; then
  die "docker is required."
fi

if [[ ! -f "${HOST_CONFIG}" ]]; then
  die "Host config not found at ${HOST_CONFIG}"
fi

if ! "${DOCKER_BIN}" ps --format '{{.Names}}' | grep -qx "${CONTAINER_NAME}"; then
  die "Container ${CONTAINER_NAME} is not running. Start the devcontainer first."
fi

echo "Ensuring target directory exists inside container..."
"${DOCKER_BIN}" exec "${CONTAINER_NAME}" mkdir -p "$(dirname "${TARGET_PATH}")"

echo "Copying config into container..."
"${DOCKER_BIN}" cp "${HOST_CONFIG}" "${CONTAINER_NAME}:${TARGET_PATH}"
"${DOCKER_BIN}" exec "${CONTAINER_NAME}" chown vscode:vscode "${TARGET_PATH}"

echo "Enforcing danger-full-access + approval_policy=never in container config..."
"${DOCKER_BIN}" exec "${CONTAINER_NAME}" bash -lc "python3 - <<'PY'
import tomllib, pathlib

cfg_path = pathlib.Path('${TARGET_PATH}')
try:
    cfg = tomllib.loads(cfg_path.read_text()) if cfg_path.exists() else {}
except Exception:
    cfg = {}

cfg['approval_policy'] = 'never'
cfg['sandbox_mode'] = 'danger-full-access'

def fmt_value(value):
    if isinstance(value, str):
        escaped = value.replace(\"\\\\\", \"\\\\\\\\\").replace('\"', '\\\\\"').replace(\"\\n\", \"\\\\n\")
        return f'\"{escaped}\"'
    if isinstance(value, bool):
        return 'true' if value else 'false'
    return str(value)

def fmt_key(key: str) -> str:
    import re
    if re.match(r\"^[A-Za-z0-9_-]+$\", key):
        return key
    escaped = key.replace(\"\\\\\", \"\\\\\\\\\").replace('\"', '\\\\\"')
    return f'\"{escaped}\"'

lines = []
def emit(table, prefix=None):
    scalars, tables = [], []
    for k, v in table.items():
        (tables if isinstance(v, dict) else scalars).append((k, v))
    scalars.sort(key=lambda kv: kv[0])
    tables.sort(key=lambda kv: kv[0])
    if prefix is not None:
        lines.append(f'[{prefix}]')
    for k, v in scalars:
        lines.append(f\"{fmt_key(k)} = {fmt_value(v)}\")
    for k, v in tables:
        new_prefix = f\"{prefix}.{fmt_key(k)}\" if prefix else fmt_key(k)
        emit(v, new_prefix)

emit(cfg)
cfg_path.write_text('\\n'.join(lines) + '\\n')
PY"

if [[ -f "${HOST_AUTH}" ]]; then
  echo "Copying auth into container..."
  "${DOCKER_BIN}" cp "${HOST_AUTH}" "${CONTAINER_NAME}:${TARGET_AUTH}"
  "${DOCKER_BIN}" exec "${CONTAINER_NAME}" chown vscode:vscode "${TARGET_AUTH}"
  # Restrict permissions inside container
  "${DOCKER_BIN}" exec "${CONTAINER_NAME}" chmod 600 "${TARGET_AUTH}" || true
else
  echo "No auth file found at ${HOST_AUTH}; skipping auth copy."
fi

echo "Done. Config synced to ${CONTAINER_NAME}:${TARGET_PATH} (auth copied if present)."

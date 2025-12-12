#!/usr/bin/env bash

# Run a command inside a throwaway user session that has its own D-Bus session
# and `systemd --user` instance. Useful for running onemount-launcher/systemd
# integration tests without touching the host's systemd.

set -euo pipefail

if ! command -v dbus-run-session >/dev/null 2>&1; then
  echo "dbus-run-session is required (install dbus-user-session)" >&2
  exit 1
fi

SYSTEMD_BIN="$(command -v systemd || true)"
if [[ -z "${SYSTEMD_BIN}" ]]; then
  echo "systemd binary not found (install systemd package)" >&2
  exit 1
fi

SYSTEMCTL_BIN="$(command -v systemctl || true)"
if [[ -z "${SYSTEMCTL_BIN}" ]]; then
  echo "systemctl not found (install systemd package)" >&2
  exit 1
fi

uid=$(id -u)
gid=$(id -g)
export XDG_RUNTIME_DIR=${XDG_RUNTIME_DIR:-/run/user/${uid}}

if [[ ! -d "${XDG_RUNTIME_DIR}" ]]; then
  if mkdir -p "${XDG_RUNTIME_DIR}" 2>/dev/null; then
    chmod 700 "${XDG_RUNTIME_DIR}" || true
  else
    echo "${XDG_RUNTIME_DIR} is missing and could not be created; ensure it exists and is owned by UID ${uid}" >&2
    exit 1
  fi
fi

# Validate ownership and writability
owner_uid=$(stat -c %u "${XDG_RUNTIME_DIR}") || owner_uid=-1
if [[ "${owner_uid}" != "${uid}" ]]; then
  echo "${XDG_RUNTIME_DIR} is owned by UID ${owner_uid}; need ownership ${uid}" >&2
  exit 1
fi

if [[ ! -w "${XDG_RUNTIME_DIR}" ]]; then
  echo "${XDG_RUNTIME_DIR} is not writable" >&2
  exit 1
fi

cmd=("$@")
if [[ ${#cmd[@]} -eq 0 ]]; then
  cmd=("bash")
fi

if dbus-run-session --help 2>/dev/null | grep -q -- '--systemd'; then
  # Preferred path: dbus-run-session supports --systemd and will manage the user instance
  exec dbus-run-session --systemd bash -lc "export XDG_RUNTIME_DIR='${XDG_RUNTIME_DIR}'; exec ${cmd[*]}"
fi

# Fallback path: older dbus-run-session without --systemd. Start user systemd manually.
dbus-run-session bash -lc "\
  export XDG_RUNTIME_DIR='${XDG_RUNTIME_DIR}';\
  systemd --user >/tmp/systemd-user.log 2>&1 &\
  SYSTEMD_PID=\$!;\
  for i in \$(seq 1 50); do\
    if ! kill -0 \$SYSTEMD_PID 2>/dev/null; then\
      echo 'systemd --user exited early; see /tmp/systemd-user.log' >&2; exit 1;\
    fi;\
    if systemctl --user is-system-running >/dev/null 2>&1; then\
      exec ${cmd[*]};\
    fi;\
    sleep 0.2;\
  done;\
  echo 'Timed out waiting for systemd --user to become ready' >&2;\
  exit 1;\
"
exit $?

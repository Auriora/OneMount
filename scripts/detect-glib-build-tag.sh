#!/usr/bin/env bash

set -euo pipefail

if ! command -v pkg-config >/dev/null 2>&1; then
    exit 0
fi

version="$(pkg-config --modversion glib-2.0 2>/dev/null || true)"

if [[ -z "${version}" ]]; then
    exit 0
fi

IFS='.' read -r major minor _ <<< "${version}"

if [[ -z "${major}" || -z "${minor}" ]]; then
    exit 0
fi

# gotk3 only needs build tags when GLib is older than 2.68.
if (( major != 2 || minor >= 68 )); then
    exit 0
fi

supported_minor_versions=(40 42 44 46 48 50 52 54 56 58 60 62 64 66)
selected_tag=""

for candidate in "${supported_minor_versions[@]}"; do
    if (( candidate <= minor )); then
        selected_tag=$(printf "glib_%d_%02d" "${major}" "${candidate}")
    fi
done

if [[ -n "${selected_tag}" ]]; then
    printf "%s\n" "${selected_tag}"
fi

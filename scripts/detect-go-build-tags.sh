#!/usr/bin/env bash
# Detect applicable Go build tags for this environment.
# Currently handles GLib compatibility tags and WebKit2GTK version tags.

set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

glib_tag=""
webkit_tag=""

if [[ -x "${script_dir}/detect-glib-build-tag.sh" ]]; then
    glib_tag=$("${script_dir}/detect-glib-build-tag.sh" 2>/dev/null || true)
fi

if [[ -x "${script_dir}/detect-webkit-version.sh" ]]; then
    webkit_version=$("${script_dir}/detect-webkit-version.sh" 2>/dev/null || true)
    case "${webkit_version}" in
        webkit2gtk-4.1)
            webkit_tag="webkit41"
            ;;
        webkit2gtk-4.0)
            webkit_tag="webkit40"
            ;;
        *)
            webkit_tag=""
            ;;
    esac
fi

tags=()
if [[ -n "${glib_tag}" ]]; then
    tags+=("${glib_tag}")
fi
if [[ -n "${webkit_tag}" ]]; then
    tags+=("${webkit_tag}")
fi

if [[ ${#tags[@]} -gt 0 ]]; then
    IFS=, printf "%s" "${tags[*]}"
fi

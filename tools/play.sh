#!/usr/bin/env bash
set -euo pipefail
project=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
version=$(cat "$project/VERSION")
edition=CHT-HD
if [[ ${1:-} == --cht ]]; then edition=CHT; shift; fi
if [[ ${1:-} == --cht-hd ]]; then edition=CHT-HD; shift; fi
appimage="$project/dist-all/$version/full-local/Zeliard-$version-$edition-Linux-x86_64.AppImage"
if [[ ! -f "$appimage" ]]; then appimage="$project/dist-all/$version/full-local/Zeliard-$version-Linux-x86_64.AppImage"; fi
if [[ -f "$appimage" ]]; then exec "$appimage" "$@"; fi
export ZELIARD_BUNDLE="$project/dist-all/$version/full-local"
exec bash "$project/tools/run-local.sh" "$@"

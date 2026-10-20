#!/usr/bin/env bash
set -euo pipefail
project=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
version=$(cat "$project/VERSION")
export ZELIARD_BUNDLE="$project/dist-all/$version/full-local"
exec bash "$project/tools/run-local.sh" "$@"

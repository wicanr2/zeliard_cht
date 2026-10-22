#!/usr/bin/env bash
set -euo pipefail
project=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
preview="$project/workplace/ui-text-v1/preview"
if [[ ! -x "$preview/zeliard-preview" || ! -f "$preview/runtime-args.json" ]]; then
  echo '找不到字級候選，建置方式見 docs/spec/064-readable-system-text.md。' >&2
  exit 1
fi
export LD_LIBRARY_PATH="$preview/lib${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
exec "$preview/zeliard-preview" "$@"

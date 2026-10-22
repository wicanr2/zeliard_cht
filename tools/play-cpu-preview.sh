#!/usr/bin/env bash
set -euo pipefail
project=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
preview="$project/workplace/cpu-speed-v1/preview"
if [[ ! -x "$preview/zeliard-preview" || ! -f "$preview/runtime-args.json" ]]; then
  echo '找不到CPU速度候選版，請參考 docs/spec/059-cpu-speed-toolbar.md 的本機建置入口。' >&2
  exit 1
fi
export LD_LIBRARY_PATH="$preview/lib${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
exec "$preview/zeliard-preview" "$@"

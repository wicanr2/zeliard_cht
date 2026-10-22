#!/usr/bin/env bash
set -euo pipefail
project=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
work="$project/workplace/ui-text-v1"
cache=/home/anr2/cht/dosgolem/workplace/gomodcache
[[ -d "$work/preview" && -f "$work/build.work" && -d "$project/workplace/gocache" && -d "$cache" ]]
[[ -d "$project/workplace/release-v104/dosgolem" && -d "$project/workplace/cpu-speed-v1/preview" ]]
common=(--rm --label project=zeliard --network none --user "$(id -u):$(id -g)" --cpus 1 --pids-limit 128)
mounts=(-v "$project:/project:ro" -v "$work:/project/workplace/ui-text-v1:rw")
timeout 30 docker run "${common[@]}" --memory 256m "${mounts[@]}" python:3.13-slim python3 -c 'from pathlib import Path; import os; p=Path("/project/workplace/ui-text-v1"); assert p.stat().st_uid==os.getuid(); bad=[str(f) for f in p.rglob("*") if not f.is_symlink() and f.stat().st_uid!=os.getuid()]; assert not bad,bad'
timeout 90 docker run "${common[@]}" --memory 768m "${mounts[@]}" eob-audio-capture:20261009-r3 python3 /project/tools/prepare_rounded_font.py --project /project --output /project/workplace/ui-text-v1/rounded-font
timeout 30 docker run "${common[@]}" --memory 256m "${mounts[@]}" python:3.13-slim python3 /project/tools/prepare_ui_text.py
timeout 180 docker run "${common[@]}" --memory 1536m "${mounts[@]}" -v "$project/workplace/gocache:/gocache:rw" -v "$cache:/gomodcache:ro" -e GOWORK=/project/workplace/ui-text-v1/build.work -e GOCACHE=/gocache -e GOMODCACHE=/gomodcache -e GOPROXY=off -e GOSUMDB=off -e GOTOOLCHAIN=local -e GOMAXPROCS=1 -e GOFLAGS=-p=1 -w /project dosgolem-go-ebiten:1.24-r1 go build -trimpath -buildvcs=false -ldflags '-X main.version=development-ui-text-v1' -o workplace/ui-text-v1/preview/zeliard-preview ./cmd/zeliard-preview

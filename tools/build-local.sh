#!/usr/bin/env bash
set -euo pipefail
project=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
version=$(cat "$project/VERSION")
[[ $version =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]] || exit 1
golem=/home/anr2/cht/dosgolem
font_dir=${ZELIARD_FONT_DIR:-/home/anr2/cht/softworld_san/fonts}
[[ -d "$project/workplace/dosgolem-source" && -d "$golem/workplace/gomodcache" && -f "$font_dir/unifont.hex.gz" ]]
mkdir -p "$project/workplace/input-integration-v1" "$project/dist-all" "$project/workplace/gocache"
common=(--rm --network none --user "$(id -u):$(id -g)" --memory 2g --cpus 2 --pids-limit 128)
inputs=(--mount "type=bind,src=$project,dst=/project,readonly"
  --mount "type=bind,src=$project/workplace/input-integration-v1,dst=/project/workplace/input-integration-v1"
  --mount "type=bind,src=$project/dist-all,dst=/project/dist-all"
  --mount "type=bind,src=$font_dir,dst=/fonts,readonly")
timeout 60 docker run "${common[@]}" "${inputs[@]}" --entrypoint python3 \
  eob-audio-capture:20261009-r3 /project/tools/prepare_local.py prepare
timeout 180 docker run "${common[@]}" \
  --mount "type=bind,src=$project/workplace/input-integration-v1/delivery-build,dst=/build" \
  --mount "type=bind,src=$project/dist-all/$version/full-local,dst=/bundle" \
  --mount "type=bind,src=$project/workplace/dosgolem-source,dst=/dosgolem,readonly" \
  --mount "type=bind,src=$golem/workplace/gomodcache,dst=/gomodcache,readonly" \
  --mount "type=bind,src=$project/workplace/gocache,dst=/gocache" \
  -e GOWORK=/build/build.work -e GOCACHE=/gocache -e GOMODCACHE=/gomodcache \
  -e GOPROXY=off -e GOTOOLCHAIN=local -e GOMAXPROCS=2 -e "ZELIARD_BUILD_VERSION=$version" \
  -w /build/sources dosgolem-go-ebiten:1.24-r1 sh -c \
  'go test -count=1 ./internal/prototype/controls > /build/tests.log 2>&1 && go vet ./internal/prototype/controls ./cmd/zeliard-preview > /build/vet.log 2>&1 && go build -buildvcs=false -ldflags "-X main.version=$ZELIARD_BUILD_VERSION" -o /bundle/zeliard-preview ./cmd/zeliard-preview > /build/build.log 2>&1'
timeout 60 docker run "${common[@]}" "${inputs[@]}" --entrypoint python3 \
  eob-audio-capture:20261009-r3 /project/tools/prepare_local.py finalize

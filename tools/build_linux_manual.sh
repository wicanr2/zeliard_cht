#!/usr/bin/env bash
# 050使用者指定的無自動驗證Linux打包入口；不執行遊戲、test、vet或smoke。
set -euo pipefail
project=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
IFS= read -r version < "$project/VERSION"
[[ $version =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]
work="$project/workplace/full-hd-v1/manual-hd-v1"
[[ -d "$work/registered" && -d "$project/dist-all" && -d "$project/workplace/gocache" ]]
[[ -d "$project/workplace/music-release-v1/release-build/dosgolem" && -d /home/anr2/cht/dosgolem/workplace/gomodcache ]]
common=(--rm --label project=zeliard --network none --user "$(id -u):$(id -g)" --memory 3g --cpus 2 --pids-limit 256)
timeout 120 docker run "${common[@]}" -v "$project:/project:ro" -v "$work:/project/workplace/full-hd-v1/manual-hd-v1" \
  eob-audio-capture:20261009-r3 python3 /project/tools/prepare_rounded_font.py --project /project
timeout 120 docker run "${common[@]}" -v "$project:/project:ro" -v "$work:/project/workplace/full-hd-v1/manual-hd-v1" -v "$project/dist-all:/project/dist-all" \
  eob-audio-capture:20261009-r3 python3 /project/tools/package_linux_manual.py --project /project prepare
build="$work/linux-build/$version"
[[ -d "$build/AppDir" && -d "$project/dist-all/$version/full-local" ]]
timeout 600 docker run "${common[@]}" -v "$project:/project:ro" -v "$build:/build" \
  -v "$project/workplace/gocache:/gocache" -v /home/anr2/cht/dosgolem/workplace/gomodcache:/gomodcache:ro \
  -e GOWORK=/build/build.work -e GOCACHE=/gocache -e GOMODCACHE=/gomodcache -e GOPROXY=off -e GOSUMDB=off -e GOTOOLCHAIN=local -e GOMAXPROCS=2 \
  -e "BUILD_VERSION=$version" -w /project dosgolem-go-ebiten:1.24-r1 sh -c \
  'go build -trimpath -buildvcs=false -ldflags "-X main.version=$BUILD_VERSION" -o /build/AppDir/zeliard-preview ./cmd/zeliard-preview > /build/compile.log 2>&1'
timeout 180 docker run "${common[@]}" -e ARCH=x86_64 -v "$build:/build:ro" \
  -v "$project/dist-all/$version/full-local:/delivery" -e "BUILD_VERSION=$version" fd2-build-appimage:latest sh -c '
    set -e
    cd /tmp
    /opt/appimage-tools/appimagetool.AppImage --appimage-extract > /tmp/extract.log
    squashfs-root/AppRun --runtime-file /opt/appimage-tools/runtime-x86_64 /build/AppDir /delivery/Zeliard-$BUILD_VERSION-Linux-x86_64.AppImage'
timeout 120 docker run "${common[@]}" -v "$project:/project:ro" -v "$project/dist-all:/project/dist-all" \
  eob-audio-capture:20261009-r3 python3 /project/tools/package_linux_manual.py --project /project finish

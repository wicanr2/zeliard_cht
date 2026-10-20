#!/usr/bin/env bash
set -euo pipefail
project=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
work="$project/workplace/music-release-v1"
version=$(cat "$project/VERSION")
[[ $version =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]] || exit 1
[[ -d "$project/workplace/dosgolem-source" && -d /home/anr2/cht/dosgolem/workplace/gomodcache && -d "$work" ]]
common=(--rm --network none --user "$(id -u):$(id -g)" --memory 3g --cpus 2 --pids-limit 256)
phase=${1:-all}
if [[ $phase == all || $phase == prepare ]]; then
  timeout 90 docker run "${common[@]}" -v "$project:/project" \
    -v "$project/workplace/orig:/project/workplace/orig:ro" \
    -v "$project/workplace/dosgolem-source:/project/workplace/dosgolem-source:ro" \
    eob-audio-capture:20261009-r3 python3 /project/tools/prepare_release.py freeze
  timeout 90 docker run "${common[@]}" -v "$project:/project:ro" -v "$work:/project/workplace/music-release-v1" \
    eob-audio-capture:20261009-r3 python3 /project/tools/prepare_release.py assets
fi
build="$work/release-build"
[[ -d "$build/sources" && -d "$build/dosgolem" ]]
mounts=(-v "$build:/build" -v /home/anr2/cht/dosgolem/workplace/gomodcache:/gomodcache:ro -v "$project/workplace/gocache:/gocache")
envs=(-e GOWORK=/build/build.work -e GOCACHE=/gocache -e GOMODCACHE=/gomodcache -e GOPROXY=off -e GOSUMDB=off -e GOTOOLCHAIN=local -e GOMAXPROCS=2 -e "BUILD_VERSION=$version")
if [[ $phase == all || $phase == compile ]]; then
  timeout 600 docker run "${common[@]}" "${mounts[@]}" "${envs[@]}" -w /build/sources \
    dosgolem-go-ebiten:1.24-r1 sh -c '
      set -e
      mkdir -p /build/linux /build/windows /build/mac-amd64 /build/mac-arm64
      gcc -O2 -fPIC -shared -fvisibility=hidden -o /build/linux/libzopl.so third_party/nuked-opl3/opl3.c third_party/nuked-opl3/bridge.c
      ZELIARD_FM_LIBRARY=/build/linux/libzopl.so go test -count=1 ./internal/prototype/fm ./internal/prototype/sound ./internal/prototype/controls > /build/tests.log 2>&1
      go vet ./internal/prototype/fm ./internal/prototype/sound ./internal/prototype/controls ./cmd/zeliard-preview > /build/vet.log 2>&1
      go build -buildvcs=false -ldflags "-X main.version=$BUILD_VERSION" -o /build/linux/zeliard-preview ./cmd/zeliard-preview > /build/linux/build.log 2>&1
      GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -ldflags "-H windowsgui -X main.version=$BUILD_VERSION" -o /build/windows/Zeliard.exe ./cmd/zeliard-preview > /build/windows/build.log 2>&1
      ldd /build/linux/zeliard-preview > /build/linux/ldd.log
      mkdir -p /build/linux/lib
      for name in libasound.so.2 libX11.so.6 libXcursor.so.1 libXrandr.so.2 libXi.so.6 libXinerama.so.1 libXxf86vm.so.1 libXext.so.6 libXrender.so.1 libXfixes.so.3 libxcb.so.1 libXau.so.6 libXdmcp.so.6 libbsd.so.0 libmd.so.0; do
        if test -f /lib/x86_64-linux-gnu/"$name"; then cp -L /lib/x86_64-linux-gnu/"$name" /build/linux/lib/; fi
      done'
  timeout 120 docker run "${common[@]}" "${mounts[@]}" -w /build/sources \
    taipei-windows-audio:go1.24.13-mingw12.2-r1 sh -c \
    'x86_64-w64-mingw32-gcc -O2 -shared -static-libgcc -o /build/windows/zopl.dll third_party/nuked-opl3/opl3.c third_party/nuked-opl3/bridge.c'
  timeout 900 docker run "${common[@]}" "${mounts[@]}" "${envs[@]}" -e MACOSX_DEPLOYMENT_TARGET=11.0 -w /build/sources \
    tokugawa-desktop-platforms:go1.24.13-sdk15.5-noble-r1 sh -c '
      set -e
      for arch in amd64 arm64; do
        if test "$arch" = amd64; then cc=x86_64-apple-darwin24.5-clang; else cc=aarch64-apple-darwin24.5-clang; fi
        "$cc" -O2 -dynamiclib -fvisibility=hidden -install_name @rpath/libzopl.dylib -o /build/mac-$arch/libzopl.dylib third_party/nuked-opl3/opl3.c third_party/nuked-opl3/bridge.c
        GOOS=darwin GOARCH=$arch CGO_ENABLED=1 CC="$cc" go build -buildvcs=false -ldflags "-X main.version=$BUILD_VERSION" -o /build/mac-$arch/zeliard-preview ./cmd/zeliard-preview > /build/mac-$arch/build.log 2>&1
      done
      mkdir -p /build/macos
      lipo -create /build/mac-amd64/zeliard-preview /build/mac-arm64/zeliard-preview -output /build/macos/zeliard-preview
      lipo -create /build/mac-amd64/libzopl.dylib /build/mac-arm64/libzopl.dylib -output /build/macos/libzopl.dylib'
fi
if [[ $phase == all || $phase == package ]]; then
  timeout 120 docker run "${common[@]}" -v "$project:/project:ro" -v "$work:/project/workplace/music-release-v1" -v "$project/dist-all:/project/dist-all" \
    fd2-promo:20261017-r1 python3 /project/tools/package_release.py
  timeout 180 docker run "${common[@]}" -e ARCH=x86_64 -e "BUILD_VERSION=$version" \
    -v "$build:/build" -v "$project/dist-all/$version/full-local:/delivery" \
    fd2-build-appimage:latest sh -c '
      set -e
      cd /tmp
      /opt/appimage-tools/appimagetool.AppImage --appimage-extract > /tmp/extract.log
      squashfs-root/AppRun --runtime-file /opt/appimage-tools/runtime-x86_64 /build/AppDir /delivery/Zeliard-$BUILD_VERSION-Linux-x86_64.AppImage'
fi

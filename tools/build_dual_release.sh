#!/usr/bin/env bash
set -euo pipefail
project=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
IFS= read -r version < "$project/VERSION"
[[ $version =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]
version_core=${version#v.}
version_core=${version_core%%-*}
work="$project/workplace/release-v${version_core//./}"
[[ -d "$work" && -d "$project/dist-all" && -d "$project/workplace/gocache" && -d /home/anr2/cht/dosgolem/workplace/gomodcache ]]
common=(--rm --label project=zeliard --network none --user "$(id -u):$(id -g)" --memory 3g --cpus 2 --pids-limit 256)
phase=${1:-all}
if [[ $phase == all || $phase == prepare ]]; then
  timeout 180 docker run "${common[@]}" -v "$project:/project:ro" -v "$work:$work:rw" -v "$project/dist-all:/project/dist-all:rw" -e "ZELIARD_RELEASE_WORK=$work" eob-audio-capture:20261009-r3 python3 /project/tools/package_dual_release.py prepare
fi
mounts=(-v "$work:/build:rw" -v /home/anr2/cht/dosgolem/workplace/gomodcache:/gomodcache:ro -v "$project/workplace/gocache:/gocache:rw")
envs=(-e GOWORK=/build/build.work -e GOCACHE=/gocache -e GOMODCACHE=/gomodcache -e GOPROXY=off -e GOSUMDB=off -e GOTOOLCHAIN=local -e GOMAXPROCS=2 -e "BUILD_VERSION=$version")
if [[ $phase == all || $phase == compile ]]; then
  timeout 600 docker run "${common[@]}" "${mounts[@]}" "${envs[@]}" -w /build/sources dosgolem-go-ebiten:1.24-r1 sh -c '
    set -e
    gcc -O2 -fPIC -shared -fvisibility=hidden -o /build/linux/libzopl.so third_party/nuked-opl3/opl3.c third_party/nuked-opl3/bridge.c
    go build -trimpath -buildvcs=false -ldflags "-X main.version=$BUILD_VERSION" -o /build/linux/zeliard-preview ./cmd/zeliard-preview > /build/linux/build.log 2>&1
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags "-H windowsgui -X main.version=$BUILD_VERSION" -o /build/windows/Zeliard.exe ./cmd/zeliard-preview > /build/windows/build.log 2>&1
    mkdir -p /build/linux/lib
    for name in libasound.so.2 libX11.so.6 libXcursor.so.1 libXrandr.so.2 libXi.so.6 libXinerama.so.1 libXxf86vm.so.1 libXext.so.6 libXrender.so.1 libXfixes.so.3 libxcb.so.1 libXau.so.6 libXdmcp.so.6 libbsd.so.0 libmd.so.0; do
      if test -f /lib/x86_64-linux-gnu/"$name"; then cp -L /lib/x86_64-linux-gnu/"$name" /build/linux/lib/; fi
    done'
  timeout 120 docker run "${common[@]}" "${mounts[@]}" -w /build/sources taipei-windows-audio:go1.24.13-mingw12.2-r1 x86_64-w64-mingw32-gcc -O2 -shared -static-libgcc -o /build/windows/zopl.dll third_party/nuked-opl3/opl3.c third_party/nuked-opl3/bridge.c
  timeout 900 docker run "${common[@]}" "${mounts[@]}" "${envs[@]}" -e MACOSX_DEPLOYMENT_TARGET=11.0 -w /build/sources tokugawa-desktop-platforms:go1.24.13-sdk15.5-noble-r1 sh -c '
    set -e
    for arch in amd64 arm64; do
      if test "$arch" = amd64; then cc=x86_64-apple-darwin24.5-clang; else cc=aarch64-apple-darwin24.5-clang; fi
      "$cc" -O2 -dynamiclib -fvisibility=hidden -install_name @rpath/libzopl.dylib -o /build/mac-$arch/libzopl.dylib third_party/nuked-opl3/opl3.c third_party/nuked-opl3/bridge.c
      GOOS=darwin GOARCH=$arch CGO_ENABLED=1 CC="$cc" go build -trimpath -buildvcs=false -ldflags "-X main.version=$BUILD_VERSION" -o /build/mac-$arch/zeliard-preview ./cmd/zeliard-preview > /build/mac-$arch/build.log 2>&1
    done
    lipo -create /build/mac-amd64/zeliard-preview /build/mac-arm64/zeliard-preview -output /build/macos/zeliard-preview
    lipo -create /build/mac-amd64/libzopl.dylib /build/mac-arm64/libzopl.dylib -output /build/macos/libzopl.dylib'
fi
if [[ $phase == all || $phase == package ]]; then
  timeout 300 docker run "${common[@]}" -v "$project:/project:ro" -v "$work:$work:rw" -v "$project/dist-all:/project/dist-all:rw" -e "ZELIARD_RELEASE_WORK=$work" eob-audio-capture:20261009-r3 python3 /project/tools/package_dual_release.py assemble
  [[ -d "$project/dist-all/$version/full-local" ]]
  for edition in CHT CHT-HD; do
    timeout 180 docker run "${common[@]}" -e ARCH=x86_64 -e "BUILD_VERSION=$version" -e "BUILD_EDITION=$edition" -v "$work:/build:ro" -v "$project/dist-all/$version/full-local:/delivery:rw" fd2-build-appimage:latest sh -c '
      set -e
      cd /tmp
      /opt/appimage-tools/appimagetool.AppImage --appimage-extract > /tmp/extract.log
      squashfs-root/AppRun --runtime-file /opt/appimage-tools/runtime-x86_64 /build/AppDir-$BUILD_EDITION /delivery/Zeliard-$BUILD_VERSION-$BUILD_EDITION-Linux-x86_64.AppImage'
  done
fi

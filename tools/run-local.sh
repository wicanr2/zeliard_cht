#!/usr/bin/env bash
set -euo pipefail
bundle=${ZELIARD_BUNDLE:-$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)}
user_dir=${ZELIARD_USER_DIR:-$bundle/userdata}
if [[ ! -f "$bundle/zeliard-preview" || ! -d "$bundle/original" ]]; then
  echo '找不到本機遊戲包，請先依 README 建置。' >&2
  exit 1
fi
if [[ -z ${DISPLAY:-} ]]; then
  echo '請在有桌面顯示的終端機執行本遊戲。' >&2
  exit 1
fi
mkdir -p "$user_dir"
options=(--rm --init --network none --user "$(id -u):$(id -g)" --memory 2g --cpus 2 --pids-limit 256
  --mount "type=bind,src=$bundle,dst=/gameapp,readonly"
  --mount "type=bind,src=$user_dir,dst=/user"
  --mount type=bind,src=/tmp/.X11-unix,dst=/tmp/.X11-unix,readonly
  -e "DISPLAY=$DISPLAY" -e LIBGL_ALWAYS_SOFTWARE=1 -e GOMAXPROCS=2 -e LP_NUM_THREADS=2)
authority=${XAUTHORITY:-${HOME}/.Xauthority}
if [[ -f "$authority" ]]; then
  options+=(--mount "type=bind,src=$authority,dst=/xauthority,readonly" -e XAUTHORITY=/xauthority)
fi
if [[ -d /dev/input ]]; then
  options+=(--mount type=bind,src=/dev/input,dst=/dev/input,readonly --device-cgroup-rule 'c 13:* r')
fi
if [[ -d /run/udev ]]; then
  options+=(--mount type=bind,src=/run/udev,dst=/run/udev,readonly)
fi
exec timeout --foreground 12h docker run "${options[@]}" --entrypoint python3 \
  eob-audio-capture:20261009-r3 /gameapp/start_game.py "$@"

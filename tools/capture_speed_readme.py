#!/usr/bin/env python3
"""Docker內擷取正式包的速度設定頁；不執行遊戲或量測效能。"""
from pathlib import Path
import hashlib
import json
import os
import shutil
import subprocess
import time

P, O = Path('/project'), Path('/output')
VERSION = (P / 'VERSION').read_text().strip()
PACKAGE = P / 'dist-all' / VERSION / 'full-local' / f'Zeliard-{VERSION}-CHT-HD-Linux-x86_64.AppImage'
assert O.is_dir() and O.stat().st_uid == os.getuid() and not any(O.iterdir())
env = {**os.environ, 'DISPLAY': ':97', 'LIBGL_ALWAYS_SOFTWARE': '1',
       'GOMAXPROCS': '1', 'LP_NUM_THREADS': '1', 'ZELIARD_USER_DIR': str(O / 'user')}
children = []


def sha(file):
    with file.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def call(command):
    return subprocess.run(command, env=env, check=True, capture_output=True, timeout=10)


def stop(child):
    if child.poll() is None:
        child.terminate()
        try:
            child.wait(timeout=5)
        except subprocess.TimeoutExpired:
            child.kill()
            child.wait(timeout=5)


try:
    with (O / 'extract.log').open('wb') as log:
        subprocess.run([str(PACKAGE), '--appimage-extract'], cwd=O, check=True,
                       stdout=log, stderr=subprocess.STDOUT, timeout=60)
    root = O / 'squashfs-root'
    binary_sha = sha(root / 'zeliard-preview')
    x = subprocess.Popen(['Xvfb', ':97', '-screen', '0', '1280x800x24', '-nolisten', 'tcp'],
                         env=env, stdout=(O / 'xvfb.log').open('wb'), stderr=subprocess.STDOUT)
    children.append(x)
    deadline = time.monotonic() + 10
    while not Path('/tmp/.X11-unix/X97').exists():
        assert time.monotonic() < deadline
        time.sleep(.05)
    command = [str(root / 'AppRun'), '-show-controller', '-audio', 'off',
               '-observe', str(O / 'observe.json'), '-capture', str(O / 'capture.png'),
               '-capture-request', str(O / 'capture.request')]
    game = subprocess.Popen(command, cwd=O, env=env,
                            stdout=(O / 'game.log').open('wb'), stderr=subprocess.STDOUT)
    children.append(game)
    deadline = time.monotonic() + 90
    window = None
    while time.monotonic() < deadline:
        assert game.poll() is None, (O / 'game.log').read_text(errors='replace')[-2000:]
        found = subprocess.run(['xdotool', 'search', '--onlyvisible', '--pid', str(game.pid)],
                               env=env, capture_output=True, timeout=3)
        if found.returncode == 0 and found.stdout.strip():
            window = found.stdout.decode().splitlines()[-1]
            break
        time.sleep(.1)
    assert window is not None, '視窗未出現'
    call(['xdotool', 'windowfocus', window])

    def state():
        assert game.poll() is None
        try:
            return json.loads((O / 'observe.json').read_text())
        except (FileNotFoundError, json.JSONDecodeError):
            return {}

    def wait(condition):
        end = time.monotonic() + 20
        while time.monotonic() < end:
            s = state()
            if s and condition(s):
                return s
            time.sleep(.05)
        raise TimeoutError('等待設定頁更新')

    def click(x, y):
        call(['xdotool', 'mousemove', '--window', window, str(x), str(y), 'click', '1'])

    initial = wait(lambda s: s['focused'] and s['panel'])
    click(668, 24)
    wait(lambda s: s['paused'])
    click(332, 24)
    wait(lambda s: s['speed_panel'])
    click(480, 341)  # 8x，經實際工具列選取並保存。
    wait(lambda s: not s['panel'] and s['cpu_multiplier'] == 8)
    click(332, 24)
    selected = wait(lambda s: s['speed_panel'] and s['cpu_multiplier'] == 8)
    call(['xdotool', 'mousemove', '--window', window, '930', '640'])
    selected = wait(lambda s: s['speed_panel'] and s['frames'] > selected['frames'] + 2)
    assert selected['steps'] == initial['steps'] and selected['cycles'] == initial['cycles']
    image = O / 'cpu-speed-settings.png'
    call(['import', '-window', window, str(image)])
    (O / 'selected-state.json').write_text(json.dumps(selected, ensure_ascii=False, indent=2) + '\n')
    (O / 'capture.request').write_text('capture\n')
    game.wait(timeout=20)
    assert game.returncode == 0
    manifest = {
        'version': VERSION, 'scope': 'README實際設定頁截圖；原版未推進，不作遊玩或HD幀率證據',
        'package': str(PACKAGE.relative_to(P)), 'package_sha256': sha(PACKAGE),
        'binary_sha256': binary_sha, 'tool': 'tools/capture_speed_readme.py',
        'tool_sha256': sha(Path(__file__)), 'image': 'cpu-speed-settings.png',
        'image_sha256': sha(image), 'width': 960, 'height': 680,
        'selected_ceiling': 8, 'actual_multiplier': selected['cpu_effective_multiplier'],
        'original_steps_unchanged': True, 'original_cycles_unchanged': True,
        'audio': 'off', 'performance_acceptance': 'NOT_MEASURED',
        'source_receipt': 'workplace/release-v103/readme-speed/selected-state.json',
        'rights': 'user-authorized README UI preview; full original data remains local',
    }
    (O / 'cpu-speed-provenance.json').write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + '\n')
    print('CAPTURED_SPEED_SETTINGS_UI_ONLY', flush=True)
finally:
    for child in reversed(children):
        stop(child)
    # 只移除本次可重生解壓層，正式包與原始截圖、設定及收據保留。
    if (O / 'squashfs-root').exists():
        shutil.rmtree(O / 'squashfs-root')

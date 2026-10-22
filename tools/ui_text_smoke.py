"""064：Docker／Xvfb限定設定頁與字級冷重開，不推進DOS或測試幀率。"""
from pathlib import Path
import hashlib
import json
import os
import signal
import subprocess
import time

P = Path('/project')
W = P / 'workplace/ui-text-v1'
B = W / 'preview'
O = W / 'smoke'
assert W.stat().st_uid == os.getuid() and not O.exists()
O.mkdir()
env = {**os.environ, 'DISPLAY': ':96', 'GOMAXPROCS': '1', 'LP_NUM_THREADS': '1', 'LD_LIBRARY_PATH': str(B / 'lib'), 'ZELIARD_USER_DIR': str(O / 'user')}
children, shots, boots = [], [], []
process = window = directory = None


def invoke(args):
    return subprocess.run(args, env=env, check=True, capture_output=True, timeout=8)


def state():
    assert process.poll() is None, '候選提早結束'
    try:
        return json.loads((directory / 'observe.json').read_text())
    except (FileNotFoundError, json.JSONDecodeError):
        return None


def wait(label, check, seconds=20):
    end = time.monotonic() + seconds
    while time.monotonic() < end:
        s = state()
        if s:
            assert s['steps'] == 0 and s['cycles'] == 0, '設定操作推進DOS'
            assert not s['held_keys'], '設定鍵洩漏至DOS'
        if s and check(s):
            return s
        time.sleep(.04)
    raise TimeoutError(label)


def click(x, y):
    invoke(['xdotool', 'mousemove', '--window', window, str(x), str(y), 'click', '1'])


def press(name, check):
    invoke(['xdotool', 'keydown', name])
    try:
        wait('按鍵已處理：'+name, check)
    finally:
        invoke(['xdotool', 'keyup', name])
    initial = state()['frames']
    wait('按鍵釋放後更新', lambda v: v['frames'] >= initial+3)


def shot(name, extra=None):
    initial = state()
    s = wait('設定畫面完成', lambda v: v['frames'] > initial['frames'])
    assert not s['ui_text_overflow'], (name, s['ui_text_overflow'])
    if extra:
        extra(s)
    path = directory / (name + '.png')
    invoke(['ffmpeg', '-nostdin', '-hide_banner', '-loglevel', 'error', '-f', 'x11grab', '-video_size', '960x680', '-draw_mouse', '0', '-i', ':96.0+0,0', '-frames:v', '1', '-threads', '1', str(path)])
    (directory / (name + '.json')).write_text(json.dumps(s, ensure_ascii=False, indent=2) + '\n')
    shots.append({'name': name, 'size': s['ui_text_size'], 'path': str(path.relative_to(W)), 'sha256': hashlib.sha256(path.read_bytes()).hexdigest()})
    print('SHOT', name, s['ui_text_size'], flush=True)


def choose(size, keyboard=False):
    click(907, 24)
    s = wait('字級頁', lambda v: v['ui_text_panel'])
    i = [16, 24, 32].index(size)
    if keyboard:
        current = [16, 24, 32].index(s['ui_text_size'])
        for _ in range((i-current) % 3):
            current = (current+1) % 3
            press('Down', lambda v: v['ui_text_focus'] == current)
        press('Return', lambda v: v['ui_text_size'] == size)
    else:
        click(480, 259+i*82)
    wait('字級立即套用', lambda v: v['ui_text_panel'] and v['ui_text_size'] == size)
    saved = json.loads((O / 'user/controller.json').read_text())
    assert saved['ui_text_size'] == size


def start():
    global directory, process, window
    directory = O / ('boot' + str(len(boots)+1))
    directory.mkdir()
    command = ['/opt/VirtualGL/bin/vglrun', '-d', 'egl', str(B / 'zeliard-preview'), '-show-controller', '-audio', 'off', '-user-dir', str(O / 'user'), '-observe', str(directory / 'observe.json'), '-capture', str(directory / 'capture.png'), '-capture-request', str(directory / 'capture.request')]
    process = subprocess.Popen(command, env=env, stdout=(directory / 'game.log').open('wb'), stderr=subprocess.STDOUT)
    children.append(process)
    end = time.monotonic() + 45
    while time.monotonic() < end:
        assert process.poll() is None
        result = subprocess.run(['xdotool', 'search', '--onlyvisible', '--pid', str(process.pid)], env=env, capture_output=True, timeout=3)
        if result.returncode == 0 and result.stdout.strip():
            window = result.stdout.decode().splitlines()[-1]
            break
        time.sleep(.1)
    else:
        raise TimeoutError('候選視窗')
    invoke(['xdotool', 'windowmove', window, '0', '0'])
    invoke(['xdotool', 'windowfocus', window])
    wait('設定就緒', lambda v: v['panel'] and v['focused'], 40)
    boots.append({'command': command})


def finish():
    (directory / 'capture.request').touch()
    assert process.wait(timeout=15) == 0


def stop(child):
    if child.poll() is None:
        child.terminate()
        try:
            child.wait(timeout=5)
        except subprocess.TimeoutExpired:
            child.kill()
            child.wait(timeout=5)


signal.signal(signal.SIGTERM, lambda *_: (_ for _ in ()).throw(TimeoutError('有界字級檢查')))
try:
    x = subprocess.Popen(['Xvfb', ':96', '-screen', '0', '1280x800x24', '-nolisten', 'tcp'], env=env, stdout=(O / 'xvfb.log').open('wb'), stderr=subprocess.STDOUT)
    children.append(x)
    end = time.monotonic() + 10
    while not Path('/tmp/.X11-unix/X96').exists():
        assert time.monotonic() < end
        time.sleep(.05)
    renderer = invoke(['/opt/VirtualGL/bin/vglrun', '-d', 'egl', '/opt/VirtualGL/bin/glxinfo', '-B']).stdout.decode()
    assert 'direct rendering: Yes' in renderer and 'llvmpipe' not in renderer.lower()
    (O / 'renderer.txt').write_text(renderer)
    start()
    assert state()['ui_text_size'] == 24
    shot('controller-24')
    choose(32, keyboard=True)
    shot('text-size-32')
    for x, panel in [(568, 'cheat'), (330, 'speed'), (800, 'keyboard'), (440, 'controller')]:
        click(x, 24)
        wait(panel, lambda v: v['panel'] and not v['ui_text_panel'])
        shot(panel+'-32')
    # 檢查鍵盤重綁等待訊息仍完整，Esc取消不送入DOS。
    click(800, 24)
    wait('鍵盤頁', lambda v: v['keyboard_panel'])
    press('Return', lambda v: bool(v['ui_listening']))
    shot('keyboard-listening-32')
    press('Escape', lambda v: not v['ui_listening'] and v['panel'])
    for size in (16, 24):
        choose(size)
        click(568, 24)
        wait('作弊頁', lambda v: v['cheat_panel'])
        shot('cheat-'+str(size))
    choose(32)
    shot('saved-32')
    finish()
    original_save = O / 'user/game/SAVETEST.usr'
    before = (hashlib.sha256(original_save.read_bytes()).hexdigest(), original_save.stat().st_mtime_ns)
    start()
    assert state()['ui_text_size'] == 32
    shot('cold-restart-32')
    click(568, 24)
    wait('冷重開作弊頁', lambda v: v['cheat_panel'])
    assert not state()['cheats']['lock_hp'] and not state()['cheats']['lock_mp']
    shot('cold-cheat-32')
    finish()
    after = (hashlib.sha256(original_save.read_bytes()).hexdigest(), original_save.stat().st_mtime_ns)
    assert before == after
    receipt = {'status': 'PASS_SCOPED_UI_TEXT_SETTINGS', 'binary_sha256': hashlib.sha256((B / 'zeliard-preview').read_bytes()).hexdigest(), 'script_sha256': hashlib.sha256(Path(__file__).read_bytes()).hexdigest(), 'boots': boots, 'shots': shots, 'sizes': [16,24,32], 'default': 24, 'cold_restart': 32, 'dos_steps': 0, 'dos_cycles': 0, 'game_save_unchanged': True, 'renderer': renderer, 'not_checked': ['gameplay', 'FPS', 'Windows/macOS real hardware', 'physical gamepad']}
    (O / 'receipt.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+'\n')
    print(receipt['status'], flush=True)
finally:
    for child in reversed(children):
        stop(child)

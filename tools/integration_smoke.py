"""Docker／Xvfb 內的第一洞窟代表性抽測，不作完整通關驗收。"""
from pathlib import Path
import base64
import hashlib
import json
import os
import shutil
import signal
import subprocess
import time

R = Path('/project')
B = Path('/gameapp')
O = Path('/output')
assert not any(O.iterdir()) and O.stat().st_uid == 1000
h = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
env = {**os.environ, 'DISPLAY': ':96', 'LIBGL_ALWAYS_SOFTWARE': '1', 'GOMAXPROCS': '2', 'LP_NUM_THREADS': '2'}
children, shots, boots = [], [], []
process = None
window = None
directory = None
names = {0x39: 'z', 0x41: 'F7', 0x15: 'y', 0xe: 'BackSpace', 0x1f: 's', 0x1e: 'a', 0x2f: 'v', 0x12: 'e', 0x14: 't', 0x1c: 'Return', 0x4b: 'Left', 0x4d: 'Right', 0x48: 'Up', 0x1d: 'Control_L', 0x10: 'q'}

def invoke(args):
    return subprocess.run(args, env=env, check=True, capture_output=True, timeout=8)

def state():
    assert process.poll() is None, '遊戲提早結束'
    try:
        return json.loads((directory / 'observe.json').read_text())
    except (FileNotFoundError, json.JSONDecodeError):
        return None

def held(s):
    raw = s['held_keys']
    return set(base64.b64decode(raw) if isinstance(raw, str) else raw or [])

def wait(label, condition, seconds=90):
    end = time.monotonic() + seconds
    while time.monotonic() < end:
        s = state()
        if s and condition(s):
            return s
        time.sleep(.02)
    raise TimeoutError(label)

def click(x, y):
    invoke(['xdotool', 'mousemove', '--window', window, str(x), str(y), 'click', '1'])

def advance(n):
    s = state()
    assert s['focused'] and not s['panel'] and not s['paused'] and not s['input_gate']
    return wait('原版執行', lambda v: v['steps'] >= s['steps'] + n, 150)

def edge(scan, pressed):
    invoke(['xdotool', 'keydown' if pressed else 'keyup', names[scan]])
    return wait('按鍵送入原版', lambda s: (scan in held(s)) == pressed, 15)

def key(scan, n):
    edge(scan, True)
    try:
        advance(n)
    finally:
        edge(scan, False)

def shot(name):
    assert not held(state()) and not state()['paused']
    click(644, 24)
    a = wait('暫停', lambda s: s['paused'] and not held(s), 15)
    s = wait('暫停畫面完成', lambda s: s['paused'] and s['frames'] > a['frames'], 15)
    assert s['steps'] == a['steps'] and s['cycles'] == a['cycles']
    path = directory / (name + '.png')
    invoke(['import', '-window', window, str(path)])
    (directory / (name + '.json')).write_text(json.dumps(s, ensure_ascii=False, indent=2) + '\n')
    shots.append({'name': name, 'boot': len(boots), 'image_sha256': h(path), 'state': s})
    click(644, 24)
    wait('恢復前端', lambda s: not s['paused'], 15)
    print('SHOT', name, s['steps'], flush=True)

def start():
    global process, window, directory
    directory = O / ('boot' + str(len(boots) + 1))
    directory.mkdir()
    command = [str(B / 'zeliard-preview'), '-orig', str(B / 'original'), '-font', str(B / 'font.hex.gz'), '-user-dir', str(O / 'user'), '-observe', str(directory / 'observe.json'), '-input-trace', str(directory / 'inputs.jsonl'), '-capture', str(directory / 'capture.png'), '-capture-request', str(directory / 'capture.request')]
    for option, relative in json.loads((B / 'runtime-args.json').read_text()):
        command += [option, str(B / relative)]
    log = (directory / 'game.log').open('wb')
    process = subprocess.Popen(command, env=env, stdout=log, stderr=log)
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
        raise TimeoutError('遊戲視窗')
    invoke(['xdotool', 'windowfocus', window])
    wait('視窗就緒', lambda s: s['focused'] and not s['input_gate'], 30)
    boots.append({'command': command})

def panel(keyboard):
    click(840 if keyboard else 440, 24)
    wait('中文設定頁', lambda s: s['panel'] and s['keyboard_panel'] == keyboard, 15)

def close_panel():
    invoke(['xdotool', 'keydown', 'Escape'])
    s = wait('關閉設定', lambda s: not s['panel'] and s['input_gate'], 15)
    time.sleep(.12)
    assert not held(state()) and state()['steps'] == s['steps'], '關閉鍵洩漏至原版'
    invoke(['xdotool', 'keyup', 'Escape'])
    wait('設定鍵已放開', lambda s: not s['input_gate'], 15)

def align(name, box, offset):
    from PIL import Image
    S = R / 'workplace/prose-v1/services-v1/source-v1'
    raw, pal = (S / (name + '.indexed')).read_bytes(), (S / (name + '.palette')).read_bytes()
    expected = bytes(pal[raw[y * 320 + x] * 3 + c] for y in range(box[1], box[3]) for x in range(box[0], box[2]) for c in range(3))
    for attempt in range(20):
        path = directory / 'navigation.png'
        invoke(['import', '-window', window, str(path)])
        im = Image.open(path).convert('RGB')
        found = []
        for left in range(56, 249):
            pixels = bytes(im.getpixel((x * 3, y * 3 + 48))[c] for y in range(box[1], box[3]) for x in range(left, left + box[2] - box[0]) for c in range(3))
            if pixels == expected:
                found.append(left)
        if not found and name.startswith('cave'):
            key(0x4d, 500000)
            advance(300000)
            continue
        assert len(found) == 1, ('入口標記不唯一', name, found)
        player = 48 + 8 * state()['native_primary_0083']
        delta = found[0] + offset - player
        if delta == 0:
            return
        key(0x4b if delta < 0 else 0x4d, 150000)
        advance(350000)
    raise RuntimeError('入口未對準')

def hd(enabled):
    if state()['hd_enabled'] != enabled:
        click(240 if enabled else 80, 24)
        wait('畫面切換', lambda s: s['hd_enabled'] == enabled, 15)

def stop(p):
    if p.poll() is None:
        p.terminate()
        try:
            p.wait(timeout=5)
        except subprocess.TimeoutExpired:
            p.kill()
            p.wait(timeout=5)

signal.signal(signal.SIGTERM, lambda *_: (_ for _ in ()).throw(TimeoutError('有界視窗抽測')))
try:
    x = subprocess.Popen(['Xvfb', ':96', '-screen', '0', '1280x800x24', '-nolisten', 'tcp'], env=env, stdout=(O / 'xvfb.log').open('wb'), stderr=subprocess.STDOUT)
    children.append(x)
    end = time.monotonic() + 10
    while not Path('/tmp/.X11-unix/X96').exists():
        assert time.monotonic() < end
        time.sleep(.05)
    assert invoke([str(B / 'zeliard-preview'), '-version']).stdout.decode().strip() == (B / 'VERSION').read_text().strip()
    user = O / 'user/game'
    user.mkdir(parents=True)
    shutil.copy2(B / 'userdata/game/SAVETEST.usr', user / 'SAVETEST.usr')
    start()
    panel(False)
    shot('controller-settings')
    close_panel()
    panel(True)
    baseline = state()['steps']
    click(680, 284)
    invoke(['xdotool', 'key', 'z'])
    wait('鍵盤綁定保存', lambda s: s['keyboard_bindings']['attack'] == 'Z', 15)
    assert state()['steps'] == baseline and not held(state())
    shot('keyboard-remapped')
    click(540, 580)
    wait('解除綁定', lambda s: s['keyboard_bindings']['attack'] == '', 15)
    click(210, 580)
    wait('還原預設', lambda s: s['keyboard_bindings']['attack'] == 'Space', 15)
    click(680, 284)
    invoke(['xdotool', 'key', 'z'])
    wait('重綁保存', lambda s: s['keyboard_bindings']['attack'] == 'Z', 15)
    close_panel()
    advance(12000000)
    for _ in range(6):
        key(0x39, 50000)
        advance(7950000)
    key(0x41, 500000)
    advance(2000000)
    key(0x15, 150000)
    advance(1500000)
    for _ in range(8):
        key(0xe, 100000)
        advance(200000)
    for scan in [0x1f, 0x1e, 0x2f, 0x12, 0x14, 0x12, 0x1f, 0x14]:
        key(scan, 100000)
        advance(200000)
    key(0x1c, 100000)
    advance(16000000)
    shot('restored-town')
    key(0x4b, 27000000)
    advance(500000)
    align('church-door', (160, 102, 176, 118), 0)
    key(0x48, 500000)
    wait('中文對話', lambda s: s['dialogue']['key'] == 'zelres2.file.01a508' and s['dialogue']['prompt'])
    hd(True)
    wait('教堂高清', lambda s: s['church_hd_matched'], 20)
    shot('church-chinese-HD')
    key(0x39, 500000)
    for attempt in range(3):
        wait('祝福完成', lambda s: s['dialogue']['key'] == 'zelres2.file.01a568' and s['dialogue']['complete'] and s['text_cursor'] == 0xa3e2 and s['advance_input_flags'] == 0)
        advance(500000)
        key(0x39, 500000)
        advance(6000000)
        if not state()['dialogue']['active']:
            break
    else:
        raise RuntimeError('教堂續頁鍵未完成離店')
    wait('城鎮高清', lambda s: s['town_scenery_hd_matched'] and s['town_ground_hd_matched'], 30)
    shot('town-HD')
    hd(False)
    key(0x4d, 36000000)
    advance(500000)
    align('cave-for-wound-door', (216, 102, 240, 118), 16)
    key(0x48, 500000)
    advance(16000000)
    hd(True)
    wait('第一洞窟高清', lambda s: s['cave_terrain_hd_matched'], 30)
    shot('first-cave-HD')
    key(0x39, 1500000)
    key(0x4d, 1000000)
    key(0x48, 500000)
    advance(2500000)
    wait('洞窟操作完成', lambda s: s['cave_terrain_hd_matched'], 30)
    shot('first-cave-played')
    key(0x1c, 500000)
    advance(500000)
    shot('chinese-inventory')
    key(0x1c, 500000)
    advance(1000000)
    edge(0x1d, True)
    key(0x10, 500000)
    edge(0x1d, False)
    advance(2000000)
    invoke(['xdotool', 'keydown', 'y'])
    try:
        process.wait(timeout=60)
    finally:
        invoke(['xdotool', 'keyup', 'y'])
    assert process.returncode == 0
    boots[-1]['normal_DOS_exit'] = True
    boots[-1]['exit_code'] = process.returncode
    boots[-1]['input_trace_sha256'] = h(directory / 'inputs.jsonl')
    start()
    panel(True)
    assert state()['keyboard_bindings']['attack'] == 'Z'
    shot('keyboard-restored-after-restart')
    (directory / 'capture.request').write_text('capture\n')
    process.wait(timeout=30)
    assert process.returncode == 0
    boots[-1]['capture_exit_code'] = 0
    assert h(user / 'SAVETEST.usr') == h(B / 'userdata/game/SAVETEST.usr')
    receipt = {'status': 'PASS_SCOPED_FIRST_CAVE_INPUT_CHINESE_HD', 'version': (B / 'VERSION').read_text().strip(), 'binary_sha256': h(B / 'zeliard-preview'), 'boots': boots, 'shots': shots, 'keyboard_rebind_unbind_reset_restart': True, 'settings_input_isolated': True, 'gamepad_hardware_tested': False, 'scope': '使用者限定第一洞窟及代表性中文／HD／工具列整合；未要求完整通關；實體手把未接入'}
    (O / 'receipt.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + '\n')
    print(receipt['status'], flush=True)
finally:
    for p in reversed(children):
        stop(p)

#!/usr/bin/env python3
"""057／058：Docker內正常冷F7、HD入洞過場及正常退出的限定視窗檢查。"""
import ast
from pathlib import Path

helper = Path('/project/tools/integration_smoke.py')
source = helper.read_text()
source = source.replace("    for option, relative in json.loads((B / 'runtime-args.json').read_text()):",
                        "    command += json.loads((B / 'bundle-options.json').read_text()) + ['-audio','off']\n    for option, relative in json.loads((B / 'runtime-args.json').read_text()):")
nodes = []
for node in ast.parse(source).body:
    if isinstance(node, ast.Try):
        break
    nodes.append(node)
exec(compile(ast.Module(body=nodes, type_ignores=[]), str(helper), 'exec'))
names[0x39] = 'space'
original_state = state
samples = []
record_begin = None


def state():
    s = original_state()
    if s is not None and record_begin is not None:
        if not samples or s['frames'] != samples[-1]['frames']:
            samples.append({'seconds': time.monotonic() - record_begin, 'frames': s['frames'],
                            'steps': s['steps'], 'cycles': s['cycles'], 'hd_enabled': s['hd_enabled'],
                            'transition': s.get('cave_transition_hd'), 'cave': s['cave_terrain_hd_matched'],
                            'town': s.get('town_frame_presentation'),
                            'memory_sha256': s['oracle_memory_sha256'], 'palette_sha256': s['oracle_palette_sha256']})
    return s


try:
    x = subprocess.Popen(['Xvfb', ':96', '-screen', '0', '1280x800x24', '-nolisten', 'tcp'], env=env,
                         stdout=(O / 'xvfb.log').open('wb'), stderr=subprocess.STDOUT)
    children.append(x)
    end = time.monotonic() + 10
    while not Path('/tmp/.X11-unix/X96').exists():
        assert time.monotonic() < end
        time.sleep(.05)
    save = R / 'workplace/orig/SAVEM1.USR'
    user = O / 'user/game'
    user.mkdir(parents=True)
    shutil.copy2(save, user / 'SAVETEST.usr')
    start()
    invoke(['xdotool', 'windowmove', window, '0', '0'])
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
    for scan in (0x1f, 0x1e, 0x2f, 0x12, 0x14, 0x12, 0x1f, 0x14):
        key(scan, 100000)
        advance(200000)
    key(0x1c, 100000)
    advance(16000000)
    hd(False)
    key(0x4d, 8000000)
    advance(250000)
    align('cave-for-wound-door', (216, 102, 240, 118), 16)
    hd(True)
    wait('入洞前HD城鎮', lambda s: s['town_scenery_hd_matched'] and s['town_ground_hd_matched'], 20)
    shot('before-entry-HD')
    record_begin = time.monotonic()
    video = subprocess.Popen(['ffmpeg', '-nostdin', '-hide_banner', '-loglevel', 'warning', '-y',
                              '-f', 'x11grab', '-video_size', '960x680', '-framerate', '30', '-draw_mouse', '0',
                              '-i', ':96.0+0,0', '-an', '-c:v', 'libx264', '-threads', '1', '-preset', 'veryfast',
                              '-crf', '20', '-pix_fmt', 'yuv420p', '-movflags', '+faststart', str(O / 'transition-playthrough.mp4')],
                             env=env, stdout=subprocess.DEVNULL, stderr=(O / 'video.log').open('wb'))
    children.append(video)
    key(0x48, 500000)
    wait('正常過場進入洞窟HD', lambda s: s['cave_terrain_hd_matched'], 35)
    shot('after-entry-HD')
    key(0x39, 1500000)
    advance(1000000)
    stop(video)
    assert any((s.get('transition') or {}).get('cave_matched') for s in samples), '沒有觀察到HD過場'
    assert any((s.get('town') or {}).get('erased', 0) > 0 for s in samples), '沒有觀察到HD擦除'
    active = [s['seconds'] for s in samples if (s.get('transition') or {}).get('cave_matched')]
    assert all((s.get('transition') or {}).get('cave_matched') for s in samples if active[0] <= s['seconds'] <= active[-1]), '過場中途撤除整幅HD'
    assert all(s['hd_enabled'] for s in samples), 'HD模式被切換'
    edge(0x1d, True)
    key(0x10, 500000)
    edge(0x1d, False)
    advance(1000000)
    invoke(['xdotool', 'keydown', 'y'])
    try:
        process.wait(timeout=45)
    finally:
        invoke(['xdotool', 'keyup', 'y'])
    assert process.returncode == 0
    assert h(user / 'SAVETEST.usr') == h(save)
    assert (user / 'SAVETEST.usr').stat().st_mtime_ns == save.stat().st_mtime_ns
    report = {'status': 'PASS_HD_CAVE_TRANSITION_GUI', 'binary_sha256': h(B / 'zeliard-preview'),
              'version': (B / 'VERSION').read_text().strip(), 'source_save_sha256': h(save), 'save_unchanged': True,
              'normal_DOS_exit': 0, 'transition_hd_verified': True, 'town_wipe_hd_verified': True,
              'samples': samples, 'shots': shots, 'boots': boots,
              'video_sha256': h(O / 'transition-playthrough.mp4'), 'video_audio': 'none; visual regression clip only',
              'limits': '只驗正常冷F7、HD入洞過場與第一洞窟；首領為另行靜態素材接入核對，未逐一到達後段'}
    (O / 'receipt.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
    print(report['status'], flush=True)
except Exception:
    (O / 'failed-samples.json').write_text(json.dumps(samples, ensure_ascii=False, indent=2) + '\n')
    if process is not None and process.poll() is None:
        invoke(['import', '-window', window, str(O / 'failed-live.png')])
        (O / 'failed-state.json').write_text(json.dumps(state(), ensure_ascii=False, indent=2) + '\n')
    raise
finally:
    for child in reversed(children):
        stop(child)

"""低負載Docker：正常F7讀檔、居民對話與第一洞窟的限定操作。"""
from pathlib import Path
import json, os, signal, subprocess, time

# 沿用既有正常鍵盤、讀檔及視窗擷取方法，僅載入函式定義。
source = Path('/project/tools/integration_smoke.py').read_text()
source = source.replace("'-orig', str(B / 'original')", "'-audio', 'off', '-orig', str(B / 'original')")
source = source.replace("'-input-trace', str(directory / 'inputs.jsonl'), ", "")
gpu = os.environ.get('ZELIARD_HD_SMOKE_GPU') == '1'
if gpu:
    source = source.replace("'LIBGL_ALWAYS_SOFTWARE': '1', ", '')
    source = source.replace("command = [str(B / 'zeliard-preview')", "command = ['/opt/VirtualGL/bin/vglrun', '-d', 'egl', str(B / 'zeliard-preview')")
    source = source.replace('    from PIL import Image\n', '')
    source = source.replace("        im = Image.open(path).convert('RGB')", "        rgb_buffer = invoke(['ffmpeg','-nostdin','-loglevel','error','-i',str(path),'-threads','1','-f','rawvideo','-pix_fmt','rgb24','-']).stdout\n        assert len(rgb_buffer) == 960*680*3\n        im = type('Frame', (), {'getpixel': lambda self, xy: rgb_buffer[(xy[1]*960+xy[0])*3:(xy[1]*960+xy[0])*3+3]})()")
exec(compile(source.split('signal.signal(signal.SIGTERM,')[0], 'integration_helpers', 'exec'))
B = Path('/project/workplace/runtime-controls-v1/preview')
env.update(GOMAXPROCS='1', LP_NUM_THREADS='1')
names[0x39] = 'space'
signal.signal(signal.SIGTERM, lambda *_: (_ for _ in ()).throw(TimeoutError('限定HD抽測')))

def key(scan, n):
    edge(scan, True)
    target = state()['steps'] + n
    end = time.monotonic() + 150
    try:
        while state()['steps'] < target:
            assert time.monotonic() < end, '正常按鍵執行逾時'
            # 場景按需載入會釋放遊戲按鍵；先放開實體鍵，避免阻住輸入閘門。
            if state()['input_gate']:
                invoke(['xdotool', 'keyup', names[scan]])
            time.sleep(.04)
    finally:
        edge(scan, False)

def samples(name, seconds):
    rows = []
    end = time.monotonic() + seconds
    while time.monotonic() < end:
        s = state()
        rows.append({k:s[k] for k in ('frames','steps','town_dialogue','town_frame_presentation','cave_frame_presentation','cave_terrain_hd_matched','town_scenery_hd_matched','performance')})
        time.sleep(.045)
    (directory/(name+'.json')).write_text(json.dumps(rows,ensure_ascii=False,indent=2)+'\n')
    return rows

try:
    x = subprocess.Popen(['Xvfb', ':96', '-screen', '0', '1280x800x24', '-nolisten', 'tcp'], env=env, stdout=(O/'xvfb.log').open('wb'), stderr=subprocess.STDOUT)
    children.append(x)
    end = time.monotonic()+10
    while not Path('/tmp/.X11-unix/X96').exists():
        assert time.monotonic()<end
        time.sleep(.05)
    if gpu:
        renderer=invoke(['/opt/VirtualGL/bin/vglrun','-d','egl','/opt/VirtualGL/bin/glxinfo','-B']).stdout.decode()
        assert not any(x in renderer.lower() for x in ('llvmpipe','softpipe','swrast')),renderer
        (O/'renderer.txt').write_text(renderer)
    user=O/'user/game'; user.mkdir(parents=True)
    shutil.copy2(B/'starter/SAVETEST.usr',user/'SAVETEST.usr')
    start()
    hd(False)
    advance(12000000)
    for _ in range(6):
        key(0x39,50000); advance(7950000)
    key(0x41,500000); advance(2000000)
    key(0x15,150000); advance(1500000)
    for _ in range(8):
        key(0xe,100000); advance(200000)
    for scan in [0x1f,0x1e,0x2f,0x12,0x14,0x12,0x1f,0x14]:
        key(scan,100000); advance(200000)
    key(0x1c,100000); advance(16000000)
    hd(True)
    wait('城鎮按需載入',lambda s:s['town_scenery_hd_matched'] and not s['input_gate'],70)
    shot('town-HD')
    for _ in range(25):
        key(0x4b,250000); advance(250000)
        key(0x39,500000); advance(100000)
        if state()['town_dialogue']['key']:break
    wait('居民中文',lambda s:bool(s['town_dialogue']['key']),20)
    talk=samples('town-dialogue-samples',3)
    shot('town-dialogue-HD')
    assert any(r['town_scenery_hd_matched'] for r in talk),'對話時未使用HD背景'
    for _ in range(3):
        if not state()['town_dialogue']['key']:break
        key(0x39,500000); advance(100000)
    assert not state()['town_dialogue']['key'],'居民對話未關閉'
    key(0x4d,8000000); advance(500000)
    hd(False)
    align('cave-for-wound-door',(216,102,240,118),16)
    key(0x48,500000); advance(16000000)
    hd(True)
    wait('第一洞窟按需載入',lambda s:s['cave_terrain_hd_matched'] and not s['input_gate'],90)
    shot('first-cave-HD')
    standing=samples('cave-standing-samples',3)
    key(0x4b,2000000); key(0x4d,2000000)
    moving=samples('cave-moving-samples',3)
    shot('first-cave-moved-HD')
    assert any(r['cave_terrain_hd_matched'] for r in standing)
    assert any(r['cave_terrain_hd_matched'] for r in moving)
    (directory/'capture.request').write_text('capture\n'); process.wait(timeout=30)
    assert process.returncode==0
    receipt={'status':'PASS_SCOPED_NORMAL_DIALOGUE_AND_FIRST_CAVE','binary_sha256':h(B/'zeliard-preview'),'shots':shots,'original_save_unchanged':h(user/'SAVETEST.usr')==h(B/'starter/SAVETEST.usr'),'standing_samples':len(standing),'moving_samples':len(moving),'scope':'正常F7、自建存檔、居民與第一洞窟；非完整通關、非使用者精確站位或30fps保證'}
    (O/'receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
    print(receipt['status'],flush=True)
finally:
    if process is not None and process.poll() is None and window is not None:
        try:invoke(['import','-window',window,str(O/'last-window.png')])
        except Exception:pass
    for p in reversed(children):stop(p)

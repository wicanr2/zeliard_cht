#!/usr/bin/env python3
"""051：以既有正常玩家路徑重新擷取原貌中文／HD中文及操作頁。"""
from pathlib import Path
import hashlib
import json
import os
import subprocess
import time

P=Path('/project');C=Path('/recording');children=[]
def stop(process):
    if process.poll() is None:
        process.terminate()
        try:process.wait(timeout=8)
        except subprocess.TimeoutExpired:process.kill();process.wait(timeout=5)

try:
    C.mkdir(exist_ok=True)
    for name in ('runtime','pulse-state','pulse-config'):(C/name).mkdir(mode=0o700)
    server='unix:/recording/runtime/native'
    (C/'server.pa').write_text('load-module module-native-protocol-unix socket=/recording/runtime/native auth-anonymous=1 auth-cookie-enabled=0\nload-module module-null-sink sink_name=zeliard_release format=s16le rate=44100 channels=2\nset-default-sink zeliard_release\n')
    (C/'asound.conf').write_text('pcm.!default { type pulse server "'+server+'" device "zeliard_release" }\nctl.!default { type pulse server "'+server+'" }\n')
    (C/'client.conf').write_text('default-server = '+server+'\nautospawn = no\n')
    os.environ.update({'PULSE_SERVER':server,'PULSE_RUNTIME_PATH':str(C/'runtime'),'PULSE_STATE_PATH':str(C/'pulse-state'),'PULSE_CONFIG_PATH':str(C/'pulse-config'),'PULSE_CLIENTCONFIG':str(C/'client.conf'),'ALSA_CONFIG_PATH':str(C/'asound.conf'),'XDG_RUNTIME_DIR':str(C/'runtime')})
    pulse=subprocess.Popen(['pulseaudio','--daemonize=no','--exit-idle-time=-1','--use-pid-file=no','-n','-F',str(C/'server.pa')],stdout=(C/'pulse.log').open('wb'),stderr=subprocess.STDOUT);children.append(pulse)
    deadline=time.monotonic()+10
    while not (C/'runtime/native').exists():
        assert time.monotonic()<deadline and pulse.poll() is None
        time.sleep(.05)
    monitor=subprocess.Popen(['parec','--device=zeliard_release.monitor','--format=s16le','--rate=44100','--channels=2'],stdout=(C/'device.pcm').open('wb'),stderr=(C/'parec.log').open('wb'));children.append(monitor)
    # 擷取程式有自己的有界Xvfb生命週期；不從其他專案借用桌面或程序。
    source=(P/'tools/integration_smoke.py').read_text()
    source=source.replace("names = {0x39: 'z'", "names = {0x39: 'space'")
    source=source.replace('    from PIL import Image\n','')
    source=source.replace("        im = Image.open(path).convert('RGB')", "        rgb_buffer = invoke(['convert', str(path), '-depth', '8', 'rgb:-']).stdout\n        assert len(rgb_buffer) == 960 * 680 * 3\n        im = type('NativeFrame', (), {'getpixel': lambda self, xy: rgb_buffer[(xy[1] * 960 + xy[0]) * 3:(xy[1] * 960 + xy[0]) * 3 + 3]})()")
    source=source.replace("B / 'userdata/game/SAVETEST.usr'","B / 'starter/SAVETEST.usr'")
    source=source.replace("    for option, relative in json.loads((B / 'runtime-args.json').read_text()):","    command += json.loads((B / 'bundle-options.json').read_text())\n    command += ['-audio-wav', str(directory / 'source.wav')]\n    for option, relative in json.loads((B / 'runtime-args.json').read_text()):")
    source=source.replace("    start()\n    panel(False)","""    start()
    invoke(['xdotool','windowmove',window,'0','0'])
    record_begin=time.monotonic()
    recording=subprocess.Popen(['ffmpeg','-nostdin','-hide_banner','-loglevel','warning','-f','x11grab','-video_size','960x680','-framerate','25','-i',':96.0+0,0','-an','-c:v','libx264','-preset','veryfast','-crf','20','-threads','2','-pix_fmt','yuv420p','/recording/gameplay.mp4'],env=env,stdout=subprocess.DEVNULL,stderr=open('/recording/ffmpeg.log','wb'))
    children.append(recording)
    panel(False)""",1)
    source=source.replace("shots.append({'name': name,", "shots.append({'wall_seconds':time.monotonic()-record_begin,'name': name,")
    source=source.replace("    shot('restored-town')", "    hd(False)\n    shot('restored-town')")
    source=source.replace("    shot('church-chinese-HD')", "    shot('church-chinese-HD')\n    hd(False)\n    shot('church-chinese-original')\n    hd(True)")
    # 空白鍵恢復預設，讓既有遊玩鍵序及影片的遊戲操作保持一致。
    source=source.replace("    close_panel()\n    advance(12000000)", "    click(210,580)\n    wait('錄影使用預設空白鍵', lambda s: s['keyboard_bindings']['attack']=='Space',15)\n    close_panel()\n    advance(12000000)")
    source=source.replace("assert state()['keyboard_bindings']['attack'] == 'Z'", "assert state()['keyboard_bindings']['attack'] == 'Space'")
    source=source.replace("    shot('town-HD')", "    shot('town-HD')\n    hd(False)\n    shot('town-original')")
    source=source.replace("    shot('first-cave-HD')", "    shot('first-cave-HD')\n    hd(False)\n    shot('first-cave-original')\n    hd(True)")
    C.mkdir(exist_ok=True)
    (C/'executed-driver.py').write_text(source)
    exec(compile(source,str(P/'tools/integration_smoke.py'),'exec'),{'__name__':'__main__'})
    (C/'provenance.json').write_text(json.dumps({'driver_sha256':hashlib.sha256(source.encode()).hexdigest(),'source':'new 051 native X11 game window recording; complete HD assets and rounded Chinese; normal display switches, not same-state parity evidence','audio':'source WAV from original MSCADLIB + Nuked-OPL3; final promo uses traced original title cue','private_local_only':True},ensure_ascii=False,indent=2)+'\n')
finally:
    for process in reversed(children):stop(process)

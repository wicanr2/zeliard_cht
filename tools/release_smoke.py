"""在Docker內執行既有玩家抽測，同時錄下真實視窗與虛擬音效裝置。"""
from pathlib import Path
import json,os,signal,subprocess,time,hashlib

R=Path('/project'); C=Path('/recording');C.mkdir(exist_ok=True)
children=[]
def stop(p):
    if p.poll() is None:
        p.terminate()
        try:p.wait(timeout=8)
        except subprocess.TimeoutExpired:p.kill();p.wait(timeout=5)

for n in ['runtime','pulse-state','pulse-config']:(C/n).mkdir(mode=0o700)
server='unix:/recording/runtime/native'
(C/'server.pa').write_text('load-module module-native-protocol-unix socket=/recording/runtime/native auth-anonymous=1 auth-cookie-enabled=0\nload-module module-null-sink sink_name=zeliard_release format=s16le rate=44100 channels=2\nset-default-sink zeliard_release\n')
(C/'asound.conf').write_text('pcm.!default { type pulse server "'+server+'" device "zeliard_release" }\nctl.!default { type pulse server "'+server+'" }\n')
(C/'client.conf').write_text('default-server = '+server+'\nautospawn = no\n')
os.environ.update({'PULSE_SERVER':server,'PULSE_RUNTIME_PATH':str(C/'runtime'),'PULSE_STATE_PATH':str(C/'pulse-state'),'PULSE_CONFIG_PATH':str(C/'pulse-config'),'PULSE_CLIENTCONFIG':str(C/'client.conf'),'ALSA_CONFIG_PATH':str(C/'asound.conf'),'XDG_RUNTIME_DIR':str(C/'runtime')})

try:
    pulse=subprocess.Popen(['pulseaudio','--daemonize=no','--exit-idle-time=-1','--use-pid-file=no','-n','-F',str(C/'server.pa')],stdout=(C/'pulse.log').open('wb'),stderr=subprocess.STDOUT);children.append(pulse)
    deadline=time.monotonic()+10
    while not (C/'runtime/native').exists():
        assert time.monotonic()<deadline and pulse.poll() is None
        time.sleep(.05)
    monitor=subprocess.Popen(['parec','--device=zeliard_release.monitor','--format=s16le','--rate=44100','--channels=2'],stdout=(C/'device.pcm').open('wb'),stderr=(C/'parec.log').open('wb'));children.append(monitor)
    source=(R/'tools/integration_smoke.py').read_text()
    # 錄影映像沿用既有ImageMagick，不另外混入另一套Python runtime。
    source=source.replace('    from PIL import Image\n','')
    source=source.replace("        im = Image.open(path).convert('RGB')", "        rgb_buffer = invoke(['convert', str(path), '-depth', '8', 'rgb:-']).stdout\n        assert len(rgb_buffer) == 960 * 680 * 3\n        im = type('NativeFrame', (), {'getpixel': lambda self, xy: rgb_buffer[(xy[1] * 960 + xy[0]) * 3:(xy[1] * 960 + xy[0]) * 3 + 3]})()")
    source=source.replace("    for option, relative in json.loads((B / 'runtime-args.json').read_text()):", "    command += ['-audio', 'ebiten', '-adlib', '-realtime', '-audio-wav', str(directory / 'source.wav')]\n    for option, relative in json.loads((B / 'runtime-args.json').read_text()):")
    source=source.replace("    start()\n    panel(False)","""    start()
    invoke(['xdotool','windowmove',window,'0','0'])
    record_begin=time.monotonic()
    recording=subprocess.Popen(['ffmpeg','-nostdin','-hide_banner','-loglevel','warning','-f','x11grab','-video_size','960x680','-framerate','25','-i',':96.0+0,0','-an','-c:v','libx264','-preset','veryfast','-crf','20','-threads','2','-pix_fmt','yuv420p','/recording/gameplay.mp4'],env=env,stdout=subprocess.DEVNULL,stderr=open('/recording/ffmpeg.log','wb'))
    children.append(recording)
    panel(False)""",1)
    source=source.replace("shots.append({'name': name,", "shots.append({'wall_seconds':time.monotonic()-record_begin,'name': name,")
    # 圖片與按鍵日誌仍由原來的抽測腳本獨立建立，不以影片代替通過條件。
    (C/'executed-driver.py').write_text(source)
    exec(compile(source,str(R/'tools/integration_smoke.py'),'exec'),{'__name__':'__main__'})
    (C/'provenance.json').write_text(json.dumps({'driver_sha256':hashlib.sha256(source.encode()).hexdigest(),'source':'actual native game X11 window; original MSCADLIB + Nuked-OPL3, source WAV and PulseAudio monitor separately retained','private_local_only':True},ensure_ascii=False,indent=2)+'\n')
finally:
    for p in reversed(children):stop(p)

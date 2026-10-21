#!/usr/bin/env python3
"""053：ada的VirtualGL EGL硬體繪圖、NVENC錄影與原版音訊擷取。"""
import ast
from pathlib import Path
import socket

P=Path('/project')
helper=P/'tools/integration_smoke.py'
source=helper.read_text()
source=source.replace("'LIBGL_ALWAYS_SOFTWARE': '1', ",'')
source=source.replace("'GOMAXPROCS': '2'","'GOMAXPROCS': '4'")
source=source.replace("command = [str(B / 'zeliard-preview')", "command = ['/opt/VirtualGL/bin/vglrun','-d','egl',str(B / 'zeliard-preview')")
source=source.replace("    for option, relative in json.loads((B / 'runtime-args.json').read_text()):", "    command += json.loads((B / 'bundle-options.json').read_text()) + ['-audio-wav',str(directory/'source.wav')]\n    for option, relative in json.loads((B / 'runtime-args.json').read_text()):")
source=source.replace('    from PIL import Image\n','')
source=source.replace("        im = Image.open(path).convert('RGB')", "        rgb_buffer=invoke(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-i',str(path),'-threads','1','-f','rawvideo','-pix_fmt','rgb24','-']).stdout\n        assert len(rgb_buffer)==960*680*3\n        im=type('Frame',(),{'getpixel':lambda self,xy:rgb_buffer[(xy[1]*960+xy[0])*3:(xy[1]*960+xy[0])*3+3]})()")
nodes=[]
for node in ast.parse(source).body:
    if isinstance(node,ast.Try):break
    nodes.append(node)
exec(compile(ast.Module(body=nodes,type_ignores=[]),str(helper),'exec'))
names.update({0x39:'space',0x50:'Down',0x49:'KP_Prior'})
env.pop('LIBGL_ALWAYS_SOFTWARE',None)
C=Path('/recording')
original_invoke=invoke
record_begin=None
markers=[]


def invoke(args):
    if args[0]=='import':
        args=['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-y','-f','x11grab','-video_size','960x680','-draw_mouse','0','-i',':96.0+0,0','-frames:v','1','-threads','1',args[-1]]
    return original_invoke(args)


def mark(name):
    item={'name':name,'wall_seconds':time.monotonic()-record_begin,'steps':state()['steps']}
    markers.append(item)
    print('MARK',name,round(item['wall_seconds'],2),flush=True)


original_shot=shot
def shot(name):
    original_shot(name)
    shots[-1]['wall_seconds']=time.monotonic()-record_begin


def hold_seconds(scan,seconds):
    edge(scan,True)
    try:time.sleep(seconds)
    finally:edge(scan,False)


try:
    for name in ('runtime','pulse-state','pulse-config'):(C/name).mkdir(mode=0o700)
    server='unix:/recording/runtime/native'
    (C/'server.pa').write_text('load-module module-native-protocol-unix socket=/recording/runtime/native auth-anonymous=1 auth-cookie-enabled=0\nload-module module-null-sink sink_name=zeliard_ada format=s16le rate=44100 channels=2\nset-default-sink zeliard_ada\n')
    (C/'asound.conf').write_text('pcm.!default { type pulse server "'+server+'" device "zeliard_ada" }\nctl.!default { type pulse server "'+server+'" }\n')
    (C/'client.conf').write_text('default-server = '+server+'\nautospawn = no\n')
    env.update({'PULSE_SERVER':server,'PULSE_RUNTIME_PATH':str(C/'runtime'),'PULSE_STATE_PATH':str(C/'pulse-state'),'PULSE_CONFIG_PATH':str(C/'pulse-config'),'PULSE_CLIENTCONFIG':str(C/'client.conf'),'ALSA_CONFIG_PATH':str(C/'asound.conf'),'XDG_RUNTIME_DIR':str(C/'runtime')})
    pulse=subprocess.Popen(['pulseaudio','--daemonize=no','--exit-idle-time=-1','--use-pid-file=no','-n','-F',str(C/'server.pa')],env=env,stdout=(C/'pulse.log').open('wb'),stderr=subprocess.STDOUT);children.append(pulse)
    deadline=time.monotonic()+10
    while not (C/'runtime/native').exists():
        assert time.monotonic()<deadline and pulse.poll() is None
        time.sleep(.05)
    monitor_begin_epoch=time.time()
    monitor=subprocess.Popen(['parec','--device=zeliard_ada.monitor','--format=s16le','--rate=44100','--channels=2'],env=env,stdout=(C/'device.pcm').open('wb'),stderr=(C/'parec.log').open('wb'));children.append(monitor)
    x=subprocess.Popen(['Xvfb',':96','-screen','0','1280x800x24','-nolisten','tcp'],env=env,stdout=(C/'xvfb.log').open('wb'),stderr=subprocess.STDOUT);children.append(x)
    deadline=time.monotonic()+10
    while not Path('/tmp/.X11-unix/X96').exists():
        assert time.monotonic()<deadline
        time.sleep(.05)
    renderer=invoke(['/opt/VirtualGL/bin/vglrun','-d','egl','/opt/VirtualGL/bin/glxinfo','-B']).stdout.decode()
    assert 'NVIDIA' in renderer and 'llvmpipe' not in renderer.lower(),renderer
    (C/'renderer.txt').write_text(renderer)
    save=B/'original/SAVEM1.USR'
    assert h(save)=='03df9abcf62b1712d6b46dd1fc5b23335a4b8711b3dc84875a1498260d37735c'
    user=O/'user/game';user.mkdir(parents=True)
    shutil.copy2(save,user/'SAVETEST.usr')
    start()
    invoke(['xdotool','windowmove',window,'0','0'])
    maps=Path('/proc')/str(process.pid)/'maps'
    loaded=[line for line in maps.read_text().splitlines() if any(s in line.lower() for s in ['nvidia','vglfaker'])]
    assert any('libvglfaker' in line for line in loaded) and any('nvidia' in line for line in loaded)
    (C/'game-gpu-libraries.txt').write_text('\n'.join(loaded)+'\n')
    record_begin_epoch=time.time()
    video=subprocess.Popen(['ffmpeg','-nostdin','-hide_banner','-loglevel','info','-f','x11grab','-video_size','960x680','-framerate','60','-draw_mouse','0','-i',':96.0+0,0','-an','-c:v','h264_nvenc','-preset','p4','-cq','20','-b:v','0','-pix_fmt','yuv420p','-movflags','+faststart',str(C/'gameplay-source.mp4')],env=env,stdout=subprocess.DEVNULL,stderr=(C/'ffmpeg.log').open('wb'));children.append(video)
    record_begin=time.monotonic()
    advance(12000000)
    for _ in range(6):key(0x39,50000);advance(7950000)
    key(0x41,500000);advance(2000000);key(0x15,150000);advance(1500000)
    for _ in range(8):key(0xe,100000);advance(200000)
    for scan in (0x1f,0x1e,0x2f,0x12,0x14,0x12,0x1f,0x14):key(scan,100000);advance(200000)
    key(0x1c,100000);advance(16000000)
    hd(True);mark('town-walk-begin');shot('town-HD')
    key(0x4b,27000000);advance(500000)
    hd(False);align('church-door',(160,102,176,118),0);hd(True)
    key(0x48,500000)
    wait('中文教堂對話',lambda s:s['dialogue']['key']=='zelres2.file.01a508' and s['dialogue']['prompt'])
    mark('church-dialogue');shot('church-HD');hd(False);shot('church-original');hd(True)
    key(0x39,500000)
    for _ in range(3):
        wait('祝福完成',lambda s:s['dialogue']['key']=='zelres2.file.01a568' and s['dialogue']['complete'] and s['text_cursor']==0xa3e2 and s['advance_input_flags']==0)
        advance(500000);key(0x39,500000);advance(6000000)
        if not state()['dialogue']['active']:break
    else:raise RuntimeError('教堂未正常離開')
    wait('回城鎮HD',lambda s:s['town_scenery_hd_matched'] and s['town_ground_hd_matched'],30)
    mark('town-return');shot('town-return-HD');hd(False);shot('town-original')
    mark('original-town-walk');key(0x4d,18000000);hd(True);mark('HD-town-walk');key(0x4d,18000000);advance(500000)
    hd(False);align('cave-for-wound-door',(216,102,240,118),16);hd(True)
    key(0x48,500000);advance(16000000)
    wait('第一洞窟HD',lambda s:s['cave_terrain_hd_matched'],30)
    mark('cave-entry');shot('cave-HD');hd(False);shot('cave-original');hd(True)
    mark('cave-action')
    for _ in range(3):
        hold_seconds(0x39,.35);hold_seconds(0x4d,.7);hold_seconds(0x48,.3)
    hold_seconds(0x4b,1.4)
    for _ in range(4):hold_seconds(0x39,.25);time.sleep(.2)
    hold_seconds(0x48,.5);hold_seconds(0x4d,.8);time.sleep(1)
    shot('cave-played-HD');mark('gameplay-end')
    panel(True);shot('keyboard-settings');close_panel()
    panel(False);shot('controller-settings');close_panel()
    edge(0x1d,True);key(0x10,500000);edge(0x1d,False);advance(2000000)
    invoke(['xdotool','keydown','y'])
    try:process.wait(timeout=60)
    finally:invoke(['xdotool','keyup','y'])
    assert process.returncode==0 and h(user/'SAVETEST.usr')==h(save)
    assert (user/'SAVETEST.usr').stat().st_mtime_ns==save.stat().st_mtime_ns
    stop(video);stop(monitor)
    assert (C/'gameplay-source.mp4').stat().st_size>100000
    report={'status':'PASS_ADA_NEW_GAMEPLAY_CAPTURE','version':(B/'VERSION').read_text().strip(),'machine':'ada','container_hostname':socket.gethostname(),'renderer':renderer,'binary_sha256':h(B/'zeliard-preview'),'source_save_sha256':h(save),'script_sha256':h(Path(__file__)),'video_sha256':h(C/'gameplay-source.mp4'),'encoder':'h264_nvenc','recording_fps':60,'record_begin_epoch':record_begin_epoch,'monitor_begin_epoch':monitor_begin_epoch,'game_gpu_libraries':'game-gpu-libraries.txt','boots':boots,'shots':shots,'markers':markers,'normal_DOS_exit':0,'save_unchanged':True,'rights':'private local only','limits':'正常F7副本遊玩；非通關或後段驗收，工具列只展示、不冒稱實體手把驗收'}
    (C/'receipt.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(report['status'],flush=True)
except Exception:
    if process is not None and process.poll() is None:
        invoke(['import','-window',window,str(O/'failed-live.png')])
        (O/'failed-state.json').write_text(json.dumps(state(),ensure_ascii=False,indent=2)+'\n')
    raise
finally:
    for child in reversed(children):stop(child)

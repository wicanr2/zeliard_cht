"""063：正式包的新功能與HD穩定短錄，只在有界Docker／VirtualGL內執行。"""
from pathlib import Path
import hashlib,json,os,re,signal,subprocess,threading,time

P=Path('/project');C=Path('/output/recording')
source=(P/'tools/integration_smoke.py').read_text()
source=source.replace("O = Path('/output')", "O = Path('/output/capture')")
source=source.replace("'LIBGL_ALWAYS_SOFTWARE': '1', ", '')
exec(compile(source.split('signal.signal(signal.SIGTERM,')[0],'normal_input_helpers','exec'))
env.update(GOMAXPROCS='2',LP_NUM_THREADS='1')
names.update({0x39:'space',0x50:'Down',0x49:'KP_Prior'})
assert C.is_dir() and C.stat().st_uid==os.getuid() and not any(C.iterdir())
markers=[];events=[];rows=[];active=None;sample_stop=threading.Event();video=None
record_begin=time.monotonic();record_epoch=time.time()

def mark(name):
    s=state() if process is not None else None
    item={'name':name,'wall_seconds':time.monotonic()-record_begin,'steps':s['steps'] if s else None}
    markers.append(item);print('MARK',name,round(item['wall_seconds'],2),flush=True)

def click(x,y):
    events.append({'kind':'click','x':x,'y':y,'wall_seconds':time.monotonic()-record_begin})
    invoke(['xdotool','mousemove','--window',window,str(x),str(y),'click','1'])

original_edge=edge
def edge(scan,pressed):
    events.append({'kind':'key','scan':scan,'pressed':pressed,'wall_seconds':time.monotonic()-record_begin})
    return original_edge(scan,pressed)

def key(scan,n):
    edge(scan,True);target=state()['steps']+n;end=time.monotonic()+100
    try:
        while state()['steps']<target:
            assert time.monotonic()<end,'原版按鍵逾時'
            if state()['input_gate']:invoke(['xdotool','keyup',names[scan]])
            time.sleep(.025)
    finally:edge(scan,False)

def picture(name):
    file=C/(name+'.png')
    invoke(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-f','x11grab','-video_size','960x680','-draw_mouse','0','-i',':96.0+0,0','-frames:v','1','-threads','1',str(file)])
    (C/(name+'.json')).write_text(json.dumps(state(),ensure_ascii=False,indent=2)+'\n')

def sample_loop():
    last=None
    while not sample_stop.wait(.02):
        label=active
        if label is None:continue
        try:s=state()
        except AssertionError:continue
        if not s or (label,s['frames'])==last:continue
        last=(label,s['frames'])
        town=s.get('town_frame_presentation') or {};cave=s.get('cave_frame_presentation') or {}
        rows.append({'label':label,'wall_seconds':time.monotonic()-record_begin,'frames':s['frames'],'steps':s['steps'],'hd_enabled':s['hd_enabled'],'paused':s['paused'],'panel':s['panel'],'town_hd':s['town_scenery_hd_matched'],'cave_hd':s['cave_terrain_hd_matched'],'town_source_used':town.get('used',False),'cave_source_used':cave.get('used',False),'dialogue_key':s['town_dialogue']['key'],'dialogue_backdrop':town.get('dialogue_backdrop',False),'platform':s.get('cave_common_hd'),'map_origin':cave.get('source_map_origin'),'cheats':s['cheats'],'performance':s['performance']})

def begin_proof(label):
    global active
    assert state()['hd_enabled'] and not state()['paused'] and not state()['panel']
    mark(label+'-begin');active=label

def end_proof(label):
    global active
    active=None;mark(label+'-end')
    group=[r for r in rows if r['label']==label]
    assert len(group)>=30,(label,'too few source observations')
    assert all(r['hd_enabled'] and not r['paused'] and not r['panel'] for r in group),label
    field='town' if label=='dialogue-proof' else 'cave'
    assert all(r[field+'_hd'] and r[field+'_source_used'] for r in group),(label,'native fallback present')
    start=next(x['wall_seconds'] for x in markers if x['name']==label+'-begin')
    assert not any(e['kind']=='click' and e['y']<48 and e['x']<264 and start<=e['wall_seconds']<=markers[-1]['wall_seconds'] for e in events),'證明區間手動切換HD'

def ready(s):return s['cave_terrain_hd_matched'] and bool((s.get('cave_frame_presentation')or{}).get('used'))
def xvalue(s):return s['native_primary_0080']+s['native_primary_0083']
def yvalue(s):return (s['native_primary_0082']+s['native_primary_0084'])&63
def platform(s):
    v=s.get('cave_common_hd')or{}
    return ready(s) and v.get('matched') and len(v.get('placements')or[])==1
def platform_row(s):return (s['native_primary_0082']+(s['cave_common_hd']['placements'][0]['Y']-14)//8)&63
def move(scan,condition,label,seconds=25):
    edge(scan,True)
    try:s=wait(label,condition,seconds)
    finally:edge(scan,False)
    advance(100000);return s

signal.signal(signal.SIGTERM,lambda *_:(_ for _ in ()).throw(TimeoutError('有界影片擷取')))
try:
    x=subprocess.Popen(['Xvfb',':96','-screen','0','1280x800x24','-nolisten','tcp'],env=env,stdout=(C/'xvfb.log').open('wb'),stderr=subprocess.STDOUT);children.append(x)
    end=time.monotonic()+10
    while not Path('/tmp/.X11-unix/X96').exists():
        assert time.monotonic()<end;time.sleep(.05)
    renderer=invoke(['/opt/VirtualGL/bin/vglrun','-d','egl','/opt/VirtualGL/bin/glxinfo','-B']).stdout.decode()
    assert 'direct rendering: Yes' in renderer and not any(s in renderer.lower() for s in ('llvmpipe','softpipe','swrast')),renderer
    (C/'renderer.txt').write_text(renderer)
    video=subprocess.Popen(['ffmpeg','-nostdin','-hide_banner','-loglevel','info','-filter_threads','1','-vaapi_device','/dev/dri/renderD128','-f','x11grab','-video_size','960x680','-framerate','30','-draw_mouse','0','-i',':96.0+0,0','-an','-vf','format=nv12,hwupload','-c:v','h264_vaapi','-qp','20','-movflags','+faststart',str(C/'feature-source.mp4')],env=env,stdout=subprocess.DEVNULL,stderr=(C/'ffmpeg.log').open('wb'));children.append(video)
    sampler=threading.Thread(target=sample_loop,daemon=True);sampler.start()
    user=O/'user/game';user.mkdir(parents=True)
    save=B/'starter/SAVETEST.usr';shutil.copy2(save,user/'SAVETEST.usr')
    directory=O/'boot1';directory.mkdir()
    command=['/opt/VirtualGL/bin/vglrun','-d','egl',str(B/'AppRun'),'-audio','off','-show-controller','-user-dir',str(O/'user'),'-observe',str(directory/'observe.json'),'-capture',str(directory/'capture.png'),'-capture-request',str(directory/'capture.request')]
    process=subprocess.Popen(command,env=env,stdout=(directory/'game.log').open('wb'),stderr=subprocess.STDOUT);children.append(process)
    end=time.monotonic()+60
    while time.monotonic()<end:
        assert process.poll() is None
        r=subprocess.run(['xdotool','search','--onlyvisible','--pid',str(process.pid)],env=env,capture_output=True,timeout=3)
        if r.returncode==0 and r.stdout.strip():window=r.stdout.decode().splitlines()[-1];break
        time.sleep(.05)
    assert window is not None
    invoke(['xdotool','windowmove',window,'0','0']);invoke(['xdotool','windowfocus',window])
    wait('啟動設定就緒',lambda s:s['panel'] and s['focused'],70)
    assert (B/'VERSION').read_text().strip()=='v.1.0.4-20261022'
    click(644,24);wait('暫停',lambda s:s['paused'],10)
    click(330,24);wait('速度選單',lambda s:s['speed_panel'],10)
    click(480,291);wait('選4x',lambda s:s['cpu_multiplier']==4 and not s['panel'],10)
    click(330,24);wait('速度頁重新開啟',lambda s:s['speed_panel'],10)
    mark('speed-feature-begin');time.sleep(8);picture('speed-settings');mark('speed-feature-end')
    click(480,191);wait('還原1x',lambda s:s['cpu_multiplier']==1 and not s['panel'],10)
    click(568,24);wait('作弊頁',lambda s:s['cheat_panel'],10)
    mark('cheat-feature-begin');time.sleep(1)
    click(480,212);wait('鎖滿生命',lambda s:s['cheats']['lock_hp'],10);time.sleep(1.3)
    click(480,302);wait('鎖滿MP',lambda s:s['cheats']['lock_mp'],10);time.sleep(1.3)
    click(480,392);wait('停用站立回血',lambda s:not s['cheats']['standing_regeneration'],10);time.sleep(5)
    picture('cheat-settings');mark('cheat-feature-end')
    # 保留生命鎖定供片中洞窟操作；其餘恢復原版設定，來源明示此差異。
    click(480,302);click(480,392);wait('MP與回血還原',lambda s:not s['cheats']['lock_mp'] and s['cheats']['standing_regeneration'],10)
    click(820,580);wait('關閉作弊頁',lambda s:not s['panel'],10)
    hd(False);click(644,24);wait('開始原版',lambda s:not s['paused'] and not s['input_gate'],10)
    advance(12000000)
    for _ in range(6):key(0x39,50000);advance(7950000)
    key(0x41,500000);advance(2000000);key(0x15,150000);advance(1500000)
    for _ in range(8):key(0xe,100000);advance(200000)
    for scan in (0x1f,0x1e,0x2f,0x12,0x14,0x12,0x1f,0x14):key(scan,100000);advance(200000)
    key(0x1c,100000);advance(16000000)
    mark('town-load-begin');hd(True)
    wait('城鎮HD',lambda s:s['town_scenery_hd_matched'] and not s['input_gate'],80);mark('town-load-end')
    for _ in range(25):
        key(0x4b,250000);advance(250000);key(0x39,500000);advance(100000)
        if state()['town_dialogue']['key']:break
    wait('中文對話HD背景',lambda s:bool(s['town_dialogue']['key']) and s['town_scenery_hd_matched'] and (s.get('town_frame_presentation')or{}).get('dialogue_backdrop'),30)
    time.sleep(.4);begin_proof('dialogue-proof');time.sleep(5)
    key(0x39,500000);time.sleep(5);end_proof('dialogue-proof');picture('dialogue-HD')
    for _ in range(3):
        if not state()['town_dialogue']['key']:break
        key(0x39,500000);advance(100000)
    assert not state()['town_dialogue']['key']
    key(0x4d,8000000);advance(500000);hd(False)
    # 入口圖樣辨識沿既有正常導航，硬體工具映像用FFmpeg讀取RGB。
    import ast
    node=next(n for n in ast.parse(source).body if isinstance(n,ast.FunctionDef) and n.name=='align')
    text=ast.unparse(node).replace('    from PIL import Image\n','')
    text=text.replace("im = Image.open(path).convert('RGB')", "rgb_buffer=invoke(['ffmpeg','-nostdin','-loglevel','error','-i',str(path),'-threads','1','-f','rawvideo','-pix_fmt','rgb24','-']).stdout\n        assert len(rgb_buffer)==960*680*3\n        im=type('Frame',(),{'getpixel':lambda self,xy:rgb_buffer[(xy[1]*960+xy[0])*3:(xy[1]*960+xy[0])*3+3]})()")
    exec(compile(text,'hardware_door_alignment','exec'))
    align('cave-for-wound-door',(216,102,240,118),16)
    key(0x48,500000);advance(16000000);mark('cave-load-begin');hd(True)
    wait('第一洞窟HD',lambda s:ready(s) and not s['input_gate'],90);mark('cave-load-end')
    move(0x4b,lambda s:xvalue(s)<=41,'走到繩索')
    move(0x50,lambda s:yvalue(s)>=22,'沿繩索下降')
    for _ in range(5):key(0x39,100000);advance(400000)
    if xvalue(state())<40:move(0x4d,lambda s:xvalue(s)>=40,'靠近平台')
    elif xvalue(state())>41:move(0x4b,lambda s:xvalue(s)<=41,'對齊跳躍')
    move(0x49,lambda s:xvalue(s)>=44,'向右跳上平台');advance(500000)
    for _ in range(12):
        s=wait('平台來源',platform,20);x=xvalue(s)
        if x==44 and yvalue(s)==(platform_row(s)-3)&63:break
        if x!=44:key(0x4b if x>44 else 0x4d,100000)
        advance(300000)
    else:raise RuntimeError('未站穩平台原點44')
    time.sleep(.4);begin_proof('platform-proof');time.sleep(4)
    move(0x48,lambda s:platform(s) and platform_row(s)<=17,'平台升起');time.sleep(1.5)
    move(0x50,lambda s:platform(s) and platform_row(s)>=24,'平台下降');time.sleep(1.5)
    move(0x48,lambda s:platform(s) and platform_row(s)<=17,'平台再升起');time.sleep(1)
    move(0x4d,lambda s:xvalue(s)>=48,'平台右移')
    time.sleep(4);end_proof('platform-proof');picture('platform-HD')
    click(644,24);wait('參考圖暫停',lambda s:s['paused'],10)
    baseline=state();hd(False);mark('platform-original-reference-begin');time.sleep(1);picture('platform-original');mark('platform-original-reference-end')
    hd(True);time.sleep(.5);picture('platform-HD-reference')
    after=state()
    assert all(baseline[k]==after[k] for k in ('steps','cycles','oracle_memory_sha256','oracle_indexed_sha256'))
    sample_stop.set();sampler.join(timeout=2)
    (directory/'capture.request').write_text('capture\n');process.wait(timeout=30);assert process.returncode==0
    video.send_signal(signal.SIGINT);video.wait(timeout=20);assert video.returncode in (0,255)
    assert h(user/'SAVETEST.usr')==h(save) and (user/'SAVETEST.usr').stat().st_mtime_ns==save.stat().st_mtime_ns
    report={'status':'PASS_V104_FEATURE_AND_HD_PROOF_CAPTURE','version':(B/'VERSION').read_text().strip(),'machine':'local','renderer':renderer,'source_binary_sha256':h(B/'zeliard-preview'),'script_sha256':h(Path(__file__)),'helper_sha256':h(P/'tools/integration_smoke.py'),'source_save_sha256':h(save),'source_save_kind':'bundled starter copied via normal F7; SHA recorded','source_video_sha256':h(C/'feature-source.mp4'),'record_begin_epoch':record_epoch,'recording_fps':30,'encoder':'h264_vaapi','markers':markers,'input_events':events,'proof_samples':rows,'hp_lock_active_during_proof':True,'proofs_edited':False,'proofs_manually_switched_hd':False,'source_audio':'off; promo reuses original title music, not synchronous sound','original_save_unchanged':True,'exit_method':'front-end capture request','rights':'local only; original game footage and derivatives','scope':'normal resident dialogue and verified green platform; no full-game or30fps claim'}
    (C/'receipt.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(report['status'],len(rows),flush=True)
except Exception:
    (C/'failed-progress.json').write_text(json.dumps({'markers':markers,'events':events,'rows':rows},ensure_ascii=False,indent=2)+'\n')
    if process is not None and process.poll() is None:
        try:picture('failed-live')
        except Exception:pass
    raise
finally:
    sample_stop.set()
    for child in reversed(children):stop(child)

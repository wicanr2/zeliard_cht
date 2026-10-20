"""Docker/Xvfb正常進武器店、局部動畫、未購物離店姿態與原貌切回。"""
import ast
from pathlib import Path

helper=Path('/project/tools/integration_smoke.py');prefix=[]
for node in ast.parse(helper.read_text(encoding='utf-8')).body:
    if isinstance(node,ast.Try):break
    prefix.append(node)
exec(compile(ast.Module(body=prefix,type_ignores=[]),str(helper),'exec'))
names[0x39]='space';names[0x50]='Down'
candidate=R/'workplace/full-hd-v1/merchant-frames-runtime-v3/zeliard-preview'
pairs=[];visible=set();original_state=state


def merchant(s):return s.get('merchant_frames_hd') or {}


def state():
    s=original_state()
    if s and merchant(s).get('matched'):visible.add(merchant(s)['key'])
    return s


def start_merchant():
    global process,window,directory
    directory=O/'boot1';directory.mkdir();user=O/'user/game';user.mkdir(parents=True)
    shutil.copy2(B/'userdata/game/SAVETEST.usr',user/'SAVETEST.usr')
    command=[str(candidate),'-orig',str(B/'original'),'-font',str(B/'font.hex.gz'),'-user-dir',str(O/'user'),
             '-observe',str(directory/'observe.json'),'-input-trace',str(directory/'inputs.jsonl'),
             '-capture',str(directory/'capture.png'),'-capture-request',str(directory/'capture.request'),'-audio','off','-realtime']
    for option,relative in json.loads((B/'runtime-args.json').read_text()):command+=[option,str(B/relative)]
    for flag,relative in [('merchant-frames-art-dir','merchant-frames-complete-v2'),('sage-frames-art-dir','sage-frames-complete-v2'),
                          ('interior-art-dir','interior-complete-v2'),('town-residents-art-dir','town-npc-art')]:
        command+=['-'+flag,str(R/'workplace/full-hd-v1'/relative)]
    process=subprocess.Popen(command,env=env,stdout=(directory/'game.log').open('wb'),stderr=subprocess.STDOUT)
    children.append(process);end=time.monotonic()+45
    while time.monotonic()<end:
        assert process.poll() is None
        result=subprocess.run(['xdotool','search','--onlyvisible','--pid',str(process.pid)],env=env,capture_output=True,timeout=3)
        if result.returncode==0 and result.stdout.strip():window=result.stdout.decode().splitlines()[-1];break
        time.sleep(.1)
    else:raise TimeoutError('武器店候選視窗')
    invoke(['xdotool','windowfocus',window]);wait('視窗就緒',lambda s:s['focused'] and not s['input_gate'],30)
    boots.append({'command':command})


def frozen(name):
    prior=state();s=wait('暫停畫格完成',lambda s:s['paused'] and s['frames']>prior['frames'],15)
    path=directory/(name+'.png');invoke(['import','-window',window,str(path)])
    (directory/(name+'.json')).write_text(json.dumps(s,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    shots.append({'name':name,'image_sha256':h(path),'state':s});return s


def compare(name,predicate):
    wait('完整武器店目標畫格',lambda s:merchant(s).get('matched') and predicate(merchant(s)),75)
    click(644,24);s=wait('暫停',lambda s:s['paused'] and not held(s),15)
    assert merchant(s).get('matched') and predicate(merchant(s)),('暫停未捕捉目標',merchant(s))
    hd(False);original=frozen(name+'-original')
    hd(True);painted=frozen(name+'-hd');assert merchant(painted).get('matched')
    hd(False);returned=frozen(name+'-original-returned')
    for field in ['steps','cycles','oracle_memory_sha256','oracle_indexed_sha256','oracle_palette_sha256']:
        assert original[field]==painted[field]==returned[field],field
    pairs.append({'name':name,'original':name+'-original.png','hd':name+'-hd.png','returned':name+'-original-returned.png',
                  'key':merchant(painted)['key'],'pose':merchant(painted)['pose'],'state_frozen':True})
    hd(True);click(644,24);wait('恢復',lambda s:not s['paused'],15);print('PAIR',name,merchant(painted),flush=True)


try:
    x=subprocess.Popen(['Xvfb',':96','-screen','0','1280x800x24','-nolisten','tcp'],env=env,
                       stdout=(O/'xvfb.log').open('wb'),stderr=subprocess.STDOUT);children.append(x)
    end=time.monotonic()+10
    while not Path('/tmp/.X11-unix/X96').exists():assert time.monotonic()<end;time.sleep(.05)
    start_merchant()
    for line in (R/'tools/merchant-hd-departure.commands').read_text(encoding='utf-8').splitlines():
        f=line.split()
        if not f or f[0].startswith('#'):continue
        if f==['shot','restored']:break
        if f[0]=='run':advance(int(f[1],0))
        elif f[0]=='key':key(int(f[1],0),int(f[2],0))
    hd(True);wait('正常Muralla',lambda s:s['town_scenery_hd_matched'],40)
    key(0x4b,24000000);advance(500000)
    # 320來自已驗正常fine-3段，不寫座標，只用方向鍵對準原版入口。
    for _ in range(50):
        s=state();world=(s['native_primary_0080']+4+s['native_primary_0083'])*8;delta=world-320
        if abs(delta)<=16:
            # 鄰近值只決定何時嘗試Up；是否進店一律由原版實際轉場判定。
            key(0x48,250000);advance(3000000)
            end=time.monotonic()+3
            while time.monotonic()<end and not merchant(state()).get('matched'):time.sleep(.03)
            if merchant(state()).get('matched'):break
            s=state();world=(s['native_primary_0080']+4+s['native_primary_0083'])*8;delta=world-320
        key(0x4b if delta>0 else 0x4d,800000 if abs(delta)>80 else 50000);advance(100000)
    else:raise RuntimeError('正常武器店入口未對準')
    wait('正常進武器店',lambda s:merchant(s).get('matched'),30)
    compare('merchant-base',lambda m:m['key'] in ('body-0','body-2'))
    compare('merchant-idle-animation',lambda m:m['key'].endswith('-animation-3'))
    key(0x39,500000)
    # 原版先演出告別手勢，再停在B23E續頁；續頁後已經離店。
    # 在第一個空白鍵後擷取，不能等續頁完成才開始觀察手勢。
    compare('merchant-departure-gesture',lambda m:m['pose']>=3)
    wait('未購物告別完成',lambda s:s['text_cursor']==0xb23e and s['advance_input_flags']==0,75)
    key(0x39,500000)
    wait('正常離店回城鎮',lambda s:s['town_scenery_hd_matched'] and not merchant(s).get('matched'),90)
    shot('outside-after-merchant')
    edge(0x1d,True);key(0x10,500000);edge(0x1d,False);advance(1000000)
    invoke(['xdotool','keydown','y'])
    try:process.wait(timeout=60)
    finally:invoke(['xdotool','keyup','y'])
    assert process.returncode==0
    assert h(O/'user/game/SAVETEST.usr')==h(B/'userdata/game/SAVETEST.usr')
    receipt={'schema':1,'status':'PASS_NORMAL_MERCHANT_FRAMES_HD_GUI','binary_sha256':h(candidate),'script_sha256':h(Path(__file__)),
             'boots':boots,'pairs':pairs,'shots':shots,'visible_keys':sorted(visible),'source_save_unchanged':True,'normal_DOS_exit':0,'audio':'off',
             'limits':'正常冷F7進武器店、局部動畫、代表未購物離店姿態及三組原貌切回；購買另以成對原版驗證，不稱本次GUI購買或全部逐幀驗收'}
    (O/'receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n',encoding='utf-8');print(receipt['status'],flush=True)
except Exception as error:
    (O/'failure.json').write_text(json.dumps({'error':repr(error),'state':state() if process and process.poll() is None else None},ensure_ascii=False,indent=2)+'\n',encoding='utf-8');raise
finally:
    for child in reversed(children):stop(child)

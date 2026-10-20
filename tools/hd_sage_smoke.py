"""Docker/Xvfb正常進長老、查看力量、HD切換及正常離開，不注入狀態。"""
import ast
from pathlib import Path

helper=Path('/project/tools/integration_smoke.py')
prefix=[]
for node in ast.parse(helper.read_text(encoding='utf-8')).body:
    if isinstance(node,ast.Try):break
    prefix.append(node)
exec(compile(ast.Module(body=prefix,type_ignores=[]),str(helper),'exec'))
names[0x39]='space';names[0x50]='Down'
candidate=R/'workplace/full-hd-v1/sage-frames-runtime-v1/zeliard-preview'
pairs=[];observed=set()


def start_sage():
    global process,window,directory
    directory=O/'boot1';directory.mkdir()
    user=O/'user/game';user.mkdir(parents=True)
    shutil.copy2(B/'userdata/game/SAVETEST.usr',user/'SAVETEST.usr')
    command=[str(candidate),'-orig',str(B/'original'),'-font',str(B/'font.hex.gz'),
             '-user-dir',str(O/'user'),'-observe',str(directory/'observe.json'),
             '-input-trace',str(directory/'inputs.jsonl'),'-capture',str(directory/'capture.png'),
             '-capture-request',str(directory/'capture.request'),'-audio','off','-realtime']
    for option,relative in json.loads((B/'runtime-args.json').read_text()):command+=[option,str(B/relative)]
    command+=['-sage-frames-art-dir',str(R/'workplace/full-hd-v1/sage-frames-complete-v2')]
    process=subprocess.Popen(command,env=env,stdout=(directory/'game.log').open('wb'),stderr=subprocess.STDOUT)
    children.append(process);end=time.monotonic()+45
    while time.monotonic()<end:
        assert process.poll() is None
        result=subprocess.run(['xdotool','search','--onlyvisible','--pid',str(process.pid)],env=env,capture_output=True,timeout=3)
        if result.returncode==0 and result.stdout.strip():window=result.stdout.decode().splitlines()[-1];break
        time.sleep(.1)
    else:raise TimeoutError('長老候選視窗')
    invoke(['xdotool','windowfocus',window]);wait('視窗就緒',lambda s:s['focused'] and not s['input_gate'],30)
    boots.append({'command':command})


def sage(s):return s.get('sage_frames_hd') or {}


def frozen(name):
    prior=state();s=wait('暫停畫格完成',lambda s:s['paused'] and s['frames']>prior['frames'],15)
    path=directory/(name+'.png');invoke(['import','-window',window,str(path)])
    (directory/(name+'.json')).write_text(json.dumps(s,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    shots.append({'name':name,'image_sha256':h(path),'state':s});return s


def compare(name,gesture):
    wait('完整長老手勢',lambda s:sage(s).get('matched') and (sage(s)['pose']>0 if gesture else sage(s)['pose']==0),40)
    click(644,24);s=wait('暫停',lambda s:s['paused'] and not held(s),15)
    assert sage(s).get('matched') and (sage(s)['pose']>0 if gesture else sage(s)['pose']==0),('暫停未捕捉目標',sage(s))
    hd(False);original=frozen(name+'-original')
    hd(True);painted=frozen(name+'-hd');assert sage(painted).get('matched')
    hd(False);returned=frozen(name+'-original-returned')
    for field in ['steps','cycles','oracle_memory_sha256','oracle_indexed_sha256','oracle_palette_sha256']:
        assert original[field]==painted[field]==returned[field],field
    observed.add(sage(painted)['pose'])
    pairs.append({'name':name,'original':name+'-original.png','hd':name+'-hd.png','returned':name+'-original-returned.png',
                  'pose':sage(painted)['pose'],'phase':sage(painted)['phase'],'animation_tile':sage(painted)['animation_tile'],
                  'native_origin':sage(painted)['native_origin'],'state_frozen':True})
    hd(True);click(644,24);wait('恢復',lambda s:not s['paused'],15);print('PAIR',name,sage(painted),flush=True)


try:
    x=subprocess.Popen(['Xvfb',':96','-screen','0','1280x800x24','-nolisten','tcp'],env=env,
                       stdout=(O/'xvfb.log').open('wb'),stderr=subprocess.STDOUT);children.append(x)
    end=time.monotonic()+10
    while not Path('/tmp/.X11-unix/X96').exists():assert time.monotonic()<end;time.sleep(.05)
    start_sage()
    # 冷開機及F7讀回，沿已核正常鍵序；忽略中間shot，不改任何機器記憶體。
    for line in (R/'tools/sage-hd.commands').read_text(encoding='utf-8').splitlines():
        f=line.split()
        if not f or f[0].startswith('#'):continue
        if f==['shot','restored']:break
        if f[0]=='run':advance(int(f[1],0))
        elif f[0]=='key':key(int(f[1],0),int(f[2],0))
    key(0x48,250000);advance(13000000);hd(True)
    wait('長老初訪選單',lambda s:sage(s).get('matched') and s['text_cursor']==0xadbf,40)
    compare('sage-menu',False)
    key(0x50,100000);advance(1500000);key(0x39,500000)
    wait('查看力量手勢',lambda s:sage(s).get('matched') and sage(s)['pose']>=3,60)
    compare('sage-gesture',True)
    end=time.monotonic()+6
    while time.monotonic()<end:
        s=state()
        if sage(s).get('matched'):observed.add(sage(s)['pose'])
        time.sleep(.03)
    key(0x39,500000);advance(16000000)
    key(0x39,500000);advance(16000000)
    key(0x48,100000);advance(1500000)
    key(0x39,500000);advance(12000000)
    key(0x39,500000);advance(8000000)
    wait('正常回城鎮',lambda s:s['town_scenery_hd_matched'] and not sage(s).get('matched'),40)
    shot('outside-after-power')
    edge(0x1d,True);key(0x10,500000);edge(0x1d,False);advance(1000000)
    invoke(['xdotool','keydown','y'])
    try:process.wait(timeout=60)
    finally:invoke(['xdotool','keyup','y'])
    assert process.returncode==0
    assert h(O/'user/game/SAVETEST.usr')==h(B/'userdata/game/SAVETEST.usr')
    receipt={'schema':1,'status':'PASS_NORMAL_SAGE_HANDS_HD_GUI','binary_sha256':h(candidate),
             'script_sha256':h(Path(__file__)),'boots':boots,'pairs':pairs,'shots':shots,'visible_poses':sorted(observed),
             'source_save_unchanged':True,'normal_DOS_exit':0,'audio':'off',
             'limits':'正常冷F7、進長老、查看力量、代表手勢及原貌切回、離店及退出；非全部逐幀GUI或置中手勢正常驗收'}
    (O/'receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n',encoding='utf-8');print(receipt['status'],flush=True)
except Exception as error:
    (O/'failure.json').write_text(json.dumps({'error':repr(error),'state':state() if process and process.poll() is None else None},ensure_ascii=False,indent=2)+'\n',encoding='utf-8');raise
finally:
    for child in reversed(children):stop(child)

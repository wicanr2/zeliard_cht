"""Docker／Xvfb內正常進城堡小屋，核石像HD、原貌切回與離開。"""
import ast
from pathlib import Path

helper=Path('/project/tools/hd_castle_smoke.py')
tree=ast.parse(helper.read_text(encoding='utf-8'));prefix=[]
for node in tree.body:
    if isinstance(node,ast.Try):break
    prefix.append(node)
exec(compile(ast.Module(body=prefix,type_ignores=[]),str(helper),'exec'))
candidate=R/'workplace/full-hd-v1/interior-runtime-v1/zeliard-preview'
assert candidate.is_file()


def start_interior():
    global process,window,directory
    directory=O/'boot1';directory.mkdir()
    command=[str(candidate),'-orig',str(B/'original'),'-font',str(B/'font.hex.gz'),
             '-user-dir',str(O/'user'),'-observe',str(directory/'observe.json'),
             '-input-trace',str(directory/'inputs.jsonl'),'-capture',str(directory/'capture.png'),
             '-capture-request',str(directory/'capture.request'),'-audio','off','-realtime']
    for option,relative in json.loads((B/'runtime-args.json').read_text()):
        command += [option,str(B/relative)]
    command += ['-interior-art-dir',str(R/'workplace/full-hd-v1/interior-complete-v2'),
                '-town-residents-art-dir',str(R/'workplace/full-hd-v1/town-npc-art'),
                '-town-player-back-art',str(R/'workplace/full-hd-v1/town-neutral-complete-v1'),
                '-castle-memory',str(R/'workplace/town-hd-v1/foreground-v1/normal-v1/cold-town.memory'),
                '-castle-art-dir',str(R/'workplace/full-hd-v1/castle-registration-v1')]
    process=subprocess.Popen(command,env=env,stdout=(directory/'game.log').open('wb'),stderr=subprocess.STDOUT)
    children.append(process);end=time.monotonic()+45
    while time.monotonic()<end:
        assert process.poll() is None
        result=subprocess.run(['xdotool','search','--onlyvisible','--pid',str(process.pid)],env=env,capture_output=True,timeout=3)
        if result.returncode==0 and result.stdout.strip():window=result.stdout.decode().splitlines()[-1];break
        time.sleep(.1)
    else:raise TimeoutError('室內候選視窗')
    invoke(['xdotool','windowfocus',window]);wait('視窗就緒',lambda s:s['focused'] and not s['input_gate'],30)
    boots.append({'command':command})


def interior_ready(s):
    return s.get('interior_hd_matched') and (s.get('interior_hd') or {}).get('profile')=='hut'


def compare_hut():
    wait('石像完整高清',interior_ready,30)
    click(644,24);wait('暫停',lambda s:s['paused'] and not held(s),15)
    hd(False);original=frozen('hut-original')
    hd(True);painted=frozen('hut-hd');assert interior_ready(painted)
    hd(False);returned=frozen('hut-original-returned')
    for key in ['steps','cycles','oracle_memory_sha256','oracle_indexed_sha256','oracle_palette_sha256']:
        assert original[key]==painted[key]==returned[key],key
    pairs.append({'name':'hut','profile':'hut','original':'hut-original.png','hd':'hut-hd.png',
                  'returned':'hut-original-returned.png','native_state_frozen':True,
                  'native_box':[96,30,232,158],'source_pixels':17408})
    hd(True);click(644,24);wait('繼續',lambda s:not s['paused'],15)
    print('PAIR hut',flush=True)


try:
    x=subprocess.Popen(['Xvfb',':96','-screen','0','1280x800x24','-nolisten','tcp'],env=env,
                       stdout=(O/'xvfb.log').open('wb'),stderr=subprocess.STDOUT);children.append(x)
    end=time.monotonic()+10
    while not Path('/tmp/.X11-unix/X96').exists():assert time.monotonic()<end;time.sleep(.05)
    start_interior();advance(12000000)
    for _ in range(6):key(0x39,50000);advance(7950000)
    hd(True);wait('正常城堡',lambda s:ready(s,'castle'),30)
    key(0x4d,11000000);advance(100000)
    for attempt in range(45):
        s=state();world=(s['native_primary_0080']+4+s['native_primary_0083'])*8
        if world==760:break
        key(0x4d if world<760 else 0x4b,50000);advance(25000)
    else:raise RuntimeError('正常亭台入口未對準')
    key(0x48,250000);advance(12000000);compare_hut()
    key(0x39,500000);advance(8000000)
    wait('正常離開小屋',lambda s:ready(s,'castle') and not s.get('interior_hd_matched'),30)
    compare('castle-after-hut','castle')
    edge(0x1d,True);key(0x10,500000);edge(0x1d,False);advance(1000000)
    invoke(['xdotool','keydown','y'])
    try:process.wait(timeout=60)
    finally:invoke(['xdotool','keyup','y'])
    assert process.returncode==0
    receipt={'schema':1,'status':'PASS_NORMAL_HUT_HD_GUI','binary_sha256':h(candidate),'script_sha256':h(Path(__file__)),
             'boots':boots,'pairs':pairs,'shots':shots,'normal_DOS_exit':0,'audio':'off',
             'limits':'冷新局、正常走到城堡亭台、進小屋看石像、離開及原貌切回；非旅館正常到達或全遊戲完成'}
    (O/'receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    print(receipt['status'],flush=True)
except Exception as error:
    (O/'failure.json').write_text(json.dumps({'error':repr(error),'state':state() if process and process.poll() is None else None},ensure_ascii=False,indent=2)+'\n',encoding='utf-8');raise
finally:
    for child in reversed(children):stop(child)

"""Docker／Xvfb內抽測城堡HD、原版四旗幟、原貌切換與城鎮往返。"""
import ast
from pathlib import Path

helper = Path('/project/tools/integration_smoke.py')
tree = ast.parse(helper.read_text(encoding='utf-8'))
prefix = []
for node in tree.body:
    if isinstance(node, ast.Try):
        break
    prefix.append(node)
exec(compile(ast.Module(body=prefix, type_ignores=[]), str(helper), 'exec'))
names[0x39] = 'space'
candidate = R/'workplace/full-hd-v1/castle-runtime-v3/zeliard-preview'
assert candidate.is_file()
pairs = []
flag_ids = set()


def start_castle():
    global process, window, directory
    directory = O/'boot1'
    directory.mkdir()
    command = [str(candidate), '-orig', str(B/'original'), '-font', str(B/'font.hex.gz'),
               '-user-dir', str(O/'user'), '-observe', str(directory/'observe.json'),
               '-input-trace', str(directory/'inputs.jsonl'), '-capture', str(directory/'capture.png'),
               '-capture-request', str(directory/'capture.request'), '-audio', 'off', '-realtime']
    for option, relative in json.loads((B/'runtime-args.json').read_text()):
        command += [option, str(B/relative)]
    command += ['-town-residents-art-dir', str(R/'workplace/full-hd-v1/town-npc-art'),
                '-castle-memory', str(R/'workplace/town-hd-v1/foreground-v1/normal-v1/cold-town.memory'),
                '-castle-art-dir', str(R/'workplace/full-hd-v1/castle-registration-v1')]
    process = subprocess.Popen(command, env=env, stdout=(directory/'game.log').open('wb'), stderr=subprocess.STDOUT)
    children.append(process)
    end = time.monotonic()+45
    while time.monotonic() < end:
        assert process.poll() is None, '城堡候選提早結束'
        result = subprocess.run(['xdotool','search','--onlyvisible','--pid',str(process.pid)],env=env,capture_output=True,timeout=3)
        if result.returncode == 0 and result.stdout.strip():
            window = result.stdout.decode().splitlines()[-1]
            break
        time.sleep(.1)
    else:
        raise TimeoutError('城堡視窗')
    invoke(['xdotool','windowfocus',window])
    wait('視窗就緒',lambda s:s['focused'] and not s['input_gate'],30)
    boots.append({'command':command})


def profile(s):
    return (s.get('town_frame_presentation') or {}).get('profile')


def ready(s, expected):
    return profile(s) == expected and s['town_scenery_hd_matched'] and s['town_ground_hd_matched'] and s['town_residents_hd_matched']


def frozen(name):
    prior = state()
    value = wait('暫停下一畫格',lambda s:s['paused'] and s['frames'] > prior['frames'],15)
    path = directory/(name+'.png')
    invoke(['import','-window',window,str(path)])
    (directory/(name+'.json')).write_text(json.dumps(value,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    shots.append({'name':name,'image_sha256':h(path),'state':value})
    return value


def compare(name, expected='castle'):
    wait('已驗場景與人物HD',lambda s:ready(s,expected),40)
    click(644,24)
    wait('暫停',lambda s:s['paused'] and not held(s),15)
    hd(False)
    original = frozen(name+'-original')
    hd(True)
    painted = frozen(name+'-hd')
    assert ready(painted,expected)
    hd(False)
    returned = frozen(name+'-original-returned')
    for field in ['steps','cycles','oracle_memory_sha256','oracle_indexed_sha256','oracle_palette_sha256']:
        assert original[field] == painted[field] == returned[field],(name,field)
    cells = painted['town_frame_presentation'].get('flag_cells') or []
    flag_ids.update(cell['id'] for cell in cells)
    pairs.append({'name':name,'profile':expected,'original':name+'-original.png','hd':name+'-hd.png',
                  'returned':name+'-original-returned.png','state_frozen':True,'flag_cells':cells,
                  'source_steps':painted['town_frame_presentation']['source_steps']})
    hd(True)
    click(644,24)
    wait('繼續',lambda s:not s['paused'],15)
    print('PAIR',name,expected,'flags',sorted(flag_ids),flush=True)
    return painted


try:
    xvfb = subprocess.Popen(['Xvfb',':96','-screen','0','1280x800x24','-nolisten','tcp'],env=env,
                            stdout=(O/'xvfb.log').open('wb'),stderr=subprocess.STDOUT)
    children.append(xvfb)
    end = time.monotonic()+10
    while not Path('/tmp/.X11-unix/X96').exists():
        assert time.monotonic() < end
        time.sleep(.05)
    start_castle()
    advance(12000000)
    for _ in range(6):
        key(0x39,50000)
        advance(7950000)
    hd(True)
    compare('castle-gate')
    movie = subprocess.Popen(['ffmpeg','-v','error','-nostdin','-thread_queue_size','64',
                              '-f','x11grab','-framerate','30','-video_size','960x680','-i',':96.0+0,0',
                              '-t','12','-an','-c:v','libx264','-preset','fast','-crf','20',str(O/'castle-animation.mkv')],
                             env=env,stdout=(O/'ffmpeg.log').open('wb'),stderr=subprocess.STDOUT)
    children.append(movie)
    for i in range(24):
        if flag_ids == {2,3,4,5}:
            break
        advance(250000)
        present = {v['id'] for v in (state()['town_frame_presentation'].get('flag_cells') or [])}
        if present-flag_ids:
            compare('castle-flags-'+str(i))
    assert flag_ids == {2,3,4,5}, ('未抽到四種原版旗幟',flag_ids)
    key(0x4d,1000000)
    advance(500000)
    compare('castle-walking-right')
    key(0x4b,2000000)
    advance(500000)
    compare('castle-walking-left')
    for _ in range(40):
        s = state()
        assert profile(s) == 'castle', '到亭台前已離開城堡'
        camera = s['town_frame_presentation']['source_camera']
        if camera >= 75:
            break
        key(0x4d,200000 if camera >= 68 else 500000)
        advance(100000)
    else:
        raise RuntimeError('正常步行尚未到亭台')
    compare('castle-pavilion')
    for _ in range(30):
        if ready(state(),'muralla'):
            break
        key(0x4d,500000)
        advance(150000)
    else:
        raise RuntimeError('正常右行尚未離開城堡')
    compare('muralla-after-castle','muralla')
    for _ in range(30):
        if ready(state(),'castle'):
            break
        key(0x4b,500000)
        advance(150000)
    else:
        raise RuntimeError('正常左行尚未返回城堡')
    compare('castle-return')
    movie.wait(timeout=20)
    assert movie.returncode == 0
    edge(0x1d,True)
    key(0x10,500000)
    edge(0x1d,False)
    advance(1000000)
    invoke(['xdotool','keydown','y'])
    try:
        process.wait(timeout=60)
    finally:
        invoke(['xdotool','keyup','y'])
    assert process.returncode == 0
    receipt = {'status':'PASS_CASTLE_HD_GUI_SCOPED','binary_sha256':h(candidate),
               'script_sha256':h(Path(__file__)),'boots':boots,'pairs':pairs,'shots':shots,
               'flag_ids':sorted(flag_ids),'normal_DOS_exit':0,'animation_video':'castle-animation.mkv',
               'limits':'正常新局、步行、四旗幟、城堡／Muralla往返與原貌切回；非全遊戲完成'}
    (O/'receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    print(receipt['status'],flush=True)
except Exception as error:
    (O/'failure.json').write_text(json.dumps({'error':repr(error),'state':state() if process and process.poll() is None else None},ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    raise
finally:
    for child in reversed(children):
        stop(child)

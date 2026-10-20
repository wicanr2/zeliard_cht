"""Docker／Xvfb內以正常冷新局進門，抽測背向HD與原貌切回。"""
import ast
from pathlib import Path

helper = Path('/project/tools/hd_castle_smoke.py')
tree = ast.parse(helper.read_text(encoding='utf-8'))
prefix = []
for node in tree.body:
    if isinstance(node, ast.Try):
        break
    prefix.append(node)
exec(compile(ast.Module(body=prefix, type_ignores=[]), str(helper), 'exec'))
candidate = R/'workplace/full-hd-v1/town-back-runtime-v4/zeliard-preview'
assert candidate.is_file()


def start_back():
    global process, window, directory
    directory = O/('boot'+str(len(boots)+1))
    directory.mkdir()
    command = [str(candidate), '-orig', str(B/'original'), '-font', str(B/'font.hex.gz'),
               '-user-dir', str(O/'user'), '-observe', str(directory/'observe.json'),
               '-input-trace', str(directory/'inputs.jsonl'), '-capture', str(directory/'capture.png'),
               '-capture-request', str(directory/'capture.request'), '-audio', 'off', '-realtime']
    for option, relative in json.loads((B/'runtime-args.json').read_text()):
        command += [option, str(B/relative)]
    command += ['-town-player-back-art', str(R/'workplace/full-hd-v1/town-neutral-complete-v1'),
                '-town-residents-art-dir', str(R/'workplace/full-hd-v1/town-npc-art'),
                '-castle-memory', str(R/'workplace/town-hd-v1/foreground-v1/normal-v1/cold-town.memory'),
                '-castle-art-dir', str(R/'workplace/full-hd-v1/castle-registration-v1')]
    process = subprocess.Popen(command, env=env, stdout=(directory/'game.log').open('wb'), stderr=subprocess.STDOUT)
    children.append(process)
    end = time.monotonic()+45
    while time.monotonic() < end:
        assert process.poll() is None
        result = subprocess.run(['xdotool','search','--onlyvisible','--pid',str(process.pid)],env=env,capture_output=True,timeout=3)
        if result.returncode == 0 and result.stdout.strip():
            window = result.stdout.decode().splitlines()[-1]
            break
        time.sleep(.1)
    else:
        raise TimeoutError('背向候選視窗')
    invoke(['xdotool','windowfocus',window])
    wait('視窗就緒',lambda s:s['focused'] and not s['input_gate'],30)
    boots.append({'command':command})


def is_back(s, bit):
    actor = s.get('town_residents_hd') or {}
    return ready(s, 'castle') and actor.get('player_phase') == 4 and actor.get('player_direction_bit') == bit and actor.get('residents') == 0


def freeze_back(bit):
    edge(0x48, True)
    try:
        wait('正常背向HD畫格',lambda s:is_back(s,bit),25)
        click(644,24)
        wait('暫停背向',lambda s:s['paused'],15)
    finally:
        invoke(['xdotool','keyup','Up'])
    hd(True)
    assert is_back(state(),bit), '背向姿態在暫停前已轉場'
    hd(False)
    original = frozen('back-bit'+str(bit)+'-original')
    hd(True)
    painted = frozen('back-bit'+str(bit)+'-hd')
    assert is_back(painted,bit)
    hd(False)
    returned = frozen('back-bit'+str(bit)+'-original-returned')
    for field in ['steps','cycles','oracle_memory_sha256','oracle_indexed_sha256','oracle_palette_sha256']:
        assert original[field] == painted[field] == returned[field],field
    pairs.append({'name':'back-bit'+str(bit),'profile':'castle','original':'back-bit'+str(bit)+'-original.png',
                  'hd':'back-bit'+str(bit)+'-hd.png','returned':'back-bit'+str(bit)+'-original-returned.png',
                  'boot':len(boots),'state_frozen':True,'player_phase':4,'direction_bit':bit,'residents':0})
    print('BACK PAIR',bit,flush=True)
    (directory/'capture.request').write_text('capture\n')
    process.wait(timeout=30)
    assert process.returncode == 0
    boots[-1]['capture_exit_code'] = 0


try:
    xvfb = subprocess.Popen(['Xvfb',':96','-screen','0','1280x800x24','-nolisten','tcp'],env=env,
                            stdout=(O/'xvfb.log').open('wb'),stderr=subprocess.STDOUT)
    children.append(xvfb)
    end = time.monotonic()+10
    while not Path('/tmp/.X11-unix/X96').exists():
        assert time.monotonic() < end
        time.sleep(.05)
    for bit in (0,1):
        start_back()
        advance(12000000)
        for _ in range(6):
            key(0x39,50000)
            advance(7950000)
        hd(True)
        wait('城堡HD就緒',lambda s:ready(s,'castle'),30)
        # 僅用正常方向鍵；依原版觀察位置逐步靠近門，未寫入位置或姿態。
        def world(s):
            return (s['native_primary_0080']+4+s['native_primary_0083'])*8
        for attempt in range(35):
            value = world(state())
            if value == 416 and (state().get('town_residents_hd') or {}).get('player_direction_bit') == bit:
                break
            direction = 0x4d if value < (424 if bit else 416) else 0x4b
            key(direction,50000)
            advance(25000)
        else:
            raise RuntimeError('正常進門位置未對準')
        freeze_back(bit)
    receipt = {'status':'PASS_TOWN_BACK_HD_GUI','binary_sha256':h(candidate),'script_sha256':h(Path(__file__)),
               'boots':boots,'pairs':pairs,'shots':shots,'audio':'off',
               'limits':'兩條正常冷新局城堡進門與背向原貌切回；擷取結束碼0，非原版DOS退出或其他城鎮驗收'}
    (O/'receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    print(receipt['status'],flush=True)
except Exception as error:
    (O/'failure.json').write_text(json.dumps({'error':repr(error),'state':state() if process and process.poll() is None else None},ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    raise
finally:
    for child in reversed(children):
        stop(child)

"""Docker／Xvfb內抽測居民HD：冷F7、步行、交談、原貌切換及正常退出。"""
import ast
from pathlib import Path

# 沿用已驗的X11按放、進度等待及工具列助手，不執行舊第一洞窟主流程。
helper = Path('/project/tools/integration_smoke.py')
tree = ast.parse(helper.read_text(encoding='utf-8'))
prefix = []
for node in tree.body:
    if isinstance(node, ast.Try):
        break
    prefix.append(node)
exec(compile(ast.Module(body=prefix, type_ignores=[]), str(helper), 'exec'))

names[0x39] = 'space'
candidate = R/'workplace/full-hd-v1/town-npc-runtime-v1/zeliard-preview'
art = R/'workplace/full-hd-v1/town-npc-art'
assert candidate.is_file() and art.is_dir()
pairs, conversations = [], []


def start_npcs():
    global process, window, directory
    directory = O/'boot1'
    directory.mkdir()
    command = [str(candidate), '-orig', str(B/'original'), '-font', str(B/'font.hex.gz'),
               '-user-dir', str(O/'user'), '-observe', str(directory/'observe.json'),
               '-input-trace', str(directory/'inputs.jsonl'), '-capture', str(directory/'capture.png'),
               '-capture-request', str(directory/'capture.request'), '-audio', 'off', '-realtime']
    for option, relative in json.loads((B/'runtime-args.json').read_text()):
        command += [option, str(B/relative)]
    command += ['-town-residents-art-dir', str(art)]
    log = (directory/'game.log').open('wb')
    process = subprocess.Popen(command, env=env, stdout=log, stderr=log)
    children.append(process)
    end = time.monotonic()+45
    while time.monotonic() < end:
        assert process.poll() is None, '候選前端提早結束'
        found = subprocess.run(['xdotool','search','--onlyvisible','--pid',str(process.pid)], env=env, capture_output=True, timeout=3)
        if found.returncode == 0 and found.stdout.strip():
            window = found.stdout.decode().splitlines()[-1]
            break
        time.sleep(.1)
    else:
        raise TimeoutError('居民候選視窗')
    invoke(['xdotool','windowfocus',window])
    wait('視窗就緒', lambda s:s['focused'] and not s['input_gate'], 30)
    boots.append({'command':command})


def paused_frame(name):
    prior = state()
    value = wait('暫停下一畫格', lambda s:s['paused'] and s['frames'] > prior['frames'], 15)
    path = directory/(name+'.png')
    invoke(['import','-window',window,str(path)])
    (directory/(name+'.json')).write_text(json.dumps(value,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    shots.append({'name':name,'image_sha256':h(path),'state':value})
    return value


def compare(name):
    wait('居民與主角HD', lambda s:s['town_residents_hd_matched'] and s['town_player_hd_matched'], 30)
    click(644,24)
    wait('工具列暫停',lambda s:s['paused'] and not held(s),15)
    hd(False)
    original = paused_frame(name+'-original')
    hd(True)
    painted = paused_frame(name+'-hd')
    assert painted['town_residents_hd_matched'] and painted['town_residents_hd']['residents'] > 0
    hd(False)
    returned = paused_frame(name+'-original-returned')
    for field in ['steps','cycles','oracle_memory_sha256','oracle_indexed_sha256','oracle_palette_sha256']:
        assert original[field] == painted[field] == returned[field],(name,field)
    pairs.append({'name':name,'original':name+'-original.png','hd':name+'-hd.png',
                  'returned':name+'-original-returned.png','state_frozen':True,
                  'residents':painted['town_residents_hd'],'steps':painted['steps']})
    hd(True)
    click(644,24)
    wait('恢復遊戲',lambda s:not s['paused'],15)
    print('PAIR',name,painted['steps'],painted['town_residents_hd'],flush=True)


def resident_ready(s):
    q = s['town_dialogue'] or {}
    return q.get('key') and not q.get('pending') and q.get('visible',0)>0


try:
    x = subprocess.Popen(['Xvfb',':96','-screen','0','1280x800x24','-nolisten','tcp'],env=env,
                         stdout=(O/'xvfb.log').open('wb'),stderr=subprocess.STDOUT)
    children.append(x)
    end = time.monotonic()+10
    while not Path('/tmp/.X11-unix/X96').exists():
        assert time.monotonic()<end
        time.sleep(.05)
    user = O/'user/game'
    user.mkdir(parents=True)
    shutil.copy2(B/'starter/SAVETEST.usr',user/'SAVETEST.usr')
    start_npcs()
    # 已驗自建存檔冷讀鍵序，沿用原版輸名；不還原機器或注入位置。
    advance(12000000)
    for _ in range(6):
        key(0x39,50000)
        advance(7950000)
    key(0x41,500000)
    advance(2000000)
    key(0x15,150000)
    advance(1500000)
    for _ in range(8):
        key(0xe,100000)
        advance(200000)
    for scan in [0x1f,0x1e,0x2f,0x12,0x14,0x12,0x1f,0x14]:
        key(scan,100000)
        advance(200000)
    key(0x1c,100000)
    advance(16000000)
    hd(True)
    compare('restored')
    key(0x4d,1000000)
    advance(1000000)
    compare('walking-right')
    key(0x4b,1000000)
    advance(1000000)
    compare('walking-left')
    # 只抽一位居民兩頁及關閉，使用原版移動／交談。
    for step in range(30):
        key(0x4b,125000)
        advance(125000)
        key(0x39,500000)
        advance(100000)
        if (state()['town_dialogue'] or {}).get('key'):
            break
    else:
        raise RuntimeError('正常路徑尚未與居民交談')
    source = state()['town_dialogue']['key']
    for page in range(8):
        wait('居民中文完成',resident_ready,60)
        shot(f'resident-dialogue-{page}')
        conversations.append({'source':source,'page':page,'state':state()['town_dialogue']})
        old = state()['town_dialogue']['glyphs']
        key(0x39,500000)
        advance(100000)
        wait('居民翻頁或關閉',lambda s:not (s['town_dialogue'] or {}).get('key') or s['town_dialogue']['glyphs']>old,60)
        if not (state()['town_dialogue'] or {}).get('key'):
            break
    else:
        raise RuntimeError('居民交談超出抽樣頁數')
    compare('dialogue-closed')
    edge(0x1d,True)
    key(0x10,500000)
    edge(0x1d,False)
    advance(1000000)
    invoke(['xdotool','keydown','y'])
    try:
        process.wait(timeout=60)
    finally:
        invoke(['xdotool','keyup','y'])
    assert process.returncode==0
    assert h(user/'SAVETEST.usr')==h(B/'starter/SAVETEST.usr')
    receipt = {'status':'PASS_RESIDENT_HD_GUI_SCOPED','binary_sha256':h(candidate),
               'script_sha256':h(Path(__file__)),'boots':boots,'pairs':pairs,'shots':shots,
               'conversations':conversations,'normal_DOS_exit':0,'save_unchanged':True,
               'limits':'代表性冷F7、步行、居民交談及原貌切換；非全城鎮、40姿態逐一GUI或全遊戲完成'}
    (O/'receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    print(receipt['status'],flush=True)
except Exception as error:
    (O/'failure.json').write_text(json.dumps({'error':repr(error),'state':state() if process and process.poll() is None else None},ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    raise
finally:
    for child in reversed(children):
        stop(child)

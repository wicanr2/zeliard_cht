"""Docker／Xvfb內正常冷F7到綠色平台，驗升降與原貌切回。"""
import ast
from pathlib import Path

catalog = Path('/project/tools/hd_cave_catalog_smoke.py')
prefix = []
for node in ast.parse(catalog.read_text()).body:
    if isinstance(node, ast.Try):
        break
    prefix.append(node)
exec(compile(ast.Module(body=prefix, type_ignores=[]), str(catalog), 'exec'))
# 工具列預設把斜跳綁定至數字鍵盤9，不能以PageUp代替。
names.update({0x50:'Down', 0x49:'KP_Prior'})
save = R/'workplace/orig/SAVEM1.USR'
assert h(save) == '03df9abcf62b1712d6b46dd1fc5b23335a4b8711b3dc84875a1498260d37735c'
actions = []

def xvalue(s):
    return s['native_primary_0080']+s['native_primary_0083']

def yvalue(s):
    return (s['native_primary_0082']+s['native_primary_0084']) & 63

def move(scan, condition, label, seconds=15):
    edge(scan, True)
    try:
        result = wait(label, condition, seconds)
    finally:
        edge(scan, False)
    actions.append({'label':label,'scan':scan,'steps':result['steps'],'x':xvalue(result),'y':yvalue(result)})
    advance(100000)
    return result

def platform(s):
    value = s.get('cave_common_hd') or {}
    return ready(s) and value.get('matched') and len(value.get('placements') or []) == 1

def platform_row(s):
    point = s['cave_common_hd']['placements'][0]
    return (s['native_primary_0082']+(point['Y']-14)//8) & 63

try:
    xvfb = subprocess.Popen(['Xvfb', ':96', '-screen', '0', '1280x800x24', '-nolisten', 'tcp'], env=env,
                            stdout=(O/'xvfb.log').open('wb'), stderr=subprocess.STDOUT)
    children.append(xvfb)
    end = time.monotonic()+10
    while not Path('/tmp/.X11-unix/X96').exists():
        assert time.monotonic()<end
        time.sleep(.05)
    user = O/'user/game'
    user.mkdir(parents=True)
    shutil.copy2(save, user/'SAVETEST.usr')
    start()
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
    for scan in (0x1f,0x1e,0x2f,0x12,0x14,0x12,0x1f,0x14):
        key(scan,100000)
        advance(200000)
    key(0x1c,100000)
    advance(16000000)
    key(0x4d,8000000)
    advance(250000)
    align_town_entrance()
    key(0x48,250000)
    advance(20000000)
    hd(True)
    wait('正常入洞',ready,30)
    # 原版0080+0083是主角繪圖原點，平台表的足部欄位另加4。
    # 正常來源對照：繩索原點41、跳躍起點40、平台原點44。
    move(0x4b,lambda s:xvalue(s)<=41,'走到繩索')
    move(0x50,lambda s:yvalue(s)>=22,'沿繩索下降')
    for _ in range(5):
        key(0x39,100000)
        advance(400000)
    if xvalue(state())<40:
        move(0x4d,lambda s:xvalue(s)>=40,'靠近平台')
    elif xvalue(state())>41:
        move(0x4b,lambda s:xvalue(s)<=41,'對齊跳躍起點')
    move(0x49,lambda s:xvalue(s)>=44,'正常向右跳上平台')
    advance(500000)
    wait('平台HD來源',platform,20)
    compare('platform-board')
    move(0x48,lambda s:platform(s) and platform_row(s)<=17,'平台升起')
    compare('platform-raised')
    move(0x50,lambda s:platform(s) and platform_row(s)>=24,'平台降下')
    compare('platform-lowered')
    move(0x48,lambda s:platform(s) and platform_row(s)<=17,'平台再次升起')
    compare('platform-reraised')
    move(0x4d,lambda s:xvalue(s)>=53,'正常離開平台')
    shot('platform-left')
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
    assert h(user/'SAVETEST.usr')==h(save)
    assert (user/'SAVETEST.usr').stat().st_mtime_ns==save.stat().st_mtime_ns
    report={'status':'PASS_CAVE_PLATFORM_GUI_SCOPED','binary_sha256':h(candidate),
            'script_sha256':h(Path(__file__)),'catalog_helper_sha256':h(catalog),'helper_sha256':h(helper),
            'source_save_sha256':h(save),'boots':boots,'shots':shots,'pairs':pairs,'actions':actions,
            'normal_DOS_exit':0,'save_unchanged':True,
            'limits':'僅綠色平台正常冷F7、跳上、升降、離開與HD切回；未驗魔法石、其他共用圖格或全遊戲'}
    (O/'receipt.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(report['status'],flush=True)
except Exception:
    (O/'failed-observations.json').write_text(json.dumps({'shots':shots,'pairs':pairs,'actions':actions},ensure_ascii=False,indent=2)+'\n')
    if process is not None and process.poll() is None:
        invoke(['import','-window',window,str(O/'failed-live.png')])
        (O/'failed-state.json').write_text(json.dumps(state(),ensure_ascii=False,indent=2)+'\n')
    raise
finally:
    for child in reversed(children):
        stop(child)

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
magia_enabled = os.environ.get('ZELIARD_GUI_MAGIA') == '1'
doors_enabled = os.environ.get('ZELIARD_GUI_DOORS') == '1'
cyan_enabled = os.environ.get('ZELIARD_GUI_CYAN_WEAPON') == '1'
extra_player_enabled = os.environ.get('ZELIARD_GUI_EXTRA_PLAYER') == '1'

def cyan_ready(s):
    weapon = s.get('cave_weapon_hd') or {}
    return ready(s) and weapon.get('matched') and weapon.get('pose',-1)>=10

def extra_player_ready(s):
    player = s.get('cave_player_hd') or {}
    return ready(s) and player.get('matched') and 55<=player.get('pose',-1)<67

def compare_action(name,scan,condition,label):
    edge(scan,True)
    try:
        wait(label,condition,20)
        click(644,24)
    finally:
        edge(scan,False)
    wait('動作畫面暫停',lambda s:s['paused'] and not held(s),15)
    capture_paused_action(name,condition)

def capture_paused_action(name,condition):
    assert condition(state()),'暫停未保留指定HD動作'
    hd(False)
    original=frozen(name+'-original')
    hd(True)
    painted=frozen(name+'-hd')
    assert condition(painted)
    hd(False)
    returned=frozen(name+'-returned')
    for field in ('steps','cycles','oracle_memory_sha256','oracle_indexed_sha256','oracle_palette_sha256'):
        assert original[field]==painted[field]==returned[field],field
    pairs.append({'name':name,'original':name+'-original.png','hd':name+'-hd.png','returned':name+'-returned.png'})
    hd(True)
    click(644,24)
    wait('動作抽測後繼續',lambda s:not s['paused'],15)
    print('PAIR '+name,flush=True)

def compare_cyan():
    # 沿正常配對的100000步短按及400000步間隔；不長按到未知FF41分支。
    for _ in range(6):
        key(0x39,100000)
        try:
            wait('短按後青藍刀光HD',cyan_ready,.6)
        except TimeoutError:
            advance(400000)
            continue
        click(644,24)
        wait('刀光暫停',lambda s:s['paused'] and not held(s),15)
        if cyan_ready(state()):
            capture_paused_action('cyan-weapon',cyan_ready)
            return
        click(644,24)
        wait('錯過刀光後繼續',lambda s:not s['paused'],15)
        advance(400000)
    raise RuntimeError('正常短按未擷取到青藍刀光')

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

def align_platform_origin():
    # 原043正常來源已驗主角原點44、足部欄48；跳躍放開後仍可能多走兩格。
    # 只送正常方向鍵修正站位，不因畫面有平台就假定已站上操作位置。
    for _ in range(12):
        s=wait('平台完成來源',platform,15)
        x=xvalue(s)
        if x==44 and yvalue(s)==(platform_row(s)-3)&63:
            return
        if x!=44:
            key(0x4b if x>44 else 0x4d,100000)
        advance(300000)
    raise RuntimeError('未站穩已驗平台原點44')

def magia_ready(s):
    value = s.get('cave_magia_hd') or {}
    return value.get('matched') and value.get('visible_positions',0)>0

original_compare = compare
def compare(name):
    if magia_enabled:
        wait('平台與魔法石HD來源',lambda s:platform(s) and magia_ready(s),30)
    original_compare(name)
    if magia_enabled:
        assert magia_ready(shots[-2]['state']), '暫停來源沒有可見魔法石'

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
    if extra_player_enabled:
        compare_action('extra-player',0x4b,extra_player_ready,'正常主角擴充姿態')
        if xvalue(state())<=41:
            move(0x4d,lambda s:xvalue(s)>=42,'回到繩索右側')
    # 原版0080+0083是主角繪圖原點，平台表的足部欄位另加4。
    # 正常來源對照：繩索原點41、跳躍起點40、平台原點44。
    move(0x4b,lambda s:xvalue(s)<=41,'走到繩索')
    move(0x50,lambda s:yvalue(s)>=22,'沿繩索下降')
    def inactive_ff41(s):
        p=(s.get("cave_frame_presentation")or{}).get("body_primary_fields")or{}
        return ready(s) and p.get("01B1:FF41")==1 and p.get("01B1:FF40")==0 and (s.get("cave_player_hd")or{}).get("matched")
    # 原失敗GUI最後一次放開攻擊在202355121步；候選在放開後完成。
    # 先觀察非零分支或走到原虛擬時間，再放開，不能一邊持續攻擊一邊等FF40=0。
    def ff41_seen(s):
        f=s.get("cave_frame_presentation")or{}
        rejected=(f.get("last_rejected_source")or{}).get("body_primary_fields")or{}
        body=f.get("body_primary_fields")or{}
        return body.get("01B1:FF41")==1 or rejected.get("01B1:FF41")==1
    edge(0x39,True)
    try:
        wait("原版攻擊來源或原日誌放開時刻",lambda s:ff41_seen(s) or s["steps"]>=202355121,90)
    finally:
        edge(0x39,False)
    wait("放開後正常FF41=1且FF40=0",inactive_ff41,15)
    click(644,24)
    wait("FF41畫格暫停",lambda s:s["paused"] and not held(s),15)
    capture_paused_action("ff41-inactive",inactive_ff41)
    edge(0x1d,True)
    key(0x10,500000)
    edge(0x1d,False)
    advance(1000000)
    invoke(["xdotool","keydown","y"])
    try:
        process.wait(timeout=60)
    finally:
        invoke(["xdotool","keyup","y"])
    assert process.returncode==0
    assert h(user/"SAVETEST.usr")==h(save)
    assert (user/"SAVETEST.usr").stat().st_mtime_ns==save.stat().st_mtime_ns
    report={"status":"PASS_CAVE_FF41_INACTIVE_GUI_SCOPED","binary_sha256":h(candidate),
            "script_sha256":h(Path(__file__)),"catalog_helper_sha256":h(catalog),"helper_sha256":h(helper),
            "source_save_sha256":h(save),"boots":boots,"shots":shots,"pairs":pairs,"actions":actions,
            "normal_DOS_exit":0,"save_unchanged":True,
            "limits":"只驗正常冷F7與FF41=1/FF40=0畫格、既有HD及原貌切回；其他FF41與全遊戲仍待完成"}
    (O/"receipt.json").write_text(json.dumps(report,ensure_ascii=False,indent=2)+"\n")
    print(report["status"],flush=True)
except Exception:
    (O/'failed-observations.json').write_text(json.dumps({'shots':shots,'pairs':pairs,'actions':actions},ensure_ascii=False,indent=2)+'\n')
    if process is not None and process.poll() is None:
        invoke(['import','-window',window,str(O/'failed-live.png')])
        (O/'failed-state.json').write_text(json.dumps(state(),ensure_ascii=False,indent=2)+'\n')
    raise
finally:
    for child in reversed(children):
        stop(child)

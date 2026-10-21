"""052：正常冷F7、接敵受傷、暫停切換HD及正常退出；只驗本輪修正。"""
import ast
from pathlib import Path

helper=Path('/project/tools/integration_smoke.py')
source=helper.read_text()
source=source.replace("    for option, relative in json.loads((B / 'runtime-args.json').read_text()):", "    command += json.loads((B / 'bundle-options.json').read_text()) + ['-audio','off']\n    for option, relative in json.loads((B / 'runtime-args.json').read_text()):")
source=source.replace('    from PIL import Image\n','')
source=source.replace("        im = Image.open(path).convert('RGB')", "        rgb_buffer = invoke(['convert',str(path),'-depth','8','rgb:-']).stdout\n        assert len(rgb_buffer)==960*680*3\n        im = type('Frame',(),{'getpixel':lambda self,xy:rgb_buffer[(xy[1]*960+xy[0])*3:(xy[1]*960+xy[0])*3+3]})()")
nodes=[]
for node in ast.parse(source).body:
    if isinstance(node,ast.Try):break
    nodes.append(node)
exec(compile(ast.Module(body=nodes,type_ignores=[]),str(helper),'exec'))
names.update({0x39:'space',0x50:'Down'})


def move(scan,condition,label):
    edge(scan,True)
    try:wait(label,condition,25)
    finally:edge(scan,False)


def injury(s):
    f=s.get('cave_frame_presentation') or {}
    return f.get('used') and (f.get('body_primary_fields') or {}).get('01B1:00E8')==255


def injury_hd(s):
    return injury(s) and s['hd_enabled'] and s['cave_terrain_hd_matched'] and (s.get('cave_player_hd') or {}).get('matched')


def hit_flash_hd(s):
    f=s.get('cave_frame_presentation') or {}
    return (f.get('used') and (f.get('body_primary_fields') or {}).get('01B1:FF36')==255
            and s['hd_enabled'] and s['cave_terrain_hd_matched'] and (s.get('cave_player_hd') or {}).get('matched'))


def frozen(name):
    before=state()
    wait('暫停繪製完成',lambda s:s['paused'] and s['frames']>before['frames'],15)
    s=state()
    assert not held(s)
    invoke(['import','-window',window,str(directory/(name+'.png'))])
    (directory/(name+'.json')).write_text(json.dumps(s,ensure_ascii=False,indent=2)+'\n')
    shots.append({'name':name,'state':s,'image_sha256':h(directory/(name+'.png'))})
    return s


def compare_paused(name,condition):
    assert condition(state()),'錯過指定完成畫面'
    painted=frozen(name+'-HD')
    hd(False);original=frozen(name+'-original')
    hd(True);returned=frozen(name+'-HD-returned')
    assert condition(returned)
    for field in ('steps','cycles','oracle_memory_sha256','oracle_indexed_sha256','oracle_palette_sha256'):
        assert painted[field]==original[field]==returned[field],field
    def game_pixels(suffix):
        return invoke(['convert',str(directory/(name+suffix+'.png')),'-crop','960x600+0+48','+repage','-depth','8','rgb:-']).stdout
    assert game_pixels('-HD')==game_pixels('-HD-returned')
    assert game_pixels('-HD')!=game_pixels('-original')
    click(644,24);wait('恢復遊戲',lambda s:not s['paused'],15)


try:
    x=subprocess.Popen(['Xvfb',':96','-screen','0','1280x800x24','-nolisten','tcp'],env=env,stdout=(O/'xvfb.log').open('wb'),stderr=subprocess.STDOUT);children.append(x)
    deadline=time.monotonic()+10
    while not Path('/tmp/.X11-unix/X96').exists():
        assert time.monotonic()<deadline
        time.sleep(.05)
    save=R/'workplace/orig/SAVEM1.USR'
    user=O/'user/game';user.mkdir(parents=True)
    shutil.copy2(save,user/'SAVETEST.usr')
    start()
    advance(12000000)
    for _ in range(6):key(0x39,50000);advance(7950000)
    key(0x41,500000);advance(2000000)
    key(0x15,150000);advance(1500000)
    for _ in range(8):key(0xe,100000);advance(200000)
    for scan in (0x1f,0x1e,0x2f,0x12,0x14,0x12,0x1f,0x14):key(scan,100000);advance(200000)
    key(0x1c,100000);advance(16000000)
    hd(False)
    key(0x4d,8000000);advance(250000)
    align('cave-for-wound-door',(216,102,240,118),16)
    key(0x48,250000);advance(20000000)
    hd(True)
    wait('正常進第一洞窟',lambda s:s['cave_terrain_hd_matched'],30)
    move(0x4b,lambda s:s['native_primary_0080']+s['native_primary_0083']<=41,'走到繩索')
    edge(0x50,True)
    try:
        wait('正常接敵變色HD',hit_flash_hd,30)
        click(644,24)
    finally:edge(0x50,False)
    wait('受擊變色畫面暫停',lambda s:s['paused'],15)
    compare_paused('hit-flash',hit_flash_hd)
    move(0x50,lambda s:(s['native_primary_0082']+s['native_primary_0084'])&63>=22,'沿繩索下降接敵')
    wait('正常接敵後受傷HD',injury_hd,45)
    click(644,24)
    wait('受傷畫面暫停',lambda s:s['paused'],15)
    compare_paused('injured',injury_hd)
    edge(0x1d,True);key(0x10,500000);edge(0x1d,False);advance(1000000)
    invoke(['xdotool','keydown','y'])
    try:process.wait(timeout=60)
    finally:invoke(['xdotool','keyup','y'])
    assert process.returncode==0 and h(user/'SAVETEST.usr')==h(save)
    assert (user/'SAVETEST.usr').stat().st_mtime_ns==save.stat().st_mtime_ns
    report={'status':'PASS_HD_RUNTIME_DAMAGE_GUI','version':(B/'VERSION').read_text().strip(),'binary_sha256':h(B/'zeliard-preview'),'script_sha256':h(Path(__file__)),'boots':boots,'shots':shots,'source_save_sha256':h(save),'save_unchanged':True,'normal_DOS_exit':0,'injured_hd_verified':True,'hit_flash_hd_verified':True,'scope':'正常接敵變色及倒地特殊姿態、原貌與HD同狀態切換；不代表完整通關或全遊戲HD'}
    (O/'receipt.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(report['status'],flush=True)
except Exception:
    if process is not None and process.poll() is None:
        invoke(['import','-window',window,str(O/'failed-live.png')])
        (O/'failed-state.json').write_text(json.dumps(state(),ensure_ascii=False,indent=2)+'\n')
    raise
finally:
    for child in reversed(children):stop(child)

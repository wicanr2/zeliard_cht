"""Docker／Xvfb：啟動進度與三個工具列開關的限定檢查，不推進遊戲。"""
from pathlib import Path
import argparse,hashlib,json,os,subprocess,time

parser=argparse.ArgumentParser()
parser.add_argument('--app',type=Path,default=Path('/project/workplace/runtime-controls-v1/preview'))
args=parser.parse_args()
B,O=args.app,Path('/output')
assert O.is_dir() and O.stat().st_uid==os.getuid() and not any(O.iterdir())
env={**os.environ,'DISPLAY':':97','LIBGL_ALWAYS_SOFTWARE':'1','LP_NUM_THREADS':'1','GOMAXPROCS':'1','ZELIARD_USER_DIR':str(O/'user')}
children=[];window=None;game=None
h=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()

def run(command):return subprocess.run(command,env=env,check=True,capture_output=True,timeout=8)
def stop(p):
    if p.poll() is None:
        p.terminate()
        try:p.wait(timeout=5)
        except subprocess.TimeoutExpired:p.kill();p.wait(timeout=5)

try:
    x=subprocess.Popen(['Xvfb',':97','-screen','0','1280x800x24','-nolisten','tcp'],env=env,stdout=(O/'xvfb.log').open('wb'),stderr=subprocess.STDOUT);children.append(x)
    end=time.monotonic()+10
    while not Path('/tmp/.X11-unix/X97').exists():
        assert time.monotonic()<end;time.sleep(.05)
    command=[str(B/'AppRun'),'-audio','off','-show-controller','-observe',str(O/'observe.json'),'-capture',str(O/'capture.png'),'-capture-request',str(O/'capture.request')]
    start=time.monotonic()
    game=subprocess.Popen(command,env=env,stdout=(O/'game.log').open('wb'),stderr=subprocess.STDOUT);children.append(game)
    end=time.monotonic()+90
    while time.monotonic()<end:
        assert game.poll() is None,(O/'game.log').read_text(errors='replace')[-3000:]
        r=subprocess.run(['xdotool','search','--onlyvisible','--pid',str(game.pid)],env=env,capture_output=True,timeout=3)
        if r.returncode==0 and r.stdout.strip():window=r.stdout.decode().splitlines()[-1];break
        time.sleep(.05)
    assert window is not None,'啟動視窗未出現'
    mapped_seconds=time.monotonic()-start
    run(['xdotool','windowfocus',window])

    def state():
        assert game.poll() is None,(O/'game.log').read_text(errors='replace')[-3000:]
        try:return json.loads((O/'observe.json').read_text())
        except (FileNotFoundError,json.JSONDecodeError):return {}
    def wait(condition):
        end=time.monotonic()+70
        while time.monotonic()<end:
            s=state()
            if s and condition(s):return s
            time.sleep(.05)
        raise TimeoutError('工具列狀態未更新')
    def click(x,y):run(['xdotool','mousemove','--window',window,str(x),str(y),'click','1'])

    initial=wait(lambda s:s['focused'] and s['panel'])
    ready_seconds=time.monotonic()-start
    assert initial['cheats']['lock_hp'] is False and initial['cheats']['lock_mp'] is False and initial['cheats']['standing_regeneration'] is True
    click(644,24);wait(lambda s:s['paused'])
    click(568,24);wait(lambda s:s['cheat_panel'])
    click(480,212);wait(lambda s:s['cheats']['lock_hp'])
    click(480,302);wait(lambda s:s['cheats']['lock_mp'])
    click(480,392);selected=wait(lambda s:not s['cheats']['standing_regeneration'])
    selected=wait(lambda s:s['frames']>selected['frames']+2)
    assert selected['steps']==initial['steps'] and selected['cycles']==initial['cycles']
    assert json.loads((O/'user/controller.json').read_text())['disable_standing_regen'] is True
    run(['xdotool','mousemove','--window',window,'930','640'])
    run(['import','-window',window,str(O/'cheat-settings.png')])
    (O/'selected-state.json').write_text(json.dumps(selected,ensure_ascii=False,indent=2)+'\n')
    # 保持HP鎖定，單獨關MP，證明兩個開關不互相控制。
    click(480,302);wait(lambda s:s['cheats']['lock_hp'] and not s['cheats']['lock_mp'])
    click(480,212);click(480,392)
    restored=wait(lambda s:not s['cheats']['lock_hp'] and not s['cheats']['lock_mp'] and s['cheats']['standing_regeneration'])
    assert restored['steps']==initial['steps'] and restored['cycles']==initial['cycles']
    (O/'capture.request').write_text('capture\n');game.wait(timeout=20)
    assert game.returncode==0
    report={'status':'PASS_STARTUP_AND_CHEAT_TOOLBAR_UI_ONLY','binary_sha256':h(B/'zeliard-preview'),'script_sha256':h(Path(__file__)),'command':command,'window_mapped_seconds':mapped_seconds,'settings_ready_seconds':ready_seconds,'default_cheats_off':True,'independent_toggles':True,'original_not_advanced':True,'startup_log':(O/'user/startup.log').read_text(),'selected_state':selected,'image_sha256':h(O/'cheat-settings.png'),'normal_gameplay':'NOT_RUN','audio':'off; not validated','fps_acceptance':'NOT_MEASURED'}
    (O/'receipt.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(report['status'],mapped_seconds,ready_seconds,flush=True)
finally:
    if game is not None and game.poll() is None and window is not None:
        try:run(['import','-window',window,str(O/'last-window.png')])
        except Exception:pass
    for child in reversed(children):stop(child)

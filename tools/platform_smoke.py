"""原生封包的有界啟動抽測；Windows的Wine不冒稱真機驗收。"""
from pathlib import Path
import hashlib,json,os,signal,subprocess,sys,time,zipfile

O=Path('/output');D=Path('/delivery');mode=sys.argv[1]
assert not any(O.iterdir()) and O.stat().st_uid==os.getuid()
h=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
children=[]
env={**os.environ,'DISPLAY':':93','LIBGL_ALWAYS_SOFTWARE':'1','LP_NUM_THREADS':'2','GOMAXPROCS':'2'}
def stop(p):
    if p.poll() is None:
        p.terminate()
        try:p.wait(timeout=5)
        except subprocess.TimeoutExpired:p.kill();p.wait(timeout=5)
try:
    x=subprocess.Popen(['Xvfb',':93','-screen','0','1280x800x24','-nolisten','tcp'],env=env,stdout=(O/'xvfb.log').open('wb'),stderr=subprocess.STDOUT);children.append(x)
    deadline=time.monotonic()+10
    while not Path('/tmp/.X11-unix/X93').exists():
        assert time.monotonic()<deadline
        time.sleep(.05)
    if mode=='linux':
        package=next(D.glob('*.AppImage'))
        subprocess.run([str(package),'--appimage-extract'],cwd=O,stdout=(O/'extract.log').open('wb'),stderr=subprocess.STDOUT,check=True,timeout=40)
        root=O/'squashfs-root';binary=root/'zeliard-preview'
        command=[str(root/'AppRun'),'-audio','off','-show-controller','-capture',str(O/'capture.png'),'-capture-frames','90']
        env['ZELIARD_USER_DIR']=str(O/'user')
    elif mode=='windows':
        package=next(D.glob('*Windows*.zip'))
        with zipfile.ZipFile(package) as z:
            for name in z.namelist():
                assert not name.startswith('/') and '..' not in Path(name).parts
            z.extractall(O)
        root=O/'Zeliard-Windows';binary=root/'Zeliard.exe'
        wine='/usr/lib/wine/wine64' if Path('/usr/lib/wine/wine64').exists() else 'wine'
        env.update({'WINEPREFIX':str(O/'wine-prefix'),'WINEDEBUG':'-all','WINEDLLOVERRIDES':'mscoree,mshtml=' ,'ZELIARD_USER_DIR':'Z:\\output\\user'})
        command=[wine,str(binary),'-audio','off','-show-controller','-capture','Z:\\output\\capture.png','-capture-frames','90']
    else:raise ValueError(mode)
    log=(O/'game.log').open('wb');p=subprocess.Popen(command,cwd=O,env=env,stdout=log,stderr=log);children.append(p)
    p.wait(timeout=100)
    assert p.returncode==0,(p.returncode,(O/'game.log').read_text(errors='replace')[-4000:])
    assert (O/'capture.png').is_file()
    messages=(O/'game.log').read_text(errors='replace');assert '未載入' not in messages,messages
    meta=json.loads((O/'capture.png.json').read_text());assert meta['panel']
    assert (O/'user/game/SAVETEST.usr').is_file() and (O/'user/game/RESOURCE.CFG').is_file()
    assert b'MSCADLIB.DRV' in (O/'user/game/RESOURCE.CFG').read_bytes()
    report={'status':'PASS_NATIVE_PACKAGE_LAUNCH','platform':mode,'method':'AppImage extraction-mode native AppRun' if mode=='linux' else 'Wine auxiliary startup; Windows hardware not tested','package_sha256':h(package),'binary_sha256':h(binary),'exit_code':0,'all_requested_assets_loaded':True,'standalone_asset_resolution':True,'separate_user_storage':True,'startup_audio':'off for transport smoke; full native audio GUI recorded separately','command':command}
    (O/'receipt.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(mode,report['status'],flush=True)
finally:
    for p in reversed(children):stop(p)
    if mode=='windows':subprocess.run(['wineserver','-k'],env=env,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL,timeout=10)

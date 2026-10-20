"""Docker內核對三平台封包、音源、來源與本機推廣片。"""
from pathlib import Path
import hashlib,json,plistlib,struct,zipfile
R=Path('/project');version=(R/'VERSION').read_text().strip();D=R/'dist-all'/version;W=R/'workplace/music-release-v1';B=W/'release-build'
h=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def pe(data):
    assert data[:2]==b'MZ';offset=struct.unpack_from('<I',data,0x3c)[0];assert data[offset:offset+4]==b'PE\0\0';return struct.unpack_from('<H',data,offset+4)[0]
def fat(data):
    magic,count=struct.unpack_from('>II',data);assert magic==0xcafebabe and count==2
    return {struct.unpack_from('>IIIII',data,8+i*20)[0] for i in range(count)}
linux=json.loads((D/'smoke/linux-appimage-v1/receipt.json').read_text());windows=json.loads((D/'smoke/windows-wine-v1/receipt.json').read_text())
assert linux['status']==windows['status']=='PASS_NATIVE_PACKAGE_LAUNCH'
assert linux['binary_sha256']==h(B/'linux/zeliard-preview')
assert windows['binary_sha256']==h(B/'windows/Zeliard.exe')
gui=json.loads((W/'gui-v4/receipt.json').read_text());assert gui['status']=='PASS_SCOPED_FIRST_CAVE_INPUT_CHINESE_HD' and gui['binary_sha256']==linux['binary_sha256']
audio=json.loads((D/'smoke/audio-quality.json').read_text());assert audio['status']=='PASS_NATIVE_ORIGINAL_MUSIC_EXCERPT'
assert '-realtime' in gui['boots'][0]['command'] and '-adlib' in gui['boots'][0]['command']
assert gui['boots'][0]['normal_DOS_exit'] and gui['boots'][0]['exit_code']==0
assert gui['keyboard_rebind_unbind_reset_restart'] and gui['settings_input_isolated']
assert len(gui['shots'])==9
for shot in gui['shots']:assert shot['state']['audio_output']['last_error']==''
assert pe((B/'windows/Zeliard.exe').read_bytes())==pe((B/'windows/zopl.dll').read_bytes())==0x8664
for p in [B/'macos/zeliard-preview',B/'macos/libzopl.dylib']:assert fat(p.read_bytes())=={0x1000007,0x100000c}
files=list((D/'full-local').iterdir());assert len(files)==3 and sum(p.suffix=='.AppImage' for p in files)==1
for p in files:
    if p.suffix=='.zip':
        with zipfile.ZipFile(p)as z:
            for n in z.namelist():assert not n.startswith('/') and '..' not in Path(n).parts
            base='Zeliard-Windows/' if 'Windows' in p.name else 'Zeliard-macOS/Zeliard.app/Contents/Resources/game/'
            assert json.loads(z.read(base+'bundle-options.json'))==['-adlib','-audio','ebiten','-realtime']
            assert z.read(base+'VERSION').decode().strip()==version
            for source in (R/'workplace/orig').rglob('*'):
                if source.is_file():assert z.read(base+'original/'+str(source.relative_to(R/'workplace/orig')))==source.read_bytes()
            if 'macOS' in p.name:
                info=plistlib.loads(z.read('Zeliard-macOS/Zeliard.app/Contents/Info.plist'));assert info['CFBundleGetInfoString']==version
                assert (z.getinfo('Zeliard-macOS/Zeliard.app/Contents/MacOS/zeliard-preview').external_attr>>16)&0o111
appdir=D/'smoke/linux-appimage-v1/squashfs-root'
assert json.loads((appdir/'bundle-options.json').read_text())==['-adlib','-audio','ebiten','-realtime']
for source in (B/'data').rglob('*'):
    if source.is_file():
        relative=source.relative_to(B/'data')
        if str(relative)=='runtime-args.json':assert json.loads((appdir/relative).read_text())==json.loads(source.read_text())+[['-fm-library','libzopl.so']]
        else:assert (appdir/relative).read_bytes()==source.read_bytes(),source
for source in (R/'workplace/dosgolem-source').rglob('*'):
    if source.is_file() and source.relative_to(R/'workplace/dosgolem-source')!=Path('oracle/live.go'):
        assert h(source)==h(B/'dosgolem'/source.relative_to(R/'workplace/dosgolem-source'))
identity=json.loads((B/'identity.json').read_text())
for name,digest in identity['sources'].items():assert h(B/'sources'/name)==digest and h(R/name)==digest,name
assert (B/'vet.log').read_text()=='' and 'FAIL' not in (B/'tests.log').read_text()
qa=json.loads((D/'promo/quality.json').read_text());assert qa['status']=='PASS_PROMO_TECHNICAL_QA'
artifacts={str(p.relative_to(D)):{'bytes':p.stat().st_size,'sha256':h(p)}for p in files+[D/'promo'/f'Zeliard-{version}-promo.mp4']}
(D/'SHA256SUMS.json').write_text(json.dumps({'version':version,'rights':'full-local，含原版資料與音樂，只限本機，不進Git或公開散布','artifacts':artifacts},ensure_ascii=False,indent=2)+'\n')
report={'status':'PASS_LOCAL_THREE_PLATFORM_SCOPED_RELEASE','version':version,'platforms':{'linux':'x86_64 AppImage；實際解出封包啟動與同一二進位正常第一洞窟抽測','windows':'x86_64 ZIP；Wine輔助啟動通過，未驗Windows真機','macos':'Intel／Apple Silicon universal .app ZIP；結構與雙架構已驗，未驗macOS真機'},'source_files':len(identity['sources']),'original_file_count':sum(p.is_file()for p in(R/'workplace/orig').rglob('*')),'music':'原版MSCADLIB事件＋可替換Nuked-OPL3核心；44100Hz','pacing':'使用者已選擇實際時間驅動，最多追趕100ms，暫停及失焦不追趕','physical_gamepads_tested':False,'full_game_localization_claimed':False,'promo_seconds':60,'artifacts':artifacts}
(D/'smoke/release-audit.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n');print(report['status'])

#!/usr/bin/env python3
"""051：核對六個完整包的來源、模式、平台、內容、啟動與新版影片。"""
from pathlib import Path
import argparse,hashlib,json,plistlib,struct,subprocess,tempfile,zipfile

parser=argparse.ArgumentParser();parser.add_argument('--runtime-fix',action='store_true',help='052受傷修正版，不要求重製051影片');args=parser.parse_args()
P=Path('/project');VERSION=(P/'VERSION').read_text().strip();W=P/'workplace'/('release-v'+VERSION[2:].split('-')[0].replace('.',''));D=P/'dist-all'/VERSION
sha=lambda data:hashlib.sha256(data).hexdigest()
def pe(data):
    assert data[:2]==b'MZ';at=struct.unpack_from('<I',data,0x3c)[0];assert data[at:at+4]==b'PE\0\0';assert struct.unpack_from('<H',data,at+4)[0]==0x8664
def fat(data):
    magic,count=struct.unpack_from('>II',data);assert magic==0xcafebabe and count==2
    assert {struct.unpack_from('>IIIII',data,8+i*20)[0] for i in range(2)}=={0x1000007,0x100000c}

identity=json.loads((W/'identity.json').read_text())
for row in identity['source_files']:assert sha((W/'sources'/row['file']).read_bytes())==row['sha256']
original={str(f.relative_to(P/'workplace/orig')):f.read_bytes() for f in (P/'workplace/orig').rglob('*') if f.is_file()}
reports=[];temporary=[]
for edition in ('CHT','CHT-HD'):
    for platform in ('Linux','Windows','macOS'):
        if platform=='Linux':
            evidence=D/'smoke'/('linux-'+edition);root=evidence/'squashfs-root'
            if not root.exists():
                temp=tempfile.TemporaryDirectory(prefix='zeliard-audit-');temporary.append(temp)
                package=D/'full-local'/f'Zeliard-{VERSION}-{edition}-Linux-x86_64.AppImage'
                subprocess.run([str(package),'--appimage-extract'],cwd=temp.name,check=True,stdout=subprocess.DEVNULL,stderr=subprocess.PIPE,timeout=60)
                root=Path(temp.name)/'squashfs-root'
            read=lambda name:(root/name).read_bytes();names={str(f.relative_to(root)) for f in root.rglob('*') if f.is_file()}
            launch=json.loads((evidence/'receipt.json').read_text());assert launch['status']=='PASS_NATIVE_PACKAGE_LAUNCH' and launch['edition']==edition
            assert launch['binary_sha256']==sha((W/'linux/zeliard-preview').read_bytes())
        else:
            arch='x86_64' if platform=='Windows' else 'universal'
            file=D/'full-local'/f'Zeliard-{VERSION}-{edition}-{platform}-{arch}.zip'
            z=zipfile.ZipFile(file)
            for info in z.infolist():
                assert not info.filename.startswith('/') and '..' not in Path(info.filename).parts
                if not info.filename.isascii():assert info.flag_bits&0x800
            prefix=f'Zeliard-{edition}-{platform}/'
            if platform=='macOS':prefix+=f'Zeliard-{edition}.app/Contents/Resources/game/'
            read=lambda name:z.read(prefix+name);names={name[len(prefix):] for name in z.namelist() if name.startswith(prefix) and not name.endswith('/')}
            if platform=='Windows':
                launch=json.loads((D/'smoke'/('windows-'+edition)/'receipt.json').read_text());assert launch['status']=='PASS_NATIVE_PACKAGE_LAUNCH' and launch['edition']==edition
                assert launch['binary_sha256']==sha(read('Zeliard.exe'))==sha((W/'windows/Zeliard.exe').read_bytes());pe(read('Zeliard.exe'));pe(read('zopl.dll'))
                text=read('README.txt');assert text.startswith(b'\xef\xbb\xbf') and b'\n' not in text.replace(b'\r\n',b'')
            else:
                base=f'Zeliard-{edition}-macOS/Zeliard-{edition}.app/Contents/'
                executable=z.read(base+'MacOS/zeliard-preview');fat(executable);fat(read('libzopl.dylib'));assert sha(executable)==sha((W/'macos/zeliard-preview').read_bytes())
                assert (z.getinfo(base+'MacOS/zeliard-preview').external_attr>>16)&0o111
                assert plistlib.loads(z.read(base+'Info.plist'))['CFBundleGetInfoString']==VERSION
        assert read('VERSION').decode().strip()==VERSION and read('EDITION').decode().strip()==edition
        entries=dict(json.loads(read('runtime-args.json')));options=json.loads(read('bundle-options.json'))
        assert options==['-adlib','-audio','ebiten','-realtime','-hd='+str(edition=='CHT-HD').lower()]
        assert entries['-dialogue-text']=='text/dialogue-prototype.tsv' and entries['-text']=='text/host.json'
        assert all(value in names or any(name.startswith(value+'/') for name in names) for value in entries.values())
        assert all(read('original/'+name)==data for name,data in original.items())
        assert 'font.hex.gz.smooth.json.gz' in names and 'licenses/jf-openhuninn/OFL.txt' in names
        if edition=='CHT':assert not any(name.startswith('assets/') for name in names) and '-town-signs-text' not in entries and not any('art' in k for k in entries)
        else:
            assert entries['-manual-hd-art']=='assets/manual-hd' and entries['-world-town-art']=='assets/manual-hd'
            assert sha(read('assets/manual-hd/complete-art.json'))==identity['complete_art_sha256']
        manifest=json.loads(read('bundle-manifest.json'));assert manifest['version']==VERSION and manifest['edition']==edition
        assert names=={r['file'] for r in manifest['files']}|{'bundle-manifest.json'}
        for row in manifest['files']:
            data=read(row['file']);assert len(data)==row['bytes'] and sha(data)==row['sha256'],(edition,platform,row['file'])
        reports.append({'edition':edition,'platform':platform,'files':len(names),'original_files':len(original),'mode_correct':True,'contents_and_hashes':'PASS','launch':'PASS_NATIVE' if platform=='Linux' else 'PASS_WINE_AUXILIARY' if platform=='Windows' else 'STRUCTURE_ONLY_NOT_MAC_HARDWARE'})
        if platform!='Linux':z.close()

gui=json.loads((D/'smoke/integration/receipt.json').read_text())
assert gui['binary_sha256']==sha((W/'linux/zeliard-preview').read_bytes())
if args.runtime_fix:
    assert gui['status']=='PASS_HD_RUNTIME_DAMAGE_GUI' and gui['normal_DOS_exit']==0 and gui['save_unchanged'] and gui['injured_hd_verified']
    qa={'seconds':None}
else:
    assert gui['status']=='PASS_SCOPED_FIRST_CAVE_INPUT_CHINESE_HD' and gui['boots'][0]['normal_DOS_exit']
    for shot in gui['shots']:assert shot['state']['audio_output']['last_error']==''
    qa=json.loads((D/'promo/quality.json').read_text());assert qa['status']=='PASS_PROMO_TECHNICAL_QA'
report={'status':'PASS_LOCAL_DUAL_EDITION_THREE_PLATFORM_RELEASE','version':VERSION,'packages':reports,'promo_seconds':qa['seconds'],'runtime_damage_fix':args.runtime_fix,'full_game_or_every_frame_verified':False,'physical_gamepads_tested':False,'windows_hardware_tested':False,'macos_hardware_tested':False,'rights':'full-local only; no remote publication'}
(D/'smoke/release-audit.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
print(report['status'],flush=True)
for temp in temporary:temp.cleanup()

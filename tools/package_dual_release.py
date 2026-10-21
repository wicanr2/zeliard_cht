#!/usr/bin/env python3
"""051：Docker內從固定輸入重建CHT／CHT-HD三平台本機完整包。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import plistlib
import re
import shutil
import zipfile

P=Path('/project')
VERSION=(P/'VERSION').read_text().strip()
if not re.fullmatch(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}',VERSION):raise ValueError('版號格式不符')
W=Path(os.environ.get('ZELIARD_RELEASE_WORK',str(P/'workplace'/('release-v'+VERSION[2:].split('-')[0].replace('.','')))))
D=P/'dist-all'/VERSION
OLD=P/'workplace/music-release-v1/release-build'
FONT=P/'workplace/full-hd-v1/manual-hd-v1/rounded-font'
MANUAL=P/'workplace/full-hd-v1/manual-hd-v1/registered'
EDITIONS=('CHT','CHT-HD')

def sha(file):
    h=hashlib.sha256()
    with file.open('rb') as stream:
        for chunk in iter(lambda:stream.read(1<<20),b''):h.update(chunk)
    return h.hexdigest()

def write(file,data):file.write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')

def inventory(root):
    return [{'file':str(f.relative_to(root)),'bytes':f.stat().st_size,'sha256':sha(f)} for f in sorted(root.rglob('*')) if f.is_file() and f.name!='bundle-manifest.json']

def prepare():
    if (W/'sources').exists() or D.exists():raise ValueError('拒絕覆寫既有製作或正式輸出')
    assert W.is_dir() and W.stat().st_uid==os.getuid()
    for name in ('cmd','internal','text','third_party'):shutil.copytree(P/name,W/'sources'/name)
    for name in ('go.mod','go.sum','VERSION','LICENSE'):shutil.copy2(P/name,W/'sources'/name)
    shutil.copytree(OLD/'dosgolem',W/'dosgolem')
    (W/'build.work').write_text('go 1.24.0\nuse (\n /build/sources\n /build/dosgolem\n)\n')
    for name in ('linux','windows','mac-amd64','mac-arm64','macos'):(W/name).mkdir()
    for name in ('full-local','smoke','promo'):(D/name).mkdir(parents=True,exist_ok=False)
    translations={k:v for k,v in json.loads((OLD/'data/runtime-args.json').read_text()) if v.startswith('text/') and k!='-town-signs-text'}
    sources=[]
    for edition in EDITIONS:
        root=W/('data-'+edition);root.mkdir()
        shutil.copytree(P/'workplace/orig',root/'original')
        shutil.copytree(P/'text',root/'text')
        shutil.copytree(P/'third_party',root/'third_party')
        for name in ('font.hex.gz','font.hex.gz.smooth.json.gz'):shutil.copy2(FONT/name,root/name)
        license_dir=root/'licenses/jf-openhuninn';license_dir.mkdir(parents=True)
        for name in ('OFL.txt','source.json','manifest.json'):shutil.copy2(FONT/name,license_dir/name)
        for name in ('LICENSE','THIRD_PARTY_NOTICES.md'):
            if (P/name).exists():shutil.copy2(P/name,root/name)
        shutil.copy2(OLD/'data/THIRD-PARTY-NOTICES.txt',root/'THIRD-PARTY-NOTICES.txt')
        with (root/'THIRD-PARTY-NOTICES.txt').open('a') as f:f.write('\njf open 粉圓2.1，justfont，OFL-1.1；來源、全文及轉製記錄見licenses/jf-openhuninn。\n')
        (root/'starter').mkdir();shutil.copy2(OLD/'data/starter/SAVETEST.usr',root/'starter/SAVETEST.usr')
        labels=json.loads((root/'text/host.json').read_text());labels['title']='神劍封魔：繁體中文'+(' HD' if edition=='CHT-HD' else '')
        write(root/'text/host.json',labels)
        entries=dict(translations)
        if edition=='CHT-HD':
            shutil.copytree(OLD/'data/assets',root/'assets')
            entries.update(dict(json.loads((OLD/'data/runtime-args.json').read_text())))
            for name in ('cave-unmatched-gui-v2','castle-gui-v2','interior-gui-v1','sage-frames-gui-v1','merchant-frames-gui-v4'):
                command=json.loads((P/'workplace/full-hd-v1'/name/'receipt.json').read_text())['boots'][0]['command']
                for n,value in enumerate(command):
                    if not value.startswith('/project/workplace/') or n==0:continue
                    flag=command[n-1]
                    if not flag.startswith('-') or flag in ('-orig','-font','-user-dir','-observe','-capture','-capture-request','-input-trace'):continue
                    source=P/Path(value).relative_to('/project');assert source.exists(),source
                    target=Path('assets/full-hd')/source.name;dest=root/target;dest.parent.mkdir(exist_ok=True,parents=True)
                    if not dest.exists():
                        if source.is_dir():shutil.copytree(source,dest)
                        else:shutil.copy2(source,dest)
                    entries[flag]=target.as_posix();sources.append({'flag':flag,'source':str(source.relative_to(P)),'target':target.as_posix()})
            shutil.copytree(MANUAL,root/'assets/manual-hd')
            entries.update({'-manual-hd-art':'assets/manual-hd','-world-town-art':'assets/manual-hd'})
        for flag,relative in entries.items():assert (root/relative).exists(),(flag,relative)
        write(root/'runtime-args.json',list(entries.items()))
        write(root/'bundle-options.json',['-adlib','-audio','ebiten','-realtime','-hd='+str(edition=='CHT-HD').lower()])
        (root/'VERSION').write_text(VERSION+'\n');(root/'EDITION').write_text(edition+'\n')
    write(W/'identity.json',{'version':VERSION,'source_files':inventory(W/'sources'),'dosgolem_files':inventory(W/'dosgolem'),'original_files':inventory(P/'workplace/orig'),'asset_sources':sources,'complete_art_sha256':sha(MANUAL/'complete-art.json'),'rights':'full-local only'})
    print('PREPARED',VERSION,EDITIONS,flush=True)

def archive(source,target):
    with zipfile.ZipFile(target,'x',compression=zipfile.ZIP_DEFLATED,compresslevel=6) as z:
        for file in sorted(source.rglob('*')):
            if file.is_file():z.write(file,source.name+'/'+str(file.relative_to(source)))

def setup(root,platform,edition):
    shutil.copytree(W/('data-'+edition),root)
    library={'linux':'libzopl.so','windows':'zopl.dll','macos':'libzopl.dylib'}[platform]
    write(root/'runtime-args.json',json.loads((root/'runtime-args.json').read_text())+[['-fm-library',library]])
    heading='原版＋中文＋HD' if edition=='CHT-HD' else '原版＋中文'
    text=f'''神劍封魔 Zeliard {VERSION} {heading}

包含完整原版遊戲、目前繁體中文、粉圓字型、原版AdLib音樂與Nuked-OPL3音色改善。
{'本版預設HD，可在工具列切回原貌；包含050已盤點補圖，未知來源保留原貌。' if edition=='CHT-HD' else '本版保留原版美術，不附HD圖庫；HD版為另一個獨立封包。'}

Windows開啟Zeliard.exe，macOS開啟.app，Linux執行AppImage。
工具列可設定鍵盤及USB／藍牙手把；藍牙先由作業系統配對。
方向鍵移動／跳躍／蹲下，空白鍵攻擊／交談，Enter裝備，Alt法術。
Esc原版暫停，F12鍵盤設定，F1／F2音樂／音效，Ctrl+Q再按Y退出。
可開新局，或F7讀SAVETEST，再從Muralla向右進第一洞窟。

兩版存檔與操作設定相容，保存在系統的Zeliard-CHT設定目錄：
Linux：~/.config/Zeliard-CHT
Windows：%AppData%\\Zeliard-CHT
macOS：~/Library/Application Support/Zeliard-CHT
可用ZELIARD_USER_DIR指定其他位置，安裝目錄不作正式存檔。

macOS與Windows真機、實體USB／藍牙手把未驗。未知來源安全退回原貌。
未宣稱全遊戲翻譯、每幀HD或通關驗收。封包驗收範圍見同版smoke記錄。
含原版資料與衍生美術，只留本機，不上傳或公開散布。
'''
    (root/'README.txt').write_text(text.replace('\n','\r\n') if platform=='windows' else text,encoding='utf-8-sig' if platform=='windows' else 'utf-8',newline='')

def assemble():
    for edition in EDITIONS:
        app=W/('AppDir-'+edition);setup(app,'linux',edition)
        for name in ('zeliard-preview','libzopl.so'):shutil.copy2(W/'linux'/name,app/name)
        shutil.copytree(W/'linux/lib',app/'lib')
        (app/'AppRun').write_text('#!/bin/sh\nset -eu\nhere="${APPDIR:-$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)}"\nexport LD_LIBRARY_PATH="$here/lib${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"\nexec "$here/zeliard-preview" "$@"\n')
        for name in ('AppRun','zeliard-preview'):(app/name).chmod(0o755)
        (app/'zeliard.desktop').write_text(f'[Desktop Entry]\nType=Application\nName=Zeliard 繁體中文 {edition}\nExec=zeliard-preview\nIcon=zeliard\nCategories=Game;RolePlaying;\nTerminal=false\n')
        shutil.copy2(OLD/'AppDir/zeliard.svg',app/'zeliard.svg');(app/'.DirIcon').symlink_to('zeliard.svg')
        win=W/('Zeliard-'+edition+'-Windows');setup(win,'windows',edition)
        for name in ('Zeliard.exe','zopl.dll'):shutil.copy2(W/'windows'/name,win/name)
        mac=W/('Zeliard-'+edition+'-macOS');bundle=mac/('Zeliard-'+edition+'.app');resources=bundle/'Contents/Resources/game';setup(resources,'macos',edition)
        binary=bundle/'Contents/MacOS';binary.mkdir();shutil.copy2(W/'macos/zeliard-preview',binary/'zeliard-preview');(binary/'zeliard-preview').chmod(0o755)
        shutil.copy2(W/'macos/libzopl.dylib',resources/'libzopl.dylib')
        (bundle/'Contents/Info.plist').write_bytes(plistlib.dumps({'CFBundleName':'神劍封魔 '+edition,'CFBundleDisplayName':'神劍封魔 '+edition,'CFBundleIdentifier':'tw.wicanr2.zeliard-'+edition.lower(),'CFBundleExecutable':'zeliard-preview','CFBundlePackageType':'APPL','CFBundleShortVersionString':VERSION[2:].split('-')[0],'CFBundleVersion':VERSION[2:].split('-')[0],'CFBundleGetInfoString':VERSION,'LSMinimumSystemVersion':'11.0','NSHighResolutionCapable':True}))
        shutil.copy2(resources/'README.txt',mac/'README.txt')
        for platform,root in [('Linux',app),('Windows',win),('macOS',resources)]:
            write(root/'bundle-manifest.json',{'version':VERSION,'edition':edition,'platform':platform,'rights':'full-local only','files':inventory(root)})
        archive(win,D/'full-local'/f'Zeliard-{VERSION}-{edition}-Windows-x86_64.zip')
        archive(mac,D/'full-local'/f'Zeliard-{VERSION}-{edition}-macOS-universal.zip')
    print('ASSEMBLED_TWO_EDITIONS',flush=True)

def finish():
    packages=list((D/'full-local').glob('*'));assert len(packages)==6
    artifacts={str(f.relative_to(D)):{'bytes':f.stat().st_size,'sha256':sha(f)} for f in packages+list((D/'promo').glob('*.mp4'))}
    write(D/'SHA256SUMS.json',{'version':VERSION,'editions':list(EDITIONS),'rights':'full-local only; contains original game and music','artifacts':artifacts,'inputs':json.loads((W/'identity.json').read_text()),'validation':json.loads((D/'smoke/release-audit.json').read_text())})
    print('FINISHED',len(artifacts),'artifacts',flush=True)

if __name__=='__main__':
    parser=argparse.ArgumentParser();parser.add_argument('phase',choices=['prepare','assemble','finish']);args=parser.parse_args()
    {'prepare':prepare,'assemble':assemble,'finish':finish}[args.phase]()

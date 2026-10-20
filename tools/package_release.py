"""Docker內組裝原生封包。含原版素材，只能留在本機。"""
from pathlib import Path
import hashlib,json,os,plistlib,shutil,zipfile

R=Path('/project'); W=R/'workplace/music-release-v1/release-build'
version=(R/'VERSION').read_text().strip();D=R/'dist-all'/version
assert not D.exists(),'正式輸出已存在，拒絕覆寫'
(D/'full-local').mkdir(parents=True);(D/'smoke').mkdir();(D/'promo').mkdir()
assets=W/'data'

def setup(dest,platform):
    shutil.copytree(assets,dest)
    args=json.loads((dest/'runtime-args.json').read_text())
    library={'linux':'libzopl.so','windows':'zopl.dll','macos':'libzopl.dylib'}[platform]
    args.append(['-fm-library',library]);(dest/'runtime-args.json').write_text(json.dumps(args,ensure_ascii=False,indent=2)+'\n')
    options=['-adlib','-audio','ebiten']
    decision=R/'workplace/music-release-v1/pacing-decision.json'
    if decision.exists() and json.loads(decision.read_text()).get('realtime'):
        options+=['-realtime']
    (dest/'bundle-options.json').write_text(json.dumps(options)+'\n')
    text=f'''# 神劍封魔：繁體中文與HD {version}

雙擊遊戲啟動。Linux AppImage 首次需允許執行。macOS 解壓縮後開啟 Zeliard.app。
含完整原版遊戲資料與目前完成的中文、手繪HD素材；抽樣驗收範圍是第一洞窟，未宣稱全遊戲翻譯或美術完成。

- 工具列可切換原貌／HD，並設定鍵盤與USB／藍牙手把。藍牙先在系統配對。
- 原版AdLib配樂使用Nuked-OPL3核心。F1為原版音樂開關，F2為原版音效開關。
- 方向鍵移動／跳躍／蹲下，空白鍵攻擊／交談，Enter裝備，Alt法術，Esc原版暫停，F12鍵盤設定。
- F7讀取SAVETEST後，從城鎮向右到洞口，按上進入第一洞窟。Ctrl+Q、Y正常退出。
- 存檔與設定位於系統使用者設定目錄的Zeliard-CHT。可用ZELIARD_USER_DIR指定其他目錄。
- Linux：~/.config/Zeliard-CHT；Windows：%AppData%\\Zeliard-CHT；macOS：~/Library/Application Support/Zeliard-CHT。
- 原版音源核心可替換，來源與LGPL條款在third_party/nuked-opl3，允許替換所需的除錯與逆向。

封包包含原版資料、字型與HD研究素材，只供本機使用，不得上傳或散布。
實體USB／藍牙手把與macOS真機尚未接機測試。Windows的Wine測試不等於Windows真機驗收。
'''
    (dest/'README.txt').write_text(text,encoding='utf-8-sig' if platform=='windows' else 'utf-8')

def archive(source,target):
    with zipfile.ZipFile(target,'x',compression=zipfile.ZIP_DEFLATED,compresslevel=6) as z:
        for p in sorted(source.rglob('*')):
            if p.is_file():z.write(p,source.name+'/'+str(p.relative_to(source)))

app=W/'AppDir';setup(app,'linux');shutil.copy2(W/'linux/zeliard-preview',app/'zeliard-preview');shutil.copy2(W/'linux/libzopl.so',app/'libzopl.so');shutil.copytree(W/'linux/lib',app/'lib')
(app/'AppRun').write_text('#!/bin/sh\nset -eu\nhere="${APPDIR:-$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)}"\nexport LD_LIBRARY_PATH="$here/lib${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"\nexec "$here/zeliard-preview" "$@"\n')
os.chmod(app/'AppRun',0o755);os.chmod(app/'zeliard-preview',0o755)
(app/'zeliard.desktop').write_text('[Desktop Entry]\nType=Application\nName=Zeliard 繁體中文 HD\nExec=zeliard-preview\nIcon=zeliard\nCategories=Game;RolePlaying;\nTerminal=false\n')
(app/'zeliard.svg').write_text('<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256" viewBox="0 0 256 256"><rect width="256" height="256" rx="36" fill="#243729"/><path d="M48 48h160v104q-12 43-80 72-68-29-80-72z" fill="#a48946" stroke="#e4d6a6" stroke-width="6"/><path d="m128 30 13 24v116h-26V54z" fill="#e6e8dd"/><path d="M87 155h82v13H87zM120 168h16v35h-16z" fill="#493221"/></svg>')
(app/'.DirIcon').symlink_to('zeliard.svg')

win=W/'Zeliard-Windows';setup(win,'windows');shutil.copy2(W/'windows/Zeliard.exe',win/'Zeliard.exe');shutil.copy2(W/'windows/zopl.dll',win/'zopl.dll')
archive(win,D/'full-local'/f'Zeliard-{version}-Windows-x86_64.zip')

mac=W/'Zeliard-macOS';resources=mac/'Zeliard.app/Contents/Resources/game';setup(resources,'macos');binary=mac/'Zeliard.app/Contents/MacOS';binary.mkdir(parents=True)
shutil.copy2(W/'macos/zeliard-preview',binary/'zeliard-preview');os.chmod(binary/'zeliard-preview',0o755);shutil.copy2(W/'macos/libzopl.dylib',resources/'libzopl.dylib')
info={'CFBundleName':'神劍封魔','CFBundleDisplayName':'神劍封魔 繁體中文 HD','CFBundleIdentifier':'tw.wicanr2.zeliard-cht','CFBundleExecutable':'zeliard-preview','CFBundlePackageType':'APPL','CFBundleShortVersionString':version[2:].split('-')[0],'CFBundleVersion':version[2:].split('-')[0],'NSHighResolutionCapable':True,'LSMinimumSystemVersion':'11.0','NSHumanReadableCopyright':'自有程式RRSAL-1.0；原版素材保留原權利。','CFBundleGetInfoString':version}
(mac/'Zeliard.app/Contents/Info.plist').write_bytes(plistlib.dumps(info))
shutil.copy2(resources/'README.txt',mac/'README.txt')
archive(mac,D/'full-local'/f'Zeliard-{version}-macOS-universal.zip')
print('PACKAGED_WINDOWS_MACOS_AND_LINUX_APPDIR',version)

#!/usr/bin/env python3
"""050本機Linux完整包。只準備封包及雜湊，不執行遊戲、測試或驗證腳本。"""
import argparse
import hashlib
import json
import os
import re
import shutil
from pathlib import Path


def digest(file):
    h=hashlib.sha256()
    with file.open('rb') as f:
        for chunk in iter(lambda:f.read(1<<20),b''):h.update(chunk)
    return h.hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project',type=Path,required=True);ap.add_argument('phase',choices=['prepare','finish','cleanup'])
    args=ap.parse_args();p=args.project;version=(p/'VERSION').read_text().strip()
    if not re.fullmatch(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}',version):raise ValueError('版號不符')
    work=p/'workplace/full-hd-v1/manual-hd-v1';build=work/'linux-build'/version;delivery=p/'dist-all'/version
    if args.phase=='prepare':
        if build.exists() or delivery.exists():raise ValueError('拒絕覆寫既有編譯或交付目錄')
        build.mkdir(parents=True);app=build/'AppDir'
        shutil.copytree(p/'workplace/music-release-v1/release-build/AppDir',app)
        # 從未改動的遊戲輸入重建，不沿用前版AppDir的可寫遊戲層。
        shutil.rmtree(app/'original');shutil.copytree(p/'workplace/orig',app/'original')
        shutil.copytree(p/'text',app/'text',dirs_exist_ok=True)
        font=work/'rounded-font'
        for name in ('font.hex.gz','font.hex.gz.smooth.json.gz'):
            shutil.copy2(font/name,app/name)
        font_license=app/'licenses/jf-openhuninn'
        font_license.mkdir(parents=True,exist_ok=True)
        for name in ('OFL.txt','source.json','manifest.json'):
            shutil.copy2(font/name,font_license/name)
        notices=app/'THIRD-PARTY-NOTICES.txt'
        with notices.open('a',encoding='utf-8') as stream:
            stream.write('\njf open 粉圓2.1，justfont，SIL Open Font License 1.1。\n字型轉製與灰階遮罩同採OFL-1.1，全文及來源見licenses/jf-openhuninn/。\nhttps://github.com/justfont/open-huninn-font/releases/tag/v2.1\n')
        entries=dict(json.loads((app/'runtime-args.json').read_text()));sources=[]
        receipts=['cave-unmatched-gui-v2','castle-gui-v2','interior-gui-v1','sage-frames-gui-v1','merchant-frames-gui-v4']
        for name in receipts:
            file=p/'workplace/full-hd-v1'/name/'receipt.json';command=json.loads(file.read_text())['boots'][0]['command']
            for i,value in enumerate(command):
                if i==0 or not value.startswith('/project/workplace/'):continue
                flag=command[i-1]
                if not flag.startswith('-'):continue
                source=p/Path(value).relative_to('/project')
                if not source.exists():raise FileNotFoundError(source)
                relative=Path('assets/full-hd')/source.name
                destination=app/relative;destination.parent.mkdir(parents=True,exist_ok=True)
                if not destination.exists():
                    if source.is_dir():shutil.copytree(source,destination)
                    else:shutil.copy2(source,destination)
                entries[flag]=relative.as_posix();sources.append({'flag':flag,'source':str(source.relative_to(p)),'destination':relative.as_posix()})
        candidate=Path('assets/manual-hd');shutil.copytree(work/'registered',app/candidate)
        entries['-manual-hd-art']=candidate.as_posix();entries['-world-town-art']=candidate.as_posix()
        (app/'runtime-args.json').write_text(json.dumps(list(entries.items()),ensure_ascii=False,indent=2)+'\n')
        (app/'bundle-options.json').write_text(json.dumps(['-adlib','-audio','ebiten','-realtime'])+'\n')
        (app/'VERSION').write_text(version+'\n')
        for name in ['README.md','LICENSE','THIRD_PARTY_NOTICES.md']:
            if (p/name).is_file():shutil.copy2(p/name,app/name)
        (app/'人工驗證.txt').write_text(
            f'神劍封魔 繁體中文與HD {version}\n\n'
            '包含完整本地遊戲資料、既有中文與操作設定、原版AdLib音樂音色改善及本次HD候選。\n'
            '啟動後在工具列選「HD」；鍵盤與手把設定亦在工具列。\n'
            '本次依使用者要求，不執行測試、對拍或啟動驗證。全遊戲與實體手把交人工檢視。\n'
            '新增八組敵人、十組首領、八城鎮前景、CMAN40姿態、共用圖格與17幅主要插畫候選。\n'
            '中文採用jf open 粉圓2.1，依16／18／24／32像素尺寸繪製柔邊筆畫，保留原字距。\n'
            '未匹配、被遮擋或尚未整理的片頭／結局動畫部件保留原版，不能宣稱100%全圖HD。\n'
            '儲存位置：${XDG_CONFIG_HOME:-~/.config}/Zeliard-CHT；可用ZELIARD_USER_DIR指定。\n'
            '此包含未獲公開散布授權的原版資料，僅留本機。\n',encoding='utf-8')
        shutil.copy2(app/'人工驗證.txt',app/'README.txt')
        (build/'build.work').write_text('go 1.24.0\nuse (\n /project\n /project/workplace/music-release-v1/release-build/dosgolem\n)\n')
        inputs=[]
        for root in [p/'cmd',p/'internal',p/'text',p/'tools']:
            for file in sorted(root.rglob('*')):
                if file.is_file() and file.suffix in ('.go','.json','.tsv','.py','.sh'):inputs.append({'file':str(file.relative_to(p)),'sha256':digest(file)})
        (build/'inputs.json').write_text(json.dumps({'version':version,'inputs':inputs,'asset_sources':sources,'font':json.loads((font/'manifest.json').read_text()),'tests_executed':False},ensure_ascii=False,indent=2)+'\n')
        (delivery/'full-local').mkdir(parents=True)
        print('PREPARED',app,'no tests or gameplay executed')
    elif args.phase=='finish':
        image=delivery/'full-local'/f'Zeliard-{version}-Linux-x86_64.AppImage'
        image.chmod(0o755)
        files=[f for f in (build/'AppDir').rglob('*') if f.is_file()]
        contents=[{'file':str(f.relative_to(build/'AppDir')),'bytes':f.stat().st_size,'sha256':digest(f)} for f in sorted(files)]
        manifest={'version':version,'status':'MANUAL_REVIEW_PENDING','platform':'Linux','architecture':'x86_64',
                  'rights':'full-local: contains original game, no redistribution authorization',
                  'automatic_tests':False,'launch_check':False,'gameplay_check':False,
                  'package':{'file':str(image.relative_to(delivery)),'bytes':image.stat().st_size,'sha256':digest(image)},
                  'contents':contents,'inputs':json.loads((build/'inputs.json').read_text()),
                  'limits':['新增素材未作自動測試、對拍或遊戲啟動。','未知或未匹配來源仍保留原版，未宣稱100%全圖HD。']}
        (delivery/'SHA256SUMS.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
        shutil.copy2(build/'AppDir/人工驗證.txt',delivery/'full-local/人工驗證.txt')
        print(json.dumps(manifest['package'],ensure_ascii=False))
    else:
        manifest=json.loads((delivery/'SHA256SUMS.json').read_text())
        image=delivery/manifest['package']['file']
        if not image.is_file() or image.stat().st_size!=manifest['package']['bytes']:raise ValueError('先完成交付，再清理')
        roots=[p/'workplace/gocache',build/'AppDir',work/'zeliard-preview',build/'appimage-pre-final.AppImage']
        removed=[]
        for root in roots:
            if not root.exists():continue
            files=[root] if root.is_file() else [f for f in root.rglob('*') if f.is_file() and not f.is_symlink()]
            if root.stat().st_uid!=os.getuid() or any(f.stat().st_uid!=os.getuid() for f in files):raise ValueError('清理目標擁有者不符：'+str(root))
            removed.append({'path':str(root.relative_to(p)),'logical_bytes':sum(f.stat().st_size for f in files),'allocated_bytes':sum(f.stat().st_blocks*512 for f in files),'files':len(files)})
            if root.name=='gocache':
                for child in root.iterdir():
                    if child.is_dir() and not child.is_symlink():shutil.rmtree(child)
                    else:child.unlink()
            elif root.is_dir():shutil.rmtree(root)
            else:root.unlink()
        report={'version':version,'removed':removed,'preserved':'原版、壓縮檔、手冊、字型、生成圖、提示詞、來源、存檔、歷史收據、正式交付與工具鏈均保留。'}
        (work/'cleanup.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
        print(json.dumps(report,ensure_ascii=False))


if __name__=='__main__':main()

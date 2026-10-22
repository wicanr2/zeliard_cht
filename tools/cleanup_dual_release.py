#!/usr/bin/env python3
"""051交付後只清除明列的重複建置／解壓／Wine及本專案Go快取。"""
from pathlib import Path
import argparse,hashlib,json,os,shutil

parser=argparse.ArgumentParser(description=__doc__)
mode=parser.add_mutually_exclusive_group()
mode.add_argument('--closeout',action='store_true',help='正式結案時追加清除舊封包解壓、Wine與可重建剪輯片段；保留來源及存檔')
mode.add_argument('--cpu-speed-repack',action='store_true',help='059六包內容已核對、執行驗收待低載；只移除本版重複組裝層，保留Go快取')
args=parser.parse_args()

P=Path('/project');VERSION=(P/'VERSION').read_text().strip();W=P/'workplace'/('release-v'+VERSION[2:].split('-')[0].replace('.',''));D=P/'dist-all'/VERSION
def sha(file):
    h=hashlib.sha256()
    with file.open('rb') as stream:
        for chunk in iter(lambda:stream.read(1<<20),b''):h.update(chunk)
    return h.hexdigest()
manifest=json.loads((D/'SHA256SUMS.json').read_text())
assert manifest['version']==VERSION
if args.cpu_speed_repack:
    assert manifest['validation']['status']=='PASS_PACKAGE_CONTENTS_RUNTIME_PENDING'
    assert manifest['validation']['cpu_speed_repack'] is True
else:
    assert manifest['validation']['status']=='PASS_LOCAL_DUAL_EDITION_THREE_PLATFORM_RELEASE'
for name,row in manifest['artifacts'].items():
    file=D/name;assert file.is_file() and file.stat().st_size==row['bytes'] and sha(file)==row['sha256']

roots=[]
for edition in ('CHT','CHT-HD'):
    roots += [W/('data-'+edition),W/('AppDir-'+edition),W/('Zeliard-'+edition+'-Windows'),W/('Zeliard-'+edition+'-macOS')]
    for base in (D/'smoke',W/'attempt-1/delivery/smoke'):
        roots += [base/('linux-'+edition)/'squashfs-root',base/('windows-'+edition)/('Zeliard-'+edition+'-Windows'),base/('windows-'+edition)/'wine-prefix']
attempt=W/'attempt-1'
if (attempt/'data-CHT/runtime-args.json').exists():shutil.copy2(attempt/'data-CHT/runtime-args.json',attempt/'failed-cht-runtime-args.json')
if (attempt/'delivery/full-local').exists():
    packages=list((attempt/'delivery/full-local').iterdir())
    (attempt/'failed-package-hashes.json').write_text(json.dumps([{'file':f.name,'bytes':f.stat().st_size,'sha256':sha(f)} for f in packages],indent=2)+'\n')
    roots += [f for f in packages if not f.name.endswith('-CHT-Linux-x86_64.AppImage')]
for name in ('sources','dosgolem','linux','windows','mac-amd64','mac-arm64','macos','data-CHT','data-CHT-HD','AppDir-CHT','AppDir-CHT-HD','Zeliard-CHT-Windows','Zeliard-CHT-HD-Windows','Zeliard-CHT-macOS','Zeliard-CHT-HD-macOS'):
    roots.append(attempt/name)
if not args.cpu_speed_repack:roots.append(P/'workplace/gocache')
if args.closeout:
    promotion=json.loads((W/'formal-promotion.json').read_text())
    assert promotion['status']=='PASS_UNCHANGED_RUNTIME_FORMAL_PROMOTION'
    old=P/'workplace/music-release-v1'
    # 這兩套資料、來源與平台二進位仍留在各自建置根，僅移除組裝副本。
    for release in ('release-build','release-build-fixed-clock'):
        base=old/release
        for required in ('data','sources','dosgolem','linux','windows','macos'):
            assert (base/required).is_dir(),base/required
        roots += [base/'Zeliard-Windows',base/'Zeliard-macOS']
        if release=='release-build-fixed-clock':roots.append(base/'AppDir')
    roots += [old/'pre-clock-delivery/smoke/linux-appimage-v1/squashfs-root',
              old/'pre-clock-delivery/smoke/windows-wine-v1/Zeliard-Windows',
              old/'pre-clock-delivery/smoke/windows-wine-v1/wine-prefix']
    # 原始錄影、PCM／WAV、時間線與卡片保留，只刪除已合併的無聲中間影片。
    edit=P/'workplace/ada-promo-v1/edit'
    roots += sorted(edit.glob('*.mp4'))
removed=[]
for root in roots:
    if not root.exists():continue
    assert root.lstat().st_uid==os.getuid(),root
    files=[root] if root.is_file() else list(root.rglob('*'))
    assert all(f.lstat().st_uid==os.getuid() for f in files),(root,'擁有者不符')
    removed.append({'file':str(root.relative_to(P)),'allocated_bytes':sum(f.lstat().st_blocks*512 for f in files),'entries':len(files)})
    if root.name=='gocache':
        for child in root.iterdir():
            if child.is_dir() and not child.is_symlink():shutil.rmtree(child)
            else:child.unlink()
    elif root.is_dir():shutil.rmtree(root)
    else:root.unlink()
report={'version':VERSION,'closeout':args.closeout,'cpu_speed_repack':args.cpu_speed_repack,'removed':removed,'allocated_bytes':sum(r['allocated_bytes'] for r in removed),'preserved':'六個正式包、推廣片、錄影原檔、源音訊、截圖、收據、唯一生成圖、提示詞、原版、字型、存檔及固定原始碼均保留。首批失敗日誌、參數與一份CHT AppImage亦保留。'}
(W/'cleanup.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
print('CLEANED_ALLOCATED_BYTES',report['allocated_bytes'],flush=True)

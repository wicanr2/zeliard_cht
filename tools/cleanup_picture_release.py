#!/usr/bin/env python3
"""069結案：核對1.0.8正式產物後，只清除明列的組裝、解壓、Wine、Go快取、探針截圖與探針基底。"""
from pathlib import Path
import hashlib,json,os,shutil,stat

P=Path('/project');VERSION=(P/'VERSION').read_text().strip();assert VERSION=='v.1.0.8-20261023'
DRY=os.environ.get('ZELIARD_CLEANUP_DRY')=='1'  # 只列出清單與bytes，不刪除也不寫收據
W=P/'workplace/release-v108';D=P/'dist-all'/VERSION;C=P/'workplace/picture-coverage-v1'
def sha(file):
    h=hashlib.sha256()
    with file.open('rb') as stream:
        for chunk in iter(lambda:stream.read(1<<20),b''):h.update(chunk)
    return h.hexdigest()
manifest=json.loads((D/'SHA256SUMS.json').read_text())
assert manifest['version']==VERSION and manifest['validation']['status']=='PASS_LOCAL_DUAL_EDITION_PICTURE_RELEASE'
assert len(manifest['artifacts'])==7
for name,row in manifest['artifacts'].items():
    file=D/name;assert file.is_file() and file.stat().st_size==row['bytes'] and sha(file)==row['sha256'],name

# 效能基準讀取的開發探針截圖先複製進068證據目錄，命名沿用其他證據的「輸出_檔名」。
source=C/'probe/out/g-story-long/probe.shot-468050000.indexed';kept=C/'evidence/g-story-long_probe.shot-468050000.indexed'
if source.exists():
    if not kept.exists():shutil.copy2(source,kept)
    assert sha(kept)==sha(source)
assert kept.is_file()

groups=[]
def tree(root):
    if root.exists():groups.append((root,[root]))
for release in ('release-v106','release-v107','release-v108'):
    base=P/'workplace'/release/'package'
    # 來源、執行器差異、各平台二進位與身份檔保留，只移除可由它們重組的資料及封包層。
    for required in ('identity.json','sources','dosgolem','linux','windows','macos','mac-amd64','mac-arm64'):assert (base/required).exists(),base/required
    for edition in ('CHT','CHT-HD'):
        for name in ('data-'+edition,'AppDir-'+edition,f'Zeliard-{edition}-Windows',f'Zeliard-{edition}-macOS'):tree(base/name)
for version in ('v.1.0.6-20261023','v.1.0.7-20261023',VERSION):
    smoke=P/'dist-all'/version/'smoke'
    for edition in ('CHT','CHT-HD'):
        assert (smoke/f'linux-{edition}/receipt.json').is_file() and (smoke/f'windows-{edition}/receipt.json').is_file()
        for root in (smoke/f'linux-{edition}/squashfs-root',smoke/f'windows-{edition}/Zeliard-{edition}-Windows',smoke/f'windows-{edition}/wine-prefix'):tree(root)
tree(P/'workplace/release-v106/probe-base');tree(C/'probe-base')
# 截圖與記憶體快照可依保存的鍵序重跑；jsonl、日誌與探針存檔保留。1.0.8推廣片來源rel-*整組保留。
SHOTS={'.png','.indexed','.mem','.memory'}
for out in (P/'workplace/release-v106/probe/out',P/'workplace/release-v107/probe/out',P/'workplace/release-v107/probe/out-attempt-1',C/'probe/out'):
    for run in sorted(r for r in out.iterdir() if r.is_dir()):
        if out==C/'probe/out' and run.name.startswith('rel-'):continue
        files=sorted(f for f in run.rglob('*') if f.is_file() and not f.is_symlink() and f.suffix in SHOTS)
        if files:groups.append((run,files))
cache=P/'workplace/gocache'
if cache.is_dir():groups.append((cache,list(cache.iterdir())))

removed=[]
for label,paths in groups:
    entries=[]
    for path in paths:entries += [path] if not path.is_dir() or path.is_symlink() else [path,*path.rglob('*')]
    assert all(e.lstat().st_uid==os.getuid() for e in entries),(label,'擁有者不符')
    removed.append({'file':str(label.relative_to(P)),'scope':'tree' if paths==[label] else ('contents' if label==cache else 'shots'),'allocated_bytes':sum(e.lstat().st_blocks*512 for e in entries),'entries':len(entries)})
    for path in [] if DRY else paths:
        if path.is_dir() and not path.is_symlink():
            # 探針基底凍結時移除了寫入權限；只在要刪除的目錄樹內恢復目錄寫入。
            for e in (path,*path.rglob('*')):
                if e.is_dir() and not e.is_symlink() and not e.stat().st_mode&stat.S_IWUSR:e.chmod(e.stat().st_mode|stat.S_IWUSR)
            shutil.rmtree(path)
        else:path.unlink()
report={'version':VERSION,'removed':removed,'allocated_bytes':sum(r['allocated_bytes'] for r in removed),
        'benchmark_dump':{'file':str(kept.relative_to(P)),'sha256':sha(kept)},
        'preserved':'六個正式包與推廣片、各版正式交付與啟動收據、來源副本、獨立dosgolem、各平台二進位、identity.json、推廣片來源畫格與rel-*探針、探針jsonl／日誌／存檔、068實機截圖證據、唯一生成圖、參考圖、提示詞、字模、原版與存檔均保留。'}
if DRY:
    for row in removed:print(row['scope'],row['allocated_bytes'],row['entries'],row['file'])
    print('DRY_RUN_ALLOCATED_BYTES',report['allocated_bytes'],flush=True)
else:
    (W/'cleanup.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print('CLEANED_ALLOCATED_BYTES',report['allocated_bytes'],flush=True)

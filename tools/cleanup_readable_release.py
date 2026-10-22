"""065：七份正式成品核SHA後，只刪本輪明列的重複組裝及剪輯層。"""
from pathlib import Path
import hashlib,json,os,shutil,subprocess

P=Path('/project');V=(P/'VERSION').read_text().strip();W=P/'workplace/release-v105';D=P/'dist-all'/V
assert V=='v.1.0.5-20261022'
def sha(p):
    h=hashlib.sha256()
    with p.open('rb')as f:
        for b in iter(lambda:f.read(1<<20),b''):h.update(b)
    return h.hexdigest()
def check():
    m=json.loads((D/'SHA256SUMS.json').read_text());assert len(m['artifacts'])==7
    for name,row in m['artifacts'].items():assert (D/name).stat().st_size==row['bytes'] and sha(D/name)==row['sha256']
    return m

assert json.loads((W/'repack-receipt.json').read_text())['status']=='PASS_V105_LOCAL_READABLE_RELEASE_AND_PROMO'
check()
subprocess.run(['python3',str(P/'tools/cleanup_dual_release.py'),'--contents-only'],check=True,timeout=120)
removed=[]
roots=sorted((W/'edit').glob('*.mp4'))+sorted((W/'edit-attempt-1').glob('*.mp4'))
roots += [D/'smoke/windows-CHT-attempt-1'/name for name in ('Zeliard-CHT-Windows','wine-prefix')]
for root in roots:
    if not root.exists():continue
    assert root.lstat().st_uid==os.getuid(),root
    entries=[root]+list(root.rglob('*'))if root.is_dir()else[root]
    assert all(f.lstat().st_uid==os.getuid()for f in entries)
    removed.append({'path':str(root.relative_to(P)),'allocated_bytes':sum(f.lstat().st_blocks*512 for f in entries)})
    if root.is_dir():shutil.rmtree(root)
    else:root.unlink()
check()
prior=json.loads((W/'cleanup.json').read_text())
report={'status':'PASS_V105_REGISTERED_DELIVERY_CLEANUP','version':V,'package_cleanup_allocated_bytes':prior['allocated_bytes'],'additional_removed':removed,'additional_allocated_bytes':sum(r['allocated_bytes']for r in removed),'preserved':'七個正式成品、原始新舊錄影、前版、固定來源、編譯二進位、字型、圖庫、存檔、診斷日誌、工作工具及Go快取'}
report['total_allocated_bytes']=report['package_cleanup_allocated_bytes']+report['additional_allocated_bytes']
(W/'cleanup-media.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
print(report['status'],report['total_allocated_bytes'],flush=True)

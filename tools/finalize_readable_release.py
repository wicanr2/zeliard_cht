"""065：品質與來源均通過後，登錄六包、影片與可重生工具身份。"""
from pathlib import Path
import hashlib,json,os,shutil,subprocess

P=Path('/project');V=(P/'VERSION').read_text().strip();W=P/'workplace/release-v105';D=P/'dist-all'/V
assert V=='v.1.0.5-20261022'
def sha(p):
    h=hashlib.sha256()
    with p.open('rb')as f:
        for b in iter(lambda:f.read(1<<20),b''):h.update(b)
    return h.hexdigest()
def read(p):return json.loads(p.read_text())
def write(p,data):p.write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')

assert not (D/'SHA256SUMS.json').exists()
audit=read(D/'smoke/release-audit.json');assert audit['status']=='PASS_LOCAL_DUAL_EDITION_SCOPED_SETTINGS_RELEASE'
quality=read(D/'promo/quality.json');source=read(D/'promo/source-review.json');visual=read(D/'promo/visual-review.json')
assert quality['status']=='PASS_PROMO_TECHNICAL_QA'
assert source['status']=='PASS_READABLE_PROMO_SOURCES_AND_REUSED_HD_PROOF'
assert visual['status']=='PASS_SCOPED_VISUAL_REVIEW'
assert quality['sha256']==source['video_sha256']==sha(D/'promo'/f'Zeliard-{V}-promo.mp4')
capture=read(D/'smoke/toolbar/receipt.json');assert capture['version']==V
assert capture['binary_sha256']==sha(W/'linux/zeliard-preview')
assert sha(P/'tools/ui_text_smoke.py')==capture['script_sha256']==sha(W/'recording/ui_text_smoke.py')
request=read(W/'request.json')
assert sha(P/'dist-all/v.1.0.4-20261022/SHA256SUMS.json')==request['old_manifest_sha256']
assert sha(P/'DOS 神劍封魔 Zeliard Remastered.txt')==request['user_txt_sha256']
for root in ('cmd','internal','text','third_party'):
    for file in (P/root).rglob('*'):
        if file.is_file():assert sha(file)==sha(W/'sources'/file.relative_to(P)),file
subprocess.run(['python3',str(P/'tools/package_dual_release.py'),'finish'],check=True,timeout=90)
manifest=read(D/'SHA256SUMS.json');assert len(manifest['artifacts'])==7
manifest['media_validation']={'quality':quality,'sources_and_reused_hd':source,'visual':visual,'gameplay_source_versions':['v.1.0.2-20261021','v.1.0.4-20261022'],'new_settings_source':V,'not_new_gameplay_or_30fps_acceptance':True}
manifest['media_provenance']={'file':'promo/provenance.json','sha256':sha(D/'promo/provenance.json')}
write(D/'SHA256SUMS.json',manifest)
write(D/'promo/media-manifest.json',{'version':V,'rights':'local only; original game footage/music; no upload','files':{f.name:{'bytes':f.stat().st_size,'sha256':sha(f)}for f in sorted((D/'promo').iterdir())if f.is_file()and f.name!='media-manifest.json'}})
shutil.copy2(Path(__file__),W/'build-tools'/Path(__file__).name)
tools={f.name:sha(f)for f in sorted((W/'build-tools').iterdir())if f.is_file()}
report={'status':'PASS_V105_LOCAL_READABLE_RELEASE_AND_PROMO','version':V,'artifacts':manifest['artifacts'],'tools':tools,'source_identity_sha256':sha(W/'identity.json'),'linux_binary_sha256':sha(W/'linux/zeliard-preview'),'windows_binary_sha256':sha(W/'windows/Zeliard.exe'),'macos_universal_binary_sha256':sha(W/'macos/zeliard-preview'),'formal_settings_shots':len(capture['shots']),'settings_original_steps':capture['dos_steps'],'settings_original_cycles':capture['dos_cycles'],'old_manifest_unchanged':True,'user_txt_unchanged':True,'rights':'full-local only; sources/docs only may sync Git','fps_already_accepted':True,'full_game_or_hardware_gamepad_tested':False}
write(W/'repack-receipt.json',report)
print(report['status'],flush=True)

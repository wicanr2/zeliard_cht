"""065：核新錄字級與具名舊來源，逐幀核重用HD區間未被改畫或替換。"""
from pathlib import Path
import hashlib,json,subprocess
from PIL import Image,ImageChops,ImageStat

P=Path('/project');V=(P/'VERSION').read_text().strip();D=P/'dist-all'/V/'promo';W=P/'workplace/release-v105'
h=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
meta=json.loads((D/'provenance.json').read_text());quality=json.loads((D/'quality.json').read_text())
assert quality['status']=='PASS_PROMO_TECHNICAL_QA'
film=D/f'Zeliard-{V}-promo.mp4';assert h(film)==meta['video_sha256']==quality['sha256']
capture=json.loads((D.parent/'smoke/toolbar/receipt.json').read_text())
assert capture['version']==V and capture['status']=='PASS_SCOPED_UI_TEXT_SETTINGS'
assert capture['binary_sha256']==h(W/'linux/zeliard-preview')==meta['source_binary_sha256']
assert capture['dos_steps']==0 and capture['dos_cycles']==0 and capture['game_save_unchanged']
assert {s['size']for s in capture['shots']}=={16,24,32}
old=P/meta['old_promo'];assert h(old)==meta['old_promo_sha256']
proof=old.parent/'hd-stability-proof.json'
assert h(proof)==meta['old_proof_receipt_sha256'] and json.loads(proof.read_text())['status']=='PASS_SCOPED_CONTINUOUS_HD_VIDEO_PROOF'
raw=Path(capture['recording']);assert h(raw)==meta['new_recording_sha256']==capture['video_sha256']
font_segments=[s for s in meta['segments']if s['caption']=='字級自由調整']
assert len(font_segments)==3
for size,segment in zip((16,24,32),font_segments):
    shot=next(s for s in capture['shots']if s['name']=='text-size-'+str(size))
    assert shot['size']==size and segment['seconds']==4
    assert abs(segment['source_seek_seconds']-(shot['wall_seconds']-meta['new_source_timing_delay']+.1))<.001

def run(args):return subprocess.run(args,check=True,capture_output=True,timeout=90)
def frames(file,start,seconds,crop=None,cfr=False):
    vf=('crop='+crop+',' if crop else '')+'scale=240:170,format=gray'
    cadence=['-r','30','-fps_mode','cfr'] if cfr else ['-fps_mode','passthrough']
    data=run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-threads','1','-ss',str(start),'-i',str(file),'-t',str(seconds),'-filter_threads','1','-vf',vf]+cadence+['-f','rawvideo','-pix_fmt','gray','-']).stdout
    n=240*170;assert data and len(data)%n==0
    return [data[i:i+n]for i in range(0,len(data),n)]

reports=[];timeline=0
for segment in meta['segments']:
    length=segment['seconds']
    if segment.get('proof_segment') or segment.get('source_version')==V:
        is_old=bool(segment.get('proof_segment'))
        source=old if is_old else raw
        # 新錄X11時間戳依製片相同的-r30常數編碼節奏重取樣；舊證明仍逐實際畫格核對。
        before=frames(source,segment['source_seek_seconds'],length,None if is_old else '960:680:0:0',cfr=not is_old)
        after=frames(film,timeline,length,None if is_old else '960:680:16:20')
        assert len(before)==len(after)==length*30,(segment['caption'],len(before),len(after))
        errors=[ImageStat.Stat(ImageChops.difference(Image.frombytes('L',(240,170),a),Image.frombytes('L',(240,170),b))).mean[0]for a,b in zip(before,after)]
        assert max(errors)<8,(segment['caption'],max(errors))
        if is_old:
            selected=run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-threads','1','-ss',str(timeline),'-i',str(film),'-t',str(length),'-filter_threads','1','-vf','crop=4:4:138:30,scale=1:1:flags=neighbor,format=rgb24','-fps_mode','passthrough','-f','rawvideo','-pix_fmt','rgb24','-']).stdout
            assert len(selected)//3==len(after)
            assert all(selected[i+1]>72 and selected[i+2]>95 for i in range(0,len(selected),3))
        reports.append({'caption':segment['caption'],'film_start':timeline,'seconds':length,'source_version':segment['source_version'],'decoded_frames':len(after),'source_comparison_cadence':'passthrough' if is_old else 'same FFmpeg -r30 CFR as production; not game FPS','source_comparison_max_mean_luma_error':max(errors),'unchanged_recorded_window':True,'reused_hd_proof':is_old,'hd_selected_all_frames':True if is_old else None})
    timeline+=length
assert timeline==60 and sum(r['decoded_frames']for r in reports if r['reused_hd_proof'])==540
report={'status':'PASS_READABLE_PROMO_SOURCES_AND_REUSED_HD_PROOF','version':V,'video_sha256':h(film),'segments':reports,'text_sizes_shown':[16,24,32],'each_text_size_seconds':4,'source_versions_shown':True,'new_settings_version':V,'proof_gameplay_version':'v.1.0.4-20261022','proof_scope':'重用且明列舊版的對話及平台，各自連續；不冒稱1.0.5遊玩實錄','physical_gamepad_or_full_game_or_30fps_claim':False,'music':'original title background music; not synchronous sound'}
(D/'source-review.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
print(report['status'],flush=True)

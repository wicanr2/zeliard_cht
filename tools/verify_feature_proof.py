"""063：核對連續區間、去重來源樣本與影片像素，限定居民／平台範圍。"""
from pathlib import Path
import hashlib,json,re,subprocess
from PIL import Image,ImageChops,ImageStat

P=Path('/project');V=(P/'VERSION').read_text().strip();W=P/'workplace/promo-features-v104';D=P/'dist-all'/V/'promo'
meta=json.loads((D/'provenance.json').read_text());capture=json.loads((W/'recording/receipt.json').read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def run(args):return subprocess.run(args,check=True,capture_output=True,timeout=100)
masks={}

def grid_ratio(raw,size):
    width,height=size
    if size not in masks:
        hi=Image.frombytes('L',(width-1,height),bytes(255 if x%3!=2 else 0 for y in range(height)for x in range(width-1)))
        vi=Image.frombytes('L',(width,height-1),bytes(255 if y%3!=2 else 0 for y in range(height-1)for x in range(width)))
        masks[size]=(hi,ImageChops.invert(hi),vi,ImageChops.invert(vi))
    im=Image.frombytes('L',size,raw)
    horizontal=ImageChops.difference(im.crop((1,0,width,height)),im.crop((0,0,width-1,height)))
    vertical=ImageChops.difference(im.crop((0,1,width,height)),im.crop((0,0,width,height-1)))
    hi,hb,vi,vb=masks[size]
    a,b,c,d=[ImageStat.Stat(image,mask)for image,mask in ((horizontal,hi),(horizontal,hb),(vertical,vi),(vertical,vb))]
    inside=(a.sum[0]+c.sum[0])/(a.count[0]+c.count[0])
    boundary=(b.sum[0]+d.sum[0])/(b.count[0]+d.count[0])
    return inside/max(boundary,.01)

def pixels(file,region,seek=None,seconds=None):
    x,y,width,height=region
    args=['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-threads','1']
    if seek is not None:args+=['-ss',str(seek)]
    args+=['-i',str(file)]
    if seconds is not None:args+=['-t',str(seconds)]
    args+=['-vf',f'crop={width}:{height}:{x}:{y},format=gray','-filter_threads','1','-fps_mode','passthrough','-f','rawvideo','-pix_fmt','gray','-']
    raw=run(args).stdout;frame=width*height
    assert len(raw)%frame==0 and raw
    return [raw[i:i+frame]for i in range(0,len(raw),frame)]

def selected_hd_frames(file,seconds=None,seek=None,offset=(0,0)):
    args=['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-threads','1']
    if seek is not None:args+=['-ss',str(seek)]
    args+=['-i',str(file)]
    if seconds is not None:args+=['-t',str(seconds)]
    args+=['-filter_threads','1','-vf',f'crop=4:4:{122+offset[0]}:{10+offset[1]},scale=1:1:flags=neighbor,format=rgb24','-fps_mode','passthrough','-f','rawvideo','-pix_fmt','rgb24','-']
    raw=run(args).stdout;assert len(raw)%3==0
    values=[raw[i+1]>72 and raw[i+2]>95 for i in range(0,len(raw),3)]
    assert values and all(values),'實際影片內HD按鈕未保持選取'
    return len(values)

regions={'platform':(144,90,672,432),'dialogue':(144,384,672,132)}
reference={}
for label,event,offset in [('native','platform-original-reference-begin',.25),('hd','platform-original-reference-end',1)]:
    start=next(m['wall_seconds']for m in capture['markers']if m['name']==event)-meta['new_source_timing_delay']+offset
    frames=pixels(W/'recording/feature-source.mp4',regions['platform'],start,.4)
    values=[grid_ratio(f,(672,432))for f in frames]
    reference[label]={'minimum':min(values),'maximum':max(values),'source_seek_seconds':start,'frames':len(values)}
assert reference['hd']['minimum']>reference['native']['maximum']*2+.05,'已知原貌與HD參考未分離'
threshold=(reference['native']['maximum']+reference['hd']['minimum'])/2
reports=[]
for item in meta['proof_clips']:
    file=D/item['file'];assert sha(file)==item['sha256']
    assert item['speed']==1 and not item['internal_cuts'] and not item['manual_hd_toggle'] and item['full_window_preserved']
    probe=json.loads(run(['ffprobe','-v','error','-show_streams','-show_format','-of','json',str(file)]).stdout)
    video=next(x for x in probe['streams']if x['codec_type']=='video');audio=next(x for x in probe['streams']if x['codec_type']=='audio')
    assert (video['width'],video['height'])==(960,680) and audio['channels']==2
    assert abs(float(video['duration'])-item['seconds'])<.15 and abs(float(audio['duration'])-item['seconds'])<.15
    key=item['label']+'-proof';samples=[s for s in capture['proof_samples']if s['label']==key and s['frames']in item['source_sample_frames']]
    assert len(samples)==item['sample_count'] and len({s['frames']for s in samples})==len(samples)
    prefix='town'if item['label']=='dialogue'else'cave'
    assert all(s['hd_enabled'] and s[prefix+'_hd'] and s[prefix+'_source_used'] and not s['paused'] and not s['panel']for s in samples)
    if prefix=='town':assert all(s['dialogue_key'] and s['dialogue_backdrop']for s in samples)
    else:assert len({s['map_origin']for s in samples})>=3,'平台未移動'
    region=regions[item['label']];frames=pixels(file,region);size=region[2:]
    assert len(frames)==int(video['nb_frames']),'解碼畫格數與實際編碼不同'
    ratios=[grid_ratio(f,size)for f in frames]
    button_frames=selected_hd_frames(file)
    assert button_frames==len(frames)
    low=[i for i,r in enumerate(ratios)if r<threshold]
    signal=run(['ffmpeg','-nostdin','-hide_banner','-threads','1','-i',str(file),'-filter_threads','1','-vf','blackdetect=d=0.5:pic_th=0.98:pix_th=0.10','-af','volumedetect','-threads','1','-f','null','-']).stderr.decode()
    mean=float(re.search(r'mean_volume: ([\-\d.]+) dB',signal).group(1));peak=float(re.search(r'max_volume: ([\-\d.]+) dB',signal).group(1))
    assert -40<mean<-8 and peak<-.1 and 'black_start:'not in signal
    (D/(item['label']+'-proof-signal.log')).write_text(signal)
    report={'file':item['file'],'label':item['label'],'seconds':float(video['duration']),'video_frames':len(frames),'distinct_region_frames':len({hashlib.sha256(f).digest()for f in frames}),'source_samples':len(samples),'source_fallback_samples':0,'required_luma_grid_ratio':threshold,'minimum_luma_grid_ratio':min(ratios),'mean_luma_grid_ratio':sum(ratios)/len(ratios),'original_like_frame_indices':low,'no_internal_cuts':True,'manual_hd_switches':0,'hp_lock_active':item['hp_lock_active'],'audio':item['audio'],'mean_db':mean,'peak_db':peak,'long_black_frames':False,'sha256':item['sha256']}
    report['actual_video_hd_button_selected_frames']=button_frames
    reports.append(report)
    (D/(item['label']+'-proof-ffprobe.json')).write_text(json.dumps(probe,indent=2)+'\n')

main_reports=[];timeline=0
for segment in meta['segments']:
    if segment.get('proof_segment'):
        label='dialogue'if 'dialogue' in Path(segment['file']).name else 'platform'
        item=next(p for p in meta['proof_clips']if p['label']==label)
        assert segment['source_version']==V and segment['speed']==1 and not segment['internal_cuts']
        assert segment['source_seek_seconds']>=item['source_seek_seconds'] and segment['source_seek_seconds']+segment['seconds']<=item['source_seek_seconds']+item['seconds']
        x,y,width,height=regions[label];region=(x+16,y+20,width,height)
        file=D/f'Zeliard-{V}-promo.mp4'
        values=[grid_ratio(f,(width,height))for f in pixels(file,region,timeline,segment['seconds'])]
        selected=selected_hd_frames(file,segment['seconds'],timeline,(16,20))
        assert len(values)==selected
        main_reports.append({'label':label,'promo_begin_seconds':timeline,'seconds':segment['seconds'],'video_frames':len(values),'actual_video_hd_button_selected_frames':selected,'minimum_luma_grid_ratio':min(values),'required_luma_grid_ratio':threshold,'original_like_frame_indices':[i for i,r in enumerate(values)if r<threshold],'internal_cuts':False})
    timeline+=segment['seconds']
assert len(main_reports)==2 and all(not r['original_like_frame_indices']for r in main_reports)
result={'status':'COLLECTED_HD_PROOF_PIXEL_AND_SOURCE_EVIDENCE','version':V,'proofs':reports,'main_promo_proof_segments':main_reports,'method':'每個解碼畫格比較原生3×3格內與跨格邊界的亮度梯度；以同狀態原貌及HD參考區間中點為門檻。居民取文字框外下方畫帶，平台取完整viewport，不含工具列或中文字幕。另逐幀核對實際影片的HD選取色。','limits':'只證明這兩段錄影；亮度格線為本場景的輔助分類，不當作通用HD偵測或全遊戲驗收。','calibration':reference,'source_video_sha256':meta['new_recording_sha256'],'capture_tool_sha256':capture['script_sha256'],'verification_tool_sha256':sha(Path(__file__))}
(D/'hd-stability-proof.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
assert all(not r['original_like_frame_indices']for r in reports),reports
assert all(r['distinct_region_frames']>=3 for r in reports),'證明區域缺少實際變化'
result['status']='PASS_SCOPED_CONTINUOUS_HD_VIDEO_PROOF'
(D/'hd-stability-proof.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print(result['status'],[(r['label'],r['video_frames'],r['source_samples'])for r in reports],flush=True)

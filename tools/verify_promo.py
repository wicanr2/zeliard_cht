"""核對本機推廣片的串流、長度、聲音、黑幀、凍結與抽樣圖。"""
from pathlib import Path
import hashlib,json,re,subprocess
R=Path('/project');version=(R/'VERSION').read_text().strip();O=R/'dist-all'/version/'promo'
video=O/f'Zeliard-{version}-promo.mp4';meta=json.loads((O/'provenance.json').read_text())
def run(args):return subprocess.run(args,check=True,capture_output=True,text=True,timeout=90)
probe=json.loads(run(['ffprobe','-v','error','-show_streams','-show_format','-of','json',str(video)]).stdout)
(O/'ffprobe.json').write_text(json.dumps(probe,indent=2)+'\n')
v=next(s for s in probe['streams'] if s['codec_type']=='video');a=next(s for s in probe['streams'] if s['codec_type']=='audio')
assert (v['width'],v['height'])==(1280,720) and a['channels']==2
assert abs(float(v['duration'])-60)<.1 and abs(float(a['duration'])-60)<.1
result=run(['ffmpeg','-hide_banner','-i',str(video),'-vf','blackdetect=d=0.5:pic_th=0.98:pix_th=0.10,freezedetect=n=-50dB:d=2','-af','silencedetect=n=-50dB:d=2,volumedetect','-threads','2','-f','null','-'])
log=result.stderr;(O/'signal-check.log').write_text(log)
peak=float(re.search(r'max_volume: ([\-\d.]+) dB',log).group(1));mean=float(re.search(r'mean_volume: ([\-\d.]+) dB',log).group(1))
assert -40<mean<-8 and peak<-.1,(mean,peak)
assert 'black_start:' not in log,'出現連續黑幀'
assert 'silence_start:' not in log,'出現兩秒以上靜音'
stills=[];time=0
for s in meta['segments']:
    if s['kind']=='intentional still':stills.append((time,time+s['seconds']))
    time+=s['seconds']
freezes=[];start=None
for line in log.splitlines():
    if 'freeze_start:' in line:start=float(line.split('freeze_start:')[-1])
    if 'freeze_end:' in line and start is not None:
        end=float(line.split('freeze_end:')[-1]);overlap=sum(max(0,min(end,b)-max(start,a))for a,b in stills)
        freezes.append({'start':start,'end':end,'planned_still_overlap':overlap,'unplanned_seconds':max(0,end-start-overlap)});start=None
if start is not None:
    end=60;overlap=sum(max(0,min(end,b)-max(start,a))for a,b in stills)
    freezes.append({'start':start,'end':end,'planned_still_overlap':overlap,'unplanned_seconds':max(0,end-start-overlap)})
assert all(f['unplanned_seconds']<=3.5 for f in freezes),freezes
run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-i',str(video),'-vf','fps=1/10,scale=426:240,tile=3x2','-frames:v','1',str(O/'contact-sheet.jpg')])
report={'status':'PASS_PROMO_TECHNICAL_QA','seconds':60,'resolution':[1280,720],'video_codec':v['codec_name'],'audio_codec':a['codec_name'],'audio_channels':2,'mean_db':mean,'peak_db':peak,'long_silence':False,'long_black_frames':False,'freeze_intervals':freezes,'planned_still_ranges':stills,'subtitles_bounds_checked':meta['subtitle_bounds_checked'],'sha256':hashlib.sha256(video.read_bytes()).hexdigest(),'visual_review':'contact-sheet.jpg；目視審查另記，不以技術通過代替'}
(O/'quality.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n');print(report['status'])

# 053另交付連續遊玩片段。場景等待、對話及正常暫停如實保留，不套用推廣片的靜音刪除標準。
if 'gameplay_clip' in meta:
    item=meta['gameplay_clip'];file=O/item['file']
    assert hashlib.sha256(file.read_bytes()).hexdigest()==item['sha256']
    gp=json.loads(run(['ffprobe','-v','error','-show_streams','-show_format','-of','json',str(file)]).stdout)
    gv=next(s for s in gp['streams'] if s['codec_type']=='video');ga=next(s for s in gp['streams'] if s['codec_type']=='audio')
    assert (gv['width'],gv['height'])==(960,680) and ga['channels']==2
    numerator,denominator=map(int,gv['avg_frame_rate'].split('/'));assert numerator/denominator==60
    assert abs(float(gv['duration'])-item['seconds'])<.15 and abs(float(ga['duration'])-item['seconds'])<.15
    glog=run(['ffmpeg','-hide_banner','-i',str(file),'-vf','blackdetect=d=1:pic_th=0.98:pix_th=0.10,freezedetect=n=-50dB:d=2','-af','silencedetect=n=-50dB:d=2,volumedetect','-threads','2','-f','null','-']).stderr
    (O/'gameplay-signal-check.log').write_text(glog)
    gm=float(re.search(r'mean_volume: ([\-\d.]+) dB',glog).group(1));gpeak=float(re.search(r'max_volume: ([\-\d.]+) dB',glog).group(1))
    assert -65<gm<0 and 'black_start:' not in glog,(gm,gpeak)
    gq={'status':'PASS_ADA_CONTINUOUS_GAMEPLAY_AV' if meta['capture_machine']=='ada' else 'PASS_LOCAL_CONTINUOUS_GAMEPLAY_AV','seconds':float(gv['duration']),'audio_seconds':float(ga['duration']),'resolution':[960,680],'fps':60,'mean_db':gm,'peak_db':gpeak,'long_black_frames':False,'source_audio':meta['capture_machine']+' PulseAudio monitor','audio_alignment':item['alignment'],'preserved_native_scene_pauses':'未刪除正常對話、暫停與換場，原始訊號日誌保留靜音及停頓位置','sha256':item['sha256']}
    (O/'gameplay-quality.json').write_text(json.dumps(gq,ensure_ascii=False,indent=2)+'\n');print(gq['status'])

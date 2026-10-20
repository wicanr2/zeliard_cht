"""Docker內以真實遊戲錄影及原版AdLib配樂製作60秒本機推廣片。"""
from pathlib import Path
import array,hashlib,json,subprocess,sys,wave

R=Path('/project');W=R/'workplace/music-release-v1';version=(R/'VERSION').read_text().strip()
S=Path(sys.argv[1]);recording=Path(sys.argv[2]);O=R/'dist-all'/version/'promo'
O.mkdir(exist_ok=True);T=W/'promo-edit';T.mkdir(exist_ok=True);assert not any(T.iterdir())
receipt=json.loads((S/'receipt.json').read_text());assert receipt['status']=='PASS_SCOPED_FIRST_CAVE_INPUT_CHINESE_HD'
shots={s['name']:s for s in receipt['shots']};source=recording/'gameplay.mp4'
font='/usr/share/fonts/opentype/noto/NotoSerifCJK-Bold.ttc'
body='/usr/share/fonts/opentype/noto/NotoSerifCJK-Regular.ttc'
assert Path(font).is_file() and Path(body).is_file()
h=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def run(args):return subprocess.run(args,check=True,capture_output=True,timeout=90)
def photo(name):return S/('boot'+str(shots[name]['boot']))/(name+'.png')
def textfits(text,size,maxwidth):
    # 保守以每字一個全形寬度計算，字幕預留至少24像素邊距。
    assert max(len(s) for s in text.split('\n'))*size<=maxwidth

def card(name,title,subtitle,shot):
    textfits(title,68,1184);textfits(subtitle,30,1184)
    out=T/(name+'.png')
    run(['convert',str(photo(shot)),'-crop','960x600+0+48','+repage','-resize','1280x800^','-gravity','center','-extent','1280x720','-fill','#102016bb','-draw','rectangle 0,0 1279,719','-stroke','#b3a06e','-strokewidth','2','-fill','none','-draw','rectangle 38,38 1241,681','-stroke','none','-font',font,'-fill','#f0e3be','-pointsize','68','-gravity','center','-annotate','+0-42',title,'-font',body,'-pointsize','30','-annotate','+0+54',subtitle,str(out)])
    return out

def panel(name,shot,title,lines):
    textfits(title,28,280);textfits(lines,24,280)
    out=T/(name+'.png')
    run(['convert','-size','1280x720','xc:#19251e',str(photo(shot)),'-gravity','northwest','-geometry','+0+20','-composite','-fill','#b3a06e','-draw','rectangle 969,52 972,668','-font',font,'-fill','#f0e3be','-pointsize','28','-annotate','+992+190',title,'-font',body,'-pointsize','24','-annotate','+992+270',lines,str(out)])
    return out

def game_still(name,shot,caption):
    textfits(caption,32,1184)
    out=T/(name+'.png')
    run(['convert','-size','1280x720','xc:#19251e','(',str(photo(shot)),'-crop','960x600+0+48','+repage',')','-gravity','northwest','-geometry','+160+20','-composite','-font',body,'-fill','#f0e3be','-pointsize','32','-gravity','south','-annotate','+0+24',caption,str(out)])
    return out

segments=[]
def still(name,image,duration):
    out=T/(name+'.mp4')
    run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-loop','1','-i',str(image),'-t',str(duration),'-r','25','-vf','format=yuv420p','-an','-c:v','libx264','-preset','veryfast','-crf','20','-threads','2',str(out)])
    segments.append({'file':out,'seconds':duration,'kind':'intentional still','source_image_sha256':h(image)})
def live(name,marker,duration,caption,before):
    textfits(caption,32,1184)
    seek=max(0,shots[marker]['wall_seconds']-before);captionfile=T/(name+'.txt');captionfile.write_text(caption)
    out=T/(name+'.mp4');vf='crop=960:600:0:48,pad=1280:720:160:20:color=0x19251e,drawbox=x=158:y=18:w=964:h=604:color=0xb3a06e:t=2,drawtext=fontfile='+body+':textfile='+str(captionfile)+':fontcolor=0xf0e3be:fontsize=32:x=(w-tw)/2:y=662,format=yuv420p'
    run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-ss',str(seek),'-i',str(source),'-t',str(duration),'-r','25','-vf',vf,'-an','-c:v','libx264','-preset','veryfast','-crf','20','-threads','2',str(out)])
    segments.append({'file':out,'seconds':duration,'kind':'actual live window','source_seek_seconds':seek,'caption':caption})

still('00',card('title','神劍封魔','繁體中文 × 手繪 HD','first-cave-played'),4)
live('01','town-HD',10,'原版冒險，重返熟悉的城鎮',2)
still('02',game_still('church','church-chinese-HD','用中文，再次踏上旅途'),7)
live('03','first-cave-played',10,'深入第一洞窟，拔劍迎戰',9)
still('04',panel('keyboard','keyboard-remapped','鍵盤自由設定','重綁按鍵\n保存設定\n隨時還原'),7)
still('05',panel('controller','controller-settings','接上你的手把','USB／藍牙手把\n先由系統辨識\n再於工具列綁定'),7)
live('06','first-cave-played',9,'原版配樂，以 AdLib 音色重現',8)
still('07',card('end','神劍封魔','Windows · macOS · Linux AppImage','town-HD'),6)
assert sum(s['seconds'] for s in segments)==60
listing=T/'list.txt';listing.write_text(''.join("file '"+str(s['file'])+"'\n" for s in segments))
silent=T/'silent.mp4';run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-f','concat','-safe','0','-i',str(listing),'-c','copy',str(silent)])
music=W/'title-music/source.wav';assert music.is_file()
with wave.open(str(music),'rb') as f:
    assert f.getnchannels()==1 and f.getsampwidth()==2 and f.getframerate()==44100
    samples=array.array('h',f.readframes(f.getnframes()))
onset=next(i for i in range(len(samples)//44100) if sum(v*v for v in samples[i*44100:(i+1)*44100])/44100>1000**2)
loop_seconds=min(28,len(samples)/44100-onset)
excerpt=T/'music-loop.wav'
run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-ss',str(onset),'-i',str(music),'-t',str(loop_seconds),'-af','afade=t=in:d=0.025,afade=t=out:st='+str(loop_seconds-.025)+':d=0.025',str(excerpt)])
out=O/('Zeliard-'+version+'-promo.mp4')
run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-i',str(silent),'-stream_loop','-1','-i',str(excerpt),'-filter_complex','[1:a]asetpts=PTS-STARTPTS,atrim=duration=60,loudnorm=I=-18:TP=-2:LRA=11,afade=t=in:d=1,afade=t=out:st=58:d=2[a]','-map','0:v','-map','[a]','-t','60','-c:v','copy','-c:a','aac','-ar','44100','-ac','2','-b:a','192k','-movflags','+faststart',str(out)])
metadata={'version':version,'source_video_sha256':h(source),'source_driver_sha256':h(recording/'executed-driver.py'),'music_source':'original MSCADLIB driver, native Nuked-OPL3 rendering from controlled cold-start with one normal Space input; no new composition','music_sha256':h(music),'music_start_seconds':onset,'music_loop_seconds':loop_seconds,'music_loop_sha256':h(excerpt),'segments':[{k:str(v) if isinstance(v,Path) else v for k,v in s.items()} for s in segments],'video_sha256':h(out),'rights':'private local demo; original music, art and game footage; do not upload','scope':'first-cave sampled content; no full-game translation or physical-controller claim','subtitle_bounds_checked':True,'actual_game_frames_unretouched':True}
(O/'provenance.json').write_text(json.dumps(metadata,ensure_ascii=False,indent=2)+'\n')
print('CREATED_60_SECOND_LOCAL_PROMO',out,flush=True)

#!/usr/bin/env python3
"""051：新版實際錄影、原貌／HD對照與原版配樂，合成60秒本機推廣片。"""
from pathlib import Path
import array,hashlib,json,subprocess,sys,wave

P=Path('/project');W=P/'workplace/release-v040';VERSION=(P/'VERSION').read_text().strip()
S=P/'dist-all'/VERSION/'smoke/integration';C=W/'recording';O=P/'dist-all'/VERSION/'promo';T=W/'promo-edit'
remux='--remux' in sys.argv
if not remux:T.mkdir()
receipt=json.loads((S/'receipt.json').read_text());assert receipt['status']=='PASS_SCOPED_FIRST_CAVE_INPUT_CHINESE_HD'
shots={s['name']:s for s in receipt['shots']};source=C/'gameplay.mp4';font=P/'workplace/full-hd-v1/manual-hd-v1/rounded-font/jf-openhuninn-2.1.ttf'
assert font.is_file();h=lambda f:hashlib.sha256(f.read_bytes()).hexdigest()
def run(args):return subprocess.run(args,check=True,capture_output=True,timeout=100)
def photo(name):return S/('boot'+str(shots[name]['boot']))/(name+'.png')
def fits(text,size,width):assert max(map(len,text.split('\n')))*size<=width,(text,size,width)

def card(name,title,subtitle,shot):
    fits(title,68,1184);fits(subtitle,30,1184);out=T/(name+'.png')
    run(['convert',str(photo(shot)),'-crop','960x600+0+48','+repage','-resize','1280x800^','-gravity','center','-extent','1280x720','-fill','#111d30cc','-draw','rectangle 0,0 1279,719','-stroke','#d9b96f','-strokewidth','2','-fill','none','-draw','rectangle 38,38 1241,681','-stroke','none','-font',str(font),'-fill','#f2e8ce','-pointsize','68','-gravity','center','-annotate','+0-42',title,'-pointsize','30','-annotate','+0+58',subtitle,str(out)])
    return out

def compare(name,left,right,title):
    fits(title,40,1184);out=T/(name+'.png')
    args=['convert','-size','1280x720','xc:#111d30']
    for shot,x in ((left,24),(right,648)):
        args+=['(',str(photo(shot)),'-crop','960x600+0+48','+repage','-resize','608x380',')','-gravity','northwest','-geometry',f'+{x}+150','-composite']
    args+=['-font',str(font),'-fill','#f2e8ce','-pointsize','40','-gravity','north','-annotate','+0+42',title,'-pointsize','30','-gravity','northwest','-annotate','+190+568','原版＋中文','-annotate','+784+568','原版＋中文＋HD',str(out)]
    run(args);return out

def panel(name,shot,title,body):
    fits(title,28,280);fits(body,24,280);out=T/(name+'.png')
    run(['convert','-size','1280x720','xc:#111d30',str(photo(shot)),'-gravity','northwest','-geometry','+0+20','-composite','-fill','#d9b96f','-draw','rectangle 969,52 972,668','-font',str(font),'-fill','#f2e8ce','-pointsize','28','-annotate','+992+190',title,'-pointsize','24','-annotate','+992+270',body,str(out)])
    return out

segments=[]
def still(name,image,seconds):
    out=T/(name+'.mp4');run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-loop','1','-i',str(image),'-t',str(seconds),'-r','25','-vf','format=yuv420p','-an','-c:v','libx264','-preset','veryfast','-crf','20','-threads','2',str(out)])
    segments.append({'file':str(out),'seconds':seconds,'kind':'intentional still','source_image_sha256':h(image)})
def live(name,seek,seconds,caption):
    fits(caption,32,1184);text=T/(name+'.txt');text.write_text(caption);out=T/(name+'.mp4')
    vf='crop=960:600:0:48,pad=1280:720:160:20:color=0x111d30,drawbox=x=158:y=18:w=964:h=604:color=0xd9b96f:t=2,drawtext=fontfile='+str(font)+':textfile='+str(text)+':fontcolor=0xf2e8ce:fontsize=32:x=(w-tw)/2:y=662,format=yuv420p'
    run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-ss',str(seek),'-i',str(source),'-t',str(seconds),'-r','25','-vf',vf,'-an','-c:v','libx264','-preset','veryfast','-crf','20','-threads','2',str(out)])
    segments.append({'file':str(out),'seconds':seconds,'kind':'actual live window','source_seek_seconds':seek,'caption':caption})

if not remux:
    still('00',card('title','神劍封魔','兩種畫面，同一段冒險','town-HD'),4)
    still('01',compare('town-compare','town-original','town-HD','熟悉的城鎮，重新走一遍'),8)
    live('02',shots['restored-town']['wall_seconds']+.5,10,'原版＋中文，保留熟悉的像素風景')
    still('03',compare('church-compare','church-chinese-original','church-chinese-HD','以繁體中文，再次踏上旅途'),6)
    live('04',max(0,shots['first-cave-played']['wall_seconds']-9),12,'切換手繪 HD，深入第一洞窟')
    still('05',panel('keyboard','keyboard-remapped','鍵盤自由設定','重綁按鍵\n自動保存\n隨時還原'),7)
    still('06',panel('controller','controller-settings','接上你的手把','USB／藍牙手把\n先在系統配對\n再於工具列綁定'),7)
    still('07',card('end','神劍封魔','原版＋中文｜原版＋中文＋HD\nWindows · macOS · Linux AppImage','town-HD'),6)
else:
    segments=json.loads((O/'provenance.json').read_text())['segments']
assert sum(s['seconds'] for s in segments)==60
listing=T/'list.txt';silent=T/'silent.mp4'
if not remux:
    listing.write_text(''.join("file '"+s['file']+"'\n" for s in segments))
    run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-f','concat','-safe','0','-i',str(listing),'-c','copy',str(silent)])
# 連續標題曲來自既有原版事件擷取；新遊玩錄音包含正常換場停頓，另留作裝置／來源證據。
music=P/'workplace/music-release-v1/title-music/source.wav'
with wave.open(str(music),'rb') as stream:
    assert stream.getnchannels()==1 and stream.getsampwidth()==2 and stream.getframerate()==44100
    samples=array.array('h',stream.readframes(stream.getnframes()))
onset=next(i for i in range(len(samples)//44100) if sum(v*v for v in samples[i*44100:(i+1)*44100])/44100>1000**2)
seconds=min(28,len(samples)/44100-onset);assert seconds>10
excerpt=T/'title-music-loop.wav';run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-ss',str(onset),'-i',str(music),'-t',str(seconds),'-af',f'afade=t=in:d=0.025,afade=t=out:st={seconds-.025}:d=0.025',str(excerpt)])
out=O/f'Zeliard-{VERSION}-promo.mp4'
run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-i',str(silent),'-stream_loop','-1','-i',str(excerpt),'-filter_complex','[1:a]asetpts=PTS-STARTPTS,atrim=duration=60,loudnorm=I=-18:TP=-2:LRA=11,afade=t=in:d=1,afade=t=out:st=58:d=2[a]','-map','0:v','-map','[a]','-t','60','-c:v','copy','-c:a','aac','-ar','44100','-ac','2','-b:a','192k','-movflags','+faststart',str(out)])
metadata={'version':VERSION,'source_video_sha256':h(source),'source_driver_sha256':h(C/'executed-driver.py'),'source_binary_sha256':receipt['binary_sha256'],'music_source':'original MSCADLIB title cue, established Nuked-OPL3 core; controlled cold start and one normal Space input; no new composition','music_file':str(music.relative_to(P)),'music_sha256':h(music),'music_start_seconds':onset,'music_loop_seconds':seconds,'music_loop_sha256':h(excerpt),'fresh_gameplay_audio_sha256':h(S/'boot1/source.wav'),'fresh_audio_not_used_as_music':'contains scene transitions and a 2.09 second pause inside the first candidate; original uninterrupted title capture used instead','segments':segments,'video_sha256':h(out),'rights':'private full-local only; original game footage and music; no upload','scope':'new normal first-cave and representative Chinese/HD display; sequential original/HD captures are not same-state parity proof; physical controllers not tested','subtitle_bounds_checked':True,'actual_game_frames_unretouched':True,'design_changes':['原貌與HD雙版本對照構成主線','深藍與金色取代前版綠色背景','新版圓體標題及雙版本結尾卡']}
(O/'provenance.json').write_text(json.dumps(metadata,ensure_ascii=False,indent=2)+'\n')
print('CREATED_NEW_60_SECOND_PROMO',out,flush=True)

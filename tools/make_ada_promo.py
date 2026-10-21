#!/usr/bin/env python3
"""053：以ada新錄實機畫面製作連續遊玩片段及60秒推廣片。"""
from pathlib import Path
import array,hashlib,json,os,re,subprocess,wave
from PIL import Image,ImageDraw,ImageFont,ImageOps

P=Path('/project');W=P/os.environ.get('ZELIARD_PROMO_WORK','workplace/ada-promo-v1');C=W/'received/recording';S=W/'received/capture'
assert W.resolve().is_relative_to((P/'workplace').resolve())
VERSION=(P/'VERSION').read_text().strip();O=P/'dist-all'/VERSION/'promo';T=W/'edit'
T.mkdir();assert O.is_dir() and not list(O.glob('*.mp4'))
receipt=json.loads((C/'receipt.json').read_text());assert receipt['status']=='PASS_ADA_NEW_GAMEPLAY_CAPTURE'
assert receipt['version']==VERSION and 'NVIDIA RTX 2000 Ada' in receipt['renderer']
assert all(s['state']['audio_output']['last_error']=='' for s in receipt['shots'])
source=C/'gameplay-source.mp4';pcm=C/'device.pcm'
sha=lambda f:hashlib.sha256(f.read_bytes()).hexdigest()
assert sha(source)==receipt['video_sha256']
fontfile=P/'workplace/full-hd-v1/manual-hd-v1/rounded-font/jf-openhuninn-2.1.ttf'
fonts={n:ImageFont.truetype(str(fontfile),n) for n in (26,30,34,44,70)}
bg,ink,gold,dim='#151c18','#f3eddf','#d49b59','#b9c2b6'
shots={s['name']:s for s in receipt['shots']}


def run(args,seconds=180):return subprocess.run(args,check=True,capture_output=True,timeout=seconds)


def photo(name):return S/('boot'+str(shots[name]['boot']))/(name+'.png')


def text(draw,xy,value,size,fill=ink,anchor=None):
    box=draw.multiline_textbbox((0,0),value,font=fonts[size],spacing=12)
    assert box[2]-box[0]<1190 and box[3]-box[1]<230,(value,box)
    if anchor=='mt':xy=(xy[0]-(box[2]-box[0])/2-box[0],xy[1]-box[1])
    draw.multiline_text(xy,value,font=fonts[size],fill=fill,align='center' if anchor else 'left',spacing=12)


def game(name):return Image.open(photo(name)).convert('RGB').crop((0,48,960,648))


def title(name,heading,body,shot):
    im=ImageOps.fit(game(shot),(1280,720)).convert('RGBA')
    im=Image.alpha_composite(im,Image.new('RGBA',im.size,(13,24,17,220))).convert('RGB')
    d=ImageDraw.Draw(im);d.rectangle((38,38,1241,681),outline=gold,width=2)
    d.line((420,438,860,438),fill=gold,width=2)
    text(d,(640,250),heading,70,anchor='mt');text(d,(640,478),body,30,anchor='mt')
    path=T/(name+'.png');im.save(path);return path


def compare():
    im=Image.new('RGB',(1280,720),bg);d=ImageDraw.Draw(im)
    text(d,(640,38),'原貌與手繪 HD，隨時切換',44,anchor='mt')
    for name,x,label in [('cave-original',24,'原版＋中文'),('cave-HD',648,'原版＋中文＋HD')]:
        picture=game(name).resize((608,380),Image.Resampling.LANCZOS)
        im.paste(picture,(x,155));d.rectangle((x-2,153,x+609,536),outline=gold,width=2)
        text(d,(x+304,570),label,30,anchor='mt')
    path=T/'compare.png';im.save(path);return path


def dialogue():
    im=Image.new('RGB',(1280,720),bg);d=ImageDraw.Draw(im)
    im.paste(game('church-HD'),(160,32));d.rectangle((158,30,1121,633),outline=gold,width=2)
    text(d,(640,658),'以繁體中文，再次踏上旅途',34,anchor='mt')
    path=T/'dialogue.png';im.save(path);return path


def controls(name,heading,body):
    im=Image.new('RGB',(1280,720),bg);d=ImageDraw.Draw(im)
    im.paste(Image.open(photo(name)).convert('RGB'),(0,20))
    d.line((975,52,975,668),fill=gold,width=3)
    text(d,(992,180),heading,26)
    for i,line in enumerate(body):text(d,(992,250+i*52),line,26,dim)
    path=T/(name+'.png');im.save(path);return path


log=(C/'ffmpeg.log').read_text(errors='replace')
match=re.search(r'start:\s*(\d+\.\d+)',log)
assert match,'需要原始X11擷取時間，不能猜測音訊對齊'
video_epoch=float(match.group(1));video_delay=video_epoch-receipt['record_begin_epoch']
audio_offset=video_epoch-receipt['monitor_begin_epoch']
assert 0<=video_delay<3 and 0<=audio_offset<90,(video_delay,audio_offset)
marks={m['name']:m['wall_seconds']-video_delay for m in receipt['markers']}
probe=json.loads(run(['ffprobe','-v','error','-show_streams','-show_format','-of','json',str(source)]).stdout)
duration=float(probe['format']['duration']);audio_seconds=pcm.stat().st_size/4/44100
assert audio_seconds>audio_offset+marks['gameplay-end']-.2

# 保留正常連續遊玩與同期裝置音訊，不加速、不循環、不重排。
begin,end=marks['town-walk-begin'],marks['gameplay-end']
assert 20<end-begin<400 and 0<=begin<end<=duration
gameplay=O/f'Zeliard-{VERSION}-ada-gameplay.mp4'
run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-ss',str(begin),'-i',str(source),'-f','s16le','-ar','44100','-ac','2','-ss',str(audio_offset+begin),'-i',str(pcm),'-map','0:v','-map','1:a','-t',str(end-begin),'-c:v','libx264','-preset','veryfast','-crf','19','-threads','2','-c:a','aac','-b:a','192k','-movflags','+faststart',str(gameplay)],300)

segments=[]
def still(name,path,seconds):
    out=T/(name+'.mp4')
    run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-loop','1','-i',str(path),'-t',str(seconds),'-r','60','-vf','format=yuv420p','-an','-c:v','libx264','-preset','veryfast','-crf','21','-threads','2',str(out)])
    segments.append({'file':str(out),'seconds':seconds,'kind':'intentional still','image_sha256':sha(path)})


def live(name,start,seconds,caption):
    assert start>=0 and start+seconds<=duration
    captionfile=T/(name+'.txt');captionfile.write_text(caption)
    out=T/(name+'.mp4')
    vf='crop=960:600:0:48,pad=1280:720:160:24:color=0x151c18,drawbox=x=158:y=22:w=964:h=604:color=0xd49b59:t=2,drawtext=fontfile='+str(fontfile)+':textfile='+str(captionfile)+':fontcolor=0xf3eddf:fontsize=32:x=(w-tw)/2:y=658,format=yuv420p'
    run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-ss',str(start),'-i',str(source),'-t',str(seconds),'-r','60','-vf',vf,'-an','-c:v','libx264','-preset','veryfast','-crf','21','-threads','2',str(out)])
    segments.append({'file':str(out),'seconds':seconds,'kind':'actual live window','source_seek_seconds':start,'caption':caption,'speed':1})


still('00',title('title','神劍封魔','繁體中文・手繪奇幻','town-HD'),3)
live('01',max(begin,marks['church-dialogue']-12),12,'走過城鎮，前往地下迷宮')
still('02',dialogue(),6)
still('03',compare(),6)
live('04',end-20,20,'揮劍、跳躍，深入第一洞窟')
still('05',controls('keyboard-settings','鍵盤自由設定',['重綁按鍵','保存操作設定']),4)
still('06',controls('controller-settings','接上你的手把',['USB／藍牙手把','先在系統配對','再於工具列綁定']),4)
still('07',title('end','神劍封魔','原版＋中文｜原版＋中文＋HD\nWindows · macOS · Linux AppImage','town-return-HD'),5)
assert sum(s['seconds'] for s in segments)==60
listing=T/'list.txt';listing.write_text(''.join("file '"+s['file']+"'\n" for s in segments))
silent=T/'silent.mp4';run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-f','concat','-safe','0','-i',str(listing),'-c','copy',str(silent)])

music=P/'workplace/music-release-v1/title-music/source.wav'
with wave.open(str(music),'rb') as stream:
    assert stream.getnchannels()==1 and stream.getsampwidth()==2 and stream.getframerate()==44100
    samples=array.array('h',stream.readframes(stream.getnframes()))
onset=next(i for i in range(len(samples)//44100) if sum(v*v for v in samples[i*44100:(i+1)*44100])/44100>1000**2)
excerpt=T/'original-title.wav'
run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-ss',str(onset),'-i',str(music),'-t','28','-af','afade=t=in:d=0.025,afade=t=out:st=27.975:d=0.025',str(excerpt)])
promo=O/f'Zeliard-{VERSION}-promo.mp4'
run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-i',str(silent),'-stream_loop','-1','-i',str(excerpt),'-filter_complex','[1:a]atrim=duration=60,loudnorm=I=-18:TP=-2:LRA=11,afade=t=in:d=1,afade=t=out:st=58:d=2[a]','-map','0:v','-map','[a]','-t','60','-c:v','copy','-c:a','aac','-ar','44100','-ac','2','-b:a','192k','-movflags','+faststart',str(promo)])
metadata={'version':VERSION,'capture_machine':'ada','renderer':receipt['renderer'],'source_video_sha256':sha(source),'source_driver_sha256':receipt['script_sha256'],'source_binary_sha256':receipt['binary_sha256'],'recording_fps':60,'source_gpu_encoder':'h264_nvenc','segments':segments,'video_sha256':sha(promo),'music_file':str(music.relative_to(P)),'music_sha256':sha(music),'music_source':'已追溯的原版標題曲：MSCADLIB事件及Nuked-OPL3音源；推廣片配樂與同期聲分開','actual_game_frames_unretouched':True,'title_card_backdrop':'只有標題卡加深背景，遊玩片段保留原色','subtitle_bounds_checked':True,'design_changes':['以ada新錄的動態遊玩為主線','洞窟岩色與城鎮植被取銅金／深綠配色','60fps呈現及較短的開場／結尾'],'gameplay_clip':{'file':gameplay.name,'sha256':sha(gameplay),'seconds':end-begin,'source_begin_seconds':begin,'source_end_seconds':end,'audio':'ada PulseAudio monitor同期聲','monitor_audio_offset_seconds':audio_offset,'alignment':'同主機X11原始擷取epoch與parec啟動epoch；含程序啟動的小幅時間誤差','audio_sha256':sha(pcm)},'rights':'private full-local only; original game footage and music; no upload','scope':'正常F7遊玩與代表中文／HD；非通關、每幀全遊戲或實體手把驗收'}
(O/'provenance.json').write_text(json.dumps(metadata,ensure_ascii=False,indent=2)+'\n')
print('CREATED_ADA_GAMEPLAY_AND_NEW_PROMO',round(end-begin,3),'seconds gameplay',flush=True)

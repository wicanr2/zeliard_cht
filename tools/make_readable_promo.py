"""065：正式1.0.5大字介面新錄與具名舊遊玩來源，60秒本機推廣片。"""
from pathlib import Path
import hashlib,json,re,subprocess
from PIL import Image,ImageDraw,ImageFont,ImageOps

P=Path('/project');V=(P/'VERSION').read_text().strip();W=P/'workplace/release-v105'
D=P/'dist-all'/V/'promo';T=W/'edit';C=W/'recording'
assert V=='v.1.0.5-20261022' and not list(D.glob('*.mp4'))
T.mkdir(exist_ok=False)
h=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def run(args,seconds=150):return subprocess.run(args,check=True,capture_output=True,timeout=seconds)
def write(p,data):p.write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')

new=json.loads((P/'dist-all'/V/'smoke/toolbar/receipt.json').read_text())
assert new['version']==V and new['status']=='PASS_SCOPED_UI_TEXT_SETTINGS'
source=Path(new['recording']);assert h(source)==new['video_sha256']
old=P/'dist-all/v.1.0.4-20261022/promo/Zeliard-v.1.0.4-20261022-promo.mp4'
prior=json.loads((old.parent/'provenance.json').read_text());assert h(old)==prior['video_sha256']
oldproof=json.loads((old.parent/'hd-stability-proof.json').read_text())
assert oldproof['status']=='PASS_SCOPED_CONTINUOUS_HD_VIDEO_PROOF'
font=P/'workplace/ui-text-v1/rounded-font/jf-openhuninn-2.1.ttf'
fonts={n:ImageFont.truetype(str(font),n)for n in (22,24,26,28,32,40,56,68)}
bg,ink,mint,gold,dim='#101c23','#f0eddf','#7bc6b3','#d5a96c','#b3c2c6'

def text(d,xy,s,size,fill=ink):
    box=d.multiline_textbbox(xy,s,font=fonts[size],spacing=10)
    assert box[0]>=0 and box[1]>=0 and box[2]<=1260 and box[3]<=704,(s,box)
    d.multiline_text(xy,s,font=fonts[size],fill=fill,spacing=10)

def panel(name,heading,lines,version=V):
    im=Image.new('RGBA',(1280,720),(0,0,0,0));d=ImageDraw.Draw(im)
    d.rectangle((14,18,977,701),outline=mint,width=2);d.rectangle((990,20,1264,700),fill=bg)
    text(d,(1006,58),heading,28,mint);d.line((1006,112,1245,112),fill=gold,width=2)
    for i,line in enumerate(lines):text(d,(1006,145+i*47),line,26)
    text(d,(1006,610),version,22,dim)
    f=T/(name+'-panel.png');im.save(f);return f

photo=P/'workplace/promo-features-v104/recording/platform-HD.png'
def title(name,heading,body):
    im=ImageOps.fit(Image.open(photo).convert('RGB').crop((144,90,816,522)),(1280,720)).convert('RGBA')
    im=Image.alpha_composite(im,Image.new('RGBA',im.size,(12,24,29,220))).convert('RGB');d=ImageDraw.Draw(im)
    d.rectangle((42,42,1238,678),outline=mint,width=2);d.line((80,124,280,124),fill=gold,width=4)
    text(d,(82,205),heading,68);text(d,(84,335),body,32)
    text(d,(84,540),'配樂：原版標題曲，非同期聲',24,dim)
    text(d,(84,595),'ZELIARD  /  '+V,24,mint)
    f=T/(name+'.png');im.save(f);return f

delay=float(re.search(r'start:\s*(\d+\.\d+)',(C/'ffmpeg.log').read_text(errors='replace')).group(1))-new['record_begin_epoch']
assert 0<=delay<15
shots={s['name']:s for s in new['shots']if '/boot1/'in s['path']}
segments=[]
def still(name,picture,seconds,caption):
    out=T/(name+'.mp4')
    run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-filter_threads','1','-loop','1','-i',str(picture),'-t',str(seconds),'-r','30','-an','-vf','format=yuv420p','-c:v','libx264','-preset','veryfast','-crf','20','-threads','1',str(out)])
    segments.append({'file':str(out),'kind':'intentional still','seconds':seconds,'caption':caption,'image_sha256':h(picture)})

def previous(name,start,seconds,caption,source_version,proof=False):
    out=T/(name+'.mp4')
    run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-ss',str(start),'-i',str(old),'-t',str(seconds),'-an','-c:v','libx264','-preset','veryfast','-crf','20','-threads','1','-pix_fmt','yuv420p',str(out)])
    segments.append({'file':str(out),'kind':'actual live window','source':'previous promo; version shown in original sidebar','source_version':source_version,'source_seek_seconds':start,'seconds':seconds,'caption':caption,'speed':1,'internal_cuts':False,'proof_segment':proof,'expected_native_pause':proof,'full_window_preserved':True})

def live(name,shot,seconds,caption,lines):
    start=shots[shot]['wall_seconds']-delay+.1
    overlay=panel(name,caption,lines);out=T/(name+'.mp4')
    vf='[0:v]pad=1280:720:16:20:color=0x101c23[base];[base][1:v]overlay=0:0:shortest=1,format=yuv420p[v]'
    run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-filter_complex_threads','1','-ss',str(start),'-i',str(source),'-loop','1','-i',str(overlay),'-t',str(seconds),'-filter_complex',vf,'-map','[v]','-r','30','-an','-c:v','libx264','-preset','veryfast','-crf','20','-threads','1',str(out)])
    segments.append({'file':str(out),'kind':'actual live window','source':'new formal HD AppImage recording','source_version':V,'source_seek_seconds':start,'seconds':seconds,'caption':caption,'speed':1,'internal_cuts':False,'expected_native_pause':True,'full_window_preserved':True})

still('00-title',title('title','神劍封魔','繁體中文與手繪 HD\n系統文字放大，看得更清楚'),3,'開場')
previous('01-town',3,7,'重返地下迷宮','v.1.0.2-20261021')
previous('02-dialogue',10,7,'HD 對話連續證明','v.1.0.4-20261022',True)
previous('03-platform',17,11,'HD 平台連續證明','v.1.0.4-20261022',True)
for size in (16,24,32):
    live('04-text-size-'+str(size),'text-size-'+str(size),4,'字級自由調整',['標準 16','大字 24','特大 32','','切換實錄節選'])
live('05-cheat','cheat-32',7,'大字難度設定',['生命與 MP 可鎖滿','兩者獨立開關','站立回血可關閉','','本文使用特大字'])
live('06-speed','speed-32',5,'速度上限設定',['1／2／4／8倍','16／32倍上限','負載高時自動調整','','鍵盤與手把可設定'])
load=P/'workplace/promo-features-v104/edit/loading-window.png'
card=Image.new('RGBA',(1280,720),bg);card.paste(Image.open(load).convert('RGBA'),(16,20))
card.alpha_composite(Image.open(panel('loading','啟動與按需載入',['啟動進度看得見','首次進場景才載入','用過的圖片留快取','','前版載入畫面'],'v.1.0.4-20261022')))
f=T/'loading-card.png';card.convert('RGB').save(f);still('07-loading',f,3,'按需載入')
still('08-end',title('end','神劍封魔','原版＋中文   ／   原版＋中文＋HD\nLinux AppImage・Windows・macOS\n可調字級・速度・鍵盤與手把設定'),5,'雙版本三平台')
assert sum(s['seconds']for s in segments)==60
listing=T/'list.txt';listing.write_text(''.join("file '"+s['file']+"'\n"for s in segments))
silent=T/'silent.mp4';run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-f','concat','-safe','0','-i',str(listing),'-c','copy',str(silent)])
music=P/prior['music_file'];assert h(music)==prior['music_sha256']
excerpt=P/'workplace/promo-features-v104/edit/original-title.wav'
promo=D/f'Zeliard-{V}-promo.mp4'
run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-i',str(silent),'-stream_loop','-1','-i',str(excerpt),'-filter_complex_threads','1','-filter_complex','[1:a]atrim=duration=60,loudnorm=I=-18:TP=-2:LRA=11,afade=t=in:d=1,afade=t=out:st=58:d=2[a]','-map','0:v','-map','[a]','-t','60','-c:v','copy','-c:a','aac','-ac','2','-ar','44100','-b:a','192k','-movflags','+faststart',str(promo)])
meta={'version':V,'capture_machine':'local','video_sha256':h(promo),'segments':segments,'new_recording_sha256':h(source),'new_capture_receipt_sha256':h(P/'dist-all'/V/'smoke/toolbar/receipt.json'),'source_binary_sha256':new['binary_sha256'],'source_script_sha256':new['script_sha256'],'old_promo':str(old.relative_to(P)),'old_promo_sha256':h(old),'old_proof_receipt_sha256':h(old.parent/'hd-stability-proof.json'),'make_script_sha256':h(Path(__file__)),'music_file':str(music.relative_to(P)),'music_sha256':h(music),'music_excerpt_sha256':h(excerpt),'music_source':'已追溯原版AdLib與Nuked-OPL3標題曲；沿前片，非同期聲','new_source_timing_delay':delay,'subtitle_bounds_checked':True,'actual_game_frames_unretouched':True,'game_window_scaled':False,'rights':'local only; original game footage/music; no upload','scope':'新錄1.0.5正式設定頁，遊玩來源1.0.2／1.0.4明列版本；不冒稱新版遊玩、全遊戲或30fps'}
write(D/'provenance.json',meta)
(T/'make_readable_promo.py').write_bytes(Path(__file__).read_bytes())
print('CREATED_V105_READABLE_PROMO',flush=True)

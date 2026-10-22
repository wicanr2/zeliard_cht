"""063：舊實錄混剪、新功能與未剪接HD證明，正式影片只留本機。"""
from pathlib import Path
import array,hashlib,json,os,re,shutil,subprocess,sys,wave
from PIL import Image,ImageDraw,ImageFont,ImageOps

P=Path('/project');V=(P/'VERSION').read_text().strip();W=P/'workplace/promo-features-v104'
C=W/'recording';OLD=P/'workplace/ada-promo-v102/received/recording';D=P/'dist-all'/V/'promo';T=W/'edit'
h=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()

def finish():
    quality=json.loads((D/'quality.json').read_text());proof=json.loads((D/'hd-stability-proof.json').read_text())
    visual=json.loads((D/'visual-review.json').read_text());provenance=json.loads((D/'provenance.json').read_text())
    assert quality['status']=='PASS_PROMO_TECHNICAL_QA' and proof['status']=='PASS_SCOPED_CONTINUOUS_HD_VIDEO_PROOF'
    assert visual['status']=='PASS_SCOPED_VISUAL_REVIEW'
    assert quality['sha256']==h(D/f'Zeliard-{V}-promo.mp4')==provenance['video_sha256']
    assert h(T/'make-feature-promo.render.py')==provenance['make_script_sha256']
    root=D.parent;file=root/'SHA256SUMS.json';manifest=json.loads(file.read_text())
    assert manifest['version']==V and not (W/'manifest-before-media.json').exists()
    packages={k:v for k,v in manifest['artifacts'].items()if k.startswith('full-local/')}
    assert len(packages)==6
    for name,row in packages.items():assert (root/name).stat().st_size==row['bytes'] and h(root/name)==row['sha256']
    shutil.copy2(file,W/'manifest-before-media.json')
    for clip in D.glob('*.mp4'):
        name=str(clip.relative_to(root));assert name not in manifest['artifacts']
        manifest['artifacts'][name]={'bytes':clip.stat().st_size,'sha256':h(clip)}
    manifest['media_validation']={'promo':quality,'hd_stability_proof':proof,'visual_review':visual,'current_fps_accepted_by_user':True,'full_game_or_30fps_claim':False}
    manifest['media_provenance']={'file':'promo/provenance.json','sha256':h(D/'provenance.json'),'finalization_script_sha256':h(Path(__file__))}
    temp=root/'SHA256SUMS.json.tmp';temp.write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n');os.replace(temp,file)
    media={'version':V,'rights':'local only; original game footage and music; no upload','files':{x.name:{'bytes':x.stat().st_size,'sha256':h(x)}for x in sorted(D.iterdir())if x.is_file()and x.name!='media-manifest.json'}}
    (D/'media-manifest.json').write_text(json.dumps(media,ensure_ascii=False,indent=2)+'\n')
    removed=[]
    for directory in [W/'app']+sorted(T.glob('*.mp4')):
        assert directory.exists() and directory.lstat().st_uid==os.getuid()
        entries=list(directory.rglob('*'))if directory.is_dir()else[directory]
        assert all(x.lstat().st_uid==os.getuid()for x in entries)
        removed.append({'file':str(directory.relative_to(W)),'allocated_bytes':sum(x.lstat().st_blocks*512 for x in entries)})
        if directory.is_dir():shutil.rmtree(directory)
        else:directory.unlink()
    for name,row in manifest['artifacts'].items():assert h(root/name)==row['sha256']
    for name,row in media['files'].items():assert h(D/name)==row['sha256']
    report={'status':'PASS_MEDIA_REGISTERED_AND_CLEANED','version':V,'formal_packages_unchanged':6,'formal_videos':3,'removed':removed,'allocated_bytes':sum(r['allocated_bytes']for r in removed),'preserved':'原始錄影、前版影片、兩段完整證明、正式推廣片、來源／SHA、畫格、音源與失敗紀錄均保留。'}
    (W/'cleanup.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n');print(report['status'],report['allocated_bytes'],flush=True)

if sys.argv[1:]==['finish']:
    finish();raise SystemExit(0)
assert not sys.argv[1:]
assert V=='v.1.0.4-20261022' and D.is_dir() and not list(D.glob('*.mp4'))
T.mkdir()
(T/'make-feature-promo.render.py').write_bytes(Path(__file__).read_bytes())
new=json.loads((C/'receipt.json').read_text());old=json.loads((OLD/'receipt.json').read_text())
assert new['status']=='PASS_V104_FEATURE_AND_HD_PROOF_CAPTURE' and new['version']==V
assert old['version']=='v.1.0.2-20261021'
sources={'new':C/'feature-source.mp4','old':OLD/'gameplay-source.mp4'}
assert h(sources['new'])==new['source_video_sha256'] and h(sources['old'])==old['video_sha256']
font=P/'workplace/runtime-controls-v1/rounded-font/jf-openhuninn-2.1.ttf'
fonts={n:ImageFont.truetype(str(font),n) for n in (22,24,28,32,40,56,68)}
bg,ink,mint,gold,dim='#101c23','#f0eddf','#7bc6b3','#d5a96c','#b3c2c6'

def run(args,seconds=150):
    return subprocess.run(args,check=True,capture_output=True,timeout=seconds)

def write(path,data):path.write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')
def text(d,xy,value,size,fill=ink):
    box=d.multiline_textbbox(xy,value,font=fonts[size],spacing=10)
    assert box[0]>=0 and box[1]>=0 and box[2]<=1260 and box[3]<=704,(value,box)
    d.multiline_text(xy,value,font=fonts[size],fill=fill,spacing=10)

def panel(name,heading,lines,version=V):
    im=Image.new('RGBA',(1280,720),(0,0,0,0));d=ImageDraw.Draw(im)
    d.rectangle((14,18,977,701),outline=mint,width=2)
    d.rectangle((990,20,1264,700),fill=bg)
    text(d,(1006,58),heading,28,mint);d.line((1006,112,1245,112),fill=gold,width=2)
    for i,line in enumerate(lines):text(d,(1006,145+i*45),line,24)
    text(d,(1006,610),version,22,dim)
    path=T/(name+'-panel.png');im.save(path);return path

def title(name,heading,body,photo):
    im=ImageOps.fit(Image.open(photo).convert('RGB').crop((144,90,816,522)),(1280,720)).convert('RGBA')
    im=Image.alpha_composite(im,Image.new('RGBA',im.size,(12,24,29,220))).convert('RGB');d=ImageDraw.Draw(im)
    d.rectangle((42,42,1238,678),outline=mint,width=2)
    d.line((80,124,280,124),fill=gold,width=4)
    text(d,(82,218),heading,68);text(d,(84,350),body,32)
    text(d,(84,595),'ZELIARD  /  '+V,24,mint)
    path=T/(name+'.png');im.save(path);return path

def timing(receipt,log):
    m=re.search(r'start:\s*(\d+\.\d+)',log.read_text(errors='replace'));assert m
    delay=float(m.group(1))-receipt['record_begin_epoch'];assert 0<=delay<15,delay
    return {x['name']:x['wall_seconds']-delay for x in receipt['markers']},delay

nmarks,ndelay=timing(new,C/'ffmpeg.log');omarks,odelay=timing(old,OLD/'ffmpeg.log')
music=P/'workplace/music-release-v1/title-music/source.wav'
prior=json.loads((P/'dist-all/v.1.0.2-20261021/promo/provenance.json').read_text())
assert h(music)==prior['music_sha256']
with wave.open(str(music),'rb') as f:
    assert (f.getnchannels(),f.getsampwidth(),f.getframerate())==(1,2,44100)
    data=array.array('h',f.readframes(f.getnframes()))
onset=next(i for i in range(len(data)//44100) if sum(v*v for v in data[i*44100:(i+1)*44100])/44100>1000**2)
excerpt=T/'original-title.wav'
run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-ss',str(onset),'-i',str(music),'-t','28','-af','afade=t=in:d=0.025,afade=t=out:st=27.975:d=0.025',str(excerpt)])

proofs=[]
for label in ('dialogue','platform'):
    key=label+'-proof';start=nmarks[key+'-begin']+.3;seconds=nmarks[key+'-end']-nmarks[key+'-begin']-.6
    assert seconds>8
    file=D/f'Zeliard-{V}-HD-{label}-proof.mp4'
    run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-filter_threads','1','-ss',str(start),'-i',str(sources['new']),'-stream_loop','-1','-i',str(excerpt),'-t',str(seconds),'-map','0:v','-map','1:a','-c:v','libx264','-preset','veryfast','-crf','18','-threads','1','-pix_fmt','yuv420p','-c:a','aac','-b:a','160k','-ac','2','-af',f'loudnorm=I=-18:TP=-2:LRA=11,afade=t=in:d=0.25,afade=t=out:st={seconds-.5}:d=0.5','-movflags','+faststart',str(file)])
    samples=[x for x in new['proof_samples'] if x['label']==key and start+ndelay<=x['wall_seconds']<=start+ndelay+seconds]
    assert samples and all(s['hd_enabled'] and s['town_hd' if label=='dialogue' else 'cave_hd'] for s in samples)
    proofs.append({'file':file.name,'sha256':h(file),'label':label,'source_seek_seconds':start,'seconds':seconds,'speed':1,'internal_cuts':False,'manual_hd_toggle':False,'sample_count':len(samples),'source_sample_frames':[s['frames'] for s in samples],'audio':'original title background music; not synchronous sound','hp_lock_active':True,'full_window_preserved':True})

segments=[]
def still(name,picture,seconds,caption):
    out=T/(name+'.mp4')
    run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-filter_threads','1','-loop','1','-i',str(picture),'-t',str(seconds),'-r','30','-an','-vf','format=yuv420p','-c:v','libx264','-preset','veryfast','-crf','20','-threads','1',str(out)])
    segments.append({'file':str(out),'kind':'intentional still','seconds':seconds,'caption':caption,'image_sha256':h(picture)})

def live(name,source_key,start,seconds,heading,lines,proof=False,native_pause=False):
    overlay=panel(name,heading,lines,V if source_key=='new' else old['version']);out=T/(name+'.mp4')
    vf='[0:v]pad=1280:720:16:20:color=0x101c23[base];[base][1:v]overlay=0:0:shortest=1,format=yuv420p[v]'
    run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-filter_complex_threads','1','-ss',str(start),'-i',str(sources[source_key]),'-loop','1','-i',str(overlay),'-t',str(seconds),'-filter_complex',vf,'-map','[v]','-r','30','-an','-c:v','libx264','-preset','veryfast','-crf','20','-threads','1',str(out)])
    segments.append({'file':str(out),'kind':'actual live window','seconds':seconds,'source':source_key,'source_seek_seconds':start,'source_version':new['version'] if source_key=='new' else old['version'],'speed':1,'caption':heading,'proof_segment':proof,'full_window_preserved':True,'internal_cuts':False,'expected_native_pause':native_pause})

still('00-title',title('title','神劍封魔','繁體中文與手繪 HD\n熟悉的冒險，新的操作選擇',C/'platform-HD.png'),3,'開場')
live('01-tour','old',omarks['town-walk-begin'],7,'重返地下迷宮',['城鎮探索','揮劍與魔法','手繪奇幻 HD','','前版實錄節選'])
live('02-dialogue','new',nmarks['dialogue-proof-begin']+.5,7,'HD 對話實錄',['HD 全程開啟','對話背景保留','不剪接、不切換','','生命鎖定開啟'],True,True)
assert nmarks['platform-proof-end']-nmarks['platform-proof-begin']>13.5
live('03-platform','new',nmarks['platform-proof-begin']+.5,13,'HD 平台實錄',['站立、升降、移動','HD 全程開啟','不剪接、不切換','','生命鎖定開啟'],True,True)
live('04-speed','new',nmarks['speed-feature-begin']+.4,7,'速度自由選擇',['1／2／4／8倍','16／32倍上限','負載高時自動調整','','工具列內設定'],False,True)
live('05-cheats','new',nmarks['cheat-feature-begin']+.2,8,'難度由你選擇',['生命與 MP 可鎖滿','兩者獨立開關','站立回血可關閉','','MP 為法術次數'],False,True)

load_start,load_end=nmarks['town-load-begin'],nmarks['town-load-end']
assert load_end>load_start+.4
loading=T/'loading-window.png'
run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-ss',str((load_start+load_end)/2),'-i',str(sources['new']),'-frames:v','1','-threads','1',str(loading)])
load_card=Image.new('RGBA',(1280,720),bg);load_card.paste(Image.open(loading).convert('RGBA'),(16,20));load_card.alpha_composite(Image.open(panel('loading','啟動與按需載入',['啟動進度看得見','HD 進入場景再載入','用過的圖片保留快取','','首次載入時暫停遊戲'])))
load_image=T/'loading-card.png';load_card.convert('RGB').save(load_image)
still('06-loading',load_image,5,'啟動進度與按需載入')

settings=Image.new('RGB',(1280,720),bg);draw=ImageDraw.Draw(settings)
photos=P/'workplace/ada-promo-v102/received/capture'
for shot,x in (('keyboard-settings',28),('controller-settings',652)):
    item=next(s for s in old['shots'] if s['name']==shot)
    im=Image.open(photos/('boot'+str(item['boot']))/(shot+'.png')).convert('RGB').resize((600,425),Image.Resampling.LANCZOS)
    settings.paste(im,(x,125));draw.rectangle((x-2,123,x+601,551),outline=mint,width=2)
text(draw,(46,30),'鍵盤與手把，照自己的習慣設定',40)
text(draw,(46,584),'USB／藍牙手把先由系統辨識，再於工具列選擇與綁定。',28)
text(draw,(46,642),'設定頁沿用前版實錄 v.1.0.2；未宣稱實體手把驗收。',22,dim)
settings_file=T/'controls.png';settings.save(settings_file)
still('07-controls',settings_file,5,'鍵盤與手把設定')
still('08-end',title('end','神劍封魔','原版＋中文   ／   原版＋中文＋HD\nLinux AppImage・Windows・macOS',C/'platform-HD.png'),5,'雙版本三平台')
assert sum(s['seconds']for s in segments)==60
listing=T/'list.txt';listing.write_text(''.join("file '"+s['file']+"'\n"for s in segments))
silent=T/'silent.mp4';run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-f','concat','-safe','0','-i',str(listing),'-c','copy',str(silent)])
promo=D/f'Zeliard-{V}-promo.mp4'
run(['ffmpeg','-nostdin','-hide_banner','-loglevel','error','-i',str(silent),'-stream_loop','-1','-i',str(excerpt),'-filter_complex_threads','1','-filter_complex','[1:a]atrim=duration=60,loudnorm=I=-18:TP=-2:LRA=11,afade=t=in:d=1,afade=t=out:st=58:d=2[a]','-map','0:v','-map','[a]','-t','60','-c:v','copy','-c:a','aac','-ac','2','-ar','44100','-b:a','192k','-movflags','+faststart',str(promo)])
meta={'version':V,'capture_machine':'local','renderer':new['renderer'],'video_sha256':h(promo),'segments':segments,'proof_clips':proofs,'new_recording_sha256':h(sources['new']),'old_recording_sha256':h(sources['old']),'source_binary_sha256':new['source_binary_sha256'],'source_script_sha256':new['script_sha256'],'make_script_sha256':h(Path(__file__)),'music_file':str(music.relative_to(P)),'music_sha256':h(music),'music_source':'已追溯原版AdLib事件與Nuked-OPL3標題曲，沿前片；配樂非同期聲','new_source_timing_delay':ndelay,'old_source_timing_delay':odelay,'subtitle_bounds_checked':True,'actual_game_frames_unretouched':True,'proof_game_window_scaled':False,'theme':'洞窟石壁、青綠工具列與暖金','design_changes':['以連續修正證明取代前版場景巡禮主線','完整工具列與原尺寸遊玩置左，右欄介紹功能與版本','另交兩段不剪接、不切換HD的原尺寸證明'],'rights':'private local only; original game footage and music; no upload','scope':'錄到的居民及綠色平台範圍；不外推全遊戲、精確未知站位或30fps'}
write(D/'provenance.json',meta)
print('CREATED_FEATURE_PROMO_AND_TWO_UNCUT_PROOFS',flush=True)

"""用實際裝置監聽核原版音樂短段的取樣與節奏，不宣稱整段硬體一致。"""
from pathlib import Path
import hashlib,json,wave
import numpy as np
R=Path('/project');W=R/'workplace/music-release-v1';D=R/'dist-all'/(R/'VERSION').read_text().strip()/'smoke'
def inspect(gui,recording):
    source=gui/'boot1/source.wav';device=recording/'device.pcm'
    with wave.open(str(source),'rb')as f:
        assert f.getframerate()==44100 and f.getnchannels()==1
        src=np.frombuffer(f.readframes(f.getnframes()),dtype='<i2').astype(np.int32)
    stereo=np.fromfile(device,dtype='<i2').reshape(-1,2).astype(np.int32);dev=stereo[:,0]
    start=int(np.flatnonzero(np.abs(src)>32)[0])+8820
    def locate(q,low,high):
        low=max(0,low);high=min(len(dev)-len(q),high)
        peak=int(np.argmax(np.abs(q)));candidates=np.flatnonzero(np.abs(dev[low+peak:high+peak]-q[peak])<=2)+low
        probes=np.arange(0,len(q),128)
        for begin in range(0,len(candidates),10000):
            ids=candidates[begin:begin+10000]
            ids=ids[np.all(np.abs(dev[ids[:,None]+probes]-q[probes])<=2,axis=1)]
            for i in ids:
                if int(np.max(np.abs(dev[i:i+len(q)]-q)))<=2:return int(i)
        return None
    matches=[]
    for offset in [0,22050,44100]:
        at=start+offset;q=src[at:at+4096]
        if matches:
            target=matches[0]['device_frame']+offset;low,high=target-44100,target+88200
        else:low,high=0,min(len(dev),44100*60)
        pos=locate(q,low,high)
        if pos is not None:matches.append({'source_frame':at,'device_frame':pos,'matched_frames':len(q),'max_sample_difference':int(np.max(np.abs(dev[pos:pos+len(q)]-q)))})
    ratio=None
    if len(matches)>=2:ratio=(matches[-1]['device_frame']-matches[0]['device_frame'])/(matches[-1]['source_frame']-matches[0]['source_frame'])
    return {'source_seconds':len(src)/44100,'device_seconds_including_UI_pauses':len(dev)/44100,'source_peak':int(np.abs(src).max()),'device_peak':int(np.abs(stereo).max()),'stereo_max_difference':int(np.abs(stereo[:,0]-stereo[:,1]).max()),'matches':matches,'excerpt_time_ratio':ratio,'source_sha256':hashlib.sha256(source.read_bytes()).hexdigest(),'monitor_sha256':hashlib.sha256(device.read_bytes()).hexdigest()}
current=inspect(W/'gui-v4',W/'recording-v4')
baseline=inspect(W/'gui-v3',W/'recording-v3')
assert current['source_peak']>0 and current['device_peak']<32767
assert len(current['matches'])>=2 and .8<=current['excerpt_time_ratio']<=1.25,current
report={'status':'PASS_NATIVE_ORIGINAL_MUSIC_EXCERPT','current_realtime':current,'prior_fixed_clock':baseline,'scope':'三個候選中有'+str(len(current['matches']))+'個4096音框片段與實際裝置相符，核起點相對節奏；不宣稱全輸出零丟樣、逐取樣時序、人耳或實體喇叭通過','audio_core':'Nuked-OPL3，原版MSCADLIB事件，不新作曲'}
(D/'audio-quality.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n');print(report['status'],current['excerpt_time_ratio'],baseline['excerpt_time_ratio'])

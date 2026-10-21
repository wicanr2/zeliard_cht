#!/usr/bin/env python3
"""核十二個已存主角來源並把生成圖集直接註冊成72x72姿態。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image, ImageDraw
from hd_mpp1_art import area_image


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project',type=Path,required=True)
    ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args();p,out=args.project,args.output
    assert out.is_dir() and not any(out.iterdir())
    sha=lambda path:hashlib.sha256(path.read_bytes()).hexdigest()
    b=p/'workplace/full-hd-v1';src_path=b/'cave-saved-gear-source-v1/receipt.json'
    src=json.loads(src_path.read_text());assert src['status']=='SAVED_GEAR_SOURCE_ONLY'
    sources=[r for r in src['sources']if r['kind']=='player'];assert len(sources)==12
    gen_path=b/'cave-saved-player-art-v1/generation.json';gen=json.loads(gen_path.read_text())
    assert gen['mode']=='built-in imagegen'
    for ref in gen['references']:assert sha(p/ref['file'])==ref['sha256']
    path=p/gen['output_file'];assert sha(path)==gen['output_sha256']
    atlas=Image.open(path).convert('RGBA');assert atlas.getchannel('A').getextrema()[0]==0
    # 沿hd_register_npcs.py的alpha96墨跡界定，避免極淡背景噪點撐大裁切框。
    # 先取實際透明列間隔；等分格線可能碰到腳尖，不能直接截斷。
    alpha=np.asarray(atlas)[:,:,3];cuts=[]
    for column in range(3):
        a=alpha[:,column*atlas.width//3:(column+1)*atlas.width//3]
        maximum=a.max(axis=1);boundaries=[0]
        for row in range(1,4):
            target=row*atlas.height//4;choices=range(target-50,target+51)
            boundary=min(choices,key=lambda y:(int(maximum[y]),abs(y-target)))
            assert int(maximum[boundary])<=1,'列間沒有完整透明間隔'
            boundaries.append(boundary)
        cuts.append(boundaries+[atlas.height])
    prior=p/'workplace/prose-v1/cave-shield-inventory-v1/receipt.json';known=json.loads(prior.read_text())
    masks=[]
    for pose in range(17):
        row=next(r for r in known['records']if r['actor_pose']==pose)
        raw=(prior.parent/f'pose-{pose:02d}.indexed').read_bytes();assert hashlib.sha256(raw).hexdigest()==row['actor_key']
        masks.append((row['actor_key'],np.frombuffer(raw,dtype=np.uint8).reshape(24,24)!=255))
    poses=[];overview=Image.new('RGB',(3*240,4*240),(24,30,40));draw=ImageDraw.Draw(overview)
    for n,r in enumerate(sources):
        raw_path=src_path.parent/(r['file']+'.raster');raw=raw_path.read_bytes();assert sha(raw_path)==r['key']
        native=Image.open(src_path.parent/(r['file']+'.png')).convert('RGBA')
        assert hashlib.sha256(native.tobytes()).hexdigest()==r['rgba_sha256']
        x,y=n%3,n//3;cell=(x*atlas.width//3,cuts[x][y],(x+1)*atlas.width//3,cuts[x][y+1])
        part=atlas.crop(cell);box=part.getchannel('A').point(lambda value:255 if value>=96 else 0).getbbox()
        assert box and box[0]>0 and box[1]>0 and box[2]<part.width and box[3]<part.height,'有效墨跡碰到格界'
        box=(max(0,box[0]-2),max(0,box[1]-2),min(part.width,box[2]+2),min(part.height,box[3]+2))
        crop=[cell[0]+box[0],cell[1]+box[1],cell[0]+box[2],cell[1]+box[3]]
        bounds=r['bounds'];native_box=[bounds['Min']['X'],bounds['Min']['Y'],bounds['Max']['X'],bounds['Max']['Y']]
        x0,y0,x1,y1=native_box;pixels=np.zeros((72,72,4),dtype=np.uint8)
        pixels[y0*3:y1*3,x0*3:x1*3]=np.asarray(area_image(atlas.crop(tuple(crop)),(x1-x0)*3,(y1-y0)*3))
        im=Image.fromarray(pixels,'RGBA');file=out/('player-'+r['key'][:12]+'-72x72.png');im.save(file)
        target=Image.new('RGB',(72,72),(0,0,0));target.paste(im,mask=im.getchannel('A'))
        overview.paste(target.resize((216,216),Image.Resampling.NEAREST),(x*240+12,y*240+24));draw.text((x*240+8,y*240+6),r['key'][:12],fill='white')
        mask=np.frombuffer(raw,dtype=np.uint8).reshape(24,24)!=255
        diff,key=min((int(np.count_nonzero(mask!=old)),key)for key,old in masks)
        poses.append({'key':r['key'],'file':file.name,'sha256':sha(file),'rgba_sha256':hashlib.sha256(pixels.tobytes()).hexdigest(),
                      'native_bounds':native_box,'crop':crop,'cell':cell,'source_prefix':r['source_prefix'],'source':r['source'],
                      'known_shape_differences':diff,'closest_shape_key':key,'palette_variant':n>=6,'size':[72,72]})
    overview.save(out/'player-hd-overview.png')
    report={'schema':1,'status':'REGISTERED_TWELVE_SAVED_PLAYER_CANDIDATES','poses':poses,'script_sha256':sha(Path(__file__)),
            'source_receipt_sha256':sha(src_path),'generation_sha256':sha(gen_path),'known_shield_receipt_sha256':sha(prior),
            'crop_policy':{'ink_alpha_threshold':96,'antialias_padding':2,'row_boundary_max_alpha':1,'row_cuts':cuts},
            'limits':'直接依原生包絡取樣，不經24x24縮圖；只為已存來源圖面，不推定色彩變化原因、正常新路徑或全遊戲完成。'}
    (out/'manifest.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(report['status'],[(r['key'][:12],r['sha256'],r['native_bounds'])for r in poses])


if __name__=='__main__':main()

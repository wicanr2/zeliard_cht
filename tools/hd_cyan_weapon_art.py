#!/usr/bin/env python3
"""把已保存來源的十種青藍刀光註冊成96x96直接高清圖面。"""
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
    b=p/'workplace/full-hd-v1';source_path=b/'cave-saved-gear-source-v1/receipt.json'
    source=json.loads(source_path.read_text());assert source['status']=='SAVED_GEAR_SOURCE_ONLY'
    generation_path=b/'cave-cyan-weapon-art-v1/generation.json';generation=json.loads(generation_path.read_text())
    assert generation['mode']=='built-in imagegen'
    for ref in generation['references']:assert sha(p/ref['file'])==ref['sha256']
    path=p/generation['output_file'];assert sha(path)==generation['output_sha256']
    atlas=Image.open(path).convert('RGBA');assert atlas.size==(1254,1254) and atlas.getchannel('A').getextrema()==(0,255)
    registration_path=p/'workplace/cave-hd-v1/weapon-art-preview-v1/registration.json'
    old={r['key']:r for r in json.loads(registration_path.read_text())['poses']}
    def rgba(row):
        raw=(source_path.parent/(row['file']+'.raster')).read_bytes();assert hashlib.sha256(raw).hexdigest()==row['key']
        im=Image.open(source_path.parent/(row['file']+'.png')).convert('RGBA')
        assert hashlib.sha256(im.tobytes()).hexdigest()==row['rgba_sha256']
        return np.asarray(im)
    known=[(r,rgba(r))for r in source['known_weapons']if r['key']in old]
    rows=[];overview=Image.new('RGB',(5*200,2*160),(24,30,40));draw=ImageDraw.Draw(overview)
    for r in [x for x in source['sources']if x['kind']=='weapon']:
        native=rgba(r);matches=[(k,a)for k,a in known if np.array_equal(native[:,:,3]>0,a[:,:,3]>0)]
        assert len(matches)==1,'輪廓未唯一符合既有圖樣'
        k,a=matches[0];prior=old[k['key']]
        diff=int(np.count_nonzero(np.any(native!=a,axis=2)));assert diff>0
        bounds=r['bounds'];box=[bounds['Min']['X'],bounds['Min']['Y'],bounds['Max']['X'],bounds['Max']['Y']]
        assert box==prior['native_bounds']
        x0,y0,x1,y1=box;crop=prior['crop']
        pixels=np.zeros((96,96,4),dtype=np.uint8)
        pixels[y0*3:y1*3,x0*3:x1*3]=np.asarray(area_image(atlas.crop(tuple(crop)),(x1-x0)*3,(y1-y0)*3))
        assert np.count_nonzero(pixels[:,:,3])>0
        im=Image.fromarray(pixels,'RGBA');file=out/('weapon-'+r['key'][:12]+'-96x96.png');im.save(file)
        n=len(rows);x,y=n%5*200,n//5*160
        preview=Image.new('RGB',(96,96),(0,0,0));preview.paste(im,mask=im.getchannel('A'))
        overview.paste(preview.resize((192,128),Image.Resampling.NEAREST),(x+4,y+24));draw.text((x+5,y+5),r['key'][:12],fill='white')
        rows.append({'key':r['key'],'file':file.name,'sha256':sha(file),'rgba_sha256':hashlib.sha256(pixels.tobytes()).hexdigest(),
                     'native_bounds':box,'crop':crop,'known_reference_key':k['key'],'native_shape_differences':0,'native_rgba_differences':diff,
                     'source_prefix':r['source_prefix'],'source':r['source']})
    assert len(rows)==10
    overview.save(out/'cyan-weapon-hd-overview.png')
    report={'schema':1,'status':'REGISTERED_TEN_CYAN_WEAPON_CANDIDATES','poses':rows,
            'source_receipt_sha256':sha(source_path),'baseline_registration_sha256':sha(registration_path),
            'generation_sha256':sha(generation_path),'script_sha256':sha(Path(__file__)),
            'limits':'十種青藍刀光只綁完整來源鍵；資料與圖面註冊不當正式接入或新正常路徑驗收。'}
    (out/'manifest.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(report['status'],[(r['key'][:12],r['sha256'])for r in rows])


if __name__=='__main__':main()

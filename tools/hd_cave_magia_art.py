#!/usr/bin/env python3
"""將內建生成的不透明結晶材質直接註冊至原版圖格102的24x24輪廓。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image
from hd_mpp1_art import area_image


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project',type=Path,required=True)
    ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args();p,out=args.project,args.output
    assert out.is_dir() and not any(out.iterdir())
    sha=lambda path:hashlib.sha256(path.read_bytes()).hexdigest()
    b=p/'workplace/full-hd-v1'
    gen_path=b/'cave-magia-art-v1/generation.json';generation=json.loads(gen_path.read_text())
    assert generation['mode']=='built-in imagegen' and not generation['production_runtime_changed']
    for ref in generation['reference_inputs']:assert sha(p/ref['file'])==ref['sha256']
    request=gen_path.parent/generation['request_file'];assert sha(request)==generation['request_sha256']
    image_path=p/generation['output_file'];assert sha(image_path)==generation['output_sha256']=='cdb7f499c247ca6cd14224121e7d0e8cae18262e0418abc56c38b397e7f4c44c'
    image=Image.open(image_path).convert('RGBA');assert image.size==(1254,1254) and image.getchannel('A').getextrema()==(255,255)
    source_path=b/'cave-magia-source-v1/point-000.memory';assert sha(source_path)=='6b758b3b0f63b0697d82ad7fa7b69fb5fff9d9bf2fb983e2eb0a0ed6e5c98034'
    memory=source_path.read_bytes();raw=memory[0x11b10+0x9320:0x11b10+0x9350]
    colors=[]
    for at in range(0,48,3):
        a,c,d=raw[at:at+3];colors.extend([c>>2,(c&3)*16+(a>>4),(a&15)*4+(d>>6),d&63])
    mask=np.repeat(np.repeat(np.array(colors,dtype=np.uint8).reshape(8,8)!=0,3,axis=0),3,axis=1)
    assert np.count_nonzero(mask)==414
    pixels=np.asarray(area_image(image,24,24)).copy();pixels[~mask]=0
    assert np.all(pixels[:,:,3][mask]==255) and not np.any(pixels[~mask])
    path=out/'magia-shard-24x24.png';Image.fromarray(pixels,'RGBA').save(path)
    report={'schema':1,'status':'REGISTERED_MAGIA_SHARD_CANDIDATE_NOT_CONNECTED','source_id':102,
            'source_file':str(source_path.relative_to(p)),'source_sha256':sha(source_path),
            'source_tile_address':'11B1:9320..934F','source_tile_sha256':hashlib.sha256(raw).hexdigest(),
            'generation_file':str(gen_path.relative_to(p)),'generation_sha256':sha(gen_path),
            'script_sha256':sha(Path(__file__)),'art_file':path.name,'art_sha256':sha(path),
            'size':[24,24],'opaque_pixels':414,'transparent_inside_mask':0,'outside_mask_written_pixels':0,
            'policy':'不透明材質直接取樣至24x24，再依原版102非零色號輪廓裁形，不經8x8縮圖。',
            'limits':'僅圖面註冊，正式來源守衛、正常HD配對及GUI待驗。'}
    (out/'manifest.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(report['status'],report['art_sha256'],414)


if __name__=='__main__':main()

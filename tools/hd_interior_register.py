#!/usr/bin/env python3
"""註冊內建imagegen室內圖面，固定共用背景與原版來源別名；完整素材留本機。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageDraw
from hd_mpp1_art import area_image


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project',type=Path,required=True)
    ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args(); project=args.project; output=args.output
    assert output.is_dir() and not any(output.iterdir())
    sha=lambda path:hashlib.sha256(path.read_bytes()).hexdigest()
    base=project/'workplace/full-hd-v1/interior-art-v1'
    generation=base/'generation-v2.json'; meta=json.loads(generation.read_text())
    assert meta['mode']=='builtin-imagegen' and len(meta['requests'])==6
    entries={row['key']:row for row in meta['requests']}
    def read(key):
        row=entries[key]; path=base/row['output']['file']
        assert sha(path)==row['output']['sha256'] and row['review'].startswith('selected')
        image=Image.open(path).convert('RGBA');assert image.getchannel('A').getextrema()==(255,255)
        return image
    hut=area_image(read('hut-pose-correction'),408,384)
    inn=area_image(read('inn-base'),288,192)
    images={'hut':hut,'inn-base':inn}
    # 只將已核角色區的變化貼回同一背景；不是重新繪畫或改原始生成圖。
    regions={'inn-ambient-1':(6,8,30,32),'inn-rest-2':(20,21,48,47),'inn-rest-3':(20,21,48,47)}
    for key,box in regions.items():
        variant=area_image(read(key),288,192); registered=inn.copy()
        region=tuple(x*3 for x in box); registered.paste(variant.crop(region),region[:2])
        for y in range(192):
            for x in range(288):
                if not (region[0]<=x<region[2] and region[1]<=y<region[3]):
                    assert registered.getpixel((x,y))==inn.getpixel((x,y))
        assert registered.tobytes()!=inn.tobytes()
        images[key]=registered
    native=project/'workplace/full-hd-v1/interior-layout-v2'
    sources=json.loads((native/'receipt.json').read_text()); rows=[]
    mapping={'hut':['OMOYA-base'],'inn-base':['INN-base','INN-ambient-0','INN-rest-0','INN-rest-1'],
             'inn-ambient-1':['INN-ambient-1'],'inn-rest-2':['INN-rest-2'],'inn-rest-3':['INN-rest-3']}
    contact=Image.new('RGB',(1024,1408),'#25303a');draw=ImageDraw.Draw(contact)
    for n,(key,image) in enumerate(images.items()):
        target=output/(key+'-hd-v1.png');image.save(target)
        originals=[Image.open(native/(name+'-original.png')).convert('RGB') for name in mapping[key]]
        assert all(originals[0].tobytes()==im.tobytes() for im in originals)
        y=n*264+24
        original=originals[0];shown=image.convert('RGB')
        if key=='hut':
            original=original.resize((204,192),Image.Resampling.NEAREST);shown=shown.resize((204,192),Image.Resampling.NEAREST)
        else:
            original=original.resize((288,192),Image.Resampling.NEAREST)
        contact.paste(original,(32,y));contact.paste(shown,(544,y));draw.text((32,y-16),key+' original',fill='white');draw.text((544,y-16),key+' HD',fill='white')
        rows.append({'key':key,'file':target.name,'sha256':sha(target),'size':list(image.size),
                     'native_frames':mapping[key],'native_rgb_sha256':hashlib.sha256(originals[0].tobytes()).hexdigest(),
                     'change_region_native':regions.get(key),
                     'native_frame_sha256':[sha(native/(name+'-original.png')) for name in mapping[key]]})
    contact.save(output/'interior-hd-contact.png')
    original_profiles={row['name']:row for row in sources['records']}
    assert not original_profiles['OMOYA']['unreferenced_tile_ids'] and not original_profiles['INN']['unreferenced_tile_ids']
    report={'schema':1,'status':'REGISTERED_INTERIOR_ART_NORMAL_PATH_PENDING','assets':rows,'native_tiles':247,
            'generation_sha256':sha(generation),'source_receipt_sha256':sha(native/'receipt.json'),
            'script_sha256':sha(Path(__file__)),'area_tool_sha256':sha(project/'tools/hd_mpp1_art.py'),
            'limits':['兩資源247個圖格均被原版索引表使用；註冊不證明正常到達或GUI已驗。',
                      '旅館的共用背景固定；動作變化限已列HD角色區，原版程式及動畫表不改。',
                      '旅館正式覆繪須保留原版一像素外框，這些圖不含新框線。']}
    (output/'manifest.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    print(report['status'],[(x['key'],x['sha256']) for x in rows])


if __name__=='__main__':
    main()

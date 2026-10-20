#!/usr/bin/env python3
"""長老手繪姿態註冊到直接3倍畫布；不回縮到原版解析度。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project',type=Path,required=True)
    ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args();p,out=args.project,args.output
    assert out.is_dir() and not any(out.iterdir())
    root=p/'workplace/full-hd-v1'
    generation=root/'sage-frames-art-v1/generation-v2.json'
    plan=json.loads(generation.read_text())
    files={row['key']:row for row in plan['outputs']}
    def art(key):
        row=files[key];path=p/row['file'];assert sha(path)==row['sha256']
        im=Image.open(path);assert im.mode in ('RGB','RGBA') and im.size==(1536,1024)
        if im.mode=='RGBA':assert im.getchannel('A').getextrema()==(255,255)
        return im.convert('RGB').resize((288,192),Image.Resampling.BOX)
    base_path=p/'workplace/hd-sage-prototype.png'
    assert sha(base_path)=='a1fa3b745422f59153239b903591db104c5c3fb320c18ee7f5181a247c33a5aa'
    # 基底沿既有直接取樣，避免未變更區的畫面漂移。
    base=Image.open(base_path).convert('RGB').resize((288,192),Image.Resampling.NEAREST)
    patch=(48,48,240,144);glint=(168,72,192,96)
    cycle=art('glint-cycle')
    records=[];contact=Image.new('RGB',(288*4,192*4))
    for pose in range(8):
        frame=base.copy()
        if pose:
            candidate=art(f'pose-{pose}-refined' if pose in (2,4,5,6) else f'pose-{pose}')
            frame.paste(candidate.crop(patch),patch[:2])
        for phase in range(2):
            result=frame.copy()
            if phase:result.paste(cycle.crop(glint),glint[:2])
            name=f'sage-pose-{pose}-phase-{phase}.png';result.save(out/name)
            if phase:
                assert result.tobytes()!=frame.tobytes(),'動畫圖面沒有差異'
                # 此動畫來源只畫8×8，高清差異不可溢出。
                restored=result.copy();restored.paste(frame.crop(glint),glint[:2]);assert restored.tobytes()==frame.tobytes()
            restored=frame.copy();restored.paste(base.crop(patch),patch[:2]);assert restored.tobytes()==base.tobytes()
            contact.paste(result,((pose%4)*288,(pose//4*2+phase)*192))
            records.append({'pose':pose,'phase':phase,'file':name,'sha256':sha(out/name),'size':[288,192]})
    contact.save(out/'sage-hd-contact.png')
    report={'schema':1,'status':'REGISTERED_SAGE_EIGHT_HANDS_TWO_HD_ANIMATION_PHASES','script_sha256':sha(Path(__file__)),
            'generation_sha256':sha(generation),'source_layout_sha256':sha(root/'sage-frames-layout-v1/receipt.json'),
            'base_art_sha256':sha(base_path),'assets':records,'target_scale':3,'hand_patch':list(patch),'animation_patch':list(glint),
            'limits':['手繪動畫以原版選圖為準；glint名稱只描述手繪效果，不推定原版單圖格的角色語意。','本工具只註冊圖面，不證明正常前端或原版狀態不變。']}
    (out/'manifest.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    print(report['status'],len(records))


if __name__=='__main__':main()

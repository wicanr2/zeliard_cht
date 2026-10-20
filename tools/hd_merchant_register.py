#!/usr/bin/env python3
"""武器店完整圖組直接3倍註冊，保留既有基底取樣與局部動畫範圍。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image


def sha(path):return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('--project',type=Path,required=True);ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args();p,out=args.project,args.output;assert out.is_dir() and not any(out.iterdir())
    root=p/'workplace/full-hd-v1';generation=root/'merchant-frames-art-v1/generation-v2.json';plan=json.loads(generation.read_text())
    generated={row['key']:row for row in plan['outputs']};layout_path=root/'merchant-frames-layout-v1/receipt.json';layout=json.loads(layout_path.read_text())
    def open_image(path):
        im=Image.open(path);assert im.size==(1536,1024) and im.mode in ('RGB','RGBA')
        if im.mode=='RGBA':assert im.getchannel('A').getextrema()==(255,255)
        return im.convert('RGB')
    def existing(file,digest):
        path=p/'workplace/merchant-hd-v1'/file;assert sha(path)==digest
        array=np.asarray(open_image(path));return Image.fromarray(array[(np.arange(186)*1024//186)[:,None],np.arange(282)*1536//282])
    def painted(key):
        row=generated[key];path=p/row['file'];assert sha(path)==row['sha256'];return open_image(path).resize((282,186),Image.Resampling.BOX)
    base=existing('merchant-prototype.png','2660fc7892f9a5229d4cee3723e992146992a460a1678197793cc890b0dfb59b')
    folded=existing('merchant-farewell-prototype.png','90243f65aef8532303b12dbbc3e1ddd9bd06c03eb58e14548a2d15419892946f')
    folded_phase=painted('folded-animation')
    sheet=Image.new('RGB',(1128,3*186));records=[]
    for i,row in enumerate(layout['frames']):
        pose=row['pose'];key=row['key']
        frame=folded.copy() if pose==2 else base.copy()
        if pose not in (0,2):
            frame=painted(f'body-{pose}')
        animation=None
        if key=='body-0-animation-3':animation=(213,165,261,186);frame.paste(painted('counter-animation').crop(animation),animation[:2])
        elif key=='body-2-animation-3':animation=(93,93,117,141);frame.paste(folded_phase.crop(animation),animation[:2])
        file=key+'-hd.png';frame.save(out/file)
        sheet.paste(frame,(i%4*282,i//4*186));records.append({'key':key,'pose':pose,'file':file,'sha256':sha(out/file),'size':[282,186],'animation_patch':list(animation) if animation else None})
    assert len(records)==10
    sheet.save(out/'merchant-hd-contact.png')
    report={'schema':1,'status':'REGISTERED_MERCHANT_EIGHT_BODIES_AND_TWO_ANIMATIONS','script_sha256':sha(Path(__file__)),
            'generation_sha256':sha(generation),'layout_sha256':sha(layout_path),'target_scale':3,'assets':records,
            'limits':['既有姿態0/2沿原來floor取樣；局部動畫只取其原版圖塊區，不重畫計時或碰撞。','只註冊高清圖，正常顯示及原版狀態另驗。']}
    (out/'manifest.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    print(report['status'],len(records))


if __name__=='__main__':main()

#!/usr/bin/env python3
"""Docker內製作使用者指定的怪物／魔王README圖庫縮圖，不改原始圖面。"""
import argparse
import hashlib
import json
from pathlib import Path
from PIL import Image, ImageDraw, ImageFont, ImageOps


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project',type=Path,required=True)
    args=ap.parse_args()
    root=args.project
    work=root/'workplace/full-hd-v1/manual-hd-v1'
    out=root/'docs/images'
    assert out.is_dir()
    digest=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
    fontfile=work/'rounded-font/jf-openhuninn-2.1.ttf'
    fonts={n:ImageFont.truetype(str(fontfile),n) for n in (20,24,30,40)}
    plan=json.loads((work/'plan.json').read_text())
    selected=json.loads((work/'registered/manual-art.json').read_text())['selected_generations']
    sources={};outputs={}

    def picture(path):
        sources[str(path.relative_to(root))]={'sha256':digest(path)}
        return Image.open(path).convert('RGBA')

    def place(canvas,art,box,pixel=False):
        x,y,w,h=box
        im=ImageOps.contain(art,(w,h),Image.Resampling.NEAREST if pixel else Image.Resampling.LANCZOS)
        canvas.paste(im,(x+(w-im.width)//2,y+(h-im.height)//2),im)

    for bosses in (False,True):
        entries=[x for x in plan['assets'] if x['key'] in selected and x['key'].startswith('enp')!=bosses]
        assert len(entries)==(10 if bosses else 8)
        height=160+440*((len(entries)+1)//2)+70
        canvas=Image.new('RGB',(1600,height),'#101923')
        d=ImageDraw.Draw(canvas)
        title='魔王與首領圖庫總覽' if bosses else '怪物圖庫總覽'
        d.text((32,20),title,font=fonts[40],fill='#f3eee1')
        d.text((32,82),'原版來源與手繪 HD；各圖庫含動畫姿態、局部元件與效果。',font=fonts[24],fill='#bac5cf')
        for i,e in enumerate(entries):
            x,y=24+(i%2)*788,144+(i//2)*440
            d.rounded_rectangle((x,y,x+764,y+420),radius=12,fill='#1c2936')
            d.text((x+20,y+12),e.get('program_name') or e['key'].upper(),font=fonts[30],fill='#e3b66a')
            d.text((x+20,y+58),'原版',font=fonts[20],fill='#bac5cf')
            d.text((x+402,y+58),'手繪 HD',font=fonts[20],fill='#bac5cf')
            reference=root/e['reference']
            art=work/'generated'/f'{e["key"]}-generated-v{selected[e["key"]]}.png'
            place(canvas,picture(reference),(x+12,y+92,360,312),True)
            place(canvas,picture(art),(x+394,y+92,358,312))
        d.text((32,height-56),'文件縮圖，非完整解析度素材；圖庫數不等於敵人種類數，後段動畫仍待人工檢視。',font=fonts[20],fill='#bac5cf')
        file=out/('bosses-overview.jpg' if bosses else 'enemies-overview.jpg')
        canvas.save(file,quality=88,subsampling=0,optimize=True)
        outputs[file.name]={'sha256':digest(file),'bytes':file.stat().st_size,'size':list(canvas.size),'families':[e['key'] for e in entries]}
    manifest={'scope':'使用者指定加入私人GitHub README的圖庫文件縮圖；原版與衍生美術不適用自有程式授權','pillow':Image.__version__,'tool_sha256':digest(Path(__file__)),'font_sha256':digest(fontfile),'plan_sha256':digest(work/'plan.json'),'selected_sha256':digest(work/'registered/manual-art.json'),'sources':sources,'outputs':outputs}
    (out/'creatures-provenance.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(outputs,ensure_ascii=False))


if __name__=='__main__':main()

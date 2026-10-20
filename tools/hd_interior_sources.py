#!/usr/bin/env python3
"""依原版室內圖塊表重建構圖；正常長老為格式對照，不冒稱旅館可達。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageDraw
from hd_planar import convert


def digest(data):
    return hashlib.sha256(data).hexdigest()


def tiles(raw):
    assert len(raw)%48==0
    packed=convert(raw); pixels=[]
    for a,b,c in zip(packed[::3],packed[1::3],packed[2::3]):
        pixels.extend([b>>2,((b&3)<<4)|(a>>4),((a&15)<<2)|(c>>6),c&63])
    return packed,[bytes(pixels[i:i+64]) for i in range(0,len(pixels),64)]


def raster(atlas, ids, columns, rows):
    assert len(ids)==columns*rows and max(ids)<len(atlas)
    width=columns*8; out=bytearray(width*rows*8)
    for n,id in enumerate(ids):
        for y in range(8):
            start=(n//columns*8+y)*width+n%columns*8
            out[start:start+8]=atlas[id][y*8:y*8+8]
    return bytes(out)


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project',type=Path,required=True)
    ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args(); project=args.project; output=args.output
    assert output.is_dir() and not any(output.iterdir())
    sar=(project/'workplace/orig/ZELRES2.SAR').read_bytes()
    assert digest(sar)=='bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1'
    ledger=json.loads((project/'docs/hd-coverage.json').read_text())
    entries={row['record']:row for row in ledger['records'] if row['archive']=='ZELRES2.SAR'}
    pal=(project/'workplace/music-release-v1/release-build/data/assets/town/frame/palette.bin').read_bytes()
    assert digest(pal)=='6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c'
    def module(number):
        row=entries[number];raw=sar[row['payload_offset']:row['payload_offset']+row['payload_bytes']]
        assert digest(raw)==row['payload_sha256']
        return raw
    reports=[]; contact=Image.new('RGB',(1536,1536),'#25303a'); draw=ImageDraw.Draw(contact)
    profiles=[('OMOYA',11,19,0xa129,17,16,(96,30)),('INN',16,24,0xa1cf,12,8,(56,23)),
              ('KENJYA',17,25,0xa9b6,12,8,(56,23))]
    cell=0
    for name,program,graphics,table,columns,rows,origin in profiles:
        code=module(program)
        source=project/f'workplace/full-hd-v1/decoded-v2/zelres2-{graphics:02d}.decoded'
        raw=source.read_bytes();assert digest(raw)==entries[graphics]['decoder']['sha256']
        packed,atlas=tiles(raw)
        indices=code[table-0xa000:table-0xa000+columns*rows]
        base=raster(atlas,indices,columns,rows)
        frames=[('base',base)]
        frame_maps=[{'name':'base','table':hex(table),'columns':columns,'rows':rows,'origin':origin,'ids':list(indices)}]
        if name=='INN':
            for kind,offset,w,h,count in [('ambient',0xa279,2,2,2),('rest',0xa281,5,4,4)]:
                for phase in range(count):
                    ids=code[offset-0xa000+phase*w*h:offset-0xa000+(phase+1)*w*h]
                    pixels=raster(atlas,ids,w,h); result=bytearray(base)
                    # 原版BX=0827，像素位置64,39，位於原96×64圖內的8,16。
                    for y in range(h*8):
                        at=(16+y)*96+8; result[at:at+w*8]=pixels[y*w*8:(y+1)*w*8]
                    label=f'{kind}-{phase}';frames.append((label,bytes(result)))
                    frame_maps.append({'name':label,'table':hex(offset+phase*w*h),'columns':w,'rows':h,
                                       'origin':[64,39],'ids':list(ids)})
        normal=None
        if name in ('KENJYA','OMOYA'):
            folder,sample=('home-normal-v1','home-stable') if name=='KENJYA' else ('hut-normal-v2','hut-stable')
            study=project/'workplace/full-hd-v1'/folder
            meta=json.loads((study/(sample+'.json')).read_text());memory=(study/(sample+'.memory')).read_bytes()
            assert digest(memory)==meta['memory_sha256']
            p,r=memory[0x1b10:0x11b10],memory[0x11b10:0x21b10]
            assert p[table:table+columns*rows]==indices and r[0x8000:0x8000+len(packed)]==packed
            indexed=(study/(sample+'.indexed')).read_bytes();positions=[]
            for x,y in ((56,23),(112,23)) if name=='KENJYA' else ((96,30),):
                crop=b''.join(indexed[(y+dy)*320+x:(y+dy)*320+x+columns*8] for dy in range(rows*8))
                if crop==base:positions.append([x,y])
            assert len(positions)==1,positions
            normal={'memory_sha256':meta['memory_sha256'],'indexed_sha256':digest(indexed),
                    'packed_bytes':len(packed),'pixels':columns*rows*64,'origin':positions[0],
                    'folder':folder,'sample':sample,
                    'scope':'正常冷F7自建存檔進長老' if name=='KENJYA' else '正常冷新局走到城堡亭台按Up進小屋；未注入狀態'}
        all_ids=set()
        for row in frame_maps:all_ids.update(row['ids'])
        for label,pixels in frames:
            image=Image.frombytes('P',(columns*8,rows*8),pixels);image.putpalette(pal)
            image.convert('RGB').save(output/f'{name}-{label}-original.png')
            (output/f'{name}-{label}.indexed').write_bytes(pixels)
            scale=3 if rows==16 else 4
            reference=image.convert('RGB').resize((columns*8*scale,rows*8*scale),Image.Resampling.NEAREST)
            reference.save(output/f'{name}-{label}-reference.png')
            x,y=(cell%3)*512,(cell//3)*448
            draw.text((x+24,y+8),name+' '+label,fill='white')
            contact.paste(reference,(x+24,y+32));cell+=1
        reports.append({'name':name,'program_record':program,'program_sha256':digest(code),'graphics_record':graphics,
                        'graphics_sha256':digest(raw),'packed_sha256':digest(packed),'tile_count':len(atlas),
                        'covered_tile_ids':sorted(all_ids),'unreferenced_tile_ids':sorted(set(range(len(atlas)))-all_ids),
                        'frame_maps':frame_maps,'normal_binding':normal,
                        'frames':[{'name':label,'indexed_sha256':digest(pixels),'pixels':len(pixels)} for label,pixels in frames],
                        'source_level':'正常來源已核' if normal else '原版靜態consumer已核，正常入口未知'})
    contact.save(output/'interior-contact.png')
    report={'schema':1,'status':'PASS_INTERIOR_STATIC_LAYOUTS_AND_NORMAL_SAGE_BINDING','records':reports,
            'script_sha256':digest(Path(__file__).read_bytes()),'palette_sha256':digest(pal),
            'address_space':'各室內BIN的base A000與SAR檔案偏移分列；GRP為解壓buffer偏移，非程式位址',
            'limits':['民房／旅館原版consumer與圖表只證結構，正常入口與實際視窗未驗。',
                      '旅館ambient與rest名稱僅導覽；不推定未驗觸發、亂數或實際事件時間。']}
    (output/'receipt.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    print(report['status'],[(x['name'],x['tile_count'],x['unreferenced_tile_ids']) for x in reports])


if __name__=='__main__':
    main()

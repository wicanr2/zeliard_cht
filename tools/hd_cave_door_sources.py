#!/usr/bin/env python3
"""盤點沒有城鎮招牌的共用門框；只核來源與圖面，不推定通行規則。"""
import argparse
import hashlib
import json
import struct
from pathlib import Path

from PIL import Image, ImageDraw
from hd_cave_maps import decode_columns


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project',type=Path,required=True)
    ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args();p,out=args.project,args.output
    assert out.is_dir()and not any(out.iterdir())
    sha=lambda path:hashlib.sha256(path.read_bytes()).hexdigest()
    read=lambda path:json.loads(path.read_text())
    b=p/'workplace/full-hd-v1';normal_path=b/'cave-station-source-v1/receipt.json';normal=read(normal_path)
    assert normal['status']=='PASS_GUI_JOURNAL_SOURCE_RECONSTRUCTION'and not normal['terrain_hd_matched']
    memory_path=normal_path.parent/'source.memory';assert sha(memory_path)==normal['memory_sha256']
    memory=memory_path.read_bytes();resident=memory[0x1b10:0x11b10];resource=memory[0x11b10:0x21b10]
    assert hashlib.sha256(resource[0x8c00:0x9350]).hexdigest()=='7b4628728e0b6ce81a25b7bc0bbb3b4eedd640b0da102924b4384da0d38f690b'
    palette_path=normal_path.parent/'source.palette';palette=palette_path.read_bytes();assert len(palette)==768
    # 這些門框不是RLE地形格，而由78DD消費C00A的12byte記錄後寫入快取。
    templates={}
    for name,address in [('filled',0x79b4),('open',0x79c8)]:
        values=list(resident[address:address+20]);values[2]=97
        templates[name]=[values[i:i+5]for i in range(0,20,5)]
    assert templates['filled']==[[73,74,97,75,76],[77,79,80,81,78],[95,82,83,84,96],[95,85,86,87,96]]
    assert templates['open']==[[73,74,97,75,76],[77,88,0,89,78],[95,90,0,91,96],[95,92,93,94,96]]
    def matrix(marker,kind):return [[marker if x==2 and y==0 else value for x,value in enumerate(row)]for y,row in enumerate(templates[kind])]
    def tile(id):
        colors=[];raw=resource[0x8030+(id-1)*48:0x8030+id*48]
        for at in range(0,48,3):
            a,c,d=raw[at:at+3];colors.extend([c>>2,(c&3)*16+(a>>4),(a&15)*4+(d>>6),d&63])
        return colors
    variants=[];sheet=Image.new('RGB',(2*360,5*330),(22,29,40));draw=ImageDraw.Draw(sheet)
    for marker,kind in [(marker,kind)for marker in range(97,102)for kind in ('filled','open')]:
        ids=matrix(marker,kind);im=Image.new('RGB',(40,32))
        packed=bytearray()
        for y,row in enumerate(ids):
            for x,id in enumerate(row):
                if not id:continue
                packed.extend(resource[0x8030+(id-1)*48:0x8030+id*48])
                for n,color in enumerate(tile(id)):
                    im.putpixel((x*8+n%8,y*8+n//8),tuple(palette[color*3:color*3+3]))
        path=out/f'door-{marker}-{kind}-original.png';im.save(path)
        reference=out/f'door-{marker}-{kind}-reference.png';im.resize((640,512),Image.Resampling.NEAREST).save(reference)
        x,y=(0 if kind=='filled'else 1)*360,(marker-97)*330
        draw.text((x+12,y+8),f'ID {marker} / {kind}',fill='white');sheet.paste(im.resize((320,256),Image.Resampling.NEAREST),(x+12,y+40))
        variants.append({'marker':marker,'kind':kind,'matrix':ids,'native_size':[40,32],'file':path.name,'sha256':sha(path),'rgb_sha256':hashlib.sha256(im.tobytes()).hexdigest(),'packed_matrix_sha256':hashlib.sha256(packed).hexdigest(),'reference_file':reference.name,'reference_sha256':sha(reference)})
    sheet.save(out/'doors-overview.png')
    map_path=b/'cave-map-sources-v1/receipt.json';maps=read(map_path);placements=[]
    for row in maps['records']:
        source_path=map_path.parent/row['source'];columns_path=map_path.parent/row['columns']
        assert sha(source_path)==row['payload_sha256']and sha(columns_path)==row['columns_sha256']
        columns,_,_=decode_columns(source_path.read_bytes());assert b''.join(columns)==columns_path.read_bytes()
        width=len(columns);data=source_path.read_bytes()
        start,end=struct.unpack_from('<HH',data,10)
        table=data[start-0xc000:end-0xc000]
        assert len(table)>=2 and (len(table)-2)%12==0 and table[-2:]==b'\xff\xff'
        for at in range(0,len(table)-2,12):
            record=table[at:at+12];x=struct.unpack_from('<H',record)[0];y=record[2];flags=record[3]
            assert x<width and y<64 and (flags&7)<=4
            placements.append({'map':row['name'],'record':row['record'],'slot':at//12,'table_address':hex(start+at),
                               'stored_x':x,'x':(x-1)%width,'y':y,'marker':97+(flags&7),
                               'kind':'open'if flags&128 else'filled','flags':flags,'source_sha256':row['payload_sha256'],
                               'record_sha256':hashlib.sha256(record).hexdigest()})
    # 正常快取中的新門框完整四列；不以可見的單一右柱當證據。
    origin=struct.unpack_from('<H',resident,0xff31)[0]
    def cell(x,y):
        id=resident[0xe000+(origin-0xe000+y*36+4+x)%0x900]
        return resident[0xed20+(id&127)]if id&128 else id
    anchor=(-4,3);observed=[[cell(anchor[0]+x,anchor[1]+y)for x in range(5)]for y in range(4)]
    assert observed==matrix(98,'open')
    assert [cell(anchor[0]+x,anchor[1]-1)for x in range(5)]==[0]*5
    world_x=(struct.unpack_from('<H',resident,128)[0]+4+anchor[0])%240;world_y=(resident[130]+anchor[1])&63
    assert any(r['map']=='MP10.MDT'and(r['x'],r['y'],r['marker'])==(world_x,world_y,98)for r in placements)
    assert len(placements)==163
    report={'schema':1,'status':'BOUND_UNTITLED_CAVE_DOOR_SOURCE','script_sha256':sha(Path(__file__)),
            'normal_receipt':str(normal_path.relative_to(p)),'normal_receipt_sha256':sha(normal_path),
            'memory_sha256':sha(memory_path),'palette_sha256':sha(palette_path),
            'map_receipt':str(map_path.relative_to(p)),'map_receipt_sha256':sha(map_path),'variants':variants,'placements':placements,
            'normal_anchor_native':[16,38],'normal_world':[world_x,world_y],'normal_marker':98,
            'normal_cache_matrix':observed,'normal_preceding_row':[0]*5,'templates':templates,
            'constructor_code_address':'01B1:78DD..79B3','constructor_code_sha256':hashlib.sha256(resident[0x78dd:0x79b4]).hexdigest(),
            'address_space':'正常dosgolem 01B1循環地圖及11B1資源；靜態地圖另為SAR3各payload從0起算。',
            'limits':['原版執行仍處理通行、鑰匙與場景切換，顏色不當規則證據。','正常只驗98/open；其他標記及filled只核原版constructor、資料與固定圖面，不當正常到達或HD完成。']}
    (out/'receipt.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(report['status'],len(variants),'variants',len(placements),'static placements',report['normal_world'])


if __name__=='__main__':main()

#!/usr/bin/env python3
"""武器店八姿態與兩組局部動畫的來源，分列正常、混合與未引用格。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageDraw
from hd_interior_sources import tiles, raster


def sha(data):return hashlib.sha256(data).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project',type=Path,required=True);ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args();p,out=args.project,args.output;assert out.is_dir() and not any(out.iterdir())
    root=p/'workplace/full-hd-v1';code=(root/'merchant-frames-source-v1/ARMRPRO.BIN').read_bytes()
    assert sha(code)=='f10d874783703f8385f36cc579b956b36e9d326af0c8c2150ee886647fe53961'
    raw=(root/'decoded-v2/zelres2-20.decoded').read_bytes();assert sha(raw)=='3dcb0eb6230c4399e8a95b013c8d0c3d525329d4cb0e8664c0f91c760f7549c3'
    packed,atlas=tiles(raw);assert len(atlas)==256 and sha(packed)=='b908b4b98dece29dfeba950cc8b996dcf89412bfa5a3286254d8ea5a5ca3a6ed'
    palette=(p/'workplace/music-release-v1/release-build/data/assets/town/frame/palette.bin').read_bytes()
    assert sha(palette)=='6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c'
    rows=[];covered=set();frames={};aliases=[];base_frames=[];descriptors=[]
    for pose in range(8):
        frame=bytearray();tables=[]
        for group in range(2):
            at=0xa10+pose*6+group*3;h=code[at];table=int.from_bytes(code[at+1:at+3],'little')
            if not h:break
            ids=code[table-0xa000:table-0xa000+h*12];assert len(ids)==h*12
            covered.update(ids);frame.extend(raster(atlas,ids,12,h));tables.append({'address':hex(table),'rows':h,'columns':12,'ids':list(ids)})
        assert len(frame)==6144
        base_frames.append(bytes(frame));descriptors.append({'pose':pose,'descriptor':hex(0xaa10+pose*6),'tables':tables})
        candidates=[(f'body-{pose}',bytes(frame))]
        if pose in (0,2):
            offset,w,h,x,y=(0xad0,2,1,72,56) if pose==0 else (0xb68,1,2,32,32)
            for phase in range(4):
                ids=code[offset+phase*2:offset+(phase+1)*2];covered.update(ids)
                patch=raster(atlas,ids,w,h);pixels=bytearray(frame)
                for dy in range(h*8):pixels[(y+dy)*96+x:(y+dy)*96+x+w*8]=patch[dy*w*8:(dy+1)*w*8]
                candidates.append((f'body-{pose}-animation-{phase}',bytes(pixels)))
        for name,pixels in candidates:
            digest=sha(pixels)
            if digest in frames:
                aliases.append({'alias':name,'canonical':frames[digest]['key']});continue
            record={'key':name,'pose':pose,'indexed_sha256':digest,'normal_sources':[]}
            frames[digest]=record;rows.append(record);(out/(name+'.indexed')).write_bytes(pixels)
            im=Image.frombytes('P',(96,64),pixels);im.putpalette(palette);im.convert('RGB').save(out/(name+'-original.png'))
    assert len(frames)==10 and covered==set(range(256))-{139,140}
    native_bounds=[]
    for pose,pixels in enumerate(base_frames):
        delta=[(i%96,i//96) for i,(a,b) in enumerate(zip(pixels,base_frames[0])) if a!=b]
        bounds=[min(x for x,y in delta),min(y for x,y in delta),max(x for x,y in delta)+1,max(y for x,y in delta)+1] if delta else None
        native_bounds.append({'pose':pose,'difference_bounds':bounds})
    normal=[];rejections=[];guards=[]
    for folder in ('merchant-frames-normal-v1','merchant-frames-normal-v2'):
        receipt_path=root/folder/'receipt.json';receipt=json.loads(receipt_path.read_text())
        assert receipt['source_save_unchanged'] and receipt['scratch_save_unchanged'] and not receipt['snapshot_restore'] and not receipt['state_injection']
        assert sha((root/folder/'probe-source.go').read_bytes())==receipt['script_sha256']
        for sample in receipt['frames']:
            memory=(root/folder/(sample['file']+'.memory')).read_bytes();indexed=(root/folder/(sample['file']+'.indexed')).read_bytes()
            assert sha(memory)==sample['memory_sha256'] and sha(indexed)==sample['indexed_sha256']
            if memory[0x1b10+0xa000:0x1b10+0xbbfd]!=code[:0x1bfd] or memory[0x11b10+0x8000:0x11b10+0xb000]!=packed:
                guards.append({'folder':folder,'file':sample['file'],'reason':'模組或圖格已變更，不能作室內完整來源'});continue
            crop=b''.join(indexed[(23+y)*320+56:(23+y)*320+152] for y in range(64));assert sha(crop)==sample['crop_sha256']
            record=frames.get(sha(crop))
            if record:
                binding={'folder':folder,'file':sample['file'],'key':record['key'],'memory_sha256':sample['memory_sha256'],'indexed_sha256':sample['indexed_sha256']}
                normal.append(binding);record['normal_sources'].append(folder+'/'+sample['file'])
            else:rejections.append({'folder':folder,'file':sample['file'],'reason':'未完成或混合畫格，不取代完整來源'})
    sheet=Image.new('RGB',(1152,4*288),'#25303a');draw=ImageDraw.Draw(sheet)
    for i,row in enumerate(rows):
        im=Image.open(out/(row['key']+'-original.png')).resize((384,256),Image.Resampling.NEAREST)
        x,y=i%3*384,i//3*288;sheet.paste(im,(x,y+24));draw.text((x,y+5),row['key'],fill='white')
    sheet.save(out/'merchant-original-contact.png')
    report={'schema':1,'status':'PASS_MERCHANT_TABLES_AND_NORMAL_BINDINGS','script_sha256':sha(Path(__file__).read_bytes()),
            'module_sha256':sha(code),'immutable_module_prefix_bytes':0x1bfd,'immutable_module_prefix_sha256':sha(code[:0x1bfd]),
            'graphics_sha256':sha(raw),'packed_sha256':sha(packed),'palette_sha256':sha(palette),'tile_count':256,'referenced_tile_count':254,
            'unreferenced_tiles':[139,140],'unreferenced_level':'八個完整描述及兩組局部動畫未引用；未將非空圖格稱空白或HD已完成',
            'descriptors':descriptors,'body_differences':native_bounds,'frames':rows,'aliases':aliases,
            'normal_poses':sorted({row['pose'] for row in rows if row['normal_sources']}),
            'normal_bindings':normal,'normal_mixed_rejections':rejections,'normal_source_guard_rejections':guards,
            'limits':['完整表及直接call3016消費端只證本模組已知圖面；未引用兩格不宣稱不存在其他用途。','原版／HD成對狀態及實際GUI另驗，不能由來源匹配代替。']}
    (out/'receipt.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    print(report['status'],'poses',report['normal_poses'],'frames',len(rows),'normal',len(normal),'mixed',len(rejections),'guards',len(guards))


if __name__=='__main__':main()

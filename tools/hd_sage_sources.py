#!/usr/bin/env python3
"""長老八手勢及四動畫圖格的來源；結構組合與正常畫格分列。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageDraw
from hd_interior_sources import tiles, raster


def sha(data):
    return hashlib.sha256(data).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--output', type=Path, required=True)
    args = ap.parse_args()
    p, out = args.project, args.output
    assert out.is_dir() and not any(out.iterdir())
    study = p/'workplace/full-hd-v1'
    code = (study/'sage-frames-source-v1/KENJPRO.BIN').read_bytes()
    assert sha(code) == 'ab2a46c7b6918b360a3db214935e73477b6ec7e9b04ae84f4790685cc7e86e9f'
    raw = (study/'decoded-v2/zelres2-25.decoded').read_bytes()
    assert sha(raw) == 'f9693d20d2eb14542dc68a21336e260f585cc4c97054bb1b385e07c20cfc996d'
    packed, atlas = tiles(raw)
    assert sha(packed) == '98fa4a0855443baf6e44bd52ae24e393ce6e147d2fce47d9a75713e60797eae8'
    pal = (p/'workplace/music-release-v1/release-build/data/assets/town/frame/palette.bin').read_bytes()
    assert sha(pal) == '6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c'
    base_ids = code[0x9b6:0xa16]
    base = raster(atlas, base_ids, 12, 8)
    covered = set(base_ids)
    covered.update(code[0xbfb:0xbff])
    maps, variants, by_hash = [], [], {}
    contact = Image.new('RGB', (1152, 936), '#25303a')
    draw = ImageDraw.Draw(contact)
    for pose in range(8):
        ids = code[0xa47+pose*32:0xa47+(pose+1)*32]
        covered.update(ids)
        patch = raster(atlas, ids, 8, 4)
        pixels = bytearray(base)
        for y in range(32):
            pixels[(16+y)*96+16:(16+y)*96+80] = patch[y*64:(y+1)*64]
        maps.append({'pose': pose, 'table': hex(0xaa47+pose*32), 'ids': list(ids)})
        im = Image.frombytes('P', (96,64), bytes(pixels)); im.putpalette(pal)
        im.convert('RGB').save(out/f'pose-{pose}-original.png')
        contact.paste(im.convert('RGB').resize((384,256), Image.Resampling.NEAREST), (pose%3*384,pose//3*312+28))
        draw.text((pose%3*384,pose//3*312+8), f'pose {pose}', fill='white')
        # 四個單圖格繪製端是獨立動畫。下列為結構可匹配組合，不推定全數會正常出現。
        for mouth in [None, *code[0xbfb:0xbff]]:
            frame = bytearray(pixels)
            if mouth is not None:
                for y in range(8):
                    frame[(24+y)*96+56:(24+y)*96+64] = atlas[mouth][y*8:(y+1)*8]
            key = sha(frame)
            if key in by_hash:
                by_hash[key]['aliases'].append({'pose':pose,'animation_tile':mouth})
                continue
            name = f'pose-{pose}-animation-{mouth if mouth is not None else "table"}'
            row = {'key':name,'pose':pose,'animation_tile':mouth,'indexed_sha256':key,'aliases':[], 'normal_sources':[]}
            by_hash[key] = row; variants.append(row)
            (out/(name+'.indexed')).write_bytes(frame)
    assert covered == set(range(166))
    contact.save(out/'sage-original-contact.png')
    normal = json.loads((study/'sage-frames-normal-v4/receipt.json').read_text())
    assert normal['source_save_unchanged'] and normal['scratch_save_unchanged'] and not normal['state_injection'] and not normal['snapshot_restore']
    matches, rejected = [], []
    for sample in normal['frames']:
        folder = study/'sage-frames-normal-v4'
        memory = (folder/(sample['file']+'.memory')).read_bytes()
        assert sha(memory) == sample['memory_sha256']
        assert memory[0x1b10+0xa000:0x1b10+0xbb12] == code[:0x1b12]
        assert memory[0x11b10+0x8000:0x11b10+0x8000+len(packed)] == packed
        indexed = (folder/(sample['file']+'.indexed')).read_bytes()
        assert sha(indexed) == sample['indexed_sha256']
        x,y = sample['origin']
        crop = b''.join(indexed[(y+dy)*320+x:(y+dy)*320+x+96] for dy in range(64))
        assert sha(crop) == sample['crop_sha256']
        row = by_hash.get(sha(crop))
        if row:
            row['normal_sources'].append(sample['file'])
            matches.append({'file': sample['file'],'key': row['key'],'memory_sha256': sample['memory_sha256'], 'indexed_sha256':sample['indexed_sha256']})
        else:
            rejected.append({'file':sample['file'],'crop_sha256':sha(crop),'scope':'未完成或混合畫格；不能作完整來源'})
    poses = sorted({by_hash[next(s['crop_sha256'] for s in normal['frames'] if s['file']==r['file'])]['pose'] for r in matches})
    result = {'schema':1,'status':'PASS_ALL_SAGE_TILES_AND_NORMAL_FRAME_BINDINGS','script_sha256':sha(Path(__file__).read_bytes()),
              'module_sha256':sha(code),'immutable_module_prefix_bytes':0x1b12,'immutable_module_prefix_sha256':sha(code[:0x1b12]),
              'graphics_sha256':sha(raw),'packed_sha256':sha(packed),'palette_sha256':sha(pal),'covered_tile_count':len(covered),
              'base_table':{'address':'0xa9b6','columns':12,'rows':8,'ids':list(base_ids)},'hand_tables':maps,
              'animation_tiles':list(code[0xbfb:0xbff]),'cycle_table':list(code[0xbff:0xc07]),'raise_table':list(code[0x1fd:0x200]),
              'structural_variants':variants,'normal_poses':poses,'normal_bindings':matches,'normal_rejected_frames':rejected,
              'normal_receipt_sha256':sha((study/'sage-frames-normal-v4/receipt.json').read_bytes()),
              'limits':['結構組合不代表每種均會正常出現。','正常鍵序為冷F7自建存檔、進長老、查看力量及空白續頁，非GUI或完整原版狀態配對。']}
    (out/'receipt.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    print(result['status'], 'tiles',len(covered),'variants',len(variants),'normal poses',poses,'matched',len(matches),'rejected',len(rejected))


if __name__ == '__main__':
    main()

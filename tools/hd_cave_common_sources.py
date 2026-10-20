#!/usr/bin/env python3
"""核DCHR共用洞窟圖格的正常RAM來源，原圖與候選均不當正式HD完成。"""
import argparse
import hashlib
import json
import platform
import struct
from collections import Counter
from pathlib import Path

from PIL import Image, ImageDraw

from hd_grp_sources import tile_image
from hd_planar import convert


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--output', type=Path, required=True)
    args = ap.parse_args()
    p, out = args.project, args.output
    assert out.is_dir() and not any(out.iterdir()), '須為既有空輸出目錄'
    sha = lambda b: hashlib.sha256(b).hexdigest()
    root = p / 'workplace/full-hd-v1'
    payload_path = root / 'decoded-v2/zelres3-54.decoded'
    raw = payload_path.read_bytes()
    assert len(raw) == 1872 and sha(raw) == 'ac077b64a652afebeb88b907dd61c2411ce2fbac138aec4663c09fb48c0af998'
    packed = convert(raw)
    assert sha(packed) == '7b4628728e0b6ce81a25b7bc0bbb3b4eedd640b0da102924b4384da0d38f690b'
    receipt_path = root / 'town-path-savem1-v2/receipt.json'
    receipt = json.loads(receipt_path.read_text())
    assert not receipt['state_injection'] and not receipt['snapshot_restore']
    assert receipt['source_save_name'] == 'SAVEM1'
    assert receipt['source_save_sha256'] == '03df9abcf62b1712d6b46dd1fc5b23335a4b8711b3dc84875a1498260d37735c'
    assert receipt['script_sha256'] == sha((receipt_path.parent / 'probe-source.go').read_bytes())
    assert receipt['source_save_unchanged'] and receipt['scratch_save_unchanged'] and receipt['save_mtime_unchanged']
    bindings, used, pixel_proof = [], Counter(), []
    palette = None
    for point in receipt['points']:
        assert point['completed_cave_source'] and point['source_ip'] == '01B1:7114'
        memory = (receipt_path.parent / (point['file'] + '.memory')).read_bytes()
        pal = (receipt_path.parent / (point['file'] + '.palette')).read_bytes()
        indexed = (receipt_path.parent / (point['file'] + '.indexed')).read_bytes()
        assert len(memory) == 1 << 20 and len(indexed) == 64000
        assert sha(memory) == point['memory_sha256'] and sha(pal) == point['palette_sha256']
        assert sha(indexed) == point['indexed_sha256']
        assert sha(pal) == '6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c'
        assert memory[0x11b10+0x8c00:0x11b10+0x9350] == packed
        primary = memory[0x1b10:0x11b10]
        origin = struct.unpack_from('<H', primary, 0xff31)[0]
        assert 0xe000 <= origin < 0xe900
        cells = []
        for y in range(18):
            for x in range(28):
                at = 0xe000 + (origin-0xe000+y*36+4+x) % 2304
                tile = primary[at]
                if tile & 128:
                    tile = primary[0xed20+(tile & 127)]
                if 64 <= tile <= 102:
                    cells.append({'id': tile, 'native_xy': [48+x*8, 14+y*8], 'map_address': f'01B1:{at:04X}'})
                    used[tile] += 1
        bindings.append({'file': str((receipt_path.parent / (point['file']+'.memory')).relative_to(p)),
                         'sha256': sha(memory), 'source_steps': point['steps'], 'source_ip': point['source_ip'],
                         'runtime_bytes_verified': len(packed), 'visible_cells': cells})
        if point['label'] == 'upper-landing':
            for cell in cells:
                if cell['id'] not in (64, 65, 66):
                    continue
                x, y = cell['native_xy']
                i = cell['id']-64
                im = tile_image(packed[i*48:(i+1)*48], bytes(8), pal)
                actual = bytes(c for yy in range(8) for xx in range(8)
                               for c in pal[indexed[(y+yy)*320+x+xx]*3:indexed[(y+yy)*320+x+xx]*3+3])
                assert actual == im.convert('RGB').tobytes(), '未遮擋平台原生畫面必須逐像素相同'
                pixel_proof.append({'id': cell['id'], 'native_xy': [x, y], 'pixels': 64,
                                    'rgb_sha256': sha(actual), 'normal_point': point['file']})
        palette = pal
    assert len(bindings) == 10 and all(used[i] for i in (64, 65, 66))
    assert [r['id'] for r in pixel_proof] == [64, 65, 66]
    tiles, rows = {}, []
    atlas = Image.new('RGB', (1024, 640), '#24303d')
    draw = ImageDraw.Draw(atlas)
    reference = Image.new('RGBA', (1024, 640))
    for i in range(39):
        tile_id = i+64
        data = packed[i*48:(i+1)*48]
        im = tile_image(data, bytes(8), palette)
        im.save(out / f'tile-{tile_id:03d}.png')
        tiles[tile_id] = im
        x, y = i % 8*128, i // 8*128
        zoom = im.resize((96, 96), Image.Resampling.NEAREST)
        atlas.paste(zoom, (x+16, y+24))
        draw.text((x+16, y+8), str(tile_id), fill='white')
        reference.alpha_composite(zoom, (x+16, y+16))
        rows.append({'id': tile_id, 'source_offset': i*48, 'runtime_address': f'11B1:{0x8c00+i*48:04X}',
                     'source_sha256': sha(raw[i*48:(i+1)*48]), 'packed_sha256': sha(data),
                     'rgba_sha256': sha(im.tobytes()), 'observed_cells': used[tile_id]})
    atlas.save(out / 'DCHR-original-overview.png')
    reference.save(out / 'DCHR-atlas-reference.png')
    for name, ids in [('green-platform', [64, 65, 66]), ('second-triplet', [67, 68, 69]), ('third-triplet', [70, 71, 72])]:
        im = Image.new('RGBA', (24, 8))
        for i, tile_id in enumerate(ids):
            im.alpha_composite(tiles[tile_id], (i*8, 0))
        im.save(out / (name+'-original.png'))
        if name == 'green-platform':
            cutout = im.copy()
            for y in range(8):
                for x in range(24):
                    if cutout.getpixel((x, y))[:3] == (0, 0, 0):
                        cutout.putpixel((x, y), (0, 0, 0, 0))
            ref = Image.new('RGBA', (1024, 1024))
            ref.alpha_composite(cutout.resize((768, 256), Image.Resampling.NEAREST), (128, 384))
            ref.save(out / 'green-platform-reference.png')
    report = {'schema': 1, 'status': 'PASS_DCHR_NORMAL_RAM_BINDING_SOURCE_ONLY', 'source_record': 'zelres3.record.54',
              'source_file': str(payload_path.relative_to(p)), 'source_sha256': sha(raw), 'source_bytes': len(raw),
              'packed_sha256': sha(packed), 'first_id': 64, 'last_id': 102, 'tiles': rows,
              'normal_receipt': str(receipt_path.relative_to(p)), 'normal_receipt_sha256': sha(receipt_path.read_bytes()),
              'normal_bindings': bindings, 'normal_observed_ids': sorted(used),
              'platform_pixel_proof': pixel_proof,
              'script_sha256': sha(Path(__file__).read_bytes()), 'tools': {'python': platform.python_version(), 'pillow': Image.__version__},
              'address_space': 'SAR3 record54解壓buffer +0；現有已核三平面轉換後，dosgolem資源段11B1:8C00..934F；地圖ID64..102，沿011的8030+(id-1)*48',
              'evidence_level': '已證實：39格1872bytes與十個正常完成來源一致；平台64..66的192個原生RGB像素相同。既有入口沿020；其他新用途與正常可達性未知。',
              'production_runtime_changed': False,
              'limits': ['來源相同不代表39格皆已在正常畫面出現；其他用途須按consumer及正常來源核對。',
                         '原版持有ID64..66時，既有只接受MPP圖集與完整入口矩陣的地形覆繪會拒絕；未放寬守衛。',
                         '本工具只解出原圖，不是HD接入、正常Satono到達或實際GUI驗收。']}
    (out / 'receipt.json').write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')
    print(report['status'], len(rows), 'tiles', len(bindings), 'normal sources', sorted(used))


if __name__ == '__main__':
    main()

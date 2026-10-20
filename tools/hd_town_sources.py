#!/usr/bin/env python3
"""核十份城鎮地圖與三圖塊家族，區分已驗正常來源與靜態候選。"""
import argparse
import hashlib
import json
import struct
from collections import Counter
from pathlib import Path

from PIL import Image, ImageDraw

from hd_grp_sources import terrain_preprocess, tile_image
from hd_inventory import EXPECTED, records


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--output', type=Path, required=True)
    args = ap.parse_args()
    project, out = args.project, args.output
    assert out.is_dir() and not any(out.iterdir())
    sha = lambda data: hashlib.sha256(data).hexdigest()
    root = project / 'workplace/full-hd-v1'
    archive = (project / 'workplace/orig/ZELRES2.SAR').read_bytes()
    assert sha(archive) == EXPECTED['ZELRES2.SAR'][1]
    entries = {row['index']: row for row in records(archive)}
    code = (root / 'castle-edge-ida-v1/primary.segment').read_bytes()
    assert sha(code) == 'aed45060ee02eafa07b41891d326ac0217a8afdd2957d79b401369a8bb2fc11d'
    palette = (project / 'workplace/music-release-v1/release-build/data/assets/town/frame/palette.bin').read_bytes()
    assert sha(palette) == '6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c'
    families = []
    for key, record, count in (('CPAT', 33, 157), ('MPAT', 34, 242), ('DPAT', 35, 191)):
        source = (root / f'decoded-v2/zelres2-{record:02d}.decoded').read_bytes()
        assert len(source) == 256 + count*48
        assert struct.unpack_from('<H', source)[0] == 6
        tiles, packed, masks = [], bytearray(), bytearray()
        for tile in range(count):
            flag = source[6+tile]
            assert flag <= 4
            pixel, mask = terrain_preprocess(source[256+tile*48:256+(tile+1)*48], flag)
            packed.extend(pixel)
            masks.extend(mask)
            tiles.append(tile_image(pixel, mask, palette))
        families.append({'key': key, 'record': record, 'count': count, 'tiles': tiles, 'flags': source[6:6+count],
                         'decoded_sha256': sha(source), 'packed_sha256': sha(packed), 'mask_candidate_sha256': sha(masks)})
    report = []
    for record, key in enumerate(('CMAP', 'MRMP', 'STMP', 'BSMP', 'HLMP', 'TMMP', 'DRMP', 'LLMP', 'PRMP', 'ESMP'), 36):
        row = entries[record]
        raw = archive[row['payload_offset']:row['payload_offset']+row['bytes']]
        assert sha(raw) == row['sha256']
        descriptor_address = 0x10c8 + (record-36)*11
        descriptor = code[descriptor_address:descriptor_address+11]
        assert descriptor == bytes((1, record+1)) + (key+'.MDT').encode('ascii') + b'\0'
        word = lambda at: struct.unpack_from('<H', raw, at)[0]
        width, map_end = word(2), 23+word(2)*8
        assert 1 <= width <= 1024 and word(17)-0xc000 == map_end <= len(raw)
        header = word(0)-0xc000
        end = raw.index(255, header)
        background, terrain = raw[end+1:end+3]
        assert background in (0, 1) and terrain < len(families)
        family = families[terrain]
        title = word(4)-0xc000
        title_bytes = raw[title+4:title+4+raw[title+3]]
        title_text = title_bytes.decode('ascii')
        actor_start, actors = word(15)-0xc000, []
        pos = actor_start
        while word(pos) != 65535:
            assert map_end <= pos and pos+8 <= len(raw)
            values = raw[pos:pos+8]
            x = struct.unpack_from('<H', values)[0]
            assert x < width and values[2] & 127 < 5
            actors.append({'source_address': hex(0xc000+pos), 'column': x, 'raw': list(values)})
            pos += 8
        body = raw[23:map_end]
        unresolved = Counter(value for value in body if value >= family['count'])
        picture = Image.new('RGBA', (width*8, 64))
        for x in range(width):
            for y in range(8):
                tile = body[x*8+y]
                if tile >= family['count']:
                    ImageDraw.Draw(picture).rectangle((x*8, y*8, x*8+7, y*8+7), fill='#a52b8aff')
                    continue
                image = family['tiles'][tile].copy()
                if y >= 3 or not family['flags'][tile]:
                    image.putalpha(255)
                picture.paste(image, (x*8, y*8))
        picture.save(out / (key+'-foreground-static.png'))
        # 保留透明原圖；棋盤背景只供目視研究，不能冒稱原版截圖。
        visible = Image.new('RGB', picture.size, '#22303a')
        visible.paste(picture, mask=picture.getchannel('A'))
        visible.resize((width*8*3, 192), Image.Resampling.NEAREST).save(out / (key+'-reference.png'))
        (out / (key+'.source')).write_bytes(raw)
        (out / (key+'.columns')).write_bytes(body)
        normalized = bytes(min(255, x) for x in body)
        report.append({'key': key, 'record': record, 'payload_offset': row['payload_offset'], 'payload_sha256': sha(raw),
                       'map_descriptor_address': hex(descriptor_address), 'map_descriptor': descriptor.hex(),
                       'map_load_AH': hex(0x80+record-36), 'town_id': raw[6], 'name_source_address': hex(word(4)+4),
                       'name_ascii': title_text, 'width': width, 'height': 8, 'body_offset': 23, 'map_ids_sha256': sha(normalized),
                       'terrain_key': family['key'], 'terrain_selector': terrain, 'background_selector': background,
                       'background_program': ('YMPD.BIN', 'CKPD.BIN')[background],
                       'selector_address': hex(0xc000+end+1), 'actors': actors, 'actor_table_address': hex(0xc000+actor_start),
                       'unresolved_tile_ids': dict(unresolved), 'world_rgba_sha256': sha(picture.tobytes()),
                       'level': '已核靜態地圖／選擇端；新增城鎮正常載入、原生像素與HD另驗'})
    family_rows = [{k: value for k, value in row.items() if k not in ('tiles', 'flags')} for row in families]
    receipt = {'schema': 1, 'status': 'PASS_TEN_TOWN_STRUCTURES_SELECTORS_WITH_EXPLICIT_UNRESOLVED',
               'script_sha256': sha(Path(__file__).read_bytes()), 'archive_sha256': sha(archive), 'code_sha256': sha(code),
               'palette_sha256': sha(palette), 'families': family_rows, 'maps': report, 'total_cells': sum(row['width']*8 for row in report),
               'limits': ['地圖前景只含八列，未合成原版背景、下層地面、居民與主角，不稱完整遊戲畫面。',
                          '洋紅格表示未核圖格，不能以相鄰圖格或空白補值；DPAT尾端及動態記錄須正常來源再核。',
                          '此為來源盤點，不計全城鎮HD完成或正常可達；原圖與完整圖面不入Git。']}
    (out / 'receipt.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')
    print(receipt['status'], receipt['total_cells'], [(row['key'], row['terrain_key'], row['unresolved_tile_ids']) for row in report])


if __name__ == '__main__':
    main()

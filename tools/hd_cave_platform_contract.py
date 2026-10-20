#!/usr/bin/env python3
"""核綠色平台的原版消費端、31地圖表及無透明缺口圖面，供043 READY審查。"""
import argparse
import hashlib
import json
import struct
from pathlib import Path

import numpy as np
from PIL import Image

from hd_grp_sources import tile_image
from hd_planar import convert


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--output', type=Path, required=True)
    args = ap.parse_args()
    p, out = args.project, args.output
    assert out.is_dir() and not any(out.iterdir())
    root = p/'workplace/full-hd-v1'
    sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
    read = lambda path: json.loads(path.read_text())
    primary_path = root/'cave-common-ida-v1/primary.segment'
    primary = primary_path.read_bytes()
    assert sha(primary_path) == '5e7f76080cc542eb06ca90ec349ad2ec0c5ab374c32acf59bc4c4307899a0a5b'
    code_hash = hashlib.sha256(primary[0x7fdc:0x8163]).hexdigest()
    assert code_hash == '494c0a9b65acd0d48ef6ede08f336d06b9ebdab9e6c418865d5f4c392437ebff'
    ida_path = root/'cave-common-ida-v3/evidence.json'
    ida = read(ida_path)
    assert ida['tool'] == 'IDA 9.4' and ida['input_sha256'] == sha(primary_path)
    assert ida['script_sha256'] == sha(p/'tools/ida_cave_common.py')
    assert ida['source_index_sha256'] == sha(p/'docs/cave-common-source-index.json')
    assert len(ida['functions']) == 7 and len(ida['tables']) == 3
    instructions = {int(row['address'], 16): row for function in ida['functions'] for row in function['instructions']}
    for address, expected in ((0x7ec3, 'bf008c'), (0x8137, '83c703'),
                              (0x806b, 'fe4502'), (0x80f3, 'fe4d02'), (0x8357, '8815')):
        assert instructions[address]['bytes'] == expected, hex(address)
    descriptor = next(row for row in ida['tables'] if row['address'] == '0x9c13')
    assert descriptor['bytes'] == '0237444348522e47525000'
    maps_path = root/'cave-map-sources-v1/receipt.json'
    maps = read(maps_path)
    tables = []
    for record in maps['records']:
        source_path = maps_path.parent/record['source']
        data = source_path.read_bytes()
        assert sha(source_path) == record['payload_sha256']
        width, start, end = struct.unpack_from('<HHH', data, 2)
        assert width == record['width'] and 0xc01b <= start < end <= 0xc000+len(data)
        table = data[start-0xc000:end-0xc000]
        assert len(table) >= 2 and (len(table)-2) % 3 == 0 and table[-2:] == b'\xff\xff'
        positions = [list(struct.unpack_from('<HB', table, at)) for at in range(0, len(table)-2, 3)]
        assert all(x < width and y < 64 for x, y in positions)
        tables.append({'map': record['name'], 'source_sha256': sha(source_path),
                       'table_address': f'01B1:{start:04X}', 'end_address': f'01B1:{end:04X}',
                       'entry_stride': 3, 'positions': positions, 'table_sha256': hashlib.sha256(table).hexdigest()})
    assert len(tables) == 31 and sum(len(row['positions']) for row in tables) == 72
    cycle_path = root/'town-path-savem1-v4/receipt.json'
    cycle = read(cycle_path)
    assert not cycle['state_injection'] and not cycle['snapshot_restore']
    assert cycle['source_save_unchanged'] and cycle['scratch_save_unchanged'] and cycle['save_mtime_unchanged']
    observed = []
    for point in cycle['points']:
        source_path = cycle_path.parent/(point['file']+'.memory')
        assert sha(source_path) == point['memory_sha256'] and point['completed_cave_source']
        memory = source_path.read_bytes()[0x1b10:0x11b10]
        start, end = struct.unpack_from('<HH', memory, 0xc004)
        table = memory[start:end]
        positions = [list(struct.unpack_from('<HB', table, at)) for at in range(0, len(table)-2, 3)]
        assert [row[0] for row in positions] == [48, 221] and positions[1] == [221, 44]
        observed.append(positions[0])
    assert observed == [[48, 17], [48, 24], [48, 17]]
    pair_path = root/'cave-platform-source-pair-v2/cave-platform/receipt.json'
    pair = read(pair_path)
    assert pair['state_equal'] and pair['save_unchanged'] and not pair['wrote'] and not pair['unimplemented']
    assert pair['source_save_sha256'] == '03df9abcf62b1712d6b46dd1fc5b23335a4b8711b3dc84875a1498260d37735c'
    assert pair['commands_sha256'] == sha(p/'tools/cave-platform.commands')
    normal_platform_points = [point for point in pair['points'] if point['name'].startswith('platform-')]
    assert len(pair['points']) == 10 and len(normal_platform_points) == 5
    assert all(point['state_equal'] and point.get('source_steps') and point['presentation']['used'] for point in normal_platform_points)
    generation_path = root/'cave-common-art-v2/generation-v2.json'
    generation = read(generation_path)
    assert generation['mode'] == 'built-in imagegen' and generation['alpha_range'] == [255, 255]
    assert sha(p/generation['output_file']) == generation['output_sha256']
    registration_path = root/'cave-common-registration-v2/manifest.json'
    registration = read(registration_path)
    assert registration['script_sha256'] == sha(p/'tools/hd_cave_common_art.py')
    assert registration['transparent_inside_mask'] == registration['outside_mask_written_pixels'] == 0
    art_path = registration_path.parent/registration['art_file']
    assert sha(art_path) == registration['art_sha256']
    art = np.asarray(Image.open(art_path).convert('RGBA'))
    common_path = root/'cave-common-sources-v2/receipt.json'
    common = read(common_path)
    raw = (p/common['source_file']).read_bytes()
    assert hashlib.sha256(raw).hexdigest() == common['source_sha256']
    packed = convert(raw)
    palette = (cycle_path.parent/(cycle['points'][0]['file']+'.palette')).read_bytes()
    native = Image.new('RGBA', (24, 8))
    for i in range(3):
        native.paste(tile_image(packed[i*48:(i+1)*48], bytes(8), palette), (i*8, 0))
    occupied = np.repeat(np.repeat(np.any(np.asarray(native)[:, :, :3] != 0, axis=2), 3, axis=0), 3, axis=1)
    assert art.shape == (24, 72, 4) and np.all(art[:, :, 3][occupied] == 255) and not np.any(art[~occupied])
    proofs = [primary_path, ida_path, maps_path, cycle_path, pair_path, common_path, generation_path, registration_path, art_path]
    report = {'schema': 1, 'status': 'READY_GREEN_PLATFORM_SOURCE_AND_ART', 'scope_spec': 'docs/spec/043-cave-common-scenery.md',
              'script_sha256': sha(Path(__file__)), 'platform_ids': [64, 65, 66], 'common_resource_bytes': 1872,
              'platform_code_address': '01B1:7FDC..8162', 'platform_code_sha256': code_hash,
              'tables': tables, 'map_count': 31, 'platform_entries': 72,
              'normal_observed_positions': observed, 'opaque_hd_pixels': int(np.count_nonzero(occupied)),
              'normal_readonly_pair_points': 10, 'presentable_normal_platform_points': 5,
              'transparent_inside_mask': 0, 'outside_mask_written_pixels': 0,
              'proofs': [{'file': str(path.relative_to(p)), 'sha256': sha(path)} for path in proofs],
              'limits': ['READY限綠色平台64～66；其他共用圖格不因來源相同而放行。',
                         '72是靜態表項數，非72次正常到達；正常平台只抽樣MP10。',
                         '尚未接正式呈現器、原版完整狀態配對或GUI，不當全遊戲HD完成。']}
    (out/'receipt.json').write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')
    print(report['status'], report['platform_entries'], 'source entries', report['opaque_hd_pixels'], 'opaque pixels')


if __name__ == '__main__':
    main()

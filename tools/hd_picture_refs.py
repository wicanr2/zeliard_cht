#!/usr/bin/env python3
"""片頭／結局插畫的正確色序參考圖。只讀實機截圖與既有圖集排版，不啟動遊戲、不生圖。規格068。

050的參考圖以IBM位元順序（4紅、2綠）合成色盤，原版是PC-98順序（2紅、4綠），HD圖面因此紅綠互換。
本工具從實機截圖裁出插畫的原生點陣與當時色盤，依picture-sheets.json的格位重組1536×1024參考圖集。
點陣須與pictures.json逐點相同；--raster-from-dump只用於pictures.json解碼錯誤的插畫，改以實機點陣為準。
"""
import argparse
import base64
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image


def digest(data):
    return hashlib.sha256(data).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--output', type=Path, required=True, help='既有空目錄')
    ap.add_argument('--sheet', action='append', default=[], help='picture-sheets.json的圖集鍵，如story1')
    ap.add_argument('--picture', action='append', default=[], metavar='KEY=DUMP@X,Y')
    ap.add_argument('--raster-from-dump', action='append', default=[], metavar='KEY=WIDTHxHEIGHT', help='以實機點陣為準，並指定尺寸')
    ap.add_argument('--ignore-row0', action='append', default=[], metavar='KEY', help='第0列登錄點陣與實機不同的插畫（end7、fin），只核第1列以後')
    ap.add_argument('--formula', action='append', default=[], metavar='KEY', help='沒有實機截圖的插畫：登錄點陣配修正後的PC-98公式色盤，收據標為未經實機核對')
    ap.add_argument('--blank', action='append', default=[], metavar='KEY', help='不重畫的插畫：圖集該格留黑，登錄時以--keep保留原圖面')
    ap.add_argument('--single', action='append', default=[], metavar='KEY', help='不在任何圖集的插畫（new1），或改以實機點陣重新登錄的插畫（dmaou）：自成一張1536×1024參考圖，格位寫入refs.json，登錄時以--extra-cells讀取')
    args = ap.parse_args()
    p, out = args.project, args.output
    if not out.is_dir() or any(out.iterdir()):
        raise ValueError('輸出須為既有空目錄')
    w = p/'workplace/full-hd-v1/manual-hd-v1'
    pictures = {r['key']: r for r in json.loads((w/'pictures.json').read_text())['scenes']}
    sheets = {s['key']: s for s in json.loads((w/'picture-sheets.json').read_text())['sheets']}
    # 完整插畫（hd_complete_sources.py）另有圖集與重新解碼的點陣；同鍵時以完整插畫為準。
    completion = json.loads((w/'completion-sources.json').read_text())
    sheets.update({s['key']: s for s in completion['sheets']})
    pictures.update({s['key']: s for s in completion['sources'] if s['kind'] == 'picture'})
    ignore_row0 = set(args.ignore_row0)
    dumps = {}
    for spec in args.picture:
        key, rest = spec.split('=')
        path, at = rest.split('@')
        dumps[key] = (path, tuple(map(int, at.split(','))))
    from_dump = {k: tuple(map(int, v.split('x'))) for k, v in (s.split('=') for s in args.raster_from_dump)}
    receipt = {'schema': 1, 'tool': 'tools/hd_picture_refs.py', 'tool_sha256': digest((p/'tools/hd_picture_refs.py').read_bytes()), 'pictures': [], 'sheets': [], 'game_launched': False}
    rendered = {}
    # 兩個三位元子像素的平均色，位元順序與PC-98相同（1藍、2紅、4綠）；結局與片頭四個場景與實機色盤逐色相同。
    formula = np.array([[(((i>>4)&2)+(i&2))*255//4, (((i>>4)&4)+(i&4))*255//8, (((i>>4)&1)+(i&1))*255//2] for i in range(256)], np.uint8)
    for key in args.formula:
        row = pictures[key]
        width, height = row['width'], row['height']
        native = np.frombuffer(base64.b64decode(row['raster']), np.uint8).reshape(height, width)
        image = Image.fromarray(formula[native]).resize((width*3, height*3), Image.Resampling.NEAREST)
        file = out/f'{key}-picture-v2.png'
        image.save(file)
        rendered[key] = image
        receipt['pictures'].append({'key': key, 'width': width, 'height': height, 'raster_source': 'registered raster, unverified at runtime', 'raster': base64.b64encode(native.tobytes()).decode(),
                                    'raster_sha256': digest(native.tobytes()), 'palette': base64.b64encode(formula.tobytes()).decode(), 'palette_source': 'corrected PC-98 formula, unverified at runtime',
                                    'file': file.name, 'sha256': digest(file.read_bytes())})
    for key, (path, (ox, oy)) in dumps.items():
        data = (p/path).read_bytes()
        if len(data) != 64768:
            raise ValueError(f'截圖格式不符：{path}')
        screen = np.frombuffer(data[:64000], np.uint8).reshape(200, 320)
        palette = np.frombuffer(data[64000:], np.uint8).reshape(256, 3)
        width, height = from_dump.get(key, (pictures[key]['width'], pictures[key]['height']))
        if ox < 0 or oy < 0 or ox+width > 320 or oy+height > 200:
            raise ValueError(f'{key} 超出畫面')
        native = screen[oy:oy+height, ox:ox+width]
        source = 'pictures.json'
        if key in from_dump:
            source = 'runtime dump'
        else:
            expected = np.frombuffer(base64.b64decode(pictures[key]['raster']), np.uint8).reshape(height, width)
            first = 1 if key in ignore_row0 else 0
            if not np.array_equal(native[first:], expected[first:]):
                raise ValueError(f'{key} 原生點陣與登錄來源不符；確認解碼錯誤後改用--raster-from-dump')
            if first:
                source = 'registered raster, row 0 ignored'
        image = Image.fromarray(palette[native]).resize((width*3, height*3), Image.Resampling.NEAREST)
        file = out/f'{key}-picture-v2.png'
        image.save(file)
        rendered[key] = image
        receipt['pictures'].append({'key': key, 'width': width, 'height': height, 'raster_source': source, 'raster': base64.b64encode(native.tobytes()).decode(),
                                    'raster_sha256': digest(native.tobytes()), 'palette': base64.b64encode(palette.tobytes()).decode(),
                                    'runtime_dump': path, 'runtime_dump_sha256': digest(data), 'origin': [ox, oy], 'file': file.name, 'sha256': digest(file.read_bytes())})
    for name in args.sheet:
        sheet = sheets[name]
        atlas = Image.new('RGB', (1536, 1024), 'black')
        for n, cell in enumerate(sheet['cells']):
            key = cell['key']
            if key in args.blank:
                continue
            if key not in rendered:
                raise ValueError(f'{name} 的 {key} 沒有實機截圖')
            im = rendered[key]
            # 與hd_manual_tail.py相同的格位公式；格位須與既有圖集逐一相同，登錄裁切才不必改。
            ratio = min(720/im.width, 460/im.height)
            size = tuple(round(v*ratio) for v in im.size)
            x, y = n % 2*768+(768-size[0])//2, n//2*512+(512-size[1])//2
            if [x, y, x+size[0], y+size[1]] != cell['box']:
                raise ValueError(f'{name} 的 {key} 格位與既有圖集不同')
            atlas.paste(im.resize(size, Image.Resampling.NEAREST), (x, y))
        file = out/f'{name}-v2-selected.png'
        atlas.save(file)
        receipt['sheets'].append({'key': name, 'file': file.name, 'sha256': digest(file.read_bytes()), 'cells': sheet['cells']})
    for key in args.single:
        if key not in rendered:
            raise ValueError(f'{key} 沒有實機截圖')
        # 已在圖集內的插畫只有改以實機點陣重新登錄時（dmaou：登錄點陣的色號解碼有誤）才另成一張。
        if key not in from_dump and any(c['key'] == key for s in sheets.values() for c in s['cells']):
            raise ValueError(f'{key} 已在圖集內')
        im = rendered[key]
        # 置中，四周至少留80點黑邊。
        ratio = min((1536-160)/im.width, (1024-160)/im.height)
        size = tuple(round(v*ratio) for v in im.size)
        x, y = (1536-size[0])//2, (1024-size[1])//2
        atlas = Image.new('RGB', (1536, 1024), 'black')
        atlas.paste(im.resize(size, Image.Resampling.NEAREST), (x, y))
        file = out/f'{key}-v2-selected.png'
        atlas.save(file)
        receipt['sheets'].append({'key': key, 'file': file.name, 'sha256': digest(file.read_bytes()), 'cells': [{'key': key, 'box': [x, y, x+size[0], y+size[1]]}], 'single': True})
    (out/'refs.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+'\n')
    print(f'REFS {len(rendered)} pictures, {len(args.sheet)} sheets; no game launched')


if __name__ == '__main__':
    main()

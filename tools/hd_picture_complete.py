#!/usr/bin/env python3
"""片頭／結局插畫補登錄為整張全圖補繪。只裁切既有圖集，不重畫、不啟動遊戲。規格068。

050人工圖格只收三色以上的8×8格，兩色抖色格（例如陽台柱底的119／112棋盤）沒有HD，原版露出。
同一張story圖集的同一次裁切改登錄為complete-art的整張插畫與8×8片段（兩色格也收），片段仍受母圖上下文守衛。
每張插畫的原生點陣須與實機截圖逐點相同才登錄；end7／fin另遮去實機永遠是黑色的第0列。
hd_complete_register.py會從頭重寫complete-art.json；重跑它之後必須以相同參數再跑本工具，否則story-群組與第0列遮罩會消失。
"""
import argparse
import base64
import hashlib
import json
import shutil
from collections import deque
from pathlib import Path

import numpy as np
from PIL import Image

PREFIX = 'story-'


def digest(data):
    return hashlib.sha256(data).hexdigest()


def border_mask(native, index):
    """與圖框邊緣連通的指定色號設為透明（0），其餘255；人物內部同色不受影響。"""
    height, width = native.shape
    bg = np.zeros(native.shape, bool)
    queue = deque((y, x) for y in range(height) for x in range(width) if (x in (0, width-1) or y in (0, height-1)) and native[y, x] == index)
    for y, x in queue:
        bg[y, x] = True
    while queue:
        y, x = queue.popleft()
        for ny, nx in ((y+1, x), (y-1, x), (y, x+1), (y, x-1)):
            if 0 <= ny < height and 0 <= nx < width and not bg[ny, nx] and native[ny, nx] == index:
                bg[ny, nx] = True
                queue.append((ny, nx))
    return np.where(bg, 0, 255).astype(np.uint8)


def dump(path):
    data = path.read_bytes()
    if len(data) != 64768:
        raise ValueError(f'截圖格式不符：{path}')
    return np.frombuffer(data[:64000], np.uint8).reshape(200, 320), data[64000:]


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--backup', type=Path, required=True, help='原complete-art.json備份位置，須不存在')
    ap.add_argument('--picture', action='append', default=[], metavar='KEY=DUMP@X,Y', help='插畫鍵、實機.indexed截圖及原生左上角')
    ap.add_argument('--row0', action='append', default=[], metavar='KEY=DUMP', help='第0列遮去後須與實機截圖逐點相同的既有插畫')
    ap.add_argument('--transparent', action='append', default=[], metavar='KEY=INDEX', help='插畫外圍背景色號；與邊緣連通者透明，露出原版底色')
    args = ap.parse_args()
    p = args.project
    w = p/'workplace/full-hd-v1/manual-hd-v1'
    out = w/'registered'
    manifest_file = out/'complete-art.json'
    if args.backup.exists():
        raise ValueError('備份已存在，拒絕覆寫')
    shutil.copy2(manifest_file, args.backup)
    m = json.loads(manifest_file.read_text())
    # 重跑時移除上一次補登錄的群組；補登錄群組必須連續排在最後，前面的群組索引不變。
    keep = len(m['groups'])
    for i, g in enumerate(m['groups']):
        if g['key'].startswith(PREFIX):
            keep = min(keep, i)
    if any(not g['key'].startswith(PREFIX) for g in m['groups'][keep:]):
        raise ValueError('補登錄群組不在清單最後')
    m['groups'] = m['groups'][:keep]
    m['entries'] = [e for e in m['entries'] if e['group'] < keep]
    # 本工具產生的圖檔都以complete-story-開頭；重跑前移除，不留未引用的舊片段。
    for stale in out.glob('complete-'+PREFIX+'*.png'):
        stale.unlink()
    pictures = {r['key']: r for r in json.loads((w/'pictures.json').read_text())['scenes']}
    cells = {c['key']: (s['key'], c['box']) for s in json.loads((w/'picture-sheets.json').read_text())['sheets'] for c in s['cells']}
    tiles = json.loads((out/'manual-tiles.json').read_text())['tiles']
    transparent = {k: int(v) for k, v in (s.split('=') for s in args.transparent)}
    supplement = []
    for spec in args.picture:
        key, rest = spec.split('=')
        path, at = rest.split('@')
        ox, oy = map(int, at.split(','))
        row = pictures[key]
        width, height = row['width'], row['height']
        raster = base64.b64decode(row['raster'])
        native = np.frombuffer(raster, np.uint8).reshape(height, width)
        screen, palette = dump(p/path)
        if ox < 0 or oy < 0 or ox+width > 320 or oy+height > 200 or not np.array_equal(screen[oy:oy+height, ox:ox+width], native):
            raise ValueError(f'{key} 原生點陣與實機截圖不符')
        sheet, box = cells[key]
        source = w/'generated'/(sheet+'-generated-v1.png')
        atlas = Image.open(source).convert('RGBA')
        box = tuple(round(v*(atlas.width/1536 if i % 2 == 0 else atlas.height/1024)) for i, v in enumerate(box))
        # 與hd_manual_art.py的人工圖格同一次裁切；逐格核對已登錄圖格，證明圖面未改。
        painted = atlas.crop(box).resize((width*3, height*3), Image.Resampling.LANCZOS)
        checked = 0
        files = {t['file']: t for t in tiles if t['category'] == key}
        for y in range(0, height-7, 8):
            for x in range(0, width-7, 8):
                name = 'tile-'+digest(native[y:y+8, x:x+8].tobytes())[:16]+'.png'
                if name in files:
                    tile = np.array(Image.open(out/name).convert('RGBA'))
                    crop = np.array(painted.crop((x*3, y*3, x*3+24, y*3+24)))
                    if not np.array_equal(tile, crop):
                        continue  # 同點陣先出現在別格時沿用先登錄者，只核首見位置
                    checked += 1
        if checked == 0:
            raise ValueError(f'{key} 找不到相同裁切的人工圖格')
        group = len(m['groups'])
        name = PREFIX+key
        m['groups'].append({'key': name, 'bindings': [], 'source_sha256': row['source_sha256'], 'archive': 1, 'record': row['record'],
                            'method': f'050 {sheet} 圖集同一次裁切；原生點陣與實機截圖逐點相同', 'runtime_dump': path, 'runtime_dump_sha256': digest((p/path).read_bytes()), 'runtime_origin': [ox, oy]})
        colors = base64.b64encode(palette).decode()
        parent = f'complete-{name}-000.png'
        painted.save(out/parent)
        base = {'group': group, 'source': name, 'colors': colors}
        large = {**base, 'width': width, 'height': height, 'raster': base64.b64encode(raster).decode(), 'file': parent, 'sha256': digest((out/parent).read_bytes())}
        # 插畫比原版底色範圍小時（國王擁抱圖），外圍背景透明，底色維持原版，不讓HD背景成為一塊矩形。
        mask = border_mask(native, transparent[key]) if key in transparent else None
        if mask is not None:
            large['mask'] = base64.b64encode(mask.tobytes()).decode()
        m['entries'].append(large)
        seen, count = set(), 0
        for y in range(0, height, 8):
            for x in range(0, width, 8):
                fw, fh = min(8, width-x), min(8, height-y)
                if fw < 2 or fh < 2:
                    continue
                part = native[y:y+fh, x:x+fw].tobytes()
                cut = mask[y:y+fh, x:x+fw].tobytes() if mask is not None else b''
                shown = (native[y:y+fh, x:x+fw][mask[y:y+fh, x:x+fw] != 0] if mask is not None else native[y:y+fh, x:x+fw]).ravel()
                if len(set(shown.tolist())) < 2 or (fw, fh, part, cut) in seen:
                    continue
                seen.add((fw, fh, part, cut))
                file = f'complete-{name}-000-x{x:03d}-y{y:03d}.png'
                painted.crop((x*3, y*3, (x+fw)*3, (y+fh)*3)).save(out/file)
                fragment = {**base, 'width': fw, 'height': fh, 'raster': base64.b64encode(part).decode(), 'file': file, 'sha256': digest((out/file).read_bytes()), 'derivative_of': parent}
                if mask is not None:
                    fragment['mask'] = base64.b64encode(cut).decode()
                m['entries'].append(fragment)
                count += 1
        supplement.append({'key': name, 'source_art': str(source.relative_to(p)), 'source_art_sha256': digest(source.read_bytes()), 'tiles_checked': checked, 'fragments': count, 'runtime_dump': path, 'origin': [ox, oy], 'transparent_index': transparent.get(key), 'transparent_pixels': int((mask == 0).sum()) if mask is not None else 0})
    for spec in args.row0:
        key, path = spec.split('=')
        screen, _ = dump(p/path)
        large = [e for e in m['entries'] if e['source'] == key and not e.get('derivative_of')]
        if len(large) != 1:
            raise ValueError(f'{key} 整張插畫不唯一')
        e = large[0]
        width, height = e['width'], e['height']
        native = np.frombuffer(base64.b64decode(e['raster']), np.uint8).reshape(height, width)
        if width != 320 or not (screen[0] == 0).all() or not np.array_equal(screen[1:height], native[1:]):
            raise ValueError(f'{key} 第0列以外與實機截圖不符')
        mask = bytes([0]*width+[255]*(width*(height-1)))
        e['mask'] = base64.b64encode(mask).decode()
        rows = 0
        for f in m['entries']:
            if f.get('derivative_of') == e['file'] and f['file'].endswith('-y000.png'):
                f['mask'] = base64.b64encode(bytes([0]*f['width']+[255]*(f['width']*(f['height']-1)))).decode()
                rows += 1
        supplement.append({'key': key, 'row0_masked': True, 'fragments_masked': rows, 'runtime_dump': path, 'runtime_dump_sha256': digest((p/path).read_bytes())})
    if len(m['entries']) > 16384:
        raise ValueError('全圖補繪項目超過載入上限')
    tool = p/'tools/hd_picture_complete.py'
    m['supplements'] = [s for s in m.get('supplements', []) if s.get('tool') != 'tools/hd_picture_complete.py']
    m['supplements'].append({'tool': 'tools/hd_picture_complete.py', 'tool_sha256': digest(tool.read_bytes()), 'backup_sha256': digest(args.backup.read_bytes()), 'items': supplement, 'game_launched': False})
    manifest_file.write_text(json.dumps(m, ensure_ascii=False, indent=2)+'\n')
    print(f'SUPPLEMENTED {len(args.picture)} pictures, {len(args.row0)} row-0 masks, {len(m["entries"])} entries; no game launched')


if __name__ == '__main__':
    main()

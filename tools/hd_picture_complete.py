#!/usr/bin/env python3
"""片頭／結局插畫補登錄為整張全圖補繪。只裁切既有圖集，不重畫、不啟動遊戲。規格068。

050人工圖格只收三色以上的8×8格，兩色抖色格（例如陽台柱底的119／112棋盤）沒有HD，原版露出。
同一張story圖集的同一次裁切改登錄為complete-art的整張插畫與8×8片段（兩色格也收），片段仍受母圖上下文守衛。
每張插畫的原生點陣須與實機截圖逐點相同才登錄；end7／fin另遮去實機永遠是黑色的第0列。
hd_complete_register.py會從頭重寫complete-art.json；重跑它之後必須以相同參數再跑本工具，否則story-群組與第0列遮罩會消失。
重畫後的圖集以--sheet-art指定，同一插畫須同時--retile，把首見位置的8×8人工圖格換成新圖面；hd_manual_art.py重跑會從v1圖集重生人工圖格，之後也須再跑本工具。
人像說話姿態與魔王眼睛循環以--patch-art指定hd_portrait_patch.py參考圖的重畫結果，部件只在顏色與底圖不同處換圖，其餘沿用本次登錄的底圖HD。
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
    ap.add_argument('--sheet-art', action='append', default=[], metavar='SHEET=PNG', help='該圖集改用重畫後的圖面（須先通過hd_hue_check.py）')
    ap.add_argument('--retile', action='append', default=[], metavar='KEY', help='同步替換該插畫首見位置的8×8人工圖格')
    ap.add_argument('--replace-art', action='append', default=[], metavar='SHEET=PNG', help='completion-sources圖集改用重畫後的圖面；只重寫該圖集插畫的整張與片段PNG（須先通過hd_hue_check.py）')
    ap.add_argument('--raster-from-dump', action='append', default=[], metavar='KEY=WIDTHxHEIGHT', help='pictures.json解碼錯誤的插畫（sei）：以實機截圖的點陣為準')
    ap.add_argument('--keep', action='append', default=[], metavar='KEY', help='--replace-art時保留原圖面的插畫')
    ap.add_argument('--composite', action='append', default=[], metavar='TARGET=BASE', help='兩張只差局部時，差異範圍外沿用BASE的圖面（fin=end7、yuu3-final=yuu3-base）')
    ap.add_argument('--tile-backup', type=Path, help='被替換人工圖格的備份目錄')
    ap.add_argument('--extra-cells', action='append', default=[], type=Path, metavar='REFS', help='hd_picture_refs.py --single的refs.json：不在任何圖集的插畫（new1）的格位')
    ap.add_argument('--patch-layout', action='append', default=[], type=Path, help='hd_portrait_patch.py的patch.json；可多份，依序與--patch-check配對')
    ap.add_argument('--patch-check', action='append', default=[], type=Path, help='hd_portrait_patch.py --check的收據：須通過，並沿用其中的對位位移')
    ap.add_argument('--patch-art', action='append', default=[], metavar='KEY=PNG', help='人像說話姿態、魔王眼睛循環改用重畫後的補丁；只在重畫範圍貼回底圖HD（須先通過hd_portrait_patch.py --check）')
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
    pictures = {r['key']: r for r in json.loads((w/'pictures.json').read_text())['scenes']}
    cells = {c['key']: (s['key'], c['box']) for s in json.loads((w/'picture-sheets.json').read_text())['sheets'] for c in s['cells']}
    for refs_file in args.extra_cells:
        for s in json.loads((refs_file if refs_file.is_absolute() else p/refs_file).read_text())['sheets']:
            if not s.get('single'):
                continue
            if s['key'] in cells:
                raise ValueError(f"{s['key']} 已在圖集內")
            cells[s['key']] = (s['key'], s['cells'][0]['box'])
    tiles_doc = json.loads((out/'manual-tiles.json').read_text())
    tiles = tiles_doc['tiles']
    transparent = {k: int(v) for k, v in (s.split('=') for s in args.transparent)}
    sheet_art = dict(s.split('=') for s in args.sheet_art)
    retile, retiled = set(args.retile), {}
    from_dump = {k: tuple(map(int, v.split('x'))) for k, v in (s.split('=') for s in args.raster_from_dump)}
    tile_backup = None if args.tile_backup is None else (args.tile_backup if args.tile_backup.is_absolute() else p/args.tile_backup)
    stale = sorted(out.glob('complete-'+PREFIX+'*.png'))
    supplement = []

    def sheet_crop(key, width, height):
        sheet, box = cells[key]
        source = p/sheet_art[sheet] if sheet in sheet_art else w/'generated'/(sheet+'-generated-v1.png')
        atlas = Image.open(source).convert('RGBA')
        box = tuple(round(v*(atlas.width/1536 if i % 2 == 0 else atlas.height/1024)) for i, v in enumerate(box))
        return sheet, source, atlas.crop(box).resize((width*3, height*3), Image.Resampling.LANCZOS)

    def replace_tiles(key, native, painted):
        # hd_manual_art.py的add_tiles依插畫順序首見登錄；同一插畫內首見位置就是登錄位置。
        if tile_backup is None or not tile_backup.is_dir():
            raise ValueError('替換人工圖格需要既有的備份目錄')
        files = {t['file']: t for t in tiles if t['category'] == key}
        height, width = native.shape
        done = set()
        for y in range(0, height-7, 8):
            for x in range(0, width-7, 8):
                cell = native[y:y+8, x:x+8]
                if len(set(cell.flatten().tolist())) < 3 or np.count_nonzero(cell) < 8:
                    continue
                name = 'tile-'+digest(cell.tobytes())[:16]+'.png'
                if name not in files or name in done:
                    continue
                done.add(name)
                if not (tile_backup/name).exists():
                    shutil.copy2(out/name, tile_backup/name)
                painted.crop((x*3, y*3, x*3+24, y*3+24)).save(out/name)
                files[name]['sha256'] = digest((out/name).read_bytes())
        retiled[key] = len(done)
        return len(done)

    def add_fragments(base, parent, native, painted, mask, seen):
        # 8×8片段（兩色以上）；同群組內同點陣與遮罩只登錄首見位置，檔名接在母圖名後。
        height, width = native.shape
        count = 0
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
                file = f'{parent[:-4]}-x{x:03d}-y{y:03d}.png'
                painted.crop((x*3, y*3, (x+fw)*3, (y+fh)*3)).save(out/file)
                fragment = {**base, 'width': fw, 'height': fh, 'raster': base64.b64encode(part).decode(), 'file': file, 'sha256': digest((out/file).read_bytes()), 'derivative_of': parent}
                if mask is not None:
                    fragment['mask'] = base64.b64encode(cut).decode()
                m['entries'].append(fragment)
                count += 1
        return count

    # 沒有實機截圖、不登錄整張插畫的圖：只把人工圖格換成重畫後的圖面（例如外框waku）。
    registered = {s.split('=')[0] for s in args.picture}
    for key in sorted(retile-registered):
        row = pictures[key]
        sheet, source, painted = sheet_crop(key, row['width'], row['height'])
        if sheet not in sheet_art:
            raise ValueError(f'{key} 只換人工圖格時必須指定重畫後的圖集')
        native = np.frombuffer(base64.b64decode(row['raster']), np.uint8).reshape(row['height'], row['width'])
        supplement.append({'key': key, 'tiles_only': True, 'source_art': str(source.relative_to(p)), 'source_art_sha256': digest(source.read_bytes()), 'retiled': replace_tiles(key, native, painted)})
    for spec in args.picture:
        key, rest = spec.split('=')
        path, at = rest.split('@')
        ox, oy = map(int, at.split(','))
        row = pictures[key]
        screen, palette = dump(p/path)
        if key in from_dump:
            # pictures.json的點陣解碼錯誤（色號半位元組換位）；以實機截圖為準，人工圖格仍依原點陣位置替換。
            width, height = from_dump[key]
            if ox < 0 or oy < 0 or ox+width > 320 or oy+height > 200:
                raise ValueError(f'{key} 超出畫面')
            native = screen[oy:oy+height, ox:ox+width].copy()
            raster = native.tobytes()
        else:
            width, height = row['width'], row['height']
            raster = base64.b64decode(row['raster'])
            native = np.frombuffer(raster, np.uint8).reshape(height, width)
            if ox < 0 or oy < 0 or ox+width > 320 or oy+height > 200 or not np.array_equal(screen[oy:oy+height, ox:ox+width], native):
                raise ValueError(f'{key} 原生點陣與實機截圖不符')
        # 與hd_manual_art.py的人工圖格同一次裁切；未換圖面時逐格核對已登錄圖格，證明圖面未改。
        sheet, source, painted = sheet_crop(key, width, height)
        if (sheet in sheet_art) != (key in retile):
            raise ValueError(f'{key} 換圖面時必須同步替換人工圖格')
        checked = 0
        files = {t['file']: t for t in tiles if t['category'] == key}
        if key in retile:
            checked = replace_tiles(key, native, painted)
        for y in range(0, height-7, 8) if key not in retile else ():
            for x in range(0, width-7, 8):
                name = 'tile-'+digest(native[y:y+8, x:x+8].tobytes())[:16]+'.png'
                if name in files:
                    tile = np.array(Image.open(out/name).convert('RGBA'))
                    crop = np.array(painted.crop((x*3, y*3, x*3+24, y*3+24)))
                    if not np.array_equal(tile, crop):
                        continue  # 同點陣先出現在別格時沿用先登錄者，只核首見位置
                    checked += 1
        if checked == 0 and key not in retile:
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
        count = add_fragments(base, parent, native, painted, mask, set())
        supplement.append({'key': name, 'source_art': str(source.relative_to(p)), 'source_art_sha256': digest(source.read_bytes()), 'tiles_checked': checked, 'fragments': count, 'runtime_dump': path, 'origin': [ox, oy], 'transparent_index': transparent.get(key), 'transparent_pixels': int((mask == 0).sum()) if mask is not None else 0, 'retiled': retiled.get(key, 0), 'raster_source': 'runtime dump' if key in from_dump else 'pictures.json'})
    # completion-sources圖集換圖：與hd_complete_register.py相同的裁切，只換PNG與SHA，點陣、遮罩與群組不變。
    completion = json.loads((w/'completion-sources.json').read_text())
    sources = {s['key']: s for s in completion['sources']}
    if args.replace_art and (tile_backup is None or not tile_backup.is_dir()):
        raise ValueError('替換圖面需要既有的備份目錄')
    painted_art, sheet_of = {}, {}
    for spec in args.replace_art:
        sheet_key, path = spec.split('=')
        sheet = next(s for s in completion['sheets'] if s['key'] == sheet_key)
        atlas = Image.open(p/path).convert('RGBA')
        for cell in sheet['cells']:
            key, source = cell['key'], sources[cell['key']]
            if key in args.keep:
                continue
            if source['kind'] != 'picture':
                raise ValueError(f'{key} 不是整張插畫')
            scaled = tuple(round(v*atlas.size[i % 2]/(1536, 1024)[i % 2]) for i, v in enumerate(cell['box']))
            painted_art[key] = atlas.crop(scaled).resize((source['width']*3, source['height']*3), Image.Resampling.LANCZOS)
            sheet_of[key] = (sheet_key, path)
    # 只差局部的兩張（fin只比end7多FIN字樣、yuu3-final只比yuu3-base多主角）：差異範圍外擴2點用目標圖面，
    # 其餘沿用基底圖面，切換時畫面不跳。第0列不計（end7、fin的第0列已遮去）。
    composites = {}
    for spec in args.composite:
        target, base = spec.split('=')
        if target not in painted_art or base not in painted_art:
            raise ValueError(f'{target}／{base} 須同時換圖面')
        a = np.frombuffer(base64.b64decode(sources[target]['raster']), np.uint8).reshape(sources[target]['height'], sources[target]['width'])
        b = np.frombuffer(base64.b64decode(sources[base]['raster']), np.uint8).reshape(sources[base]['height'], sources[base]['width'])
        if a.shape != b.shape:
            raise ValueError(f'{target}／{base} 尺寸不同')
        grow = a != b
        grow[0] = False
        for _ in range(2):
            g = grow.copy()
            g[1:] |= grow[:-1]
            g[:-1] |= grow[1:]
            g[:, 1:] |= grow[:, :-1]
            g[:, :-1] |= grow[:, 1:]
            grow = g
        mask = Image.fromarray(np.repeat(np.repeat(grow, 3, axis=0), 3, axis=1).astype(np.uint8)*255)
        painted_art[target] = Image.composite(painted_art[target], painted_art[base], mask)
        composites[target] = {'base': base, 'target_pixels': int(grow.sum())}
    def rewrite(key, parent, painted):
        # 整張PNG與其8×8片段都從同一張圖面裁出；第一次改寫前備份原檔。
        count = 0
        for e in m['entries']:
            if e['source'] != key or (e['file'] != parent and e.get('derivative_of') != parent):
                continue
            if not (tile_backup/e['file']).exists():
                shutil.copy2(out/e['file'], tile_backup/e['file'])
            if e['file'] == parent:
                image = painted
            else:
                x, y = map(int, e['file'][len(parent)-4:-4].replace('-x', '').split('-y'))
                image = painted.crop((x*3, y*3, (x+e['width'])*3, (y+e['height'])*3))
            image.save(out/e['file'])
            e['sha256'] = digest((out/e['file']).read_bytes())
            count += 1
        return count

    replaced = {}
    for key, painted in painted_art.items():
        replaced[key] = rewrite(key, f'complete-{key}-000.png', painted)
    # 人像說話姿態與魔王眼睛循環：部件都是底圖HD，只在顏色與底圖不同的重畫範圍貼回生成結果並漸變銜接。
    # 底圖HD須是本次登錄寫出的版本，且與做參考圖時相同；否則補丁四周的上下文已不同，須重做參考圖。
    if args.patch_art:
        if not args.patch_layout or len(args.patch_layout) != len(args.patch_check) or tile_backup is None or not tile_backup.is_dir():
            raise ValueError('補丁需要成對的--patch-layout、--patch-check與既有的備份目錄')
        layout, checked, files_of = {}, {}, {}
        for layout_file, check_file in zip(args.patch_layout, args.patch_check):
            layout_file = layout_file if layout_file.is_absolute() else p/layout_file
            check_file = check_file if check_file.is_absolute() else p/check_file
            check = json.loads(check_file.read_text())
            if check['layout_sha256'] != digest(layout_file.read_bytes()):
                raise ValueError('補丁檢查收據與參考圖排版不同')
            for s in json.loads(layout_file.read_text())['sets']:
                layout[s['key']] = s
                files_of[s['key']] = (layout_file, check_file)
            checked.update({s['key']: s for s in check['sets']})
    for spec in args.patch_art:
        key, path = spec.split('=')
        st = layout[key]
        source = st if st.get('dump_parts') else sources[key]
        layout_file, check_file = files_of[key]
        result = checked.get(key)
        if result is None or result['art_sha256'] != digest((p/path).read_bytes()) or not all(r['pass'] for r in result['parts']):
            raise ValueError(f'{key} 補丁圖沒有通過hd_portrait_patch.py --check')
        offsets = {r['index']: r['offset'] for r in result['parts']}
        base_file = p/st['base_png']
        if digest(base_file.read_bytes()) != st['base_png_sha256']:
            raise ValueError(f'{key} 底圖HD與參考圖不同，須重做補丁參考圖')
        base_im = Image.open(base_file).convert('RGBA')
        art = Image.open(p/path).convert('RGBA')
        if art.width*1024 != art.height*1536:
            raise ValueError(f'{key} 補丁圖長寬比不是3:2')
        files, same = 0, []
        if st.get('dump_parts'):
            # 實機畫格組沒有既有項目：另建story-群組，整張部件與8×8片段都從貼好補丁的圖面裁出。
            group, name, seen = len(m['groups']), PREFIX+key, set()
            m['groups'].append({'key': name, 'bindings': [], 'method': 'hd_portrait_patch.py實機畫格組：底圖HD只在顏色不同處換圖', 'base_key': st['base_key'],
                                'runtime_dumps': [{'file': q['dump'], 'sha256': q['dump_sha256'], 'origin': q['origin']} for q in st['parts']]})
            colors = base64.b64encode((p/st['palette_source']).read_bytes()[64000:]).decode()
        for part in st['parts']:
            n, pw, ph = part['index'], part['width'], part['height']
            if digest(base64.b64decode(source['parts'][n]['raster'])) != part['raster_sha256']:
                raise ValueError(f'{key}#{n} 部件點陣與參考圖不同')
            image = base_im.crop((0, 0, pw*3, ph*3))
            if part.get('identical'):
                same.append(n)
            else:
                rx0, ry0, rx1, ry1 = part['region']
                size = ((rx1-rx0)*3, (ry1-ry0)*3)
                # 沿用檢查收據的對位位移（參考圖1536×1024座標）。
                dx, dy = offsets[n]
                region_box = [part['sheet_region'][0]+dx, part['sheet_region'][1]+dy, part['sheet_region'][2]+dx, part['sheet_region'][3]+dy]
                box = tuple(round(v*(art.width/1536 if i % 2 == 0 else art.height/1024)) for i, v in enumerate(region_box))
                patch = art.crop(box).resize(size, Image.Resampling.LANCZOS)
                # 漸變遮罩以最近點放大，範圍外為0，與底圖HD逐點相同。
                alpha = Image.frombytes('L', (rx1-rx0, ry1-ry0), base64.b64decode(part['alpha'])).resize(size, Image.Resampling.NEAREST)
                image.paste(Image.composite(patch, image.crop((rx0*3, ry0*3, rx1*3, ry1*3)), alpha), (rx0*3, ry0*3))
            image.putalpha(255)
            if not st.get('dump_parts'):
                files += rewrite(key, f'complete-{key}-{n:03d}.png', image)
            elif not part.get('identical'):
                parent = f'complete-{name}-{n:03d}.png'
                image.save(out/parent)
                native = np.frombuffer(base64.b64decode(part['raster']), np.uint8).reshape(ph, pw)
                base = {'group': group, 'source': name, 'colors': colors}
                m['entries'].append({**base, 'width': pw, 'height': ph, 'raster': part['raster'], 'file': parent, 'sha256': digest((out/parent).read_bytes())})
                files += 1+add_fragments(base, parent, native, image, None, seen)
        supplement.append({'patch_set': key, 'patch_art': path, 'patch_art_sha256': digest((p/path).read_bytes()), 'patch_layout': str(layout_file.relative_to(p)),
                           'patch_layout_sha256': digest(layout_file.read_bytes()), 'patch_check': str(check_file.relative_to(p)), 'patch_check_sha256': digest(check_file.read_bytes()), 'offsets': {str(k): v for k, v in offsets.items()}, 'base_png': st['base_png'], 'base_png_sha256': st['base_png_sha256'],
                           'parts': len(st['parts']), 'identical_parts': same, 'files': files})
    for spec in args.replace_art:
        sheet_key, path = spec.split('=')
        sheet = next(s for s in completion['sheets'] if s['key'] == sheet_key)
        supplement.append({'sheet': sheet_key, 'replaced_art': path, 'replaced_art_sha256': digest((p/path).read_bytes()), 'files': {c['key']: replaced.get(c['key'], 0) for c in sheet['cells']},
                           'kept': [c['key'] for c in sheet['cells'] if c['key'] in args.keep], 'composites': {k: v for k, v in composites.items() if sheet_of[k][0] == sheet_key}})
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
    # 本工具產生的圖檔都以complete-story-開頭；全部寫完後才移除本次未再引用的舊片段。
    referenced = {e['file'] for e in m['entries']}
    for old in stale:
        if old.name not in referenced:
            old.unlink()
    if retiled:
        if not (tile_backup/'manual-tiles.json').exists():
            shutil.copy2(out/'manual-tiles.json', tile_backup/'manual-tiles.json')
        (out/'manual-tiles.json').write_text(json.dumps(tiles_doc, ensure_ascii=False, indent=2)+'\n')
    print(f'SUPPLEMENTED {len(args.picture)} pictures, {len(args.row0)} row-0 masks, {len(m["entries"])} entries; no game launched')


if __name__ == '__main__':
    main()

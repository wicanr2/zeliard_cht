#!/usr/bin/env python3
"""人像說話姿態與魔王眼睛循環的補丁參考圖。只讀已登錄的HD底圖與原版點陣，不生圖、不啟動遊戲。規格068。

人像組（portrait-*-all）每個姿態都是同一張96×88底圖，只在嘴或眼換成另一個狀態；魔王眼睛循環
（dmaou-cycle-all）是魔王插畫上方136×48，換了色號但只有雙眼附近的顏色不同。兩者都從底圖左上角對齊。
本工具以實機色盤比較每個部件與底圖的顏色：顏色不同的點外擴1點為重畫範圍，參考圖在範圍內貼原版點陣的
實機色，四周保留已重畫的底圖HD作上下文，排成1536×1024交給生圖工具只重畫這些範圍。
--check以面板上未改動的上下文找出生成時的位移，並比對重畫範圍與底圖HD同色號的平均色，收據供登錄沿用位移。
登錄時（hd_picture_complete.py --patch-art）每個部件都是底圖HD，只在重畫範圍貼回生成結果，再往外2點
漸變銜接；範圍外與底圖HD逐點相同。顏色與底圖相同的部件直接用底圖HD。說話與眨眼時臉不會跳。
--dump-set的部件取自實機截圖：片頭yuu3的灰光剪影畫格沒有來源記錄，new1的說話嘴型只在實機看過；登錄時另建群組。
"""
import argparse
import base64
import colorsys
import hashlib
import json
import math
from pathlib import Path

import numpy as np
from PIL import Image

CORE = 1      # 顏色不同的點外擴成重畫範圍的原生點數
GROW = 2      # 重畫範圍外漸變銜接的原生點數
CONTEXT = 12  # 參考圖上範圍四周保留的底圖原生點數
SHEET = (1536, 1024)
# --check的對位搜尋半徑（參考圖點數）。
SEARCH = 48
# --check的上限（亮度0..255、RGB距離）。對位正確時上下文差不超過8、漸變帶差不超過17，錯位一欄時為29到60、35到44。
# 色號平均色以整張底圖為準，嘴內紅色在底圖可能對到披風同色號，實測最高約90／105；紅綠互換在177以上。
MAX_CONTEXT_MAD = 12
MAX_RING_MAD = 20
MAX_MEAN_DISTANCE = 100
MAX_MAJOR_DISTANCE = 130


def digest(data):
    return hashlib.sha256(data).hexdigest()


def palette_of(path):
    data = path.read_bytes()
    if len(data) == 64768:   # 實機.indexed截圖：320×200色號加256色盤
        return np.frombuffer(data[64000:], np.uint8).reshape(256, 3)
    if len(data) == 768:     # 已存的色盤檔
        return np.frombuffer(data, np.uint8).reshape(256, 3)
    raise ValueError(f'色盤來源格式不符：{path}')


def dump_at(p, spec):
    """DUMP@X,Y：實機.indexed截圖、原生左上角。"""
    path, at = spec.split('@')
    x, y = map(int, at.split(','))
    data = (p/path).read_bytes()
    if len(data) != 64768:
        raise ValueError(f'截圖格式不符：{path}')
    return path, (x, y), np.frombuffer(data[:64000], np.uint8).reshape(200, 320), np.frombuffer(data[64000:], np.uint8).reshape(256, 3)


def base_raster_of(p, pictures, base_key, painting):
    """底圖點陣：pictures.json／completion-sources的鍵，或DUMP@X,Y（點陣解碼錯誤、以實機為準的插畫，如new1）。"""
    bw, bh = painting.width//3, painting.height//3
    if '@' in base_key:
        path, (x, y), screen, _ = dump_at(p, base_key)
        if x < 0 or y < 0 or x+bw > 320 or y+bh > 200:
            raise ValueError(f'{base_key} 超出畫面')
        return screen[y:y+bh, x:x+bw].copy()
    base = pictures[base_key]
    if (base['width'], base['height']) != (bw, bh):
        raise ValueError(f'{base_key} 底圖HD尺寸不符')
    return np.frombuffer(base64.b64decode(base['raster']), np.uint8).reshape(bh, bw)


def dilate(mask):
    out = mask.copy()
    out[1:] |= mask[:-1]
    out[:-1] |= mask[1:]
    out[:, 1:] |= out[:, :-1].copy()
    out[:, :-1] |= out[:, 1:].copy()
    return out


def check(args):
    """重畫結果的對位與一致性。生圖常把面板內容整體挪動數點，直接照格位裁切會在漸變帶露出錯位；
    以面板上未改動的底圖上下文（重畫範圍外再擴2點不計）找最小平均亮度差的位移，登錄時沿用。
    一致性以底圖HD為準：重畫範圍每個色號的平均色，與同色號在整張底圖HD的平均色比較（底圖已過實機色盤閘門）；
    底圖沒有的色號才比實機色盤色相。"""
    p = args.project
    if args.layout is None or args.output.exists():
        raise ValueError('--check需要--layout，且收據不得已存在')
    layout_file = args.layout if args.layout.is_absolute() else p/args.layout
    sets = {s['key']: s for s in json.loads(layout_file.read_text())['sets']}
    w = p/'workplace/full-hd-v1/manual-hd-v1'
    completion = json.loads((w/'completion-sources.json').read_text())
    sources = {s['key']: s for s in completion['sources']}
    pictures = {r['key']: r for r in json.loads((w/'pictures.json').read_text())['scenes']}
    pictures.update({k: v for k, v in sources.items() if v['kind'] == 'picture'})
    gray = lambda a: a[..., 0]*0.299+a[..., 1]*0.587+a[..., 2]*0.114
    result = {'schema': 1, 'tool': 'tools/hd_portrait_patch.py', 'tool_sha256': digest((p/'tools/hd_portrait_patch.py').read_bytes()),
              'layout': str(layout_file.relative_to(p)), 'layout_sha256': digest(layout_file.read_bytes()),
              'limits': {'context_mad': MAX_CONTEXT_MAD, 'ring_mad': MAX_RING_MAD, 'mean_distance': MAX_MEAN_DISTANCE, 'major_distance': MAX_MAJOR_DISTANCE},
              'sets': [], 'game_launched': False}
    passed = True
    for spec in args.check:
        key, path = spec.split('=')
        st = sets[key]
        ref_sheet = gray(np.asarray(Image.open(layout_file.parent/st['file']).convert('RGB'), np.float64))
        art = Image.open(p/path).convert('RGB')
        if art.width*SHEET[1] != art.height*SHEET[0]:
            raise ValueError(f'{key} 重畫圖長寬比不是3:2')
        art = art.resize(SHEET, Image.Resampling.LANCZOS) if art.size != SHEET else art
        gen = gray(np.asarray(art, np.float64))
        palette = palette_of(p/st['palette_source'])
        painting = Image.open(p/st['base_png']).convert('RGB')
        bw, bh = painting.width//3, painting.height//3
        base_raster = base_raster_of(p, pictures, st['base_key'], painting)
        native_art = np.asarray(painting, np.float64).reshape(bh, 3, bw, 3, 3).mean(axis=(1, 3))
        means = {int(k): native_art[base_raster == k].mean(axis=0) for k in np.unique(base_raster)}
        rows = []
        for part in st['parts']:
            if part.get('identical'):
                continue
            n, s = part['index'], part['scale']
            ox, oy = part['sheet_origin']
            cx0, cy0, cx1, cy1 = part['context']
            rx0, ry0, rx1, ry1 = part['region']
            ph, pw = (cy1-cy0)*s, (cx1-cx0)*s
            alpha = np.frombuffer(base64.b64decode(part['alpha']), np.uint8).reshape(ry1-ry0, rx1-rx0)
            excl = np.zeros((cy1-cy0, cx1-cx0), bool)
            excl[ry0-cy0:ry1-cy0, rx0-cx0:rx1-cx0] = alpha > 0
            for _ in range(2):
                excl = dilate(excl)
            keep = np.repeat(np.repeat(~excl, s, axis=0), s, axis=1)
            ref = ref_sheet[oy:oy+ph, ox:ox+pw]
            radius, step = SEARCH, 4  # 生成時整欄可能平移數十點；先粗搜再細搜

            def mad(dx, dy):
                y0, x0 = oy+dy, ox+dx
                if y0 < 0 or x0 < 0 or y0+ph > SHEET[1] or x0+pw > SHEET[0]:
                    return float('inf')
                return float(np.abs(gen[y0:y0+ph, x0:x0+pw]-ref)[keep].mean())
            best = min((mad(dx, dy), dx, dy) for dy in range(-radius, radius+1, step) for dx in range(-radius, radius+1, step))
            _, bx, by = best
            best = min((mad(dx, dy), dx, dy) for dy in range(by-step, by+step+1) for dx in range(bx-step, bx+step+1))
            context_mad, dx, dy = best
            # 對位後裁出重畫範圍，與登錄相同的裁切。
            sx0, sy0, sx1, sy1 = part['sheet_region']
            crop = art.crop((sx0+dx, sy0+dy, sx1+dx, sy1+dy)).resize(((rx1-rx0)*3, (ry1-ry0)*3), Image.Resampling.LANCZOS)
            hd = np.asarray(crop, np.float64)
            under = np.asarray(painting.crop((rx0*3, ry0*3, rx1*3, ry1*3)), np.float64)
            ring = np.repeat(np.repeat((alpha > 0) & (alpha < 255), 3, axis=0), 3, axis=1)
            ring_mad = float(np.abs(gray(hd)-gray(under))[ring].mean()) if ring.any() else 0.0
            native = hd.reshape(ry1-ry0, 3, rx1-rx0, 3, 3).mean(axis=(1, 3))
            raster = np.frombuffer(base64.b64decode(part['raster'] if 'raster' in part else sources[key]['parts'][n]['raster']), np.uint8).reshape(part['height'], part['width'])[ry0:ry1, rx0:rx1]
            core = alpha == 255
            total, weighted, worst, colors = 0, 0.0, 0.0, []
            for k in np.unique(raster[core]):
                m = core & (raster == k)
                count = int(m.sum())
                got = native[m].mean(axis=0)
                if int(k) in means:
                    want, basis = means[int(k)], 'base painting'
                    d = float(np.linalg.norm(got-want))
                else:
                    want, basis = palette[k].astype(np.float64), 'runtime palette hue'
                    hw, hg = colorsys.rgb_to_hsv(*(want/255)), colorsys.rgb_to_hsv(*(got/255))
                    if hw[1] < 0.35 or hw[2] < 0.25:
                        continue
                    diff = abs(hw[0]-hg[0])*360
                    d = min(diff, 360-diff)*2  # 色相差1度約當RGB距離2，與底圖比較同一尺度
                total += count
                weighted += d*count
                if count >= 0.1*core.sum():
                    worst = max(worst, d)
                colors.append({'index': int(k), 'pixels': count, 'basis': basis, 'want_rgb': [round(c) for c in want], 'got_rgb': [round(c) for c in got], 'distance': round(d, 1)})
            mean = weighted/total if total else 0.0
            ok = bool(context_mad <= MAX_CONTEXT_MAD and ring_mad <= MAX_RING_MAD and mean <= MAX_MEAN_DISTANCE and worst <= MAX_MAJOR_DISTANCE)
            passed = passed and ok
            colors.sort(key=lambda c: -c['pixels'])
            rows.append({'index': n, 'offset': [dx, dy], 'context_mad': round(context_mad, 2), 'unaligned_context_mad': round(mad(0, 0), 2), 'ring_mad': round(ring_mad, 2),
                         'mean_distance': round(mean, 1), 'major_distance': round(worst, 1), 'pass': ok, 'colors': colors[:10]})
            print(f'{key}#{n}: offset ({dx},{dy}) context {context_mad:.1f} (unaligned {mad(0, 0):.1f}) ring {ring_mad:.1f} mean {mean:.1f} major {worst:.1f} {"PASS" if ok else "FAIL"}')
        result['sets'].append({'key': key, 'art': path, 'art_sha256': digest((p/path).read_bytes()), 'parts': rows})
    result['pass'] = passed
    args.output.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print('PATCH_CHECK', 'PASS' if passed else 'FAIL')


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--output', type=Path, required=True, help='做參考圖：既有空目錄；--check：收據檔，須不存在')
    ap.add_argument('--set', action='append', default=[], metavar='KEY=BASEKEY=BASEPNG=PALETTE',
                    help='部件組、底圖點陣鍵（pictures.json或completion-sources）、底圖HD，以及實機截圖或色盤檔')
    ap.add_argument('--dump-set', action='append', default=[], metavar='KEY=BASE=BASEPNG=DUMP@X,Y[+DUMP@X,Y...]',
                    help='實機畫格組（剪影yuu3-dark、new1說話嘴型）：部件取自實機截圖；BASE為底圖點陣鍵或DUMP@X,Y')
    ap.add_argument('--close', action='append', default=[], metavar='KEY=N', help='該組的重畫範圍先閉合N點並補洞')
    ap.add_argument('--layout', type=Path, help='--check用：做參考圖時的patch.json')
    ap.add_argument('--check', action='append', default=[], metavar='KEY=PNG', help='檢查重畫結果：對位並比對底圖HD同色號平均色')
    args = ap.parse_args()
    if args.check:
        return check(args)
    p, out = args.project, args.output
    if not out.is_dir() or any(out.iterdir()):
        raise ValueError('輸出須為既有空目錄')
    w = p/'workplace/full-hd-v1/manual-hd-v1'
    completion = json.loads((w/'completion-sources.json').read_text())
    sources = {s['key']: s for s in completion['sources']}
    pictures = {r['key']: r for r in json.loads((w/'pictures.json').read_text())['scenes']}
    pictures.update({k: v for k, v in sources.items() if v['kind'] == 'picture'})
    receipt = {'schema': 2, 'tool': 'tools/hd_portrait_patch.py', 'tool_sha256': digest((p/'tools/hd_portrait_patch.py').read_bytes()),
               'core': CORE, 'grow': GROW, 'context': CONTEXT, 'sets': [], 'game_launched': False}
    refs = {'schema': 1, 'tool': 'tools/hd_portrait_patch.py', 'pictures': [], 'sheets': [], 'game_launched': False}
    closing = {k: int(v) for k, v in (s.split('=') for s in args.close)}
    jobs = [(*spec.split('='), None) for spec in args.set]
    jobs += [(key, base_key, base_png, None, frames.split('+')) for key, base_key, base_png, frames in (spec.split('=') for spec in args.dump_set)]
    for key, base_key, base_png, palette_file, frames in jobs:
        painting = Image.open(p/base_png).convert('RGB')
        bw, bh = painting.width//3, painting.height//3
        base_raster = base_raster_of(p, pictures, base_key, painting)
        if frames is None:
            source = sources[key]
            palette = palette_of(p/palette_file)
        else:
            # 實機畫格組：部件取自實機截圖（與底圖同大小、同色盤），登錄時另建群組。
            if key in sources:
                raise ValueError(f'{key} 與既有來源同名')
            source = {'parts': []}
            for i, frame in enumerate(frames):
                path, (x, y), screen, pal = dump_at(p, frame)
                if x < 0 or y < 0 or x+bw > 320 or y+bh > 200:
                    raise ValueError(f'{frame} 超出畫面')
                if i == 0:
                    palette, palette_file = pal, path
                elif not np.array_equal(pal, palette):
                    raise ValueError(f'{key} 各畫格色盤不同')
                source['parts'].append({'width': bw, 'height': bh, 'raster': base64.b64encode(screen[y:y+bh, x:x+bw].tobytes()).decode(),
                                        'dump': path, 'dump_sha256': digest((p/path).read_bytes()), 'origin': [x, y]})
        parts, changed = [], []
        for n, part in enumerate(source['parts']):
            pw, ph = part['width'], part['height']
            if pw > bw or ph > bh:
                raise ValueError(f'{key}#{n} 比底圖大')
            raster = np.frombuffer(base64.b64decode(part['raster']), np.uint8).reshape(ph, pw)
            diff = (palette[raster] != palette[base_raster[:ph, :pw]]).any(axis=2)
            if part.get('patch'):
                x, y, w_, h_ = part['patch']
                outside = np.ones((ph, pw), bool)
                outside[y:y+h_, x:x+w_] = False
                if diff[outside].any():
                    raise ValueError(f'{key}#{n} 宣告的改寫矩形外顏色與底圖不同')
            entry = {'index': n, 'width': pw, 'height': ph, 'raster_sha256': digest(raster.tobytes()), 'patch': part.get('patch')}
            if 'dump' in part:
                entry.update({k: part[k] for k in ('raster', 'dump', 'dump_sha256', 'origin')})
            parts.append(entry)
            if not diff.any():
                entry['identical'] = True
                continue
            core = diff
            if closing.get(key):
                # 閉合並補洞：原版人物常以不變的黑點勾身體，換色時只有黑點之間的色點變了；
                # 不納入這些黑點，底圖HD的舊顏色會從縫隙露出（yuu3剪影畫格的主角）。
                for _ in range(closing[key]):
                    core = dilate(core)
                for _ in range(closing[key]):
                    core = ~dilate(~core)
                core = core | diff
                outside = np.zeros_like(core)
                outside[0], outside[-1], outside[:, 0], outside[:, -1] = ~core[0], ~core[-1], ~core[:, 0], ~core[:, -1]
                while True:
                    grown = dilate(outside) & ~core
                    if np.array_equal(grown, outside):
                        break
                    outside = grown
                core = ~outside
            for _ in range(CORE):
                core = dilate(core)
            # 漸變：重畫範圍內為255，往外每點遞減，GROW點後為0。
            alpha = core.astype(np.float64)
            ring = core
            for k in range(1, GROW+1):
                grown = dilate(ring)
                alpha[grown & ~ring] = 1-k/(GROW+1)
                ring = grown
            ys, xs = np.nonzero(ring)
            region = [int(xs.min()), int(ys.min()), int(xs.max())+1, int(ys.max())+1]
            context = [max(0, region[0]-CONTEXT), max(0, region[1]-CONTEXT), min(bw, region[2]+CONTEXT), min(bh, region[3]+CONTEXT)]
            rx0, ry0, rx1, ry1 = region
            entry.update({'diff_pixels': int(diff.sum()), 'core_pixels': int(core.sum()), 'region': region, 'context': context,
                          'alpha': base64.b64encode(np.round(alpha[ry0:ry1, rx0:rx1]*255).astype(np.uint8).tobytes()).decode()})
            changed.append((entry, raster, core))
        # 格位：變化部件依序排成格子，欄數取讓最小放大倍率最大者；每格以整數倍率放大上下文。
        def fit(cols):
            cw, ch = SHEET[0]//cols, SHEET[1]//math.ceil(len(changed)/cols)
            return min(min((cw-24)//(e['context'][2]-e['context'][0]), (ch-24)//(e['context'][3]-e['context'][1])) for e, _, _ in changed)
        cols = max(range(1, len(changed)+1), key=fit)
        rows = math.ceil(len(changed)/cols)
        cw, ch = SHEET[0]//cols, SHEET[1]//rows
        sheet = Image.new('RGB', SHEET, 'black')
        cells = []
        for i, (entry, raster, core) in enumerate(changed):
            cx0, cy0, cx1, cy1 = entry['context']
            scale = min((cw-24)//(cx1-cx0), (ch-24)//(cy1-cy0))
            if scale < 3:
                raise ValueError(f'{key}#{entry["index"]} 格子太小')
            size = ((cx1-cx0)*scale, (cy1-cy0)*scale)
            ox, oy = i % cols*cw+(cw-size[0])//2, i//cols*ch+(ch-size[1])//2
            panel = np.array(painting.crop((cx0*3, cy0*3, cx1*3, cy1*3)).resize(size, Image.Resampling.LANCZOS))
            # 上下文可超出部件（魔王眼睛循環只有上方48列）；部件外沒有點陣，只留底圖HD。
            hh, ww = min(cy1, raster.shape[0])-cy0, min(cx1, raster.shape[1])-cx0
            flat = np.zeros((cy1-cy0, cx1-cx0, 3), np.uint8)
            sel = np.zeros((cy1-cy0, cx1-cx0), bool)
            flat[:hh, :ww] = palette[raster[cy0:cy0+hh, cx0:cx0+ww]]
            sel[:hh, :ww] = core[cy0:cy0+hh, cx0:cx0+ww]
            flat = np.repeat(np.repeat(flat, scale, axis=0), scale, axis=1)
            sel = np.repeat(np.repeat(sel, scale, axis=0), scale, axis=1)
            panel[sel] = flat[sel]
            sheet.paste(Image.fromarray(panel), (ox, oy))
            rx0, ry0, rx1, ry1 = entry['region']
            entry.update({'scale': scale, 'sheet_origin': [ox, oy],
                          'sheet_region': [ox+(rx0-cx0)*scale, oy+(ry0-cy0)*scale, ox+(rx1-cx0)*scale, oy+(ry1-cy0)*scale]})
            name = f'{key}#{entry["index"]}'
            region = raster[ry0:ry1, rx0:rx1]
            cells.append({'key': name, 'box': entry['sheet_region']})
            refs['pictures'].append({'key': name, 'width': rx1-rx0, 'height': ry1-ry0, 'raster_source': 'completion-sources part region',
                                     'raster': base64.b64encode(region.tobytes()).decode(), 'raster_sha256': digest(region.tobytes()),
                                     'palette': base64.b64encode(palette.tobytes()).decode(), 'palette_source': palette_file})
        file = out/f'{key}-v2-selected.png'
        sheet.save(file)
        refs['sheets'].append({'key': key, 'file': file.name, 'sha256': digest(file.read_bytes()), 'cells': cells})
        receipt['sets'].append({'key': key, 'base_key': base_key, 'base_png': base_png, 'base_png_sha256': digest((p/base_png).read_bytes()),
                                'palette_source': palette_file, 'palette_source_sha256': digest((p/palette_file).read_bytes()),
                                'file': file.name, 'sha256': digest(file.read_bytes()), 'dump_parts': frames is not None, 'close': closing.get(key, 0), 'parts': parts})
    (out/'patch.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+'\n')
    (out/'refs.json').write_text(json.dumps(refs, ensure_ascii=False, indent=2)+'\n')
    print(f'PATCH_SHEETS {len(jobs)} sets; no game launched')


if __name__ == '__main__':
    main()

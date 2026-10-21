#!/usr/bin/env python3
"""Docker 內以原版 8×8 圖格及既有 HD 元件重組 README 完整角色。"""
import argparse
import base64
import hashlib
import json
import os
from pathlib import Path
from urllib.request import urlopen

import numpy as np
from PIL import Image, ImageDraw, ImageFont, ImageOps


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def color_code(rgb):
    """保留全部 RGB 三階色值，不能只靠輪廓或灰階匹配。"""
    levels = np.rint(np.asarray(rgb) / 125.5).astype(np.uint8)
    return levels[..., 0] * 9 + levels[..., 1] * 3 + levels[..., 2]


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--fetch-references', action='store_true',
                    help='只在允許網路的 Docker 內下載固定雜湊的布局參考')
    args = ap.parse_args()
    root = args.project.resolve()
    artdir = root / 'workplace/full-hd-v1/manual-hd-v1'
    work = root / 'workplace/readme-creatures-v2'
    out = root / 'docs/images'
    for folder in (work, out):
        assert folder.is_dir() and folder.stat().st_uid == os.getuid(), folder
    for name in ('enemies-overview.jpg', 'bosses-overview.jpg', 'creatures-provenance.json'):
        path = out / name
        assert not path.exists() or path.stat().st_uid == os.getuid(), path
    refs = work / 'reference'
    refs.mkdir(exist_ok=True)
    assembled = work / 'assembled'
    assembled.mkdir(exist_ok=True)
    sources = {}

    def source(path, role, **extra):
        sources[str(path.relative_to(root))] = {'sha256': digest(path), 'role': role, **extra}

    layoutfile = out / 'creatures-layout.json'
    layout = json.loads(layoutfile.read_text())
    source(layoutfile, '固定代表姿態布局')
    for name, record in layout['references'].items():
        path = refs / name
        if not path.exists() and args.fetch_references:
            with urlopen(record['url'], timeout=30) as response:
                data = response.read()
            assert hashlib.sha256(data).hexdigest() == record['sha256'], name
            path.write_bytes(data)
        assert digest(path) == record['sha256'], name
        source(path, '僅用於布局比對，不複製參考像素', url=record['url'])

    catalogfile = artdir / 'registered/manual-art.json'
    catalog = json.loads(catalogfile.read_text())
    source(catalogfile, '原生 16×16 模板與既有 HD 元件登錄')
    palettefile = root / 'workplace/music-release-v1/release-build/AppDir/assets/town/frame/palette.bin'
    palette = np.frombuffer(palettefile.read_bytes(), np.uint8).reshape(256, 3)
    source(palettefile, '本地原版來源色表')
    wholefile = artdir / 'registered/coherent-bosses.json'
    whole = {}
    if wholefile.exists():
        source(wholefile, '058完整構圖首領，同一份素材接入正式覆繪')
        whole = {r['key']: r for r in json.loads(wholefile.read_text())['figures']}
    source(root / 'workplace/orig/ZELRES3.SAR', '原版怪物來源封存檔')
    fontfile = artdir / 'rounded-font/jf-openhuninn-2.1.ttf'
    source(fontfile, '既有粉圓字型')
    fonts = {n: ImageFont.truetype(str(fontfile), n) for n in (18, 20, 24, 30, 40)}
    index = {}
    for obj in catalog['objects']:
        if 'bank' not in obj:
            continue
        raw = np.frombuffer(base64.b64decode(obj['raster']), np.uint8).reshape(16, 16)
        codes = color_code(palette[raw])
        bank = index.setdefault(obj['bank'], {})
        for q in range(4):
            tile = codes[q // 2 * 8:q // 2 * 8 + 8, q % 2 * 8:q % 2 * 8 + 8]
            if np.any(tile):
                bank.setdefault(tile.tobytes(), []).append((obj, q, raw, codes))
    hd_cache = {}

    def hd_art(obj):
        filename = obj['file']
        if filename not in hd_cache:
            path = artdir / 'registered' / filename
            assert digest(path) == obj['sha256'], filename
            source(path, '既有已登錄 HD 元件；逐象限原樣取用')
            hd_cache[filename] = Image.open(path).convert('RGBA')
            assert hd_cache[filename].size == (48, 48), filename
        return hd_cache[filename]

    def assemble(figure):
        pixels = np.array(Image.open(refs / figure['reference']).convert('RGB').crop(figure['box']))
        pixels[~np.all((pixels == 0) | (pixels == 255), axis=2)] = 0
        h2, w2, _ = pixels.shape
        assert h2 % 2 == w2 % 2 == 0
        target = color_code(pixels.reshape(h2 // 2, 2, w2 // 2, 2, 3).mean(axis=(1, 3)) * 251 / 255)
        lookup = np.arange(27, dtype=np.uint8)
        for a, b in figure.get('layout_palette_map', {}).items():
            lookup[int(a)] = int(b)
        target = lookup[target]
        h, w = target.shape
        bank = index[figure['bank']]
        best = None
        for oy in range(8):
            for ox in range(8):
                padded = np.pad(target, ((oy, 16), (ox, 16)))
                missing, parts = 0, []
                for y in range(0, h + oy, 8):
                    for x in range(0, w + ox, 8):
                        tile = padded[y:y + 8, x:x + 8]
                        if not np.any(tile):
                            continue
                        matches = bank.get(tile.tobytes())
                        if matches is None:
                            missing += int(np.count_nonzero(tile))
                        else:
                            parts.append((x - ox, y - oy, matches))
                if best is None or missing < best[0]:
                    best = missing, ox, oy, parts
                if missing == 0:
                    break
            if best[0] == 0:
                break
        missing, ox, oy, parts = best
        assert missing == 0, (figure['key'], missing)
        native = Image.new('RGBA', (w, h))
        hd = Image.new('RGBA', (w * 3, h * 3))
        proof = []
        padded = np.pad(target, ((16, 16), (16, 16)))
        for x, y, choices in parts:
            # 相同 8×8 圖格優先採周圍 16×16 最吻合的登錄母圖。
            def context_score(choice):
                _, q, _, codes = choice
                left, top = x - q % 2 * 8, y - q // 2 * 8
                view = padded[top + 16:top + 32, left + 16:left + 32]
                return int(np.count_nonzero(view == codes)) if view.shape == codes.shape else 0
            obj, q, raw, _ = max(choices, key=context_score)
            tile = raw[q // 2 * 8:q // 2 * 8 + 8, q % 2 * 8:q % 2 * 8 + 8]
            rgba = np.dstack((palette[tile], np.where(tile != 0, 255, 0).astype(np.uint8)))
            native.paste(Image.fromarray(rgba), (x, y))
            art = hd_art(obj).crop((q % 2 * 24, q // 2 * 24, q % 2 * 24 + 24, q // 2 * 24 + 24))
            hd.paste(art, (x * 3, y * 3))
            proof.append({'native_position': [x, y], 'registered_file': obj['file'],
                          'registered_sha256': obj['sha256'], 'object_key': obj['key'],
                          'template': obj['template'], 'quad': q})
        assert np.array_equal(color_code(np.array(native.convert('RGB'))), target), figure['key']
        expected_alpha = np.repeat(np.repeat(target != 0, 3, axis=0), 3, axis=1)
        actual_alpha = np.array(hd.getchannel('A')) != 0
        # 既有 HD 含柔化透明邊及內部留白；保留原 RGBA，只拒絕輪廓外貼片。
        assert np.any(actual_alpha) and not np.any(actual_alpha & ~expected_alpha), figure['key']
        uncovered = int(np.count_nonzero(expected_alpha & ~actual_alpha))
        # 與正式覆繪層相同，以原版像素作下層，不把透明接縫變成缺片。
        hd = Image.alpha_composite(native.resize((w * 3, h * 3), Image.Resampling.NEAREST), hd)
        assert np.array_equal(np.array(hd.getchannel('A')) != 0, expected_alpha), figure['key']
        bounds = native.getbbox()
        assert bounds is not None
        native = native.crop(bounds)
        hd = hd.crop(tuple(v * 3 for v in bounds))
        whole_source = None
        if figure['key'] in whole:
            r = whole[figure['key']]
            raw = np.frombuffer(base64.b64decode(r['raster']), np.uint8).reshape(r['height'], r['width'])
            assert np.array_equal(np.array(native.convert('RGB')), palette[raw]), figure['key']
            path = artdir / 'registered' / r['file']
            assert digest(path) == r['sha256'], path
            source(path, '058完整首領構圖，與遊戲接入相同')
            hd = Image.open(path).convert('RGBA')
            assert hd.size == (native.width * 3, native.height * 3)
            assert np.array_equal(np.array(hd.getchannel('A')) != 0, np.repeat(np.repeat(raw != 0, 3, axis=0), 3, axis=1))
            whole_source = {'file': r['file'], 'sha256': r['sha256'], 'method': '058完整構圖優先'}
        images = []
        for name, picture in (('original', native), ('hd', hd)):
            path = assembled / f'{figure["key"]}-{name}.png'
            picture.save(path)
            images.append({'path': str(path.relative_to(root)), 'sha256': digest(path), 'size': list(picture.size)})
        record = {**figure, 'unmatched_native_pixels': missing, 'native_pixels': int(np.count_nonzero(target)),
                  'grid_phase': [ox, oy], 'native_color_match': True, 'hd_within_native_silhouette': True,
                  'hd_transparent_subpixels_within_native': uncovered,
                  'hd_base_layer': 'original-nearest-3x', 'final_hd_silhouette_match': True,
                  'whole_figure_override': whole_source,
                  'alpha_crop': list(bounds), 'parts': proof, 'assembled': images}
        print('ASSEMBLED', figure['key'], 'missing', missing, flush=True)
        return native, hd, record

    figures = [(f, *assemble(f)) for f in layout['figures']]
    assert len(figures) == 40

    def place(canvas, picture, box, pixel=False):
        x, y, w, h = map(int, box)
        im = ImageOps.contain(picture, (w, h), Image.Resampling.NEAREST if pixel else Image.Resampling.LANCZOS)
        canvas.paste(im, (x + (w - im.width) // 2, y + (h - im.height) // 2), im)

    outputs = {}
    for bosses in (False, True):
        entries = [row for row in figures if (row[0]['kind'] == 'boss') == bosses]
        assert len(entries) == (10 if bosses else 30)
        height = 2420 if bosses else 2300
        canvas = Image.new('RGB', (1600, height), '#101923')
        d = ImageDraw.Draw(canvas)
        title = '魔王與首領完整姿態' if bosses else '怪物完整姿態總覽'
        d.text((32, 20), title, font=fonts[40], fill='#f3eee1')
        d.text((32, 82), '每組左為原版素材，右為手繪 HD；依原生圖格重組完整角色。', font=fonts[24], fill='#bac5cf')
        if bosses:
            for i, (f, native, hd, _) in enumerate(entries):
                x, y = 24 + (i % 2) * 788, 144 + (i // 2) * 440
                d.rounded_rectangle((x, y, x + 764, y + 420), radius=12, fill='#1c2936')
                d.text((x + 20, y + 12), f['title'], font=fonts[30], fill='#e3b66a')
                d.text((x + 20, y + 58), '原版素材', font=fonts[20], fill='#bac5cf')
                d.text((x + 402, y + 58), '手繪 HD', font=fonts[20], fill='#bac5cf')
                place(canvas, native, (x + 12, y + 92, 360, 312), True)
                place(canvas, hd, (x + 394, y + 92, 358, 312))
        else:
            for area in range(1, 9):
                y = 144 + (area - 1) * 260
                group = [row for row in entries if row[0]['area'] == area]
                d.rounded_rectangle((24, y, 1576, y + 244), radius=12, fill='#1c2936')
                d.text((44, y + 10), f'區域 {area}', font=fonts[24], fill='#e3b66a')
                step = 1504 / len(group)
                for i, (_, native, hd, _) in enumerate(group):
                    x, half = 48 + i * step, step / 2
                    d.text((int(x), y + 46), '原版', font=fonts[18], fill='#bac5cf')
                    d.text((int(x + half), y + 46), 'HD', font=fonts[18], fill='#bac5cf')
                    place(canvas, native, (x + 4, y + 78, half - 16, 150), True)
                    place(canvas, hd, (x + half + 4, y + 78, half - 16, 150))
        d.text((32, height - 56), '素材重組示意；展示代表姿態，不代表後段實際出場或全動畫驗收。', font=fonts[20], fill='#bac5cf')
        name = 'bosses-overview.jpg' if bosses else 'enemies-overview.jpg'
        file = work / name
        canvas.save(file, quality=90, subsampling=0, optimize=True)
        outputs[name] = {'sha256': digest(file), 'bytes': file.stat().st_size,
                         'size': list(canvas.size), 'figures': [row[0]['key'] for row in entries]}

    manifest = {'schema': 2, 'scope': '使用者指定的私人 GitHub README 文件縮圖；不改遊戲及正式封包，原版及衍生美術不適用自有程式授權',
                'method': '原版 8×8 全色圖樣逐像素匹配，依 16×16 模板上下文選取原生圖格及既有 3 倍 HD 象限；原版作下層，沿正式覆繪方式混合 HD 透明邊。不複製外部參考像素、不補畫、不改來源色',
                'layout_reference_credit': layout['layout_reference_credit'], 'palette_note': layout['palette_note'],
                'pillow': Image.__version__, 'tool_sha256': digest(Path(__file__)),
                'normal_player_path_verified': False, 'all_figures_have_complete_native_coverage': True,
                'sources': sources, 'figures': [row[3] for row in figures], 'outputs': outputs}
    # 所有角色完整重組後才更新文件；暫存檔與目的檔位於同一掛載。
    for name in outputs:
        temporary = out / (name + '.tmp')
        temporary.write_bytes((work / name).read_bytes())
        temporary.replace(out / name)
    temporary = out / 'creatures-provenance.json.tmp'
    temporary.write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + '\n')
    temporary.replace(out / 'creatures-provenance.json')
    print('ASSEMBLED_COMPLETE_FIGURES 40 ZERO_UNMATCHED_NATIVE_PIXELS')


if __name__ == '__main__':
    main()

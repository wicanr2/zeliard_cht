#!/usr/bin/env python3
"""插畫HD色相閘門：逐原生色號比較HD圖面平均色與實機色盤的色相。只寫收據檔。規格068。

圖面依hd_manual_art.py的裁切方式縮放到插畫的3倍大小；每個色號取原生位置對應3×3點的平均色。
只比較飽和度足夠的色號（灰、黑、白不比色相），以點數加權。紅綠互換的圖面在主要色號上會差約120度。
另以亮度位移相關檢查對位：最佳位移不是(0,0)就不通過。另以4×4原生區塊查灰階：實機區塊是中性灰（純黑除外）而HD區塊彩度超過0.3者佔5%以上，也不通過。
"""
import argparse
import base64
import colorsys
import json
from pathlib import Path

import numpy as np
from PIL import Image


def hue(rgb):
    h, s, v = colorsys.rgb_to_hsv(*(c/255 for c in rgb))
    return h*360, s, v


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--refs', type=Path, required=True, help='hd_picture_refs.py的refs.json')
    ap.add_argument('--art', action='append', default=[], metavar='SHEET=PNG', help='要檢查的生成圖集')
    ap.add_argument('--output', type=Path, required=True, help='收據檔，須不存在')
    ap.add_argument('--max-mean', type=float, default=25.0, help='加權平均色相差上限（度）')
    ap.add_argument('--max-major', type=float, default=60.0, help='佔5%以上的色號色相差上限（度）；紅綠互換約差120度，暗綠畫成橄欖色草地約差45度')
    args = ap.parse_args()
    if args.output.exists():
        raise ValueError('收據已存在')
    refs = json.loads(args.refs.read_text())
    pictures = {r['key']: r for r in refs['pictures']}
    sheets = {s['key']: s for s in refs['sheets']}
    result = {'schema': 1, 'max_mean': args.max_mean, 'max_major': args.max_major, 'sheets': []}
    passed = True
    for spec in args.art:
        name, path = spec.split('=')
        atlas = Image.open(path).convert('RGB')
        row = {'sheet': name, 'art': path, 'size': list(atlas.size), 'pictures': []}
        want = sheets[name].get('atlas_size', [1536, 1024])
        if atlas.size[0]*want[1] != atlas.size[1]*want[0]:
            row['error'] = '圖集長寬比與參考不同'
            passed = False
        for cell in sheets[name]['cells']:
            if cell['key'] not in pictures:
                continue  # 參考圖留黑、不重畫的格
            pic = pictures[cell['key']]
            w, h = pic['width'], pic['height']
            size = sheets[name].get('atlas_size', [1536, 1024])
            box = tuple(round(v*(atlas.width/size[0] if i % 2 == 0 else atlas.height/size[1])) for i, v in enumerate(cell['box']))
            art = np.asarray(atlas.crop(box).resize((w*3, h*3), Image.Resampling.LANCZOS), np.float64).reshape(h, 3, w, 3, 3).mean(axis=(1, 3))
            native = np.frombuffer(base64.b64decode(pic['raster']), np.uint8).reshape(h, w)
            palette = np.frombuffer(base64.b64decode(pic['palette']), np.uint8).reshape(256, 3)
            total, weighted, worst, details = 0, 0.0, 0.0, []
            for k in np.unique(native):
                mask = native == k
                count = int(mask.sum())
                want = hue(palette[k])
                share = count/native.size
                got = hue(art[mask].mean(axis=0))
                if want[1] < 0.35 or want[2] < 0.25:
                    continue
                diff = abs(want[0]-got[0])
                diff = min(diff, 360-diff)
                total += count
                weighted += diff*count
                if share >= 0.05:
                    worst = max(worst, diff)
                details.append({'index': int(k), 'pixels': count, 'share': round(share, 4), 'palette_rgb': palette[k].tolist(), 'art_rgb': [round(c) for c in art[mask].mean(axis=0)], 'hue_diff': round(float(diff), 1)})
            # 對位：HD縮回原生解析度後與實機亮度做位移相關，最佳位移須為(0,0)。
            lum = lambda rgb: rgb[..., 0]*0.299+rgb[..., 1]*0.587+rgb[..., 2]*0.114
            a, b = lum(art), lum(palette[native].astype(np.float64))
            best, shift = -2.0, (0, 0)
            for dy in range(-2, 3):
                for dx in range(-2, 3):
                    x0, x1, y0, y1 = max(0, dx), w+min(0, dx), max(0, dy), h+min(0, dy)
                    u = a[y0:y1, x0:x1].ravel()
                    v = b[y0-dy:y1-dy, x0-dx:x1-dx].ravel()
                    c = float(np.corrcoef(u, v)[0, 1]) if u.std() > 0 and v.std() > 0 else 0.0
                    if c > best:
                        best, shift = c, (dx, dy)
            # 灰階：以4×4原生區塊平均色比較，抖色混合不誤判；黑色描線常被畫成陰影，純黑區塊不計。
            # 實機區塊是中性灰而HD區塊彩度超過0.3，佔區塊5%以上不通過。
            bh, bw = max(1, h//4), max(1, w//4)
            ref = palette[native].astype(np.float64)[:bh*4, :bw*4]
            rb = ref.reshape(bh, ref.shape[0]//bh, bw, ref.shape[1]//bw, 3).mean(axis=(1, 3))
            ab = art[:bh*4, :bw*4].reshape(bh, ref.shape[0]//bh, bw, ref.shape[1]//bw, 3).mean(axis=(1, 3))
            chroma = lambda x: (x.max(axis=2)-x.min(axis=2))/255
            bad = (chroma(rb) < 0.08) & (rb.max(axis=2)/255 >= 0.12) & (chroma(ab) > 0.3)
            tone = float(bad.sum()/bad.size)
            mean = weighted/total if total else 0.0
            ok = bool(mean <= args.max_mean and worst <= args.max_major) and shift == (0, 0) and tone < 0.05
            passed = passed and ok
            details.sort(key=lambda d: -d['pixels'])
            row['pictures'].append({'key': cell['key'], 'alignment': {'best_shift': list(shift), 'correlation': round(best, 3)}, 'grey_block_mismatch': round(tone, 4), 'mean_hue_diff': round(float(mean), 1), 'worst_major_hue_diff': round(float(worst), 1), 'pass': ok, 'colors': details[:12]})
            print(f"{name} {cell['key']}: mean {mean:.1f}° worst-major {worst:.1f}° shift {shift} corr {best:.3f} grey {tone:.3f} {'PASS' if ok else 'FAIL'}")
        result['sheets'].append(row)
    result['pass'] = passed
    args.output.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print('HUE_CHECK', 'PASS' if passed else 'FAIL')


if __name__ == '__main__':
    main()

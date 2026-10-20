#!/usr/bin/env python3
"""以生成基底及標記圖集直接註冊十種120x96門框，保留原版輪廓。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image, ImageDraw
from hd_mpp1_art import area_image


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--output', type=Path, required=True)
    args = ap.parse_args()
    p, out = args.project, args.output
    assert out.is_dir() and not any(out.iterdir())
    sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
    root = p / 'workplace/full-hd-v1'
    source_path = root / 'cave-door-source-v2/receipt.json'
    source = json.loads(source_path.read_text())
    generation_path = root / 'cave-door-art-v1/generation.json'
    generation = json.loads(generation_path.read_text())
    assert generation['mode'] == 'built-in imagegen'
    for row in generation['references']:
        assert sha(p / row['file']) == row['sha256']
    images = {}
    for row in generation['outputs']:
        path = p / row['file']
        assert sha(path) == row['sha256']
        im = Image.open(path).convert('RGBA')
        assert im.getchannel('A').getextrema() == (255, 255)
        images[row['key']] = im
    memory_path = root / 'cave-station-source-v1/source.memory'
    assert sha(memory_path) == source['memory_sha256']
    resource = memory_path.read_bytes()[0x11b10:0x21b10]
    bases = {key: np.asarray(area_image(images[key], 120, 96)) for key in ('filled', 'open')}
    atlas = images['markers']
    gems = [np.asarray(area_image(atlas.crop((i*atlas.width//4, 0, (i+1)*atlas.width//4, atlas.height)), 24, 24)) for i in range(4)]
    rows = []
    overview = Image.new('RGB', (520, 5*224), (22, 29, 40))
    draw = ImageDraw.Draw(overview)
    for v in source['variants']:
        mask = np.zeros((32, 40), dtype=bool)
        for y, row in enumerate(v['matrix']):
            for x, tile in enumerate(row):
                if not tile:
                    continue
                raw = resource[0x8030+(tile-1)*48:0x8030+tile*48]
                colors = []
                for at in range(0, 48, 3):
                    a, c, d = raw[at:at+3]
                    colors.extend([c>>2, (c&3)*16+(a>>4), (a&15)*4+(d>>6), d&63])
                mask[y*8:y*8+8, x*8:x*8+8] = np.array(colors).reshape(8, 8) != 0
        mask = np.repeat(np.repeat(mask, 3, axis=0), 3, axis=1)
        pixels = bases[v['kind']].copy()
        if v['marker'] != 97:
            pixels[0:24, 48:72] = gems[v['marker']-98]
        pixels[~mask] = 0
        assert np.all(pixels[:, :, 3][mask] == 255) and not np.any(pixels[~mask])
        path = out / f"door-{v['marker']}-{v['kind']}-120x96.png"
        im = Image.fromarray(pixels, 'RGBA')
        im.save(path)
        x, y = (0 if v['kind']=='filled' else 260), (v['marker']-97)*224
        draw.text((x+8, y+4), f"{v['marker']} / {v['kind']}", fill='white')
        overview.paste(im.resize((240,192), Image.Resampling.NEAREST).convert('RGB'), (x+8,y+24))
        rows.append({'marker': v['marker'], 'kind': v['kind'], 'file': path.name, 'sha256': sha(path),
                     'rgba_sha256': hashlib.sha256(pixels.tobytes()).hexdigest(), 'size': [120,96],
                     'opaque_pixels': int(np.count_nonzero(mask)), 'transparent_inside_mask': 0,
                     'outside_mask_written_pixels': 0, 'original_sha256': v['sha256']})
    overview.save(out/'doors-hd-overview.png')
    report = {'schema': 1, 'status': 'REGISTERED_TEN_CAVE_DOOR_VARIANTS', 'art': rows,
              'script_sha256': sha(Path(__file__)), 'source_receipt_sha256': sha(source_path),
              'generation_sha256': sha(generation_path),
              'policy': '生成基底直接取樣120x96；四標記各直接取樣24x24；原始非零色號輪廓裁形，不經40x32縮圖。',
              'limits': '圖面註冊不代表正式接入或正常路徑通過。'}
    (out/'manifest.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(report['status'], len(rows), 'variants')


if __name__ == '__main__':
    main()

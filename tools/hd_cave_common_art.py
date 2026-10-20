#!/usr/bin/env python3
"""將內建生成的共用平台候選直接註冊至72×24，不修改原圖或正式前端。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image

from hd_mpp1_art import area_image


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--output', type=Path, required=True)
    ap.add_argument('--material', action='store_true', help='第二版不透明材質底，由原版平台包絡裁出形狀')
    args = ap.parse_args()
    p, out = args.project, args.output
    assert out.is_dir() and not any(out.iterdir()), '須為既有空輸出目錄'
    sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
    root = p / 'workplace/full-hd-v1'
    gen_path = root / ('cave-common-art-v2/generation-v2.json' if args.material else 'cave-common-art-v1/generation-v1.json')
    gen = json.loads(gen_path.read_text())
    assert gen['mode'] == 'built-in imagegen' and not gen['production_runtime_changed']
    image_path = p / gen['output_file']
    expected = '300b88c5d3d4f4b744c4733b420fcbbfed6930d83135241cef3478b75136c266' if args.material else '47d2622cc795a3b45683da5c1269aa5985ce1839f77723987af9d0b2e0987ae8'
    assert sha(image_path) == gen['output_sha256'] == expected
    if args.material:
        for ref in gen['reference_inputs']:
            assert sha(p/ref['file']) == ref['sha256']
    else:
        assert sha(p/gen['reference_file']) == gen['reference_sha256']
    request_path = gen_path.parent / gen['request_file']
    assert sha(request_path) == gen['request_sha256']
    source_path = root / 'cave-common-sources-v2/receipt.json'
    source = json.loads(source_path.read_text())
    assert sha(source_path) == gen['source_receipt_sha256']
    assert source['status'] == 'PASS_DCHR_NORMAL_RAM_BINDING_SOURCE_ONLY'
    assert sum(row['pixels'] for row in source['platform_pixel_proof']) == 192
    image = Image.open(image_path).convert('RGBA')
    if args.material:
        assert image.size == (1536, 1024) and image.getchannel('A').getextrema() == (255, 255)
        crop = (0, 0, 1536, 1024)
    else:
        assert image.size == (1254, 1254)
        # 第一版以目視主體框裁切；保留舊候選，不當本版不透明材質。
        crop = (128, 514, 1130, 752)
    registered = area_image(image.crop(crop), 72, 24)
    native = Image.open(source_path.parent / 'green-platform-original.png').convert('RGB')
    assert native.size == (24, 8)
    occupied = np.repeat(np.repeat(np.any(np.asarray(native) != 0, axis=2), 3, axis=0), 3, axis=1)
    pixels = np.asarray(registered).copy()
    pixels[~occupied] = 0
    registered = Image.fromarray(pixels, 'RGBA')
    registered.save(out / 'green-platform-72x24.png')
    rows = []
    for i, tile_id in enumerate((64, 65, 66)):
        tile = registered.crop((i*24, 0, (i+1)*24, 24))
        path = out / f'tile-{tile_id:03d}.png'
        tile.save(path)
        source_tile = next(row for row in source['tiles'] if row['id'] == tile_id)
        rows.append({'id': tile_id, 'file': path.name, 'sha256': sha(path), 'size': [24, 24],
                     'rgba_sha256': hashlib.sha256(tile.tobytes()).hexdigest(),
                     'source_packed_sha256': source_tile['packed_sha256']})
    # 使用copy而非alpha-composite重接，半透明像素的RGB不得被預乘改寫。
    joined = Image.new('RGBA', (72, 24))
    for i, row in enumerate(rows):
        joined.paste(Image.open(out/row['file']), (i*24, 0))
    assert joined.tobytes() == registered.tobytes()
    assert not np.any(pixels[~occupied])
    if args.material:
        assert np.all(pixels[:, :, 3][occupied] == 255)
    report = {'schema': 1, 'status': 'REGISTERED_GREEN_PLATFORM_CANDIDATE_NOT_CONNECTED',
              'generation_file': str(gen_path.relative_to(p)), 'generation_sha256': sha(gen_path),
              'source_file': str(source_path.relative_to(p)), 'source_sha256': sha(source_path),
              'script_sha256': sha(Path(__file__)), 'source_crop': list(crop), 'target_size': [72, 24],
              'art_file': 'green-platform-72x24.png', 'art_sha256': sha(out/'green-platform-72x24.png'),
              'tiles': rows, 'native_nonzero_mask_pixels': int(np.count_nonzero(occupied)),
              'transparent_inside_mask': int(np.count_nonzero((pixels[:, :, 3] == 0) & occupied)),
              'outside_mask_written_pixels': 0, 'direct_target_resolution': True,
              'registration_policy': ('不透明材質全圖直接面積取樣到72x24，再依本平台原版非零包絡裁切；包絡內全不透明，不經24x8縮圖。' if args.material else '主體固定裁切直接面積取樣到72x24，再依本平台原版非零色號包絡裁切；不經24x8縮圖、不改碰撞或來源。'),
              'production_runtime_changed': False,
              'limits': ['此為可重生的註冊候選，正式來源守衛、平台移動完整原版配對及GUI仍待驗。',
                         '半透明輪廓與原生包絡的差異須在實際遊戲畫面檢視，不把註冊通過當HD完成。']}
    (out/'manifest.json').write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')
    print(report['status'], report['target_size'], 'transparent inside mask', report['transparent_inside_mask'])


if __name__ == '__main__':
    main()

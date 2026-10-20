#!/usr/bin/env python3
"""按八份後續城鎮原始地圖準備內建imagegen參考與完整提示詞，不接正式前端。"""
import argparse
import hashlib
import json
import math
from pathlib import Path

from PIL import Image


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--output', type=Path, required=True)
    args = ap.parse_args()
    project, out = args.project, args.output
    assert out.is_dir() and not any(out.iterdir())
    sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
    study = project / 'workplace/full-hd-v1'
    source_dir = study / 'town-resources-source-v1'
    receipt_path = source_dir / 'receipt.json'
    source = json.loads(receipt_path.read_text())
    assert source['script_sha256'] == sha(project / 'tools/hd_town_sources.py')
    style = project / 'workplace/town-hd-v1/foreground-v1/town-left-art-v2.png'
    jobs = []
    for row in source['maps'][2:]:
        key = row['key']
        im = Image.open(source_dir / (key + '-foreground-static.png')).convert('RGBA')
        assert im.size == (row['width']*8, 64)
        assert hashlib.sha256(im.tobytes()).hexdigest() == row['world_rgba_sha256']
        count = math.ceil(im.width/224)
        rows = math.ceil(count/2)
        height = 1024 if rows <= 4 else 1536
        atlas = Image.new('RGBA', (1536, height))
        panels = []
        for index in range(count):
            start = index*224
            stop = min(im.width, start+240)
            x, y = index%2*768+24, index//2*256+32
            panel = im.crop((start, 0, stop, 64)).resize(((stop-start)*3, 192), Image.Resampling.NEAREST)
            atlas.paste(panel, (x, y))
            panels.append({'index': index, 'world_native': [start, 0, stop, 64],
                           'reference_box': [x, y, x+panel.width, y+192]})
        path = out / (key + '-atlas-reference.png')
        atlas.save(path)
        unknown = ''
        if row['unresolved_tile_ids']:
            unknown = ('The thin magenta vertical band in the final strip is an UNRESOLVED source marker, not scenery. '
                       'Leave that band fully transparent. Do not turn it into a building, wall, door or decorative element. ')
        prompt = (
            'Use case: style-transfer. Asset type: hand-painted fantasy platform-game foreground atlas. '
            f'Edit target image1 is the original {row["name_ascii"]} foreground atlas, {1536} by {height} pixels. '
            'Image2 is STYLE REFERENCE ONLY from this same project: its painterly texture, crisp game-readable forms, '
            'material depth and restrained lighting. Do not copy its buildings or change the subjects in image1. '
            f'Repaint the {count} occupied horizontal strips in image1 into the same B hand-painted fantasy style. '
            'The atlas has TWO COLUMNS of independent strips, read left-to-right then top-to-bottom. '
            'Each cell is 768x256. The scene starts 24 pixels from the cell left and 32 pixels from its top, '
            'and its exact foreground height is 192 pixels. Preserve the reference canvas size and exact cell layout. '
            'Keep every building, tree, fence, post, torch, arch, doorway and cave edge at the SAME position and width. '
            'Keep their original silhouettes, heights, opening widths, roof lines and ground baseline. '
            'These strips are adjoining slices of ONE scrolling town, with 16 native pixels of overlap. '
            'Repeated details and overlapping buildings must match across neighboring strips. '
            'Maintain the original palette identities: warm logs and green pines where present, and blue, green or purple '
            'stone/brick with warm torchlight where present. Paint detailed stone blocks, wood grain, foliage and flame light '
            'inside the existing shapes, with clear side-view orthographic geometry and no perspective camera. '
            'Preserve the original shop-sign words exactly where visible: MAGIC, INN, BANK, WEAPON. '
            'Do not add other words or labels. Keep black door interiors dark and open, never fill an opening with a door or wall. '
            + unknown +
            'The areas outside source silhouettes, gaps between cells and unused cells must be genuinely transparent alpha. '
            'Do not add a sky, cave-wall backdrop, ground extension, characters, enemies, props, captions, framing or watermark. '
            'This is a direct high-resolution repaint, not pixel art and not an illustrated scene redesign.'
        )
        jobs.append({'key': key, 'town_name_source': row['name_ascii'], 'source_record': row['record'],
                     'terrain_key': row['terrain_key'], 'target_reference_size': [1536, height], 'panels': panels,
                     'unresolved_tile_ids': row['unresolved_tile_ids'], 'prompt': prompt,
                     'reference_inputs': [{'file': str(path.relative_to(project)), 'sha256': sha(path), 'role': 'edit target; exact geometry'},
                                          {'file': str(style.relative_to(project)), 'sha256': sha(style), 'role': 'style only'}]})
    report = {'schema': 1, 'mode': 'built-in imagegen', 'status': 'PREPARED_EIGHT_TOWN_ART_CANDIDATES',
              'script_sha256': sha(Path(__file__)), 'source_receipt_sha256': sha(receipt_path), 'jobs': jobs,
              'limits': ['只準備八份美術候選；正常來源、動畫、背景、下層地面、居民與前端接入仍待驗。',
                         'Pureza原版圖格191未核，生成圖不能補出其用途或算已完成。']}
    (out / 'generation-prepared.json').write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')
    print(report['status'], [(job['key'], len(job['panels'])) for job in jobs])


if __name__ == '__main__':
    main()

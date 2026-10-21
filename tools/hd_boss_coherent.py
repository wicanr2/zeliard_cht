#!/usr/bin/env python3
"""058：由完整首領圖裁出既有原生模板，原圖及舊登錄保留本機。"""
import argparse
import base64
import hashlib
import json
import os
import shutil
from pathlib import Path

import numpy as np
from PIL import Image


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    args = ap.parse_args()
    root = args.project
    work = root / 'workplace/hd-boss-coherent-v1'
    target = root / 'workplace/full-hd-v1/manual-hd-v1/registered'
    assert work.is_dir() and target.is_dir()
    assert work.stat().st_uid == target.stat().st_uid == os.getuid()
    snapshots = work / 'input-snapshot'
    snapshots.mkdir(exist_ok=True)
    for name in ('manual-art.json', 'manual-tiles.json'):
        assert (target / name).stat().st_uid == os.getuid()
        if not (snapshots / name).exists():
            shutil.copy2(target / name, snapshots / name)
    if not (snapshots / 'creatures-provenance.json').exists():
        shutil.copy2(root / 'docs/images/creatures-provenance.json', snapshots / 'creatures-provenance.json')
    catalog = json.loads((snapshots / 'manual-art.json').read_text())
    tiles = json.loads((snapshots / 'manual-tiles.json').read_text())
    proven = json.loads((snapshots / 'creatures-provenance.json').read_text())
    byfile = {o['file']: o for o in catalog['objects']}
    palette = np.frombuffer((root / 'workplace/music-release-v1/release-build/AppDir/assets/town/frame/palette.bin').read_bytes(), np.uint8).reshape(256, 3)
    selected = {'tako': 1, 'tori': 1, 'zela': 1, 'meda': 1, 'mao1': 2}
    masters, whole, reports = {}, [], []
    for bank, version in selected.items():
        figure = next(f for f in proven['figures'] if f['key'] == bank)
        w = (figure['box'][2] - figure['box'][0]) // 2
        h = (figure['box'][3] - figure['box'][1]) // 2
        native = np.zeros((h, w), np.uint8)
        for part in figure['parts']:
            obj = byfile[part['registered_file']]
            raw = np.frombuffer(base64.b64decode(obj['raster']), np.uint8).reshape(16, 16)
            q = part['quad']
            x, y = part['native_position']
            native[y:y + 8, x:x + 8] = raw[q // 2 * 8:q // 2 * 8 + 8, q % 2 * 8:q % 2 * 8 + 8]
        bounds = tuple(figure['alpha_crop'])
        x0, y0, x1, y1 = bounds
        cropped = native[y0:y1, x0:x1]
        ref = work / 'references' / (bank + '-original.png')
        assert np.array_equal(np.array(Image.open(ref).convert('RGB')), palette[cropped])
        source = work / 'generated' / f'{bank}-whole-v{version}.png'
        original = Image.open(source).convert('RGBA')
        a = original.getchannel('A').point(lambda v: 255 if v >= 32 else 0)
        box = a.getbbox()
        assert box is not None
        painted = original.crop(box).resize(((x1 - x0) * 3, (y1 - y0) * 3), Image.Resampling.LANCZOS)
        mask = Image.fromarray(np.where(cropped != 0, 255, 0).astype(np.uint8)).resize(painted.size, Image.Resampling.NEAREST)
        rgba = np.dstack((palette[cropped], np.where(cropped != 0, 255, 0).astype(np.uint8)))
        base = Image.fromarray(rgba).resize(painted.size, Image.Resampling.NEAREST)
        painted = Image.alpha_composite(base, painted)
        painted.putalpha(mask)
        file = f'coherent-{bank}-whole-v1.png'
        painted.save(target / file)
        canvas = Image.new('RGBA', (w * 3, h * 3))
        canvas.paste(painted, (x0 * 3, y0 * 3))
        masters[bank] = native, canvas
        entry = {'key': bank, 'width': x1 - x0, 'height': y1 - y0,
                 'raster': base64.b64encode(cropped.tobytes()).decode(),
                 'colors': base64.b64encode(palette.tobytes()).decode(),
                 'file': file, 'sha256': sha(target / file), 'source_art': str(source.relative_to(root)),
                 'source_art_sha256': sha(source), 'source_crop': list(box), 'native_bounds': list(bounds)}
        whole.append(entry)
        reports.append({'bank': bank, 'source_sha256': sha(source), 'whole_sha256': entry['sha256'], 'objects': 0, 'native_fallback_quads': 0})
    # 只取同一完整圖的原生象限；每個位置保留來源可回查。
    for obj in catalog['objects']:
        bank = obj.get('bank')
        if bank not in masters:
            continue
        native, painted = masters[bank]
        raw = np.frombuffer(base64.b64decode(obj['raster']), np.uint8).reshape(16, 16)
        rgba = np.dstack((palette[raw], np.where(raw != 0, 255, 0).astype(np.uint8)))
        result = Image.fromarray(rgba).resize((48, 48), Image.Resampling.NEAREST)
        parts = []
        report = next(r for r in reports if r['bank'] == bank)
        report['objects'] += 1
        for q in range(4):
            tile = raw[q // 2 * 8:q // 2 * 8 + 8, q % 2 * 8:q % 2 * 8 + 8]
            if not np.any(tile):
                continue
            choices = []
            for y in range(0, native.shape[0] - 7, 8):
                for x in range(0, native.shape[1] - 7, 8):
                    if not np.array_equal(native[y:y + 8, x:x + 8], tile):
                        continue
                    left, top = x - q % 2 * 8, y - q // 2 * 8
                    view = native[max(0, top):top + 16, max(0, left):left + 16]
                    score = int(np.count_nonzero(view == raw)) if view.shape == raw.shape else 0
                    choices.append((score, x, y))
            if not choices:
                report['native_fallback_quads'] += 1
                continue
            _, x, y = max(choices, key=lambda c: c[0])
            art = painted.crop((x * 3, y * 3, (x + 8) * 3, (y + 8) * 3))
            result.paste(art, (q % 2 * 24, q // 2 * 24))
            parts.append({'quad': q, 'whole_native_position': [x, y]})
        file = 'coherent-' + obj['file']
        result.save(target / file)
        obj.update({'file': file, 'sha256': sha(target / file), 'coherent_bank': bank, 'coherent_parts': parts,
                    'source_art': str((work / 'generated' / f'{bank}-whole-v{selected[bank]}.png').relative_to(root)),
                    'source_art_sha256': sha(work / 'generated' / f'{bank}-whole-v{selected[bank]}.png'),
                    'source_crop': None})
    tiles['tiles'] = [t for t in tiles['tiles'] if t.get('category') not in selected]
    for bank, (native, painted) in masters.items():
        seen = set()
        for y in range(0, native.shape[0] - 7, 8):
            for x in range(0, native.shape[1] - 7, 8):
                tile = native[y:y + 8, x:x + 8]
                if not np.any(tile) or tile.tobytes() in seen:
                    continue
                seen.add(tile.tobytes())
                key = hashlib.sha256(tile.tobytes()).hexdigest()[:16]
                file = f'coherent-{bank}-tile-{key}.png'
                painted.crop((x * 3, y * 3, (x + 8) * 3, (y + 8) * 3)).save(target / file)
                tiles['tiles'].append({'raster': base64.b64encode(tile.tobytes()).decode(),
                                       'rgb': base64.b64encode(palette[tile].tobytes()).decode(),
                                       'file': file, 'sha256': sha(target / file), 'category': bank})
    catalog['coherent_revision'] = '058-complete-figure-first'
    write(target / 'manual-art.json', catalog)
    write(target / 'manual-tiles.json', tiles)
    manifest = {'schema': 1, 'scale': 3, 'figures': whole, 'reports': reports,
                'method': '完整圖先生成，再依原生圖格裁切；未知象限保留原版，不引用舊錯誤圖',
                'normal_boss_gameplay_verified': False, 'tool_sha256': sha(Path(__file__))}
    write(target / 'coherent-bosses.json', manifest)
    write(work / 'registration.json', manifest)
    print(json.dumps(reports, ensure_ascii=False))


if __name__ == '__main__':
    main()

#!/usr/bin/env python3
"""057：登錄正常原版入洞背景與觀察到的角色圖樣，不改原始來源。"""
import argparse
import base64
import hashlib
import json
import os
from pathlib import Path

import numpy as np
from PIL import Image


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    args = ap.parse_args()
    root = args.project
    work = root / 'workplace/hd-cave-transition-v1'
    target = root / 'workplace/full-hd-v1/manual-hd-v1/registered'
    assert work.is_dir() and target.is_dir() and target.stat().st_uid == os.getuid()
    sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
    data = (work / 'transition-background.indexed').read_bytes()
    assert hashlib.sha256(data).hexdigest() == '054e31ebab40662de49d9023c6d3b3561e5a4dfb88a8ba25481f8cbc56c74329'
    background = np.frombuffer(data, np.uint8).reshape(144, 224)
    palette = (work / 'oracle/frame-0035.palette').read_bytes()
    inputs = []
    receipt = json.loads((work / 'oracle-dense/receipt.json').read_text())
    poses = {}
    matched = []
    for shot in receipt['shots']:
        path = work / 'oracle-dense' / (shot['file'] + '.indexed')
        raw = np.frombuffer(path.read_bytes(), np.uint8).reshape(200, 320)[14:158, 48:272]
        difference = (raw != background) & (raw != 0)
        y, x = np.where(difference)
        ink = np.count_nonzero((raw == background) & (background != 0))
        if ink < 1000:
            continue
        if len(x):
            x0, y0, x1, y1 = map(int, (x.min(), y.min(), x.max() + 1, y.max() + 1))
            if y0 < 96 or y1 > 120 or x1 - x0 > 24 or y1 - y0 > 24 or np.any(background[difference] != 0):
                continue
            actor = raw[y0:y1, x0:x1]
            key = (x1 - x0, y1 - y0, actor.tobytes())
            poses[key] = {'width': x1 - x0, 'height': y1 - y0, 'raster': base64.b64encode(actor.tobytes()).decode(), 'source_frame': shot['file']}
        matched.append(shot['file'])
        inputs.append({'file': str(path.relative_to(root)), 'sha256': sha(path), 'memory_sha256': shot['memory_sha256'], 'steps': shot['steps']})
    assert poses and len(matched) > 100
    components = json.loads((target / 'complete-art.json').read_text())['components']
    index = {}
    for component in components:
        if component['source'] == 'fman-all':
            raw = bytes(0 if v == 255 else v for v in base64.b64decode(component['raster']))
            index.setdefault(raw, component)
    actor_hd = 0
    colors = np.frombuffer(palette, np.uint8).reshape(256, 3)
    for n, pose in enumerate(poses.values()):
        raw = np.frombuffer(base64.b64decode(pose['raster']), np.uint8).reshape(pose['height'], pose['width'])
        selection = None
        best_missing = None
        for oy in range(25 - pose['height']):
            for ox in range(25 - pose['width']):
                padded = np.zeros((24, 24), np.uint8)
                padded[oy:oy + pose['height'], ox:ox + pose['width']] = raw
                parts, missing = [], 0
                for y in range(0, 24, 8):
                    for x in range(0, 24, 8):
                        tile = padded[y:y + 8, x:x + 8]
                        if not np.any(tile):
                            continue
                        component = index.get(tile.tobytes())
                        if component is None:
                            missing += int(np.count_nonzero(tile))
                        else:
                            parts.append((x, y, component))
                if best_missing is None or missing < best_missing:
                    best_missing = missing
                    selection = ox, oy, parts
        if selection is None or not selection[2]:
            pose['hd_matched'] = False
            continue
        ox, oy, parts = selection
        canvas = Image.new('RGBA', (72, 72))
        records = []
        for x, y, component in parts:
            path = target / component['file']
            assert sha(path) == component['sha256']
            canvas.paste(Image.open(path).convert('RGBA'), (x * 3, y * 3))
            records.append({'native_position': [x - ox, y - oy], 'file': component['file'], 'sha256': component['sha256']})
        image = canvas.crop((ox * 3, oy * 3, (ox + pose['width']) * 3, (oy + pose['height']) * 3))
        rgba = np.dstack((colors[raw], np.where(raw != 0, 255, 0).astype(np.uint8)))
        base = Image.fromarray(rgba).resize(image.size, Image.Resampling.NEAREST)
        image = Image.alpha_composite(base, image)
        image.putalpha(base.getchannel('A'))
        file = f'cave-transition-actor-{n:03d}.png'
        image.save(target / file)
        pose.update({'hd_matched': True, 'file': file, 'sha256': sha(target / file), 'parts': records,
                     'unmatched_native_pixels_kept_original': best_missing})
        actor_hd += 1
    generated = work / 'transition-background-generated-v1.png'
    image = Image.open(generated).convert('RGBA').resize((672, 432), Image.Resampling.LANCZOS)
    rgba = np.array(image)
    black = np.repeat(np.repeat(background == 0, 3, axis=0), 3, axis=1)
    rgba[black, :3] = 0
    rgba[:, :, 3] = 255
    file = 'cave-transition-HD-v1.png'
    Image.fromarray(rgba).save(target / file)
    manifest = {'schema': 1, 'scale': 3, 'width': 224, 'height': 144, 'file': file, 'sha256': sha(target / file),
                'raster': base64.b64encode(data).decode(), 'colors': base64.b64encode(palette).decode(),
                'poses': list(poses.values()), 'matched_source_frames': matched,
                'generated_source': {'file': str(generated.relative_to(root)), 'sha256': sha(generated)},
                'inputs': inputs, 'oracle_receipt_sha256': sha(work / 'oracle-dense/receipt.json'),
                'tool_sha256': sha(Path(__file__)), 'scope': '正常入洞來源；非其他城鎮、反向或完整遊戲驗收'}
    manifest['anchors'] = [[0,0,32,8],[192,0,224,8]]
    encoded = json.dumps(manifest, ensure_ascii=False, indent=2) + '\n'
    (target / 'cave-transition.json').write_text(encoded)
    (work / 'registration.json').write_text(encoded)
    print('REGISTERED_TRANSITION', len(matched), 'observed actor patterns', len(poses), 'HD actor patterns', actor_hd)


if __name__ == '__main__':
    main()

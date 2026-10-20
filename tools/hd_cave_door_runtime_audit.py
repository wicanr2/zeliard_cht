#!/usr/bin/env python3
"""核門框十圖、來源守衛、正常配對及實際視窗；不外推其他變體已正常到達。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--output', type=Path, required=True)
    args = ap.parse_args()
    p, b = args.project, args.project/'workplace/full-hd-v1'
    assert not args.output.exists(), '不得覆寫歷史稽核'
    sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
    read = lambda path: json.loads(path.read_text())
    source_path = b/'cave-door-source-v2/receipt.json'
    source = read(source_path)
    assert source['status'] == 'BOUND_UNTITLED_CAVE_DOOR_SOURCE'
    assert len(source['placements']) == 163 and len(source['variants']) == 10
    assert source['script_sha256'] == sha(p/'tools/hd_cave_door_sources.py')
    replay_path = b/'cave-station-source-v1/receipt.json'
    replay = read(replay_path)
    assert replay['status'] == 'PASS_GUI_JOURNAL_SOURCE_RECONSTRUCTION' and not replay['terrain_hd_matched']
    assert replay['records'] == 4614 and replay['save_unchanged']
    generation_path = b/'cave-door-art-v1/generation.json'
    generation = read(generation_path)
    assert generation['mode'] == 'built-in imagegen' and len(generation['outputs']) == 3
    assert len(generation['prompts']) == 3
    for row in generation['outputs'] + generation['references']:
        assert sha(p/row['file']) == row['sha256']
    reg_path = b/'cave-door-complete-v1/manifest.json'
    reg = read(reg_path)
    assert reg['script_sha256'] == sha(p/'tools/hd_cave_doors_art.py')
    assert reg['source_receipt_sha256'] == sha(source_path) and reg['generation_sha256'] == sha(generation_path)
    assert len(reg['art']) == 10
    atlases = {}
    for row in reg['art']:
        file = reg_path.parent/row['file']
        assert sha(file) == row['sha256'] and row['size'] == [120,96]
        assert row['transparent_inside_mask'] == row['outside_mask_written_pixels'] == 0
        im = np.asarray(Image.open(file).convert('RGBA'))
        assert im.shape == (96,120,4) and set(np.unique(im[:,:,3])) <= {0,255}
        assert hashlib.sha256(im.tobytes()).hexdigest() == row['rgba_sha256']
        atlases[(row['marker'], row['kind']=='open')] = im
    runtime = b/'cave-door-runtime-v1'
    counts = {}
    for file, expected in [(runtime/'regression-tests.jsonl',34),(b/'cave-door-normal-pair-v1/tests.jsonl',3)]:
        rows = [json.loads(line) for line in file.read_text().splitlines()]
        assert rows[-1]['Action']=='pass' and not any(r['Action'] in ('fail','skip') for r in rows)
        count = sum(r['Action']=='pass' and 'Test' in r and '/' not in r['Test'] for r in rows)
        assert count == expected, (file, count)
        counts[str(file.relative_to(p))] = count
    assert '獨立像素2304，改變1008' in (runtime/'regression-tests.jsonl').read_text()
    assert '獨立結構像素144000' in (runtime/'regression-tests.jsonl').read_text()
    assert not (runtime/'vet.log').read_bytes() and not (runtime/'build.log').read_bytes()
    binary = runtime/'zeliard-preview'
    assert binary.read_bytes()[:4] == b'\x7fELF'
    pairs = []
    for name, count in [('cave-platform',10),('cave-platform-magia',3)]:
        file = b/'cave-door-normal-pair-v1'/name/'receipt.json'
        data = read(file)
        assert data['state_equal'] and data['save_unchanged'] and not data['wrote'] and not data['unimplemented']
        assert len(data['points']) == data['hd_draws'] == count
        assert all(row['state_equal'] and row['terrain_hd_drawn'] and row['presentation']['used'] for row in data['points'])
        pairs.append({'path': str(file.relative_to(p)), 'points': count, 'captured': data['captured'], 'refused': data['refused']})
    gui_path = b/'cave-door-gui-v1/receipt.json'
    gui = read(gui_path)
    assert gui['status']=='PASS_CAVE_DOOR_GUI_SCOPED' and gui['doors_enabled'] and gui['magia_enabled']
    assert gui['binary_sha256']==sha(binary) and gui['normal_DOS_exit']==0 and gui['save_unchanged']
    assert gui['script_sha256']==sha(p/'tools/hd_cave_platform_smoke.py')
    assert gui['catalog_helper_sha256']==sha(p/'tools/hd_cave_catalog_smoke.py')
    shots = {row['name']: row for row in gui['shots']}
    assert len(gui['pairs']) == 5
    return_pixels, door_pixels = 0, 0
    for pair in gui['pairs']:
        images, states = [], []
        for suffix in ['original','hd','returned']:
            row = shots[pair['name']+'-'+suffix]
            file = gui_path.parent/'boot1'/(row['name']+'.png')
            assert sha(file)==row['image_sha256']
            im = np.asarray(Image.open(file).convert('RGB'))
            assert im.shape == (680,960,3)
            images.append(im)
            states.append(row['state'])
        for key in ['steps','cycles','oracle_memory_sha256','oracle_indexed_sha256','oracle_palette_sha256']:
            assert states[0][key] == states[1][key] == states[2][key]
        assert np.array_equal(images[0][48:648], images[2][48:648])
        return_pixels += 960*600
        assert states[1]['cave_terrain_hd_matched'] and states[1]['cave_magia_hd']['matched']
        if pair['name'] == 'door-after-item':
            doors = states[1]['cave_door_hd']
            assert doors['matched'] and len(doors['placements']) == 1
            place = doors['placements'][0]
            assert place['Marker']==98 and place['Open']
            x, y = place['Position']['X']*3, place['Position']['Y']*3
            left, top, right, bottom = max(x,144), max(y,42), min(x+120,816), min(y+96,474)
            assert right>left and bottom>top
            atlas = atlases[(98,True)][top-y:bottom-y,left-x:right-x]
            expected = (atlas[:,:,:3].astype(np.uint32)*atlas[:,:,3,None]+127)//255
            actual = images[1][48+top:48+bottom,left:right]
            assert np.array_equal(actual, expected.astype(np.uint8)), '實際門框像素不符'
            door_pixels += actual.shape[0]*actual.shape[1]
    assert door_pixels >= 2304
    user_save = gui_path.parent/'user/game/SAVETEST.usr'
    original_save = p/'workplace/orig/SAVEM1.USR'
    assert sha(user_save)==sha(original_save) and user_save.stat().st_mtime_ns==original_save.stat().st_mtime_ns
    files = ['cmd/zeliard-preview/main.go', 'internal/prototype/hd/cave_doors.go', 'internal/prototype/hd/cave_doors_test.go',
             'internal/prototype/hd/cave_terrain.go', 'internal/prototype/hd/cave_remaining_families_test.go',
             'tools/hd_cave_door_sources.py', 'tools/hd_cave_doors_art.py', 'tools/hd_cave_door_runtime_audit.py',
             'tools/hd_cave_platform_smoke.py', 'tools/hd_cave_catalog_smoke.py', 'tools/hd_catalog.py']
    proofs = [source_path, replay_path, generation_path, reg_path, gui_path, binary, runtime/'regression-tests.jsonl',
              runtime/'vet.log', runtime/'build.log', b/'cave-door-normal-pair-v1/tests.jsonl'] + [p/row['path'] for row in pairs]
    report = {'schema':1, 'status':'PASS_CAVE_DOOR_SOURCE_ART_NORMAL_GUI_SCOPED', 'spec':'docs/spec/045-cave-untitled-doors.md',
              'binary_sha256':sha(binary), 'registered_variants':10, 'normal_variant':'98/open',
              'normal_state_equal_points':13, 'normal_pairs':pairs, 'independent_source_hd_pixels':2304,
              'structural_hd_pixels':144000, 'gui_original_return_pixels':return_pixels, 'gui_door_pixels':door_pixels,
              'test_counts':counts, 'source_files':[{'path':file,'sha256':sha(p/file)}for file in files],
              'proofs':[{'path':str(file.relative_to(p)), 'sha256':sha(file)}for file in proofs],
              'limits':['其他九種門框只有原資料／結構驗證，不當正常到達。',
                        '既有未匹配主角／武器、其他共用物件及後段仍待補；不宣稱全遊戲HD完成。',
                        '封包、版號與影片未更新。']}
    args.output.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(report['status'], return_pixels, 'original pixels', door_pixels, 'GUI door pixels')


if __name__ == '__main__':
    main()

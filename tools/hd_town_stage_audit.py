#!/usr/bin/env python3
"""核後續城鎮來源與八份美術候選，正常接入未驗時維持待驗。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--output', type=Path, required=True)
    args = ap.parse_args()
    p = args.project
    root = p / 'workplace/full-hd-v1'
    sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
    read = lambda path: json.loads(path.read_text())
    source_path = root / 'town-resources-source-v1/receipt.json'
    source = read(source_path)
    assert source['script_sha256'] == sha(p / 'tools/hd_town_sources.py')
    assert source['total_cells'] == 17784 and len(source['maps']) == 10
    assert [row['record'] for row in source['maps']] == list(range(36, 46))
    unresolved = {row['key']: row['unresolved_tile_ids'] for row in source['maps'] if row['unresolved_tile_ids']}
    assert unresolved == {'PRMP': {'191': 8}}
    for row in source['maps']:
        assert sha(source_path.parent / (row['key'] + '.source')) == row['payload_sha256']
        assert sha(source_path.parent / (row['key'] + '.columns')) == row['map_ids_sha256']
        im = Image.open(source_path.parent / (row['key'] + '-foreground-static.png')).convert('RGBA')
        assert im.size == (row['width']*8, 64)
        assert hashlib.sha256(im.tobytes()).hexdigest() == row['world_rgba_sha256']
    native = read(root / 'grp-sources-v3/receipt.json')
    for family in source['families'][:2]:
        bound = next(row for row in native['records'] if row['record'] == family['record'])
        assert family['packed_sha256'] == bound['pixels_sha256']
    ida_path = root / 'town-resources-ida-v2/evidence.json'
    ida = read(ida_path)
    assert ida['tool'] == 'IDA 9.4' and ida['input_sha256'] == source['code_sha256']
    assert ida['source_index_sha256'] == sha(p / 'docs/town-resources-source-index.json')
    assert ida['script_sha256'] == sha(p / 'tools/ida_town_resources.py')
    assert len(ida['functions']) == 9 and len(ida['tables']) == 5
    generation_path = root / 'town-world-art-v1/generation-v1.json'
    generation = read(generation_path)
    assert generation['mode'] == 'built-in imagegen'
    assert generation['source_receipt_sha256'] == sha(source_path)
    assert generation['script_sha256'] == sha(p / 'tools/hd_town_art.py')
    assert generation['prepared_sha256'] == sha(generation_path.parent / 'generation-prepared.json')
    assert len(generation['jobs']) == len(generation['outputs']) == 8
    for job in generation['jobs']:
        assert job['prompt'] and len(job['reference_inputs']) == 2
        for ref in job['reference_inputs']:
            assert sha(p / ref['file']) == ref['sha256']
    for row in generation['outputs']:
        assert sha(p / row['file']) == row['sha256']
        im = Image.open(p / row['file'])
        assert im.mode == 'RGBA' and list(im.size) == row['size']
        assert im.getchannel('A').getextrema()[0] == 0
    assert {row['key'] for row in generation['outputs']} == {row['key'] for row in source['maps'][2:]}
    probes = []
    for name in ('town-path-normal-v1', 'town-path-normal-v2', 'town-path-normal-v3', 'town-path-normal-v4'):
        path = root / name / 'receipt.json'
        receipt = read(path)
        assert receipt['script_sha256'] == sha(path.parent / 'probe-source.go')
        assert receipt['script_sha256'] == sha(p / 'tools/hd_town_path_probe.go')
        assert receipt['source_save_unchanged'] and receipt['scratch_save_unchanged'] and receipt['save_mtime_unchanged']
        assert not receipt['state_injection'] and not receipt['snapshot_restore']
        for point in receipt['points']:
            for ext, field in (('memory', 'memory_sha256'), ('indexed', 'indexed_sha256'), ('palette', 'palette_sha256')):
                assert sha(path.parent / (point['file'] + '.' + ext)) == point[field]
        reached = any(point['label'].startswith('town-') and point['label'].endswith('-complete') for point in receipt['points'])
        assert not reached
        probes.append({'path': str(path.relative_to(p)), 'sha256': sha(path), 'points': len(receipt['points']), 'new_town_reached': False})
    files = ['tools/hd_town_sources.py', 'tools/ida_town_resources.py', 'tools/hd_town_path_probe.go', 'tools/hd_town_art.py',
             'tools/hd_town_stage_audit.py', 'tools/hd_catalog.py', 'docs/town-resources-source-index.json']
    report = {'schema': 1, 'status': 'SOURCE_AND_EIGHT_ART_CANDIDATES_NORMAL_PATH_PENDING', 'town_maps': 10,
              'map_cells': 17784, 'additional_towns': 8, 'generated_candidates': 8, 'unresolved': unresolved,
              'normal_new_town_reached': False, 'production_runtime_changed': False,
              'source_files': [{'path': name, 'sha256': sha(p/name)} for name in files],
              'proofs': [{'path': str(path.relative_to(p)), 'sha256': sha(path)} for path in (source_path, ida_path, generation_path)],
              'normal_probes': probes,
              'limits': ['八份生成圖均是美術候選，沒有正常新城鎮來源、前端接入或GUI驗收。',
                         '接縫、門洞、招牌與透明邊緣仍須校準；Pureza未知標記仍存在，不拿生成圖補原版用途。',
                         'DPAT只核靜態解碼及原版選擇端，正常載入與原生像素待驗；原型不進正式路徑。',
                         '全遊戲HD目標保持進行中，未更新封包、影片或README縮圖。']}
    assert not args.output.exists()
    args.output.write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')
    print(report['status'], report['map_cells'], 'cells', report['generated_candidates'], 'art candidates')


if __name__ == '__main__':
    main()

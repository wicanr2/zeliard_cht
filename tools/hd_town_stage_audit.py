#!/usr/bin/env python3
"""核後續城鎮與共用洞窟來源及美術候選，正常接入未驗時維持待驗。"""
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
    correction_path = root / 'town-world-art-v1/ESMP-correction-v2.receipt.json'
    correction = read(correction_path)
    assert correction['mode'] == 'built-in imagegen' and not correction['production_runtime_changed']
    request_path = correction_path.parent / correction['request_file']
    assert sha(request_path) == correction['request_sha256']
    request = read(request_path)
    assert request['prompt'] and len(request['reference_inputs']) == 2
    for ref in request['reference_inputs']:
        assert sha(p/ref['file']) == ref['sha256']
    corrected_image = correction_path.parent / correction['output_file']
    assert sha(corrected_image) == correction['output_sha256']
    im = Image.open(corrected_image)
    assert im.mode == 'RGBA' and list(im.size) == correction['size'] == [1536, 1024]
    assert im.getchannel('A').getextrema()[0] == 0
    assert correction['review']['viewed']
    probes = []
    probe_versions = {
        **{f'town-path-normal-v{i}': 'fba6806e65c933242cc9ebd2332433d7b46557daa5db5d34f8959da065c28536' for i in range(1, 5)},
        'town-path-savem1-v1': 'f8ce902463a4e9b56a9f446ddc0fb5547aafc2cf8ab86656fce2dbb197fb228e',
        'town-path-savem1-v2': 'b9031e48e20ee846b6fef60cbddfe986fd9b033d962977db3a0bf8ab679c49af',
        'town-path-savem1-v3': 'b9031e48e20ee846b6fef60cbddfe986fd9b033d962977db3a0bf8ab679c49af',
        'town-path-savem1-v4': sha(p / 'tools/hd_town_path_probe.go'),
    }
    for name, source_hash in probe_versions.items():
        path = root / name / 'receipt.json'
        receipt = read(path)
        assert receipt['script_sha256'] == sha(path.parent / 'probe-source.go')
        # 舊探針以當次保存的來源及固定雜湊核對；不能換成新來源身份。
        assert receipt['script_sha256'] == source_hash
        assert receipt['source_save_unchanged'] and receipt['scratch_save_unchanged'] and receipt['save_mtime_unchanged']
        assert not receipt['state_injection'] and not receipt['snapshot_restore']
        for point in receipt['points']:
            for ext, field in (('memory', 'memory_sha256'), ('indexed', 'indexed_sha256'), ('palette', 'palette_sha256')):
                assert sha(path.parent / (point['file'] + '.' + ext)) == point[field]
        reached = any(point['label'].startswith('town-') and point['label'].endswith('-complete') for point in receipt['points'])
        assert not reached
        probes.append({'path': str(path.relative_to(p)), 'sha256': sha(path), 'points': len(receipt['points']),
                       'script_sha256': source_hash, 'source_save_name': receipt.get('source_save_name', 'SAVETEST'),
                       'source_save_sha256': receipt['source_save_sha256'], 'new_town_reached': False})
    common_path = root / 'cave-common-sources-v2/receipt.json'
    common = read(common_path)
    assert common['status'] == 'PASS_DCHR_NORMAL_RAM_BINDING_SOURCE_ONLY'
    assert common['script_sha256'] == sha(p / 'tools/hd_cave_common_sources.py')
    assert len(common['tiles']) == 39 and common['source_bytes'] == 1872
    assert [row['id'] for row in common['tiles']] == list(range(64, 103))
    assert len(common['normal_bindings']) == 10
    assert sum(row['pixels'] for row in common['platform_pixel_proof']) == 192
    assert sha(p / common['normal_receipt']) == common['normal_receipt_sha256']
    for binding in common['normal_bindings']:
        assert sha(p/binding['file']) == binding['sha256'] and binding['runtime_bytes_verified'] == 1872
    common_generation_path = root / 'cave-common-art-v1/generation-v1.json'
    common_generation = read(common_generation_path)
    assert common_generation['mode'] == 'built-in imagegen'
    assert common_generation['source_receipt_sha256'] == sha(common_path)
    assert sha(p/common_generation['reference_file']) == common_generation['reference_sha256']
    common_request = common_generation_path.parent / common_generation['request_file']
    assert sha(common_request) == common_generation['request_sha256'] and read(common_request)['prompt']
    assert sha(p/common_generation['output_file']) == common_generation['output_sha256']
    registration_path = root / 'cave-common-registration-v1/manifest.json'
    registration = read(registration_path)
    assert registration['status'] == 'REGISTERED_GREEN_PLATFORM_CANDIDATE_NOT_CONNECTED'
    # v1保存舊註冊器身份；不透明材質v2及目前來源由平台新稽核銜接。
    assert registration['script_sha256'] == '9f5094be9e60f07b7775aaa373d9c0a737335708b12e8d83e887a89c560d9e66'
    assert registration['source_sha256'] == sha(common_path)
    assert registration['generation_sha256'] == sha(common_generation_path)
    assert registration['target_size'] == [72, 24] and registration['outside_mask_written_pixels'] == 0
    assert registration['transparent_inside_mask'] == 51
    for row in registration['tiles']:
        assert sha(registration_path.parent/row['file']) == row['sha256'] and row['size'] == [24, 24]
    cycle_path = root / 'town-path-savem1-v4/receipt.json'
    cycle = read(cycle_path)
    earlier = read(root / 'town-path-savem1-v3/receipt.json')['points'][2]
    first = cycle['points'][0]
    assert all(first[key] == earlier[key] for key in ('memory_sha256', 'indexed_sha256', 'palette_sha256', 'regs', 'steps', 'cycles'))
    assert [point['label'] for point in cycle['points']] == ['raised-first', 'lowered', 'raised-again']
    platform_rows = []
    for point in cycle['points']:
        memory = (cycle_path.parent/(point['file']+'.memory')).read_bytes()
        primary = memory[0x1b10:0x11b10]
        table = int.from_bytes(primary[0xc004:0xc006], 'little')
        platform_rows.append(primary[table+2])
    assert platform_rows == [17, 24, 17]
    assert first['indexed_sha256'] == cycle['points'][2]['indexed_sha256']
    proof_paths = (source_path, ida_path, generation_path, correction_path, request_path,
                   common_path, common_generation_path, common_request, registration_path)
    files = ['tools/hd_town_sources.py', 'tools/ida_town_resources.py', 'tools/hd_town_path_probe.go', 'tools/hd_town_art.py',
             'tools/hd_town_stage_audit.py', 'tools/hd_catalog.py', 'docs/town-resources-source-index.json',
             'tools/hd_cave_common_sources.py', 'tools/hd_cave_common_art.py']
    report = {'schema': 1, 'status': 'SOURCE_AND_EIGHT_ART_CANDIDATES_NORMAL_PATH_PENDING', 'town_maps': 10,
              'map_cells': 17784, 'additional_towns': 8, 'generated_candidates': 8, 'unresolved': unresolved,
              'normal_new_town_reached': False, 'production_runtime_changed': False,
              'source_files': [{'path': name, 'sha256': sha(p/name)} for name in files],
              'proofs': [{'path': str(path.relative_to(p)), 'sha256': sha(path)} for path in proof_paths],
              'normal_probes': probes,
              'art_corrections': [{'key': 'ESMP', 'image': str(corrected_image.relative_to(p)),
                                   'sha256': sha(corrected_image), 'status': 'sign_and_four_windows_corrected_not_connected'}],
              'common_cave': {'source_record': 'zelres3.record.54', 'tiles': 39, 'normal_source_points': 10,
                              'platform_native_pixels': 192, 'registered_platform_ids': [64, 65, 66],
                              'normal_platform_rows': platform_rows, 'independent_normal_replay_ram_regs_video_equal': True,
                              'status': 'source_and_registered_candidate_normal_hd_pair_pending',
                              'transparent_inside_mask': 51, 'evidence': str(registration_path.relative_to(p))},
              'limits': ['八份生成圖均是美術候選，沒有正常新城鎮來源、前端接入或GUI驗收。',
                         '接縫、門洞、招牌與透明邊緣仍須校準；Pureza未知標記仍存在，不拿生成圖補原版用途。',
                         'DPAT只核靜態解碼及原版選擇端，正常載入與原生像素待驗；原型不進正式路徑。',
                         '全遊戲HD目標保持進行中，未更新封包、影片或README縮圖。']}
    assert not args.output.exists()
    args.output.write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')
    print(report['status'], report['map_cells'], 'cells', report['generated_candidates'], 'art candidates')


if __name__ == '__main__':
    main()

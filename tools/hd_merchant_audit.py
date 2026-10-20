#!/usr/bin/env python3
"""核對武器店已知圖面、正常交易／告別及實際視窗，不外推全遊戲HD。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageChops


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--output', type=Path, required=True)
    args = ap.parse_args()
    p = args.project
    root = p / 'workplace/full-hd-v1'
    sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
    read = lambda path: json.loads(path.read_text(encoding='utf-8'))
    layout = read(root / 'merchant-frames-layout-v1/receipt.json')
    assert layout['status'] == 'PASS_MERCHANT_TABLES_AND_NORMAL_BINDINGS'
    assert layout['script_sha256'] == sha(p / 'tools/hd_merchant_sources.py')
    assert layout['tile_count'] == 256 and layout['referenced_tile_count'] == 254
    assert layout['unreferenced_tiles'] == [139, 140]
    assert layout['normal_poses'] == list(range(8)) and len(layout['frames']) == 10
    assert len(layout['normal_bindings']) == 21
    for folder in ('merchant-frames-normal-v1', 'merchant-frames-normal-v2', 'merchant-frames-normal-v3', 'merchant-frames-normal-v5', 'merchant-frames-normal-v6'):
        receipt = read(root / folder / 'receipt.json')
        assert receipt['script_sha256'] == sha(root / folder / 'probe-source.go')
        assert receipt['source_index_sha256'] == sha(root / folder / 'source-index.json')
        assert receipt['source_save_unchanged'] and receipt['scratch_save_unchanged']
        assert receipt['source_save_mtime_ns'] == receipt['scratch_mtime_ns']
        assert not receipt['state_injection'] and not receipt['snapshot_restore']
        for sample in receipt['frames']:
            for suffix, field in (('.memory', 'memory_sha256'), ('.indexed', 'indexed_sha256'), ('.palette', 'palette_sha256')):
                assert sha(root / folder / (sample['file'] + suffix)) == sample[field]
    adlib = read(root / 'merchant-frames-normal-v3/receipt.json')
    assert adlib['adlib'] and sorted(map(int, adlib['selected_pose_indices'])) == [0, 3, 4, 5, 6, 7]
    final_probe = read(root / 'merchant-frames-normal-v6/receipt.json')
    assert final_probe['script_sha256'] == sha(p / 'tools/hd_room_probe.go')
    for i, point in enumerate(final_probe['points']):
        assert sha(root / 'merchant-frames-normal-v6' / f'point-{i:03d}.memory') == point['memory_sha256']
        assert sha(root / 'merchant-frames-normal-v6' / f'point-{i:03d}.indexed') == point['indexed_sha256']
    ida = read(root / 'merchant-frames-ida-v2/room-frames-evidence.json')
    assert ida['tool'] == 'IDA 9.4' and ida['sha256'] == layout['module_sha256']
    assert ida['source_index_sha256'] == sha(p / 'docs/merchant-frame-source-index.json')
    assert ida['script_sha256'] == sha(p / 'tools/ida_room_frames.py')
    assert len(ida['functions']) == 11 and len(ida['tables']) == 4
    assert (root / 'merchant-frames-ida-v2/ARMRPRO.i64').stat().st_size > 10000
    manifest_path = root / 'merchant-frames-complete-v2/manifest.json'
    manifest = read(manifest_path)
    assert sha(manifest_path) == '3b5f4da6b39ab356c06e8512d8d0c9023fecc5604db6f9a3b4ab85512aa3f064'
    assert manifest['script_sha256'] == sha(p / 'tools/hd_merchant_register.py')
    assert manifest['layout_sha256'] == sha(root / 'merchant-frames-layout-v1/receipt.json')
    generation_path = root / 'merchant-frames-art-v1/generation-v2.json'
    generation = read(generation_path)
    assert manifest['generation_sha256'] == sha(generation_path)
    assert generation['mode'] == 'built-in imagegen' and len(generation['jobs']) == len(generation['outputs']) == 8
    for job in generation['jobs']:
        assert job['prompt'] and len(job['refs']) == len(job['reference_inputs'])
        for ref in job['reference_inputs']:
            assert sha(p / ref['file']) == ref['sha256']
    for row in generation['outputs']:
        assert sha(p / row['file']) == row['sha256']
    for row in manifest['assets']:
        path = manifest_path.parent / row['file']
        assert sha(path) == row['sha256']
        im = Image.open(path)
        assert im.size == (282, 186) and im.mode == 'RGB'
    test_groups = [
        ('merchant-frames-runtime-v3', ['TestMerchantFramesAllSourcesAndPixels', 'TestMerchantFramesAllNativeRejections', 'TestMerchantFramesMissingAssetGroup', 'TestMerchantFramesNormalBranchesReadOnly',
                                       'TestMerchantReturnTownHDMaskFootprint', 'TestTownSceneryUnusedBytesAndSourceBoundaries', 'TestTownBackgroundOriginalMasksAndFallback',
                                       'TestTownFramesCompleteAndPartialIsolation', 'TestCastleNormalSourcesAndLayers', 'TestCastleRefusalIsAtomic']),
        ('merchant-frames-regression-v1', ['TestInteriorsSourcesAliasesAndPixels', 'TestMerchantPortraitPosesAndFallback', 'TestSageFramesAllSourcesAndDirectPixels', 'TestSageFramesFallbackAndNormalSnapshots']),
    ]
    for folder, names in test_groups:
        rows = [json.loads(line) for line in (root / folder / 'tests.jsonl').read_text().splitlines()]
        assert not any(row['Action'] in ('fail', 'skip') for row in rows)
        assert all(any(row['Action'] == 'pass' and row.get('Test') == name for row in rows) for name in names)
    render = read(root / 'merchant-frames-runtime-v3/render.json')
    assert len(render['frames']) == 10 and all(row['input_unchanged'] for row in render['frames'])
    assert sum(row['drawn_pixels'] for row in render['frames']) == 524520 and render['legacy_bases_preserved'] == 2
    native = read(root / 'merchant-frames-runtime-v3/native-sources.json')
    assert (native['accepted'], native['mixed_rejected'], native['source_guard_rejected']) == (21, 77, 45)
    normal = read(root / 'merchant-frames-runtime-v3/normal.json')
    assert normal['status'] == 'PASS_MERCHANT_TWO_NORMAL_BRANCHES_READ_ONLY'
    normal_points = 0
    for branch, command, gold in zip(normal['branches'], ('purchase', 'departure'), (800, 1000)):
        assert branch['branch'] == command and branch['gold'] == gold and branch['save_and_mtime_unchanged']
        assert branch['commands_sha256'] == sha(p / ('tools/merchant-hd-' + command + '.commands'))
        assert all(point['state_equal'] for point in branch['points'])
        normal_points += len(branch['points'])
    assert normal_points == 49
    returned = read(root / 'merchant-frames-runtime-v3/return-town.json')
    assert returned['status'] == 'PASS_NORMAL_MERCHANT_RETURN_TOWN_FOUR_HD_LAYERS'
    assert returned['memory_sha256'] == final_probe['points'][-1]['memory_sha256']
    assert returned['inputs_unchanged'] and returned['tail_variation_pixels_equal'] and returned['live_mask_boundary_rejected']
    assert (root / 'merchant-frames-runtime-v3/vet.log').stat().st_size == 0
    binary = root / 'merchant-frames-runtime-v3/zeliard-preview'
    gui_dir = root / 'merchant-frames-gui-v4'
    gui = read(gui_dir / 'receipt.json')
    assert gui['status'] == 'PASS_NORMAL_MERCHANT_FRAMES_HD_GUI'
    assert gui['binary_sha256'] == sha(binary) and gui['script_sha256'] == sha(p / 'tools/hd_merchant_smoke.py')
    assert sha(gui_dir / 'driver.py') == gui['script_sha256']
    assert gui['normal_DOS_exit'] == 0 and gui['source_save_unchanged'] and len(gui['pairs']) == 3
    shots = {row['name']: row for row in gui['shots']}
    palette = (p / 'workplace/music-release-v1/release-build/data/assets/town/frame/palette.bin').read_bytes()
    original_pixels = hd_pixels = restored_pixels = 0
    for pair in gui['pairs']:
        images, states = [], []
        for kind in ('original', 'hd', 'returned'):
            path = gui_dir / 'boot1' / pair[kind]
            assert sha(path) == shots[path.stem]['image_sha256']
            im = Image.open(path).convert('RGB')
            assert im.size == (960, 680)
            images.append(im)
            states.append(shots[path.stem]['state'])
        for field in ('steps', 'cycles', 'oracle_memory_sha256', 'oracle_indexed_sha256', 'oracle_palette_sha256'):
            assert states[0][field] == states[1][field] == states[2][field]
        assert ImageChops.difference(images[0].crop((0, 48, 960, 648)), images[2].crop((0, 48, 960, 648))).getbbox() is None
        indexed = (root / 'merchant-frames-layout-v1' / (pair['key'] + '.indexed')).read_bytes()
        expected = Image.frombytes('P', (96, 64), indexed)
        expected.putpalette(palette)
        assert ImageChops.difference(images[0].crop((168, 117, 456, 309)), expected.convert('RGB').resize((288, 192), Image.Resampling.NEAREST)).getbbox() is None
        art = Image.open(manifest_path.parent / (pair['key'] + '-hd.png')).convert('RGB')
        assert ImageChops.difference(images[1].crop((171, 120, 453, 306)), art).getbbox() is None
        original_pixels += 55296
        hd_pixels += 52452
        restored_pixels += 576000
    outside = shots['outside-after-merchant']['state']
    assert outside['town_scenery_hd_matched'] and not outside['merchant_frames_hd']['matched']
    files = ['cmd/zeliard-preview/main.go', 'internal/prototype/hd/merchant_frames.go', 'internal/prototype/hd/merchant_frames_test.go',
             'internal/prototype/hd/town_scenery.go', 'internal/prototype/hd/town_background.go', 'internal/prototype/hd/town_scenery_reload_test.go',
             'tools/hd_room_probe.go', 'tools/ida_room_frames.py', 'tools/hd_merchant_sources.py', 'tools/hd_merchant_register.py',
             'tools/hd_merchant_smoke.py', 'tools/hd_merchant_audit.py', 'tools/hd_catalog.py', 'docs/merchant-frame-source-index.json',
             'tools/merchant-hd-purchase.commands', 'tools/merchant-hd-departure.commands']
    proofs = ['merchant-frames-layout-v1/receipt.json', 'merchant-frames-ida-v2/room-frames-evidence.json',
              'merchant-frames-complete-v2/manifest.json', 'merchant-frames-art-v1/generation-v2.json',
              'merchant-frames-runtime-v3/tests.jsonl', 'merchant-frames-runtime-v3/normal.json', 'merchant-frames-runtime-v3/render.json',
              'merchant-frames-runtime-v3/native-sources.json', 'merchant-frames-runtime-v3/return-town.json',
              'merchant-frames-regression-v1/tests.jsonl', 'merchant-frames-gui-v4/receipt.json', 'merchant-frames-gui-v3/failure.json']
    proofs += [folder + '/receipt.json' for folder in ('merchant-frames-normal-v1', 'merchant-frames-normal-v2', 'merchant-frames-normal-v3', 'merchant-frames-normal-v5', 'merchant-frames-normal-v6')]
    report = {'schema': 1, 'status': 'PASS_MERCHANT_KNOWN_FRAMES_SOURCE_ART_NORMAL_GUI', 'binary_sha256': sha(binary),
              'source_tiles': 256, 'referenced_tiles': 254, 'unreferenced_tiles': [139, 140], 'registered_images': 10, 'poses': 8,
              'normal_points': normal_points, 'direct_render_pixels': 524520, 'gui_original_pixels': original_pixels,
              'gui_hd_pixels': hd_pixels, 'restored_game_pixels': restored_pixels, 'gui_keys': gui['visible_keys'], 'normal_DOS_exit': 0,
              'source_files': [{'path': name, 'sha256': sha(p / name)} for name in files],
              'proofs': [{'path': str((root / name).relative_to(p)), 'sha256': sha(root / name)} for name in proofs],
              'limits': ['只驗武器店已知八姿態及兩局部動畫，未引用非空圖格139、140用途仍未知。',
                         '購買與告別有49點完整原版狀態配對；GUI只抽代表圖，不稱每個動畫逐幀驗收。',
                         '未更新封包、影片與README縮圖；其他城鎮、敵人及首領仍有缺口，全遊戲HD未完成。']}
    assert not args.output.exists()
    args.output.write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
    print(report['status'], normal_points, gui['visible_keys'])


if __name__ == '__main__':
    main()

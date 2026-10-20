#!/usr/bin/env python3
"""核綠色平台來源、圖面、完整原版配對與實際視窗，不外推全遊戲。"""
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
    ap.add_argument('--gui', default='cave-platform-gui-v3')
    args = ap.parse_args()
    p = args.project
    b = p/'workplace/full-hd-v1'
    read = lambda path: json.loads(path.read_text())
    sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
    assert not args.output.exists(), '不得覆寫舊稽核'
    contract_path = b/'cave-platform-contract-v2/receipt.json'
    contract = read(contract_path)
    assert contract['status']=='READY_GREEN_PLATFORM_SOURCE_AND_ART'
    assert contract['script_sha256']==sha(p/'tools/hd_cave_platform_contract.py')
    assert contract['platform_entries']==72 and contract['opaque_hd_pixels']==1287
    for row in contract['proofs']:
        assert sha(p/row['file'])==row['sha256']
    reg_path = b/'cave-common-registration-v2/manifest.json'
    reg = read(reg_path)
    assert reg['script_sha256']==sha(p/'tools/hd_cave_common_art.py')
    assert reg['transparent_inside_mask']==reg['outside_mask_written_pixels']==0
    art_path = reg_path.parent/reg['art_file']
    assert sha(art_path)==reg['art_sha256']
    atlas = np.asarray(Image.open(art_path).convert('RGBA'))
    assert atlas.shape==(24,72,4) and np.count_nonzero(atlas[:,:,3])==1287
    gen_path = b/'cave-common-art-v2/generation-v2.json'
    generation = read(gen_path)
    assert generation['mode']=='built-in imagegen'
    assert sha(p/generation['output_file'])==generation['output_sha256']
    runtime = b/'cave-platform-runtime-v1'
    test_counts = {}
    for path, minimum in [(runtime/'unit-tests.jsonl',28),(runtime/'edge-tests.jsonl',1),(b/'cave-platform-hd-pair-v1/tests.jsonl',3)]:
        rows = [json.loads(line) for line in path.read_text().splitlines()]
        assert not any(row['Action'] in ('fail','skip') for row in rows)
        assert rows[-1]['Action']=='pass'
        passed = [row['Test'] for row in rows if row['Action']=='pass' and 'Test' in row and '/' not in row['Test']]
        assert len(passed)>=minimum
        test_counts[str(path.relative_to(p))]=len(passed)
    assert '6435' in (runtime/'unit-tests.jsonl').read_text()
    assert not (runtime/'vet.log').read_bytes() and not (runtime/'build.log').read_bytes()
    binary = runtime/'zeliard-preview'
    assert binary.read_bytes()[:4]==b'\x7fELF'
    normal_path = b/'cave-platform-hd-pair-v1/cave-platform/receipt.json'
    normal = read(normal_path)
    assert normal['state_equal'] and normal['save_unchanged'] and not normal['wrote'] and not normal['unimplemented']
    assert normal['hd_draws']==len(normal['points'])==10
    assert normal['commands_sha256']==sha(p/'tools/cave-platform.commands')
    targets = [row for row in normal['points'] if row['name'].startswith('platform-')]
    assert len(targets)==5 and all(row['terrain_hd_drawn'] and row['common_hd']['matched'] and row['state_equal'] for row in targets)
    for row in normal['points']:
        source = row['presentation']
        for suffix,key in [('memory','source_memory_sha256'),('body.memory','body_memory_sha256'),('indexed','source_indexed_sha256')]:
            assert sha(normal_path.parent/(row['name']+'.presented.'+suffix))==source[key]
    gui_path = b/args.gui/'receipt.json'
    gui = read(gui_path)
    assert gui['status']=='PASS_CAVE_PLATFORM_GUI_SCOPED'
    assert gui['binary_sha256']==sha(binary) and gui['save_unchanged'] and gui['normal_DOS_exit']==0
    assert gui['source_save_sha256']==normal['source_save_sha256']
    assert gui['script_sha256']==sha(p/'tools/hd_cave_platform_smoke.py')
    assert gui['catalog_helper_sha256']==sha(p/'tools/hd_cave_catalog_smoke.py')
    assert gui['helper_sha256']==sha(p/'tools/integration_smoke.py')
    assert len(gui['pairs'])==4
    shots = {row['name']:row for row in gui['shots']}
    native_equal, hd_pixels = 0,0
    for pair in gui['pairs']:
        images, states = [],[]
        for suffix in ['original','hd','returned']:
            row = shots[pair['name']+'-'+suffix]
            path = gui_path.parent/'boot1'/(row['name']+'.png')
            assert sha(path)==row['image_sha256']
            images.append(np.asarray(Image.open(path).convert('RGB')))
            states.append(row['state'])
        for key in ['steps','cycles','oracle_memory_sha256','oracle_indexed_sha256','oracle_palette_sha256']:
            assert states[0][key]==states[1][key]==states[2][key]
        # 工具列顯示切換狀態，獨立核遊戲畫布960x600。
        assert all(im.shape==(680,960,3) for im in images)
        assert np.array_equal(images[0][48:648],images[2][48:648])
        native_equal += 960*600
        common = states[1]['cave_common_hd']
        assert common['matched'] and len(common['placements'])==1
        point = common['placements'][0]
        x,y=point['X']*3,48+point['Y']*3
        actual=images[1][y:y+24,x:x+72]
        assert actual.shape==(24,72,3)
        mask=atlas[:,:,3]>0
        # 來源角色的24x24包絡排除合法遮擋；不以裁掉差分代替對拍。
        source=states[1]['cave_frame_presentation']
        actor_x,actor_y=source['source_actor_x']*3,48+source['source_actor_y']*3
        yy,xx=np.indices(mask.shape)
        mask &= ~((xx+x>=actor_x)&(xx+x<actor_x+72)&(yy+y>=actor_y)&(yy+y<actor_y+72))
        assert np.count_nonzero(mask)>=1000
        assert np.array_equal(actual[mask],atlas[:,:,:3][mask]),pair['name']
        hd_pixels += int(np.count_nonzero(mask))
    files = ['cmd/zeliard-preview/main.go','internal/prototype/hd/cave_common.go',
             'internal/prototype/hd/cave_common_test.go','internal/prototype/hd/cave_terrain.go',
             'internal/prototype/hd/cave_frames_pair_test.go','internal/prototype/hd/cave_remaining_families_test.go',
             'tools/hd_cave_common_art.py','tools/hd_cave_platform_contract.py','tools/ida_cave_common.py',
             'tools/cave-platform.commands','tools/cave-platform-magia.commands','tools/verify_cave_platform.sh',
             'tools/hd_cave_platform_smoke.py','tools/hd_cave_catalog_smoke.py','tools/hd_town_stage_audit.py',
             'tools/hd_cave_platform_audit.py','tools/hd_catalog.py','docs/cave-common-source-index.json']
    proofs = [contract_path,reg_path,gen_path,normal_path,gui_path,binary,runtime/'unit-tests.jsonl',runtime/'edge-tests.jsonl',
              runtime/'vet.log',runtime/'build.log',b/'cave-platform-hd-pair-v1/tests.jsonl']
    report = {'schema':1,'status':'PASS_GREEN_PLATFORM_SOURCE_ART_NORMAL_GUI',
              'spec':'docs/spec/043-cave-common-scenery.md','binary_sha256':sha(binary),
              'normal_points':10,'normal_platform_points':5,'platform_ids':[64,65,66],
              'independent_hd_pixels':6435,'gui_independent_hd_pixels':hd_pixels,
              'gui_original_return_pixels':native_equal,'test_counts':test_counts,
              'source_files':[{'path':f,'sha256':sha(p/f)}for f in files],
              'proofs':[{'path':str(f.relative_to(p)),'sha256':sha(f)}for f in proofs],
              'limits':['只驗MP10綠色平台，不當39格共用圖格或全洞窟完成。',
                        '魔法石路徑三個平台目標點沒有可呈現來源，缺口仍保留。',
                        '正常配對42個繪製中來源拒絕，十個具名點均有來源；不宣稱所有畫格覆蓋。',
                        'SAVEM1存檔前史未知；未驗其他地圖72表項的正常到達，未更新封包。']}
    args.output.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(report['status'],native_equal,'original pixels',hd_pixels,'HD pixels')


if __name__=='__main__':
    main()

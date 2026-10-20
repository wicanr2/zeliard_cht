#!/usr/bin/env python3
"""核魔法石正式來源、高清、正常原版配對與實際視窗，保留未覆蓋站位。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project',type=Path,required=True)
    ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args();p=args.project;b=p/'workplace/full-hd-v1'
    assert not args.output.exists(),'不得覆寫歷史稽核'
    sha=lambda path:hashlib.sha256(path.read_bytes()).hexdigest()
    read=lambda path:json.loads(path.read_text())
    stage_path=b/'cave-magia-stage-audit-v1.json';stage=read(stage_path)
    assert stage['status']=='READY_MAGIA_SOURCE_AND_ART_NOT_CONNECTED'
    # READY保存當時來源身份，正式增量另列當前來源，不改寫舊收據。
    for row in stage['proofs']:assert sha(p/row['path'])==row['sha256']
    reg_path=b/'cave-magia-complete-v1/manifest.json';reg=read(reg_path)
    assert reg['script_sha256']==sha(p/'tools/hd_cave_magia_art.py')
    art_path=reg_path.parent/reg['art_file'];assert sha(art_path)==reg['art_sha256']
    atlas=np.asarray(Image.open(art_path).convert('RGBA'));mask=atlas[:,:,3]>0
    assert atlas.shape==(24,24,4) and np.count_nonzero(mask)==414
    runtime=b/'cave-magia-runtime-v1'
    test_counts={}
    for path,expected in [(runtime/'unit-tests.jsonl',31),(runtime/'atomic-current-tests.jsonl',1),(b/'cave-magia-hd-pair-v1/tests.jsonl',3)]:
        rows=[json.loads(line)for line in path.read_text().splitlines()]
        assert rows[-1]['Action']=='pass' and not any(r['Action']in ('fail','skip')for r in rows)
        passed=[r['Test']for r in rows if r['Action']=='pass'and 'Test'in r and '/'not in r['Test']]
        assert len(passed)==expected,(str(path),len(passed))
        test_counts[str(path.relative_to(p))]=len(passed)
    assert '獨立HD像素31878，已清除槽4，重疊槽27'in (runtime/'unit-tests.jsonl').read_text()
    assert not (runtime/'vet.log').read_bytes() and not (runtime/'build.log').read_bytes()
    binary=runtime/'zeliard-preview';assert binary.read_bytes()[:4]==b'\x7fELF'
    pairs=[]
    for name,count in [('cave-platform-magia',3),('cave-platform',10)]:
        path=b/'cave-magia-hd-pair-v1'/name/'receipt.json';data=read(path)
        assert data['state_equal'] and data['save_unchanged'] and not data['unimplemented'] and not data['wrote']
        assert data['hd_draws']==len(data['points'])==count
        assert all(r['state_equal'] and r.get('terrain_hd_drawn')and r['presentation']['used']for r in data['points'])
        if count==3:assert all(r['magia_hd']['matched']and r['magia_hd']['visible_positions']>0 for r in data['points'])
        else:assert all(r['magia_hd']['empty']and r['magia_hd']['visible_positions']==0 for r in data['points'])
        pairs.append({'path':str(path.relative_to(p)),'points':count,'captured':data['captured'],'refused':data['refused'],
                      'missing_player_frames':data['player_coverage']['missing_visible_frames'],
                      'missing_player_source_keys':len(data['player_coverage']['missing_sources']),
                      'missing_weapon_source_keys':len(data['weapon_coverage']['missing_sources']or[])})
    gui_path=b/'cave-magia-gui-v2/receipt.json';gui=read(gui_path)
    assert gui['status']=='PASS_CAVE_MAGIA_GUI_SCOPED' and gui['magia_enabled']
    assert gui['binary_sha256']==sha(binary) and gui['save_unchanged'] and gui['normal_DOS_exit']==0
    assert gui['script_sha256']==sha(p/'tools/hd_cave_platform_smoke.py')
    assert gui['catalog_helper_sha256']==sha(p/'tools/hd_cave_catalog_smoke.py')
    shots={row['name']:row for row in gui['shots']}
    assert len(gui['pairs'])==4
    original_pixels,hd_pixels=0,0
    for pair in gui['pairs']:
        images,states=[],[]
        for suffix in ['original','hd','returned']:
            row=shots[pair['name']+'-'+suffix];path=gui_path.parent/'boot1'/(row['name']+'.png')
            assert sha(path)==row['image_sha256']
            im=np.asarray(Image.open(path).convert('RGB'));assert im.shape==(680,960,3)
            images.append(im);states.append(row['state'])
        for key in ['steps','cycles','oracle_memory_sha256','oracle_indexed_sha256','oracle_palette_sha256']:
            assert states[0][key]==states[1][key]==states[2][key]
        assert np.array_equal(images[0][48:648],images[2][48:648]);original_pixels+=960*600
        state=states[1];assert state['cave_terrain_hd_matched'] and state['cave_magia_hd']['matched']
        positions={(r['X'],r['Y'])for r in state['cave_magia_hd']['records']if r['Visibility']=='visible'}
        assert len(positions)==state['cave_magia_hd']['visible_positions']>0
        for x,y in positions:
            actual=images[1][48+y*3:48+y*3+24,x*3:x*3+24]
            assert actual.shape==(24,24,3)and np.array_equal(actual[mask],atlas[:,:,:3][mask]),(pair['name'],x,y)
            hd_pixels+=414
    # 離開平台的觀察不強行算HD通過，保留真正未匹配的站位。
    unmatched=[{'name':row['name'],'steps':row['state']['steps'],'source_steps':(row['state'].get('cave_frame_presentation')or{}).get('source_steps'),
                'scope':'未匹配整幅地形HD，根因待核，不當本次效果失敗或全場景完成。'}
               for row in gui['shots']if not row['name'].endswith(('-original','-returned'))and not row['state']['cave_terrain_hd_matched']]
    failed_path=b/'cave-magia-gui-v1/failed-state.json';failed=read(failed_path)
    assert not failed['cave_terrain_hd_matched'] and any(r['Visibility']=='visible'for r in failed['cave_frame_presentation']['magia_records'])
    files=['cmd/zeliard-preview/main.go','internal/prototype/hd/cave_magia.go','internal/prototype/hd/cave_magia_test.go',
           'internal/prototype/hd/cave_sources.go','internal/prototype/hd/cave_frames.go','internal/prototype/hd/cave_terrain.go',
           'internal/prototype/hd/cave_frames_pair_test.go','internal/prototype/hd/cave_remaining_families_test.go',
           'tools/verify_cave_platform.sh','tools/hd_cave_platform_smoke.py','tools/hd_cave_catalog_smoke.py',
           'tools/hd_cave_magia_runtime_audit.py','tools/hd_catalog.py']
    proofs=[stage_path,reg_path,gui_path,failed_path,binary,runtime/'unit-tests.jsonl',runtime/'atomic-current-tests.jsonl',runtime/'vet.log',runtime/'build.log',b/'cave-magia-hd-pair-v1/tests.jsonl']
    proofs += [p/row['path']for row in pairs]
    report={'schema':1,'status':'PASS_MAGIA_SOURCE_ART_NORMAL_GUI_SCOPED','spec':'docs/spec/044-cave-magia-presentation.md',
            'binary_sha256':sha(binary),'source_frames':27,'independent_hd_pixels':31878,'normal_state_equal_points':13,
            'normal_pairs':pairs,'gui_original_return_pixels':original_pixels,'gui_independent_hd_pixels':hd_pixels,
            'gui_unmatched_observations':unmatched,'earlier_unmatched_station':{'path':str(failed_path.relative_to(p)),'source_effect_recognized':True,'terrain_hd_matched':False,'scope':'v1物品使用後站位；原因待核。v2平台及離開截圖另按實際狀態記錄，不推定沿途全覆蓋。'},'test_counts':test_counts,
            'source_files':[{'path':f,'sha256':sha(p/f)}for f in files],
            'proofs':[{'path':str(f.relative_to(p)),'sha256':sha(f)}for f in proofs],
            'limits':['只驗可見／已清除／重疊來源及此正常使用與平台路徑，未宣稱所有魔法或全遊戲完成。',
                      'GUI v1物品使用後站位未匹配地形HD，保留原圖與日誌；v2具名截圖依實際狀態記錄，不宣稱沿途全覆蓋。',
                      '魔法石路徑92次、普通路徑42次來源拒絕未全部消除；未驗耗盡、部分遮擋及邊緣。',
                      '主角／武器及其他場景美術缺口保留，封包與影片未更新。']}
    args.output.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n');print(report['status'],original_pixels,'original pixels',hd_pixels,'HD pixels')


if __name__=='__main__':main()

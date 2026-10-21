#!/usr/bin/env python3
"""核十二主角來源、六色彩變體、正常完成來源缺圖清零與實際GUI。"""
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
    sha=lambda file:hashlib.sha256(file.read_bytes()).hexdigest()
    read=lambda file:json.loads(file.read_text())
    source_path=b/'cave-saved-gear-source-v1/receipt.json';source=read(source_path)
    assert source['status']=='SAVED_GEAR_SOURCE_ONLY'and len(source['sources'])==22
    generation_path=b/'cave-saved-player-art-v1/generation.json';gen=read(generation_path)
    assert gen['mode']=='built-in imagegen'and gen['prompt']
    for ref in gen['references']:assert sha(p/ref['file'])==ref['sha256']
    assert sha(p/gen['output_file'])==gen['output_sha256']
    reg_path=b/'cave-saved-player-complete-v1/manifest.json';reg=read(reg_path)
    assert reg['script_sha256']==sha(p/'tools/hd_saved_player_art.py')
    assert reg['source_receipt_sha256']==sha(source_path)and reg['generation_sha256']==sha(generation_path)
    assert len(reg['poses'])==12 and sum(r['palette_variant']for r in reg['poses'])==6
    assert sum(r['known_shape_differences']==0 for r in reg['poses'])==5
    keys={r['key']for r in reg['poses']}
    for r in reg['poses']:
        file=reg_path.parent/r['file'];assert sha(file)==r['sha256']
        im=np.asarray(Image.open(file).convert('RGBA'));assert im.shape==(72,72,4)
        assert hashlib.sha256(im.tobytes()).hexdigest()==r['rgba_sha256']
    runtime=b/'cave-saved-player-runtime-v1';binary=runtime/'zeliard-preview';assert binary.read_bytes()[:4]==b'\x7fELF'
    test_counts={}
    for file,count in[(runtime/'tests.jsonl',42),(b/'cave-saved-player-normal-pair-v1/tests.jsonl',3)]:
        rows=[json.loads(line)for line in file.read_text().splitlines()]
        assert rows[-1]['Action']=='pass'and not any(r['Action']in('fail','skip')for r in rows)
        actual=sum(r['Action']=='pass'and'Test'in r and'/'not in r['Test']for r in rows)
        assert actual==count,(file,actual);test_counts[str(file.relative_to(p))]=actual
    assert '62208註冊RGBA與31376個未遮擋混色像素獨立相同'in(runtime/'tests.jsonl').read_text()
    assert not(runtime/'vet.log').read_bytes()and not(runtime/'build.log').read_bytes()
    normal=[];observed=set()
    for name,count in[('cave-platform',10),('cave-platform-magia',3)]:
        file=b/'cave-saved-player-normal-pair-v1'/name/'receipt.json';data=read(file)
        assert data['state_equal']and data['save_unchanged']and not data['wrote']and not data['unimplemented']
        assert len(data['points'])==data['hd_draws']==count
        assert all(r['state_equal']and r['terrain_hd_drawn']and r['presentation']['used']for r in data['points'])
        for kind in['player','weapon']:
            c=data[kind+'_coverage'];assert c['missing_visible_frames']==0 and not c['missing_sources'],(name,kind)
        observed.update(data['player_coverage']['keys'])
        normal.append({'path':str(file.relative_to(p)),'points':count,'known_player_frames':data['player_coverage']['known_frames'],
                       'missing_player_frames':0,'missing_weapon_frames':0,'captured':data['captured'],'refused':data['refused']})
    assert keys<=observed
    gui_path=b/'cave-saved-player-gui-v3/receipt.json';gui=read(gui_path)
    assert gui['status']=='PASS_CAVE_EXTRA_PLAYER_GUI_SCOPED'and gui['extra_player_enabled']and gui['cyan_enabled']
    assert gui['normal_DOS_exit']==0 and gui['save_unchanged']and gui['binary_sha256']==sha(binary)
    assert gui['script_sha256']==sha(p/'tools/hd_cave_platform_smoke.py')and gui['catalog_helper_sha256']==sha(p/'tools/hd_cave_catalog_smoke.py')
    assert len(gui['pairs'])==7
    shots={r['name']:r for r in gui['shots']};original_pixels=0;extra_key=None;extra_pose=None
    for pair in gui['pairs']:
        images=[];states=[]
        for suffix in['original','hd','returned']:
            row=shots[pair['name']+'-'+suffix];file=gui_path.parent/'boot1'/(row['name']+'.png')
            assert sha(file)==row['image_sha256'];im=np.asarray(Image.open(file).convert('RGB'));assert im.shape==(680,960,3)
            images.append(im);states.append(row['state'])
        for field in['steps','cycles','oracle_memory_sha256','oracle_indexed_sha256','oracle_palette_sha256']:
            assert states[0][field]==states[1][field]==states[2][field]
        assert np.array_equal(images[0][48:648],images[2][48:648]);original_pixels+=960*600
        assert states[1]['cave_terrain_hd_matched']
        if pair['name']=='extra-player':
            player=states[1]['cave_player_hd'];assert player['matched']and 55<=player['pose']<67 and player['source_key']in keys
            extra_key,extra_pose=player['source_key'],player['pose'];assert np.any(images[0][90:522]!=images[1][90:522])
    assert extra_key
    save=gui_path.parent/'user/game/SAVETEST.usr';orig=p/'workplace/orig/SAVEM1.USR'
    assert sha(save)==sha(orig)and save.stat().st_mtime_ns==orig.stat().st_mtime_ns
    failure_path=b/'cave-saved-player-gui-v1/failed-state.json';failure=read(failure_path)
    assert failure['native_primary_0080']+failure['native_primary_0083']==46
    branch_path=b/'cave-saved-player-gui-v2/failed-state.json';branch=read(branch_path)
    refused=branch['cave_frame_presentation']
    assert refused['last_refusal']['steps']==refused['last_rejected_source']['steps']
    assert refused['last_refusal']['primary_fields']['01B1:FF41']==1
    assert refused['last_rejected_source']['body_primary_fields']['01B1:FF41']==1
    files=['cmd/zeliard-preview/main.go','internal/prototype/hd/cave_player_extra.go','internal/prototype/hd/cave_player_extra_test.go',
           'internal/prototype/hd/cave_frames_pair_test.go','tools/hd_saved_player_art.py','tools/hd_saved_player_runtime_audit.py',
           'tools/hd_cave_platform_smoke.py','tools/hd_cave_catalog_smoke.py','tools/hd_catalog.py']
    proofs=[source_path,generation_path,reg_path,gui_path,binary,runtime/'tests.jsonl',runtime/'vet.log',runtime/'build.log',
            b/'cave-saved-player-normal-pair-v1/tests.jsonl',failure_path,b/'cave-saved-player-gui-v1/driver.py',branch_path,b/'cave-saved-player-gui-v2/driver.py']+[p/r['path']for r in normal]
    report={'schema':1,'status':'PASS_SAVED_PLAYER_SOURCE_ART_NORMAL_GUI_SCOPED','binary_sha256':sha(binary),
            'registered_source_keys':12,'palette_variants':6,'existing_poses_preserved':55,'normal_state_equal_points':13,
            'normal_pairs':normal,'gui_original_return_pixels':original_pixels,'gui_extra_player_key':extra_key,'gui_extra_player_pose':extra_pose,
            'registered_rgba_pixels':62208,'independent_blended_pixels':31376,'test_counts':test_counts,
            'prior_gui_failure':{'path':str(failure_path.relative_to(p)),'player_origin':46,'verified_platform_origin':44,'scope':'外部GUI鍵序跳過頭，v2只用正常方向鍵修正站位，不改產品。'},
            'unresolved_ff41_source':{'path':str(branch_path.relative_to(p)),'value':1,'scope':'長按攻擊測試途中遇到未支援來源，仍退回原貌；v3沿已驗短按抽樣，不將此缺口稱為完成。'},
            'source_files':[{'path':f,'sha256':sha(p/f)}for f in files],
            'proofs':[{'path':str(f.relative_to(p)),'sha256':sha(f)}for f in proofs],
            'limits':['兩條路徑的已選完成來源無主角／武器缺圖；原42／92次來源拒絕仍保留，不當每畫格皆HD。',
                      '六個色彩變體均驗已存正常來源；GUI只驗具名實際姿態，不冒稱十二鍵逐一GUI。',
                      '31376混色像素是已存來源驗證，不當GUI逐像素證據。其他場景、美術、封包與影片仍待補。']}
    args.output.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(report['status'],original_pixels,'original pixels',extra_key[:12])


if __name__=='__main__':main()

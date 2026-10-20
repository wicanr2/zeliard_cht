#!/usr/bin/env python3
"""核青藍刀光的固定來源、十圖、正常缺圖清零及GUI切回，保留主角缺口。"""
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
    source_path=b/'cave-saved-gear-source-v1/receipt.json';source=read(source_path)
    assert source['status']=='SAVED_GEAR_SOURCE_ONLY' and len(source['sources'])==22
    assert source['script_sha256']==sha(p/'internal/prototype/hd/cave_saved_gear_study_test.go')
    generation_path=b/'cave-cyan-weapon-art-v1/generation.json';generation=read(generation_path)
    assert generation['mode']=='built-in imagegen' and generation['prompt']
    for ref in generation['references']:assert sha(p/ref['file'])==ref['sha256']
    assert sha(p/generation['output_file'])==generation['output_sha256']
    reg_path=b/'cave-cyan-weapon-complete-v1/manifest.json';reg=read(reg_path)
    assert reg['script_sha256']==sha(p/'tools/hd_cyan_weapon_art.py')
    assert reg['source_receipt_sha256']==sha(source_path) and reg['generation_sha256']==sha(generation_path)
    assert len(reg['poses'])==10 and all(r['native_shape_differences']==0 and 16<=r['native_rgba_differences']<=54 for r in reg['poses'])
    keys={r['key']for r in reg['poses']}
    for r in reg['poses']:
        file=reg_path.parent/r['file'];assert sha(file)==r['sha256']
        im=np.asarray(Image.open(file).convert('RGBA'));assert im.shape==(96,96,4)
        assert hashlib.sha256(im.tobytes()).hexdigest()==r['rgba_sha256']
    runtime=b/'cave-cyan-weapon-runtime-v2';binary=runtime/'zeliard-preview'
    assert binary.read_bytes()[:4]==b'\x7fELF'
    test_counts={}
    for file,count in [(runtime/'tests.jsonl',38),(runtime/'atomic-current-tests.jsonl',1),(b/'cave-cyan-weapon-normal-pair-v1/tests.jsonl',3)]:
        rows=[json.loads(s)for s in file.read_text().splitlines()]
        assert rows[-1]['Action']=='pass' and not any(r['Action']in('fail','skip')for r in rows)
        actual=sum(r['Action']=='pass'and'Test'in r and'/'not in r['Test']for r in rows)
        assert actual==count,(file,actual);test_counts[str(file.relative_to(p))]=actual
    assert '92160註冊RGBA與5152個透明度合成像素獨立相同'in(runtime/'tests.jsonl').read_text()
    assert not(runtime/'vet.log').read_bytes() and not(runtime/'build.log').read_bytes()
    normal=[]
    for name,count in [('cave-platform',10),('cave-platform-magia',3)]:
        file=b/'cave-cyan-weapon-normal-pair-v1'/name/'receipt.json';data=read(file)
        assert data['state_equal']and data['save_unchanged']and not data['wrote']and not data['unimplemented']
        assert len(data['points'])==data['hd_draws']==count
        assert all(r['state_equal']and r['terrain_hd_drawn']and r['presentation']['used']for r in data['points'])
        weapons=data['weapon_coverage'];assert weapons['missing_visible_frames']==0 and not weapons['missing_sources']
        assert weapons['known_frames']>0
        normal.append({'path':str(file.relative_to(p)),'points':count,'known_weapon_frames':weapons['known_frames'],
                       'missing_weapon_frames':0,'missing_player_frames':data['player_coverage']['missing_visible_frames'],
                       'captured':data['captured'],'refused':data['refused']})
    gui_path=b/'cave-cyan-weapon-gui-v1/receipt.json';gui=read(gui_path)
    assert gui['status']=='PASS_CAVE_CYAN_WEAPON_GUI_SCOPED'and gui['cyan_enabled']and gui['doors_enabled']and gui['magia_enabled']
    assert gui['normal_DOS_exit']==0 and gui['save_unchanged']and gui['binary_sha256']==sha(binary)
    assert gui['script_sha256']==sha(p/'tools/hd_cave_platform_smoke.py')and gui['catalog_helper_sha256']==sha(p/'tools/hd_cave_catalog_smoke.py')
    assert len(gui['pairs'])==6
    shots={r['name']:r for r in gui['shots']};original_pixels=0;cyan_key=None
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
        if pair['name']=='cyan-weapon':
            weapon=states[1]['cave_weapon_hd'];assert weapon['matched']and weapon['pose']>=10 and weapon['source_key']in keys
            cyan_key=weapon['source_key'];assert np.any(images[0][90:522]!=images[1][90:522])
    assert cyan_key
    save=gui_path.parent/'user/game/SAVETEST.usr';orig=p/'workplace/orig/SAVEM1.USR'
    assert sha(save)==sha(orig)and save.stat().st_mtime_ns==orig.stat().st_mtime_ns
    files=['cmd/zeliard-preview/main.go','internal/prototype/hd/cave_weapon_cyan.go','internal/prototype/hd/cave_weapon_cyan_test.go',
           'internal/prototype/hd/cave_saved_gear_study_test.go','internal/prototype/hd/cave_frames_pair_test.go',
           'tools/hd_cyan_weapon_art.py','tools/hd_cyan_weapon_runtime_audit.py','tools/hd_cave_platform_smoke.py','tools/hd_cave_catalog_smoke.py','tools/hd_catalog.py']
    proofs=[source_path,generation_path,reg_path,gui_path,binary,runtime/'tests.jsonl',runtime/'atomic-current-tests.jsonl',runtime/'vet.log',runtime/'build.log',
            b/'cave-cyan-weapon-normal-pair-v1/tests.jsonl']+[p/r['path']for r in normal]
    report={'schema':1,'status':'PASS_CYAN_WEAPON_SOURCE_ART_NORMAL_GUI_SCOPED','binary_sha256':sha(binary),
            'registered_weapons':10,'normal_state_equal_points':13,'normal_pairs':normal,'gui_original_return_pixels':original_pixels,
            'gui_weapon_key':cyan_key,'registered_rgba_pixels':92160,'independent_blended_pixels':5152,'test_counts':test_counts,
            'source_files':[{'path':f,'sha256':sha(p/f)}for f in files],
            'proofs':[{'path':str(f.relative_to(p)),'sha256':sha(f)}for f in proofs],
            'limits':['只驗已列青藍十來源與正常路徑，不推定裝備名稱或全遊戲武器完成。',
                      'GUI驗實際出劍、狀態與原貌切回；5152個獨立混色像素是已存來源驗證，不冒稱GUI逐像素比對。',
                      '主角12來源及其餘美術缺口仍待補，封包與影片未更新。']}
    args.output.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(report['status'],original_pixels,'original pixels',cyan_key[:12])


if __name__=='__main__':main()

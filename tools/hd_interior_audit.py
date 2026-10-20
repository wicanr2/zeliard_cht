#!/usr/bin/env python3
"""Docker內稽核室內來源、圖面與正常小屋GUI，保留旅館待驗範圍。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageChops


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project',type=Path,required=True)
    ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args();project=args.project;study=project/'workplace/full-hd-v1'
    sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
    read=lambda p:json.loads(p.read_text())
    sources=read(study/'interior-layout-v2/receipt.json')
    assert sources['script_sha256']==sha(project/'tools/hd_interior_sources.py')
    profiles={x['name']:x for x in sources['records']}
    assert profiles['OMOYA']['tile_count']==135 and profiles['INN']['tile_count']==112
    assert not profiles['OMOYA']['unreferenced_tile_ids'] and not profiles['INN']['unreferenced_tile_ids']
    assert profiles['OMOYA']['normal_binding']['pixels']==17408
    assert profiles['KENJYA']['normal_binding']['pixels']==6144 and profiles['KENJYA']['normal_binding']['packed_bytes']==7968
    assert profiles['INN']['normal_binding'] is None
    registration=read(study/'interior-complete-v2/manifest.json')
    assert registration['native_tiles']==247 and len(registration['assets'])==5
    assert registration['script_sha256']==sha(project/'tools/hd_interior_register.py')
    assert registration['source_receipt_sha256']==sha(study/'interior-layout-v2/receipt.json')
    assert registration['generation_sha256']==sha(study/'interior-art-v1/generation-v2.json')
    assert registration['area_tool_sha256']==sha(project/'tools/hd_mpp1_art.py')
    for asset in registration['assets']:
        path=study/'interior-complete-v2'/asset['file'];assert sha(path)==asset['sha256']
        assert Image.open(path).size==tuple(asset['size'])
        assert sha(path)==sha(study/'interior-complete-v1'/asset['file'])
    generation=read(study/'interior-art-v1/generation-v2.json')
    assert generation['mode']=='builtin-imagegen' and len(generation['requests'])==6
    assert sum(x['review'].startswith('selected') for x in generation['requests'])==5
    for entry in generation['requests']:
        assert sha(study/'interior-art-v1'/entry['output']['file'])==entry['output']['sha256']
        for ref in entry['references']:
            reference=project/Path(ref['path']).relative_to('/home/anr2/cht/zeliard')
            assert sha(reference)==ref['sha256']
    for file in ('OMOYPRO-evidence.json','INNAPRO-evidence.json'):
        ida=read(study/'interior-ida-v2'/file)
        assert ida['tool']=='IDA 9.4' and ida['functions'] and ida['tables']
        assert ida['source_index_sha256']==sha(project/'docs/interior-source-index.json')
        assert ida['script_sha256']==sha(project/'tools/ida_interior_evidence.py')
        assert (study/'interior-ida-v2'/(file.split('-')[0]+'.i64')).stat().st_size>10000
    tests=[json.loads(line) for line in (study/'interior-runtime-v1/tests.jsonl').read_text().splitlines()]
    test_names=['TestInteriorsSourcesAliasesAndPixels','TestInteriorsAtomicFallback','TestInteriorsMissingAssets','TestInteriorsNormalHutReadOnly']
    assert not any(x['Action'] in ('fail','skip') for x in tests)
    assert all(any(x['Action']=='pass' and x.get('Test')==name for x in tests) for name in test_names)
    render=read(study/'interior-runtime-v1/render.json')
    assert len(render['frames'])==8 and all(x['input_unchanged'] for x in render['frames'])
    assert sum(x['drawn_pixels'] for x in render['frames'])==523836
    normal=read(study/'interior-runtime-v1/normal-hut.json')
    assert normal['status']=='PASS_NORMAL_HUT_READ_ONLY' and len(normal['checkpoints'])==5
    assert all(x['state_equal'] for x in normal['checkpoints'])
    assert [x['matched'] for x in normal['checkpoints']]==[False,False,True,True,False]
    assert normal['commands_sha256']==sha(project/'tools/hut-hd.commands')
    gui_dir=study/'interior-gui-v1';gui=read(gui_dir/'receipt.json')
    binary=study/'interior-runtime-v1/zeliard-preview'
    assert gui['status']=='PASS_NORMAL_HUT_HD_GUI' and gui['normal_DOS_exit']==0
    assert gui['binary_sha256']==sha(binary) and gui['script_sha256']==sha(project/'tools/hd_interior_smoke.py')
    assert len(gui['pairs'])==2
    shots={x['name']:x for x in gui['shots']}
    restored=0;hut_pixels=0
    for pair in gui['pairs']:
        images=[];states=[]
        for key in ('original','hd','returned'):
            path=gui_dir/'boot1'/pair[key];meta=shots[path.stem]
            assert sha(path)==meta['image_sha256']
            im=Image.open(path).convert('RGB');assert im.size==(960,680)
            images.append(im);states.append(meta['state'])
        for key in ('steps','cycles','oracle_memory_sha256','oracle_indexed_sha256','oracle_palette_sha256'):
            assert states[0][key]==states[1][key]==states[2][key]
        assert ImageChops.difference(images[0].crop((0,48,960,648)),images[2].crop((0,48,960,648))).getbbox() is None
        restored+=576000
        if pair['name']=='hut':
            assert states[1]['interior_hd_matched'] and states[1]['interior_hd']['profile']=='hut'
            crop=(288,138,696,522)
            original=Image.open(study/'interior-layout-v2/OMOYA-base-original.png').convert('RGB').resize((408,384),Image.Resampling.NEAREST)
            art=Image.open(study/'interior-complete-v2/hut-hd-v1.png').convert('RGB')
            assert ImageChops.difference(images[0].crop(crop),original).getbbox() is None
            assert ImageChops.difference(images[1].crop(crop),art).getbbox() is None
            hut_pixels=408*384
        else:
            assert not states[1]['interior_hd_matched'] and states[1]['town_scenery_hd_matched']
    assert restored==1152000 and hut_pixels==156672
    back=read(study/'town-back-audit-v1.json')
    for row in back['source_files']:
        if row['path']!='cmd/zeliard-preview/main.go':assert sha(project/row['path'])==row['sha256']
    source_files=['cmd/zeliard-preview/main.go','internal/prototype/hd/interior.go','internal/prototype/hd/interior_test.go',
                  'tools/hd_interior_sources.py','tools/hd_interior_register.py','tools/ida_interior_evidence.py',
                  'tools/hd_interior_smoke.py','tools/hd_interior_audit.py','tools/hd_catalog.py','tools/hut-hd.commands',
                  'docs/interior-source-index.json']
    proof_files=['interior-layout-v2/receipt.json','interior-complete-v2/manifest.json','interior-art-v1/generation-v2.json',
                 'interior-ida-v2/OMOYPRO-evidence.json','interior-ida-v2/INNAPRO-evidence.json',
                 'interior-runtime-v1/tests.jsonl','interior-runtime-v1/render.json','interior-runtime-v1/normal-hut.json',
                 'interior-gui-v1/receipt.json','hut-normal-v2/receipt.json','home-normal-v1/receipt.json']
    report={'schema':1,'status':'PASS_INTERIOR_SOURCES_ART_AND_NORMAL_HUT_GUI','binary_sha256':sha(binary),
            'source_tiles':247,'registered_images':5,'native_frame_entries':8,'unique_hut_inn_frames':5,
            'normal_source_pixels':23552,'source_and_art_test_pixels':523836,'normal_points':5,
            'gui_hut_source_pixels':156672,'gui_hut_hd_pixels':hut_pixels,'restored_game_pixels':restored,
            'test_names':test_names,'normal_DOS_exit':0,
            'source_files':[{'path':name,'sha256':sha(project/name)} for name in source_files],
            'proofs':[{'path':str((study/name).relative_to(project)),'sha256':sha(study/name)} for name in proof_files],
            'limits':['小屋已驗正常冷新局入亭台、石像HD、原貌切回及離開；只限這條正常路徑。',
                      '旅館四種圖面接入並通過結構完整像素與退回，正常到達、觸發及GUI未驗。',
                      '長老只新增來源格式核對，不宣稱其所有圖格與動畫已HD。',
                      '未更新三平台封包或推廣片；其他城鎮、敵人／首領及其餘美術仍未完成。']}
    assert not args.output.exists()
    args.output.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    print(report['status'],restored,hut_pixels,sha(binary))


if __name__=='__main__':
    main()

#!/usr/bin/env python3
"""核對長老圖塊、手繪圖組、正常原版狀態與實際視窗；不改寫歷史身份。"""
import argparse
import datetime
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageChops


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project',type=Path,required=True)
    ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args();p=args.project;study=p/'workplace/full-hd-v1'
    sha=lambda path:hashlib.sha256(path.read_bytes()).hexdigest()
    read=lambda path:json.loads(path.read_text(encoding='utf-8'))
    layout=read(study/'sage-frames-layout-v1/receipt.json')
    assert layout['status']=='PASS_ALL_SAGE_TILES_AND_NORMAL_FRAME_BINDINGS'
    assert layout['script_sha256']==sha(p/'tools/hd_sage_sources.py')
    assert layout['covered_tile_count']==166 and layout['normal_poses']==list(range(8))
    assert len(layout['structural_variants'])==38 and len(layout['normal_bindings'])==16 and len(layout['normal_rejected_frames'])==3
    ida=read(study/'sage-frames-ida-v4/sage-frames-evidence.json')
    assert ida['tool']=='IDA 9.4' and ida['sha256']==layout['module_sha256']
    assert ida['source_index_sha256']==sha(p/'docs/sage-frame-source-index.json') and ida['script_sha256']==sha(p/'tools/ida_sage_frames.py')
    assert len(ida['tables'])==5 and len(ida['functions'])==17
    assert (study/'sage-frames-ida-v4/KENJPRO.i64').stat().st_size>10000
    # 格式化後的來源工具獨立重播。mtime差異不能包裝成同狀態重播。
    replay_old=read(study/'sage-frames-normal-v4/receipt.json');replay=read(study/'sage-frames-normal-v5/receipt.json')
    assert replay['script_sha256']==sha(p/'tools/hd_sage_probe.go')
    assert replay['source_save_unchanged'] and replay['scratch_save_unchanged'] and not replay['state_injection'] and not replay['snapshot_restore']
    assert len(replay['frames'])==len(replay_old['frames'])==19
    difference_offsets=set()
    for old,new in zip(replay_old['frames'],replay['frames']):
        assert old['file']==new['file'] and old['indexed_sha256']==new['indexed_sha256']
        a=(study/'sage-frames-normal-v4'/(old['file']+'.memory')).read_bytes()
        b=(study/'sage-frames-normal-v5'/(new['file']+'.memory')).read_bytes()
        assert hashlib.sha256(a).hexdigest()==old['memory_sha256'] and hashlib.sha256(b).hexdigest()==new['memory_sha256']
        difference_offsets.update(i for i,(x,y) in enumerate(zip(a,b)) if x!=y)
    assert difference_offsets=={0x257f,0x2580}
    for folder in ('sage-frames-normal-v4','sage-frames-normal-v5'):
        stamp=datetime.datetime.fromtimestamp((study/folder/'scratch/SAVETEST.usr').stat().st_mtime,datetime.timezone.utc)
        packed=stamp.hour<<11|stamp.minute<<5|stamp.second//2
        memory=(study/folder/'frame-002.memory').read_bytes()
        assert int.from_bytes(memory[0x257f:0x2581],'little')==packed
    manifest=read(study/'sage-frames-complete-v2/manifest.json')
    assert sha(study/'sage-frames-complete-v2/manifest.json')=='41bd587485d0e54b48063bce43b26433c560b064cd9c339ba84f6d61df1e470c'
    assert manifest['script_sha256']==sha(p/'tools/hd_sage_register.py')
    generation=read(study/'sage-frames-art-v1/generation-v3.json')
    assert generation['mode']=='built-in imagegen' and len(generation['jobs'])==12 and len(generation['selected'])==8
    assert generation['parent_generation_sha256']==sha(study/'sage-frames-art-v1/generation-v2.json')==manifest['generation_sha256']
    for job in generation['jobs']:
        assert job['prompt'] and len(job['reference_inputs'])==len(job['refs'])
        for ref in job['reference_inputs']:assert sha(p/ref['file'])==ref['sha256']
    for row in generation['outputs']:assert sha(p/row['file'])==row['sha256']
    for row in manifest['assets']:
        path=study/'sage-frames-complete-v2'/row['file'];assert sha(path)==row['sha256']
        im=Image.open(path);assert im.size==(288,192) and im.mode=='RGB'
    test_groups=[('sage-frames-runtime-v1',['TestSageFramesAllSourcesAndDirectPixels','TestSageFramesFallbackAndNormalSnapshots','TestSageFramesMissingAssetGroup','TestSageFramesNormalReadOnly']),
                 ('sage-frames-regression-v1',['TestInteriorsSourcesAliasesAndPixels','TestInteriorsAtomicFallback','TestInteriorsMissingAssets','TestInteriorsNormalHutReadOnly','TestSagePortraitPreservesOriginalAnimation','TestSageCenteredSourceFrames'])]
    for folder,names in test_groups:
        rows=[json.loads(s) for s in (study/folder/'tests.jsonl').read_text().splitlines()]
        assert not any(x['Action'] in ('fail','skip') for x in rows)
        assert all(any(x['Action']=='pass' and x.get('Test')==name for x in rows) for name in names)
    render=read(study/'sage-frames-runtime-v1/render.json')
    assert len(render['samples'])==76 and all(x['inputs_unchanged'] and x['structural_fixture'] for x in render['samples'])
    direct_pixels=sum(x['drawn_pixels'] for x in render['samples']);assert direct_pixels==3986352
    normal=read(study/'sage-frames-runtime-v1/normal.json')
    assert normal['status']=='PASS_NORMAL_SAGE_READ_ONLY' and sorted(map(int,normal['poses']))==list(range(8))
    assert normal['command_sha256']==sha(p/'tools/sage-hd.commands') and all(x['state_equal'] for x in normal['points'])
    assert len(normal['points'])==24 and normal['points'][-1]['name']=='outside' and not normal['points'][-1]['matched']
    assert (study/'sage-frames-runtime-v1/vet.log').stat().st_size==0
    binary=study/'sage-frames-runtime-v1/zeliard-preview'
    gui_dir=study/'sage-frames-gui-v1';gui=read(gui_dir/'receipt.json')
    assert gui['status']=='PASS_NORMAL_SAGE_HANDS_HD_GUI' and gui['binary_sha256']==sha(binary) and gui['script_sha256']==sha(p/'tools/hd_sage_smoke.py')
    assert gui['normal_DOS_exit']==0 and gui['source_save_unchanged'] and len(gui['pairs'])==2
    assert gui['visible_poses']==[0,3,4,5,6,7]
    shots={row['name']:row for row in gui['shots']};restored=0;original_pixels=0;hd_pixels=0
    palette=(p/'workplace/music-release-v1/release-build/data/assets/town/frame/palette.bin').read_bytes()
    for pair in gui['pairs']:
        images=[];states=[]
        for kind in ('original','hd','returned'):
            path=gui_dir/'boot1'/pair[kind];assert sha(path)==shots[path.stem]['image_sha256']
            im=Image.open(path).convert('RGB');assert im.size==(960,680);images.append(im);states.append(shots[path.stem]['state'])
        for field in ('steps','cycles','oracle_memory_sha256','oracle_indexed_sha256','oracle_palette_sha256'):
            assert states[0][field]==states[1][field]==states[2][field]
        assert ImageChops.difference(images[0].crop((0,48,960,648)),images[2].crop((0,48,960,648))).getbbox() is None
        restored+=576000
        pose,phase,animation=pair['pose'],pair['phase'],pair['animation_tile'];x,y=pair['native_origin']
        key=f'pose-{pose}-animation-'+('table' if animation<0 else str(animation))
        indexed=(study/'sage-frames-layout-v1'/(key+'.indexed')).read_bytes()
        expected=Image.frombytes('P',(96,64),indexed);expected.putpalette(palette)
        assert ImageChops.difference(images[0].crop((x*3,y*3+48,(x+96)*3,(y+64)*3+48)),expected.convert('RGB').resize((288,192),Image.Resampling.NEAREST)).getbbox() is None
        art=Image.open(study/'sage-frames-complete-v2'/f'sage-pose-{pose}-phase-{phase}.png').convert('RGB')
        assert ImageChops.difference(images[1].crop(((x+1)*3,(y+1)*3+48,(x+95)*3,(y+63)*3+48)),art.crop((3,3,285,189))).getbbox() is None
        original_pixels+=55296;hd_pixels+=52452
    assert restored==1152000 and original_pixels==110592 and hd_pixels==104904
    outside=shots['outside-after-power']['state'];assert not outside['sage_frames_hd']['matched'] and outside['town_scenery_hd_matched']
    files=['cmd/zeliard-preview/main.go','internal/prototype/hd/sage_frames.go','internal/prototype/hd/sage_frames_test.go',
           'tools/hd_sage_sources.py','tools/hd_sage_register.py','tools/hd_sage_probe.go','tools/ida_sage_frames.py',
           'tools/hd_sage_smoke.py','tools/hd_sage_audit.py','tools/sage-hd.commands','tools/hd_catalog.py','docs/sage-frame-source-index.json']
    proofs=['sage-frames-layout-v1/receipt.json','sage-frames-ida-v4/sage-frames-evidence.json','sage-frames-normal-v4/receipt.json',
            'sage-frames-normal-v5/receipt.json','sage-frames-complete-v2/manifest.json','sage-frames-art-v1/generation-v3.json',
            'sage-frames-runtime-v1/tests.jsonl','sage-frames-runtime-v1/normal.json','sage-frames-runtime-v1/render.json',
            'sage-frames-regression-v1/tests.jsonl','sage-frames-gui-v1/receipt.json']
    report={'schema':1,'status':'PASS_SAGE_SOURCE_ART_NORMAL_GUI','binary_sha256':sha(binary),'source_tiles':166,'hand_poses':8,
            'registered_images':16,'structural_variants':38,'direct_render_pixels':direct_pixels,'normal_points':24,
            'normal_source_frames':16,'normal_mixed_rejections':3,'gui_visible_poses':gui['visible_poses'],
            'gui_original_pixels':original_pixels,'gui_hd_pixels':hd_pixels,'restored_game_pixels':restored,'normal_DOS_exit':0,
            'replay':{'indexed_equal_frames':19,'ram_difference_physical_offsets':['0x257f','0x2580'],
                      'mtime_packed_word_equal':True,'level':'兩byte與scratch副本DOS時間欄相同；未宣稱跨兩次來源擷取的完整RAM一致'},
            'source_files':[{'path':name,'sha256':sha(p/name)} for name in files],
            'proofs':[{'path':str((study/name).relative_to(p)),'sha256':sha(study/name)} for name in proofs],
            'limits':['只驗Muralla正常查看力量八手勢及代表GUI，不外推置中手勢或所有城鎮正常路徑。',
                      '38結構組合不代表每種均會出現；繪製中混合畫格保持原子拒絕。',
                      '手繪形狀、紋理與動畫光點屬美術差異；原版時序、規則、輸入及存檔不改。',
                      '音訊關閉；未更新三平台封包、影片或README縮圖；全遊戲HD仍進行中。']}
    assert not args.output.exists()
    args.output.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    print(report['status'],direct_pixels,len(normal['points']),sha(binary))


if __name__=='__main__':main()

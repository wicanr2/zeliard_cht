#!/usr/bin/env python3
"""核魔法石來源及高清候選；只給READY，不宣稱正式HD已接通。"""
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
    assert not args.output.exists(),'不能覆寫舊稽核'
    sha=lambda path:hashlib.sha256(path.read_bytes()).hexdigest()
    read=lambda path:json.loads(path.read_text())
    source_path=b/'cave-magia-source-v3/receipt.json';source=read(source_path)
    assert source['status']=='OBSERVED_REJECTED_CAVE_SOURCES' and source['state_equal'] and source['save_unchanged'] and not source['state_injection']
    assert source['script_sha256']==sha(p/'tools/hd_cave_rejected_probe.go')==sha(source_path.parent/'probe-source.go')
    assert source['commands_sha256']==sha(p/'tools/cave-platform-magia.commands')
    assert len(source['points'])==27
    assert sum(row['state_equal']for row in source['points'])==3
    for row in source['points']:
        for suffix,key in [('memory','memory_sha256'),('body.memory','body_memory_sha256'),('indexed','indexed_sha256'),('palette','palette_sha256')]:
            path=source_path.parent/(row['file']+'.'+suffix);assert sha(path)==row[key]
        for suffix in ['memory','body.memory']:
            memory=(source_path.parent/(row['file']+'.'+suffix)).read_bytes()
            assert hashlib.sha256(memory[0x11b10+0x8c00:0x11b10+0x9350]).hexdigest()=='7b4628728e0b6ce81a25b7bc0bbb3b4eedd640b0da102924b4384da0d38f690b'
    diff_path=b/'cave-magia-diff-v5/receipt.json';diff=read(diff_path)
    assert diff['status']=='DRAFT_REJECTED_DIFFERENCE_STUDY' and len(diff['points'])==27
    assert diff['study_source_sha256']==sha(p/'internal/prototype/hd/cave_rejected_study_test.go')==sha(diff_path.parent/'study-source.go')
    assert all(row['candidate_mapped_difference_count']==0 for row in diff['points'])
    placements=[v for row in diff['points']for v in row['candidate_placements']]
    assert all(v['visibility']in (0,1)for v in placements)
    cleared=sum(v['visibility']==0 for v in placements);assert cleared>=4
    phases=sorted({v['phase']for v in placements});assert len(phases)==11
    assert not (diff_path.parent/'vet.log').read_bytes()
    ida_path=b/'cave-magia-ida-v9/evidence.json';ida=read(ida_path)
    assert ida['tool']=='IDA 9.4' and ida['input_sha256']=='5e7f76080cc542eb06ca90ec349ad2ec0c5ab374c32acf59bc4c4307899a0a5b'
    assert ida['script_sha256']==sha(p/'tools/ida_cave_magia.py')
    assert ida['source_index_sha256']==sha(p/'docs/cave-magia-source-index.json')
    assert len(ida['windows'])==6 and len(ida['tables'])==2
    instructions={row['address']:row for w in ida['windows']for row in w['instructions']}
    for address in ['0x8679','0x867c','0x86cf','0x86d6','0x86dc','0x412f','0x713e','0x7147']:
        assert instructions[address]['semantic_index']['level']=='已證實' and not instructions[address]['semantic_index']['warning']
    assert next(row for row in ida['tables']if row['address']=='0x3008')['bytes']=='2f41ad44'
    gen_path=b/'cave-magia-art-v1/generation.json';gen=read(gen_path)
    assert gen['mode']=='built-in imagegen' and gen['alpha_range']==[255,255] and not gen['production_runtime_changed']
    assert sha(p/gen['output_file'])==gen['output_sha256']
    request_path=gen_path.parent/gen['request_file'];assert sha(request_path)==gen['request_sha256'] and read(request_path)['prompt']==gen['prompt']
    for ref in gen['reference_inputs']:assert sha(p/ref['file'])==ref['sha256']
    reg_path=b/'cave-magia-complete-v1/manifest.json';reg=read(reg_path)
    assert reg['script_sha256']==sha(p/'tools/hd_cave_magia_art.py')
    assert reg['generation_sha256']==sha(gen_path) and reg['source_id']==102
    assert reg['opaque_pixels']==414 and reg['transparent_inside_mask']==reg['outside_mask_written_pixels']==0
    art_path=reg_path.parent/reg['art_file'];assert sha(art_path)==reg['art_sha256']
    art=np.asarray(Image.open(art_path).convert('RGBA'));assert art.shape==(24,24,4)
    assert np.count_nonzero(art[:,:,3])==414 and set(np.unique(art[:,:,3]))=={0,255}
    files=['tools/hd_cave_rejected_probe.go','internal/prototype/hd/cave_rejected_study_test.go',
           'tools/ida_cave_magia.py','docs/cave-magia-source-index.json','tools/hd_cave_magia_art.py',
           'tools/hd_cave_magia_audit.py','tools/hd_catalog.py']
    proofs=[source_path,diff_path,diff_path.parent/'tests.jsonl',ida_path,gen_path,request_path,reg_path,art_path]
    report={'schema':1,'status':'READY_MAGIA_SOURCE_AND_ART_NOT_CONNECTED','spec':'docs/spec/044-cave-magia-presentation.md',
            'source_points':27,'command_boundary_state_equal_points':3,'sampled_phases':phases,
            'original_viewport_pixels_reconstructed':27*224*144,'cleared_record_observations':cleared,
            'source_id':102,'opaque_hd_pixels':414,'production_runtime_changed':False,
            'source_files':[{'path':f,'sha256':sha(p/f)}for f in files],
            'proofs':[{'path':str(f.relative_to(p)),'sha256':sha(f)}for f in proofs],
            'limits':['DRAFT研究的27張原版重建，不當27點完整機器配對或正式呈現器驗收。',
                      '102圖面已註冊，正式圖層、來源守衛、正常HD配對及GUI尚未實作。',
                      '部分遮擋、耗盡、邊緣及其他魔法仍待補；全遊戲目標維持進行中。']}
    args.output.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n');print(report['status'],27,'sources',414,'HD pixels')


if __name__=='__main__':main()

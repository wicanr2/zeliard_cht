#!/usr/bin/env python3
"""Docker內核對背向圖面、正常原版配對及GUI證據；未捕捉不計通過。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageChops


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project',type=Path,required=True)
    ap.add_argument('--output',type=Path,required=True)
    args=ap.parse_args(); project=args.project; study=project/'workplace/full-hd-v1'
    sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
    read=lambda p:json.loads(p.read_text())
    binary=study/'town-back-runtime-v4/zeliard-preview'
    units=study/'town-back-runtime-v2/source-art-live-tests.jsonl'
    regression=study/'town-back-regression-v1/test.jsonl'
    tests=[json.loads(line) for p in (units,regression) for line in p.read_text().splitlines()]
    names=['TestTownBackNormalSource','TestTownBackCompositeAndFallback','TestTownBackLiveEntryReadOnly',
           'TestCastleNormalPathsReadOnly','TestCastleNormalSourcesAndLayers','TestCastleRefusalIsAtomic',
           'TestCastlePresentationCrossProfile','TestCastlePlayerHiddenAtRightBoundary','TestTownActorsNormalPathsReadOnly',
           'TestTownActorsAllSources','TestTownActorsAtomicFallback','TestTownActorsLayerOrder','TestTownActorsMissingAssets']
    assert not any(x['Action'] in ('fail','skip') for x in tests)
    assert all(any(x['Action']=='pass' and x.get('Test')==name for x in tests) for name in names)
    live=read(study/'town-back-runtime-v2/live-entry.json')
    assert live['status']=='PASS_NORMAL_BACK_ENTRY_READ_ONLY'
    assert [case['bit'] for case in live['cases']]==[0,1]
    assert all(case['full_state_equal'] and case['samples'] for case in live['cases'])
    native=read(study/'town-back-source-audit-v2/native-check.json')
    assert native['status']=='PASS_NORMAL_BACK_SOURCE_PIXELS'
    assert native['script_sha256']==sha(project/'tools/hd_town_back_art.py')
    assert sum(x['original_pixels'] for x in native['cases'])==28672
    assert sum(x['independent_hd_pixels'] for x in native['cases'])==6912
    assert all(x['max_channel_deviation']<=x['hd_tolerance']==2 for x in native['cases'])
    registration=study/'town-neutral-complete-v1/town-player-back-v1.png'
    replay=study/'town-back-art-replay-v1/town-player-back-v1.png'
    assert sha(registration)==sha(replay)=='d84525e69589decf0045cc32b35032eaea862c967e9876df5294cf208d3096f4'
    ida=read(study/'town-back-ida-v4/evidence.json')
    assert ida['tool']=='IDA 9.4' and ida['source_index_sha256']==sha(project/'docs/town-back-source-index.json')
    assert ida['script_sha256']==sha(project/'tools/ida_town_back_evidence.py')
    assert len(ida['functions'])>=4
    normal_points=0
    for folder in ('castle','muralla'):
        normal=read(study/'town-back-regression-v1'/folder/'receipt.json')
        assert not normal['uncovered'] and not normal['held_keys']
        assert all(x['state_equal'] for x in normal['checkpoints'])
        normal_points+=len(normal['checkpoints'])
    assert normal_points==19
    gui=study/'town-back-gui-v3'; gui_verified=False; restored=0
    if (gui/'receipt.json').exists():
        evidence=read(gui/'receipt.json')
        assert evidence['status']=='PASS_TOWN_BACK_HD_GUI' and evidence['binary_sha256']==sha(binary)
        assert evidence['script_sha256']==sha(project/'tools/hd_town_back_smoke.py')
        assert len(evidence['pairs'])==2
        for pair in evidence['pairs']:
            images=[Image.open(gui/('boot'+str(pair['boot']))/pair[key]).convert('RGB') for key in ('original','hd','returned')]
            assert all(im.size==(960,680) for im in images)
            assert ImageChops.difference(images[0].crop((0,48,960,648)),images[2].crop((0,48,960,648))).getbbox() is None
            assert ImageChops.difference(images[0],images[1]).getbbox() is not None
            assert pair['player_phase']==4 and pair['residents']==0
            restored+=576000
        gui_verified=True
    else:
        assert (gui/'failure.json').exists(), 'GUI仍在進行或尚無結果'
    sources=['cmd/zeliard-preview/main.go','internal/prototype/hd/town_player.go','internal/prototype/hd/town_scenery.go',
             'internal/prototype/hd/town_actors.go','internal/prototype/hd/town_player_back_test.go',
             'tools/hd_town_back_art.py','tools/ida_town_back_evidence.py','tools/hd_town_back_smoke.py',
             'tools/hd_town_back_audit.py','docs/town-back-source-index.json']
    proofs=[units,regression,study/'town-back-runtime-v2/live-entry.json',study/'town-back-source-audit-v2/native-check.json',
            study/'town-back-ida-v4/evidence.json',study/'town-back-art-replay-v1/manifest.json',
            gui/('receipt.json' if gui_verified else 'failure.json')]
    report={'schema':1,'status':'PASS_TOWN_BACK_HD_SCOPED' if gui_verified else 'SOURCE_ART_NORMAL_PASSED_GUI_PENDING',
            'binary_sha256':sha(binary),'normal_points':normal_points,'tests':names,'native_pixels':28672,
            'independent_hd_pixels':6912,'hd_channel_tolerance':2,'art_replay_equal':True,
            'gui_verified':gui_verified,'restored_gui_pixels':restored,
            'source_files':[{'path':p,'sha256':sha(project/p)} for p in sources],
            'proofs':[{'path':str(p.relative_to(project)),'sha256':sha(p)} for p in proofs],
            'limits':['正常原版完成回呼的HD合成已驗，不能代替實際視窗的短暫背向姿態。',
                      '未取得背向GUI畫格時規格038保持READY，不為擷取改變原版時序。',
                      '其他城鎮、交談背景、敵人／首領及介面等仍待補，三平台封包未更新。']}
    assert not args.output.exists()
    args.output.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    print(report['status'],sha(binary),normal_points)


if __name__=='__main__':
    main()

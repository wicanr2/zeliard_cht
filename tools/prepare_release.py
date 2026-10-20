"""Docker內固定來源並整理三平台本機完整素材包。"""
from pathlib import Path
import difflib, hashlib, json, os, re, shutil, sys

R=Path('/project'); O=R/'workplace/music-release-v1'; version=(R/'VERSION').read_text().strip()
assert re.fullmatch(r'v\.\d+\.\d+\.\d+-\d{8}',version)
h=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()

if sys.argv[1]=='freeze':
    build=O/'release-build'; assert not build.exists(), '已存在固定建置，拒絕覆寫'
    build.mkdir()
    for sub in ['cmd','internal','text','third_party']:
        shutil.copytree(R/sub,build/'sources'/sub)
    for n in ['go.mod','go.sum','VERSION','LICENSE']:
        shutil.copy2(R/n,build/'sources'/n)
    golem=build/'dosgolem';shutil.copytree(R/'workplace/dosgolem-source',golem)
    f=golem/'oracle/live.go';before=f.read_text();after=before.replace('synth    *opl2.Synth','synth    FMSynth')
    marker='// Rate 回取樣率。'
    assert marker in after and 'synth    FMSynth' in after
    after=after.replace(marker,'''// FMSynth 是作品中立的輸出音源，不持有或改寫機器。
type FMSynth interface { Write(uint8, uint8); Sample() float64 }
// SetFMSynth 只更換輸出合成器，須在首次 Render 前呼叫。
func (a *Audio) SetFMSynth(s FMSynth) { a.synth = s }

'''+marker)
    f.write_text(after)
    patch=''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/oracle/live.go',tofile='b/oracle/live.go'))
    (R/'tools/dosgolem-fm.patch').write_text(patch)
    (build/'build.work').write_text('go 1.24.0\nuse (\n /build/sources\n /build/dosgolem\n)\n')
    files={str(p.relative_to(build/'sources')):h(p) for p in (build/'sources').rglob('*') if p.is_file()}
    (build/'identity.json').write_text(json.dumps({'version':version,'sources':files,'dosgolem_original_live_sha256':hashlib.sha256(before.encode()).hexdigest(),'dosgolem_patched_live_sha256':h(f),'generic_patch_sha256':h(R/'tools/dosgolem-fm.patch'),'audio031_promoted':True},indent=2)+'\n')
    print('FROZEN',version,len(files))

elif sys.argv[1]=='assets':
    data=O/'release-build/data';assert not data.exists()
    prior=R/'dist-all/v.0.1.0-20261020/full-local'
    shutil.copytree(prior,data,ignore=shutil.ignore_patterns('zeliard-preview','userdata','play.sh','start_game.py','README.md'))
    args=json.loads((data/'runtime-args.json').read_text())
    def add(option,source,target):
        assert source.is_file(),source
        dest=data/target;dest.parent.mkdir(parents=True,exist_ok=True);shutil.copy2(source,dest)
        args.append([option,target])
    W=R/'workplace'
    for flag,source in [('king-reference-a','probe-king-memory-84m/screen.bin'),('king-reference-b','probe-king-pose-96m/screen.bin'),('king-art','hd-king-prototype.png'),('king-speaking-art','hd-king-speaking-prototype.png')]:
        add('-'+flag,W/source,'assets/king/'+source)
    for name in ['probe-king-memory-84m','probe-king-pose-96m']:
        shutil.copy2(W/name/'palette.bin',data/'assets/king'/name/'palette.bin')
    M=W/'merchant-hd-gui-v2-tools'
    for flag,name in [('merchant-reference-a','greeting.indexed'),('merchant-reference-b','farewell.indexed'),('merchant-reference-c','no-purchase-farewell.indexed'),('merchant-palette','greeting.palette'),('merchant-art-a','merchant-prototype.png'),('merchant-art-b','merchant-farewell-prototype.png'),('merchant-art-c','merchant-no-purchase-prototype.png')]:
        add('-'+flag,M/name,'assets/merchant/'+name)
    bank=W/'bank-hd-v1/gui-tools-v2'
    for flag,name in [('bank-reference-a','bank-farewell.indexed'),('bank-reference-b','bank-farewell-ready.indexed'),('bank-reference-c','bank-greeting.indexed'),('bank-palette','bank-greeting.palette'),('bank-art-a','bank-bent-prototype.png'),('bank-art-b','bank-lowered-prototype.png'),('bank-art-c','bank-raised-prototype.png')]:
        add('-'+flag,bank/name,'assets/bank/'+name)
    item=W/'item-hd-v1/gui-tools-v1'
    for i,(ref,art) in enumerate(zip(['item-greeting.indexed','item-list.indexed','item-farewell.indexed','item-farewell-ready.indexed'],['item-shop-base-prototype.png','item-shop-steam-b-prototype.png','item-shop-arms-prototype-v3.png','item-shop-steam-c-prototype.png'])):
        add('-item-reference-'+str(i),item/ref,'assets/item/'+ref)
        add('-item-art-'+str(i),item/art,'assets/item/'+art)
    add('-item-palette',item/'item-greeting.palette','assets/item/item-greeting.palette')
    for _,path in args:assert (data/path).exists(),path
    (data/'runtime-args.json').write_text(json.dumps(args,ensure_ascii=False,indent=2)+'\n')
    (data/'starter').mkdir();shutil.copy2(prior/'userdata/game/SAVETEST.usr',data/'starter/SAVETEST.usr')
    shutil.copy2(R/'VERSION',data/'VERSION');shutil.copytree(R/'third_party',data/'third_party')
    (data/'THIRD-PARTY-NOTICES.txt').write_text('Nuked-OPL3 ©2013–2024 Alexey Khokholov，LGPL-2.1-or-later。來源與條款見 third_party/nuked-opl3。音源為可替換動態函式庫，允許替換所需的逆向與除錯。\n原版遊戲、字型與HD研究素材僅供本機，不屬自有程式授權。\n',encoding='utf-8')
    print('ASSETS',len(args))
else:
    raise SystemExit('freeze 或 assets')

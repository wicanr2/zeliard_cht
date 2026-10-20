"""在 Docker 內建立固定來源的本機試玩包；原版素材不得散布。"""
from pathlib import Path
import hashlib
import json
import os
import re
import shutil
import sys

R = Path('/project')
version = (R / 'VERSION').read_text().strip()
if not re.fullmatch(r'v\.\d+\.\d+\.\d+-\d{8}', version):
    raise SystemExit('VERSION 格式錯誤')
O = R / 'workplace/input-integration-v1/delivery-build'
D = R / 'dist-all' / version
B = D / 'full-local'
digest = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()

if sys.argv[1] == 'prepare':
    if O.exists() or D.exists():
        raise SystemExit('此版號或建置目錄已存在，拒絕覆寫。')
    O.mkdir()
    B.mkdir(parents=True)
    (D / 'smoke').mkdir()
    stable = R / 'workplace/prose-v1/services-v1/hd-candidate-v3/candidate-sources'
    files = [R / 'go.mod', R / 'go.sum']
    for sub in ['cmd', 'internal']:
        files += list((R / sub).rglob('*.go'))
    files += [p for p in (R / 'text').iterdir() if p.is_file()]
    sources, checkout = {}, {}
    for p in files:
        relative = str(p.relative_to(R))
        checkout[relative] = digest(p)
        selected = stable / relative if relative in ['cmd/zeliard-preview/audio.go', 'internal/prototype/sound/stream.go'] else p
        out = O / 'sources' / relative
        out.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(selected, out)
        sources[relative] = digest(out)
    (O / 'identity.json').write_text(json.dumps({'version': version, 'sources': sources, 'checkout_sources': checkout, 'audio031_promoted': False}, indent=2) + '\n')
    (O / 'build.work').write_text('go 1.24.0\n\nuse (\n /dosgolem\n /build/sources\n)\n')
    args = []

    def include(option, source, target):
        assert source.exists(), source
        dest = B / target
        dest.parent.mkdir(parents=True, exist_ok=True)
        if not dest.exists():
            if source.is_dir():
                shutil.copytree(source, dest)
            else:
                shutil.copy2(source, dest)
        args.append([option, target])

    shutil.copytree(R / 'workplace/orig', B / 'original')
    shutil.copytree(O / 'sources/text', B / 'text')
    shutil.copy2('/fonts/unifont.hex.gz', B / 'font.hex.gz')
    for option, name in [('text', 'host.json'), ('dialogue-text', 'dialogue-prototype.tsv'), ('labels-text', 'labels-prototype.tsv'), ('hud-text', 'hud-prototype.tsv'), ('town-text', 'town-prototype.tsv'), ('town-signs-text', 'town-signs-prototype.tsv'), ('intro-scroll', 'intro-scroll-prototype.tsv'), ('intro-credits', 'intro-credits-prototype.tsv'), ('intro-outro', 'intro-outro-prototype.tsv'), ('intro-threat', 'intro-threat-prototype.tsv'), ('intro-story', 'intro-story-prototype.tsv')]:
        args.append(['-' + option, 'text/' + name])
    T = R / 'workplace/prose-v1/shield-transition-tools-v1'
    for option, name in [('hd-reference', 'frame/screen.bin'), ('hd-art', 'hd-border-prototype.png'), ('town-sky-reference', 'sky-reference.indexed'), ('town-sky-palette', 'sky-reference.palette'), ('town-background-memory', 'restored.memory'), ('town-background-art', 'background-handpaint-v2.png'), ('town-scenery-art-dir', 'scenery-art'), ('town-ground-art', 'ground-handpaint-v1.png'), ('town-player-art-dir', 'player-art'), ('sage-reference-a', 'sage-a.indexed'), ('sage-reference-b', 'sage-b.indexed'), ('sage-palette', 'sage.palette'), ('sage-art', 'sage-handpaint.png')]:
        include('-' + option, T / name, 'assets/town/' + name)
    shutil.copy2(T / 'frame/palette.bin', B / 'assets/town/frame/palette.bin')
    cave = B / 'assets/cave'
    cave.mkdir()
    for path in (R / 'internal/prototype/hd').glob('cave_*.go'):
        if path.name.endswith('_test.go'):
            continue
        for name in re.findall(r'"([^"/]+\.png)"', path.read_text()):
            source = R / 'workplace/cave-hd-v1' / name
            if source.is_file() and not (cave / name).exists():
                shutil.copy2(source, cave / name)
    for option in ['cave-art-dir', 'cave-player-art-dir', 'cave-weapon-art-dir', 'cave-object-art-dir']:
        args.append(['-' + option, 'assets/cave'])
    shutil.copy2(R / 'workplace/cave-hd-v1/portal-handpaint-v1.png', cave / 'portal-handpaint-v1.png')
    args.append(['-cave-portal-art', 'assets/cave/portal-handpaint-v1.png'])
    C = R / 'workplace/prose-v1/services-v1'
    include('-church-reference-dir', C / 'hd-references-v2', 'assets/church/reference')
    art = B / 'assets/church/art'
    art.mkdir()
    for name in re.findall(r'"([^"/]+\.png)"', (R / 'internal/prototype/hd/church.go').read_text()):
        if (C / name).is_file() and not (art / name).exists():
            shutil.copy2(C / name, art / name)
    args.append(['-church-art-dir', 'assets/church/art'])
    for _, relative in args:
        assert (B / relative).exists(), relative
    for name in ['start_game.py', 'run-local.sh']:
        shutil.copy2(R / 'tools' / name, B / ('play.sh' if name == 'run-local.sh' else name))
    shutil.copy2(R / 'LICENSE', B / 'LICENSE')
    shutil.copy2(R / 'VERSION', B / 'VERSION')
    (B / 'runtime-args.json').write_text(json.dumps(args, ensure_ascii=False, indent=2) + '\n')
    user = B / 'userdata/game'
    user.mkdir(parents=True)
    shutil.copy2(R / 'workplace/prose-v1/npc-normal-v2/scratch/SAVETEST.usr', user / 'SAVETEST.usr')
    (B / 'README.md').write_text('# 神劍封魔：第一洞窟試玩版\n\n執行 `bash play.sh`。需要 Docker 及 Linux 桌面。\n\n- 點工具列「HD 畫面」開啟高清。\n- 工具列可設定鍵盤與搖桿／手把；鍵盤預設 F12 開啟。\n- 方向鍵移動，空白鍵攻擊／交談，Enter 開關裝備，Alt 施法，Esc 原版暫停。\n- F7 讀取 SAVETEST 可回到Muralla鎮，向右走到洞口，按上進入第一洞窟。\n- 存檔與設定保存於 userdata；Ctrl+Q、Y 正常離開。\n\n本機包含原版資料、使用者字型與研究素材，不得上傳或散布。只驗第一洞窟與代表性中文／HD，不宣稱完整通關或全遊戲素材完成。USB／藍牙配對由作業系統處理。\n', encoding='utf-8')
    print('PREPARED_LOCAL_TRIAL', version, len(sources))
elif sys.argv[1] == 'finalize':
    assert (B / 'zeliard-preview').is_file()
    for relative in ['play.sh', 'zeliard-preview']:
        os.chmod(B / relative, 0o755)
    files = {str(p.relative_to(B)): {'bytes': p.stat().st_size, 'sha256': digest(p)} for p in B.rglob('*') if p.is_file() and 'userdata' not in p.relative_to(B).parts}
    report = {'version': version, 'platform': 'linux', 'architecture': 'amd64', 'rights': 'full-local; 原版資料、字型與圖面只限本機，不進Git或Release', 'files': files}
    (D / 'SHA256SUMS.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
    print('MANIFEST_LOCAL_TRIAL', version, len(files))
else:
    raise SystemExit('須指定 prepare 或 finalize')

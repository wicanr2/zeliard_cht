"""064：在Docker內建立借用既有唯讀素材的字級候選，不改正式包。"""
from pathlib import Path
import hashlib
import json
import os
import shutil

P = Path('/project')
W = P / 'workplace/ui-text-v1'
assert P.is_dir() and W.is_dir() and W.stat().st_uid == os.getuid()
base = P / 'workplace/cpu-speed-v1/preview'
preview = W / 'preview'
preview.mkdir(exist_ok=True)
for name in ('original', 'assets', 'starter', 'lib', 'libzopl.so', 'runtime-args.json', 'bundle-options.json'):
    source = (base / name).resolve(strict=True)
    target = preview / name
    assert not target.exists() or target.is_symlink(), target
    if target.is_symlink():
        target.unlink()
    target.symlink_to(os.path.relpath(source, preview), target_is_directory=source.is_dir())
text = preview / 'text'
text.mkdir(exist_ok=True)
for source in (P / 'text').iterdir():
    if source.is_file():
        target = text / source.name
        assert not target.exists() or target.stat().st_uid == os.getuid()
        shutil.copyfile(source, target)
for name in ('font.hex.gz', 'font.hex.gz.smooth.json.gz'):
    target = preview / name
    assert not target.exists() or target.stat().st_uid == os.getuid()
    shutil.copyfile(W / 'rounded-font' / name, target)
(preview / 'VERSION').write_text('development-ui-text-v1\n')
inputs = [P / 'text/host.json', preview / 'font.hex.gz', preview / 'font.hex.gz.smooth.json.gz', preview / 'original/ZELIARD.EXE']
(W / 'candidate-inputs.json').write_text(json.dumps({
    'identity': 'development-ui-text-v1',
    'base_code': 'v.1.0.4-20261022 plus 064',
    'art_source': 'unchanged local art reused through CPU preview; no official package overwritten',
    'inputs': [{'path': str(p.relative_to(P)), 'sha256': hashlib.sha256(p.read_bytes()).hexdigest()} for p in inputs],
}, ensure_ascii=False, indent=2) + '\n')
print('PREPARED_UI_TEXT_CANDIDATE')

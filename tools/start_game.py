"""Docker 容器內的本機試玩入口。"""
from pathlib import Path
import json
import os
import sys

root = Path('/gameapp')
args = json.loads((root / 'runtime-args.json').read_text())
command = [str(root / 'zeliard-preview'), '-orig', str(root / 'original'),
           '-font', str(root / 'font.hex.gz'), '-user-dir', '/user']
for option, relative in args:
    path = root / relative
    if not path.exists():
        raise SystemExit(f'缺少遊戲資料：{relative}')
    command += [option, str(path)]
os.execv(command[0], command + sys.argv[1:])

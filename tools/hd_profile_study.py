#!/usr/bin/env python3
"""只建立Go研究編譯覆蓋；不改正式來源、原版機器或遊戲資料。"""
import argparse
import hashlib
import json
from pathlib import Path

ap = argparse.ArgumentParser(description=__doc__)
ap.add_argument("--project", type=Path, required=True)
ap.add_argument("--output", type=Path, required=True)
args = ap.parse_args()
assert args.output.is_dir() and not any(args.output.iterdir())
source = args.project / "internal/prototype/hd/cave_sources.go"
raw = source.read_bytes()
text = raw.decode()
changes = [
    ("inactiveFF41 := bp[0xff41] == 1 && bp[0xff40] == 0",
     "inactiveFF41 := bp[0xff41] == 1 && (bp[0xff40] == 0 || bp[0xff40] == 255)"),
    ("shieldTransition := shield && bp[0xff3a] == 255 && bp[0xe8] == 0 && bp[0xff39] == 0 && bp[0xff38] == 0 && bp[0xff40] == 0",
     "shieldTransition := shield && (bp[0xff3a] == 255 || bp[0xff39] == 255) && bp[0xe8] == 0 && bp[0xff38] == 0 && bp[0xff40] == 0"),
    ("back = 0x61b9 + 0x24 + offset", "back = 0x61b9 + 0x24 + offset\n\t\t\tif bp[0xff41] == 1 { back += 0x24 }"),
    ("front = 0x6117 + 0x24 + offset", "front = 0x6117 + 0x24 + offset\n\t\t\tif bp[0xff41] == 1 { front += 0x24 }"),
]
for before, after in changes:
    assert text.count(before) == 1, before
    text = text.replace(before, after)
replacement = args.output / "prototype-source.go"
replacement.write_text(text)
overlay = {"Replace": {str(source): str(replacement)}}
(args.output / "compiler-overlay.json").write_text(json.dumps(overlay, indent=2) + "\n")
report = {"status": "DRAFT_COMPILER_OVERLAY_ONLY", "source_sha256": hashlib.sha256(raw).hexdigest(),
          "prototype_sha256": hashlib.sha256(replacement.read_bytes()).hexdigest(),
          "changes": [{"before": a, "after": b} for a, b in changes],
          "limits": "只用Go -overlay研究，正式來源不變；沒有改原版記憶體或將研究decoder接入前端。"}
(args.output / "prototype.json").write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n")
print(report["status"])

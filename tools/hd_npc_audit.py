#!/usr/bin/env python3
"""Docker內稽核規格035的正常視窗、原貌切回與原版配對收據。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageChops


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    p = args.project
    study = p/"workplace/full-hd-v1"
    sha = lambda b: hashlib.sha256(b).hexdigest()
    gui_path = study/"town-npc-gui-v1/receipt.json"
    pair_path = study/"town-npc-pair-v2/receipt.json"
    gui, pair = (json.loads(f.read_text()) for f in (gui_path, pair_path))
    assert gui["status"] == "PASS_RESIDENT_HD_GUI_SCOPED"
    assert gui["normal_DOS_exit"] == 0 and gui["save_unchanged"]
    binary = study/"town-npc-runtime-v1/zeliard-preview"
    assert sha(binary.read_bytes()) == gui["binary_sha256"]
    assert sha((p/"tools/hd_npc_smoke.py").read_bytes()) == gui["script_sha256"]
    assert pair["uncovered"] == 0 and not pair["held_keys"]
    assert len(pair["checkpoints"]) == 6
    assert all(x["state_equal"] for x in pair["checkpoints"])
    assert sum(bool(x["residents_hd_matched"]) for x in pair["checkpoints"]) == 4
    frozen = ("steps", "cycles", "oracle_memory_sha256", "oracle_indexed_sha256", "oracle_palette_sha256")
    shots = {x["name"]: x for x in gui["shots"]}
    comparisons = []
    for group in gui["pairs"]:
        states, images = [], []
        for key in ("original", "hd", "returned"):
            path = gui_path.parent/"boot1"/group[key]
            shot = shots[path.stem]
            assert sha(path.read_bytes()) == shot["image_sha256"]
            states.append({k: shot["state"][k] for k in frozen})
            images.append(Image.open(path).convert("RGB"))
        assert states[0] == states[1] == states[2] and group["state_frozen"]
        assert all(im.size == (960, 680) for im in images)
        crop = (0, 48, 960, 648)
        assert ImageChops.difference(images[0].crop(crop), images[2].crop(crop)).getbbox() is None
        assert ImageChops.difference(images[0].crop(crop), images[1].crop(crop)).getbbox()
        assert group["residents"]["matched"] and group["residents"]["residents"] == 1
        comparisons.append({"name": group["name"], "game_pixels_restored": 576000,
                            "native_state_equal": True, "resident_keys": group["residents"]["keys"]})
    assert len(comparisons) == 4 and len(gui["conversations"]) == 2
    sources = ["cmd/zeliard-preview/main.go", "internal/prototype/hd/town_actors.go",
               "internal/prototype/hd/town_npc_assets.go", "internal/prototype/hd/town_actors_test.go",
               "internal/prototype/hd/town_actors_pair_test.go", "internal/prototype/hd/town_frames_pair_test.go",
               "tools/hd_compile_npcs.py", "tools/hd_npc_smoke.py", "tools/hd_npc_audit.py",
               "tools/town-residents-hd.commands"]
    report = {"schema": 1, "status": "PASS_MURALLA_RESIDENTS_SCOPED", "spec": "docs/spec/035-town-residents-hd.md",
              "binary_sha256": gui["binary_sha256"], "pairs": comparisons,
              "original_state_checkpoints": 6, "source_files": [{"path": f, "sha256": sha((p/f).read_bytes())} for f in sources],
              "proofs": [{"path": str(f.relative_to(p)), "sha256": sha(f.read_bytes())} for f in (gui_path, pair_path)],
              "visual_review": {"reviewer": "Codex", "images": ["restored-hd.png", "walking-right-hd.png", "resident-dialogue-1.png"],
                                "result": "衛兵與居民在原位置，主角及地面完整；中文第二頁在對話框內。"},
              "limits": ["Muralla固定來源40項資產檢查，正常視窗只抽樣兩種居民姿態；不是全部城鎮驗收。",
                         "交談期間既有整區守衛退回原貌背景，其他主角姿態與場景仍有HD缺口。",
                         "未更新v.0.2.0封包；完整遊戲HD目標仍進行中。"]}
    assert not args.output.exists(), "preserve previous audit identity"
    args.output.write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
    print(report["status"], len(comparisons), "pairs")


if __name__ == "__main__":
    main()

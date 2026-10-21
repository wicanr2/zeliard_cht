#!/usr/bin/env python3
"""核FF41=1／FF40=0來源修正、原日誌、既有HD及正常視窗。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    p, b = args.project, args.project / "workplace/full-hd-v1"
    assert not args.output.exists(), "不得覆寫歷史稽核"
    sha = lambda f: hashlib.sha256(f.read_bytes()).hexdigest()
    read = lambda f: json.loads(f.read_text())
    before_path = b / "cave-ff41-source-v1/receipt.json"
    after_path = b / "cave-ff41-accepted-source-v1/receipt.json"
    before, after = read(before_path), read(after_path)
    assert before["status"] == "PASS_REJECTED_GUI_JOURNAL_SOURCE_RECONSTRUCTION"
    assert after["status"] == "PASS_PREVIOUSLY_REJECTED_GUI_SOURCE_NOW_PRESENTABLE"
    fields = ["records", "journal_sha256", "observation_sha256", "target_steps", "source_steps", "body_steps",
              "source_cycles", "body_cycles", "memory_sha256", "body_memory_sha256", "indexed_sha256", "palette_sha256"]
    assert all(before[f] == after[f] for f in fields)
    assert before["source_steps"] == 202440421 and before["target_steps"] == 202514118
    assert before["save_unchanged"] and after["save_unchanged"]
    proofs = [before_path, after_path]
    for receipt_path, receipt in [(before_path, before), (after_path, after)]:
        for ext, field in [("memory", "memory_sha256"), ("body.memory", "body_memory_sha256"),
                           ("indexed", "indexed_sha256"), ("palette", "palette_sha256")]:
            file = receipt_path.parent / ("source." + ext)
            assert sha(file) == receipt[field]
            proofs.append(file)
        script = receipt_path.parent / "probe-source.go"
        assert sha(script) == receipt["script_sha256"]
        proofs.append(script)
    index_path = p / "docs/cave-ff41-source-index.json"
    ida_path = b / "cave-ff41-ida-v5/evidence.json"
    index, ida = read(index_path), read(ida_path)
    assert ida["tool"] == "IDA 9.4" and ida["input_sha256"] == index["input_sha256"]
    assert ida["source_index_sha256"] == sha(index_path)
    assert ida["script_sha256"] == sha(p / "tools/ida_cave_ff41.py")
    assert len(index["entries"]) == 11
    assert sum(not r["semantic_index"]["warning"] for r in ida["instructions"]) == 11
    proofs.append(ida_path)
    tests = {}
    for file, expected in [(b / "cave-ff41-runtime-v1/unit-tests.jsonl", 44),
                           (b / "cave-ff41-runtime-v3/focused-tests.jsonl", 4),
                           (b / "cave-ff41-normal-pair-v2/tests.jsonl", 3)]:
        rows = [json.loads(line) for line in file.read_text().splitlines()]
        assert rows[-1]["Action"] == "pass" and not any(r["Action"] in ("skip", "fail") for r in rows)
        count = sum(r["Action"] == "pass" and "Test" in r and "/" not in r["Test"] for r in rows)
        assert count == expected, (file, count)
        tests[str(file.relative_to(p))] = count
        proofs.append(file)
    normal = []
    for name, count in [("cave-platform", 10), ("cave-platform-magia", 3)]:
        file = b / "cave-ff41-normal-pair-v2" / name / "receipt.json"
        data = read(file)
        assert data["state_equal"] and data["save_unchanged"] and not data["wrote"] and not data["unimplemented"]
        assert len(data["points"]) == data["hd_draws"] == count
        assert all(r["state_equal"] and r["terrain_hd_drawn"] and r["presentation"]["used"] for r in data["points"])
        for kind in ("player", "weapon"):
            assert data[kind + "_coverage"]["missing_visible_frames"] == 0
        normal.append({"path": str(file.relative_to(p)), "points": count,
                       "captured": data["captured"], "refused": data["refused"]})
        proofs.append(file)
    binary = b / "cave-ff41-runtime-v3/zeliard-preview"
    assert binary.read_bytes()[:4] == b"\x7fELF"
    for name in ("vet.log", "build.log"):
        file = binary.parent / name
        assert not file.read_bytes()
        proofs.append(file)
    gui_path = b / "cave-ff41-gui-v3/receipt.json"
    gui = read(gui_path)
    assert gui["status"] == "PASS_CAVE_FF41_INACTIVE_GUI_SCOPED"
    assert gui["binary_sha256"] == sha(binary) and gui["normal_DOS_exit"] == 0 and gui["save_unchanged"]
    assert gui["script_sha256"] == sha(p / "tools/hd_cave_ff41_smoke.py")
    assert gui["catalog_helper_sha256"] == sha(p / "tools/hd_cave_catalog_smoke.py")
    assert gui["helper_sha256"] == sha(p / "tools/integration_smoke.py")
    shots = {r["name"]: r for r in gui["shots"]}
    original_pixels = 0
    assert {r["name"] for r in gui["pairs"]} == {"ff41-inactive"}
    for pair in gui["pairs"]:
        states, images = [], []
        for suffix in ("original", "hd", "returned"):
            row = shots[pair["name"] + "-" + suffix]
            file = gui_path.parent / "boot1" / (row["name"] + ".png")
            assert sha(file) == row["image_sha256"]
            im = np.asarray(Image.open(file).convert("RGB"))
            assert im.shape == (680, 960, 3)
            images.append(im)
            states.append(row["state"])
            proofs.append(file)
        for field in ("steps", "cycles", "oracle_memory_sha256", "oracle_indexed_sha256", "oracle_palette_sha256"):
            assert states[0][field] == states[1][field] == states[2][field]
        assert np.array_equal(images[0][48:648], images[2][48:648])
        original_pixels += 960 * 600
        assert states[1]["cave_terrain_hd_matched"] and states[1]["cave_player_hd"]["matched"]
        if pair["name"] == "ff41-inactive":
            source = states[1]["cave_frame_presentation"]
            assert source["used"] and source["body_primary_fields"]["01B1:FF41"] == 1
            assert source["body_primary_fields"]["01B1:FF40"] == 0
            assert np.any(images[0][90:522] != images[1][90:522])
    save, orig = gui_path.parent / "user/game/SAVETEST.usr", p / "workplace/orig/SAVEM1.USR"
    assert sha(save) == sha(orig) and save.stat().st_mtime_ns == orig.stat().st_mtime_ns
    proofs.extend([binary, gui_path])
    sources = ["internal/prototype/hd/cave_sources.go", "internal/prototype/hd/cave_frames.go",
               "internal/prototype/hd/cave_frames_test.go", "internal/prototype/hd/cave_ff41_test.go",
               "internal/prototype/hd/cave_ff41_study_test.go", "tools/hd_gui_rejected_probe.go",
               "tools/ida_cave_ff41.py", "docs/cave-ff41-source-index.json",
               "tools/hd_cave_ff41_smoke.py", "tools/hd_ff41_audit.py", "tools/hd_catalog.py"]
    report = {"schema": 1, "status": "PASS_FF41_INACTIVE_SOURCE_NORMAL_GUI_SCOPED",
              "binary_sha256": sha(binary), "normal_state_equal_points": 13,
              "normal_pairs": normal, "gui_original_return_pixels": original_pixels,
              "test_counts": tests,
              "source_files": [{"path": f, "sha256": sha(p / f)} for f in sources],
              "proofs": [{"path": str(f.relative_to(p)), "sha256": sha(f)} for f in proofs],
              "limits": ["只擴充FF41=1且FF40=0；其他非零模式與既有來源拒絕仍待補。",
                         "既有高清圖面沿用，未新增或完成其他敵人、首領、城鎮及後段素材。",
                         "原GUI失敗日誌重生與新視窗分開記錄；沒有重建封包或影片。"]}
    args.output.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(report["status"], original_pixels, "original return pixels")


if __name__ == "__main__":
    main()

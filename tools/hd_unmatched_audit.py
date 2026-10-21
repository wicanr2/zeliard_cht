#!/usr/bin/env python3
"""核049八圖、九個原生遮罩、正常路徑、舊物件回歸及代表視窗。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", required=True, type=Path)
    ap.add_argument("--output", required=True, type=Path)
    args = ap.parse_args()
    p, out = args.project, args.output
    b = p / "workplace/full-hd-v1"
    assert not out.exists(), "不覆寫歷史稽核"
    sha = lambda f: hashlib.sha256(f.read_bytes()).hexdigest()
    read = lambda f: json.loads(f.read_text())
    proofs = []
    source_path = b / "cave-unmatched-assets-source-v1/receipt.json"
    source = read(source_path)
    assert source["status"] == "UNMATCHED_ASSETS_SOURCE_ONLY" and len(source["sources"]) == 17
    assert source["script_sha256"] == sha(p / "internal/prototype/hd/cave_unmatched_assets_study_test.go")
    sources = {r["key"]: r for r in source["sources"]}
    known = {r["key"]: r for r in source["known"]}
    for row in sources.values():
        for ext, key in [("memory", "memory_sha256"), ("body.memory", "body_memory_sha256"), ("indexed", "indexed_sha256")]:
            file = p / (row["source_prefix"] + "." + ext)
            assert sha(file) == row["source"][key]
        raw = source_path.parent / (row["file"] + ".raster")
        assert sha(raw) == row["key"]
        proofs.append(raw)
    gen_path = b / "cave-unmatched-art-v1/generation.json"
    gen = read(gen_path)
    assert gen["mode"] == "built-in imagegen" and len(gen["generations"]) == 8
    for row in gen["generations"]:
        assert row["prompt"] and sha(p / row["output_file"]) == row["output_sha256"]
        for ref in row["references"]:
            assert sha(p / ref["file"]) == ref["sha256"]
        proofs.append(p / row["output_file"])
    reg_path = b / "cave-unmatched-complete-v1/manifest.json"
    reg = read(reg_path)
    assert reg["source_receipt_sha256"] == sha(source_path) and reg["generation_sha256"] == sha(gen_path)
    assert reg["script_sha256"] == sha(p / "tools/hd_unmatched_art.py")
    assert len(reg["poses"]) == 8 and len(reg["aliases"]) == 9
    native_masks = rgba = 0
    for row in reg["poses"]:
        file = reg_path.parent / row["file"]
        assert sha(file) == row["sha256"]
        im = np.asarray(Image.open(file).convert("RGBA"))
        size = 96 if row["kind"] == "weapon" else 48
        assert im.shape == (size, size, 4) and hashlib.sha256(im.tobytes()).hexdigest() == row["rgba_sha256"]
        rgba += size * size
        proofs.append(file)
    for row in reg["aliases"]:
        src, old = sources[row["key"]], known[row["target_key"]]
        raw = (source_path.parent / (src["file"] + ".raster")).read_bytes()
        prior = (source_path.parent / (old["file"] + ".raster")).read_bytes()
        assert sha(source_path.parent / (old["file"] + ".raster")) == row["target_key"]
        assert src["object"]["Tiles"] == row["tiles"] and src["object"]["Backgrounds"] == row["backgrounds"]
        memory = (p / (src["source_prefix"] + ".memory")).read_bytes()
        for i, v in enumerate(raw):
            x, y = i % 16, i // 16
            q = y // 8 * 2 + x // 8
            tile, bg = row["tiles"][q], row["backgrounds"][q]
            transparent = bg != 0 and memory[0x11b10 + 0xa000 + tile * 8 + y % 8] & (128 >> (x % 8))
            assert prior[i] != 255 and v == (255 if transparent else prior[i])
            native_masks += 1
    assert native_masks == reg["verified_masked_native_pixels"] == 2304 and rgba == 46080
    runtime = b / "cave-unmatched-runtime-v1"
    for name in ("vet.log", "build.log"):
        assert not (runtime / name).read_bytes()
        proofs.append(runtime / name)
    # 首檔唯一skip為缺少輸出目錄的舊物件像素測試；第二檔已依原契約單獨補跑。
    expected_skip = {"TestCaveObjectsIndependentCompletedPixels"}
    for file, count, skips in [(runtime / "tests.jsonl", 53, expected_skip),
                               (runtime / "legacy-tests.jsonl", 1, set()),
                               (b / "cave-unmatched-normal-pair-v1/tests.jsonl", 3, set())]:
        rows = [json.loads(l) for l in file.read_text().splitlines()]
        assert rows[-1]["Action"] == "pass" and not any(r["Action"] == "fail" for r in rows)
        assert {r.get("Test") for r in rows if r["Action"] == "skip"} == skips
        assert sum(r["Action"] == "pass" and "Test" in r and "/" not in r["Test"] for r in rows) == count
        proofs.append(file)
    log = (runtime / "tests.jsonl").read_text()
    assert "17份來源548352原生像素零差" in log and "四個刀光獨立混色像素 2066" in log
    assert "混色像素 26692" in log
    normal = []
    for name, count, captures, refused in [("cave-platform", 10, 253, 0), ("cave-platform-magia", 3, 564, 2)]:
        file = b / "cave-unmatched-normal-pair-v1" / name / "receipt.json"
        d, old = read(file), read(b / "cave-profile-normal-pair-v2" / name / "receipt.json")
        assert d["state_equal"] and d["save_unchanged"] and not d["wrote"] and not d["unimplemented"]
        assert d["captured"] == captures and d["refused"] == refused and d["refusal_reasons"] == old["refusal_reasons"]
        assert len(d["points"]) == d["hd_draws"] == count
        for row, prior in zip(d["points"], old["points"]):
            assert row["state_equal"] and row["terrain_hd_drawn"] and row["presentation"]["used"]
            for key in ("steps", "cycles", "memory_sha256", "indexed_sha256"):
                assert row[key] == prior[key]
            assert row["objects_hd"]["terrain_underlay"]
        assert d["player_coverage"]["missing_visible_frames"] == d["weapon_coverage"]["missing_visible_frames"] == d["object_coverage"]["unknown_instances"] == 0
        normal.append({"path": str(file.relative_to(p)), "points": count, "captured": captures, "refused": refused,
                       "refusal_reasons": d["refusal_reasons"], "missing_player_frames": 0,
                       "missing_weapon_frames": 0, "unknown_object_instances": 0})
        proofs.append(file)
    binary = runtime / "zeliard-preview"
    assert binary.read_bytes()[:4] == b"\x7fELF"
    gui_path = b / "cave-unmatched-gui-v2/receipt.json"
    gui = read(gui_path)
    assert gui["status"] == "PASS_CAVE_UNMATCHED_ASSETS_GUI_SCOPED" and gui["normal_DOS_exit"] == 0 and gui["save_unchanged"]
    assert gui["binary_sha256"] == sha(binary) and gui["script_sha256"] == sha(p / "tools/hd_cave_unmatched_smoke.py")
    assert gui["platform_helper_sha256"] == sha(p / "tools/hd_cave_platform_smoke.py")
    assert gui["catalog_helper_sha256"] == sha(p / "tools/hd_cave_catalog_smoke.py")
    assert len(gui["pairs"]) == 1
    shots = {r["name"]: r for r in gui["shots"]}
    images, states = [], []
    for suffix in ("original", "hd", "returned"):
        row = shots["masked-object-" + suffix]
        file = gui_path.parent / "boot1" / (row["name"] + ".png")
        assert sha(file) == row["image_sha256"]
        im = np.asarray(Image.open(file).convert("RGB"))
        assert im.shape == (680, 960, 3)
        images.append(im)
        states.append(row["state"])
        proofs.append(file)
    for key in ("steps", "cycles", "oracle_memory_sha256", "oracle_indexed_sha256", "oracle_palette_sha256"):
        assert states[0][key] == states[1][key] == states[2][key]
    assert np.array_equal(images[0][48:648], images[2][48:648])
    assert np.any(images[0][90:522, 144:816] != images[1][90:522, 144:816])
    assert states[1]["cave_objects_hd"]["masked_known"] > 0 and states[1]["cave_objects_hd"]["terrain_underlay"]
    save, orig = gui_path.parent / "user/game/SAVETEST.usr", p / "workplace/orig/SAVEM1.USR"
    assert sha(save) == sha(orig) and save.stat().st_mtime_ns == orig.stat().st_mtime_ns
    review_path = gui_path.parent / "visual-review.json"
    review = read(review_path)
    assert review["reviewed_by"] == "root" and review["passed"]
    proofs.extend([source_path, gen_path, reg_path, binary, gui_path, review_path])
    files = ["cmd/zeliard-preview/main.go", "internal/prototype/hd/cave_object.go", "internal/prototype/hd/cave_terrain.go",
             "internal/prototype/hd/cave_extra_assets.go", "internal/prototype/hd/cave_extra_assets_test.go",
             "internal/prototype/hd/cave_frames_pair_test.go", "internal/prototype/hd/cave_weapon_cyan_test.go",
             "internal/prototype/hd/cave_unmatched_assets_study_test.go", "tools/hd_unmatched_art.py",
             "tools/hd_cave_catalog_smoke.py", "tools/hd_cave_unmatched_smoke.py", "tools/hd_unmatched_audit.py", "tools/hd_catalog.py"]
    report = {"schema": 1, "status": "PASS_UNMATCHED_ASSETS_SOURCE_ART_NORMAL_GUI_SCOPED", "binary_sha256": sha(binary),
              "normal_pairs": normal, "normal_state_equal_points": 13, "new_artworks": 8, "masked_source_keys": 9,
              "saved_native_pixels_equal": 548352, "verified_masked_native_pixels": native_masks,
              "registered_rgba_pixels": rgba, "independent_weapon_blended_pixels": 2066,
              "independent_object_blended_pixels": 26692, "gui_original_return_pixels": 576000,
              "source_files": [{"path": f, "sha256": sha(p / f)} for f in files],
              "proofs": [{"path": str(f.relative_to(p)), "sha256": sha(f)} for f in proofs],
              "limits": ["兩條路徑缺圖為零只涵蓋已選定完成來源；來源拒絕0／2維持原身份。",
                         "GUI只驗代表遮罩物件，八圖另沿17保存正常來源；其他城鎮、後段、首領與完整美術仍待完成。",
                         "首輪GUI腳本欄位錯誤保留失敗；未更新封包、版號、影片或README縮圖。"]}
    out.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n")
    print(report["status"])


if __name__ == "__main__":
    main()

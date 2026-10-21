#!/usr/bin/env python3
"""核048組合來源、五圖、正式原生重建、正常配對與代表GUI。"""
import argparse
import csv
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
    read = lambda f: json.loads(f.read_text())
    sha = lambda f: hashlib.sha256(f.read_bytes()).hexdigest()
    proofs, normal = [], []
    ledger_path = p / "docs/re-resolution-backlinks.tsv"
    ledger = list(csv.DictReader(ledger_path.read_text().splitlines(), delimiter="\t"))
    assert len(ledger) == 6
    for row in ledger:
        evidence, older = p / row["evidence_spec"], p / row["older_spec"]
        assert evidence.is_file() and older.is_file()
        assert row["address"].removeprefix("0x").lower() in evidence.read_text().lower()
        assert row["required_correction"] in older.read_text()
    proofs.append(ledger_path)
    for source_dir, path, expected in [("cave-profile-source-v1", "cave-platform", 15),
                                      ("cave-profile-magia-source-v1", "cave-platform-magia", 31)]:
        file = b / source_dir / "receipt.json"
        d = read(file)
        assert d["profile_signatures"] == expected and d["state_equal"] and d["save_unchanged"] and not d["state_injection"]
        assert d["script_sha256"] == sha(file.parent / "probe-source.go")
        old_path = b / "cave-ff41-normal-pair-v2" / path / "receipt.json"
        old = read(old_path)
        assert d["commands_sha256"] == old["commands_sha256"]
        assert len(d["observations"]) == len(old["points"])
        for row, prior in zip(d["observations"], old["points"]):
            assert all(row[k] == prior[k] for k in ("name", "steps", "cycles", "memory_sha256", "indexed_sha256"))
        for row in d["points"]:
            for ext, field in [("memory", "memory_sha256"), ("body.memory", "body_memory_sha256"),
                               ("indexed", "indexed_sha256"), ("palette", "palette_sha256")]:
                assert sha(file.parent / (row["file"] + "." + ext)) == row[field]
        proofs.extend([file, old_path, file.parent / "probe-source.go"])
    study_path = b / "cave-combined-profile-study-v1/receipt.json"
    study = read(study_path)
    assert study["native_matched"] == 46 and study["native_failed"] == 0 and study["unique_actor_keys"] == 6
    prototype_path = study_path.parent / "prototype.json"
    prototype = read(prototype_path)
    assert sha(study_path.parent / "original-source.go") == prototype["source_sha256"]
    assert sha(study_path.parent / "prototype-source.go") == prototype["prototype_sha256"]
    proofs.extend([study_path, prototype_path, study_path.parent / "original-source.go", study_path.parent / "prototype-source.go"])
    ida_path = b / "cave-profile-ida-v1/evidence.json"
    ida = read(ida_path)
    assert ida["tool"] == "IDA 9.4" and len(ida["instructions"]) == 262
    assert sum(not row["semantic_index"]["warning"] for row in ida["instructions"]) == 9
    assert ida["source_index_sha256"] == sha(p / "docs/cave-profile-source-index.json")
    assert ida["script_sha256"] == sha(p / "tools/ida_cave_profiles.py")
    proofs.append(ida_path)
    generation_path = b / "cave-combined-profile-art-v1/generation.json"
    generation = read(generation_path)
    assert generation["mode"] == "built-in imagegen" and len(generation["generations"]) == 6
    assert generation["selected"][-1] == "generated-4-v2.png"
    for row in generation["generations"]:
        assert row["prompt"] and sha(p / row["output_file"]) == row["output_sha256"]
        for ref in row["references"]:
            assert sha(p / ref["file"]) == ref["sha256"]
    reg_path = b / "cave-combined-profile-complete-v1/manifest.json"
    reg = read(reg_path)
    assert reg["generation_sha256"] == sha(generation_path) and reg["script_sha256"] == sha(p / "tools/hd_profile_art.py")
    assert reg["source_sha256"] == sha(b / "cave-combined-profile-art-v1/sources.json")
    assert len(reg["poses"]) == 5
    keys = {row["key"] for row in reg["poses"]}
    for row in reg["poses"]:
        file = reg_path.parent / row["file"]
        assert sha(file) == row["sha256"]
        im = np.asarray(Image.open(file).convert("RGBA"))
        assert im.shape == (72, 72, 4) and hashlib.sha256(im.tobytes()).hexdigest() == row["rgba_sha256"]
        proofs.append(file)
    proofs.extend([reg_path, generation_path, b / "cave-combined-profile-art-v1/sources.json"])
    runtime = b / "cave-profile-runtime-v1"
    binary = runtime / "zeliard-preview"
    assert binary.read_bytes()[:4] == b"\x7fELF"
    for name in ("vet.log", "build.log"):
        assert not (runtime / name).read_bytes()
        proofs.append(runtime / name)
    for file, count in [(runtime / "tests.jsonl", 46), (b / "cave-profile-normal-pair-v2/tests.jsonl", 3)]:
        rows = [json.loads(line) for line in file.read_text().splitlines()]
        assert rows[-1]["Action"] == "pass" and not any(r["Action"] in ("skip", "fail") for r in rows)
        assert sum(r["Action"] == "pass" and "Test" in r and "/" not in r["Test"] for r in rows) == count
        proofs.append(file)
    assert "五新圖25920註冊RGBA、8057混色像素相同" in (runtime / "tests.jsonl").read_text()
    observed = set()
    for name, count in [("cave-platform", 10), ("cave-platform-magia", 3)]:
        file = b / "cave-profile-normal-pair-v2" / name / "receipt.json"
        d = read(file)
        assert d["state_equal"] and d["save_unchanged"] and not d["wrote"] and not d["unimplemented"]
        assert d["refused"] == {"cave-platform": 0, "cave-platform-magia": 2}[name]
        assert len(d["points"]) == d["hd_draws"] == count
        assert d["player_coverage"]["missing_visible_frames"] == 0
        assert all(row["state_equal"] and row["terrain_hd_drawn"] and row["presentation"]["used"] for row in d["points"])
        observed.update(d["player_coverage"]["keys"])
        normal.append({"path": str(file.relative_to(p)), "points": count, "captured": d["captured"], "refused": d["refused"],
                       "refusal_reasons": d["refusal_reasons"],
                       "missing_player_frames": 0, "missing_weapon_frames": d["weapon_coverage"]["missing_visible_frames"],
                       "unknown_object_instances": d["object_coverage"]["unknown_instances"],
                       "unknown_weapon_sources": d["weapon_coverage"]["missing_sources"],
                       "unknown_object_sources": d["object_coverage"]["unknown_sources"]})
        proofs.append(file)
    assert keys <= observed
    gui_path = b / "cave-profile-gui-v1/receipt.json"
    gui = read(gui_path)
    assert gui["status"] == "PASS_CAVE_COMBINED_PROFILE_GUI_SCOPED" and gui["normal_DOS_exit"] == 0 and gui["save_unchanged"]
    assert gui["binary_sha256"] == sha(binary) and gui["script_sha256"] == sha(p / "tools/hd_cave_profile_smoke.py")
    assert gui["platform_helper_sha256"] == sha(p / "tools/hd_cave_platform_smoke.py")
    assert gui["catalog_helper_sha256"] == sha(p / "tools/hd_cave_catalog_smoke.py")
    assert len(gui["pairs"]) == 1
    shots = {r["name"]: r for r in gui["shots"]}
    images, states = [], []
    for suffix in ("original", "hd", "returned"):
        row = shots["profile-player-" + suffix]
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
    player = states[1]["cave_player_hd"]
    assert player["matched"] and player["source_key"] in keys and 67 <= player["pose"] < 71
    assert states[1]["cave_terrain_hd_matched"] and states[1]["cave_frame_presentation"]["body_primary_fields"]["01B1:FF39"] == 255
    save, orig = gui_path.parent / "user/game/SAVETEST.usr", p / "workplace/orig/SAVEM1.USR"
    assert sha(save) == sha(orig) and save.stat().st_mtime_ns == orig.stat().st_mtime_ns
    proofs.extend([gui_path, binary])
    files = ["cmd/zeliard-preview/main.go", "internal/prototype/hd/cave_sources.go", "internal/prototype/hd/cave_ff41_test.go",
             "internal/prototype/hd/cave_player_extra_test.go", "internal/prototype/hd/cave_player_profiles.go",
             "internal/prototype/hd/cave_player_profiles_test.go", "internal/prototype/hd/cave_profile_study_test.go",
             "internal/prototype/hd/cave_frames_pair_test.go",
             "tools/hd_cave_rejected_probe.go", "tools/hd_profile_study.py", "tools/hd_profile_art.py",
             "tools/ida_cave_profiles.py", "docs/cave-profile-source-index.json", "tools/hd_cave_profile_smoke.py",
             "tools/hd_cave_catalog_smoke.py", "tools/hd_profiles_audit.py", "tools/hd_catalog.py"]
    report = {"schema": 1, "status": "PASS_COMBINED_PROFILE_SOURCE_ART_NORMAL_GUI_SCOPED", "binary_sha256": sha(binary),
              "normal_pairs": normal, "normal_state_equal_points": 13, "registered_source_keys": 5,
              "existing_poses_preserved": 67, "saved_native_pixels_equal": 1483776,
              "registered_rgba_pixels": 25920, "independent_blended_pixels": 8057, "gui_original_return_pixels": 576000,
              "source_files": [{"path": f, "sha256": sha(p / f)} for f in files],
              "proofs": [{"path": str(f.relative_to(p)), "sha256": sha(f)} for f in proofs],
              "limits": ["兩條已驗平台路徑主角缺圖為零，來源拒絕0／2；不當全遊戲所有來源已支援。",
                         "尚有刀光與動態物件缺圖，來源鍵與保存位置見normal_pairs，沒有藉此稱完整HD。",
                         "五圖皆驗保存正常來源，GUI只驗代表背向姿態；封包、版號與影片未更新。"]}
    args.output.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(report["status"], "five poses", "selected-path source refusals 0/2")


if __name__ == "__main__":
    main()

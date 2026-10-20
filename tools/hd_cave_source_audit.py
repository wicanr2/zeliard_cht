#!/usr/bin/env python3
"""核對洞窟來源、原版地圖常式、正常初始化與美術候選；不冒稱HD完成。"""
import argparse
import hashlib
import json
import struct
from pathlib import Path

from PIL import Image


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    p = args.project
    base = p/"workplace/full-hd-v1"
    sha = lambda b: hashlib.sha256(b).hexdigest()
    read = lambda name: json.loads((base/name).read_text())
    sources = read("cave-terrain-sources-v1/receipt.json")
    maps = read("cave-map-sources-v1/receipt.json")
    oracle = read("cave-map-oracle-v1/receipt.json")
    survey = read("cave-save-survey-v1/receipt.json")
    normal = read("cave-save-survey-v2/receipt.json")
    ida = read("cave-map-ida-v2/evidence.json")
    primary = (p/"workplace/combat-v1/cave-entered.primary").read_bytes()
    assert sha(primary) == ida["input_sha256"] == oracle["primary_sha256"]
    assert oracle["status"] == "PASS_EXTRACTED_ORIGINAL_CAVE_MAPS"
    assert len(oracle["records"]) == len(maps["records"]) == 31
    assert all(x["matched"] and x["exit_code"] == 0 for x in oracle["records"])
    assert sum(x["cells"] for x in oracle["records"]) == maps["total_cells"] == 306048
    assert sources["total_tiles"] == 378 and len(sources["profiles"]) == 11
    for profile in sources["profiles"]:
        raw = (p/profile["source"]).read_bytes()
        assert sha(raw) == profile["source_sha256"]
        i = profile["record"]-74
        descriptor = primary[0x9c43+i*11:0x9c43+(i+1)*11]
        assert descriptor[:2] == bytes((2, profile["record"]+1))
        assert descriptor[2:].rstrip(b"\0").decode() == profile["key"]+".GRP"
    for entry, original in zip(maps["records"], oracle["records"]):
        assert entry["record"] == original["record"]
        original_bytes = (base/"cave-map-oracle-v1"/entry["columns"]).read_bytes()
        offline_bytes = (base/"cave-map-sources-v1"/entry["columns"]).read_bytes()
        assert original_bytes == offline_bytes and sha(original_bytes) == entry["columns_sha256"]
        assert not entry["unrendered_values"]
    assert normal["script_sha256"] == sha((p/"tools/hd_cave_save_probe.go").read_bytes())
    assert len(survey["boots"]) == 4 and len(normal["boots"]) == 1
    checkpoint_count = 0
    for record, directory in ((survey, "cave-save-survey-v1"), (normal, "cave-save-survey-v2")):
        assert record["status"] == "PASS_NORMAL_SAVE_SOURCE_SURVEY"
        for boot in record["boots"]:
            assert not boot["snapshot_restore"] and not boot["state_injection"]
            assert sha((p/"workplace/orig"/(boot["save"]+".USR")).read_bytes()) == boot["source_save_sha256"]
            for point in boot["points"]:
                assert point["state_equal"]
                checkpoint_count += 1
                memory = (base/directory/boot["save"]/(point["name"]+".memory")).read_bytes()
                assert sha(memory) == point["memory_sha256"]
                if "source_memory_sha256" in point:
                    source = (base/directory/boot["save"]/(point["name"]+"-source.memory")).read_bytes()
                    assert sha(source) == point["source_memory_sha256"]
                    assert source[0x1b10+0x9ef7] == 0  # 全部正常來源仍選MPP1。
    ring_cells = 0
    map_entry = next(x for x in maps["records"] if x["name"] == "MP10.MDT")
    columns = (base/"cave-map-sources-v1"/map_entry["columns"]).read_bytes()
    for load in normal["boots"][0]["map_loads"]:
        memory = (base/"cave-save-survey-v2/SAVEM1"/load["memory"]).read_bytes()
        assert sha(memory) == load["memory_sha256"] and load["source_ip"] == "01B1:6CEC"
        segment = memory[0x1b10:0x11b10]
        start = struct.unpack_from("<H", segment, 0x80)[0]
        width = struct.unpack_from("<H", segment, 0xc002)[0]
        assert width == map_entry["width"]
        for col in range(36):
            for row in range(64):
                assert segment[0xe000+row*36+col] == columns[((start+col)%width)*64+row]
                ring_cells += 1
    assert ring_cells >= 2304
    previous = json.loads((p/"workplace/prose-v1/progression-v1/receipt.json").read_text())
    reference = next(x for x in previous["snapshots"] if x["name"] == "dungeon-entry")
    assert reference["memory_sha256"] == sources["normal_memory_sha256"]
    assert sha((p/"workplace/prose-v1/progression-v1/dungeon-entry.memory").read_bytes()) == sources["normal_memory_sha256"]
    index = (p/"docs/cave-terrain-source-index.json").read_bytes()
    assert sha(index) == ida["source_index_sha256"]
    annotations = [x for x in ida["instructions"] if x["semantic_index"].get("address")]
    assert len(annotations) == 5 and all(x["semantic_index"]["level"] == "已證實" for x in annotations)
    art = read("cave-ice-art-v2/generation.json")
    art_path = base/"cave-ice-art-v2"/art["output"]
    assert sha(art_path.read_bytes()) == art["sha256"]
    im = Image.open(art_path).convert("RGBA")
    assert im.size == (2048, 768) and im.getchannel("A").getextrema()[0] == 0
    files = ["tools/hd_cave_sources.py", "tools/hd_cave_maps.py", "tools/hd_cave_map_oracle.go",
             "tools/hd_cave_save_probe.go", "tools/ida_cave_resource_evidence.py", "tools/hd_cave_source_audit.py",
             "docs/cave-terrain-source-index.json"]
    proofs = ["cave-terrain-sources-v1/receipt.json", "cave-map-sources-v1/receipt.json", "cave-map-oracle-v1/receipt.json",
              "cave-save-survey-v1/receipt.json", "cave-save-survey-v1/driver-v1.go", "cave-save-survey-v2/receipt.json",
              "cave-map-ida-v2/evidence.json", "cave-ice-art-v1/generation.json", "cave-ice-art-v2/generation.json"]
    report = {"schema": 1, "status": "PASS_CAVE_SOURCE_STAGE", "spec": "docs/spec/037-all-cave-terrain-hd.md",
              "terrain_profiles": 11, "terrain_tiles": 378, "maps": 31, "decoded_map_cells": 306048,
              "normal_checkpoints": checkpoint_count, "normal_initial_ring_cells": ring_cells,
              "normal_terrain_profiles": ["MPP1"], "ice_art_status": "candidate_not_registered",
              "source_files": [{"path": f, "sha256": sha((p/f).read_bytes())} for f in files],
              "proofs": [{"path": str((base/f).relative_to(p)), "sha256": sha((base/f).read_bytes())} for f in proofs],
              "limits": ["來源與格式研究，不是正常到達其餘洞窟或新增HD呈現的證據。",
                         "冰窟圖集兩版仍未註冊；空白格殘留alpha，尚未通過正式素材驗收。",
                         "全遊戲HD進行中，正式程式、配樂、VERSION及現行封包未改。"]}
    assert not args.output.exists(), "保留既有收據身份"
    args.output.write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
    print("PASS_CAVE_SOURCE_STAGE", 378, "tiles", 31, "maps", ring_cells, "normal ring cells")


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""建立全遊戲HD覆蓋清單；原始位元組與圖面只留本機。"""
import argparse
import hashlib
import json
import re
from collections import Counter
from pathlib import Path

from hd_inventory import EXPECTED, records


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    p = args.project
    sha = lambda b: hashlib.sha256(b).hexdigest()
    resident = p/"workplace/probe-intro-memory/memory-01B1_0000.bin"
    resident_bytes = resident.read_bytes()
    assert sha(resident_bytes) == "c9f188ea847e8d364ccf8dbc9f6a9b9c74e192c0c83bd7a012b552e6d89b13d8"
    originals = {name: (p/"workplace/orig"/name).read_bytes() for name in EXPECTED}
    for name, data in originals.items():
        assert sha(data) == EXPECTED[name][1]
    inputs = [(name, b, "SAR檔案從0起算偏移") for name, b in originals.items()]
    inputs.append(("resident", resident_bytes[:0x1136], "固定快照dosgolem01B1段內偏移，亦為IDA base0"))
    names = {}
    for source, data, space in inputs:
        for m in re.finditer(rb"([\x00-\x02])([\x01-\x60])([A-Za-z0-9_]{1,8}\.[A-Za-z0-9]{3})\x00", data):
            archive, index = "ZELRES"+str(m[1][0]+1)+".SAR", m[2][0]-1
            if index >= EXPECTED[archive][0]:
                continue
            names.setdefault((archive, index), []).append({
                "name": m[3].decode("ascii"), "source": source, "offset": m.start(), "address_space": space,
                "level": "強推論：與已證實11byte descriptor形狀相同，逐一consumer仍須核對",
            })
    decoded = p/"workplace/full-hd-v1/decoded-v2/decoded-inventory.json"
    oracle = p/"workplace/full-hd-v1/decoder-oracle-v4/receipt.json"
    raw_report, oracle_report = json.loads(decoded.read_text()), json.loads(oracle.read_text())
    assert oracle_report["status"] == "PASS_EXTRACTED_ORIGINAL_DECODER"
    verified = {(x["archive"], x["index"]): x for x in oracle_report["records"]}
    compressed = {(x["archive"], x["index"]): x for x in raw_report["records"]}
    assert len(verified) == len(compressed) == 84
    planar = p/"workplace/full-hd-v1/planar-oracle-v2/receipt.json"
    grp = p/"workplace/full-hd-v1/grp-sources-v3/receipt.json"
    planar_report, grp_report = json.loads(planar.read_text()), json.loads(grp.read_text())
    assert planar_report["status"] == "PASS_EXTRACTED_ORIGINAL_PLANAR" and len(planar_report["records"]) == 80
    assert all(x["matched"] and x["exit_code"] == 0 for x in planar_report["records"])
    assert grp_report["status"] == "PASS_FOUR_NORMAL_RESOURCE_BINDINGS"
    runtime_catalog = p/"workplace/full-hd-v1/cave-mpp6-final-audit-v1.json"
    runtime_catalog_report = json.loads(runtime_catalog.read_text())
    assert runtime_catalog_report["status"] == "PASS_MPP6_STRUCTURAL_AND_MPP1_REGRESSION"
    for source in runtime_catalog_report["source_files"]+runtime_catalog_report["proofs"]:
        assert sha((p/source["path"]).read_bytes()) == source["sha256"]
    current_sources = {r["path"]:r["sha256"] for r in runtime_catalog_report["source_files"]}
    castle = p/"workplace/full-hd-v1/castle-final-audit-v2.json"
    castle_report = json.loads(castle.read_text())
    assert castle_report["status"] == "PASS_CASTLE_HD_SCOPED"
    for source in castle_report["source_files"]:
        actual = sha((p/source["path"]).read_bytes())
        if source["path"] == "cmd/zeliard-preview/main.go":
            # 入口另增冰窟旗標，身份沿本輪稽核；城堡完整視窗仍保留6670698f原身份。
            assert actual == current_sources[source["path"]]
        else:
            assert actual == source["sha256"]
    cave = p/"workplace/full-hd-v1/cave-source-stage-audit-v1.json"
    cave_report = json.loads(cave.read_text())
    assert cave_report["status"] == "PASS_CAVE_SOURCE_STAGE"
    for source in cave_report["source_files"]+cave_report["proofs"]:
        assert sha((p/source["path"]).read_bytes()) == source["sha256"]
    terrain_path = p/"workplace/full-hd-v1/cave-terrain-sources-v1/receipt.json"
    terrain_sources = {x["record"]: x for x in json.loads(terrain_path.read_text())["profiles"]}
    cave_maps_path = p/"workplace/full-hd-v1/cave-map-sources-v1/receipt.json"
    cave_maps = {x["record"]: x for x in json.loads(cave_maps_path.read_text())["records"]}
    grp_sources = {("ZELRES2.SAR", x["record"]): x for x in grp_report["records"]}
    catalog = []
    for archive, data in originals.items():
        for entry in records(data):
            key = archive, entry["index"]
            labels = names.get(key, [])
            suffixes = {Path(x["name"]).suffix.upper() for x in labels}
            role = {".GRP": "graphics_candidate", ".MDT": "map_candidate", ".MSD": "audio_candidate", ".BIN": "program_candidate"}
            classification = role[next(iter(suffixes))] if len(suffixes) == 1 and next(iter(suffixes)) in role else "unknown"
            item = {"key": archive.lower().replace(".sar", "")+f".record.{entry['index']:02d}",
                    "archive": archive, "record": entry["index"], "payload_offset": entry["payload_offset"],
                    "payload_bytes": entry["bytes"], "payload_sha256": entry["sha256"],
                    "classification": classification, "names": labels,
                    "hd_coverage": "not_assessed", "normal_path": "not_assessed"}
            if key in compressed:
                d, v = compressed[key], verified[key]
                raw = (decoded.parent/d["decoded_path"]).read_bytes()
                actual = (oracle.parent/d["decoded_path"]).read_bytes()
                assert d["result"] == "DECODED_CANDIDATE" and v["matched"] and v["exit_code"] == 0
                assert raw == actual and sha(raw) == v["original_sha256"] == v["offline_sha256"]
                item["decoder"] = {"kind": d["rle_kind"], "bytes": len(raw), "sha256": sha(raw),
                                   "level": "已證實：與擷取原版常式逐byte相同，非正常遊戲路徑驗收"}
            if key in grp_sources:
                source = grp_sources[key]
                item["graphics_source"] = {k: source[k] for k in ("tiles", "tile_offset", "pixels_sha256", "masks_sha256", "status", "normal_memory_sha256", "runtime_bytes_verified", "runtime_mask_bytes_verified", "uninterpreted_tail_bytes")}
                item["graphics_source"]["evidence"] = str(grp.relative_to(p))
                if key == ("ZELRES2.SAR", 29):
                    item["hd_coverage"] = "partial_scoped_candidate_verified"
                    item["normal_path"] = "Muralla居民；規格035，交談背景仍退回原貌"
                if key == ("ZELRES2.SAR", 33):
                    item["hd_coverage"] = "partial_scoped_candidate_verified"
                    item["normal_path"] = "城堡前景與四旗幟；規格036，正常Castle／Muralla往返，未更新封包"
            if archive == "ZELRES3.SAR" and entry["index"] in terrain_sources:
                source = terrain_sources[entry["index"]]
                item["cave_terrain_source"] = {k: source[k] for k in ("key", "tiles", "source_sha256", "packed_sha256", "status", "runtime_bytes_verified")}
                item["cave_terrain_source"]["evidence"] = str(terrain_path.relative_to(p))
                if entry["index"] in (75,76,77,78,79):
                    item["hd_coverage"] = "connected_structural_verified_normal_path_pending"
                    item["normal_path"] = "MPP2／MPP3／MPP5／MPP6／冰窟正常進出、捲動與事件待驗；結構像素見037，MPP1回歸另列"
            if archive == "ZELRES3.SAR" and entry["index"] in cave_maps:
                source = cave_maps[entry["index"]]
                item["map_source"] = {k: source[k] for k in ("width", "height", "columns_sha256", "family_candidate", "selector_offset", "selector_raw")}
                item["map_source"]["level"] = "已證實的原版解壓及descriptor選擇；非正常到達或HD驗收"
                item["map_source"]["evidence"] = str(cave.relative_to(p))
            catalog.append(item)
    report = {"schema": 1, "objective": "完成全遊戲HD", "status": "in_progress",
              "scope_spec": "docs/spec/034-full-game-hd.md",
              "counts": {"container_records": len(catalog), "named_candidates": sum(bool(x["names"]) for x in catalog),
                         "verified_decoded_records": len(verified), "verified_cave_maps": len(cave_maps),
                         "verified_cave_map_cells": cave_report["decoded_map_cells"], "cave_terrain_source_tiles": cave_report["terrain_tiles"],
                         "classification_candidates": dict(Counter(x["classification"] for x in catalog))},
              "limits": ["194是SAR容器記錄數，尚不是美術素材分母；不換算HD完成百分比",
                         "not_assessed表示尚未將既有局部HD與此全量記錄逐項對應，不否定前版已驗成果",
                         "檔名／副檔名只作分類候選，不能單憑GRP或MDT宣稱已解圖、全部可達或已HD"],
              "prior_hd_evidence": "VERIFICATION-MATRIX.md",
              "runtime_components": [{"key": "muralla.residents", "status": "scoped_candidate_verified",
                                      "spec": "docs/spec/035-town-residents-hd.md", "groups": 5, "phases": 40,
                                      "evidence": "workplace/full-hd-v1/town-npc-final-audit-v1.json",
                                      "source_record": "zelres2.record.29",
                                      "limits": "只驗Muralla固定來源；其他城鎮及交談背景仍有缺口"},
                                     {"key": "castle.foreground", "status": "scoped_candidate_verified",
                                      "spec": "docs/spec/036-castle-hd.md", "source_record": "zelres2.record.33",
                                      "evidence": str(castle.relative_to(p)),
                                      "limits": "只驗城堡前景、四旗幟與Muralla往返；phase4、交談背景及其他場景仍有缺口，未更新封包"},
                                     {"key": "cave.terrain.ice", "status": "connected_structural_verified_normal_path_pending",
                                      "spec": "docs/spec/037-all-cave-terrain-hd.md", "source_record": "zelres3.record.77",
                                      "evidence": str(runtime_catalog.relative_to(p)), "registered_tiles":21,
                                      "limits": "21格完整手繪圖集接入可選前端，三份合成結構快照與光球取樣通過；正常冰窟進出、捲動及20／21事件用途未驗；MPP1回歸不當冰窟驗收"},
                                     {"key": "cave.terrain.MPP2", "status": "connected_structural_verified_normal_path_pending",
                                      "spec": "docs/spec/037-all-cave-terrain-hd.md", "source_record": "zelres3.record.75",
                                      "evidence": str(runtime_catalog.relative_to(p)), "registered_tiles":34,
                                      "limits": "34格固定圖集接入可選前端，三份結構快照及全34格繪圖通過；正常進出／捲動及26..34事件用途待驗；不當全洞窟完成"},
                                     {"key": "cave.terrain.MPP3", "status": "connected_structural_verified_normal_path_pending",
                                      "spec": "docs/spec/037-all-cave-terrain-hd.md", "source_record": "zelres3.record.76",
                                      "evidence": str(runtime_catalog.relative_to(p)), "registered_tiles":43,
                                      "limits": "43格、九個材質ID及完整BOSQUE字帶已接；三份結構快照、全格及三家族共存通過；正常進出／捲動／事件待驗，不當全洞窟完成"},
                                     {"key": "cave.terrain.MPP5", "status": "connected_structural_verified_normal_path_pending",
                                      "spec": "docs/spec/037-all-cave-terrain-hd.md", "source_record": "zelres3.record.78",
                                      "evidence": str(runtime_catalog.relative_to(p)), "registered_tiles":52,
                                      "limits": "52格、20個材質ID及32個細部已接；三份結構快照、全格及四家族共存通過；正常進出／捲動及48..52事件用途待驗，不當全洞窟完成"},
                                     {"key": "cave.terrain.MPP6", "status": "connected_structural_verified_normal_path_pending",
                                      "spec": "docs/spec/037-all-cave-terrain-hd.md", "source_record": "zelres3.record.79",
                                      "evidence": str(runtime_catalog.relative_to(p)), "registered_tiles":36,
                                      "limits": "36格、金黃／紫岩雙材質與非連號DORADO字帶已接；四份結構快照、全格及五家族共存通過；正常進出／捲動與尾端圖樣事件用途待驗，不當全洞窟完成"},
                                     {"key": "cave.source.catalog", "status": "implemented_scoped_sampling",
                                      "spec": "docs/spec/037-all-cave-terrain-hd.md", "maps":31,"terrain_profiles":11,
                                      "evidence":str(runtime_catalog.relative_to(p)),
                                      "limits":"來源辨識已接入，正常35點與實際視窗只驗MPP1；不當其他家族正常可達或HD完成"}],
              "proofs": [{"path": str(f.relative_to(p)), "sha256": sha(f.read_bytes())} for f in (resident, decoded, oracle, planar, grp, castle, cave, runtime_catalog)],
              "archives": [{"name": n, "sha256": EXPECTED[n][1]} for n in originals], "records": catalog}
    args.output.write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
    print(json.dumps(report["counts"], ensure_ascii=False))


if __name__ == "__main__":
    main()

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
    back = p/"workplace/full-hd-v1/town-back-audit-v1.json"
    interior = p/"workplace/full-hd-v1/interior-final-audit-v2.json"
    back_report, interior_report = json.loads(back.read_text()), json.loads(interior.read_text())
    sage = p/"workplace/full-hd-v1/sage-frames-final-audit-v1.json"
    sage_report = json.loads(sage.read_text())
    merchant = p/"workplace/full-hd-v1/merchant-frames-final-audit-v1.json"
    merchant_report = json.loads(merchant.read_text())
    assert merchant_report["status"] == "PASS_MERCHANT_KNOWN_FRAMES_SOURCE_ART_NORMAL_GUI"
    towns = p/"workplace/full-hd-v1/town-resources-stage-audit-v2.json"
    towns_report = json.loads(towns.read_text())
    assert towns_report["status"] == "SOURCE_AND_EIGHT_ART_CANDIDATES_NORMAL_PATH_PENDING"
    assert sage_report["status"] == "PASS_SAGE_SOURCE_ART_NORMAL_GUI"
    assert back_report["status"] == "SOURCE_ART_NORMAL_PASSED_GUI_PENDING"
    assert interior_report["status"] == "PASS_INTERIOR_SOURCES_ART_AND_NORMAL_HUT_GUI"
    # 舊收據保留原程式身份。來源後續有修訂時，必須由具名新稽核銜接，不能把舊GUI換稱新程式。
    latest_sources, replacement_evidence, source_revisions = {}, {}, []
    for audit_path, audit in ((back, back_report), (interior, interior_report), (sage, sage_report), (merchant, merchant_report), (towns, towns_report)):
        for row in audit["source_files"]:
            latest_sources[row["path"]] = row["sha256"]
            replacement_evidence[row["path"]] = str(audit_path.relative_to(p))
    def validate_audit(audit_path, audit):
        for row in audit.get("source_files", []):
            actual = sha((p/row["path"]).read_bytes())
            if actual != row["sha256"]:
                assert actual == latest_sources.get(row["path"]), (str(audit_path), row["path"])
                source_revisions.append({"path":row["path"],"historical_audit":str(audit_path.relative_to(p)),
                                         "historical_sha256":row["sha256"],"current_sha256":actual,
                                         "replacement_evidence":replacement_evidence[row["path"]],
                                         "limits":"原收據仍屬原程式；只將已核修訂與現行來源銜接，不升舊GUI身份。"})
        for row in audit.get("proofs", []):
            assert sha((p/row["path"]).read_bytes()) == row["sha256"]
    validate_audit(back, back_report)
    validate_audit(interior, interior_report)
    validate_audit(sage, sage_report)
    validate_audit(merchant, merchant_report)
    validate_audit(towns, towns_report)
    town_sources_path = p/"workplace/full-hd-v1/town-resources-source-v1/receipt.json"
    town_sources = {row["record"]:row for row in json.loads(town_sources_path.read_text())["maps"]}
    interior_path = p/"workplace/full-hd-v1/interior-layout-v2/receipt.json"
    interior_sources = {row["graphics_record"]:row for row in json.loads(interior_path.read_text())["records"]}
    runtime_catalog = p/"workplace/full-hd-v1/cave-mpp1-final-audit-v1.json"
    runtime_catalog_report = json.loads(runtime_catalog.read_text())
    assert runtime_catalog_report["status"] == "PASS_MPP1_EXTRAS_AND_ELEVEN_TERRAIN_FAMILIES"
    validate_audit(runtime_catalog, runtime_catalog_report)
    castle = p/"workplace/full-hd-v1/castle-final-audit-v2.json"
    castle_report = json.loads(castle.read_text())
    assert castle_report["status"] == "PASS_CASTLE_HD_SCOPED"
    validate_audit(castle, castle_report)
    cave = p/"workplace/full-hd-v1/cave-source-stage-audit-v1.json"
    cave_report = json.loads(cave.read_text())
    assert cave_report["status"] == "PASS_CAVE_SOURCE_STAGE"
    validate_audit(cave, cave_report)
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
                if entry["index"] == 74:
                    item["hd_coverage"] = "connected_structural_verified_normal_path_scoped"
                    item["normal_path"] = "25格圖面具備；SATONO／MURALLA靜態構圖、首戰及回城回歸已驗。正常Satono到達與其餘分支未驗，不當全洞窟完成"
                if entry["index"] in (75,76,77,78,79,80):
                    item["hd_coverage"] = "connected_structural_verified_normal_path_pending"
                    item["normal_path"] = "MPP2／MPP3／MPP5／MPP6／MPP7／冰窟正常進出、捲動與事件待驗；結構像素見037，MPP1回歸另列"
                if entry["index"] in (81,82,83,84):
                    item["hd_coverage"] = "connected_structural_verified_normal_path_pending"
                    item["normal_path"] = "MPP8／MPP9／MPPA／MPPB圖面與九份結構地圖已驗；正常到達、進出、捲動及事件待驗，不當全洞窟完成"
            if archive == "ZELRES2.SAR" and entry["index"] in interior_sources:
                source = interior_sources[entry["index"]]
                item["interior_source"] = {k:source[k] for k in ("name","program_record","program_sha256","graphics_sha256","packed_sha256","tile_count","covered_tile_ids","unreferenced_tile_ids","normal_binding","frames","source_level")}
                item["interior_source"]["evidence"] = str(interior_path.relative_to(p))
                if entry["index"] == 19:
                    item["hd_coverage"] = "connected_scoped_normal_path_verified"
                    item["normal_path"] = "冷新局走到城堡亭台、Up進小屋看石像、空白離開；完整原版配對與實際視窗通過，見039。不代表全遊戲完成"
                elif entry["index"] == 24:
                    item["hd_coverage"] = "connected_structural_verified_normal_path_pending"
                    item["normal_path"] = "112圖格及七表項／四種圖像已接固定圖組，結構合成與退回通過；旅館正常到達、事件與GUI待驗"
                else:
                    item["hd_coverage"] = "connected_scoped_normal_path_verified"
                    item["normal_path"] = "166圖格、八手勢及兩高清動畫相位已接，正常查看力量八手勢配對與代表GUI、離店通過；置中手勢及其他城鎮正常路徑未驗"
                    item["sage_frames"] = {"tiles":166,"hand_poses":8,"registered_images":16,"structural_variants":38,
                                           "normal_points":sage_report["normal_points"],"evidence":str(sage.relative_to(p)),
                                           "limits":"interior_source仍保存前輪基底圖表身份；完整圖塊表及手勢來源另沿規格040"}
            if archive == "ZELRES2.SAR" and entry["index"] == 20:
                item["hd_coverage"] = "known_consumers_scoped_normal_path_verified"
                item["normal_path"] = "八姿態及兩局部動畫、正常購買／告別49點及代表GUI通過；254引用圖格已核，非空139、140用途未知，不當全資源或全遊戲完成"
                item["merchant_frames"] = {"tiles":256,"referenced_tiles":254,"unreferenced_tiles":[139,140],
                                           "poses":8,"registered_images":10,"normal_points":49,
                                           "evidence":str(merchant.relative_to(p))}
            if archive == "ZELRES2.SAR" and entry["index"] in town_sources:
                source = town_sources[entry["index"]]
                item["town_source"] = {k:source[k] for k in ("key","width","height","terrain_key","terrain_selector","background_program","background_selector","unresolved_tile_ids")}
                item["town_source"]["evidence"] = str(town_sources_path.relative_to(p))
                item["town_source"]["level"] = "已核靜態地圖與原版選擇端；新增城鎮正常到達另驗"
                if entry["index"] >= 38:
                    item["hd_coverage"] = "art_candidate_not_connected_normal_path_pending"
                    item["normal_path"] = "原始前景及八份生成候選已保存；正常新城鎮未到達，接縫／門洞／招牌與前端接入待驗，不當HD完成"
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
                         "interior_source_tiles":sum(x["tile_count"] for x in interior_sources.values()),
                         "new_hut_inn_source_tiles":247,"normal_bound_interior_resources":2,
                         "sage_hand_poses":8,"sage_registered_images":16,"sage_structural_variants":38,
                         "merchant_known_consumer_tiles":254,"merchant_unreferenced_tiles":2,"merchant_registered_images":10,
                         "town_maps_structurally_checked":10,"town_map_cells":17784,"additional_town_art_candidates":8,
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
                                      "limits": "原城堡視窗保留6670698f身份；背向另沿038，GUI未驗。交談背景及其他場景仍有缺口，未更新封包"},
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
                                     {"key": "cave.terrain.MPP7", "status": "connected_structural_verified_normal_path_pending",
                                      "spec": "docs/spec/037-all-cave-terrain-hd.md", "source_record": "zelres3.record.80",
                                      "evidence": str(runtime_catalog.relative_to(p)), "registered_tiles":61,"visible_tiles":58,"source_empty_slots":[40,41,42],
                                      "limits": "61槽、赤紅／白熱雙材質、LLAMA字帶及三個原版空白槽已接；四份結構快照、58個可見圖格、十一尾端圖樣差異與六家族共存通過；正常進出／捲動／事件待驗，不當全洞窟完成"},
                                     {"key": "cave.source.catalog", "status": "implemented_scoped_sampling",
                                      "spec": "docs/spec/037-all-cave-terrain-hd.md", "maps":31,"terrain_profiles":11,
                                      "evidence":str(runtime_catalog.relative_to(p)),
                                      "limits":"來源辨識已接入，正常35點與實際視窗只驗MPP1；不當其他家族正常可達或HD完成"}],
              "art_candidates": [{"key":row["key"],"source_record":f"zelres2.record.{row['record']}",
                                  "status":"not_connected_normal_path_and_calibration_pending","spec":"docs/spec/042-all-town-resources.md",
                                  "evidence":str(towns.relative_to(p)),"unresolved_tile_ids":row["unresolved_tile_ids"]}
                                 for record,row in town_sources.items() if record>=38],
              "proofs": [{"path": str(f.relative_to(p)), "sha256": sha(f.read_bytes())} for f in (resident, decoded, oracle, planar, grp, castle, cave, runtime_catalog, back, interior, interior_path, sage, merchant, towns, town_sources_path)],
              "source_revisions":source_revisions,
              "archives": [{"name": n, "sha256": EXPECTED[n][1]} for n in originals], "records": catalog}
    for family,count,visible,record in (("MPP8",43,40,81),("MPP9",38,38,82),("MPPA",20,19,83),("MPPB",5,5,84)):
        report["runtime_components"].insert(-1, {
            "key":"cave.terrain."+family,"status":"connected_structural_verified_normal_path_pending",
            "spec":"docs/spec/037-all-cave-terrain-hd.md","source_record":f"zelres3.record.{record}",
            "evidence":str(runtime_catalog.relative_to(p)),"registered_tiles":count,"visible_tiles":visible,
            "limits":"固定圖面、來源、全槽及結構地圖已驗；十家族共存與正常MPP1回歸通過。正常新增家族可達與事件未驗，不當全遊戲HD完成。"})
    report["runtime_components"].insert(-1, {
        "key":"cave.terrain.MPP1","status":"connected_structural_verified_normal_path_scoped",
        "spec":"docs/spec/037-all-cave-terrain-hd.md","source_record":"zelres3.record.74",
        "evidence":str(runtime_catalog.relative_to(p)),"registered_tiles":25,
        "limits":"原有14格與材質保持相同，另補15及兩組圖牌。三構圖與十一家族共存、正常MPP1首戰／回城回歸通過；正常Satono到達與後段事件仍未驗。"})
    for component in [
        {"key":"town.player.back","status":"source_art_normal_passed_gui_pending","spec":"docs/spec/038-town-player-back.md",
         "source_record":"zelres2.record.31","evidence":str(back.relative_to(p)),
         "limits":"兩朝向共用背向圖，正常完成回呼合成已核；實際背向視窗未捕捉，保持待驗。"},
        {"key":"interior.hut","status":"scoped_normal_path_and_gui_verified","spec":"docs/spec/039-interior-resources.md",
         "source_record":"zelres2.record.19","evidence":str(interior.relative_to(p)),"registered_images":1,
         "limits":"只驗正常進小屋看石像及離開，不把旅館或其他劇情分支算完成。"},
        {"key":"interior.sage.hands","status":"scoped_normal_path_and_gui_verified","spec":"docs/spec/040-sage-frames-hd.md",
         "source_record":"zelres2.record.25","evidence":str(sage.relative_to(p)),"registered_images":16,"hand_poses":8,
         "limits":"正常Muralla查看力量八手勢及代表GUI已核；置中手勢與其他城鎮正常路徑未驗，不外推全遊戲HD。"},
        {"key":"interior.merchant.frames","status":"known_consumers_scoped_normal_path_and_gui_verified","spec":"docs/spec/041-merchant-frames-hd.md",
         "source_record":"zelres2.record.20","evidence":str(merchant.relative_to(p)),"registered_images":10,"poses":8,
         "limits":"254引用圖格、八姿態與兩局部動畫、正常購買／告別及代表GUI已核；未引用139、140用途未知，不外推全資源或全遊戲HD。"},
        {"key":"interior.inn","status":"connected_structural_verified_normal_path_pending","spec":"docs/spec/039-interior-resources.md",
         "source_record":"zelres2.record.24","evidence":str(interior.relative_to(p)),"registered_images":4,
         "limits":"七表項按像素共用四張圖；正常入口、觸發與GUI未驗，不宣稱正常旅館完成。"}]:
        report["runtime_components"].insert(-1,component)
    args.output.write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
    print(json.dumps(report["counts"], ensure_ascii=False))


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""核洞窟來源辨識候選、正常路徑與實際視窗；美術候選另列。"""
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
    base = p/"workplace/full-hd-v1"
    sha = lambda f: hashlib.sha256(f.read_bytes()).hexdigest()
    read = lambda name: json.loads((base/name).read_text())
    gui = read("cave-catalog-gui-v3/receipt.json")
    binary = base/"cave-catalog-runtime-v1/zeliard-preview"
    assert gui["status"] == "PASS_CAVE_CATALOG_GUI_SCOPED" and gui["normal_DOS_exit"] == 0
    assert gui["binary_sha256"] == sha(binary)
    assert gui["script_sha256"] == sha(p/"tools/hd_cave_catalog_smoke.py")
    assert gui["helper_sha256"] == sha(p/"tools/integration_smoke.py")
    rows = []
    for file in ("unit-tests.jsonl", "normal-and-refusal-tests.jsonl"):
        rows += [json.loads(line) for line in (base/"cave-catalog-runtime-v1"/file).read_text().splitlines()]
    assert not any(x["Action"] in ("fail", "skip") for x in rows)
    names = {x.get("Test") for x in rows if x["Action"] == "pass"}
    required = {"TestCaveCatalogRealSourceKeys", "TestCaveCatalogUnknownAndCrossMap", "TestCaveSourcesNormalCompletedFrames",
                "TestCaveTerrainIndependentOfflinePixels", "TestCaveFramesNormalPathsReadOnly",
                "TestCaveFramesRejectSourceAndTransitions", "TestCaveSourcesRejectUnknownOrCorrupt", "TestCaveTerrainAtomicFallback"}
    assert required <= names
    points = 0
    for name, count in (("combat",25),("lifecycle-return-v1",10)):
        receipt = read(f"cave-catalog-pair-v1/{name}/receipt.json")
        assert receipt["state_equal"] and receipt["save_unchanged"] and len(receipt["points"]) == count
        assert not receipt["held_keys"] and not receipt["wrote"] and not receipt["unimplemented"]
        points += count
    shots = {x["name"]: x for x in gui["shots"]}
    restored_pixels = changed_pixels = 0
    for pair in gui["pairs"]:
        images, states = [], []
        for key in ("original", "hd", "returned"):
            file = base/"cave-catalog-gui-v3/boot1"/pair[key]
            item = shots[file.stem]
            assert sha(file) == item["image_sha256"]
            im = Image.open(file).convert("RGB")
            assert im.size == (960,680)
            images.append(im.crop((0,48,960,648)))
            states.append(item["state"])
        for field in ("steps","cycles","oracle_memory_sha256","oracle_indexed_sha256","oracle_palette_sha256"):
            assert states[0][field] == states[1][field] == states[2][field]
        assert ImageChops.difference(images[0],images[2]).getbbox() is None
        restored_pixels += 576000
        diff = ImageChops.difference(images[0],images[1])
        changed_pixels += sum(any(c) for c in diff.getdata())
        assert states[1]["cave_frame_presentation"]["source_map_key"] == "MP10"
        assert states[1]["cave_frame_presentation"]["source_terrain_key"] == "MPP1"
    assert len(gui["pairs"]) == 2 and changed_pixels > 1000
    town = shots["returned-town"]["state"]
    assert town["town_scenery_hd_matched"] and not town["cave_frame_presentation"]["used"]
    ice = read("cave-ice-registration-v3/manifest.json")
    assert ice["status"] == "DRAFT_ICE_REGISTRATION_OFFLINE_ONLY" and len(ice["tiles"]) == 14 and len(ice["previews"]) == 3
    assert ice["script_sha256"] == sha(p/"tools/hd_register_ice.py")
    for tile in ice["tiles"]:
        assert sha(base/"cave-ice-registration-v3"/tile["registered"]) == tile["sha256"]
    for preview in ice["previews"]:
        assert sha(base/"cave-ice-registration-v3"/(preview["map"]+"-offline-hd.png")) == preview["hd_sha256"]
    for folder in ("cave-ice-art-v3", "cave-ice-material-v1"):
        generation = read(folder+"/generation.json")
        assert sha(base/folder/generation["output"]) == generation["sha256"]
        for source in generation["inputs"]:
            assert sha(p/Path(source["path"]).relative_to("/home/anr2/cht/zeliard")) == source["sha256"]
    files = ["internal/prototype/hd/cave_catalog.go","internal/prototype/hd/cave_catalog_data.go","internal/prototype/hd/cave_catalog_test.go",
             "internal/prototype/hd/cave_frames.go","tools/hd_compile_cave_catalog.py","tools/hd_register_ice.py",
             "tools/hd_cave_catalog_smoke.py","tools/hd_cave_catalog_audit.py"]
    proofs = ["cave-catalog-fixtures-v1/manifest.json","cave-catalog-runtime-v1/unit-tests.jsonl",
              "cave-catalog-runtime-v1/normal-and-refusal-tests.jsonl","cave-catalog-runtime-v1/vet.log","cave-catalog-runtime-v1/build.log",
              "cave-catalog-pair-v1/combat/receipt.json","cave-catalog-pair-v1/lifecycle-return-v1/receipt.json",
              "cave-catalog-gui-v3/receipt.json","cave-ice-art-v3/generation.json","cave-ice-material-v1/generation.json","cave-ice-registration-v3/manifest.json"]
    report = {"schema":1,"status":"PASS_CAVE_CATALOG_SCOPED","binary_sha256":sha(binary),"source_maps":31,"source_terrain_profiles":11,
              "normal_points":points,"gui_pairs":len(gui["pairs"]),"restored_game_pixels":restored_pixels,"changed_hd_pixels":changed_pixels,
              "source_files":[{"path":f,"sha256":sha(p/f)} for f in files],
              "proofs":[{"path":str((base/f).relative_to(p)),"sha256":sha(base/f)} for f in proofs],
              "limits":["正常路徑與實際視窗只驗MPP1；其餘來源以真實檔案的結構fixture驗識別，不當可達性或HD驗收。",
                        "冰窟僅三份離線預覽，15..19圖形文字及20..21保留缺口，尚未接正式前端。",
                        "全遊戲HD仍在進行，未更新v.0.2.0封包或推廣片。"]}
    assert not args.output.exists(), "不覆寫既有收據"
    args.output.write_text(json.dumps(report,ensure_ascii=False,indent=2)+"\n",encoding="utf-8")
    print("PASS_CAVE_CATALOG_SCOPED",points,"normal points",restored_pixels,"restored pixels")


if __name__ == "__main__":
    main()

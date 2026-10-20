#!/usr/bin/env python3
"""稽核四組洞窟圖面、結構測試與MPP1正常回歸；不宣稱後段可達。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image, ImageChops


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    p, b = args.project, args.project/"workplace/full-hd-v1"
    assert not args.output.exists(), "不覆寫歷史收據"
    sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
    read = lambda name: json.loads((b/name).read_text())
    profiles = {r["key"]: r for r in read("cave-terrain-sources-v1/receipt.json")["profiles"]}
    maps = {r["name"].removesuffix(".MDT"): r for r in read("cave-map-sources-v1/receipt.json")["records"]}
    assert len(profiles) == 11 and len(maps) == 31
    expected = {
        "MPP8": (81,43,[17,18,19],{"MP80","MP81","MP82","MP83","MP84","MP8D"}),
        "MPP9": (82,38,[],{"MP90"}),
        "MPPA": (83,20,[20],{"MPA0"}),
        "MPPB": (84,5,[],{"MP73"}),
    }
    proofs, registered = [], []
    for family, (record,count,blanks,names) in expected.items():
        prefix = "cave-"+family.lower()
        folder = b/(prefix+"-complete-v1")
        meta = read(prefix+"-complete-v1/manifest.json")
        source = profiles[family]
        assert source["record"] == record and source["tiles"] == count
        assert source["source_sha256"] == meta["source_sha256"] == sha(p/source["source"])
        assert meta["script_sha256"] == sha(p/"tools/hd_cave_family_art.py")
        assert len(meta["tiles"]) == count and meta.get("blank_ids",[]) == blanks
        raw = (b/"cave-catalog-fixtures-v1"/(family+".resource")).read_bytes()
        assert len(raw) == 48+count*48
        assert hashlib.sha256(raw[:48]).hexdigest() == source["header_sha256"]
        assert hashlib.sha256(raw[48:]).hexdigest() == source["packed_sha256"]
        for i, row in enumerate(meta["tiles"],1):
            assert row["id"] == i and row["native_rgba_sha256"] == source["tile_records"][i-1]["rgba_sha256"]
            assert sha(folder/row["registered"]) == row["sha256"]
            if i in blanks:
                assert raw[48+(i-1)*48:48+i*48] == bytes(48)
                assert row["kind"] == "source_empty"
                assert Image.open(folder/row["registered"]).getchannel("A").getbbox() is None
        for suffix,key in [("tiles","atlas_sha256"),("material","material_sha256")]+([("secondary-material","secondary_material_sha256")] if family=="MPPA" else []):
            image = folder/(family.lower()+"-"+suffix+"-v1.png")
            assert sha(image) == meta[key]
            assert Image.open(image).size == ((count*24,24) if suffix=="tiles" else (432,768))
        assert {r["map"] for r in meta["previews"]} == names
        for row in meta["previews"]:
            original = maps[row["map"]]
            assert original["family_candidate"] == family
            path = b/"cave-map-sources-v1"/original["columns"]
            assert sha(path) == original["columns_sha256"]
            columns = np.frombuffer(path.read_bytes(),dtype="uint8").reshape((original["width"],64))
            cx,cy = row["camera"]
            grid = bytes(int(columns[(cx+x)%original["width"],(cy+y)%64]) for y in range(18) for x in range(28))
            assert (folder/(row["map"]+"-grid.bin")).read_bytes() == grid
            assert row["native_fallback_cells"] == 0
            assert sha(folder/(row["map"]+"-offline-hd.png")) == row["hd_sha256"]
            if row["map"] in ("MP80","MP81"):
                ids = [29,30,31,32,33] if row["map"]=="MP80" else [34,35,36,13,14]
                assert any(grid[i:i+5]==bytes(ids) for i in range(len(grid)-4) if i%28<=23)
        generation = read(prefix+"-art-v1/generation.json")
        assert generation["mode"] == "builtin-imagegen"
        expected_images = {"MPP8":{"material","details_a","details_b","details_c","label","label2"},"MPP9":{"material","details_a","details_b"},"MPPA":{"material","secondary_material","details_a","details_b"},"MPPB":{"material","details_a"}}[family]
        assert {r["key"] for r in generation["outputs"]} == {r["key"] for r in generation["requests"]} == expected_images
        assert sha(b/(prefix+"-art-v1/generation.json")) == meta["generation_sha256"]
        for row in generation["outputs"]:
            assert sha(b/(prefix+"-art-v1")/row["file"]) == row["sha256"]
        for request in generation["requests"]:
            assert request["prompt"] and request["refs"]
            for inp in request["inputs"]:
                assert sha(p/inp["path"]) == inp["sha256"]
        reference = read(prefix+"-art-v1/reference-check.json")
        assert reference["status"] == "PASS_REFERENCES_AND_SOURCE_IDENTITY"
        assert reference["script_sha256"] == meta["script_sha256"]
        for row in reference["references"]:
            assert sha(b/(prefix+"-source-v1/references")/row["file"]) == row["sha256"]
        replay = read(prefix+"-complete-v1/offline-check.json")
        assert replay["status"] == "PASS_OFFLINE_ART_REPLAY_ONLY" and replay["registered_slots"] == count
        assert replay["visible_tiles"] == count-len(blanks) and replay["art_tool_sha256"] == meta["script_sha256"]
        for row in replay["replayed_files"]:
            assert sha(folder/row["file"]) == row["sha256"]
        proofs += [prefix+"-complete-v1/manifest.json",prefix+"-complete-v1/offline-check.json",prefix+"-art-v1/generation.json",prefix+"-art-v1/reference-check.json"]
        registered.append({"family":family,"slots":count,"visible":count-len(blanks),"maps":len(names)})
    labels = read("cave-mpp8-complete-v1/manifest.json")["labels"]
    assert [(r["text"],r["ids"]) for r in labels] == [("PUREZA",[29,30,31,32,33]),("ESCO",[34,35,36,13,14])]
    legacy = read("cave-mpp8-art-v1/legacy-reference-check.json")
    assert legacy["status"] == "PASS_LEGACY_REFERENCE_PNGS"
    assert legacy["script_sha256"] == sha(p/"tools/hd_cave_family_art.py")
    assert len(legacy["families"]) == 4
    for family in legacy["families"]:
        for row in family["unchanged_references"]:
            assert sha(p/row["path"]) == row["sha256"]
    prefix = "cave-final-families"
    required = {"TestCaveRemainingFamiliesIndependentStaticMaps","TestCaveRemainingFamiliesAllStructuralSlots","TestCaveRemainingFamiliesAtomicLoadAndRefusal",
                "TestCaveFramesNormalPathsReadOnly","TestCaveFramesRejectSourceAndTransitions","TestCaveSourcesRejectUnknownOrCorrupt"}
    for family in ("MPP3","MPP5","MPP6","MPP7","Peligro"):
        required |= {"TestCave"+family+suffix for suffix in ("IndependentStaticMaps","AllStructuralFrames","AtomicLoadAndRefusal")}
    required |= {"TestCaveIceIndependentStaticMaps","TestCaveIceStructuralOrbs","TestCaveIceAtomicLoadAndRefusal","TestCaveTerrainIndependentOfflinePixels","TestCaveTerrainAtomicFallback"}
    for kind in ("unit","normal"):
        path = b/(prefix+"-runtime-v1")/(kind+"-tests.jsonl")
        rows = [json.loads(line) for line in path.read_text().splitlines()]
        assert rows[-1]["Action"] == "pass" and not any(r["Action"] in ("fail","skip") for r in rows)
        required -= {r.get("Test") for r in rows if r["Action"] == "pass"}
        assert sum(r["Action"]=="pass" and bool(r.get("Test")) and "/" not in r["Test"] for r in rows) == (23 if kind=="unit" else 3)
    assert not required
    for name,count in (("combat",25),("lifecycle-return-v1",10)):
        pair = read(prefix+f"-pair-v1/{name}/receipt.json")
        assert pair["state_equal"] and pair["save_unchanged"] and len(pair["points"]) == count
        assert not pair["held_keys"] and not pair["wrote"] and not pair["unimplemented"] and pair["refused"]==0
    binary = b/(prefix+"-runtime-v1/zeliard-preview")
    gui = read(prefix+"-gui-v2/receipt.json")
    assert gui["status"] == "PASS_CAVE_CATALOG_GUI_SCOPED" and gui["normal_DOS_exit"]==0
    assert gui["binary_sha256"] == sha(binary) and gui["script_sha256"] == sha(p/"tools/hd_cave_catalog_smoke.py")
    assert gui["helper_sha256"] == sha(p/"tools/integration_smoke.py")
    command = gui["boots"][0]["command"]
    assert command[0] == str(binary)
    for family in expected:
        flag = "-cave-"+family.lower()+"-art"
        assert command[command.index(flag)+1] == gui["remaining_family_art_directories"][family] == str(b/("cave-"+family.lower()+"-complete-v1"))
    for family,folder in [("ice","cave-ice-complete-v1"),("peligro","cave-peligro-complete-v1"),("mpp3","cave-mpp3-complete-v2"),("mpp5","cave-mpp5-complete-v1"),("mpp6","cave-mpp6-complete-v2"),("mpp7","cave-mpp7-complete-v1")]:
        flag = "-cave-"+family+"-art"
        assert command[command.index(flag)+1] == str(b/folder)
    log = (b/(prefix+"-gui-v2/boot1/game.log")).read_text()
    assert "地形HD未載入" not in log
    shots = {r["name"]:r for r in gui["shots"]}
    changed = restored = 0
    assert len(gui["pairs"])==2
    for pair in gui["pairs"]:
        images,states = [],[]
        for key in ("original","hd","returned"):
            path = b/(prefix+"-gui-v2/boot1")/pair[key]
            shot = shots[path.stem]; assert sha(path)==shot["image_sha256"]
            im = Image.open(path).convert("RGB"); assert im.size==(960,680)
            images.append(im.crop((0,48,960,648))); states.append(shot["state"])
        assert ImageChops.difference(images[0],images[2]).getbbox() is None
        for key in ("steps","cycles","oracle_memory_sha256","oracle_indexed_sha256","oracle_palette_sha256"):
            assert states[0][key]==states[1][key]==states[2][key]
        assert states[1]["cave_frame_presentation"]["source_terrain_key"]=="MPP1"
        restored += 576000
        changed += sum(any(c) for c in ImageChops.difference(images[0],images[1]).getdata())
    assert changed>1000 and shots["returned-town"]["state"]["town_scenery_hd_matched"]
    prior_name = "cave-mpp7-final-audit-v1.json"
    prior = read(prior_name); assert prior["status"]=="PASS_MPP7_STRUCTURAL_AND_MPP1_REGRESSION"
    for row in prior["proofs"]:
        assert sha(p/row["path"])==row["sha256"]
    files = sorted({r["path"] for r in prior["source_files"]}|{"tools/hd_cave_remaining_audit.py","internal/prototype/hd/cave_remaining_families.go","internal/prototype/hd/cave_remaining_families_test.go"})
    proofs += [prior_name,"cave-mpp8-art-v1/legacy-reference-check.json","cave-mpp8-source-v1/label-review.json"]
    proofs += [prefix+"-runtime-v1/"+name for name in ("unit-tests.jsonl","normal-tests.jsonl","vet.log","build.log","build-info.txt","run.sh","docker-command.txt","gui-docker-command.txt","gui-docker-command-v2.txt")]
    proofs += [prefix+"-pair-v1/"+name+"/receipt.json" for name in ("combat","lifecycle-return-v1")]
    proofs += [prefix+"-gui-v2/receipt.json"]
    report = {"schema":1,"status":"PASS_REMAINING_CAVE_FAMILIES_STRUCTURAL_AND_MPP1_REGRESSION","binary_sha256":sha(binary),
              "audit_tool_sha256":sha(Path(__file__)),"source_maps":31,"source_terrain_profiles":11,"registered_families":registered,
              "registered_tiles":106,"visible_tiles":102,"new_structural_map_pixels":2612736,"normal_points":35,"gui_pairs":2,"restored_game_pixels":restored,"changed_hd_pixels":changed,
              "source_files":[{"path":f,"sha256":sha(p/f)} for f in files],
              "proofs":[{"path":str((b/f).relative_to(p)),"sha256":sha(b/f)} for f in proofs],
              "limits":["四組地形的結構快照不執行VM；正常到達、進出、捲動與事件仍未驗。","35點正常配對及實際視窗只驗MPP1回歸；不外推後段場景、敵人或首領。","全遊戲HD維持進行中，封包、版本與影片未更新。"]}
    args.output.write_text(json.dumps(report,ensure_ascii=False,indent=2)+"\n")
    print(report["status"],report["binary_sha256"])


if __name__ == "__main__":
    main()

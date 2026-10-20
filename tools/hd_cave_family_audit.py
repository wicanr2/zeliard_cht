#!/usr/bin/env python3
"""核新增洞窟圖集與目前入口；結構快照及正常玩家回歸分列。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image, ImageChops


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project",type=Path,required=True)
    ap.add_argument("--family",choices=("MPP2","MPP3","MPP5"),required=True)
    ap.add_argument("--output",type=Path,required=True)
    a = ap.parse_args()
    p,b = a.project,a.project/"workplace/full-hd-v1"
    sha = lambda path:hashlib.sha256(path.read_bytes()).hexdigest()
    read = lambda path:json.loads((b/path).read_text())
    assert not a.output.exists(),"不覆寫既有收據"
    cfg = {
        "MPP2":{"prefix":"cave-peligro","art_key":"peligro","complete":"cave-peligro-complete-v1","count":34,"record":75,
                "art_status":"MPP2_ART_REGISTERED_OFFLINE_ONLY","art_tool":"tools/hd_peligro_art.py","generation":"generation.json",
                "maps":{"MP20","MP21","MP2D"},"prior":"cave-ice-final-audit-v1.json","prior_status":"PASS_ICE_RUNTIME_STRUCTURAL_AND_MPP1_REGRESSION"},
        "MPP3":{"prefix":"cave-mpp3","art_key":"mpp3","complete":"cave-mpp3-complete-v2","count":43,"record":76,
                "art_status":"REGISTERED_FAMILY_ART_OFFLINE_ONLY","art_tool":"tools/hd_cave_family_art.py","generation":"generation-v2.json",
                "maps":{"MP30","MP31","MP3D"},"prior":"cave-peligro-final-audit-v1.json","prior_status":"PASS_MPP2_STRUCTURAL_AND_MPP1_REGRESSION"},
        "MPP5":{"prefix":"cave-mpp5","art_key":"mpp5","complete":"cave-mpp5-complete-v1","count":52,"record":78,
                "art_status":"REGISTERED_FAMILY_ART_OFFLINE_ONLY","art_tool":"tools/hd_cave_family_art.py","generation":"generation.json",
                "maps":{"MP50","MP51","MP5D"},"prior":"cave-mpp3-final-audit-v1.json","prior_status":"PASS_MPP3_STRUCTURAL_AND_MPP1_REGRESSION",
                "gui":"cave-mpp5-gui-v2"},
    }[a.family]
    prefix,complete = cfg["prefix"],cfg["complete"]
    gui_folder = cfg.get("gui",prefix+"-gui-v1")
    art = read(complete+"/manifest.json")
    assert art["status"] == cfg["art_status"] and len(art["tiles"]) == cfg["count"]
    assert art["script_sha256"] == sha(p/cfg["art_tool"])
    source = next(r for r in read("cave-terrain-sources-v1/receipt.json")["profiles"] if r["key"] == a.family)
    assert source["record"] == cfg["record"] and source["source_sha256"] == art["source_sha256"] == sha(p/source["source"])
    for i,row in enumerate(art["tiles"],1):
        assert row["id"] == i and row["native_rgba_sha256"] == source["tile_records"][i-1]["rgba_sha256"]
        assert sha(b/complete/row["registered"]) == row["sha256"]
    for file,key in ((cfg["art_key"]+"-tiles-v1.png","atlas_sha256"),(cfg["art_key"]+"-material-v1.png","material_sha256")):
        assert sha(b/complete/file) == art[key]
    maps = {r["name"].removesuffix(".MDT"):r for r in read("cave-map-sources-v1/receipt.json")["records"]}
    assert {r["map"] for r in art["previews"]} == cfg["maps"]
    for row in art["previews"]:
        name = row["map"]
        original = maps[name]
        assert original["family_candidate"] == a.family
        raw = b/"cave-map-sources-v1"/original["columns"]
        assert sha(raw) == original["columns_sha256"]
        columns = np.frombuffer(raw.read_bytes(),dtype="uint8").reshape((original["width"],64))
        cx,cy = row["camera"]
        grid = bytes(int(columns[(cx+x)%original["width"],(cy+y)%64]) for y in range(18) for x in range(28))
        grid_path = b/complete/(name+"-grid.bin")
        assert grid_path.read_bytes() == grid and sha(grid_path) == row["grid_sha256"]
        assert row["native_fallback_cells"] == 0
        assert sha(b/complete/(name+"-offline-hd.png")) == row["hd_sha256"]
        if a.family == "MPP3" and name in ("MP30","MP31"):
            assert any(grid[i:i+5] == bytes(range(34,39)) for i in range(len(grid)-4) if i%28 <= 23)
    generation_path = b/(prefix+"-art-v1")/cfg["generation"]
    generation = json.loads(generation_path.read_text())
    assert generation["mode"] == "builtin-imagegen" and sha(generation_path) == art["generation_sha256"]
    expected = {"material","details","text"} if a.family == "MPP2" else {"material","details_a","details_b","details_c","label"}
    if a.family == "MPP5":
        expected.remove("label")
        assert art["material_ids"] == [*range(3,13),24,25,*range(35,39),*range(41,45)]
        assert art["text_ids"] == [] and art["text"] == ""
    assert {r["key"] for r in generation["outputs"]} == expected
    if a.family == "MPP3":
        assert art["material_ids"] == [3,4,6,7,8,9,10,32,33] and art["text_ids"] == list(range(34,39)) and art["text"] == "BOSQUE"
    for row in generation["outputs"]:
        assert sha(generation_path.parent/row["file"]) == row["sha256"]
        assert row["alpha_extrema"] == ([255,255] if row["key"] == "material" else [0,255])
        for source_input in row["inputs"]:
            assert sha(p/Path(source_input["path"]).relative_to("/home/anr2/cht/zeliard")) == source_input["sha256"]
    rows = []
    for name in ("unit-tests.jsonl","normal-tests.jsonl"):
        rows += [json.loads(l) for l in (b/(prefix+"-runtime-v1")/name).read_text().splitlines()]
    assert not any(r["Action"] in ("skip","fail") for r in rows)
    passed = {r.get("Test") for r in rows if r["Action"] == "pass"}
    required = {"TestCavePeligroIndependentStaticMaps","TestCavePeligroAllStructuralFrames","TestCavePeligroAtomicLoadAndRefusal",
                "TestCaveIceIndependentStaticMaps","TestCaveIceStructuralOrbs","TestCaveIceAtomicLoadAndRefusal",
                "TestCaveTerrainIndependentOfflinePixels","TestCaveTerrainAtomicFallback","TestCaveFramesNormalPathsReadOnly",
                "TestCaveFramesRejectSourceAndTransitions","TestCaveSourcesRejectUnknownOrCorrupt"}
    if a.family in ("MPP3","MPP5"):
        required |= {"TestCaveMPP3IndependentStaticMaps","TestCaveMPP3AllStructuralFrames","TestCaveMPP3AtomicLoadAndRefusal"}
    if a.family == "MPP5":
        required |= {"TestCaveMPP5IndependentStaticMaps","TestCaveMPP5AllStructuralFrames","TestCaveMPP5AtomicLoadAndRefusal"}
    assert required <= passed
    for name,count in (("combat",25),("lifecycle-return-v1",10)):
        pair = read(prefix+f"-pair-v1/{name}/receipt.json")
        assert pair["state_equal"] and pair["save_unchanged"] and len(pair["points"]) == count
        assert not pair["held_keys"] and not pair["wrote"] and not pair["unimplemented"] and pair["refused"] == 0
    gui = read(gui_folder+"/receipt.json")
    binary = b/(prefix+"-runtime-v1/zeliard-preview")
    assert gui["status"] == "PASS_CAVE_CATALOG_GUI_SCOPED" and gui["normal_DOS_exit"] == 0
    assert gui["binary_sha256"] == sha(binary) and gui["script_sha256"] == sha(p/"tools/hd_cave_catalog_smoke.py")
    assert gui["helper_sha256"] == sha(p/"tools/integration_smoke.py")
    command = gui["boots"][0]["command"]
    assert command[0] == str(binary)
    bindings = [("-cave-ice-art","ice_art_directory","cave-ice-complete-v1"),("-cave-peligro-art","peligro_art_directory","cave-peligro-complete-v1")]
    if a.family in ("MPP3","MPP5"):
        bindings.append(("-cave-mpp3-art","mpp3_art_directory","cave-mpp3-complete-v2"))
    if a.family == "MPP5":
        bindings.append(("-cave-mpp5-art","mpp5_art_directory",complete))
    for key,field,folder in bindings:
        assert gui[field] == command[command.index(key)+1] == str(b/folder)
    log = (b/(gui_folder+"/boot1/game.log")).read_text()
    assert "冰窟地形HD未載入" not in log and "MPP2地形HD未載入" not in log
    assert "MPP3地形HD未載入" not in log
    assert "MPP5地形HD未載入" not in log
    shots = {r["name"]:r for r in gui["shots"]}
    restored = changed = 0
    for pair in gui["pairs"]:
        ims,states = [],[]
        for key in ("original","hd","returned"):
            path = b/(gui_folder+"/boot1")/pair[key]
            item = shots[path.stem]
            assert sha(path) == item["image_sha256"]
            im = Image.open(path).convert("RGB")
            assert im.size == (960,680)
            ims.append(im.crop((0,48,960,648)))
            states.append(item["state"])
        for key in ("steps","cycles","oracle_memory_sha256","oracle_indexed_sha256","oracle_palette_sha256"):
            assert states[0][key] == states[1][key] == states[2][key]
        assert ImageChops.difference(ims[0],ims[2]).getbbox() is None
        assert states[1]["cave_frame_presentation"]["source_terrain_key"] == "MPP1"
        restored += 576000
        changed += sum(any(c) for c in ImageChops.difference(ims[0],ims[1]).getdata())
    assert len(gui["pairs"]) == 2 and changed > 1000
    assert shots["returned-town"]["state"]["town_scenery_hd_matched"]
    prior = read(cfg["prior"])
    assert prior["status"] == cfg["prior_status"]
    for row in prior["proofs"]:
        assert sha(p/row["path"]) == row["sha256"]
    files = sorted({r["path"] for r in prior["source_files"]} | {"tools/hd_peligro_art.py","tools/hd_cave_family_audit.py",
                   "internal/prototype/hd/cave_peligro.go","internal/prototype/hd/cave_peligro_test.go"})
    if a.family in ("MPP3","MPP5"):
        files = sorted(set(files)|{cfg["art_tool"],"internal/prototype/hd/cave_mpp3.go","internal/prototype/hd/cave_mpp3_test.go"})
        if a.family == "MPP5":
            files = sorted(set(files)|{"internal/prototype/hd/cave_mpp5.go","internal/prototype/hd/cave_mpp5_test.go"})
        replay = read(prefix+"-reference-replay-v1/check.json")
        assert replay["status"] == "PASS_SOURCE_REFERENCE_REPLAY" and len(replay["references"]) == (4 if a.family == "MPP5" else 5)
        assert replay["script_sha256"] == sha(p/cfg["art_tool"])
        reference_folder = prefix+"-source-v1"+("/references" if a.family == "MPP5" else "")
        for row in replay["references"]:
            assert sha(b/reference_folder/row["file"]) == row["sha256"] == sha(b/(prefix+"-reference-replay-v1")/row["file"])
    proofs = [cfg["prior"],complete+"/manifest.json",prefix+"-art-v1/"+cfg["generation"],
              prefix+"-runtime-v1/unit-tests.jsonl",prefix+"-runtime-v1/normal-tests.jsonl",prefix+"-runtime-v1/vet.log",
              prefix+"-runtime-v1/build.log",prefix+"-runtime-v1/build-info.txt",prefix+"-pair-v1/combat/receipt.json",
              prefix+"-pair-v1/lifecycle-return-v1/receipt.json",gui_folder+"/receipt.json"]
    if a.family in ("MPP3","MPP5"):
        proofs.append(prefix+"-reference-replay-v1/check.json")
    report = {"schema":1,"status":"PASS_"+a.family+"_STRUCTURAL_AND_MPP1_REGRESSION","binary_sha256":sha(binary),
              "audit_tool_sha256":sha(Path(__file__)),
              "source_maps":31,"source_terrain_profiles":11,"registered_family":a.family,"registered_tiles":cfg["count"],
              "new_structural_map_pixels":870912,"ice_structural_map_pixels":870912,"normal_points":35,
              "gui_pairs":2,"restored_game_pixels":restored,"changed_hd_pixels":changed,
              "source_files":[{"path":f,"sha256":sha(p/f)} for f in files],
              "proofs":[{"path":str((b/f).relative_to(p)),"sha256":sha(b/f)} for f in proofs],
              "limits":["新增家族及冰窟像素為合成結構快照，未執行VM；不當正常玩家可達證據。",
                        "正常35點及實際視窗均為MPP1回歸；新增家族正常進出、捲動、事件與切回未驗。",
                        "其他家族與全遊戲HD未完成，未更新封包或影片。"]}
    a.output.write_text(json.dumps(report,ensure_ascii=False,indent=2)+"\n",encoding="utf-8")
    print(report["status"],report["binary_sha256"])


if __name__ == "__main__":
    main()

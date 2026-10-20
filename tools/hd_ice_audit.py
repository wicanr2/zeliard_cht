#!/usr/bin/env python3
"""核冰窟圖集、結構像素與MPP1正常回歸，明確保留正常冰窟未驗限制。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageChops


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    a = ap.parse_args()
    p = a.project
    b = p/"workplace/full-hd-v1"
    sha = lambda f: hashlib.sha256(f.read_bytes()).hexdigest()
    read = lambda f: json.loads((b/f).read_text())
    assert not a.output.exists(), "不覆寫收據"
    art = read("cave-ice-complete-v1/manifest.json")
    assert art["status"] == "COMPLETE_MPP4_ART_OFFLINE_ONLY"
    assert [r["id"] for r in art["tiles"]] == list(range(1,22))
    assert art["script_sha256"] == sha(p/"tools/hd_complete_ice.py")
    for r in art["tiles"]:
        assert sha(b/"cave-ice-complete-v1"/r["registered"]) == r["sha256"]
    for r in art["previews"]:
        assert r["native_fallback_cells"] == 0
        assert sha(b/"cave-ice-complete-v1"/(r["map"]+"-offline-hd.png")) == r["hd_sha256"]
        assert sha(b/"cave-ice-complete-v1"/(r["map"]+"-grid.bin")) == r["grid_sha256"]
    for name,key in (("ice-tiles-v1.png","atlas_sha256"),("ice-material-v1.png","material_sha256")):
        assert sha(b/"cave-ice-complete-v1"/name) == art[key]
    generation = read("cave-ice-details-v1/generation.json")
    assert generation["mode"] == "builtin-imagegen" and generation["transparent_background"]
    assert generation["output"]["sha256"] == art["details_sha256"] == sha(b/"cave-ice-details-v1"/generation["output"]["file"])
    assert generation["output"]["alpha_extrema"] == [0,255]
    for source in generation["inputs"]:
        assert sha(p/Path(source["path"]).relative_to("/home/anr2/cht/zeliard")) == source["sha256"]
    rows = [json.loads(l) for l in (b/"cave-ice-runtime-v1/checks-v2.jsonl").read_text().splitlines()]
    assert rows[-1]["Action"] == "pass" and not any(r["Action"] in ("skip","fail") for r in rows)
    passed = {r.get("Test") for r in rows if r["Action"] == "pass"}
    assert {"TestCaveIceIndependentStaticMaps", "TestCaveIceStructuralOrbs", "TestCaveIceAtomicLoadAndRefusal",
            "TestCaveTerrainIndependentOfflinePixels", "TestCaveTerrainAtomicFallback", "TestCaveFramesNormalPathsReadOnly",
            "TestCaveFramesRejectSourceAndTransitions", "TestCaveSourcesRejectUnknownOrCorrupt"} <= passed
    for name,count in (("combat",25),("lifecycle-return-v1",10)):
        pair = read(f"cave-ice-pair-v1/{name}/receipt.json")
        assert pair["state_equal"] and pair["save_unchanged"] and len(pair["points"]) == count
        assert not pair["held_keys"] and not pair["wrote"] and not pair["unimplemented"] and pair["refused"] == 0
    gui = read("cave-ice-gui-v1/receipt.json")
    binary = b/"cave-ice-runtime-v1/zeliard-preview"
    assert gui["status"] == "PASS_CAVE_CATALOG_GUI_SCOPED" and gui["normal_DOS_exit"] == 0
    assert gui["binary_sha256"] == sha(binary)
    assert gui["script_sha256"] == sha(p/"tools/hd_cave_catalog_smoke.py")
    assert gui["helper_sha256"] == sha(p/"tools/integration_smoke.py")
    assert gui["ice_art_directory"] == "/project/workplace/full-hd-v1/cave-ice-complete-v1"
    command = gui["boots"][0]["command"]
    assert command[0] == "/project/workplace/full-hd-v1/cave-ice-runtime-v1/zeliard-preview"
    assert command[command.index("-cave-ice-art")+1] == gui["ice_art_directory"]
    logs = (b/"cave-ice-gui-v1/boot1/game.log").read_text()
    assert "冰窟地形HD未載入" not in logs
    shots = {r["name"]:r for r in gui["shots"]}
    restored = changed = 0
    for pair in gui["pairs"]:
        ims, states = [], []
        for key in ("original","hd","returned"):
            path = b/"cave-ice-gui-v1/boot1"/pair[key]
            item = shots[path.stem]
            assert sha(path) == item["image_sha256"]
            im = Image.open(path).convert("RGB")
            assert im.size == (960,680)
            ims.append(im.crop((0,48,960,648)))
            states.append(item["state"])
        for key in ("steps","cycles","oracle_memory_sha256","oracle_indexed_sha256","oracle_palette_sha256"):
            assert states[0][key] == states[1][key] == states[2][key]
        assert ImageChops.difference(ims[0],ims[2]).getbbox() is None
        restored += 576000
        changed += sum(any(c) for c in ImageChops.difference(ims[0],ims[1]).getdata())
        assert states[1]["cave_frame_presentation"]["source_terrain_key"] == "MPP1"
    assert len(gui["pairs"]) == 2 and changed > 1000
    assert shots["returned-town"]["state"]["town_scenery_hd_matched"]
    old = read("cave-catalog-final-audit-v1.json")
    assert old["status"] == "PASS_CAVE_CATALOG_SCOPED"
    # 舊收據保留原版身份，只核其歷史證據未變；目前來源另以本輪source_files固定。
    for source in old["proofs"]:
        assert sha(p/source["path"]) == source["sha256"]
    files = sorted({r["path"] for r in old["source_files"]} | {
        "cmd/zeliard-preview/main.go", "internal/prototype/hd/cave_ice.go", "internal/prototype/hd/cave_ice_test.go",
        "internal/prototype/hd/cave_terrain.go", "internal/prototype/hd/cave_terrain_test.go",
        "internal/prototype/hd/cave_frames_pair_test.go", "tools/hd_complete_ice.py", "tools/hd_ice_audit.py"})
    proofs = ["cave-catalog-final-audit-v1.json","cave-ice-complete-v1/manifest.json","cave-ice-details-v1/generation.json",
              "cave-ice-runtime-v1/checks-v2.jsonl","cave-ice-runtime-v1/vet.log","cave-ice-runtime-v1/build.log","cave-ice-runtime-v1/build-info.txt",
              "cave-ice-pair-v1/combat/receipt.json","cave-ice-pair-v1/lifecycle-return-v1/receipt.json","cave-ice-gui-v1/receipt.json"]
    report = {"schema":1,"status":"PASS_ICE_RUNTIME_STRUCTURAL_AND_MPP1_REGRESSION",
              "binary_sha256":sha(binary),"source_maps":31,"source_terrain_profiles":11,"ice_art_tiles":21,
              "ice_structural_map_pixels":3*672*432,"normal_points":35,"gui_pairs":2,"restored_game_pixels":restored,
              "source_files":[{"path":f,"sha256":sha(p/f)} for f in files],
              "proofs":[{"path":str((b/f).relative_to(p)),"sha256":sha(b/f)} for f in proofs],
              "limits":["冰窟圖集已接可選前端；三圖合成結構快照非正常冰窟玩家路徑。",
                        "正常35點與實際視窗均為MPP1回歸，不能稱冰窟進出／捲動／事件已驗。",
                        "20與21只驗結構圖面及渲染；正常用途未證實。",
                        "其餘家族、敵人、首領、介面與全遊戲HD未完成；未更新封包或影片。"]}
    a.output.write_text(json.dumps(report,ensure_ascii=False,indent=2)+"\n",encoding="utf-8")
    print(report["status"],report["binary_sha256"])


if __name__ == "__main__":
    main()

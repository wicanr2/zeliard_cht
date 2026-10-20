#!/usr/bin/env python3
"""核MPP1補充圖面、十一家族共存與正常MPP1回歸，保留後段限制。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image, ImageChops


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project",type=Path,required=True)
    ap.add_argument("--output",type=Path,required=True)
    args=ap.parse_args(); assert not args.output.exists()
    p=args.project; b=p/"workplace/full-hd-v1"
    sha=lambda path:hashlib.sha256(path.read_bytes()).hexdigest()
    read=lambda name:json.loads((b/name).read_text())
    meta=read("cave-mpp1-complete-v2/manifest.json")
    source=next(r for r in read("cave-terrain-sources-v1/receipt.json")["profiles"] if r["key"]=="MPP1")
    assert source["record"]==74 and source["tiles"]==25
    assert source["source_sha256"]==meta["source_sha256"]==sha(p/source["source"])
    assert meta["script_sha256"]==sha(p/"tools/hd_mpp1_art.py")
    assert [r["id"] for r in meta["tiles"]]==list(range(15,26))
    folder=b/"cave-mpp1-complete-v2"
    assert sha(folder/"mpp1-extra-tiles-v1.png")==meta["atlas_sha256"]
    assert Image.open(folder/"mpp1-extra-tiles-v1.png").size==(264,24)
    for row in meta["tiles"]:
        assert row["native_rgba_sha256"]==source["tile_records"][row["id"]-1]["rgba_sha256"]
        path=folder/row["registered"]; assert sha(path)==row["sha256"]
        assert Image.open(path).getchannel("A").getbbox()
    assert [(r["text"],r["ids"]) for r in meta["labels"]]==[("SATONO",list(range(16,21))),("MURALLA",list(range(21,26)))]
    generation=read("cave-mpp1-art-v1/generation-v2.json")
    assert sha(b/"cave-mpp1-art-v1/generation-v2.json")==meta["generation_sha256"]
    assert generation["mode"]=="builtin-imagegen"
    assert {r["key"] for r in generation["outputs"]}=={"detail_15","satono","muralla"}
    for row in generation["outputs"]:
        assert sha(b/"cave-mpp1-art-v1"/row["file"])==row["sha256"]
    for row in generation["requests"]:
        assert row["prompt"] and row["refs"]
        for inp in row["inputs"]:
            assert sha(p/inp["path"])==inp["sha256"]
    for name,key in [("terrain-handpaint-v2.png","legacy_art_sha256"),("rock-material-v1.png","legacy_material_sha256")]:
        assert sha(p/"workplace/cave-hd-v1"/name)==meta[key]
    reference=read("cave-mpp1-art-v1/reference-check-v2.json")
    assert reference["status"]=="PASS_MPP1_SOURCE_REFERENCES" and reference["script_sha256"]==meta["script_sha256"]
    for row in reference["references"]:
        assert sha(b/"cave-mpp1-source-v1/references"/row["file"])==row["sha256"]
    shared=read("cave-mpp1-source-v1/normal-common-evidence.json")
    assert shared["status"]=="PASS_MPP1_ORIGINAL_PORTAL_SOURCE"
    for row in shared["evidence"]:
        assert sha(p/row["file"])==row["sha256"] and sha(p/row["body_file"])==row["body_sha256"]
        assert row["resource_pack_sha256"]=="3fd046b8c1281a013bfe1c8dfa627a663fec2eac5cebe7e3df025a63c1db6633" and row["positions"]
    replay=read("cave-mpp1-complete-v2/offline-check.json")
    assert replay["status"]=="PASS_MPP1_OFFLINE_REPLAY" and replay["script_sha256"]==meta["script_sha256"]
    for row in replay["files"]:
        assert sha(folder/row["file"])==row["sha256"]
    maps={r["name"].removesuffix(".MDT"):r for r in read("cave-map-sources-v1/receipt.json")["records"]}
    assert {r["tag"] for r in meta["previews"]}=={"MP10-Satono","MP10-Muralla","MP1D"}
    for row in meta["previews"]:
        m=maps[row["map"]]; path=b/"cave-map-sources-v1"/m["columns"]; assert sha(path)==m["columns_sha256"]
        columns=np.frombuffer(path.read_bytes(),dtype="uint8").reshape((m["width"],64)); cx,cy=row["camera"]
        grid=bytes(int(columns[(cx+x)%m["width"],(cy+y)%64]) for y in range(18) for x in range(28))
        assert (folder/(row["tag"]+"-grid.bin")).read_bytes()==grid
        assert sha(folder/(row["tag"]+"-offline-hd.png"))==row["hd_sha256"] and row["native_fallback_cells"]==0
        if row["tag"].startswith("MP10-"):
            ids=range(16,21) if row["tag"].endswith("Satono") else range(21,26)
            assert any(grid[i:i+5]==bytes(ids) for i in range(len(grid)-4) if i%28<=23)
    for kind,count in [("unit",26),("normal",3)]:
        rows=[json.loads(l) for l in (b/"cave-mpp1-runtime-v2"/(kind+"-tests.jsonl")).read_text().splitlines()]
        assert rows[-1]["Action"]=="pass" and not any(r["Action"] in ("fail","skip") for r in rows)
        passed={r["Test"] for r in rows if r["Action"]=="pass" and r.get("Test") and "/" not in r["Test"]}
        assert len(passed)==count
        if kind=="unit":
            assert {"TestCaveMPP1IndependentStaticMaps","TestCaveMPP1LegacyCompatibilityAndSlots","TestCaveMPP1AtomicLoadAndRefusal","TestCaveRemainingFamiliesIndependentStaticMaps"}<=passed
    for name,count in [("combat",25),("lifecycle-return-v1",10)]:
        pair=read("cave-mpp1-pair-v2/"+name+"/receipt.json")
        assert pair["state_equal"] and pair["save_unchanged"] and len(pair["points"])==count and pair["refused"]==0
        assert not pair["held_keys"] and not pair["wrote"] and not pair["unimplemented"]
    gui=read("cave-mpp1-gui-v1/receipt.json"); binary=b/"cave-mpp1-runtime-v2/zeliard-preview"
    assert gui["status"]=="PASS_CAVE_CATALOG_GUI_SCOPED" and gui["normal_DOS_exit"]==0 and gui["binary_sha256"]==sha(binary)
    assert gui["script_sha256"]==sha(p/"tools/hd_cave_catalog_smoke.py") and gui["helper_sha256"]==sha(p/"tools/integration_smoke.py")
    command=gui["boots"][0]["command"]; assert command[0]==str(binary)
    families={"mpp1":"cave-mpp1-complete-v2","ice":"cave-ice-complete-v1","peligro":"cave-peligro-complete-v1","mpp3":"cave-mpp3-complete-v2","mpp5":"cave-mpp5-complete-v1","mpp6":"cave-mpp6-complete-v2","mpp7":"cave-mpp7-complete-v1","mpp8":"cave-mpp8-complete-v1","mpp9":"cave-mpp9-complete-v1","mppa":"cave-mppa-complete-v1","mppb":"cave-mppb-complete-v1"}
    for key,dir in families.items():
        flag="-cave-"+key+"-art"; assert command[command.index(flag)+1]==str(b/dir)
    shots={r["name"]:r for r in gui["shots"]}; assert len(gui["pairs"])==2
    restored=changed=0
    for pair in gui["pairs"]:
        images,states=[],[]
        for key in ("original","hd","returned"):
            path=b/"cave-mpp1-gui-v1/boot1"/pair[key]; item=shots[path.stem]
            assert sha(path)==item["image_sha256"]
            im=Image.open(path).convert("RGB"); assert im.size==(960,680)
            images.append(im.crop((0,48,960,648))); states.append(item["state"])
        assert ImageChops.difference(images[0],images[2]).getbbox() is None
        for key in ("steps","cycles","oracle_memory_sha256","oracle_indexed_sha256","oracle_palette_sha256"):
            assert states[0][key]==states[1][key]==states[2][key]
        assert states[1]["cave_frame_presentation"]["source_terrain_key"]=="MPP1"
        restored+=576000; changed+=sum(any(c) for c in ImageChops.difference(images[0],images[1]).getdata())
    assert changed>1000 and shots["returned-town"]["state"]["town_scenery_hd_matched"]
    assert "地形HD未載入" not in (b/"cave-mpp1-gui-v1/boot1/game.log").read_text()
    prior=read("cave-final-families-audit-v1.json")
    assert prior["status"]=="PASS_REMAINING_CAVE_FAMILIES_STRUCTURAL_AND_MPP1_REGRESSION"
    for row in prior["proofs"]:
        assert sha(p/row["path"])==row["sha256"]
    files=sorted({r["path"] for r in prior["source_files"]}|{"tools/hd_mpp1_art.py","tools/hd_mpp1_audit.py","internal/prototype/hd/cave_mpp1.go","internal/prototype/hd/cave_mpp1_test.go"})
    proofs=["cave-final-families-audit-v1.json","cave-mpp1-complete-v2/manifest.json","cave-mpp1-complete-v2/offline-check.json","cave-mpp1-art-v1/generation-v2.json","cave-mpp1-art-v1/reference-check-v2.json","cave-mpp1-source-v1/source-review.json","cave-mpp1-source-v1/spelling-reference.json","cave-mpp1-gui-v1/receipt.json"]
    proofs += ["cave-mpp1-runtime-v2/"+name for name in ["unit-tests.jsonl","normal-tests.jsonl","vet.log","build.log","build-info.txt","run.sh","docker-command.txt","gui-docker-command.txt"]]
    proofs += ["cave-mpp1-pair-v2/"+name+"/receipt.json" for name in ["combat","lifecycle-return-v1"]]
    proofs += ["cave-mpp1-source-v1/normal-common-evidence.json","cave-mpp1-runtime-v1/normal-tests.jsonl"]
    report={"schema":1,"status":"PASS_MPP1_EXTRAS_AND_ELEVEN_TERRAIN_FAMILIES","binary_sha256":sha(binary),"audit_tool_sha256":sha(Path(__file__)),"source_maps":31,"source_terrain_profiles":11,"registered_tiles":378,"registered_mpp1_tiles":25,"new_tiles":11,"new_structural_map_pixels":870912,"normal_points":35,"gui_pairs":2,"restored_game_pixels":restored,"changed_hd_pixels":changed,
            "source_files":[{"path":f,"sha256":sha(p/f)} for f in files],"proofs":[{"path":str((b/f).relative_to(p)),"sha256":sha(b/f)} for f in proofs],
            "limits":["378是地形來源槽數，含原版空白；不代表全洞窟人物、敵人或首領完成。","Satono圖牌及後段家族只做結構構圖；正常新家族到達、捲動與事件仍未驗。","正常35點與GUI只驗MPP1首戰及回城，未更新封包、版號或影片。"]}
    args.output.write_text(json.dumps(report,ensure_ascii=False,indent=2)+"\n"); print(report["status"],report["binary_sha256"])


if __name__=="__main__":
    main()

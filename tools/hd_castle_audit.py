#!/usr/bin/env python3
"""Docker內稽核城堡GUI、原貌還原與旗幟來源像素，保留限定完成範圍。"""
import argparse
import hashlib
import json
import subprocess
from pathlib import Path

from PIL import Image, ImageChops


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    p = args.project
    study = p/"workplace/full-hd-v1"
    sha = lambda b: hashlib.sha256(b).hexdigest()
    read = lambda f: json.loads((study/f).read_text())
    gui = read("castle-gui-v2/receipt.json")
    pair = read("castle-pair-v6/receipt.json")
    assert gui["status"] == "PASS_CASTLE_HD_GUI_SCOPED" and gui["normal_DOS_exit"] == 0
    assert gui["flag_ids"] == [2, 3, 4, 5]
    assert len(pair["checkpoints"]) == 13 and all(x["state_equal"] for x in pair["checkpoints"])
    assert pair["uncovered"] == 0 and not pair["held_keys"]
    assert sum(x["presentation"].get("used", False) for x in pair["checkpoints"]) == 11
    binary = study/"castle-runtime-v3/zeliard-preview"
    assert sha(binary.read_bytes()) == gui["binary_sha256"]
    assert sha((p/"tools/hd_castle_smoke.py").read_bytes()) == gui["script_sha256"]
    source_tests = [json.loads(line) for line in (study/"castle-runtime-v3/final-unit-tests-v2.jsonl").read_text().splitlines()]
    units = ["TestCastleNormalSourcesAndLayers", "TestCastleRefusalIsAtomic", "TestCastlePresentationCrossProfile",
             "TestCastlePlayerHiddenAtRightBoundary", "TestTownActorsAllSources", "TestTownActorsAtomicFallback",
             "TestTownActorsLayerOrder", "TestTownActorsMissingAssets"]
    assert all(any(x["Action"] == "pass" and x.get("Test") == name for x in source_tests) for name in units)
    assert not any(x["Action"] in ("fail", "skip") for x in source_tests)
    # 舊驗收失敗保留；最終正常路徑只取v6，不把舊FAIL改稱PASS。
    final_tests = [json.loads(line) for line in (study/"castle-pair-v6/test.jsonl").read_text().splitlines()]
    assert any(x["Action"] == "pass" and x.get("Test") == "TestCastleNormalPathsReadOnly" for x in final_tests)
    assert not any(x["Action"] in ("fail", "skip") for x in final_tests)
    regression = [json.loads(line) for line in (study/"town-regression-after-castle-v1/test.jsonl").read_text().splitlines()]
    for name in ("TestTownActorsNormalPathsReadOnly", "TestTownSceneryOriginalFramesAndActorIsolation"):
        assert any(x["Action"] == "pass" and x.get("Test") == name for x in regression)
    assert not any(x["Action"] in ("fail", "skip") for x in regression)
    normal = read("castle-normal-v2/receipt.json")
    assert normal["status"] == "PASS_NORMAL_CASTLE_OBSERVATION" and normal["observer_state_equal"]
    assert len(normal["checkpoints"]) == 13 and all(x["state_equal"] for x in normal["checkpoints"])
    assert normal["source_index_sha256"] == sha((p/"docs/castle-source-index.json").read_bytes())
    assert {(x["before"], x["selected"]) for x in normal["update_events"]} == {(2, 3), (3, 4), (4, 5), (5, 2)}
    assert all(x["after"] == x["selected"] and x["semantic_index"]["level"] == "已證實"
               and x["write_semantic_index"]["level"] == "已證實" for x in normal["update_events"])
    registration = read("castle-registration-v1/manifest.json")
    for row in registration["flags"]:
        assert sha((study/"castle-registration-v1"/row["file"]).read_bytes()) == row["sha256"]
    background = Image.open(p/"workplace/music-release-v1/release-build/data/assets/town/background-handpaint-v2.png").convert("RGBA")
    assert background.size == (2126, 740)
    flags = {i: Image.open(study/f"castle-registration-v1/castle-flag-{i}-hd-v1.png").convert("RGBA") for i in range(2, 6)}
    shots = {x["name"]: x for x in gui["shots"]}
    totals = {"restored_game_pixels": 0, "opaque_flag_pixels": 0, "transparent_background_pixels": 0, "partial_alpha_pixels": 0}
    counts = {i: 0 for i in range(2, 6)}
    comparisons = []
    for group in gui["pairs"]:
        images, states = [], []
        for key in ("original", "hd", "returned"):
            path = study/"castle-gui-v2/boot1"/group[key]
            meta = shots[path.stem]
            assert sha(path.read_bytes()) == meta["image_sha256"]
            im = Image.open(path).convert("RGB")
            assert im.size == (960, 680)
            images.append(im)
            states.append(meta["state"])
        for key in ("steps", "cycles", "oracle_memory_sha256", "oracle_indexed_sha256", "oracle_palette_sha256"):
            assert states[0][key] == states[1][key] == states[2][key]
        assert ImageChops.difference(images[0].crop((0, 48, 960, 648)), images[2].crop((0, 48, 960, 648))).getbbox() is None
        totals["restored_game_pixels"] += 576000
        assert states[1]["town_frame_presentation"]["profile"] == group["profile"]
        assert states[1]["town_scenery_hd_matched"] and states[1]["town_residents_hd_matched"]
        for cell in group["flag_cells"]:
            col, row, tile = cell["column"], cell["row"], cell["id"]
            sprite = flags[tile]
            for y in range(24):
                for x in range(24):
                    red, green, blue, alpha = sprite.getpixel((x, y))
                    actual = images[1].getpixel((144+col*24+x, 48+234+row*24+y))
                    if alpha == 255:
                        assert actual == (red, green, blue), (group["name"], tile, x, y, actual)
                        totals["opaque_flag_pixels"] += 1
                        counts[tile] += 1
                    elif alpha == 0:
                        source_x = (col*24+x)*2126//672
                        source_y = ((54+row*8)*3+y)*740//234
                        expected = background.getpixel((source_x, source_y))
                        assert expected[3] == 255 and actual == expected[:3], (group["name"], "background", tile, x, y)
                        totals["transparent_background_pixels"] += 1
                    else:
                        totals["partial_alpha_pixels"] += 1
        comparisons.append({"name": group["name"], "profile": group["profile"], "native_state_frozen": True, "game_restored": True})
    assert all(counts.values()) and len(comparisons) >= 7
    movie = study/"castle-gui-v2"/gui["animation_video"]
    probe = json.loads(subprocess.check_output(["ffprobe", "-v", "error", "-show_streams", "-show_format", "-of", "json", str(movie)], timeout=15))
    video = [x for x in probe["streams"] if x["codec_type"] == "video"]
    assert len(video) == 1 and video[0]["width"] == 960 and video[0]["height"] == 680
    assert 11.9 <= float(probe["format"]["duration"]) <= 12.2 and movie.stat().st_size > 10000
    sources = ["internal/prototype/hd/castle.go", "internal/prototype/hd/castle_test.go", "internal/prototype/hd/town_scenery.go",
               "internal/prototype/hd/town_player.go", "internal/prototype/hd/town_actors.go", "internal/prototype/hd/town_frames.go",
               "internal/prototype/hd/town_ground.go", "internal/prototype/hd/town_background.go", "internal/prototype/hd/town_frames_pair_test.go",
               "cmd/zeliard-preview/main.go", "tools/hd_castle_probe.go", "tools/hd_castle_sources.py", "tools/hd_castle_register.py",
               "tools/hd_castle_smoke.py", "tools/hd_castle_audit.py", "tools/ida_castle_evidence.py", "tools/castle-hd.commands", "docs/castle-source-index.json"]
    report = {"schema": 1, "status": "PASS_CASTLE_HD_SCOPED", "spec": "docs/spec/036-castle-hd.md",
              "binary_sha256": gui["binary_sha256"], "comparisons": comparisons, "pixels": totals, "flag_opaque_by_id": counts,
              "source_files": [{"path": f, "sha256": sha((p/f).read_bytes())} for f in sources],
              "proofs": [{"path": str((study/f).relative_to(p)), "sha256": sha((study/f).read_bytes())} for f in
                         ["castle-gui-v2/receipt.json", "castle-pair-v6/receipt.json", "castle-pair-v6/test.jsonl",
                          "castle-runtime-v3/final-unit-tests-v2.jsonl", "town-regression-after-castle-v1/test.jsonl",
                          "castle-registration-v1/manifest.json", "castle-edge-ida-v2/evidence.json", "castle-normal-v2/receipt.json"]],
              "video": {"path": str(movie.relative_to(p)), "sha256": sha(movie.read_bytes()), "duration": probe["format"]["duration"], "audio": "off，僅驗畫面"},
              "limits": ["只驗036列城堡與Muralla往返，不代表其餘場景HD完成或全遊戲已通關。",
                         "獨立PNG像素核對限不透明旗幟與全透明背景；中間alpha由Go合成測試及目視另驗。",
                         "原版轉場及新場景第一個完整來源前維持原版；既有phase4與交談背景HD缺口仍在。",
                         "未更新v.0.2.0三平台封包或推廣片。"]}
    assert not args.output.exists(), "保留舊收據身份"
    args.output.write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
    print(report["status"], len(comparisons), "pairs", totals)


if __name__ == "__main__":
    main()

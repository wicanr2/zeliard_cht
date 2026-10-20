#!/usr/bin/env python3
"""Docker內編譯城堡手繪到3倍世界圖面與四旗幟姿態；不修改原版或宣稱已接入。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image

PANELS = [
    {"world_x": [0, 25, 32, 160, 176, 192, 208, 224, 240],
     "source_x": [0, 90, 137, 805, 877, 979, 1031, 1147, 1215],
     "world_y": [0, 8, 16, 24, 32, 64], "source_y": [0, 55, 149, 188, 235, 364]},
    {"world_x": [224, 238, 352, 364, 407, 421, 427, 441, 464],
     "source_x": [0, 46, 650, 704, 899, 988, 1041, 1138, 1215],
     "world_y": [0, 8, 16, 24, 32, 64], "source_y": [365, 407, 444, 517, 556, 694]},
    {"world_x": [448, 480, 496, 608, 656, 672, 688],
     "source_x": [0, 164, 211, 814, 1033, 1148, 1215],
     "world_y": [0, 8, 16, 24, 32, 64], "source_y": [695, 743, 776, 849, 885, 1015]},
    {"world_x": [672, 720, 736, 752, 765, 774, 790, 800, 832, 888, 912],
     "source_x": [0, 260, 370, 468, 514, 584, 627, 705, 816, 1100, 1215],
     "world_y": [0, 8, 24, 40, 56, 64], "source_y": [1015, 1032, 1125, 1200, 1260, 1295]},
]


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--study", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    study, out = args.study, args.output
    if not out.is_dir() or any(out.iterdir()):
        ap.error("輸出必須是既有空目錄")
    sha = lambda b: hashlib.sha256(b).hexdigest()
    base_path = study/"castle-art-v2/castle-panels-no-flags-v2.png"
    flag_path = study/"castle-animation-v1/flags-handpaint-v1.png"
    assert sha(base_path.read_bytes()) == "7ffb17e6ac4e5076446a2454c366bc17467a1584dcafe438a768bf4dc98c13d5"
    assert sha(flag_path.read_bytes()) == "aa38d78ceec13660d5f3a4b864b9963809a7307e1de8719c78cdf24e99ebfd8a"
    base = Image.open(base_path)
    flags = Image.open(flag_path)
    assert base.size == (1215, 1295) and flags.size == (1254, 1254)
    assert base.mode == flags.mode == "RGBA"
    # 全程直接編譯到3倍像素，不經原版解析度。預乘alpha避免透明邊界染色。
    premultiplied = base.convert("RGBa")
    world = Image.new("RGBa", (2736, 192))
    panel_images = []
    for panel in PANELS:
        xs, ys = panel["world_x"], panel["world_y"]
        us, vs = panel["source_x"], panel["source_y"]
        assert len(xs) == len(us) and len(ys) == len(vs)
        assert all(a < b for values in (xs, ys, us, vs) for a, b in zip(values, values[1:]))
        im = Image.new("RGBa", ((xs[-1]-xs[0])*3, 192))
        for row in range(len(ys)-1):
            for col in range(len(xs)-1):
                crop = premultiplied.crop((us[col], vs[row], us[col+1], vs[row+1]))
                crop = crop.resize(((xs[col+1]-xs[col])*3, (ys[row+1]-ys[row])*3), Image.Resampling.LANCZOS)
                im.paste(crop, ((xs[col]-xs[0])*3, ys[row]*3))
        panel_images.append(im)
    # 重疊區的中點固定交接；原版世界座標不變。
    handoffs = [0, 232, 456, 680, 912]
    for i, im in enumerate(panel_images):
        x0, x1 = handoffs[i], handoffs[i+1]
        part = im.crop(((x0-PANELS[i]["world_x"][0])*3, 0, (x1-PANELS[i]["world_x"][0])*3, 192))
        world.paste(part, (x0*3, 0))
    world.convert("RGBA").save(out/"castle-world-hd-v1.png")
    flag_rows = []
    for i, tile in enumerate(range(2, 6)):
        quadrant = (i%2*627, i//2*627, (i%2+1)*627, (i//2+1)*627)
        cell = flags.crop(quadrant)
        bbox = cell.getchannel("A").point(lambda v: 255 if v > 128 else 0).getbbox()
        assert bbox
        crop = (max(0, bbox[0]-2), max(0, bbox[1]-2), min(627, bbox[2]+2), min(627, bbox[3]+2))
        native = Image.open(study/f"castle-animation-v1/flag-{tile}-native.png")
        target = native.getchannel("A").getbbox()
        assert target and target[0] == 0 and target[1] == 0 and target[3] == 8
        sprite = Image.new("RGBa", (24, 24))
        painted = cell.convert("RGBa").crop(crop).resize(((target[2]-target[0])*3, (target[3]-target[1])*3), Image.Resampling.LANCZOS)
        sprite.paste(painted, (target[0]*3, target[1]*3))
        path = out/f"castle-flag-{tile}-hd-v1.png"
        sprite.convert("RGBA").save(path)
        flag_rows.append({"tile": tile, "quadrant": list(quadrant), "crop_in_quadrant": list(crop), "target_native": list(target),
                          "file": path.name, "sha256": sha(path.read_bytes())})
    output = out/"castle-world-hd-v1.png"
    report = {"schema": 1, "status": "COMPILED_HD_CANDIDATE", "spec": "docs/spec/036-castle-hd.md",
              "source_base_sha256": sha(base_path.read_bytes()), "source_flags_sha256": sha(flag_path.read_bytes()),
              "world_file": output.name, "world_sha256": sha(output.read_bytes()), "world_size": [2736, 192],
              "panels": PANELS, "handoffs_native": handoffs, "flags": flag_rows,
              "script_sha256": sha(Path(__file__).read_bytes()),
              "resampling": "Pillow預乘alpha LANCZOS，直接到原生3倍目標；runtime載入此固定HD像素，不重新縮放至native",
              "limits": "只編譯美術候選，需目視映射及原版遮罩／快取接入；不代表正常玩家路徑已HD"}
    (out/"manifest.json").write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
    print("COMPILED_HD_CANDIDATE", report["world_sha256"])


if __name__ == "__main__":
    main()

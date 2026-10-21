#!/usr/bin/env python3
"""依原生包絡註冊五張組合主角圖；不重畫、不改生成RGBA。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image, ImageDraw
from hd_mpp1_art import area_image

ap = argparse.ArgumentParser(description=__doc__)
ap.add_argument("--project", type=Path, required=True)
ap.add_argument("--output", type=Path, required=True)
args = ap.parse_args()
p, out = args.project, args.output
assert out.is_dir() and not any(out.iterdir())
b = p / "workplace/full-hd-v1"
art = b / "cave-combined-profile-art-v1"
sha = lambda f: hashlib.sha256(f.read_bytes()).hexdigest()
sources = json.loads((art / "sources.json").read_text())["sources"]
generation = json.loads((art / "generation.json").read_text())
assert generation["mode"] == "built-in imagegen" and len(sources) == 5
for row in generation["generations"]:
    assert sha(p / row["output_file"]) == row["output_sha256"]
    for ref in row["references"]:
        assert sha(p / ref["file"]) == ref["sha256"]
poses = []
overview = Image.new("RGB", (5 * 192, 230), (24, 30, 40))
draw = ImageDraw.Draw(overview)
for n, source in enumerate(sources):
    assert sha(p / source["native_png"]) == source["native_png_sha256"]
    raw = b / "cave-combined-profile-study-v1" / ("actor-" + source["key"][:12] + ".raster")
    assert sha(raw) == source["key"] == source["raster_sha256"]
    im = Image.open(art / generation["selected"][n]).convert("RGBA")
    assert im.getchannel("A").getextrema()[0] == 0
    box = im.getchannel("A").point(lambda a: 255 if a >= 96 else 0).getbbox()
    assert box and box[0] > 0 and box[1] > 0 and box[2] < im.width and box[3] < im.height
    crop = (box[0] - 2, box[1] - 2, box[2] + 2, box[3] + 2)
    x0, y0, x1, y1 = source["native_bounds"]
    pixels = np.zeros((72, 72, 4), dtype=np.uint8)
    pixels[y0 * 3:y1 * 3, x0 * 3:x1 * 3] = np.asarray(area_image(im.crop(crop), (x1 - x0) * 3, (y1 - y0) * 3))
    registered = Image.fromarray(pixels, "RGBA")
    target = out / ("player-" + source["key"][:12] + "-72x72.png")
    registered.save(target)
    black = Image.new("RGB", (72, 72), "black")
    black.paste(registered, mask=registered.getchannel("A"))
    overview.paste(black.resize((180, 180), Image.Resampling.NEAREST), (n * 192 + 6, 30))
    draw.text((n * 192 + 8, 8), source["key"][:12], fill="white")
    poses.append({**source, "file": target.name, "sha256": sha(target),
                  "rgba_sha256": hashlib.sha256(pixels.tobytes()).hexdigest(), "size": [72, 72],
                  "generated_file": generation["selected"][n], "crop": list(crop)})
overview.save(out / "player-profile-overview.png")
report = {"schema": 1, "status": "REGISTERED_FIVE_COMBINED_PROFILE_CANDIDATES", "poses": poses,
          "script_sha256": sha(Path(__file__)), "generation_sha256": sha(art / "generation.json"),
          "source_sha256": sha(art / "sources.json"),
          "crop_policy": {"ink_alpha_threshold": 96, "antialias_padding": 2},
          "limits": "直接依原生包絡取樣至三倍，不經原解析度縮圖；不當正式接入或正常GUI驗收。"}
(out / "manifest.json").write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n")
print(report["status"], [(r["key"][:12], r["sha256"], r["native_bounds"]) for r in poses])

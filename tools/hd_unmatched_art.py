#!/usr/bin/env python3
"""核九個原生遮罩來源，直接依包絡註冊八張新圖；原RGBA不重畫。"""
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
source = b / "cave-unmatched-assets-source-v1"
sha = lambda f: hashlib.sha256(f.read_bytes()).hexdigest()
receipt = json.loads((source / "receipt.json").read_text())
classification = json.loads((source / "classification.json").read_text())
assert classification["source_receipt_sha256"] == sha(source / "receipt.json")
known = {r["key"]: r for r in receipt["known"]}
sources = {r["key"]: r for r in receipt["sources"]}
aliases, verified = [], 0
for row in classification["sources"]:
    if row["kind"] != "object" or row["best_nonzero_pixel_differences"] != 0:
        continue
    current, prior = sources[row["key"]], known[row["closest_key"]]
    raw = (source / (current["file"] + ".raster")).read_bytes()
    old = (source / (prior["file"] + ".raster")).read_bytes()
    assert hashlib.sha256(raw).hexdigest() == current["key"]
    assert hashlib.sha256(old).hexdigest() == prior["key"]
    memory = (p / (current["source_prefix"] + ".memory")).read_bytes()
    assert hashlib.sha256(memory).hexdigest() == current["source"]["memory_sha256"]
    obj = current["object"]
    for n, value in enumerate(raw):
        x, y = n % 16, n // 16
        quad = y // 8 * 2 + x // 8
        tile, bg = obj["Tiles"][quad], obj["Backgrounds"][quad]
        assert bg in (0, 2)
        mask = memory[0x11b10 + 0xa000 + tile * 8 + y % 8]
        transparent = bg != 0 and mask & (128 >> (x % 8)) != 0
        assert old[n] != 255 and (old[n] != 0 or value in (0, 255))
        assert value == (255 if transparent else old[n])
        verified += 1
    aliases.append({"key": current["key"], "target_key": prior["key"],
                    "backgrounds": obj["Backgrounds"], "tiles": obj["Tiles"],
                    "verified_native_pixels": 256, "source_prefix": current["source_prefix"],
                    "source": current["source"]})
assert len(aliases) == 9 and verified == 2304
generation_path = b / "cave-unmatched-art-v1/generation.json"
generation = json.loads(generation_path.read_text())
assert generation["mode"] == "built-in imagegen" and len(generation["generations"]) == 8
poses = []
overview = Image.new("RGB", (4 * 240, 2 * 240), (24, 30, 40))
draw = ImageDraw.Draw(overview)
for n, row in enumerate(generation["generations"]):
    assert sha(p / row["output_file"]) == row["output_sha256"]
    for ref in row["references"]:
        assert sha(p / ref["file"]) == ref["sha256"]
    r = sources[row["key"]]
    raw = (source / (r["file"] + ".raster")).read_bytes()
    assert hashlib.sha256(raw).hexdigest() == r["key"]
    im = Image.open(p / row["output_file"]).convert("RGBA")
    box = im.getchannel("A").point(lambda a: 255 if a >= 96 else 0).getbbox()
    assert box and box[0] > 1 and box[1] > 1 and box[2] < im.width - 1 and box[3] < im.height - 1
    crop = (box[0] - 2, box[1] - 2, box[2] + 2, box[3] + 2)
    rect = r["foreground_bounds"]
    x0, y0, x1, y1 = rect["Min"]["X"], rect["Min"]["Y"], rect["Max"]["X"], rect["Max"]["Y"]
    size = r["size"] * 3
    pixels = np.zeros((size, size, 4), dtype=np.uint8)
    pixels[y0*3:y1*3, x0*3:x1*3] = np.asarray(area_image(im.crop(crop), (x1-x0)*3, (y1-y0)*3))
    registered = Image.fromarray(pixels, "RGBA")
    file = out / (r["kind"] + "-" + r["key"][:12] + "-" + str(size) + "x" + str(size) + ".png")
    registered.save(file)
    target = Image.new("RGB", (size, size), "black")
    target.paste(registered, mask=registered.getchannel("A"))
    overview.paste(target.resize((216,216), Image.Resampling.NEAREST), (n%4*240+12,n//4*240+24))
    draw.text((n%4*240+8,n//4*240+6), r["kind"]+" "+r["key"][:12], fill="white")
    poses.append({"kind": r["kind"], "key": r["key"], "file": file.name, "sha256": sha(file),
                  "rgba_sha256": hashlib.sha256(pixels.tobytes()).hexdigest(), "size": [size,size],
                  "native_bounds": [x0,y0,x1,y1], "crop": list(crop),
                  "source_prefix": r["source_prefix"], "source": r["source"]})
overview.save(out / "registered-overview.png")
report = {"schema":1, "status":"READY_EIGHT_ART_AND_NINE_MASKED_IDENTITIES", "poses":poses, "aliases":aliases,
          "verified_masked_native_pixels":verified, "script_sha256":sha(Path(__file__)),
          "source_receipt_sha256":sha(source/"receipt.json"), "generation_sha256":sha(generation_path),
          "limits":"九個來源逐像素核原版遮罩，黑色0仍保留寫入；八圖直接包絡乘三取樣，不當正式合成或GUI通過。"}
(out / "manifest.json").write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n")
print(report["status"], len(poses), "poses", verified, "masked native pixels")

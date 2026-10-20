#!/usr/bin/env python3
"""Docker內校準居民圖集裁切與原版像素邊界，只產生候選註冊，不接正式繪製。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageDraw


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--sources", type=Path, required=True)
    ap.add_argument("--art", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    if not args.sources.is_dir() or not args.art.is_dir() or not args.output.is_dir() or any(args.output.iterdir()):
        ap.error("來源及圖面須存在，輸出須為既有空目錄")
    sha = lambda f: hashlib.sha256(f.read_bytes()).hexdigest()
    source = json.loads((args.sources/"sources.json").read_text())
    assert len(source["frames"]) == 40
    artworks, frames = [], []
    review = Image.new("RGB", (1280, 1200), "#202832")
    labels = ImageDraw.Draw(review)
    for group in range(5):
        file = args.art/f"g{group}-handpaint-v1.png"
        metadata = args.art/f"g{group}-generation-v1.json"
        generation = json.loads(metadata.read_text())
        assert generation["sha256"] == sha(file)
        im = Image.open(file)
        assert im.mode == "RGBA" and im.size == (1448, 1086)
        artworks.append({"group": group, "file": file.name, "sha256": sha(file), "size": list(im.size),
                         "generation_metadata_sha256": sha(metadata)})
        for phase in range(8):
            entry = next(x for x in source["frames"] if x["group"] == group and x["phase"] == phase)
            original = args.sources/entry["file"]
            assert sha(original) == entry["sha256"]
            native = Image.open(original)
            target = tuple(entry["bbox"])
            assert native.getbbox() == target
            canonical = entry["canonical"]
            assert canonical["group"] == group
            selected = canonical["phase"]
            # 原版圖面完全相同時，重用同一候選，避免生成差異製造額外動畫。
            cell = (selected%4*im.width//4, selected//4*im.height//2,
                    (selected%4+1)*im.width//4, (selected//4+1)*im.height//2)
            tile = im.crop(cell)
            bounds = tile.getchannel("A").point(lambda v: 255 if v >= 96 else 0).getbbox()
            assert bounds is not None
            crop = (max(0, bounds[0]-2)+cell[0], max(0, bounds[1]-2)+cell[1],
                    min(tile.width, bounds[2]+2)+cell[0], min(tile.height, bounds[3]+2)+cell[1])
            assert crop[0] >= cell[0] and crop[1] >= cell[1] and crop[2] <= cell[2] and crop[3] <= cell[3]
            sprite = Image.new("RGBA", (48, 72))
            area = im.crop(crop).resize(((target[2]-target[0])*3, (target[3]-target[1])*3), Image.Resampling.LANCZOS)
            sprite.alpha_composite(area, (target[0]*3, target[1]*3))
            png = args.output/f"g{group}-p{phase}-registered.png"
            sprite.save(png)
            frame = {"group": group, "phase": phase, "source_rgba_sha256": entry["sha256"],
                     "target_native": list(target), "art_group": group, "art_phase": selected,
                     "crop": list(crop), "registered_preview": png.name, "registered_sha256": sha(png)}
            frames.append(frame)
            enlarged = sprite.resize((144, 216), Image.Resampling.NEAREST)
            review.paste(enlarged, (phase*160+8, group*240+20), enlarged)
            labels.text((phase*160+8, group*240+4), f"g{group} p{phase}", fill="white")
    manifest = {"schema": 1, "status": "DRAFT_REGISTRATION_NOT_RUNTIME", "source": source,
                "artworks": artworks, "frames": frames,
                "geometry": "原版16x24框、3倍合成；目標外框取自各姿態原始alpha，保留原版位置／碰撞",
                "resampling": "Pillow LANCZOS only for registration preview; runtime sampling contract still pending",
                "limits": ["未接正式前端，未驗連續動畫、遮擋、下層還原或正常視窗", "不把素材候選與註冊成功當成HD完成"]}
    (args.output/"manifest.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
    review.save(args.output/"registered-contact.jpg", quality=92)
    print(json.dumps({"groups": len(artworks), "frames": len(frames), "status": manifest["status"]}))


if __name__ == "__main__":
    main()

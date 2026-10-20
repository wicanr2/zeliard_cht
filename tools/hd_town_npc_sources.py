#!/usr/bin/env python3
"""從已驗原版城鎮模板擷取5類居民、40姿態，供HD製圖；不聲稱全城鎮可達。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageDraw


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    if not args.output.is_dir() or any(args.output.iterdir()):
        ap.error("輸出必須是既有空目錄")
    root, out = args.project, args.output
    path = root/"workplace/music-release-v1/release-build/data/assets/town/restored.memory"
    memory = path.read_bytes()
    sha = lambda b: hashlib.sha256(b).hexdigest()
    assert sha(memory) == "9dd3657107f25c097debf5f6be68137d055d98bda2e22a59f6b854d88be5fdd9"
    r, e = memory[0x11b10:0x21b10], memory[0x21b10:0x31b10]
    palette = (root/"workplace/music-release-v1/release-build/data/assets/town/frame/palette.bin").read_bytes()
    assert sha(palette) == "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c"
    assert sha(r[0x4000:0x5d80]) == "396a3b47dd240301fae669a494230194944987e2c46e32eaf35baa8e50d36825"
    assert sha(e[0x7000:0x74c0]) == "90de4921fb15fa26b9214b33e5c89daaa65106984db0f7417b3bd09f384392ce"
    frames, unique = [], {}
    overview = Image.new("RGB", (1280, 1200), "#202832")
    labels = ImageDraw.Draw(overview)
    for group in range(5):
        reference = Image.new("RGBA", (1024, 768))
        for phase in range(8):
            offset = 0x4000+group*48+phase*6
            ids = r[offset:offset+6]
            im = Image.new("RGBA", (16, 24))
            for q, one in enumerate(ids):
                tile = one-1
                assert 0 <= tile < 152
                data = r[0x4100+tile*48:0x4100+(tile+1)*48]
                colors = []
                for a, b, c in zip(data[::3], data[1::3], data[2::3]):
                    colors.extend((b>>2, (b&3)<<4 | a>>4, (a&15)<<2 | c>>6, c&63))
                for y in range(8):
                    for x in range(8):
                        if e[0x7000+tile*8+y] & (128>>x) == 0:
                            color = colors[y*8+x]
                            im.putpixel((q//3*8+x, q%3*8+y), (*palette[color*3:color*3+3], 255))
            name = f"g{group}-p{phase}.png"
            im.save(out/name)
            digest = sha((out/name).read_bytes())
            canonical = unique.setdefault(digest, {"group": group, "phase": phase})
            frames.append({"group": group, "phase": phase, "template": hex(offset), "tiles": list(ids),
                           "file": name, "sha256": digest, "canonical": canonical, "bbox": list(im.getbbox())})
            reference.alpha_composite(im.resize((192, 288), Image.Resampling.NEAREST), (phase%4*256+32, phase//4*384+48))
            tile = im.resize((144, 216), Image.Resampling.NEAREST)
            overview.paste(tile, (phase*160+8, group*240+20), tile)
            labels.text((phase*160+8, group*240+4), f"g{group} p{phase}", fill="white")
        reference.save(out/f"g{group}-reference.png")
    overview.save(out/"source-atlas.png")
    report = {"schema": 1, "scope": "固定Muralla已驗資源模板；不宣稱全部40姿態已在正常路徑逐一遇到",
              "source": str(path.relative_to(root)), "source_memory_sha256": sha(memory),
              "source_contract": "docs/spec/007-town-scenery-hd.md及internal/prototype/hd/town_scenery.go",
              "address_space": "dosgolem資源11B1:4000模板／4100圖格；21B1:7000遮罩",
              "script_sha256": sha(Path(__file__).read_bytes()), "unique_rgba_frames": len(unique), "frames": frames}
    (out/"sources.json").write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
    print(json.dumps({"groups": 5, "frames": len(frames), "unique": len(unique)}))


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Docker內註冊冰窟候選並以真實地圖合成離線預覽，原始AI圖面不改寫。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image, ImageDraw


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", type=Path, required=True)
    ap.add_argument("--material", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    p, out = args.project, args.output
    assert out.is_dir() and not any(out.iterdir()), "輸出須為既有空目錄"
    sha = lambda b: hashlib.sha256(b).hexdigest()
    base = p/"workplace/full-hd-v1"
    source = json.loads((base/"cave-terrain-sources-v1/receipt.json").read_text())
    profile = next(x for x in source["profiles"] if x["key"] == "MPP4")
    art_path = base/"cave-ice-art-v3/ice-terrain-handpaint-v3.png"
    assert sha(art_path.read_bytes()) == "9a09979a2c70a893c73a1dc876fe248862202573f4ce0b0c898a596bbd71b118"
    art = Image.open(art_path).convert("RGBA")
    assert art.size == (1254, 1254)
    material = Image.open(args.material).convert("RGBA")
    assert material.getchannel("A").getextrema() == (255, 255), "內部材質必須不透明"
    texture = material.resize((432, 768), Image.Resampling.LANCZOS).convert("RGB")
    tex = np.asarray(texture)
    originals, sprites, masks, reports = {}, {}, {}, []
    contact = Image.new("RGB", (768, 512), "#202b35")
    label = ImageDraw.Draw(contact)
    for i in range(1, 22):
        path = base/"cave-terrain-sources-v1/MPP4"/f"tile-{i:02d}.png"
        native = Image.open(path).convert("RGBA")
        entry = profile["tile_records"][i-1]
        assert sha(native.tobytes()) == entry["rgba_sha256"]
        originals[i] = native.convert("RGB")
        mask = Image.fromarray((np.any(np.asarray(native)[:, :, :3] != 0, axis=2)*255).astype("uint8"))
        masks[i] = mask.resize((24, 24), Image.Resampling.NEAREST)
        if i > 14:
            continue
        target = mask.getbbox()
        sprite = Image.new("RGBA", (24, 24))
        crop = None
        if target:
            if i in (13, 14):
                # 原生13／14為16x8三角形的左右半部，共用生成圖裁切再分半。
                crop = (0, 940, 625, 1254)
                pair = art.crop(crop).convert("RGBa").resize((48,24),Image.Resampling.LANCZOS).convert("RGBA")
                sprite = pair.crop((0 if i == 13 else 24, 0, 24 if i == 13 else 48, 24))
            else:
                cell = ((i-1)%4*1254//4, (i-1)//4*1254//4,
                        ((i-1)%4+1)*1254//4, ((i-1)//4+1)*1254//4)
                alpha = art.crop(cell).getchannel("A")
                bounds = alpha.point(lambda v: 255 if v >= 128 else 0).getbbox()
                assert bounds is not None
                crop = tuple(bounds[k]+cell[k%2] for k in range(4))
            if i not in (13,14):
                piece = art.crop(crop).convert("RGBa").resize(((target[2]-target[0])*3, (target[3]-target[1])*3), Image.Resampling.LANCZOS).convert("RGBA")
                sprite.alpha_composite(piece, (target[0]*3, target[1]*3))
            alpha = np.minimum(np.asarray(sprite.getchannel("A")), np.asarray(masks[i]))
            sprite.putalpha(Image.fromarray(alpha))
        sprites[i] = sprite
        filename = f"tile-{i:02d}-registered.png"
        sprite.save(out/filename)
        label.text(((i-1)%4*192+4, (i-1)//4*128+4), f"ID {i} original / HD candidate", fill="white")
        contact.paste(native.resize((64,64), Image.Resampling.NEAREST), ((i-1)%4*192+8, (i-1)//4*128+28))
        large = sprite.resize((72,72), Image.Resampling.NEAREST)
        contact.paste(large, ((i-1)%4*192+96, (i-1)//4*128+24), large)
        reports.append({"id": i, "native_rgba_sha256": entry["rgba_sha256"], "target_native": target,
                        "art_crop": crop, "registered": filename, "sha256": sha((out/filename).read_bytes())})
    assert masks[13].getbbox() is not None and masks[14].getbbox() is not None
    contact.save(out/"registered-contact.png")
    maps = json.loads((base/"cave-map-sources-v1/receipt.json").read_text())
    previews = []
    for row in maps["records"]:
        if row["family_candidate"] != "MPP4":
            continue
        data = (base/"cave-map-sources-v1"/row["columns"]).read_bytes()
        assert sha(data) == row["columns_sha256"]
        columns = np.frombuffer(data, dtype="uint8").reshape((row["width"], 64))
        markers = np.argwhere(columns == 15)
        cx, cy = (int(markers[0,0])-10, int(markers[0,1])-8) if len(markers) else (20, 8)
        grid = np.array([[int(columns[(cx+x)%row["width"],(cy+y)%64]) for x in range(28)] for y in range(18)])
        envelope = np.zeros((144,224),dtype="uint8")
        for x in range(28):
            y = 0
            while y < 18:
                if not 3 <= grid[y,x] <= 12:
                    y += 1
                    continue
                first = y
                while y < 18 and 3 <= grid[y,x] <= 12:
                    y += 1
                for dx in range(8):
                    occupied = [yy for yy in range(first*8,y*8) if np.asarray(masks[int(grid[yy//8,x])])[yy%8*3,dx*3] != 0]
                    if occupied:
                        envelope[min(occupied):max(occupied)+1,x*8+dx] = 255
        original = Image.new("RGB", (672, 432))
        result = Image.new("RGBA", (672, 432), (0,0,0,255))
        native_cells = []
        for x in range(28):
            for y in range(18):
                wx, wy = (cx+x)%row["width"], (cy+y)%64
                tile = int(columns[wx,wy])
                if tile == 0:
                    continue
                original.paste(originals[tile].resize((24,24), Image.Resampling.NEAREST), (x*24,y*24))
                if 3 <= tile <= 12:
                    xs = (wx*24+np.arange(24))%864
                    ys = (wy*24+np.arange(24))%1536
                    xs, ys = np.minimum(xs,863-xs), np.minimum(ys,1535-ys)
                    block = Image.fromarray(tex[ys[:,None],xs[None,:]])
                    contour = Image.fromarray(envelope[y*8:y*8+8,x*8:x*8+8]).resize((24,24),Image.Resampling.NEAREST)
                    result.paste(block, (x*24,y*24), contour)
                elif tile in sprites:
                    result.alpha_composite(sprites[tile], (x*24,y*24))
                else:
                    result.paste(originals[tile].resize((24,24),Image.Resampling.NEAREST), (x*24,y*24))
                    native_cells.append({"x":x,"y":y,"id":tile})
        name = row["name"].removesuffix(".MDT")
        original.save(out/(name+"-offline-original.png"))
        result.convert("RGB").save(out/(name+"-offline-hd.png"))
        previews.append({"map":name,"camera":[cx,cy],"native_cells":native_cells,
                         "hd_sha256":sha((out/(name+"-offline-hd.png")).read_bytes())})
    report = {"schema":1,"status":"DRAFT_ICE_REGISTRATION_OFFLINE_ONLY","tiles":reports,"previews":previews,
              "art_sha256":sha(art_path.read_bytes()),"material_sha256":sha(args.material.read_bytes()),
              "script_sha256":sha(Path(__file__).read_bytes()),
              "limits":["三份真實地圖的離線研究；不是正常玩家畫面或GUI。",
                        "3..12沿每條連續冰壁欄的首末非零像素形成美術包絡，再用直接HD材質填入；非原版透明或碰撞遮罩。",
                        "1／2／13／14保留原生邊界；13與14共用16x8原生三角形的對應裁切。",
                        "15..19圖形文字保留原貌，20..21未接入；不計全冰窟HD完成。",
                        "原始生成圖面未修改；註冊與來源遮罩只用於本次離線合成。"]}
    (out/"manifest.json").write_text(json.dumps(report,ensure_ascii=False,indent=2)+"\n",encoding="utf-8")
    print("DRAFT_ICE_REGISTRATION_OFFLINE_ONLY",len(reports),len(previews))


if __name__ == "__main__":
    main()

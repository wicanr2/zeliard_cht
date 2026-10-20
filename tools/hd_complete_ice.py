#!/usr/bin/env python3
"""在Docker內補齊MPP4的21格圖集；只合成離線研究，不執行或修改遊戲。"""
import argparse
import hashlib
import json
import shutil
from pathlib import Path

import numpy as np
from PIL import Image, ImageDraw


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    a = ap.parse_args()
    b = a.project / "workplace/full-hd-v1"
    out = a.output
    assert out.is_dir() and not any(out.iterdir())
    old = b / "cave-ice-registration-v3"
    old_meta = json.loads((old / "manifest.json").read_text())
    details = b / "cave-ice-details-v1/ice-details-handpaint-v1.png"
    assert digest(details) == "ac8fab5311db8071ddf4a90013cfe078e15aa8966b48b365721f70bbee21986c"
    im = Image.open(details).convert("RGBA")
    assert im.size == (1691, 930) and im.getchannel("A").getextrema() == (0, 255)
    native_meta = json.loads((b / "cave-terrain-sources-v1/receipt.json").read_text())
    profile = next(p for p in native_meta["profiles"] if p["key"] == "MPP4")
    natives = {}
    tiles = {}
    reports = []
    for rec in profile["tile_records"]:
        i = rec["id"]
        native = Image.open(b / "cave-terrain-sources-v1/MPP4" / f"tile-{i:02d}.png").convert("RGBA")
        assert hashlib.sha256(native.tobytes()).hexdigest() == rec["rgba_sha256"]
        natives[i] = native
    for row in old_meta["tiles"]:
        i = row["id"]
        source = old / row["registered"]
        assert digest(source) == row["sha256"]
        tiles[i] = Image.open(source).convert("RGBA")
        reports.append({"id": i, "source": str(source.relative_to(a.project)), "source_sha256": digest(source)})

    # 全字共用一次高解析度取樣後切成五格；不把一個圖格誤當一個字母。
    strip_crop = (58, 61, 1655, 503)
    strip = im.crop(strip_crop).convert("RGBa").resize((120, 24), Image.Resampling.LANCZOS).convert("RGBA")
    for i in range(15, 20):
        tile = Image.new("RGBA", (24, 24), (0, 0, 0, 255))
        tile.alpha_composite(strip.crop(((i-15)*24, 0, (i-14)*24, 24)))
        tiles[i] = tile
        reports.append({"id": i, "art_crop": strip_crop, "strip_part": i-15,
                        "policy": "120x24完整HELADA分五格；黑底清除原字形，後續角色圖層仍在上方"})
    for i, cell in ((20, (0, 516, 845, 930)), (21, (845, 516, 1691, 930))):
        alpha = im.crop(cell).getchannel("A").point(lambda v: 255 if v >= 128 else 0)
        box = alpha.getbbox()
        assert box is not None
        crop = (cell[0]+box[0], cell[1]+box[1], cell[0]+box[2], cell[1]+box[3])
        native_mask = np.any(np.asarray(natives[i])[:, :, :3] != 0, axis=2)
        bounds = Image.fromarray(native_mask.astype("uint8")*255).getbbox()
        tile = Image.new("RGBA", (24, 24))
        piece = im.crop(crop).convert("RGBa").resize(((bounds[2]-bounds[0])*3, (bounds[3]-bounds[1])*3), Image.Resampling.LANCZOS).convert("RGBA")
        tile.alpha_composite(piece, (bounds[0]*3, bounds[1]*3))
        mask = np.repeat(np.repeat(native_mask, 3, axis=0), 3, axis=1)*255
        tile.putalpha(Image.fromarray(np.minimum(np.asarray(tile.getchannel("A")), mask).astype("uint8")))
        tiles[i] = tile
        reports.append({"id": i, "art_crop": crop, "policy": "高解析度直接取樣，保留原生非零外形；用途未證實"})
    assert sorted(tiles) == list(range(1, 22))
    atlas = Image.new("RGBA", (504, 24))
    contact = Image.new("RGB", (896, 384), "#1c2833")
    draw = ImageDraw.Draw(contact)
    for i in range(1, 22):
        atlas.paste(tiles[i], ((i-1)*24, 0))
        filename = f"tile-{i:02d}-registered.png"
        tiles[i].save(out / filename)
        reports[i-1].update(native_rgba_sha256=profile["tile_records"][i-1]["rgba_sha256"],
                            registered=filename, sha256=digest(out / filename))
        x, y = (i-1)%7*128, (i-1)//7*128
        draw.text((x+4, y+4), f"{i}: native / HD", fill="white")
        contact.paste(natives[i].resize((48,48),Image.Resampling.NEAREST),(x+4,y+28))
        hd = tiles[i].resize((60,60),Image.Resampling.NEAREST)
        contact.paste(hd,(x+64,y+24),hd)
    atlas.save(out / "ice-tiles-v1.png")
    contact.save(out / "registered-contact.png")
    material = b / "cave-ice-material-v1/ice-material-handpaint-v1.png"
    assert digest(material) == "a8fef8473861ffe9a41d7f7cfe44dd6fda8ec94db154023034714c8fd88d4bc2"
    texture = Image.open(material).convert("RGB").resize((432,768),Image.Resampling.LANCZOS)
    texture.save(out / "ice-material-v1.png")
    tex = np.asarray(texture)
    previews = []
    maps = json.loads((b / "cave-map-sources-v1/receipt.json").read_text())
    for row in maps["records"]:
        if row["family_candidate"] != "MPP4":
            continue
        columns_path = b / "cave-map-sources-v1" / row["columns"]
        assert digest(columns_path) == row["columns_sha256"]
        columns = np.frombuffer(columns_path.read_bytes(),dtype="uint8").reshape((row["width"],64))
        marker = np.argwhere(columns == 15)
        cx,cy = (int(marker[0,0])-10,int(marker[0,1])-8) if len(marker) else (20,8)
        grid = np.array([[int(columns[(cx+x)%row["width"],(cy+y)%64]) for x in range(28)] for y in range(18)], dtype="uint8")
        assert int(grid.max()) <= 21
        grid.tofile(out / (row["name"].removesuffix(".MDT")+"-grid.bin"))
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
                    occupied = [yy for yy in range(first*8,y*8) if np.any(np.asarray(natives[int(grid[yy//8,x])])[yy%8,dx,:3] != 0)]
                    if occupied:
                        envelope[min(occupied):max(occupied)+1,x*8+dx] = 255
        native = Image.new("RGB",(672,432))
        hd = Image.new("RGBA",(672,432),(0,0,0,255))
        counts = {}
        for y in range(18):
            for x in range(28):
                i = int(grid[y,x])
                if not i:
                    continue
                counts[i] = counts.get(i,0)+1
                native.paste(natives[i].resize((24,24),Image.Resampling.NEAREST),(x*24,y*24))
                if 3 <= i <= 12:
                    # 研究快照的環形來源起點E000，視窗左側加4欄，沿011鏡射週期。
                    xs = (x+4)*24+np.arange(24)
                    ys = y*24+np.arange(24)
                    block = Image.fromarray(tex[np.minimum(ys,1535-ys)[:,None],np.minimum(xs,863-xs)[None,:]])
                    mask = Image.fromarray(envelope[y*8:y*8+8,x*8:x*8+8]).resize((24,24),Image.Resampling.NEAREST)
                    hd.paste(block,(x*24,y*24),mask)
                else:
                    hd.alpha_composite(tiles[i],(x*24,y*24))
        name = row["name"].removesuffix(".MDT")
        native.save(out / (name+"-offline-original.png"))
        hd.convert("RGB").save(out / (name+"-offline-hd.png"))
        previews.append({"map":name,"camera":[cx,cy],"grid_sha256":digest(out/(name+"-grid.bin")),
                         "tile_counts":counts,"native_fallback_cells":0,"hd_sha256":digest(out/(name+"-offline-hd.png")),
                         "columns_sha256":row["columns_sha256"]})
    meta = {"schema":1,"status":"COMPLETE_MPP4_ART_OFFLINE_ONLY","tiles":reports,"previews":previews,
            "atlas_sha256":digest(out/"ice-tiles-v1.png"),"material_sha256":digest(out/"ice-material-v1.png"),
            "details_sha256":digest(details),"script_sha256":digest(Path(__file__)),
            "limits":["21格圖面齊備；三份真實靜態地圖離線預覽，非正常玩家畫面。",
                      "20及21未出現在這三份靜態地圖；用途與正常事件來源待核。",
                      "環形UV起點為研究條件；正常冰窟的時序與捲動未驗。"]}
    (out/"manifest.json").write_text(json.dumps(meta,ensure_ascii=False,indent=2)+"\n",encoding="utf-8")
    print(meta["status"],meta["atlas_sha256"],meta["material_sha256"])


if __name__ == "__main__":
    main()

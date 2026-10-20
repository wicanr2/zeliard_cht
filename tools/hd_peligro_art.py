#!/usr/bin/env python3
"""Docker內準備MPP2原圖參考及註冊手繪圖集；完整圖面只留本機。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image, ImageDraw


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def source(project):
    b = project/"workplace/full-hd-v1"
    report = json.loads((b/"cave-terrain-sources-v1/receipt.json").read_text())
    profile = next(p for p in report["profiles"] if p["key"] == "MPP2")
    tiles = {}
    for row in profile["tile_records"]:
        im = Image.open(b/"cave-terrain-sources-v1/MPP2"/f"tile-{row['id']:02d}.png").convert("RGBA")
        assert hashlib.sha256(im.tobytes()).hexdigest() == row["rgba_sha256"]
        tiles[row["id"]] = im
    assert len(tiles) == 34
    return b, profile, tiles


def prepare(project, out):
    b, profile, tiles = source(project)
    details = Image.new("RGBA", (1024,768), (0,0,0,255))
    ids = [1,2,*range(26,35)]
    draw = ImageDraw.Draw(details)
    for n,i in enumerate(ids):
        x,y = n%4*256,n//4*256
        draw.text((x+8,y+6),f"Original MPP2 ID {i}",fill="white")
        details.paste(tiles[i].resize((192,192),Image.Resampling.NEAREST),(x+32,y+40))
    details.save(out/"details-original.png")
    text = Image.new("RGBA",(40,8))
    for i in range(21,26):
        text.paste(tiles[i],((i-21)*8,0))
    text.resize((1000,200),Image.Resampling.NEAREST).save(out/"peligro-original.png")
    rows = {"schema":1,"resource":"MPP2","record":75,"source_sha256":profile["source_sha256"],
            "detail_ids":ids,"text_ids":list(range(21,26)),"material_ids":list(range(3,21)),
            "references":[{"file":f,"sha256":sha(out/f)} for f in ("details-original.png","peligro-original.png")],
            "limits":["原圖等比例展示供imagegen參考，不是HD產物。","素材分組為顯示用途，不推定原版玩法或事件語意。"]}
    (out/"manifest.json").write_text(json.dumps(rows,ensure_ascii=False,indent=2)+"\n",encoding="utf-8")
    print("MPP2 references",len(ids),"detail frames")


def register(project, out):
    b, profile, natives = source(project)
    art_dir = b/"cave-peligro-art-v1"
    generation = json.loads((art_dir/"generation.json").read_text())
    outputs = {r["key"]:r for r in generation["outputs"]}
    images = {}
    for key,row in outputs.items():
        path = art_dir/row["file"]
        assert sha(path) == row["sha256"]
        images[key] = Image.open(path).convert("RGBA")
    assert set(images) == {"details","text","material"}
    material = images["material"]
    assert material.getchannel("A").getextrema() == (255,255)
    texture = material.convert("RGB").resize((432,768),Image.Resampling.LANCZOS)
    texture.save(out/"peligro-material-v1.png")
    tex = np.asarray(texture)
    masks = {i:np.any(np.asarray(im)[:,:,:3] != 0,axis=2) for i,im in natives.items()}
    tiles,rows = {},{}
    details = images["details"]
    w,h = details.size
    assert details.getchannel("A").getextrema() == (0,255)
    for n,i in enumerate([1,2,*range(26,35)]):
        cell = (n%4*w//4,n//4*h//3,(n%4+1)*w//4,(n//4+1)*h//3)
        bounds = details.crop(cell).getchannel("A").point(lambda v:255 if v>=128 else 0).getbbox()
        assert bounds is not None,(i,"生成圖缺格")
        crop = (cell[0]+bounds[0],cell[1]+bounds[1],cell[0]+bounds[2],cell[1]+bounds[3])
        native_bounds = Image.fromarray(masks[i].astype("uint8")*255).getbbox()
        assert native_bounds is not None
        target = ((native_bounds[2]-native_bounds[0])*3,(native_bounds[3]-native_bounds[1])*3)
        sprite = Image.new("RGBA",(24,24))
        piece = details.crop(crop).convert("RGBa").resize(target,Image.Resampling.LANCZOS).convert("RGBA")
        sprite.alpha_composite(piece,(native_bounds[0]*3,native_bounds[1]*3))
        mask = np.repeat(np.repeat(masks[i],3,axis=0),3,axis=1)*255
        sprite.putalpha(Image.fromarray(np.minimum(np.asarray(sprite.getchannel("A")),mask).astype("uint8")))
        tiles[i] = sprite
        rows[i] = {"id":i,"kind":"detail","art_crop":crop,"source_cell":n,"native_bounds":native_bounds}
    lettering = images["text"]
    assert lettering.getchannel("A").getextrema() == (0,255)
    bounds = lettering.getchannel("A").point(lambda v:255 if v>=128 else 0).getbbox()
    assert bounds is not None
    strip = lettering.crop(bounds).convert("RGBa").resize((120,24),Image.Resampling.LANCZOS).convert("RGBA")
    for i in range(21,26):
        sprite = Image.new("RGBA",(24,24),(0,0,0,255))
        sprite.alpha_composite(strip.crop(((i-21)*24,0,(i-20)*24,24)))
        tiles[i] = sprite
        rows[i] = {"id":i,"kind":"PELIGRO_strip","art_crop":bounds,"strip_part":i-21}
    for i in range(3,21):
        x,y = (i-3)%18*24,(i-3)//18*24
        sprite = texture.crop((x,y,x+24,y+24)).convert("RGBA")
        sprite.putalpha(Image.fromarray(np.repeat(np.repeat(masks[i],3,axis=0),3,axis=1).astype("uint8")*255))
        tiles[i] = sprite
        rows[i] = {"id":i,"kind":"continuous_material","contact_sample":[x,y]}
    assert sorted(tiles) == list(range(1,35))
    atlas = Image.new("RGBA",(816,24))
    contact = Image.new("RGB",(1024,640),"#1d2733")
    draw = ImageDraw.Draw(contact)
    for i in range(1,35):
        atlas.paste(tiles[i],((i-1)*24,0))
        filename = f"tile-{i:02d}-registered.png"
        tiles[i].save(out/filename)
        rows[i].update(native_rgba_sha256=profile["tile_records"][i-1]["rgba_sha256"],registered=filename,sha256=sha(out/filename))
        x,y = (i-1)%8*128,(i-1)//8*128
        draw.text((x+4,y+3),f"{i}: native / HD",fill="white")
        contact.paste(natives[i].resize((48,48),Image.Resampling.NEAREST),(x+4,y+28))
        hd = tiles[i].resize((60,60),Image.Resampling.NEAREST)
        contact.paste(hd,(x+64,y+22),hd)
    atlas.save(out/"peligro-tiles-v1.png")
    contact.save(out/"registered-contact.png")
    previews = []
    maps = json.loads((b/"cave-map-sources-v1/receipt.json").read_text())
    for row in maps["records"]:
        if row["family_candidate"] != "MPP2":
            continue
        raw = b/"cave-map-sources-v1"/row["columns"]
        assert sha(raw) == row["columns_sha256"]
        columns = np.frombuffer(raw.read_bytes(),dtype="uint8").reshape((row["width"],64))
        marker = np.argwhere(columns == 21)
        cx,cy = (int(marker[0,0])-10,int(marker[0,1])-8) if len(marker) else (10,8)
        grid = np.array([[int(columns[(cx+x)%row["width"],(cy+y)%64]) for x in range(28)] for y in range(18)],dtype="uint8")
        assert int(grid.max()) <= 34
        envelope = np.zeros((144,224),dtype="uint8")
        for x in range(28):
            y = 0
            while y < 18:
                if not 3 <= grid[y,x] <= 20:
                    y += 1
                    continue
                first = y
                while y < 18 and 3 <= grid[y,x] <= 20:
                    y += 1
                for dx in range(8):
                    occupied = [yy for yy in range(first*8,y*8) if masks[int(grid[yy//8,x])][yy%8,dx]]
                    if occupied:
                        envelope[min(occupied):max(occupied)+1,x*8+dx] = 255
        original = Image.new("RGB",(672,432))
        hd = Image.new("RGBA",(672,432),(0,0,0,255))
        for y in range(18):
            for x in range(28):
                i = int(grid[y,x])
                if i == 0:
                    continue
                original.paste(natives[i].resize((24,24),Image.Resampling.NEAREST),(x*24,y*24))
                if 3 <= i <= 20:
                    xs,ys = (x+4)*24+np.arange(24),y*24+np.arange(24)
                    block = Image.fromarray(tex[np.minimum(ys,1535-ys)[:,None],np.minimum(xs,863-xs)[None,:]])
                    mask = Image.fromarray(envelope[y*8:y*8+8,x*8:x*8+8]).resize((24,24),Image.Resampling.NEAREST)
                    hd.paste(block,(x*24,y*24),mask)
                else:
                    hd.alpha_composite(tiles[i],(x*24,y*24))
        name = row["name"].removesuffix(".MDT")
        grid.tofile(out/(name+"-grid.bin"))
        original.save(out/(name+"-offline-original.png"))
        hd.convert("RGB").save(out/(name+"-offline-hd.png"))
        previews.append({"map":name,"camera":[cx,cy],"grid_sha256":sha(out/(name+"-grid.bin")),
                         "hd_sha256":sha(out/(name+"-offline-hd.png")),"native_fallback_cells":0})
    report = {"schema":1,"status":"MPP2_ART_REGISTERED_OFFLINE_ONLY","tiles":[rows[i] for i in sorted(rows)],"previews":previews,
              "source_sha256":profile["source_sha256"],"atlas_sha256":sha(out/"peligro-tiles-v1.png"),
              "material_sha256":sha(out/"peligro-material-v1.png"),"generation_sha256":sha(art_dir/"generation.json"),"script_sha256":sha(Path(__file__)),
              "limits":["三份真實地圖的離線研究，不是正常玩家路徑。","3..20為顯示材質分組，首末非零包絡不作原版碰撞或透明語意。","26..34只依原圖保留形狀／配色，不推定玩法或動畫順序。"]}
    (out/"manifest.json").write_text(json.dumps(report,ensure_ascii=False,indent=2)+"\n",encoding="utf-8")
    print(report["status"],report["atlas_sha256"],report["material_sha256"])


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project",type=Path,required=True)
    ap.add_argument("--output",type=Path,required=True)
    ap.add_argument("--mode",choices=("prepare","register"),default="prepare")
    a = ap.parse_args()
    assert a.output.is_dir() and not any(a.output.iterdir())
    (prepare if a.mode == "prepare" else register)(a.project,a.output)


if __name__ == "__main__":
    main()

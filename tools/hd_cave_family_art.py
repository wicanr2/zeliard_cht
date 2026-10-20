#!/usr/bin/env python3
"""Docker內準備及註冊已核來源的洞窟圖集；角色分類僅供顯示，不推定玩法。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image, ImageDraw


PROFILES = {
    "MPP3": {"count":43,"material_ids":[3,4,6,7,8,9,10,32,33],
             "sheets":{"a":[1,2,5,11,12,13,14,15,16,17,18,19],
                       "b":list(range(20,32)),"c":list(range(34,44))},
             "maps":["MP30","MP31","MP3D"],"text_ids":list(range(34,39)),
             "text":"BOSQUE","generation":"generation-v2.json"},
    "MPP5": {"count":52,"material_ids":[*range(3,13),24,25,*range(35,39),*range(41,45)],
             "sheets":{"a":[1,2,*range(13,23)],
                       "b":[23,*range(26,35),39,40],"c":list(range(45,53))},
             "maps":["MP50","MP51","MP5D"],"text_ids":[],
             "text":"","generation":"generation.json"},
}


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def source(project,key):
    base = project/"workplace/full-hd-v1"
    profile = next(p for p in json.loads((base/"cave-terrain-sources-v1/receipt.json").read_text())["profiles"] if p["key"] == key)
    cfg = PROFILES[key]
    assert profile["tiles"] == cfg["count"]
    natives = {}
    for row in profile["tile_records"]:
        path = base/"cave-terrain-sources-v1"/key/f"tile-{row['id']:02d}.png"
        im = Image.open(path).convert("RGBA")
        assert hashlib.sha256(im.tobytes()).hexdigest() == row["rgba_sha256"]
        natives[row["id"]] = im
    ids = cfg["material_ids"]+[i for group in cfg["sheets"].values() for i in group if i not in cfg["text_ids"]]+cfg["text_ids"]
    assert sorted(ids) == list(range(1,cfg["count"]+1))
    return base,profile,natives,cfg


def prepare(project,key,out):
    base,profile,natives,cfg = source(project,key)
    for group,ids in cfg["sheets"].items():
        im = Image.new("RGBA",(1024,768),(0,0,0,255))
        d = ImageDraw.Draw(im)
        for n,i in enumerate(ids):
            x,y = n%4*256,n//4*256
            d.text((x+8,y+6),f"Original {key} ID {i}",fill="white")
            im.paste(natives[i].resize((192,192),Image.Resampling.NEAREST),(x+32,y+40))
        im.save(out/f"details-{group}-original.png")
    columns = 3 if key == "MPP3" else 5
    im = Image.new("RGB",(columns*256,((len(cfg["material_ids"])+columns-1)//columns)*256))
    d = ImageDraw.Draw(im)
    for n,i in enumerate(cfg["material_ids"]):
        x,y = n%columns*256,n//columns*256
        d.text((x+8,y+6),str(i),fill="white")
        im.paste(natives[i].resize((192,192),Image.Resampling.NEAREST),(x+32,y+40))
    im.save(out/"green-material-reference.png")
    if cfg["text_ids"]:
        strip = Image.new("RGBA",(len(cfg["text_ids"])*8,8))
        for n,i in enumerate(cfg["text_ids"]):
            strip.paste(natives[i],(n*8,0))
        strip.resize((1000,200),Image.Resampling.NEAREST).save(out/"label-34-38-original.png")
    meta = {"schema":1,"profile":key,"source_sha256":profile["source_sha256"],**cfg,
            "references":[{"file":f.name,"sha256":sha(f)} for f in out.glob("*.png")],
            "limits":["原圖等比例展示供imagegen參考，不計HD產物。","視覺分類不推定圖格用途、碰撞或事件；字帶另依實際相鄰順序核對。"]}
    (out/"manifest.json").write_text(json.dumps(meta,ensure_ascii=False,indent=2)+"\n",encoding="utf-8")
    print(key,"SOURCE_REFERENCES",len(natives))


def render_preview(grid,natives,tiles,texture,material_ids):
    masks = {i:np.any(np.asarray(im)[:,:,:3] != 0,axis=2) for i,im in natives.items()}
    envelope = np.zeros((144,224),dtype="uint8")
    for x in range(28):
        y = 0
        while y < 18:
            if int(grid[y,x]) not in material_ids:
                y += 1
                continue
            first = y
            while y < 18 and int(grid[y,x]) in material_ids:
                y += 1
            for dx in range(8):
                occupied = [yy for yy in range(first*8,y*8) if masks[int(grid[yy//8,x])][yy%8,dx]]
                if occupied:
                    envelope[min(occupied):max(occupied)+1,x*8+dx] = 255
    original = Image.new("RGB",(672,432))
    hd = Image.new("RGBA",(672,432),(0,0,0,255))
    tex = np.asarray(texture)
    for y in range(18):
        for x in range(28):
            i = int(grid[y,x])
            if not i:
                continue
            original.paste(natives[i].resize((24,24),Image.Resampling.NEAREST),(x*24,y*24))
            if i in material_ids:
                xs,ys = (x+4)*24+np.arange(24),y*24+np.arange(24)
                block = Image.fromarray(tex[np.minimum(ys,1535-ys)[:,None],np.minimum(xs,863-xs)[None,:]])
                mask = Image.fromarray(envelope[y*8:y*8+8,x*8:x*8+8]).resize((24,24),Image.Resampling.NEAREST)
                hd.paste(block,(x*24,y*24),mask)
            else:
                hd.alpha_composite(tiles[i],(x*24,y*24))
    return original,hd.convert("RGB")


def register(project,key,out):
    base,profile,natives,cfg = source(project,key)
    prefix = "cave-"+key.lower()
    art_dir = base/(prefix+"-art-v1")
    generation_path = art_dir/cfg["generation"]
    generation = json.loads(generation_path.read_text())
    images = {}
    for row in generation["outputs"]:
        path = art_dir/row["file"]
        assert sha(path) == row["sha256"]
        images[row["key"]] = Image.open(path).convert("RGBA")
    assert set(images) == {"material",*("details_"+k for k in cfg["sheets"]),*(["label"] if cfg["text_ids"] else [])}
    assert images["material"].getchannel("A").getextrema() == (255,255)
    texture = images["material"].convert("RGB").resize((432,768),Image.Resampling.LANCZOS)
    texture.save(out/(key.lower()+"-material-v1.png"))
    masks = {i:np.any(np.asarray(im)[:,:,:3] != 0,axis=2) for i,im in natives.items()}
    tiles,rows = {},{}
    for group,ids in cfg["sheets"].items():
        art = images["details_"+group]
        w,h = art.size
        assert art.getchannel("A").getextrema() == (0,255)
        for n,i in enumerate(ids):
            cell = (n%4*w//4,n//4*h//3,(n%4+1)*w//4,(n//4+1)*h//3)
            bounds = art.crop(cell).getchannel("A").point(lambda v:255 if v>=128 else 0).getbbox()
            assert bounds is not None,(group,i,"生成圖缺格")
            crop = (cell[0]+bounds[0],cell[1]+bounds[1],cell[0]+bounds[2],cell[1]+bounds[3])
            native_bounds = Image.fromarray(masks[i].astype("uint8")*255).getbbox()
            assert native_bounds is not None
            piece = art.crop(crop).convert("RGBa").resize(((native_bounds[2]-native_bounds[0])*3,(native_bounds[3]-native_bounds[1])*3),Image.Resampling.LANCZOS).convert("RGBA")
            sprite = Image.new("RGBA",(24,24))
            sprite.alpha_composite(piece,(native_bounds[0]*3,native_bounds[1]*3))
            mask = np.repeat(np.repeat(masks[i],3,axis=0),3,axis=1)*255
            sprite.putalpha(Image.fromarray(np.minimum(np.asarray(sprite.getchannel("A")),mask).astype("uint8")))
            tiles[i] = sprite
            rows[i] = {"id":i,"kind":"detail","sheet":group,"slot":n,"art_crop":crop,"native_bounds":native_bounds}
    if cfg["text_ids"]:
        label = images["label"]
        assert label.getchannel("A").getextrema() == (0,255)
        crop = label.getchannel("A").point(lambda v:255 if v>=128 else 0).getbbox()
        assert crop is not None
        strip = label.crop(crop).convert("RGBa").resize((len(cfg["text_ids"])*24,24),Image.Resampling.LANCZOS).convert("RGBA")
    for n,i in enumerate(cfg["text_ids"]):
        sprite = Image.new("RGBA",(24,24),(0,0,0,255))
        sprite.alpha_composite(strip.crop((n*24,0,(n+1)*24,24)))
        tiles[i] = sprite
        rows[i] = {"id":i,"kind":"graphic_label","text":cfg["text"],"art_crop":crop,"strip_part":n,
                   "superseded_detail":rows[i]}
    for n,i in enumerate(cfg["material_ids"]):
        x,y = n%18*24,n//18*24
        sprite = texture.crop((x,y,x+24,y+24)).convert("RGBA")
        sprite.putalpha(Image.fromarray(np.repeat(np.repeat(masks[i],3,axis=0),3,axis=1).astype("uint8")*255))
        tiles[i] = sprite
        rows[i] = {"id":i,"kind":"continuous_material","contact_sample":[x,y]}
    assert sorted(tiles) == list(range(1,cfg["count"]+1))
    atlas = Image.new("RGBA",(cfg["count"]*24,24))
    contact = Image.new("RGB",(1024,((cfg["count"]+7)//8)*128),"#1d2733")
    draw = ImageDraw.Draw(contact)
    for i in range(1,cfg["count"]+1):
        atlas.paste(tiles[i],((i-1)*24,0))
        filename = f"tile-{i:02d}-registered.png"
        tiles[i].save(out/filename)
        rows[i].update(native_rgba_sha256=profile["tile_records"][i-1]["rgba_sha256"],registered=filename,sha256=sha(out/filename))
        x,y = (i-1)%8*128,(i-1)//8*128
        draw.text((x+4,y+3),f"{i}: native / HD",fill="white")
        contact.paste(natives[i].resize((48,48),Image.Resampling.NEAREST),(x+4,y+28))
        im = tiles[i].resize((60,60),Image.Resampling.NEAREST)
        contact.paste(im,(x+64,y+22),im)
    atlas_path = out/(key.lower()+"-tiles-v1.png")
    atlas.save(atlas_path)
    contact.save(out/"registered-contact.png")
    maps = json.loads((base/"cave-map-sources-v1/receipt.json").read_text())
    previews = []
    for row in maps["records"]:
        if row["family_candidate"] != key:
            continue
        path = base/"cave-map-sources-v1"/row["columns"]
        assert sha(path) == row["columns_sha256"]
        columns = np.frombuffer(path.read_bytes(),dtype="uint8").reshape((row["width"],64))
        # 固定原始圖格ID為離線構圖錨點，不推定文字或事件含義。
        marker = np.argwhere(columns == 35)
        if not len(marker):
            marker = np.argwhere(columns == 15)
        cx,cy = (int(marker[0,0])-10,int(marker[0,1])-8) if len(marker) else (10,8)
        grid = np.array([[int(columns[(cx+x)%row["width"],(cy+y)%64]) for x in range(28)] for y in range(18)],dtype="uint8")
        assert int(grid.max()) <= cfg["count"]
        original,hd = render_preview(grid,natives,tiles,texture,set(cfg["material_ids"]))
        name = row["name"].removesuffix(".MDT")
        grid.tofile(out/(name+"-grid.bin"))
        original.save(out/(name+"-offline-original.png"))
        hd.save(out/(name+"-offline-hd.png"))
        previews.append({"map":name,"camera":[cx,cy],"grid_sha256":sha(out/(name+"-grid.bin")),
                         "hd_sha256":sha(out/(name+"-offline-hd.png")),"native_fallback_cells":0})
    assert {r["map"] for r in previews} == set(cfg["maps"])
    report = {"schema":1,"status":"REGISTERED_FAMILY_ART_OFFLINE_ONLY","family":key,"source_sha256":profile["source_sha256"],
              "material_ids":cfg["material_ids"],"text_ids":cfg["text_ids"],"text":cfg["text"],"tiles":[rows[i] for i in sorted(rows)],"previews":previews,
              "atlas_sha256":sha(atlas_path),"material_sha256":sha(out/(key.lower()+"-material-v1.png")),
              "generation_sha256":sha(generation_path),"script_sha256":sha(Path(__file__)),
              "limits":["原圖格／構圖只作結構資料，正常玩家可達性及事件未驗。","連續材質僅限明列ID，不以大範圍覆蓋其他圖樣。","無字帶的家族不新增字樣；不推定事件或官方中文譯名。"]}
    (out/"manifest.json").write_text(json.dumps(report,ensure_ascii=False,indent=2)+"\n",encoding="utf-8")
    print(report["status"],key,report["atlas_sha256"],report["material_sha256"])


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project",type=Path,required=True)
    ap.add_argument("--family",choices=tuple(PROFILES),required=True)
    ap.add_argument("--mode",choices=("prepare","register"),required=True)
    ap.add_argument("--output",type=Path,required=True)
    a = ap.parse_args()
    assert a.output.is_dir() and not any(a.output.iterdir()),"輸出須為既有空目錄"
    (prepare if a.mode == "prepare" else register)(a.project,a.family,a.output)


if __name__ == "__main__":
    main()

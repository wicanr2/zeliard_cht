#!/usr/bin/env python3
"""註冊MPP1的15..25格；舊14格與材質沿既有生成圖直接取樣。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image, ImageDraw

LABELS = [("satono", "SATONO", list(range(16,21))), ("muralla", "MURALLA", list(range(21,26)))]
LEGACY_SHA = "07004b059f8ed0ecf50c9aaf3a8b35c965050725fc652325d5ef248842a0b1bc"
ROCK_SHA = "09d331a411515a441790479646484467ca529517adc3f3cc235936921e57bf8f"


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def area_image(image, width, height):
    """與caveArea相同整數面積權重，分軸累積且只在最後四捨五入。"""
    src = np.asarray(image.convert("RGBA"), dtype="uint64")
    sh, sw = src.shape[:2]
    channels = src.copy()
    channels[:,:,:3] *= src[:,:,3,None]
    horizontal = np.zeros((sh,width,4), dtype="uint64")
    for x in range(width):
        left, right = x*sw, (x+1)*sw
        for xx in range(left//width, (right+width-1)//width):
            weight = min(right,(xx+1)*width)-max(left,xx*width)
            horizontal[:,x] += channels[:,xx]*weight
    summed = np.zeros((height,width,4), dtype="uint64")
    for y in range(height):
        top, bottom = y*sh, (y+1)*sh
        for yy in range(top//height, (bottom+height-1)//height):
            weight = min(bottom,(yy+1)*height)-max(top,yy*height)
            summed[y] += horizontal[yy]*weight
    alpha = summed[:,:,3]
    result = np.zeros((height,width,4), dtype="uint8")
    divisor = np.maximum(alpha,1)
    result[:,:,:3] = ((summed[:,:,:3]+alpha[:,:,None]//2)//divisor[:,:,None]).astype("uint8")
    result[:,:,3] = ((alpha+(sw*sh)//2)//(sw*sh)).astype("uint8")
    return Image.fromarray(result,"RGBA")


def source(project):
    b = project/"workplace/full-hd-v1"
    profile = next(r for r in json.loads((b/"cave-terrain-sources-v1/receipt.json").read_text())["profiles"] if r["key"]=="MPP1")
    assert profile["tiles"]==25 and sha(project/profile["source"])==profile["source_sha256"]
    natives = {}
    for row in profile["tile_records"]:
        im = Image.open(b/"cave-terrain-sources-v1/MPP1"/f"tile-{row['id']:02d}.png").convert("RGBA")
        assert hashlib.sha256(im.tobytes()).hexdigest()==row["rgba_sha256"]
        natives[row["id"]] = im
    return b,profile,natives


def prepare(project, output):
    _,_,natives = source(project)
    for _,_,ids in LABELS:
        strip = Image.new("RGB",(40,8))
        for n,i in enumerate(ids):
            strip.paste(natives[i].convert("RGB"),(n*8,0))
        strip.resize((1000,200),Image.Resampling.NEAREST).save(output/f"label-{ids[0]}-{ids[-1]}-original.png")


def preview(grid, natives, tiles, texture):
    masks = {i:np.any(np.asarray(im)[:,:,:3]!=0,axis=2) for i,im in natives.items()}
    occupied = np.zeros((144,224),dtype="uint8")
    for x in range(28):
        y = 0
        while y<18:
            if not 3<=int(grid[y,x])<=8:
                y+=1
                continue
            first=y
            while y<18 and 3<=int(grid[y,x])<=8:
                y+=1
            for dx in range(8):
                ys=[yy for yy in range(first*8,y*8) if masks[int(grid[yy//8,x])][yy%8,dx]]
                if ys:
                    occupied[min(ys):max(ys)+1,x*8+dx]=1
    original=Image.new("RGB",(672,432)); hd=np.zeros((432,672,3),dtype="uint8")
    tex=np.asarray(texture.convert("RGBA"))
    for y in range(18):
        for x in range(28):
            i=int(grid[y,x])
            if not i:
                continue
            native=natives[i].resize((24,24),Image.Resampling.NEAREST)
            original.paste(native.convert("RGB"),(x*24,y*24))
            old=np.asarray(native)[:,:,:3].astype("uint32")
            if 3<=i<=8:
                xs=(x+4)*24+np.arange(24); ys=y*24+np.arange(24)
                art=tex[np.minimum(ys,1535-ys)[:,None],np.minimum(xs,863-xs)[None,:]]
                mask=np.repeat(np.repeat(occupied[y*8:y*8+8,x*8:x*8+8],3,axis=0),3,axis=1).astype(bool)
            else:
                art=np.asarray(tiles[i]); mask=np.ones((24,24),dtype=bool) if i>=16 else np.repeat(np.repeat(masks[i],3,axis=0),3,axis=1)
            if i>=16:
                old=np.zeros_like(old)
            alpha=art[:,:,3,None].astype("uint32")
            composed=((art[:,:,:3].astype("uint32")*alpha+old*(255-alpha)+127)//255).astype("uint8")
            background=np.asarray(native)[:,:,:3].copy()
            background[mask]=composed[mask]
            hd[y*24:y*24+24,x*24:x*24+24]=background
    return original,Image.fromarray(hd,"RGB")


def register(project, output):
    b,profile,natives=source(project)
    legacy=project/"workplace/cave-hd-v1/terrain-handpaint-v2.png"
    rock=project/"workplace/cave-hd-v1/rock-material-v1.png"
    assert sha(legacy)==LEGACY_SHA and sha(rock)==ROCK_SHA
    atlas=area_image(Image.open(legacy),96,96)
    tiles={i:atlas.crop((i%4*24,i//4*24,i%4*24+24,i//4*24+24)) for i in range(1,15)}
    texture=area_image(Image.open(rock),432,768)
    generation_path=b/"cave-mpp1-art-v1/generation-v2.json"
    generation=json.loads(generation_path.read_text())
    assert generation["mode"]=="builtin-imagegen" and {r["key"] for r in generation["outputs"]}=={"satono","muralla","detail_15"}
    outputs={r["key"]:r for r in generation["outputs"]}
    row=outputs["detail_15"]; path=generation_path.parent/row["file"]; assert sha(path)==row["sha256"]
    im=Image.open(path).convert("RGBA"); alpha_min,alpha_max=im.getchannel("A").getextrema()
    assert alpha_min==0 and alpha_max>=128, "細部須有透明背景及可見主體，不要求生成alpha恰為255"
    bounds=im.getchannel("A").point(lambda v:255 if v>=128 else 0).getbbox(); assert bounds
    mask=np.any(np.asarray(natives[15])[:,:,:3]!=0,axis=2)
    native_bounds=Image.fromarray(mask.astype("uint8")*255).getbbox(); assert native_bounds
    piece=im.crop(bounds).convert("RGBa").resize(((native_bounds[2]-native_bounds[0])*3,(native_bounds[3]-native_bounds[1])*3),Image.Resampling.LANCZOS).convert("RGBA")
    tiles[15]=Image.new("RGBA",(24,24)); tiles[15].alpha_composite(piece,(native_bounds[0]*3,native_bounds[1]*3))
    tiles[15].putalpha(Image.fromarray(np.minimum(np.asarray(tiles[15].getchannel("A")),np.repeat(np.repeat(mask,3,axis=0),3,axis=1)*255).astype("uint8")))
    for key,text,ids in LABELS:
        row=outputs[key]; path=generation_path.parent/row["file"]; assert sha(path)==row["sha256"]
        im=Image.open(path).convert("RGBA"); assert im.getchannel("A").getextrema()==(0,255)
        crop=im.getchannel("A").point(lambda v:255 if v>=128 else 0).getbbox(); assert crop
        strip=im.crop(crop).convert("RGBa").resize((120,24),Image.Resampling.LANCZOS).convert("RGBA")
        for n,i in enumerate(ids):
            tiles[i]=strip.crop((n*24,0,(n+1)*24,24))
    registered=Image.new("RGBA",(264,24)); rows=[]
    contact=Image.new("RGB",(1024,3*128),"#1d2733"); draw=ImageDraw.Draw(contact)
    for i in range(15,26):
        im=tiles[i]; assert im.getchannel("A").getbbox() and np.any(np.asarray(im)[:,:,:3]),("缺少可見圖格",i)
        file=f"tile-{i:02d}-registered.png"; im.save(output/file)
        registered.paste(im,((i-15)*24,0))
        rows.append({"id":i,"registered":file,"sha256":sha(output/file),"native_rgba_sha256":profile["tile_records"][i-1]["rgba_sha256"],"kind":"detail" if i==15 else "graphic_label"})
        x,y=(i-15)%8*128,(i-15)//8*128
        draw.text((x+4,y+3),f"{i}: native / HD",fill="white")
        contact.paste(natives[i].resize((48,48),Image.Resampling.NEAREST),(x+4,y+28))
        shown=im.resize((60,60),Image.Resampling.NEAREST); contact.paste(shown,(x+64,y+22),shown)
    registered.save(output/"mpp1-extra-tiles-v1.png"); contact.save(output/"registered-contact.png")
    records={r["name"].removesuffix(".MDT"):r for r in json.loads((b/"cave-map-sources-v1/receipt.json").read_text())["records"]}
    views=[]
    for name,tag,marker in [("MP10","MP10-Satono",16),("MP10","MP10-Muralla",21),("MP1D","MP1D",3)]:
        row=records[name]; path=b/"cave-map-sources-v1"/row["columns"]; assert sha(path)==row["columns_sha256"]
        columns=np.frombuffer(path.read_bytes(),dtype="uint8").reshape((row["width"],64))
        positions=np.argwhere(columns==marker); assert len(positions)
        cx,cy=int(positions[0,0])-10,int(positions[0,1])-8
        grid=np.array([[int(columns[(cx+x)%row["width"],(cy+y)%64]) for x in range(28)] for y in range(18)],dtype="uint8")
        original,hd=preview(grid,natives,tiles,texture)
        grid.tofile(output/(tag+"-grid.bin")); original.save(output/(tag+"-offline-original.png")); hd.save(output/(tag+"-offline-hd.png"))
        views.append({"map":name,"tag":tag,"camera":[cx,cy],"grid_sha256":sha(output/(tag+"-grid.bin")),"hd_sha256":sha(output/(tag+"-offline-hd.png")),"native_fallback_cells":0})
    report={"schema":1,"status":"REGISTERED_MPP1_EXTRAS_OFFLINE_ONLY","source_sha256":profile["source_sha256"],"legacy_art_sha256":LEGACY_SHA,"legacy_material_sha256":ROCK_SHA,
            "atlas_sha256":sha(output/"mpp1-extra-tiles-v1.png"),"generation_sha256":sha(generation_path),"script_sha256":sha(Path(__file__)),"tiles":rows,"labels":[{"text":text,"ids":ids} for _,text,ids in LABELS],"previews":views,
            "limits":["只註冊15..25，1..14及岩面沿既有圖面與合成方式。","靜態網格不是正常Satono到達或事件驗收。"]}
    (output/"manifest.json").write_text(json.dumps(report,ensure_ascii=False,indent=2)+"\n")
    print(report["status"],report["atlas_sha256"])


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project",type=Path,required=True)
    ap.add_argument("--output",type=Path,required=True)
    ap.add_argument("--mode",choices=("prepare","register"),required=True)
    a=ap.parse_args(); assert a.output.is_dir() and not any(a.output.iterdir())
    (prepare if a.mode=="prepare" else register)(a.project,a.output)


if __name__=="__main__":
    main()

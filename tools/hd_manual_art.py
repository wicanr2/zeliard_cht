#!/usr/bin/env python3
"""Docker內整理050人工驗證候選圖組；不執行測試、對拍或遊戲。"""
import argparse
import base64
import hashlib
import json
import math
import shutil
import struct
from pathlib import Path

import numpy as np
from PIL import Image


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", type=Path, required=True)
    ap.add_argument("--cache", type=Path, required=True)
    args = ap.parse_args()
    p = args.project
    w = p / "workplace/full-hd-v1/manual-hd-v1"
    sha = lambda f: hashlib.sha256(f.read_bytes()).hexdigest()
    plan = json.loads((w / "plan.json").read_text())
    # 請求與cache本身保留；複製選用版本，不覆寫舊生成圖。
    for request in sorted((w / "generated").glob("*-request.json")):
        meta = json.loads(request.read_text())
        version = meta.get("version", 1)
        target = w / "generated" / f'{meta["key"]}-generated-v{version}.png'
        source = args.cache / Path(meta["cache"]).relative_to("/home/anr2/.codex/generated_images")
        if not target.exists():
            shutil.copy2(source, target)
        meta["output_file"] = str(target.relative_to(p))
        meta["output_sha256"] = sha(target)
        meta["references"] = [{"file": str(Path(f).relative_to("/home/anr2/cht/zeliard")),
                                "sha256": sha(p / Path(f).relative_to("/home/anr2/cht/zeliard"))} for f in meta["refs"]]
        request.write_text(json.dumps(meta, ensure_ascii=False, indent=2) + "\n")
    selected = {"enp1": 1, "enp2": 2, "enp3": 2, "enp4": 2,
                "enp5": 3, "enp6": 3, "enp7": 2, "enp8": 3}
    selected.update({key:1 for key in ('crab','tako','tori','zela','meda','lega','drgn','akma','mao1','mao2')})
    selected.update({key:2 for key in ('meda','lega','drgn','akma','mao1','mao2')})
    memory = (p / "workplace/full-hd-v1/cave-profile-normal-pair-v2/cave-platform/object-missing-00.memory").read_bytes()
    primary = memory[0x1b10:0x11b10]
    word = lambda at: struct.unpack_from("<H", primary, at)[0]
    poses, seen = [], set()
    registered = w / "registered"
    for bank in plan["assets"]:
        if bank["key"] not in selected:
            continue
        version = selected[bank["key"]]
        source = w / "generated" / f'{bank["key"]}-generated-v{version}.png'
        if not source.exists():
            raise SystemExit(f"圖面尚未產生：{source}")
        art = Image.open(source).convert("RGBA")
        columns, rows = bank["columns"], bank["rows"]
        for n, frame in enumerate(bank["frames"]):
            pal = primary[word(0x4f8c + frame["palette"] * 2):word(0x4f8c + frame["palette"] * 2) + 16]
            raw = bytearray(256)
            for q, tile in enumerate(frame["tiles"]):
                for i, v in enumerate(bank["tiles"][tile]["nibbles"]):
                    raw[(q // 2 * 8 + i // 8) * 16 + q % 2 * 8 + i % 8] = pal[v]
            key = hashlib.sha256(raw).hexdigest()
            if key in seen or not any(raw):
                continue
            seen.add(key)
            x0, x1 = n % columns * art.width / columns, (n % columns + 1) * art.width / columns
            y0, y1 = n // columns * art.height / rows, (n // columns + 1) * art.height / rows
            # 參考圖每格24原生pixel，中間16pixel是原圖；直接取三倍，不縮回原版。
            dx, dy = (x1 - x0) / 6, (y1 - y0) / 6
            crop = (round(x0 + dx), round(y0 + dy), round(x1 - dx), round(y1 - dy))
            image = art.crop(crop).resize((48, 48), Image.Resampling.LANCZOS)
            rgba = np.array(image)
            mask = np.asarray(list(raw), dtype=np.uint8).reshape(16, 16) != 0
            mask = np.repeat(np.repeat(mask, 3, axis=0), 3, axis=1)
            rgba[:, :, 3] = np.where(mask, rgba[:, :, 3], 0)
            file = registered / ("object-" + key[:16] + ".png")
            Image.fromarray(rgba, "RGBA").save(file)
            poses.append({"key": key, "file": file.name, "sha256": sha(file), "bank": bank["key"],
                          "template": frame, "source_art": str(source.relative_to(p)), "source_art_sha256": sha(source),
                          "source_crop": crop, "raster": base64.b64encode(raw).decode(), "bounds": [0, 0, 48, 48]})
    report = {"schema": 1, "status": "MANUAL_REVIEW_PENDING", "automatic_tests": False,
              "objects": poses, "selected_generations": selected,
              "limits": "僅製作及接入候選；未跑測試、對拍或遊戲。保留原版來源、位置及透明幾何，未知來源保留原貌。"}
    (registered / "manual-art.json").write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n")
    tiles=[]; tile_seen=set()
    palette=np.frombuffer((p/'workplace/music-release-v1/release-build/AppDir/assets/town/frame/palette.bin').read_bytes(),np.uint8).reshape(256,3)
    def add_tiles(raw,width,height,art,category,match_rgb=True):
        raster=np.frombuffer(raw,np.uint8).reshape(height,width)
        for y in range(0,height-7,8):
            for x in range(0,width-7,8):
                native=raster[y:y+8,x:x+8];colors=set(native.flatten().tolist())
                if len(colors)<3 or np.count_nonzero(native)<8:continue
                key=hashlib.sha256(native.tobytes()).hexdigest()
                if key in tile_seen:continue
                tile_seen.add(key)
                image=art.crop((x*3,y*3,x*3+24,y*3+24));file=registered/('tile-'+key[:16]+'.png');image.save(file)
                tiles.append({'raster':base64.b64encode(native.tobytes()).decode(),'rgb':base64.b64encode(palette[native].tobytes()).decode() if match_rgb else None,
                              'file':file.name,'sha256':sha(file),'category':category})
    for row in poses:
        add_tiles(base64.b64decode(row['raster']),16,16,Image.open(registered/row['file']).convert('RGBA'),row['bank'])
    pictures=json.loads((w/'pictures.json').read_text())['scenes']
    by_key={r['key']:r for r in pictures}
    for sheet in json.loads((w/'picture-sheets.json').read_text())['sheets']:
        source=w/'generated'/(sheet['key']+'-generated-v1.png')
        if not source.exists():continue
        atlas=Image.open(source).convert('RGBA')
        for cell in sheet['cells']:
            row=by_key[cell['key']];box=cell['box'];box=tuple(round(v*(atlas.width/1536 if i%2==0 else atlas.height/1024)) for i,v in enumerate(box))
            art=atlas.crop(box).resize((row['width']*3,row['height']*3),Image.Resampling.LANCZOS)
            add_tiles(base64.b64decode(row['raster']),row['width'],row['height'],art,row['key'],False)
    common_file=w/'generated/common-generated-v1.png'
    if common_file.exists():
        meta=json.loads((w/'common-tiles.json').read_text());atlas=Image.open(common_file).convert('RGBA')
        for n,row in enumerate(meta['tiles']):
            x0=n%16*atlas.width/16;y0=n//16*atlas.height/13;cw=atlas.width/16;ch=atlas.height/13
            image=atlas.crop((round(x0+cw/6),round(y0+ch/6),round(x0+cw*5/6),round(y0+ch*5/6))).resize((24,24),Image.Resampling.LANCZOS)
            rgba=np.array(image);native=np.frombuffer(base64.b64decode(row['raster']),np.uint8).reshape(8,8)
            rgba[:,:,3]=np.where(np.repeat(np.repeat(native!=0,3,axis=0),3,axis=1),rgba[:,:,3],0)
            add_tiles(native.tobytes(),8,8,Image.fromarray(rgba),'common',True)
    credits_file=w/'generated/credits-generated-v1.png'
    if credits_file.exists():
        meta=json.loads((w/'credits-scroll.json').read_text());art=Image.open(credits_file).convert('RGBA').resize((288,795),Image.Resampling.LANCZOS)
        add_tiles(base64.b64decode(meta['raster']),96,265,art,'credits-scroll',False)
    (registered/'manual-tiles.json').write_text(json.dumps({'schema':1,'status':'MANUAL_REVIEW_PENDING','tiles':tiles},ensure_ascii=False,indent=2)+'\n')
    # 八城鎮沿用既有完整手繪條帶，不重畫已完成圖面。
    town_source=p/'workplace/full-hd-v1/town-resources-source-v1'
    towns=json.loads((town_source/'receipt.json').read_text())['maps']
    town_jobs=json.loads((p/'workplace/full-hd-v1/town-world-art-v1/generation-v1.json').read_text())['jobs']
    world=[]
    from hd_grp_sources import terrain_preprocess
    for job in town_jobs:
        row=next(t for t in towns if t['key']==job['key']);key=row['key']
        version=2 if key=='ESMP' else 1
        source=p/f'workplace/full-hd-v1/town-world-art-v1/{key}-generated-v{version}.png'
        atlas=Image.open(source).convert('RGBA');art=Image.new('RGBA',(row['width']*24,192))
        target_width,target_height=job['target_reference_size']
        for panel in job['panels']:
            box=tuple(round(v*(atlas.width/target_width if i%2==0 else atlas.height/target_height)) for i,v in enumerate(panel['reference_box']))
            x0,_,x1,_=panel['world_native'];im=atlas.crop(box).resize(((x1-x0)*3,192),Image.Resampling.LANCZOS);art.paste(im,(x0*3,0))
        native=Image.open(town_source/(key+'-foreground-static.png')).convert('RGBA')
        # 原版透明範圍與未知191圖格不補造幾何。
        rgba=np.array(art);mask=np.repeat(np.repeat(np.array(native)[:,:,3],3,axis=0),3,axis=1)
        rgba[:,:,3]=np.minimum(rgba[:,:,3],mask)
        columns=(town_source/(key+'.columns')).read_bytes()
        fam=next(f for f in json.loads((town_source/'receipt.json').read_text())['families'] if f['key']==row['terrain_key'])
        for col in range(row['width']):
            for y in range(8):
                if columns[col*8+y]>=fam['count']:rgba[y*24:y*24+24,col*24:col*24+24,3]=0
        file=registered/(key+'-world.png');Image.fromarray(rgba).save(file)
        native_file=registered/(key+'-native.png');native.save(native_file)
        raw=(p/f'workplace/full-hd-v1/decoded-v2/zelres2-{fam["record"]:02d}.decoded').read_bytes()
        packed=b''.join(terrain_preprocess(raw[256+t*48:256+(t+1)*48],raw[6+t])[0] for t in range(fam['count']))
        world.append({'key':key,'width':row['width'],'town_id':row['town_id'],'columns':base64.b64encode(columns).decode(),
                      'terrain':base64.b64encode(packed).decode(),'file':file.name,'sha256':sha(file),'native_file':native_file.name,'native_sha256':sha(native_file)})
    (registered/'world-towns.json').write_text(json.dumps({'schema':1,'status':'MANUAL_REVIEW_PENDING','towns':world},ensure_ascii=False,indent=2)+'\n')
    cman=w/'generated/cman-generated-v1.png'
    if cman.exists():
        from hd_grp_sources import preprocess
        tail=json.loads((w/'tail.json').read_text());atlas=Image.open(cman).convert('RGBA');figures=[]
        for row in tail['figures']:
            x,y=row['phase'],row['group']
            box=(round((x*24+4)/192*atlas.width),round((y*32+4)/160*atlas.height),round((x*24+20)/192*atlas.width),round((y*32+28)/160*atlas.height))
            art=atlas.crop(box).resize((48,72),Image.Resampling.LANCZOS);rgba=np.array(art)
            mask=np.frombuffer(base64.b64decode(row['mask']),np.uint8).reshape(24,16);mask=np.repeat(np.repeat(mask,3,axis=0),3,axis=1)
            rgba[:,:,3]=np.minimum(rgba[:,:,3],mask)
            file=registered/f'cman-g{y}-p{x}.png';Image.fromarray(rgba).save(file)
            figures.append({**row,'file':file.name,'sha256':sha(file)})
        data=(p/'workplace/full-hd-v1/decoded-v2/zelres2-30.decoded').read_bytes()
        resource=data[:256]+b''.join(preprocess(data[256+t*48:256+(t+1)*48],7,0)[0] for t in range(164))
        (registered/'manual-figures.json').write_text(json.dumps({'schema':1,'status':'MANUAL_REVIEW_PENDING','resource':base64.b64encode(resource).decode(),'palette':base64.b64encode(palette.tobytes()).decode(),'figures':figures},ensure_ascii=False,indent=2)+'\n')
    print("PREPARED_MANUAL_ART", len(poses), "objects; no tests or gameplay executed")


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""050其餘圖格及居民候選。保留來源圖格，組裝與遊戲到達交人工檢視。"""
import argparse, base64, hashlib, json, math
from pathlib import Path
import numpy as np
from PIL import Image
from hd_planar import convert
from hd_grp_sources import preprocess
from hd_manual_sources import img_open, picture_six


def unpack(data):
    values=[]
    for a,b,c in zip(data[::3],data[1::3],data[2::3]):values.extend((b>>2,(b&3)*16+(a>>4),(a&15)*4+(c>>6),c&63))
    return bytes(values)


def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('--project',type=Path,required=True);p=ap.parse_args().project
    w=p/'workplace/full-hd-v1/manual-hd-v1';root=p/'workplace/full-hd-v1/decoded-v2'
    pal=np.frombuffer((p/'workplace/music-release-v1/release-build/AppDir/assets/town/frame/palette.bin').read_bytes(),np.uint8).reshape(256,3)
    entries=[]; seen=set()
    # 這些拆分資料只建立可回退候選，不把研究圖格當作已確認的完整插畫。
    specs=[(1,i,0) for i in (14,17,19,21,23,26,30,35,36)]+[(2,i,0) for i in range(51,58)]+[(3,53,0),(3,54,0)]
    for archive,record,offset in specs:
        data=(root/f'zelres{archive}-{record:02d}.decoded').read_bytes()
        inner=archive==1 or archive==2 and record in range(52,57)
        if inner:
            try:data=img_open(data)
            except (IndexError,ValueError):pass
        data=data[offset:]
        for i in range(len(data)//48):
            raw=unpack(convert(data[i*48:(i+1)*48]))
            if archive==1:raw=bytes((v>>3)*16+(v&7) for v in raw)
            if len(set(raw))<3 or not any(raw):continue
            key=hashlib.sha256(raw).hexdigest()
            if key in seen:continue
            seen.add(key);entries.append({'key':key,'raster':base64.b64encode(raw).decode(),'archive':archive,'record':record,'tile':i,'source_kind':'fragment_candidate','source_sha256':hashlib.sha256(data).hexdigest()})
    pages=[]
    for start in range(0,len(entries),320):
        batch=entries[start:start+320];rows=math.ceil(len(batch)/16);atlas=Image.new('RGBA',(192,rows*12))
        for n,row in enumerate(batch):
            raw=np.frombuffer(base64.b64decode(row['raster']),np.uint8).reshape(8,8)
            rgba=np.concatenate([pal[raw],np.where(raw!=0,255,0).astype(np.uint8)[:,:,None]],axis=2)
            atlas.paste(Image.fromarray(rgba),(n%16*12+2,n//16*12+2))
        key=f'tail{start//320+1}';file=w/'references'/(key+'.png');atlas.resize((1152,rows*72),Image.Resampling.NEAREST).save(file)
        pages.append({'key':key,'columns':16,'rows':rows,'reference':str(file.relative_to(p)),'tiles':batch})
    # CMAN五類居民、各八姿態，沿既有MMAN的6圖格模板方式整理。
    data=(root/'zelres2-30.decoded').read_bytes();figures=[];atlas=Image.new('RGBA',(192,160))
    tiles=[]
    for i in range(164):
        packed,mask=preprocess(data[256+i*48:256+(i+1)*48],7,0)
        tiles.append((unpack(packed),mask))
    for group in range(5):
        for phase in range(8):
            raw=np.zeros((24,16),np.uint8);alpha=np.zeros((24,16),np.uint8)
            ids=data[group*48+phase*6:group*48+phase*6+6]
            for q,id in enumerate(ids):
                colors,mask=tiles[id-1];x,y=q//3*8,q%3*8
                raw[y:y+8,x:x+8]=np.frombuffer(colors,np.uint8).reshape(8,8)
                alpha[y:y+8,x:x+8]=np.array([[0 if mask[yy]&(128>>xx) else 255 for xx in range(8)] for yy in range(8)],np.uint8)
            rgba=np.concatenate([pal[raw],alpha[:,:,None]],axis=2)
            atlas.paste(Image.fromarray(rgba),(phase*24+4,group*32+4))
            figures.append({'group':group,'phase':phase,'width':16,'height':24,'raster':base64.b64encode(raw.tobytes()).decode(),'mask':base64.b64encode(alpha.tobytes()).decode()})
    file=w/'references/cman-figures.png';atlas.resize((1152,960),Image.Resampling.NEAREST).save(file)
    (w/'tail.json').write_text(json.dumps({'status':'MANUAL_REVIEW_PENDING','pages':pages,'figures':figures,'figure_reference':str(file.relative_to(p))},ensure_ascii=False,indent=2)+'\n')
    # 主插畫使用已整理的完整構圖，其餘動畫部件留在上列候選圖格。
    good={'ame','hime','himp','isi','ne80','nec','new2','oui','sei','seip','ttl1','ttl3','waku','yuu1','yuu2','yuup'}
    scenes=[r for r in json.loads((w/'pictures.json').read_text())['scenes'] if r['key'] in good]
    sheets=[]
    for start in range(0,len(scenes),4):
        sheet=Image.new('RGB',(1536,1024),'black');cells=[]
        for n,row in enumerate(scenes[start:start+4]):
            im=Image.open(p/row['reference']);ratio=min(720/im.width,460/im.height);size=tuple(round(v*ratio) for v in im.size)
            im=im.resize(size,Image.Resampling.NEAREST);x=n%2*768+(768-size[0])//2;y=n//2*512+(512-size[1])//2;sheet.paste(im,(x,y))
            cells.append({'key':row['key'],'box':[x,y,x+size[0],y+size[1]]})
        key=f'story{start//4+1}';file=w/'references'/(key+'-selected.png');sheet.save(file)
        sheets.append({'key':key,'reference':str(file.relative_to(p)),'cells':cells})
    (w/'picture-sheets.json').write_text(json.dumps({'sheets':sheets},ensure_ascii=False,indent=2)+'\n')
    # 已有48-byte消費端的共用圖格另列，不混入上方未解組裝候選。
    common=[];atlas=Image.new('RGBA',(192,math.ceil(199/16)*12))
    for archive,record in [(2,57),(3,53),(3,54)]:
        data=(root/f'zelres{archive}-{record:02d}.decoded').read_bytes()
        for tile in range(len(data)//48):
            raw=unpack(convert(data[tile*48:(tile+1)*48]));n=len(common)
            native=np.frombuffer(raw,np.uint8).reshape(8,8);rgba=np.concatenate([pal[native],np.where(native!=0,255,0).astype(np.uint8)[:,:,None]],axis=2)
            atlas.paste(Image.fromarray(rgba),(n%16*12+2,n//16*12+2))
            common.append({'raster':base64.b64encode(raw).decode(),'archive':archive,'record':record,'tile':tile})
    atlas.resize((1152,936),Image.Resampling.NEAREST).save(w/'references/common-tiles.png')
    (w/'common-tiles.json').write_text(json.dumps({'columns':16,'rows':13,'tiles':common},ensure_ascii=False,indent=2)+'\n')
    data=img_open((root/'zelres1-23.decoded').read_bytes())
    # ENDMO fill_credits_triplane: plane stride18D8、每列24bytes，96×265捲動圖。
    raw=picture_six(data,96,265)
    # 位元順序與PC-98相同（1藍、2紅、4綠）。規格068。
    colors=np.array([[(((i>>4)&2)+(i&2))*255//4,(((i>>4)&4)+(i&4))*255//8,(((i>>4)&1)+(i&1))*255//2] for i in range(256)],np.uint8)
    Image.fromarray(colors[np.frombuffer(raw,np.uint8)].reshape(265,96,3)).resize((384,1060),Image.Resampling.NEAREST).save(w/'references/credits-scroll.png')
    (w/'credits-scroll.json').write_text(json.dumps({'width':96,'height':265,'raster':base64.b64encode(raw).decode(),'source':'zelres1-23.decoded','plane_stride':0x18d8},indent=2)+'\n')
    print('PREPARED',len(entries),'fragment candidates',len(pages),'pages; CMAN40; story',len(scenes))


if __name__=='__main__':main()

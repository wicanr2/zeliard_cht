#!/usr/bin/env python3
"""全圖HD的剩餘來源整理。只解碼與組圖，不執行遊戲或測試。"""
import argparse
import base64
import hashlib
import json
import math
import struct
from pathlib import Path

import numpy as np
from PIL import Image, ImageDraw

from hd_manual_sources import img_open, picture, picture_six, rle
from hd_interior_sources import tiles
from hd_inventory import records
from hd_grp_sources import preprocess, terrain_preprocess


def sha(data):
    return hashlib.sha256(data).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    p = ap.parse_args().project
    w = p/'workplace/full-hd-v1/manual-hd-v1'
    root = p/'workplace/full-hd-v1/decoded-v2'
    palette = np.frombuffer((p/'workplace/music-release-v1/release-build/AppDir/assets/town/frame/palette.bin').read_bytes(), np.uint8).reshape(256,3)
    intro = np.frombuffer((p/'workplace/probe-intro-display/palette.bin').read_bytes(), np.uint8).reshape(256,3)
    # 兩個三位元子像素的平均色，位元順序與PC-98相同（1藍、2紅、4綠）。規格068。
    rgb_pair = np.array([[(((i>>4)&2)+(i&2))*255//4,(((i>>4)&4)+(i&4))*255//8,(((i>>4)&1)+(i&1))*255//2] for i in range(256)], np.uint8)
    rows = []

    def load(archive, record):
        return (root/f'zelres{archive}-{record:02d}.decoded').read_bytes()

    def scene(key, raw, width, height, colors, archive, record, method, source=None):
        if len(raw)!=width*height: raise ValueError(key+' 圖面尺寸不符')
        native = Image.fromarray(colors[np.frombuffer(raw,np.uint8)].reshape(height,width,3))
        file = w/'references'/f'complete-{key}-v1.png'
        native.resize((width*4,height*4),Image.Resampling.NEAREST).save(file)
        rows.append({'key':key,'kind':'picture','width':width,'height':height,'raster':base64.b64encode(raw).decode(),
                     'reference':str(file.relative_to(p)),'archive':archive,'record':record,'source_sha256':sha(load(archive,record) if source is None else source),
                     'method':method,'level':'strong inference: source layout; integrated review deferred','colors':base64.b64encode(colors.tobytes()).decode()})

    def part_atlas(key, parts, archive, record, method, columns=8, cell=32):
        height=math.ceil(len(parts)/columns);atlas=Image.new('RGBA',(columns*cell,height*cell))
        for n,row in enumerate(parts):
            raw=base64.b64decode(row['raster']);width,h=row['width'],row['height'];x=n%columns*cell+(cell-width)//2;y=n//columns*cell+(cell-h)//2
            colors=rgb_pair if row.get('intro') else palette
            image=Image.fromarray(colors[np.frombuffer(raw,np.uint8)].reshape(h,width,3)).convert('RGBA')
            if 'mask' in row:image.putalpha(Image.frombytes('L',(width,h),base64.b64decode(row['mask'])))
            atlas.paste(image,(x,y));row['box']=[x,y,x+width,y+h]
        file=w/'references'/f'complete-{key}-v1.png';atlas.resize((columns*cell*4,height*cell*4),Image.Resampling.NEAREST).save(file)
        rows.append({'key':key,'kind':'parts','parts':parts,'reference':str(file.relative_to(p)),'atlas_size':list(atlas.size),'archive':archive,'record':record,'source_sha256':sha(load(archive,record)),'method':method})

    # DMAOU的兩平面由三個分離矩形組成，不能把壓縮輸出當連續圖面。
    data = img_open(load(1,14)); seg=bytearray(65536);seg[0x97c0:0x97c0+len(data)]=data;scratch=bytearray(0x2ca0)
    for start,bx,cx,passes in [(0xab40,0,0x2230,(1,0)),(0xa9c0,0x0f30,0x0620,(1,0)),(0x9c40,0x0850,0x1220,(0,1))]:
        at=start
        for font in passes:
            for y in range(cx&255):
                for x in range(cx>>8):
                    dest=(bx&255)*34+(bx>>8)+font*0xee0+y*34+x
                    scratch[dest]=seg[at];at+=1
    scene('dmaou',picture(scratch,136,112),136,112,intro,1,14,'OPDMO palette_lookup scratch 22×70, plane stride0EE0')
    def two_planes(data,width,height,lookup):
        stride=width//4*height;out=bytearray()
        for a,b in zip(data[:stride],data[stride:stride*2]):
            logical=[lookup[((a>>bit)&1)|(((b>>bit)&1)<<1)] for bit in range(7,-1,-1)]
            out.extend(logical[i]*16+logical[i+1] for i in range(0,8,2))
        return bytes(out)
    data=img_open(load(1,17));parts=[];atlas=Image.new('RGB',(160,96),'black')
    for n,(offset,width,height) in enumerate([(0,24,32),(0x180,24,32),(0x300,24,32),(0x480,24,32),(0x600,16,24),(0x6c0,16,24),(0x780,16,24),(0x840,16,24)]):
        raw=two_planes(data[offset:],width,height,(0,5,2,7));x=n%4*40+(40-width)//2;y=n//4*48+(48-height)//2
        atlas.paste(Image.fromarray(intro[np.frombuffer(raw,np.uint8)].reshape(height,width,3)),(x,y));parts.append({'raster':base64.b64encode(raw).decode(),'width':width,'height':height,'box':[x,y,x+width,y+height]})
    file=w/'references/complete-hou-v2.png';atlas.resize((960,576),Image.Resampling.NEAREST).save(file)
    rows.append({'key':'hou','kind':'parts','parts':parts,'reference':str(file.relative_to(p)),'atlas_size':[160,96],'archive':1,'record':17,'source_sha256':sha(load(1,17)),'method':'OPDMO SPRITE_A_FRAME_TABLE: four0620 plus four0418 frames; native two-plane mappings A+C/B'})
    data=bytearray(img_open(load(1,19)));bp=0x1028
    for i in range(bp):
        a,b,c=data[i],data[i+bp],data[i+bp*2]; mask=~(a&b&~c)&255
        fill=c&~a&~b&255;data[i]=(a&mask)|fill;data[i+bp]=(b&mask)|fill;data[i+bp*2]=(c&mask)&~fill&255
    scene('maop',picture_six(data,188,88),188,88,rgb_pair,1,19,'OPDMO script-area 2F58 / plane stride1028, original color preparation')
    scene('ne81',picture_six(img_open(load(1,21)),72,81),72,81,rgb_pair,1,21,'ENDMO secondary portrait CX1251')
    scene('oup',picture_six(img_open(load(1,26)),96,88),96,88,rgb_pair,1,26,'OPDMO right-side portrait CX1858')
    data=rle(load(1,30));sar=(p/'workplace/orig/ZELRES1.SAR').read_bytes();record=records(sar)[0];program=sar[record['payload_offset']:record['payload_offset']+record['bytes']]
    base=0x9547-(program.index(b'hou.grp\0')-2);ids=program[0x912b-base:0x912b-base+850];planes=bytearray(0x1a90*3)
    for row in range(25):
        for col in range(34):
            tile=ids[row*34+col];src=tile//40*0x140+tile%40
            for plane in range(3):
                for y in range(8):
                    at=src+plane*0x640+y*40;planes[plane*0x1a90+row*0x110+col+y*34]=data[at] if at<len(data) else 0
    raw=two_planes(planes,136,200,(0,1,2,3))
    scene('ttl2',raw,136,200,rgb_pair,1,30,'OPDMO scene_sprite_d 912B:34×25 tiles of4×8, source pitch28/plane640')
    # 最後一幕先合併YUU3兩平面，再按原遮罩位置加入YUU4，不任意拉伸拆分圖。
    seg=bytearray(65536);data=img_open(load(1,35));seg[0x4000:0x4000+len(data)]=data
    for i in range(0x3000):
        a,b,c=0x4000+i,0x7000+i,0xa000+i
        v=seg[b]&~seg[a]&255;seg[a]|=v;seg[c]=v|(seg[b]&seg[a]);seg[b]&=~v&255
    scene('yuu3-base',picture_six(seg[0x4000:],256,192),256,192,rgb_pair,1,35,'OPDMO merge_gfx_planes, three planes stride3000')
    data=img_open(load(1,36));seg[0xd000:0xd000+len(data)]=data
    for y in range(160):
        for x in range(21):
            src=0xd000+y*21+x;dest=0x4000+0x819+y*64+x
            a,b,c=seg[src],seg[src+0xd20],seg[src+0x1a40]
            keep=~(a|b|c)&255;clear=~(a&b&~c)&255
            seg[dest]=(seg[dest]&keep)|(a&clear);seg[dest+0x3000]=(seg[dest+0x3000]&keep)|(b&clear);seg[dest+0x6000]=(seg[dest+0x6000]&keep)|c
    scene('yuu3-final',picture_six(seg[0x4000:],256,192),256,192,rgb_pair,1,36,'OPDMO YUU3+YUU4 mask composition, three planes stride3000')
    # 結局的原始完整圖面。END7的第三平面由EN72另外載入。
    for key,record,width,height in [('end5',53,228,154),('end4',52,188,114),('end6',54,188,114)]:
        scene(key,picture_six(img_open(load(2,record)),width,height),width,height,rgb_pair,2,record,'ENDMO credits source, 3-plane image')
    data=img_open(load(2,55))+load(2,51)
    scene('end7',picture_six(data,320,134),320,134,rgb_pair,2,55,'ENDMO END7 two planes + EN72 third plane at93C0')
    data=bytearray(data);encoded=load(2,56);controls=int.from_bytes(encoded[:2],'little');cursor=2+controls;fin=bytearray()
    # FIN只執行稀疏解開，OR/AND遮罩不經圖片用的2-bit累積XOR。
    for c in encoded[2:2+controls]:
        for bit in range(7,-1,-1):
            fin.append(encoded[cursor] if c&(1<<bit) else 0);cursor+=bool(c&(1<<bit))
    bp=0x29e0;mask_stride=38*53
    for y in range(53):
        for x in range(38):
            source=y*38+x;target=0xce6+y*80+x
            for plane in range(3):data[plane*bp+target]=(data[plane*bp+target]|fin[source])&(~fin[mask_stride+source]&255)
    scene('fin',picture_six(data,320,134),320,134,rgb_pair,2,56,'ENDMO FIN sparse-only OR/AND mask:19words×53rows at4CE6, no cumulative XOR; source planes29E0/53C0')
    # 物品頁：首段為六件20×18裝備圖，其餘六段為31件16×16圖示。
    def icon(raw,width,height,large=False):
        result=bytearray();pitch=15 if large else 12
        for y in range(height):
            row=raw[y*pitch:(y+1)*pitch]
            groups=[(int.from_bytes(row[0:2],'big'),int.from_bytes(row[8:10],'little'),int.from_bytes(row[10:12],'big')),
                    (int.from_bytes(row[2:4],'big'),int.from_bytes(row[6:8],'little'),int.from_bytes(row[12:14],'big')),
                    (row[4]<<8,row[5]<<8,row[14]<<8)] if large else [
                    (int.from_bytes(row[0:2],'big'),int.from_bytes(row[6:8],'little'),int.from_bytes(row[8:10],'big')),
                    (int.from_bytes(row[2:4],'big'),int.from_bytes(row[4:6],'little'),int.from_bytes(row[10:12],'big'))]
            for q,(a,b,c) in enumerate(groups):
                for x in range(4 if large and q==2 else 8):
                    hi=sum(((v>>(15-x*2))&1)<<k for k,v in enumerate((a,b,c)))
                    lo=sum(((v>>(14-x*2))&1)<<k for k,v in enumerate((a,b,c)))
                    result.append(hi*8+lo)
        return bytes(result)
    data=load(2,27);pointers=list(struct.unpack_from('<7H',data));parts=[]
    for group,start in enumerate(pointers):
        end=pointers[group+1] if group+1<len(pointers) else len(data);size=270 if group==0 else 192
        for n in range((end-start)//size):
            width,height=(20,18) if group==0 else (16,16);raw=icon(data[start+n*size:start+(n+1)*size],width,height,group==0)
            parts.append({'group':group,'phase':n,'width':width,'height':height,'raster':base64.b64encode(raw).decode()})
    part_atlas('inventory-icons',parts,2,27,'GMMCGA large20×18/stride10E; small16×16/strideC0; original word permutation')
    # 法術資料按原六個段指標保留，各段以16×16、192-byte三平面圖格組成。
    data=load(2,28);pointers=sorted(set(struct.unpack_from('<6H',data)));parts=[]
    for group,start in enumerate(pointers):
        end=pointers[group+1] if group+1<len(pointers) else len(data)
        for n in range((end-start)//192):
            chunk=data[start+n*192:start+(n+1)*192];pixels=bytearray()
            for y in range(16):
                for q in range(2):
                    a,b,c=struct.unpack_from('>HHH',chunk,y*12+q*6)
                    for x in range(8):
                        hi=sum(((v>>(15-x*2))&1)<<k for k,v in enumerate((a,b,c)));lo=sum(((v>>(14-x*2))&1)<<k for k,v in enumerate((a,b,c)));pixels.append(hi*8+lo)
            raw=bytes(pixels);parts.append({'group':group,'phase':n,'width':16,'height':16,'raster':base64.b64encode(raw).decode(),'mask':base64.b64encode(bytes(255 if v else 0 for v in raw)).decode()})
    part_atlas('magic-effects',parts,2,28,'MAGIC three-plane16×16 source frames; repeated root0C is shared')
    # 所有三組刀光的2-bit字模，用實際根指標及字模起點，不只列已走過的十姿態。
    data=load(2,26);pointers=list(struct.unpack_from('<3H',data));parts=[]
    for group,start in enumerate(pointers):
        end=pointers[group+1] if group+1<len(pointers) else len(data);begin=start+struct.unpack_from('<H',data,start)[0]
        for n in range((end-begin)//16):
            raw=bytearray()
            for word in struct.unpack_from('>8H',data,begin+n*16):
                raw.extend(0 if ((word>>(14-x*2))&3)==0 else 63 if ((word>>(14-x*2))&3)==3 else 45 for x in range(8))
            parts.append({'group':group,'index':n,'width':8,'height':8,'raster':base64.b64encode(raw).decode(),'mask':base64.b64encode(bytes(255 if v else 0 for v in raw)).decode()})
    part_atlas('weapon-glyphs',parts,2,26,'SWORD three roots, each first-word bitmap pointer, 8×8 two-bit glyphs',16,12)
    # 主角完整230圖格沿原本已保存的FMAN資源快照；不重新執行遊戲。
    memory=(p/'workplace/full-hd-v1/cave-profile-normal-pair-v2/cave-platform/object-missing-00.memory').read_bytes();resource=memory[0x11b10:0x21b10];primary=memory[0x1b10:0x11b10]
    address=struct.unpack_from('<H',primary,0x4f8c)[0];colors=primary[address:address+16];parts=[]
    for n in range(230):
        raw=bytearray();mask=resource[0xd000+n*8:0xd000+(n+1)*8];tile=resource[0x6333+n*32:0x6333+(n+1)*32]
        for y in range(8):
            words=struct.unpack_from('<HH',tile,y*4)
            raw.extend(colors[(words[x//4]>>(12-x%4*4))&15] for x in range(8))
        alpha=bytes(0 if mask[y]&(128>>x) else 255 for y in range(8) for x in range(8));parts.append({'index':n,'width':8,'height':8,'raster':base64.b64encode(raw).decode(),'mask':base64.b64encode(alpha).decode()})
    part_atlas('fman-all',parts,3,51,'original cached FMAN resource6333, D000 masks, countE6 from200FIGHT; palette4F8C',16,12)
    rows[-1]['palette_variants']=[base64.b64encode(primary[struct.unpack_from('<H',primary,0x4f8c+i*2)[0]:struct.unpack_from('<H',primary,0x4f8c+i*2)[0]+16]).decode() for i in range(4)]
    # 城鎮人物及完整地形來源。原專用圖優先，這些圖格補交談與未列姿態。
    for record,key,count,start,family in [(29,'mman-all',152,256,'npc'),(30,'cman-all',164,256,'npc'),(31,'tman-all',46,0,'npc'),(33,'cpat-all',157,256,'terrain'),(34,'mpat-all',242,256,'terrain'),(35,'dpat-all',191,256,'terrain')]:
        data=load(2,record);parts=[]
        for n in range(count):
            raw=data[start+n*48:start+(n+1)*48]
            packed,mask=terrain_preprocess(raw,data[6+n]) if family=='terrain' else preprocess(raw,7,0)
            pixels=bytearray()
            for a,b,c in zip(packed[::3],packed[1::3],packed[2::3]):pixels.extend((b>>2,(b&3)*16+(a>>4),(a&15)*4+(c>>6),c&63))
            alpha=bytes(0 if mask[y]&(128>>x) else 255 for y in range(8) for x in range(8))
            parts.append({'index':n,'width':8,'height':8,'raster':base64.b64encode(pixels).decode(),'mask':base64.b64encode(alpha).decode()})
        part_atlas(key,parts,2,record,'existing 48-byte planar preprocessing and source mask; full resource',16,12)
    # 交談後景內嵌在YMPD/CKPD程式，按圖形資料範圍解開，不執行程式。
    sar=(p/'workplace/orig/ZELRES2.SAR').read_bytes();index=records(sar)
    def program(n):
        row=index[n];return sar[row['payload_offset']:row['payload_offset']+row['bytes']]
    def mountain_decode(code,offset,count):
        output=bytearray()
        while len(output)<count:
            v=code[offset];offset+=1
            if v==6:v,n=code[offset:offset+2];offset+=2
            else:n=1
            output.extend([v]*n)
        return bytes(output[:count])
    def paired(a,b):
        return bytes(((vb>>bit)&1)*32+((va>>bit)&1)*8+((vb>>(bit-1))&1)*4+((va>>(bit-1))&1) for va,vb in zip(a,b) for bit in (7,5,3,1))
    code=program(8);a=mountain_decode(code,0x38e7-0x3300,4928);b=mountain_decode(code,0x4759-0x3300,4928)
    scene('ympd-mountains',paired(a,b),224,88,palette,2,8,'YMPD embedded mountain RLE at38E7/4759; byte-pair expansion',code)
    def ground_decode(code,offset):
        out=bytearray()
        for _ in range(16):
            row=bytearray()
            while len(row)<28:
                v=code[offset];offset+=1
                row.extend(bytes(v&15 or 256) if v&0xf0==0x60 else bytes([v]))
            out.extend(row)
        return out
    ground=ground_decode(code,0x559e-0x3300)+ground_decode(code,0x56f1-0x3300);raw=bytearray()
    for y in range(16):
        base=(0 if y<8 else 448)+(y%8)*56;line=bytearray()
        for a,b in zip(ground[base:base+28],ground[base+28:base+56]):
            line.extend(((a>>bit)&1)*32+((b>>bit)&1)*16+((a>>(bit-1))&1)*4+((b>>(bit-1))&1)*2 for bit in (7,5,3,1))
        raw.extend(line*2)
    scene('ympd-ground',bytes(raw),224,16,palette,2,8,'YMPD ground zero-run558? / ground1 56F1, duplicated112px bands',code)
    code=program(9)
    def ckpd_decode(at):
        out=bytearray();cursor=at-0x3300
        while True:
            v=code[cursor];cursor+=1
            if v==0:return bytes(out)
            if v&0xf0 in (0x10,0x40):out.extend([0 if v&0xf0==0x10 else 0xaa]*(v&15 or 256))
            else:out.append(v)
    for key,pa,pb,width,height in [('ckpd-backdrop',0x38e0,0x426a,224,72),('ckpd-band',0x4c6d,0x4db8,112,16)]:
        a,b=ckpd_decode(pa),ckpd_decode(pb);n=width//4*height
        scene(key,paired(a[:n].ljust(n,b'\0'),b[:n].ljust(n,b'\0')),width,height,palette,2,9,'CKPD embedded RLE; original init clears0FC0 words before decode, unwritten tail stays zero; byte-pair expansion',code)
    scene('ckpd-foreground',paired(code[0x4f25-0x3300:0x4f25-0x3300+448],code[0x50e5-0x3300:0x50e5-0x3300+448]),112,16,palette,2,9,'CKPD two embedded foreground planes4F25/50E5',code)
    # ZEL2與ZELA共用圖格，程式模板另列，避免只製作第一組模板。
    plan=json.loads((w/'plan.json').read_text());bank=next(r for r in plan['assets'] if r['key']=='zela')
    sar=(p/'workplace/orig/ZELRES3.SAR').read_bytes();index=records(sar);entry=index[15];code=sar[entry['payload_offset']:entry['payload_offset']+entry['bytes']]
    end=struct.unpack_from('<H',code)[0];pointers=sorted(set(struct.unpack_from('<H',code,at)[0] for at in range(0x30,0xb0,2) if 0xa030<=struct.unpack_from('<H',code,at)[0]<min(end,0xa000+len(code))));parts=[];seen=set()
    for n,begin in enumerate(pointers):
        limit=pointers[n+1] if n+1<len(pointers) else end
        for at in range(begin,min(limit,begin+5*64)-4,5):
            values=code[at-0xa000:at-0xa000+5]
            if len(values)!=5 or values[0]>=8 or max(values[1:])>=bank['tile_count']:break
            if values in seen:continue
            seen.add(values);address=struct.unpack_from('<H',primary,0x4f8c+values[0]*2)[0];pal=primary[address:address+16];raw=bytearray(256)
            for q,tile in enumerate(values[1:]):
                for i,v in enumerate(bank['tiles'][tile]['nibbles']):raw[(q//2*8+i//8)*16+q%2*8+i%8]=pal[v]
            parts.append({'width':16,'height':16,'raster':base64.b64encode(raw).decode(),'mask':base64.b64encode(bytes(255 if v else 0 for v in raw)).decode(),'template':at,'palette':values[0],'tiles':list(values[1:])})
    part_atlas('zel2-poses',parts,3,67,'ZEL2.BIN record15 own A030/A070 templates over ZELA.GRP record67',8,24)
    rows[-1]['program_sha256']=sha(code)
    bank=next(r for r in plan['assets'] if r['key']=='encnt');address=struct.unpack_from('<H',primary,0x4f8c)[0];pal=primary[address:address+16];parts=[]
    for n,tile in enumerate(bank['tiles']):
        raw=bytes(pal[v] for v in tile['nibbles']);parts.append({'width':8,'height':8,'index':n,'raster':base64.b64encode(raw).decode(),'mask':base64.b64encode(bytes(255 if v else 0 for v in raw)).decode(),'nibbles':tile['nibbles']})
    part_atlas('encnt-all',parts,3,55,'existing ENCNT nibble source, all32 tile components; usage remains manual-review pending',8,12)
    # 原版portrait指令的嘴形矩形。先回貼完整頭像，避免生成孤立嘴形而失去角色一致性。
    # OPDMO portrait_lg/sm；ENDMO portrait80/90/B0，位址均相對4000/8000載入底。
    for record,key,regions in [
        (37,'portrait-yuup-all',[(0x18c0,36,32,6,32,56),(0x2d00,44,16,6,28,32)]),
        (26,'portrait-oup-all',[(0x18c0,56,32,6,24,56),(0x3840,44,16,3,24,32)]),
        (16,'portrait-himp-all',[(0x18c0,36,24,6,28,56),(0x27f0,40,24,3,24,32)]),
        (28,'portrait-seip-all',[(0x18c0,28,24,6,44,56)])]:
        data=img_open(load(1,record));base=picture_six(data,96,88);parts=[]
        for offset,width,height,count,x,y in regions:
            stride=width*height*3//4
            for phase in range(count):
                patch=picture_six(data[offset+phase*stride:offset+(phase+1)*stride],width,height);frame=bytearray(base)
                for row in range(height):frame[(y+row)*96+x:(y+row)*96+x+width]=patch[row*width:(row+1)*width]
                parts.append({'width':96,'height':88,'raster':base64.b64encode(frame).decode(),'intro':True,'offset':offset+phase*stride,'patch':[x,y,width,height]})
        part_atlas(key,parts,1,record,'OPDMO portrait_lg/sm and ENDMO portrait80/90/B0 source rectangles over original96×88 base',4,112)
        rows[-1]['colors']=base64.b64encode(rgb_pair.tobytes()).decode()
    # 魔王幻影的五個原始雙平面階段，沿busy_wait_delay/cycle_palette_colors的轉換。
    data=img_open(load(1,14));parts=[]
    for phase in range(5):
        start=0x1380+phase*0xcc0;chunk=data[start:start+0xcc0];stride=34*48;a,b=chunk[:stride],chunk[stride:]
        planes=bytes(a)+bytes(vb&~va for va,vb in zip(a,b))+bytes(va^(vb&~va) for va,vb in zip(a,b))
        raw=picture_six(planes,136,48);parts.append({'width':136,'height':48,'raster':base64.b64encode(raw).decode(),'intro':True,'phase':phase})
    part_atlas('dmaou-cycle-all',parts,1,14,'OPDMO busy_wait_delay five CC0-byte states, original A / B&~A / A xor(B&~A) expansion',3,144)
    rows[-1]['colors']=base64.b64encode(rgb_pair.tobytes()).decode()
    # 室內完整圖格保留正確的48-byte格式；專用HD優先，未列動畫由完整來源接續。
    for record,key in [(18,'king-all'),(19,'omoya-all'),(20,'armor-all'),(21,'bank-all'),(22,'church-all'),(23,'drug-all'),(24,'inn-all'),(25,'kenjya-all')]:
        _,native=tiles(load(2,record));count=len(native);columns=16;h=math.ceil(count/columns)
        atlas=Image.new('RGBA',(columns*12,h*12));parts=[]
        for n,raw in enumerate(native):
            im=Image.fromarray(palette[np.frombuffer(raw,np.uint8)].reshape(8,8,3)).convert('RGBA');atlas.paste(im,(n%columns*12+2,n//columns*12+2))
            parts.append({'raster':base64.b64encode(raw).decode(),'index':n})
        file=w/'references'/f'complete-{key}-v1.png';atlas.resize((1152,h*72),Image.Resampling.NEAREST).save(file)
        rows.append({'key':key,'kind':'tile-atlas','columns':columns,'rows':h,'parts':parts,'reference':str(file.relative_to(p)),
                     'archive':2,'record':record,'source_sha256':sha(load(2,record)),'method':'existing 48-byte planar tile format; dedicated HD retains priority'})
    contact=Image.new('RGB',(1600,math.ceil(len(rows)/4)*340),'#16202a');draw=ImageDraw.Draw(contact)
    for n,row in enumerate(rows):
        im=Image.open(p/row['reference']).convert('RGB');im.thumbnail((388,304),Image.Resampling.NEAREST)
        x=n%4*400;y=n//4*340;draw.text((x+8,y+6),row['key'],fill='white');contact.paste(im,(x+(400-im.width)//2,y+28))
    contact.save(w/'references/completion-contact-v1.png')
    sheets=[]
    for number,keys in enumerate([['dmaou','maop','yuu3-base','yuu3-final'],['end5','end4','end6','end7'],['ne81','oup','ttl2','fin']],1):
        sheet=Image.new('RGB',(1536,1024),'black');cells=[]
        for n,key in enumerate(keys):
            row=next(r for r in rows if r['key']==key);im=Image.open(p/row['reference']).convert('RGB');ratio=min(720/im.width,460/im.height)
            im=im.resize((round(im.width*ratio),round(im.height*ratio)),Image.Resampling.NEAREST)
            x=n%2*768+(768-im.width)//2;y=n//2*512+(512-im.height)//2;sheet.paste(im,(x,y))
            cells.append({'key':key,'box':[x,y,x+im.width,y+im.height]})
        file=w/'references'/f'complete-scenes-{number}-v1.png';sheet.save(file)
        sheets.append({'key':f'complete-scenes-{number}','reference':str(file.relative_to(p)),'cells':cells})
    sheet=Image.new('RGB',(1536,1024),'black');cells=[]
    for key,slot in [('ympd-mountains',(24,24,744,400)),('ckpd-backdrop',(792,24,1512,400)),('ympd-ground',(24,460,1512,600)),('ckpd-band',(24,650,1512,790)),('ckpd-foreground',(24,840,1512,980))]:
        row=next(r for r in rows if r['key']==key);im=Image.open(p/row['reference']).convert('RGB');x0,y0,x1,y1=slot;ratio=min((x1-x0)/im.width,(y1-y0)/im.height)
        im=im.resize((round(im.width*ratio),round(im.height*ratio)),Image.Resampling.NEAREST);x=x0+(x1-x0-im.width)//2;y=y0+(y1-y0-im.height)//2;sheet.paste(im,(x,y));cells.append({'key':key,'box':[x,y,x+im.width,y+im.height]})
    file=w/'references/complete-backgrounds-v1.png';sheet.save(file);sheets.append({'key':'complete-backgrounds','reference':str(file.relative_to(p)),'cells':cells})
    (w/'completion-sources.json').write_text(json.dumps({'schema':1,'status':'ART_IN_PRODUCTION','automatic_tests':False,'game_launched':False,'sources':rows,'sheets':sheets},ensure_ascii=False,indent=2)+'\n')
    print('PREPARED',len(rows),'source groups; no game or tests executed')


if __name__=='__main__':main()

"""洞窟HD離線研究工具。DRAFT，不接正式前端或改寫原版。

在Docker內執行。baseline與study應唯讀，out須是另外掛載的空白可寫目錄。
來源與限制見docs/spec/011-cave-hd.md及CONTEXT.md。
"""
import argparse
from pathlib import Path
import hashlib, json, struct, zlib, binascii

parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--baseline',type=Path,required=True)
parser.add_argument('--study',type=Path,required=True)
parser.add_argument('--out',type=Path,required=True)
args=parser.parse_args()
b,study,out=args.baseline,args.study,args.out
if not b.is_dir() or not study.is_dir() or not out.is_dir():parser.error('輸入與輸出目錄必須已存在')
if any(out.iterdir()):parser.error('輸出目錄必須為空，研究產物不覆寫')
h=lambda data:hashlib.sha256(data).hexdigest()
w=lambda data,off:struct.unpack_from('<H',data,off)[0]

def decode_png(path):
    data=path.read_bytes();assert data[:8]==b'\x89PNG\r\n\x1a\n'
    pos=8;stream=bytearray()
    while pos<len(data):
        n=struct.unpack_from('>I',data,pos)[0];kind=data[pos+4:pos+8];payload=data[pos+8:pos+8+n];pos+=12+n
        if kind==b'IHDR':width,height,depth,color,compression,filtering,interlace=struct.unpack('>IIBBBBB',payload)
        if kind==b'IDAT':stream.extend(payload)
    assert depth==8 and color in [2,6] and interlace==0
    stride=width*(3 if color==2 else 4);channels=3 if color==2 else 4
    raw=zlib.decompress(stream);rows=[];prev=bytearray(stride)
    for y in range(height):
        f=raw[y*(stride+1)];row=bytearray(raw[y*(stride+1)+1:(y+1)*(stride+1)])
        for i in range(stride):
            a=row[i-channels]if i>=channels else 0;bb=prev[i];c=prev[i-channels]if i>=channels else 0
            if f==1:row[i]=(row[i]+a)&255
            elif f==2:row[i]=(row[i]+bb)&255
            elif f==3:row[i]=(row[i]+(a+bb)//2)&255
            elif f==4:
                p=a+bb-c;pa,pb,pc=abs(p-a),abs(p-bb),abs(p-c);predict=a if pa<=pb and pa<=pc else bb if pb<=pc else c
                row[i]=(row[i]+predict)&255
            else:assert f==0
        rows.extend(row);prev=row
    if channels==3:
        rows=bytes(v for i in range(width*height)for v in [*rows[i*3:i*3+3],255])
    return width,height,bytes(rows)

def encode_png(width,height,pixels):
    def chunk(name,data):return struct.pack('>I',len(data))+name+data+struct.pack('>I',binascii.crc32(name+data)&0xffffffff)
    return b'\x89PNG\r\n\x1a\n'+chunk(b'IHDR',struct.pack('>IIBBBBB',width,height,8,6,0,0,0))+chunk(b'IDAT',zlib.compress(b''.join(b'\0'+pixels[y*width*4:(y+1)*width*4]for y in range(height))))+chunk(b'IEND',b'')

aw,ah,art=decode_png(study/'terrain-handpaint-v2.png')
assert h((study/'terrain-handpaint-v2.png').read_bytes())=='07004b059f8ed0ecf50c9aaf3a8b35c965050725fc652325d5ef248842a0b1bc'
tiles={}
for slot in range(1,15):
    colors=[]
    for y in range(24):
        top=ah*((slot//4)*24+y);bottom=top+ah
        for x in range(24):
            left=aw*((slot%4)*24+x);right=left+aw;totals=[0,0,0,0]
            for yy in range(top//96,(bottom+95)//96):
                wy=min(bottom,(yy+1)*96)-max(top,yy*96)
                for xx in range(left//96,(right+95)//96):
                    wx=min(right,(xx+1)*96)-max(left,xx*96);i=(yy*aw+xx)*4;weight=wx*wy
                    for c in range(3):totals[c]+=art[i+c]*art[i+3]*weight
                    totals[3]+=art[i+3]*weight
            area=aw*ah;alpha=(totals[3]+area//2)//area
            colors.append(bytes((v+totals[3]//2)//totals[3]for v in totals[:3])+bytes([alpha])if totals[3]else b'\0\0\0\0')
    tiles[slot]=colors

# 連續材質沿原版36x64循環地圖地址定位；鏡射閉合接縫，不新增地形。
mw,mh,material=decode_png(study/'rock-material-v1.png')
assert h((study/'rock-material-v1.png').read_bytes())=='09d331a411515a441790479646484467ca529517adc3f3cc235936921e57bf8f'
tw,th=432,768
weights_x=[[(i,min((x+1)*mw,(i+1)*tw)-max(x*mw,i*tw))for i in range(x*mw//tw,((x+1)*mw+tw-1)//tw)]for x in range(tw)]
weights_y=[[(i,min((y+1)*mh,(i+1)*th)-max(y*mh,i*th))for i in range(y*mh//th,((y+1)*mh+th-1)//th)]for y in range(th)]
material_cache=[]
for y in range(th):
    for x in range(tw):
        total=[0,0,0,0]
        for yy,wy in weights_y[y]:
            for xx,wx in weights_x[x]:
                i=(yy*mw+xx)*4;weight=wx*wy
                for k in range(3):total[k]+=material[i+k]*material[i+3]*weight
                total[3]+=material[i+3]*weight
        alpha=(total[3]+mw*mh//2)//(mw*mh)
        material_cache.append(bytes((v+total[3]//2)//total[3]for v in total[:3])+bytes([alpha])if total[3]else b'\0\0\0\0')

records=[]
for name in ['cave-idle-1','advance-strike-2','jump-key-500k','down-key-250k']:
    frame_dir=study/'frames-v3';frame_name=name+'-pc-3123'
    point=next(x for x in json.loads((frame_dir/'completed-receipt.json').read_text())['checkpoints']if x['name']==frame_name)
    memory=(frame_dir/(frame_name+'.memory')).read_bytes();original_raw=(frame_dir/(frame_name+'.indexed')).read_bytes();raw=bytearray(original_raw)
    assert len(memory)==1048576 and len(original_raw)==64000
    assert h(memory)==point['memory_sha256'] and h(original_raw)==point['indexed_sha256']
    assert h(memory[0x1b10+0xc000:0x1b10+0xc040])=='3f5edbc8b51aa017211fb0ca808f807a1ade51d95e14b49dadf593e90ada24eb'
    palette=(frame_dir/(frame_name+'.palette')).read_bytes();mp=memory[0x1b10:0x11b10]
    body_memory=(frame_dir/point['body']['path']).read_bytes()if point.get('body')else memory
    bodyp=body_memory[0x1b10:0x11b10]
    underlay=(frame_dir/point['underlay']['path']).read_bytes()[0x1b10:0x11b10]
    actor=(frame_dir/(frame_name+'.actor-v2.indexed')).read_bytes()
    bodypos=w(bodyp,0x4fed);bx=bodypos%320;by=bodypos//320
    weapon_pixels={}
    if mp[0xff44]:
        resource=w(mp,0xff2c)*16;res=memory[resource:resource+65536]
        template=w(mp,0x4ff9);atlas=w(res,0xb000);position=w(mp,0x4ff7);wx=position%320;wy=position//320
        saved=memory[0xafa40:0xafe40]
        for y in range(32):
            for x in range(32):raw[(wy+y)*320+wx+x]=saved[y*32+x]
        for n,tile in enumerate(res[template:template+16]):
            if tile==255:continue
            data=res[atlas+tile*16:atlas+(tile+1)*16]
            for y in range(8):
                bits=(data[y*2]<<8)|data[y*2+1]
                for x in range(8):
                    v=(bits>>(14-x*2))&3
                    if v:weapon_pixels[wx+(n//4)*8+x,wy+(n%4)*8+y]=mp[0x4ff6]if v==3 else mp[0x4ff5]
    for y in range(24):
        for x in range(24):raw[(by+y)*320+bx+x]=underlay[0x511d+(y//8*3+x//8)*64+(y%8)*8+x%8]
    actor_pixels={(bx+x,by+y):actor[y*24+x]for y in range(24)for x in range(24)if actor[y*24+x]!=255}
    rebuild=bytearray(raw)
    for (x,y),v in actor_pixels.items():rebuild[y*320+x]=v
    for (x,y),v in weapon_pixels.items():rebuild[y*320+x]=v
    assert bytes(rebuild)==original_raw,'entire native original reconstruction'

    p=memory[0x1b10:0x11b10];seg=w(p,0xff2c);r=memory[seg*16:seg*16+65536];origin=w(p,0xff31);assert 0xe000<=origin<0xe900 and seg==0x11b1
    width,height=960,600
    def target_from_raw(values):
        return bytes(v for y in range(600)for x in range(960)for v in [*palette[values[(y//3)*320+x//3]*3:values[(y//3)*320+x//3]*3+3],255])
    base=target_from_raw(original_raw);canvas=bytearray(target_from_raw(raw))
    _,_,chinese=decode_png(b/'pair-v3'/(name+'.png'))
    patch_boxes=[(56,163,21,8),(123,187,26,8),(53,187,21,8),(53,175,26,8),(88,175,86,8)]
    baseline_rgb=(b/'entry-v3'/(name+'.rgb')).read_bytes()
    for px,py,pw,ph in patch_boxes:
        for yy in range(py,py+ph):
            for xx in range(px,px+pw):
                assert palette[original_raw[yy*320+xx]*3:original_raw[yy*320+xx]*3+3]==baseline_rgb[(yy*320+xx)*3:(yy*320+xx)*3+3],'native label source mismatch'

    # 靜態圖塊以黑色直接覆寫；0不是通用透明色。
    # 美術輪廓：保留原版空格與每條連續岩面欄的首末可見像素，內部黑點視為材質。
    decoded_cache={}
    def source_tile(identity):
        if identity not in decoded_cache:
            values=[];off=0x8030+(identity-1)*48
            for n in range(off,off+48,3):
                a,bb,c=r[n:n+3];values.extend([bb>>2,((bb&3)<<4)|(a>>4),((a&15)<<2)|(c>>6),c&63])
            decoded_cache[identity]=values
        return decoded_cache[identity]
    rock_geometry=set()
    for column in range(28):
        identities=[p[0xe000+(origin-0xe000+row*36+4+column)%0x900]for row in range(18)]
        row=0
        while row<18:
            if not 3<=identities[row]<=8:row+=1;continue
            first=row
            while row<18 and 3<=identities[row]<=8:row+=1
            last=row
            for x in range(8):
                occupied=[y for y in range(first*8,last*8)if source_tile(identities[y//8])[(y%8)*8+x]!=0]
                if occupied:
                    for y in range(min(occupied),max(occupied)+1):rock_geometry.add((column*8+x,y))
    cells=[];mask=set();rejected=[]
    for row in range(18):
        for col in range(28):
            addr=0xe000+(origin-0xe000+row*36+4+col)%0x900;identity=p[addr]
            if identity not in tiles:continue
            decoded=[];off=0x8030+(identity-1)*48
            for n in range(off,off+48,3):
                a,bb,c=r[n:n+3];decoded.extend([bb>>2,((bb&3)<<4)|(a>>4),((a&15)<<2)|(c>>6),c&63])
            actual=bytes(raw[(14+row*8+y)*320+48+col*8+x]for y in range(8)for x in range(8))
            if actual!=bytes(decoded):rejected.append([col,row]);continue
            cells.append({'cell':[col,row],'identity':identity})
            for y in range(24):
                for x in range(24):
                    if 3<=identity<=8:
                        if (col*8+x//3,row*8+y//3)not in rock_geometry:continue
                    elif decoded[(y//3)*8+x//3]==0:continue
                    xx=(48+col*8)*3+x;yy=(14+row*8)*3+y;i=(yy*960+xx)*4
                    c=tiles[identity][y*24+x]
                    if 3<=identity<=8:
                        ring=addr-0xe000;tx=(ring%36)*24+x;ty=(ring//36)*24+y
                        tx=min(tx,863-tx);ty=min(ty,1535-ty)
                        c=material_cache[ty*432+tx]
                    alpha=c[3]
                    for k in range(3):canvas[i+k]=(c[k]*alpha+canvas[i+k]*(255-alpha)+127)//255
                    mask.add(yy*960+xx)
    # 地形之後還原原版角色及武器；其不透明像素保持原貌。
    for positions in [actor_pixels,weapon_pixels]:
        for (xx,yy),value in positions.items():
            color=palette[value*3:value*3+3]+bytes([255])
            for dy in range(3):
                for dx in range(3):
                    i=(((yy*3+dy)*960)+xx*3+dx)*4;canvas[i:i+4]=color
    for (xx,yy) in set(actor_pixels)|set(weapon_pixels):
        for dy in range(3):
            for dx in range(3):
                i=(((yy*3+dy)*960)+xx*3+dx)*4;assert canvas[i:i+4]==base[i:i+4],'original actor/weapon pixel changed'
    # 只複製已驗的五個中文靜態標籤，原版數字與生命條仍取當次完成來源。
    base=bytearray(base)
    for px,py,pw,ph in patch_boxes:
        for yy in range(py*3,(py+ph)*3):
            for xx in range(px*3,(px+pw)*3):
                i=(yy*960+xx)*4;canvas[i:i+4]=chinese[i:i+4];base[i:i+4]=chinese[i:i+4]
    changed=sum(canvas[i:i+4]!=base[i:i+4]for i in range(0,len(base),4))
    outside=sum(canvas[i:i+4]!=base[i:i+4]for i in range(0,len(base),4)if i//4 not in mask)
    assert outside==0
    path=name+'-complete-terrain-hd-v5.png';data=encode_png(960,600,canvas);(out/path).write_bytes(data)
    records.append({'name':name,'source_kind':'normal original3123 completion snapshot; not command live','source_steps':point['steps'],'source_cycles':point['cycles'],'entire_native_original_reconstruction_differences':0,'original_actor_weapon_pixels_differences':0,'native_label_bitmaps_equal':5,'source_memory_sha256':h(memory),'original_indexed_sha256':h(original_raw),'stripped_underlay_indexed_sha256':h(raw),'Chinese_base_png_sha256':h((b/'pair-v3'/(name+'.png')).read_bytes()),'image':path,'image_sha256':h(data),'accepted_cells':len(cells),'source_cells':cells,'rejected_cells':rejected,'allowed_target_pixels':len(mask),'changed_target_pixels':changed,'outside_mask_differences':outside})
(out/'receipt.json').write_text(json.dumps({'records':records,'art_sha256':h((study/'terrain-handpaint-v2.png').read_bytes()),'continuous_material_sha256':h((study/'rock-material-v1.png').read_bytes()),'material_mapping':'IDs3..8 use36x64 circular-map byte address,24HDpixel percell,mirrored864x1536 period and explicit artistic rock-envelope; otherknown IDs keep atlas with source nonzero mask','sampling':'rational-boundary premultiplied RGBA integer area average directly to24x24 per8x8native;1254px atlas never reduced to original8px','geometry':'artistic rock-envelope within original IDs3..8 runs; preserve mapempty cells and first/last visible sourcepixel percolumn, replace internal black shade; otherIDs keep source nonzero mask; no collision or source writes','scope':'DRAFT offline completed-source static terrain; full original reconstruction beforeHD; character/weapon restored afterHD, five verified Chinese label patches; no production or GUI integration'},ensure_ascii=False,indent=2)+'\n')
print(json.dumps([{'name':x['name'],'cells':x['accepted_cells'],'changed':x['changed_target_pixels'],'outside':x['outside_mask_differences']}for x in records]))

#!/usr/bin/env python3
"""050人工候選的首領與片頭參考圖。只處理本地資產，不啟動遊戲。"""
import argparse
import base64
import hashlib
import json
import math
import struct
from pathlib import Path
from PIL import Image
import numpy as np
from hd_inventory import records


def rle(data):
    out = bytearray()
    i = 0
    while i < len(data):
        token = data[i]; i += 1
        if token & 64:
            token = token * 256 + data[i]; i += 1
            if token == 65535: break
            count, fill = token & 16383, token & 32768
        else:
            count, fill = token & 63, token & 128
        if fill:
            out.extend(data[i:i+1] * count); i += 1
        else:
            out.extend(data[i:i+count]); i += count
        if len(out) > 200000: raise ValueError('圖片過大')
    return bytes(out)


def img_open(data):
    count = int.from_bytes(data[:2], 'little')
    i = 2 + count
    out = bytearray()
    state = 0
    for ctrl in data[2:2+count]:
        for bit in range(7, -1, -1):
            v = data[i] if ctrl & (1 << bit) else 0
            i += bool(ctrl & (1 << bit))
            result = 0
            for shift in (6, 4, 2, 0):
                state ^= (v >> shift) & 3
                result |= state << shift
            out.append(result)
    return bytes(out)


def picture(planes, width, height, gfx=False):
    bp = width // 4 * height
    if len(planes) < bp*2: raise ValueError(f'平面過短 {len(planes)} / {bp*2}')
    lookup = (0, 1, 8, 9) if gfx else (0, 10, 12, 8)
    raster = bytearray()
    for a, b in zip(planes[:bp], planes[bp:bp*2]):
        nibbles = [lookup[((a >> bit) & 1) | (((b >> bit) & 1) << 1)] for bit in range(7, -1, -1)]
        raster.extend(nibbles[i]*16+nibbles[i+1] for i in range(0, 8, 2))
    return bytes(raster)


def picture_six(planes, width, height):
    bp = width // 4 * height
    result = bytearray()
    for i in range(bp):
        a,b,c=planes[i],planes[bp+i],planes[bp*2+i]
        for bit in (7,5,3,1):
            hi=((c>>bit)&1)*4+((b>>bit)&1)*2+((a>>bit)&1)
            lo=((c>>(bit-1))&1)*4+((b>>(bit-1))&1)*2+((a>>(bit-1))&1)
            result.append(hi*16+lo)
    return bytes(result)


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--only-boss', action='store_true')
    args = ap.parse_args(); p = args.project
    w = p/'workplace/full-hd-v1/manual-hd-v1'; refs = w/'references'
    plan = json.loads((w/'plan.json').read_text())
    prior=w/'plan-v4-before-index-correction.json'
    if not prior.exists():prior.write_text(json.dumps(plan,ensure_ascii=False,indent=2)+'\n')
    palette = (p/'workplace/music-release-v1/release-build/AppDir/assets/town/frame/palette.bin').read_bytes()
    pal = np.frombuffer(palette, np.uint8).reshape(256, 3)
    memory = (p/'workplace/full-hd-v1/cave-profile-normal-pair-v2/cave-platform/object-missing-00.memory').read_bytes()
    primary = memory[0x1b10:0x11b10]
    word = lambda b, at: struct.unpack_from('<H', b, at)[0]
    sar = (p/'workplace/orig/ZELRES3.SAR').read_bytes(); rec = records(sar)
    ledger = json.loads((p/'docs/hd-coverage.json').read_text())['records']
    programs = {}
    program_names = {}
    for bank in plan['assets']:
        if not 64 <= bank['record'] <= 73: continue
        name = bank['key'].upper()+'.BIN'
        candidates = [row for row in ledger if row['archive']=='ZELRES3.SAR'
                      and any(n['name']==name for n in row['names'])]
        if len(candidates)!=1: raise ValueError('首領程式名稱必須唯一：'+name)
        programs[bank['record']] = candidates[0]['record']
        program_names[bank['record']] = name
    # 首領各自的A030模板；保存元件而不推測整隻首領的行為或排列。
    for bank in plan['assets']:
        if bank['record'] not in programs: continue
        pr = rec[programs[bank['record']]]
        program = sar[pr['payload_offset']:pr['payload_offset']+pr['bytes']]
        start = word(program, 0)
        pointers = sorted(set(word(program, off) for off in range(0x30, min(0xb0,len(program)-1), 2)
                              if 0xa030 <= word(program,off) < min(start,0xa000+len(program))))
        frames, seen = [], set()
        for n, pointer in enumerate(pointers):
            end = pointers[n+1] if n+1 < len(pointers) else start
            for pos in range(pointer, min(end,pointer+5*64)-4, 5):
                entry = program[pos-0xa000:pos-0xa000+5]
                if len(entry)!=5 or entry[0]>=8 or max(entry[1:])>=bank['tile_count']: break
                if entry in seen: continue
                seen.add(entry); frames.append({'address':pos,'palette':entry[0],'tiles':list(entry[1:])})
        bank.update(frames=frames,columns=8,rows=max(1,math.ceil(len(frames)/8)),program_record=programs[bank['record']],program_name=program_names[bank['record']],program_mapping_level='strong inference: filename descriptor ledger; manual review pending',program_sha256=hashlib.sha256(program).hexdigest())
        atlas = Image.new('RGBA',(192,bank['rows']*24))
        for n,f in enumerate(frames):
            pt = word(primary,0x4f8c+f['palette']*2); colors = primary[pt:pt+16]
            raw = np.zeros((16,16),np.uint8)
            for q,t in enumerate(f['tiles']): raw[q//2*8:q//2*8+8,q%2*8:q%2*8+8]=np.array([colors[v] for v in bank['tiles'][t]['nibbles']],np.uint8).reshape(8,8)
            rgba=np.concatenate([pal[raw],np.where(raw!=0,255,0).astype(np.uint8)[:,:,None]],axis=2)
            atlas.paste(Image.fromarray(rgba),(n%8*24+4,n//8*24+4))
        file=refs/(bank['key']+'-own-components-v5.png'); atlas.resize((1152,bank['rows']*144),Image.Resampling.NEAREST).save(file)
        bank['reference']=str(file.relative_to(p))
        print(bank['key'],len(frames),'components')
    (w/'plan.json').write_text(json.dumps(plan,ensure_ascii=False,indent=2)+'\n')
    if args.only_boss:return
    # 固定尺寸來自同版本公開格式資料；GRP與img_open為不同內層編碼。
    defs = [(13,'ame',288,104,False),(14,'dmaou',136,112,True),(15,'hime',288,104,False),
            (16,'himp',96,88,False),(17,'hou',64,64,True),(18,'isi',288,104,False),
            (19,'maop',192,93,False),(20,'ne80',104,100,False),(21,'ne81',104,56,False),(22,'nec',176,104,True),
            (23,'new1',96,88,False),(24,'new2',112,100,False),(25,'oui',288,104,False),
            (26,'oup',112,100,False),(27,'sei',144,104,False),(28,'seip',96,88,False),
            (29,'ttl1',196,128,False),(30,'ttl2',128,50,False),(31,'ttl3',260,112,False),
            (32,'waku',320,136,False),(33,'yuu1',288,104,False),
            (34,'yuu2',196,96,False),(35,'yuu3',256,192,True),(36,'yuu4',176,76,False),(37,'yuup',96,88,False)]
    # 使用原生色號，執行時依色盤呈現；美術參考採原版片頭既有色盤。
    intro_pal=(p/'workplace/probe-intro-display/palette.bin').read_bytes()
    intro_pal=np.frombuffer(intro_pal,np.uint8).reshape(256,3)
    scenes=[]
    for recno,key,width,height,gfx in defs:
        src=p/f'workplace/full-hd-v1/decoded-v2/zelres1-{recno:02d}.decoded'
        data=src.read_bytes()
        try:
            codec='6de1' if key.startswith('ttl') else 'img_open'
            planes=rle(data) if codec=='6de1' else img_open(data)
            if key=='oup':planes=planes[0x1000:]
            bp=width//4*height
            six=len(planes)>=bp*3
            raw=picture_six(planes,width,height) if six else picture(planes,width,height,gfx)
        except (ValueError,IndexError) as e:
            print('DEFER_FORMAT',key,str(e)); continue
        # 六位元為兩個三位元子像素，位元順序與PC-98相同（1藍、2紅、4綠）；參考圖用其平均色，不借用城鎮色盤。規格068。
        reference_pal=np.array([[(((i>>4)&2)+(i&2))*255//4,(((i>>4)&4)+(i&4))*255//8,(((i>>4)&1)+(i&1))*255//2] for i in range(256)],np.uint8) if six else intro_pal
        pixels=reference_pal[np.frombuffer(raw,np.uint8)].reshape(height,width,3)
        ref=refs/(key+'-picture-v1.png'); Image.fromarray(pixels).resize((width*3,height*3),Image.Resampling.NEAREST).save(ref)
        scenes.append({'key':key,'record':recno,'width':width,'height':height,'reference':str(ref.relative_to(p)),
                       'raster':base64.b64encode(raw).decode(),'source_sha256':hashlib.sha256(data).hexdigest(),
                       'codec':codec,'six_bit':six,'format_status':'MANUAL_REVIEW_PENDING'})
        print(key,width,height,'image')
    (w/'pictures.json').write_text(json.dumps({'scenes':scenes,'status':'MANUAL_REVIEW_PENDING'},ensure_ascii=False,indent=2)+'\n')
    sheets=[]
    for start in range(0,len(scenes),4):
        batch=scenes[start:start+4];sheet=Image.new('RGB',(1536,1024),'black');cells=[]
        for n,row in enumerate(batch):
            im=Image.open(p/row['reference']);ratio=min(720/im.width,460/im.height)
            size=tuple(round(v*ratio) for v in im.size);im=im.resize(size,Image.Resampling.NEAREST)
            x=n%2*768+(768-size[0])//2;y=n//2*512+(512-size[1])//2
            sheet.paste(im,(x,y));cells.append({'key':row['key'],'box':[x,y,x+size[0],y+size[1]]})
        file=refs/f'story-sheet-{start//4+1}.png';sheet.save(file)
        sheets.append({'key':f'story{start//4+1}','reference':str(file.relative_to(p)),'cells':cells})
    (w/'picture-sheets.json').write_text(json.dumps({'sheets':sheets},ensure_ascii=False,indent=2)+'\n')


if __name__ == '__main__': main()

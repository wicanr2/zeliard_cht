#!/usr/bin/env python3
"""全圖補繪登錄。只裁切、取樣與套用原透明遮罩，不重畫、不啟動遊戲。"""
import argparse
import base64
import hashlib
import json
import shutil
import struct
from pathlib import Path

from PIL import Image
from hd_grp_sources import preprocess, terrain_preprocess
from hd_interior_sources import tiles


def digest(data):
    return hashlib.sha256(data).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--cache', type=Path, required=True)
    args = ap.parse_args()
    p = args.project
    w = p/'workplace/full-hd-v1/manual-hd-v1'
    out = w/'registered'
    sources = json.loads((w/'completion-sources.json').read_text())
    images, requests = {}, []
    for request in sorted((w/'generated').glob('complete-*-request.json')):
        meta = json.loads(request.read_text())
        if not meta.get('selected'):
            continue
        cache = args.cache/Path(meta['cache']).relative_to('/home/anr2/.codex/generated_images')
        target = w/'generated'/f'{meta["key"]}-generated-v1.png'
        shutil.copyfile(cache, target)
        images[meta['key']] = Image.open(target).convert('RGBA')
        meta['file'] = str(target.relative_to(p))
        meta['sha256'] = digest(target.read_bytes())
        refs = [p/Path(f).relative_to('/home/anr2/cht/zeliard') for f in meta['refs']]
        meta['reference_sha256'] = {str(f.relative_to(p)): digest(f.read_bytes()) for f in refs}
        request.write_text(json.dumps(meta, ensure_ascii=False, indent=2)+'\n')
        requests.append({'key':meta['key'], 'sha256':meta['sha256'], 'request':str(request.relative_to(p))})
    # 所有請求完成才一次更新接入清單，避免半套圖組。
    expected = {s['key'] for s in sources['sheets']}
    expected.update('complete-'+s['key'] for s in sources['sources'] if s['kind'] != 'picture')
    if set(images) != expected:
        raise ValueError(f'圖組未齊：missing={expected-set(images)} extra={set(images)-expected}')
    picture_cells = {c['key']:(s['key'],c['box']) for s in sources['sheets'] for c in s['cells']}
    groups, entries, components = [], [], []
    fragments=set()
    def data(archive,record):
        return (p/f'workplace/full-hd-v1/decoded-v2/zelres{archive}-{record:02d}.decoded').read_bytes()
    def binding(segment,offset,raw):
        return {'segment':segment,'offset':offset,'data':base64.b64encode(raw).decode()}
    def crop(im,box,size,atlas):
        scaled=tuple(round(v*im.size[i%2]/atlas[i%2]) for i,v in enumerate(box))
        return im.crop(scaled).resize(size,Image.Resampling.LANCZOS)
    memory=(p/'workplace/full-hd-v1/cave-profile-normal-pair-v2/cave-platform/object-missing-00.memory').read_bytes()
    resource=memory[0x11b10:0x21b10]
    for source in sources['sources']:
        key=source['key']; binds=[]
        raw=data(source['archive'],source['record']) if source['kind']=='tile-atlas' or key in ('mman-all','cman-all','tman-all','cpat-all','mpat-all','dpat-all','inventory-icons','magic-effects') else b''
        if source['kind']=='tile-atlas':
            packed,_=tiles(raw);binds=[binding('r',0x8000,packed)]
        elif key in ('mman-all','cman-all','tman-all','cpat-all','mpat-all','dpat-all'):
            terrain='pat-' in key;start=0 if key=='tman-all' else 256;packed=[]
            for n in range(len(source['parts'])):
                chunk=raw[start+n*48:start+(n+1)*48]
                packed.append((terrain_preprocess(chunk,raw[6+n]) if terrain else preprocess(chunk,7,0))[0])
            offset=0x8100 if terrain else 0x6000 if key=='tman-all' else 0x4100
            binds=[binding('r',offset,b''.join(packed))]
        elif key=='fman-all':
            binds=[binding('r',0x6333,resource[0x6333:0x6333+230*32]),binding('r',0xd000,resource[0xd000:0xd000+230*8])]
        elif key=='inventory-icons':
            binds=[binding('r',0xe20e,raw[14:])]
        elif key=='magic-effects':
            binds=[binding('e',12,raw[12:])]
        group=len(groups);groups.append({'key':key,'bindings':binds,'source_sha256':source['source_sha256'],'archive':source['archive'],'record':source['record'],'method':source['method']})
        if source['kind']=='picture':
            sheet,box=picture_cells[key];im=crop(images[sheet],box,(source['width']*3,source['height']*3),(1536,1024))
            parts=[(source,im)]
        else:
            im=images['complete-'+key];parts=[]
            for n,part in enumerate(source['parts']):
                if source['kind']=='tile-atlas':
                    part={**part,'width':8,'height':8,'box':[n%16*12+2,n//16*12+2,n%16*12+10,n//16*12+10]}
                    atlas=(source['columns']*12,source['rows']*12)
                else:
                    atlas=source['atlas_size']
                painted=crop(im,part['box'],(part['width']*3,part['height']*3),atlas)
                # 原版透明包絡為唯一幾何；遮罩外不讓AI細節影響構圖。
                if 'mask' in part:
                    mask=Image.frombytes('L',(part['width'],part['height']),base64.b64decode(part['mask'])).resize(painted.size,Image.Resampling.NEAREST)
                    painted.putalpha(mask)
                else:
                    painted.putalpha(255)
                parts.append((part,painted))
        for n,(part,painted) in enumerate(parts):
            file=f'complete-{key}-{n:03d}.png';painted.save(out/file)
            entry={'group':group,'source':key,'width':part['width'],'height':part['height'],'raster':part['raster'],'file':file,'sha256':digest((out/file).read_bytes())}
            if 'mask' in part:entry['mask']=part['mask']
            entry['colors']=source.get('colors',base64.b64encode((p/'workplace/music-release-v1/release-build/AppDir/assets/town/frame/palette.bin').read_bytes()).decode())
            entries.append(entry)
            if key in ('fman-all','weapon-glyphs'):
                canonical=bytearray(base64.b64decode(part['raster']));mask=base64.b64decode(part['mask'])
                for i,v in enumerate(mask):
                    if not v:canonical[i]=255
                components.append({**entry,'raster':base64.b64encode(canonical).decode()})
                if key=='fman-all':
                    packed=resource[0x6333+n*32:0x6333+(n+1)*32]
                    indices=[(struct.unpack_from('<H',packed,y*4+x//4*2)[0]>>(12-x%4*4))&15 for y in range(8) for x in range(8)]
                    for variant in source['palette_variants'][1:]:
                        palette=base64.b64decode(variant)
                        canonical=bytes(palette[v] if mask[i] else 255 for i,v in enumerate(indices))
                        components.append({**entry,'raster':base64.b64encode(canonical).decode()})
            # 捲動、擦入及說話會局部改寫圖面。沿同張HD圖切出完整8×8來源，仍須逐像素匹配。
            if part['width']>32 or part['height']>32:
                raster=base64.b64decode(part['raster']);width,height=part['width'],part['height']
                for y in range(0,height,8):
                    for x in range(0,width,8):
                        fw,fh=min(8,width-x),min(8,height-y)
                        if fw<2 or fh<2:continue
                        native=b''.join(raster[(y+dy)*width+x:(y+dy)*width+x+fw] for dy in range(fh))
                        identity=(group,fw,fh,native)
                        if len(set(native))<2 or identity in fragments:continue
                        fragments.add(identity)
                        fragment=painted.crop((x*3,y*3,(x+fw)*3,(y+fh)*3));name=f'complete-{key}-{n:03d}-x{x:03d}-y{y:03d}.png';fragment.save(out/name)
                        entries.append({**entry,'width':fw,'height':fh,'raster':base64.b64encode(native).decode(),'file':name,'sha256':digest((out/name).read_bytes()),'derivative_of':file})
    input_files=[p/'workplace/orig/ZELIARD.EXE',p/'workplace/orig/ZELRES1.SAR',p/'workplace/orig/ZELRES2.SAR',p/'workplace/orig/ZELRES3.SAR',p/'workplace/full-hd-v1/cave-profile-normal-pair-v2/cave-platform/object-missing-00.memory']
    input_files+=sorted((p/'workplace/full-hd-v1/decoded-v2').glob('*.decoded'))
    input_files+=[p/'workplace/music-release-v1/release-build/AppDir/assets/town/frame/palette.bin',p/'workplace/probe-intro-display/palette.bin']
    input_files+=[p/'tools'/name for name in ('hd_complete_sources.py','hd_complete_register.py','hd_manual_sources.py','hd_grp_sources.py','hd_interior_sources.py','hd_inventory.py')]
    inputs=[{'file':str(file.relative_to(p)),'sha256':digest(file.read_bytes())} for file in input_files]
    manifest={'schema':1,'status':'INTEGRATED_MANUAL_REVIEW_PENDING','scale':3,'groups':groups,'entries':entries,'components':components,'requests':requests,'inputs':inputs,'source_level':'strong inference; exact generated-input identity, player-path review pending','environment':'eob-audio-capture:20261009-r3; built-in imagegen','game_launched':False,'tests_run':False}
    (out/'complete-art.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
    # 首領第二階段共用原本的完整16×16物件匹配入口。
    manifest=json.loads((out/'manual-art.json').read_text());objects=manifest['objects'];seen={r['key'] for r in objects}
    for entry in entries:
        if entry['source']!='zel2-poses':continue
        raw=base64.b64decode(entry['raster']);key=digest(raw)
        if key not in seen:objects.append({**entry,'key':key});seen.add(key)
    (out/'manual-art.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
    sources['status']='ART_INTEGRATED_MANUAL_REVIEW_PENDING'
    (w/'completion-sources.json').write_text(json.dumps(sources,ensure_ascii=False,indent=2)+'\n')
    print(f'REGISTERED {len(images)} sheets, {len(entries)} entries, {len(components)} actor/weapon parts; no game or tests executed')


if __name__=='__main__':main()

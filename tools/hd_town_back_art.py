#!/usr/bin/env python3
"""重建TMAN背向來源並註冊48×72手繪圖；不執行或改寫遊戲。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageDraw
from hd_mpp1_art import area_image


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def source(project):
    path=project/"workplace/music-release-v1/release-build/data/assets/town/restored.memory"
    assert sha(path)=="9dd3657107f25c097debf5f6be68137d055d98bda2e22a59f6b854d88be5fdd9"
    memory=path.read_bytes(); p=memory[0x1b10:0x11b10]; r=memory[0x11b10:0x21b10]; e=memory[0x21b10:0x31b10]
    palette=(project/"workplace/town-hd-v1/foreground-v1/normal-v1/restored.palette").read_bytes()
    assert hashlib.sha256(palette).hexdigest()=="6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c"
    frames=[]
    for bit,table in [(0,0x6a71),(1,0x6a53)]:
        ids=p[table:table+6]; assert list(ids)==[20,22,24,21,23,25]
        im=Image.new("RGBA",(16,24))
        for n,tile in enumerate(ids):
            packed=r[0x6000+tile*48:0x6000+(tile+1)*48]; pixels=[]
            for i in range(16):
                a,c,d=packed[i*3:i*3+3]
                pixels.extend([c>>2,((c&3)<<4)|(a>>4),((a&15)<<2)|(d>>6),d&63])
            for y in range(8):
                for x in range(8):
                    v=pixels[y*8+x]; keep=e[0x8000+tile*8+y]&(128>>x)
                    assert not keep or v==0
                    if not keep:
                        im.putpixel((n//3*8+x,n%3*8+y),tuple(palette[v*3:v*3+3])+(255,))
        digest=hashlib.sha256(im.tobytes()).hexdigest()
        assert digest=="70ca7bed8492e9f7c4a49a951ba602782c57cc94ba62d3b07ee4b1ddd5fbf2d8"
        frames.append((bit,im))
    return frames


def prepare(project,output):
    for bit,im in source(project):
        im.save(output/f"neutral-bit{bit}-original.png")
        im.resize((512,768),Image.Resampling.NEAREST).save(output/f"neutral-bit{bit}-reference.png")


def audit(project, output):
    """獨立重建兩條正常進門來源的14336像素，居民是否存在由地圖核對。"""
    cases=[]
    word=lambda b,i:int.from_bytes(b[i:i+2],"little")
    for folder, name, bit, identity in [
        ("town-neutral-normal-v3","observed-back-bit0",0,"50ba7bef235d3c9ac684feffd86b2b318a9e9f43b9b45f1bdeb8184d9e19be0d"),
        ("town-neutral-normal-v4","observed-back-bit1",1,"368b1ce2e672bfe5c48c88537401c3fbe5b3917c8a8216c60659cb5379ad9d00")]:
        base=project/"workplace/full-hd-v1"/folder
        memory=base/(name+".memory"); assert sha(memory)==identity
        data=memory.read_bytes(); p=data[0x1b10:0x11b10]; r=data[0x11b10:0x21b10]; e=data[0x21b10:0x31b10]
        raw=(base/(name+".indexed")).read_bytes(); pal=(base/(name+".palette")).read_bytes()
        assert p[0xe7]==4 and p[0xc2]&1==bit and p[0x83]<27 and len(raw)==64000
        actors=[]; cursor=word(p,0xc00f)
        while word(p,cursor)!=65535:
            x=word(p,cursor); under=p[cursor+3]; mark=p[0xc017+x*8+5]
            assert mark==under and mark<157
            actors.append({"record":hex(cursor),"x":x,"map":mark,"under":under})
            cursor+=8
        def tile(off):
            pixels=[]
            for i in range(16):
                a,c,d=r[off+i*3:off+i*3+3]
                pixels.extend([c>>2,((c&3)<<4)|(a>>4),((a&15)<<2)|(d>>6),d&63])
            return pixels
        decoded=bytearray(224*64)
        for col in range(28):
            for row in range(8):
                id=p[0xe000+col*8+row]
                if id>=253:id=p[word(p,0xff2a)+32+col*8+row]
                assert id<157
                pixels=tile(0x8100+id*48)
                for dy in range(8):
                    for dx in range(8):
                        x,y=col*8+dx,row*8+dy
                        bg=row<3 and r[0x8006+id]!=0 and r[0xd000+id*8+dy]&(128>>dx)
                        decoded[y*224+x]=p[0xa000+(x//8)*192+y*8+x%8] if bg else pixels[dy*8+dx]
        underlay=bytes(decoded)
        ids=p[0x6a71:0x6a77]; assert list(ids)==[20,22,24,21,23,25]
        for n,id in enumerate(ids):
            pixels=tile(0x6000+id*48)
            for y in range(8):
                for x in range(8):
                    v=pixels[y*8+x]; keep=e[0x8000+id*8+y]&(128>>x)
                    assert not keep or v==0
                    if not keep:decoded[(40+n%3*8+y)*224+p[0x83]*8+n//3*8+x]=v
        original=b"".join(raw[(78+y)*320+48:(78+y)*320+272] for y in range(64))
        assert decoded==original
        image=Image.frombytes("P",(224,64),bytes(decoded)); image.putpalette(pal)
        image.convert("RGB").resize((672,192),Image.Resampling.NEAREST).save(output/(name+"-reconstructed.png"))
        # 獨立核對直接3倍合成：原版底圖、世界圖、背向圖；中間alpha容許2階整數取捨。
        castle_path=project/"workplace/full-hd-v1/castle-registration-v1/castle-world-hd-v1.png"
        sprite_path=project/"workplace/full-hd-v1/town-neutral-complete-v1/town-player-back-v1.png"
        assert sha(castle_path)=="288ac2c28d625582a401343dedbc108ad8e50a7c78e1956cbe92b32ae9c63bb5"
        assert sha(sprite_path)=="d84525e69589decf0045cc32b35032eaea862c967e9876df5294cf208d3096f4"
        castle=Image.open(castle_path).convert("RGBA"); sprite=Image.open(sprite_path).convert("RGBA")
        rendered=project/"workplace/full-hd-v1/town-back-runtime-v2"/(name+"-hd.png")
        actual=Image.open(rendered).convert("RGB"); world=(word(p,0x80)+4+p[0x83])*24
        difference=0; checked=0
        def over(src, dst):
            *rgb,alpha=src
            return tuple(rgb[i]*alpha//255+dst[i]*(255-alpha)//255 for i in range(3))
        expected_image=Image.new("RGB",(48,72))
        for y in range(72):
            for x in range(48):
                v=underlay[(40+y//3)*224+p[0x83]*8+x//3]
                pixel=over(castle.getpixel((world+x,120+y)),tuple(pal[v*3:v*3+3]))
                pixel=over(sprite.getpixel((x,y)),pixel)
                found=actual.getpixel(((48+p[0x83]*8)*3+x,354+y))
                delta=max(abs(a-b) for a,b in zip(pixel,found)); assert delta<=2,(name,x,y,pixel,found)
                difference=max(difference,delta); checked+=1; expected_image.putpixel((x,y),pixel)
        expected_image.save(output/(name+"-independent-hd.png"))
        cases.append({"name":name,"memory_sha256":identity,"bit":bit,"phase":4,
                      "original_pixels":14336,"hidden_residents":actors,"indexed_sha256":sha(base/(name+".indexed")),
                      "independent_hd_pixels":checked,"max_channel_deviation":difference,
                      "hd_tolerance":2,"rendered_sha256":sha(rendered)})
    report={"schema":1,"status":"PASS_NORMAL_BACK_SOURCE_PIXELS","script_sha256":sha(Path(__file__)),"cases":cases}
    (output/"native-check.json").write_text(json.dumps(report,ensure_ascii=False,indent=2)+"\n",encoding="utf-8")
    print(report["status"],len(cases)*14336)


def register(project,output):
    frames=source(project)
    base=project/"workplace/full-hd-v1/town-neutral-art-v1"
    generation=base/"generation.json"; meta=json.loads(generation.read_text())
    assert meta["mode"]=="builtin-imagegen"
    image_path=base/meta["output"]["file"]; assert sha(image_path)==meta["output"]["sha256"]
    im=Image.open(image_path).convert("RGBA"); low,high=im.getchannel("A").getextrema()
    assert low==0 and high>=128
    crop=im.getchannel("A").point(lambda v:255 if v>=128 else 0).getbbox(); assert crop
    sprite=area_image(im.crop(crop),48,72)
    assert sprite.getchannel("A").getbbox() and sprite.getchannel("A").getextrema()[0]==0
    target=output/"town-player-back-v1.png"; sprite.save(target)
    contact=Image.new("RGB",(512,320),(64,76,88)); draw=ImageDraw.Draw(contact)
    native=frames[0][1].resize((160,240),Image.Resampling.NEAREST)
    shown=sprite.resize((160,240),Image.Resampling.NEAREST)
    contact.paste(native,(48,48),native); contact.paste(shown,(304,48),shown)
    draw.text((48,24),"Original phase 4",fill="white"); draw.text((304,24),"HD phase 4",fill="white")
    contact.save(output/"registered-contact.png")
    report={"schema":1,"status":"REGISTERED_TOWN_BACK_POSE_OFFLINE_ONLY","registered":target.name,"sha256":sha(target),"size":[48,72],"art_crop":crop,
            "source_rgba_sha256":"70ca7bed8492e9f7c4a49a951ba602782c57cc94ba62d3b07ee4b1ddd5fbf2d8","shared_slots":[4,9],"generation_sha256":sha(generation),"script_sha256":sha(Path(__file__)),"area_tool_sha256":sha(project/"tools/hd_mpp1_art.py"),
            "limits":["只提供固定背向圖面，不設定姿態或改變遊戲輸入。","兩朝向共用同一原始RGBA，其他場景及姿態仍依各自來源驗證。"]}
    (output/"manifest.json").write_text(json.dumps(report,ensure_ascii=False,indent=2)+"\n")
    print(report["status"],report["sha256"])


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project",type=Path,required=True)
    ap.add_argument("--output",type=Path,required=True)
    ap.add_argument("--mode",choices=("prepare","register","audit"),required=True)
    args=ap.parse_args(); assert args.output.is_dir() and not any(args.output.iterdir())
    {"prepare":prepare,"register":register,"audit":audit}[args.mode](args.project,args.output)


if __name__=="__main__":
    main()

#!/usr/bin/env python3
"""從已核SAR解壓資料重建城鎮圖格，對照正常原版RAM；完整圖片只留本機。"""
import argparse
import hashlib
import json
import struct
from pathlib import Path

from PIL import Image, ImageDraw

from hd_planar import convert

MURALLA = "workplace/music-release-v1/release-build/data/assets/town/restored.memory"
CASTLE = "workplace/town-hd-v1/foreground-v1/normal-v1/cold-town.memory"
MEMORY_HASHES = {MURALLA: "9dd3657107f25c097debf5f6be68137d055d98bda2e22a59f6b854d88be5fdd9",
                 CASTLE: "5c5b610c97bd2cd5266ae25b3da3764d0b46f02b30e8392a3b09c286a0db45a2"}


def preprocess(raw, clear_color, transparent_color):
    if len(raw) != 48:
        raise ValueError("8列、每列6bytes")
    cleaned, masks = bytearray(), bytearray()
    for row in range(8):
        words = list(struct.unpack_from(">HHH", raw, row*6))
        colors = [sum(((words[q] >> bit) & 1) << q for q in range(3)) for bit in range(15, -1, -1)]
        mask = 0
        for col in range(8):
            mask = mask*2 + int(colors[col*2] == transparent_color and colors[col*2+1] == transparent_color)
        masks.append(mask)
        if clear_color is not None:
            equal = 65535
            for q in range(3):
                equal &= words[q] if clear_color & (1 << q) else ~words[q] & 65535
            words = [word & ~equal for word in words]
        cleaned.extend(struct.pack(">HHH", *words))
    return convert(cleaned), bytes(masks)


def tile_image(data, mask, palette):
    colors = []
    for a, b, c in zip(data[::3], data[1::3], data[2::3]):
        colors.extend((b >> 2, (b & 3) << 4 | a >> 4, (a & 15) << 2 | c >> 6, c & 63))
    im = Image.new("RGBA", (8, 8))
    for y in range(8):
        for x in range(8):
            color = colors[y*8+x]
            im.putpixel((x, y), (*palette[color*3:color*3+3], 0 if mask[y] & (128 >> x) else 255))
    return im


def terrain_preprocess(raw, flag):
    key = 1 << (3-flag) if flag in (1, 2, 3) else None
    data, mask = preprocess(raw, key, key)
    # 兩個已驗正常profile的flag4圖格，原版D000遮罩均為八個FF。
    if flag == 4:
        mask = bytes([255])*8
    return data, mask


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    p, out = args.project, args.output
    if not out.is_dir() or any(out.iterdir()):
        ap.error("輸出必須是既有空目錄")
    sha = lambda b: hashlib.sha256(b).hexdigest()
    palette = (p/"workplace/music-release-v1/release-build/data/assets/town/frame/palette.bin").read_bytes()
    assert sha(palette) == "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c"
    ledger = json.loads((p/"docs/hd-coverage.json").read_text())
    sources = {x["record"]: x for x in ledger["records"] if x["archive"] == "ZELRES2.SAR"}
    memories = {name: (p/name).read_bytes() for name in MEMORY_HASHES}
    for name, memory in memories.items():
        assert len(memory) == 1048576 and sha(memory) == MEMORY_HASHES[name]
    profiles = [(29, "MMAN", "npc", 152, MURALLA, 0x4100, 0x7000),
                (30, "CMAN", "npc", 164, None, None, None),
                (31, "TMAN", "player", 46, MURALLA, 0x6000, 0x8000),
                (33, "CPAT", "terrain", 157, CASTLE, 0x8100, 0xd000),
                (34, "MPAT", "terrain", 242, MURALLA, 0x8100, 0xd000),
                (35, "DPAT", "terrain", 191, None, None, None)]
    reports = []
    for record, name, family, count, memory_path, target, mask_target in profiles:
        path = p/f"workplace/full-hd-v1/decoded-v2/zelres2-{record:02d}.decoded"
        raw = path.read_bytes()
        assert sha(raw) == sources[record]["decoder"]["sha256"]
        start = 0 if family == "player" else 256
        assert len(raw) >= start+count*48
        memory = memories.get(memory_path)
        r = memory[0x11b10:0x21b10] if memory else None
        e = memory[0x21b10:0x31b10] if memory and family != "terrain" else r
        if family == "terrain":
            pointers = struct.unpack_from("<HHH", raw)
            assert pointers[0] == 6 and pointers[1]-6 == count
            assert len(raw) == start+count*48 and max(raw[6:6+count]) <= 4
            if r:
                assert all(struct.unpack_from("<H", r, 0x8000+i*2)[0] == pointer+0x8000 for i, pointer in enumerate(pointers))
                assert r[0x8006:0x8100] == raw[6:256]
        elif family == "npc":
            assert max(raw[:240]) == count and min(raw[:240]) > 0
            if r:
                assert r[0x4000:0x4100] == raw[:256]
        tiles, pixels, masks = [], [], []
        checked_masks = 0
        for tile in range(count):
            flag = raw[6+tile] if family == "terrain" else None
            key = 1 << (3-flag) if flag in (1, 2, 3) else None
            clear, transparent = (key, key) if family == "terrain" else (7, 0)
            chunk = raw[start+tile*48:start+(tile+1)*48]
            data, mask = terrain_preprocess(chunk, flag) if family == "terrain" else preprocess(chunk, clear, transparent)
            if r:
                assert data == r[target+tile*48:target+(tile+1)*48], (name, "pixel", tile)
                if family != "terrain" or flag != 0:
                    assert mask == e[mask_target+tile*8:mask_target+(tile+1)*8], (name, "mask", tile)
                    checked_masks += 8
            pixels.append(data)
            masks.append(mask)
            tiles.append(tile_image(data, mask, palette))
        atlas = Image.new("RGB", (1280, ((count+15)//16)*80), "#26313a")
        label = ImageDraw.Draw(atlas)
        for i, tile in enumerate(tiles):
            x, y = i%16*80+8, i//16*80+14
            zoom = tile.resize((64, 64), Image.Resampling.NEAREST)
            atlas.paste(zoom, (x, y), zoom)
            label.text((x, y-12), str(i), fill="white")
        atlas.save(out/f"{name}-tiles.png")
        frame_records = []
        if family == "npc":
            for group in range(5):
                sheet = Image.new("RGBA", (1024, 768))
                for phase in range(8):
                    ids = raw[group*48+phase*6:group*48+phase*6+6]
                    frame = Image.new("RGBA", (16, 24))
                    for i, tile_id in enumerate(ids):
                        frame.alpha_composite(tiles[tile_id-1], (i//3*8, i%3*8))
                    frame.save(out/f"{name}-g{group}-p{phase}.png")
                    frame_records.append({"group": group, "phase": phase, "rgba_sha256": sha(frame.tobytes()), "tiles": list(ids)})
                    sheet.alpha_composite(frame.resize((192, 288), Image.Resampling.NEAREST), (phase%4*256+32, phase//4*384+48))
                sheet.save(out/f"{name}-g{group}-reference.png")
        report = {"record": record, "name": name, "family": family, "tiles": count,
                  "source": str(path.relative_to(p)), "source_sha256": sha(raw), "tile_offset": start,
                  "uninterpreted_tail_bytes": len(raw)-start-count*48,
                  "pixels_sha256": sha(b"".join(pixels)), "masks_sha256": sha(b"".join(masks)),
                  "normal_memory": memory_path, "normal_memory_sha256": MEMORY_HASHES.get(memory_path),
                  "runtime_bytes_verified": count*48 if r else 0, "runtime_mask_bytes_verified": checked_masks,
                  "status": "PASS_NORMAL_RUNTIME_SOURCES" if r else "STRUCTURAL_CANDIDATE_ONLY",
                  "address_space": f"SAR2解壓buffer +{start}; dosgolem11B1:{target:04X}，遮罩"+("21B1" if family != "terrain" else "11B1")+f":{mask_target:04X}" if r else "SAR2解壓buffer，未核正常載入位址",
                  "frames": frame_records}
        reports.append(report)
        print(name, report["status"], count)
    result = {"schema": 1, "status": "PASS_FOUR_NORMAL_RESOURCE_BINDINGS", "records": reports,
              "script_sha256": sha(Path(__file__).read_bytes()), "palette_sha256": sha(palette),
              "planar_oracle": "workplace/full-hd-v1/planar-oracle-v2/receipt.json",
              "level": "已證實：四資源的指定像素、mask及header與既有正常原版快照一致；CMAN／DPAT僅結構候選",
              "limits": ["不宣稱全部原始bytes用途，NPC尾端資料尚未解釋。", "圖格來源證據不等於全地圖可達、HD圖面生成或正式前端接入。"]}
    (out/"receipt.json").write_text(json.dumps(result, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")


if __name__ == "__main__":
    main()

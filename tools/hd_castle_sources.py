#!/usr/bin/env python3
"""沿既有城鎮地圖契約擷取冷新局城堡前景，供本地HD製圖與來源審查。"""
import argparse
import hashlib
import json
import struct
from pathlib import Path

from PIL import Image

from hd_grp_sources import CASTLE, MEMORY_HASHES, terrain_preprocess, tile_image


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    p, out = args.project, args.output
    if not out.is_dir() or any(out.iterdir()):
        ap.error("輸出必須是既有空目錄")
    sha = lambda b: hashlib.sha256(b).hexdigest()
    path = p/CASTLE
    memory = path.read_bytes()
    assert sha(memory) == MEMORY_HASHES[CASTLE]
    primary, resources = memory[0x1b10:0x11b10], memory[0x11b10:0x21b10]
    word = lambda pos: struct.unpack_from("<H", primary, pos)[0]
    assert word(0xc002) == 114 and word(0xff2c) == 0x11b1
    width, camera = word(0xc002), word(0x80)
    assert word(0xff2a)-camera*8 == 0xc017
    palette = (p/"workplace/town-hd-v1/foreground-v1/normal-v1/cold-town.palette").read_bytes()
    assert sha(palette) == "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c"
    raw = (p/"workplace/full-hd-v1/decoded-v2/zelres2-33.decoded").read_bytes()
    assert sha(raw) == "831bf7b9a3cb0d31870d57d23f06e69a56d420fe17f8c27dc0c2323ddbce1339"
    actors, actor_records, pos = {}, [], word(0xc00f)
    while word(pos) != 65535:
        assert 0xc017+width*8 <= pos < 0xffff-8
        x = word(pos)
        assert x not in actors and x < width
        actors[x] = primary[pos+3]
        actor_records.append((x, primary[pos+2], primary[pos+4] & 3))
        pos += 8
    tiles = []
    for tile in range(157):
        flag = raw[6+tile]
        data, mask = terrain_preprocess(raw[256+tile*48:256+(tile+1)*48], flag)
        assert data == resources[0x8100+tile*48:0x8130+tile*48]
        tiles.append((tile_image(data, mask, palette), flag))
    world = Image.new("RGBA", (width*8, 64))
    ids = []
    for x in range(width):
        for y in range(8):
            tile = primary[0xc017+x*8+y]
            if tile == 253:
                assert x in actors
                tile = actors[x]
            assert tile < len(tiles), (x, y, tile)
            ids.append(tile)
            im, flag = tiles[tile]
            for dy in range(8):
                for dx in range(8):
                    red, green, blue, alpha = im.getpixel((dx, dy))
                    if y >= 3 or not flag:
                        alpha = 255
                    world.putpixel((x*8+dx, y*8+dy), (red, green, blue, alpha))
    indexed = (path.with_suffix(".indexed")).read_bytes()
    # 原版E000是已畫來源，與地圖候選不同時記錄差異，不以靜態全圖替代動畫。
    window = Image.new("RGBA", (224, 64))
    cache_differences = []
    for col in range(28):
        for row in range(8):
            canonical = ids[(camera+4+col)*8+row]
            tile = primary[0xe000+col*8+row]
            if tile >= 253:
                tile = canonical
            assert tile < len(tiles)
            if tile != canonical:
                cache_differences.append({"world_column": camera+4+col, "row": row, "map_tile": canonical, "drawn_tile": tile})
            im, flag = tiles[tile]
            for dy in range(8):
                for dx in range(8):
                    x, y = col*8+dx, row*8+dy
                    color = im.getpixel((dx, dy))
                    if row < 3 and flag and not color[3]:
                        bg = primary[0xa000+col*192+y*8+dx]
                        color = (*palette[bg*3:bg*3+3], 255)
                    window.putpixel((x, y), (*color[:3], 255))
    extra = memory[0x21b10:0x31b10]

    def sprite(x, tile_ids, atlas, masks):
        for col in range(2):
            for row in range(3):
                tile = tile_ids[col*3+row]
                im = tile_image(resources[atlas+tile*48:atlas+(tile+1)*48], extra[masks+tile*8:masks+(tile+1)*8], palette)
                window.alpha_composite(im, (x+col*8, 40+row*8))

    for x, group_direction, phase in actor_records:
        px = (x-camera-4)*8
        if px <= -16 or px >= 224:
            continue
        group = group_direction & 127
        assert group < 5
        if not group_direction & 128:
            phase += 4
        pos = 0x4000+group*48+phase*6
        sprite(px, [v-1 for v in resources[pos:pos+6]], 0x4100, 0x7000)
    phase = primary[0xe7]
    assert phase < 5
    pos = (0x6a3b if primary[0xc2] & 1 else 0x6a59)+phase*6
    sprite(primary[0x83]*8, primary[pos:pos+6], 0x6000, 0x8000)
    for y in range(64):
        for x in range(224):
            color = indexed[(78+y)*320+48+x]
            assert bytes(window.getpixel((x, y))[:3]) == palette[color*3:color*3+3], (x, y)
    verified = 224*64
    window.save(out/"castle-normal-window.png")
    world.save(out/"castle-world-native.png")
    world.resize((2736, 192), Image.Resampling.NEAREST).save(out/"castle-world-reference.png")
    sheet = Image.new("RGBA", (960, 1024))
    panels = []
    for n, start in enumerate((0, 224, 448, 672)):
        panel = world.crop((start, 0, start+240, 64)).resize((960, 256), Image.Resampling.NEAREST)
        panel.save(out/f"castle-panel-{n}-reference.png")
        sheet.alpha_composite(panel, (0, n*256))
        panels.append({"index": n, "world_native": [start, 0, start+240, 64], "source_rgba_sha256": sha(panel.tobytes())})
    sheet.save(out/"castle-panels-reference.png")
    report = {"schema": 1, "status": "PASS_CASTLE_FOREGROUND_SOURCE_SCOPED", "memory": CASTLE,
              "memory_sha256": sha(memory), "palette_sha256": sha(palette), "map_address_space": "dosgolem01B1:C017，114欄各8圖格；CPAT圖格11B1:8100",
              "map_ids_sha256": sha(bytes(ids)), "world_rgba_sha256": sha(world.tobytes()), "world_native_size": [912, 64],
              "normal_camera": camera, "normal_foreground_pixels_verified": verified, "cache_map_differences": cache_differences, "panels": panels,
              "script_sha256": sha(Path(__file__).read_bytes()),
              "limits": "全圖為地圖當次靜態候選；正常viewport按原版快取選圖並驗全畫帶。快取與地圖差異原因未證實，正式HD仍須保留相應變化；未宣稱全圖已走過、HD已生成或正式接入"}
    (out/"receipt.json").write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
    print(report["status"], verified, "pixels")


if __name__ == "__main__":
    main()

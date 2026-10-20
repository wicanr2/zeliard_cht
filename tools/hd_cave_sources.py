#!/usr/bin/env python3
"""Docker內擷取11組MPP原版地形；來源證據不等於HD或正常可達驗收。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageDraw

from hd_grp_sources import tile_image
from hd_planar import convert


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    p, out = args.project, args.output
    assert out.is_dir() and not any(out.iterdir()), "輸出須為既有空目錄"
    sha = lambda b: hashlib.sha256(b).hexdigest()
    ledger = json.loads((p/"docs/hd-coverage.json").read_text())
    rows = {x["record"]: x for x in ledger["records"] if x["archive"] == "ZELRES3.SAR"}
    palette_path = p/"workplace/prose-v1/progression-v1/dungeon-entry.palette"
    palette = palette_path.read_bytes()
    assert sha(palette) == "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c"
    memory_path = p/"workplace/prose-v1/progression-v1/dungeon-entry.memory"
    memory = memory_path.read_bytes()
    assert len(memory) == 1 << 20
    native = memory[0x11b10:0x21b10]
    profiles = []
    overview = Image.new("RGB", (1200, 2080), "#18232d")
    for record in range(74, 85):
        name = "MPP"+"123456789AB"[record-74]
        path = p/f"workplace/full-hd-v1/decoded-v2/zelres3-{record:02d}.decoded"
        raw = path.read_bytes()
        assert sha(raw) == rows[record]["decoder"]["sha256"]
        assert len(raw) >= 96 and len(raw) % 48 == 0
        header, data = raw[:48], convert(raw[48:])
        count = len(data)//48
        assert max(header) <= count, (name, "header參照超過圖格數")
        target = out/name
        target.mkdir()
        sheet = Image.new("RGB", (384, ((count+7)//8)*60+24), "#26313a")
        label = ImageDraw.Draw(sheet)
        label.text((8, 4), f"{name}: {count} original 8x8 tiles", fill="white")
        tile_records = []
        for tile in range(count):
            block = data[tile*48:tile*48+48]
            image = tile_image(block, bytes(8), palette)
            image.save(target/f"tile-{tile+1:02d}.png")
            x, y = tile%8*48, tile//8*60+24
            sheet.paste(image.resize((48, 48), Image.Resampling.NEAREST), (x, y+12))
            label.text((x+2, y), str(tile+1), fill="white")
            tile_records.append({"id": tile+1, "decoded_offset": 48+tile*48,
                                 "source_sha256": sha(raw[48+tile*48:96+tile*48]),
                                 "packed_sha256": sha(block), "rgba_sha256": sha(image.tobytes())})
        sheet.save(target/"original-atlas.png")
        overview.paste(sheet, ((record-74)%3*400, (record-74)//3*520))
        matched = record == 74
        if matched:
            assert native[0x8000:0x8030] == header
            assert native[0x8030:0x8030+len(data)] == data
        profiles.append({"key": name, "record": record, "tiles": count,
                         "source": str(path.relative_to(p)), "source_sha256": sha(raw),
                         "header_sha256": sha(header), "packed_sha256": sha(data),
                         "header_values": list(header), "tile_records": tile_records,
                         "normal_memory": str(memory_path.relative_to(p)) if matched else None,
                         "runtime_bytes_verified": len(raw) if matched else 0,
                         "status": "PASS_NORMAL_RESOURCE_BINDING" if matched else "STRUCTURAL_SOURCE_CANDIDATE",
                         "normal_path": "既有SAVEM1冷讀後正常入第一洞窟" if matched else "未驗正常載入與可達性"})
        print(name, count, profiles[-1]["status"])
    overview.save(out/"original-terrain-overview.png")
    report = {"schema": 1, "status": "PASS_MPP_SOURCE_EXTRACTION_SCOPED",
              "profiles": profiles, "total_tiles": sum(x["tiles"] for x in profiles),
              "script_sha256": sha(Path(__file__).read_bytes()),
              "tools": {"python": __import__("platform").python_version(), "pillow": Image.__version__},
              "normal_memory_sha256": sha(memory), "palette_sha256": sha(palette),
              "address_space": "SAR3解壓buffer前48bytes為header；圖格offset48+(id-1)*48；MPP1對應dosgolem11B1:8030",
              "normal_evidence": "workplace/prose-v1/progression-v1/receipt.json",
              "limits": ["只有MPP1完整來源與正常RAM核對，其他10組只驗已證實的解壓／變換及結構候選。",
                         "378圖格尚不是全部美術分母；未證明各header欄位用途、所有場景或HD覆蓋。",
                         "原圖及完整產物只留本機；本工具不改正式前端、原版、輸入或存檔。"]}
    (out/"receipt.json").write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")


if __name__ == "__main__":
    main()

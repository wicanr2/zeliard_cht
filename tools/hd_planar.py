#!/usr/bin/env python3
"""固定GRP候選的三平面轉換研究；輸出不等於已解出的圖像或HD素材。"""
import argparse
import hashlib
import json
from pathlib import Path


def convert(data):
    if len(data) % 6:
        raise ValueError("每列必須六bytes")
    out = bytearray()
    for i in range(0, len(data), 6):
        words = [int.from_bytes(data[i+j:i+j+2], "big") for j in (0, 2, 4)]
        for start in (15, 7):
            value = 0
            for bit in range(start, start-8, -1):
                for word in reversed(words):
                    value = value*2 + ((word >> bit) & 1)
            out.extend((value >> 8).to_bytes(2, "little"))
            out.append(value & 255)
    return bytes(out)


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    if not args.output.is_dir() or any(args.output.iterdir()):
        ap.error("輸出必須是既有空目錄")
    p, out = args.project, args.output
    sha = lambda b: hashlib.sha256(b).hexdigest()
    ledger = json.loads((p/"docs/hd-coverage.json").read_text())
    chunks = json.loads((p/"workplace/full-hd-v1/decoded-v2/decoded-inventory.json").read_text())
    decoded = {(x["archive"], x["index"]): x for x in chunks["records"]}
    rows = []
    for item in ledger["records"]:
        key = item["archive"], item["record"]
        if item["classification"] != "graphics_candidate" or key not in decoded:
            continue
        source = p/"workplace/full-hd-v1/decoded-v2"/decoded[key]["decoded_path"]
        raw = source.read_bytes()
        assert sha(raw) == item["decoder"]["sha256"] and 0 < len(raw) < 0xa000
        padded = raw + bytes(-len(raw) % 48)
        result = convert(padded)
        filename = source.stem+".planar"
        (out/filename).write_bytes(result)
        rows.append({"archive": key[0], "record": key[1], "names": [n["name"] for n in item["names"]],
                     "source": str(source.relative_to(p)), "source_sha256": sha(raw), "source_bytes": len(raw),
                     "padded_bytes": len(padded), "output": filename, "sha256": sha(result)})
    report = {"schema": 1, "status": "CANDIDATE_PLANAR_TRANSFORM", "records": rows,
              "script_sha256": sha(Path(__file__).read_bytes()),
              "address_space": "IDA base0／dosgolem01B1:2C2A..2CBC，輸出為研究buffer偏移",
              "limits": "對整筆解壓資料作候選變換，含非圖像header及補零；尚未認定GRP圖像起點、色盤、遮罩或可見尺寸"}
    (out/"manifest.json").write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
    print("CANDIDATE_PLANAR_TRANSFORM", len(rows))


if __name__ == "__main__":
    main()

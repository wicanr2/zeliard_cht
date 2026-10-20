#!/usr/bin/env python3
"""Docker 研究工具：依 IDA 9.4 常駐 DAD..F16 的八種 RLE 分支解開 SAR 候選。"""
import argparse
import hashlib
import json
from pathlib import Path

from hd_inventory import EXPECTED, records


def decode(stream, limit=65536):
    if not stream:
        raise ValueError("缺少RLE類型")
    kind, data = stream[0] & 7, stream[1:]
    if kind == 0:
        if len(data) > limit:
            raise ValueError("解碼超限")
        return data
    table, pos = [], 0
    if kind in (1, 3, 6):
        while True:
            if pos >= len(data):
                raise ValueError("缺少字典結尾")
            if kind in (1, 3) and data[pos] == 255:
                pos += 1
                break
            if pos+1 >= len(data):
                raise ValueError("不完整字典項目")
            a, b = data[pos:pos+2]
            pos += 2
            if kind == 6 and a == b == 255:
                break
            table.append((a, b))
    marker = None
    if kind in (2, 4, 7):
        if pos >= len(data):
            raise ValueError("缺少RLE標記")
        marker = data[pos]
        pos += 1
    out = bytearray()

    def take():
        nonlocal pos
        if pos >= len(data):
            raise ValueError("RLE項目越界")
        value = data[pos]
        pos += 1
        return value

    while pos < len(data):
        value, count = take(), 1
        if kind in (1, 3):
            mask, literal_mask = (240, 15) if kind == 1 else (15, 240)
            for token, replacement in table:
                if token & literal_mask:
                    break
                if value & mask == token:
                    count = (value & 15)+2 if kind == 1 else (value >> 4)+2
                    value = replacement
                    break
        elif kind in (2, 4):
            token = value & (240 if kind == 2 else 15)
            if token == marker:
                count = (value & 15)+3 if kind == 2 else (value >> 4)+3
                value = take()
        elif kind == 5:
            if pos < len(data) and data[pos] == value:
                take()
                count = take()+2
        elif kind == 6:
            for token, replacement in table:
                if value == token:
                    count = take()+2
                    value = replacement
                    break
        elif kind == 7 and value == marker:
            value = take()
            count = take()+3
        if len(out)+count > limit:
            raise ValueError("解碼超限")
        out.extend(bytes([value])*count)
    return bytes(out)


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--original", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    if not args.original.is_dir() or not args.output.is_dir() or any(args.output.iterdir()):
        ap.error("來源須存在，輸出須為既有空目錄")
    result = []
    for filename, (_, expected) in EXPECTED.items():
        source = (args.original/filename).read_bytes()
        assert hashlib.sha256(source).hexdigest() == expected
        for entry in records(source):
            start = entry["payload_offset"]
            payload = source[start:start+entry["bytes"]]
            if not payload or payload[0] != 0:
                continue
            item = {"archive": filename, **entry, "rle_kind": payload[1] & 7,
                    "level": "強推論：依固定IDA指令解碼，仍須原版消費端比對"}
            try:
                raw = decode(payload[1:])
                name = filename.lower().replace(".sar", "")+f"-{entry['index']:02d}.decoded"
                (args.output/name).write_bytes(raw)
                item.update({"decoded_path": name, "decoded_bytes": len(raw),
                             "decoded_sha256": hashlib.sha256(raw).hexdigest(), "result": "DECODED_CANDIDATE"})
            except ValueError as e:
                item.update({"result": "REJECTED", "reason": str(e)})
            result.append(item)
    report = {"schema": 1, "tool": "Python 標準函式庫", "script_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
              "evidence": "ida-resident-v2/resident.json，IDA9.4，固定常駐01B1:DAD..F16及B1D..B6C",
              "limits": ["只解壓縮，不推斷圖像尺寸、色盤、語意或HD完成", "已解碼仍須原版載入結果或同狀態畫面比對"],
              "records": result}
    (args.output/"decoded-inventory.json").write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
    print(json.dumps({"candidates": len(result), "decoded": sum(x["result"] == "DECODED_CANDIDATE" for x in result),
                      "rejected": [x for x in result if x["result"] == "REJECTED"]}, ensure_ascii=False))


if __name__ == "__main__":
    main()

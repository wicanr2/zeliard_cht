#!/usr/bin/env python3
"""Docker內解出31份洞窟地圖，保留原始圖格與尚未繪製的特殊值。"""
import argparse
import hashlib
import json
import struct
from collections import Counter
from pathlib import Path

from PIL import Image, ImageDraw


def decode_columns(data):
    width = struct.unpack_from("<H", data, 2)[0]
    stop = struct.unpack_from("<H", data, 25)[0]-0xc000
    if not 0 < width <= 1024 or not 27 <= stop <= len(data):
        raise ValueError("地圖長度或寬度越界")
    pos, grid, cases = 27, [], Counter()
    for col in range(width):
        column = []
        while len(column) < 64:
            if pos >= stop:
                raise ValueError("地圖壓縮輸入不足")
            value = data[pos]
            pos += 1
            kind = value >> 6
            cases[kind] += 1
            if kind == 0:
                if pos >= stop:
                    raise ValueError("雙byte run缺少圖格")
                count, tile = value+1, data[pos]
                pos += 1
            elif kind == 1:
                count, tile = ((value >> 4) & 3)+2, (value & 15)+1
            elif kind == 2:
                count, tile = value & 63, 0
            else:
                count, tile = 1, value & 63
            if count == 0 or len(column)+count > 64:
                raise ValueError("run跨越64列或沒有前進")
            column.extend([tile]*count)
        grid.append(bytes(column))
    if pos != stop:
        raise ValueError("地圖寬度解完但未抵達原始C019邊界")
    return grid, dict(cases), stop


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--project", type=Path, required=True)
    ap.add_argument("--sources", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    p, out = args.project, args.output
    assert out.is_dir() and not any(out.iterdir()), "輸出須為既有空目錄"
    sha = lambda b: hashlib.sha256(b).hexdigest()
    ledger = json.loads((p/"docs/hd-coverage.json").read_text())
    archive_path = p/"workplace/orig/ZELRES3.SAR"
    archive = archive_path.read_bytes()
    assert sha(archive) == next(x["sha256"] for x in ledger["archives"] if x["name"] == "ZELRES3.SAR")
    source_report = json.loads((args.sources/"receipt.json").read_text())
    profiles = source_report["profiles"]
    proof, cases_total = [], Counter()
    for row in ledger["records"]:
        if row["archive"] != "ZELRES3.SAR" or row["classification"] != "map_candidate":
            continue
        data = archive[row["payload_offset"]:row["payload_offset"]+row["payload_bytes"]]
        assert sha(data) == row["payload_sha256"]
        columns, cases, stop = decode_columns(data)
        cases_total.update(cases)
        end = struct.unpack_from("<H", data)[0]-0xc000
        assert 0 <= end <= len(data)-3
        selector = data[end+2]
        assert selector < len(profiles)
        profile = profiles[selector]
        name = sorted(set(x["name"] for x in row["names"]))
        assert len(name) == 1
        key = name[0].removesuffix(".MDT")
        raw_path = key+".source"
        grid_path = key+".columns"
        (out/raw_path).write_bytes(data)
        (out/grid_path).write_bytes(b"".join(columns))
        tiles = {i: Image.open(args.sources/profile["key"]/f"tile-{i:02d}.png").convert("RGB") for i in range(1, profile["tiles"]+1)}
        picture = Image.new("RGB", (len(columns)*8, 512))
        unresolved = Counter()
        for x, column in enumerate(columns):
            for y, tile in enumerate(column):
                if tile in tiles:
                    picture.paste(tiles[tile], (x*8, y*8))
                elif tile != 0:
                    unresolved[tile] += 1
                    # 洋紅格只標未分類資料，不能當成原版圖像或遊戲截圖。
                    ImageDraw.Draw(picture).rectangle((x*8, y*8, x*8+7, y*8+7), fill="#7d286e")
        picture.save(out/(key+"-static-research.png"))
        proof.append({"record": row["record"], "name": name[0], "payload_sha256": sha(data),
                      "source": raw_path, "columns": grid_path, "columns_sha256": sha(b"".join(columns)),
                      "width": len(columns), "height": 64, "compressed_start": 27, "compressed_end": stop,
                      "selector_offset": end+2, "selector_raw": selector,
                      "family_candidate": profile["key"], "unrendered_values": dict(unresolved), "cases": cases,
                      "level": "結構候選；須核原版常式、選擇端與正常來源"})
        print(key, len(columns), profile["key"], "unrendered", sum(unresolved.values()))
    report = {"schema": 1, "status": "DECODED_CAVE_MAP_CANDIDATES", "records": proof,
              "total_cells": sum(x["width"]*64 for x in proof), "case_counts": dict(cases_total),
              "archive_sha256": sha(archive), "script_sha256": sha(Path(__file__).read_bytes()),
              "source_receipt_sha256": sha((args.sources/"receipt.json").read_bytes()),
              "address_space": "SAR3各record的payload從0起算；載入C000後，C01B為壓縮起點、C019為結束指標",
              "limits": ["圖為離線靜態研究，洋紅格是未渲染的特殊值，不是正常原版畫面。",
                         "只讀原始資料，不解析敵人規則、不證明可達性、不計HD覆蓋。"]}
    (out/"receipt.json").write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")


if __name__ == "__main__":
    main()

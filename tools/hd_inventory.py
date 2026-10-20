#!/usr/bin/env python3
"""在 Docker 內盤點完整 SAR 記錄；形狀分類不代表圖像已解碼或 HD 完成。"""
import argparse
import hashlib
import json
import struct
from pathlib import Path

EXPECTED = {
    "ZELRES1.SAR": (40, "56a73b59e952dab47fb0171a3fb1cdb7d641baceafa97e2826057f877a590eb0"),
    "ZELRES2.SAR": (58, "bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1"),
    "ZELRES3.SAR": (96, "9c19b9c317d3006e11f8a5cad8ea1f226f8043e11c6467fa27f10eef33175ef0"),
}


def sha(data):
    return hashlib.sha256(data).hexdigest()


def records(data):
    if len(data) < 8:
        raise ValueError("SAR 太短")
    first, = struct.unpack_from("<I", data)
    if first < 4 or first % 4 or first > len(data)-4:
        raise ValueError("SAR 目錄邊界錯誤")
    offsets = struct.unpack_from("<"+"I"*(first//4), data)
    result = []
    for i, offset in enumerate(offsets):
        if offset < first or offset > len(data)-4:
            raise ValueError("SAR 記錄位址越界")
        size, = struct.unpack_from("<I", data, offset)
        end = offset+4+size
        expected = offsets[i+1] if i+1 < len(offsets) else len(data)
        if end != expected or end > len(data):
            raise ValueError("SAR 記錄不連續或長度越界")
        payload = data[offset+4:end]
        result.append({"index": i, "record_offset": offset, "payload_offset": offset+4,
                       "bytes": size, "sha256": sha(payload),
                       "classification": "unknown", "hd_status": "not_assessed"})
    return result


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--original", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    args = ap.parse_args()
    if not args.original.is_dir() or not args.output.is_dir():
        ap.error("輸入與輸出目錄必須已存在")
    destination = args.output/"sar-inventory.json"
    if destination.exists():
        ap.error("拒絕覆寫既有收據")
    archives = []
    for filename, (count, expected) in EXPECTED.items():
        data = (args.original/filename).read_bytes()
        if sha(data) != expected:
            raise ValueError("原版輸入身份不符："+filename)
        entries = records(data)
        if len(entries) != count:
            raise ValueError("記錄數不符："+filename)
        archives.append({"name": filename, "bytes": len(data), "sha256": expected,
                         "count": count, "records": entries})
    report = {"schema": 1, "tool": "Python 標準函式庫", "script_sha256": sha(Path(__file__).read_bytes()),
              "address_space": "各 SAR 檔案從 0 起算的位元組偏移，非執行期位址",
              "level": "已證實：固定原版目錄、長度與每筆雜湊",
              "limits": ["記錄可能含程式、地圖、文字、圖像或其他資料，不把記錄數當美術總數",
                         "分類、圖像解碼、HD接入與正常路徑驗證須各有證據，盤點不算完成"],
              "archives": archives}
    destination.write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
    print(json.dumps({"status": "PASS_BOUNDARIES_ONLY", "archives": len(archives),
                      "records": sum(x["count"] for x in archives)}, ensure_ascii=False))


if __name__ == "__main__":
    main()

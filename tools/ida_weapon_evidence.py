"""IDA 9.4 匯出正常武器店的名稱準備與搬移鏈，保留原始定位及分級索引。"""
import hashlib
import json
import os
import pathlib
import struct
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_lines
import ida_nalt
import ida_pro
import ida_segment
import ida_ua
import idautils
import idc


def main():
    source = pathlib.Path(ida_nalt.get_input_file_path())
    raw = source.read_bytes()
    index_path = pathlib.Path(__file__).resolve().parent.parent / "docs/weapon-source-index.json"
    index = json.loads(index_path.read_text(encoding="utf-8"))
    assert index["schema"] == 1 and len(raw) == 65536
    assert hashlib.sha256(raw).hexdigest() == index["input_sha256"]
    annotations = {int(e["address"], 16): e for e in index["entries"]}
    assert len(annotations) == len(index["entries"])

    def annotation(address):
        entry = annotations.get(address, {"level": "未知", "semantic": None, "evidence": []})
        assert entry["level"] in ("已證實", "強推論", "假說", "未知")
        return {**entry, "warning": entry["level"] != "已證實"}

    ida_ida.inf_set_app_bitness(16)
    for address in idautils.Segments():
        assert ida_segment.set_segm_addressing(ida_segment.getseg(address), 0)
    offsets = [0xA000, 0xA002, 0x2000] + list(range(0x6000, 0x601E, 2)) + [0x301A, 0x301C]
    seeds = [struct.unpack_from("<H", raw, off)[0] for off in offsets]
    for address in seeds:
        ida_ua.create_insn(address)
        ida_funcs.add_func(address)
    ida_auto.auto_wait()
    functions = []
    for start in idautils.Functions():
        if start not in (0x2046, 0x20B6, 0x20E8, 0xA10E, 0x7539, 0x3805, 0x383E, 0x3854, 0x37CC):
            continue
        function = ida_funcs.get_func(start)
        instructions = [{"address": hex(a),
                         "bytes": (ida_bytes.get_bytes(a, idc.get_item_size(a)) or b"").hex(),
                         "operand": ida_lines.tag_remove(idc.generate_disasm_line(a, 0) or ""),
                         "semantic_index": annotation(a)} for a in idautils.FuncItems(start)]
        functions.append({"original_name": idc.get_func_name(start), "start": hex(start),
                          "end": hex(function.end_ea), "semantic_index": annotation(start),
                          "instructions": instructions})
    assert {f["start"] for f in functions} == {"0x2046", "0x20b6", "0x20e8", "0xa10e", "0x7539", "0x3805", "0x383e", "0x3854", "0x37cc"}
    result = {"schema": 1, "tool": "IDA " + ida_kernwin.get_kernel_version(), "bits": 16,
              "input_name": source.name, "input_sha256": hashlib.sha256(raw).hexdigest(),
              "source_index_sha256": hashlib.sha256(index_path.read_bytes()).hexdigest(),
              "address_space": index["address_space"], "functions": functions,
              "limitation": "函式名稱只供導覽；未證實語意附 warning，不從推測性名稱或反編譯變數推論"}
    pathlib.Path(os.environ["ZELIARD_WEAPON_IDA_OUTPUT"]).write_text(
        json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    ida_pro.qexit(0)


try:
    main()
except Exception:
    pathlib.Path(os.environ["ZELIARD_WEAPON_IDA_OUTPUT"] + ".error").write_text(
        traceback.format_exc(), encoding="utf-8")
    ida_pro.qexit(1)

"""IDA 9.4 從國王正常初訪段傾印匯出原始定位與分級索引，不修改正式資料庫。"""
import hashlib
import json
import os
import pathlib

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
    index_path = pathlib.Path(__file__).resolve().parent.parent / "docs/dialogue-source-index.json"
    index = json.loads(index_path.read_text(encoding="utf-8"))
    assert len(raw) == 65536 and index["schema"] == 1
    assert hashlib.sha256(raw).hexdigest() == index["input_sha256"]
    annotations = {int(e["address"], 16): e for e in index["entries"]}
    assert len(annotations) == len(index["entries"])

    def annotation(address):
        entry = annotations.get(address, {"level":"未知", "semantic":None, "evidence":[]})
        assert entry["level"] in ("已證實", "強推論", "假說", "未知")
        return {**entry, "warning":entry["level"] != "已證實"}

    ida_ida.inf_set_app_bitness(16)
    for address in idautils.Segments():
        assert ida_segment.set_segm_addressing(ida_segment.getseg(address), 0)
    assert not ida_ida.inf_is_32bit_exactly() and not ida_ida.inf_is_64bit()
    entries = [int.from_bytes(raw[off:off+2], "little") for off in (0xA000, 0xA002, 0x6004, 0x2010, 0x2022)]
    assert entries == [0xA004, 0xA302, 0x706C, 0x22CD, 0x27E9]
    for address in entries:
        ida_ua.create_insn(address)
        ida_funcs.add_func(address)
    ida_auto.auto_wait()
    functions = []
    for start in idautils.Functions():
        if not (start == 0x27E9 or 0x7000 <= start < 0x7300 or 0xA004 <= start < 0xA40F):
            continue
        function = ida_funcs.get_func(start)
        instructions = [{"address":hex(a), "bytes":(ida_bytes.get_bytes(a, idc.get_item_size(a)) or b"").hex(),
                         "operand":ida_lines.tag_remove(idc.generate_disasm_line(a, 0) or ""),
                         "semantic_index":annotation(a)} for a in idautils.FuncItems(start)]
        functions.append({"original_name":idc.get_func_name(start), "start":hex(start), "end":hex(function.end_ea),
                          "semantic_index":annotation(start), "instructions":instructions})
    result = {"schema":1, "tool":"IDA " + ida_kernwin.get_kernel_version(), "input_name":source.name,
              "input_sha256":hashlib.sha256(raw).hexdigest(), "source_index_sha256":hashlib.sha256(index_path.read_bytes()).hexdigest(),
              "bits":16, "address_space":index["address_space"], "functions":functions,
              "limitation":"函式名稱僅供導覽；未證實語意帶 warning，不從反編譯器變數名推論"}
    pathlib.Path(os.environ["ZELIARD_IDA_OUTPUT"]).write_text(json.dumps(result, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
    ida_pro.qexit(0)


try:
    main()
except Exception:
    import traceback
    pathlib.Path(os.environ["ZELIARD_IDA_OUTPUT"]+".error").write_text(traceback.format_exc(), encoding="utf-8")
    ida_pro.qexit(1)

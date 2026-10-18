"""由 IDA 9.4 匯出片頭記憶體段的原始定位；只供研究，不替換函式名稱。"""

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
import ida_xref
import idautils
import idc


def main():
    source = pathlib.Path(ida_nalt.get_input_file_path())
    data = source.read_bytes()
    assert len(data) == 65536, len(data)
    assert data[0x6FFB:0x700E] == b"Two thousand years,"
    # Binary loader 預設為 32 位元；本工具只接受此固定 16 位元段。
    # 僅調整由唯讀輸入新建的一次性 DB，不載入或修改正式資料庫。
    ida_ida.inf_set_app_bitness(16)
    for segment in idautils.Segments():
        assert ida_segment.set_segm_addressing(ida_segment.getseg(segment), 0)
    assert not ida_ida.inf_is_64bit() and not ida_ida.inf_is_32bit_exactly()
    # 兩個入口來自 dosgolem 的實際 SAR 載入記錄，不對其用途命名。
    callback = int.from_bytes(data[0x300C:0x300E], "little")
    for address in (0x3000, 0x6000, callback):
        ida_ua.create_insn(address)
        ida_funcs.add_func(address)
    ida_auto.auto_wait()
    needle = b"Two thousand years,"
    offsets = [i for i in range(len(data)) if data.startswith(needle, i)]
    assert offsets == [0x6FFB], offsets
    targets = range(0x6FE0, 0x7040)
    references = []
    for address in targets:
        for reference in idautils.XrefsTo(address):
            if reference.type == ida_xref.fl_F:
                continue
            function = ida_funcs.get_func(reference.frm)
            references.append({
                "from":hex(reference.frm), "to":hex(address), "xref_type":reference.type,
                "original_function":idc.get_func_name(reference.frm),
                "function_start":hex(function.start_ea) if function else None,
                "bytes":(ida_bytes.get_bytes(reference.frm, idc.get_item_size(reference.frm)) or b"").hex(),
                "operand":ida_lines.tag_remove(idc.generate_disasm_line(reference.frm, 0) or ""),
                "evidence_level":"unknown; navigation until consumer verified",
            })
    functions = []
    starts = {int(r["function_start"], 16) for r in references if r["function_start"]}
    for address in (0x6363, callback):
        function = ida_funcs.get_func(address)
        if function:
            starts.add(function.start_ea)
    starts = sorted(starts)
    for start in starts:
        function = ida_funcs.get_func(start)
        instructions = []
        for address in idautils.FuncItems(start):
            instructions.append({
                "address":hex(address),
                "bytes":(ida_bytes.get_bytes(address, idc.get_item_size(address)) or b"").hex(),
                "operand":ida_lines.tag_remove(idc.generate_disasm_line(address, 0) or ""),
            })
        functions.append({"original_name":idc.get_func_name(start), "start":hex(start),
                          "end":hex(function.end_ea), "instructions":instructions})
    result = {
        "schema":1, "tool":"IDA " + ida_kernwin.get_kernel_version(),
        "input_name":source.name, "input_sha256":hashlib.sha256(data).hexdigest(),
        "processor":ida_ida.inf_get_procname(), "bits":16,
        "address_space":"IDA binary image base 0; offsets in dosgolem segment 01B1 at 10729940 steps",
        "runtime_mapping":"IDA 0x6FFB corresponds to dosgolem 01B1:6FFB, linear 0x8B0B",
        "function_count":ida_funcs.get_func_qty(), "text_offsets":[hex(i) for i in offsets],
        "navigation_anchors":[
            {"address":"0x6363", "bytes":data[0x6363:0x636B].hex(),
             "reason":"original immediate 0x6FF0 preceding indirect call; semantics unknown"},
            {"address":"0x300C", "bytes":data[0x300C:0x300E].hex(),
             "pointer_value":hex(callback), "reason":"snapshot word; call contract unknown"},
        ],
        "references":references, "functions":functions,
        "limitation":"xref and function boundaries do not alone prove text rendering semantics",
    }
    output = pathlib.Path(os.environ["ZELIARD_IDA_OUTPUT"])
    output.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
    ida_pro.qexit(0)


try:
    main()
except Exception:
    import traceback
    pathlib.Path(os.environ["ZELIARD_IDA_OUTPUT"] + ".error").write_text(traceback.format_exc())
    ida_pro.qexit(1)

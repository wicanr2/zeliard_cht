"""IDA 9.4：匯出固定常駐程式的原始指令，定位美術解碼入口。只用一次性 DB。"""
import hashlib
import json
import os
from pathlib import Path

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

source = Path(ida_nalt.get_input_file_path())
data = source.read_bytes()
assert len(data) == 65536
assert hashlib.sha256(data).hexdigest() == "c9f188ea847e8d364ccf8dbc9f6a9b9c74e192c0c83bd7a012b552e6d89b13d8"
index_path = Path(__file__).resolve().parent.parent/"docs/source-index.json"
index_data = index_path.read_bytes()
index = json.loads(index_data)
assert index["input_sha256"] == hashlib.sha256(data).hexdigest()
annotations = {int(x["address"], 16): x for x in index["entries"]}


def annotation(address):
    entry = annotations.get(address, {"level": "未知", "semantic": None, "evidence": []})
    return {**entry, "warning": entry["level"] != "已證實"}


ida_ida.inf_set_app_bitness(16)
for start in idautils.Segments():
    assert ida_segment.set_segm_addressing(ida_segment.getseg(start), 0)
ida_auto.auto_wait()
instructions = []
address = 0x100
while address < 0x3000:
    insn = ida_ua.insn_t()
    size = ida_ua.decode_insn(insn, address)
    if not size or address+size > 0x3000:
        address += 1
        continue
    ida_ua.create_insn(address)
    function = ida_funcs.get_func(address)
    instructions.append({"address": hex(address), "bytes": ida_bytes.get_bytes(address, size).hex(),
                         "original_function": idc.get_func_name(address),
                         "function_start": hex(function.start_ea) if function else None,
                         "operand": ida_lines.tag_remove(idc.generate_disasm_line(address, 0) or ""),
                         "semantic_index": annotation(address)})
    address += size
result = {"schema": 1, "tool": "IDA "+ida_kernwin.get_kernel_version(),
          "input_name": "probe-intro-memory/memory-01B1_0000.bin",
          "input_sha256": hashlib.sha256(data).hexdigest(), "bits": 16,
          "source_index_sha256": hashlib.sha256(index_data).hexdigest(),
          "address_space": "IDA binary base 0，對應此快照 dosgolem 01B1 段內偏移",
          "scope": "0100..3000 線性候選指令供查詢，不把資料誤反組譯當函式或已證實語意",
          "navigation_anchors": [{"address": hex(a), "original_function": idc.get_func_name(a),
                                   "bytes": data[a:a+8].hex(), "semantic_index": annotation(a)}
                                  for a in (0xdad, 0xdbc, 0xdcc, 0xdd1, 0xe13, 0xe34, 0xe73, 0xe9c, 0xeba, 0xef5, 0x2c2a)],
          "instructions": instructions}
Path(os.environ["ZELIARD_HD_IDA_OUTPUT"]).write_text(json.dumps(result, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
ida_pro.qexit(0)

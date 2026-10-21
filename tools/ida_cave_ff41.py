"""IDA 9.4：原樣匯出FF41繪製分支，不覆寫名稱或運算元。"""
import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_kernwin
import ida_lines
import ida_nalt
import ida_pro
import ida_ua
import idautils
import idc

source = Path(ida_nalt.get_input_file_path())
data = source.read_bytes()
assert len(data) == 65536
assert hashlib.sha256(data).hexdigest() == "5e7f76080cc542eb06ca90ec349ad2ec0c5ab374c32acf59bc4c4307899a0a5b"
assert ida_bytes.get_bytes(0, 65536) == data
index_raw = Path("/project/docs/cave-ff41-source-index.json").read_bytes()
index = json.loads(index_raw)
assert index["input_sha256"] == hashlib.sha256(data).hexdigest()
annotations = {int(row["address"], 16): row for row in index["entries"]}
for address, row in annotations.items():
    assert data[address:address + len(bytes.fromhex(row["bytes"]))].hex() == row["bytes"]
# 3A85的三byte JMP止於3A88；只在一次性副本補後續未分析指令。
address = 0x3a88
while address < 0x3d34:
    size = ida_ua.create_insn(address)
    assert size and address + size <= 0x3d34, hex(address)
    address += size
ida_auto.auto_wait()
rows = []
addresses = list(idautils.Heads(0x3a60, 0x3e20)) + list(idautils.Heads(0x703c, 0x7114))
for address in addresses:
    if not ida_bytes.is_code(ida_bytes.get_full_flags(address)):
        continue
    function = ida_funcs.get_func(address)
    semantic = annotations.get(address, {"level": "未知", "semantic": None, "evidence": []})
    rows.append({"address": hex(address),
                 "bytes": ida_bytes.get_bytes(address, idc.get_item_size(address)).hex(),
                 "operand": ida_lines.tag_remove(idc.generate_disasm_line(address, 0) or ""),
                 "original_function": idc.get_func_name(address),
                 "function_start": hex(function.start_ea) if function else None,
                 "code_refs": [hex(x) for x in idautils.CodeRefsFrom(address, False)],
                 "data_refs": [hex(x) for x in idautils.DataRefsFrom(address)],
                 "semantic_index": {**semantic, "warning": semantic["level"] != "已證實"}})
assert len(rows) > 100
report = {"schema": 1, "tool": "IDA " + ida_kernwin.get_kernel_version(),
          "input_name": source.name, "input_sha256": hashlib.sha256(data).hexdigest(),
          "script_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
          "source_index_sha256": hashlib.sha256(index_raw).hexdigest(),
          "address_space": "IDA base0，對應dosgolem 01B1段內偏移；不是EXE檔案偏移",
          "instructions": rows,
          "queries": [{"address": hex(address),
                       "xrefs": [{"from": hex(x.frm), "type": x.type}
                                 for x in idautils.XrefsTo(address)]}
                      for address in (0xff41, 0x4129, 0x3a60)],
          "limits": "只匯出原始定位；語意待正常來源與像素重建核對，不改資料庫名稱。"}
Path("/output/evidence.json").write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
ida_pro.qexit(0)

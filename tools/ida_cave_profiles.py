"""IDA9.4：保留原名與bytes，合併048主角組合的分級索引。"""
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
import idautils
import idc

source = Path(ida_nalt.get_input_file_path())
data = source.read_bytes()
index_raw = Path("/project/docs/cave-profile-source-index.json").read_bytes()
index = json.loads(index_raw)
assert len(data) == 65536 and hashlib.sha256(data).hexdigest() == index["input_sha256"]
assert ida_bytes.get_bytes(0, 65536) == data
annotations = {int(row["address"], 16): row for row in index["entries"]}
for address, row in annotations.items():
    assert data[address:address + len(bytes.fromhex(row["bytes"]))].hex() == row["bytes"]
ida_auto.auto_wait()
rows = []
for address in idautils.Heads(0x3a95, 0x3d34):
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
assert len(rows) > 200 and sum(not r["semantic_index"]["warning"] for r in rows) == 9
report = {"schema": 1, "tool": "IDA " + ida_kernwin.get_kernel_version(),
          "input_name": source.name, "input_sha256": index["input_sha256"],
          "source_index_sha256": hashlib.sha256(index_raw).hexdigest(),
          "script_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
          "address_space": index["address_space"], "instructions": rows,
          "limits": "只核正常來源所需繪製consumer；未知指令語意仍附警示，不改原始定位或函式名。"}
Path("/output/evidence.json").write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
ida_pro.qexit(0)

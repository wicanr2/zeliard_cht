"""IDA9.4：匯出固定洞窟段的地圖解壓與圖集選擇，合併分級索引。"""
import hashlib
import json
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
raw = source.read_bytes()
sha = lambda b: hashlib.sha256(b).hexdigest()
assert len(raw) == 65536 and sha(raw) == "12d91918212c4688a575bc31d395c133c7232f99ead769c7cfbc680333545c93"
index_data = (Path(__file__).resolve().parent.parent/"docs/cave-terrain-source-index.json").read_bytes()
index = json.loads(index_data)
assert index["input_sha256"] == sha(raw)
annotations = {int(x["address"], 16): x for x in index["entries"]}
ida_ida.inf_set_app_bitness(16)
for start in idautils.Segments():
    assert ida_segment.set_segm_addressing(ida_segment.getseg(start), 0)
anchors = (0x100, 0x6c98, 0x6ced, 0x6d06, 0x6d57, 0x7e93, 0x7ebb)
for start in anchors:
    ida_ua.create_insn(start)
    ida_funcs.add_func(start)
ida_auto.auto_wait()
rows = []
for lo, hi in ((0x6c98, 0x6d82), (0x7e93, 0x7f13)):
    for address in idautils.Heads(lo, hi):
        if not ida_bytes.is_code(ida_bytes.get_full_flags(address)):
            continue
        item = annotations.get(address, {"level": "未知", "semantic": None, "evidence": []})
        rows.append({"address": hex(address), "original_function": idc.get_func_name(address),
                     "bytes": ida_bytes.get_bytes(address, idc.get_item_size(address)).hex(),
                     "operand": ida_lines.tag_remove(idc.generate_disasm_line(address, 0) or ""),
                     "semantic_index": {**item, "warning": item["level"] != "已證實"}})
report = {"schema": 1, "tool": "IDA "+ida_kernwin.get_kernel_version(),
          "input_name": "workplace/combat-v1/cave-entered.primary", "input_sha256": sha(raw),
          "source_index_sha256": sha(index_data), "script_sha256": sha(Path(__file__).read_bytes()),
          "address_space": index["address_space"], "instructions": rows,
          "switch_words": [int.from_bytes(raw[a:a+2], "little") for a in range(0x6cfe, 0x6d06, 2)],
          "scope": "唯讀固定正常洞窟段；一次性DB，不改名稱或原始定位，未驗指令有警示"}
Path("/output/evidence.json").write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
ida_pro.qexit(0)

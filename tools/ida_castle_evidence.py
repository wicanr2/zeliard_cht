"""IDA9.4：唯讀城堡邊界段，保存原始函式與像素呈現相關指令。"""
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
assert len(raw) == 65536 and hashlib.sha256(raw).hexdigest() == "aed45060ee02eafa07b41891d326ac0217a8afdd2957d79b401369a8bb2fc11d"
ida_ida.inf_set_app_bitness(16)
for start in idautils.Segments():
    assert ida_segment.set_segm_addressing(ida_segment.getseg(start), 0)
index_path = Path(__file__).resolve().parent.parent/'docs/castle-source-index.json'
index_data = index_path.read_bytes()
index = json.loads(index_data)
assert index['input_sha256'] == hashlib.sha256(raw).hexdigest()
annotations = {int(x['address'], 16): x for x in index['entries']}
for start in (0x100, 0x3051, 0x31d8, 0x359a, 0x35d7, 0x68ac, 0x6950, 0x6975):
    ida_ua.create_insn(start)
    ida_funcs.add_func(start)
ida_auto.auto_wait()
functions = []
seen = set()
for anchor in (0x3051, 0x31d8, 0x359a, 0x35d7, 0x68ac, 0x6950, 0x6975):
    function = ida_funcs.get_func(anchor)
    if not function or function.start_ea in seen:
        continue
    seen.add(function.start_ea)
    rows = []
    for address in idautils.FuncItems(function.start_ea):
        rows.append({"address": hex(address), "bytes": ida_bytes.get_bytes(address, idc.get_item_size(address)).hex(),
                     "operand": ida_lines.tag_remove(idc.generate_disasm_line(address, 0) or ""),
                     "semantic_index": {**annotations.get(address, {'level': '未知', 'semantic': None, 'evidence': []}),
                                        'warning': annotations.get(address, {}).get('level') != '已證實'}})
    functions.append({"original_function": idc.get_func_name(function.start_ea), "start": hex(function.start_ea), "end": hex(function.end_ea), "instructions": rows})
report = {"schema": 1, "tool": "IDA "+ida_kernwin.get_kernel_version(), "input_sha256": hashlib.sha256(raw).hexdigest(),
          "source_index_sha256": hashlib.sha256(index_data).hexdigest(), "script_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
          "address_space": "IDA binary base0＝此快照dosgolem01B1段內偏移；原1MiB RAM檔案偏移需另加1B10",
          "scope": "固定正常城堡右邊界失配；不改函式名、DB或原版bytes", "functions": functions}
Path('/output/evidence.json').write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')
ida_pro.qexit(0)

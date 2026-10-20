"""IDA9.4匯出城鎮地圖及圖像選擇端，保留原始定位與證據分級。"""
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
data = source.read_bytes()
assert len(data) == 65536
assert hashlib.sha256(data).hexdigest() == 'aed45060ee02eafa07b41891d326ac0217a8afdd2957d79b401369a8bb2fc11d'
index_path = Path(__file__).resolve().parent.parent / 'docs/town-resources-source-index.json'
index_data = index_path.read_bytes()
index = json.loads(index_data)
assert index['input_sha256'] == hashlib.sha256(data).hexdigest()
annotations = {int(row['address'], 16): row for row in index['entries']}
ida_ida.inf_set_app_bitness(16)
for start in idautils.Segments():
    assert ida_segment.set_segm_addressing(ida_segment.getseg(start), 0)
anchors = (0xa84, 0xa8b, 0xad6, 0x6aa2, 0x6aaf, 0x6aed, 0x6c72, 0x6d9e, 0x6def, 0x6e29, 0x6ff8)
for start in anchors:
    ida_ua.create_insn(start)
    ida_funcs.add_func(start)
ida_auto.auto_wait()
functions, seen = [], set()
for anchor in anchors:
    function = ida_funcs.get_func(anchor)
    assert function is not None, hex(anchor)
    if function.start_ea in seen:
        continue
    seen.add(function.start_ea)
    rows = []
    for address in idautils.FuncItems(function.start_ea):
        semantic = annotations.get(address, {'level': '未知', 'semantic': None, 'evidence': []})
        rows.append({'address': hex(address), 'bytes': ida_bytes.get_bytes(address, idc.get_item_size(address)).hex(),
                     'operand': ida_lines.tag_remove(idc.generate_disasm_line(address, 0) or ''),
                     'original_function': idc.get_func_name(address),
                     'code_refs': [hex(x) for x in idautils.CodeRefsFrom(address, False)],
                     'data_refs': [hex(x) for x in idautils.DataRefsFrom(address)],
                     'semantic_index': {**semantic, 'warning': semantic['level'] != '已證實'}})
    functions.append({'original_function': idc.get_func_name(function.start_ea), 'start': hex(function.start_ea),
                      'end': hex(function.end_ea), 'instructions': rows})
tables = [{'address': hex(address), 'bytes': data[address:address+size].hex(),
           'xrefs': [{'from': hex(x.frm), 'type': x.type} for x in idautils.XrefsTo(address)]}
          for address, size in ((0x10c, 2), (0xaca, 12), (0x10c8, 110), (0x6ad3, 22), (0x6dce, 33))]
report = {'schema': 1, 'tool': 'IDA ' + ida_kernwin.get_kernel_version(), 'input_name': index['input_name'],
          'input_sha256': hashlib.sha256(data).hexdigest(), 'source_index_sha256': hashlib.sha256(index_data).hexdigest(),
          'script_sha256': hashlib.sha256(Path(__file__).read_bytes()).hexdigest(), 'address_space': index['address_space'],
          'functions': functions, 'tables': tables,
          'limits': '只匯出原始圖像與地圖選擇端；靜態選擇及結構資料不代表正常可達或HD已完成。'}
Path('/output/evidence.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
ida_pro.qexit(0)

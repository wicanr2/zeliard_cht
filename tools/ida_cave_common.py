"""IDA9.4匯出DCHR載入與平台表的窄範圍，原始名稱及分級語意並列。"""
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
index_path = Path('/project/docs/cave-common-source-index.json')
index_raw = index_path.read_bytes()
index = json.loads(index_raw)
assert len(data) == 65536 and hashlib.sha256(data).hexdigest() == index['input_sha256']
# archive byte從0起算，記錄編號從1起算；02,37即SAR3 record54。
assert data[0x9c13:0x9c15] == bytes((2, 55)) and data[0x9c15:0x9c1d] == b'DCHR.GRP'
annotations = {int(row['address'], 16): row for row in index['entries']}
ida_ida.inf_set_app_bitness(16)
for start in idautils.Segments():
    assert ida_segment.set_segm_addressing(ida_segment.getseg(start), 0)
windows = ((0x7ebb, 0x7ecd), (0x7fdc, 0x8074), (0x8074, 0x8109),
           (0x8109, 0x814c), (0x814c, 0x8163), (0x8352, 0x8366), (0x6db1, 0x6dcb))
for start, end in windows:
    ida_ua.create_insn(start)
    ida_funcs.add_func(start)
ida_auto.auto_wait()
functions = []
for start, end in windows:
    function = ida_funcs.get_func(start)
    assert function is not None, hex(start)
    rows = []
    for address in idautils.Heads(start, end):
        assert ida_bytes.is_code(ida_bytes.get_full_flags(address)), hex(address)
        semantic = annotations.get(address, {'level': '未知', 'semantic': None, 'evidence': []})
        rows.append({'address': hex(address), 'bytes': ida_bytes.get_bytes(address, idc.get_item_size(address)).hex(),
                     'operand': ida_lines.tag_remove(idc.generate_disasm_line(address, 0) or ''),
                     'original_function': idc.get_func_name(address),
                     'code_refs': [hex(x) for x in idautils.CodeRefsFrom(address, False)],
                     'data_refs': [hex(x) for x in idautils.DataRefsFrom(address)],
                     'semantic_index': {**semantic, 'warning': semantic['level'] != '已證實'}})
    functions.append({'window': [hex(start), hex(end)], 'original_function': idc.get_func_name(start),
                      'function_start': hex(function.start_ea), 'function_end': hex(function.end_ea), 'instructions': rows})
tables = [{'address': hex(address), 'bytes': data[address:address+size].hex(),
           'xrefs': [{'from': hex(x.frm), 'type': x.type} for x in idautils.XrefsTo(address)],
           'semantic_index': annotations.get(address, {'level': '未知', 'semantic': None, 'evidence': []})}
          for address, size in ((0x10c, 2), (0xc004, 4), (0x9c13, 11))]
report = {'schema': 1, 'tool': 'IDA '+ida_kernwin.get_kernel_version(), 'input_name': index['input_name'],
          'input_sha256': index['input_sha256'], 'source_index_sha256': hashlib.sha256(index_raw).hexdigest(),
          'script_sha256': hashlib.sha256(Path(__file__).read_bytes()).hexdigest(), 'address_space': index['address_space'],
          'functions': functions, 'tables': tables,
          'limits': '只核DCHR載入與平台圖格／表；不重製控制或將靜態來源當全部正常可達。'}
Path('/output/evidence.json').write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')
ida_pro.qexit(0)

"""IDA9.4：長老室內動畫的原始函式與圖塊表交叉參照。"""
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
assert hashlib.sha256(raw).hexdigest() == "ab2a46c7b6918b360a3db214935e73477b6ec7e9b04ae84f4790685cc7e86e9f"
ida_ida.inf_set_app_bitness(16)
for start in idautils.Segments():
    assert ida_segment.set_segm_addressing(ida_segment.getseg(start), 0)
minimum = ida_ida.inf_get_min_ea()
if minimum != 0xa000:
    assert minimum == 0
    ida_segment.rebase_program(0xa000, ida_segment.MSF_FIXONCE)
assert ida_bytes.get_bytes(0xa000, len(raw)) == raw
for offset in (0, 2):
    address = int.from_bytes(raw[offset:offset+2], "little")
    ida_ua.create_insn(address)
    ida_funcs.add_func(address)
ida_auto.auto_wait()
index_path = Path('/project/docs/sage-frame-source-index.json')
annotations = {}
if index_path.is_file():
    annotations = {int(row['address'], 16): row for row in json.loads(index_path.read_text(encoding='utf-8'))['entries']}
functions = []
for start in idautils.Functions(0xa000, 0xa000+len(raw)):
    fn = ida_funcs.get_func(start)
    rows = []
    for address in idautils.FuncItems(start):
        semantic = annotations.get(address, {'level': '未知', 'semantic': None})
        rows.append({'address': hex(address), 'bytes': ida_bytes.get_bytes(address, idc.get_item_size(address)).hex(),
                     'operand': ida_lines.tag_remove(idc.generate_disasm_line(address, 0) or ''),
                     'semantic_index': {**semantic, 'warning': semantic['level'] != '已證實'},
                     'data_refs': [hex(x) for x in idautils.DataRefsFrom(address)]})
    functions.append({'original_function': idc.get_func_name(start), 'start': hex(start), 'end': hex(fn.end_ea), 'instructions': rows})
report = {'schema': 1, 'tool': 'IDA '+ida_kernwin.get_kernel_version(), 'file': source.name,
          'sha256': hashlib.sha256(raw).hexdigest(), 'script_sha256': hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
          'address_space': 'IDA base A000 = KENJPRO.BIN 檔案偏移0；正常01B1段另核',
          'source_index_sha256': hashlib.sha256(index_path.read_bytes()).hexdigest() if index_path.is_file() else None,
          'functions': functions, 'limits': '未分級指令保留未知警示；不推定事件、金額或動畫時序。'}
report['tables'] = []
if index_path.is_file():
    for table in json.loads(index_path.read_text(encoding='utf-8'))['tables']:
        address = int(table['address'], 16)
        data = ida_bytes.get_bytes(address, table['bytes'])
        assert data == raw[address-0xa000:address-0xa000+table['bytes']]
        report['tables'].append({**table, 'sha256': hashlib.sha256(data).hexdigest(),
                                 'xrefs': [{'from':hex(x.frm),'kind':x.type} for x in idautils.XrefsTo(address)]})
Path('/output/sage-frames-evidence.json').write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')
ida_pro.qexit(0)

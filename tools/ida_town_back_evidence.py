"""IDA 9.4：城鎮背向姿態與居民可見性，只匯出、不改原名。"""
import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_kernwin
import ida_lines
import ida_pro
import ida_ua
import idautils
import idc

source = Path('/project/workplace/full-hd-v1/castle-edge-ida-v1/primary.segment')
raw = source.read_bytes()
assert len(raw) == 65536
assert hashlib.sha256(raw).hexdigest() == 'aed45060ee02eafa07b41891d326ac0217a8afdd2957d79b401369a8bb2fc11d'
assert ida_bytes.get_bytes(0, len(raw)) == raw
index_path = Path('/project/docs/town-back-source-index.json')
index_data = index_path.read_bytes()
index = json.loads(index_data)
assert index['input_sha256'] == hashlib.sha256(raw).hexdigest()
annotations = {int(entry['address'], 16): entry for entry in index['entries']}
anchors = [0x68ac, 0x6975, 0x6a77, 0x6c4e]
links = []
for target in (0x68ac, 0x68ba, 0x6c4e, 0xe7):
    for ref in idautils.XrefsTo(target):
        fn = ida_funcs.get_func(ref.frm)
        links.append({'target': hex(target), 'from': hex(ref.frm), 'kind': ref.type,
                      'original_function': idc.get_func_name(ref.frm)})
        if fn is not None and fn.start_ea not in anchors:
            anchors.append(fn.start_ea)
for start in anchors:
    ida_ua.create_insn(start)
    ida_funcs.add_func(start)
ida_auto.auto_wait()
functions = []
for anchor in anchors:
    fn = ida_funcs.get_func(anchor)
    assert fn is not None and fn.start_ea == anchor
    rows = []
    for address in idautils.FuncItems(fn.start_ea):
        rows.append({'address': hex(address), 'bytes': ida_bytes.get_bytes(address, idc.get_item_size(address)).hex(),
                     'operand': ida_lines.tag_remove(idc.generate_disasm_line(address, 0) or ''),
                     'semantic_index': {**annotations.get(address, {'level': '未知', 'semantic': None, 'evidence': []}),
                                        'warning': annotations.get(address, {}).get('level') != '已證實'}})
    functions.append({'original_function': idc.get_func_name(fn.start_ea), 'start': hex(fn.start_ea),
                      'end': hex(fn.end_ea), 'instructions': rows})
report = {'schema': 1, 'tool': 'IDA '+ida_kernwin.get_kernel_version(),
          'input_name': str(source), 'input_sha256': hashlib.sha256(raw).hexdigest(),
          'script_sha256': hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
          'source_index_sha256': hashlib.sha256(index_data).hexdigest(),
          'address_space': 'IDA binary base0＝此快照dosgolem01B1段內偏移；1MiB RAM檔案定位另加1B10',
          'scope': '正常進門phase4的居民可見性與背向繪製；所有原名與bytes保留',
          'xrefs': links, 'functions': functions}
Path('/output/evidence.json').write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')
ida_pro.qexit(0)

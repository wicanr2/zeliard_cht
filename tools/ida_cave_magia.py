"""IDA9.4窄範圍匯出；原始定位保留，未分級項醒目標記未知。"""
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
index_path = Path('/project/docs/cave-magia-source-index.json')
index_raw = index_path.read_bytes()
index = json.loads(index_raw)
assert len(data)==65536 and hashlib.sha256(data).hexdigest()==index['input_sha256']
assert ida_bytes.get_bytes(0,65536)==data, '資料庫原始位元組與固定輸入不符'
annotations = {int(row['address'],16):row for row in index['entries']}
ida_auto.auto_wait()
windows=[]
for start,end in index['windows']:
    assert 0<=start<end<=65536 and end-start<=4096
    rows=[]
    for address in idautils.Heads(start,end):
        if not ida_bytes.is_code(ida_bytes.get_full_flags(address)):
            continue
        semantic=annotations.get(address,{'level':'未知','semantic':None,'evidence':[]})
        function=ida_funcs.get_func(address)
        rows.append({'address':hex(address),'bytes':ida_bytes.get_bytes(address,idc.get_item_size(address)).hex(),
                     'operand':ida_lines.tag_remove(idc.generate_disasm_line(address,0) or ''),
                     'original_function':idc.get_func_name(address),
                     'function_start':hex(function.start_ea) if function else None,
                     'code_refs':[hex(x)for x in idautils.CodeRefsFrom(address,False)],
                     'data_refs':[hex(x)for x in idautils.DataRefsFrom(address)],
                     'semantic_index':{**semantic,'warning':semantic['level']!='已證實'}})
    assert rows
    windows.append({'start':hex(start),'end':hex(end),'instructions':rows})
report={'schema':1,'tool':'IDA '+ida_kernwin.get_kernel_version(),'input_sha256':index['input_sha256'],
        'input_name':index['input_name'],'address_space':index['address_space'],
        'source_index_sha256':hashlib.sha256(index_raw).hexdigest(),
        'script_sha256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        'queries':[{'address':hex(address),'xrefs':[{'from':hex(x.frm),'type':x.type}for x in idautils.XrefsTo(address)]}for address in (0x412f,)],
        'tables':[{'address':hex(address),'bytes':data[address:address+size].hex(),'xrefs':[{'from':hex(x.frm),'type':x.type}for x in idautils.XrefsTo(address)],'semantic_index':{**annotations.get(address,{'level':'未知','semantic':None,'evidence':[]}),'warning':annotations.get(address,{}).get('level')!='已證實'}}for address,size in index['tables']],
        'windows':windows,'limits':'窄範圍導航證據；未分級語意不當已證實，不改函式名或運算元。'}
Path('/output/evidence.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
ida_pro.qexit(0)

"""IDA9.4：唯讀室內模組入口及圖塊繪製端；保留原名與未知警示。"""
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

source=Path(ida_nalt.get_input_file_path())
raw=source.read_bytes()
manifest=json.loads(Path('/project/workplace/full-hd-v1/interior-code-v1/inputs.json').read_text())
entry=next(x for x in manifest['inputs'] if x['file']==source.name)
assert hashlib.sha256(raw).hexdigest()==entry['sha256']
index_data=Path('/project/docs/interior-source-index.json').read_bytes()
index=json.loads(index_data)
source_index=next(x for x in index['inputs'] if x['file']==source.name)
assert source_index['sha256']==entry['sha256']
annotations={int(x['address'],16):x for x in source_index['entries']}
ida_ida.inf_set_app_bitness(16)
for start in idautils.Segments():
    assert ida_segment.set_segm_addressing(ida_segment.getseg(start),0)
minimum=ida_ida.inf_get_min_ea()
if minimum!=0xa000:
    assert minimum==0
    ida_segment.rebase_program(0xa000,ida_segment.MSF_FIXONCE)
assert ida_bytes.get_bytes(0xa000,len(raw))==raw
anchors=[int.from_bytes(raw[i:i+2],'little') for i in (0,2)]
for address in anchors:
    assert 0xa000<=address<0xa000+len(raw)
    ida_ua.create_insn(address)
    ida_funcs.add_func(address)
ida_auto.auto_wait()
functions=[]
for start in idautils.Functions(0xa000,0xa000+len(raw)):
    fn=ida_funcs.get_func(start)
    assert fn is not None
    rows=[]
    for address in idautils.FuncItems(start):
        rows.append({'address':hex(address),'bytes':ida_bytes.get_bytes(address,idc.get_item_size(address)).hex(),
                     'operand':ida_lines.tag_remove(idc.generate_disasm_line(address,0) or ''),
                     'semantic_index':{**annotations.get(address,{'level':'未知','semantic':None}),
                                       'warning':annotations.get(address,{}).get('level')!='已證實'}})
    functions.append({'original_function':idc.get_func_name(start),'start':hex(start),'end':hex(fn.end_ea),'instructions':rows})
report={'schema':1,'tool':'IDA '+ida_kernwin.get_kernel_version(),'input':entry,
        'script_sha256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        'source_index_sha256':hashlib.sha256(index_data).hexdigest(),
        'address_space':'IDA baseA000＝本BIN檔案偏移0；與正常載入對應須另核，不稱正常執行快照',
        'entry_words':[hex(x) for x in anchors],'functions':functions,
        'limits':'只匯出原始bytes與函式。未知及資料誤解碼候選不當已證實語意，不改函式名。'}
report['tables']=[]
for table in source_index['tables']:
    address=int(table['address'],16);data=ida_bytes.get_bytes(address,table['bytes'])
    assert data==raw[address-0xa000:address-0xa000+table['bytes']]
    report['tables'].append({**table,'sha256':hashlib.sha256(data).hexdigest(),
                             'xrefs':[{'from':hex(x.frm),'kind':x.type} for x in idautils.XrefsTo(address)]})
Path('/output/'+source.stem+'-evidence.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
ida_pro.qexit(0)

"""IDA9.4窄範圍門框載入證據，保留原始定位，語意預設未知。"""
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

path=Path(ida_nalt.get_input_file_path());data=path.read_bytes()
assert hashlib.sha256(data).hexdigest()=='5e7f76080cc542eb06ca90ec349ad2ec0c5ab374c32acf59bc4c4307899a0a5b'
assert ida_bytes.get_bytes(0,65536)==data
ida_funcs.add_func(0x78dd);ida_auto.auto_wait()
rows=[]
for address in idautils.Heads(0x78dd,0x7a20):
    if not ida_bytes.is_code(ida_bytes.get_full_flags(address)):continue
    rows.append({'address':hex(address),'bytes':ida_bytes.get_bytes(address,idc.get_item_size(address)).hex(),
                 'operand':ida_lines.tag_remove(idc.generate_disasm_line(address,0)or''),
                 'original_function':idc.get_func_name(address),'code_refs':[hex(x)for x in idautils.CodeRefsFrom(address,False)],
                 'data_refs':[hex(x)for x in idautils.DataRefsFrom(address)],
                 'level':'未知','semantic':None,'warning':True})
assert rows
report={'schema':1,'tool':'IDA '+ida_kernwin.get_kernel_version(),'input_name':path.name,
        'input_sha256':hashlib.sha256(data).hexdigest(),'script_sha256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        'address_space':'IDA base0，對應01B1段內偏移，不是EXE檔案偏移。','instructions':rows,
        'tables':[{'address':hex(a),'bytes':data[a:a+n].hex(),'xrefs':[{'from':hex(x.frm),'type':x.type}for x in idautils.XrefsTo(a)],'level':'未知','warning':True}for a,n in [(0x79b4,40),(0xc008,8)]],
        'limits':'導航與原始證據，不改原名、指令或原始資料庫，不因圖形相似推論通行規則。'}
Path('/output/evidence.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
ida_pro.qexit(0)

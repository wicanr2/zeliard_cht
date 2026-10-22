"""IDA9.4一次性DB：匯出生命、法術次數及回血寫入線索，不改原名。"""
from pathlib import Path
import hashlib,json,os,re,traceback
import ida_auto,ida_bytes,ida_funcs,ida_kernwin,ida_lines,ida_pro,idautils,idc

out=Path(os.environ['ZELIARD_RESOURCES_OUTPUT'])
try:
    raw=Path(os.environ['ZELIARD_RESOURCES_INPUT']).read_bytes()
    assert len(raw)==65536
    ida_auto.auto_wait()
    targets=[0x90,0x91,0xb2,0xb3,0x9d,*range(0xab,0xb2),*range(0xb4,0xbb)]
    wanted=set();refs=[];lines=[]
    def row(a):
        f=ida_funcs.get_func(a);b=ida_bytes.get_bytes(a,idc.get_item_size(a))or b''
        assert b==raw[a:a+len(b)]
        return {'address':hex(a),'bytes':b.hex(),'operand':ida_lines.tag_remove(idc.generate_disasm_line(a,0)or''),'original_function':idc.get_func_name(f.start_ea)if f else None,'function_start':hex(f.start_ea)if f else None,'level':'未知'}
    for target in targets:
        for x in idautils.XrefsTo(target):
            r=row(x.frm);r.update({'to':hex(target),'xref_type':x.type});refs.append(r)
            f=ida_funcs.get_func(x.frm)
            if f:wanted.add(f.start_ea)
    for a in idautils.Heads(0x100,0x10000):
        text=ida_lines.tag_remove(idc.generate_disasm_line(a,0)or'')
        if re.search(r'\b(?:word|byte)_(?:90|B2|9D|AB|AC|AD|AE|AF|B0|B1|B4|B5|B6|B7|B8|B9|BA)\b|\b(?:90h|0ABh|0B4h)\b',text,re.I):
            lines.append(row(a));f=ida_funcs.get_func(a)
            if f:wanted.add(f.start_ea)
    funcs=[]
    for a in sorted(wanted):
        f=ida_funcs.get_func(a)
        funcs.append({'original_name':idc.get_func_name(a),'start':hex(a),'end':hex(f.end_ea),'level':'未知','instructions':[row(i) for i in idautils.FuncItems(a)]})
    out.write_text(json.dumps({'schema':1,'tool':'IDA '+ida_kernwin.get_kernel_version(),'input_name':Path(os.environ['ZELIARD_RESOURCES_INPUT']).name,'input_sha256':hashlib.sha256(raw).hexdigest(),'address_space':'IDA binary base0 = dosgolem 01B1段內偏移；非EXE或SAR檔案偏移','function_count':ida_funcs.get_func_qty(),'references':refs,'lines':lines,'functions':funcs,'limits':'既有函式邊界及原名保留；輸出均為線索，語意須經來源及消費端核對。'},ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    ida_pro.qexit(0)
except Exception:
    out.with_suffix('.error').write_text(traceback.format_exc(),encoding='utf-8')
    ida_pro.qexit(1)

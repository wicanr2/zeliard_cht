"""066：盤點 SAR 模組中對話常式（706C）讀取的文字頁，比對現有譯表找出缺譯。

腳本格式依國王模組原始 bytes：FF 04 開始一段文字、FF 05 結束，FF 02 等待按鍵，
0D 換頁，11 或 FF FF 結束。一頁從第一個 FF 04 之後的文字起，跨過 FF 05 FF 02 FF 04 接續，
直到 0D、11 或 FF FF；與現有 dialogue.go 來源清單同一種切法。
頁內任一位置已有譯表鍵即視為已譯（報價、銀行等子對話以片段為鍵）。
用法：python3 dialogue_inventory.py --project <專案> --out <新JSON>
"""
import argparse, csv, json
from pathlib import Path

ap = argparse.ArgumentParser()
ap.add_argument('--project', required=True)
ap.add_argument('--out', required=True)
a = ap.parse_args()
P = Path(a.project)
coverage = json.loads((P / 'docs/hd-coverage.json').read_text())
keys = set()
for name in ('dialogue-prototype.tsv', 'labels-prototype.tsv'):
    with (P / 'text' / name).open(encoding='utf-8') as f:
        for row in csv.reader(f, delimiter='\t'):
            if row and row[0].startswith('zelres2.file.'):
                keys.add(int(row[0].split('.')[-1], 16))


def pages(data, base):
    """回傳 (起, 訖, 英文) 頁清單；起訖為檔案偏移。"""
    out, i, n = [], 0, len(data)
    while i < n - 1:
        if data[i] != 0xff or data[i + 1] != 0x04:
            i += 1
            continue
        start, j, text = i + 2, i + 2, []
        while j < n:
            c = data[j]
            if c == 0xff and j + 1 < n:
                code = data[j + 1]
                if code == 0x05:
                    k = j + 2
                    # FF 05 FF 02 FF 04 接續同頁；FF 05 FF 02 0D 或其他則結束本頁。
                    while k + 1 < n and data[k] == 0xff and data[k + 1] in (0x02, 0x01):
                        k += 2
                    if k + 1 < n and data[k] == 0xff and data[k + 1] == 0x04:
                        j = k + 2
                        continue
                    break
                if code == 0xff:
                    break
                j += 2
                continue
            if c in (0x0d, 0x11):
                break
            if 0x20 <= c < 0x7f:
                text.append(chr(c))
            j += 1
        if text:
            out.append((base + start, base + j, ''.join(text)))
        i = j
    return out


result = []
for r in coverage['records']:
    if r['archive'] != 'ZELRES2.SAR':
        continue
    data = (P / 'workplace/orig' / r['archive']).read_bytes()[r['payload_offset']:r['payload_offset'] + r['payload_bytes']]
    for start, end, text in pages(data, r['payload_offset']):
        translated = any(start <= k < end for k in keys)
        result.append({'record': r['key'], 'payload': hex(r['payload_offset']), 'start': hex(start), 'end': hex(end), 'translated': translated, 'text': text})
missing = [p for p in result if not p['translated']]
Path(a.out).write_text(json.dumps({'pages': len(result), 'missing': len(missing), 'records': sorted({p['record'] for p in missing}), 'items': result}, ensure_ascii=False, indent=1) + '\n')
print('pages', len(result), 'missing', len(missing), 'records', len({p['record'] for p in missing}))

"""閃爍探針輸出的判定：英文先現後換中文、同場景內HD／中文層中斷後恢復。

輸入為 cmd/zeliard-preview -flicker-probe 的 <prefix>.composes.jsonl 與 <prefix>.glyphs.jsonl。
每列合成點都是 GUI Draw 可能落下的位置；任何一列露出的狀態都可能被玩家看到。
用法：python3 flicker_report.py <prefix> [<prefix> ...]，結果印成 JSON。
"""
import collections, json, sys


def runs(rows, key):
    """同場景內 key 層 true→false→true 的中斷，回傳每段 false 的合成點數與指令數。
    城鎮對話另以來源鍵分段：兩段對話之間沒有對話的空檔不是中斷。"""
    gaps, scene, seen, start = [], None, False, None
    for r in rows:
        identity = (r['scene'], r.get('town', '')) if key == 'text_town' else r['scene']
        if identity != scene:
            scene, seen, start = identity, False, None
        on = r['layers'].get(key)
        if on is None:
            continue
        if on:
            if seen and start is not None:
                gaps.append({'scene': scene, 'from': start[0], 'to': r['step'], 'composes': start[1]})
            seen, start = True, None
        elif seen:
            start = (r['step'], 1) if start is None else (start[0], start[1] + 1)
    return gaps


def strings(glyphs):
    """依呼叫端、場景與列把字形接回字串，附該組英文露出的最長合成點數。"""
    groups = collections.defaultdict(list)
    for g in glyphs:
        groups[(g['caller'], g['scene'], g['y'])].append(g)
    out = []
    for (caller, scene, y), items in groups.items():
        items.sort(key=lambda g: (g['step']))
        out.append({'caller': caller, 'scene': scene, 'y': y, 'first_step': items[0]['step'], 'text': ''.join(chr(g['char']) if 32 <= g['char'] < 127 else '?' for g in items), 'max_visible_composes': max(g['visible_composes'] for g in items)})
    return sorted(out, key=lambda x: x['first_step'])


def report(prefix):
    rows = [json.loads(l) for l in open(prefix + '.composes.jsonl')]
    glyphs = [json.loads(l) for l in open(prefix + '.glyphs.jsonl')]
    # 英文先露出、之後被中文層蓋掉：玩家看到的「英文先閃再變中文」。
    flashes = [g for g in glyphs if g['visible_composes'] and g['text_covered_composes'] and g['first_visible_step'] < g['first_text_covered_step']]
    english = [g for g in glyphs if g['visible_composes'] and not g['text_covered_composes']]
    layers = sorted({k for r in rows for k in r['layers']})
    interruptions = {}
    for key in layers:
        gaps = runs(rows, key)
        if gaps:
            sizes = sorted(g['composes'] for g in gaps)
            interruptions[key] = {'count': len(gaps), 'max_composes': sizes[-1], 'median_composes': sizes[len(sizes) // 2], 'examples': gaps[:3]}
    by_caller = collections.Counter(g['caller'] for g in flashes)
    return {
        'prefix': prefix,
        'composes': len(rows), 'scenes': len({r['scene'] for r in rows}),
        'glyph_events': len(glyphs),
        'english_then_covered': len(flashes),
        'english_then_covered_by_caller': dict(by_caller.most_common()),
        'english_never_covered': len(english),
        'english_never_covered_by_caller': dict(collections.Counter(g['caller'] for g in english).most_common()),
        'composes_with_english': sum(1 for r in rows if r['english_pixels']),
        'layer_interruptions': interruptions,
        'flash_strings': strings(flashes),
        'english_strings': strings(english),
    }


if __name__ == '__main__':
    print(json.dumps([report(p) for p in sys.argv[1:]], ensure_ascii=False, indent=1))

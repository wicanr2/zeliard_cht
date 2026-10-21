#!/usr/bin/env python3
"""在Docker內由粉圓2.1製作固定字距的中文顯示字型，不啟動遊戲。"""
import argparse
import base64
import gzip
import hashlib
import json
import urllib.request
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont


def sha(data):
    return hashlib.sha256(data).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--download', action='store_true')
    args = ap.parse_args()
    p = args.project
    out = p / 'workplace/full-hd-v1/manual-hd-v1/rounded-font'
    out.mkdir(exist_ok=True)
    if args.download:
        commit = '98d53b3dac1730889edf548359c326c53624fa80'
        urls = {
            'jf-openhuninn-2.1.ttf': 'https://github.com/justfont/open-huninn-font/releases/download/v2.1/jf-openhuninn-2.1.ttf',
            'OFL.txt': f'https://raw.githubusercontent.com/justfont/open-huninn-font/{commit}/LICENSE',
        }
        records = []
        expected = {'jf-openhuninn-2.1.ttf': '9d5bf4932d31fe94c18cd8cfddc98bc1b14ce10f4e354c682179db290a99c825',
                    'OFL.txt': '5df6de36ba2fe26119ac161edbed97c62a9da3c1493f6930b9d9af7f5c93004b'}
        for name, url in urls.items():
            data = urllib.request.urlopen(url).read()
            if sha(data) != expected[name]:
                raise ValueError('字型輸入雜湊不符：' + name)
            (out / name).write_bytes(data)
            records.append({'file': name, 'url': url, 'sha256': sha(data), 'bytes': len(data)})
        (out / 'source.json').write_text(json.dumps({'font': 'jf open 粉圓 2.1', 'tag': 'v2.1', 'commit': commit, 'license': 'OFL-1.1', 'files': records}, ensure_ascii=False, indent=2) + '\n')
    native = p / 'workplace/music-release-v1/release-build/AppDir/font.hex.gz'
    raw = gzip.decompress(native.read_bytes())
    glyphs = {}
    for line in raw.decode().splitlines():
        cp, data = line.split(':', 1)
        glyphs[int(cp, 16)] = bytes.fromhex(data)
    # 只替換中文與中日韓標點；保留半形字、數字及既有排版寬度。
    used = set()
    for root in ('text', 'cmd', 'internal'):
        for file in (p / root).rglob('*'):
            if file.is_file() and file.suffix in ('.json', '.tsv', '.go') and not file.name.endswith('_test.go'):
                used.update(ord(c) for c in file.read_text() if ord(c) >= 0x2e80)
    ttf = out / 'jf-openhuninn-2.1.ttf'
    sizes = (16, 18, 24, 32)
    faces = {size: ImageFont.truetype(str(ttf), size * 4) for size in sizes}
    missing = bytes(faces[16].getmask('\U0010ffff'))
    masks, unavailable = {}, []
    for cp in sorted(used):
        data = glyphs.get(cp)
        if data is None or len(data) not in (16, 32):
            continue
        ch = chr(cp)
        if bytes(faces[16].getmask(ch)) == missing:
            unavailable.append(ch)
            continue
        width = len(data) // 16 * 8
        packed = bytearray()
        for size in sizes:
            face = faces[size]
            box = face.getbbox(ch)
            # 以墨跡垂直置中，四倍取樣後縮至正式尺寸。筆畫與抗鋸齒一併保留。
            target = Image.new('L', (width * size // 16 * 4, size * 4))
            ink = Image.new('L', (max(1, box[2]-box[0]), max(1, box[3]-box[1])))
            ImageDraw.Draw(ink).text((-box[0], -box[1]), ch, font=face, fill=255)
            available_width, available_height = target.width-4, target.height-4
            ratio = min(1, available_width/ink.width, available_height/ink.height)
            if ratio < 1:
                ink = ink.resize((max(1, round(ink.width*ratio)), max(1, round(ink.height*ratio))), Image.Resampling.LANCZOS)
            target.paste(ink, ((target.width-ink.width)//2, (target.height-ink.height)//2))
            mask = target.resize((width * size // 16, size), Image.Resampling.LANCZOS)
            packed.extend(mask.tobytes())
        masks[str(cp)] = base64.b64encode(packed).decode()
    # 基底HEX保持原檔；粉圓衍生物單獨保存於OFL灰階遮罩，不混寫兩種字型。
    compressed = native.read_bytes()
    (out / 'font.hex.gz').write_bytes(compressed)
    sidecar = {'schema': 1, 'font_sha256': sha(compressed), 'sizes': sizes, 'glyphs': masks}
    (out / 'font.hex.gz.smooth.json.gz').write_bytes(gzip.compress(json.dumps(sidecar, separators=(',', ':')).encode(), mtime=0))
    (out / 'manifest.json').write_text(json.dumps({'source': json.loads((out/'source.json').read_text()), 'base_sha256': sha(native.read_bytes()), 'font_sha256': sha(compressed), 'rounded_glyphs': len(masks), 'fallback_glyphs': unavailable, 'sizes': sizes, 'method': '4x supersampling; fixed original advance; grayscale coverage; legacy threshold 128', 'tests_executed': False}, ensure_ascii=False, indent=2)+'\n')
    print('ROUNDED_FONT', len(masks), 'glyphs;', len(unavailable), 'keep original; no game or tests executed')


if __name__ == '__main__':
    main()

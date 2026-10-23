#!/usr/bin/env python3
"""069：1.0.8片頭與結局插畫重畫版60秒本機推廣片。

畫面全部取自1.0.8正式HD包解出的執行檔，以閃爍探針跑原版按鍵鍵序（片頭四次空白鍵、開機不按鍵；
結局另以poke 01B1:0049=1結構觸發），存下與視窗同一條合成路徑的畫面。主機負載下視窗實錄跑不到原速，
使用者選定以這些合成畫面製片。每格是靜態畫面；魔王宣言與國王說話以原版畫格順序連續呈現，時間間隔不代表原速。
配樂是既有原版MSCAdLib標題曲錄音（music-release-v1/title-music），與畫面非同期。
"""
from pathlib import Path
import array
import hashlib
import json
import subprocess
import wave
from PIL import Image, ImageDraw, ImageFont, ImageOps

P = Path('/project')
V = (P / 'VERSION').read_text().strip()
W = P / 'workplace/release-v108'
D = P / 'dist-all' / V / 'promo'
T = W / 'edit'
PROBE = P / 'workplace/picture-coverage-v1/probe'
BASE = P / 'workplace/picture-coverage-v1/probe-base/v108-rel-CHT-HD'
assert V == 'v.1.0.8-20261023' and not list(D.glob('*.mp4'))
T.mkdir(exist_ok=False)
(T / 'frames').mkdir()
h = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
binary = h(BASE / 'zeliard-preview')
assert binary == h(W / 'package/linux/zeliard-preview'), '探針基底不是本版正式執行檔'
package = P / 'dist-all' / V / 'full-local' / f'Zeliard-{V}-CHT-HD-Linux-x86_64.AppImage'


def run(args, seconds=300):
    return subprocess.run(args, check=True, capture_output=True, timeout=seconds)


def write(p, data):
    p.write_text(json.dumps(data, ensure_ascii=False, indent=2) + '\n')


RUNS = {'boot': ('rel-boot-threat', 'probe/research/boot-threat.commands'), 'story': ('rel-story-title', 'probe/research/story-title.commands'),
        'ending': ('rel-ending-poke', 'tools/ending-poke.commands')}
frames = []


def frame(run_name, step):
    """探針在指定指令數之後的第一個合成點；複製到剪輯目錄，記錄來源與SHA。"""
    directory = PROBE / 'out' / RUNS[run_name][0]
    shots = sorted((int(f.stem.split('-')[-1]), f) for f in directory.glob('probe.shot-*.png'))
    at, source = next((s, f) for s, f in shots if s >= step)
    assert at - step < 200_000, (run_name, step, at)
    target = T / 'frames' / f'{run_name}-{at}.png'
    if not target.exists():
        target.write_bytes(source.read_bytes())
        frames.append({'run': run_name, 'probe_output': str(directory.relative_to(P)), 'commands': RUNS[run_name][1], 'step': at, 'file': str(target.relative_to(P)),
                       'sha256': h(target), 'indexed_sha256': h(source.with_suffix('.indexed'))})
    return target


font = P / 'workplace/release-v106/rounded-font/jf-openhuninn-2.1.ttf'
fonts = {n: ImageFont.truetype(str(font), n) for n in (22, 24, 26, 28, 32, 68)}
bg, ink, mint, gold, dim = '#101c23', '#f0eddf', '#7bc6b3', '#d5a96c', '#b3c2c6'
bounds = []


def text(d, xy, s, size, fill=ink):
    box = d.multiline_textbbox(xy, s, font=fonts[size], spacing=10)
    assert box[0] >= 0 and box[1] >= 0 and box[2] <= 1260 and box[3] <= 704, (s, box)
    bounds.append(list(box))
    d.multiline_text(xy, s, font=fonts[size], fill=fill, spacing=10)


def card(name, picture, heading, lines, note):
    """遊戲畫面960×600（合成畫布去掉上下工具列）放左側，右側說明欄。"""
    im = Image.new('RGB', (1280, 720), bg)
    im.paste(Image.open(picture).convert('RGB').crop((0, 48, 960, 648)), (16, 60))
    d = ImageDraw.Draw(im)
    d.rectangle((14, 58, 977, 661), outline=mint, width=2)
    text(d, (1006, 58), heading, 28, mint)
    d.line((1006, 112, 1245, 112), fill=gold, width=2)
    for i, line in enumerate(lines):
        text(d, (1006, 145 + i * 47), line, 26)
    text(d, (1006, 540), note, 22, dim)
    text(d, (1006, 650), V, 22, dim)
    f = T / (name + '.png')
    im.save(f)
    return f


def title(name, picture, heading, body):
    # 只取插畫範圍（原生第8到144列），不帶入下方的故事字列與版權字。
    im = ImageOps.fit(Image.open(picture).convert('RGB').crop((0, 72, 960, 480)), (1280, 720)).convert('RGBA')
    im = Image.alpha_composite(im, Image.new('RGBA', im.size, (12, 24, 29, 215))).convert('RGB')
    d = ImageDraw.Draw(im)
    d.rectangle((42, 42, 1238, 678), outline=mint, width=2)
    d.line((80, 124, 280, 124), fill=gold, width=4)
    text(d, (82, 205), heading, 68)
    text(d, (84, 335), body, 32)
    text(d, (84, 540), '配樂：原版標題曲（AdLib），非同期聲', 24, dim)
    text(d, (84, 595), 'ZELIARD  /  ' + V, 24, mint)
    f = T / (name + '.png')
    im.save(f)
    return f


segments = []
NOTE = '1.0.8正式執行檔\n合成畫面'
ENDING_NOTE = '結局以探針結構觸發\n1.0.8正式執行檔合成畫面'


def still(name, pictures, seconds, caption, sequence=False):
    """pictures依序各佔相同時間；單張即靜態畫面。"""
    out = T / (name + '.mp4')
    if len(pictures) == 1:
        source = ['-loop', '1', '-i', str(pictures[0])]
    else:
        # concat讀圖片序列；最後一格以tpad延長，再精確裁到指定秒數。
        listing = T / (name + '.txt')
        each = seconds / len(pictures)
        listing.write_text(''.join(f"file '{p}'\nduration {each}\n" for p in pictures) + f"file '{pictures[-1]}'\n")
        source = ['-f', 'concat', '-safe', '0', '-i', str(listing)]
    run(['ffmpeg', '-nostdin', '-hide_banner', '-loglevel', 'error', *source, '-t', str(seconds), '-r', '30', '-an',
         '-vf', 'tpad=stop_mode=clone:stop_duration=1,format=yuv420p', '-t', str(seconds), '-c:v', 'libx264', '-preset', 'veryfast', '-crf', '20', '-threads', '1', str(out)])
    segments.append({'file': str(out.relative_to(P)), 'kind': 'intentional still', 'seconds': seconds, 'caption': caption,
                     'frames': [str(p.relative_to(P)) for p in pictures], 'frame_sha256': [h(p) for p in pictures],
                     'sequence': 'original frame order; spacing not original speed' if sequence else None})


def scene(name, run_name, steps, seconds, heading, lines, note=NOTE):
    pictures = [card(f'{name}-{n}', frame(run_name, s), heading, lines, note) for n, s in enumerate(steps)]
    still(name, pictures, seconds, heading, len(steps) > 1)


hero = frame('story', 100_000_000)
still('00-title', [title('title', hero, '神劍封魔', '片頭與結局插畫\n依原版色盤重畫')], 3, '開場')
scene('01-threat', 'boot', [167_600_000, 168_000_000, 169_000_000, 172_000_000], 4, '魔王宣言', ['藍色線條、紅眼', '依實機點陣重畫', '說話嘴型維持HD'])
scene('02-balcony', 'story', [100_000_000], 6, '陽台公主', ['紅色禮服', '綠色山丘', '雨中陽台'])
scene('03-princess', 'story', [430_000_000], 3, '公主特寫', ['金髮、洋紅衣裳'])
scene('04-statue', 'story', [620_000_000], 3, '化為石像', ['灰白石像', '深藍夜色'])
scene('05-king', 'story', [690_000_000], 3, '國王', ['原版色盤', '紅綠不再互換'])
scene('06-sei', 'story', [760_000_000], 3, '守護神靈', ['青色頭髮', '綠色披風', '點陣以實機為準'])
scene('07-garland', 'story', [1_110_000_000], 3, '加蘭公爵', ['紅髮紅衣'])
scene('08-king-talk', 'story', [1_183_000_000, 1_184_000_000, 1_185_000_000, 1_186_000_000, 1_187_000_000], 3, '國王說話', ['說話時嘴型維持HD'])
scene('09-knight', 'story', [1_370_000_000], 3, '黑衣騎士', ['紅色披風'])
scene('10-door', 'story', [1_704_600_000], 2, '命運之門', ['門後的光'])
scene('11-silhouette', 'story', [1_795_000_000], 4, '踏入迷宮', ['灰光中的剪影', '新增HD畫格'])
for n, (step, heading, lines) in enumerate([(116_000_000, '凱旋', ['主角與公主', '公主人像補上HD']), (266_000_000, '父女重逢', ['國王擁抱公主']),
                                             (386_000_000, '守護神靈', ['再次現身']), (656_000_000, '加蘭離城', ['紅髮紅衣']),
                                             (1_016_000_000, '製作名單', ['公主與騎士']), (1_286_000_000, 'FIN', ['結局最後一張'])]):
    scene(f'12-ending-{n}', 'ending', [step], 2.5, heading, lines, ENDING_NOTE)
still('13-end', [title('end', frame('ending', 1_286_000_000), '神劍封魔', '原版＋中文   ／   原版＋中文＋HD\nLinux AppImage・Windows・macOS\n片頭與結局插畫重畫')], 5, '雙版本三平台')
assert abs(sum(s['seconds'] for s in segments) - 60) < 1e-9, sum(s['seconds'] for s in segments)
listing = T / 'list.txt'
listing.write_text(''.join(f"file '{P / s['file']}'\n" for s in segments))
silent = T / 'silent.mp4'
run(['ffmpeg', '-nostdin', '-hide_banner', '-loglevel', 'error', '-f', 'concat', '-safe', '0', '-i', str(listing), '-c', 'copy', str(silent)])
# 配樂：沿065的原版標題曲錄音，從第一個有聲的秒取28秒循環。
music = P / 'workplace/music-release-v1/title-music/source.wav'
assert h(music) == json.loads((music.parent / 'receipt.json').read_text())['sha256']
with wave.open(str(music), 'rb') as f:
    assert (f.getnchannels(), f.getsampwidth(), f.getframerate()) == (1, 2, 44100)
    data = array.array('h', f.readframes(f.getnframes()))
onset = next(i for i in range(len(data) // 44100) if sum(v * v for v in data[i * 44100:(i + 1) * 44100]) / 44100 > 1000 ** 2)
excerpt = T / 'original-title.wav'
run(['ffmpeg', '-nostdin', '-hide_banner', '-loglevel', 'error', '-ss', str(onset), '-i', str(music), '-t', '28', '-af', 'afade=t=in:d=0.025,afade=t=out:st=27.975:d=0.025', str(excerpt)])
promo = D / f'Zeliard-{V}-promo.mp4'
run(['ffmpeg', '-nostdin', '-hide_banner', '-loglevel', 'error', '-i', str(silent), '-stream_loop', '-1', '-i', str(excerpt), '-filter_complex_threads', '1', '-filter_complex',
     '[1:a]atrim=duration=60,loudnorm=I=-18:TP=-2:LRA=11,afade=t=in:d=1,afade=t=out:st=58:d=2[a]', '-map', '0:v', '-map', '[a]', '-t', '60',
     '-c:v', 'copy', '-c:a', 'aac', '-ac', '2', '-ar', '44100', '-b:a', '192k', '-movflags', '+faststart', str(promo)])
write(D / 'provenance.json', {'version': V, 'capture_machine': 'local', 'video_sha256': h(promo), 'segments': segments, 'frames': frames,
                              'source': 'formal 1.0.8 binary flicker probe composites (same compose path as the window); window recording ran at 0.08x under host load and was not used',
                              'source_binary_sha256': binary, 'package': package.name, 'package_sha256': h(package), 'complete_art_sha256': h(BASE / 'assets/manual-hd/complete-art.json'),
                              'music_file': str(music.relative_to(P)), 'music_sha256': h(music), 'music_onset_seconds': onset, 'music_excerpt_sha256': h(excerpt),
                              'music': 'original MSCADLIB title music, recorded by tools/capture_music.go; looped excerpt, not synchronous',
                              'subtitle_bounds_checked': len(bounds), 'tool_sha256': h(Path(__file__)), 'private_local_only': True})
print('CREATED_V108_PICTURE_PROMO', flush=True)

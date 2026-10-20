#!/usr/bin/env python3
"""在 Docker 內從本機證據製作 README 縮圖，不改動原始素材。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont, ImageOps


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--project", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--font", type=Path, required=True)
    args = parser.parse_args()
    root, out = args.project, args.output
    if not out.is_dir():
        parser.error("輸出目錄必須先建立")
    fonts = {n: ImageFont.truetype(str(args.font), n, index=3) for n in (18, 22, 26, 36)}
    sources = {}
    outputs = {}
    bg, panel, ink, muted, accent = "#101923", "#1c2936", "#f3eee1", "#b8c4cd", "#e3b66a"
    assets = "workplace/music-release-v1/release-build/data/assets/"
    cave = "workplace/cave-hd-v1/"
    town = "workplace/town-hd-v1/foreground-v1/"
    gui = "workplace/music-release-v1/gui-v4/boot1/"
    town_pair = cave + "gui-current-return-v6/"
    church_pair = "workplace/prose-v1/services-v1/hd-gui-v2/boot1/"

    def digest(path):
        return hashlib.sha256(path.read_bytes()).hexdigest()

    def read(name, crop=None, palette=None):
        path = root / name
        entry = {"sha256": digest(path)}
        if palette:
            pal = root / palette
            raw = path.read_bytes()
            assert len(raw) == 64000 and len(pal.read_bytes()) == 768
            im = Image.frombytes("P", (320, 200), raw)
            im.putpalette(pal.read_bytes())
            sources.setdefault(palette, [{"sha256": digest(pal)}])
        else:
            im = Image.open(path)
        im = im.convert("RGBA")
        if crop:
            assert 0 <= crop[0] < crop[2] <= im.width
            assert 0 <= crop[1] < crop[3] <= im.height
            im = im.crop(crop)
            entry["crop"] = list(crop)
        if entry not in sources.setdefault(name, []):
            sources[name].append(entry)
        return im

    def place(canvas, im, box, pixel=False):
        x, y, w, h = box
        im = ImageOps.contain(im, (w, h), Image.Resampling.NEAREST if pixel else Image.Resampling.LANCZOS)
        canvas.paste(im, (x + (w-im.width)//2, y + (h-im.height)//2), im)

    def save(im, name):
        path = out / name
        im.save(path, quality=91, subsampling=0, optimize=True)
        outputs[name] = {"sha256": digest(path), "width": im.width, "height": im.height, "bytes": path.stat().st_size}

    # 實際視窗的遊戲區。原貌仍保留中文覆繪，不能標成英文原版。
    pairs = [
        ("Muralla 城鎮", town_pair+"returned-town-original.png", town_pair+"returned-town-HD.png"),
        ("教堂對話", church_pair+"church-frame-1-original.png", church_pair+"church-frame-1-hd.png"),
    ]
    im = Image.new("RGB", (1600, 1240), bg)
    d = ImageDraw.Draw(im)
    d.text((32, 18), "原貌與 HD｜實際視窗對照", font=fonts[36], fill=ink)
    d.text((32, 72), "同一暫停狀態切換顯示；兩側皆保留繁體中文覆繪。", font=fonts[22], fill=muted)
    for row, (label, original, hd) in enumerate(pairs):
        y = 122 + row*540
        for col, (name, source) in enumerate((("原貌", original), ("手繪 HD", hd))):
            x = 24 + col*788
            d.text((x+8, y), label+" · "+name, font=fonts[26], fill=accent)
            place(im, read(source, (0, 48, 960, 648)), (x, y+44, 764, 478), True)
    d.text((32, 1208), "既有 GUI 驗證截圖，裁去工具列與頁尾；僅展示上述場景。", font=fonts[18], fill=muted)
    save(im, "original-hd-comparison.jpg")

    im = Image.new("RGB", (1600, 704), bg)
    d = ImageDraw.Draw(im)
    d.text((32, 18), "目前封包｜城鎮與第一洞窟", font=fonts[36], fill=ink)
    for col, (label, name) in enumerate((("Muralla 城鎮", "town-HD.png"), ("Malicia 第一洞窟", "first-cave-HD.png"))):
        x = 24 + col*788
        d.text((x+8, 82), label, font=fonts[26], fill=accent)
        place(im, read(gui+name), (x, 128, 764, 542))
    save(im, "gameplay-hd.jpg")

    # 每格一類素材；圖集只抽樣，完整執行期原圖不加入 Git。
    entries = []
    def portrait(label, reference, palette, art, box=(56,23,152,87)):
        entries.append((label, read(assets+reference, box, assets+palette), read(assets+art)))

    portrait("國王", "king/probe-king-memory-84m/screen.bin", "king/probe-king-memory-84m/palette.bin", "king/hd-king-prototype.png", (112,23,208,87))
    portrait("長老", "town/sage-a.indexed", "town/sage.palette", "town/sage-handpaint.png")
    portrait("武器商人", "merchant/greeting.indexed", "merchant/greeting.palette", "merchant/merchant-prototype.png")
    portrait("銀行", "bank/bank-greeting.indexed", "bank/bank-greeting.palette", "bank/bank-raised-prototype.png")
    portrait("道具店", "item/item-greeting.indexed", "item/item-greeting.palette", "item/item-shop-base-prototype.png")
    portrait("教堂", "church/reference/church-entered.indexed", "church/reference/church-entered.palette", "church/art/church-rest-handpaint-v1.png", (112,23,208,87))
    for label, original, art in [
        ("城鎮遠山", town+"background-native.png", "town/background-handpaint-v2.png"),
        ("城鎮建築與前景", town+"muralla-foreground-native.png", "town/scenery-art/town-middle-art-v2.png"),
        ("主角姿態圖集", cave+"player-source-v1/reference-atlas.png", "cave/player-atlas-handpaint-v1.png"),
        ("武器動作圖集", cave+"weapon-inventory-v1/source-sheet.png", "cave/weapon-atlas-handpaint-v1.png"),
        ("綠色物件圖集", cave+"object-art-reference-v1/green-source-sheet.png", "cave/object-green-handpaint-v1.png"),
        ("紅色物件圖集", cave+"object-art-reference-v1/red-source-sheet.png", "cave/object-red-handpaint-v1.png"),
        ("效果圖集", cave+"object-art-reference-v1/effects-source-sheet.png", "cave/object-effects-handpaint-v1.png"),
        ("洞窟出入口", cave+"portal-source-v1/reference.png", "cave/portal-handpaint-v1.png"),
    ]:
        entries.append((label, read(original), read(assets+art)))
    entries.append(("城鎮地面", read(town_pair+"returned-town-original.png", (144,486,816,528)), read(assets+"town/ground-handpaint-v1.png")))
    entries.append(("洞窟岩壁", read(cave+"gui-current-movement-v7/entry-terrain-original-a.png", (144,96,816,222)), read(assets+"cave/rock-material-v1.png")))
    for variant, name in ((0, "original-assets-overview.jpg"), (1, "hd-assets-overview.jpg")):
        im = Image.new("RGB", (1600, 1420), bg)
        d = ImageDraw.Draw(im)
        d.text((32, 18), ("原版素材總覽" if variant == 0 else "HD 素材總覽")+"｜目前接入範圍抽樣", font=fonts[36], fill=ink)
        d.text((32, 76), "依相同類別排列；人物取一姿態，圖集與場景以縮圖展示。", font=fonts[22], fill=muted)
        for i, (label, original, hd) in enumerate(entries):
            x, y = 24+(i%4)*394, 126+(i//4)*312
            d.rounded_rectangle((x,y,x+370,y+292), radius=10, fill=panel)
            d.text((x+14,y+10), f"{i+1:02d}  {label}", font=fonts[22], fill=accent)
            place(im, hd if variant else original, (x+12,y+50,346,228), not variant)
        d.text((32, 1390), "本圖為文件預覽，非全遊戲資產清單；原版及衍生素材不納入自有程式授權。", font=fonts[18], fill=muted)
        save(im, name)

    receipts = [town_pair+"receipt.json", "workplace/prose-v1/services-v1/hd-gui-v2/receipt.json", "workplace/music-release-v1/gui-v4/receipt.json"]
    manifest = {
        "scope": "使用者指定的 README 圖片；僅文件縮圖，不包含原始遊戲或完整解析度素材",
        "font": {"name": args.font.name, "sha256": digest(args.font), "ttc_index": 3},
        "tool": {"script_sha256": digest(Path(__file__)), "pillow": Image.__version__},
        "comparison_receipts": [town_pair+"receipt.json", "workplace/prose-v1/services-v1/hd-gui-v2/receipt.json"],
        "current_screenshots_receipt": "workplace/music-release-v1/gui-v4/receipt.json",
        "receipt_sha256": {name: digest(root/name) for name in receipts},
        "sources": sources, "outputs": outputs,
    }
    (out/"provenance.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2)+"\n")
    print(json.dumps(outputs, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()

package overlay

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

// 輸入來自手冊與正常新局探索；不載入來源存檔、不寫入座標或記憶體。
func TestSageMenuNormalPathReadOnly(t *testing.T) {
	root, fontPath, output := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_LABEL_OUTPUT")
	if root == "" || fontPath == "" || output == "" {
		t.Skip("須掛載原版、字型與選單收據輸出目錄")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	load := func() *oracle.Oracle {
		o, e := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if e != nil {
			t.Fatal(e)
		}
		o.SetDOSBoxCycles(3000)
		o.SetScratch(t.TempDir())
		t.Cleanup(func() { o.Close() })
		return o
	}
	a, b := load(), load()
	d, err := NewDialogue(b, filepath.Join(root, "ZELRES2.SAR"), "../../../text/dialogue-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	l, err := NewLabels(b, filepath.Join(root, "ZELRES2.SAR"), "../../../text/labels-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(output, 0755); err != nil {
		t.Fatal(err)
	}
	type event struct {
		step uint64
		scan uint8
		down bool
	}
	var events []event
	for _, s := range []uint64{12000000, 20000000, 28000000, 36000000, 44000000, 52000000, 86000000, 94000000, 102000000, 110000000, 118000000, 126000000} {
		events = append(events, event{s, 0x39, true}, event{s + 50000, 0x39, false})
	}
	events = append(events, event{60000000, 0x4d, true}, event{62000000, 0x4d, false}, event{64000000, 0x48, true}, event{64250000, 0x48, false})
	for _, k := range []struct {
		start, end uint64
		scan       uint8
	}{
		{130000000, 140000000, 0x4d}, {142000000, 144000000, 0x4d}, {144000000, 144250000, 0x48},
		{154250000, 156750000, 0x4b}, {156750000, 157000000, 0x48}, {167000000, 168000000, 0x4d}, {168000000, 168250000, 0x48},
		{188250000, 188300000, 0x39}, {198300000, 210300000, 0x4d}, {211300000, 219300000, 0x4d},
		{220300000, 228300000, 0x4d}, {229300000, 237300000, 0x4d}, {238300000, 244300000, 0x4d},
		{245300000, 251300000, 0x4d}, {252300000, 253600000, 0x4d}, {253600000, 253850000, 0x48},
		{273850000, 273900000, 0x39},
	} {
		events = append(events, event{k.start, k.scan, true}, event{k.end, k.scan, false})
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].step < events[j].step })
	for _, e := range events {
		for _, o := range []*oracle.Oracle{a, b} {
			if err = o.Run(e.step - o.Steps()); err != nil {
				t.Fatal(err)
			}
			if e.down {
				o.KeyDown(e.scan)
			} else {
				o.KeyUp(e.scan)
			}
		}
	}
	var receipts []map[string]any
	for _, target := range []uint64{283900000, 284650000} {
		if target == 284650000 {
			for i := 0; i < 3; i++ {
				for _, o := range []*oracle.Oracle{a, b} {
					o.KeyDown(0x50)
					if err = o.Run(50000); err != nil {
						t.Fatal(err)
					}
					o.KeyUp(0x50)
					if err = o.Run(200000); err != nil {
						t.Fatal(err)
					}
				}
			}
		}
		for _, o := range []*oracle.Oracle{a, b} {
			if err = o.Run(target - o.Steps()); err != nil {
				t.Fatal(err)
			}
		}
		raw := b.Indexed()
		pal := b.Palette()
		mem := b.Bytes(oracle.Far(0, 0), 1<<20)
		if !bytes.Equal(a.Indexed(), raw) || a.Palette() != pal || a.Cycles() != b.Cycles() || a.IP() != b.IP() || !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), mem) {
			t.Fatal("選單觀察改寫原版狀態")
		}
		_, _, rgb := b.ScreenRGB()
		dst := image.NewRGBA(image.Rect(0, 0, 960, 600))
		original := image.NewRGBA(dst.Bounds())
		for y := 0; y < 600; y++ {
			for x := 0; x < 960; x++ {
				i := (y/3*320 + x/3) * 3
				c := color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255}
				dst.SetRGBA(x, y, c)
				original.SetRGBA(x, y, c)
			}
		}
		n := l.Draw(dst, raw, pal, image.Point{})
		if n != 4 {
			t.Fatalf("長老四列選單未覆繪：drawn=%d completed=%d stamps=%d", n, l.Completed, len(l.stamps))
		}
		dialogueRendered := d.Draw(dst, raw, pal, image.Point{})
		if !dialogueRendered || d.current < 0 || d.paragraphs[d.current].key != "zelres2.file.01ce7e" || d.visible != len(d.paragraphs[d.current].text) || len(d.Rejections) != 0 {
			t.Fatalf("長老初訪中文未完整顯示：rendered=%v current=%d glyphs=%d rejects=%v", dialogueRendered, d.current, d.Glyphs, d.Rejections)
		}
		positions, e := dialogueLayout(font, d.paragraphs[d.current].text)
		if e != nil {
			t.Fatal(e)
		}
		for i, ch := range d.paragraphs[d.current].text {
			if positions[i].X == 6 && (ch == '，' || ch == '。') {
				t.Fatal("中文標點落在行首")
			}
		}
		inside, outside := 0, 0
		for y := 0; y < 600; y++ {
			for x := 0; x < 960; x++ {
				if dst.RGBAAt(x, y) == original.RGBAAt(x, y) {
					continue
				}
				in := dialogueRendered && image.Pt(x/3, y/3).In(dialogueBox)
				for _, s := range l.stamps {
					if image.Pt(x/3, y/3).In(s.box) {
						in = true
					}
				}
				if in {
					inside++
				} else {
					outside++
				}
			}
		}
		if inside == 0 || outside != 0 || !bytes.Equal(raw, b.Indexed()) || !bytes.Equal(mem, b.Bytes(oracle.Far(0, 0), 1<<20)) {
			t.Fatal("標籤未顯示、越界或改寫原版")
		}
		fp, e := os.Create(filepath.Join(output, fmtStep(target)+".png"))
		if e != nil {
			t.Fatal(e)
		}
		e = png.Encode(fp, dst)
		fp.Close()
		if e != nil {
			t.Fatal(e)
		}
		var boxes []image.Rectangle
		for _, s := range l.stamps {
			boxes = append(boxes, s.box)
		}
		receipts = append(receipts, map[string]any{"steps": target, "cycles": b.Cycles(), "raw_sha256": fmtHash(raw), "memory_sha256": fmtHash(mem), "state_equal": true, "labels": n, "completed": l.Completed, "boxes": boxes, "inside_changed": inside, "outside_changed": outside, "dialogue_glyphs": d.Glyphs, "dialogue_paragraphs": d.Paragraphs, "dialogue_key": d.paragraphs[d.current].key, "dialogue_rejections": d.Rejections})
	}
	// 失配或離開後不可把舊選單蓋到另一個畫面。
	altered := append([]byte(nil), b.Indexed()...)
	for _, s := range l.stamps {
		altered[s.box.Min.Y*320+s.box.Min.X] ^= 1
	}
	if n := l.Draw(image.NewRGBA(image.Rect(0, 0, 960, 600)), altered, b.Palette(), image.Point{}); n != 0 {
		t.Fatal("失配後仍殘留選單")
	}
	data, _ := json.MarshalIndent(receipts, "", "  ")
	if err = os.WriteFile(filepath.Join(output, "receipt.json"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
}

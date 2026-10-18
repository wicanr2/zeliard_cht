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

func TestKingDialogueNormalPathReadOnly(t *testing.T) {
	root, fontPath, output := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_DIALOGUE_OUTPUT")
	if root == "" || fontPath == "" || output == "" {
		t.Skip("須掛載原版、字型與對話收據輸出目錄")
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
	if err = os.MkdirAll(output, 0755); err != nil {
		t.Fatal(err)
	}
	type event struct {
		step    uint64
		scan    uint8
		down    bool
		capture bool
	}
	var events []event
	for _, step := range []uint64{12000000, 20000000, 28000000, 36000000, 44000000, 52000000, 86000000, 94000000, 102000000, 110000000, 118000000, 126000000} {
		events = append(events, event{step, 0x39, true, false}, event{step + 50000, 0x39, false, false})
	}
	events = append(events, event{60000000, 0x4d, true, false}, event{62000000, 0x4d, false, false}, event{64000000, 0x48, true, false}, event{64250000, 0x48, false, false})
	for _, step := range []uint64{70000000, 74000000, 80000000, 84000000, 88000000, 96000000, 104000000, 106000000, 130000000} {
		events = append(events, event{step, 0, false, true})
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].step < events[j].step })
	var previous uint64
	var receipts []map[string]any
	for _, e := range events {
		if err = a.Run(e.step - previous); err != nil {
			t.Fatal(err)
		}
		if err = b.Run(e.step - previous); err != nil {
			t.Fatal(err)
		}
		previous = e.step
		if !e.capture {
			if e.down {
				a.KeyDown(e.scan)
				b.KeyDown(e.scan)
			} else {
				a.KeyUp(e.scan)
				b.KeyUp(e.scan)
			}
			continue
		}
		raw := b.Indexed()
		if !bytes.Equal(a.Indexed(), raw) || a.Palette() != b.Palette() || a.Cycles() != b.Cycles() || a.IP() != b.IP() || !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) {
			t.Fatalf("%d：對話覆繪觀察改寫原版狀態", e.step)
		}
		_, _, rgb := b.ScreenRGB()
		dst := image.NewRGBA(image.Rect(0, 0, 960, 600))
		orig := image.NewRGBA(dst.Bounds())
		for y := 0; y < 600; y++ {
			for x := 0; x < 960; x++ {
				i := (y/3*320 + x/3) * 3
				c := color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255}
				dst.SetRGBA(x, y, c)
				orig.SetRGBA(x, y, c)
			}
		}
		rendered := d.Draw(dst, raw, b.Palette(), image.Point{})
		inside, outside := 0, 0
		for y := 0; y < 600; y++ {
			for x := 0; x < 960; x++ {
				if dst.RGBAAt(x, y) != orig.RGBAAt(x, y) {
					if image.Pt(x/3, y/3).In(dialogueBox) {
						inside++
					} else {
						outside++
					}
				}
			}
		}
		if outside != 0 || !bytes.Equal(raw, b.Indexed()) {
			t.Fatal("顯示覆繪越界或改寫原版")
		}
		if e.step == 84000000 && (!rendered || d.current != 0 || d.visible != len(d.paragraphs[0].text)) {
			t.Fatalf("國王首段未完整覆繪：%d glyphs, %d visible", d.Glyphs, d.visible)
		}
		if e.step == 96000000 && d.Paragraphs < 2 {
			t.Fatal("第二段未走到")
		}
		f, er := os.Create(filepath.Join(output, fmtStep(e.step)+".png"))
		if er != nil {
			t.Fatal(er)
		}
		er = png.Encode(f, dst)
		f.Close()
		if er != nil {
			t.Fatal(er)
		}
		receipts = append(receipts, map[string]any{"steps": e.step, "cycles": b.Cycles(), "raw_sha256": fmtHash(raw), "memory_sha256": fmtHash(b.Bytes(oracle.Far(0, 0), 1<<20)), "gold_bytes": b.Bytes(oracle.Far(0x01b1, 0x85), 3), "state_equal": true, "rendered": rendered, "glyphs": d.Glyphs, "paragraphs": d.Paragraphs, "scrolls": d.Scrolls, "visible": d.visible, "prompt": d.prompt, "inside_changed": inside, "outside_changed": outside, "source_pointer": b.Word(oracle.Far(0x01b1, 0xff4c))})
	}
	if d.active {
		t.Fatal("離開國王對話後仍留字")
	}
	data, _ := json.MarshalIndent(receipts, "", "  ")
	if err = os.WriteFile(filepath.Join(output, "receipt.json"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
}

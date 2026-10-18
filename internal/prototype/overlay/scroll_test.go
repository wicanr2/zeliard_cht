package overlay

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

// 跑原版的自然片頭，對照唯讀掛鉤前後，不以合成字串事件取代正常執行。
func TestScrollingIntroOriginalState(t *testing.T) {
	root, fontPath, output := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_SCROLL_OUTPUT")
	if root == "" || fontPath == "" || output == "" {
		t.Skip("須掛載本地原版、字型與收據目錄")
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
	baseline, translated := load(), load()
	s, err := NewScroll(translated, filepath.Join(root, "ZELRES1.SAR"), "../../../text/intro-scroll-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(output, 0755); err != nil {
		t.Fatal(err)
	}
	var receipts []map[string]any
	var previous uint64
	for _, target := range []uint64{10729940, 14000000, 30000000, 65000000, 110000000, 160000000} {
		if err = baseline.Run(target - previous); err != nil {
			t.Fatal(err)
		}
		if err = translated.Run(target - previous); err != nil {
			t.Fatal(err)
		}
		previous = target
		a, b := baseline.Indexed(), translated.Indexed()
		if !bytes.Equal(a, b) || baseline.Palette() != translated.Palette() || baseline.Cycles() != translated.Cycles() || baseline.IP() != translated.IP() || !bytes.Equal(baseline.Bytes(oracle.Far(0, 0), 1<<20), translated.Bytes(oracle.Far(0, 0), 1<<20)) {
			t.Fatalf("%d：覆繪觀察改變了原版狀態", target)
		}
		_, _, rgb := translated.ScreenRGB()
		dst := image.NewRGBA(image.Rect(0, 0, 960, 600))
		raw := image.NewRGBA(dst.Bounds())
		for y := 0; y < 600; y++ {
			for x := 0; x < 960; x++ {
				i := (y/3*320 + x/3) * 3
				c := color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255}
				dst.SetRGBA(x, y, c)
				raw.SetRGBA(x, y, c)
			}
		}
		before := append([]byte(nil), b...)
		rendered := s.Draw(dst, b, translated.Palette(), image.Point{})
		if !bytes.Equal(before, translated.Indexed()) {
			t.Fatal("Draw 改寫色號畫格")
		}
		inside, outside := 0, 0
		for y := 0; y < 600; y++ {
			for x := 0; x < 960; x++ {
				if dst.RGBAAt(x, y) != raw.RGBAAt(x, y) {
					if y >= 96 && y < 456 {
						inside++
					} else {
						outside++
					}
				}
			}
		}
		if outside != 0 {
			t.Fatal("片頭覆繪越界", outside)
		}
		if (target == 30000000 || target == 110000000) && (!rendered || len(s.lines) < 2 || inside == 0) {
			t.Fatalf("未顯示多行中文：%d rows %d lines", s.Rows, len(s.lines))
		}
		name := filepath.Join(output, fmtStep(target)+".png")
		f, e := os.Create(name)
		if e != nil {
			t.Fatal(e)
		}
		e = png.Encode(f, dst)
		f.Close()
		if e != nil {
			t.Fatal(e)
		}
		receipts = append(receipts, map[string]any{"steps": target, "cycles": translated.Cycles(), "raw_sha256": fmtHash(b), "full_memory_sha256": fmtHash(translated.Bytes(oracle.Far(0, 0), 1<<20)), "source_events": s.Sources, "completed_rows": s.Rows, "visible_lines": len(s.lines), "rendered": rendered, "pending": s.pending, "inside_changed": inside, "outside_changed": outside, "state_equal": true})
	}
	if s.Sources != 30 {
		t.Fatalf("首章來源行數不符：%d", s.Sources)
	}
	if s.active || len(s.lines) != 0 {
		t.Fatal("離開首章後仍留中文")
	}
	data, _ := json.MarshalIndent(receipts, "", "  ")
	if err = os.WriteFile(filepath.Join(output, "receipt.json"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
}

func fmtHash(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }
func fmtStep(step uint64) string { return fmt.Sprintf("%09d", step) }

package overlay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

func TestIntroThreatOriginal(t *testing.T) {
	root, fontPath, out := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_THREAT_OUTPUT")
	if root == "" || fontPath == "" || out == "" {
		t.Skip("須掛載原版、字型及獨立收據目錄")
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
		t.Cleanup(o.Close)
		return o
	}
	a, b := load(), load()
	s, err := NewIntroThreat(b, filepath.Join(root, "ZELRES1.SAR"), "../../../text/intro-threat-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(out, 0755); err != nil {
		t.Fatal(err)
	}
	var records []map[string]any
	compare := func() {
		t.Helper()
		if err := a.Run(b.Steps() - a.Steps()); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) || !reflect.DeepEqual(a.Regs(), b.Regs()) || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) {
			t.Fatal("宣言覆繪改變原版狀態", b.Steps())
		}
	}
	draw := func() (*image.RGBA, bool, int) {
		_, _, rgb := b.ScreenRGB()
		dst := image.NewRGBA(image.Rect(0, 0, 960, 600))
		for y := 0; y < 600; y++ {
			for x := 0; x < 960; x++ {
				i := (y/3*320 + x/3) * 3
				dst.SetRGBA(x, y, color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255})
			}
		}
		before := append([]byte(nil), dst.Pix...)
		raw := b.Indexed()
		rendered := s.Draw(dst, raw, b.Palette(), image.Point{})
		inside := 0
		for y := 0; y < 600; y++ {
			for x := 0; x < 960; x++ {
				i := dst.PixOffset(x, y)
				if !bytes.Equal(before[i:i+4], dst.Pix[i:i+4]) {
					if y < 444 || y >= 531 {
						t.Fatal("宣言覆繪越出安全區")
					}
					inside++
				}
			}
		}
		if !bytes.Equal(raw, b.Indexed()) {
			t.Fatal("宣言Draw改原版色號")
		}
		return dst, rendered, inside
	}
	checks := 0
	pendingChecks := 0
	b.OnCall(oracle.Far(0x1b1, 0x6336), func(o *oracle.Oracle) {
		if !s.pending {
			return
		}
		_, rendered, _ := draw()
		if !rendered {
			t.Fatal("陰影階段失去中文", s.State())
		}
		pendingChecks++
	})
	b.OnCall(oracle.Far(0x1b1, 0x6346), func(o *oracle.Oracle) {
		if !s.active {
			t.Fatal("已觀察宣言被拒絕", s.State())
		}
		compare()
		dst, rendered, inside := draw()
		if !rendered {
			t.Fatal("字形完成後未覆繪")
		}
		checks++
		records = append(records, map[string]any{"steps": b.Steps(), "cycles": b.Cycles(), "glyphs": s.Glyphs, "state_equal": true, "outside_differences": 0, "inside_changed": inside, "memory_sha256": fmtHash(b.Bytes(oracle.Far(0, 0), 1<<20))})
		if s.done == 24 || s.done == 52 || s.done == 88 {
			f, e := os.Create(filepath.Join(out, fmt.Sprintf("glyph-%02d.png", s.done)))
			if e != nil {
				t.Fatal(e)
			}
			if e = png.Encode(f, dst); e != nil {
				t.Fatal(e)
			}
			if e = f.Close(); e != nil {
				t.Fatal(e)
			}
		}
		if s.done == 88 {
			for _, kind := range []string{"unknown-pixel", "short-frame", "small-canvas"} {
				clone := *s
				raw := append([]byte(nil), b.Indexed()...)
				canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
				if kind == "unknown-pixel" {
					raw[150*320+2] = 255
				}
				if kind == "short-frame" {
					raw = raw[:100]
				}
				if kind == "small-canvas" {
					canvas = image.NewRGBA(image.Rect(0, 0, 10, 10))
				}
				before := append([]byte(nil), canvas.Pix...)
				if clone.Draw(canvas, raw, b.Palette(), image.Point{}) || !bytes.Equal(before, canvas.Pix) {
					t.Fatal("未知畫面沒有原子退回", kind)
				}
			}
		}
	})
	// 原版200m收據已進標題；185m仍在魔王宣言，不能作撤除斷言。
	if err = b.Run(200000000); err != nil {
		t.Fatal(err)
	}
	compare()
	_, rendered, inside := draw()
	if rendered || inside != 0 || s.active || s.pending {
		t.Fatal("進標題後仍有中文")
	}
	if checks != 88 || pendingChecks != 88 || s.Glyphs != 88 {
		t.Fatal("宣言字形／途中檢查不足", checks, pendingChecks, s.Glyphs)
	}
	j := map[string]any{"status": "PASS_SCOPED_THREAT_ORIGINAL_PAIR", "points": records, "completed_glyphs": checks, "interrupted_shadow_checks": pendingChecks, "title_removed": true, "final_steps": b.Steps(), "full_original_state_equal": true, "input": "none; cold startup", "state_injection": false}
	data, e := json.MarshalIndent(j, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, 10), 0644); e != nil {
		t.Fatal(e)
	}
}

func TestIntroThreatCatalog(t *testing.T) {
	root := os.Getenv("ZELIARD_ORIG")
	if root == "" {
		t.Skip("須掛載本地原版")
	}
	original, err := os.ReadFile("../../../text/intro-threat-prototype.tsv")
	if err != nil {
		t.Fatal(err)
	}
	font := bitmap.Font{}
	for _, ch := range string(original) {
		font[ch] = make([]byte, 32)
	}
	o, err := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
	if err != nil {
		t.Fatal(err)
	}
	defer o.Close()
	base := string(original)
	for _, tc := range []struct {
		name    string
		data    []byte
		missing bool
	}{{"duplicate", []byte(strings.ReplaceAll(base, "003163", "00313e")), false}, {"missing", []byte(strings.Join(strings.Split(base, "\n")[:3], "\n")), false}, {"empty", []byte(strings.Replace(base, "當心吧，我將甦醒，", "", 1)), false}, {"overflow", []byte(strings.Replace(base, "當心吧，我將甦醒，", strings.Repeat("當", 50), 1)), false}, {"control", []byte(strings.Replace(base, "當心吧，我將甦醒，", "當\x01", 1)), false}, {"utf8", append(append([]byte(nil), original...), 255), false}, {"missing-glyph", original, true}} {
		t.Run(tc.name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "text.tsv")
			if e := os.WriteFile(p, tc.data, 0644); e != nil {
				t.Fatal(e)
			}
			f := font
			if tc.missing {
				f = bitmap.Font{}
			}
			if _, e := NewIntroThreat(o, filepath.Join(root, "ZELRES1.SAR"), p, f); e == nil {
				t.Fatal("錯誤譯表未被拒絕")
			}
		})
	}
}

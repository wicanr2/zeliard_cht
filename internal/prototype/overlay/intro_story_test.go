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

func TestIntroStoryIndependentLayout(t *testing.T) {
	root, fontPath, layout := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_STORY_LAYOUT")
	if root == "" || fontPath == "" || layout == "" {
		t.Skip("須掛載原版、字型與獨立排版收據")
	}
	font, e := bitmap.Load(fontPath)
	if e != nil {
		t.Fatal(e)
	}
	data, e := os.ReadFile(filepath.Join(root, "ZELRES1.SAR"))
	if e != nil {
		t.Fatal(e)
	}
	entries, _, _, e := parseStory(data)
	if e != nil {
		t.Fatal(e)
	}
	text, e := os.ReadFile("../../../text/intro-story-prototype.tsv")
	if e != nil {
		t.Fatal(e)
	}
	if e = loadStoryText(entries, text, font); e != nil {
		t.Fatal(e)
	}
	var reference struct {
		Positions map[string][]struct {
			Rune      int    `json:"rune"`
			Character string `json:"character"`
			X, Y      int
		}
	}
	b, e := os.ReadFile(layout)
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &reference); e != nil {
		t.Fatal(e)
	}
	if len(reference.Positions) != 58 {
		t.Fatal("獨立排版來源數不符")
	}
	for _, entry := range entries {
		want := reference.Positions[entry.key]
		if len(want) != len(entry.positions) {
			t.Fatal("獨立排版字數不同", entry.key)
		}
		for i, p := range entry.positions {
			w := want[i]
			if p.ordinal != w.Rune || string(p.character) != w.Character || p.at != image.Pt(w.X, w.Y) {
				t.Fatal("獨立排版位置不同", entry.key, i, p, w)
			}
		}
	}
}

func TestIntroStoryOriginal(t *testing.T) {
	root, fontPath, out := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_STORY_OUTPUT")
	if root == "" || fontPath == "" || out == "" {
		t.Skip("須掛載本地原版、字型與全新輸出")
	}
	font, e := bitmap.Load(fontPath)
	if e != nil {
		t.Fatal(e)
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
	s, e := NewIntroStory(b, filepath.Join(root, "ZELRES1.SAR"), "../../../text/intro-story-prototype.tsv", font)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.Mkdir(out, 0755); e != nil {
		t.Fatal(e)
	}
	var records []map[string]any
	pageDraws := 0
	shadowDraws := 0
	clearDraws := 0
	compare := func(label string) {
		t.Helper()
		if a.Steps() > b.Steps() {
			t.Fatal("原版基線時間超前")
		}
		if e := a.Run(b.Steps() - a.Steps()); e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) || !reflect.DeepEqual(a.Regs(), b.Regs()) || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) {
			t.Fatal("故事覆繪改變原版狀態", label, b.Steps())
		}
		records = append(records, map[string]any{"label": label, "steps": b.Steps(), "cycles": b.Cycles(), "state_equal": true, "memory_sha256": fmtHash(b.Bytes(oracle.Far(0, 0), 1<<20))})
	}
	draw := func() (*image.RGBA, bool) {
		t.Helper()
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
		shown := s.Draw(dst, raw, b.Palette(), image.Point{})
		for y := 0; y < 429; y++ {
			start := y * dst.Stride
			if !bytes.Equal(before[start:start+960*4], dst.Pix[start:start+960*4]) {
				t.Fatal("故事覆繪寫入圖片區")
			}
		}
		if !bytes.Equal(raw, b.Indexed()) {
			t.Fatal("故事Draw改原版VGA")
		}
		return dst, shown
	}
	endFiles := map[int]bool{}
	for _, entry := range s.entries {
		endFiles[entry.files[len(entry.files)-1]] = true
	}
	b.OnCall(oracle.Far(0x1b1, 0x6ae3), func(o *oracle.Oracle) {
		if s.pending == 1 && s.lastFile != storyFileStart {
			c := s.chars[s.lastFile]
			if c.index == 0 {
				_, shown := draw()
				if !shown {
					t.Fatal("故事陰影途中無呈現", s.State())
				}
				shadowDraws++
			}
		}
	})
	b.OnCall(oracle.Far(0x1b1, 0x6aef), func(o *oracle.Oracle) {
		if !s.tracking {
			t.Fatal("故事字形被拒絕", s.State())
		}
		if endFiles[s.lastFile] {
			compare("fragment-" + s.entries[s.chars[s.lastFile].entry].key)
		}
	})
	savePage := func() {
		state := s.State()
		if active, _ := state["active"].(bool); !active {
			return
		}
		if complete, _ := state["page_complete"].(bool); !complete {
			t.Fatal("清框前故事未完整", state)
		}
		dst, shown := draw()
		if !shown {
			t.Fatal("完整故事頁未呈現", state)
		}
		name := filepath.Join(out, fmt.Sprintf("page-%02d.png", s.page))
		f, e := os.Create(name)
		if e != nil {
			t.Fatal(e)
		}
		if e = png.Encode(f, dst); e != nil {
			t.Fatal(e)
		}
		if e = f.Close(); e != nil {
			t.Fatal(e)
		}
		pageDraws++
	}
	b.OnCall(oracle.Far(0x1b1, 0x6bfd), func(o *oracle.Oracle) {
		if s.pending != 2 {
			t.Fatal("清框未建立原版批次", s.State())
		}
		compare(fmt.Sprintf("clear-start-%d", s.page))
		savePage()
	})
	b.OnCall(oracle.Far(0x1b1, 0x6c05), func(o *oracle.Oracle) {
		if s.pending == 2 && len(s.pageEntries()) > 0 {
			_, shown := draw()
			if !shown {
				t.Fatal("故事清框途中無呈現", s.State())
			}
			clearDraws++
		}
	})
	b.OnCall(oracle.Far(0x1b1, 0x6c0a), func(o *oracle.Oracle) {
		if s.Glyphs > 0 {
			compare(fmt.Sprintf("clear-end-%d", s.page))
			if s.pending != 0 || s.Rejected != 0 {
				t.Fatal("故事清框拒絕", s.State())
			}
		}
	})
	b.OnCall(oracle.Far(0x1b1, 0x6b25), func(o *oracle.Oracle) {
		if s.lastByte != 0xff || !s.finished {
			t.Fatal("故事結尾來源不符", s.State())
		}
		savePage()
		for _, kind := range []string{"unknown-pixel", "short-frame", "small-canvas"} {
			clone := *s
			raw := append([]byte(nil), b.Indexed()...)
			dst := image.NewRGBA(image.Rect(0, 0, 960, 600))
			if kind == "unknown-pixel" {
				raw[145*320+1] = 255
			}
			if kind == "short-frame" {
				raw = raw[:100]
			}
			if kind == "small-canvas" {
				dst = image.NewRGBA(image.Rect(0, 0, 12, 12))
			}
			before := append([]byte(nil), dst.Pix...)
			if clone.Draw(dst, raw, b.Palette(), image.Point{}) || !bytes.Equal(before, dst.Pix) {
				t.Fatal("故事退回非原子", kind)
			}
		}
	})
	for _, o := range []*oracle.Oracle{a, b} {
		if e = o.Run(200000000); e != nil {
			t.Fatal(e)
		}
		o.KeyDown(0x39)
		if e = o.Run(100000); e != nil {
			t.Fatal(e)
		}
		o.KeyUp(0x39)
	}
	if e = b.Run(2200000000 - b.Steps()); e != nil {
		t.Fatal(e)
	}
	compare("post-story")
	_, shown := draw()
	if shown {
		t.Fatal("故事原版撤除後仍有中文", s.State())
	}
	if s.Glyphs != 5083 || s.Clears != 52 || s.Fragments != 58 || s.Rejected != 0 || pageDraws != 47 || shadowDraws != 58 {
		t.Fatal("故事驗收數量不符", s.State(), pageDraws, shadowDraws, clearDraws)
	}
	j := map[string]any{"status": "PASS_COMPLETE_STORY_ORIGINAL_PAIR", "points": records, "glyphs": s.Glyphs, "clears": s.Clears, "fragments": s.Fragments, "rejected": s.Rejected, "page_draws": pageDraws, "shadow_stage_draws": shadowDraws, "clear_stage_draws": clearDraws, "final_state": s.State(), "full_original_state_equal": true, "post_story_removed": true, "input": "cold startup,Space200m/hold100k/release", "state_injection": false}
	data, e := json.MarshalIndent(j, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, 10), 0644); e != nil {
		t.Fatal(e)
	}
}

func TestIntroStoryMalformedCatalog(t *testing.T) {
	root, fontPath := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT")
	if root == "" || fontPath == "" {
		t.Skip("須掛載原版與字型")
	}
	font, e := bitmap.Load(fontPath)
	if e != nil {
		t.Fatal(e)
	}
	data, e := os.ReadFile(filepath.Join(root, "ZELRES1.SAR"))
	if e != nil {
		t.Fatal(e)
	}
	base, e := os.ReadFile("../../../text/intro-story-prototype.tsv")
	if e != nil {
		t.Fatal(e)
	}
	old := "很久以前，一場可怕的暴風雨襲擊了 Zeliard 大地。"
	for _, tc := range []struct {
		name    string
		text    []byte
		missing bool
	}{{"duplicate", []byte(strings.ReplaceAll(string(base), "001ab3", "001a6f")), false}, {"missing", []byte(strings.Join(strings.Split(string(base), "\n")[:58], "\n")), false}, {"blank", []byte(strings.Replace(string(base), old, "  ", 1)), false}, {"control", []byte(strings.Replace(string(base), old, "\x01", 1)), false}, {"utf8", append(append([]byte(nil), base...), 255), false}, {"word-width", []byte(strings.Replace(string(base), old, strings.Repeat("A", 100), 1)), false}, {"height", []byte(strings.Replace(string(base), old, strings.Repeat("國", 300), 1)), false}, {"missing-glyph", base, true}} {
		t.Run(tc.name, func(t *testing.T) {
			entries, _, _, e := parseStory(data)
			if e != nil {
				t.Fatal(e)
			}
			f := font
			if tc.missing {
				f = bitmap.Font{}
			}
			if e = loadStoryText(entries, tc.text, f); e == nil {
				t.Fatal("錯誤故事譯表未拒絕")
			}
		})
	}
	if _, _, _, e = parseStory([]byte{}); e == nil {
		t.Fatal("短來源未拒絕")
	}
}

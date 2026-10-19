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

func TestScrollBlocksCatalog(t *testing.T) {
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
	base, e := os.ReadFile("../../../text/intro-credits-prototype.tsv")
	if e != nil {
		t.Fatal(e)
	}
	block := scrollBlock{"credits", 0x14d3, 0x1a69, 51, 32, 120}
	entries, e := parseScrollCatalog(data, base, &block, font)
	if e != nil || len(entries) != 52 {
		t.Fatal("有效名單未載入", len(entries), e)
	}
	valid := []byte(strings.Replace(string(base), "奇幻動作遊戲", strings.Repeat("A", 75), 1))
	if _, e = parseScrollCatalog(data, valid, &block, font); e != nil {
		t.Fatal("900像素邊界未接受", e)
	}
	cases := []struct {
		name      string
		text      []byte
		emptyFont bool
	}{
		{"duplicate", []byte(strings.Replace(string(base), "0014fc", "0014d3", 1)), false},
		{"missing", []byte(strings.Join(strings.Split(string(base), "\n")[:51], "\n")), false},
		{"orphan", append(append([]byte(nil), base...), []byte("unknown\t未知\n")...), false},
		{"blank-original", []byte(strings.Replace(string(base), "zelres1.file.001525\t", "zelres1.file.001525\t新增", 1)), false},
		{"blank-translation", []byte(strings.Replace(string(base), "奇幻動作遊戲", "  ", 1)), false},
		{"control", []byte(strings.Replace(string(base), "奇幻動作遊戲", "\x01", 1)), false},
		{"utf8", append(append([]byte(nil), base...), 255), false},
		{"width", []byte(strings.Replace(string(base), "奇幻動作遊戲", strings.Repeat("A", 76), 1)), false},
		{"missing-glyph", base, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := font
			if tc.emptyFont {
				f = bitmap.Font{}
			}
			if _, e := parseScrollCatalog(data, tc.text, &block, f); e == nil {
				t.Fatal("錯誤名單未拒絕")
			}
		})
	}
	if _, e = parseScrollCatalog(nil, base, &block, font); e == nil {
		t.Fatal("短來源未拒絕")
	}
}

func TestScrollBlocksOriginal(t *testing.T) {
	root, fontPath, out, reference := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_SCROLL_BLOCKS_OUTPUT"), os.Getenv("ZELIARD_SCROLL_BLOCKS_LAYOUT")
	if root == "" || fontPath == "" || out == "" || reference == "" {
		t.Skip("須掛載原版、字型、獨立排版及新收據目錄")
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
	s, e := NewScrollBlocks(b, filepath.Join(root, "ZELRES1.SAR"), ScrollCatalogs{First: "../../../text/intro-scroll-prototype.tsv", Credits: "../../../text/intro-credits-prototype.tsv", Outro: "../../../text/intro-outro-prototype.tsv"}, font)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.Mkdir(out, 0755); e != nil {
		t.Fatal(e)
	}
	var layout struct {
		Frames []struct {
			Name, Block string
			SourceFile  int    `json:"source_file"`
			RGBSHA      string `json:"RGB_sha256"`
		}
	}
	data, e := os.ReadFile(reference)
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(data, &layout); e != nil {
		t.Fatal(e)
	}
	names := map[int]string{}
	hashes := map[int]string{}
	for _, f := range layout.Frames {
		names[f.SourceFile] = f.Name
		hashes[f.SourceFile] = f.RGBSHA
	}
	if len(names) != 90 {
		t.Fatal("獨立來源不完整")
	}
	points := []map[string]any{}
	frames := []map[string]any{}
	compare := func(label string) {
		t.Helper()
		if a.Steps() > b.Steps() {
			t.Fatal("原版基線時間超前")
		}
		if a.Steps() < b.Steps() {
			if e := a.Run(b.Steps() - a.Steps()); e != nil {
				t.Fatal(e)
			}
		}
		if !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) || !reflect.DeepEqual(a.Regs(), b.Regs()) || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) {
			t.Fatal("捲字覆繪改變原版狀態", label, b.Steps())
		}
		points = append(points, map[string]any{"label": label, "steps": b.Steps(), "cycles": b.Cycles(), "state_equal": true, "memory_sha256": fmtHash(b.Bytes(oracle.Far(0, 0), 1<<20))})
	}
	draw := func() (*image.RGBA, bool) {
		t.Helper()
		raw := b.Indexed()
		pal := b.Palette()
		dst := image.NewRGBA(image.Rect(0, 0, 960, 600))
		for y := 0; y < 600; y++ {
			for x := 0; x < 960; x++ {
				c := pal[raw[y/3*320+x/3]]
				dst.SetRGBA(x, y, color.RGBA{c[0], c[1], c[2], 255})
			}
		}
		before := append([]byte(nil), dst.Pix...)
		shown := s.Draw(dst, raw, pal, image.Point{})
		if !bytes.Equal(raw, b.Indexed()) {
			t.Fatal("捲字Draw改原版VGA")
		}
		top, bottom := s.block.top*3, (s.block.top+s.block.height)*3
		for y := 0; y < 600; y++ {
			if y < top || y >= bottom {
				at := y * dst.Stride
				if !bytes.Equal(before[at:at+960*4], dst.Pix[at:at+960*4]) {
					t.Fatal("捲字區外寫入", y)
				}
			}
		}
		return dst, shown
	}
	pendingSeen := map[int]bool{}
	pendingByBlock := map[string]int{}
	b.OnCall(oracle.Far(0x1b1, 0x33a3), func(o *oracle.Oracle) {
		if !s.active || !s.pending || s.sourceRows != 5 || pendingSeen[s.file] || len(s.completed) == 0 {
			return
		}
		i := int(o.Regs().DI) - s.block.top*320
		if i < 0 || i+1 >= len(s.completed) || (s.completed[i]&0x66 == 0 && s.completed[i+1]&0x66 == 0) {
			return
		}
		if _, shown := draw(); !shown {
			t.Fatal("AND途中未保留完整中文", s.State())
		}
		pendingSeen[s.file] = true
		pendingByBlock[s.block.name]++
	})
	lastRows := 0
	fallbacks := 0
	b.OnCall(oracle.Far(0x1b1, 0x33b6), func(o *oracle.Oracle) {
		if s.Rejected != 0 {
			t.Fatal("正常捲字被拒絕", s.State())
		}
		if s.Rows == lastRows {
			return
		}
		if s.Rows != lastRows+1 {
			t.Fatal("逐列計數跳躍")
		}
		lastRows = s.Rows
		if s.sourceRows == 10 {
			compare(fmt.Sprintf("source-%06x", s.file))
			dst, shown := draw()
			if !shown {
				t.Fatal("完整來源未呈現", s.State())
			}
			rgb := make([]byte, 0, 960*600*3)
			for i := 0; i < len(dst.Pix); i += 4 {
				rgb = append(rgb, dst.Pix[i:i+3]...)
			}
			if names[s.file] == "" || fmtHash(rgb) != hashes[s.file] {
				t.Fatal("正式捲字與獨立中文排版不同", s.file, s.State(), fmtHash(rgb), hashes[s.file])
			}
			path := filepath.Join(out, names[s.file]+".png")
			f, e := os.Create(path)
			if e != nil {
				t.Fatal(e)
			}
			if e = png.Encode(f, dst); e != nil {
				t.Fatal(e)
			}
			if e = f.Close(); e != nil {
				t.Fatal(e)
			}
			frames = append(frames, map[string]any{"file": filepath.Base(path), "source_file": s.file, "block": s.block.name, "steps": b.Steps(), "RGB_sha256": fmtHash(rgb), "independent_pixels_equal": true})
			if s.file == 0x14a9 {
				for _, kind := range []string{"unknown-pixel", "short-frame", "small-canvas", "pool-identity"} {
					clone := *s
					raw := append([]byte(nil), b.Indexed()...)
					canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
					switch kind {
					case "unknown-pixel":
						raw[(s.block.top+5)*320+1] = 255
					case "short-frame":
						raw = raw[:100]
					case "small-canvas":
						canvas = image.NewRGBA(image.Rect(0, 0, 12, 12))
					case "pool-identity":
						clone.pool = append([]byte(nil), s.pool...)
						clone.pool[0] ^= 1
					}
					before := append([]byte(nil), canvas.Pix...)
					if clone.Draw(canvas, raw, b.Palette(), image.Point{}) || !bytes.Equal(before, canvas.Pix) {
						t.Fatal("捲字退回不原子", kind)
					}
					fallbacks++
				}
			}
		}
		if !s.active {
			compare("block-end-" + s.block.name)
			if _, shown := draw(); shown {
				t.Fatal("捲字尾端仍覆繪", s.State())
			}
		}
	})
	if e = b.Run(200000000); e != nil {
		t.Fatal(e)
	}
	compare("title-200m")
	a.KeyDown(0x39)
	b.KeyDown(0x39)
	if e = b.Run(100000); e != nil {
		t.Fatal(e)
	}
	if e = a.Run(100000); e != nil {
		t.Fatal(e)
	}
	a.KeyUp(0x39)
	b.KeyUp(0x39)
	if e = b.Run(2200000000 - b.Steps()); e != nil {
		t.Fatal(e)
	}
	compare("post-outro")
	if s.Rows != 1300 || s.Sources != 87 || s.BlocksCompleted != 3 || s.Rejected != 0 || len(frames) != 90 || fallbacks != 4 || len(points) != 95 {
		t.Fatal("三組驗收計數不符", s.State(), len(frames), fallbacks, len(points))
	}
	for name, want := range map[string]int{"first": 430, "credits": 640, "outro": 230} {
		if s.blockRows[name] != want || pendingByBlock[name] == 0 {
			t.Fatal("區塊逐列或途中呈現未覆蓋", name, s.blockRows, pendingByBlock)
		}
	}
	result := map[string]any{"status": "PASS_THREE_SCROLL_BLOCKS_ORIGINAL_PAIR", "points": points, "frames": frames, "final_state": s.State(), "full_original_state_equal": true, "independent_layout_frames": 90, "pending_AND_draws": pendingByBlock, "atomic_fallbacks": fallbacks, "input": "cold startup,Space200m/hold100k/release", "state_injection": false}
	data, e = json.MarshalIndent(result, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, 10), 0644); e != nil {
		t.Fatal(e)
	}
}

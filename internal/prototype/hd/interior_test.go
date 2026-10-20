package hd

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

func interiorInputs(t *testing.T) (*Interiors, string, string, []byte) {
	t.Helper()
	project, original := os.Getenv("ZELIARD_PROJECT"), os.Getenv("ZELIARD_ORIG")
	if project == "" || original == "" {
		t.Skip("須掛載原版及完整室內圖組；skip不算驗收")
	}
	study := filepath.Join(project, "workplace/full-hd-v1")
	s, err := NewInteriors(original, filepath.Join(study, "interior-complete-v2"))
	if err != nil {
		t.Fatal(err)
	}
	pal, err := os.ReadFile(filepath.Join(project, "workplace/music-release-v1/release-build/data/assets/town/frame/palette.bin"))
	if err != nil {
		t.Fatal(err)
	}
	return s, project, study, pal
}

func interiorRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func interiorFixture(t *testing.T, study, frame string, palette []byte) ([]byte, []byte, []byte, []byte) {
	t.Helper()
	if frame == "OMOYA-base" {
		memory := interiorRead(t, filepath.Join(study, "hut-normal-v2/hut-stable.memory"))
		raw := interiorRead(t, filepath.Join(study, "hut-normal-v2/hut-stable.indexed"))
		rgb, _ := skyOriginal(raw, palette)
		return memory[0x1b10:0x11b10], memory[0x11b10:0x21b10], raw, rgb
	}
	// 旅館是結構fixture，沒有執行VM、恢復存檔或宣稱正常入口。
	p, r, raw := make([]byte, 65536), make([]byte, 65536), make([]byte, 64000)
	copy(p[0xa000:], interiorRead(t, filepath.Join(study, "interior-code-v1/INNAPRO.BIN")))
	copy(r[0x8000:], interiorRead(t, filepath.Join(study, "planar-v1/zelres2-24.planar")))
	pixels := interiorRead(t, filepath.Join(study, "interior-layout-v2", frame+".indexed"))
	if len(pixels) != 6144 {
		t.Fatal("旅館結構畫格尺寸錯誤")
	}
	for y := 0; y < 64; y++ {
		copy(raw[(23+y)*320+56:(23+y)*320+152], pixels[y*96:(y+1)*96])
	}
	rgb, _ := skyOriginal(raw, palette)
	return p, r, raw, rgb
}

func interiorCanvas(t *testing.T, rgb []byte) *image.RGBA {
	t.Helper()
	canvas := image.NewRGBA(image.Rect(0, 0, 960, 680))
	if !bitmap.DrawRGB3(canvas, rgb, image.Pt(0, 48)) {
		t.Fatal("原生畫格形狀錯誤")
	}
	return canvas
}

func TestInteriorsSourcesAliasesAndPixels(t *testing.T) {
	s, _, study, palette := interiorInputs(t)
	samples := []struct{ frame, artwork, pose string }{
		{"OMOYA-base", "hut", "base"}, {"INN-base", "inn-base", "base-aliases"},
		{"INN-ambient-0", "inn-base", "base-aliases"}, {"INN-rest-0", "inn-base", "base-aliases"},
		{"INN-rest-1", "inn-base", "base-aliases"}, {"INN-ambient-1", "inn-ambient-1", "talk"},
		{"INN-rest-2", "inn-rest-2", "hand-near"}, {"INN-rest-3", "inn-rest-3", "hand-open"},
	}
	var receipts []map[string]any
	for _, sample := range samples {
		t.Run(sample.frame, func(t *testing.T) {
			p, r, raw, rgb := interiorFixture(t, study, sample.frame, palette)
			input := [][]byte{cloneTownBytes(p), cloneTownBytes(r), cloneTownBytes(raw), cloneTownBytes(rgb), cloneTownBytes(palette)}
			canvas := interiorCanvas(t, rgb)
			before := cloneTownBytes(canvas.Pix)
			if !s.Draw(canvas, 0x13, p, r, raw, rgb, palette, image.Pt(0, 48)) || s.State()["frame"] != sample.pose {
				t.Fatal("室內原版來源或圖面失配", s.State())
			}
			profile := interiorProfiles[1]
			if sample.frame == "OMOYA-base" {
				profile = interiorProfiles[0]
			}
			box := profile.box.Inset(profile.inset)
			box = image.Rectangle{Min: box.Min.Mul(3), Max: box.Max.Mul(3)}.Add(image.Pt(0, 48))
			f, err := os.Open(filepath.Join(study, "interior-complete-v2", sample.artwork+"-hd-v1.png"))
			if err != nil {
				t.Fatal(err)
			}
			art, err := png.Decode(f)
			f.Close()
			if err != nil {
				t.Fatal(err)
			}
			pixels := 0
			for y := 0; y < 680; y++ {
				for x := 0; x < 960; x++ {
					at := canvas.PixOffset(x, y)
					if !image.Pt(x, y).In(box) {
						if !bytes.Equal(before[at:at+4], canvas.Pix[at:at+4]) {
							t.Fatal("覆繪寫入區外或原版外框", x, y)
						}
						continue
					}
					want := art.At(x-profile.box.Min.X*3, y-48-profile.box.Min.Y*3)
					if canvas.RGBAAt(x, y) != color.RGBAModel.Convert(want).(color.RGBA) {
						t.Fatal("註冊圖直接3倍合成失配", x, y)
					}
					pixels++
				}
			}
			for i, data := range [][]byte{p, r, raw, rgb, palette} {
				if !bytes.Equal(input[i], data) {
					t.Fatal("覆繪改寫來源", i)
				}
			}
			if out := os.Getenv("ZELIARD_INTERIOR_OUTPUT"); out != "" {
				file, err := os.Create(filepath.Join(out, sample.frame+"-hd.png"))
				if err != nil {
					t.Fatal(err)
				}
				err = png.Encode(file, canvas)
				file.Close()
				if err != nil {
					t.Fatal(err)
				}
			}
			receipts = append(receipts, map[string]any{"frame": sample.frame, "pose": sample.pose, "drawn_pixels": pixels,
				"input_unchanged": true, "normal_hut_source": sample.frame == "OMOYA-base", "inn_structural_fixture": sample.frame != "OMOYA-base"})
		})
	}
	if out := os.Getenv("ZELIARD_INTERIOR_OUTPUT"); out != "" {
		data, _ := json.MarshalIndent(map[string]any{"status": "PASS_INTERIOR_SOURCE_AND_ART_PIXELS", "frames": receipts}, "", "  ")
		if err := os.WriteFile(filepath.Join(out, "render.json"), append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestInteriorsAtomicFallback(t *testing.T) {
	s, _, study, palette := interiorInputs(t)
	for _, label := range []string{"mode", "short", "module", "graphics", "native-index", "native-rgb", "palette", "canvas", "missing-art"} {
		t.Run(label, func(t *testing.T) {
			p, r, raw, rgb := interiorFixture(t, study, "INN-rest-3", palette)
			pal := cloneTownBytes(palette)
			canvas := interiorCanvas(t, rgb)
			mode := byte(0x13)
			switch label {
			case "mode":
				mode = 3
			case "short":
				p = p[:32]
			case "module":
				p[0xa000] ^= 1
			case "graphics":
				r[0x8000] ^= 1
			case "native-index":
				raw[24*320+57] ^= 1
			case "native-rgb":
				rgb[(24*320+57)*3] ^= 1
			case "palette":
				pal[0] ^= 1
			case "canvas":
				canvas = image.NewRGBA(image.Rect(0, 0, 100, 100))
			case "missing-art":
				old := s.art[4]
				s.art[4] = nil
				defer func() { s.art[4] = old }()
			}
			before := cloneTownBytes(canvas.Pix)
			if s.Draw(canvas, mode, p, r, raw, rgb, pal, image.Pt(0, 48)) || !bytes.Equal(canvas.Pix, before) {
				t.Fatal("失配沒有原子退回", label)
			}
		})
	}
	// 原版會寫A505及後續暫存，不能把這些可寫資料誤作固定程式。
	p, r, raw, rgb := interiorFixture(t, study, "INN-base", palette)
	for at := 0xa505; at < 0xa510; at++ {
		p[at] ^= 0x55
	}
	if !s.Draw(interiorCanvas(t, rgb), 0x13, p, r, raw, rgb, palette, image.Pt(0, 48)) {
		t.Fatal("不相關可寫資料造成HD消失")
	}
	memory := interiorRead(t, filepath.Join(study, "home-normal-v1/home-stable.memory"))
	raw = interiorRead(t, filepath.Join(study, "home-normal-v1/home-stable.indexed"))
	rgb, _ = skyOriginal(raw, palette)
	canvas := interiorCanvas(t, rgb)
	before := cloneTownBytes(canvas.Pix)
	if s.Draw(canvas, 0x13, memory[0x1b10:0x11b10], memory[0x11b10:0x21b10], raw, rgb, palette, image.Pt(0, 48)) || !bytes.Equal(before, canvas.Pix) {
		t.Fatal("新室內圖覆蓋正常長老")
	}
}

func TestInteriorsMissingAssets(t *testing.T) {
	_, project, study, _ := interiorInputs(t)
	directory := t.TempDir()
	for _, name := range []string{"hut", "inn-base", "inn-ambient-1", "inn-rest-2", "inn-rest-3"} {
		file := name + "-hd-v1.png"
		if err := os.WriteFile(filepath.Join(directory, file), interiorRead(t, filepath.Join(study, "interior-complete-v2", file)), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(directory, "inn-rest-3-hd-v1.png"), []byte("corrupt"), 0644); err != nil {
		t.Fatal(err)
	}
	if s, err := NewInteriors(filepath.Join(project, "workplace/orig"), directory); err == nil || s != nil {
		t.Fatal("損圖仍載入部分室內圖組")
	}
}

func TestInteriorsNormalHutReadOnly(t *testing.T) {
	s, project, _, _ := interiorInputs(t)
	root := os.Getenv("ZELIARD_ORIG")
	var machines [2]*oracle.Oracle
	for i := range machines {
		o, err := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if err != nil {
			t.Fatal(err)
		}
		machines[i] = o
		defer o.Close()
		o.SetDOSBoxCycles(3000)
		o.SetScratch(t.TempDir())
	}
	a, b := machines[0], machines[1]
	run := func(n uint64) {
		for _, o := range machines {
			if err := o.Run(n); err != nil {
				t.Fatal(err)
			}
		}
	}
	var points []map[string]any
	commands := interiorRead(t, filepath.Join(project, "tools/hut-hd.commands"))
	number := func(value string) uint64 {
		n, err := strconv.ParseUint(value, 0, 64)
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	for _, line := range strings.Split(string(commands), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		switch fields[0] {
		case "run":
			run(number(fields[1]))
		case "key":
			for _, o := range machines {
				o.KeyDown(uint8(number(fields[1])))
			}
			run(number(fields[2]))
			for _, o := range machines {
				o.KeyUp(uint8(number(fields[1])))
			}
		case "shot":
			before := b.Bytes(oracle.Far(0, 0), 1<<20)
			raw := b.Indexed()
			_, _, rgb := b.ScreenRGB()
			var palette []byte
			for _, color := range b.Palette() {
				palette = append(palette, color[:]...)
			}
			canvas := interiorCanvas(t, rgb)
			matched := s.Draw(canvas, b.VideoMode(), before[0x1b10:0x11b10], before[0x11b10:0x21b10], raw, rgb, palette, image.Pt(0, 48))
			want := strings.HasPrefix(fields[1], "hut-")
			if matched != want {
				t.Fatal("正常小屋呈現與離開不符", fields[1], s.State())
			}
			if !bytes.Equal(before, b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), before) ||
				!bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Regs() != b.Regs() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() ||
				!reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) {
				t.Fatal("正常原版完整狀態不同", fields[1])
			}
			points = append(points, map[string]any{"name": fields[1], "matched": matched, "state_equal": true, "steps": b.Steps(), "cycles": b.Cycles(), "memory_sha256": townBackgroundHash(before)})
		default:
			t.Fatal("未知命令", line)
		}
	}
	if len(points) != 5 || len(b.HeldKeys()) != 0 {
		t.Fatal("正常抽樣不足或按鍵未釋放")
	}
	if output := os.Getenv("ZELIARD_INTERIOR_OUTPUT"); output != "" {
		data, _ := json.MarshalIndent(map[string]any{"status": "PASS_NORMAL_HUT_READ_ONLY", "checkpoints": points, "commands_sha256": townBackgroundHash(commands), "scope": "正常冷新局、亭台入小屋、看石像及離開；非旅館正常入口或GUI"}, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "normal-hut.json"), append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

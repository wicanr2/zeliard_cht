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
)

func sageFramesInputs(t *testing.T) (*SageFrames, string, []byte) {
	t.Helper()
	p := os.Getenv("ZELIARD_PROJECT")
	if p == "" || os.Getenv("ZELIARD_ORIG") == "" {
		t.Skip("須掛載原版與長老完整圖組，skip不算驗收")
	}
	s, e := NewSageFrames(os.Getenv("ZELIARD_ORIG"), filepath.Join(p, "workplace/full-hd-v1/sage-frames-complete-v2"))
	if e != nil {
		t.Fatal(e)
	}
	pal := interiorRead(t, filepath.Join(p, "workplace/music-release-v1/release-build/data/assets/town/frame/palette.bin"))
	return s, p, pal
}

type sageVariant struct {
	Key       string `json:"key"`
	Pose      int    `json:"pose"`
	Animation *int   `json:"animation_tile"`
	SHA       string `json:"indexed_sha256"`
}

func sageFramesFixture(t *testing.T, project string, v sageVariant, x int, pal []byte) ([]byte, []byte, []byte, []byte) {
	t.Helper()
	study := filepath.Join(project, "workplace/full-hd-v1")
	p, r, raw := make([]byte, 65536), make([]byte, 65536), make([]byte, 64000)
	copy(p[0xa000:], interiorRead(t, filepath.Join(study, "sage-frames-source-v1/KENJPRO.BIN")))
	p[0xbb12] = 23
	p[0xbb13] = byte(x / 8)
	copy(r[0x8000:], interiorRead(t, filepath.Join(study, "planar-v1/zelres2-25.planar")))
	pixels := interiorRead(t, filepath.Join(study, "sage-frames-layout-v1", v.Key+".indexed"))
	if townBackgroundHash(pixels) != v.SHA {
		t.Fatal("結構圖形身份錯誤")
	}
	for y := 0; y < 64; y++ {
		copy(raw[(23+y)*320+x:(23+y)*320+x+96], pixels[y*96:(y+1)*96])
	}
	rgb, _ := skyOriginal(raw, pal)
	return p, r, raw, rgb
}

func sageFramesVariants(t *testing.T, project string) []sageVariant {
	t.Helper()
	var layout struct {
		Variants []sageVariant `json:"structural_variants"`
	}
	if e := json.Unmarshal(interiorRead(t, filepath.Join(project, "workplace/full-hd-v1/sage-frames-layout-v1/receipt.json")), &layout); e != nil {
		t.Fatal(e)
	}
	if len(layout.Variants) != 38 {
		t.Fatal("來源組合分母不符", len(layout.Variants))
	}
	return layout.Variants
}

func TestSageFramesAllSourcesAndDirectPixels(t *testing.T) {
	s, project, pal := sageFramesInputs(t)
	var results []map[string]any
	for _, v := range sageFramesVariants(t, project) {
		for _, x := range []int{56, 112} {
			p, r, raw, rgb := sageFramesFixture(t, project, v, x, pal)
			inputs := [][]byte{bytes.Clone(p), bytes.Clone(r), bytes.Clone(raw), bytes.Clone(rgb), bytes.Clone(pal)}
			dst := interiorCanvas(t, rgb)
			before := bytes.Clone(dst.Pix)
			phase := 0
			if v.Animation != nil && (*v.Animation == 42 || *v.Animation == 104) {
				phase = 1
			}
			if !s.Draw(dst, 0x13, p, r, raw, rgb, pal, image.Pt(0, 48)) || s.State()["pose"] != v.Pose || s.State()["phase"] != phase {
				t.Fatal("來源姿態失配", v, x, s.State())
			}
			path := filepath.Join(project, "workplace/full-hd-v1/sage-frames-complete-v2", "sage-pose-"+strconv.Itoa(v.Pose)+"-phase-"+strconv.Itoa(phase)+".png")
			art, e := png.Decode(bytes.NewReader(interiorRead(t, path)))
			if e != nil {
				t.Fatal(e)
			}
			box := image.Rect((x+1)*3, 24*3+48, (x+95)*3, 86*3+48)
			pixels := 0
			for y := 0; y < 680; y++ {
				for cx := 0; cx < 960; cx++ {
					at := dst.PixOffset(cx, y)
					if !image.Pt(cx, y).In(box) {
						if !bytes.Equal(before[at:at+4], dst.Pix[at:at+4]) {
							t.Fatal("區外或原版外框被改寫", v, cx, y)
						}
						continue
					}
					want := color.RGBAModel.Convert(art.At(cx-x*3, y-23*3-48)).(color.RGBA)
					if dst.RGBAAt(cx, y) != want {
						t.Fatal("獨立3倍圖面不符", v, cx, y)
					}
					pixels++
				}
			}
			for i, b := range [][]byte{p, r, raw, rgb, pal} {
				if !bytes.Equal(inputs[i], b) {
					t.Fatal("改寫原版來源", i)
				}
			}
			results = append(results, map[string]any{"key": v.Key, "pose": v.Pose, "phase": phase, "origin_x": x, "drawn_pixels": pixels, "structural_fixture": true, "inputs_unchanged": true})
		}
	}
	sageFramesReceipt(t, "render.json", map[string]any{"status": "PASS_SAGE_ALL_STRUCTURAL_PIXELS", "samples": results})
}

func sageFramesReceipt(t *testing.T, name string, value any) {
	t.Helper()
	if out := os.Getenv("ZELIARD_SAGE_OUTPUT"); out != "" {
		b, e := json.MarshalIndent(value, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(out, name), append(b, 10), 0644); e != nil {
			t.Fatal(e)
		}
	}
}

func TestSageFramesFallbackAndNormalSnapshots(t *testing.T) {
	s, project, pal := sageFramesInputs(t)
	study := filepath.Join(project, "workplace/full-hd-v1")
	var layout struct {
		Bindings []struct {
			File string `json:"file"`
		} `json:"normal_bindings"`
		Rejected []struct {
			File string `json:"file"`
		} `json:"normal_rejected_frames"`
	}
	if e := json.Unmarshal(interiorRead(t, filepath.Join(study, "sage-frames-layout-v1/receipt.json")), &layout); e != nil {
		t.Fatal(e)
	}
	count := 0
	for _, group := range []struct {
		files []string
		want  bool
	}{
		{func() []string {
			var names []string
			for _, b := range layout.Bindings {
				names = append(names, b.File)
			}
			return names
		}(), true},
		{func() []string {
			var names []string
			for _, b := range layout.Rejected {
				names = append(names, b.File)
			}
			return names
		}(), false},
	} {
		for _, name := range group.files {
			mem := interiorRead(t, filepath.Join(study, "sage-frames-normal-v4", name+".memory"))
			raw := interiorRead(t, filepath.Join(study, "sage-frames-normal-v4", name+".indexed"))
			rgb, _ := skyOriginal(raw, pal)
			canvas := interiorCanvas(t, rgb)
			before := bytes.Clone(canvas.Pix)
			if s.Draw(canvas, 0x13, mem[0x1b10:0x11b10], mem[0x11b10:0x21b10], raw, rgb, pal, image.Pt(0, 48)) != group.want {
				t.Fatal("正常來源／混合畫格判斷不符", name)
			}
			if !group.want && !bytes.Equal(before, canvas.Pix) {
				t.Fatal("混合畫格沒有原子拒絕", name)
			}
			count++
		}
	}
	if count != 19 {
		t.Fatal("正常來源分母錯誤", count)
	}
	v := sageFramesVariants(t, project)[0]
	for _, bad := range []string{"mode", "length", "module", "graphics", "palette", "indexed", "rgb", "position", "canvas", "asset"} {
		t.Run(bad, func(t *testing.T) {
			p, r, raw, rgb := sageFramesFixture(t, project, v, 56, pal)
			palette := bytes.Clone(pal)
			dst := interiorCanvas(t, rgb)
			mode := byte(0x13)
			switch bad {
			case "mode":
				mode = 3
			case "length":
				raw = raw[:32]
			case "module":
				p[0xa000] ^= 1
			case "graphics":
				r[0x8000] ^= 1
			case "palette":
				palette[0] ^= 1
			case "indexed":
				raw[24*320+57] ^= 1
			case "rgb":
				rgb[(24*320+57)*3] ^= 1
			case "position":
				p[0xbb13] = 8
			case "canvas":
				dst = image.NewRGBA(image.Rect(0, 0, 100, 100))
			case "asset":
				old := s.art[0][0]
				s.art[0][0] = nil
				defer func() { s.art[0][0] = old }()
			}
			before := bytes.Clone(dst.Pix)
			if s.Draw(dst, mode, p, r, raw, rgb, palette, image.Pt(0, 48)) || !bytes.Equal(before, dst.Pix) {
				t.Fatal("失配未原子拒絕", bad)
			}
		})
	}
	p, r, raw, rgb := sageFramesFixture(t, project, v, 56, pal)
	for i := 0xbb14; i < 0xbb3d; i++ {
		p[i] ^= 0x55
	}
	if !s.Draw(interiorCanvas(t, rgb), 0x13, p, r, raw, rgb, pal, image.Pt(0, 48)) {
		t.Fatal("不相關可寫資料阻止HD")
	}
}

func TestSageFramesMissingAssetGroup(t *testing.T) {
	_, project, _ := sageFramesInputs(t)
	dir := t.TempDir()
	source := filepath.Join(project, "workplace/full-hd-v1/sage-frames-complete-v2")
	files, e := os.ReadDir(source)
	if e != nil {
		t.Fatal(e)
	}
	for _, file := range files {
		if file.Name() == "manifest.json" || strings.HasPrefix(file.Name(), "sage-pose-") {
			if e = os.WriteFile(filepath.Join(dir, file.Name()), interiorRead(t, filepath.Join(source, file.Name())), 0644); e != nil {
				t.Fatal(e)
			}
		}
	}
	if e = os.WriteFile(filepath.Join(dir, "sage-pose-7-phase-1.png"), []byte("corrupt"), 0644); e != nil {
		t.Fatal(e)
	}
	if s, e := NewSageFrames(os.Getenv("ZELIARD_ORIG"), dir); e == nil || s != nil {
		t.Fatal("損壞圖組未整組拒絕")
	}
}

func TestSageFramesNormalReadOnly(t *testing.T) {
	s, project, pal := sageFramesInputs(t)
	save := interiorRead(t, filepath.Join(project, "workplace/prose-v1/npc-normal-v2/scratch/SAVETEST.usr"))
	if townBackgroundHash(save) != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
		t.Fatal("自建存檔身份")
	}
	var machines [2]*oracle.Oracle
	for i := range machines {
		o, e := oracle.Load(filepath.Join(os.Getenv("ZELIARD_ORIG"), "ZELIARD.EXE"), os.Getenv("ZELIARD_ORIG"))
		if e != nil {
			t.Fatal(e)
		}
		defer o.Close()
		o.SetDOSBoxCycles(3000)
		scratch := t.TempDir()
		if e = os.WriteFile(filepath.Join(scratch, "SAVETEST.usr"), save, 0644); e != nil {
			t.Fatal(e)
		}
		o.SetScratch(scratch)
		machines[i] = o
	}
	a, b := machines[0], machines[1]
	seen := map[string]bool{}
	poses := map[int]bool{}
	var points []map[string]any
	sample := func(name string, force bool) {
		if b.VideoMode() != 0x13 {
			return
		}
		raw := b.Indexed()
		_, _, rgb := b.ScreenRGB()
		memory := b.Bytes(oracle.Far(0, 0), 1<<20)
		canvas := interiorCanvas(t, rgb)
		matched := s.Draw(canvas, b.VideoMode(), memory[0x1b10:0x11b10], memory[0x11b10:0x21b10], raw, rgb, pal, image.Pt(0, 48))
		key := ""
		if matched {
			key = strconv.Itoa(s.State()["pose"].(int)) + "/" + strconv.Itoa(s.State()["phase"].(int))
			poses[s.State()["pose"].(int)] = true
		}
		if !force && (!matched || seen[key]) {
			return
		}
		seen[key] = true
		if !bytes.Equal(memory, b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(memory, a.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Regs() != b.Regs() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) {
			t.Fatal("顯示改變完整原版狀態", name, key)
		}
		points = append(points, map[string]any{"name": name, "key": key, "matched": matched, "state_equal": true, "memory_sha256": townBackgroundHash(memory), "steps": b.Steps(), "cycles": b.Cycles()})
		if matched {
			if out := os.Getenv("ZELIARD_SAGE_OUTPUT"); out != "" {
				f, e := os.Create(filepath.Join(out, "normal-pose-"+strings.ReplaceAll(key, "/", "-")+".png"))
				if e != nil {
					t.Fatal(e)
				}
				if e = png.Encode(f, canvas); e != nil {
					t.Fatal(e)
				}
				f.Close()
			}
		}
	}
	run := func(n uint64) {
		for n > 0 {
			step := uint64(100000)
			if n < step {
				step = n
			}
			for _, o := range machines {
				if e := o.Run(step); e != nil {
					t.Fatal(e)
				}
			}
			n -= step
			sample("animation", false)
		}
	}
	number := func(v string) uint64 {
		n, e := strconv.ParseUint(v, 0, 64)
		if e != nil {
			t.Fatal(e)
		}
		return n
	}
	commands := interiorRead(t, filepath.Join(project, "tools/sage-hd.commands"))
	for _, line := range strings.Split(string(commands), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		switch f[0] {
		case "run":
			run(number(f[1]))
		case "key":
			for _, o := range machines {
				o.KeyDown(uint8(number(f[1])))
			}
			run(number(f[2]))
			for _, o := range machines {
				o.KeyUp(uint8(number(f[1])))
			}
		case "shot":
			sample(f[1], true)
		default:
			t.Fatal("未知鍵序", line)
		}
	}
	if len(poses) != 8 || len(b.HeldKeys()) != 0 {
		t.Fatal("正常手勢抽樣不足", poses)
	}
	if s.State()["matched"] != false {
		t.Fatal("離店後長老HD未撤除")
	}
	sageFramesReceipt(t, "normal.json", map[string]any{"status": "PASS_NORMAL_SAGE_READ_ONLY", "poses": poses, "points": points, "command_sha256": townBackgroundHash(commands), "source_save_sha256": townBackgroundHash(save), "scope": "兩個獨立dosgolem冷F7、正常進長老、查看力量八手勢及離店；非GUI"})
}

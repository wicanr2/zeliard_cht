package hd

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
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
	"github.com/wicanr2/zeliard_cht/internal/prototype/overlay"
)

func backgroundTestInputs(t *testing.T) (string, string, *TownBackground) {
	t.Helper()
	root, project := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_PROJECT")
	ref, pal := os.Getenv("ZELIARD_SKY_REFERENCE"), os.Getenv("ZELIARD_SKY_PALETTE")
	memory, art := os.Getenv("ZELIARD_BACKGROUND_MEMORY"), os.Getenv("ZELIARD_BACKGROUND_ART")
	if root == "" || project == "" || ref == "" || pal == "" || memory == "" || art == "" {
		t.Skip("須掛載原版、正常記憶體、背景來源與本地手繪圖；skip不計驗收")
	}
	s, err := NewTownBackground(root, ref, pal, memory, art)
	if err != nil {
		t.Fatal(err)
	}
	return root, project, s
}

func TestTownBackgroundOriginalMasksAndFallback(t *testing.T) {
	root, project, s := backgroundTestInputs(t)
	pal, err := os.ReadFile(os.Getenv("ZELIARD_SKY_PALETTE"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"cold-town", "restored", "citizen-near", "citizen-outside", "citizen-page", "citizen-last"} {
		base := filepath.Join(project, "workplace/town-hd-v1/foreground-v1/normal-v1", name)
		memory, err := os.ReadFile(base + ".memory")
		if err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(base + ".indexed")
		if err != nil {
			t.Fatal(err)
		}
		primary, resources := memory[0x1b10:0x11b10], memory[0x11b10:0x21b10]
		rgb, dst := skyOriginal(raw, pal)
		before := append([]byte(nil), dst.Pix...)
		memoryBefore := append([]byte(nil), memory...)
		mask, matched := s.backgroundMask(0x13, primary, resources, raw, rgb)
		want := name != "citizen-page" && name != "citizen-last"
		if matched != want || s.Draw(dst, 0x13, primary, resources, raw, rgb, image.Point{}) != want {
			t.Fatal("正常背景來源狀態錯誤", name)
		}
		changed, lowerChanged := 0, 0
		for y := 0; y < 600; y++ {
			for x := 0; x < 960; x++ {
				i := dst.PixOffset(x, y)
				if bytes.Equal(before[i:i+4], dst.Pix[i:i+4]) {
					continue
				}
				changed++
				p := image.Pt(x/3, y/3)
				if !want || !p.In(townBackgroundBox) || !mask[(p.Y-24)*224+p.X-48] {
					t.Fatal("HD覆寫原版前景或區外", name, p)
				}
				if p.Y >= 78 {
					lowerChanged++
				}
			}
		}
		if want && (changed < 130000 || lowerChanged < 24000) || !want && changed != 0 || !bytes.Equal(memory, memoryBefore) {
			t.Fatal("背景未延伸或寫入原版", name, changed, lowerChanged)
		}
		if !want {
			continue
		}
		for _, bad := range []string{"module", "graphics", "cache", "atlas", "mask", "flags", "display-cache", "resource-pointer", "flags-pointer", "upper-index", "lower-index", "rgb", "mode", "short-primary", "short-resources", "short-indexed", "short-rgb", "canvas"} {
			p, r, indices, col := append([]byte(nil), primary...), append([]byte(nil), resources...), append([]byte(nil), raw...), append([]byte(nil), rgb...)
			mode := uint8(0x13)
			canvas := image.NewRGBA(dst.Bounds())
			switch bad {
			case "module":
				p[0xc000] ^= 1
			case "graphics":
				p[0x3000] ^= 1
			case "cache":
				p[0xa000] ^= 1
			case "atlas":
				r[0x8100] ^= 1
			case "mask":
				r[0xd000] ^= 1
			case "flags":
				r[0x8006] ^= 1
			case "display-cache":
				for c := 0; c < 28; c++ {
					for row := 0; row < 3; row++ {
						p[0xe000+c*8+row] = 255
					}
				}
			case "resource-pointer":
				p[0xff2c] ^= 1
			case "flags-pointer":
				r[0x8000] ^= 1
			case "upper-index":
				indices[24*320+48] ^= 1
			case "lower-index":
				indices[101*320+271] ^= 1
			case "rgb":
				col[(101*320+271)*3] ^= 1
			case "mode":
				mode = 3
			case "short-primary":
				p = p[:100]
			case "short-resources":
				r = r[:100]
			case "short-indexed":
				indices = indices[:100]
			case "short-rgb":
				col = col[:100]
			case "canvas":
				canvas = image.NewRGBA(image.Rect(0, 0, 300, 200))
			}
			before := append([]byte(nil), canvas.Pix...)
			if s.Draw(canvas, mode, p, r, indices, col, image.Point{}) || !bytes.Equal(canvas.Pix, before) {
				t.Fatal("失配未整層退回", name, bad)
			}
		}
	}
	data, err := os.ReadFile(os.Getenv("ZELIARD_BACKGROUND_ART"))
	if err != nil {
		t.Fatal(err)
	}
	art, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	for y := 0; y < 234; y++ {
		for x := 0; x < 672; x++ {
			r, g, b, a := art.At(x*2126/672, y*740/234).RGBA()
			if s.art.RGBAAt(x, y) != (color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}) {
				t.Fatal("未直接取樣高清圖", x, y)
			}
		}
	}
	if _, err := NewTownBackground(root, os.Getenv("ZELIARD_SKY_REFERENCE"), os.Getenv("ZELIARD_SKY_PALETTE"), os.Getenv("ZELIARD_BACKGROUND_MEMORY"), filepath.Join(t.TempDir(), "missing.png")); err == nil {
		t.Fatal("缺圖未拒絕")
	}
}

func TestTownBackgroundNormalWalkAndDialogueReadOnly(t *testing.T) {
	root, project, s := backgroundTestInputs(t)
	fontPath, out := os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_BACKGROUND_OUTPUT")
	commandsPath, savePath := os.Getenv("ZELIARD_SKY_COMMANDS"), os.Getenv("ZELIARD_SKY_SAVE")
	if fontPath == "" || out == "" || commandsPath == "" || savePath == "" {
		t.Skip("須掛載正常鍵序、自建存檔、字型與輸出")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	save, err := os.ReadFile(savePath)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(save)) != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
		t.Fatal("起始存檔不符", err)
	}
	info, err := os.Stat(savePath)
	if err != nil {
		t.Fatal(err)
	}
	var machines [2]*oracle.Oracle
	for i := range machines {
		o, e := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if e != nil {
			t.Fatal(e)
		}
		machines[i] = o
		defer o.Close()
		o.SetDOSBoxCycles(3000)
		scratch := t.TempDir()
		p := filepath.Join(scratch, "SAVETEST.usr")
		if err = os.WriteFile(p, save, 0644); err != nil {
			t.Fatal(err)
		}
		if err = os.Chtimes(p, info.ModTime(), info.ModTime()); err != nil {
			t.Fatal(err)
		}
		o.SetScratch(scratch)
	}
	d, err := overlay.NewTownDialogue(machines[1], filepath.Join(root, "ZELRES2.SAR"), filepath.Join(project, "text/town-prototype.tsv"), font)
	if err != nil {
		t.Fatal(err)
	}
	commands, err := os.ReadFile(commandsPath)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	run := func(n uint64) {
		for _, o := range machines {
			if err := o.Run(n); err != nil {
				t.Fatal(err)
			}
		}
	}
	number := func(v string) uint64 {
		n, e := strconv.ParseUint(v, 0, 64)
		if e != nil {
			t.Fatal(e)
		}
		return n
	}
	var receipts []map[string]any
	for _, line := range strings.Split(string(commands), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
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
			a, b := machines[0], machines[1]
			mem := a.Bytes(oracle.Far(0, 0), 1<<20)
			if !bytes.Equal(mem, b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.IP() != b.IP() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) {
				t.Fatal("HD改動原版狀態", f[1])
			}
			pal := b.Palette()
			var palette []byte
			for _, c := range pal {
				palette = append(palette, c[:]...)
			}
			rgb, canvas := skyOriginal(b.Indexed(), palette)
			before := append([]byte(nil), canvas.Pix...)
			matched := s.Draw(canvas, b.VideoMode(), b.Bytes(oracle.Far(0x1b1, 0), 65536), b.Bytes(oracle.Far(0x11b1, 0), 65536), b.Indexed(), rgb, image.Point{})
			if f[1] == "cold-town" || f[1] == "restored" || f[1] == "citizen-near" || f[1] == "citizen-outside" {
				if !matched {
					t.Fatal("正常遠景未接入", f[1])
				}
			}
			if strings.HasPrefix(f[1], "citizen-") && (f[1] == "citizen-page" || f[1] == "citizen-last") && matched {
				t.Fatal("HD蓋住對話", f[1])
			}
			changed := 0
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := canvas.PixOffset(x, y)
					if !bytes.Equal(before[i:i+4], canvas.Pix[i:i+4]) {
						changed++
						if !image.Pt(x/3, y/3).In(townBackgroundBox) {
							t.Fatal("正常遠景越界", f[1])
						}
					}
				}
			}
			textDrawn := d.Draw(canvas, b.Indexed(), b.Palette(), image.Point{})
			if (f[1] == "citizen-page" || f[1] == "citizen-last") && !textDrawn {
				t.Fatal("HD模式居民中文遺失", f[1], d.State())
			}
			if !bytes.Equal(mem, b.Bytes(oracle.Far(0, 0), 1<<20)) {
				t.Fatal("HD繪圖寫入原版")
			}
			fp, e := os.Create(filepath.Join(out, f[1]+".png"))
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(fp, canvas)
			fp.Close()
			if e != nil {
				t.Fatal(e)
			}
			receipts = append(receipts, map[string]any{"name": f[1], "state_equal": true, "memory_sha256": fmt.Sprintf("%x", sha256.Sum256(mem)), "indexed_sha256": fmt.Sprintf("%x", sha256.Sum256(b.Indexed())), "steps": b.Steps(), "cycles": b.Cycles(), "background_matched": matched, "background_pixels_changed": changed, "outside_background_changed": 0, "town_drawn": textDrawn, "town": d.State(), "held_keys": b.HeldKeys()})
		default:
			t.Fatal("未知正常鍵序", line)
		}
	}
	if len(receipts) != 9 || len(machines[1].HeldKeys()) != 0 {
		t.Fatal("未完成正常遠景路徑")
	}
	data, err := json.MarshalIndent(map[string]any{"commands_sha256": fmt.Sprintf("%x", sha256.Sum256(commands)), "checkpoints": receipts, "scope": "正常城堡／冷F7／Muralla走路／居民兩頁／結束；HD差異只在遠景，原版1MiB相同"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
}

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

func skyTestInputs(t *testing.T) (string, string, *TownSky) {
	t.Helper()
	root, project := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_PROJECT")
	ref, pal, art := os.Getenv("ZELIARD_SKY_REFERENCE"), os.Getenv("ZELIARD_SKY_PALETTE"), os.Getenv("ZELIARD_SKY_ART")
	if root == "" || project == "" || ref == "" || pal == "" || art == "" {
		t.Skip("須掛載原版、遠景原生收據與本地手繪圖，skip不計驗收")
	}
	s, err := NewTownSky(root, ref, pal, art)
	if err != nil {
		t.Fatal(err)
	}
	return root, project, s
}

func skyOriginal(indexed, palette []byte) ([]byte, *image.RGBA) {
	rgb := make([]byte, 64000*3)
	dst := image.NewRGBA(image.Rect(0, 0, 960, 600))
	for i, v := range indexed {
		copy(rgb[i*3:i*3+3], palette[int(v)*3:int(v)*3+3])
	}
	for y := 0; y < 600; y++ {
		for x := 0; x < 960; x++ {
			i := (y/3*320 + x/3) * 3
			dst.SetRGBA(x, y, color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255})
		}
	}
	return rgb, dst
}

func TestTownSkySourcesPixelsAndFallback(t *testing.T) {
	root, project, s := skyTestInputs(t)
	pal, err := os.ReadFile(os.Getenv("ZELIARD_SKY_PALETTE"))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		path string
		want bool
	}{
		{"prose-v1/npc-normal-v2/cold-town", true}, {"prose-v1/npc-normal-v2/restored", true}, {"prose-v1/npc-normal-v2/citizen-near", true}, {"prose-v1/navigation-v1/citizen-near", true}, {"prose-v1/navigation-v1/citizen-outside", true}, {"prose-v1/npc-normal-v2/citizen-page", false}, {"prose-v1/npc-normal-v2/citizen-last", false},
	} {
		base := filepath.Join(project, "workplace", test.path)
		raw, e := os.ReadFile(base + ".indexed")
		if e != nil {
			t.Fatal(e)
		}
		seg, e := os.ReadFile(base + ".segment")
		if e != nil {
			t.Fatal(e)
		}
		rgb, dst := skyOriginal(raw, pal)
		before := append([]byte(nil), dst.Pix...)
		rawBefore := append([]byte(nil), raw...)
		if got := s.Draw(dst, 0x13, seg[0xc000:0xc040], raw, rgb, image.Point{}); got != test.want {
			t.Fatal("正常來源覆繪狀態不符", test.path, got)
		}
		changed := 0
		for y := 0; y < 600; y++ {
			for x := 0; x < 960; x++ {
				i := dst.PixOffset(x, y)
				if bytes.Equal(before[i:i+4], dst.Pix[i:i+4]) {
					continue
				}
				changed++
				if !image.Pt(x/3, y/3).In(townSkyBox) {
					t.Fatal("遠山HD越界", test.path, x, y)
				}
			}
		}
		if test.want && changed < 100000 || !test.want && changed != 0 || !bytes.Equal(raw, rawBefore) {
			t.Fatal("遠山未實際替換或改原版", test.path, changed)
		}
		if !test.want {
			continue
		}
		for _, bad := range []string{"first-pixel", "last-pixel", "rgb", "module", "mode", "short-indexed", "short-rgb", "canvas"} {
			input, col, prefix := append([]byte(nil), raw...), append([]byte(nil), rgb...), append([]byte(nil), seg[0xc000:0xc040]...)
			mode := uint8(0x13)
			canvas := image.NewRGBA(dst.Bounds())
			switch bad {
			case "first-pixel":
				input[24*320+48] ^= 1
			case "last-pixel":
				input[77*320+271] ^= 1
			case "rgb":
				col[(24*320+48)*3] ^= 1
			case "module":
				prefix[0] ^= 1
			case "mode":
				mode = 3
			case "short-indexed":
				input = input[:10]
			case "short-rgb":
				col = col[:10]
			case "canvas":
				canvas = image.NewRGBA(image.Rect(0, 0, 300, 200))
			}
			before := append([]byte(nil), canvas.Pix...)
			if s.Draw(canvas, mode, prefix, input, col, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
				t.Fatal("未知畫面未原子退回", test.path, bad)
			}
		}
	}
	data, err := os.ReadFile(os.Getenv("ZELIARD_SKY_ART"))
	if err != nil {
		t.Fatal(err)
	}
	art, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	unique := map[color.RGBA]bool{}
	for y := 0; y < 162; y++ {
		for x := 0; x < 672; x++ {
			r, g, b, a := art.At(x*1973/672, 207+y*382/162).RGBA()
			want := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
			if s.art.RGBAAt(x, y) != want {
				t.Fatal("不是直接高解析度取樣", x, y)
			}
			unique[want] = true
		}
	}
	if len(unique) < 1000 {
		t.Fatal("HD圖面未保留足夠繪畫細節", len(unique))
	}
	if _, err := NewTownSky(root, os.Getenv("ZELIARD_SKY_REFERENCE"), os.Getenv("ZELIARD_SKY_PALETTE"), filepath.Join(t.TempDir(), "missing.png")); err == nil {
		t.Fatal("缺圖未拒絕")
	}
}

func TestTownSkyNormalWalkAndDialogueReadOnly(t *testing.T) {
	root, project, s := skyTestInputs(t)
	fontPath, out := os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_SKY_OUTPUT")
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
			matched := s.Draw(canvas, b.VideoMode(), b.Bytes(oracle.Far(0x1b1, 0xc000), 64), b.Indexed(), rgb, image.Point{})
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
						if !image.Pt(x/3, y/3).In(townSkyBox) {
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
			receipts = append(receipts, map[string]any{"name": f[1], "state_equal": true, "memory_sha256": fmt.Sprintf("%x", sha256.Sum256(mem)), "indexed_sha256": fmt.Sprintf("%x", sha256.Sum256(b.Indexed())), "steps": b.Steps(), "cycles": b.Cycles(), "sky_matched": matched, "sky_pixels_changed": changed, "outside_sky_changed": 0, "town_drawn": textDrawn, "town": d.State(), "held_keys": b.HeldKeys()})
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

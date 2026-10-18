package hd

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
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

func TestTownSceneryNormalPathsReadOnly(t *testing.T) {
	paths := os.Getenv("ZELIARD_SCENERY_COMMANDS")
	out := os.Getenv("ZELIARD_SCENERY_OUTPUT")
	if paths == "" || out == "" {
		t.Skip("須掛載正常鍵序與配對輸出；skip不算驗收")
	}
	for _, path := range filepath.SplitList(paths) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			runSceneryNormalPair(t, path, filepath.Join(out, strings.TrimSuffix(filepath.Base(path), ".commands")))
		})
	}
}
func runSceneryNormalPair(t *testing.T, commandsPath, out string) {
	s, _ := sceneryTestInputs(t)
	root, project := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_PROJECT")
	fontPath := os.Getenv("ZELIARD_FONT")
	savePath := os.Getenv("ZELIARD_SKY_SAVE")
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
			matched := s.Draw(canvas, b.VideoMode(), b.Bytes(oracle.Far(0x1b1, 0), 65536), b.Bytes(oracle.Far(0x11b1, 0), 65536), b.Bytes(oracle.Far(0x21b1, 0), 65536), b.Indexed(), rgb, image.Point{}, true)
			if f[1] == "restored" || f[1] == "citizen-near" || f[1] == "citizen-outside" || (strings.HasPrefix(f[1], "walk-left-") && f[1] != "walk-left-7") || f[1] == "cave-near" {
				if !matched {
					t.Fatal("正常前景未接入", f[1])
				}
			}
			if (f[1] == "cold-town" || f[1] == "walk-left-7" || f[1] == "cave-entered" || f[1] == "citizen-page" || f[1] == "citizen-last") && matched {
				t.Fatal("HD蓋住對話", f[1])
			}
			changed := 0
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := canvas.PixOffset(x, y)
					if !bytes.Equal(before[i:i+4], canvas.Pix[i:i+4]) {
						changed++
						if !image.Pt(x/3, y/3).In(image.Rect(48, 78, 272, 142)) {
							t.Fatal("正常前景越界", f[1])
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
			receipts = append(receipts, map[string]any{"name": f[1], "state_equal": true, "memory_sha256": fmt.Sprintf("%x", sha256.Sum256(mem)), "indexed_sha256": fmt.Sprintf("%x", sha256.Sum256(b.Indexed())), "steps": b.Steps(), "cycles": b.Cycles(), "scenery_matched": matched, "scenery_pixels_changed": changed, "outside_scenery_changed": 0, "town_drawn": textDrawn, "town": d.State(), "held_keys": b.HeldKeys()})
		default:
			t.Fatal("未知正常鍵序", line)
		}
	}
	if len(receipts) < 7 || len(machines[1].HeldKeys()) != 0 {
		t.Fatal("未完成正常前景路徑")
	}
	data, err := json.MarshalIndent(map[string]any{"commands_sha256": fmt.Sprintf("%x", sha256.Sum256(commands)), "checkpoints": receipts, "scope": "正常城堡／冷F7／Muralla走路／居民兩頁／結束；HD差異只在城鎮前景，原版1MiB相同"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
}

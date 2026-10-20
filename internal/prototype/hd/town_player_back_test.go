package hd

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
)

func TestTownBackNormalSource(t *testing.T) {
	scene, _, _, study := castleTestInputs(t)
	for _, sample := range []struct {
		folder, name string
		bit          byte
	}{{"town-neutral-normal-v3", "observed-back-bit0", 0}, {"town-neutral-normal-v4", "observed-back-bit1", 1}} {
		t.Run(sample.name, func(t *testing.T) {
			read := func(suffix string) []byte {
				t.Helper()
				b, err := os.ReadFile(filepath.Join(study, sample.folder, sample.name+suffix))
				if err != nil {
					t.Fatal(err)
				}
				return b
			}
			m, raw, palette := read(".memory"), read(".indexed"), read(".palette")
			p, r, e := m[0x1b10:0x11b10], m[0x11b10:0x21b10], m[0x21b10:0x31b10]
			if p[0xe7] != 4 || p[0xc2]&1 != sample.bit || p[0x83] >= 27 {
				t.Fatal("不是已驗正常背向來源")
			}
			rgb, _ := skyOriginal(raw, palette)
			if _, err := scene.decode(0x13, p, r, e, raw, rgb); err != nil {
				t.Fatal("正常14336原生像素失配", err)
			}
		})
	}
}

func TestTownBackLiveEntryReadOnly(t *testing.T) {
	scene, _, _, study := castleTestInputs(t)
	project, root := os.Getenv("ZELIARD_PROJECT"), os.Getenv("ZELIARD_ORIG")
	player, err := NewTownPlayer(scene, filepath.Join(project, "workplace/town-hd-v1/foreground-v1"))
	if err != nil {
		t.Fatal(err)
	}
	if err = player.LoadBackPose(filepath.Join(study, "town-neutral-complete-v1")); err != nil {
		t.Fatal(err)
	}
	actors, err := NewTownActors(player, filepath.Join(study, "town-npc-art"))
	if err != nil {
		t.Fatal(err)
	}
	var points []map[string]any
	for bit := 0; bit < 2; bit++ {
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
		frames := NewTownFrames(b, scene)
		var samples []map[string]any
		b.OnCall(oracle.Far(0x1b1, 0x68ba), func(o *oracle.Oracle) {
			f := frames.complete
			if f == nil || f.Primary[0xe7] != 4 || f.Steps != o.Steps() {
				return
			}
			if int(f.Primary[0xc2]&1) != bit {
				t.Fatal("正常進門朝向不同")
			}
			before := o.Bytes(oracle.Far(0, 0), 1<<20)
			_, canvas := skyOriginal(f.Indexed, f.Palette)
			if !scene.Draw(canvas, 0x13, f.Primary, f.Resources, f.Extra, f.Indexed, f.RGB, image.Point{}, true) ||
				!actors.Draw(canvas, 0x13, f.Primary, f.Resources, f.Extra, f.Indexed, f.RGB, f.Palette, image.Point{}) ||
				actors.State()["residents"] != 0 {
				t.Fatal("正常背向共同HD合成拒絕")
			}
			if !bytes.Equal(before, o.Bytes(oracle.Far(0, 0), 1<<20)) {
				t.Fatal("完成回呼合成改變原版RAM")
			}
			samples = append(samples, map[string]any{"steps": f.Steps, "cycles": f.Cycles, "phase": 4,
				"bit": bit, "memory_sha256": townBackgroundHash(before), "hd_sha256": townBackgroundHash(canvas.Pix), "residents": 0})
		})
		equal := func() {
			if !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) ||
				!bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Regs() != b.Regs() ||
				a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) ||
				!reflect.DeepEqual(a.KeyReads(), b.KeyReads()) || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) {
				t.Fatal("正常原版完整狀態不同", bit)
			}
		}
		var inputs []map[string]any
		run := func(n uint64) {
			for _, o := range machines {
				if err := o.Run(n); err != nil {
					t.Fatal(err)
				}
			}
			equal()
		}
		key := func(scan uint8, n uint64) {
			inputs = append(inputs, map[string]any{"scan": scan, "instructions": n})
			for _, o := range machines {
				o.KeyDown(scan)
			}
			run(n)
			for _, o := range machines {
				o.KeyUp(scan)
			}
		}
		run(12000000)
		for i := 0; i < 6; i++ {
			key(0x39, 50000)
			run(7950000)
		}
		if bit == 0 {
			key(0x4d, 2000000)
		} else {
			key(0x4d, 3000000)
			key(0x4b, 1000000)
		}
		run(250000)
		key(0x48, 250000)
		run(10000000)
		if len(samples) == 0 || len(b.HeldKeys()) != 0 {
			t.Fatal("未由正常進門走到背向HD", bit)
		}
		points = append(points, map[string]any{"bit": bit, "inputs": inputs, "samples": samples, "full_state_equal": true})
	}
	if output := os.Getenv("ZELIARD_BACK_TEST_OUTPUT"); output != "" {
		data, err := json.MarshalIndent(map[string]any{"status": "PASS_NORMAL_BACK_ENTRY_READ_ONLY", "cases": points,
			"scope": "兩條冷新局正常鍵序及完成回呼合成，非實際視窗、未注入姿態或位置"}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(output, "live-entry.json"), append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTownBackCompositeAndFallback(t *testing.T) {
	scene, _, _, study := castleTestInputs(t)
	project := os.Getenv("ZELIARD_PROJECT")
	player, err := NewTownPlayer(scene, filepath.Join(project, "workplace/town-hd-v1/foreground-v1"))
	if err != nil {
		t.Fatal(err)
	}
	old := player.art
	directory := filepath.Join(study, "town-neutral-complete-v1")
	if err := player.LoadBackPose(directory); err != nil {
		t.Fatal(err)
	}
	if player.art[4] == nil || player.art[4] != player.art[9] || player.art[4].Bounds() != image.Rect(0, 0, 48, 72) {
		t.Fatal("兩朝向背向圖未完整載入")
	}
	for _, key := range []int{0, 1, 2, 3, 5, 6, 7, 8} {
		if old[key] != player.art[key] {
			t.Fatal("改變既有行走素材", key)
		}
	}
	loaded := player.art
	bad := t.TempDir()
	if err := os.WriteFile(filepath.Join(bad, "town-player-back-v1.png"), []byte("corrupt"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, directory := range []string{bad, filepath.Join(bad, "missing")} {
		if player.LoadBackPose(directory) == nil || player.art != loaded {
			t.Fatal("載入失敗改變既有圖組")
		}
	}
	actors, err := NewTownActors(player, filepath.Join(study, "town-npc-art"))
	if err != nil {
		t.Fatal(err)
	}
	for bit, folder := range []string{"town-neutral-normal-v3", "town-neutral-normal-v4"} {
		t.Run(folder, func(t *testing.T) {
			name := "observed-back-bit" + string(rune('0'+bit))
			read := func(suffix string) []byte {
				t.Helper()
				b, e := os.ReadFile(filepath.Join(study, folder, name+suffix))
				if e != nil {
					t.Fatal(e)
				}
				return b
			}
			memory, raw, palette := read(".memory"), read(".indexed"), read(".palette")
			identity := townBackgroundHash(memory)
			p, r, e := memory[0x1b10:0x11b10], memory[0x11b10:0x21b10], memory[0x21b10:0x31b10]
			rgb, original := skyOriginal(raw, palette)
			canvas := image.NewRGBA(original.Bounds())
			copy(canvas.Pix, original.Pix)
			if !scene.Draw(canvas, 0x13, p, r, e, raw, rgb, image.Point{}, true) {
				t.Fatal("正常背向前景拒絕")
			}
			before := cloneTownBytes(canvas.Pix)
			unloaded := &TownPlayer{scene: scene, art: old}
			if unloaded.Draw(canvas, 0x13, p, r, e, raw, rgb, palette, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
				t.Fatal("未載入背向圖仍改畫面")
			}
			if !actors.Draw(canvas, 0x13, p, r, e, raw, rgb, palette, image.Point{}) || actors.State()["residents"] != 0 {
				t.Fatal("背向共同合成未依原版隱藏居民", actors.State())
			}
			if bytes.Equal(before, canvas.Pix) || townBackgroundHash(memory) != identity {
				t.Fatal("未畫HD或改變原版來源")
			}
			box := image.Rect((48+int(p[0x83])*8)*3, 354, (48+int(p[0x83])*8)*3+48, 426)
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					if box.Min.X <= x && x < box.Max.X && box.Min.Y <= y && y < box.Max.Y {
						continue
					}
					i := canvas.PixOffset(x, y)
					if !bytes.Equal(before[i:i+4], canvas.Pix[i:i+4]) {
						t.Fatal("背向人物層寫入區外", x, y)
					}
				}
			}
			if output := os.Getenv("ZELIARD_BACK_TEST_OUTPUT"); output != "" {
				for suffix, im := range map[string]*image.RGBA{"-original": original, "-hd": canvas} {
					f, err := os.Create(filepath.Join(output, name+suffix+".png"))
					if err != nil {
						t.Fatal(err)
					}
					err = png.Encode(f, im)
					f.Close()
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			// 拒絕已驗條件之外的居民配置，不能把記錄表當成畫面。
			for _, mutate := range []func(){
				func() { p[0xc01c+48*8] = 253 },
				func() { p[0xe7] = 5 },
				func() { r[0x6000] ^= 1 },
			} {
				copy(memory, read(".memory"))
				mutate()
				before := cloneTownBytes(canvas.Pix)
				if actors.Draw(canvas, 0x13, p, r, e, raw, rgb, palette, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
					t.Fatal("未知來源未原子拒絕")
				}
			}
		})
	}
}

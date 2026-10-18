package hd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
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

func TestTownFramesNormalPathsReadOnly(t *testing.T) {
	paths, out := os.Getenv("ZELIARD_SCENERY_COMMANDS"), os.Getenv("ZELIARD_SCENERY_OUTPUT")
	if paths == "" || out == "" {
		t.Skip("須掛載正常鍵序及輸出；skip不算驗收")
	}
	for _, path := range filepath.SplitList(paths) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			runTownFramesPair(t, path, filepath.Join(out, strings.TrimSuffix(filepath.Base(path), ".commands")))
		})
	}
}
func runTownFramesPair(t *testing.T, path, out string) {
	scene, _ := sceneryTestInputs(t)
	root, project := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_PROJECT")
	savePath := os.Getenv("ZELIARD_SKY_SAVE")
	save, err := os.ReadFile(savePath)
	if err != nil {
		t.Fatal(err)
	}
	if townBackgroundHash(save) != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
		t.Fatal("存檔不符")
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
		if e = os.WriteFile(p, save, 0644); e != nil {
			t.Fatal(e)
		}
		if e = os.Chtimes(p, info.ModTime(), info.ModTime()); e != nil {
			t.Fatal(e)
		}
		o.SetScratch(scratch)
	}
	a, b := machines[0], machines[1]
	font, err := bitmap.Load(os.Getenv("ZELIARD_FONT"))
	if err != nil {
		t.Fatal(err)
	}
	dialogue, err := overlay.NewTownDialogue(b, filepath.Join(root, "ZELRES2.SAR"), filepath.Join(project, "text/town-prototype.tsv"), font)
	if err != nil {
		t.Fatal(err)
	}
	frames := NewTownFrames(b, scene)
	var ground *TownGround
	if artwork := os.Getenv("ZELIARD_GROUND_ART"); artwork != "" {
		ground, err = NewTownGround(scene, filepath.Join(project, "workplace/town-hd-v1/foreground-v1/normal-v1/restored.indexed"), artwork)
		if err != nil {
			t.Fatal(err)
		}
	}
	canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
	var selected *TownPresentation
	draws, held, uncovered, modal := 0, 0, 0, 0
	loadInProgress := false
	var loadRequested uint64
	equal := func(label string) {
		if !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Regs() != b.Regs() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) {
			t.Fatal("完整原版狀態不相同", label)
		}
	}
	present := func() {
		draws++
		raw := b.Indexed()
		_, _, rgb := b.ScreenRGB()
		dialogue.Draw(canvas, raw, b.Palette(), image.Point{})
		blocked := dialogue.State()["key"] != ""
		in := frames.read(raw, rgb)
		_, nativeError := scene.decode(in.mode, in.primary, in.resources, in.extra, raw, rgb)
		selected = frames.Present(raw, rgb, blocked)
		if selected != nil {
			if blocked {
				t.Fatal("快照蓋住中文對話")
			}
			if _, err := scene.decode(in.mode, selected.Primary, selected.Resources, selected.Extra, selected.Indexed, selected.RGB); err != nil {
				t.Fatal("完整呈現未通過像素守衛", err)
			}
			if nativeError != nil {
				held++
			}
			for y := 0; y < 200; y++ {
				for x := 0; x < 320; x++ {
					if image.Pt(x, y).In(townPresentationBox) {
						continue
					}
					i := y*320 + x
					if raw[i] != selected.Indexed[i] || !bytes.Equal(rgb[i*3:i*3+3], selected.RGB[i*3:i*3+3]) {
						t.Fatal("呈現改變畫帶外原版")
					}
				}
			}
		} else if blocked {
			modal++
		} else if nativeError != nil && townFrameSource(in) && frames.captured > 0 && !loadInProgress {
			uncovered++
			t.Fatal("正常路徑存在未涵蓋中途畫格", b.Steps(), b.IP())
		}
	}
	run := func(n uint64) {
		for n > 0 {
			step := uint64(50000)
			if n < step {
				step = n
			}
			if e := a.Run(step); e != nil {
				t.Fatal(e)
			}
			if e := b.Run(step); e != nil {
				t.Fatal(e)
			}
			present()
			n -= step
		}
	}
	num := func(s string) uint64 {
		n, e := strconv.ParseUint(s, 0, 64)
		if e != nil {
			t.Fatal(e)
		}
		return n
	}
	commands, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var points []map[string]any
	for _, line := range strings.Split(string(commands), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "run":
			run(num(f[1]))
		case "key":
			if num(f[1]) == 0x41 {
				loadInProgress = true
				loadRequested = b.Steps()
			}
			for _, o := range machines {
				o.KeyDown(uint8(num(f[1])))
			}
			run(num(f[2]))
			for _, o := range machines {
				o.KeyUp(uint8(num(f[1])))
			}
		case "shot":
			equal(f[1])
			if strings.HasSuffix(f[1], "restored") {
				if selected == nil || selected.Steps <= loadRequested {
					t.Fatal("讀檔未建立新的完成來源", f[1], frames.State())
				}
				loadInProgress = false
			}
			if strings.Contains(f[1], "load-") && selected != nil {
				t.Fatal("城鎮快照蓋住F7選單", f[1])
			}
			before := b.Bytes(oracle.Far(0, 0), 1<<20)
			if selected != nil && !scene.Draw(canvas, b.VideoMode(), selected.Primary, selected.Resources, selected.Extra, selected.Indexed, selected.RGB, image.Point{}, true) {
				t.Fatal("正常完成來源HD失配", f[1])
			}
			if ground != nil {
				if selected != nil {
					if !ground.Draw(canvas, b.VideoMode(), selected.Primary, selected.Resources, selected.Extra, selected.Indexed, selected.RGB, selected.Palette, image.Point{}) {
						t.Fatal("正常完成來源地面HD失配", f[1])
					}
				} else if strings.HasSuffix(f[1], "cave-entered") {
					in := frames.read(b.Indexed(), nil)
					_, _, rgb := b.ScreenRGB()
					unchanged := cloneTownBytes(canvas.Pix)
					if ground.Draw(canvas, in.mode, in.primary, in.resources, in.extra, in.indexed, rgb, in.palette, image.Point{}) || !bytes.Equal(unchanged, canvas.Pix) {
						t.Fatal("地面HD蓋住洞窟")
					}
				}
			}
			if !bytes.Equal(before, b.Bytes(oracle.Far(0, 0), 1<<20)) {
				t.Fatal("HD改寫原版")
			}
			if f[1] == "citizen-page" || f[1] == "citizen-last" {
				if selected != nil || !dialogue.Draw(canvas, b.Indexed(), b.Palette(), image.Point{}) {
					t.Fatal("居民中文被完整畫格破壞", f[1])
				}
			}
			if strings.HasSuffix(f[1], "cave-entered") && selected != nil {
				t.Fatal("城鎮快照蓋住洞窟")
			}
			points = append(points, map[string]any{"name": f[1], "steps": b.Steps(), "cycles": b.Cycles(), "registers": b.Regs(), "state_equal": true, "memory_sha256": townBackgroundHash(before), "presentation": frames.State(), "town": dialogue.State()})
		default:
			t.Fatal("未知正常命令", line)
		}
	}
	equal("final")
	if held == 0 || uncovered != 0 || len(b.HeldKeys()) != 0 {
		t.Fatal("正常呈現驗收不足", held, uncovered)
	}
	data, e := json.MarshalIndent(map[string]any{"commands_sha256": townBackgroundHash(commands), "draws": draws, "held_complete": held, "uncovered": uncovered, "modal_frames": modal, "checkpoints": points, "captured": frames.captured, "refused": frames.refused, "held_keys": b.HeldKeys(), "scope": "兩個獨立dosgolem正常鍵序；各檢查點完整原版狀態與每50k完整呈現；F7選單用當次原版，恢復必須新完成來源；非GUI或其他城鎮驗收"}, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.MkdirAll(out, 0755); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, '\n'), 0644); e != nil {
		t.Fatal(e)
	}
	t.Log(fmt.Sprintf("%d呈現／%d完整保留／未涵蓋%d", draws, held, uncovered))
}

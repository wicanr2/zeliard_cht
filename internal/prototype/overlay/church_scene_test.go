package overlay

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

func TestHUDChurchNormalPathReadOnly(t *testing.T) {
	root, fontPath, savePath := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_BANK_SAVE")
	out, sourceDir := os.Getenv("ZELIARD_CHURCH_SCENE_OUTPUT"), os.Getenv("ZELIARD_CHURCH_SOURCE")
	if root == "" || fontPath == "" || savePath == "" || out == "" || sourceDir == "" {
		t.Skip("須掛載原版、字型、自建存檔與正常教堂來源")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	commands, err := os.ReadFile("../../../tools/church.commands")
	if err != nil {
		t.Fatal(err)
	}
	save, err := os.ReadFile(savePath)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(savePath)
	if err != nil {
		t.Fatal(err)
	}
	var machines [2]*oracle.Oracle
	for i := range machines {
		machines[i], err = oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if err != nil {
			t.Fatal(err)
		}
		defer machines[i].Close()
		machines[i].SetDOSBoxCycles(3000)
		scratch := t.TempDir()
		path := filepath.Join(scratch, "SAVETEST.usr")
		if err = os.WriteFile(path, save, 0644); err != nil {
			t.Fatal(err)
		}
		if err = os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
			t.Fatal(err)
		}
		machines[i].SetScratch(scratch)
	}
	hud, err := NewHUD(machines[1], root, "../../../text/hud-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	var events, checkpoints []map[string]any
	entryCount := 0
	machines[1].OnCall(oracle.Far(0x1b1, 0x22cd), func(o *oracle.Oracle) {
		if o.DSReg() != 0x1b1 || o.SI() != 0xa2a6 {
			return
		}
		entryCount++
		events = append(events, map[string]any{"stage": "entry", "steps": o.Steps(), "cycles": o.Cycles(), "SI": o.SI(), "source_sha256": fmtHash(o.Bytes(oracle.Far(0x1b1, 0xa2a6), 14)), "prefix_sha256": fmtHash(o.Bytes(oracle.Far(0x1b1, 0xa000), 64)), "font_pointer": o.Word(oracle.Far(0x1b1, 0xf504)), "font_sha256": fmtHash(o.Bytes(oracle.Far(0x1b1, 0xf856), 768)), "code_sha256": fmtHash(o.Bytes(oracle.Far(0x1b1, 0x22bf), 0xc7))})
	})
	lastCompleted := 0
	machines[1].OnCall(oracle.Far(0x1b1, 0x2344), func(o *oracle.Oracle) {
		changed := hud.Completed != lastCompleted
		lastCompleted = hud.Completed
		s := hud.fields["scene"]
		if !changed || s == nil || s.key != "zelres2.file.01a448" || o.SI() != 0xa2b4 {
			return
		}
		events = append(events, map[string]any{"stage": "complete", "steps": o.Steps(), "cycles": o.Cycles(), "SI": o.SI(), "key": s.key, "runtime_base": s.base, "runtime_header": s.header, "raw_sha256": s.rawHash, "prefix_sha256": s.prefixHash, "safe_box": s.box, "bitmap_sha256": fmtHash(labelCrop(o.Indexed(), s.box)), "admitted": true})
	})
	number := func(s string) uint64 {
		n, e := strconv.ParseUint(s, 0, 64)
		if e != nil {
			t.Fatal(e)
		}
		return n
	}
	run := func(n uint64) {
		for _, o := range machines {
			if e := o.Run(n); e != nil {
				t.Fatal(e)
			}
		}
	}
	finished := false
	for _, line := range strings.Split(string(commands), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "run":
			run(number(f[1]))
		case "down", "up":
			for _, o := range machines {
				if f[0] == "down" {
					o.KeyDown(uint8(number(f[1])))
				} else {
					o.KeyUp(uint8(number(f[1])))
				}
			}
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
			memory := a.Bytes(oracle.Far(0, 0), 1<<20)
			if !bytes.Equal(memory, b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Regs() != b.Regs() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() {
				t.Fatal("HUD改動原版狀態", f[1])
			}
			var original struct {
				Steps, Cycles uint64
				Memory        string `json:"memory_sha256"`
				Indexed       string `json:"indexed_sha256"`
			}
			data, e := os.ReadFile(filepath.Join(sourceDir, f[1]+".json"))
			if e != nil {
				t.Fatal(e)
			}
			if e = json.Unmarshal(data, &original); e != nil {
				t.Fatal(e)
			}
			if b.Steps() != original.Steps || b.Cycles() != original.Cycles || fmtHash(memory) != original.Memory || fmtHash(b.Indexed()) != original.Indexed {
				t.Fatal("正常來源未重現", f[1])
			}
			s := hud.fields["scene"]
			if s == nil {
				t.Fatal("沒有已完成場景", f[1])
			}
			wantChurch := f[1] == "church-entered" || f[1] == "church-healing-page2" || f[1] == "church-second-visit" || f[1] == "church-second-final"
			if (s.key == "zelres2.file.01a448") != wantChurch {
				t.Fatal("場景來源未隨正常進出改變", f[1], s.key)
			}
			if wantChurch && (s.base != 0xa000 || s.header != 0xa2a6 || string(s.text) != "教堂") {
				t.Fatal("教堂來源或譯文不符")
			}
			isolated := *hud
			isolated.fields = map[string]*hudSource{"scene": s}
			canvas := inventoryOriginalCanvas(b)
			before := append([]byte(nil), canvas.Pix...)
			if isolated.Draw(canvas, b.Indexed(), b.Palette(), image.Point{}) != 1 {
				t.Fatal("完整名稱未顯示", f[1])
			}
			box := image.Rectangle{Min: s.box.Min.Mul(3), Max: s.box.Max.Mul(3)}
			changed := 0
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := canvas.PixOffset(x, y)
					if !bytes.Equal(before[i:i+4], canvas.Pix[i:i+4]) {
						if !image.Pt(x, y).In(box) {
							t.Fatal("名稱越界")
						}
						changed++
					}
				}
			}
			if changed == 0 || !bytes.Equal(memory, b.Bytes(oracle.Far(0, 0), 1<<20)) {
				t.Fatal("名稱未繪製或寫入原版")
			}
			bad := append([]byte(nil), b.Indexed()...)
			bad[s.box.Min.Y*320+s.box.Min.X] ^= 128
			blank := image.NewRGBA(canvas.Bounds())
			prior := append([]byte(nil), blank.Pix...)
			if isolated.Draw(blank, bad, b.Palette(), image.Point{}) != 0 || !bytes.Equal(prior, blank.Pix) {
				t.Fatal("未知字形未原子退回")
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
			checkpoints = append(checkpoints, map[string]any{"name": f[1], "steps": b.Steps(), "cycles": b.Cycles(), "scene_key": s.key, "state_equal": true, "source_equal": true, "outside_changed": 0, "changed_pixels": changed, "gold": b.Word(oracle.Far(0x1b1, 0x86))})
			if f[1] == "church-actually-outside" {
				finished = true
			}
		default:
			t.Fatal("未知正常鍵序", line)
		}
		if finished {
			break
		}
	}
	if !finished || len(checkpoints) != 8 || entryCount != 2 || len(events) != 4 || len(machines[1].HeldKeys()) != 0 {
		t.Fatal("教堂名稱驗證未完整", len(checkpoints), entryCount, len(events))
	}
	data, e := json.MarshalIndent(map[string]any{"status": "PASS_NORMAL_CHURCH_SCENE_HUD_SOURCE", "input_exe_sha256": fmtHash(mustReadChurchScene(t, filepath.Join(root, "ZELIARD.EXE"))), "source_save_sha256": fmtHash(save), "source_save_mtime_ns": info.ModTime().UnixNano(), "commands_sha256": fmtHash(commands), "checkpoints": checkpoints, "events": events, "scope": "cold normal F7, two Church entries and departures; HUD scene-only independent display with complete native state and prior source snapshots equal; no GUI claim"}, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, 10), 0644); e != nil {
		t.Fatal(e)
	}
}

func mustReadChurchScene(t *testing.T, path string) []byte {
	t.Helper()
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	return b
}

package overlay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
)

// 真正GUI的IRQ及RunCycles日誌由冷啟動重播，沒有還原或注入。
func TestExhaustionReturnGUIJournalReadOnly(t *testing.T) {
	root, project, study, font := exhaustionInputs(t)
	gui, out := os.Getenv("ZELIARD_RETURN_GUI"), os.Getenv("ZELIARD_RETURN_GUI_OUTPUT")
	if gui == "" || out == "" {
		t.Skip("須提供已完成GUI及新的重播輸出")
	}
	read := func(path string) []byte {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	var targets struct {
		Records []struct {
			Label         string
			Steps, Cycles uint64
			Memory        string `json:"oracle_memory_sha256"`
			Indexed       string `json:"oracle_indexed_sha256"`
			Palette       string `json:"oracle_palette_sha256"`
			Dialogue      map[string]any
			Sage          bool `json:"sage_hd_matched"`
			Pose          int  `json:"sage_hd_pose"`
		}
	}
	if e := json.Unmarshal(read(filepath.Join(gui, "targets.json")), &targets); e != nil || len(targets.Records) < 4 {
		t.Fatal("缺少真正GUI目標", e)
	}
	wanted := map[uint64]bool{}
	for _, r := range targets.Records {
		wanted[r.Steps] = true
	}
	type record struct {
		Kind          string      `json:"kind"`
		CyclesPerMS   uint64      `json:"cycles_per_ms"`
		SampleRate    int         `json:"audio_sample_rate"`
		InitialMemory string      `json:"initial_memory_sha256"`
		InitialRegs   oracle.Regs `json:"initial_regs"`
		Scan          uint8       `json:"scan"`
		Pressed       bool        `json:"pressed"`
		Budget        uint64      `json:"cycle_budget"`
		BeforeSteps   uint64      `json:"before_steps"`
		BeforeCycles  uint64      `json:"before_cycles"`
		AfterSteps    uint64      `json:"after_steps"`
		AfterCycles   uint64      `json:"after_cycles"`
		AfterRegs     oracle.Regs `json:"after_regs"`
		Audio         bool        `json:"audio_render"`
		Error         string      `json:"error"`
	}
	data := read(filepath.Join(gui, "guest-inputs.jsonl"))
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 || len(lines) > 12000 {
		t.Fatal("日誌數量")
	}
	var header record
	if e := json.Unmarshal([]byte(lines[0]), &header); e != nil || header.Kind != "header" {
		t.Fatal("header", e)
	}
	savePath := os.Getenv("ZELIARD_CAVE_SAVE")
	save := read(savePath)
	info, e := os.Stat(savePath)
	if e != nil || fmtHash(save) != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
		t.Fatal("自建存檔", e)
	}
	var machines [2]*oracle.Oracle
	var audio [2]*oracle.Audio
	var saves []string
	for i := range machines {
		o, e := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if e != nil {
			t.Fatal(e)
		}
		machines[i] = o
		t.Cleanup(o.Close)
		scratch := t.TempDir()
		p := filepath.Join(scratch, "SAVETEST.usr")
		if e = os.WriteFile(p, save, 0644); e != nil {
			t.Fatal(e)
		}
		if e = os.Chtimes(p, info.ModTime(), info.ModTime()); e != nil {
			t.Fatal(e)
		}
		o.SetScratch(scratch)
		saves = append(saves, p)
		o.SetDOSBoxCycles(header.CyclesPerMS)
		o.TraceFiles()
		audio[i] = o.NewAudio(header.SampleRate)
		if fmtHash(o.Bytes(oracle.Far(0, 0), 1<<20)) != header.InitialMemory || o.Regs() != header.InitialRegs {
			t.Fatal("GUI冷啟動身份不符")
		}
	}
	a, b := machines[0], machines[1]
	d, e := NewDialogue(b, filepath.Join(root, "ZELRES2.SAR"), filepath.Join(project, "text/dialogue-prototype.tsv"), font)
	if e != nil {
		t.Fatal(e)
	}
	sage := exhaustionSage(t, project, study)
	// 原版側獨立記錄原字元與指標，Python像素工具不依譯文狀態推定進度。
	sar := read(filepath.Join(root, "ZELRES2.SAR"))
	var glyphs []map[string]any
	a.OnCall(oracle.Far(0x1b1, 0x711b), func(o *oracle.Oracle) {
		if !bytes.Equal(o.Bytes(oracle.Far(0x1b1, 0xa000), 64), sar[0x1bcc6:0x1bcc6+64]) {
			return
		}
		pos := int(o.Word(oracle.Far(0x1b1, 0xff4c))) - 1
		for _, s := range [][2]int{{0x1d72e, 0x1d76c}, {0x1d76c, 0x1d79c}, {0x1d79c, 0x1d7b2}, {0x1d7b2, 0x1d7d5}} {
			lo, hi := 0xa000+s[0]-0x1bcc6, 0xa000+s[1]-0x1bcc6
			if pos < lo || pos >= hi {
				continue
			}
			offset := s[0] + pos - lo
			if o.DSReg() != 0x1b1 || o.AX()>>8 != 1 || byte(o.AX()) != sar[offset] || !bytes.Equal(o.Bytes(oracle.Far(0x1b1, 0x711b), 5), []byte{0x2e, 0xff, 0x16, 0x22, 0x20}) {
				t.Fatal("原版字元來源不符")
			}
			if len(glyphs) >= 4096 {
				t.Fatal("字元事件預算")
			}
			glyphs = append(glyphs, map[string]any{"steps": o.Steps(), "cycles": o.Cycles(), "key": fmt.Sprintf("zelres2.file.%06x", s[0]), "file_offset": offset, "source_pointer": pos, "byte": byte(o.AX()), "bx": o.BX(), "cx": o.CX(), "ds": o.DSReg()})
		}
	})
	if e = os.Mkdir(out, 0755); e != nil {
		t.Fatal(e)
	}
	var captures []map[string]any
	for n, line := range lines[1:] {
		var r record
		if e = json.Unmarshal([]byte(line), &r); e != nil {
			t.Fatal(e)
		}
		for i, o := range machines {
			if o.Steps() != r.BeforeSteps || o.Cycles() != r.BeforeCycles {
				t.Fatal("before", n+2, i)
			}
			switch r.Kind {
			case "key":
				if r.Pressed {
					o.KeyDown(r.Scan)
				} else {
					o.KeyUp(r.Scan)
				}
			case "run":
				if r.Budget > 5000000 || o.Steps() > 400000000 {
					t.Fatal("日誌預算")
				}
				err := o.RunCycles(r.Budget)
				message := ""
				if err != nil {
					message = err.Error()
				}
				if r.Audio {
					audio[i].Render()
				}
				if message != r.Error || o.Regs() != r.AfterRegs {
					t.Fatal("after regs", n+2, i)
				}
			default:
				t.Fatal("unknown kind", r.Kind)
			}
			if o.Steps() != r.AfterSteps || o.Cycles() != r.AfterCycles {
				t.Fatal("after time", n+2, i)
			}
		}
		if !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) || a.Regs() != b.Regs() || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) {
			t.Fatal("每筆完整原版不同", n+2)
		}
		if d.Active() {
			d.Draw(image.NewRGBA(image.Rect(0, 0, 960, 600)), b.Indexed(), b.Palette(), image.Point{})
		}
		if !wanted[b.Steps()] || len(b.HeldKeys()) != 0 {
			continue
		}
		memory, raw := b.Bytes(oracle.Far(0, 0), 1<<20), b.Indexed()
		pal := []byte{}
		for _, c := range b.Palette() {
			pal = append(pal, c[:]...)
		}
		canvas := inventoryOriginalCanvas(b)
		_, _, rgb := b.ScreenRGB()
		sageDrawn, pose := sage.Draw(canvas, raw, rgb, image.Point{})
		drawn := d.Draw(canvas, raw, b.Palette(), image.Point{})
		state := d.State()
		stateJSON, e := json.Marshal(state)
		if e != nil {
			t.Fatal(e)
		}
		var normalized map[string]any
		if e = json.Unmarshal(stateJSON, &normalized); e != nil {
			t.Fatal(e)
		}
		for _, target := range targets.Records {
			if target.Steps != b.Steps() {
				continue
			}
			if target.Cycles != b.Cycles() || target.Memory != fmtHash(memory) || target.Indexed != fmtHash(raw) || target.Palette != fmtHash(pal) || !reflect.DeepEqual(normalized, target.Dialogue) || sageDrawn != target.Sage || pose != target.Pose {
				t.Fatal("GUI指定同狀態不同", target.Label, state, target.Dialogue, sageDrawn, pose)
			}
		}
		stem := fmt.Sprintf("source-%02d", len(captures))
		for ext, value := range map[string][]byte{"memory": memory, "indexed": raw, "palette": pal} {
			if e = os.WriteFile(filepath.Join(out, stem+"."+ext), value, 0644); e != nil {
				t.Fatal(e)
			}
		}
		captures = append(captures, map[string]any{"file": stem, "steps": b.Steps(), "cycles": b.Cycles(), "memory_sha256": fmtHash(memory), "indexed_sha256": fmtHash(raw), "palette_sha256": fmtHash(pal), "dialogue": state, "dialogue_drawn": drawn, "sage_HD": sageDrawn, "sage_pose": pose})
		wanted[b.Steps()] = false
	}
	for step, pending := range wanted {
		if pending {
			t.Fatal("未捕獲GUI目標", step)
		}
	}
	if !bytes.Equal(b.Bytes(oracle.Far(0, 0), 1<<20), read(filepath.Join(gui, "town-final.png.memory"))) || !bytes.Equal(b.Indexed(), read(filepath.Join(gui, "town-final.png.indexed"))) {
		t.Fatal("GUI末端原版不同")
	}
	if !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) || len(b.Wrote()) != 0 || len(b.HeldKeys()) != 0 {
		t.Fatal("輸入或檔案不同")
	}
	for _, p := range saves {
		stat, e := os.Stat(p)
		if e != nil || !bytes.Equal(read(p), save) || !stat.ModTime().Equal(info.ModTime()) {
			t.Fatal("存檔不符", e)
		}
	}
	result := map[string]any{"status": "PASS_EXHAUSTION_ACTUAL_GUI_JOURNAL", "records": len(lines), "full_original_state_each_record_equal": true, "final_GUI_memory_indexed_equal": true, "save_unchanged": true, "sources": captures, "original_glyph_events": glyphs, "fileops": len(b.FileOps()), "rejections": d.Rejections, "limits": "真實GUI原生日誌及指定CPU時刻完整同狀態；獨立像素另驗，未注入或重擲"}
	encoded, e := json.MarshalIndent(result, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(encoded, 10), 0644); e != nil {
		t.Fatal(e)
	}
}

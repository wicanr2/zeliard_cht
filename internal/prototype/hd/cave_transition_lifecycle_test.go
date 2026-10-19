package hd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
)

// 正常失敗視窗的實際原生日誌。每個Run後依Draw呼叫呈現選取，不還原CPU。
func TestCaveFramesPreservesFreshBodyBeforeBatch(t *testing.T) {
	study := caveShieldStudy(t)
	root, out := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_TRANSITION_LIFECYCLE_OUTPUT")
	if root == "" || out == "" {
		t.Skip("須提供實際視窗日誌及新的生命週期輸出；skip不算驗收")
	}
	gui := filepath.Join(study, "shield-transition-gui-v1/SAVEM1")
	journal := caveShieldRead(t, filepath.Join(gui, "guest-inputs.jsonl"))
	scanner := bufio.NewScanner(bytes.NewReader(journal))
	scanner.Buffer(make([]byte, 4096), 1<<20)
	if !scanner.Scan() {
		t.Fatal("缺少實際日誌標頭")
	}
	var header struct {
		Kind   string
		Cycles uint64      `json:"cycles_per_ms"`
		Rate   int         `json:"audio_sample_rate"`
		Memory string      `json:"initial_memory_sha256"`
		Regs   oracle.Regs `json:"initial_regs"`
	}
	if e := json.Unmarshal(scanner.Bytes(), &header); e != nil || header.Kind != "header" {
		t.Fatal(e)
	}
	save := caveShieldRead(t, filepath.Join(root, "SAVEM1.USR"))
	info, e := os.Stat(filepath.Join(root, "SAVEM1.USR"))
	if e != nil {
		t.Fatal(e)
	}
	scratch := t.TempDir()
	dst := filepath.Join(scratch, "SAVEM1.USR")
	if e = os.WriteFile(dst, save, 0644); e != nil {
		t.Fatal(e)
	}
	if e = os.Chtimes(dst, info.ModTime(), info.ModTime()); e != nil {
		t.Fatal(e)
	}
	o, e := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
	if e != nil {
		t.Fatal(e)
	}
	defer o.Close()
	o.SetScratch(scratch)
	o.SetDOSBoxCycles(header.Cycles)
	audio := o.NewAudio(header.Rate)
	if townBackgroundHash(o.Bytes(oracle.Far(0, 0), 1<<20)) != header.Memory || o.Regs() != header.Regs {
		t.Fatal("冷啟動身份不同")
	}
	frames := NewCaveFrames(o)
	fresh, cleared, records := 0, 0, 1
	var observations []map[string]any
	for scanner.Scan() {
		var r caveTransitionInput
		if e = json.Unmarshal(scanner.Bytes(), &r); e != nil {
			t.Fatal(e)
		}
		if o.Steps() != r.BeforeSteps || o.Cycles() != r.BeforeCycles {
			t.Fatal("原生日誌前置時間")
		}
		switch r.Kind {
		case "key":
			if r.Pressed {
				o.KeyDown(r.Scan)
			} else {
				o.KeyUp(r.Scan)
			}
		case "run":
			if r.Budget > 5000000 || o.Steps() > 150000000 {
				t.Fatal("原版預算")
			}
			err := o.RunCycles(r.Budget)
			if r.Audio {
				audio.Render()
			}
			message := ""
			if err != nil {
				message = err.Error()
			}
			if message != r.Error || o.Regs() != r.AfterRegs {
				t.Fatal("原版暫存器不同")
			}
		default:
			t.Fatal("未知原生日誌")
		}
		if o.Steps() != r.AfterSteps || o.Cycles() != r.AfterCycles {
			t.Fatal("原生日誌完成時間")
		}
		records++
		if r.Kind != "run" {
			continue
		}
		_, _, rgb := o.ScreenRGB()
		in := frames.read(o.Indexed(), rgb)
		body, last := frames.body, frames.complete
		pending := body != nil && last != nil && body.steps > last.Steps && body.cycles >= last.Cycles && body.steps <= in.steps && body.cycles <= in.cycles && caveFrameSource(in) && caveFrameSource(*body) && !frames.drawing && !caveFrameEqual(in.indexed, last.SourceIndexed)
		var obs map[string]any
		if pending {
			fresh++
			p := body.memory[0x1b10:0x11b10]
			obs = map[string]any{"steps": in.steps, "cycles": in.cycles, "body_steps": body.steps, "body_cycles": body.cycles, "last_complete_steps": last.Steps, "last_complete_cycles": last.Cycles, "drawing": frames.drawing, "body_0093": p[0x93], "body_FF3A": p[0xff3a], "body_memory_sha256": townBackgroundHash(body.memory), "current_memory_sha256": townBackgroundHash(in.memory), "old_view_differs": true, "source_guards_valid": true}
			if fresh == 1 {
				for ext, b := range map[string][]byte{"memory": in.memory, "body.memory": body.memory, "old.memory": last.Memory, "indexed": in.indexed, "old.indexed": last.SourceIndexed, "palette": in.palette} {
					if e = os.WriteFile(filepath.Join(out, "fresh-before-batch."+ext), b, 0644); e != nil {
						t.Fatal(e)
					}
				}
			}
		}
		selected := frames.selectFrame(in, false)
		if pending {
			lost := frames.body == nil
			if lost {
				cleared++
			}
			obs["fresh_body_cleared"] = lost
			obs["old_presentation_withdrawn"] = selected == nil
			observations = append(observations, obs)
		}
	}
	if e = scanner.Err(); e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(o.Bytes(oracle.Far(0, 0), 1<<20), caveShieldRead(t, filepath.Join(gui, "town-final.png.memory"))) || !bytes.Equal(o.Indexed(), caveShieldRead(t, filepath.Join(gui, "town-final.png.indexed"))) {
		t.Fatal("實際失敗視窗的最終原版狀態不同")
	}
	report := map[string]any{"records": records, "fresh_body_observations": fresh, "cleared_fresh_bodies": cleared, "refused": frames.refused, "observations": observations, "final_original_RAM_VGA_equal": true, "journal_sha256": townBackgroundHash(journal), "scope": "實際正常冷F7與Up日誌，每個Run之後呼叫selectFrame；不注入或還原CPU，未完成畫格不當作HD收據"}
	data, e := json.MarshalIndent(report, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "lifecycle.json"), append(data, 10), 0644); e != nil {
		t.Fatal(e)
	}
	if fresh == 0 || cleared != 0 || frames.refused != 0 {
		t.Fatal(fmt.Sprint(report))
	}
}

// 只組裝呈現器的真實來源參數，不還原原版機器；負向資料只測守衛。
func TestCaveFramesPendingBodyGuardBoundaries(t *testing.T) {
	base := filepath.Join(caveShieldStudy(t), "shield-transition-lifecycle-v1")
	var receipt struct {
		Observations []struct {
			Steps, Cycles uint64
			BodySteps     uint64 `json:"body_steps"`
			BodyCycles    uint64 `json:"body_cycles"`
			LastSteps     uint64 `json:"last_complete_steps"`
			LastCycles    uint64 `json:"last_complete_cycles"`
		}
	}
	if e := json.Unmarshal(caveShieldRead(t, filepath.Join(base, "lifecycle.json")), &receipt); e != nil || len(receipt.Observations) != 1 {
		t.Fatal("真實生命週期來源不符", e)
	}
	r := receipt.Observations[0]
	memory, bodyMemory, raw, oldRaw, palette := caveShieldRead(t, filepath.Join(base, "fresh-before-batch.memory")), caveShieldRead(t, filepath.Join(base, "fresh-before-batch.body.memory")), caveShieldRead(t, filepath.Join(base, "fresh-before-batch.indexed")), caveShieldRead(t, filepath.Join(base, "fresh-before-batch.old.indexed")), caveShieldRead(t, filepath.Join(base, "fresh-before-batch.palette"))
	rgb := make([]byte, 192000)
	for i, v := range raw {
		copy(rgb[i*3:], palette[int(v)*3:int(v)*3+3])
	}
	cases := []struct {
		name          string
		blocked, keep bool
		change        func(*caveFrameInput, *caveFrameInput)
	}{
		{"fresh-body", false, true, func(_, _ *caveFrameInput) {}},
		{"stale-body", false, false, func(_ *caveFrameInput, b *caveFrameInput) { b.steps = r.LastSteps }},
		{"future-body", false, false, func(i, b *caveFrameInput) { b.steps = i.steps + 1 }},
		{"old-body-cycles", false, false, func(_ *caveFrameInput, b *caveFrameInput) { b.cycles = r.LastCycles - 1 }},
		{"future-body-cycles", false, false, func(i, b *caveFrameInput) { b.cycles = i.cycles + 1 }},
		{"unknown-body-source", false, false, func(_ *caveFrameInput, b *caveFrameInput) {
			b.memory = cloneTownBytes(b.memory)
			b.memory[0x1b10+0xc000] ^= 1
		}},
		{"unknown-current-source", false, false, func(i, _ *caveFrameInput) { i.memory = cloneTownBytes(i.memory); i.memory[0x1b10+0x302c] ^= 1 }},
		{"blocked", true, false, func(_, _ *caveFrameInput) {}},
		{"time-rollback", false, false, func(i, _ *caveFrameInput) { i.steps = r.LastSteps - 1 }},
		{"cycle-rollback", false, false, func(i, _ *caveFrameInput) { i.cycles = r.LastCycles - 1 }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := caveFrameInput{mode: 0x13, memory: memory, indexed: raw, rgb: rgb, palette: palette, steps: r.Steps, cycles: r.Cycles}
			body := &caveFrameInput{mode: 0x13, memory: bodyMemory, palette: palette, steps: r.BodySteps, cycles: r.BodyCycles}
			c.change(&in, body)
			f := &CaveFrames{body: body, complete: &CavePresentation{Steps: r.LastSteps, Cycles: r.LastCycles, SourceIndexed: oldRaw}}
			if selected := f.selectFrame(in, c.blocked); selected != nil || f.complete != nil {
				t.Fatal("撤除期間仍回傳舊呈現")
			}
			if (f.body == body) != c.keep || (f.body != nil && !c.keep) {
				t.Fatal("新主體保留／清除邊界錯誤")
			}
		})
	}
}

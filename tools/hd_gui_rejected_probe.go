// 重播已存GUI的真實鍵緣與RunCycles，取具名被拒來源；不注入狀態。
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/hd"
)

func main() {
	project := flag.String("project", "", "唯讀專案")
	gui := flag.String("gui", "", "已有GUI研究目錄")
	output := flag.String("output", "", "既有空輸出")
	verifyFixed := flag.Bool("verify-fixed", false, "核對原被拒畫格現在可呈現；仍須匹配原來源SHA")
	flag.Parse()
	need := func(e error) {
		if e != nil {
			panic(e)
		}
	}
	read := func(p string) []byte { b, e := os.ReadFile(p); need(e); return b }
	sha := func(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
	write := func(name string, b []byte) { need(os.WriteFile(filepath.Join(*output, name), b, 0644)) }
	files, e := os.ReadDir(*output)
	need(e)
	if len(files) != 0 {
		panic("fresh output required")
	}
	var wanted struct {
		Steps     uint64
		MemorySHA string `json:"oracle_memory_sha256"`
		Cave      struct {
			Rejected struct {
				Steps      uint64 `json:"steps"`
				MemorySHA  string `json:"memory_sha256"`
				BodySHA    string `json:"body_memory_sha256"`
				IndexedSHA string `json:"indexed_sha256"`
				PaletteSHA string `json:"palette_sha256"`
				Reason     string `json:"reason"`
			} `json:"last_rejected_source"`
		} `json:"cave_frame_presentation"`
	}
	statePath := filepath.Join(*gui, "failed-state.json")
	need(json.Unmarshal(read(statePath), &wanted))
	journalPath := filepath.Join(*gui, "boot1/inputs.jsonl")
	journal := read(journalPath)
	type record struct {
		Kind          string
		CyclesPerMS   uint64      `json:"cycles_per_ms"`
		SampleRate    int         `json:"audio_sample_rate"`
		InitialMemory string      `json:"initial_memory_sha256"`
		InitialRegs   oracle.Regs `json:"initial_regs"`
		BeforeSteps   uint64      `json:"before_steps"`
		BeforeCycles  uint64      `json:"before_cycles"`
		AfterSteps    uint64      `json:"after_steps"`
		AfterCycles   uint64      `json:"after_cycles"`
		AfterRegs     oracle.Regs `json:"after_regs"`
		Budget        uint64      `json:"cycle_budget"`
		Audio         bool        `json:"audio_render"`
		Scan          byte
		Pressed       bool
		Error         string
	}
	scanner := bufio.NewScanner(bytes.NewReader(journal))
	if !scanner.Scan() {
		panic("missing journal header")
	}
	var header record
	need(json.Unmarshal(scanner.Bytes(), &header))
	if header.Kind != "header" {
		panic("header kind")
	}
	app := filepath.Join(*project, "workplace/music-release-v1/release-build/AppDir")
	orig := filepath.Join(app, "original")
	o, e := oracle.Load(filepath.Join(orig, "ZELIARD.EXE"), orig)
	need(e)
	defer o.Close()
	if sha(o.Bytes(oracle.Far(0, 0), 1<<20)) != header.InitialMemory || o.Regs() != header.InitialRegs {
		panic("cold GUI source differs")
	}
	savePath := filepath.Join(*project, "workplace/orig/SAVEM1.USR")
	save := read(savePath)
	stamp, e := os.Stat(savePath)
	need(e)
	if sha(save) != "03df9abcf62b1712d6b46dd1fc5b23335a4b8711b3dc84875a1498260d37735c" {
		panic("save identity")
	}
	scratch := filepath.Join(*output, "scratch")
	need(os.Mkdir(scratch, 0755))
	copySave := filepath.Join(scratch, "SAVETEST.usr")
	need(os.WriteFile(copySave, save, 0644))
	need(os.Chtimes(copySave, stamp.ModTime(), stamp.ModTime()))
	o.SetScratch(scratch)
	o.SetDOSBoxCycles(header.CyclesPerMS)
	o.TraceFiles()
	audio := o.NewAudio(header.SampleRate)
	frames := hd.NewCaveFrames(o)
	var selected *hd.CaveRejected
	for _, hook := range []uint16{0x7114, 0x714c} {
		o.OnCall(oracle.Far(0x1b1, hook), func(q *oracle.Oracle) {
			if q.Steps() != wanted.Cave.Rejected.Steps {
				return
			}
			if *verifyFixed {
				_, _, rgb := q.ScreenRGB()
				p := frames.Present(q.Indexed(), rgb, false)
				if p == nil || p.Steps != wanted.Cave.Rejected.Steps {
					panic("原被拒來源仍無完成呈現")
				}
				selected = &hd.CaveRejected{Memory: p.Memory, BodyMemory: p.BodyMemory, Indexed: p.SourceIndexed, Palette: p.Palette, Steps: p.Steps, Cycles: p.Cycles, BodySteps: p.BodySteps, BodyCycles: p.BodyCycles, Hook: q.IP().Off, Paints: p.Paints}
			} else {
				selected = frames.RejectedSource()
			}
		})
	}
	count := 1
	found := false
	for scanner.Scan() {
		count++
		var r record
		need(json.Unmarshal(scanner.Bytes(), &r))
		if o.Steps() != r.BeforeSteps || o.Cycles() != r.BeforeCycles {
			panic(fmt.Sprintf("before record %d", count))
		}
		switch r.Kind {
		case "key":
			if r.Pressed {
				o.KeyDown(r.Scan)
			} else {
				o.KeyUp(r.Scan)
			}
		case "run":
			if r.Budget > 5000000 || r.AfterSteps > 300000000 {
				panic("bounded GUI replay")
			}
			message := ""
			if e := o.RunCycles(r.Budget); e != nil {
				message = e.Error()
			}
			if r.Audio {
				audio.Render()
			}
			if message != r.Error || o.Regs() != r.AfterRegs {
				panic(fmt.Sprintf("registers record %d", count))
			}
		default:
			panic("unknown record")
		}
		if o.Steps() != r.AfterSteps || o.Cycles() != r.AfterCycles {
			panic("after record")
		}
		if o.Steps() == wanted.Steps {
			if sha(o.Bytes(oracle.Far(0, 0), 1<<20)) != wanted.MemorySHA {
				panic("GUI observation RAM differs")
			}
			found = true
			break
		}
		if o.Steps() > wanted.Steps {
			panic("target step missing from journal")
		}
	}
	need(scanner.Err())
	if !found || selected == nil {
		panic("target or rejected source missing")
	}
	if sha(selected.Memory) != wanted.Cave.Rejected.MemorySHA || sha(selected.BodyMemory) != wanted.Cave.Rejected.BodySHA || sha(selected.Indexed) != wanted.Cave.Rejected.IndexedSHA {
		panic("GUI rejected source differs")
	}
	for suffix, data := range map[string][]byte{"memory": selected.Memory, "body.memory": selected.BodyMemory, "indexed": selected.Indexed, "palette": selected.Palette} {
		write("source."+suffix, data)
	}
	paints, e := json.Marshal(selected.Paints)
	need(e)
	write("source.paints.json", paints)
	if sha(selected.Palette) != wanted.Cave.Rejected.PaletteSHA || (!*verifyFixed && selected.Reason != wanted.Cave.Rejected.Reason) {
		panic("rejected palette or reason differs")
	}
	info, e := os.Stat(copySave)
	need(e)
	if !bytes.Equal(read(copySave), save) || !info.ModTime().Equal(stamp.ModTime()) || len(o.Wrote()) != 0 {
		panic("save changed")
	}
	write("probe-source.go", read(filepath.Join(*project, "tools/hd_gui_rejected_probe.go")))
	report := map[string]any{"status": "PASS_REJECTED_GUI_JOURNAL_SOURCE_RECONSTRUCTION", "script_sha256": sha(read(filepath.Join(*output, "probe-source.go"))), "records": count, "journal_sha256": sha(journal), "observation_sha256": sha(read(statePath)), "target_steps": wanted.Steps, "source_steps": selected.Steps, "body_steps": selected.BodySteps, "memory_sha256": sha(selected.Memory), "body_memory_sha256": sha(selected.BodyMemory), "indexed_sha256": sha(selected.Indexed), "palette_sha256": sha(selected.Palette), "reason": selected.Reason, "source_cycles": selected.Cycles, "body_cycles": selected.BodyCycles, "hook": fmt.Sprintf("01B1:%04X", selected.Hook), "save_unchanged": true, "limits": "同GUI日誌之原版被拒來源重生；不啟用HD、不放寬守衛，非新GUI或HD通過。"}
	if *verifyFixed {
		report["status"] = "PASS_PREVIOUSLY_REJECTED_GUI_SOURCE_NOW_PRESENTABLE"
		report["previous_reason"] = wanted.Cave.Rejected.Reason
		report["limits"] = "同GUI正常日誌與原版完整來源SHA相同，新解碼器可選取完成來源；非新實際GUI或全部HD驗收。"
	}
	data, e := json.MarshalIndent(report, "", "  ")
	need(e)
	write("receipt.json", append(data, 10))
	fmt.Println("PASS_REJECTED_GUI_SOURCE", count, selected.Steps, selected.Reason)
}

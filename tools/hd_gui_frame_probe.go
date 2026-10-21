// 重播已存GUI的真實鍵緣與RunCycles，取具名完成來源；不注入狀態。
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/hd"
)

func main() {
	project := flag.String("project", "", "唯讀專案")
	gui := flag.String("gui", "", "已有GUI研究目錄")
	output := flag.String("output", "", "既有空輸出")
	adlib := flag.Bool("adlib", false, "重播明確採AdLib的GUI，沿前端建立獨立RESOURCE.CFG")
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
			Steps      uint64 `json:"source_steps"`
			MemorySHA  string `json:"source_memory_sha256"`
			BodySHA    string `json:"body_memory_sha256"`
			IndexedSHA string `json:"source_indexed_sha256"`
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
	if *adlib {
		config := read(filepath.Join(orig, "RESOURCE.CFG"))
		config = []byte(strings.ReplaceAll(string(config), "MSCSTD.DRV", "MSCADLIB.DRV"))
		need(os.WriteFile(filepath.Join(scratch, "RESOURCE.CFG"), config, 0600))
		o.SetAdLib(true)
	}
	o.SetDOSBoxCycles(header.CyclesPerMS)
	o.TraceFiles()
	audio := o.NewAudio(header.SampleRate)
	frames := hd.NewCaveFrames(o)
	var selected *hd.CavePresentation
	for _, hook := range []uint16{0x7114, 0x714c} {
		o.OnCall(oracle.Far(0x1b1, hook), func(q *oracle.Oracle) {
			if q.Steps() != wanted.Cave.Steps {
				return
			}
			_, _, rgb := q.ScreenRGB()
			selected = frames.Present(q.Indexed(), rgb, false)
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
		panic("target or complete source missing")
	}
	if sha(selected.Memory) != wanted.Cave.MemorySHA || sha(selected.BodyMemory) != wanted.Cave.BodySHA || sha(selected.SourceIndexed) != wanted.Cave.IndexedSHA {
		panic("GUI complete source differs")
	}
	for suffix, data := range map[string][]byte{"memory": selected.Memory, "body.memory": selected.BodyMemory, "indexed": selected.SourceIndexed, "palette": selected.Palette} {
		write("source."+suffix, data)
	}
	paints, e := json.Marshal(selected.Paints)
	need(e)
	write("source.paints.json", paints)
	art, e := hd.NewCaveTerrain(filepath.Join(app, "assets/cave"))
	need(e)
	need(art.LoadMPP1(filepath.Join(*project, "workplace/full-hd-v1/cave-mpp1-complete-v2")))
	need(art.LoadCommon(filepath.Join(*project, "workplace/full-hd-v1/cave-common-registration-v2")))
	need(art.LoadMagia(filepath.Join(*project, "workplace/full-hd-v1/cave-magia-complete-v1")))
	matched := art.Draw(image.NewRGBA(image.Rect(0, 0, 960, 600)), selected, image.Point{})
	ids := map[byte]int{}
	for _, c := range selected.Decoded.Cells {
		ids[c.ID]++
	}
	info, e := os.Stat(copySave)
	need(e)
	if !bytes.Equal(read(copySave), save) || !info.ModTime().Equal(stamp.ModTime()) || len(o.Wrote()) != 0 {
		panic("save changed")
	}
	write("probe-source.go", read(filepath.Join(*project, "tools/hd_gui_frame_probe.go")))
	report := map[string]any{"status": "PASS_GUI_JOURNAL_SOURCE_RECONSTRUCTION", "script_sha256": sha(read(filepath.Join(*output, "probe-source.go"))), "records": count, "journal_sha256": sha(journal), "observation_sha256": sha(read(statePath)), "target_steps": wanted.Steps, "source_steps": selected.Steps, "body_steps": selected.BodySteps, "memory_sha256": sha(selected.Memory), "body_memory_sha256": sha(selected.BodyMemory), "indexed_sha256": sha(selected.SourceIndexed), "native_cells": selected.Decoded.Cells, "cell_id_counts": ids, "terrain_hd_matched": matched, "common_hd": art.CommonState(), "magia_hd": art.MagiaState(), "save_unchanged": true, "limits": "同GUI日誌之原版來源重生；只載地形、平台與魔法石作拒絕診斷，非GUI或完整前端畫面驗收。"}
	report["adlib"] = *adlib
	data, e := json.MarshalIndent(report, "", "  ")
	need(e)
	write("receipt.json", append(data, 10))
	fmt.Println("PASS_GUI_SOURCE", count, selected.Steps, "terrain matched", matched, ids)
}

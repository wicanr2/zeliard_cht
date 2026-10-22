// 057：重播既有正常入洞日誌，只讀原版畫格、色盤與記憶體。
package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/dosgolem/oracle"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	project := flag.String("project", "/project", "唯讀專案")
	output := flag.String("output", "", "既有空輸出目錄")
	journalFile := flag.String("journal", "", "原生IRQ與RunCycles日誌")
	first := flag.Uint64("first-step", 190000000, "擷取起點")
	last := flag.Uint64("last-step", 214000000, "有界重播終點")
	denseFrom := flag.Uint64("dense-from", 0, "逐筆擷取起點，零值停用")
	denseTo := flag.Uint64("dense-to", 0, "逐筆擷取終點")
	flag.Parse()
	if *output == "" || *journalFile == "" || *first > *last {
		panic("須指定日誌與既有空輸出")
	}
	root := *project
	out := *output
	must := func(e error) {
		if e != nil {
			panic(e)
		}
	}
	read := func(p string) []byte { b, e := os.ReadFile(p); must(e); return b }
	sha := func(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
	existing, e := os.ReadDir(out)
	must(e)
	if len(existing) != 0 {
		panic("fresh output required")
	}
	journal := read(*journalFile)
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
		Cheats        *struct {
			LockHP               bool  `json:"lock_hp"`
			LockMP               bool  `json:"lock_mp"`
			StandingRegeneration *bool `json:"standing_regeneration"`
		} `json:"cheats"`
	}
	sc := bufio.NewScanner(bytes.NewReader(journal))
	if !sc.Scan() {
		panic("header")
	}
	var header record
	must(json.Unmarshal(sc.Bytes(), &header))
	if c := header.Cheats; c != nil && (c.LockHP || c.LockMP || c.StandingRegeneration != nil && !*c.StandingRegeneration) {
		panic("active cheats are not original oracle evidence")
	}
	orig := root + "/workplace/music-release-v1/release-build/AppDir/original"
	o, e := oracle.Load(orig+"/ZELIARD.EXE", orig)
	must(e)
	defer o.Close()
	if sha(o.Bytes(oracle.Far(0, 0), 1<<20)) != header.InitialMemory || o.Regs() != header.InitialRegs {
		panic("cold identity")
	}
	scratch := out + "/scratch"
	must(os.Mkdir(scratch, 0755))
	save := read(root + "/workplace/orig/SAVEM1.USR")
	must(os.WriteFile(scratch+"/SAVETEST.usr", save, 0644))
	stamp, e := os.Stat(root + "/workplace/orig/SAVEM1.USR")
	must(e)
	must(os.Chtimes(scratch+"/SAVETEST.usr", stamp.ModTime(), stamp.ModTime()))
	o.SetScratch(scratch)
	cfg := []byte(strings.ReplaceAll(string(read(orig+"/RESOURCE.CFG")), "MSCSTD.DRV", "MSCADLIB.DRV"))
	must(os.WriteFile(scratch+"/RESOURCE.CFG", cfg, 0600))
	o.SetAdLib(true)
	o.SetDOSBoxCycles(header.CyclesPerMS)
	audio := o.NewAudio(header.SampleRate)
	n := 0
	var shots []map[string]any
	lastPal := ""
	lastRecord := 0
	for sc.Scan() {
		n++
		var r record
		must(json.Unmarshal(sc.Bytes(), &r))
		if o.Steps() != r.BeforeSteps || o.Cycles() != r.BeforeCycles {
			panic(fmt.Sprintf("before %d", n))
		}
		switch r.Kind {
		case "cheat_settings":
			panic("cheat journal is not original oracle evidence")
		case "cpu_speed":
			audio.Render()
			o.SetDOSBoxCycles(r.CyclesPerMS)
		case "audio":
			audio.Render()
		case "key":
			if r.Pressed {
				o.KeyDown(r.Scan)
			} else {
				o.KeyUp(r.Scan)
			}
		case "run":
			if r.Budget > 5000000 {
				panic("bounded cycle budget")
			}
			msg := ""
			if e = o.RunCycles(r.Budget); e != nil {
				msg = e.Error()
			}
			if r.Audio {
				audio.Render()
			}
			if msg != r.Error || o.Regs() != r.AfterRegs {
				panic(fmt.Sprintf("regs %d", n))
			}
		default:
			panic(r.Kind)
		}
		if o.Steps() != r.AfterSteps || o.Cycles() != r.AfterCycles {
			panic("after")
		}
		if o.Steps() >= *first && o.Steps() <= *last && r.Kind == "run" {
			var pal []byte
			for _, c := range o.Palette() {
				pal = append(pal, c[:]...)
			}
			ph := sha(pal)
			if n-lastRecord >= 8 || ph != lastPal || (*denseFrom > 0 && o.Steps() >= *denseFrom && o.Steps() <= *denseTo) {
				raw := o.Indexed()
				w, h, rgb := o.ScreenRGB()
				if w != 320 || h != 200 {
					panic("mode")
				}
				mem := o.Bytes(oracle.Far(0, 0), 1<<20)
				p := mem[0x1b10:0x11b10]
				base := fmt.Sprintf("frame-%04d", len(shots))
				name := filepath.Join(out, base)
				must(os.WriteFile(name+".indexed", raw, 0644))
				must(os.WriteFile(name+".palette", pal, 0644))
				f, e := os.Create(name + ".memory.gz")
				must(e)
				z := gzip.NewWriter(f)
				_, e = z.Write(mem)
				must(e)
				must(z.Close())
				must(f.Close())
				im := image.NewRGBA(image.Rect(0, 0, 320, 200))
				for i := range raw {
					copy(im.Pix[i*4:i*4+3], rgb[i*3:i*3+3])
					im.Pix[i*4+3] = 255
				}
				f, e = os.Create(name + ".png")
				must(e)
				must(png.Encode(f, im))
				must(f.Close())
				s := map[string]any{"file": base, "journal_record": n, "steps": o.Steps(), "cycles": o.Cycles(), "regs": o.Regs(), "memory_sha256": sha(mem), "indexed_sha256": sha(raw), "palette_sha256": ph, "ff2c": int(p[0xff2c]) + 256*int(p[0xff2d]), "ff31": int(p[0xff31]) + 256*int(p[0xff32]), "00d0": p[0xd0], "0093": p[0x93]}
				shots = append(shots, s)
				lastRecord = n
				lastPal = ph
				fmt.Println(base, o.Steps(), ph[:8])
			}
		}
		if o.Steps() > *last {
			break
		}
	}
	must(sc.Err())
	if !bytes.Equal(save, read(scratch+"/SAVETEST.usr")) {
		panic("save")
	}
	report := map[string]any{"status": "PASS_ORIGINAL_JOURNAL_REPLAY", "journal_sha256": sha(journal), "records": n, "shots": shots, "save_unchanged": true, "scope": "原生暫存器與時鐘日誌重生；完整RAM逐點SHA僅保存，不宣稱與GUI全RAM相同", "go_version": runtime.Version(), "tool_sha256": sha(read(root + "/tools/hd_transition_probe.go"))}
	data, e := json.MarshalIndent(report, "", "  ")
	must(e)
	must(os.WriteFile(out+"/receipt.json", append(data, 10), 0644))
	fmt.Println("PASS", len(shots))
}

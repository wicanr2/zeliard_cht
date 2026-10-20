// 正常鍵盤探索後續城鎮入口，只讀原版位置與完成畫格，不注入狀態。
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/wicanr2/dosgolem/oracle"
)

func main() {
	project := flag.String("project", "", "唯讀專案")
	out := flag.String("output", "", "既有空輸出目錄")
	prefix := flag.String("prefix", "", "正常初始鍵序，之後由stdin繼續")
	saveName := flag.String("save", "SAVETEST", "固定來源存檔：SAVETEST或包內SAVEM1／SAVEM2／SAVEM3／SAVEM")
	flag.Parse()
	need := func(e error) {
		if e != nil {
			panic(e)
		}
	}
	read := func(path string) []byte { b, e := os.ReadFile(path); need(e); return b }
	sha := func(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
	entries, e := os.ReadDir(*out)
	need(e)
	if len(entries) != 0 {
		panic("fresh empty output required")
	}
	write := func(name string, data []byte) { need(os.WriteFile(filepath.Join(*out, name), data, 0644)) }
	original := filepath.Join(*project, "workplace/orig")
	if sha(read(filepath.Join(original, "ZELIARD.EXE"))) != "fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341" {
		panic("original identity")
	}
	type saveSource struct{ path, hash string }
	sources := map[string]saveSource{
		"SAVETEST": {"workplace/prose-v1/npc-normal-v2/scratch/SAVETEST.usr", "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb"},
		"SAVEM1":   {"workplace/orig/SAVEM1.USR", "03df9abcf62b1712d6b46dd1fc5b23335a4b8711b3dc84875a1498260d37735c"},
		"SAVEM2":   {"workplace/orig/SAVEM2.USR", "9b0be23b08292118dc8de8d28a917480a906133eb4a52791f294011cafab089c"},
		"SAVEM3":   {"workplace/orig/SAVEM3.USR", "e6e37ae9e79c88bf506a731cc4db5c4c4fe16db3febc475760e8c0370a1e421c"},
		"SAVEM":    {"workplace/orig/SAVEM.USR", "cd081139852ad3ffcf0acecc7a60db5bc1fdd0a3c9d2c049d78f3e368e6b3ddc"},
	}
	source, ok := sources[*saveName]
	if !ok {
		panic("unknown source save")
	}
	savePath := filepath.Join(*project, source.path)
	save := read(savePath)
	if sha(save) != source.hash {
		panic("save identity")
	}
	stamp, e := os.Stat(savePath)
	need(e)
	scratch := filepath.Join(*out, "scratch")
	need(os.Mkdir(scratch, 0755))
	copySave := filepath.Join(scratch, "SAVETEST.usr")
	need(os.WriteFile(copySave, save, 0644))
	need(os.Chtimes(copySave, stamp.ModTime(), stamp.ModTime()))
	o, e := oracle.Load(filepath.Join(original, "ZELIARD.EXE"), original)
	need(e)
	defer o.Close()
	o.SetScratch(scratch)
	o.SetDOSBoxCycles(3000)
	o.TraceFiles()
	write("probe-source.go", read(filepath.Join(*project, "tools/hd_town_path_probe.go")))
	var points []map[string]any
	var commands []string
	var journal []map[string]any
	validName := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	var caveMemory, caveRaw, cavePalette, caveHeld []byte
	var caveSteps, caveCycles uint64
	var caveRegs any
	var caveMode byte
	o.OnCall(oracle.Far(0x1b1, 0x7114), func(q *oracle.Oracle) {
		if q.DSReg() != 0x1b1 || sha(q.Bytes(oracle.Far(0x1b1, 0x302c), 0x1f60)) != "018193b8bc08668248cd7b831c9b5be1b26743fadfbf534ea09b7066b7dd5310" {
			return
		}
		caveMemory, caveRaw = q.Bytes(oracle.Far(0, 0), 1<<20), q.Indexed()
		cavePalette = nil
		for _, c := range q.Palette() {
			cavePalette = append(cavePalette, c[:]...)
		}
		caveSteps, caveCycles, caveRegs = q.Steps(), q.Cycles(), q.Regs()
		caveHeld, caveMode = q.HeldKeys(), q.VideoMode()
	})
	snapshot := func(label string, completedCave bool) {
		if !validName.MatchString(label) || len(points) >= 96 {
			panic("bounded snapshot labels/count")
		}
		stem := fmt.Sprintf("point-%03d-%s", len(points), label)
		memory, raw := o.Bytes(oracle.Far(0, 0), 1<<20), o.Indexed()
		var palette []byte
		for _, c := range o.Palette() {
			palette = append(palette, c[:]...)
		}
		steps, cycles, regs, sourceIP := o.Steps(), o.Cycles(), any(o.Regs()), "manual observation; may be mid-paint"
		held, mode := o.HeldKeys(), o.VideoMode()
		if completedCave {
			if caveMemory == nil || caveSteps > o.Steps() {
				panic("no completed normal cave source")
			}
			memory, raw, palette = caveMemory, caveRaw, cavePalette
			steps, cycles, regs, sourceIP = caveSteps, caveCycles, caveRegs, "01B1:7114"
			held, mode = caveHeld, caveMode
		}
		write(stem+".memory", memory)
		write(stem+".indexed", raw)
		write(stem+".palette", palette)
		if len(raw) == 64000 {
			im := image.NewRGBA(image.Rect(0, 0, 320, 200))
			for i, v := range raw {
				at := int(v) * 3
				im.SetRGBA(i%320, i/320, color.RGBA{palette[at], palette[at+1], palette[at+2], 255})
			}
			f, e := os.Create(filepath.Join(*out, stem+".png"))
			need(e)
			need(png.Encode(f, im))
			need(f.Close())
		}
		p := memory[0x1b10:0x11b10]
		word := func(at int) int { return int(p[at]) | int(p[at+1])<<8 }
		point := map[string]any{"file": stem, "label": label, "steps": steps, "cycles": cycles, "regs": regs, "mode": mode,
			"source_ip": sourceIP, "observed_steps": o.Steps(), "completed_cave_source": completedCave,
			"observed_held_keys": o.HeldKeys(),
			"memory_sha256":      sha(memory), "indexed_sha256": sha(raw), "palette_sha256": sha(palette),
			"native0080": word(0x80), "native0082": p[0x82], "native0083": p[0x83], "native0084": p[0x84], "gold": word(0x86),
			"life0090": word(0x90), "native0092_00BF": fmt.Sprintf("%x", p[0x92:0xc0]),
			"nativeC4": p[0xc4], "nativeC006": p[0xc006], "nativeC002": word(0xc002), "native4FED": word(0x4fed), "held_keys": held}
		points = append(points, point)
		b, e := json.Marshal(point)
		need(e)
		fmt.Println("SHOT", string(b))
		write("points.json", func() []byte { b, e := json.MarshalIndent(points, "", "  "); need(e); return append(b, 10) }())
	}
	// 此回呼不控制原版；只記錄首次出現的新城鎮完成入口。
	seenTowns := map[uint8]bool{}
	o.OnCall(oracle.Far(0x1b1, 0x68ba), func(q *oracle.Oracle) {
		if q.DSReg() != 0x1b1 || sha(q.Bytes(oracle.Far(0x1b1, 0x3000), 0x71c)) != "cc52036cd4c7c43747d3235e3026b0e333c7eb692ed651dff569886b2233588d" {
			return
		}
		id := q.Byte(oracle.Far(0x1b1, 0xc006))
		if id >= 2 && id <= 9 && !seenTowns[id] {
			seenTowns[id] = true
			snapshot(fmt.Sprintf("town-%d-complete", id), false)
		}
	})
	run := func(n uint64) {
		if n > 100000000 || o.Steps()+n > 1200000000 {
			panic("bounded normal execution")
		}
		need(o.Run(n))
	}
	number := func(s string) uint64 { n, e := strconv.ParseUint(s, 0, 64); need(e); return n }
	consume := func(line string) bool {
		fields := strings.Fields(line)
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			return true
		}
		commands = append(commands, line)
		switch fields[0] {
		case "run":
			run(number(fields[1]))
		case "key":
			scan := uint8(number(fields[1]))
			o.KeyDown(scan)
			run(number(fields[2]))
			o.KeyUp(scan)
		case "down":
			o.KeyDown(uint8(number(fields[1])))
		case "up":
			o.KeyUp(uint8(number(fields[1])))
		case "shot":
			snapshot(fields[1], false)
		case "cave":
			snapshot(fields[1], true)
		case "quit":
			return false
		default:
			panic(line)
		}
		journal = append(journal, map[string]any{"command": line, "steps": o.Steps(), "cycles": o.Cycles(), "regs": o.Regs()})
		return true
	}
	if *prefix != "" {
		for _, line := range strings.Split(string(read(*prefix)), "\n") {
			if !consume(line) {
				panic("prefix must not quit")
			}
		}
	}
	fmt.Println("READY stdin normal keys; shot LABEL; quit")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if !consume(scanner.Text()) {
			break
		}
	}
	need(scanner.Err())
	currentStamp, e := os.Stat(copySave)
	need(e)
	receipt := map[string]any{"schema": 1, "status": "OBSERVED_NORMAL_TOWN_PATH", "go": runtime.Version(), "tool": "dosgolem oracle",
		"source_save_name": *saveName, "source_save_path": source.path, "scratch_save_name": "SAVETEST.usr", "source_save_history": "包內USR取得歷史未知；自建SAVETEST沿既有正常存檔收據。副本只改檔名，保留bytes與mtime，F7由原版讀取。",
		"script_sha256": sha(read(filepath.Join(*out, "probe-source.go"))), "source_save_sha256": sha(save), "source_save_unchanged": bytes.Equal(save, read(savePath)),
		"scratch_save_unchanged": bytes.Equal(save, read(copySave)), "save_mtime_unchanged": currentStamp.ModTime().Equal(stamp.ModTime()),
		"state_injection": false, "snapshot_restore": false, "commands": commands, "journal": journal, "points": points, "file_ops": o.FileOps(),
		"limits": "正常鍵盤研究；單機來源不是成對原版狀態或實際GUI驗收。新城鎮完成點只作唯讀來源。"}
	b, e := json.MarshalIndent(receipt, "", "  ")
	need(e)
	write("receipt.json", append(b, 10))
	fmt.Println("DONE", len(points))
}

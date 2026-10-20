// 冷讀包內原版存檔，只送正常鍵盤並盤點地形來源；不注入位置或其他狀態。
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/wicanr2/dosgolem/oracle"
)

func caveProbeNeed(e error) {
	if e != nil {
		panic(e)
	}
}
func caveProbeHash(b []byte) string     { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func caveProbeRead(p string) []byte     { b, e := os.ReadFile(p); caveProbeNeed(e); return b }
func caveProbeWrite(p string, b []byte) { caveProbeNeed(os.WriteFile(p, b, 0644)) }

func main() {
	root := flag.String("original", "", "唯讀原版目錄")
	output := flag.String("output", "", "既有空輸出目錄")
	saves := flag.String("saves", "SAVEM1,SAVEM2,SAVEM3,SAVEM", "包內存檔名稱，以逗號分隔")
	flag.Parse()
	entries, e := os.ReadDir(*output)
	caveProbeNeed(e)
	if len(entries) != 0 {
		panic("fresh output required")
	}
	if caveProbeHash(caveProbeRead(filepath.Join(*root, "ZELIARD.EXE"))) != "fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341" {
		panic("original identity")
	}
	var boots []map[string]any
	for _, saveName := range strings.Split(*saves, ",") {
		if saveName != "SAVEM1" && saveName != "SAVEM2" && saveName != "SAVEM3" && saveName != "SAVEM" {
			panic("unknown source save")
		}
		savePath := filepath.Join(*root, saveName+".USR")
		save := caveProbeRead(savePath)
		info, e := os.Stat(savePath)
		caveProbeNeed(e)
		dir := filepath.Join(*output, saveName)
		caveProbeNeed(os.Mkdir(dir, 0755))
		var machines [2]*oracle.Oracle
		for i := range machines {
			o, e := oracle.Load(filepath.Join(*root, "ZELIARD.EXE"), *root)
			caveProbeNeed(e)
			machines[i] = o
			o.SetDOSBoxCycles(3000)
			scratch := filepath.Join(dir, fmt.Sprintf("scratch-%d", i))
			caveProbeNeed(os.Mkdir(scratch, 0755))
			copyPath := filepath.Join(scratch, saveName+".USR")
			caveProbeWrite(copyPath, save)
			caveProbeNeed(os.Chtimes(copyPath, info.ModTime(), info.ModTime()))
			o.SetScratch(scratch)
			o.TraceFiles()
		}
		a, b := machines[0], machines[1]
		var sourceMemory, sourceRaw, sourcePalette []byte
		var sourceSteps, sourceCycles, captures uint64
		var mapLoads []map[string]any
		b.OnCall(oracle.Far(0x1b1, 0x6cec), func(o *oracle.Oracle) {
			if o.DSReg() != 0x1b1 || caveProbeHash(o.Bytes(oracle.Far(0x1b1, 0x302c), 0x1f60)) != "018193b8bc08668248cd7b831c9b5be1b26743fadfbf534ea09b7066b7dd5310" {
				return
			}
			if len(mapLoads) >= 32 {
				panic("bounded map-load captures")
			}
			memory := o.Bytes(oracle.Far(0, 0), 1<<20)
			file := fmt.Sprintf("map-load-%02d.memory", len(mapLoads))
			caveProbeWrite(filepath.Join(dir, file), memory)
			mapLoads = append(mapLoads, map[string]any{"memory": file, "memory_sha256": caveProbeHash(memory), "steps": o.Steps(), "cycles": o.Cycles(), "source_ip": "01B1:6CEC"})
		})
		for _, off := range []uint16{0x7114, 0x714c} {
			b.OnCall(oracle.Far(0x1b1, off), func(o *oracle.Oracle) {
				if o.DSReg() != 0x1b1 || caveProbeHash(o.Bytes(oracle.Far(0x1b1, 0x302c), 0x1f60)) != "018193b8bc08668248cd7b831c9b5be1b26743fadfbf534ea09b7066b7dd5310" {
					return
				}
				sourceMemory = o.Bytes(oracle.Far(0, 0), 1<<20)
				sourceRaw = o.Indexed()
				sourcePalette = nil
				for _, c := range o.Palette() {
					sourcePalette = append(sourcePalette, c[:]...)
				}
				sourceSteps, sourceCycles = o.Steps(), o.Cycles()
				captures++
			})
		}
		var actions, points []map[string]any
		equal := func() {
			if !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Regs() != b.Regs() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) {
				panic("observer state difference")
			}
		}
		run := func(n uint64) {
			if n > 20000000 || b.Steps()+n > 180000000 {
				panic("bounded instruction budget")
			}
			before := b.Steps()
			for _, o := range machines {
				caveProbeNeed(o.Run(n))
			}
			actions = append(actions, map[string]any{"run": n, "before_steps": before, "after_steps": b.Steps()})
		}
		key := func(scan uint8, n uint64) {
			for _, o := range machines {
				o.KeyDown(scan)
			}
			actions = append(actions, map[string]any{"scan": scan, "pressed": true, "steps": b.Steps()})
			run(n)
			for _, o := range machines {
				o.KeyUp(scan)
			}
			actions = append(actions, map[string]any{"scan": scan, "pressed": false, "steps": b.Steps()})
		}
		shot := func(name string) {
			equal()
			m, raw := b.Bytes(oracle.Far(0, 0), 1<<20), b.Indexed()
			var pal []byte
			for _, c := range b.Palette() {
				pal = append(pal, c[:]...)
			}
			for ext, data := range map[string][]byte{"memory": m, "indexed": raw, "palette": pal} {
				caveProbeWrite(filepath.Join(dir, name+"."+ext), data)
			}
			picture := image.NewRGBA(image.Rect(0, 0, 320, 200))
			for i, v := range raw {
				copy(picture.Pix[i*4:i*4+3], pal[int(v)*3:int(v)*3+3])
				picture.Pix[i*4+3] = 255
			}
			f, e := os.Create(filepath.Join(dir, name+".png"))
			caveProbeNeed(e)
			caveProbeNeed(png.Encode(f, picture))
			caveProbeNeed(f.Close())
			point := map[string]any{"name": name, "state_equal": true, "steps": b.Steps(), "cycles": b.Cycles(), "memory_sha256": caveProbeHash(m), "indexed_sha256": caveProbeHash(raw), "palette_sha256": caveProbeHash(pal), "module_prefix_sha256": caveProbeHash(m[0x7b10:0x7b50]), "map_prefix_sha256": caveProbeHash(m[0xdb10:0xdb50]), "raw_c010": m[0xdb20], "raw_c012": m[0xdb22], "raw_9ef7": m[0xba07], "captures": captures}
			if sourceMemory != nil {
				for ext, data := range map[string][]byte{"memory": sourceMemory, "indexed": sourceRaw, "palette": sourcePalette} {
					caveProbeWrite(filepath.Join(dir, name+"-source."+ext), data)
				}
				point["source_steps"], point["source_cycles"] = sourceSteps, sourceCycles
				point["source_memory_sha256"] = caveProbeHash(sourceMemory)
			}
			points = append(points, point)
			fmt.Println(saveName, name, point["raw_c012"], point["raw_9ef7"], captures)
		}
		run(12000000)
		for i := 0; i < 6; i++ {
			key(0x39, 50000)
			run(7950000)
		}
		key(0x41, 500000)
		run(2000000)
		key(0x15, 150000)
		run(1500000)
		for i := 0; i < 8; i++ {
			key(0xe, 100000)
			run(200000)
		}
		scans := []uint8{31, 30, 47, 18, 50}
		if len(saveName) == 6 {
			scans = append(scans, uint8(2+saveName[5]-'1'))
		}
		for _, scan := range scans {
			key(scan, 100000)
			run(200000)
		}
		key(0x1c, 100000)
		run(16000000)
		shot("restored")
		key(0x4d, 8000000)
		run(250000)
		key(0x48, 250000)
		run(20000000)
		shot("east-entry")
		key(0x4b, 1500000)
		run(500000)
		shot("walk-left")
		for i, o := range machines {
			o.Close()
			if !bytes.Equal(caveProbeRead(filepath.Join(dir, fmt.Sprintf("scratch-%d", i), saveName+".USR")), save) {
				panic("source save modified")
			}
		}
		boots = append(boots, map[string]any{"save": saveName, "source_save_sha256": caveProbeHash(save), "source_save_mtime": info.ModTime(), "actions": actions, "points": points, "map_loads": mapLoads, "file_ops": b.FileOps(), "snapshot_restore": false, "state_injection": false})
	}
	_, sourcePath, _, ok := runtime.Caller(0)
	if !ok {
		panic("source identity unavailable")
	}
	report := map[string]any{"schema": 1, "status": "PASS_NORMAL_SAVE_SOURCE_SURVEY", "boots": boots, "go_version": runtime.Version(), "script_sha256": caveProbeHash(caveProbeRead(sourcePath)), "exe_sha256": caveProbeHash(caveProbeRead(filepath.Join(*root, "ZELIARD.EXE"))), "scope": "指定包內原版USR的冷F7及正常按鍵；存檔取得歷史未知，不宣稱新局全流程或HD完成"}
	data, e := json.MarshalIndent(report, "", "  ")
	caveProbeNeed(e)
	caveProbeWrite(filepath.Join(*output, "receipt.json"), append(data, 10))
}

// 正常冷新局城堡鍵序；只讀完成畫格與旗幟更新端，另跑獨立原版核對狀態。
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/wicanr2/dosgolem/oracle"
)

func castleNeed(e error) {
	if e != nil {
		panic(e)
	}
}
func castleHash(b []byte) string     { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func castleRead(p string) []byte     { b, e := os.ReadFile(p); castleNeed(e); return b }
func castleWrite(p string, b []byte) { castleNeed(os.WriteFile(p, b, 0644)) }

type castleFrame struct {
	Memory, Indexed, Palette []byte
	Steps, Cycles            uint64
}

func main() {
	root := flag.String("original", "", "唯讀原版")
	commands := flag.String("commands", "", "正常鍵序")
	output := flag.String("output", "", "既有空輸出目錄")
	flag.Parse()
	indexPath := filepath.Join(filepath.Dir(*commands), "..", "docs", "castle-source-index.json")
	indexData := castleRead(indexPath)
	var index struct {
		Entries []map[string]any `json:"entries"`
	}
	castleNeed(json.Unmarshal(indexData, &index))
	annotations := map[string]map[string]any{}
	for _, entry := range index.Entries {
		annotations[entry["address"].(string)] = entry
	}
	entries, e := os.ReadDir(*output)
	castleNeed(e)
	if len(entries) != 0 {
		panic("fresh output required")
	}
	exe := filepath.Join(*root, "ZELIARD.EXE")
	if castleHash(castleRead(exe)) != "fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341" {
		panic("original identity")
	}
	var machines [2]*oracle.Oracle
	for i := range machines {
		o, e := oracle.Load(exe, *root)
		castleNeed(e)
		machines[i] = o
		defer o.Close()
		o.SetDOSBoxCycles(3000)
		scratch := filepath.Join(*output, fmt.Sprintf("scratch-%d", i))
		castleNeed(os.Mkdir(scratch, 0755))
		o.SetScratch(scratch)
	}
	a, b := machines[0], machines[1]
	var latest *castleFrame
	var captures uint64
	var events, points []map[string]any
	var pending map[string]any
	var eventCount uint64
	isCastle := func(o *oracle.Oracle) bool {
		return o.VideoMode() == 0x13 && o.Word(oracle.Far(0x1b1, 0xc002)) == 114 && castleHash(o.Bytes(oracle.Far(0x1b1, 0x3000), 0x71c)) == "cc52036cd4c7c43747d3235e3026b0e333c7eb692ed651dff569886b2233588d" && castleHash(o.Bytes(oracle.Far(0x11b1, 0x8100), 157*48)) == "1decae0802e6c839ad05f0522a18612d800fc5a2d8df5972064332af198c19e8"
	}
	b.OnCall(oracle.Far(0x1b1, 0x68ba), func(o *oracle.Oracle) {
		if !isCastle(o) {
			latest = nil
			return
		}
		pal := []byte{}
		for _, c := range o.Palette() {
			pal = append(pal, c[:]...)
		}
		latest = &castleFrame{o.Bytes(oracle.Far(0, 0), 1<<20), o.Indexed(), pal, o.Steps(), o.Cycles()}
		captures++
	})
	b.OnCall(oracle.Far(0x1b1, 0x32ea), func(o *oracle.Oracle) {
		if !isCastle(o) {
			return
		}
		r := o.Regs()
		target := uint16(r.SI - 1)
		pending = map[string]any{"steps": o.Steps(), "cycles": o.Cycles(), "source_ip": "01B1:32EA", "ds": r.DS, "target_offset": target, "es": r.ES, "table_offset": r.DI, "before": o.Byte(oracle.Far(r.DS, target)), "selected": o.Byte(oracle.Far(r.ES, r.DI+1))}
		pending["semantic_index"] = annotations["0x32EA"]
		pending["write_semantic_index"] = annotations["0x32EE"]
	})
	b.OnCall(oracle.Far(0x1b1, 0x32f9), func(o *oracle.Oracle) {
		if pending == nil {
			return
		}
		r := o.Regs()
		pending["after"] = o.Byte(oracle.Far(r.DS, pending["target_offset"].(uint16)))
		eventCount++
		if pending["after"] != pending["selected"] {
			panic("observed source update mismatch")
		}
		if len(events) < 128 {
			events = append(events, pending)
		}
		pending = nil
	})
	equal := func() {
		if !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Regs() != b.Regs() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) {
			panic("observer changed original state")
		}
	}
	run := func(n uint64) {
		for _, o := range machines {
			castleNeed(o.Run(n))
		}
	}
	num := func(s string) uint64 { v, e := strconv.ParseUint(s, 0, 64); castleNeed(e); return v }
	input := castleRead(*commands)
	for _, line := range strings.Split(string(input), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		switch f[0] {
		case "run":
			run(num(f[1]))
		case "key":
			for _, o := range machines {
				o.KeyDown(uint8(num(f[1])))
			}
			run(num(f[2]))
			for _, o := range machines {
				o.KeyUp(uint8(num(f[1])))
			}
		case "shot":
			equal()
			name := f[1]
			memory := b.Bytes(oracle.Far(0, 0), 1<<20)
			point := map[string]any{"name": name, "steps": b.Steps(), "cycles": b.Cycles(), "memory_sha256": castleHash(memory), "state_equal": true, "castle": isCastle(b), "camera": b.Word(oracle.Far(0x1b1, 0x80)), "player_x": b.Byte(oracle.Far(0x1b1, 0x83)), "map_width": b.Word(oracle.Far(0x1b1, 0xc002)), "captured": captures}
			castleWrite(filepath.Join(*output, name+".memory"), memory)
			castleWrite(filepath.Join(*output, name+".indexed"), b.Indexed())
			if latest != nil && isCastle(b) {
				castleWrite(filepath.Join(*output, name+"-source.memory"), latest.Memory)
				castleWrite(filepath.Join(*output, name+"-source.indexed"), latest.Indexed)
				castleWrite(filepath.Join(*output, name+"-source.palette"), latest.Palette)
				point["source_steps"] = latest.Steps
				point["source_cycles"] = latest.Cycles
				point["source_memory_sha256"] = castleHash(latest.Memory)
				p := latest.Memory[0x1b10:0x11b10]
				camera := int(binary.LittleEndian.Uint16(p[0x80:]))
				var differences []map[string]int
				for col := 0; col < 28; col++ {
					for row := 0; row < 3; row++ {
						old := int(p[0xe000+col*8+row])
						current := int(p[0xc017+(camera+4+col)*8+row])
						if old != current {
							differences = append(differences, map[string]int{"column": camera + 4 + col, "row": row, "map": current, "drawn": old})
						}
					}
				}
				point["cache_differences"] = differences
			}
			points = append(points, point)
			fmt.Println(name, point["castle"], point["camera"], point["player_x"], point["map_width"])
		default:
			panic("unknown command " + line)
		}
	}
	equal()
	if len(b.HeldKeys()) != 0 {
		panic("held keys")
	}
	report := map[string]any{"schema": 1, "status": "PASS_NORMAL_CASTLE_OBSERVATION", "source_index_sha256": castleHash(indexData), "commands_sha256": castleHash(input), "checkpoints": points, "update_events": events, "total_update_events": eventCount, "completed_frames": captures, "observer_state_equal": true, "wrote": b.Wrote(), "scope": "正常冷新局與鍵盤；兩原版完整狀態核對。source檔屬最近完成時刻，非當次CPU狀態。非HD完成證據"}
	data, e := json.MarshalIndent(report, "", "  ")
	castleNeed(e)
	castleWrite(filepath.Join(*output, "receipt.json"), append(data, 10))
}

// 只用正常鍵序匯出洞窟被拒絕來源；不注入狀態或繞過呈現守衛。
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/hd"
)

func main() {
	project := flag.String("project", "", "唯讀專案")
	output := flag.String("output", "", "既有空輸出")
	commands := flag.String("commands", "", "正常鍵序")
	studyFrames := flag.Int("study-frames", 0, "另收最多64個不同護身記錄的拒絕來源；只作研究")
	flag.Parse()
	need := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	read := func(path string) []byte { b, e := os.ReadFile(path); need(e); return b }
	sha := func(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
	files, e := os.ReadDir(*output)
	need(e)
	if len(files) != 0 {
		panic("fresh output required")
	}
	write := func(name string, b []byte) { need(os.WriteFile(filepath.Join(*output, name), b, 0644)) }
	orig := filepath.Join(*project, "workplace/orig")
	if sha(read(filepath.Join(orig, "ZELIARD.EXE"))) != "fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341" {
		panic("EXE identity")
	}
	savePath := filepath.Join(orig, "SAVEM1.USR")
	save := read(savePath)
	if sha(save) != "03df9abcf62b1712d6b46dd1fc5b23335a4b8711b3dc84875a1498260d37735c" {
		panic("save identity")
	}
	stamp, e := os.Stat(savePath)
	need(e)
	var machines [2]*oracle.Oracle
	for i := range machines {
		o, e := oracle.Load(filepath.Join(orig, "ZELIARD.EXE"), orig)
		need(e)
		defer o.Close()
		scratch := filepath.Join(*output, fmt.Sprintf("scratch-%d", i))
		need(os.Mkdir(scratch, 0755))
		path := filepath.Join(scratch, "SAVETEST.usr")
		need(os.WriteFile(path, save, 0644))
		need(os.Chtimes(path, stamp.ModTime(), stamp.ModTime()))
		o.SetScratch(scratch)
		o.SetDOSBoxCycles(3000)
		o.TraceFiles()
		machines[i] = o
	}
	a, b := machines[0], machines[1]
	frames := hd.NewCaveFrames(b)
	var alphaPaints []map[string]any
	b.OnCall(oracle.Far(0x1b1, 0x412f), func(q *oracle.Oracle) {
		regs := q.Regs()
		if byte(regs.AX) != 102 {
			return
		}
		ret := q.Bytes(oracle.Far(regs.SS, regs.SP), 2)
		alphaPaints = append(alphaPaints, map[string]any{"steps": q.Steps(), "cycles": q.Cycles(), "id": byte(regs.AX), "di": regs.DI, "ds": regs.DS, "caller_return": int(ret[0]) | int(ret[1])<<8})
		if len(alphaPaints) > 64 {
			alphaPaints = alphaPaints[len(alphaPaints)-64:]
		}
	})
	equal := func() {
		if !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) || a.Regs() != b.Regs() || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) {
			panic("adapter changed original state")
		}
	}
	run := func(n uint64) {
		if n > 40000000 || b.Steps()+n > 240000000 {
			panic("instruction budget")
		}
		for n > 0 {
			chunk := n
			if b.Steps() >= 97000000 && chunk > 20000 {
				chunk = 20000
			}
			for _, o := range machines {
				need(o.Run(chunk))
			}
			n -= chunk
			_, _, rgb := b.ScreenRGB()
			frames.Present(b.Indexed(), rgb, false)
		}
	}
	num := func(s string) uint64 { n, e := strconv.ParseUint(s, 0, 64); need(e); return n }
	var points []map[string]any
	dump := func(label string, r *hd.CaveRejected, checked bool) {
		name := fmt.Sprintf("point-%03d", len(points))
		for suffix, data := range map[string][]byte{"memory": r.Memory, "body.memory": r.BodyMemory, "indexed": r.Indexed, "palette": r.Palette} {
			write(name+"."+suffix, data)
		}
		point := map[string]any{"name": label, "file": name, "steps": r.Steps, "cycles": r.Cycles, "body_steps": r.BodySteps, "body_cycles": r.BodyCycles, "reason": r.Reason, "hook": r.Hook, "paints": r.Paints, "memory_sha256": sha(r.Memory), "body_memory_sha256": sha(r.BodyMemory), "indexed_sha256": sha(r.Indexed), "palette_sha256": sha(r.Palette), "observed_steps": b.Steps(), "state_equal": checked, "presentation": frames.State(), "alpha_paints": append([]map[string]any{}, alphaPaints...)}
		points = append(points, point)
		fmt.Println(label, r.Steps, r.Reason)
	}
	if *studyFrames < 0 || *studyFrames > 64 {
		panic("study frame bound")
	}
	seenRecords := map[string]bool{}
	if *studyFrames > 0 {
		for _, hook := range []uint16{0x7114, 0x714c} {
			b.OnCall(oracle.Far(0x1b1, hook), func(q *oracle.Oracle) {
				if len(seenRecords) >= *studyFrames {
					return
				}
				r := frames.RejectedSource()
				if r == nil || r.Steps != q.Steps() || !strings.HasPrefix(r.Reason, "full viewport differs:") {
					return
				}
				records := r.Memory[0x1b10+0xeb60 : 0x1b10+0xeb7c]
				active := false
				for i := 0; i < 4; i++ {
					if records[i*7] != 255 && records[i*7+2] != 0 {
						active = true
					}
				}
				key := sha(records)
				if !active || seenRecords[key] {
					return
				}
				seenRecords[key] = true
				// 此hook沒有同步另一台機器；完整配對仍在命令觀察點及最後核對。
				dump(fmt.Sprintf("phase-%02d", len(seenRecords)-1), r, false)
			})
		}
	}
	input := read(*commands)
	for _, line := range strings.Split(string(input), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		switch f[0] {
		case "run":
			run(num(f[1]))
		case "key":
			scan := num(f[1])
			if scan > 255 {
				panic("scan bounds")
			}
			for _, o := range machines {
				o.KeyDown(uint8(scan))
			}
			run(num(f[2]))
			for _, o := range machines {
				o.KeyUp(uint8(scan))
			}
		case "down", "up":
			scan := num(f[1])
			if scan > 255 {
				panic("scan bounds")
			}
			for _, o := range machines {
				if f[0] == "down" {
					o.KeyDown(uint8(scan))
				} else {
					o.KeyUp(uint8(scan))
				}
			}
		case "shot":
			equal()
			r := frames.RejectedSource()
			if r == nil {
				panic("missing rejected source")
			}
			dump(f[1], r, true)
		default:
			panic(line)
		}
	}
	equal()
	for i, o := range machines {
		path := filepath.Join(*output, fmt.Sprintf("scratch-%d/SAVETEST.usr", i))
		fi, e := os.Stat(path)
		need(e)
		if !bytes.Equal(read(path), save) || !fi.ModTime().Equal(stamp.ModTime()) || len(o.Wrote()) != 0 || len(o.Unimplemented()) != 0 || len(o.HeldKeys()) != 0 {
			panic("normal path side effects")
		}
	}
	write("probe-source.go", read(filepath.Join(*project, "tools/hd_cave_rejected_probe.go")))
	report := map[string]any{"schema": 1, "status": "OBSERVED_REJECTED_CAVE_SOURCES", "go": runtime.Version(), "script_sha256": sha(read(filepath.Join(*output, "probe-source.go"))), "commands_sha256": sha(input), "source_save_sha256": sha(save), "save_unchanged": true, "state_equal": true, "state_injection": false, "points": points, "limits": "被拒絕候選，不是已呈現畫格或HD完成證據。"}
	data, e := json.MarshalIndent(report, "", "  ")
	need(e)
	write("receipt.json", append(data, 10))
}

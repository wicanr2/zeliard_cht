// 062：既有包內存檔經正常F7與平台鍵序，只讀來源連續性，不作GUI或fps驗收。
package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/hd"
)

func main() {
	if len(os.Args) != 3 {
		panic("project output")
	}
	root, out := os.Args[1], os.Args[2]
	need := func(e error) {
		if e != nil {
			panic(e)
		}
	}
	read := func(p string) []byte { b, e := os.ReadFile(p); need(e); return b }
	sha := func(p string) string { return fmt.Sprintf("%x", sha256.Sum256(read(p))) }
	orig := filepath.Join(root, "workplace/orig")
	if sha(filepath.Join(orig, "ZELIARD.EXE")) != "fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341" || sha(filepath.Join(orig, "SAVEM1.USR")) != "03df9abcf62b1712d6b46dd1fc5b23335a4b8711b3dc84875a1498260d37735c" {
		panic("source identity")
	}
	o, e := oracle.Load(filepath.Join(orig, "ZELIARD.EXE"), orig)
	need(e)
	defer o.Close()
	o.SetDOSBoxCycles(3000)
	scratch := filepath.Join(out, "scratch")
	need(os.Mkdir(scratch, 0700))
	save := filepath.Join(orig, "SAVEM1.USR")
	info, e := os.Stat(save)
	need(e)
	copy := filepath.Join(scratch, "SAVETEST.usr")
	need(os.WriteFile(copy, read(save), 0600))
	need(os.Chtimes(copy, info.ModTime(), info.ModTime()))
	o.SetScratch(scratch)
	f := hd.NewCaveFrames(o)
	f.HoldFreshBody(true)
	seen, missing := 0, 0
	var shots []map[string]any
	run := func(n uint64) {
		for n > 0 {
			step := uint64(25000)
			if n < step {
				step = n
			}
			need(o.Run(step))
			n -= step
			if !hd.CaveRuntimeModule(o) {
				continue
			}
			_, _, rgb := o.ScreenRGB()
			p := f.Present(o.Indexed(), rgb, false)
			if p != nil {
				seen++
			} else if c, ok := f.State()["captured"].(uint64); ok && c > 0 {
				missing++
			}
		}
	}
	commands := read(filepath.Join(root, "tools/cave-platform.commands"))
	for _, line := range strings.Split(string(commands), "\n") {
		a := strings.Fields(line)
		if len(a) == 0 || strings.HasPrefix(a[0], "#") {
			continue
		}
		num := func(s string) uint64 { n, e := strconv.ParseUint(s, 0, 64); need(e); return n }
		switch a[0] {
		case "run":
			run(num(a[1]))
		case "key":
			scan := byte(num(a[1]))
			o.KeyDown(scan)
			run(num(a[2]))
			o.KeyUp(scan)
		case "shot":
			_, _, rgb := o.ScreenRGB()
			p := f.Present(o.Indexed(), rgb, false)
			shots = append(shots, map[string]any{"name": a[1], "steps": o.Steps(), "presented": p != nil, "state": f.State()})
			fmt.Println(a[1], p != nil)
		default:
			panic(line)
		}
	}
	b, e := json.MarshalIndent(map[string]any{"source": "normal F7 copy of bundled SAVEM1; prior history unknown", "executable_sha256": sha(filepath.Join(orig, "ZELIARD.EXE")), "save_sha256": sha(save), "commands_sha256": sha(filepath.Join(root, "tools/cave-platform.commands")), "tool_sha256": sha(filepath.Join(root, "tools/hd_platform_stability_probe.go")), "go_version": runtime.Version(), "seen_samples": seen, "missing_after_first_complete": missing, "shots": shots, "scope": "source continuity only; not GUI, rules, random parity or fps acceptance"}, "", "  ")
	need(e)
	need(os.WriteFile(filepath.Join(out, "receipt.json"), append(b, 10), 0600))
}

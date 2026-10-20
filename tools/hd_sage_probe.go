// 正常鍵序的長老手勢擷取；只讀機器，不恢復快照或寫入遊戲狀態。
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wicanr2/dosgolem/oracle"
)

func main() {
	out := os.Getenv("ZELIARD_SAGE_OUTPUT")
	if out == "" {
		panic("output required")
	}
	if e, err := os.ReadDir(out); err != nil || len(e) != 0 {
		panic("fresh empty output required")
	}
	read := func(path string) []byte {
		b, e := os.ReadFile(path)
		if e != nil {
			panic(e)
		}
		return b
	}
	write := func(path string, b []byte) {
		if e := os.WriteFile(filepath.Join(out, path), b, 0644); e != nil {
			panic(e)
		}
	}
	sha := func(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
	project := os.Getenv("ZELIARD_PROJECT")
	source := read(filepath.Join(project, "workplace/full-hd-v1/sage-frames-source-v1/KENJPRO.BIN"))
	savePath := filepath.Join(project, "workplace/prose-v1/npc-normal-v2/scratch/SAVETEST.usr")
	save := read(savePath)
	if sha(save) != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
		panic("save identity")
	}
	scratch := filepath.Join(out, "scratch")
	if e := os.Mkdir(scratch, 0755); e != nil {
		panic(e)
	}
	if e := os.WriteFile(filepath.Join(scratch, "SAVETEST.usr"), save, 0644); e != nil {
		panic(e)
	}
	original := os.Getenv("ZELIARD_ORIG")
	o, e := oracle.Load(filepath.Join(original, "ZELIARD.EXE"), original)
	if e != nil {
		panic(e)
	}
	defer o.Close()
	o.SetScratch(scratch)
	o.SetDOSBoxCycles(3000)
	o.TraceFiles()
	var records []map[string]any
	var checkpoints []map[string]any
	seen := map[string]bool{}
	seenKeys := map[string]bool{}
	sample := func(label string, force bool) {
		mem := o.Bytes(oracle.Far(0, 0), 1<<20)
		p := mem[0x1b10:0x11b10]
		if !bytes.Equal(p[0xa000:0xbb12], source[:0x1b12]) {
			return
		}
		raw := o.Indexed()
		x, y := int(p[0xbb13])*8, int(p[0xbb12])
		if (x != 56 && x != 112) || y != 23 {
			return
		}
		crop := make([]byte, 6144)
		for dy := 0; dy < 64; dy++ {
			copy(crop[dy*96:(dy+1)*96], raw[(y+dy)*320+x:(y+dy)*320+x+96])
		}
		key := sha(crop)
		if seen[key] && !force {
			return
		}
		seen[key] = true
		stem := fmt.Sprintf("frame-%03d", len(records))
		if len(records) >= 80 {
			panic("bounded source count")
		}
		palette := []byte{}
		for _, c := range o.Palette() {
			palette = append(palette, c[:]...)
		}
		im := image.NewRGBA(image.Rect(0, 0, 320, 200))
		for i, v := range raw {
			at := int(v) * 3
			im.SetRGBA(i%320, i/320, color.RGBA{palette[at], palette[at+1], palette[at+2], 255})
		}
		f, e := os.Create(filepath.Join(out, stem+".png"))
		if e != nil {
			panic(e)
		}
		if e = png.Encode(f, im); e != nil {
			panic(e)
		}
		f.Close()
		write(stem+".indexed", raw)
		write(stem+".memory", mem)
		write(stem+".palette", palette)
		records = append(records, map[string]any{"file": stem, "label": label, "steps": o.Steps(), "cycles": o.Cycles(), "origin": []int{x, y}, "crop_sha256": key, "memory_sha256": sha(mem), "indexed_sha256": sha(raw), "palette_sha256": sha(palette), "regs": o.Regs()})
		fmt.Println("SOURCE", stem, label, key, o.Steps())
	}
	// 只記實際到達的選擇，不推定來源旗標或時序。
	o.OnCall(oracle.Far(0x1b1, 0xaa16), func(q *oracle.Oracle) {
		if bytes.Equal(q.Bytes(oracle.Far(0x1b1, 0xa000), 32), source[:32]) {
			seenKeys[fmt.Sprintf("hand-%d", q.AX()&255)] = true
		}
	})
	run := func(n uint64) {
		for n > 0 {
			step := uint64(100000)
			if n < step {
				step = n
			}
			if o.Steps()+step > 350000000 {
				panic("bounded350M")
			}
			if e := o.Run(step); e != nil {
				panic(e)
			}
			n -= step
			sample("periodic", false)
		}
	}
	commands := read(os.Getenv("ZELIARD_SAGE_COMMANDS"))
	num := func(s string) uint64 {
		n, e := strconv.ParseUint(s, 0, 64)
		if e != nil {
			panic(e)
		}
		return n
	}
	for _, line := range strings.Split(string(commands), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		switch f[0] {
		case "run":
			run(num(f[1]))
		case "key":
			o.KeyDown(uint8(num(f[1])))
			run(num(f[2]))
			o.KeyUp(uint8(num(f[1])))
		case "shot":
			sample(f[1], true)
			checkpoints = append(checkpoints, map[string]any{"name": f[1], "steps": o.Steps(), "regs": o.Regs(), "indexed_sha256": sha(o.Indexed())})
		default:
			panic(line)
		}
	}
	result := map[string]any{"status": "OBSERVED_NORMAL_SAGE_FRAMES", "tool": "dosgolem oracle", "command_sha256": sha(commands), "script_sha256": sha(read(filepath.Join(project, "tools/hd_sage_probe.go"))), "module_sha256": sha(source), "source_save_sha256": sha(save), "source_save_unchanged": bytes.Equal(save, read(savePath)), "scratch_save_unchanged": bytes.Equal(save, read(filepath.Join(scratch, "SAVETEST.usr"))), "state_injection": false, "snapshot_restore": false, "selected_hand_indices": seenKeys, "frames": records, "checkpoints": checkpoints, "file_ops": o.FileOps(), "final_held_keys": o.HeldKeys()}
	b, e := json.MarshalIndent(result, "", "  ")
	if e != nil {
		panic(e)
	}
	write("receipt.json", append(b, 10))
}

// 正常鍵序的室內圖面擷取；不注入狀態，不代替成對原版或GUI驗收。
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
	"runtime"
	"strconv"
	"strings"

	"github.com/wicanr2/dosgolem/oracle"
)

func main() {
	out := os.Getenv("ZELIARD_ROOM_OUTPUT")
	entries, e := os.ReadDir(out)
	if e != nil || len(entries) != 0 {
		panic("fresh empty output required")
	}
	read := func(path string) []byte {
		b, e := os.ReadFile(path)
		if e != nil {
			panic(e)
		}
		return b
	}
	write := func(name string, b []byte) {
		if e := os.WriteFile(filepath.Join(out, name), b, 0644); e != nil {
			panic(e)
		}
	}
	sha := func(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
	project, original := os.Getenv("ZELIARD_PROJECT"), os.Getenv("ZELIARD_ORIG")
	source := read(os.Getenv("ZELIARD_ROOM_MODULE"))
	indexBytes := read(os.Getenv("ZELIARD_ROOM_INDEX"))
	var index struct {
		File          string `json:"file"`
		SHA           string `json:"sha256"`
		GraphicsBytes int    `json:"graphics_bytes"`
		GraphicsSHA   string `json:"graphics_sha256"`
	}
	if e = json.Unmarshal(indexBytes, &index); e != nil {
		panic(e)
	}
	if filepath.Base(os.Getenv("ZELIARD_ROOM_MODULE")) != index.File || sha(source) != index.SHA {
		panic("module identity")
	}
	probeSource := read(filepath.Join(project, "tools/hd_room_probe.go"))
	write("probe-source.go", probeSource)
	write("source-index.json", indexBytes)
	savePath := filepath.Join(project, "workplace/prose-v1/npc-normal-v2/scratch/SAVETEST.usr")
	save := read(savePath)
	if sha(save) != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
		panic("source USR identity")
	}
	info, e := os.Stat(savePath)
	if e != nil {
		panic(e)
	}
	scratch := filepath.Join(out, "scratch")
	if e = os.Mkdir(scratch, 0755); e != nil {
		panic(e)
	}
	copyPath := filepath.Join(scratch, "SAVETEST.usr")
	if e = os.WriteFile(copyPath, save, 0644); e != nil {
		panic(e)
	}
	if e = os.Chtimes(copyPath, info.ModTime(), info.ModTime()); e != nil {
		panic(e)
	}
	o, e := oracle.Load(filepath.Join(original, "ZELIARD.EXE"), original)
	if e != nil {
		panic(e)
	}
	defer o.Close()
	o.SetScratch(scratch)
	o.SetDOSBoxCycles(3000)
	o.TraceFiles()
	var audio *oracle.Audio
	adlib := os.Getenv("ZELIARD_ROOM_ADLIB") == "1"
	if adlib {
		config := bytes.ReplaceAll(read(filepath.Join(original, "RESOURCE.CFG")), []byte("MSCSTD.DRV"), []byte("MSCADLIB.DRV"))
		if e = os.WriteFile(filepath.Join(scratch, "RESOURCE.CFG"), config, 0644); e != nil {
			panic(e)
		}
		o.SetAdLib(true)
		audio = o.NewAudio(44100)
	}
	var frames, points []map[string]any
	seen := map[string]bool{}
	selected := map[int]bool{}
	if address := os.Getenv("ZELIARD_ROOM_POSE_ENTRY"); address != "" {
		n, e := strconv.ParseUint(address, 0, 16)
		if e != nil {
			panic(e)
		}
		o.OnCall(oracle.Far(0x1b1, uint16(n)), func(q *oracle.Oracle) {
			if bytes.Equal(q.Bytes(oracle.Far(0x1b1, 0xa000), 32), source[:32]) {
				selected[int(q.AX()&255)] = true
			}
		})
	}
	sample := func(label string, force bool) {
		if o.VideoMode() != 0x13 || !bytes.Equal(o.Bytes(oracle.Far(0x1b1, 0xa000), 32), source[:32]) {
			return
		}
		if index.GraphicsBytes > 0 && sha(o.Bytes(oracle.Far(0x11b1, 0x8000), index.GraphicsBytes)) != index.GraphicsSHA {
			return
		}
		raw := o.Indexed()
		if len(raw) != 64000 {
			return
		}
		crop := make([]byte, 6144)
		for y := 0; y < 64; y++ {
			copy(crop[y*96:(y+1)*96], raw[(23+y)*320+56:(23+y)*320+152])
		}
		key := sha(crop)
		if seen[key] && !force {
			return
		}
		seen[key] = true
		if len(frames) >= 160 {
			panic("bounded160frames")
		}
		mem := o.Bytes(oracle.Far(0, 0), 1<<20)
		palette := []byte{}
		for _, c := range o.Palette() {
			palette = append(palette, c[:]...)
		}
		im := image.NewRGBA(image.Rect(0, 0, 320, 200))
		for i, v := range raw {
			at := int(v) * 3
			im.SetRGBA(i%320, i/320, color.RGBA{palette[at], palette[at+1], palette[at+2], 255})
		}
		stem := fmt.Sprintf("frame-%03d", len(frames))
		f, e := os.Create(filepath.Join(out, stem+".png"))
		if e != nil {
			panic(e)
		}
		if e = png.Encode(f, im); e != nil {
			panic(e)
		}
		f.Close()
		write(stem+".memory", mem)
		write(stem+".indexed", raw)
		write(stem+".palette", palette)
		frames = append(frames, map[string]any{"file": stem, "label": label, "steps": o.Steps(), "cycles": o.Cycles(), "crop_sha256": key, "indexed_sha256": sha(raw), "memory_sha256": sha(mem), "palette_sha256": sha(palette), "origin": []int{56, 23}, "regs": o.Regs(), "text_cursor": o.Word(oracle.Far(0x1b1, 0xff4c)), "raw0086": o.Word(oracle.Far(0x1b1, 0x86))})
		fmt.Println("SOURCE", stem, label, key, o.Steps())
	}
	run := func(n uint64) {
		for n > 0 {
			step := uint64(25000)
			if n < step {
				step = n
			}
			if o.Steps()+step > 350000000 {
				panic("bounded350Msteps")
			}
			if e := o.Run(step); e != nil {
				panic(e)
			}
			if audio != nil {
				_ = audio.Render()
			}
			n -= step
			sample("periodic", false)
		}
	}
	number := func(s string) uint64 {
		n, e := strconv.ParseUint(s, 0, 64)
		if e != nil {
			panic(e)
		}
		return n
	}
	commands := read(os.Getenv("ZELIARD_ROOM_COMMANDS"))
	for _, line := range strings.Split(string(commands), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		switch f[0] {
		case "run":
			run(number(f[1]))
		case "key":
			o.KeyDown(uint8(number(f[1])))
			run(number(f[2]))
			o.KeyUp(uint8(number(f[1])))
		case "down":
			o.KeyDown(uint8(number(f[1])))
		case "up":
			o.KeyUp(uint8(number(f[1])))
		case "shot":
			sample(f[1], true)
			if os.Getenv("ZELIARD_ROOM_POINT_MEMORY") == "1" {
				write(fmt.Sprintf("point-%03d.memory", len(points)), o.Bytes(oracle.Far(0, 0), 1<<20))
				write(fmt.Sprintf("point-%03d.indexed", len(points)), o.Indexed())
				var palette []byte
				for _, c := range o.Palette() {
					palette = append(palette, c[:]...)
				}
				write(fmt.Sprintf("point-%03d.palette", len(points)), palette)
			}
			points = append(points, map[string]any{"name": f[1], "steps": o.Steps(), "cycles": o.Cycles(), "regs": o.Regs(), "indexed_sha256": sha(o.Indexed()), "memory_sha256": sha(o.Bytes(oracle.Far(0, 0), 1<<20)), "raw0086": o.Word(oracle.Far(0x1b1, 0x86))})
		default:
			panic(line)
		}
	}
	result := map[string]any{"schema": 1, "status": "OBSERVED_NORMAL_ROOM_FRAMES", "go": runtime.Version(), "tool": "dosgolem oracle", "module": index.File, "module_sha256": sha(source), "source_index_sha256": sha(indexBytes), "script_sha256": sha(probeSource), "command_sha256": sha(commands), "source_save_sha256": sha(save), "source_save_unchanged": bytes.Equal(save, read(savePath)), "scratch_save_unchanged": bytes.Equal(save, read(copyPath)), "source_save_mtime_ns": info.ModTime().UnixNano(), "scratch_mtime_ns": func() int64 {
		info, e := os.Stat(copyPath)
		if e != nil {
			panic(e)
		}
		return info.ModTime().UnixNano()
	}(), "snapshot_restore": false, "state_injection": false, "frames": frames, "points": points, "selected_pose_indices": selected, "file_ops": o.FileOps(), "held_keys": o.HeldKeys(), "scope": "正常鍵序來源擷取；不是成對原版或GUI驗收"}
	result["adlib"] = adlib
	data, e := json.MarshalIndent(result, "", "  ")
	if e != nil {
		panic(e)
	}
	write("receipt.json", append(data, 10))
}

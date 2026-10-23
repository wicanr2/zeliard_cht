// 066：原版與 patch 版驅動同步執行同一鍵序，證明字形閘只改「畫不畫進畫面」。
// patch 版掛上真實的城鎮與對話覆繪層，認領照正常流程發生。每個切片點比對：
// 1. patch 版守衛視圖（VRAM 疊上平面）等於原版 VRAM；
// 2. 1 MiB 記憶體相同，排除 B000 段、patch 位址、目的段變數與 VRAM，31B1 方框緩衝以平面重建；
// 3. 步數、週期與 CS:IP 相同。
// 用法：go run tools/glyph_patch_diff.go -orig <原版> -work <空目錄> -commands <鍵序> -font <font.hex.gz> -save <USR> -out <新JSON>
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
	"github.com/wicanr2/zeliard_cht/internal/prototype/glyphpatch"
	"github.com/wicanr2/zeliard_cht/internal/prototype/overlay"
)

type mismatch struct {
	Step   uint64 `json:"step"`
	Kind   string `json:"kind"`
	Offset uint32 `json:"offset"`
	Want   uint8  `json:"original"`
	Got    uint8  `json:"patched"`
	IP     string `json:"ip"`
	Stack  string `json:"stack,omitempty"`
}

func main() {
	orig := flag.String("orig", "", "原版目錄（唯讀）")
	work := flag.String("work", "", "可寫空目錄")
	commands := flag.String("commands", "", "正常鍵序檔")
	fontPath := flag.String("font", "", "中文字型 font.hex.gz")
	save := flag.String("save", "", "可選：兩側可寫層的起始存檔")
	slice := flag.Uint64("slice", 20000, "比對間隔指令數")
	out := flag.String("out", "", "輸出 JSON，須不存在")
	tracePixel := flag.Int("trace-pixel", -1, "可選：記錄兩台機器對這個 VRAM 偏移的每次寫入")
	flag.Parse()
	if *orig == "" || *work == "" || *commands == "" || *fontPath == "" || *out == "" {
		log.Fatal("須指定 -orig、-work、-commands、-font、-out")
	}
	font, err := bitmap.Load(*fontPath)
	if err != nil {
		log.Fatal(err)
	}
	machines := make([]*oracle.Oracle, 2)
	for i, name := range []string{"original", "patched"} {
		scratch := filepath.Join(*work, name)
		if err := os.Mkdir(scratch, 0700); err != nil {
			log.Fatal(err)
		}
		if *save != "" {
			data, err := os.ReadFile(*save)
			if err != nil {
				log.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(scratch, filepath.Base(*save)), data, 0600); err != nil {
				log.Fatal(err)
			}
		}
		if i == 1 {
			if _, err := glyphpatch.Install(*orig, scratch); err != nil {
				log.Fatal(err)
			}
		}
		o, err := oracle.Load(filepath.Join(*orig, "ZELIARD.EXE"), *orig)
		if err != nil {
			log.Fatal(err)
		}
		defer o.Close()
		o.SetScratch(scratch)
		o.SetDOSBoxCycles(3000)
		machines[i] = o
	}
	a, b := machines[0], machines[1]
	gate := glyphpatch.Enable(b)
	var trace []string
	if *tracePixel >= 0 {
		at := uint32(*tracePixel)
		for _, w := range []struct {
			name string
			o    *oracle.Oracle
			base uint32
		}{{"original A000", a, 0xa0000}, {"patched A000", b, 0xa0000}, {"patched B000", b, 0xb0000}} {
			w := w
			w.o.WatchEveryWrite(w.base+at, w.base+at, func(_ uint32, old, v uint8) {
				trace = append(trace, fmt.Sprintf("%d %s %v %d->%d", w.o.Steps(), w.name, w.o.IP(), old, v))
			})
		}
	}
	// 與正式封包相同的中文層組合，讓每一種認領都照正常流程發生。
	sar1, sar2 := filepath.Join(*orig, "ZELRES1.SAR"), filepath.Join(*orig, "ZELRES2.SAR")
	if _, err := overlay.NewIntroStory(b, sar1, "text/intro-story-prototype.tsv", font); err != nil {
		log.Fatal(err)
	}
	if _, err := overlay.NewIntroThreat(b, sar1, "text/intro-threat-prototype.tsv", font); err != nil {
		log.Fatal(err)
	}
	if _, err := overlay.NewScrollBlocks(b, sar1, overlay.ScrollCatalogs{First: "text/intro-scroll-prototype.tsv", Credits: "text/intro-credits-prototype.tsv", Outro: "text/intro-outro-prototype.tsv"}, font); err != nil {
		log.Fatal(err)
	}
	dialogue, err := overlay.NewDialogue(b, sar2, "text/dialogue-prototype.tsv", font)
	if err != nil {
		log.Fatal(err)
	}
	labels, err := overlay.NewLabels(b, sar2, "text/labels-prototype.tsv", font)
	if err != nil {
		log.Fatal(err)
	}
	dialogue.UseLabels(labels)
	if _, err := overlay.NewHUD(b, *orig, "text/hud-prototype.tsv", font); err != nil {
		log.Fatal(err)
	}
	if _, err := overlay.NewTownDialogue(b, sar2, "text/town-prototype.tsv", font); err != nil {
		log.Fatal(err)
	}
	excluded := func(lin uint32) bool {
		switch {
		case lin >= 0xa0000 && lin < 0xc0000:
			return true // VRAM 另以守衛視圖比對；B000 為不顯示的字形去處
		case lin >= 0x1b10+0x2832 && lin < 0x1b10+0x2835, lin >= 0x1b10+0x22fe && lin < 0x1b10+0x2301, lin >= 0x1b10+0x2333 && lin < 0x1b10+0x2336:
			return true
		case lin >= 0x1b10+glyphpatch.VarOffset && lin < 0x1b10+glyphpatch.VarOffset+2:
			return true
		}
		return false
	}
	var found []mismatch
	compares, total, residue := 0, 0, 0
	// 全部差異依種類、原版當下 CS:IP 與百萬步分組，不只看前幾筆。
	groups := map[string]int{}
	check := func() {
		compares++
		if a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || a.IP() != b.IP() {
			log.Fatalf("執行軌跡分岔：steps %d/%d cycles %d/%d ip %v/%v", a.Steps(), b.Steps(), a.Cycles(), b.Cycles(), a.IP(), b.IP())
		}
		want, got := a.Indexed(), glyphpatch.Indexed(b)
		if !bytes.Equal(want, got) {
			for i := range want {
				if want[i] != got[i] {
					total++
					groups[fmt.Sprintf("vram %v @%dM", a.IP(), a.Steps()/1000000)]++
					if len(found) < 64 {
						found = append(found, mismatch{a.Steps(), "vram", uint32(i), want[i], got[i], fmt.Sprint(a.IP()), ""})
					}
				}
			}
		}
		ma, mb := a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)
		for off, v := range gate.Buffer() {
			mb[0x31b10+uint32(off)] = v
		}
		// 壓下期間 AX=B000：中斷處理推入的暫存器與返回後的堆疊殘值是原版 A0、patch 版 B0。
		// 只接受這一種 byte 差異，且須位於目前 SS:SP 上下 512 bytes 內。
		regs := b.Regs()
		top := uint32(regs.SS)*16 + uint32(regs.SP)
		if !bytes.Equal(ma, mb) {
			for i := range ma {
				if ma[i] != mb[i] && ma[i] == 0xa0 && mb[i] == 0xb0 && uint32(i)+512 >= top && uint32(i) < top+512 {
					residue++
					continue
				}
				if ma[i] != mb[i] && !excluded(uint32(i)) {
					total++
					groups[fmt.Sprintf("memory %05X %v @%dM", i, a.IP(), a.Steps()/1000000)]++
					if len(found) < 64 {
						found = append(found, mismatch{a.Steps(), "memory", uint32(i), ma[i], mb[i], fmt.Sprint(a.IP()), fmt.Sprintf("%04X:%04X", regs.SS, regs.SP)})
					}
				}
			}
		}
	}
	exited := false
	run := func(n uint64) {
		for n > 0 && !exited {
			s := min(*slice, n)
			for _, o := range machines {
				if err := o.Run(s); err != nil {
					var exit *oracle.ExitError
					if !errors.As(err, &exit) {
						log.Fatal(err)
					}
					exited = true
				}
			}
			n -= s
			check()
		}
	}
	data, err := os.ReadFile(*commands)
	if err != nil {
		log.Fatal(err)
	}
	num := func(s string) uint64 {
		v, e := strconv.ParseUint(s, 0, 64)
		if e != nil {
			log.Fatal(e)
		}
		return v
	}
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		switch f[0] {
		case "run":
			run(num(f[1]))
		case "key", "keys":
			var scans []uint8
			for _, s := range strings.Split(f[1], ",") {
				scans = append(scans, uint8(num(s)))
			}
			for _, o := range machines {
				for _, s := range scans {
					o.KeyDown(s)
				}
			}
			run(num(f[2]))
			for _, o := range machines {
				for _, s := range scans {
					o.KeyUp(s)
				}
			}
		case "down", "up":
			for _, o := range machines {
				if f[0] == "down" {
					o.KeyDown(uint8(num(f[1])))
				} else {
					o.KeyUp(uint8(num(f[1])))
				}
			}
		case "king":
			// 與 save_load_test 的 king 相同：冷開機後固定 Space、右移與進門，到國王初訪。
			type event struct {
				at   uint64
				scan uint8
				down bool
			}
			var events []event
			for _, s := range []uint64{12000000, 20000000, 28000000, 36000000, 44000000, 52000000, 86000000, 94000000, 102000000, 110000000, 118000000, 126000000} {
				events = append(events, event{s, 0x39, true}, event{s + 50000, 0x39, false})
			}
			events = append(events, event{60000000, 0x4d, true}, event{62000000, 0x4d, false}, event{64000000, 0x48, true}, event{64250000, 0x48, false})
			sort.SliceStable(events, func(i, j int) bool { return events[i].at < events[j].at })
			for _, e := range events {
				run(e.at - a.Steps())
				for _, o := range machines {
					if e.down {
						o.KeyDown(e.scan)
					} else {
						o.KeyUp(e.scan)
					}
				}
			}
			run(130000000 - a.Steps())
		case "shot", "quit":
		default:
			log.Fatalf("不支援 %q", f[0])
		}
	}
	status := "PASS_GUARD_VIEW_AND_MEMORY_EQUAL"
	if total > 0 {
		status = "FAIL_DIFFERENCES"
	}
	report := map[string]any{"status": status, "steps": a.Steps(), "compares": compares, "differences": total, "stack_segment_residue": residue, "first_differences": found, "difference_groups": groups, "pixel_trace": trace, "suppressed_glyphs": gate.Suppressed, "passed_glyphs": gate.Passed, "ds_anomalies": gate.Anomalies, "slice": *slice, "commands": *commands}
	body, _ := json.MarshalIndent(report, "", " ")
	f, err := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err = f.Write(append(body, '\n')); err != nil {
		log.Fatal(err)
	}
	fmt.Println(status, "steps", a.Steps(), "compares", compares, "differences", total, "suppressed", gate.Suppressed, "passed", gate.Passed)
}

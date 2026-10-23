// 066：在原版正常鍵序下蒐集字形改寫的前提證據，只讀原版。
// 1. GMMCGA.BIN 的開啟與讀取量，判斷載入端是否依檔案長度讀入；
// 2. 驅動檔尾之後 01B1:2CC9..2FFF 是否有任何讀寫；
// 3. 27E9 與 2345 兩個字形常式實際出現的 FF77、AH 與色號。
// 用法：go run tools/glyph_patch_study.go -orig <原版目錄> -scratch <可寫空目錄> -commands <鍵序> -out <新JSON>
package main

import (
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
)

func main() {
	orig := flag.String("orig", "", "原版目錄（唯讀）")
	scratch := flag.String("scratch", "", "原版可寫層，須為空目錄")
	commands := flag.String("commands", "", "正常鍵序檔")
	out := flag.String("out", "", "輸出 JSON，須不存在")
	save := flag.String("save", "", "可選：先複製到可寫層的存檔")
	flag.Parse()
	if *orig == "" || *scratch == "" || *commands == "" || *out == "" {
		log.Fatal("須指定 -orig、-scratch、-commands、-out")
	}
	if *save != "" {
		data, err := os.ReadFile(*save)
		if err != nil {
			log.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(*scratch, filepath.Base(*save)), data, 0600); err != nil {
			log.Fatal(err)
		}
	}
	o, err := oracle.Load(filepath.Join(*orig, "ZELIARD.EXE"), *orig)
	if err != nil {
		log.Fatal(err)
	}
	defer o.Close()
	o.SetScratch(*scratch)
	o.SetDOSBoxCycles(3000)
	o.SetAdLib(true)

	// 驅動檔尾之後的區段：線性 0x1B10+0x2CC9 .. 0x1B10+0x2FFF。
	lo, hi := uint32(0x1b10+0x2cc9), uint32(0x1b10+0x2fff)
	writes := o.WatchWritesAt(lo, hi)
	reads := o.WatchReadsAt(lo, hi)
	glyph := map[string]int{}
	o.OnCall(oracle.Far(0x1b1, 0x27e9), func(o *oracle.Oracle) {
		ah := uint8(o.AX() >> 8)
		ff77 := o.Byte(oracle.Far(0x1b1, 0xff77))
		color := o.Byte(oracle.Far(0x1b1, 0x24ea+uint16(ah)))
		if ff77 != 0 {
			color = ah * 17
		}
		glyph[fmt.Sprintf("27e9 ff77=%02x ah=%d color=%d caller=%04x", ff77, ah, color, o.StackWord(0))]++
	})
	o.OnCall(oracle.Far(0x1b1, 0x2345), func(o *oracle.Oracle) {
		glyph[fmt.Sprintf("2345 color=%d shadow=%d", o.Byte(oracle.Far(0x1b1, 0x2cbd)), o.Byte(oracle.Far(0x1b1, 0x2cbe)))]++
	})
	data, err := os.ReadFile(*commands)
	if err != nil {
		log.Fatal(err)
	}
	run := func(n uint64) {
		if err := o.Run(n); err != nil {
			var exit *oracle.ExitError
			if !errors.As(err, &exit) {
				log.Fatal(err)
			}
		}
	}
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		num := func(s string) uint64 {
			v, e := strconv.ParseUint(s, 0, 64)
			if e != nil {
				log.Fatal(e)
			}
			return v
		}
		switch f[0] {
		case "run":
			run(num(f[1]))
		case "key", "keys":
			var scans []uint8
			for _, s := range strings.Split(f[1], ",") {
				scans = append(scans, uint8(num(s)))
			}
			for _, s := range scans {
				o.KeyDown(s)
			}
			run(num(f[2]))
			for _, s := range scans {
				o.KeyUp(s)
			}
		case "down":
			o.KeyDown(uint8(num(f[1])))
		case "up":
			o.KeyUp(uint8(num(f[1])))
		case "shot", "quit":
		default:
			log.Fatalf("不支援 %q", f[0])
		}
	}
	var ops []oracle.FileOp
	for _, op := range o.FileOps() {
		if strings.EqualFold(filepath.Base(op.Name), "GMMCGA.BIN") {
			ops = append(ops, op)
		}
	}
	keys := make([]string, 0, len(glyph))
	for k := range glyph {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	counts := map[string]int{}
	for _, k := range keys {
		counts[k] = glyph[k]
	}
	report := map[string]any{"steps": o.Steps(), "driver_file_ops": ops, "tail_writes": len(*writes), "tail_reads": len(*reads), "glyph_calls": counts}
	if len(*writes) > 0 {
		report["tail_write_examples"] = (*writes)[:min(8, len(*writes))]
	}
	if len(*reads) > 0 {
		report["tail_read_examples"] = (*reads)[:min(8, len(*reads))]
	}
	b, _ := json.MarshalIndent(report, "", " ")
	f, err := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err = f.Write(append(b, '\n')); err != nil {
		log.Fatal(err)
	}
}

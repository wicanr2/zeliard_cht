// 在dosgolem執行固定原版三平面前處理，核對離線轉換；只作格式研究。
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wicanr2/dosgolem/oracle"
)

func planarNeed(e error) {
	if e != nil {
		panic(e)
	}
}
func planarRead(p string) []byte { b, e := os.ReadFile(p); planarNeed(e); return b }
func planarHash(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }

func main() {
	project := flag.String("project", "", "唯讀專案")
	converted := flag.String("converted", "", "離線轉換輸出")
	output := flag.String("output", "", "既有空輸出目錄")
	flag.Parse()
	entries, e := os.ReadDir(*output)
	planarNeed(e)
	if len(entries) != 0 {
		panic("fresh output required")
	}
	resident := planarRead(filepath.Join(*project, "workplace/probe-intro-memory/memory-01B1_0000.bin"))
	if planarHash(resident) != "c9f188ea847e8d364ccf8dbc9f6a9b9c74e192c0c83bd7a012b552e6d89b13d8" {
		panic("resident identity")
	}
	var manifest struct {
		Records []struct {
			Archive    string `json:"archive"`
			Record     int    `json:"record"`
			Source     string `json:"source"`
			SourceHash string `json:"source_sha256"`
			Padded     int    `json:"padded_bytes"`
			Output     string `json:"output"`
		} `json:"records"`
	}
	planarNeed(json.Unmarshal(planarRead(filepath.Join(*converted, "manifest.json")), &manifest))
	var receipts []map[string]any
	for _, r := range manifest.Records {
		raw := planarRead(filepath.Join(*project, r.Source))
		if planarHash(raw) != r.SourceHash || r.Padded != (len(raw)+47)/48*48 || r.Padded+0x3000 > 0xffff {
			panic("source identity or bounds")
		}
		image := make([]byte, 0x3000+r.Padded)
		copy(image[0x2c2a:0x2cc9], resident[0x2c2a:0x2cc9])
		copy(image[0x3000:], raw)
		count := r.Padded / 48
		stub := []byte{0x0e, 0x1f, 0xbe, 0x00, 0x30, 0xb9, byte(count), byte(count >> 8), 0xe8, 0, 0, 0xb8, 0, 0x4c, 0xcd, 0x21}
		binary.LittleEndian.PutUint16(stub[9:11], 0x2c2a-(0x100+11))
		copy(image[0x100:], stub)
		header := make([]byte, 32)
		size := len(header) + len(image)
		for i, v := range []uint16{0x5a4d, uint16(size % 512), uint16((size + 511) / 512), 0, 2, 0x4000, 0xffff, 0, 0xfffe, 0, 0x100, 0, 0x1c, 0} {
			binary.LittleEndian.PutUint16(header[i*2:], v)
		}
		fixture := filepath.Join(*output, fmt.Sprintf("%s-%02d.exe", r.Archive, r.Record))
		planarNeed(os.WriteFile(fixture, append(header, image...), 0644))
		o, e := oracle.Load(fixture, *output)
		planarNeed(e)
		cs := o.IP().Seg
		e = o.Run(8_000_000)
		var exited *oracle.ExitError
		if !errors.As(e, &exited) {
			panic(fmt.Sprintf("bounded converter %s #%d %v %v", r.Archive, r.Record, e, o.IP()))
		}
		actual := o.Bytes(oracle.Far(cs, 0x3000), r.Padded)
		expected := planarRead(filepath.Join(*converted, r.Output))
		matched := bytes.Equal(actual, expected) && exited.Code == 0 && int(o.Regs().DI) == 0x3000+r.Padded
		planarNeed(os.WriteFile(filepath.Join(*output, r.Output), actual, 0644))
		receipts = append(receipts, map[string]any{"archive": r.Archive, "record": r.Record, "source_sha256": r.SourceHash,
			"fixture_sha256": planarHash(planarRead(fixture)), "output_sha256": planarHash(actual), "bytes": len(actual),
			"matched": matched, "steps": o.Steps(), "cycles": o.Cycles(), "exit_code": exited.Code})
		o.Close()
		if !matched {
			panic(fmt.Sprintf("planar mismatch %s #%d", r.Archive, r.Record))
		}
	}
	report := map[string]any{"schema": 1, "status": "PASS_EXTRACTED_ORIGINAL_PLANAR", "go_version": runtime.Version(),
		"records": receipts, "original_routine_sha256": planarHash(resident[0x2c2a:0x2cc9]), "resident_sha256": planarHash(resident),
		"script_sha256": planarHash(planarRead(filepath.Join(*project, "tools/hd_planar_oracle.go"))),
		"address_space": "原版dosgolem01B1:2C2A..2CC8，測試MZ保留段內位址，輸入／輸出CS:3000，scratch為CS+3000段",
		"limits":        "隔離原常式及補零研究buffer，不是正常遊戲路徑、GRP完整格式或HD完成證據"}
	b, e := json.MarshalIndent(report, "", "  ")
	planarNeed(e)
	planarNeed(os.WriteFile(filepath.Join(*output, "receipt.json"), append(b, 10), 0644))
	fmt.Println("PASS_EXTRACTED_ORIGINAL_PLANAR", len(receipts))
}

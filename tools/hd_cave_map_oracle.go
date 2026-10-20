// 在dosgolem執行原版6CED..6D6D柱狀地圖解壓；只作格式研究，不當正常路徑。
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

func mapNeed(e error) {
	if e != nil {
		panic(e)
	}
}
func mapRead(p string) []byte { b, e := os.ReadFile(p); mapNeed(e); return b }
func mapHash(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }

func main() {
	project := flag.String("project", "", "唯讀專案")
	maps := flag.String("maps", "", "離線地圖來源")
	output := flag.String("output", "", "既有空輸出目錄")
	flag.Parse()
	entries, e := os.ReadDir(*output)
	mapNeed(e)
	if len(entries) != 0 {
		panic("fresh output required")
	}
	primary := mapRead(filepath.Join(*project, "workplace/combat-v1/cave-entered.primary"))
	if mapHash(primary) != "12d91918212c4688a575bc31d395c133c7232f99ead769c7cfbc680333545c93" {
		panic("original primary identity")
	}
	var manifest struct {
		Records []struct {
			Record      int    `json:"record"`
			Name        string `json:"name"`
			Source      string `json:"source"`
			Hash        string `json:"payload_sha256"`
			Columns     string `json:"columns"`
			ColumnsHash string `json:"columns_sha256"`
			Width       int    `json:"width"`
			Stop        int    `json:"compressed_end"`
		} `json:"records"`
	}
	mapNeed(json.Unmarshal(mapRead(filepath.Join(*maps, "receipt.json")), &manifest))
	var receipts []map[string]any
	for _, entry := range manifest.Records {
		source, expected := mapRead(filepath.Join(*maps, entry.Source)), mapRead(filepath.Join(*maps, entry.Columns))
		if mapHash(source) != entry.Hash || mapHash(expected) != entry.ColumnsHash || len(expected) != entry.Width*64 || entry.Width < 1 || entry.Width > 1024 || len(source) > 0x2000 {
			panic("source identity or bounds")
		}
		program := make([]byte, 65536)
		copy(program[0x6ced:0x6d6e], primary[0x6ced:0x6d6e])
		copy(program[0xc000:], source)
		// DS=ES=CS，SI讀C01B。每欄重設DI=8000，保留原版36-byte列距。
		stub := []byte{0x0e, 0x1f, 0x0e, 0x07, 0xbe, 0x1b, 0xc0, 0xbd, byte(entry.Width), byte(entry.Width >> 8), 0xbf, 0x00, 0x80, 0xe8, 0, 0, 0x4d, 0x75, 0xf7, 0xb8, 0x00, 0x4c, 0xcd, 0x21}
		binary.LittleEndian.PutUint16(stub[14:16], 0x6d57-0x110)
		copy(program[0x100:], stub)
		header := make([]byte, 32)
		size := len(header) + len(program)
		for i, v := range []uint16{0x5a4d, uint16(size % 512), uint16((size + 511) / 512), 0, 2, 0x4000, 0xffff, 0, 0xfffe, 0, 0x100, 0, 0x1c, 0} {
			binary.LittleEndian.PutUint16(header[i*2:], v)
		}
		fixture := filepath.Join(*output, entry.Name+".exe")
		mapNeed(os.WriteFile(fixture, append(header, program...), 0644))
		o, e := oracle.Load(fixture, *output)
		mapNeed(e)
		cs := o.IP().Seg
		var actual []byte
		o.OnCall(oracle.Far(cs, 0x110), func(o *oracle.Oracle) {
			if len(actual) >= len(expected) {
				panic("column limit")
			}
			for y := 0; y < 64; y++ {
				actual = append(actual, o.Byte(oracle.Far(cs, uint16(0x8000+y*36))))
			}
		})
		e = o.Run(8000000)
		var exited *oracle.ExitError
		if !errors.As(e, &exited) {
			panic(fmt.Sprintf("bounded original decoder: %v %v", e, o.IP()))
		}
		matched := bytes.Equal(actual, expected) && exited.Code == 0 && int(o.Regs().SI) == 0xc000+entry.Stop
		mapNeed(os.WriteFile(filepath.Join(*output, entry.Columns), actual, 0644))
		receipts = append(receipts, map[string]any{"record": entry.Record, "name": entry.Name, "payload_sha256": entry.Hash, "fixture_sha256": mapHash(mapRead(fixture)), "columns_sha256": mapHash(actual), "columns": entry.Width, "cells": len(actual), "matched": matched, "exit_code": exited.Code, "steps": o.Steps(), "cycles": o.Cycles(), "final_si": o.Regs().SI})
		o.Close()
		if !matched {
			panic("map column mismatch: " + entry.Name)
		}
	}
	report := map[string]any{"schema": 1, "status": "PASS_EXTRACTED_ORIGINAL_CAVE_MAPS", "records": receipts, "go_version": runtime.Version(), "primary_sha256": mapHash(primary), "original_routine_sha256": mapHash(primary[0x6ced:0x6d6e]), "script_sha256": mapHash(mapRead(filepath.Join(*project, "tools/hd_cave_map_oracle.go"))), "address_space": "原版01B1:6CED..6D6D，隔離MZ保留段內位址；DS:SI讀C01B，ES:DI每欄8000、列距36", "limits": "隔離原常式與真實31份地圖，不是正常場景路徑、圖集選擇或HD完成證據"}
	data, e := json.MarshalIndent(report, "", "  ")
	mapNeed(e)
	mapNeed(os.WriteFile(filepath.Join(*output, "receipt.json"), append(data, 10), 0644))
	fmt.Println("PASS_EXTRACTED_ORIGINAL_CAVE_MAPS", len(receipts))
}

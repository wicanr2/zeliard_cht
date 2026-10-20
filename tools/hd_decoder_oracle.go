// 在 dosgolem 執行擷取的原版 RLE 常式，比對離線解碼器；不是正常遊戲路徑驗收。
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

func needHD(err error) {
	if err != nil {
		panic(err)
	}
}
func readHD(name string) []byte { b, e := os.ReadFile(name); needHD(e); return b }
func hashHD(b []byte) string    { return fmt.Sprintf("%x", sha256.Sum256(b)) }

func main() {
	project := flag.String("project", "", "唯讀專案")
	decoded := flag.String("decoded", "", "離線解碼輸出")
	output := flag.String("output", "", "既有空輸出目錄")
	flag.Parse()
	entries, err := os.ReadDir(*output)
	needHD(err)
	if len(entries) != 0 {
		panic("fresh output required")
	}
	resident := readHD(filepath.Join(*project, "workplace/probe-intro-memory/memory-01B1_0000.bin"))
	if hashHD(resident) != "c9f188ea847e8d364ccf8dbc9f6a9b9c74e192c0c83bd7a012b552e6d89b13d8" {
		panic("resident identity")
	}
	var manifest struct {
		Records []struct {
			Archive string `json:"archive"`
			Index   int    `json:"index"`
			Offset  int    `json:"payload_offset"`
			Size    int    `json:"bytes"`
			Path    string `json:"decoded_path"`
			Result  string `json:"result"`
		} `json:"records"`
	}
	needHD(json.Unmarshal(readHD(filepath.Join(*decoded, "decoded-inventory.json")), &manifest))
	var receipts []map[string]any
	for _, record := range manifest.Records {
		if record.Result != "DECODED_CANDIDATE" {
			panic("rejected offline input")
		}
		sar := readHD(filepath.Join(*project, "workplace/orig", record.Archive))
		stream := sar[record.Offset+1 : record.Offset+record.Size]
		image := make([]byte, 0x1000+len(stream))
		copy(image[0xdad:0xf17], resident[0xdad:0xf17])
		copy(image[0x1000:], stream)
		// 測試框架只設來源、目的地及長度，呼叫原始段內DAD；常式bytes不改。
		stub := []byte{0x0e, 0x1f, 0x8c, 0xc8, 0x05, 0x00, 0x10, 0x8e, 0xc0, 0xbe, 0x00, 0x10, 0xbf, 0, 0, 0xba, byte(len(stream)), byte(len(stream) >> 8), 0xe8, 0, 0, 0xb8, 0, 0x4c, 0xcd, 0x21}
		binary.LittleEndian.PutUint16(stub[19:21], 0xdad-(0x100+21))
		copy(image[0x100:], stub)
		header := make([]byte, 32)
		fileSize := len(header) + len(image)
		for i, value := range []uint16{0x5a4d, uint16(fileSize % 512), uint16((fileSize + 511) / 512), 0, 2, 0x2000, 0xffff, 0, 0xfffe, 0, 0x100, 0, 0x1c, 0} {
			binary.LittleEndian.PutUint16(header[i*2:], value)
		}
		fixture := filepath.Join(*output, fmt.Sprintf("%s-%02d.exe", record.Archive, record.Index))
		needHD(os.WriteFile(fixture, append(header, image...), 0644))
		o, e := oracle.Load(fixture, *output)
		needHD(e)
		cs := o.IP().Seg
		// 字典型逐項搜尋；12KiB輸入乘256項及每項指令，預算32M。
		err = o.Run(32_000_000)
		var exited *oracle.ExitError
		if !errors.As(err, &exited) {
			panic(fmt.Sprintf("bounded original decoder did not exit: %s #%d %v IP=%v regs=%+v", record.Archive, record.Index, err, o.IP(), o.Regs()))
		}
		expected := readHD(filepath.Join(*decoded, record.Path))
		actual := o.Bytes(oracle.Far(cs+0x1000, 0), len(expected))
		matched := bytes.Equal(actual, expected) && int(o.Regs().DI) == len(expected) && exited.Code == 0
		needHD(os.WriteFile(filepath.Join(*output, record.Path), actual, 0644))
		receipts = append(receipts, map[string]any{"archive": record.Archive, "index": record.Index, "source_payload_offset": record.Offset,
			"stream_sha256": hashHD(stream), "fixture_sha256": hashHD(readHD(fixture)), "original_routine_sha256": hashHD(resident[0xdad:0xf17]),
			"bytes": len(actual), "original_sha256": hashHD(actual), "offline_sha256": hashHD(expected), "di": o.Regs().DI, "dx": o.Regs().DX,
			"steps": o.Steps(), "cycles": o.Cycles(), "matched": matched, "exit_code": exited.Code})
		o.Close()
		if !matched {
			failure, e := json.MarshalIndent(receipts, "", "  ")
			needHD(e)
			needHD(os.WriteFile(filepath.Join(*output, "failure.json"), append(failure, 10), 0644))
			panic(fmt.Sprintf("RLE mismatch %s #%d", record.Archive, record.Index))
		}
	}
	result := map[string]any{"schema": 1, "status": "PASS_EXTRACTED_ORIGINAL_DECODER", "records": receipts,
		"go_version": runtime.Version(), "script_sha256": hashHD(readHD(filepath.Join(*project, "tools/hd_decoder_oracle.go"))),
		"dosgolem_source_manifest_sha256": hashHD(readHD(filepath.Join(*project, "workplace/dosgolem-source-manifest.json"))),
		"input_sha256":                    hashHD(resident), "address_space": "原常式IDA base0／原dosgolem 01B1:DAD..F16；測試MZ保留段內位址，載入段由dosgolem配置",
		"limits": []string{"隔離原版常式測試，不是冷啟動遊戲、正常玩家路徑或HD完成證據", "只驗84筆零前綴壓縮候選的解碼bytes，不推定圖像尺寸或用途"}}
	b, e := json.MarshalIndent(result, "", "  ")
	needHD(e)
	needHD(os.WriteFile(filepath.Join(*output, "receipt.json"), append(b, 10), 0644))
	fmt.Println("PASS_EXTRACTED_ORIGINAL_DECODER", len(receipts))
}

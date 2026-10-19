package sound

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestWAVOriginalSourceAndExclusiveCreation(t *testing.T) {
	b, _ := originalPCM(t)
	path := filepath.Join(t.TempDir(), "original.wav")
	w, err := CreateWAV(path)
	if err != nil {
		t.Fatal(err)
	}
	values := make([]int16, len(b)/2)
	for i := range values {
		values[i] = int16(binary.LittleEndian.Uint16(b[2*i:]))
	}
	for i := 0; i < len(values); i += 735 {
		end := i + 735
		if end > len(values) {
			end = len(values)
		}
		if err = w.Write(values[i:end]); err != nil {
			t.Fatal(err)
		}
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 44+len(b) || !bytes.Equal(out[44:], b) || string(out[:4]) != "RIFF" || binary.LittleEndian.Uint32(out[40:44]) != uint32(len(b)) || binary.LittleEndian.Uint16(out[22:24]) != 1 || binary.LittleEndian.Uint32(out[24:28]) != Rate {
		t.Fatal("WAV內容或檔頭不符")
	}
	if _, err = CreateWAV(path); !os.IsExist(err) {
		t.Fatal("允許覆寫已有錄音", err)
	}
	if err = w.Write([]int16{1}); err == nil {
		t.Fatal("已關閉錄音仍可寫入")
	}
	if s := w.State(); s["samples"] != uint64(511929) || s["closed"] != true {
		t.Fatal(s)
	}
}

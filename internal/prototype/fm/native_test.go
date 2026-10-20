package fm

import (
	"os"
	"testing"
)

func TestNativeCoreDeterministicAndRelease(t *testing.T) {
	path := os.Getenv("ZELIARD_FM_LIBRARY")
	if path == "" {
		t.Skip("需指定剛建置的動態音源核心")
	}
	a, e := Open(path, 44100)
	if e != nil {
		t.Fatal(e)
	}
	defer a.Close()
	b, e := Open(path, 44100)
	if e != nil {
		t.Fatal(e)
	}
	defer b.Close()
	for i := 0; i < 441; i++ {
		if a.Sample() != 0 || b.Sample() != 0 {
			t.Fatal("未起音便有訊號")
		}
	}
	// YM3812標準雙運算子樂音，核相同事件可重播、有效且不越界。
	regs := [][2]uint8{{1, 32}, {0x20, 1}, {0x23, 1}, {0x40, 0x10}, {0x43, 0}, {0x60, 0xf2}, {0x63, 0xf2}, {0x80, 0x77}, {0x83, 0x77}, {0xa0, 0x98}, {0xb0, 0x31}}
	for _, r := range regs {
		a.Write(r[0], r[1])
		b.Write(r[0], r[1])
	}
	nonzero := 0
	for i := 0; i < 4410; i++ {
		x, y := a.Sample(), b.Sample()
		if x != y || x < -1 || x > 1 {
			t.Fatal("音源不穩定", i, x, y)
		}
		if x != 0 {
			nonzero++
		}
	}
	if nonzero < 100 {
		t.Fatal("核心沒有實際出聲")
	}
	a.Close()
	a.Close()
}

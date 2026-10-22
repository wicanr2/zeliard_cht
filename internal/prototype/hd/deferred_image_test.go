package hd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func deferredFixture(t *testing.T) ([]byte, string) {
	t.Helper()
	im := image.NewNRGBA(image.Rect(0, 0, 6, 3))
	for y := 0; y < 3; y++ {
		for x := 0; x < 6; x++ {
			im.SetNRGBA(x, y, color.NRGBA{byte(x * 20), byte(y * 70), 42, 255})
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	return b.Bytes(), fmt.Sprintf("%x", sha256.Sum256(b.Bytes()))
}

func TestDeferredTerrainReadOnUseAndCache(t *testing.T) {
	b, hash := deferredFixture(t)
	path := filepath.Join(t.TempDir(), "later.png")
	lazy, err := deferredTerrainImage(path, hash, 6, 3, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = deferredTerrainImage(path, hash, 6, 3, false); err == nil {
		t.Fatal("eager未拒絕缺檔")
	}
	if err = os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	im := resolvedImage(lazy)
	if im == nil {
		t.Fatal("首次使用未載入")
	}
	for y := 0; y < 3; y++ {
		for x := 0; x < 6; x++ {
			want := color.NRGBA{byte(x * 20), byte(y * 70), 42, 255}
			if got := color.NRGBAModel.Convert(im.At(x, y)); got != want {
				t.Fatalf("pixel %d,%d: %v", x, y, got)
			}
		}
	}
	if err = os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if resolvedImage(lazy) != im {
		t.Fatal("已載圖未快取")
	}
}

func TestDeferredTerrainBadImageNeverPublishes(t *testing.T) {
	b, hash := deferredFixture(t)
	path := filepath.Join(t.TempDir(), "bad.png")
	if err := os.WriteFile(path, []byte("bad"), 0600); err != nil {
		t.Fatal(err)
	}
	lazy, _ := deferredTerrainImage(path, hash, 6, 3, true)
	if resolvedImage(lazy) != nil {
		t.Fatal("壞圖仍可見")
	}
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	if resolvedImage(lazy) != nil {
		t.Fatal("失敗圖反覆讀檔")
	}
	wrong, _ := deferredTerrainImage(path, hash, 9, 3, true)
	if resolvedImage(wrong) != nil {
		t.Fatal("錯尺寸仍可見")
	}
}

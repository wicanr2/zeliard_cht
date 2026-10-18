package hd

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"testing"
)

func TestRealTownFrameBoundariesAndFallback(t *testing.T) {
	ref, art := os.Getenv("ZELIARD_HD_REFERENCE"), os.Getenv("ZELIARD_HD_ART")
	if ref == "" || art == "" {
		t.Skip("需掛載真實城鎮收據與本輪生成圖，skip 不算驗收")
	}
	f, err := NewFrame(ref, filepath.Join(filepath.Dir(ref), "palette.bin"), art)
	if err != nil {
		t.Fatal(err)
	}
	indexed, rgb := append([]byte(nil), f.indices...), append([]byte(nil), f.rgb...)
	dst := image.NewRGBA(image.Rect(0, 0, 960, 680))
	for i := range dst.Pix {
		dst.Pix[i] = 123
	}
	before := append([]byte(nil), dst.Pix...)
	if !f.Draw(dst, indexed, rgb, image.Pt(0, 48)) {
		t.Fatal("真實城鎮邊框未辨識")
	}
	changed := 0
	for y := 0; y < 680; y++ {
		for x := 0; x < 960; x++ {
			i := y*dst.Stride + x*4
			if !bytes.Equal(before[i:i+4], dst.Pix[i:i+4]) {
				changed++
				if y < 48 || y >= 648 || x >= 144 && x < 816 {
					t.Fatalf("覆繪越界：%d,%d", x, y)
				}
			}
		}
	}
	if changed == 0 || !bytes.Equal(indexed, f.indices) || !bytes.Equal(rgb, f.rgb) {
		t.Fatal("沒有實際繪圖或改寫了原版輸入")
	}
	stable := append([]byte(nil), dst.Pix...)
	indexed[199*320+319] ^= 1
	if f.Draw(dst, indexed, rgb, image.Pt(0, 48)) || !bytes.Equal(stable, dst.Pix) {
		t.Fatal("右下角失配時沒有整組退回")
	}
	indexed[199*320+319] ^= 1
	rgb[3*100*320] ^= 1
	if f.Draw(dst, indexed, rgb, image.Pt(0, 48)) || !bytes.Equal(stable, dst.Pix) {
		t.Fatal("色盤／淡入失配仍覆繪")
	}
	if f.Match(indexed[:100], rgb) || f.Match(indexed, rgb[:100]) {
		t.Fatal("不完整畫格被接受")
	}
	if _, err := NewFrame(ref, filepath.Join(filepath.Dir(ref), "palette.bin"), filepath.Join(t.TempDir(), "missing.png")); err == nil {
		t.Fatal("缺圖沒有回報錯誤供前端退回原貌")
	}
}

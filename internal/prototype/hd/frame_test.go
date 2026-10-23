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
	indexed[199*320+319] = otherIndex(f.indices[199*320+319])
	if f.Draw(dst, indexed, rgb, image.Pt(0, 48)) || !bytes.Equal(stable, dst.Pix) {
		t.Fatal("右下角失配時沒有整組退回")
	}
	indexed[199*320+319] = f.indices[199*320+319]
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

// otherIndex 回傳既不是0也不是原值的色號，模擬非揭開造成的失配。
func otherIndex(v byte) byte {
	for _, c := range []byte{v ^ 1, v ^ 2, 200} {
		if c != 0 && c != v {
			return c
		}
	}
	return 201
}

// 規格067：原版逐格揭開兩側柱子時，只在已有像素的8×8格畫 HD，尚未畫到的格保持原狀；
// 眼睛暗色畫格視為揭開途中；其他色號一律整組退回。
func TestFrameRevealDrawsOnlyInkedCells(t *testing.T) {
	ref, art := os.Getenv("ZELIARD_HD_REFERENCE"), os.Getenv("ZELIARD_HD_ART")
	if ref == "" || art == "" {
		t.Skip("需掛載真實城鎮收據與本輪生成圖，skip 不算驗收")
	}
	f, err := NewFrame(ref, filepath.Join(filepath.Dir(ref), "palette.bin"), art)
	if err != nil {
		t.Fatal(err)
	}
	indexed, rgb := append([]byte(nil), f.indices...), append([]byte(nil), f.rgb...)
	// 左柱畫到第13列格（y<104），右柱只畫了第一欄格的一半（x 272..275），其餘尚未畫。
	for y := 0; y < 200; y++ {
		for x := 0; x < 320; x++ {
			left := x < 48 && y < 104
			right := x >= 272 && x < 276
			if !left && !right && (x < 48 || x >= 272) {
				indexed[y*320+x] = 0
			}
		}
	}
	for at, v := range frameEyes {
		if at%320 < 48 && at/320 < 104 {
			indexed[at] = v
			copy(rgb[at*3:at*3+3], f.palette[int(v)*3:int(v)*3+3])
		}
	}
	full := image.NewRGBA(image.Rect(0, 0, 960, 680))
	if !f.Draw(full, f.indices, f.rgb, image.Pt(0, 48)) {
		t.Fatal("完整邊框未辨識")
	}
	dst := image.NewRGBA(image.Rect(0, 0, 960, 680))
	for i := range dst.Pix {
		dst.Pix[i] = 123
	}
	if !f.Draw(dst, indexed, rgb, image.Pt(0, 48)) {
		t.Fatal("逐格揭開途中未畫 HD")
	}
	at := func(img *image.RGBA, x, y int) []byte {
		i := img.PixOffset(x, y)
		return img.Pix[i : i+4]
	}
	for _, c := range []struct {
		x, y int
		hd   bool
	}{{10, 48 + 10, true}, {140, 48 + 300, true}, {140, 48 + 320, false}, {820, 48 + 10, true}, {820, 48 + 590, true}, {850, 48 + 10, false}, {500, 48 + 300, false}} {
		got := at(dst, c.x, c.y)
		if c.hd && !bytes.Equal(got, at(full, c.x, c.y)) || !c.hd && !bytes.Equal(got, []byte{123, 123, 123, 123}) {
			t.Fatalf("%d,%d 揭開覆繪不符：%v", c.x, c.y, got)
		}
	}
	indexed[150*320+300] = otherIndex(f.indices[150*320+300])
	if f.Draw(dst, indexed, rgb, image.Pt(0, 48)) {
		t.Fatal("揭開期間出現其他色號仍覆繪")
	}
	if f.Draw(dst, make([]byte, 64000), make([]byte, 192000), image.Pt(0, 48)) {
		t.Fatal("全黑畫面不該當成揭開")
	}
}

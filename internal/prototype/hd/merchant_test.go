package hd

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"testing"
)

func TestMerchantPortraitPosesAndFallback(t *testing.T) {
	a, b, palette := os.Getenv("ZELIARD_MERCHANT_A"), os.Getenv("ZELIARD_MERCHANT_B"), os.Getenv("ZELIARD_MERCHANT_PALETTE")
	artA, artB := os.Getenv("ZELIARD_MERCHANT_ART_A"), os.Getenv("ZELIARD_MERCHANT_ART_B")
	c, artC := os.Getenv("ZELIARD_MERCHANT_C"), os.Getenv("ZELIARD_MERCHANT_ART_C")
	if a == "" || b == "" || c == "" || palette == "" || artA == "" || artB == "" || artC == "" {
		t.Skip("須掛載武器商人原版三畫格、色盤與三張手繪圖")
	}
	m, err := NewMerchant(a, b, c, palette, artA, artB, artC)
	if err != nil {
		t.Fatal(err)
	}
	pal, err := os.ReadFile(palette)
	if err != nil {
		t.Fatal(err)
	}
	for pose, path := range []string{a, b, c} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		rgb := make([]byte, len(raw)*3)
		for i, v := range raw {
			copy(rgb[i*3:i*3+3], pal[int(v)*3:int(v)*3+3])
		}
		beforeRaw, beforeRGB := append([]byte(nil), raw...), append([]byte(nil), rgb...)
		dst := image.NewRGBA(image.Rect(0, 0, 960, 680))
		for y := 48; y < 648; y++ {
			for x := 0; x < 960; x++ {
				i := ((y-48)/3*320 + x/3) * 3
				dst.SetRGBA(x, y, color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255})
			}
		}
		original := append([]byte(nil), dst.Pix...)
		matched, got := m.Draw(dst, raw, rgb, image.Pt(0, 48))
		if !matched || got != pose {
			t.Fatal("未匹配武器商人原版姿態", pose, got)
		}
		inner := merchantBox.Inset(1)
		box := image.Rect(inner.Min.X*scale, inner.Min.Y*scale+48, inner.Max.X*scale, inner.Max.Y*scale+48)
		changed := 0
		for y := 0; y < 680; y++ {
			for x := 0; x < 960; x++ {
				i := dst.PixOffset(x, y)
				if !bytes.Equal(original[i:i+4], dst.Pix[i:i+4]) {
					if !image.Pt(x, y).In(box) {
						t.Fatal("商人HD改寫內框以外像素", x, y)
					}
					changed++
				}
			}
		}
		if changed < 50000 || !bytes.Equal(raw, beforeRaw) || !bytes.Equal(rgb, beforeRGB) {
			t.Fatal("商人HD沒有完整顯示或改寫原版輸入", changed)
		}
		after := append([]byte(nil), dst.Pix...)
		raw[23*320+56] ^= 1
		if yes, _ := m.Draw(dst, raw, rgb, image.Pt(0, 48)); yes || !bytes.Equal(after, dst.Pix) {
			t.Fatal("未知商人畫格沒有原子退回")
		}
		raw[23*320+56] ^= 1
		rgb[(24*320+57)*3] ^= 1
		if yes, _ := m.Draw(dst, raw, rgb, image.Pt(0, 48)); yes || !bytes.Equal(after, dst.Pix) {
			t.Fatal("色盤失配沒有原子退回")
		}
		if yes, _ := m.Draw(dst, raw[:10], rgb, image.Pt(0, 48)); yes || !bytes.Equal(after, dst.Pix) {
			t.Fatal("短畫格被接受")
		}
	}
	if bytes.Equal(m.art[0].Pix, m.art[1].Pix) || bytes.Equal(m.art[0].Pix, m.art[2].Pix) || bytes.Equal(m.art[1].Pix, m.art[2].Pix) {
		t.Fatal("商人三個HD姿態沒有差異")
	}
	if _, err := NewMerchant(a, b, c, palette, artA, artB, artC+".missing"); err == nil {
		t.Fatal("缺一張圖時未拒絕整組")
	}
	if _, err := NewMerchant(a, a, c, palette, artA, artB, artC); err == nil {
		t.Fatal("錯誤第二原版姿態未拒絕")
	}
	if _, err := NewMerchant(a, b, a, palette, artA, artB, artC); err == nil {
		t.Fatal("錯誤第三原版姿態未拒絕")
	}
}

package hd

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"testing"
)

func TestSagePortraitPreservesOriginalAnimation(t *testing.T) {
	a, b, palette, art := os.Getenv("ZELIARD_SAGE_A"), os.Getenv("ZELIARD_SAGE_B"), os.Getenv("ZELIARD_SAGE_PALETTE"), os.Getenv("ZELIARD_SAGE_ART")
	if a == "" || b == "" || palette == "" || art == "" {
		t.Skip("須掛載長老原版兩畫格、色盤及本地手繪圖")
	}
	s, err := NewSage(a, b, palette, art)
	if err != nil {
		t.Fatal(err)
	}
	pal, err := os.ReadFile(palette)
	if err != nil {
		t.Fatal(err)
	}
	animated := 0
	for _, on := range s.animated {
		if on {
			animated++
		}
	}
	if animated != 4 {
		t.Fatal("原版動畫像素數量不符", animated)
	}
	for pose, path := range []string{a, b} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		beforeRaw := append([]byte(nil), raw...)
		rgb := make([]byte, len(raw)*3)
		for i, v := range raw {
			copy(rgb[i*3:i*3+3], pal[int(v)*3:int(v)*3+3])
		}
		beforeRGB := append([]byte(nil), rgb...)
		dst := image.NewRGBA(image.Rect(0, 0, 960, 680))
		for y := 0; y < 600; y++ {
			for x := 0; x < 960; x++ {
				i := (y/3*320 + x/3) * 3
				dst.SetRGBA(x, y+48, color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255})
			}
		}
		original := append([]byte(nil), dst.Pix...)
		matched, got := s.Draw(dst, raw, rgb, image.Pt(0, 48))
		if !matched || got != pose {
			t.Fatal("原版長老畫格未匹配", pose, got)
		}
		changed := 0
		for y := 0; y < 680; y++ {
			for x := 0; x < 960; x++ {
				i := dst.PixOffset(x, y)
				if bytes.Equal(original[i:i+4], dst.Pix[i:i+4]) {
					continue
				}
				p := image.Pt(x/3, (y-48)/3)
				if !p.In(sageBox.Inset(1)) {
					t.Fatal("改寫長老內框以外像素", x, y)
				}
				local := p.Sub(sageBox.Min)
				if s.animated[local.Y*sageBox.Dx()+local.X] {
					t.Fatal("改寫原版動畫像素", p)
				}
				changed++
			}
		}
		if changed < 50000 {
			t.Fatal("長老高清圖面沒有完整顯示", changed)
		}
		if !bytes.Equal(raw, beforeRaw) || !bytes.Equal(rgb, beforeRGB) {
			t.Fatal("改寫原版畫格或RGB")
		}
		copyAfter := append([]byte(nil), dst.Pix...)
		raw[23*320+56] ^= 1
		if yes, _ := s.Draw(dst, raw, rgb, image.Point{}); yes || !bytes.Equal(copyAfter, dst.Pix) {
			t.Fatal("未知畫格未完整退回")
		}
		raw[23*320+56] ^= 1
		rgb[(24*320+57)*3] ^= 1
		if yes, _ := s.Draw(dst, raw, rgb, image.Point{}); yes || !bytes.Equal(copyAfter, dst.Pix) {
			t.Fatal("色盤失配未完整退回")
		}
		if yes, _ := s.Draw(dst, raw[:10], rgb, image.Point{}); yes {
			t.Fatal("短畫格被接受")
		}
	}
	if _, err := NewSage(a, b, palette, art+".missing"); err == nil {
		t.Fatal("缺圖未拒絕整組素材")
	}
	if _, err := NewSage(a, a, palette, art); err == nil {
		t.Fatal("錯誤第二畫格未拒絕")
	}
}

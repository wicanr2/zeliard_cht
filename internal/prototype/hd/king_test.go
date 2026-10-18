package hd

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestKingPortraitPosesAndAtomicFallback(t *testing.T) {
	a, b, closed, open := os.Getenv("ZELIARD_KING_A"), os.Getenv("ZELIARD_KING_B"), os.Getenv("ZELIARD_KING_CLOSED"), os.Getenv("ZELIARD_KING_OPEN")
	if a == "" || b == "" || closed == "" || open == "" {
		t.Skip("須掛載國王原版兩姿態及兩張本地手繪圖")
	}
	palette := filepath.Join(filepath.Dir(a), "palette.bin")
	k, err := NewKing(a, b, palette, closed, open)
	if err != nil {
		t.Fatal(err)
	}
	pal, err := os.ReadFile(palette)
	if err != nil {
		t.Fatal(err)
	}
	for pose, path := range []string{a, b} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before := append([]byte(nil), raw...)
		rgb := make([]byte, len(raw)*3)
		for i, v := range raw {
			copy(rgb[3*i:3*i+3], pal[int(v)*3:int(v)*3+3])
		}
		dst := image.NewRGBA(image.Rect(0, 0, 960, 680))
		for y := 0; y < 680; y++ {
			for x := 0; x < 960; x++ {
				dst.SetRGBA(x, y, color.RGBA{3, 4, 5, 255})
			}
		}
		matched, got := k.Draw(dst, raw, rgb, image.Pt(0, 48))
		if !matched || got != pose {
			t.Fatal("未匹配原版姿態", pose, got)
		}
		if !bytes.Equal(raw, before) {
			t.Fatal("改寫原版畫格")
		}
		box := image.Rect(339, 120, 621, 306)
		changed := 0
		for y := 0; y < 680; y++ {
			for x := 0; x < 960; x++ {
				if dst.RGBAAt(x, y) != (color.RGBA{3, 4, 5, 255}) {
					if !image.Pt(x, y).In(box) {
						t.Fatal("覆繪超出肖像內框", x, y)
					}
					changed++
				}
			}
		}
		if changed < 50000 {
			t.Fatal("HD 圖面未完整合成", changed)
		}
		raw[23*320+112] ^= 1
		if yes, _ := k.Draw(dst, raw, rgb, image.Pt(0, 48)); yes {
			t.Fatal("未知姿態未退回")
		}
		raw[23*320+112] ^= 1
		rgb[(24*320+113)*3] ^= 1
		if yes, _ := k.Draw(dst, raw, rgb, image.Pt(0, 48)); yes {
			t.Fatal("淡入色盤未退回")
		}
		if yes, _ := k.Draw(dst, raw[:10], rgb, image.Point{}); yes {
			t.Fatal("不完整畫格未退回")
		}
	}
	diff := 0
	for y := 0; y < 192; y++ {
		for x := 0; x < 288; x++ {
			if k.art[0].RGBAAt(x, y) != k.art[1].RGBAAt(x, y) {
				if !image.Pt(x, y).In(image.Rect(104, 86, 184, 125)) {
					t.Fatal("姿態變化改動口部以外的美術")
				}
				diff++
			}
		}
	}
	if diff == 0 {
		t.Fatal("兩個 HD 姿態完全相同")
	}
	if _, err := NewKing(a, b, palette, closed, filepath.Join(t.TempDir(), "missing.png")); err == nil {
		t.Fatal("少一張姿態時必須拒絕整組")
	}
}

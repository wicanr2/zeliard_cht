package overlay

import (
	"bytes"
	"os"
	"testing"

	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

func TestOriginalIntroReadOnlyAndRemoval(t *testing.T) {
	ref, fontPath, textPath := os.Getenv("ZELIARD_REFERENCE"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_INTRO_TEXT")
	if ref == "" || fontPath == "" || textPath == "" {
		t.Skip("須掛載原版片頭收據、使用者字型與譯文；未驗原版覆繪")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	layer, err := NewIntro(ref, textPath, font)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(ref)
	if err != nil {
		t.Fatal(err)
	}
	before := append([]byte(nil), raw...)
	rgb := make([]byte, len(raw)*3)
	for i, v := range raw {
		rgb[i*3], rgb[i*3+1], rgb[i*3+2] = v, v, v
	}
	layer.Frame(raw, rgb)
	dst := make([]byte, 960*600*4)
	if !layer.Draw(dst, 3, nil) || len(layer.Stamps) != 1 {
		t.Fatal("原版匹配時必須覆繪")
	}
	if !bytes.Equal(raw, before) {
		t.Fatal("覆繪改動原版畫格")
	}
	for i := 0; i < len(dst); i += 4 {
		if dst[i+3] != 0 {
			x, y := (i/4)%960, (i/4)/960
			if x < 88*3 || x >= 248*3 || y < 136*3 || y >= 152*3 {
				t.Fatalf("覆繪越界 %d,%d", x, y)
			}
		}
	}
	blank := make([]byte, len(raw))
	blankRGB := make([]byte, len(rgb))
	for i := 0; i < 3; i++ {
		layer.Frame(blank, blankRGB)
	}
	clear(dst)
	if layer.Draw(dst, 3, nil) || len(layer.Stamps) != 0 {
		t.Fatal("原文消失後不可留字")
	}
	negative, err := NewIntro(ref, textPath, font)
	if err != nil {
		t.Fatal(err)
	}
	negative.Frame(blank, blankRGB)
	if len(negative.Stamps) != 0 {
		t.Fatal("不相符畫面不可誤蓋")
	}
}

package overlay

import (
	"image"
	"testing"
)

// 規格067：場景名比原文長時，向兩側整欄都是欄底色5的範圍延伸，兩端各留一欄，不跨進其他欄位。
func TestHUDExtend(t *testing.T) {
	raw := make([]byte, 64000)
	for y := 175; y < 183; y++ {
		for x := 53; x < 187; x++ {
			raw[y*320+x] = 5
		}
		for x := 100; x < 161; x += 2 {
			raw[y*320+x] = 9 // 原版字形
		}
	}
	box := image.Rect(100, 175, 161, 183)
	got, fits := hudExtend(raw, box, 300, nil)
	if !fits || got != image.Rect(54, 175, 186, 183) {
		t.Fatal("延伸範圍不符", got, fits)
	}
	got, fits = hudExtend(raw, box, 300, []image.Rectangle{image.Rect(53, 175, 80, 183)})
	if !fits || got.Min.X != 81 {
		t.Fatal("跨進已畫的地點標籤", got, fits)
	}
	if _, fits = hudExtend(raw, box, 500, nil); fits {
		t.Fatal("超過欄寬仍回報放得下")
	}
	raw[178*320+170] = 1
	if got, _ = hudExtend(raw, box, 300, nil); got.Max.X != 169 {
		t.Fatal("右側非底色欄未停止", got)
	}
}

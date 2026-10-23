package main

import (
	"bytes"
	"image"
	"image/color"
)

// restoreNativeBox在HD背景之後放回原版對話框，中文層仍在其後繪製。
func restoreNativeBox(dst *image.RGBA, rgb []byte, box image.Rectangle) {
	if len(rgb) != 320*200*3 || box.Empty() || !box.In(image.Rect(0, 0, 320, 200)) {
		return
	}
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			i := (y*320 + x) * 3
			c := color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255}
			for dy := 0; dy < 3; dy++ {
				for dx := 0; dx < 3; dx++ {
					dst.SetRGBA(x*3+dx, y*3+dy+barHeight, c)
				}
			}
		}
	}
}

// restoreChangedRows 只把對話框內已和完成背景不同的列放回原版。居民對話來源辨識後、原版
// 還沒畫框時，框內仍是已驗城鎮，整框放回原版會讓HD與原版拼在一起（規格066）。
func restoreChangedRows(dst *image.RGBA, raw, base, rgb []byte, box image.Rectangle) {
	if len(raw) != 64000 || len(base) != 64000 || box.Empty() || !box.In(image.Rect(0, 0, 320, 200)) {
		restoreNativeBox(dst, rgb, box)
		return
	}
	for y := box.Min.Y; y < box.Max.Y; y++ {
		a, b := y*320+box.Min.X, y*320+box.Max.X
		if !bytes.Equal(raw[a:b], base[a:b]) {
			restoreNativeBox(dst, rgb, image.Rect(box.Min.X, y, box.Max.X, y+1))
		}
	}
}

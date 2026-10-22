package main

import (
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

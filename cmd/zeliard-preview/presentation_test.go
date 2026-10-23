package main

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

// 居民對話框尚未畫出時保留HD；原版畫到哪一列，那一列才放回原版。
func TestRestoreChangedRows(t *testing.T) {
	base := make([]byte, 64000)
	raw := append([]byte(nil), base...)
	rgb := make([]byte, 64000*3)
	for i := range rgb {
		rgb[i] = 50
	}
	box := image.Rect(56, 40, 232, 66)
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	hd := color.RGBA{200, 0, 0, 255}
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{hd}, image.Point{}, draw.Src)
	restoreChangedRows(canvas, raw, base, rgb, box)
	if blockColor(canvas, 60, 45) != hd {
		t.Fatal("原版尚未畫框時框內不該放回原版")
	}
	raw[41*320+56] = 7
	restoreChangedRows(canvas, raw, base, rgb, box)
	if c := blockColor(canvas, 200, 41); c.R != 50 {
		t.Fatal("已改變的列未放回原版", c)
	}
	if blockColor(canvas, 60, 45) != hd {
		t.Fatal("未改變的列被放回原版")
	}
}

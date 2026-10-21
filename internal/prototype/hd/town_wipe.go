package hd

import (
	"bytes"
	"image"
	"image/color"
)

// 原版會在兩次Draw之間完成城鎮並開始擦除。基準必須取完成回呼，不能取舊Draw。
func nativeTownWipe(in townFrameInput, source *TownPresentation) (*TownPresentation, int) {
	if source == nil || in.mode != 0x13 || len(in.indexed) != 64000 || len(in.rgb) != 192000 || len(source.SourceIndexed) != 64000 || len(source.RGB) != 192000 || len(in.palette) != 768 || !bytes.Equal(in.palette, source.Palette) || in.steps < source.Steps || in.cycles < source.Cycles {
		return nil, 0
	}
	mask := make([]byte, 224*144)
	erased := 0
	for y := 14; y < 158; y++ {
		for x := 48; x < 272; x++ {
			i := y*320 + x
			if in.indexed[i] == source.SourceIndexed[i] && bytes.Equal(in.rgb[i*3:i*3+3], source.RGB[i*3:i*3+3]) {
				if in.indexed[i] == 0 {
					mask[(y-14)*224+x-48] = 1
				}
				continue
			}
			if in.indexed[i] != 0 || in.rgb[i*3] != 0 || in.rgb[i*3+1] != 0 || in.rgb[i*3+2] != 0 {
				return nil, 0
			}
			mask[(y-14)*224+x-48] = 1
			erased++
		}
	}
	if erased == 0 {
		return nil, 0
	}
	result := *source
	result.Indexed = cloneTownBytes(in.indexed)
	result.RGB = cloneTownBytes(in.rgb)
	result.Wiping = true
	result.EraseMask = mask
	for y := 14; y < 158; y++ {
		a, b := y*320+48, y*320+272
		copy(result.Indexed[a:b], source.SourceIndexed[a:b])
		copy(result.RGB[a*3:b*3], source.RGB[a*3:b*3])
	}
	return &result, erased
}

func (f *TownPresentation) ApplyErase(dst *image.RGBA, origin image.Point) bool {
	if f == nil || !f.Wiping || len(f.EraseMask) != 224*144 || dst == nil || !image.Rect(144, 42, 816, 474).Add(origin).In(dst.Bounds()) {
		return false
	}
	for y := 0; y < 144; y++ {
		for x := 0; x < 224; x++ {
			if f.EraseMask[y*224+x] == 0 {
				continue
			}
			for dy := 0; dy < 3; dy++ {
				for dx := 0; dx < 3; dx++ {
					dst.SetRGBA((x+48)*3+dx+origin.X, (y+14)*3+dy+origin.Y, color.RGBA{A: 255})
				}
			}
		}
	}
	return true
}

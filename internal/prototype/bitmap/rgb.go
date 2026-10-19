package bitmap

import "image"

// DrawRGB3將完整320×200 RGB最近鄰複製到既有三倍雛形畫帶。
// 先寫一列再複製兩列，避開逐點SetRGBA；輸入失配時不寫目標。
func DrawRGB3(dst *image.RGBA, rgb []byte, at image.Point) bool {
	if dst == nil || len(rgb) != 320*200*3 {
		return false
	}
	box := image.Rect(at.X, at.Y, at.X+960, at.Y+600)
	if !box.In(dst.Rect) {
		return false
	}
	for y := 0; y < 200; y++ {
		start := dst.PixOffset(at.X, at.Y+y*3)
		row := dst.Pix[start : start+960*4]
		src := rgb[y*320*3 : (y+1)*320*3]
		for x := 0; x < 320; x++ {
			s, d := x*3, x*12
			for k := 0; k < 12; k += 4 {
				row[d+k], row[d+k+1], row[d+k+2], row[d+k+3] = src[s], src[s+1], src[s+2], 255
			}
		}
		copy(dst.Pix[start+dst.Stride:start+dst.Stride+960*4], row)
		copy(dst.Pix[start+2*dst.Stride:start+2*dst.Stride+960*4], row)
	}
	return true
}

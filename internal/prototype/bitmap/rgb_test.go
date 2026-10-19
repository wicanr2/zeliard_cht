package bitmap

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

func rgbPattern() []byte {
	b := make([]byte, 320*200*3)
	for i := range b {
		b[i] = byte((i*173 + i/23) % 256)
	}
	return b
}

// 保留前端原迴圈作參考，完整比較底圖與區外。
func drawRGBPrevious(dst *image.RGBA, rgb []byte, at image.Point) {
	for y := 0; y < 200; y++ {
		for x := 0; x < 320; x++ {
			i := (y*320 + x) * 3
			c := color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255}
			for yy := 0; yy < 3; yy++ {
				for xx := 0; xx < 3; xx++ {
					dst.SetRGBA(at.X+x*3+xx, at.Y+y*3+yy, c)
				}
			}
		}
	}
}

func TestDrawRGB3WholeCanvasPreviousPixels(t *testing.T) {
	for _, tc := range []struct {
		name      string
		root, sub image.Rectangle
		at        image.Point
	}{
		{"frontend", image.Rect(0, 0, 960, 680), image.Rect(0, 0, 960, 680), image.Pt(0, 48)},
		{"subimage_stride", image.Rect(-40, -60, 1040, 750), image.Rect(-15, -20, 985, 710), image.Pt(-4, 17)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, b := image.NewRGBA(tc.root), image.NewRGBA(tc.root)
			for i := range a.Pix {
				a.Pix[i] = byte(i*13 + 7)
			}
			copy(b.Pix, a.Pix)
			input := rgbPattern()
			before := append([]byte(nil), input...)
			drawRGBPrevious(a.SubImage(tc.sub).(*image.RGBA), input, tc.at)
			if !DrawRGB3(b.SubImage(tc.sub).(*image.RGBA), input, tc.at) {
				t.Fatal("有效來源被拒絕")
			}
			if !bytes.Equal(a.Pix, b.Pix) {
				t.Fatal("整幅含區外像素不符")
			}
			if !bytes.Equal(before, input) {
				t.Fatal("改寫RGB來源")
			}
		})
	}
}

func TestDrawRGB3RejectWithoutWrites(t *testing.T) {
	b := image.NewRGBA(image.Rect(0, 0, 960, 680))
	for i := range b.Pix {
		b.Pix[i] = byte(i*17 + 3)
	}
	before := append([]byte(nil), b.Pix...)
	input := rgbPattern()
	for _, tc := range []struct {
		rgb []byte
		at  image.Point
	}{
		{nil, image.Pt(0, 48)}, {input[:len(input)-1], image.Pt(0, 48)},
		{append(input, 0), image.Pt(0, 48)}, {input, image.Pt(1, 48)},
		{input, image.Pt(-1, 48)}, {input, image.Pt(0, 81)},
	} {
		if DrawRGB3(b, tc.rgb, tc.at) || !bytes.Equal(before, b.Pix) {
			t.Fatal("失配來源寫入目標")
		}
	}
	if DrawRGB3(nil, input, image.Point{}) {
		t.Fatal("空目標通過")
	}
}

func BenchmarkRGB3Previous(b *testing.B) {
	dst, input := image.NewRGBA(image.Rect(0, 0, 960, 680)), rgbPattern()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		drawRGBPrevious(dst, input, image.Pt(0, 48))
	}
}
func BenchmarkRGB3Rows(b *testing.B) {
	dst, input := image.NewRGBA(image.Rect(0, 0, 960, 680)), rgbPattern()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawRGB3(dst, input, image.Pt(0, 48))
	}
}

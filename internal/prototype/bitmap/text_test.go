package bitmap

import (
	"image"
	"image/color"
	"reflect"
	"testing"
)

func TestSystemTextWrapAndClip(t *testing.T) {
	f := Font{'字': make([]byte, 32), 'A': make([]byte, 16), '?': make([]byte, 16), '…': make([]byte, 32)}
	for _, size := range []int{16, 24, 32} {
		if got := f.TextLines("字A字\nA字", size, size*2); !reflect.DeepEqual(got, []string{"字A", "字", "A字"}) {
			t.Fatalf("size %d: %v", size, got)
		}
		dst := image.NewRGBA(image.Rect(0, 0, 300, 200))
		if !f.DrawText(dst, image.Rect(8, 8, 8+size*2, 8+size*3+8), "字A字\nA字", size, color.RGBA{255, 255, 255, 255}) {
			t.Fatalf("fitting text rejected at %d", size)
		}
		if f.DrawText(dst, image.Rect(8, 8, 8+size*2, 8+size), "字A字", size, color.RGBA{255, 255, 255, 255}) {
			t.Fatalf("overflow not reported at %d", size)
		}
	}
}

func TestSystemTextUsesTargetSizeMask(t *testing.T) {
	f := Font{'字': make([]byte, 32), '?': make([]byte, 16)}
	mask := make([]byte, 16*16+18*18+24*24+32*32)
	// 只在32像素遮罩放一個像素，16像素來源全空。
	mask[16*16+18*18+24*24+7*32+11] = 255
	f[-'字'-1] = mask
	dst := image.NewRGBA(image.Rect(0, 0, 64, 64))
	box := image.Rect(5, 6, 37, 38)
	ink := color.RGBA{255, 255, 255, 255}
	if !f.DrawText(dst, box, "字", 32, ink) || dst.RGBAAt(16, 13) != ink {
		t.Fatal("target-resolution mask was not used")
	}
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			if (x != 16 || y != 13) && dst.RGBAAt(x, y).A != 0 {
				t.Fatalf("unexpected ink at %d,%d", x, y)
			}
		}
	}
}

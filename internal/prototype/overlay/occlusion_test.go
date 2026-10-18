package overlay

import (
	"bytes"
	"image"
	"testing"
)

func TestModalOwnershipRestorationAndUnknownChanges(t *testing.T) {
	raw := make([]byte, 64000)
	row := image.Rect(80, 110, 230, 118)
	for x := row.Min.X; x < row.Max.X; x++ {
		raw[row.Min.Y*320+x] = 9
	}
	expected := labelCrop(raw, row)
	baseline := append([]byte(nil), raw...)
	h := &paintHistory{}
	modal := image.Rect(188, 43, 236, 68)
	h.record(modal, raw)
	raw[44*320+189] = 9
	// 畫框尚在，即使某一文字列完整相同，也不能撤掉遮罩。
	masks, known := h.masks(0, raw)
	if !known || len(masks) != 1 {
		t.Fatal("未保留仍開啟的畫框")
	}
	modal2 := image.Rect(180, 110, 240, 140)
	h.record(modal2, raw)
	for y := modal2.Min.Y; y < modal2.Max.Y; y++ {
		for x := modal2.Min.X; x < modal2.Max.X; x++ {
			raw[y*320+x] = 7
		}
	}
	masks, known = h.masks(0, raw)
	if !known || len(masks) != 2 || !visibleRowsEqual(expected, raw, []image.Rectangle{row}, masks) {
		t.Fatal("已知遮擋造成可見文字退回")
	}
	raw[110*320+100] = 1
	if visibleRowsEqual(expected, raw, []image.Rectangle{row}, masks) {
		t.Fatal("未拒絕遮擋範圍外的未知改變")
	}
	copy(raw, baseline)
	masks, known = h.masks(0, raw)
	if !known || len(masks) != 0 || !visibleRowsEqual(expected, raw, []image.Rectangle{row}, masks) {
		t.Fatal("原版完整還原後未解除遮擋")
	}
	if !bytes.Equal(expected, labelCrop(raw, row)) {
		t.Fatal("觀察器改寫底圖")
	}
	h.invalidate()
	if _, known = h.masks(0, raw); known {
		t.Fatal("未知畫框程式仍被當成已知遮擋")
	}
	for i := 0; i < 65; i++ {
		h.record(modal, raw)
	}
	if _, known = h.masks(1, raw); known || len(h.events) != 64 {
		t.Fatal("已淘汰歷史未安全退回")
	}
}

func assertLabelOcclusion(t *testing.T, l *Labels, raw []byte, pal [256][3]uint8, before []byte) {
	t.Helper()
	for _, stamp := range l.stamps {
		masks, known := l.paint.masks(stamp.paintSerial, raw)
		if !known {
			t.Fatal("留下所需歷史已失效的標籤")
		}
		isolated := *l
		isolated.stamps = []labelStamp{stamp}
		canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
		copy(canvas.Pix, before)
		isolated.Draw(canvas, raw, pal, image.Point{})
		for _, mask := range masks {
			for y := mask.Min.Y * 3; y < mask.Max.Y*3; y++ {
				for x := mask.Min.X * 3; x < mask.Max.X*3; x++ {
					i := canvas.PixOffset(x, y)
					if !bytes.Equal(before[i:i+4], canvas.Pix[i:i+4]) {
						t.Fatal("舊中文覆蓋原版彈窗")
					}
				}
			}
		}
		// 真正進入 Draw 的未知變化也必須撤除，不能只測比對 helper。
		checked := false
		for _, row := range stamp.rows {
			for y := row.Min.Y; y < row.Max.Y && !checked; y++ {
				for x := row.Min.X; x < row.Max.X; x++ {
					if occluded(image.Pt(x, y), masks) {
						continue
					}
					bad := append([]byte(nil), raw...)
					bad[y*320+x] ^= 1
					isolated.stamps = []labelStamp{stamp}
					copy(canvas.Pix, before)
					if isolated.Draw(canvas, bad, pal, image.Point{}) != 0 || !bytes.Equal(canvas.Pix, before) {
						t.Fatal("未知的可見像素變化仍覆繪中文")
					}
					checked = true
					break
				}
			}
		}
	}
}

package overlay

import (
	"crypto/sha256"
	"fmt"
	"image"

	"github.com/wicanr2/dosgolem/oracle"
)

type paintEvent struct {
	serial uint64
	box    image.Rectangle
	before []byte
}

// paintHistory 只觀察固定 MCGA 的清框入口。新畫框擁有整個矩形，
// 較舊的中文只能畫在矩形外；未知的畫面改變仍會撤除譯文。
type paintHistory struct {
	serial uint64
	floor  uint64
	events []paintEvent
}

func observePaint(o *oracle.Oracle) *paintHistory {
	h := &paintHistory{}
	o.OnCall(oracle.Far(0x01b1, 0x2046), func(o *oracle.Oracle) {
		// IDA binary base 0，對應正常武器店 01B1:2046..2105。
		// 完整函式鏈包含內部填色及邊框，不只比對入口幾個 bytes。
		if o.VideoMode() != 0x13 || fmt.Sprintf("%x", sha256.Sum256(o.Bytes(oracle.Far(0x01b1, 0x2046), 0xc0))) != "3e91f3bc6ad5a812f3b4c89c49aa3c8e146f8cef714fe0cade0009e4303f72c4" {
			h.invalidate()
			return
		}
		x, y := int(o.BX()>>8)*4, int(o.BX()&255)
		w, height := int(o.CX()>>8)*4, int(o.CX()&255)
		box := image.Rect(x, y, x+w, y+height)
		if box.Empty() || !box.In(image.Rect(0, 0, 320, 200)) || byte(o.AX()) != 0 && (w < 4 || height <= 4) {
			h.invalidate()
			return
		}
		h.record(box, o.Indexed())
	})
	return h
}

func (h *paintHistory) invalidate() {
	h.serial++
	h.floor = h.serial
	h.events = nil
}

func (h *paintHistory) record(box image.Rectangle, raw []byte) {
	h.serial++
	if len(h.events) == 64 {
		h.floor = h.events[0].serial
		h.events = h.events[1:]
	}
	h.events = append(h.events, paintEvent{serial: h.serial, box: box, before: labelCrop(raw, box)})
}

// masks 的版本來自文字完成時。歷史被淘汰時不能推測遮擋範圍。
func (h *paintHistory) masks(since uint64, raw []byte) ([]image.Rectangle, bool) {
	if since < h.floor {
		return nil, false
	}
	var boxes []image.Rectangle
	for _, e := range h.events {
		// 必須整個原版畫框範圍還原，不能只因某一列文字相同就撤遮罩。
		if e.serial > since && !visibleRowsEqual(e.before, raw, []image.Rectangle{e.box}, nil) {
			boxes = append(boxes, e.box)
		}
	}
	return boxes, true
}

func occluded(p image.Point, boxes []image.Rectangle) bool {
	for _, box := range boxes {
		if p.In(box) {
			return true
		}
	}
	return false
}

// visibleRowsEqual 比對所有仍由舊文字擁有的像素，不接受任意差異。
func visibleRowsEqual(expected, raw []byte, rows, masks []image.Rectangle) bool {
	if len(raw) != 64000 {
		return false
	}
	i := 0
	for _, row := range rows {
		if !row.In(image.Rect(0, 0, 320, 200)) {
			return false
		}
		for y := row.Min.Y; y < row.Max.Y; y++ {
			for x := row.Min.X; x < row.Max.X; x++ {
				if i >= len(expected) || !occluded(image.Pt(x, y), masks) && expected[i] != raw[y*320+x] {
					return false
				}
				i++
			}
		}
	}
	return i == len(expected)
}

func visiblePixel(p, origin image.Point, masks []image.Rectangle) bool {
	return !occluded(image.Pt((p.X-origin.X)/3, (p.Y-origin.Y)/3), masks)
}

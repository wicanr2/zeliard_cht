package overlay

import (
	"bytes"
	"image"

	"github.com/wicanr2/dosgolem/oracle"
)

// 裝備頁 AA2B 的 ASCIIZ 迴圈每字重新進入。只有完整來源、逐字位置、
// 顏色及原版完成事件全部相符，才保存可覆繪標籤。
func (l *Labels) beginInventory(o *oracle.Oracle) {
	valid := o.VideoMode() == 0x13 && o.DSReg() == 0x01b1 &&
		bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0xaa2b), 6), []byte{0xac, 0x0a, 0xc0, 0x75, 0x01, 0xc3}) &&
		bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0xaa4a), 5), []byte{0x2e, 0xff, 0x16, 0x22, 0x20}) &&
		bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0x24ea), 8), []byte{0, 9, 18, 27, 36, 45, 54, 63}) &&
		o.Byte(oracle.Far(0x01b1, 0xff77)) == 0
	if !valid {
		l.inventoryPending = nil
		return
	}
	if l.inventoryPending == nil {
		for _, s := range l.sources {
			if !s.inventory || int(s.base)+s.offset != int(o.SI()) {
				continue
			}
			ah := uint8(o.AX() >> 8)
			if s.offset == 0xa92 && ah != 1 || s.offset != 0xa92 && ah != 2 && ah != 3 {
				return
			}
			h := 8
			shadow := uint8(0)
			if ah != 1 {
				h, shadow = 9, 45
			}
			x, y := int(o.BX()), int(o.CX()&255)
			box := image.Rect(x, y, x+(len(s.raw)-1)*8, y+h)
			if !box.In(image.Rect(0, 0, 320, 200)) {
				return
			}
			l.inventoryPending = &labelStamp{source: s, box: box, rows: []image.Rectangle{box}, foreground: ah * 9, shadow: shadow}
			l.inventoryIndex = 0
			break
		}
	}
	p := l.inventoryPending
	if p == nil {
		return
	}
	s := p.source
	if !bytes.Equal(s.prefix, o.Bytes(oracle.Far(0x01b1, s.base), len(s.prefix))) ||
		!bytes.Equal(s.raw, o.Bytes(oracle.Far(0x01b1, uint16(int(s.base)+s.offset)), len(s.raw))) ||
		l.inventoryIndex >= len(s.raw) || int(o.SI()) != int(s.base)+s.offset+l.inventoryIndex ||
		int(o.BX()) != p.box.Min.X+l.inventoryIndex*8 || int(o.CX()&255) != p.box.Min.Y ||
		uint8(o.AX()>>8)*9 != p.foreground {
		l.inventoryPending = nil
		return
	}
	if s.raw[l.inventoryIndex] != 0 {
		l.inventoryIndex++
	}
}

func (l *Labels) finishInventory(o *oracle.Oracle) {
	p := l.inventoryPending
	l.inventoryPending = nil
	if p == nil || l.inventoryIndex != len(p.source.raw)-1 || int(o.SI()) != int(p.source.base)+p.source.offset+len(p.source.raw) ||
		o.DSReg() != 0x01b1 || uint8(o.AX()>>8)*9 != p.foreground {
		return
	}
	p.raw = labelRowsCrop(o.Indexed(), p.rows)
	ink := false
	for _, v := range p.raw {
		if v != 0 && v != p.foreground && (p.shadow == 0 || v != p.shadow) {
			return
		}
		ink = ink || v == p.foreground
	}
	if !ink {
		return
	}
	p.palette, p.paintSerial = o.Palette(), l.paint.serial
	l.store(*p)
}

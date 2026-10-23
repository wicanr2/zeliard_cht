package overlay

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/glyphpatch"
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
		l.dropInventory(o)
		return
	}
	if l.inventoryPending == nil {
		for _, s := range l.sources {
			if !s.inventory || int(s.base)+s.offset != int(o.SI()) {
				continue
			}
			ah := uint8(o.AX() >> 8)
			white := s.inventoryWhite
			if white && ah != 1 || !white && ah != 2 && ah != 3 {
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
		l.dropInventory(o)
		return
	}
	if s.raw[l.inventoryIndex] != 0 {
		l.inventoryIndex++
		// 規格066：已核對的物品名字形由中文取代；AA44 為陰影、AA4F 為字色。
		glyphpatch.Claim(o, 0xaa44, 0xaa4f)
	}
}

// dropInventory 放棄進行中的物品名；已壓下的字形照原版寫回。
func (l *Labels) dropInventory(o *oracle.Oracle) {
	if p := l.inventoryPending; p != nil {
		glyphpatch.Unclaim(o, 0xaa44, 0xaa4f)
		glyphpatch.Reveal(o, p.rows...)
	}
	l.inventoryPending = nil
}

// inventoryIconFrame 只接受本輪正常路徑的完整面板、零物品圖示與來源。
// 原版NO USE圖示並非AA2B繪字事件，證據見CONTEXT的item-use-v1入口。
func (l *Labels) inventoryIconFrame(raw []byte, pal [256][3]uint8) bool {
	s := l.sources["zelres2.file.004ab8"]
	o := l.original
	if s == nil || o == nil || len(raw) != 64000 || o.VideoMode() != 0x13 || !s.inventory || len(s.text) != 2 ||
		!bytes.Equal(o.Bytes(oracle.Far(0x01b1, s.base), len(s.prefix)), s.prefix) ||
		!bytes.Equal(o.Bytes(oracle.Far(0x01b1, uint16(int(s.base)+s.offset)), len(s.raw)), s.raw) {
		return false
	}
	for _, ch := range s.text {
		if len(l.font[ch]) != 32 {
			return false
		}
	}
	var palette []byte
	for _, c := range pal {
		palette = append(palette, c[:]...)
	}
	if fmt.Sprintf("%x", sha256.Sum256(palette)) != "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c" ||
		fmt.Sprintf("%x", sha256.Sum256(labelCrop(raw, image.Rect(58, 131, 74, 147)))) != "1eee9c14d388ec401a73b41898b6b9cfc7497315a292fd1b6828ae293c7278a3" {
		return false
	}
	top := fmt.Sprintf("%x", sha256.Sum256(labelCrop(raw, image.Rect(47, 12, 272, 63))))
	return top == "8e9944c44008748fdc0780a518f123327fd04830882fcbc9a001af54cf95c6d7" || top == "a2bec73d332b999c8094fd6a6df45e379ecf278c1173a5541bf756685e74a672"
}

func (l *Labels) observeInventoryIcon(raw []byte, pal [256][3]uint8) {
	if !l.inventoryIconFrame(raw, pal) {
		return
	}
	box := image.Rect(59, 132, 73, 146)
	s := l.sources["zelres2.file.004ab8"]
	for _, stamp := range l.stamps {
		if stamp.noUseIcon && stamp.source == s && stamp.box == box && stamp.palette == pal && stamp.paintSerial == l.paint.serial && bytes.Equal(stamp.raw, labelCrop(raw, box)) {
			return
		}
	}
	l.store(labelStamp{source: s, box: box, rows: []image.Rectangle{box}, raw: labelCrop(raw, box), palette: pal,
		paintSerial: l.paint.serial, foreground: 9, noUseIcon: true})
}

func (l *Labels) finishInventory(o *oracle.Oracle) {
	glyphpatch.Unclaim(o, 0xaa44, 0xaa4f)
	p := l.inventoryPending
	l.inventoryPending = nil
	// 規格066：已認領的物品名若最後沒有存入，照原版寫回，不留空白。
	stored := false
	defer func() {
		if p != nil && !stored {
			glyphpatch.Reveal(o, p.rows...)
		}
	}()
	if p == nil || l.inventoryIndex != len(p.source.raw)-1 || int(o.SI()) != int(p.source.base)+p.source.offset+len(p.source.raw) ||
		o.DSReg() != 0x01b1 || uint8(o.AX()>>8)*9 != p.foreground {
		return
	}
	p.raw = labelRowsCrop(glyphpatch.Indexed(o), p.rows)
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
	stored = true
	l.store(*p)
}

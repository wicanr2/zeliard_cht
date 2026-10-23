package overlay

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/glyphpatch"
	"image"
)

// 商店標題使用 22CD 的 counted-row，並非 ASCIIZ。來源只限 countedTitles 列出的位置：
// 魔法物品店已以正常路徑驗證；旅館同格式，後段城鎮未正常到達。
func (l *Labels) beginItemTitle(o *oracle.Oracle) {
	l.itemTitlePending = nil
	if o.VideoMode() != 0x13 || o.DSReg() != 0x01b1 ||
		fmt.Sprintf("%x", sha256.Sum256(glyphpatch.CodeView(o, oracle.Far(0x01b1, 0x22cd), 0x78))) != "f7415a3888eb9e7bafd8d68b76f3e281a7fe4a6c5b15275e5536d40a7594fef4" {
		return
	}
	for _, s := range l.sources {
		if !s.itemTitle || len(s.titleHeader) != 4 {
			continue
		}
		at := uint16(int(s.base) + s.offset)
		if o.SI() != at-4 || !bytes.Equal(s.prefix, o.Bytes(oracle.Far(0x01b1, s.base), len(s.prefix))) ||
			!bytes.Equal(o.Bytes(oracle.Far(0x01b1, at-4), 4), s.titleHeader) ||
			!bytes.Equal(s.raw, o.Bytes(oracle.Far(0x01b1, at), len(s.raw))) {
			continue
		}
		h := s.titleHeader
		x, y := int(h[0])*4+int(h[2]), int(h[1])
		box := image.Rect(x, y, x+int(h[3])*5, y+8)
		offset := image.Pt(1, 0)
		l.itemTitlePending = &labelStamp{source: s, box: box, rows: []image.Rectangle{box}, foreground: 9, background: 5, shadow: 45, shadowOffset: &offset}
		glyphpatch.ClaimText(o)
		return
	}
}

func (l *Labels) finishItemTitle(o *oracle.Oracle) {
	p := l.itemTitlePending
	l.itemTitlePending = nil
	// 規格066：已認領的標題若最後沒有存入，照原版寫回，不留空白。
	stored := false
	defer func() {
		if p != nil && !stored {
			glyphpatch.Reveal(o, p.rows...)
		}
	}()
	if p == nil || o.DSReg() != 0x01b1 || int(o.SI()) != int(p.source.base)+p.source.offset+len(p.source.raw) || o.Byte(oracle.Far(0x01b1, 0x2cbd)) != 9 || o.Byte(oracle.Far(0x01b1, 0x2cbe)) != 45 {
		return
	}
	p.raw = labelRowsCrop(glyphpatch.Indexed(o), p.rows)
	ink := false
	for _, v := range p.raw {
		if v != 5 && v != 9 && v != 45 {
			return
		}
		ink = ink || v == 9
	}
	if !ink {
		return
	}
	p.palette, p.paintSerial = o.Palette(), l.paint.serial
	stored = true
	l.store(*p)
}

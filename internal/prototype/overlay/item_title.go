package overlay

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/wicanr2/dosgolem/oracle"
	"image"
)

// 魔法物品店標題使用固定25字元 counted-row，並非 ASCIIZ。
// 此入口只接受已驗魔法物品店來源，不把其他場景的 header 格式推廣為通用契約。
func (l *Labels) beginItemTitle(o *oracle.Oracle) {
	l.itemTitlePending = nil
	s := l.sources["zelres2.file.01ada8"]
	if s == nil || !s.itemTitle || o.VideoMode() != 0x13 || o.DSReg() != 0x01b1 || o.SI() != 0xa81c ||
		fmt.Sprintf("%x", sha256.Sum256(o.Bytes(oracle.Far(0x01b1, 0x22cd), 0x78))) != "f7415a3888eb9e7bafd8d68b76f3e281a7fe4a6c5b15275e5536d40a7594fef4" ||
		!bytes.Equal(s.prefix, o.Bytes(oracle.Far(0x01b1, 0xa000), len(s.prefix))) ||
		!bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0xa81c), 4), []byte{0x0e, 0xaf, 0, 25}) ||
		!bytes.Equal(s.raw, o.Bytes(oracle.Far(0x01b1, 0xa820), len(s.raw))) {
		return
	}
	box := image.Rect(56, 175, 181, 183)
	offset := image.Pt(1, 0)
	l.itemTitlePending = &labelStamp{source: s, box: box, rows: []image.Rectangle{box}, foreground: 9, background: 5, shadow: 45, shadowOffset: &offset}
}

func (l *Labels) finishItemTitle(o *oracle.Oracle) {
	p := l.itemTitlePending
	l.itemTitlePending = nil
	if p == nil || o.DSReg() != 0x01b1 || o.SI() != 0xa839 || o.Byte(oracle.Far(0x01b1, 0x2cbd)) != 9 || o.Byte(oracle.Far(0x01b1, 0x2cbe)) != 45 {
		return
	}
	p.raw = labelRowsCrop(o.Indexed(), p.rows)
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
	l.store(*p)
}

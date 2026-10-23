package overlay

import (
	"bytes"
	"fmt"
	"image"
	"strings"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
	"github.com/wicanr2/zeliard_cht/internal/prototype/glyphpatch"
)

// 洞窟告示由洞窟模組 740C 逐字繪製（規格066，靜態反組譯，強推論）：
// SI 指向 [Y][行數] 之後各行「X 位移＋字串」，以「/」分行、FF 結束。
// Y 加 19h 為第一列，每列加 12；X 為 58h 加位移；字形經 7460 呼叫 27E9，返回 7465，色號 AH=1。
// 讀到 FF 時由 7456 返回。告示所在的地圖模組載在 01B1:C000。
const (
	signGlyphReturn = 0x7465
	signInk         = 1
)

func (l *Labels) addSign(key, text string, data []byte, module, off int, font bitmap.Font) error {
	if off+3 >= len(data) || l.sources[key] != nil {
		return fmt.Errorf("告示來源錯誤：%s", key)
	}
	y0, count := int(data[off])+0x19, int(data[off+1])
	var rows []image.Rectangle
	var lines []string
	i := off + 2
	for {
		if i >= len(data) {
			return fmt.Errorf("告示缺結束碼：%s", key)
		}
		x := 0x58 + int(data[i])
		i++
		j := i
		for j < len(data) && data[j] != '/' && data[j] != 0xff {
			if data[j] < 0x20 || data[j] >= 0x7f {
				return fmt.Errorf("告示含未知控制碼：%s", key)
			}
			j++
		}
		if j >= len(data) || j == i {
			return fmt.Errorf("告示行錯誤：%s", key)
		}
		y := y0 + 12*len(lines)
		rows = append(rows, image.Rect(x, y, x+(j-i)*8, y+8))
		lines = append(lines, string(data[i:j]))
		if data[j] == 0xff {
			i = j
			break
		}
		i = j + 1
	}
	if len(lines) != count {
		return fmt.Errorf("告示行數不符：%s", key)
	}
	translated := strings.Split(text, "\n")
	if text == "" || len(translated) > len(lines) {
		return fmt.Errorf("告示譯文超過原版行數：%s", key)
	}
	for n, line := range translated {
		width := 0
		for _, ch := range line {
			g := font[ch]
			if len(g) != 16 && len(g) != 32 {
				return fmt.Errorf("告示字型缺字：%U", ch)
			}
			width += len(g) / 16 * 12
		}
		if width > len(lines[n])*8*3 {
			return fmt.Errorf("告示譯文超出原文安全寬度：%s", key)
		}
	}
	l.sources[key] = &labelSource{offset: off - module, base: 0xc000, prefix: append([]byte(nil), data[module:module+64]...), width: 8, raw: append([]byte(nil), data[off:i+1]...), text: []rune(text), signRows: rows}
	return nil
}

func (l *Labels) beginSign(o *oracle.Oracle) {
	l.signPending = nil
	if o.VideoMode() != 0x13 || o.DSReg() != 0x01b1 || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0x6000), len(l.caveModulePrefix)), l.caveModulePrefix) {
		return
	}
	for _, s := range l.sources {
		if s.signRows == nil || int(s.base)+s.offset != int(o.SI()) || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, s.base), len(s.prefix)), s.prefix) || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, o.SI()), len(s.raw)), s.raw) {
			continue
		}
		box := s.signRows[0]
		for _, r := range s.signRows[1:] {
			box = box.Union(r)
		}
		l.signPending = &labelStamp{source: s, box: box, rows: s.signRows}
		glyphpatch.HoldGlyphs(o, signGlyphReturn)
		return
	}
}

// finishSign 在 7456 讀到 FF 返回時收下告示；字列只能有一種底色與原版字色，否則照原版寫回。
func (l *Labels) finishSign(o *oracle.Oracle) {
	p := l.signPending
	if p == nil || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0x6000), len(l.caveModulePrefix)), l.caveModulePrefix) {
		return
	}
	l.signPending = nil
	glyphpatch.ReleaseLines(o)
	p.raw = labelRowsCrop(glyphpatch.Indexed(o), p.rows)
	background, ink := -1, false
	for _, v := range p.raw {
		if v == signInk {
			ink = true
			continue
		}
		if background >= 0 && int(v) != background {
			glyphpatch.Reveal(o, p.rows...)
			return
		}
		background = int(v)
	}
	if !ink || background < 0 {
		glyphpatch.Reveal(o, p.rows...)
		return
	}
	p.foreground, p.background = signInk, uint8(background)
	p.palette, p.paintSerial = o.Palette(), l.paint.serial
	l.store(*p)
}

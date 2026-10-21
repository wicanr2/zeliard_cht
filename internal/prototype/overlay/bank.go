package overlay

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

const bankNumberAddress uint16 = 0xad30

// 銀行標題使用固定八字元 counted-row，並非 ASCIIZ。
// 此入口只接受已驗銀行來源，不把其他場景的 header 格式推廣為通用契約。
func (l *Labels) beginBankTitle(o *oracle.Oracle) {
	l.bankTitlePending = nil
	s := l.sources["zelres2.file.019d54"]
	if s == nil || !s.bankTitle || o.VideoMode() != 0x13 || o.DSReg() != 0x01b1 || o.SI() != 0xa8ee ||
		fmt.Sprintf("%x", sha256.Sum256(o.Bytes(oracle.Far(0x01b1, 0x22cd), 0x78))) != "f7415a3888eb9e7bafd8d68b76f3e281a7fe4a6c5b15275e5536d40a7594fef4" ||
		!bytes.Equal(s.prefix, o.Bytes(oracle.Far(0x01b1, 0xa000), len(s.prefix))) ||
		!bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0xa8ee), 4), []byte{0x18, 0xaf, 2, 8}) ||
		!bytes.Equal(s.raw, o.Bytes(oracle.Far(0x01b1, 0xa8f2), len(s.raw))) {
		return
	}
	box := image.Rect(98, 175, 138, 183)
	offset := image.Pt(1, 0)
	l.bankTitlePending = &labelStamp{source: s, box: box, rows: []image.Rectangle{box}, foreground: 9, background: 5, shadow: 45, shadowOffset: &offset}
}

func (l *Labels) finishBankTitle(o *oracle.Oracle) {
	p := l.bankTitlePending
	l.bankTitlePending = nil
	if p == nil || o.DSReg() != 0x01b1 || o.SI() != 0xa8fa || o.Byte(oracle.Far(0x01b1, 0x2cbd)) != 9 || o.Byte(oracle.Far(0x01b1, 0x2cbe)) != 45 {
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

type bankSource struct {
	key, text   string
	first, last quoteFragment
	after       []string
}
type bankPart struct {
	source *bankSource
	number quoteNumber
}

// bankDialogue 只觀察銀行的已驗文字鏈。問句與後接的兩筆結果分開保存，
// 每筆數字取自當時已完成的原版字形，不讀帳戶欄位計算餘額。
type bankDialogue struct {
	font                                   bitmap.Font
	paint                                  *paintHistory
	prefix                                 []byte
	sources                                []*bankSource
	parts                                  []*bankPart
	current                                *bankPart
	phase, index                           int
	active, pending, pendingNumber, prompt bool
	completed                              []byte
	palette                                [256][3]uint8
	paintSerial                            uint64
	Glyphs, Completed                      int
}

func bankFragment(data []byte, start, end int) quoteFragment {
	f := quoteFragment{start: uint16(0xa000 + start - 0x19462), raw: append([]byte(nil), data[start:end]...)}
	for i := 0; i < len(f.raw); i++ {
		if f.raw[i] == 255 {
			i++
			continue
		}
		if f.raw[i] >= 32 && f.raw[i] < 127 && f.raw[i] != '/' {
			f.positions = append(f.positions, f.start+uint16(i))
		}
	}
	return f
}

func newBankDialogue(o *oracle.Oracle, data []byte, texts map[string]string, font bitmap.Font, paint *paintHistory) (*bankDialogue, error) {
	b := &bankDialogue{font: font, paint: paint, prefix: append([]byte(nil), data[0x19462:0x194a2]...)}
	// SAR2 檔案偏移；動態片段中間的數字與 & 控制碼由原版消費。
	bindings := [][4]int{
		{0x19def, 0x19e15, 0, 0}, {0x19e15, 0x19e3c, 0, 0},
		{0x19f2d, 0x19f57, 0, 0}, {0x19fe3, 0x1a006, 0, 0},
		{0x1a05a, 0x1a073, 0, 0}, {0x1a100, 0x1a137, 0, 0},
		{0x19f57, 0x19f69, 0x19f69, 0x19f73},
		{0x1a006, 0x1a01b, 0x1a01b, 0x1a023},
		{0x1a073, 0x1a07e, 0x1a07e, 0x1a098},
	}
	for _, v := range bindings {
		key := fmt.Sprintf("zelres2.file.%06x", v[0])
		text, ok := texts[key]
		if !ok {
			continue
		}
		delete(texts, key)
		plain := text
		if v[2] != 0 {
			if strings.Count(text, "{amount}") != 1 {
				return nil, fmt.Errorf("銀行數字欄位不符：%s", key)
			}
			plain = strings.ReplaceAll(text, "{amount}", "")
		}
		if plain == "" || strings.ContainsAny(plain, "{}\n\r\t") {
			return nil, fmt.Errorf("銀行試譯不相容：%s", key)
		}
		if _, err := dialogueLayout(font, []rune(plain)); err != nil {
			return nil, err
		}
		s := &bankSource{key: key, text: text, first: bankFragment(data, v[0], v[1])}
		if v[2] != 0 {
			s.last = bankFragment(data, v[2], v[3])
		}
		switch v[0] {
		case 0x19f57:
			s.after = []string{"zelres2.file.019f2d", "zelres2.file.01a006"}
		case 0x1a006:
			s.after = []string{"zelres2.file.019fe3"}
		}
		b.sources = append(b.sources, s)
	}
	if len(b.sources) == 0 {
		return nil, nil
	}
	if len(b.sources) != len(bindings) {
		return nil, fmt.Errorf("銀行試譯須完整提供九個來源")
	}
	o.OnCall(oracle.Far(0x01b1, 0x711b), b.begin)
	o.OnCall(oracle.Far(0x01b1, 0x7120), b.capture)
	o.OnCall(oracle.Far(0x01b1, 0x7205), func(*oracle.Oracle) { b.reset() })
	o.OnCall(oracle.Far(0x01b1, 0x71c9), func(o *oracle.Oracle) {
		if b.active {
			b.prompt = true
			b.captureFrame(o)
		}
	})
	o.OnCall(oracle.Far(0x01b1, 0x71d9), func(o *oracle.Oracle) {
		if b.active {
			b.prompt = false
			b.captureFrame(o)
		}
	})
	return b, nil
}

func (b *bankDialogue) reset() {
	b.active, b.pending, b.pendingNumber, b.prompt = false, false, false, false
	b.current = nil
	b.parts = nil
	b.completed = nil
	b.phase, b.index = 0, 0
}

func (b *bankDialogue) begin(o *oracle.Oracle) {
	pos := o.Word(oracle.Far(0x01b1, 0xff4c)) - 1
	var start *bankSource
	for _, s := range b.sources {
		if len(s.first.positions) > 0 && pos == s.first.positions[0] {
			start = s
			break
		}
	}
	if !b.active && start == nil {
		return
	}
	if o.DSReg() != 0x01b1 || o.AX()>>8 != 1 || o.BX() < 54 || o.BX()+8 > 264 || o.CX()&255 < 99 || o.CX()&255 > 139 || byte(o.AX()) != o.Byte(oracle.Far(0x01b1, pos)) || !bytes.Equal(b.prefix, o.Bytes(oracle.Far(0x01b1, 0xa000), len(b.prefix))) || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0x711b), 5), []byte{0x2e, 0xff, 0x16, 0x22, 0x20}) {
		b.reset()
		return
	}
	if start != nil {
		if len(start.after) == 0 {
			b.reset()
		} else {
			valid := false
			if len(b.parts) > 0 && b.current == nil {
				for _, key := range start.after {
					valid = valid || b.parts[len(b.parts)-1].source.key == key
				}
			}
			if !valid {
				b.reset()
				return
			}
		}
		b.active = true
		b.current = &bankPart{source: start}
		b.phase, b.index = 0, 0
	}
	if b.current == nil {
		b.reset()
		return
	}
	s := b.current.source
	fragment := s.first
	b.pendingNumber = false
	switch b.phase {
	case 1:
		n := &b.current.number
		if b.index == 0 {
			// 固定銀行來源的數字暫存位置由正常交易字形追蹤核對。
			if pos != bankNumberAddress {
				b.reset()
				return
			}
			raw := o.Bytes(oracle.Far(0x01b1, pos), 10)
			end := 0
			for end < len(raw) && raw[end] >= '0' && raw[end] <= '9' {
				end++
			}
			if end < 1 || end > 8 || end >= len(raw) || raw[end] != 255 {
				b.reset()
				return
			}
			n.raw = append([]byte(nil), raw[:end+1]...)
			n.box = image.Rect(int(o.BX()), int(o.CX()&255), int(o.BX())+8, int(o.CX()&255)+8)
		}
		if int(o.CX()&255) != n.box.Min.Y || int(o.BX()) < n.box.Min.X || int(o.BX()) > n.box.Max.X {
			b.reset()
			return
		}
		n.box.Max.X = int(o.BX()) + 8
		fragment = quoteFragment{start: bankNumberAddress, raw: n.raw}
		for i := 0; i < len(n.raw)-1; i++ {
			fragment.positions = append(fragment.positions, bankNumberAddress+uint16(i))
		}
		b.pendingNumber = true
	case 2:
		fragment = s.last
	}
	if b.index >= len(fragment.positions) || pos != fragment.positions[b.index] || !bytes.Equal(fragment.raw, o.Bytes(oracle.Far(0x01b1, fragment.start), len(fragment.raw))) {
		b.reset()
		return
	}
	b.index++
	if b.index == len(fragment.positions) {
		b.phase++
		b.index = 0
	}
	b.pending = true
}

func (b *bankDialogue) captureFrame(o *oracle.Oracle) {
	b.completed = dialogueCrop(o.Indexed())
	b.palette = o.Palette()
	b.paintSerial = b.paint.serial
}
func (b *bankDialogue) capture(o *oracle.Oracle) {
	if !b.active || !b.pending || b.current == nil {
		return
	}
	if b.pendingNumber {
		n := &b.current.number
		n.pixels = labelCrop(o.Indexed(), n.box)
		for _, v := range n.pixels {
			if v != 0 && v != 9 {
				b.reset()
				return
			}
		}
	}
	b.captureFrame(o)
	b.pending, b.pendingNumber = false, false
	b.Glyphs++
	if b.phase == 3 || b.phase == 1 && len(b.current.source.last.raw) == 0 {
		b.parts = append(b.parts, b.current)
		b.current = nil
		b.Completed++
	}
}

func (b *bankDialogue) tokens() []quoteToken {
	var tokens []quoteToken
	for _, p := range b.parts {
		text := p.source.text
		for len(text) > 0 {
			at := strings.Index(text, "{amount}")
			if at < 0 {
				at = len(text)
			}
			for _, ch := range text[:at] {
				tokens = append(tokens, quoteToken{ch: ch, width: len(b.font[ch]) / 16 * 16})
			}
			text = text[at:]
			if text == "" {
				break
			}
			tokens = append(tokens, quoteToken{number: &p.number, width: p.number.box.Dx() * 3})
			text = text[len("{amount}"):]
		}
	}
	return tokens
}

func (b *bankDialogue) Draw(dst *image.RGBA, raw []byte, pal [256][3]uint8, origin image.Point) bool {
	if !b.active || len(b.completed) != dialogueBox.Dx()*dialogueBox.Dy() {
		return false
	}
	masks, known := b.paint.masks(b.paintSerial, raw)
	if pal != b.palette || !known || !b.pending && !visibleRowsEqual(b.completed, raw, []image.Rectangle{dialogueBox}, masks) {
		b.reset()
		return false
	}
	tokens := b.tokens()
	positions, valid := quoteLayout(tokens)
	if !valid {
		b.reset()
		return false
	}
	for _, t := range tokens {
		if n := t.number; n != nil {
			if n.box.Empty() || len(n.pixels) != n.box.Dx()*n.box.Dy() {
				b.reset()
				return false
			}
		}
	}
	box := image.Rect(origin.X+dialogueBox.Min.X*3, origin.Y+dialogueBox.Min.Y*3, origin.X+dialogueBox.Max.X*3, origin.Y+dialogueBox.Max.Y*3)
	prompt := image.Rect(origin.X+dialoguePrompt.Min.X*3, origin.Y+dialoguePrompt.Min.Y*3, origin.X+dialoguePrompt.Max.X*3, origin.Y+dialoguePrompt.Max.Y*3)
	put := func(p image.Point, c [3]uint8) {
		if p.In(dst.Bounds()) && (!b.prompt || !p.In(prompt)) && visiblePixel(p, origin, masks) {
			dst.SetRGBA(p.X, p.Y, color.RGBA{c[0], c[1], c[2], 255})
		}
	}
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			put(image.Pt(x, y), pal[0])
		}
	}
	for i, t := range tokens {
		at := positions[i].Add(box.Min)
		if n := t.number; n != nil {
			if len(n.pixels) != n.box.Dx()*n.box.Dy() {
				b.reset()
				return false
			}
			for y := 0; y < n.box.Dy()*3; y++ {
				for x := 0; x < n.box.Dx()*3; x++ {
					if n.pixels[y/3*n.box.Dx()+x/3] == 9 {
						put(at.Add(image.Pt(x, y+4)), pal[9])
					}
				}
			}
			continue
		}
		c := pal[9]
		b.font.DrawGlyph(dst, at, t.ch, 32, color.RGBA{c[0], c[1], c[2], 255}, func(p image.Point) bool { return (!b.prompt || !p.In(prompt)) && visiblePixel(p, origin, masks) })
	}
	return true
}

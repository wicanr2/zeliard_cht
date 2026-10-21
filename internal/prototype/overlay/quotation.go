package overlay

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

type quoteFragment struct {
	start     uint16
	raw       []byte
	positions []uint16
}
type quoteNumber struct {
	raw    []byte
	box    image.Rectangle
	pixels []byte
}
type quoteToken struct {
	ch     rune
	number *quoteNumber
	width  int
}

// Quotation 只合成已驗證的武器報價來源鏈。數字保留原版字形，
// 商品譯名引用 Labels 的同一份來源；交易與資金判定仍由原版處理。
type Quotation struct {
	font                bitmap.Font
	paint               *paintHistory
	labels              *Labels
	prefix              []byte
	fragments           map[int]quoteFragment
	texts               [4]string
	stage, index, ready int
	active, pending     bool
	item                *labelSource
	price, trade        quoteNumber
	pendingNumber       *quoteNumber
	completed           []byte
	paintSerial         uint64
	Glyphs, Completed   int
}

func newQuotation(o *oracle.Oracle, data []byte, texts map[string]string, font bitmap.Font, paint *paintHistory) (*Quotation, error) {
	keys := [4]string{"zelres2.file.0188f4", "zelres2.file.01891d", "zelres2.file.01885d", "zelres2.file.018904"}
	placeholders := [4]string{"item", "price", "trade", ""}
	q := &Quotation{font: font, paint: paint, fragments: map[int]quoteFragment{}, prefix: append([]byte(nil), data[0x17817:0x17857]...)}
	found := 0
	for i, key := range keys {
		text, ok := texts[key]
		if !ok {
			continue
		}
		found++
		delete(texts, key)
		if strings.ContainsAny(text, "\n\r\t") {
			return nil, fmt.Errorf("報價試譯不可內嵌換行")
		}
		plain := text
		if placeholders[i] != "" {
			placeholder := "{" + placeholders[i] + "}"
			if strings.Count(text, placeholder) != 1 {
				return nil, fmt.Errorf("報價 placeholder 不符：%s", key)
			}
			plain = strings.ReplaceAll(text, placeholder, "")
		}
		if strings.ContainsAny(plain, "{}") {
			return nil, fmt.Errorf("未知報價 placeholder：%s", key)
		}
		for _, ch := range plain {
			if len(font[ch]) != 16 && len(font[ch]) != 32 {
				return nil, fmt.Errorf("報價字型缺字：%U", ch)
			}
		}
		q.texts[i] = text
	}
	if found == 0 {
		return nil, nil
	}
	if found != 4 {
		return nil, fmt.Errorf("報價試譯須完整提供四個片段")
	}
	for _, binding := range [][3]int{{0, 0xb0dd, 0xb0e7}, {2, 0xb0e7, 0xb0ed}, {3, 0xb106, 0xb115}, {5, 0xb115, 0xb11f}, {6, 0xb046, 0xb056}, {8, 0xb056, 0xb081}, {9, 0xb0ed, 0xb106}} {
		start, end := binding[1], binding[2]
		raw := append([]byte(nil), data[0x17817+start-0xa000:0x17817+end-0xa000]...)
		fragment := quoteFragment{start: uint16(start), raw: raw}
		for i := 0; i < len(raw); i++ {
			if raw[i] == 0xff {
				i++
				continue
			}
			if raw[i] >= 0x20 && raw[i] < 0x7f && raw[i] != '/' {
				fragment.positions = append(fragment.positions, uint16(start+i))
			}
		}
		q.fragments[binding[0]] = fragment
	}
	o.OnCall(oracle.Far(0x01b1, 0x711b), q.begin)
	o.OnCall(oracle.Far(0x01b1, 0x7120), q.capture)
	o.OnCall(oracle.Far(0x01b1, 0x7205), func(*oracle.Oracle) { q.reset() })
	return q, nil
}

func (q *Quotation) reset() {
	q.active, q.pending = false, false
	q.stage, q.index, q.ready = 0, 0, 0
	q.item = nil
	q.price, q.trade = quoteNumber{}, quoteNumber{}
	q.pendingNumber = nil
	q.completed = nil
}

func (q *Quotation) begin(o *oracle.Oracle) {
	pos := o.Word(oracle.Far(0x01b1, 0xff4c)) - 1
	if !q.active {
		if pos != 0xb0dd || q.labels == nil {
			return
		}
		q.reset()
		q.active = true
	}
	if q.stage >= 10 || o.DSReg() != 0x01b1 || o.AX()>>8 != 1 || int(o.BX()) < 54 || int(o.BX())+8 > 264 || o.CX()&255 < 99 || o.CX()&255 > 139 || byte(o.AX()) != o.Byte(oracle.Far(0x01b1, pos)) || !bytes.Equal(q.prefix, o.Bytes(oracle.Far(0x01b1, 0xa000), len(q.prefix))) || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0x711b), 5), []byte{0x2e, 0xff, 0x16, 0x22, 0x20}) {
		q.reset()
		return
	}
	var fragment quoteFragment
	q.pendingNumber = nil
	switch q.stage {
	case 1:
		if q.item == nil {
			for _, s := range q.labels.sources {
				if s.base == 0xa000 && s.width == 5 && int(pos) == int(s.base)+s.offset && bytes.Equal(s.prefix, q.prefix) && bytes.Equal(s.raw, o.Bytes(oracle.Far(0x01b1, pos), len(s.raw))) {
					q.item = s
					break
				}
			}
		}
		if q.item == nil {
			q.reset()
			return
		}
		fragment.start = uint16(int(q.item.base) + q.item.offset)
		fragment.raw = q.item.raw
		for i := 0; i < len(fragment.raw)-1; i++ {
			fragment.positions = append(fragment.positions, fragment.start+uint16(i))
		}
	case 4, 7:
		number := &q.price
		if q.stage == 7 {
			number = &q.trade
		}
		if q.index == 0 {
			if pos != 0xbc33 {
				q.reset()
				return
			}
			raw := o.Bytes(oracle.Far(0x01b1, pos), 10)
			end := 0
			for end < len(raw) && raw[end] >= '0' && raw[end] <= '9' {
				end++
			}
			if end < 1 || end > 8 || end >= len(raw) || raw[end] != 0xff {
				q.reset()
				return
			}
			number.raw = append([]byte(nil), raw[:end+1]...)
			number.box = image.Rect(int(o.BX()), int(o.CX()&255), int(o.BX())+8, int(o.CX()&255)+8)
		}
		if int(o.CX()&255) != number.box.Min.Y || int(o.BX()) < number.box.Min.X || int(o.BX()) > number.box.Max.X {
			q.reset()
			return
		}
		number.box.Max.X = int(o.BX()) + 8
		fragment.start = 0xbc33
		fragment.raw = number.raw
		for i := 0; i < len(number.raw)-1; i++ {
			fragment.positions = append(fragment.positions, 0xbc33+uint16(i))
		}
		q.pendingNumber = number
	default:
		fragment = q.fragments[q.stage]
	}
	if q.index >= len(fragment.positions) || pos != fragment.positions[q.index] || !bytes.Equal(fragment.raw, o.Bytes(oracle.Far(0x01b1, fragment.start), len(fragment.raw))) {
		q.reset()
		return
	}
	q.index++
	if q.index == len(fragment.positions) {
		q.stage++
		q.index = 0
	}
	q.pending = true
}

func (q *Quotation) capture(o *oracle.Oracle) {
	if !q.active || !q.pending {
		return
	}
	raw := o.Indexed()
	if number := q.pendingNumber; number != nil {
		number.pixels = labelCrop(raw, number.box)
		for _, v := range number.pixels {
			if v != 0 && v != 9 {
				q.reset()
				return
			}
		}
	}
	q.completed = dialogueCrop(raw)
	q.paintSerial = q.paint.serial
	q.pending = false
	q.pendingNumber = nil
	q.ready = q.stage
	q.Glyphs++
	if q.ready == 10 {
		q.Completed++
	}
}

func (q *Quotation) tokens() []quoteToken {
	var tokens []quoteToken
	for i, stage := range []int{3, 6, 9, 10} {
		if q.ready < stage {
			break
		}
		text := q.texts[i]
		for len(text) > 0 {
			at := strings.IndexByte(text, '{')
			if at < 0 {
				at = len(text)
			}
			for _, ch := range text[:at] {
				tokens = append(tokens, quoteToken{ch: ch, width: len(q.font[ch]) / 16 * 16})
			}
			text = text[at:]
			if text == "" {
				break
			}
			end := strings.IndexByte(text, '}')
			switch text[1:end] {
			case "item":
				for _, ch := range q.item.text {
					tokens = append(tokens, quoteToken{ch: ch, width: len(q.font[ch]) / 16 * 16})
				}
			case "price":
				tokens = append(tokens, quoteToken{number: &q.price, width: q.price.box.Dx() * 3})
			case "trade":
				tokens = append(tokens, quoteToken{number: &q.trade, width: q.trade.box.Dx() * 3})
			}
			text = text[end+1:]
		}
	}
	return tokens
}

func quoteLayout(tokens []quoteToken) ([]image.Point, bool) {
	x, y := 6, 4
	var positions []image.Point
	for i, token := range tokens {
		needed := token.width
		for j := i + 1; j < len(tokens) && strings.ContainsRune("，。！？；：、）］」』】〉》…", tokens[j].ch); j++ {
			needed += tokens[j].width
		}
		if needed > dialogueBox.Dx()*3-12 {
			return nil, false
		}
		if x+needed > dialogueBox.Dx()*3-6 {
			x = 6
			y += 36
		}
		if token.width <= 0 || y+32 > 112 {
			return nil, false
		}
		positions = append(positions, image.Pt(x, y))
		x += token.width
	}
	return positions, true
}

func (q *Quotation) Draw(dst *image.RGBA, raw []byte, pal [256][3]uint8, origin image.Point) bool {
	if !q.active || len(q.completed) != dialogueBox.Dx()*dialogueBox.Dy() {
		return false
	}
	masks, known := q.paint.masks(q.paintSerial, raw)
	if !known || !q.pending && !visibleRowsEqual(q.completed, raw, []image.Rectangle{dialogueBox}, masks) {
		q.reset()
		return false
	}
	tokens := q.tokens()
	positions, valid := quoteLayout(tokens)
	if !valid {
		q.reset()
		return false
	}
	bg, fg := pal[0], pal[9]
	box := image.Rect(origin.X+dialogueBox.Min.X*3, origin.Y+dialogueBox.Min.Y*3, origin.X+dialogueBox.Max.X*3, origin.Y+dialogueBox.Max.Y*3)
	put := func(p image.Point, c [3]uint8) {
		if p.In(dst.Bounds()) && visiblePixel(p, origin, masks) {
			dst.SetRGBA(p.X, p.Y, color.RGBA{c[0], c[1], c[2], 255})
		}
	}
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			put(image.Pt(x, y), bg)
		}
	}
	for i, token := range tokens {
		at := positions[i].Add(box.Min)
		if n := token.number; n != nil {
			for y := 0; y < n.box.Dy()*3; y++ {
				for x := 0; x < n.box.Dx()*3; x++ {
					if n.pixels[y/3*n.box.Dx()+x/3] == 9 {
						put(at.Add(image.Pt(x, y+4)), fg)
					}
				}
			}
			continue
		}
		q.font.DrawGlyph(dst, at, token.ch, 32, color.RGBA{fg[0], fg[1], fg[2], 255}, func(p image.Point) bool { return visiblePixel(p, origin, masks) })
	}
	return true
}

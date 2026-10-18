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

// itemPurchase 只觀察已購物後的商品、原版金額與繼續購物詢問。
// 金額已由原版扣除；此顯示層不建立購買前確認或改變交易。
type itemPurchase struct {
	font                bitmap.Font
	paint               *paintHistory
	labels              *Labels
	prefix              []byte
	fragments           map[int]quoteFragment
	texts               [3]string
	stage, index, ready int
	active, pending     bool
	item                *labelSource
	price               quoteNumber
	pendingNumber       *quoteNumber
	completed           []byte
	palette             [256][3]uint8
	paintSerial         uint64
	Glyphs, Completed   int
}

func newItemPurchase(o *oracle.Oracle, data []byte, texts map[string]string, font bitmap.Font, paint *paintHistory) (*itemPurchase, error) {
	keys := [3]string{"zelres2.file.01ae4d", "zelres2.file.01ae7a", "zelres2.file.01ae92"}
	placeholders := [3]string{"item", "price", ""}
	q := &itemPurchase{font: font, paint: paint, fragments: map[int]quoteFragment{}, prefix: append([]byte(nil), data[0x1a588:0x1a5c8]...)}
	found := 0
	for i, key := range keys {
		text, ok := texts[key]
		if !ok {
			continue
		}
		found++
		delete(texts, key)
		if strings.ContainsAny(text, "\n\r\t") {
			return nil, fmt.Errorf("道具店購物試譯不可內嵌換行")
		}
		plain := text
		if placeholders[i] != "" {
			placeholder := "{" + placeholders[i] + "}"
			if strings.Count(text, placeholder) != 1 {
				return nil, fmt.Errorf("道具店購物 placeholder 不符：%s", key)
			}
			plain = strings.ReplaceAll(text, placeholder, "")
		}
		if strings.ContainsAny(plain, "{}") {
			return nil, fmt.Errorf("未知道具店購物 placeholder：%s", key)
		}
		for _, ch := range plain {
			if len(font[ch]) != 16 && len(font[ch]) != 32 {
				return nil, fmt.Errorf("道具店購物字型缺字：%U", ch)
			}
		}
		q.texts[i] = text
	}
	if found == 0 {
		return nil, nil
	}
	if found != 3 {
		return nil, fmt.Errorf("道具店購物試譯須完整提供三個片段")
	}
	for _, binding := range [][3]int{{0, 0xa8c5, 0xa8d4}, {2, 0xa8d4, 0xa8d8}, {3, 0xa8f2, 0xa901}, {5, 0xa901, 0xa90a}, {6, 0xa90a, 0xa929}} {
		start, end := binding[1], binding[2]
		raw := append([]byte(nil), data[0x1a588+start-0xa000:0x1a588+end-0xa000]...)
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

func (q *itemPurchase) reset() {
	q.active, q.pending = false, false
	q.stage, q.index, q.ready = 0, 0, 0
	q.item = nil
	q.price = quoteNumber{}
	q.pendingNumber = nil
	q.completed = nil
}

func (q *itemPurchase) begin(o *oracle.Oracle) {
	pos := o.Word(oracle.Far(0x01b1, 0xff4c)) - 1
	if !q.active {
		if pos != 0xa8c5 || q.labels == nil {
			return
		}
		q.reset()
		q.active = true
	}
	if q.stage >= 7 || o.DSReg() != 0x01b1 || o.AX()>>8 != 1 || int(o.BX()) < 54 || int(o.BX())+8 > 264 || o.CX()&255 < 99 || o.CX()&255 > 139 || byte(o.AX()) != o.Byte(oracle.Far(0x01b1, pos)) || !bytes.Equal(q.prefix, o.Bytes(oracle.Far(0x01b1, 0xa000), len(q.prefix))) || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0x711b), 5), []byte{0x2e, 0xff, 0x16, 0x22, 0x20}) {
		q.reset()
		return
	}
	var fragment quoteFragment
	q.pendingNumber = nil
	switch q.stage {
	case 1:
		if q.item == nil {
			if pos != 0xb09a && pos != 0xb0d4 && pos != 0xb0f4 {
				q.reset()
				return
			}
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
	case 4:
		number := &q.price
		if q.index == 0 {
			if pos != 0xb21e {
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
		fragment.start = 0xb21e
		fragment.raw = number.raw
		for i := 0; i < len(number.raw)-1; i++ {
			fragment.positions = append(fragment.positions, 0xb21e+uint16(i))
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

func (q *itemPurchase) capture(o *oracle.Oracle) {
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
	q.palette = o.Palette()
	q.paintSerial = q.paint.serial
	q.pending = false
	q.pendingNumber = nil
	q.ready = q.stage
	q.Glyphs++
	if q.ready == 7 {
		q.Completed++
	}
}

func (q *itemPurchase) tokens() []quoteToken {
	var tokens []quoteToken
	for i, stage := range []int{3, 6, 7} {
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
			}
			text = text[end+1:]
		}
	}
	return tokens
}

func (q *itemPurchase) Draw(dst *image.RGBA, raw []byte, pal [256][3]uint8, origin image.Point) bool {
	if !q.active || len(q.completed) != dialogueBox.Dx()*dialogueBox.Dy() {
		return false
	}
	masks, known := q.paint.masks(q.paintSerial, raw)
	if pal != q.palette || !known || !q.pending && !visibleRowsEqual(q.completed, raw, []image.Rectangle{dialogueBox}, masks) {
		q.reset()
		return false
	}
	tokens := q.tokens()
	positions, valid := quoteLayout(tokens)
	if !valid {
		q.reset()
		return false
	}
	for _, token := range tokens {
		if n := token.number; n != nil && (n.box.Empty() || len(n.pixels) != n.box.Dx()*n.box.Dy()) {
			q.reset()
			return false
		}
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
		g := q.font[token.ch]
		w := len(g) / 16 * 8
		for y := 0; y < 16; y++ {
			for x := 0; x < w; x++ {
				if g[y*(w/8)+x/8]&(0x80>>uint(x%8)) == 0 {
					continue
				}
				for yy := 0; yy < 2; yy++ {
					for xx := 0; xx < 2; xx++ {
						put(at.Add(image.Pt(x*2+xx, y*2+yy)), fg)
					}
				}
			}
		}
	}
	return true
}

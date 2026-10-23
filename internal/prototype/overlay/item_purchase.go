package overlay

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
	"github.com/wicanr2/zeliard_cht/internal/prototype/glyphpatch"
)

// itemPurchase 共用道具店購買後訊息、出售確認與商品說明的唯讀顯示鏈。
// 三種來源分開配對；原版決定交易、翻頁與等待輸入。
type itemPurchase struct {
	original            *oracle.Oracle
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
	start               uint16
	terminal            int
	thresholds          [3]int
	explanation, prompt bool
	details             map[uint16][2]string
}

func newItemPurchase(o *oracle.Oracle, data []byte, texts map[string]string, font bitmap.Font, paint *paintHistory) (*itemPurchase, error) {
	return newItemChain(o, data, texts, font, paint, itemChainSpec{
		keys: []string{"zelres2.file.01ae4d", "zelres2.file.01ae7a", "zelres2.file.01ae92"}, placeholders: []string{"item", "price", ""},
		start: 0xa8c5, terminal: 7, thresholds: [3]int{3, 6, 7},
		bindings: [][3]int{{0, 0xa8c5, 0xa8d4}, {2, 0xa8d4, 0xa8d8}, {3, 0xa8f2, 0xa901}, {5, 0xa901, 0xa90a}, {6, 0xa90a, 0xa929}},
	})
}

type itemChainSpec struct {
	keys, placeholders []string
	start              uint16
	terminal           int
	thresholds         [3]int
	bindings           [][3]int
	explanation        bool
}

func newItemSale(o *oracle.Oracle, data []byte, texts map[string]string, font bitmap.Font, paint *paintHistory) (*itemPurchase, error) {
	return newItemChain(o, data, texts, font, paint, itemChainSpec{
		keys: []string{"zelres2.file.01ae60", "zelres2.file.01af4c", "zelres2.file.01af6d"}, placeholders: []string{"item", "price", ""},
		start: 0xa8d8, terminal: 7, thresholds: [3]int{3, 6, 7},
		bindings: [][3]int{{0, 0xa8d8, 0xa8ef}, {2, 0xa8ef, 0xa8f2}, {3, 0xa9c4, 0xa9d4}, {5, 0xa9d4, 0xa9e5}, {6, 0xa9e5, 0xa9fe}},
	})
}

// itemDetail 是道具店商品名稱的模組位址，以及說明前段、後段起點與結束位置。
// 八件依固定 SAR 的名稱表與說明順序；前段止於第一個「/」之後，刀劍油沿 1.0.5 已驗切點。
type itemDetail struct {
	name   uint16
	bounds [3]uint16
}

var itemDetails = []itemDetail{
	{0xb09a, [3]uint16{0xab4a, 0xab86, 0xabc5}}, // Ken'ko 藥水
	{0xb0a8, [3]uint16{0xabc5, 0xac11, 0xac9c}}, // Juu-en 果實
	{0xb0b5, [3]uint16{0xac9c, 0xacf2, 0xad39}}, // Kashi 靈藥
	{0xb0c5, [3]uint16{0xad39, 0xadae, 0xadfd}}, // Chikara 粉末
	{0xb0d4, [3]uint16{0xadfd, 0xae34, 0xaea6}}, // Magia 石
	{0xb0e0, [3]uint16{0xaea6, 0xaedc, 0xaf3d}}, // Acero 聖水
	{0xb0f4, [3]uint16{0xaf3d, 0xaf7c, 0xafd3}}, // 刀劍油
	{0xb0fe, [3]uint16{0xafd3, 0xb01b, 0xb08a}}, // Kioku 羽毛
}

func itemDetailOf(name uint16) (itemDetail, bool) {
	for _, item := range itemDetails {
		if item.name == name {
			return item, true
		}
	}
	return itemDetail{}, false
}

func newItemExplanation(o *oracle.Oracle, data []byte, texts map[string]string, font bitmap.Font, paint *paintHistory) (*itemPurchase, error) {
	spec := itemChainSpec{
		keys: []string{"zelres2.file.01b053"}, placeholders: []string{"item"},
		start: 0xaacb, terminal: 5, thresholds: [3]int{3, 4, 5}, explanation: true,
		bindings: [][3]int{{0, 0xaacb, 0xaae6}, {2, 0xaae6, 0xaaea}},
	}
	for _, item := range itemDetails {
		for i := 0; i < 2; i++ {
			start, end := int(item.bounds[i]), int(item.bounds[i+1])
			spec.keys = append(spec.keys, fmt.Sprintf("zelres2.file.%06x", 0x1a588+start-0xa000))
			spec.placeholders = append(spec.placeholders, "")
			spec.bindings = append(spec.bindings, [3]int{start, start, end})
		}
	}
	return newItemChain(o, data, texts, font, paint, spec)
}

func newItemChain(o *oracle.Oracle, data []byte, texts map[string]string, font bitmap.Font, paint *paintHistory, spec itemChainSpec) (*itemPurchase, error) {
	q := &itemPurchase{original: o, font: font, paint: paint, fragments: map[int]quoteFragment{}, prefix: append([]byte(nil), data[0x1a588:0x1a5c8]...), start: spec.start, terminal: spec.terminal, thresholds: spec.thresholds, explanation: spec.explanation, details: map[uint16][2]string{}}
	translated := make([]string, len(spec.keys))
	found := 0
	for i, key := range spec.keys {
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
		if spec.placeholders[i] != "" {
			placeholder := "{" + spec.placeholders[i] + "}"
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
		translated[i] = text
	}
	if found == 0 {
		return nil, nil
	}
	if found != len(spec.keys) {
		return nil, fmt.Errorf("道具店試譯須完整提供此來源鏈的%d個片段", len(spec.keys))
	}
	if spec.explanation {
		q.texts[0] = translated[0]
		for i, item := range itemDetails {
			q.details[item.name] = [2]string{translated[1+i*2], translated[2+i*2]}
		}
	} else {
		copy(q.texts[:], translated)
	}
	for _, binding := range spec.bindings {
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
	if spec.explanation {
		o.OnCall(oracle.Far(0x01b1, 0x71a7), func(*oracle.Oracle) {
			if q.active {
				q.pending = true
			}
		})
		o.OnCall(oracle.Far(0x01b1, 0x71ac), func(o *oracle.Oracle) {
			if q.active {
				q.captureFrame(o)
			}
		})
		o.OnCall(oracle.Far(0x01b1, 0x71c9), func(o *oracle.Oracle) {
			if q.active {
				q.prompt = true
				q.captureFrame(o)
			}
		})
		o.OnCall(oracle.Far(0x01b1, 0x71d9), func(o *oracle.Oracle) {
			if q.active {
				q.prompt = false
				q.captureFrame(o)
			}
		})
	}
	return q, nil
}

// abort 用於已認領字形後的來源不符：壓下的字形照原版寫回後退出，不留空白框。
func (q *itemPurchase) abort(o *oracle.Oracle) {
	if q.active {
		glyphpatch.Reveal(o, dialogueBox)
	}
	q.reset()
}

func (q *itemPurchase) reset() {
	q.active, q.pending = false, false
	q.stage, q.index, q.ready = 0, 0, 0
	q.item = nil
	q.price = quoteNumber{}
	q.pendingNumber = nil
	q.completed = nil
	q.prompt = false
}

func (q *itemPurchase) begin(o *oracle.Oracle) {
	pos := o.Word(oracle.Far(0x01b1, 0xff4c)) - 1
	if !q.active {
		if pos != q.start || q.labels == nil {
			return
		}
		q.reset()
		q.active = true
	}
	// 介紹中的商品名在BX=258畫空白後換行，不能用八像素包圍盒拒絕可變字距文字。
	// 數字仍保留完整八列／八像素字形的邊界檢查。
	if q.stage >= q.terminal || o.DSReg() != 0x01b1 || o.AX()>>8 != 1 || int(o.BX()) < 54 || int(o.BX()) >= 264 || !q.explanation && int(o.BX())+8 > 264 || o.CX()&255 < 99 || o.CX()&255 > 139 || byte(o.AX()) != o.Byte(oracle.Far(0x01b1, pos)) || !bytes.Equal(q.prefix, o.Bytes(oracle.Far(0x01b1, 0xa000), len(q.prefix))) || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0x711b), 5), []byte{0x2e, 0xff, 0x16, 0x22, 0x20}) {
		q.abort(o)
		return
	}
	var fragment quoteFragment
	q.pendingNumber = nil
	if q.explanation && q.stage >= 3 {
		if q.item == nil {
			q.abort(o)
			return
		}
		address := uint16(int(q.item.base) + q.item.offset)
		item, ok := itemDetailOf(address)
		bounds := item.bounds
		if !ok {
			q.abort(o)
			return
		}
		start, end := bounds[q.stage-3], bounds[q.stage-2]
		fragment = quoteFragment{start: start, raw: o.Bytes(oracle.Far(0x01b1, start), int(end-start))}
		// 原始來源必須與固定 SAR 的對應片段完全相同，不能只接受執行中任意字串。
		if !bytes.Equal(fragment.raw, q.fragments[int(start)].raw) {
			q.abort(o)
			return
		}
		fragment.positions = q.fragments[int(start)].positions
		detail := q.details[address]
		q.texts[1], q.texts[2] = detail[0], detail[1]
	} else {
		switch q.stage {
		case 1:
			if q.item == nil {
				if _, known := itemDetailOf(pos); !known {
					q.abort(o)
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
				q.abort(o)
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
					q.abort(o)
					return
				}
				raw := o.Bytes(oracle.Far(0x01b1, pos), 10)
				end := 0
				for end < len(raw) && raw[end] >= '0' && raw[end] <= '9' {
					end++
				}
				if end < 1 || end > 8 || end >= len(raw) || raw[end] != 0xff {
					q.abort(o)
					return
				}
				number.raw = append([]byte(nil), raw[:end+1]...)
				number.box = image.Rect(int(o.BX()), int(o.CX()&255), int(o.BX())+8, int(o.CX()&255)+8)
			}
			if int(o.CX()&255) != number.box.Min.Y || int(o.BX()) < number.box.Min.X || int(o.BX()) > number.box.Max.X {
				q.abort(o)
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
	}
	if q.index >= len(fragment.positions) || pos != fragment.positions[q.index] || !bytes.Equal(fragment.raw, o.Bytes(oracle.Far(0x01b1, fragment.start), len(fragment.raw))) {
		q.abort(o)
		return
	}
	q.index++
	if q.index == len(fragment.positions) {
		q.stage++
		q.index = 0
	}
	q.pending = true
	// 規格066：文字字形由中文取代；價格等數字保留原版圖形，不認領。
	if q.pendingNumber == nil {
		glyphpatch.Claim(o, 0x7120)
	}
}

func (q *itemPurchase) capture(o *oracle.Oracle) {
	if !q.active || !q.pending {
		return
	}
	raw := glyphpatch.Indexed(o)
	if number := q.pendingNumber; number != nil {
		number.pixels = labelCrop(raw, number.box)
		for _, v := range number.pixels {
			if v != 0 && v != 9 {
				q.abort(o)
				return
			}
		}
	}
	q.captureFrame(o)
	q.pendingNumber = nil
	q.ready = q.stage
	q.Glyphs++
	if q.ready == q.terminal {
		q.Completed++
	}
}

func (q *itemPurchase) captureFrame(o *oracle.Oracle) {
	q.completed = dialogueCrop(glyphpatch.Indexed(o))
	q.palette = o.Palette()
	q.paintSerial = q.paint.serial
	q.pending = false
}

func (q *itemPurchase) tokens() []quoteToken {
	var tokens []quoteToken
	for i, stage := range q.thresholds {
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
		q.abort(q.original)
		return false
	}
	tokens := q.tokens()
	positions, valid := quoteLayout(tokens)
	if !valid {
		q.abort(q.original)
		return false
	}
	for _, token := range tokens {
		if n := token.number; n != nil && (n.box.Empty() || len(n.pixels) != n.box.Dx()*n.box.Dy()) {
			q.abort(q.original)
			return false
		}
	}
	bg, fg := pal[0], pal[9]
	box := image.Rect(origin.X+dialogueBox.Min.X*3, origin.Y+dialogueBox.Min.Y*3, origin.X+dialogueBox.Max.X*3, origin.Y+dialogueBox.Max.Y*3)
	prompt := image.Rectangle{Min: dialoguePrompt.Min.Mul(3), Max: dialoguePrompt.Max.Mul(3)}.Add(origin)
	put := func(p image.Point, c [3]uint8) {
		if p.In(dst.Bounds()) && (!q.prompt || !p.In(prompt)) && visiblePixel(p, origin, masks) {
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
		q.font.DrawGlyph(dst, at, token.ch, 32, color.RGBA{fg[0], fg[1], fg[2], 255}, func(p image.Point) bool { return (!q.prompt || !p.In(prompt)) && visiblePixel(p, origin, masks) })
	}
	return true
}

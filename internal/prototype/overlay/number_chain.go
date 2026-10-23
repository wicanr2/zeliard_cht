package overlay

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
	"github.com/wicanr2/zeliard_cht/internal/prototype/glyphpatch"
)

// numberSentenceSpec 是「文字、原版數字、文字」交錯的單句來源。
// texts 為 SAR2 檔案偏移 [起, 訖)，相鄰兩段之間由原版在 number 暫存位址印一個數字。
// 暫存位址取自各模組在 FF 00 之後設定 FF4C 的指令（規格066）。
type numberSentenceSpec struct {
	module int
	number uint16
	texts  [][2]int
}

var numberSentences = []numberSentenceSpec{
	// 武器店修理盾牌報價：模組 A1C6 設 AEF8 文字，A1E3 設 BC33 數字。
	{0x17817, 0xbc33, [][2]int{{0x18710, 0x1874e}, {0x18750, 0x18768}}},
	// 銀行匯率：A180 設 A9D9 文字，A193／A1B1 兩次設 AD27 數字。
	{0x19462, 0xad27, [][2]int{{0x19e3c, 0x19e51}, {0x19e53, 0x19e5d}, {0x19e5f, 0x19e7e}}},
	// 旅館問候與房價：A046 設 A2F6 文字，A0AD 設 A508 數字。
	{0x1b7b2, 0xa508, [][2]int{{0x1baa9, 0x1bb54}, {0x1bb56, 0x1bb6d}}},
}

type numberSentence struct {
	key    string
	text   string
	prefix []byte
	number uint16
	parts  []quoteFragment
}

// numberChain 只在整句完成後顯示中文；數字沿用原版字形，順序或位置不符就寫回原版字形並退出。
type numberChain struct {
	original                      *oracle.Oracle
	font                          bitmap.Font
	paint                         *paintHistory
	sources                       []*numberSentence
	current                       *numberSentence
	numbers                       []quoteNumber
	part, index                   int
	active, pending, prompt, done bool
	pendingNumber                 *quoteNumber
	completed                     []byte
	palette                       [256][3]uint8
	paintSerial                   uint64
	Glyphs, Completed, Rejections int
}

func newNumberChain(o *oracle.Oracle, data []byte, texts map[string]string, font bitmap.Font, paint *paintHistory) (*numberChain, error) {
	c := &numberChain{original: o, font: font, paint: paint}
	for _, spec := range numberSentences {
		key := fmt.Sprintf("zelres2.file.%06x", spec.texts[0][0])
		text, ok := texts[key]
		if !ok {
			continue
		}
		delete(texts, key)
		plain := text
		for i := 1; i < len(spec.texts); i++ {
			slot := "{" + strconv.Itoa(i) + "}"
			if strings.Count(text, slot) != 1 {
				return nil, fmt.Errorf("數字句欄位不符：%s", key)
			}
			plain = strings.ReplaceAll(plain, slot, "")
		}
		if plain == "" || strings.ContainsAny(plain, "{}\n\r\t") {
			return nil, fmt.Errorf("數字句試譯不相容：%s", key)
		}
		if _, err := dialogueLayout(font, []rune(plain)); err != nil {
			return nil, fmt.Errorf("%s：%w", key, err)
		}
		s := &numberSentence{key: key, text: text, prefix: append([]byte(nil), data[spec.module:spec.module+64]...), number: spec.number}
		for _, r := range spec.texts {
			f := quoteFragment{start: uint16(0xa000 + r[0] - spec.module), raw: append([]byte(nil), data[r[0]:r[1]]...)}
			for i := 0; i < len(f.raw); i++ {
				if f.raw[i] == 0xff {
					i++
					continue
				}
				if f.raw[i] >= 0x20 && f.raw[i] < 0x7f && f.raw[i] != '/' {
					f.positions = append(f.positions, f.start+uint16(i))
				}
			}
			if len(f.positions) == 0 || data[r[1]] != 0xff {
				return nil, fmt.Errorf("數字句來源邊界不符：%s", key)
			}
			s.parts = append(s.parts, f)
		}
		c.sources = append(c.sources, s)
	}
	if len(c.sources) == 0 {
		return nil, nil
	}
	o.OnCall(oracle.Far(0x01b1, 0x711b), c.begin)
	o.OnCall(oracle.Far(0x01b1, 0x7120), c.capture)
	o.OnCall(oracle.Far(0x01b1, 0x7205), func(*oracle.Oracle) { c.reset() })
	o.OnCall(oracle.Far(0x01b1, 0x71c9), func(o *oracle.Oracle) {
		if c.active {
			c.prompt = true
			c.captureFrame(o)
		}
	})
	o.OnCall(oracle.Far(0x01b1, 0x71d9), func(o *oracle.Oracle) {
		if c.active {
			c.prompt = false
			c.captureFrame(o)
		}
	})
	return c, nil
}

func (c *numberChain) reset() {
	c.active, c.pending, c.prompt, c.done = false, false, false, false
	c.current, c.pendingNumber, c.completed = nil, nil, nil
	c.numbers = nil
	c.part, c.index = 0, 0
}

// abort 用於已認領字形後的來源不符：先把壓下的字形照原版寫回，再退出。
func (c *numberChain) abort(o *oracle.Oracle) {
	if c.active {
		glyphpatch.Reveal(o, dialogueBox)
		c.Rejections++
	}
	c.reset()
}

func (c *numberChain) begin(o *oracle.Oracle) {
	pos := o.Word(oracle.Far(0x01b1, 0xff4c)) - 1
	if !c.active || c.done {
		var start *numberSentence
		for _, s := range c.sources {
			if pos == s.parts[0].positions[0] {
				start = s
				break
			}
		}
		// 已完成的句子後面接了別的字：先把壓下的字形寫回，再交給其他覆繪層或原版。
		if start == nil {
			if c.done {
				c.abort(o)
			}
			return
		}
		if c.done {
			c.abort(o)
		}
		c.reset()
		c.active, c.current = true, start
	}
	s := c.current
	if o.DSReg() != 0x01b1 || o.AX()>>8 != 1 || o.BX() < 54 || o.BX()+8 > 264 || o.CX()&255 < 99 || o.CX()&255 > 139 || byte(o.AX()) != o.Byte(oracle.Far(0x01b1, pos)) || !bytes.Equal(s.prefix, o.Bytes(oracle.Far(0x01b1, 0xa000), len(s.prefix))) || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0x711b), 5), []byte{0x2e, 0xff, 0x16, 0x22, 0x20}) {
		c.abort(o)
		return
	}
	var fragment quoteFragment
	c.pendingNumber = nil
	if c.part%2 == 1 {
		if c.index == 0 {
			if pos != s.number {
				c.abort(o)
				return
			}
			raw := o.Bytes(oracle.Far(0x01b1, pos), 10)
			end := 0
			for end < len(raw) && raw[end] >= '0' && raw[end] <= '9' {
				end++
			}
			if end < 1 || end > 8 || end >= len(raw) || raw[end] != 0xff {
				c.abort(o)
				return
			}
			c.numbers = append(c.numbers, quoteNumber{raw: append([]byte(nil), raw[:end+1]...), box: image.Rect(int(o.BX()), int(o.CX()&255), int(o.BX())+8, int(o.CX()&255)+8)})
		}
		n := &c.numbers[len(c.numbers)-1]
		if int(o.CX()&255) != n.box.Min.Y || int(o.BX()) < n.box.Min.X || int(o.BX()) > n.box.Max.X {
			c.abort(o)
			return
		}
		n.box.Max.X = int(o.BX()) + 8
		fragment = quoteFragment{start: s.number, raw: n.raw}
		for i := 0; i < len(n.raw)-1; i++ {
			fragment.positions = append(fragment.positions, s.number+uint16(i))
		}
		c.pendingNumber = n
	} else {
		fragment = s.parts[c.part/2]
	}
	if c.index >= len(fragment.positions) || pos != fragment.positions[c.index] || !bytes.Equal(fragment.raw, o.Bytes(oracle.Far(0x01b1, fragment.start), len(fragment.raw))) {
		c.abort(o)
		return
	}
	c.index++
	if c.index == len(fragment.positions) {
		c.part++
		c.index = 0
	}
	c.pending = true
	// 規格066：文字字形由中文取代；數字保留原版圖形，不認領。
	if c.pendingNumber == nil {
		glyphpatch.Claim(o, 0x7120)
	}
}

func (c *numberChain) captureFrame(o *oracle.Oracle) {
	c.completed = dialogueCrop(glyphpatch.Indexed(o))
	c.palette = o.Palette()
	c.paintSerial = c.paint.serial
}

func (c *numberChain) capture(o *oracle.Oracle) {
	if !c.active || !c.pending || c.current == nil {
		return
	}
	if n := c.pendingNumber; n != nil {
		n.pixels = labelCrop(glyphpatch.Indexed(o), n.box)
		for _, v := range n.pixels {
			if v != 0 && v != 9 {
				c.abort(o)
				return
			}
		}
	}
	c.captureFrame(o)
	c.pending, c.pendingNumber = false, nil
	c.Glyphs++
	if c.part == 2*len(c.current.parts)-1 && !c.done {
		c.done = true
		c.Completed++
	}
}

func (c *numberChain) tokens() []quoteToken {
	var tokens []quoteToken
	if !c.done {
		return nil
	}
	text := c.current.text
	for len(text) > 0 {
		at := strings.IndexByte(text, '{')
		if at < 0 {
			at = len(text)
		}
		for _, ch := range text[:at] {
			tokens = append(tokens, quoteToken{ch: ch, width: len(c.font[ch]) / 16 * 16})
		}
		text = text[at:]
		if text == "" {
			break
		}
		end := strings.IndexByte(text, '}')
		i, _ := strconv.Atoi(text[1:end])
		n := &c.numbers[i-1]
		tokens = append(tokens, quoteToken{number: n, width: n.box.Dx() * 3})
		text = text[end+1:]
	}
	return tokens
}

func (c *numberChain) Draw(dst *image.RGBA, raw []byte, pal [256][3]uint8, origin image.Point) bool {
	if !c.active || len(c.completed) != dialogueBox.Dx()*dialogueBox.Dy() {
		return false
	}
	masks, known := c.paint.masks(c.paintSerial, raw)
	if pal != c.palette || !known || !c.pending && !visibleRowsEqual(c.completed, raw, []image.Rectangle{dialogueBox}, masks) {
		c.abort(c.original)
		return false
	}
	tokens := c.tokens()
	positions, valid := quoteLayout(tokens)
	if c.done && !valid {
		c.reset()
		return false
	}
	for _, t := range tokens {
		if n := t.number; n != nil && (n.box.Empty() || len(n.pixels) != n.box.Dx()*n.box.Dy()) {
			c.reset()
			return false
		}
	}
	box := image.Rect(origin.X+dialogueBox.Min.X*3, origin.Y+dialogueBox.Min.Y*3, origin.X+dialogueBox.Max.X*3, origin.Y+dialogueBox.Max.Y*3)
	prompt := image.Rect(origin.X+dialoguePrompt.Min.X*3, origin.Y+dialoguePrompt.Min.Y*3, origin.X+dialoguePrompt.Max.X*3, origin.Y+dialoguePrompt.Max.Y*3)
	put := func(p image.Point, col [3]uint8) {
		if p.In(dst.Bounds()) && (!c.prompt || !p.In(prompt)) && visiblePixel(p, origin, masks) {
			dst.SetRGBA(p.X, p.Y, color.RGBA{col[0], col[1], col[2], 255})
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
			for y := 0; y < n.box.Dy()*3; y++ {
				for x := 0; x < n.box.Dx()*3; x++ {
					if n.pixels[y/3*n.box.Dx()+x/3] == 9 {
						put(at.Add(image.Pt(x, y+4)), pal[9])
					}
				}
			}
			continue
		}
		col := pal[9]
		c.font.DrawGlyph(dst, at, t.ch, 32, color.RGBA{col[0], col[1], col[2], 255}, func(p image.Point) bool { return (!c.prompt || !p.In(prompt)) && visiblePixel(p, origin, masks) })
	}
	return true
}

func (c *numberChain) State() map[string]any {
	if c == nil {
		return nil
	}
	key := ""
	if c.current != nil {
		key = c.current.key
	}
	return map[string]any{"active": c.active, "key": key, "part": c.part, "done": c.done, "glyphs": c.Glyphs, "completed": c.Completed, "rejections": c.Rejections}
}

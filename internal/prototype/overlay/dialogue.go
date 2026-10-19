package overlay

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

var dialogueBox = image.Rect(54, 98, 264, 150)
var dialoguePrompt = image.Rect(156, 139, 164, 147)

type dialogueParagraph struct {
	key            string
	start, end     uint16
	text           []rune
	progress       map[uint16]int
	total          int
	prefix, script []byte
}

// Dialogue 是已核對來源的唯讀顯示試作。原版處理按鍵、捲頁、交易與存檔。
type Dialogue struct {
	font                        bitmap.Font
	paragraphs                  []dialogueParagraph
	current, visible            int
	lead                        []rune
	active, pending, prompt     bool
	completed                   []byte
	paint                       *paintHistory
	paintSerial                 uint64
	quotation                   *Quotation
	bank                        *bankDialogue
	itemPurchase                *itemPurchase
	itemSale                    *itemPurchase
	itemExplanation             *itemPurchase
	Glyphs, Paragraphs, Scrolls int
	Rejections                  []map[string]int
}

func NewDialogue(o *oracle.Oracle, sar, catalog string, font bitmap.Font) (*Dialogue, error) {
	data, err := os.ReadFile(sar)
	if err != nil {
		return nil, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1" {
		return nil, fmt.Errorf("對話來源版本不符")
	}
	tsv, err := os.ReadFile(catalog)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(tsv) {
		return nil, fmt.Errorf("對話 TSV 必須是 UTF-8")
	}
	r := csv.NewReader(bytes.NewReader(tsv))
	r.Comma = '\t'
	r.FieldsPerRecord = 2
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 || records[0][0] != "key" || records[0][1] != "zh-TW" {
		return nil, fmt.Errorf("對話 TSV 須有 key／zh-TW 欄位")
	}
	texts := map[string]string{}
	for _, record := range records[1:] {
		if texts[record[0]] != "" || record[1] == "" {
			return nil, fmt.Errorf("對話鍵重複或漏譯")
		}
		texts[record[0]] = record[1]
	}
	d := &Dialogue{font: font, current: -1, paint: observePaint(o)}
	// 檔案偏移與模組 payload 起點來自固定 SAR 及正常路徑傾印。
	// 數值、控制碼與存檔動作仍由原版執行。
	for _, source := range [][3]int{{0x17250, 0x172b2, 0x16e1a}, {0x172bb, 0x17351, 0x16e1a}, {0x1ca64, 0x1ca83, 0x1bcc6}, {0x1ca86, 0x1caaf, 0x1bcc6}, {0x1cc43, 0x1cc64, 0x1bcc6}, {0x1cc66, 0x1cca0, 0x1bcc6}, {0x1ce7e, 0x1cef1, 0x1bcc6}, {0x185ea, 0x18606, 0x17817}, {0x18607, 0x18633, 0x17817}, {0x18634, 0x18661, 0x17817}, {0x1883e, 0x1885d, 0x17817}, {0x18899, 0x188b8, 0x17817}, {0x189f6, 0x18a16, 0x17817}, {0x1876b, 0x187c6, 0x17817}, {0x18a17, 0x18a57, 0x17817}, {0x1adf3, 0x1ae11, 0x1a588}, {0x1ae15, 0x1ae2e, 0x1a588}, {0x1aeee, 0x1af12, 0x1a588}, {0x1b099, 0x1b0bd, 0x1a588}, {0x1a0bd, 0x1a0ff, 0x19462}, {0x1b02f, 0x1b04f, 0x1a588}, {0x1b072, 0x1b096, 0x1a588}, {0x1b002, 0x1b02e, 0x1a588}, {0x1af16, 0x1af36, 0x1a588}, {0x1af36, 0x1af4c, 0x1a588}, {0x1d72e, 0x1d76c, 0x1bcc6}, {0x1d76c, 0x1d79c, 0x1bcc6}, {0x1d79c, 0x1d7b2, 0x1bcc6}, {0x1d7b2, 0x1d7d5, 0x1bcc6}, {0x19f95, 0x19fe3, 0x19462}, {0x1cab2, 0x1cace, 0x1bcc6}} {
		key := fmt.Sprintf("zelres2.file.%06x", source[0])
		text, ok := texts[key]
		if !ok {
			continue
		}
		delete(texts, key)
		if text == "" || strings.ContainsAny(text, "\n\r\t") {
			return nil, fmt.Errorf("對話鍵或文字不符：%s", key)
		}
		if _, err := dialogueLayout(font, []rune(text)); err != nil {
			return nil, err
		}
		start, end := uint16(0xa000+source[0]-source[2]), uint16(0xa000+source[1]-source[2])
		p := dialogueParagraph{key: key, start: start, end: end, text: []rune(text), progress: map[uint16]int{}, prefix: append([]byte(nil), data[source[2]:source[2]+64]...), script: append([]byte(nil), data[source[0]:source[1]]...)}
		for pos := start; pos < end; pos++ {
			ch := p.script[int(pos-start)]
			if ch == 0xff {
				pos++
				continue
			}
			if ch >= 0x20 && ch < 0x7f && ch != '/' {
				p.total++
				p.progress[pos] = p.total
			}
		}
		d.paragraphs = append(d.paragraphs, p)
	}
	if err = d.validateRetainedDialogue(); err != nil {
		return nil, err
	}
	d.quotation, err = newQuotation(o, data, texts, font, d.paint)
	if err != nil {
		return nil, err
	}
	d.bank, err = newBankDialogue(o, data, texts, font, d.paint)
	if err != nil {
		return nil, err
	}
	d.itemPurchase, err = newItemPurchase(o, data, texts, font, d.paint)
	if err != nil {
		return nil, err
	}
	d.itemSale, err = newItemSale(o, data, texts, font, d.paint)
	if err != nil {
		return nil, err
	}
	d.itemExplanation, err = newItemExplanation(o, data, texts, font, d.paint)
	if err != nil {
		return nil, err
	}
	if len(texts) != 0 {
		return nil, fmt.Errorf("對話 catalog 含未知來源鍵")
	}
	o.OnCall(oracle.Far(0x01b1, 0x711b), d.beginGlyph)
	o.OnCall(oracle.Far(0x01b1, 0x7120), func(o *oracle.Oracle) {
		if d.pending {
			d.capture(o)
			d.Glyphs++
		}
	})
	o.OnCall(oracle.Far(0x01b1, 0x7205), func(*oracle.Oracle) { d.reset() })
	o.OnCall(oracle.Far(0x01b1, 0x71a7), func(*oracle.Oracle) {
		if d.active {
			d.pending = true
		}
	})
	o.OnCall(oracle.Far(0x01b1, 0x71ac), func(o *oracle.Oracle) {
		if d.active {
			d.capture(o)
			d.Scrolls++
		}
	})
	o.OnCall(oracle.Far(0x01b1, 0x71c9), func(o *oracle.Oracle) {
		if d.active {
			d.prompt = true
			d.capture(o)
		}
	})
	o.OnCall(oracle.Far(0x01b1, 0x71d9), func(o *oracle.Oracle) {
		if d.active {
			d.prompt = false
			d.capture(o)
		}
	})
	o.OnCall(oracle.Far(0x01b1, 0xa068), func(o *oracle.Oracle) {
		if d.active && bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0xa068), 5), []byte{0x2e, 0xff, 0x26, 0x40, 0x20}) {
			d.reset()
		}
	})
	return d, nil
}

func (d *Dialogue) reset() {
	d.active = false
	d.pending = false
	d.prompt = false
	d.completed = nil
	d.current = -1
	d.visible = 0
	d.lead = nil
}
func (d *Dialogue) beginGlyph(o *oracle.Oracle) {
	pos := o.Word(oracle.Far(0x01b1, 0xff4c)) - 1
	idx := -1
	count := 0
	for i, p := range d.paragraphs {
		if n, ok := p.progress[pos]; ok && bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0xa000), len(p.prefix)), p.prefix) && bytes.Equal(o.Bytes(oracle.Far(0x01b1, p.start), len(p.script)), p.script) {
			idx = i
			count = n
			break
		}
	}
	// 7118h 只更新 CL。正常翻頁可帶回 CH=2，不能把整個 CX 當作 Y。
	y := o.CX() & 0xff
	if idx < 0 || o.DSReg() != 0x01b1 || o.AX()>>8 != 1 || int(o.BX()) < 54 || int(o.BX()) >= 264 || y < 99 || y > 139 || byte(o.AX()) != o.Byte(oracle.Far(0x01b1, pos)) || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0x711b), 5), []byte{0x2e, 0xff, 0x16, 0x22, 0x20}) {
		if (d.active || idx >= 0) && len(d.Rejections) < 16 {
			d.Rejections = append(d.Rejections, map[string]int{"kind": 1, "steps": int(o.Steps()), "source": int(pos), "index": idx, "ax": int(o.AX()), "bx": int(o.BX()), "cx": int(o.CX()), "ds": int(o.DSReg())})
		}
		d.reset()
		return
	}
	if d.current != idx {
		lead, allowed := d.retainedDialogueLead(idx)
		if !allowed {
			if len(d.Rejections) < 16 {
				d.Rejections = append(d.Rejections, map[string]int{"kind": 3, "steps": int(o.Steps()), "source": int(pos), "index": idx})
			}
			d.reset()
			return
		}
		d.lead = lead
		d.Paragraphs++
		d.current = idx
		d.prompt = false
	}
	d.active = true
	d.pending = true
	p := d.paragraphs[idx]
	d.visible = (count*len(p.text) + p.total - 1) / p.total
}

func dialogueCrop(indexed []byte) []byte {
	if len(indexed) != 320*200 {
		return nil
	}
	result := make([]byte, dialogueBox.Dx()*dialogueBox.Dy())
	for y := dialogueBox.Min.Y; y < dialogueBox.Max.Y; y++ {
		copy(result[(y-dialogueBox.Min.Y)*dialogueBox.Dx():], indexed[y*320+dialogueBox.Min.X:y*320+dialogueBox.Max.X])
	}
	return result
}
func (d *Dialogue) capture(o *oracle.Oracle) {
	d.completed = dialogueCrop(o.Indexed())
	d.paintSerial = d.paint.serial
	d.pending = false
}

// dialogueLayout 用實際字模寬度決定三行版面。超寬／缺字直接拒絕，不靠裁切藏字。
func dialogueLayout(font bitmap.Font, text []rune) ([]image.Point, error) {
	x, y := 6, 4
	points := make([]image.Point, 0, len(text))
	noLineStart := func(r rune) bool { return strings.ContainsRune("，。！？；：、）］」』】〉》…", r) }
	for i, ch := range text {
		glyph := font[ch]
		if len(glyph) != 16 && len(glyph) != 32 {
			return nil, fmt.Errorf("對話字型缺字：%U", ch)
		}
		w := len(glyph) / 16 * 16
		// 拉丁專名與數字整組換行，避免 Jashiin 被拆成兩行。
		latin := func(r rune) bool { return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' }
		if latin(ch) && (i == 0 || !latin(text[i-1])) {
			wordWidth := 0
			j := i
			for ; j < len(text) && latin(text[j]); j++ {
				wordWidth += len(font[text[j]]) / 16 * 16
			}
			for ; j < len(text) && noLineStart(text[j]); j++ {
				wordWidth += len(font[text[j]]) / 16 * 16
			}
			if wordWidth > dialogueBox.Dx()*3-12 {
				return nil, fmt.Errorf("對話專名超出安全寬度")
			}
			if x+wordWidth > dialogueBox.Dx()*3-6 {
				x = 6
				y += 36
			}
		}
		// 保留下一串句尾標點的位置，避免逗號或句號單獨落在行首。
		needed := w
		for j := i + 1; j < len(text) && noLineStart(text[j]); j++ {
			needed += len(font[text[j]]) / 16 * 16
		}
		if x+needed > dialogueBox.Dx()*3-6 && !noLineStart(ch) || x+w > dialogueBox.Dx()*3-6 {
			x = 6
			y += 36
		}
		if y+32 > 112 {
			return nil, fmt.Errorf("對話試譯超出三行安全區")
		}
		points = append(points, image.Pt(x, y))
		x += w
	}
	return points, nil
}

func (d *Dialogue) Active() bool {
	return d.active || d.quotation != nil && d.quotation.active || d.bank != nil && d.bank.active || d.itemPurchase != nil && d.itemPurchase.active || d.itemSale != nil && d.itemSale.active || d.itemExplanation != nil && d.itemExplanation.active
}

// UseLabels 共用已驗證的商品來源與譯名，不複製另一份商品翻譯。
func (d *Dialogue) UseLabels(l *Labels) {
	if d.quotation != nil {
		d.quotation.labels = l
	}
	if d.itemPurchase != nil {
		d.itemPurchase.labels = l
	}
	if d.itemSale != nil {
		d.itemSale.labels = l
	}
	if d.itemExplanation != nil {
		d.itemExplanation.labels = l
	}
}

func (d *Dialogue) QuotationState() map[string]any {
	q := d.quotation
	if q == nil {
		return nil
	}
	price, trade := "", ""
	if len(q.price.raw) > 0 {
		price = string(q.price.raw[:len(q.price.raw)-1])
	}
	if len(q.trade.raw) > 0 {
		trade = string(q.trade.raw[:len(q.trade.raw)-1])
	}
	return map[string]any{"active": q.active, "stage": q.stage, "ready": q.ready, "glyphs": q.Glyphs, "completed": q.Completed, "price": price, "trade": trade}
}

func (d *Dialogue) BankState() map[string]any {
	if d.bank == nil {
		return nil
	}
	b := d.bank
	parts := []map[string]string{}
	for _, p := range b.parts {
		number := ""
		if len(p.number.raw) > 0 {
			number = string(p.number.raw[:len(p.number.raw)-1])
		}
		parts = append(parts, map[string]string{"key": p.source.key, "amount": number})
	}
	current := ""
	if b.current != nil {
		current = b.current.source.key
	}
	return map[string]any{"active": b.active, "parts": parts, "current": current, "phase": b.phase, "glyphs": b.Glyphs, "completed": b.Completed}
}

func (d *Dialogue) Draw(dst *image.RGBA, indexed []byte, palette [256][3]uint8, origin image.Point) bool {
	if d.itemExplanation != nil && d.itemExplanation.active {
		return d.itemExplanation.Draw(dst, indexed, palette, origin)
	}
	if d.itemSale != nil && d.itemSale.active {
		return d.itemSale.Draw(dst, indexed, palette, origin)
	}
	if d.itemPurchase != nil && d.itemPurchase.active {
		return d.itemPurchase.Draw(dst, indexed, palette, origin)
	}
	if d.bank != nil && d.bank.active {
		return d.bank.Draw(dst, indexed, palette, origin)
	}
	if d.quotation != nil && d.quotation.active {
		return d.quotation.Draw(dst, indexed, palette, origin)
	}
	if !d.active || len(d.completed) != dialogueBox.Dx()*dialogueBox.Dy() {
		return false
	}
	masks, known := d.paint.masks(d.paintSerial, indexed)
	valid := known && visibleRowsEqual(d.completed, indexed, []image.Rectangle{dialogueBox}, masks)
	if !d.pending && !valid {
		if len(d.Rejections) < 16 {
			actual := dialogueCrop(indexed)
			first := -1
			for i, v := range actual {
				if v != d.completed[i] {
					first = i
					break
				}
			}
			d.Rejections = append(d.Rejections, map[string]int{"kind": 2, "glyphs": d.Glyphs, "paragraph": d.current, "visible": d.visible, "first_difference": first})
		}
		d.reset()
		return false
	}
	p := d.paragraphs[d.current]
	text := append(append([]rune(nil), d.lead...), p.text...)
	positions, err := dialogueLayout(d.font, text)
	if err != nil {
		d.reset()
		return false
	}
	box := image.Rect(origin.X+dialogueBox.Min.X*3, origin.Y+dialogueBox.Min.Y*3, origin.X+dialogueBox.Max.X*3, origin.Y+dialogueBox.Max.Y*3)
	prompt := image.Rect(origin.X+dialoguePrompt.Min.X*3, origin.Y+dialoguePrompt.Min.Y*3, origin.X+dialoguePrompt.Max.X*3, origin.Y+dialoguePrompt.Max.Y*3)
	bg, fg := palette[0], palette[9]
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			p := image.Pt(x, y)
			if p.In(dst.Bounds()) && (!d.prompt || !p.In(prompt)) && visiblePixel(p, origin, masks) {
				dst.SetRGBA(x, y, color.RGBA{bg[0], bg[1], bg[2], 255})
			}
		}
	}
	for i, ch := range text {
		if i >= len(d.lead)+d.visible {
			break
		}
		g := d.font[ch]
		w := len(g) / 16 * 8
		at := positions[i].Add(box.Min)
		for y := 0; y < 16; y++ {
			for x := 0; x < w; x++ {
				if g[y*(w/8)+x/8]&(0x80>>uint(x%8)) != 0 {
					for yy := 0; yy < 2; yy++ {
						for xx := 0; xx < 2; xx++ {
							point := image.Pt(at.X+x*2+xx, at.Y+y*2+yy)
							if point.In(dst.Bounds()) && visiblePixel(point, origin, masks) {
								dst.SetRGBA(point.X, point.Y, color.RGBA{fg[0], fg[1], fg[2], 255})
							}
						}
					}
				}
			}
		}
	}
	return true
}

func (d *Dialogue) ItemPurchaseState() map[string]any {
	return itemChainState(d.itemPurchase)
}

func (d *Dialogue) ItemSaleState() map[string]any        { return itemChainState(d.itemSale) }
func (d *Dialogue) ItemExplanationState() map[string]any { return itemChainState(d.itemExplanation) }

func itemChainState(q *itemPurchase) map[string]any {
	if q == nil {
		return nil
	}
	price := ""
	if len(q.price.raw) > 0 {
		price = string(q.price.raw[:len(q.price.raw)-1])
	}
	item := ""
	if q.item != nil {
		item = string(q.item.text)
	}
	return map[string]any{"active": q.active, "stage": q.stage, "ready": q.ready, "glyphs": q.Glyphs, "completed": q.Completed, "price": price, "item": item, "prompt": q.prompt}
}

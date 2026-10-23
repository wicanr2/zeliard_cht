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
	"github.com/wicanr2/zeliard_cht/internal/prototype/glyphpatch"
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
	original                    *oracle.Oracle
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
	numbers                     *numberChain
	Glyphs, Paragraphs, Scrolls int
	Rejections                  []map[string]int
}

// dialogueSources 是 706C 對話段落的 SAR2 檔案偏移 {起, 訖, 模組 payload 起點}。
// 起點為第一個可印字元；FF 05 FF 02 FF 04 接續同段，其他 FF、0D、11 結束一段。
// 數值、控制碼與存檔動作仍由原版執行。
var dialogueSources = [][3]int{
	// 1.0.5 以前已接入的正常路徑來源。
	{0x17250, 0x172b2, 0x16e1a}, {0x172bb, 0x17351, 0x16e1a}, {0x1ca64, 0x1ca83, 0x1bcc6}, {0x1ca86, 0x1caaf, 0x1bcc6}, {0x1cc43, 0x1cc64, 0x1bcc6}, {0x1cc66, 0x1cca0, 0x1bcc6}, {0x1ce7e, 0x1cef1, 0x1bcc6}, {0x185ea, 0x18606, 0x17817}, {0x18607, 0x18633, 0x17817}, {0x18634, 0x18661, 0x17817}, {0x1883e, 0x1885d, 0x17817}, {0x18899, 0x188b8, 0x17817}, {0x189f6, 0x18a16, 0x17817}, {0x1876b, 0x187c6, 0x17817}, {0x18a17, 0x18a57, 0x17817}, {0x1adf3, 0x1ae11, 0x1a588}, {0x1ae15, 0x1ae2e, 0x1a588}, {0x1aeee, 0x1af12, 0x1a588}, {0x1b099, 0x1b0bd, 0x1a588}, {0x1a0bd, 0x1a0ff, 0x19462}, {0x1b02f, 0x1b04f, 0x1a588}, {0x1b072, 0x1b096, 0x1a588}, {0x1b002, 0x1b02e, 0x1a588}, {0x1af16, 0x1af36, 0x1a588}, {0x1af36, 0x1af4c, 0x1a588}, {0x1d72e, 0x1d76c, 0x1bcc6}, {0x1d76c, 0x1d79c, 0x1bcc6}, {0x1d79c, 0x1d7b2, 0x1bcc6}, {0x1d7b2, 0x1d7d5, 0x1bcc6}, {0x19f95, 0x19fe3, 0x19462}, {0x1cab2, 0x1cace, 0x1bcc6}, {0x1a453, 0x1a490, 0x1a19e}, {0x1a491, 0x1a4cd, 0x1a19e}, {0x1a4cd, 0x1a508, 0x1a19e}, {0x1a508, 0x1a568, 0x1a19e}, {0x1a568, 0x1a582, 0x1a19e},
	// 規格066補譯：國王二訪、三訪與結局前。
	{0x1735d, 0x17382, 0x16e1a}, {0x17389, 0x173be, 0x16e1a}, {0x173c3, 0x173e7, 0x16e1a}, {0x173f3, 0x17466, 0x16e1a},
	{0x1746d, 0x174d6, 0x16e1a}, {0x174e0, 0x1754e, 0x16e1a}, {0x17555, 0x175b2, 0x16e1a},
	// 長老查看力量、七位長老傳授法術、聆聽知識與存檔錯誤。
	{0x1cacf, 0x1cb04, 0x1bcc6}, {0x1cb09, 0x1cb6b, 0x1bcc6}, {0x1cb6e, 0x1cbc7, 0x1bcc6}, {0x1cbca, 0x1cc40, 0x1bcc6},
	{0x1cca7, 0x1cce3, 0x1bcc6}, {0x1ccf9, 0x1cd2d, 0x1bcc6}, {0x1cd2f, 0x1cd53, 0x1bcc6}, {0x1cd55, 0x1cda6, 0x1bcc6},
	{0x1cda8, 0x1cddf, 0x1bcc6}, {0x1cde2, 0x1cdff, 0x1bcc6}, {0x1ce05, 0x1ce7c, 0x1bcc6},
	{0x1cef3, 0x1cf61, 0x1bcc6}, {0x1cf65, 0x1cfd9, 0x1bcc6}, {0x1cfdd, 0x1d04e, 0x1bcc6}, {0x1d052, 0x1d0c2, 0x1bcc6},
	{0x1d0c6, 0x1d14a, 0x1bcc6}, {0x1d14e, 0x1d1e0, 0x1bcc6}, {0x1d1e4, 0x1d26e, 0x1bcc6}, {0x1d272, 0x1d2b0, 0x1bcc6},
	{0x1d2c2, 0x1d333, 0x1bcc6}, {0x1d337, 0x1d3ae, 0x1bcc6}, {0x1d3b2, 0x1d430, 0x1bcc6}, {0x1d434, 0x1d4df, 0x1bcc6},
	{0x1d4e3, 0x1d575, 0x1bcc6}, {0x1d579, 0x1d617, 0x1bcc6}, {0x1d61b, 0x1d672, 0x1bcc6}, {0x1d676, 0x1d72a, 0x1bcc6},
	// 武器店盾牌修理、商品說明、榮耀徽章交換與缺貨。
	{0x18662, 0x186c6, 0x17817}, {0x186c9, 0x1870d, 0x17817}, {0x187c7, 0x187ef, 0x17817}, {0x187f8, 0x1883b, 0x17817},
	{0x18937, 0x18993, 0x17817}, {0x18996, 0x189bf, 0x17817}, {0x189c0, 0x189f4, 0x17817}, {0x18a64, 0x18ab7, 0x17817},
	{0x18ad7, 0x18b4b, 0x17817}, {0x18b4e, 0x18b89, 0x17817}, {0x18b8d, 0x18bcb, 0x17817}, {0x18bcd, 0x18bf2, 0x17817},
	{0x18c0d, 0x18cb2, 0x17817}, {0x18cb6, 0x18d79, 0x17817}, {0x18d7d, 0x18ddc, 0x17817}, {0x18de0, 0x18e5f, 0x17817},
	{0x18e63, 0x18eee, 0x17817}, {0x18ef2, 0x18f1f, 0x17817}, {0x18f23, 0x18fce, 0x17817}, {0x18fd2, 0x19052, 0x17817},
	{0x19056, 0x190dd, 0x17817}, {0x190e1, 0x1916b, 0x17817}, {0x1916f, 0x19222, 0x17817}, {0x19226, 0x192ba, 0x17817},
	// 銀行不含數字的回覆。
	{0x19e80, 0x19ea8, 0x19462}, {0x19eab, 0x19ee2, 0x19462}, {0x19ee5, 0x19f01, 0x19462}, {0x19f04, 0x19f2a, 0x19462},
	{0x19f73, 0x19f90, 0x19462}, {0x1a023, 0x1a03f, 0x19462}, {0x1a041, 0x1a057, 0x19462}, {0x1a098, 0x1a0ba, 0x19462},
	{0x1a137, 0x1a17b, 0x19462},
	// 道具店與旅館不含數字的回覆。
	{0x1ae31, 0x1ae4a, 0x1a588}, {0x1aeb0, 0x1aec7, 0x1a588}, {0x1aec8, 0x1aeeb, 0x1a588}, {0x1af87, 0x1afd1, 0x1a588},
	{0x1afd3, 0x1b000, 0x1a588},
	{0x1bb70, 0x1bbc9, 0x1b7b2}, {0x1bbcd, 0x1bc07, 0x1b7b2}, {0x1bc09, 0x1bc32, 0x1b7b2}, {0x1bc36, 0x1bc57, 0x1b7b2},
	{0x1bc60, 0x1bcb4, 0x1b7b2},
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
	d := &Dialogue{original: o, font: font, current: -1, paint: observePaint(o)}
	for _, source := range dialogueSources {
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
			return nil, fmt.Errorf("%s：%w", key, err)
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
	d.numbers, err = newNumberChain(o, data, texts, font, d.paint)
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
		if d.active && idx >= 0 {
			// 規格066：認得的段落幾何或來源不符而拒收時，已壓下的字形照原版寫回。
			glyphpatch.Reveal(o, dialogueBox)
		}
		if (d.active || idx >= 0) && len(d.Rejections) < 16 {
			d.Rejections = append(d.Rejections, map[string]int{"kind": 1, "steps": int(o.Steps()), "source": int(pos), "index": idx, "ax": int(o.AX()), "bx": int(o.BX()), "cx": int(o.CX()), "ds": int(o.DSReg())})
		}
		d.reset()
		return
	}
	if d.current != idx {
		lead, allowed := d.retainedDialogueLead(idx)
		if !allowed {
			if d.active {
				glyphpatch.Reveal(o, dialogueBox)
			}
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
	// 規格066：已核對的段落字形由中文取代，原版不畫進畫面。
	glyphpatch.Claim(o, 0x7120)
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
	d.completed = dialogueCrop(glyphpatch.Indexed(o))
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
	return d.active || d.numbers != nil && d.numbers.active || d.quotation != nil && d.quotation.active || d.bank != nil && d.bank.active || d.itemPurchase != nil && d.itemPurchase.active || d.itemSale != nil && d.itemSale.active || d.itemExplanation != nil && d.itemExplanation.active
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
	if d.numbers != nil && d.numbers.active {
		return d.numbers.Draw(dst, indexed, palette, origin)
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
		if d.original != nil {
			glyphpatch.Reveal(d.original, dialogueBox)
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
		at := positions[i].Add(box.Min)
		d.font.DrawGlyph(dst, at, ch, 32, color.RGBA{fg[0], fg[1], fg[2], 255}, func(p image.Point) bool { return visiblePixel(p, origin, masks) })
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

func (d *Dialogue) NumberState() map[string]any { return d.numbers.State() }

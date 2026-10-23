package overlay

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
	"github.com/wicanr2/zeliard_cht/internal/prototype/glyphpatch"
)

// 結局製作名單（enddemo 66CD 直譯器，規格067）。指標 6965 由 66BF 設為 787E（SAR2 檔案 0x45DFA），
// 到 FF 結束（0x462A6）。控制碼：F8+字組設等待、FA+byte設打字間隔、FB+字組設列（低byte）欄（高byte）、
// FC 清 y=140..199、FD 換列、09 跳到下一個4的倍數欄、F7/F9 等待、FE 換下一張圖、FF 結束。
// 每字以 AH=7（色號119）畫在 x=欄*8、y=列*14+144，畫完在下一格畫游標方塊。
const (
	creditsFileStart, creditsFileEnd = 0x45dfa, 0x462a7
	creditsModule                    = 0x4457c
	creditsEntries, creditsChars     = 60, 868
	creditsRegionTop                 = 140
	creditsRegionSize                = 60 * 320
)

type creditsEntry struct {
	key       string
	files     []int
	screen    int
	row, col  int
	text      []rune
	positions []image.Point
}

type creditsChar struct{ entry, index int }

// parseCredits 依原版直譯器的規則把名單分成「畫面」（兩次 FC 之間）與「字段」（一段連續可見字元，
// 開頭空白不計）。控制碼仍由原版執行，這裡只用來對應位置與譯文。
func parseCredits(data []byte) ([]creditsEntry, map[int]creditsChar, error) {
	if len(data) < creditsFileEnd || data[creditsFileEnd-1] != 0xff {
		return nil, nil, fmt.Errorf("名單來源長度或結尾不符")
	}
	var entries []creditsEntry
	chars := map[int]creditsChar{}
	screen, row, col := 0, 0, 0
	var cur *creditsEntry
	end := func() {
		if cur != nil {
			entries = append(entries, *cur)
			cur = nil
		}
	}
	for file := creditsFileStart; file < creditsFileEnd; file++ {
		b := data[file]
		switch {
		case b == 0xf8 || b == 0xfb:
			end()
			if b == 0xfb {
				row, col = int(data[file+1]), int(data[file+2])
			}
			file += 2
		case b == 0xfa:
			end()
			file++
		case b == 0xfc:
			end()
			screen++
			row, col = 0, 0
		case b == 0xfd:
			end()
			row++
			col = 0
		case b == 0x09:
			end()
			col = (col + 4) &^ 3
		case b >= 0x20 && b < 0x7f:
			if cur == nil && b == ' ' {
				col++
				continue
			}
			if cur == nil {
				cur = &creditsEntry{key: fmt.Sprintf("zelres2.file.%06x", file), screen: screen, row: row, col: col}
			}
			chars[file] = creditsChar{len(entries), len(cur.files)}
			cur.files = append(cur.files, file)
			col++
		default: // F7、F9、FE、FF 不改位置。
			end()
		}
		if row > 3 || col > 40 {
			return nil, nil, fmt.Errorf("名單位置超出原版文字區：0x%x", file)
		}
	}
	end()
	if len(entries) != creditsEntries || len(chars) != creditsChars {
		return nil, nil, fmt.Errorf("名單來源數量不符：%d字段、%d字", len(entries), len(chars))
	}
	return entries, chars, nil
}

// loadCreditsText 讀譯文並以實際字模排版：中文從原版字段的起點畫，24像素字，不得碰到同一列下一個字段的起點或畫面右緣。
func loadCreditsText(entries []creditsEntry, data []byte, font bitmap.Font) error {
	if !utf8.Valid(data) {
		return fmt.Errorf("名單譯表須為UTF-8")
	}
	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = '\t'
	r.FieldsPerRecord = 2
	rows, err := r.ReadAll()
	if err != nil {
		return err
	}
	if len(rows) != len(entries)+1 || rows[0][0] != "key" || rows[0][1] != "zh-TW" {
		return fmt.Errorf("名單須含%d字段及key／zh-TW欄位", len(entries))
	}
	for i := range entries {
		e := &entries[i]
		if rows[i+1][0] != e.key {
			return fmt.Errorf("名單鍵或順序不符：%s", rows[i+1][0])
		}
		text := strings.TrimSpace(rows[i+1][1])
		if text == "" {
			return fmt.Errorf("名單譯文不可空白：%s", e.key)
		}
		e.text = []rune(text)
		limit := 960
		for _, o := range entries {
			if o.screen == e.screen && o.row == e.row && o.col > e.col {
				limit = min(limit, o.col*24-12)
			}
		}
		x, y := e.col*24, (e.row*14+144)*3
		e.positions = nil
		for _, ch := range e.text {
			g := font[ch]
			if unicode.IsControl(ch) || len(g) != 16 && len(g) != 32 {
				return fmt.Errorf("名單缺字或控制碼：%U", ch)
			}
			e.positions = append(e.positions, image.Pt(x, y))
			x += len(g) / 16 * 12
		}
		if x > limit {
			return fmt.Errorf("名單譯文超出安全寬度：%s 寬到%d、上限%d", e.key, x, limit)
		}
	}
	return nil
}

// EndingCredits 只讀原版的名單字元、清框與換圖，中文依已畫字數逐字出現。
type EndingCredits struct {
	o                         *oracle.Oracle
	font                      bitmap.Font
	pool                      []byte
	entries                   []creditsEntry
	chars                     map[int]creditsChar
	done                      []int
	nextFile, lastFile        int
	lastByte                  byte
	screen, row, col          int
	tracking, ready, finished bool
	// expected 是文字區應有的色號；pending 1 字形、2 清框、3 換圖。
	expected, full     []byte
	pending            byte
	Glyphs, Rejected   int
	lastError          string
	pendingX, pendingY int
}

func NewEndingCredits(o *oracle.Oracle, sar, catalog string, font bitmap.Font) (*EndingCredits, error) {
	if o == nil {
		return nil, fmt.Errorf("名單缺少原版執行器")
	}
	data, err := os.ReadFile(sar)
	if err != nil {
		return nil, err
	}
	if threatHash(data) != endingStoryProfile.sarSHA {
		return nil, fmt.Errorf("名單來源版本不符")
	}
	entries, chars, err := parseCredits(data)
	if err != nil {
		return nil, err
	}
	text, err := os.ReadFile(catalog)
	if err != nil {
		return nil, err
	}
	if err = loadCreditsText(entries, text, font); err != nil {
		return nil, err
	}
	c := &EndingCredits{o: o, font: font, pool: append([]byte(nil), data[creditsModule:creditsFileEnd]...), entries: entries, chars: chars, done: make([]int, len(entries))}
	o.OnCall(oracle.Far(0x1b1, 0x66db), c.readByte)
	o.OnCall(oracle.Far(0x1b1, 0x673f), c.beginGlyph)
	o.OnCall(oracle.Far(0x1b1, 0x6744), c.endGlyph)
	o.OnCall(oracle.Far(0x1b1, 0x67b5), c.beginClear)
	o.OnCall(oracle.Far(0x1b1, 0x67ba), c.endClear)
	o.OnCall(oracle.Far(0x1b1, 0x6815), c.beginPicture)
	o.OnCall(oracle.Far(0x1b1, 0x6819), c.endPicture)
	return c, nil
}

func creditsRegion(raw []byte) []byte {
	if len(raw) != 64000 {
		return nil
	}
	return raw[creditsRegionTop*320:]
}

// source 核對結局模組整段（程式與名單文字）仍在 01B1:6000，且為片頭模式與同一字型。
func (c *EndingCredits) source() ([]byte, bool) {
	if c.o.VideoMode() != 0x13 {
		return nil, false
	}
	p := glyphpatch.CodeView(c.o, oracle.Far(0x1b1, 0), 65536)
	if p[0xff77] != 255 || storyWord(p, 0xf500) != 0xf506 {
		return nil, false
	}
	// 程式區 6000..6630 與名單文字逐 byte 相同；6630 起的變數區不比。
	if !bytes.Equal(p[0x6000:0x6630], c.pool[:0x630]) || !bytes.Equal(p[0x6000+creditsFileStart-creditsModule:0x6000+creditsFileEnd-creditsModule], c.pool[creditsFileStart-creditsModule:]) {
		return nil, false
	}
	if !bytes.Equal(p[0x66cd:0x6830], c.pool[0x6cd:0x830]) {
		return nil, false
	}
	for _, g := range storyResidentGuards {
		if threatHash(p[g.lo:g.hi]) != g.sha {
			return nil, false
		}
	}
	return p, true
}

func (c *EndingCredits) halt(reason string) {
	if c.tracking && !c.finished && reason != "" {
		c.Rejected++
		c.lastError = reason
	}
	c.tracking, c.ready, c.pending = false, false, 0
	c.expected, c.full = nil, nil
}

func (c *EndingCredits) readByte(o *oracle.Oracle) {
	p, ok := c.source()
	if !ok || o.DSReg() != 0x1b1 {
		c.halt("byte source")
		return
	}
	ptr := storyWord(p, 0x6965)
	file := int(ptr) - 1 - 0x6000 + creditsModule
	if file == creditsFileStart {
		c.tracking, c.ready, c.finished, c.pending = true, false, false, 0
		c.nextFile, c.screen, c.row, c.col = creditsFileStart, 0, 0, 0
		c.expected, c.full = nil, nil
		c.done = make([]int, len(c.entries))
		c.Glyphs, c.Rejected, c.lastError = 0, 0, ""
	}
	if !c.tracking {
		return
	}
	value := byte(o.Regs().AX)
	if file != c.nextFile || file >= creditsFileEnd || value != p[int(ptr)-1] || o.SI() != ptr || c.pending != 0 {
		c.halt("byte order")
		return
	}
	c.lastFile, c.lastByte = file, value
	c.nextFile++
	switch value {
	case 0xf8:
		c.nextFile += 2
	case 0xfb:
		c.row, c.col = int(p[int(ptr)]), int(p[int(ptr)+1])
		c.nextFile += 2
	case 0xfa:
		c.nextFile++
	case 0xfd:
		c.row++
		c.col = 0
	case 0x09:
		c.col = (c.col + 4) &^ 3
	case 0xff:
		c.finished = true
	}
}

func (c *EndingCredits) beginGlyph(o *oracle.Oracle) {
	if !c.tracking {
		return
	}
	p, ok := c.source()
	r := o.Regs()
	x, y := int(r.BX), int(byte(r.CX))
	if !ok || !c.ready || c.pending != 0 || c.lastByte < 32 || c.lastByte >= 127 || byte(r.AX) != c.lastByte || byte(r.AX>>8) != 7 || x != c.col*8 || y != c.row*14+144 {
		c.halt("glyph start")
		return
	}
	if ch, known := c.chars[c.lastFile]; c.lastByte != ' ' && (!known || c.entries[ch.entry].screen != c.screen || c.done[ch.entry] != ch.index) {
		c.halt("glyph identity")
		return
	}
	c.full = append([]byte(nil), c.expected...)
	glyph := p[0xf506+int(c.lastByte-32)*8 : 0xf506+int(c.lastByte-32)*8+8]
	for gy, bits := range glyph {
		for gx := 0; gx < 8; gx++ {
			if bits&(128>>gx) != 0 {
				at := (y+gy-creditsRegionTop)*320 + x + gx
				if at < 0 || at >= creditsRegionSize {
					c.halt("glyph bounds")
					return
				}
				c.full[at] = 119
			}
		}
	}
	c.pending, c.pendingX, c.pendingY = 1, x, y
}

func (c *EndingCredits) endGlyph(o *oracle.Oracle) {
	if !c.tracking || c.pending != 1 {
		return
	}
	if _, ok := c.source(); !ok || !c.matches(creditsRegion(glyphpatch.Indexed(o)), c.full, false) {
		c.halt("glyph completion")
		return
	}
	c.expected, c.full, c.pending = c.full, nil, 0
	c.Glyphs++
	if ch, ok := c.chars[c.lastFile]; ok {
		c.done[ch.entry]++
	}
	c.col++
}

func (c *EndingCredits) beginClear(o *oracle.Oracle) {
	if !c.tracking {
		return
	}
	if _, ok := c.source(); !ok || c.pending != 0 || c.lastByte != 0xfc {
		c.halt("clear start")
		return
	}
	c.full = make([]byte, creditsRegionSize)
	c.pending = 2
}

func (c *EndingCredits) endClear(o *oracle.Oracle) {
	if !c.tracking || c.pending != 2 {
		return
	}
	if _, ok := c.source(); !ok || !bytes.Equal(creditsRegion(glyphpatch.Indexed(o)), c.full) {
		c.halt("clear completion")
		return
	}
	c.expected, c.full, c.pending = c.full, nil, 0
	c.screen++
	c.row, c.col = 0, 0
	c.ready = true
}

// 換圖不寫文字區時原樣保留；寫了只容許清成0（整幅淡出），其他寫入停止覆繪。
func (c *EndingCredits) beginPicture(o *oracle.Oracle) {
	if !c.tracking {
		return
	}
	if _, ok := c.source(); !ok || c.pending != 0 || c.lastByte != 0xfe {
		c.halt("picture start")
		return
	}
	c.pending = 3
}

func (c *EndingCredits) endPicture(o *oracle.Oracle) {
	if !c.tracking || c.pending != 3 {
		return
	}
	region := creditsRegion(glyphpatch.Indexed(o))
	// 目前畫面還沒畫字時（例如版權前整幅清除），換圖後的文字區就是新的起點。
	blank := true
	for i, e := range c.entries {
		if e.screen == c.screen && c.done[i] > 0 {
			blank = false
		}
	}
	if _, ok := c.source(); !ok || c.ready && !blank && !c.matches(region, nil, true) {
		c.halt("picture completion")
		return
	}
	c.pending = 0
	if c.ready {
		c.expected = append([]byte(nil), region...)
	}
}

// matches 比對文字區：與預期相同，cleared 時也接受已清成0的點；游標格（目前欄列的8×14格）不比。
func (c *EndingCredits) matches(region, alt []byte, cleared bool) bool {
	if len(region) != creditsRegionSize || len(c.expected) != creditsRegionSize {
		return false
	}
	cursor := image.Rect(c.col*8, c.row*14+144-creditsRegionTop, c.col*8+8, c.row*14+144-creditsRegionTop+14)
	for i, v := range region {
		if v == c.expected[i] || alt != nil && v == alt[i] || cleared && v == 0 {
			continue
		}
		if image.Pt(i%320, i/320).In(cursor) {
			continue
		}
		return false
	}
	return true
}

// Draw 在已驗的名單畫面上，把文字區塗成原版底色後畫出目前畫面的中文。
func (c *EndingCredits) Draw(dst *image.RGBA, raw []byte, palette [256][3]uint8, origin image.Point) bool {
	if c == nil || !c.tracking || !c.ready || dst == nil || !image.Rect(0, creditsRegionTop*3, 960, 600).Add(origin).In(dst.Bounds()) {
		return false
	}
	if _, ok := c.source(); !ok {
		c.halt("display source")
		return false
	}
	region := creditsRegion(raw)
	if !c.matches(region, c.full, c.pending == 2 || c.pending == 3) {
		c.halt("unobserved display write")
		return false
	}
	any := false
	for i, e := range c.entries {
		if e.screen == c.screen && c.done[i] > 0 {
			any = true
		}
	}
	if !any {
		return false
	}
	bg, ink := palette[0], palette[119]
	for y := creditsRegionTop * 3; y < 600; y++ {
		for x := 0; x < 960; x++ {
			dst.SetRGBA(origin.X+x, origin.Y+y, color.RGBA{bg[0], bg[1], bg[2], 255})
		}
	}
	for i, e := range c.entries {
		if e.screen != c.screen || c.done[i] == 0 {
			continue
		}
		n := c.done[i] * len(e.text) / len(e.files)
		if c.done[i] == len(e.files) {
			n = len(e.text)
		}
		for j, p := range e.positions[:n] {
			c.font.DrawGlyph(dst, p.Add(origin), e.text[j], 24, color.RGBA{ink[0], ink[1], ink[2], 255}, nil)
		}
	}
	return true
}

func (c *EndingCredits) State() map[string]any {
	if c == nil {
		return nil
	}
	return map[string]any{"active": c.tracking && c.ready, "screen": c.screen, "glyphs": c.Glyphs, "pending": c.pending, "finished": c.finished, "last_source_file": c.lastFile, "rejected": c.Rejected, "last_error": c.lastError, "scope": "結局製作名單、登場魔物與版權60字段；原版字元、清框、換圖來源"}
}

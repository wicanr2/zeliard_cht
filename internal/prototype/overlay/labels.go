package overlay

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
	"github.com/wicanr2/zeliard_cht/internal/prototype/glyphpatch"
)

type labelSource struct {
	offset         int
	base           uint16
	prefix         []byte
	width          int
	raw            []byte
	text           []rune
	inventory      bool
	inventoryWhite bool
	bankTitle      bool
	itemTitle      bool
	leftPadding    int
	// titleHeader 是 22CD counted 標題的 4 byte header：X/4、Y、X 補位、字數。
	titleHeader []byte
	// ink 是原版字形色號；一般為 9，以色碼 8Nh 起頭的字串為 N×11h（片頭標題 87h 實測為 119）。
	ink uint8
	// alias 是原版把字串複製過去再繪製的執行期位址；0 表示直接從來源位址繪製。
	alias uint16
	// signRows 是洞窟告示各行的原版字列；非 nil 表示由 740C 繪製。
	signRows []image.Rectangle
	// lead 是字串前的原版控制 byte 數（片頭標題的 87h 色碼），不畫字、不占寬度。
	lead int
}
type labelStamp struct {
	source                         *labelSource
	box                            image.Rectangle
	rows                           []image.Rectangle
	raw                            []byte
	palette                        [256][3]uint8
	paintSerial                    uint64
	foreground, shadow, background uint8
	shadowOffset                   *image.Point
	listViewport                   bool
	noUseIcon                      bool
}
type listLabel struct {
	source  *labelSource
	pixels  []byte
	palette [256][3]uint8
}

// Labels 只觀察已知 MCGA 繪字入口，來源與完成畫面不符時保留原貌。
// 選單選擇、金額、存檔名稱及原版輸入完全由原版處理。
type Labels struct {
	font             bitmap.Font
	sources          map[string]*labelSource
	stamps           []labelStamp
	pending          *labelStamp
	listPrepared     *listLabel
	Completed        int
	paint            *paintHistory
	inventoryPending *labelStamp
	inventoryIndex   int
	bankTitlePending *labelStamp
	itemTitlePending *labelStamp
	// signPending 是洞窟告示 740C 繪製中的字串；caveModulePrefix 是 01B1:6000 洞窟模組的前 64 bytes。
	signPending      *labelStamp
	caveModulePrefix []byte
	original         *oracle.Oracle
	itemListPrefix   []byte
	itemListCache    map[int]*listLabel
}

func NewLabels(o *oracle.Oracle, sar, catalog string, font bitmap.Font) (*Labels, error) {
	type binding struct {
		data    []byte
		modules []int
	}
	files := map[string]binding{}
	for id, source := range map[string][2]string{
		"zelres2": {sar, "bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1"},
		"zelres1": {filepath.Join(filepath.Dir(sar), "ZELRES1.SAR"), "56a73b59e952dab47fb0171a3fb1cdb7d641baceafa97e2826057f877a590eb0"},
		"stick":   {filepath.Join(filepath.Dir(sar), "STICK.BIN"), "3255d2a0f1f3d5c4a603e61fa8bca57dcf2130a44d8c37ed7e1df63c6ea7f8a9"},
		"zelres3": {filepath.Join(filepath.Dir(sar), "ZELRES3.SAR"), "9c19b9c317d3006e11f8a5cad8ea1f226f8043e11c6467fa27f10eef33175ef0"},
	} {
		data, err := os.ReadFile(source[0])
		if err != nil {
			return nil, err
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != source[1] {
			return nil, fmt.Errorf("介面來源版本不符：%s", id)
		}
		b := binding{data: data}
		if id != "stick" {
			table := int(binary.LittleEndian.Uint32(data))
			if table < 4 || table%4 != 0 || table > len(data) {
				return nil, fmt.Errorf("SAR 索引長度錯誤")
			}
			for i := 0; i < table; i += 4 {
				p := int(binary.LittleEndian.Uint32(data[i:]))
				end := len(data)
				if i+4 < table {
					end = int(binary.LittleEndian.Uint32(data[i+4:]))
				}
				if p < table || p+4 > end || end > len(data) || p+4+int(binary.LittleEndian.Uint32(data[p:])) != end {
					return nil, fmt.Errorf("SAR 項目邊界錯誤")
				}
				b.modules = append(b.modules, p+4)
			}
		}
		files[id] = b
	}
	inventoryKinds, err := inventorySourceKinds(files["zelres2"].data)
	if err != nil {
		return nil, err
	}
	l := &Labels{font: font, sources: map[string]*labelSource{}, paint: observePaint(o), original: o, itemListPrefix: append([]byte(nil), files["zelres2"].data[0x1a588:0x1a5c8]...), itemListCache: map[int]*listLabel{}}
	l.caveModulePrefix = append([]byte(nil), files["zelres2"].data[0xec:0xec+64]...)
	tsv, err := os.ReadFile(catalog)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(tsv) {
		return nil, fmt.Errorf("選單 TSV 必須是 UTF-8")
	}
	r := csv.NewReader(bytes.NewReader(tsv))
	r.Comma = '\t'
	r.FieldsPerRecord = 2
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 || rows[0][0] != "key" || rows[0][1] != "zh-TW" {
		return nil, fmt.Errorf("選單 TSV 欄位錯誤")
	}
	for _, row := range rows[1:] {
		parts := strings.Split(row[0], ".")
		if len(parts) != 3 || parts[1] != "file" {
			return nil, fmt.Errorf("選單鍵錯誤")
		}
		b, ok := files[parts[0]]
		if !ok {
			return nil, fmt.Errorf("未知介面來源")
		}
		data := b.data
		v, e := strconv.ParseUint(parts[2], 16, 32)
		off := int(v)
		if e != nil || fmt.Sprintf("%s.file.%06x", parts[0], off) != row[0] || off >= len(data) || l.sources[row[0]] != nil || row[1] == "" || strings.ContainsAny(row[1], "\t\r\n") {
			return nil, fmt.Errorf("選單鍵重複或文字錯誤：%s", row[0])
		}
		module := -1
		for _, p := range b.modules {
			end := p + int(binary.LittleEndian.Uint32(data[p-4:]))
			if off >= p && off < end {
				module = p
				break
			}
		}
		base := uint16(0xa000)
		limit := len(data)
		if parts[0] == "stick" {
			module = 0
			base = 0x100
		} else if module >= 0 {
			limit = module + int(binary.LittleEndian.Uint32(data[module-4:]))
		}
		// SAR1 第0項（片頭與標題）及第6項（讀檔選單）都載在 01B1:6000。
		if parts[0] == "zelres1" {
			if module != 0xceb5 && module != 0xa4 {
				return nil, fmt.Errorf("未接入的 SAR1 介面模組")
			}
			base = 0x6000
		}
		// 規格066：SAR3 第18項的指標表 A448 等指向「X 位移字＋字串」，模組載在 01B1:A000，
		// 由洞窟模組 73DE 交 291A 繪製。其他 SAR3 項目未接入。
		// SAR3 第22、25項的洞窟地圖模組載在 01B1:C000，各有一個告示，見 addSign。
		if parts[0] == "zelres3" && (module == 0xc8d2 || module == 0xf3c2) {
			if err := l.addSign(row[0], strings.ReplaceAll(row[1], `\n`, "\n"), data, module, off, font); err != nil {
				return nil, err
			}
			continue
		}
		// SAR3 第49項以 C000 為基準的指標表指向 C1C3 的 counted 標題。
		if parts[0] == "zelres3" && module != 0x9aee && module != 0x27e69 {
			return nil, fmt.Errorf("未接入的 SAR3 介面模組")
		}
		if parts[0] == "zelres3" && module == 0x27e69 {
			base = 0xc000
		}
		// 規格066：洞窟模組（SAR2 第0項）載在 01B1:6000。訊息以 DX 指向「X 位移字＋字串」，
		// 73DE 加 3Ah 後交 291A 繪製；字串前的位移字不屬於譯文來源。
		if parts[0] == "zelres2" && module == 0xec {
			base = 0x6000
		}
		if module < 0 || module+64 > limit {
			return nil, fmt.Errorf("選單鍵不在資源項目")
		}
		// 片頭標題版權字串以 87h 起頭，291A 不為它前進 X；只接受此模組的單一前導 byte。
		lead := 0
		if parts[0] == "zelres1" && module == 0xa4 && data[off] >= 0x80 {
			lead = 1
		}
		end := off + lead
		ink := uint8(9)
		if lead > 0 {
			ink = (data[off] & 0x0f) * 0x11
		}
		for end < limit && (data[end] >= 0x20 && data[end] < 0x7f || data[end] == 13) {
			end++
		}
		bankTitle := parts[0] == "zelres2" && module == 0x19462 && off == 0x19d54
		titleHeader := countedTitles[[2]int{module, off}]
		itemTitle := (parts[0] == "zelres2" || parts[0] == "zelres3") && titleHeader != nil
		if bankTitle {
			if !bytes.Equal(data[off-4:off], []byte{0x18, 0xaf, 2, 8}) {
				return nil, fmt.Errorf("銀行標題來源不符")
			}
			end = off + 8
		}
		if itemTitle {
			if !bytes.Equal(data[off-4:off], titleHeader) {
				return nil, fmt.Errorf("商店標題來源不符：%s", row[0])
			}
			end = off + int(titleHeader[3])
		}
		if end == off+lead || end >= limit || !bankTitle && !itemTitle && data[end] != 0 && data[end] != 255 {
			return nil, fmt.Errorf("介面來源須以 NUL 或 FF 結束：%s", row[0])
		}
		charWidth := 5
		leftPadding := 0
		inventory, inventoryWhite := false, false
		if parts[0] == "zelres2" && module == 0x401e {
			kind, known := inventoryKinds[off]
			if !known {
				return nil, fmt.Errorf("未分類的裝備文字來源：%s", row[0])
			}
			inventory = kind == inventoryLargeWhite || kind == inventoryLargeColor
			inventoryWhite = kind == inventoryLargeWhite
			if kind == inventoryShieldSecondLine || kind == inventorySwordSecondLine {
				for off+leftPadding < end && data[off+leftPadding] == ' ' {
					leftPadding++
				}
				letters := end - off - leftPadding
				valid := (leftPadding == 5 || leftPadding == 6) && letters == 6
				if kind == inventorySwordSecondLine {
					valid = leftPadding >= 4 && leftPadding <= 7 && letters == 5
				}
				if !valid {
					return nil, fmt.Errorf("裝備第二行的空白與文字範圍不符")
				}
			}
		}
		if inventory {
			charWidth = 8
		}
		if data[end] == 255 {
			charWidth = 8
		}
		originalLines := strings.Split(string(data[off+lead:end]), "\r")
		if charWidth == 5 && len(originalLines) != 1 {
			return nil, fmt.Errorf("ASCIIZ 不支援換行")
		}
		if leftPadding > 0 {
			originalLines[0] = originalLines[0][leftPadding:]
		}
		text := strings.ReplaceAll(row[1], `\n`, "\n")
		lines := strings.Split(text, "\n")
		if len(lines) > len(originalLines) {
			return nil, fmt.Errorf("介面譯文超過安全行數")
		}
		width, rowIndex := 0, 0
		for _, ch := range []rune(text) {
			if ch == '\n' {
				width = 0
				rowIndex++
				continue
			}
			g := font[ch]
			if len(g) != 16 && len(g) != 32 {
				return nil, fmt.Errorf("選單字型缺字：%U", ch)
			}
			width += len(g) / 16 * 12
			if width > len(originalLines[rowIndex])*charWidth*3 {
				return nil, fmt.Errorf("選單譯文超出原文安全寬度：%s", row[0])
			}
		}
		rawEnd := end + 1
		if bankTitle || itemTitle {
			rawEnd = end
		}
		l.sources[row[0]] = &labelSource{offset: off - module, base: base, prefix: append([]byte(nil), data[module:module+64]...), width: charWidth, raw: append([]byte(nil), data[off:rawEnd]...), text: []rune(text), inventory: inventory, inventoryWhite: inventoryWhite, bankTitle: bankTitle, itemTitle: itemTitle, titleHeader: titleHeader, leftPadding: leftPadding, lead: lead, ink: ink}
		// 規格066：讀檔清單由 E001 指標表以 22DB 畫出 77BA 的 Re-Start；名稱欄空白時，選單 76F0..7723
		// 另把它複製到 7C67 的 8 字欄位，以 FF 結尾交 291A 繪製。欄位依位址與內容比對，使用者輸入的名稱照原版。
		if parts[0] == "zelres1" && module == 0xceb5 && off == 0xe66f {
			field := *l.sources[row[0]]
			field.width, field.alias = 8, 0x7c67
			field.raw = append(append([]byte(nil), data[off:end]...), 0xff)
			l.sources[row[0]+"#field"] = &field
		}
	}
	o.OnCall(oracle.Far(0x01b1, 0x22db), func(q *oracle.Oracle) { l.begin(q, 5) })
	o.OnCall(oracle.Far(0x01b1, 0x2308), func(q *oracle.Oracle) { l.finish(q, 5) })
	o.OnCall(oracle.Far(0x01b1, 0x291a), func(q *oracle.Oracle) { l.begin(q, 8) })
	o.OnCall(oracle.Far(0x01b1, 0x2939), func(q *oracle.Oracle) { l.finish(q, 8) })
	o.OnCall(oracle.Far(0x01b1, 0x383e), l.prepareListSource)
	o.OnCall(oracle.Far(0x01b1, 0x384a), l.captureListSource)
	o.OnCall(oracle.Far(0x01b1, 0x37cc), l.beginListRow)
	o.OnCall(oracle.Far(0x01b1, 0x3804), func(q *oracle.Oracle) { l.finish(q, 5) })
	o.OnCall(oracle.Far(0x01b1, 0xaa2b), l.beginInventory)
	o.OnCall(oracle.Far(0x01b1, 0xaa30), l.finishInventory)
	o.OnCall(oracle.Far(0x01b1, 0x22cd), l.beginBankTitle)
	o.OnCall(oracle.Far(0x01b1, 0x2344), l.finishBankTitle)
	o.OnCall(oracle.Far(0x01b1, 0x22cd), l.beginItemTitle)
	o.OnCall(oracle.Far(0x01b1, 0x2344), l.finishItemTitle)
	o.OnCall(oracle.Far(0x01b1, 0x740c), l.beginSign)
	o.OnCall(oracle.Far(0x01b1, 0x7456), l.finishSign)
	return l, nil
}

// 商品名先以五像素步距烘到 stride=160 的原版 buffer，之後才搬到 VGA。
// 只記錄已知來源的名稱前綴，價格及其他欄位完全保留原版。
func (l *Labels) prepareListSource(o *oracle.Oracle) {
	l.listPrepared = nil
	if o.VideoMode() != 0x13 || o.DSReg() != 0x01b1 || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0x383e), 7), []byte{0x0e, 0x07, 0xbf, 0x4f, 0x3d, 0x32, 0xdb}) {
		return
	}
	for _, s := range l.sources {
		if !s.bankTitle && !s.itemTitle && s.width == 5 && int(s.base)+s.offset == int(o.SI()) && bytes.Equal(o.Bytes(oracle.Far(0x01b1, s.base), len(s.prefix)), s.prefix) && bytes.Equal(o.Bytes(oracle.Far(0x01b1, o.SI()), len(s.raw)), s.raw) {
			l.listPrepared = &listLabel{source: s}
			return
		}
	}
}
func listNameCrop(o *oracle.Oracle, s *labelSource) []byte {
	w := (len(s.raw) - 1) * 5
	if w <= 0 || w > 160 {
		return nil
	}
	var pixels []byte
	for y := 0; y < 8; y++ {
		pixels = append(pixels, o.Bytes(oracle.Far(0x01b1, uint16(0x3d4f+y*160)), w)...)
	}
	return pixels
}
func (l *Labels) captureListSource(o *oracle.Oracle) {
	if l.listPrepared == nil {
		return
	}
	pixels := listNameCrop(o, l.listPrepared.source)
	for _, v := range pixels {
		if v != 0 && v != 9 {
			l.listPrepared = nil
			return
		}
	}
	l.listPrepared.pixels = pixels
	l.listPrepared.palette = o.Palette()
	if l.isItemListSource(l.listPrepared.source) && len(pixels) > 0 {
		cached := *l.listPrepared
		l.itemListCache[cached.source.offset] = &cached
	}
}
func (l *Labels) beginListRow(o *oracle.Oracle) {
	l.pending = nil
	p := l.listPrepared
	l.listPrepared = nil
	if p == nil || len(p.pixels) == 0 || o.DSReg() != 0x01b1 || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0x37cc), 6), []byte{0x33, 0xc0, 0x8a, 0xc7, 0x8a, 0xfc}) || !bytes.Equal(p.pixels, listNameCrop(o, p.source)) {
		return
	}
	w := (len(p.source.raw) - 1) * 5
	if int(o.Word(oracle.Far(0x01b1, 0xff6a)))*4 < w {
		return
	}
	x, y := int(o.BX()>>8)*4, int(o.BX()&255)+1
	box := image.Rect(x, y, x+w, y+8)
	if !box.In(image.Rect(0, 0, 320, 200)) {
		return
	}
	l.pending = &labelStamp{source: p.source, box: box, rows: []image.Rectangle{box}}
}

func (l *Labels) begin(o *oracle.Oracle, charWidth int) {
	l.pending = nil
	if charWidth == 8 {
		glyphpatch.ReleaseLines(o)
	}
	entry, signature := uint16(0x22db), []byte{0x2e, 0xc6, 0x06, 0xbd, 0x2c, 0x09}
	if charWidth == 8 {
		entry, signature = 0x291a, []byte{0x2e, 0x89, 0x1e, 0xc0, 0x2c}
	}
	if o.VideoMode() != 0x13 || o.DSReg() != 0x01b1 || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, entry), len(signature)), signature) {
		return
	}
	pos := int(o.SI())
	for _, s := range l.sources {
		addr := s.offset + int(s.base)
		if s.alias != 0 {
			addr = int(s.alias)
		}
		if s.bankTitle || s.itemTitle || s.inventory || s.signRows != nil || s.width != charWidth || addr != pos {
			continue
		}
		// 模組固定前綴與完整來源字串均須相符，不能只靠當次 SI 認出文字。
		if !bytes.Equal(o.Bytes(oracle.Far(0x01b1, s.base), len(s.prefix)), s.prefix) || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, uint16(pos)), len(s.raw)), s.raw) {
			continue
		}
		x := int(o.BX()>>8)*4 + int(o.CX()&255)
		y := int(o.BX() & 255)
		if charWidth == 8 {
			x = int(o.BX())
			y = int(o.CX() & 255)
		}
		lines := strings.Split(string(s.raw[s.lead:len(s.raw)-1]), "\r")
		// 024／025：原版在裝備第二行的前置空白後畫數字，只覆繪字母區。
		if s.leftPadding > 0 {
			if charWidth != 5 || len(lines) != 1 || s.leftPadding >= len(lines[0]) {
				return
			}
			x += s.leftPadding * charWidth
			lines[0] = lines[0][s.leftPadding:]
		}
		maxChars := 0
		for _, line := range lines {
			maxChars = max(maxChars, len(line))
		}
		box := image.Rect(x, y, x+maxChars*charWidth, y+len(lines)*8)
		if box.Empty() || !box.In(image.Rect(0, 0, 320, 200)) {
			return
		}
		var rows []image.Rectangle
		for i, line := range lines {
			rows = append(rows, image.Rect(x, y+i*8, x+len(line)*charWidth, y+(i+1)*8))
		}
		l.pending = &labelStamp{source: s, box: box, rows: rows}
		// 規格066：整串都由中文取代時才認領；前置空白後接原版數字的字串保留原版繪字。
		if s.leftPadding == 0 {
			if charWidth == 8 {
				glyphpatch.HoldLines(o)
			} else {
				glyphpatch.ClaimText(o)
			}
		}
		return
	}
}
func labelCrop(raw []byte, box image.Rectangle) []byte {
	if len(raw) != 64000 {
		return nil
	}
	out := make([]byte, box.Dx()*box.Dy())
	for y := box.Min.Y; y < box.Max.Y; y++ {
		copy(out[(y-box.Min.Y)*box.Dx():], raw[y*320+box.Min.X:y*320+box.Max.X])
	}
	return out
}
func labelRowsCrop(raw []byte, rows []image.Rectangle) []byte {
	if len(raw) != 64000 {
		return nil
	}
	var data []byte
	for _, row := range rows {
		data = append(data, labelCrop(raw, row)...)
	}
	return data
}
func (l *Labels) finish(o *oracle.Oracle, charWidth int) {
	if charWidth == 8 {
		glyphpatch.ReleaseLines(o)
	}
	if l.pending == nil || l.pending.source.width != charWidth {
		return
	}
	s := *l.pending
	l.pending = nil
	s.raw = labelRowsCrop(glyphpatch.Indexed(o), s.rows)
	s.palette = o.Palette()
	s.paintSerial = l.paint.serial
	// 規格066：已認領卻不收的字串照原版寫回，不留空白。
	reject := func() {
		if s.source.leftPadding == 0 {
			glyphpatch.Reveal(o, s.rows...)
		}
	}
	if len(s.raw) == 0 {
		reject()
		return
	}
	ink := s.source.ink
	if ink == 0 {
		ink = 9
	}
	for _, v := range s.raw {
		if v != 0 && v != ink {
			reject()
			return
		}
	}
	s.foreground = ink
	l.store(s)
}

func (l *Labels) store(s labelStamp) {
	// 重繪同一位置時取代舊標籤，不累積過期選單。
	for i := range l.stamps {
		if l.stamps[i].box == s.box {
			l.stamps[i] = s
			l.Completed++
			return
		}
	}
	if len(l.stamps) >= 64 {
		l.stamps = l.stamps[1:]
	}
	l.stamps = append(l.stamps, s)
	l.Completed++
}
func (l *Labels) Draw(dst *image.RGBA, raw []byte, pal [256][3]uint8, origin image.Point) int {
	_, viewport := l.itemListFrame(raw, pal)
	kept := l.stamps[:0]
	drawn := 0
	for _, s := range l.stamps {
		if s.noUseIcon && !l.inventoryIconFrame(raw, pal) {
			continue
		}
		if s.listViewport && !viewport {
			continue
		}
		if pal != s.palette {
			continue
		}
		masks, known := l.paint.masks(s.paintSerial, raw)
		if !known || !visibleRowsEqual(s.raw, raw, s.rows, masks) {
			continue
		}
		kept = append(kept, s)
		drawn++
		box := image.Rect(origin.X+s.box.Min.X*3, origin.Y+s.box.Min.Y*3, origin.X+s.box.Max.X*3, origin.Y+s.box.Max.Y*3)
		bg, fg := pal[s.background], pal[s.foreground]
		var rows []image.Rectangle
		for _, row := range s.rows {
			rows = append(rows, image.Rect(origin.X+row.Min.X*3, origin.Y+row.Min.Y*3, origin.X+row.Max.X*3, origin.Y+row.Max.Y*3))
		}
		for _, row := range rows {
			for y := row.Min.Y; y < row.Max.Y; y++ {
				for x := row.Min.X; x < row.Max.X; x++ {
					if image.Pt(x, y).In(dst.Bounds()) && visiblePixel(image.Pt(x, y), origin, masks) {
						dst.SetRGBA(x, y, color.RGBA{bg[0], bg[1], bg[2], 255})
					}
				}
			}
		}
		x, y := box.Min.X, box.Min.Y
		if len(rows) > 0 && !s.noUseIcon {
			// 每列從原版該列的起點畫；一般多行字串各列同起點，洞窟告示各列置中。
			x, y = rows[0].Min.X, rows[0].Min.Y
		}
		rowIndex := 0
		size := 24
		if s.noUseIcon {
			size = 18
			x += (box.Dx() - size) / 2
			y += (box.Dy() - size*len(s.source.text)) / 2
		}
		for index, ch := range s.source.text {
			if s.noUseIcon && index > 0 {
				x = box.Min.X + (box.Dx()-size)/2
				y += size
			}
			if ch == '\n' {
				x = box.Min.X
				y += 24
				rowIndex++
				if rowIndex < len(rows) {
					x, y = rows[rowIndex].Min.X, rows[rowIndex].Min.Y
				}
				continue
			}
			w := len(l.font[ch]) / 16 * 8
			dw := w * size / 16
			visible := func(p image.Point) bool { return p.In(rows[rowIndex]) && visiblePixel(p, origin, masks) }
			if s.shadow != 0 {
				offset := image.Pt(3, 3)
				if s.shadowOffset != nil {
					offset = s.shadowOffset.Mul(3)
				}
				c := pal[s.shadow]
				l.font.DrawGlyph(dst, image.Pt(x, y).Add(offset), ch, size, color.RGBA{c[0], c[1], c[2], 255}, visible)
			}
			l.font.DrawGlyph(dst, image.Pt(x, y), ch, size, color.RGBA{fg[0], fg[1], fg[2], 255}, visible)
			x += dw
		}
	}
	l.stamps = kept
	// 規格066：片頭模式不壓字形；以色碼起頭的標題字串在原版畫字期間就先畫中文，不等 2939 完成。
	if p := l.pending; p != nil && p.source.lead > 0 && pal[0] != pal[p.source.ink] {
		l.drawPending(dst, p, pal, origin)
		drawn++
	}
	return drawn
}

// isItemListSource 只接受道具店名稱表的八件商品名稱。
func (l *Labels) isItemListSource(s *labelSource) bool {
	if s == nil || s.base != 0xa000 || s.width != 5 || !bytes.Equal(s.prefix, l.itemListPrefix) {
		return false
	}
	_, known := itemDetailOf(uint16(0xa000 + s.offset))
	return known
}

// itemListFrame 辨識正常捲動未再走37CC/3804的兩列顯示。
// 只接受已觀察383E/384A產生的完整字形、固定viewport及兩列不同的已知商品。
// 位址與bitmap證據見 CONTEXT 的 scroll-fix-v1，不推定未知copy routine語意。
func (l *Labels) itemListFrame(raw []byte, pal [256][3]uint8) ([2]*listLabel, bool) {
	o := l.original
	if o == nil || len(raw) != 64000 || o.VideoMode() != 0x13 || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0xa000), 64), l.itemListPrefix) {
		return [2]*listLabel{}, false
	}
	var fixed []byte
	for y := 120; y < 146; y++ {
		for x := 92; x < 224; x++ {
			if (y >= 124 && y < 132 || y >= 134 && y < 142) && x >= 104 && x < 177 {
				continue
			}
			fixed = append(fixed, raw[y*320+x])
		}
	}
	if fmt.Sprintf("%x", sha256.Sum256(fixed)) != "ac575320f6b3ef26a3208535c532847f7ccbd8bfb9f42cc349d4633804d79e2e" {
		return [2]*listLabel{}, false
	}
	selected := 0
	var found [2]*listLabel
	for row, y := range []int{124, 134} {
		arrow := fmt.Sprintf("%x", sha256.Sum256(labelCrop(raw, image.Rect(104, y, 112, y+8))))
		switch arrow {
		case "7d5fb6dffdb75d8d0fbdd85c03ccc0574528d4302b919d73be048e6cbdf06a0e":
			selected++
		case "f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b":
		default:
			return [2]*listLabel{}, false
		}
		actual := labelCrop(raw, image.Rect(112, y, 177, y+8))
		for _, cached := range l.itemListCache {
			s := cached.source
			w := (len(s.raw) - 1) * 5
			if !l.isItemListSource(s) || cached.palette != pal || w > 65 || w < 1 || len(cached.pixels) != w*8 || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, uint16(0xa000+s.offset)), len(s.raw)), s.raw) {
				continue
			}
			padded := make([]byte, 65*8)
			for j := 0; j < 8; j++ {
				copy(padded[j*65:j*65+w], cached.pixels[j*w:(j+1)*w])
			}
			if bytes.Equal(actual, padded) {
				found[row] = cached
				break
			}
		}
		if found[row] == nil {
			return [2]*listLabel{}, false
		}
	}
	// 兩列須是不同商品；各列已由快取字形逐像素比對，城鎮商品組合不另列舉。
	if selected != 1 || found[0].source == found[1].source {
		return [2]*listLabel{}, false
	}
	return found, true
}

// ObserveFrame 在繪製前記錄已確認的新兩列，Draw 只繪製現有記錄。
// 單一舊標籤的隔離繪製不會順帶建立其他標籤。
func (l *Labels) ObserveFrame(raw []byte, pal [256][3]uint8) {
	l.observeInventoryIcon(raw, pal)
	o := l.original
	if o == nil || o.VideoMode() != 0x13 || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0xa000), 64), l.itemListPrefix) {
		if len(l.itemListCache) > 0 {
			l.itemListCache = map[int]*listLabel{}
		}
		return
	}
	found, valid := l.itemListFrame(raw, pal)
	if !valid {
		return
	}
	for row, cached := range found {
		y, w := 124+row*10, (len(cached.source.raw)-1)*5
		box := image.Rect(112, y, 112+w, y+8)
		current := false
		for _, stamp := range l.stamps {
			if stamp.listViewport && stamp.source == cached.source && stamp.box == box && stamp.palette == pal && bytes.Equal(stamp.raw, cached.pixels) {
				current = true
				break
			}
		}
		if current {
			continue
		}
		kept := l.stamps[:0]
		for _, stamp := range l.stamps {
			if l.isItemListSource(stamp.source) && stamp.box.Min == box.Min {
				continue
			}
			kept = append(kept, stamp)
		}
		l.stamps = kept
		l.store(labelStamp{source: cached.source, box: box, rows: []image.Rectangle{box}, raw: append([]byte(nil), cached.pixels...), palette: pal, paintSerial: l.paint.serial, foreground: 9, listViewport: true})
	}
}

// countedTitles 是以 22CD 繪製的 counted 標題：SAR {模組, 字串起點} 與字串前的 4 byte header。
// 魔法物品店已以正常路徑驗證；旅館與 SAR3 第49項的房間標題同格式，規格066加入，正常到達未驗。
var countedTitles = map[[2]int][]byte{
	{0x1a588, 0x1ada8}: {0x0e, 0xaf, 0, 25},
	{0x1b7b2, 0x1baa1}: {0x19, 0xaf, 0, 7},
	{0x27e69, 0x28030}: {0x18, 0xaf, 4, 14},
}

// drawPending 畫尚在原版繪字中的字串：只填原版字列範圍，中文以原版字形色號繪製。
func (l *Labels) drawPending(dst *image.RGBA, p *labelStamp, pal [256][3]uint8, origin image.Point) {
	var rows []image.Rectangle
	for _, row := range p.rows {
		rows = append(rows, image.Rect(origin.X+row.Min.X*3, origin.Y+row.Min.Y*3, origin.X+row.Max.X*3, origin.Y+row.Max.Y*3))
	}
	if len(rows) == 0 {
		return
	}
	bg, fg := pal[0], pal[p.source.ink]
	for _, row := range rows {
		draw.Draw(dst, row.Intersect(dst.Bounds()), &image.Uniform{color.RGBA{bg[0], bg[1], bg[2], 255}}, image.Point{}, draw.Src)
	}
	x, y, rowIndex := rows[0].Min.X, rows[0].Min.Y, 0
	for _, ch := range p.source.text {
		if ch == '\n' {
			if rowIndex+1 >= len(rows) {
				return
			}
			rowIndex++
			x, y = rows[rowIndex].Min.X, y+24
			continue
		}
		clip := rows[rowIndex]
		l.font.DrawGlyph(dst, image.Pt(x, y), ch, 24, color.RGBA{fg[0], fg[1], fg[2], 255}, func(q image.Point) bool { return q.In(clip) })
		x += len(l.font[ch]) / 16 * 8 * 24 / 16
	}
}

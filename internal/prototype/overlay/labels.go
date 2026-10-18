package overlay

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

type labelSource struct {
	offset int
	base   uint16
	prefix []byte
	width  int
	raw    []byte
	text   []rune
}
type labelStamp struct {
	source  *labelSource
	box     image.Rectangle
	rows    []image.Rectangle
	raw     []byte
	palette [256][3]uint8
}
type listLabel struct {
	source *labelSource
	pixels []byte
}

// Labels 只觀察已知 MCGA 繪字入口，來源與完成畫面不符時保留原貌。
// 選單選擇、金額、存檔名稱及原版輸入完全由原版處理。
type Labels struct {
	font         bitmap.Font
	sources      map[string]*labelSource
	stamps       []labelStamp
	pending      *labelStamp
	listPrepared *listLabel
	Completed    int
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
	l := &Labels{font: font, sources: map[string]*labelSource{}}
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
		if parts[0] == "zelres1" {
			if module != 0xceb5 {
				return nil, fmt.Errorf("未接入的 SAR1 介面模組")
			}
			base = 0x6000
		}
		if module < 0 || module+64 > limit {
			return nil, fmt.Errorf("選單鍵不在資源項目")
		}
		end := off
		for end < limit && (data[end] >= 0x20 && data[end] < 0x7f || data[end] == 13) {
			end++
		}
		if end == off || end >= limit || data[end] != 0 && data[end] != 255 {
			return nil, fmt.Errorf("介面來源須以 NUL 或 FF 結束：%s", row[0])
		}
		charWidth := 5
		if data[end] == 255 {
			charWidth = 8
		}
		originalLines := strings.Split(string(data[off:end]), "\r")
		if charWidth == 5 && len(originalLines) != 1 {
			return nil, fmt.Errorf("ASCIIZ 不支援換行")
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
		l.sources[row[0]] = &labelSource{offset: off - module, base: base, prefix: append([]byte(nil), data[module:module+64]...), width: charWidth, raw: append([]byte(nil), data[off:end+1]...), text: []rune(text)}
	}
	o.OnCall(oracle.Far(0x01b1, 0x22db), func(q *oracle.Oracle) { l.begin(q, 5) })
	o.OnCall(oracle.Far(0x01b1, 0x2308), func(q *oracle.Oracle) { l.finish(q, 5) })
	o.OnCall(oracle.Far(0x01b1, 0x291a), func(q *oracle.Oracle) { l.begin(q, 8) })
	o.OnCall(oracle.Far(0x01b1, 0x2939), func(q *oracle.Oracle) { l.finish(q, 8) })
	o.OnCall(oracle.Far(0x01b1, 0x383e), l.prepareListSource)
	o.OnCall(oracle.Far(0x01b1, 0x384a), l.captureListSource)
	o.OnCall(oracle.Far(0x01b1, 0x37cc), l.beginListRow)
	o.OnCall(oracle.Far(0x01b1, 0x3804), func(q *oracle.Oracle) { l.finish(q, 5) })
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
		if s.width == 5 && int(s.base)+s.offset == int(o.SI()) && bytes.Equal(o.Bytes(oracle.Far(0x01b1, s.base), len(s.prefix)), s.prefix) && bytes.Equal(o.Bytes(oracle.Far(0x01b1, o.SI()), len(s.raw)), s.raw) {
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
	entry, signature := uint16(0x22db), []byte{0x2e, 0xc6, 0x06, 0xbd, 0x2c, 0x09}
	if charWidth == 8 {
		entry, signature = 0x291a, []byte{0x2e, 0x89, 0x1e, 0xc0, 0x2c}
	}
	if o.VideoMode() != 0x13 || o.DSReg() != 0x01b1 || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, entry), len(signature)), signature) {
		return
	}
	pos := int(o.SI())
	for _, s := range l.sources {
		if s.width != charWidth || s.offset+int(s.base) != pos {
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
		lines := strings.Split(string(s.raw[:len(s.raw)-1]), "\r")
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
	if l.pending == nil || l.pending.source.width != charWidth {
		return
	}
	s := *l.pending
	l.pending = nil
	s.raw = labelRowsCrop(o.Indexed(), s.rows)
	s.palette = o.Palette()
	if len(s.raw) == 0 {
		return
	}
	for _, v := range s.raw {
		if v != 0 && v != 9 {
			return
		}
	}
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
	kept := l.stamps[:0]
	drawn := 0
	for _, s := range l.stamps {
		if pal != s.palette || !bytes.Equal(s.raw, labelRowsCrop(raw, s.rows)) {
			continue
		}
		kept = append(kept, s)
		drawn++
		box := image.Rect(origin.X+s.box.Min.X*3, origin.Y+s.box.Min.Y*3, origin.X+s.box.Max.X*3, origin.Y+s.box.Max.Y*3)
		bg, fg := pal[0], pal[9]
		var rows []image.Rectangle
		for _, row := range s.rows {
			rows = append(rows, image.Rect(origin.X+row.Min.X*3, origin.Y+row.Min.Y*3, origin.X+row.Max.X*3, origin.Y+row.Max.Y*3))
		}
		for _, row := range rows {
			for y := row.Min.Y; y < row.Max.Y; y++ {
				for x := row.Min.X; x < row.Max.X; x++ {
					if image.Pt(x, y).In(dst.Bounds()) {
						dst.SetRGBA(x, y, color.RGBA{bg[0], bg[1], bg[2], 255})
					}
				}
			}
		}
		x, y := box.Min.X, box.Min.Y
		rowIndex := 0
		for _, ch := range s.source.text {
			if ch == '\n' {
				x = box.Min.X
				y += 24
				rowIndex++
				continue
			}
			g := l.font[ch]
			w := len(g) / 16 * 8
			dw := w * 3 / 2
			for yy := 0; yy < 24; yy++ {
				for xx := 0; xx < dw; xx++ {
					sx, sy := xx*2/3, yy*2/3
					if g[sy*(w/8)+sx/8]&(0x80>>uint(sx%8)) != 0 {
						p := image.Pt(x+xx, y+yy)
						if p.In(rows[rowIndex]) && p.In(dst.Bounds()) {
							dst.SetRGBA(p.X, p.Y, color.RGBA{fg[0], fg[1], fg[2], 255})
						}
					}
				}
			}
			x += dw
		}
	}
	l.stamps = kept
	return drawn
}

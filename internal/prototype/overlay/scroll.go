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
	"unicode"
	"unicode/utf8"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

// Scroll只讀三組已驗捲動來源，契約見規格015。
type Scroll struct {
	o                                        *oracle.Oracle
	font                                     bitmap.Font
	entries                                  map[uint16]scrollEntry
	pool                                     []byte
	lines                                    []scrollLine
	current                                  scrollEntry
	block                                    *scrollBlock
	row                                      uint16
	active, pending, newLine, ending         bool
	completed, expected, glyphBuffer         []byte
	expectedFile, file, sourceRows           int
	Rows, Sources, Rejected, BlocksCompleted int
	blockRows                                map[string]int
	lastError                                string
}
type ScrollCatalogs struct{ First, Credits, Outro string }
type scrollBlock struct {
	name                           string
	start, end, count, top, height int
}
type scrollEntry struct {
	key, text  string
	raw        []byte
	file, next int
	block      *scrollBlock
	eof        bool
}
type scrollLine struct {
	key, text string
	y         int
}

// NewScroll保留既有只載首章的入口及TSV格式。
func NewScroll(o *oracle.Oracle, sar, catalog string, font bitmap.Font) (*Scroll, error) {
	return NewScrollBlocks(o, sar, ScrollCatalogs{First: catalog}, font)
}
func NewScrollBlocks(o *oracle.Oracle, sar string, catalogs ScrollCatalogs, font bitmap.Font) (*Scroll, error) {
	if o == nil {
		return nil, fmt.Errorf("捲字缺少原版執行器")
	}
	data, err := os.ReadFile(sar)
	if err != nil {
		return nil, err
	}
	if scrollHash(data) != "56a73b59e952dab47fb0171a3fb1cdb7d641baceafa97e2826057f877a590eb0" {
		return nil, fmt.Errorf("捲字來源版本不符")
	}
	s := &Scroll{o: o, font: font, entries: map[uint16]scrollEntry{}, pool: append([]byte(nil), data[0x1094:0x1a6a]...), blockRows: map[string]int{}}
	definitions := []struct {
		block scrollBlock
		path  string
	}{
		{scrollBlock{"first", 0x1094, 0x13db, 30, 32, 120}, catalogs.First},
		{scrollBlock{"credits", 0x14d3, 0x1a69, 51, 32, 120}, catalogs.Credits},
		{scrollBlock{"outro", 0x13dc, 0x14d2, 6, 20, 160}, catalogs.Outro},
	}
	for _, def := range definitions {
		if def.path == "" {
			continue
		}
		text, e := os.ReadFile(def.path)
		if e != nil {
			return nil, e
		}
		b := def.block
		entries, e := parseScrollCatalog(data, text, &b, font)
		if e != nil {
			return nil, e
		}
		for ptr, entry := range entries {
			s.entries[ptr] = entry
		}
	}
	if len(s.entries) == 0 {
		return nil, fmt.Errorf("捲字沒有譯表")
	}
	o.OnCall(oracle.Far(0x1b1, 0x32c9), s.source)
	o.OnCall(oracle.Far(0x1b1, 0x332c), s.beginRow)
	o.OnCall(oracle.Far(0x1b1, 0x33b6), s.endRow)
	return s, nil
}
func scrollHash(b []byte) string    { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func scrollPointer(file int) uint16 { return uint16(file - 0xa4 + 0x6000) }
func parseScrollCatalog(data, text []byte, b *scrollBlock, font bitmap.Font) (map[uint16]scrollEntry, error) {
	if b == nil || b.start < 0 || b.end >= len(data) || b.start >= b.end || data[b.end] != 255 {
		return nil, fmt.Errorf("捲字來源範圍不符")
	}
	if !utf8.Valid(text) {
		return nil, fmt.Errorf("捲字譯表須為UTF-8")
	}
	r := csv.NewReader(bytes.NewReader(text))
	r.Comma = '\t'
	r.FieldsPerRecord = 2
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) != b.count+1 || records[0][0] != "key" || records[0][1] != "zh-TW" {
		return nil, fmt.Errorf("捲字譯表行數或欄位不符：%s", b.name)
	}
	entries := map[uint16]scrollEntry{}
	off := b.start
	for _, record := range records[1:] {
		if off >= b.end {
			return nil, fmt.Errorf("捲字來源提前結束")
		}
		n := bytes.IndexByte(data[off:b.end], 13)
		if n < 0 || n > 40 {
			return nil, fmt.Errorf("捲字行分隔不符")
		}
		key := fmt.Sprintf("zelres1.file.%06x", off)
		if record[0] != key {
			return nil, fmt.Errorf("捲字來源鍵或順序不符：%s", record[0])
		}
		for _, ch := range data[off : off+n] {
			if ch < 32 || ch >= 127 {
				return nil, fmt.Errorf("捲字來源不是已驗字元")
			}
		}
		value := record[1]
		blank := strings.TrimSpace(string(data[off:off+n])) == ""
		if (blank && value != "") || (!blank && strings.TrimSpace(value) == "") {
			return nil, fmt.Errorf("捲字空白來源與譯文不符：%s", key)
		}
		for _, ch := range value {
			if unicode.IsControl(ch) || (len(font[ch]) != 16 && len(font[ch]) != 32) {
				return nil, fmt.Errorf("捲字缺字或控制碼：%U", ch)
			}
		}
		if textWidth(font, value) > 900 {
			return nil, fmt.Errorf("捲字超出單行安全寬度：%s", key)
		}
		entries[scrollPointer(off)] = scrollEntry{key: key, text: value, raw: append([]byte(nil), data[off:off+n+1]...), file: off, next: off + n + 1, block: b}
		off += n + 1
	}
	if off != b.end {
		return nil, fmt.Errorf("捲字來源行數不符")
	}
	entries[scrollPointer(off)] = scrollEntry{raw: []byte{255}, file: off, next: off, block: b, eof: true}
	return entries, nil
}
func (s *Scroll) reset(reason string) {
	if (s.active || s.pending) && reason != "" {
		s.Rejected++
		s.lastError = reason
	}
	s.active = false
	s.pending = false
	s.newLine = false
	s.lines = nil
	s.completed = nil
	s.expected = nil
	s.glyphBuffer = nil
}
func (s *Scroll) sourceSnapshot() ([]byte, bool) {
	if s.o.VideoMode() != 0x13 {
		return nil, false
	}
	p := s.o.Bytes(oracle.Far(0x1b1, 0), 65536)
	if p[0xf500] != 6 || p[0xf501] != 0xf5 || !bytes.Equal(p[0x6ff0:0x79c6], s.pool) {
		return nil, false
	}
	if scrollHash(p[0x32c9:0x33b7]) != "504241bf9c5662c6e2ee2059e11f3915468d94035e7de83dc2860fbf15537623" || scrollHash(p[0xf506:0xf806]) != "8ff0729de496f9d84d4cb3e4931a349a028981232507939f23af95a821d955d6" {
		return nil, false
	}
	return p, true
}
func (s *Scroll) source(o *oracle.Oracle) {
	p, ok := s.sourceSnapshot()
	e, known := s.entries[o.SI()]
	if !ok || !known || o.DSReg() != 0x1b1 || !bytes.Equal(p[int(o.SI()):int(o.SI())+len(e.raw)], e.raw) {
		s.reset("source")
		return
	}
	if e.file == e.block.start {
		s.reset("")
		s.block = e.block
		s.expectedFile = e.block.start
		s.active = true
		s.lastError = ""
	} else if !s.active || s.block != e.block || s.sourceRows != 10 {
		s.reset("source order")
		return
	}
	if s.pending || e.file != s.expectedFile {
		s.reset("source sequence")
		return
	}
	s.current = e
	s.file = e.file
	s.expectedFile = e.next
	s.sourceRows = 0
	s.newLine = !e.eof
	s.ending = e.eof
	s.glyphBuffer = make([]byte, 3200)
	if !e.eof {
		for x, ch := range e.raw[:len(e.raw)-1] {
			glyph := p[0xf506+int(ch-32)*8 : 0xf506+int(ch-32+1)*8]
			for y, bits := range glyph {
				for bit := 0; bit < 8; bit++ {
					if bits&(128>>bit) != 0 {
						s.glyphBuffer[y*320+x*8+bit] = 255
					}
				}
			}
		}
		s.Sources++
	}
}
func (s *Scroll) beginRow(o *oracle.Oracle) {
	if !s.active {
		return
	}
	p, ok := s.sourceSnapshot()
	r := o.Regs()
	row := s.sourceRows
	if row >= 10 {
		row = 0
	}
	if !ok || o.DSReg() != 0x1b1 || s.pending || (!s.ending && s.sourceRows >= 10) || int(r.BX) != s.block.top || int(r.CX) != (0x5000|s.block.height) || int(r.AX) != row || !bytes.Equal(p[0x4511:0x5191], s.glyphBuffer) {
		s.reset("row start")
		return
	}
	raw := o.Indexed()
	lo := s.block.top * 320
	hi := (s.block.top + s.block.height) * 320
	if len(raw) != 64000 || (s.completed != nil && !bytes.Equal(raw[lo:hi], s.completed)) {
		s.reset("row prior frame")
		return
	}
	back := o.Bytes(oracle.Far(0x21b1, uint16((s.block.top+1)*320)), s.block.height*320)
	s.expected = make([]byte, len(back))
	for i, v := range back {
		s.expected[i] = (raw[lo+i] & 0x99) | (v & 0x66)
	}
	s.row = r.AX
	s.pending = true
}
func (s *Scroll) endRow(o *oracle.Oracle) {
	if !s.active || !s.pending {
		return
	}
	_, ok := s.sourceSnapshot()
	raw := o.Indexed()
	lo := s.block.top * 320
	hi := (s.block.top + s.block.height) * 320
	if !ok || o.DSReg() != 0x1b1 || len(raw) != 64000 || !bytes.Equal(raw[lo:hi], s.expected) {
		s.reset("row completion")
		return
	}
	if s.row == 0 && s.newLine {
		if s.current.text != "" {
			s.lines = append(s.lines, scrollLine{s.current.key, s.current.text, s.block.top + s.block.height})
		}
		s.newLine = false
	}
	for i := range s.lines {
		s.lines[i].y--
	}
	for len(s.lines) > 0 && s.lines[0].y+10 <= s.block.top {
		s.lines = s.lines[1:]
	}
	s.completed = append(s.completed[:0], s.expected...)
	s.pending = false
	s.Rows++
	s.sourceRows++
	s.blockRows[s.block.name]++
	if s.ending && s.sourceRows == 10+s.block.height {
		if len(s.lines) != 0 {
			s.reset("tail lines remain")
			return
		}
		s.BlocksCompleted++
		s.reset("")
	}
}
func textWidth(font bitmap.Font, text string) int {
	w := 0
	for _, ch := range text {
		w += len(font[ch]) / 16 * 12
	}
	return w
}

// Draw直接在三倍畫布取樣24像素字模，僅替換已驗完成的文字視窗。
func (s *Scroll) Draw(dst *image.RGBA, indexed []byte, palette [256][3]uint8, origin image.Point) bool {
	if s == nil || !s.active || s.block == nil || dst == nil {
		return false
	}
	top, bottom := s.block.top, s.block.top+s.block.height
	box := image.Rect(0, top*3, 960, bottom*3).Add(origin)
	if !box.In(dst.Bounds()) || len(s.completed) != s.block.height*320 {
		return false
	}
	if len(indexed) != 64000 {
		s.reset("display size")
		return false
	}
	if _, ok := s.sourceSnapshot(); !ok {
		s.reset("display source")
		return false
	}
	region := indexed[top*320 : bottom*320]
	for i, v := range region {
		if v == s.completed[i] {
			continue
		}
		if s.pending && (v == s.expected[i] || v == (s.completed[i]&0x99)) {
			continue
		}
		s.reset("unobserved display write")
		return false
	}
	for y := top; y < bottom; y++ {
		for x := 0; x < 320; x++ {
			v := s.completed[(y-top)*320+x] & 0x99
			c := palette[v]
			for yy := 0; yy < 3; yy++ {
				for xx := 0; xx < 3; xx++ {
					dst.SetRGBA(origin.X+x*3+xx, origin.Y+y*3+yy, color.RGBA{c[0], c[1], c[2], 255})
				}
			}
		}
	}
	ink := palette[0x66]
	for _, line := range s.lines {
		x := origin.X + (960-textWidth(s.font, line.text))/2
		y := origin.Y + line.y*3 + 3
		for _, ch := range line.text {
			glyph := s.font[ch]
			gw := len(glyph) / 16 * 8
			dw := gw * 3 / 2
			for gy := 0; gy < 24; gy++ {
				for gx := 0; gx < dw; gx++ {
					sx, sy := gx*2/3, gy*2/3
					at := image.Pt(x+gx, y+gy)
					if at.In(box) && glyph[sy*(gw/8)+sx/8]&(128>>uint(sx%8)) != 0 {
						dst.SetRGBA(at.X, at.Y, color.RGBA{ink[0], ink[1], ink[2], 255})
					}
				}
			}
			x += dw
		}
	}
	return true
}
func (s *Scroll) State() map[string]any {
	if s == nil {
		return nil
	}
	name := ""
	top, height := 0, 0
	if s.block != nil {
		name = s.block.name
		top = s.block.top
		height = s.block.height
	}
	keys := []string{}
	for _, line := range s.lines {
		keys = append(keys, line.key)
	}
	counts := map[string]int{}
	for k, v := range s.blockRows {
		counts[k] = v
	}
	return map[string]any{"active": s.active, "block": name, "top": top, "height": height, "source_file": s.file, "source_key": s.current.key, "source_rows": s.sourceRows, "pending": s.pending, "visible_keys": keys, "rows": s.Rows, "sources": s.Sources, "finished_blocks": s.BlocksCompleted, "block_rows": counts, "rejected": s.Rejected, "last_error": s.lastError}
}

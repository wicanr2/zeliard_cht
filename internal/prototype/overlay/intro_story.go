package overlay

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

// IntroStory只讀原版的文字與清框，規格014記錄來源及停止線。
type IntroStory struct {
	o                                   *oracle.Oracle
	font                                bitmap.Font
	pool                                []byte
	entries                             []storyEntry
	chars                               map[int]storyChar
	clearPages                          map[int]int
	done                                []int
	nextFile, lastFile, page            int
	lastByte                            byte
	tracking, ready, finished           bool
	completed, shadow, full             []byte
	pending                             byte // 1字形、2清框；都是原版自己的批次。
	pendingChar                         byte
	pendingFile                         int
	Glyphs, Clears, Fragments, Rejected int
	lastError                           string
}

func NewIntroStory(o *oracle.Oracle, sar, catalog string, font bitmap.Font) (*IntroStory, error) {
	if o == nil {
		return nil, fmt.Errorf("故事缺少原版執行器")
	}
	data, err := os.ReadFile(sar)
	if err != nil {
		return nil, err
	}
	if threatHash(data) != "56a73b59e952dab47fb0171a3fb1cdb7d641baceafa97e2826057f877a590eb0" {
		return nil, fmt.Errorf("故事來源版本不符")
	}
	entries, chars, clears, err := parseStory(data)
	if err != nil {
		return nil, err
	}
	text, err := os.ReadFile(catalog)
	if err != nil {
		return nil, err
	}
	if err = loadStoryText(entries, text, font); err != nil {
		return nil, err
	}
	s := &IntroStory{o: o, font: font, pool: append([]byte(nil), data[storyFileStart:storyFileEnd]...), entries: entries, chars: chars, clearPages: clears, done: make([]int, len(entries))}
	o.OnCall(oracle.Far(0x1b1, 0x6a8b), s.readByte)
	o.OnCall(oracle.Far(0x1b1, 0x6ade), s.beginGlyph)
	o.OnCall(oracle.Far(0x1b1, 0x6aef), s.endGlyph)
	o.OnCall(oracle.Far(0x1b1, 0x6bfd), s.beginClear)
	o.OnCall(oracle.Far(0x1b1, 0x6c0a), s.endClear)
	return s, nil
}
func storyWord(p []byte, i int) uint16 { return uint16(p[i]) | uint16(p[i+1])<<8 }
func storyRegion(raw []byte) []byte {
	if len(raw) != 64000 {
		return nil
	}
	return raw[143*320 : 200*320]
}
func (s *IntroStory) source() ([]byte, bool) {
	if s.o.VideoMode() != 0x13 {
		return nil, false
	}
	p := s.o.Bytes(oracle.Far(0x1b1, 0), 65536)
	if p[0xff77] != 255 || storyWord(p, 0xf500) != 0xf506 || !bytes.Equal(p[0x79c6:0x9060], s.pool) {
		return nil, false
	}
	for _, g := range []struct {
		lo, hi int
		sha    string
	}{{0x6000, 0x6040, "0dc4cbac83b93284bdb4c90cb875a07f768bfd45e833dfd135a386f610a09e3e"}, {0x6a75, 0x6cc4, "e1c479c686b20282701e9928caf016841931a5876250e0fc4dd5659492563aa5"}, {0x27e9, 0x2857, "e2e41fc888266454bb405a02919fee48682538bb741964383da854a83b559fca"}, {0x947d, 0x953d, "15eda0237134c09a79b7a068929c292ef9161f96c9f46d3b79b5555b5bfe9cf9"}, {0xf506, 0xf806, "8ff0729de496f9d84d4cb3e4931a349a028981232507939f23af95a821d955d6"}} {
		if threatHash(p[g.lo:g.hi]) != g.sha {
			return nil, false
		}
	}
	return p, true
}
func (s *IntroStory) halt(reason string) {
	if s.tracking && !s.finished && reason != "" {
		s.Rejected++
		s.lastError = reason
	}
	s.tracking = false
	s.ready = false
	s.pending = 0
	s.completed = nil
	s.shadow = nil
	s.full = nil
}
func (s *IntroStory) start() {
	s.tracking = true
	s.ready = false
	s.finished = false
	s.nextFile = storyFileStart
	s.lastFile = 0
	s.page = 0
	s.pending = 0
	s.completed = make([]byte, 18240)
	s.shadow = nil
	s.full = nil
	s.done = make([]int, len(s.entries))
	s.Glyphs = 0
	s.Clears = 0
	s.Fragments = 0
	s.Rejected = 0
	s.lastError = ""
}
func (s *IntroStory) readByte(o *oracle.Oracle) {
	p, ok := s.source()
	if !ok || o.DSReg() != 0x1b1 {
		s.halt("byte source")
		return
	}
	ptr := storyWord(p, 0x6d56)
	file := int(ptr) - 1 - 0x6000 + 0xa4
	if file == storyFileStart {
		s.start()
	}
	if !s.tracking {
		return
	}
	value := byte(o.Regs().AX)
	if file != s.nextFile || file < storyFileStart || file >= storyFileEnd || value != s.pool[file-storyFileStart] || o.SI() != ptr || s.pending != 0 {
		s.halt("byte order")
		return
	}
	s.lastFile = file
	s.lastByte = value
	s.nextFile++
	if value == 0xff {
		s.finished = true
	}
}

func storyPaint(dst, glyph []byte, x, y int, ink byte) bool {
	if len(dst) != 18240 || len(glyph) != 8 {
		return false
	}
	for gy, b := range glyph {
		for gx := 0; gx < 8; gx++ {
			if b&(128>>gx) != 0 {
				at := (y+gy-143)*320 + x + gx
				if at < 0 || at >= len(dst) {
					return false
				}
				dst[at] = ink
			}
		}
	}
	return true
}
func storyStyle(style byte) (byte, byte) {
	switch style {
	case 0xfa:
		return 0, 7
	case 0xfb:
		return 1, 7
	case 0xf9:
		return 2, 6
	}
	return 255, 255
}
func (s *IntroStory) beginGlyph(o *oracle.Oracle) {
	if !s.tracking {
		return
	}
	p, ok := s.source()
	r := o.Regs()
	if !ok || o.DSReg() != 0x1b1 || s.pending != 0 || s.lastByte < 32 || s.lastByte >= 127 || byte(r.AX) != s.lastByte || int(storyWord(p, 0x6d56)) != s.lastFile-0xa4+0x6001 || !bytes.Equal(storyRegion(o.Indexed()), s.completed) {
		s.halt("glyph start")
		return
	}
	if s.lastFile != storyFileStart {
		c, known := s.chars[s.lastFile]
		if !known {
			s.halt("unknown glyph")
			return
		}
		e := s.entries[c.entry]
		shadow, front := storyStyle(e.style)
		if e.page != s.page || c.index != s.done[c.entry] || p[0x6d5b] != shadow || p[0x6d5c] != front {
			s.halt("glyph identity or style")
			return
		}
	} else if s.lastByte != 'P' || p[0x6d5b] != 0 || p[0x6d5c] != 0 {
		s.halt("initial glyph")
		return
	}
	x, y := int(r.BX)-1, int(byte(r.CX))-1
	if x < 0 || x >= 320 || y < 143 || y > 183 {
		s.halt("glyph bounds")
		return
	}
	glyph := p[0xf506+int(s.lastByte-32)*8 : 0xf506+int(s.lastByte-32)*8+8]
	s.shadow = append([]byte(nil), s.completed...)
	if !storyPaint(s.shadow, glyph, x+1, y+1, p[0x6d5b]*17) {
		s.halt("shadow bounds")
		return
	}
	s.full = append([]byte(nil), s.shadow...)
	if !storyPaint(s.full, glyph, x, y, p[0x6d5c]*17) {
		s.halt("foreground bounds")
		return
	}
	s.pending = 1
	s.pendingChar = s.lastByte
	s.pendingFile = s.lastFile
}
func (s *IntroStory) endGlyph(o *oracle.Oracle) {
	if !s.tracking || s.pending != 1 {
		return
	}
	p, ok := s.source()
	r := o.Regs()
	stack := o.Bytes(oracle.Far(r.SS, r.SP), 2)
	if !ok || o.DSReg() != 0x1b1 || stack[0] != s.pendingChar || int(storyWord(p, 0x6d56)) != s.pendingFile-0xa4+0x6001 || !bytes.Equal(storyRegion(o.Indexed()), s.full) {
		s.halt("glyph completion")
		return
	}
	s.completed = append(s.completed[:0], s.full...)
	s.pending = 0
	s.Glyphs++
	if c, ok := s.chars[s.pendingFile]; ok {
		s.done[c.entry]++
		if s.done[c.entry] == len(s.entries[c.entry].files) {
			s.Fragments++
		}
	}
}
func (s *IntroStory) beginClear(o *oracle.Oracle) {
	if !s.tracking {
		return
	}
	_, ok := s.source()
	next, known := s.clearPages[s.lastFile]
	if !ok || o.DSReg() != 0x1b1 || s.pending != 0 || s.lastByte != 0xfe || !known || next != s.page+1 || !bytes.Equal(storyRegion(o.Indexed()), s.completed) {
		s.halt("clear start")
		return
	}
	s.full = make([]byte, 18240)
	s.shadow = nil
	s.pending = 2
	s.pendingFile = s.lastFile
}
func (s *IntroStory) endClear(o *oracle.Oracle) {
	if !s.tracking || s.pending != 2 {
		return
	}
	p, ok := s.source()
	if !ok || o.DSReg() != 0x1b1 || int(storyWord(p, 0x6d56)) != s.pendingFile-0xa4+0x6001 || !bytes.Equal(storyRegion(o.Indexed()), s.full) {
		s.halt("clear completion")
		return
	}
	s.completed = append(s.completed[:0], s.full...)
	s.pending = 0
	s.page = s.clearPages[s.pendingFile]
	s.Clears++
	s.ready = true
}
func (s *IntroStory) pageEntries() []int {
	var out []int
	for i, e := range s.entries {
		if e.page == s.page {
			out = append(out, i)
		}
	}
	return out
}

// Draw只替換已完整驗證的文字區，原版原始畫格及CPU狀態不變。
func (s *IntroStory) Draw(dst *image.RGBA, raw []byte, palette [256][3]uint8, origin image.Point) bool {
	if s == nil || !s.tracking || !s.ready || dst == nil || !image.Rect(0, 429, 960, 600).Add(origin).In(dst.Bounds()) {
		return false
	}
	if _, ok := s.source(); !ok {
		s.halt("display source")
		return false
	}
	region := storyRegion(raw)
	if len(region) != 18240 {
		s.halt("display size")
		return false
	}
	for i, v := range region {
		if v == s.completed[i] {
			continue
		}
		if s.pending != 0 && v == s.full[i] {
			continue
		}
		if s.pending == 1 && v == s.shadow[i] {
			continue
		}
		s.halt("unobserved display write")
		return false
	}
	entries := s.pageEntries()
	if len(entries) == 0 {
		return false
	}
	bg := palette[0]
	for y := 429; y < 600; y++ {
		for x := 0; x < 960; x++ {
			dst.SetRGBA(origin.X+x, origin.Y+y, color.RGBA{bg[0], bg[1], bg[2], 255})
		}
	}
	for _, i := range entries {
		e := s.entries[i]
		n := s.done[i] * len(e.text) / len(e.files)
		shadow, front := storyStyle(e.style)
		for _, ink := range []struct {
			dx, dy int
			value  byte
		}{{3, 3, shadow * 17}, {0, 0, front * 17}} {
			c := palette[ink.value]
			for _, p := range e.positions {
				if p.ordinal >= n {
					continue
				}
				glyph := s.font[p.character]
				gw := len(glyph) / 16 * 8
				for y := 0; y < 24; y++ {
					for x := 0; x < gw*3/2; x++ {
						sx, sy := x*2/3, y*2/3
						if glyph[sy*(gw/8)+sx/8]&(128>>uint(sx%8)) != 0 {
							at := p.at.Add(origin).Add(image.Pt(x+ink.dx, y+ink.dy))
							dst.SetRGBA(at.X, at.Y, color.RGBA{c[0], c[1], c[2], 255})
						}
					}
				}
			}
		}
	}
	return true
}
func (s *IntroStory) State() map[string]any {
	if s == nil {
		return nil
	}
	done, total := 0, 0
	var keys []string
	for _, i := range s.pageEntries() {
		done += s.done[i]
		total += len(s.entries[i].files)
		keys = append(keys, s.entries[i].key)
	}
	return map[string]any{"active": s.tracking && s.ready && total > 0, "page": s.page, "page_native_done": done, "page_native_total": total, "page_complete": total > 0 && done == total, "keys": keys, "glyphs": s.Glyphs, "clears": s.Clears, "fragments": s.Fragments, "pending": s.pending, "finished": s.finished, "last_source_file": s.lastFile, "rejected": s.Rejected, "last_error": s.lastError, "scope": "固定故事58片段，原始清框／字形來源；其餘片頭另驗"}
}

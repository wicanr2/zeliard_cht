package overlay

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"os"
	"unicode"
	"unicode/utf8"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

// IntroThreat只觀察原版宣言，契約見規格013。
type IntroThreat struct {
	o                        *oracle.Oracle
	font                     bitmap.Font
	lines                    []threatLine
	chars                    []threatChar
	completed, shadow, final []byte
	done                     int
	active, pending          bool
	Glyphs                   int
}
type threatLine struct {
	start, end, x, y, count int
	raw                     []byte
	text                    []rune
}
type threatChar struct {
	source            uint16
	value             byte
	x, y, line, index int
}

var threatBox = image.Rect(0, 148, 320, 177)

func NewIntroThreat(o *oracle.Oracle, sar, catalog string, font bitmap.Font) (*IntroThreat, error) {
	if o == nil {
		return nil, fmt.Errorf("宣言缺少原版執行器")
	}
	data, err := os.ReadFile(sar)
	if err != nil {
		return nil, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "56a73b59e952dab47fb0171a3fb1cdb7d641baceafa97e2826057f877a590eb0" {
		return nil, fmt.Errorf("宣言來源版本不符")
	}
	b, err := os.ReadFile(catalog)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(b) {
		return nil, fmt.Errorf("宣言須為UTF-8")
	}
	r := csv.NewReader(bytes.NewReader(b))
	r.Comma = '\t'
	r.FieldsPerRecord = 2
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) != 4 || rows[0][0] != "key" || rows[0][1] != "zh-TW" {
		return nil, fmt.Errorf("宣言須含三行及key／zh-TW欄位")
	}
	s := &IntroThreat{o: o, font: font}
	for i, v := range [][5]int{{0x313e, 0x315f, 64, 148, 24}, {0x3163, 0x318a, 48, 158, 28}, {0x318e, 0x31c0, 16, 168, 36}} {
		if rows[i+1][0] != fmt.Sprintf("zelres1.file.%06x", v[0]) {
			return nil, fmt.Errorf("宣言來源鍵或順序不符")
		}
		text := rows[i+1][1]
		if text == "" {
			return nil, fmt.Errorf("宣言譯文不可空白")
		}
		for _, ch := range text {
			if unicode.IsControl(ch) || (len(font[ch]) != 16 && len(font[ch]) != 32) {
				return nil, fmt.Errorf("宣言缺字或控制碼：%U", ch)
			}
		}
		if textWidth(font, text)+6 > v[4]*24 {
			return nil, fmt.Errorf("宣言超出安全寬度")
		}
		line := threatLine{v[0], v[1], v[2], v[3], v[4], append([]byte(nil), data[v[0]:v[1]]...), []rune(text)}
		n := 0
		for file := v[0]; file < v[1]; file++ {
			ch := data[file]
			if ch >= 1 && ch <= 4 {
				continue
			}
			if ch < 32 || ch >= 127 {
				return nil, fmt.Errorf("宣言來源含未知控制碼")
			}
			s.chars = append(s.chars, threatChar{uint16(file - 0xa4 + 0x6000), ch, v[2] + 8*n, v[3], i, n})
			n++
		}
		if n != v[4] {
			return nil, fmt.Errorf("宣言原文字數不符")
		}
		s.lines = append(s.lines, line)
	}
	o.OnCall(oracle.Far(0x1b1, 0x631e), s.begin)
	o.OnCall(oracle.Far(0x1b1, 0x6346), s.end)
	return s, nil
}
func (s *IntroThreat) reset() {
	s.active = false
	s.pending = false
	s.done = 0
	s.completed = nil
	s.shadow = nil
	s.final = nil
}
func threatHash(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func (s *IntroThreat) source() ([]byte, bool) {
	if s.o.VideoMode() != 0x13 {
		return nil, false
	}
	p := s.o.Bytes(oracle.Far(0x1b1, 0), 65536)
	if p[0xff77] != 255 || uint16(p[0xf500])|uint16(p[0xf501])<<8 != 0xf506 {
		return nil, false
	}
	for _, g := range []struct {
		lo, hi int
		sha    string
	}{{0x6000, 0x6040, "0dc4cbac83b93284bdb4c90cb875a07f768bfd45e833dfd135a386f610a09e3e"}, {0x62d1, 0x6358, "d925e4b319917cd29b4a282f36b60cdefcfea3b56ad96a350ea871ca88f978cc"}, {0x27e9, 0x2857, "e2e41fc888266454bb405a02919fee48682538bb741964383da854a83b559fca"}, {0xf506, 0xf806, "8ff0729de496f9d84d4cb3e4931a349a028981232507939f23af95a821d955d6"}} {
		if threatHash(p[g.lo:g.hi]) != g.sha {
			return nil, false
		}
	}
	for _, l := range s.lines {
		start := l.start - 0xa4 + 0x6000
		if !bytes.Equal(p[start:start+len(l.raw)], l.raw) {
			return nil, false
		}
	}
	return p, true
}
func threatRegion(raw []byte) []byte {
	if len(raw) != 64000 {
		return nil
	}
	return raw[148*320 : 177*320]
}
func (s *IntroThreat) begin(o *oracle.Oracle) {
	p, ok := s.source()
	if !ok || o.DSReg() != 0x1b1 {
		s.reset()
		return
	}
	if o.SI() == s.chars[0].source+1 {
		s.reset()
		s.completed = make([]byte, 9280)
		s.active = true
	}
	if !s.active || s.pending || s.done >= len(s.chars) {
		s.reset()
		return
	}
	c := s.chars[s.done]
	r := o.Regs()
	x := int(p[0x653d]) | int(p[0x653e])<<8
	if o.SI() != c.source+1 || byte(r.AX) != c.value || x != c.x || int(p[0x653f]) != c.y || !bytes.Equal(threatRegion(o.Indexed()), s.completed) {
		s.reset()
		return
	}
	s.shadow = append([]byte(nil), s.completed...)
	s.final = append([]byte(nil), s.completed...)
	glyph := p[0xf506+int(c.value-32)*8 : 0xf506+int(c.value-32)*8+8]
	paint := func(dst []byte, x, y int, ink byte) {
		for gy, b := range glyph {
			for gx := 0; gx < 8; gx++ {
				if b&(0x80>>gx) != 0 {
					dst[(y-148+gy)*320+x+gx] = ink
				}
			}
		}
	}
	paint(s.shadow, c.x+2, c.y+1, 34)
	copy(s.final, s.shadow)
	paint(s.final, c.x, c.y, 119)
	s.pending = true
}
func (s *IntroThreat) end(o *oracle.Oracle) {
	if !s.pending {
		return
	}
	_, ok := s.source()
	r := o.Regs()
	stack := o.Bytes(oracle.Far(r.SS, r.SP), 4)
	c := s.chars[s.done]
	if !ok || o.DSReg() != 0x1b1 || uint16(stack[0])|uint16(stack[1])<<8 != c.source+1 || stack[2] != c.value || !bytes.Equal(threatRegion(o.Indexed()), s.final) {
		s.reset()
		return
	}
	s.completed = append(s.completed[:0], s.final...)
	s.done++
	s.Glyphs++
	s.pending = false
}

// Draw的字形途中只接受已證實的三個原版寫入階段。所有守衛先於畫布寫入。
func (s *IntroThreat) Draw(dst *image.RGBA, raw []byte, palette [256][3]uint8, origin image.Point) bool {
	if s == nil || !s.active || dst == nil || !image.Rect(0, 444, 960, 531).Add(origin).In(dst.Bounds()) {
		return false
	}
	if _, ok := s.source(); !ok {
		s.reset()
		return false
	}
	actual := threatRegion(raw)
	if len(actual) != 9280 {
		s.reset()
		return false
	}
	for i, v := range actual {
		if v != s.completed[i] && (!s.pending || (v != s.shadow[i] && v != s.final[i])) {
			s.reset()
			return false
		}
	}
	bg := palette[0]
	for y := 444; y < 531; y++ {
		for x := 0; x < 960; x++ {
			dst.SetRGBA(origin.X+x, origin.Y+y, color.RGBA{bg[0], bg[1], bg[2], 255})
		}
	}
	for i, l := range s.lines {
		n := 0
		for k := 0; k < s.done; k++ {
			if s.chars[k].line == i {
				n++
			}
		}
		full := string(l.text)
		count := n * len(l.text) / l.count
		text := string(l.text[:count])
		x := origin.X + l.x*3 + (l.count*24-textWidth(s.font, full))/2
		y := origin.Y + l.y*3
		for _, ink := range []struct {
			dx, dy int
			value  byte
		}{{6, 3, 34}, {0, 0, 119}} {
			c := palette[ink.value]
			xx := x
			for _, ch := range text {
				g := s.font[ch]
				gw := len(g) / 16 * 8
				dw := gw * 3 / 2
				for gy := 0; gy < 24; gy++ {
					for gx := 0; gx < dw; gx++ {
						sx, sy := gx*2/3, gy*2/3
						if g[sy*(gw/8)+sx/8]&(0x80>>uint(sx%8)) != 0 {
							dst.SetRGBA(xx+gx+ink.dx, y+gy+ink.dy, color.RGBA{c[0], c[1], c[2], 255})
						}
					}
				}
				xx += dw
			}
		}
	}
	return true
}
func (s *IntroThreat) State() map[string]any {
	if s == nil {
		return nil
	}
	return map[string]any{"active": s.active, "pending": s.pending, "completed": s.done, "glyphs": s.Glyphs, "scope": "原版魔王三行宣言；其他片頭另驗"}
}

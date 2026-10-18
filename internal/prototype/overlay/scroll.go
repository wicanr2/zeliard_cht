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

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

// Scroll 只觀察片頭首章的來源與完成繪製事件。原版仍負責捲動與時序。
// 定位以 IDA base=0 的 16-bit CS 偏移為準，見 docs/source-index.json。
type Scroll struct {
	font                     bitmap.Font
	entries                  map[uint16]scrollEntry
	lines                    []scrollLine
	current                  scrollEntry
	row                      uint16
	active, pending, newLine bool
	completed                []byte
	Rows, Sources            int
}
type scrollEntry struct {
	key, text string
	raw       []byte
}
type scrollLine struct {
	key, text string
	y         int
}

// NewScroll 的 TSV 只保存來源鍵與試譯。原文從使用者本地資料讀取。
func NewScroll(o *oracle.Oracle, sar, catalog string, font bitmap.Font) (*Scroll, error) {
	data, err := os.ReadFile(sar)
	if err != nil {
		return nil, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "56a73b59e952dab47fb0171a3fb1cdb7d641baceafa97e2826057f877a590eb0" {
		return nil, fmt.Errorf("片頭來源版本不符，保留原文")
	}
	f, err := os.Open(catalog)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = '\t'
	r.FieldsPerRecord = 2
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) != 31 || records[0][0] != "key" || records[0][1] != "zh-TW" {
		return nil, fmt.Errorf("首章試譯須有 30 行及 key／zh-TW 欄位")
	}
	s := &Scroll{font: font, entries: map[uint16]scrollEntry{}}
	off := 0x1094
	for _, record := range records[1:] {
		key := fmt.Sprintf("zelres1.file.%06x", off)
		if record[0] != key {
			return nil, fmt.Errorf("來源行鍵不符：%s，預期 %s", record[0], key)
		}
		n := bytes.IndexByte(data[off:], '\r')
		if n < 0 || n > 40 {
			return nil, fmt.Errorf("片頭行分隔不符")
		}
		text := record[1]
		if strings.TrimSpace(string(data[off:off+n])) == "" && text != "" {
			return nil, fmt.Errorf("空白來源行不可新增敘事")
		}
		for _, ch := range text {
			if len(font[ch]) != 16 && len(font[ch]) != 32 {
				return nil, fmt.Errorf("片頭缺字：%U", ch)
			}
		}
		if textWidth(font, text) > 900 {
			return nil, fmt.Errorf("片頭試譯超出安全寬度：%s", key)
		}
		s.entries[uint16(off-0xa4+0x6000)] = scrollEntry{key: key, text: text, raw: append([]byte(nil), data[off:off+n+1]...)}
		off += n + 1
	}
	if off != 0x13db || data[off] != 0xff {
		return nil, fmt.Errorf("首章結尾不符")
	}
	o.OnCall(oracle.Far(0x01b1, 0x32c9), s.source)
	o.OnCall(oracle.Far(0x01b1, 0x332c), s.beginRow)
	o.OnCall(oracle.Far(0x01b1, 0x33b6), s.endRow)
	return s, nil
}

func (s *Scroll) reset() { s.active = false; s.pending = false; s.lines = nil; s.completed = nil }
func (s *Scroll) source(o *oracle.Oracle) {
	// 結束碼仍交給原版處理。接下來的空白列把已顯示中文捲出視窗。
	if s.active && o.DSReg() == 0x01b1 && o.SI() == 0x7337 && o.Byte(oracle.Far(0x01b1, 0x7337)) == 0xff {
		s.current = scrollEntry{}
		s.newLine = false
		return
	}
	e, ok := s.entries[o.SI()]
	if !ok || o.DSReg() != 0x01b1 || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0x32c9), 6), []byte{0x0e, 0x07, 0xbf, 0x11, 0x45, 0x33}) || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, o.SI()), len(e.raw)), e.raw) {
		s.reset()
		return
	}
	if o.SI() == 0x6ff0 {
		s.reset()
	}
	s.active = true
	s.current = e
	s.newLine = true
	s.Sources++
}
func (s *Scroll) beginRow(o *oracle.Oracle) {
	if !s.active {
		return
	}
	if o.BX() != 0x0020 || o.CX() != 0x5078 || o.AX() > 9 || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0x332c), 5), []byte{0x1e, 0x51, 0x53, 0xb2, 0xa0}) {
		s.reset()
		return
	}
	s.row = o.AX()
	s.pending = true
}
func (s *Scroll) endRow(o *oracle.Oracle) {
	if !s.active || !s.pending {
		return
	}
	if o.Byte(oracle.Far(0x01b1, 0x33b6)) != 0xc3 {
		s.reset()
		return
	}
	if s.row == 0 && s.newLine {
		if s.current.text != "" {
			s.lines = append(s.lines, scrollLine{s.current.key, s.current.text, 152})
		}
		s.newLine = false
	}
	for i := range s.lines {
		s.lines[i].y--
	}
	for len(s.lines) > 0 && s.lines[0].y+10 <= 32 {
		s.lines = s.lines[1:]
	}
	s.completed = append(s.completed[:0], o.Indexed()[32*320:152*320]...)
	s.pending = false
	s.Rows++
}

func textWidth(font bitmap.Font, text string) int {
	w := 0
	for _, ch := range text {
		w += len(font[ch]) / 16 * 12
	}
	return w
}

// Draw 直接在三倍畫布繪字。每字 16 點轉成 24 顯示像素，不先降成 DOS 字模。
// 原版 33A0/33A4/33A6 將背景保留於 99h，文字放於 66h；只清除文字位元。
// 單次 VGA 複製尚未完成時，暫留上一個完整 viewport，避免半行英文閃現。
func (s *Scroll) Draw(dst *image.RGBA, indexed []byte, palette [256][3]uint8, origin image.Point) bool {
	if !s.active || len(s.completed) != 120*320 || len(indexed) != 320*200 {
		return false
	}
	if !s.pending && !bytes.Equal(s.completed, indexed[32*320:152*320]) {
		s.reset()
		return false
	}
	clip := image.Rect(origin.X, origin.Y+32*3, origin.X+960, origin.Y+152*3).Intersect(dst.Bounds())
	for y := 32; y < 152; y++ {
		for x := 0; x < 320; x++ {
			v := s.completed[(y-32)*320+x] & 0x99
			c := palette[v]
			for yy := 0; yy < 3; yy++ {
				for xx := 0; xx < 3; xx++ {
					p := image.Pt(origin.X+x*3+xx, origin.Y+y*3+yy)
					if p.In(clip) {
						dst.SetRGBA(p.X, p.Y, color.RGBA{c[0], c[1], c[2], 255})
					}
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
					p := image.Pt(x+gx, y+gy)
					if p.In(clip) && glyph[sy*(gw/8)+sx/8]&(0x80>>uint(sx%8)) != 0 {
						dst.SetRGBA(p.X, p.Y, color.RGBA{ink[0], ink[1], ink[2], 255})
					}
				}
			}
			x += dw
		}
	}
	return true
}

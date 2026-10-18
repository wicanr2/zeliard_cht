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
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

type labelSource struct {
	offset int
	module int
	raw    []byte
	text   []rune
}
type labelStamp struct {
	source  *labelSource
	box     image.Rectangle
	raw     []byte
	palette [256][3]uint8
}

// Labels 只觀察已知 MCGA 繪字入口，來源與完成畫面不符時保留原貌。
// 選單選擇、金額、存檔名稱及原版輸入完全由原版處理。
type Labels struct {
	font      bitmap.Font
	data      []byte
	modules   []int
	sources   map[int]*labelSource
	stamps    []labelStamp
	pending   *labelStamp
	Completed int
}

func NewLabels(o *oracle.Oracle, sar, catalog string, font bitmap.Font) (*Labels, error) {
	data, err := os.ReadFile(sar)
	if err != nil {
		return nil, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1" {
		return nil, fmt.Errorf("選單來源版本不符")
	}
	l := &Labels{font: font, data: data, sources: map[int]*labelSource{}}
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
		l.modules = append(l.modules, p+4)
	}
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
		if !strings.HasPrefix(row[0], "zelres2.file.") {
			return nil, fmt.Errorf("選單鍵錯誤")
		}
		v, e := strconv.ParseUint(strings.TrimPrefix(row[0], "zelres2.file."), 16, 32)
		off := int(v)
		if e != nil || off >= len(data) || l.sources[off] != nil || row[1] == "" || strings.ContainsAny(row[1], "\t\r\n") {
			return nil, fmt.Errorf("選單鍵重複或文字錯誤：%s", row[0])
		}
		module := -1
		for _, p := range l.modules {
			end := p + int(binary.LittleEndian.Uint32(data[p-4:]))
			if off >= p && off < end {
				module = p
				break
			}
		}
		if module < 0 {
			return nil, fmt.Errorf("選單鍵不在資源項目")
		}
		end := off
		for end < len(data) && data[end] >= 0x20 && data[end] < 0x7f {
			end++
		}
		if end == off || end >= module+int(binary.LittleEndian.Uint32(data[module-4:])) || data[end] != 0 {
			return nil, fmt.Errorf("選單來源須為原版 ASCIIZ：%s", row[0])
		}
		width := 0
		for _, ch := range []rune(row[1]) {
			g := font[ch]
			if len(g) != 16 && len(g) != 32 {
				return nil, fmt.Errorf("選單字型缺字：%U", ch)
			}
			width += len(g) / 16 * 12
		}
		if width > (end-off)*15 {
			return nil, fmt.Errorf("選單譯文超出原文安全寬度：%s", row[0])
		}
		l.sources[off] = &labelSource{offset: off, module: module, raw: append([]byte(nil), data[off:end]...), text: []rune(row[1])}
	}
	o.OnCall(oracle.Far(0x01b1, 0x22db), l.begin)
	o.OnCall(oracle.Far(0x01b1, 0x2308), l.finish)
	return l, nil
}

func (l *Labels) begin(o *oracle.Oracle) {
	l.pending = nil
	if o.DSReg() != 0x01b1 || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0x22db), 6), []byte{0x2e, 0xc6, 0x06, 0xbd, 0x2c, 0x09}) {
		return
	}
	pos := int(o.SI())
	if pos < 0xa000 {
		return
	}
	for off, s := range l.sources {
		if off-s.module != pos-0xa000 {
			continue
		}
		// 模組固定前綴與完整來源字串均須相符，不能只靠當次 SI 認出文字。
		if !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0xa000), 64), l.data[s.module:s.module+64]) || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, uint16(pos)), len(s.raw)+1), append(append([]byte(nil), s.raw...), 0)) {
			continue
		}
		x := int(o.BX()>>8)*4 + int(o.CX()&255)
		y := int(o.BX() & 255)
		box := image.Rect(x, y, x+len(s.raw)*5, y+8)
		if box.Empty() || !box.In(image.Rect(0, 0, 320, 200)) {
			return
		}
		l.pending = &labelStamp{source: s, box: box}
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
func (l *Labels) finish(o *oracle.Oracle) {
	if l.pending == nil {
		return
	}
	s := *l.pending
	l.pending = nil
	s.raw = labelCrop(o.Indexed(), s.box)
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
		if pal != s.palette || !bytes.Equal(s.raw, labelCrop(raw, s.box)) {
			continue
		}
		kept = append(kept, s)
		drawn++
		box := image.Rect(origin.X+s.box.Min.X*3, origin.Y+s.box.Min.Y*3, origin.X+s.box.Max.X*3, origin.Y+s.box.Max.Y*3)
		bg, fg := pal[0], pal[9]
		for y := box.Min.Y; y < box.Max.Y; y++ {
			for x := box.Min.X; x < box.Max.X; x++ {
				if image.Pt(x, y).In(dst.Bounds()) {
					dst.SetRGBA(x, y, color.RGBA{bg[0], bg[1], bg[2], 255})
				}
			}
		}
		x := box.Min.X
		for _, ch := range s.source.text {
			g := l.font[ch]
			w := len(g) / 16 * 8
			dw := w * 3 / 2
			for yy := 0; yy < 24; yy++ {
				for xx := 0; xx < dw; xx++ {
					sx, sy := xx*2/3, yy*2/3
					if g[sy*(w/8)+sx/8]&(0x80>>uint(sx%8)) != 0 {
						p := image.Pt(x+xx, box.Min.Y+yy)
						if p.In(box) && p.In(dst.Bounds()) {
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

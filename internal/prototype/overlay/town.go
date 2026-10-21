package overlay

import (
	"bytes"
	"encoding/binary"
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

type townSource struct {
	key                 string
	start, tableAddress uint16
	index               int
	prefix, table, raw  []byte
	text                []rune
	progress            map[uint16]int
	total               int
}

// TownDialogue 只讀居民專用繪字端。原版保留腳本、輸入、續頁及背景還原。
// 來源與位址基準見 docs/spec/004-town-dialogue.md。
type TownDialogue struct {
	original                   *oracle.Oracle
	font                       bitmap.Font
	sources                    []*townSource
	current                    *townSource
	box, safe, marker          image.Rectangle
	positions                  []image.Point
	completed                  []byte
	visible                    int
	pending, prompt            bool
	Glyphs, Messages, Rejected int
	LastRejection              map[string]any
}

func NewTownDialogue(o *oracle.Oracle, sar, catalog string, font bitmap.Font) (*TownDialogue, error) {
	data, err := os.ReadFile(sar)
	if err != nil {
		return nil, err
	}
	if hudHash(data) != "bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1" {
		return nil, fmt.Errorf("居民對話來源版本不符")
	}
	tsv, err := os.ReadFile(catalog)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(tsv) {
		return nil, fmt.Errorf("居民對話 TSV 必須是 UTF-8")
	}
	r := csv.NewReader(bytes.NewReader(tsv))
	r.Comma = '\t'
	r.FieldsPerRecord = 2
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 || records[0][0] != "key" || records[0][1] != "zh-TW" {
		return nil, fmt.Errorf("居民對話 TSV 欄位不符")
	}
	texts := map[string]string{}
	for _, row := range records[1:] {
		if _, exists := texts[row[0]]; exists || row[0] == "" || row[1] == "" {
			return nil, fmt.Errorf("居民對話鍵重複或漏譯")
		}
		for _, ch := range row[1] {
			if unicode.IsControl(ch) {
				return nil, fmt.Errorf("居民對話含控制字元")
			}
		}
		texts[row[0]] = row[1]
	}
	d := &TownDialogue{original: o, font: font}
	// 固定版本的十個模組。C00D 指標表與單 FF 純文字終止由 63CA 消費端證實。
	for _, limits := range [][2]int{{0x38a1a, 0x392d8}, {0x392dc, 0x3a1d5}, {0x3a1d9, 0x3ae85}, {0x3ae89, 0x3bbdf}, {0x3bbe3, 0x3c9e3}, {0x3c9e7, 0x3d9d5}, {0x3d9d9, 0x3e80f}, {0x3e813, 0x3f928}, {0x3f92c, 0x409e7}, {0x409eb, 0x413ed}} {
		module, end := limits[0], limits[1]
		tableAddr := binary.LittleEndian.Uint16(data[module+13:])
		table := module + int(tableAddr) - 0xc000
		if table < module || table+2 > end {
			return nil, fmt.Errorf("居民指標表越界")
		}
		first := binary.LittleEndian.Uint16(data[table:])
		count := (int(first) - int(tableAddr)) / 2
		if first <= tableAddr || (first-tableAddr)%2 != 0 || count > 32 || table+count*2 > end {
			return nil, fmt.Errorf("居民指標表長度不符")
		}
		for index := 0; index < count; index++ {
			start := binary.LittleEndian.Uint16(data[table+index*2:])
			body := module + int(start) - 0xc000
			if body < module || body >= end {
				return nil, fmt.Errorf("居民腳本越界")
			}
			n := bytes.IndexByte(data[body:end], 0xff)
			if n < 0 {
				return nil, fmt.Errorf("居民腳本缺終止符")
			}
			raw := data[body : body+n+1]
			plain := n > 0
			for _, ch := range raw[:n] {
				plain = plain && ch >= 32 && ch < 127
			}
			if !plain {
				continue
			} // 混合高位控制碼不猜回呼語意。
			key := fmt.Sprintf("zelres2.file.%06x", body)
			text, exists := texts[key]
			if !exists {
				return nil, fmt.Errorf("居民對話漏譯：%s", key)
			}
			delete(texts, key)
			if _, err := townLayout(font, []rune(text), image.Rect(0, 0, 168, 78)); err != nil {
				return nil, fmt.Errorf("%s：%w", key, err)
			}
			s := &townSource{key: key, start: start, tableAddress: tableAddr, index: index, prefix: append([]byte(nil), data[module:module+64]...), table: append([]byte(nil), data[table:table+count*2]...), raw: append([]byte(nil), raw...), text: []rune(text), progress: map[uint16]int{}}
			for i, ch := range raw[:n] {
				if ch != '/' {
					s.total++
					s.progress[start+uint16(i)] = s.total
				}
			}
			d.sources = append(d.sources, s)
		}
	}
	if len(texts) != 0 || len(d.sources) != 113 {
		return nil, fmt.Errorf("居民 catalog 含孤兒鍵或來源數不符")
	}
	if o != nil {
		o.OnCall(oracle.Far(0x1b1, 0x63da), d.begin)
		o.OnCall(oracle.Far(0x1b1, 0x64b2), d.beginGlyph)
		o.OnCall(oracle.Far(0x1b1, 0x64b7), func(o *oracle.Oracle) {
			if d.current != nil && d.pending {
				d.Glyphs++
				d.capture(o)
			}
		})
		// 原版寫入可以跨過前端呈現點。pending涵蓋完整捲動及標記／背景寫入。
		for _, site := range []uint16{0x64fa, 0x654c, 0x658a, 0x639e} {
			o.OnCall(oracle.Far(0x1b1, site), func(*oracle.Oracle) {
				if d.current != nil {
					d.pending = true
				}
			})
		}
		o.OnCall(oracle.Far(0x1b1, 0x6516), func(o *oracle.Oracle) {
			if d.current != nil && !d.box.Empty() {
				d.capture(o)
			}
		})
		o.OnCall(oracle.Far(0x1b1, 0x6551), func(o *oracle.Oracle) {
			if d.current != nil {
				d.prompt = true
				d.capture(o)
			}
		})
		o.OnCall(oracle.Far(0x1b1, 0x658f), func(o *oracle.Oracle) {
			if d.current != nil {
				d.prompt = false
				d.capture(o)
			}
		})
		o.OnCall(oracle.Far(0x1b1, 0x65a1), func(o *oracle.Oracle) {
			if d.current != nil {
				d.visible = len(d.current.text)
				d.prompt = false
				d.capture(o)
			}
		})
		o.OnCall(oracle.Far(0x1b1, 0x63a3), func(*oracle.Oracle) { d.reset() })
	}
	return d, nil
}

func townCode(o *oracle.Oracle) bool {
	return o.DSReg() == 0x1b1 && townDisplayCode(o)
}

// 呈現可落在原版driver內。DS只約束消費端callback，固定段身份仍完整驗證。
func townDisplayCode(o *oracle.Oracle) bool {
	return o.VideoMode() == 0x13 &&
		hudHash(o.Bytes(oracle.Far(0x1b1, 0x63ca), 0x664d-0x63ca)) == "2011d284c03ced34946616a7dde903d3e9972d9af6e00301f78d4ea0c562581c" &&
		hudHash(o.Bytes(oracle.Far(0x1b1, 0x7b82), 192)) == "15eda0237134c09a79b7a068929c292ef9161f96c9f46d3b79b5555b5bfe9cf9"
}

func (s *townSource) valid(o *oracle.Oracle) bool {
	return bytes.Equal(o.Bytes(oracle.Far(0x1b1, 0xc000), len(s.prefix)), s.prefix) && bytes.Equal(o.Bytes(oracle.Far(0x1b1, s.tableAddress), len(s.table)), s.table) && bytes.Equal(o.Bytes(oracle.Far(0x1b1, s.start), len(s.raw)), s.raw)
}

func (d *TownDialogue) reset() {
	d.current = nil
	d.completed = nil
	d.positions = nil
	d.visible = 0
	d.pending = false
	d.prompt = false
	d.box, d.safe, d.marker = image.Rectangle{}, image.Rectangle{}, image.Rectangle{}
}

func (d *TownDialogue) begin(o *oracle.Oracle) {
	d.reset()
	if !townCode(o) {
		return
	}
	for _, s := range d.sources {
		if o.SI() != s.start || o.BX() != s.tableAddress+uint16(s.index*2) || !s.valid(o) {
			continue
		}
		if d.current != nil {
			d.reset()
			d.Rejected++
			return
		}
		d.current = s
	}
	if d.current != nil {
		d.Messages++
	}
}

func townBox(o *oracle.Oracle) (image.Rectangle, bool) {
	anchor, size := o.Word(oracle.Far(0x1b1, 0x7c4e)), o.Word(oracle.Far(0x1b1, 0x7c5a))
	w, h := int(size>>8)*4, int(size&255)
	x, y := int(anchor>>8)*8, int(anchor&255)
	box := image.Rect(x, y, x+w, y+h)
	return box, w == 176 && h >= 16 && h <= 86 && (h-6)%10 == 0 && (x == 56 || x == 88) && box.In(image.Rect(48, 24, 272, 148))
}

func (d *TownDialogue) beginGlyph(o *oracle.Oracle) {
	if d.current == nil {
		return
	}
	s := d.current
	pos := o.Word(oracle.Far(0x1b1, 0x7c58)) - 1
	count, ok := s.progress[pos]
	box, geometry := townBox(o)
	x, y := int(o.BX()), int(o.CX()&255) // 原版只改 CL，CH 不是 y。
	ch := o.Byte(oracle.Far(0x1b1, pos))
	// 64A6/64AE 會扣除字形左側空白；s 的原點可為59，文字安全區仍由60開始。
	expectedX := box.Min.X + 4 + int(o.Byte(oracle.Far(0x1b1, 0x7c53)))
	if ch >= 32 && ch < 127 {
		expectedX -= int(o.Byte(oracle.Far(0x1b1, 0x7b82+uint16(ch-32))))
	}
	expectedY := box.Min.Y + 4 + int(o.Byte(oracle.Far(0x1b1, 0x7c54)))*10
	if !ok || !geometry || !townCode(o) || !s.valid(o) || o.AX()>>8 != 1 || byte(o.AX()) != ch || x != expectedX || y != expectedY {
		d.LastRejection = map[string]any{"site": "glyph", "pos": pos, "ax": o.AX(), "bx": o.BX(), "cx": o.CX(), "ds": o.DSReg(), "source": s.key, "progress_valid": ok, "geometry_valid": geometry, "code_valid": townCode(o), "source_valid": s.valid(o), "box": box, "glyphs": d.Glyphs}
		d.Rejected++
		d.reset()
		return
	}
	safe := box.Inset(4)
	positions, err := townLayout(d.font, s.text, safe)
	if err != nil {
		d.Rejected++
		d.reset()
		return
	}
	d.box, d.safe = box, safe
	d.positions = positions
	d.marker = image.Rect(box.Min.X+84, box.Min.Y+74, box.Min.X+92, box.Min.Y+82)
	d.visible = (count*len(s.text) + s.total - 1) / s.total
	d.pending = true
}

func townCrop(indexed []byte, box image.Rectangle) []byte {
	if len(indexed) != 320*200 || box.Empty() || !box.In(image.Rect(0, 0, 320, 200)) {
		return nil
	}
	data := make([]byte, box.Dx()*box.Dy())
	for y := box.Min.Y; y < box.Max.Y; y++ {
		copy(data[(y-box.Min.Y)*box.Dx():], indexed[y*320+box.Min.X:y*320+box.Max.X])
	}
	return data
}

func (d *TownDialogue) capture(o *oracle.Oracle) {
	if !townCode(o) || d.current == nil || !d.current.valid(o) || d.box.Empty() {
		d.reset()
		return
	}
	d.completed = townCrop(o.Indexed(), d.box)
	d.pending = false
}

func townLayout(font bitmap.Font, text []rune, safe image.Rectangle) ([]image.Point, error) {
	x, y := 0, 0
	maxX, maxY := safe.Dx()*3, safe.Dy()*3
	points := make([]image.Point, 0, len(text))
	width := func(ch rune) int { return len(font[ch]) / 16 * 12 }
	latin := func(ch rune) bool { return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' }
	noStart := func(ch rune) bool { return strings.ContainsRune("，。！？；：、）］」』】〉》…", ch) }
	for i, ch := range text {
		if len(font[ch]) != 16 && len(font[ch]) != 32 {
			return nil, fmt.Errorf("居民字型缺字：%U", ch)
		}
		w := width(ch)
		if latin(ch) && (i == 0 || !latin(text[i-1])) {
			wordWidth, j := 0, i
			for ; j < len(text) && latin(text[j]); j++ {
				wordWidth += width(text[j])
			}
			for ; j < len(text) && noStart(text[j]); j++ {
				wordWidth += width(text[j])
			}
			if wordWidth > maxX {
				return nil, fmt.Errorf("居民專名超寬")
			}
			if x+wordWidth > maxX {
				x = 0
				y += 30
			}
		}
		needed := w
		for j := i + 1; j < len(text) && noStart(text[j]); j++ {
			needed += width(text[j])
		}
		if x+needed > maxX && !noStart(ch) || x+w > maxX {
			x = 0
			y += 30
		}
		if y+24 > maxY {
			return nil, fmt.Errorf("居民譯文超出當次安全區")
		}
		points = append(points, image.Pt(x, y))
		x += w
	}
	return points, nil
}

func (d *TownDialogue) State() map[string]any {
	key := ""
	if d.current != nil {
		key = d.current.key
	}
	return map[string]any{"key": key, "visible": d.visible, "pending": d.pending, "prompt": d.prompt, "box": d.box, "glyphs": d.Glyphs, "messages": d.Messages, "rejected": d.Rejected, "last_rejection": d.LastRejection}
}

func (d *TownDialogue) Draw(dst *image.RGBA, indexed []byte, palette [256][3]uint8, origin image.Point) bool {
	if d.current == nil || d.pending || len(d.completed) == 0 {
		return false
	}
	if d.original == nil || !townDisplayCode(d.original) || !d.current.valid(d.original) || !hudPalette(palette) || !bytes.Equal(d.completed, townCrop(indexed, d.box)) {
		d.Rejected++
		d.reset()
		return false
	}
	box := image.Rect(origin.X+d.safe.Min.X*3, origin.Y+d.safe.Min.Y*3, origin.X+d.safe.Max.X*3, origin.Y+d.safe.Max.Y*3)
	if !box.In(dst.Bounds()) {
		return false
	}
	marker := image.Rect(origin.X+d.marker.Min.X*3, origin.Y+d.marker.Min.Y*3, origin.X+d.marker.Max.X*3, origin.Y+d.marker.Max.Y*3)
	bg, fg := palette[0], palette[9]
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			if !d.prompt || !image.Pt(x, y).In(marker) {
				dst.SetRGBA(x, y, color.RGBA{bg[0], bg[1], bg[2], 255})
			}
		}
	}
	for i, ch := range d.current.text {
		if i >= d.visible {
			break
		}
		at := box.Min.Add(d.positions[i])
		d.font.DrawGlyph(dst, at, ch, 24, color.RGBA{fg[0], fg[1], fg[2], 255}, func(p image.Point) bool { return !d.prompt || !p.In(marker) })
	}
	return true
}

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
)

var dialogueBox = image.Rect(54, 98, 264, 150)
var dialoguePrompt = image.Rect(156, 139, 164, 147)

type dialogueParagraph struct {
	key        string
	start, end uint16
	text       []rune
	progress   map[uint16]int
	total      int
}

// Dialogue 是國王初訪兩段話的唯讀顯示試作。原版處理按鍵、捲頁與賞金。
type Dialogue struct {
	font                        bitmap.Font
	script                      []byte
	paragraphs                  []dialogueParagraph
	current, visible            int
	active, pending, prompt     bool
	completed                   []byte
	Glyphs, Paragraphs, Scrolls int
	Rejections                  []map[string]int
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
	if len(records) != 3 || records[0][0] != "key" || records[0][1] != "zh-TW" {
		return nil, fmt.Errorf("國王試譯須有兩筆及 key／zh-TW 欄位")
	}
	d := &Dialogue{font: font, current: -1, script: append([]byte(nil), data[0x17249:0x17356]...)}
	for i, bounds := range [][2]uint16{{0xa436, 0xa498}, {0xa4a1, 0xa537}} {
		key := fmt.Sprintf("zelres2.file.%06x", int(bounds[0])+0xce1a)
		text := records[i+1][1]
		if records[i+1][0] != key || text == "" || strings.ContainsAny(text, "\n\r\t") {
			return nil, fmt.Errorf("對話鍵或文字不符：%s", key)
		}
		if _, err := dialogueLayout(font, []rune(text)); err != nil {
			return nil, err
		}
		p := dialogueParagraph{key: key, start: bounds[0], end: bounds[1], text: []rune(text), progress: map[uint16]int{}}
		for pos := bounds[0]; pos < bounds[1]; pos++ {
			ch := data[int(pos)+0xce1a]
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
}
func (d *Dialogue) beginGlyph(o *oracle.Oracle) {
	pos := o.Word(oracle.Far(0x01b1, 0xff4c)) - 1
	idx := -1
	count := 0
	for i, p := range d.paragraphs {
		if n, ok := p.progress[pos]; ok {
			idx = i
			count = n
			break
		}
	}
	// 7118h 只更新 CL。正常翻頁可帶回 CH=2，不能把整個 CX 當作 Y。
	y := o.CX() & 0xff
	if idx < 0 || o.DSReg() != 0x01b1 || o.AX()>>8 != 1 || int(o.BX()) < 54 || int(o.BX()) >= 264 || y < 99 || y > 139 || byte(o.AX()) != o.Byte(oracle.Far(0x01b1, pos)) || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0x711b), 5), []byte{0x2e, 0xff, 0x16, 0x22, 0x20}) || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, 0xa42f), len(d.script)), d.script) {
		if (d.active || idx >= 0) && len(d.Rejections) < 16 {
			d.Rejections = append(d.Rejections, map[string]int{"kind": 1, "steps": int(o.Steps()), "source": int(pos), "index": idx, "ax": int(o.AX()), "bx": int(o.BX()), "cx": int(o.CX()), "ds": int(o.DSReg())})
		}
		d.reset()
		return
	}
	if d.current != idx {
		d.Paragraphs++
		d.current = idx
		d.prompt = false
	}
	d.active = true
	d.pending = true
	p := d.paragraphs[idx]
	d.visible = (count*len(p.text) + p.total - 1) / p.total
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
	d.completed = dialogueCrop(o.Indexed())
	d.pending = false
}

// dialogueLayout 用實際字模寬度決定三行版面。超寬／缺字直接拒絕，不靠裁切藏字。
func dialogueLayout(font bitmap.Font, text []rune) ([]image.Point, error) {
	x, y := 6, 4
	points := make([]image.Point, 0, len(text))
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
			for j := i; j < len(text) && latin(text[j]); j++ {
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
		if x+w > dialogueBox.Dx()*3-6 {
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

func (d *Dialogue) Active() bool { return d.active }

func (d *Dialogue) Draw(dst *image.RGBA, indexed []byte, palette [256][3]uint8, origin image.Point) bool {
	if !d.active || len(d.completed) != dialogueBox.Dx()*dialogueBox.Dy() {
		return false
	}
	if !d.pending && !bytes.Equal(d.completed, dialogueCrop(indexed)) {
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
		d.reset()
		return false
	}
	box := image.Rect(origin.X+dialogueBox.Min.X*3, origin.Y+dialogueBox.Min.Y*3, origin.X+dialogueBox.Max.X*3, origin.Y+dialogueBox.Max.Y*3)
	prompt := image.Rect(origin.X+dialoguePrompt.Min.X*3, origin.Y+dialoguePrompt.Min.Y*3, origin.X+dialoguePrompt.Max.X*3, origin.Y+dialoguePrompt.Max.Y*3)
	bg, fg := palette[0], palette[9]
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			p := image.Pt(x, y)
			if p.In(dst.Bounds()) && (!d.prompt || !p.In(prompt)) {
				dst.SetRGBA(x, y, color.RGBA{bg[0], bg[1], bg[2], 255})
			}
		}
	}
	p := d.paragraphs[d.current]
	positions, _ := dialogueLayout(d.font, p.text)
	for i, ch := range p.text {
		if i >= d.visible {
			break
		}
		g := d.font[ch]
		w := len(g) / 16 * 8
		at := positions[i].Add(box.Min)
		for y := 0; y < 16; y++ {
			for x := 0; x < w; x++ {
				if g[y*(w/8)+x/8]&(0x80>>uint(x%8)) != 0 {
					for yy := 0; yy < 2; yy++ {
						for xx := 0; xx < 2; xx++ {
							point := image.Pt(at.X+x*2+xx, at.Y+y*2+yy)
							if point.In(dst.Bounds()) {
								dst.SetRGBA(point.X, point.Y, color.RGBA{fg[0], fg[1], fg[2], 255})
							}
						}
					}
				}
			}
		}
	}
	return true
}

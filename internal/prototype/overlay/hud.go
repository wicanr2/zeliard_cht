package overlay

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

type hudDefinition struct {
	key, file, kind, pixels string
	module, body            int
	base, entry             uint16
}

// 位址、完整字形與模組版本只限hud-source-index已證實的九個來源。
var hudDefinitions = []hudDefinition{
	{"zelres1.file.00db4c", "ZELRES1.SAR", "life", "1fdff0f8705807c751ef48c519413afed06ce2e38973f117b79177b07c9eb50a", 0xceb5, 0xdb4c, 0x6000, 0x22bf},
	{"zelres1.file.00db54", "ZELRES1.SAR", "almas", "0e4ce579ca3cd17fb3e0a5321ccf347fb592eba3c16ed630ec2f19450541f5f3", 0xceb5, 0xdb54, 0x6000, 0x22bf},
	{"zelres1.file.00db5d", "ZELRES1.SAR", "gold", "a0c227264c80c53a38e82bf80e4c19d77fde57a8fe22d457f75142f7edcf2fad", 0xceb5, 0xdb5d, 0x6000, 0x22bf},
	{"zelres1.file.00db65", "ZELRES1.SAR", "place", "ee64bb99d91a5d70f62559a5f18b725af14c4f826fe6243e55dba3ebf0ca339d", 0xceb5, 0xdb65, 0x6000, 0x22bf},
	{"zelres2.file.000d34", "ZELRES2.SAR", "gold", "a0c227264c80c53a38e82bf80e4c19d77fde57a8fe22d457f75142f7edcf2fad", 0xec, 0xd34, 0x6000, 0x22bf},
	{"zelres2.file.000d3c", "ZELRES2.SAR", "place", "ee64bb99d91a5d70f62559a5f18b725af14c4f826fe6243e55dba3ebf0ca339d", 0xec, 0xd3c, 0x6000, 0x22bf},
	{"zelres2.file.038dce", "ZELRES2.SAR", "scene", "d3b08035b85e002fb6cab531f6bd0291aef9cff6f7ec6aaaadf41fcecdd28509", 0x38a1a, 0x38dce, 0xc000, 0x22cd},
	{"zelres2.file.0399b8", "ZELRES2.SAR", "scene", "f0a7fb5da7dec2411619381a07e88e33fb37052448b4b32bd9a2bb84e4d83075", 0x392dc, 0x399b8, 0xc000, 0x22cd},
	{"zelres3.file.00c31a", "ZELRES3.SAR", "scene", "2c3ad1e64e2227cb08a8a95cd7a17a9d67d7edc3ec1c58828769288074d86d8c", 0xad02, 0xc31a, 0xc000, 0x22cd},
}

type hudSource struct {
	hudDefinition
	raw, prefix         []byte
	rawHash, prefixHash string
	text                []rune
	box                 image.Rectangle
	header              uint16
	foreground, shadow  uint8
}

// HUD只觀察原版完成的欄位。原版數值、條件、存檔與輸入不依賴譯文。
type HUD struct {
	original  *oracle.Oracle
	font      bitmap.Font
	sources   map[string]*hudSource
	fields    map[string]*hudSource
	pending   *hudSource
	visible   []string
	Completed int
}

func hudHash(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

func NewHUD(o *oracle.Oracle, root, catalog string, font bitmap.Font) (*HUD, error) {
	versions := map[string]string{
		"ZELRES1.SAR": "56a73b59e952dab47fb0171a3fb1cdb7d641baceafa97e2826057f877a590eb0",
		"ZELRES2.SAR": "bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1",
		"ZELRES3.SAR": "9c19b9c317d3006e11f8a5cad8ea1f226f8043e11c6467fa27f10eef33175ef0",
	}
	files := map[string][]byte{}
	for name, digest := range versions {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return nil, err
		}
		if hudHash(data) != digest {
			return nil, fmt.Errorf("狀態列來源版本不符：%s", name)
		}
		files[name] = data
	}
	data, err := os.ReadFile(catalog)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("狀態列TSV須為UTF-8")
	}
	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = '\t'
	r.FieldsPerRecord = 2
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 || rows[0][0] != "key" || rows[0][1] != "zh-TW" {
		return nil, fmt.Errorf("狀態列TSV欄位不符")
	}
	h := &HUD{original: o, font: font, sources: map[string]*hudSource{}, fields: map[string]*hudSource{}}
	for _, row := range rows[1:] {
		var definition *hudDefinition
		for i := range hudDefinitions {
			if hudDefinitions[i].key == row[0] {
				definition = &hudDefinitions[i]
				break
			}
		}
		if definition == nil || h.sources[row[0]] != nil || row[1] == "" || strings.ContainsAny(row[1], "\\\r\n\t") {
			return nil, fmt.Errorf("狀態列鍵或單行譯文不符：%s", row[0])
		}
		d := *definition
		file := files[d.file]
		header := d.body - 4
		if header < d.module || d.module+64 > len(file) || d.body > len(file) {
			return nil, fmt.Errorf("狀態列來源越界")
		}
		n := int(file[header+3])
		if n == 0 || d.body+n > len(file) {
			return nil, fmt.Errorf("狀態列字數越界")
		}
		raw := append([]byte(nil), file[header:d.body+n]...)
		x, y := int(raw[0])*4+int(raw[2]), int(raw[1])
		box := image.Rect(x, y, x+n*5+1, y+8)
		if !box.In(image.Rect(0, 0, 320, 200)) {
			return nil, fmt.Errorf("狀態列安全區越界")
		}
		text := []rune(row[1])
		width := 3
		for _, ch := range text {
			g := font[ch]
			if unicode.IsControl(ch) || len(g) != 16 && len(g) != 32 {
				return nil, fmt.Errorf("狀態列控制碼或字型缺字：%U", ch)
			}
			width += len(g) / 16 * 12
		}
		if width > box.Dx()*3 || box.Dy()*3 < 24 {
			return nil, fmt.Errorf("狀態列譯文超出安全區：%s", row[0])
		}
		prefix := append([]byte(nil), file[d.module:d.module+64]...)
		fg, shadow := uint8(27), uint8(18)
		if d.entry == 0x22cd {
			fg, shadow = 9, 45
		}
		h.sources[d.key] = &hudSource{hudDefinition: d, raw: raw, prefix: prefix, rawHash: hudHash(raw), prefixHash: hudHash(prefix), text: text, box: box, header: uint16(int(d.base) + header - d.module), foreground: fg, shadow: shadow}
	}
	if len(h.sources) != len(hudDefinitions) {
		return nil, fmt.Errorf("狀態列catalog來源缺漏：%d／%d", len(h.sources), len(hudDefinitions))
	}
	if o != nil {
		o.OnCall(oracle.Far(0x01b1, 0x22bf), func(q *oracle.Oracle) { h.begin(q, 0x22bf) })
		o.OnCall(oracle.Far(0x01b1, 0x22cd), func(q *oracle.Oracle) { h.begin(q, 0x22cd) })
		o.OnCall(oracle.Far(0x01b1, 0x2344), h.finish)
	}
	return h, nil
}

func (h *HUD) begin(o *oracle.Oracle, entry uint16) {
	h.pending = nil
	if o.VideoMode() != 0x13 || o.DSReg() != 0x01b1 || hudHash(o.Bytes(oracle.Far(0x01b1, 0x22bf), 0xc7)) != "a4564867c4b2fa886ea061551023ca7343b7e6b33f2ec9dd392b6aa452627504" {
		return
	}
	for _, s := range h.sources {
		if s.entry == entry && s.header == o.SI() && bytes.Equal(o.Bytes(oracle.Far(0x01b1, s.base), 64), s.prefix) && bytes.Equal(o.Bytes(oracle.Far(0x01b1, s.header), len(s.raw)), s.raw) {
			h.pending = s
			return
		}
	}
}

func hudPalette(pal [256][3]uint8) bool {
	var data []byte
	for _, c := range pal {
		data = append(data, c[:]...)
	}
	return hudHash(data) == "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c"
}

func (h *HUD) finish(o *oracle.Oracle) {
	s := h.pending
	h.pending = nil
	if s == nil || o.DSReg() != 0x01b1 || o.SI() != s.header+uint16(len(s.raw)) || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, s.header), len(s.raw)), s.raw) {
		return
	}
	if !h.valid(s, o.Indexed(), o.Palette()) {
		return
	}
	h.fields[s.kind] = s
	h.Completed++
}

func (h *HUD) valid(s *hudSource, raw []byte, pal [256][3]uint8) bool {
	return len(raw) == 64000 && hudPalette(pal) && hudHash(s.raw) == s.rawHash && hudHash(s.prefix) == s.prefixHash && hudHash(labelCrop(raw, s.box)) == s.pixels
}

// 模組卸載不等於畫面消失。保留最多五欄，只在完整原版字形仍存在時覆繪。
func (h *HUD) Draw(dst *image.RGBA, raw []byte, pal [256][3]uint8, origin image.Point) int {
	h.visible = nil
	if h.original != nil && h.original.VideoMode() != 0x13 {
		return 0
	}
	for _, kind := range []string{"life", "almas", "gold", "place", "scene"} {
		s := h.fields[kind]
		if s == nil || !h.valid(s, raw, pal) {
			continue
		}
		box := image.Rect(origin.X+s.box.Min.X*3, origin.Y+s.box.Min.Y*3, origin.X+s.box.Max.X*3, origin.Y+s.box.Max.Y*3)
		if !box.In(dst.Bounds()) {
			continue
		}
		bg, fg, shadow := pal[5], pal[s.foreground], pal[s.shadow]
		for y := box.Min.Y; y < box.Max.Y; y++ {
			for x := box.Min.X; x < box.Max.X; x++ {
				dst.SetRGBA(x, y, color.RGBA{bg[0], bg[1], bg[2], 255})
			}
		}
		x := box.Min.X
		for _, ch := range s.text {
			g := h.font[ch]
			w := len(g) / 16 * 8
			dw := w * 3 / 2
			for yy := 0; yy < 24; yy++ {
				for xx := 0; xx < dw; xx++ {
					sx, sy := xx*2/3, yy*2/3
					if g[sy*(w/8)+sx/8]&(0x80>>uint(sx%8)) == 0 {
						continue
					}
					dst.SetRGBA(x+xx+3, box.Min.Y+yy, color.RGBA{shadow[0], shadow[1], shadow[2], 255})
					dst.SetRGBA(x+xx, box.Min.Y+yy, color.RGBA{fg[0], fg[1], fg[2], 255})
				}
			}
			x += dw
		}
		h.visible = append(h.visible, s.key)
	}
	return len(h.visible)
}

func (h *HUD) VisibleKeys() []string {
	keys := append([]string(nil), h.visible...)
	sort.Strings(keys)
	return keys
}

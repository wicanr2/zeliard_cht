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
	"github.com/wicanr2/zeliard_cht/internal/prototype/glyphpatch"
)

type hudDefinition struct {
	key, file, kind, pixels string
	module, body            int
	base, entry             uint16
}

// 位址與字形依hud-source-index。base為0的候選須由實際入口唯一驗證載入基址。
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
	{"zelres2.file.017238", "ZELRES2.SAR", "scene", "f1485c6e92876f81e2681bdbd1bb01a20691b56103611c778492003075bd82e5", 0x16e1a, 0x17238, 0, 0x22cd},
	{"zelres2.file.017809", "ZELRES2.SAR", "scene", "cd68aab4314b80bbe89a8796f0c0fc00bfe2a018aaf06b90dfe0c96224e54f26", 0x175c0, 0x17809, 0, 0x22cd},
	{"zelres2.file.0184c9", "ZELRES2.SAR", "scene", "9f1b0ccfee9e8ff11343ce98c30347a88edb5a1e04f4686222e99a2901b2d8f0", 0x17817, 0x184c9, 0, 0x22cd},
	{"zelres2.file.01a448", "ZELRES2.SAR", "scene", "b4d572d4fecabce8ae5f8128c6c8731c49119f5983e49c1f0af105e6a112ba0e", 0x1a19e, 0x1a448, 0, 0x22cd},
	{"zelres2.file.01baa1", "ZELRES2.SAR", "scene", "62773fb3f5eabd9e9a5ad7b8aceecea5fe12cd1663ff1c8935adac61a7b563c5", 0x1b7b2, 0x1baa1, 0, 0x22cd},
	{"zelres2.file.01c997", "ZELRES2.SAR", "scene", "e1d3bb2e95c89c78c09c33cd02b0756a84d53f9e3979f705c8a628ec06c71af7", 0x1bcc6, 0x1c997, 0, 0x22cd},
	{"zelres2.file.01c9a9", "ZELRES2.SAR", "scene", "0ec5272dd2c2dd92ab9d50a9f0a61a097b86693b55ef43ae1aff5fc4dd96fc7a", 0x1bcc6, 0x1c9a9, 0, 0x22cd},
	{"zelres2.file.01c9bc", "ZELRES2.SAR", "scene", "4e833461eaf256657c8f271c063b46191017c2fd5e57fa4d02ecde1094a25e05", 0x1bcc6, 0x1c9bc, 0, 0x22cd},
	{"zelres2.file.01c9cf", "ZELRES2.SAR", "scene", "95ff5582ae79d3eea43ec58f9390fa456bdf51743ce3a2044704a467a52306bd", 0x1bcc6, 0x1c9cf, 0, 0x22cd},
	{"zelres2.file.01c9e3", "ZELRES2.SAR", "scene", "f8a499ccd7f111f0c6f78479a731fe23723aa13dd6d1225f8d7c059e76e41937", 0x1bcc6, 0x1c9e3, 0, 0x22cd},
	{"zelres2.file.01c9f6", "ZELRES2.SAR", "scene", "2752f62ebb68f0f31d5601713cbd2586084f6328391d48660e9d0728b014a1b0", 0x1bcc6, 0x1c9f6, 0, 0x22cd},
	{"zelres2.file.01ca09", "ZELRES2.SAR", "scene", "342257ef2252bf9e6a998d7d4b2e146ce62b6cb6f2d841a2af0e0a6b85dde71f", 0x1bcc6, 0x1ca09, 0, 0x22cd},
	{"zelres2.file.01ca1b", "ZELRES2.SAR", "scene", "4e9f51c6ba2e111be148aa704a28d09049cc53d89f60b3d4b5776282ac50245c", 0x1bcc6, 0x1ca1b, 0, 0x22cd},
	{"zelres2.file.03a8b5", "ZELRES2.SAR", "scene", "05492ccb82d35e6649ef237354fae16c3c2a47ae12ea8781acf9c5708e4e6ed0", 0x3a1d9, 0x3a8b5, 0, 0x22cd},
	{"zelres2.file.03b36d", "ZELRES2.SAR", "scene", "30646f231d96352832d0e3e6083edc0b5881654b3b6ace8b680de2d6740a1b22", 0x3ae89, 0x3b36d, 0, 0x22cd},
	{"zelres2.file.03c31f", "ZELRES2.SAR", "scene", "1eac1b7226a02f6da5b2d018f9ed25a7cb4a756fe96eb88edcee06c9df76cc6d", 0x3bbe3, 0x3c31f, 0, 0x22cd},
	{"zelres2.file.03d27b", "ZELRES2.SAR", "scene", "2e4d35d760710448205195f0d5511398b25a937970b8e33c1084540ffff3e3dc", 0x3c9e7, 0x3d27b, 0, 0x22cd},
	{"zelres2.file.03e0b5", "ZELRES2.SAR", "scene", "004cd493b8d21387f7ea44f1a3a838cbf04c9a07a1a83e7808d25253b751e890", 0x3d9d9, 0x3e0b5, 0, 0x22cd},
	{"zelres2.file.03f0f7", "ZELRES2.SAR", "scene", "10ab9005b633b62aac0329f47b8369ba12bfb3625069dfd6e5ebe6da7a8a9f02", 0x3e813, 0x3f0f7, 0, 0x22cd},
	{"zelres2.file.040350", "ZELRES2.SAR", "scene", "0bb7805470db13d9ce3965372e3fdca4813392bc519f0b0fb33de29512bf0b44", 0x3f92c, 0x40350, 0, 0x22cd},
	{"zelres2.file.0410c7", "ZELRES2.SAR", "scene", "edcb1251e3aa54967795e0cd8c394ff5278560cbcef267a3201835477977a80d", 0x409eb, 0x410c7, 0, 0x22cd},
	{"zelres3.file.00c88f", "ZELRES3.SAR", "scene", "2c3ad1e64e2227cb08a8a95cd7a17a9d67d7edc3ec1c58828769288074d86d8c", 0xc696, 0xc88f, 0, 0x22cd},
	{"zelres3.file.00e0f0", "ZELRES3.SAR", "scene", "a92f0d27c0b1d7ed7a13ef05cd9248d87315b55d37dee3c13d3e4310ac5ff502", 0xc8d2, 0xe0f0, 0, 0x22cd},
	{"zelres3.file.00efb8", "ZELRES3.SAR", "scene", "a92f0d27c0b1d7ed7a13ef05cd9248d87315b55d37dee3c13d3e4310ac5ff502", 0xe48c, 0xefb8, 0, 0x22cd},
	{"zelres3.file.00f37f", "ZELRES3.SAR", "scene", "a92f0d27c0b1d7ed7a13ef05cd9248d87315b55d37dee3c13d3e4310ac5ff502", 0xf064, 0xf37f, 0, 0x22cd},
	{"zelres3.file.01000c", "ZELRES3.SAR", "scene", "9dcebbb47963a53de340059c0ec309696d8d42f30541d25999d44b34556b8dbf", 0xf3c2, 0x1000c, 0, 0x22cd},
	{"zelres3.file.01118c", "ZELRES3.SAR", "scene", "bd89d4a14ae9a6445139cf1db67222f60ecd54c59083fcb07d0ba35e71cd8ec8", 0x102f7, 0x1118c, 0, 0x22cd},
	{"zelres3.file.01164f", "ZELRES3.SAR", "scene", "bd89d4a14ae9a6445139cf1db67222f60ecd54c59083fcb07d0ba35e71cd8ec8", 0x114a5, 0x1164f, 0, 0x22cd},
	{"zelres3.file.012f9b", "ZELRES3.SAR", "scene", "75283194597e01860530cd8a8a7694b14a13493c4aa9f7e50d4845720a7fccb6", 0x1168f, 0x12f9b, 0, 0x22cd},
	{"zelres3.file.0145ac", "ZELRES3.SAR", "scene", "58dd11c7b24540817bc1bc53345f2cf98de2302753629b4471ab4d1a8cdf7364", 0x13407, 0x145ac, 0, 0x22cd},
	{"zelres3.file.014a2b", "ZELRES3.SAR", "scene", "75283194597e01860530cd8a8a7694b14a13493c4aa9f7e50d4845720a7fccb6", 0x14809, 0x14a2b, 0, 0x22cd},
	{"zelres3.file.016346", "ZELRES3.SAR", "scene", "95d75333cdaab451ab01951c8ca2ff04a4b48f8d1e6ab6023b1c9ea8b3fef33d", 0x14a5a, 0x16346, 0, 0x22cd},
	{"zelres3.file.0182ea", "ZELRES3.SAR", "scene", "fa75876b1769a0f5a5bae069856aae12ca297370081d6ae15b98c8a0e20d8bb4", 0x167f2, 0x182ea, 0, 0x22cd},
	{"zelres3.file.018992", "ZELRES3.SAR", "scene", "fa75876b1769a0f5a5bae069856aae12ca297370081d6ae15b98c8a0e20d8bb4", 0x18767, 0x18992, 0, 0x22cd},
	{"zelres3.file.01a2f0", "ZELRES3.SAR", "scene", "f82a0f9dacf7a0c042493fadfd1a87ade8e4f8c975bc3fe086fc77d5cb07f432", 0x189d6, 0x1a2f0, 0, 0x22cd},
	{"zelres3.file.01bbf1", "ZELRES3.SAR", "scene", "a5214eeafa087fde13281b2b67c41ff2efd2a8b89fb1ceca02605ce5c4d46d0d", 0x1a6db, 0x1bbf1, 0, 0x22cd},
	{"zelres3.file.01c4c9", "ZELRES3.SAR", "scene", "30600d020725370b4a81a0d53ae4ffb44d00ca88e7bf26838529b413402daff1", 0x1bfcb, 0x1c4c9, 0, 0x22cd},
	{"zelres3.file.01c965", "ZELRES3.SAR", "scene", "f82a0f9dacf7a0c042493fadfd1a87ade8e4f8c975bc3fe086fc77d5cb07f432", 0x1c6a5, 0x1c965, 0, 0x22cd},
	{"zelres3.file.01e6f8", "ZELRES3.SAR", "scene", "8a0490546492f70b02f7ae258241bad4278f335e59cb605de7c95a84aa23205a", 0x1c9a7, 0x1e6f8, 0, 0x22cd},
	{"zelres3.file.02009b", "ZELRES3.SAR", "scene", "9d5ad7136d195fc08dc2865cd22352c5da2c36a182fe07dfe8e44f96bed18704", 0x1e8e5, 0x2009b, 0, 0x22cd},
	{"zelres3.file.02146d", "ZELRES3.SAR", "scene", "4dff5b8580d4aea9df8da5af5e165585d99228414ab47a16354902e6fbaa84c4", 0x202e8, 0x2146d, 0, 0x22cd},
	{"zelres3.file.02171f", "ZELRES3.SAR", "scene", "cd68aab4314b80bbe89a8796f0c0fc00bfe2a018aaf06b90dfe0c96224e54f26", 0x21598, 0x2171f, 0, 0x22cd},
	{"zelres3.file.0219eb", "ZELRES3.SAR", "scene", "8a0490546492f70b02f7ae258241bad4278f335e59cb605de7c95a84aa23205a", 0x21747, 0x219eb, 0, 0x22cd},
	{"zelres3.file.0275b1", "ZELRES3.SAR", "scene", "d4ccaf5f769814038ac2c5e260bdd7d62bdb42917ea16c5d02702e07e2a73a5f", 0x26989, 0x275b1, 0, 0x22cd},
}

type hudSource struct {
	hudDefinition
	raw, prefix         []byte
	rawHash, prefixHash string
	text                []rune
	box                 image.Rectangle
	header              uint16
	foreground, shadow  uint8
	width               int
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
	// Undrawn 是上次 Draw 原版欄位字形完整、卻放不下而沒畫中文的欄數（探針用）。規格067。
	Undrawn int
}

func hudHash(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

// Translation只回傳已核對原件來源的譯文，不參與DOS條件判定。
func (h *HUD) Translation(key string) (string, bool) {
	if h == nil {
		return "", false
	}
	s := h.sources[key]
	if s == nil {
		return "", false
	}
	return string(s.text), true
}

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
		// 場景名寫成「中文（English）」可比原文長，執行時向兩側的空白欄延伸，上限是地點欄寬。規格067。
		if width > box.Dx()*3 && (d.kind != "scene" || width > hudSceneMaxWidth) || box.Dy()*3 < 24 {
			return nil, fmt.Errorf("狀態列譯文超出安全區：%s", row[0])
		}
		prefix := append([]byte(nil), file[d.module:d.module+64]...)
		fg, shadow := uint8(27), uint8(18)
		if d.entry == 0x22cd {
			fg, shadow = 9, 45
		}
		h.sources[d.key] = &hudSource{hudDefinition: d, raw: raw, prefix: prefix, rawHash: hudHash(raw), prefixHash: hudHash(prefix), text: text, box: box, header: uint16(int(d.base) + header - d.module), foreground: fg, shadow: shadow, width: width}
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

// 026：只接受同一名稱來源已觀察的完整前綴，不遮罩遊戲狀態欄位。
func hudPrefixMatches(s *hudSource, prefix []byte) bool {
	if len(prefix) != 64 || len(s.prefix) != 64 {
		return false
	}
	if bytes.Equal(prefix, s.prefix) {
		return true
	}
	return s.key == "zelres3.file.00c31a" && hudHash(prefix) == "6c853b5dab1da5682a959dd804b0f23e5acdc3aa67099596335e1293694f03b2"
}

func (h *HUD) begin(o *oracle.Oracle, entry uint16) {
	h.pending = nil
	if o.VideoMode() != 0x13 || o.DSReg() != 0x01b1 || hudHash(glyphpatch.CodeView(o, oracle.Far(0x01b1, 0x22bf), 0xc7)) != "a4564867c4b2fa886ea061551023ca7343b7e6b33f2ec9dd392b6aa452627504" {
		return
	}
	if o.Word(oracle.Far(0x01b1, 0xf504)) != 0xf856 || hudHash(o.Bytes(oracle.Far(0x01b1, 0xf856), 768)) != "f16d564fee5c4472914b4dc64cf467bcd05da89125288a83b0b065d164707d07" {
		return
	}
	for _, s := range h.sources {
		if s.entry != entry {
			continue
		}
		candidate := *s
		if s.base == 0 {
			base := int(o.SI()) - (s.body - 4 - s.module)
			if base < 0 || base+64 > 65536 || int(o.SI())+len(s.raw) > 65536 {
				continue
			}
			candidate.base, candidate.header = uint16(base), o.SI()
		}
		if candidate.header == o.SI() && hudPrefixMatches(s, o.Bytes(oracle.Far(0x01b1, candidate.base), 64)) && bytes.Equal(o.Bytes(oracle.Far(0x01b1, candidate.header), len(s.raw)), s.raw) {
			if h.pending != nil {
				h.pending = nil
				return // 相同入口有多個來源解釋時不猜選其中一個。
			}
			h.pending = &candidate
		}
	}
	if h.pending != nil {
		// 規格066：唯一來源的名稱整串由中文取代。
		glyphpatch.ClaimText(o)
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
	// 規格066：已認領的名稱若最後沒有接受，照原版寫回，不留空白。
	stored := false
	defer func() {
		if s != nil && !stored {
			glyphpatch.Reveal(o, s.box)
		}
	}()
	if s == nil || o.DSReg() != 0x01b1 || o.SI() != s.header+uint16(len(s.raw)) || !bytes.Equal(o.Bytes(oracle.Far(0x01b1, s.header), len(s.raw)), s.raw) {
		return
	}
	if !h.valid(s, glyphpatch.Indexed(o), o.Palette()) {
		return
	}
	stored = true
	h.fields[s.kind] = s
	h.Completed++
}

func (h *HUD) valid(s *hudSource, raw []byte, pal [256][3]uint8) bool {
	return len(raw) == 64000 && hudPalette(pal) && hudHash(s.raw) == s.rawHash && hudHash(s.prefix) == s.prefixHash && hudHash(labelCrop(raw, s.box)) == s.pixels
}

// 模組卸載不等於畫面消失。保留最多五欄，只在完整原版字形仍存在時覆繪。
func (h *HUD) Draw(dst *image.RGBA, raw []byte, pal [256][3]uint8, origin image.Point) int {
	h.visible = nil
	h.Undrawn = 0
	if h.original != nil && h.original.VideoMode() != 0x13 {
		return 0
	}
	var drawn []image.Rectangle
	for _, kind := range []string{"life", "almas", "gold", "place", "scene"} {
		s := h.fields[kind]
		if s == nil || !h.valid(s, raw, pal) {
			continue
		}
		native, size, width := s.box, 24, s.width
		if width > native.Dx()*3 {
			var fits bool
			if native, fits = hudExtend(raw, s.box, width, drawn); !fits {
				// 延伸後仍放不下時改用16像素字，不退回原版英文。
				size, width = 16, (width-3)*2/3+3
				if width > native.Dx()*3 {
					h.Undrawn++
					continue
				}
			}
		}
		box := image.Rect(origin.X+native.Min.X*3, origin.Y+native.Min.Y*3, origin.X+native.Max.X*3, origin.Y+native.Max.Y*3)
		if !box.In(dst.Bounds()) {
			h.Undrawn++
			continue
		}
		drawn = append(drawn, native)
		bg, fg, shadow := pal[5], pal[s.foreground], pal[s.shadow]
		for y := box.Min.Y; y < box.Max.Y; y++ {
			for x := box.Min.X; x < box.Max.X; x++ {
				dst.SetRGBA(x, y, color.RGBA{bg[0], bg[1], bg[2], 255})
			}
		}
		x, y := box.Min.X, box.Min.Y+(24-size)/2
		if native != s.box || size != 24 {
			x += (box.Dx() - width) / 2
		}
		for _, ch := range s.text {
			dw := len(h.font[ch]) / 16 * 8 * size / 16
			h.font.DrawGlyph(dst, image.Pt(x+size/8, y), ch, size, color.RGBA{shadow[0], shadow[1], shadow[2], 255}, nil)
			h.font.DrawGlyph(dst, image.Pt(x, y), ch, size, color.RGBA{fg[0], fg[1], fg[2], 255}, nil)
			x += dw
		}
		h.visible = append(h.visible, s.key)
	}
	return len(h.visible)
}

// hudSceneMaxWidth 是地點欄去掉「地點」標籤後的寬度（原版x=80..187，乘3）。
const hudSceneMaxWidth = (187 - 80) * 3

// hudExtend 把場景名欄向左右延伸到原版畫面上整欄都是欄底色（色號5）的範圍，兩端各留一欄；
// 不跨進已畫的其他欄位。第二個回傳值表示24像素字是否放得下。規格067。
func hudExtend(raw []byte, box image.Rectangle, width int, taken []image.Rectangle) (image.Rectangle, bool) {
	blank := func(x int) bool {
		if x < 0 || x >= 320 {
			return false
		}
		for _, r := range taken {
			if x >= r.Min.X && x < r.Max.X {
				return false
			}
		}
		for y := box.Min.Y; y < box.Max.Y; y++ {
			if raw[y*320+x] != 5 {
				return false
			}
		}
		return true
	}
	lo, hi := box.Min.X, box.Max.X
	for blank(lo - 1) {
		lo--
	}
	for blank(hi) {
		hi++
	}
	if lo < box.Min.X {
		lo++
	}
	if hi > box.Max.X {
		hi--
	}
	out := image.Rect(lo, box.Min.Y, hi, box.Max.Y)
	return out, width <= out.Dx()*3
}

func (h *HUD) VisibleKeys() []string {
	keys := append([]string(nil), h.visible...)
	sort.Strings(keys)
	return keys
}

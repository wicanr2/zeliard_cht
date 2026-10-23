package overlay

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

func TestQuotationCatalogAndNumberBitmapFallback(t *testing.T) {
	root, fontPath := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT")
	if root == "" || fontPath == "" {
		t.Skip("須掛載原版及字型")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "ZELRES2.SAR"))
	if err != nil || fmtHash(data) != "bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1" {
		t.Fatal("來源版本不符", err)
	}
	o, err := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
	if err != nil {
		t.Fatal(err)
	}
	defer o.Close()
	keys := []string{"zelres2.file.0188f4", "zelres2.file.01891d", "zelres2.file.01885d", "zelres2.file.018904", "zelres2.file.0188b8"}
	for _, bad := range []string{"missing", "duplicate", "unknown", "incomplete", "font"} {
		t.Run(bad, func(t *testing.T) {
			texts := map[string]string{keys[0]: "購買{item}？", keys[1]: "售價{price}金幣。", keys[2]: "舊武器折抵{trade}金幣。", keys[3]: "確定交易嗎？", keys[4]: "舊盾牌折抵{trade}金幣。"}
			switch bad {
			case "missing":
				texts[keys[1]] = "售價金幣。"
			case "duplicate":
				texts[keys[1]] = "{price}{price}"
			case "unknown":
				texts[keys[1]] = "{price}{other}"
			case "incomplete":
				delete(texts, keys[2])
			case "font":
				texts[keys[3]] = string(rune(0x10ffff))
			}
			if _, err := newQuotation(o, data, texts, font, &paintHistory{}); err == nil {
				t.Fatal("未拒絕不相容的報價catalog")
			}
		})
	}
	if q, err := newQuotation(o, data, map[string]string{}, font, &paintHistory{}); err != nil || q != nil {
		t.Fatal("舊catalog無報價時應保留原文")
	}
	// 合成數字圖形測試只證明顯示契約，不當作原版987／65的行為證據。
	raw := make([]byte, 64000)
	q := &Quotation{font: font, paint: &paintHistory{}, active: true, ready: 10, completed: dialogueCrop(raw), item: &labelSource{text: []rune("測試")}, texts: [5]string{"購買{item}？", "售價{price}金幣。", "舊武器折抵{trade}金幣。", "確定交易嗎？", "舊盾牌折抵{trade}金幣。"}}
	q.price = quoteNumber{raw: []byte("987\xff"), box: image.Rect(0, 0, 21, 8), pixels: make([]byte, 21*8)}
	q.trade = quoteNumber{raw: []byte("65\xff"), box: image.Rect(0, 0, 14, 8), pixels: make([]byte, 14*8)}
	for i := range q.price.pixels {
		if i%5 == 0 {
			q.price.pixels[i] = 9
		}
	}
	for i := range q.trade.pixels {
		if i%3 == 0 {
			q.trade.pixels[i] = 9
		}
	}
	var pal [256][3]uint8
	pal[9] = [3]uint8{220, 230, 240}
	dst := image.NewRGBA(image.Rect(0, 0, 960, 600))
	if !q.Draw(dst, raw, pal, image.Point{}) {
		t.Fatal("已知原版數字圖形未合成")
	}
	tokens := q.tokens()
	positions, ok := quoteLayout(tokens)
	if !ok {
		t.Fatal("安全報價版面失效")
	}
	for i, token := range tokens {
		if n := token.number; n != nil {
			at := positions[i].Add(image.Pt(dialogueBox.Min.X*3, dialogueBox.Min.Y*3+4))
			for y := 0; y < n.box.Dy()*3; y++ {
				for x := 0; x < n.box.Dx()*3; x++ {
					v := pal[n.pixels[y/3*n.box.Dx()+x/3]]
					if dst.RGBAAt(at.X+x, at.Y+y) != (color.RGBA{v[0], v[1], v[2], 255}) {
						t.Fatal("合成數字像素被改寫")
					}
				}
			}
		}
	}
	bad := append([]byte(nil), raw...)
	bad[100*320+55] = 9
	before := append([]byte(nil), dst.Pix...)
	if q.Draw(dst, bad, pal, image.Point{}) || q.active || !bytes.Equal(before, dst.Pix) {
		t.Fatal("未知畫面仍顯示舊報價")
	}
	if _, ok := quoteLayout([]quoteToken{{number: &q.price, width: 650}}); ok {
		t.Fatal("超寬數字未安全退回")
	}
}

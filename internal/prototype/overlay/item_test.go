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

func TestItemShopPurchaseNormalPathReadOnly(t *testing.T) {
	if os.Getenv("ZELIARD_ITEM_SAVE") == "" {
		t.Skip("須掛載正常建立的起始存檔")
	}
	if out := os.Getenv("ZELIARD_SHOP_OUTPUT"); out != "" {
		t.Setenv("ZELIARD_SHOP_OUTPUT", filepath.Join(out, "item"))
	}
	testWeaponShopNormalPath(t, "../../../tools/item-shop.commands", 32)
}

func assertItemCheckpoint(t *testing.T, name string, d *Dialogue, l *Labels, o *oracle.Oracle, dst *image.RGBA, drawn bool) {
	t.Helper()
	wants := map[string]string{"item-greeting": "zelres2.file.01adf3", "item-list": "zelres2.file.01ae15", "after-buy-no": "zelres2.file.01aeee", "item-farewell": "zelres2.file.01b099", "item-farewell-ready": "zelres2.file.01b099"}
	if key, ok := wants[name]; ok {
		if !drawn || d.current < 0 || d.paragraphs[d.current].key != key || d.visible != len(d.paragraphs[d.current].text) {
			t.Fatalf("%s道具店完整中文未顯示：%v current=%d", name, drawn, d.current)
		}
	}
	keys := []string{}
	if name == "item-greeting" {
		keys = []string{"zelres2.file.01ada8", "zelres2.file.01adc1", "zelres2.file.01adcc", "zelres2.file.01add5", "zelres2.file.01addf"}
	}
	if name == "item-list" {
		keys = []string{"zelres2.file.01b622", "zelres2.file.01b65c", "zelres2.file.01b67c"}
	}
	if name == "first-item-quote" {
		keys = []string{"zelres1.file.00e3c8", "zelres1.file.00e3cc"}
	}
	if name == "item-inventory" {
		keys = []string{"zelres2.file.004a1d", "zelres2.file.004a2e", "zelres2.file.004a37", "zelres2.file.004a3f", "zelres2.file.004d03", "zelres2.file.004d0c"}
	}
	for _, key := range keys {
		found := false
		for _, s := range l.stamps {
			found = found || s.source == l.sources[key]
		}
		if !found {
			t.Fatal("道具店中文標籤缺漏", name, key)
		}
	}
	if name == "item-greeting" {
		for _, s := range l.stamps {
			if !s.source.itemTitle {
				continue
			}
			if s.background != 5 || s.foreground != 9 || s.shadow != 45 || s.shadowOffset == nil || *s.shadowOffset != image.Pt(1, 0) || s.box != image.Rect(56, 175, 181, 183) {
				t.Fatal("魔法物品店標題幾何或色彩不符")
			}
			isolated := *l
			isolated.stamps = []labelStamp{s}
			raw := append([]byte(nil), o.Indexed()...)
			raw[175*320+56] ^= 128
			if isolated.Draw(image.NewRGBA(dst.Bounds()), raw, o.Palette(), image.Point{}) != 0 {
				t.Fatal("道具店標題未知像素未退回")
			}
		}
	}
	if name == "first-item-quote" {
		q := d.itemPurchase
		if !drawn || q == nil || !q.active || q.ready != 7 || q.Completed != 1 || string(q.price.raw) != "50\xff" || q.item != l.sources["zelres2.file.01b622"] || o.Word(oracle.Far(0x01b1, 0x86)) != 950 {
			t.Fatalf("購買後商品、50金幣或續購中文不符：%v %v", drawn, d.ItemPurchaseState())
		}
		points, ok := quoteLayout(q.tokens())
		if !ok {
			t.Fatal("道具店購物中文超出安全區")
		}
		copied := 0
		for i, token := range q.tokens() {
			if n := token.number; n != nil {
				copied++
				at := points[i].Add(image.Pt(dialogueBox.Min.X*3, dialogueBox.Min.Y*3+4))
				pal := o.Palette()
				for y := 0; y < n.box.Dy()*3; y++ {
					for x := 0; x < n.box.Dx()*3; x++ {
						v := pal[n.pixels[y/3*n.box.Dx()+x/3]]
						if dst.RGBAAt(at.X+x, at.Y+y) != (color.RGBA{v[0], v[1], v[2], 255}) {
							t.Fatal("道具店50數字圖形被改寫")
						}
					}
				}
			}
		}
		if copied != 1 {
			t.Fatal("道具店購物缺少原版數字")
		}
		for _, bad := range []string{"pixel", "palette", "number"} {
			isolated := *q
			raw := append([]byte(nil), o.Indexed()...)
			pal := o.Palette()
			canvas := image.NewRGBA(dst.Bounds())
			before := append([]byte(nil), canvas.Pix...)
			switch bad {
			case "pixel":
				raw[100*320+55] ^= 128
			case "palette":
				pal[9][0] ^= 1
			case "number":
				isolated.price.pixels = nil
			}
			if isolated.Draw(canvas, raw, pal, image.Point{}) || isolated.active || !bytes.Equal(canvas.Pix, before) {
				t.Fatal("道具店未知畫面未原子退回", bad)
			}
		}
	}
	if name == "item-outside" || name == "item-inventory-closed" {
		if drawn || d.itemPurchase.active || o.Word(oracle.Far(0x01b1, 0x86)) != 950 {
			t.Fatal("離店或關閉物品頁後殘留中文或金額變動", name)
		}
	}
}

func TestItemPurchaseCatalogValidation(t *testing.T) {
	root, fontPath := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT")
	if root == "" || fontPath == "" {
		t.Skip("須掛載原版與字型")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "ZELRES2.SAR"))
	if err != nil {
		t.Fatal(err)
	}
	o, err := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
	if err != nil {
		t.Fatal(err)
	}
	defer o.Close()
	for _, bad := range []string{"missing", "duplicate", "unknown", "incomplete", "font"} {
		t.Run(bad, func(t *testing.T) {
			texts := map[string]string{"zelres2.file.01ae4d": "您購買了{item}。", "zelres2.file.01ae7a": "售價{price}金幣。", "zelres2.file.01ae92": "還需要其他物品嗎？"}
			switch bad {
			case "missing":
				texts["zelres2.file.01ae4d"] = "您購買了。"
			case "duplicate":
				texts["zelres2.file.01ae7a"] = "{price}{price}"
			case "unknown":
				texts["zelres2.file.01ae7a"] = "{price}{other}"
			case "incomplete":
				delete(texts, "zelres2.file.01ae92")
			case "font":
				texts["zelres2.file.01ae92"] = string(rune(0x10ffff))
			}
			if _, err := newItemPurchase(o, data, texts, font, &paintHistory{}); err == nil {
				t.Fatal("道具店不相容catalog未拒絕")
			}
		})
	}
	if q, err := newItemPurchase(o, data, map[string]string{}, font, &paintHistory{}); err != nil || q != nil {
		t.Fatal("舊catalog應保留道具店原文")
	}
}

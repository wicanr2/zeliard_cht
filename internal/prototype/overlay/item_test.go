package overlay

import (
	"bytes"
	"encoding/csv"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
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

func TestItemUseCavernNormalPathReadOnly(t *testing.T) {
	if os.Getenv("ZELIARD_ITEM_SAVE") == "" {
		t.Skip("須掛載正常建立的起始存檔")
	}
	if out := os.Getenv("ZELIARD_SHOP_OUTPUT"); out != "" {
		t.Setenv("ZELIARD_SHOP_OUTPUT", filepath.Join(out, "item-use"))
	}
	testWeaponShopNormalPath(t, "../../../tools/item-shop-use.commands", 48)
}

func assertItemUseCheckpoint(t *testing.T, name string, l *Labels, o *oracle.Oracle, dst *image.RGBA, drawn bool) {
	t.Helper()
	wants := map[string][]string{
		"cave-item-inventory":  {"zelres2.file.004ab8"},
		"cave-potion-selected": {"zelres2.file.004c62", "zelres2.file.004c69"},
		"cave-potion-used":     {"zelres2.file.004aca", "zelres2.file.004b90", "zelres2.file.004ab8"},
		"cave-inventory-empty": {"zelres2.file.004ab0"},
	}
	for _, key := range wants[name] {
		found := false
		for _, stamp := range l.stamps {
			if stamp.noUseIcon || stamp.source != l.sources[key] {
				continue
			}
			isolated := *l
			isolated.stamps = []labelStamp{stamp}
			canvas := inventoryOriginalCanvas(o)
			before := append([]byte(nil), canvas.Pix...)
			isolated.Draw(canvas, o.Indexed(), o.Palette(), image.Point{})
			found = found || !bytes.Equal(before, canvas.Pix)
		}
		if !found {
			t.Fatal("道具使用完整中文缺漏", name, key)
		}
	}
	if name == "item-inventory" || name == "cave-item-inventory" || name == "cave-potion-selected" || name == "cave-potion-used" {
		icons := 0
		for _, stamp := range l.stamps {
			if !stamp.noUseIcon {
				continue
			}
			icons++
			if string(stamp.source.text) != "不用" || stamp.box != image.Rect(59, 132, 73, 146) {
				t.Fatal("NO USE中文來源或安全區不符")
			}
			isolated := *l
			isolated.stamps = []labelStamp{stamp}
			for _, bad := range []string{"tile", "panel", "palette", "source", "text"} {
				copyL := isolated
				copyS := *stamp.source
				copyL.sources = map[string]*labelSource{"zelres2.file.004ab8": &copyS}
				raw, pal := append([]byte(nil), o.Indexed()...), o.Palette()
				switch bad {
				case "tile":
					raw[133*320+61] ^= 128
				case "panel":
					raw[12*320+47] ^= 128
				case "palette":
					pal[9][0] ^= 1
				case "source":
					copyS.prefix = append([]byte(nil), copyS.prefix...)
					copyS.prefix[0] ^= 1
				case "text":
					copyS.text = []rune("不使用")
				}
				canvas := image.NewRGBA(dst.Bounds())
				before := append([]byte(nil), canvas.Pix...)
				if copyL.Draw(canvas, raw, pal, image.Point{}) != 0 || !bytes.Equal(before, canvas.Pix) {
					t.Fatal("NO USE未知來源未原子退回", bad)
				}
			}
			canvas := image.NewRGBA(dst.Bounds())
			before := append([]byte(nil), canvas.Pix...)
			if isolated.Draw(canvas, o.Indexed(), o.Palette(), image.Point{}) != 1 {
				t.Fatal("NO USE獨立繪製失敗")
			}
			changed := 0
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := canvas.PixOffset(x, y)
					if bytes.Equal(before[i:i+4], canvas.Pix[i:i+4]) {
						continue
					}
					changed++
					if !image.Pt(x/3, y/3).In(stamp.box) {
						t.Fatal("NO USE字模或背景越界")
					}
				}
			}
			if changed == 0 {
				masks, known := l.paint.masks(stamp.paintSerial, o.Indexed())
				t.Fatal("NO USE未產生中文字模", name, stamp.paintSerial, l.paint.serial, masks, known)
			}
		}
		if icons != 1 {
			t.Fatal("NO USE圖示缺漏或重複", name, icons)
		}
		completed := l.Completed
		l.ObserveFrame(o.Indexed(), o.Palette())
		if l.Completed != completed {
			t.Fatal("穩定物品圖示重複建立")
		}
	}
	if name == "cave-potion-used" || name == "cave-inventory-empty" {
		if o.Byte(oracle.Far(0x01b1, 0xa6)) != 0 || o.Word(oracle.Far(0x01b1, 0x90)) != 80 {
			t.Fatal("滿生命的藥水消耗與原版證據不符")
		}
	}
	if name == "cave-inventory-empty" {
		for _, stamp := range l.stamps {
			if stamp.noUseIcon {
				t.Fatal("藥水移除後殘留NO USE圖示")
			}
		}
	}
	if name == "cave-item-inventory" {
		if o.Byte(oracle.Far(0x01b1, 0xadf8)) != 0 {
			t.Fatal("洞窟裝備頁未走可使用分支")
		}
		for _, stamp := range l.stamps {
			if stamp.source == l.sources["zelres2.file.004a37"] && (stamp.foreground != 18 || stamp.shadow != 45) {
				t.Fatal("使用區原版紅字或陰影遺失")
			}
		}
	}
	if name == "cave-use-closed" || name == "cave-inventory-closed" || name == "use-return-town" {
		if !bytes.Equal(dst.Pix, inventoryOriginalCanvas(o).Pix) || o.Word(oracle.Far(0x01b1, 0x86)) != 950 {
			t.Fatal("使用後關閉或返回城鎮有殘留或改變金幣", name)
		}
	}
}

func inventoryOriginalCanvas(o *oracle.Oracle) *image.RGBA {
	_, _, rgb := o.ScreenRGB()
	canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
	for y := 0; y < 600; y++ {
		for x := 0; x < 960; x++ {
			i := (y/3*320 + x/3) * 3
			canvas.SetRGBA(x, y, color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255})
		}
	}
	return canvas
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

func TestItemShopDescriptionsAndSaleNormalPathReadOnly(t *testing.T) {
	if os.Getenv("ZELIARD_ITEM_SAVE") == "" {
		t.Skip("須掛載正常建立的起始存檔")
	}
	if out := os.Getenv("ZELIARD_SHOP_OUTPUT"); out != "" {
		t.Setenv("ZELIARD_SHOP_OUTPUT", filepath.Join(out, "item-branches"))
	}
	testWeaponShopNormalPath(t, "../../../tools/item-shop-branches.commands", 63)
}

func assertItemBranchCheckpoint(t *testing.T, name string, d *Dialogue, l *Labels, o *oracle.Oracle, dst *image.RGBA, drawn bool) {
	t.Helper()
	if name == "description-list-third-attempt" || name == "description-oil-selected" {
		var rows []labelStamp
		for _, s := range l.stamps {
			if s.listViewport {
				rows = append(rows, s)
			}
		}
		if len(rows) != 2 {
			t.Fatal("捲動清單未辨識完整兩列", name, len(rows))
		}
		for _, want := range []struct {
			key string
			y   int
		}{{"zelres2.file.01b65c", 124}, {"zelres2.file.01b67c", 134}} {
			found := false
			for _, s := range rows {
				found = found || s.source == l.sources[want.key] && s.box.Min == image.Pt(112, want.y)
			}
			if !found {
				t.Fatal("捲動商品譯名／位置不符", name, want.key)
			}
		}
		completed := l.Completed
		l.ObserveFrame(o.Indexed(), o.Palette())
		l.Draw(image.NewRGBA(dst.Bounds()), o.Indexed(), o.Palette(), image.Point{})
		if l.Completed != completed {
			t.Fatal("同一畫格重複建立清單來源")
		}
		for _, bad := range []string{"border", "glyph", "palette", "missing-cache", "source"} {
			isolated := *l
			isolated.stamps = append([]labelStamp(nil), rows...)
			isolated.itemListCache = map[int]*listLabel{}
			for key, value := range l.itemListCache {
				isolated.itemListCache[key] = value
			}
			raw, pal := append([]byte(nil), o.Indexed()...), o.Palette()
			switch bad {
			case "border":
				raw[122*320+92] ^= 128
			case "glyph":
				raw[124*320+112] ^= 128
			case "palette":
				pal[9][0] ^= 1
			case "missing-cache":
				delete(isolated.itemListCache, 0x10f4)
			case "source":
				isolated.itemListPrefix = append([]byte(nil), l.itemListPrefix...)
				isolated.itemListPrefix[0] ^= 1
			}
			canvas := image.NewRGBA(dst.Bounds())
			before := append([]byte(nil), canvas.Pix...)
			if isolated.Draw(canvas, raw, pal, image.Point{}) != 0 || !bytes.Equal(before, canvas.Pix) {
				t.Fatal("捲動清單未知來源未原子退回", bad)
			}
		}
	}
	q := d.itemExplanation
	for _, kind := range []string{"kenko", "stone", "oil"} {
		if !strings.HasPrefix(name, "description-"+kind+"-page") {
			continue
		}
		wantReady := 4
		if strings.Contains(name, "page2") {
			wantReady = 5
		}
		key := map[string]string{"kenko": "zelres2.file.01b622", "stone": "zelres2.file.01b65c", "oil": "zelres2.file.01b67c"}[kind]
		if !drawn || q == nil || !q.active || q.ready != wantReady || q.item != l.sources[key] {
			t.Fatalf("商品介紹來源或中文不完整：%s drawn=%v %v", name, drawn, d.ItemExplanationState())
		}
		if _, ok := quoteLayout(q.tokens()); !ok {
			t.Fatal("商品介紹超出三行安全區", name)
		}
		if wantReady == 4 {
			if !q.prompt {
				t.Fatal("商品介紹第一頁遺失待續狀態", name)
			}
			_, _, rgb := o.ScreenRGB()
			for y := dialoguePrompt.Min.Y * 3; y < dialoguePrompt.Max.Y*3; y++ {
				for x := dialoguePrompt.Min.X * 3; x < dialoguePrompt.Max.X*3; x++ {
					i := (y/3*320 + x/3) * 3
					if dst.RGBAAt(x, y) != (color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255}) {
						t.Fatal("紅色待續三角遭覆寫", name)
					}
				}
			}
		}
	}
	wants := map[string]string{"description-list": "zelres2.file.01b02f", "description-kenko-followup": "zelres2.file.01b072", "description-stone-followup": "zelres2.file.01b072", "description-oil-followup": "zelres2.file.01b072", "empty-sell-message": "zelres2.file.01b002", "sell-potion-list": "zelres2.file.01af16", "sell-potion-result": "zelres2.file.01aeee", "sell-done-farewell": "zelres2.file.01b099", "sell-done-farewell-ready": "zelres2.file.01b099"}
	if key, ok := wants[name]; ok {
		if !drawn || d.current < 0 || d.paragraphs[d.current].key != key || d.visible != len(d.paragraphs[d.current].text) {
			t.Fatalf("商品介紹／出售靜態中文缺漏：%s %v", name, drawn)
		}
	}
	if name == "sell-potion-quote" || name == "sell-potion-quote-ready" {
		s := d.itemSale
		if !drawn || s == nil || !s.active || s.ready != 7 || s.Completed != 1 || string(s.price.raw) != "25\xff" || s.item != l.sources["zelres2.file.01b622"] || o.Word(oracle.Far(0x01b1, 0x86)) != 950 {
			t.Fatalf("出售確認前金額／來源／中文不符：%v", d.ItemSaleState())
		}
		points, ok := quoteLayout(s.tokens())
		if !ok {
			t.Fatal("出售確認超寬")
		}
		copied := 0
		for i, token := range s.tokens() {
			if n := token.number; n != nil {
				copied++
				at := points[i].Add(image.Pt(dialogueBox.Min.X*3, dialogueBox.Min.Y*3+4))
				pal := o.Palette()
				if n.box != image.Rect(138, 119, 153, 127) {
					t.Fatal("出售數字來源幾何不符", n.box)
				}
				for y := 0; y < n.box.Dy()*3; y++ {
					for x := 0; x < n.box.Dx()*3; x++ {
						v := pal[n.pixels[y/3*n.box.Dx()+x/3]]
						if dst.RGBAAt(at.X+x, at.Y+y) != (color.RGBA{v[0], v[1], v[2], 255}) {
							t.Fatal("出售25的原版字形被改寫")
						}
					}
				}
			}
		}
		if copied != 1 {
			t.Fatal("出售數字缺漏")
		}
		for _, bad := range []string{"pixel", "palette", "number"} {
			isolated := *s
			raw, pal := append([]byte(nil), o.Indexed()...), o.Palette()
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
			if isolated.Draw(canvas, raw, pal, image.Point{}) || isolated.active || !bytes.Equal(before, canvas.Pix) {
				t.Fatal("出售未知來源未原子退回", bad)
			}
		}
	}
	if name == "sell-potion-result" && (o.Word(oracle.Far(0x01b1, 0x86)) != 975 || string(d.lead) != "非常感謝。") {
		t.Fatal("出售結果或謝詞遺失", string(d.lead))
	}
	if name == "sell-done-outside" || name == "sold-inventory-closed" {
		if drawn || d.Active() || o.Word(oracle.Far(0x01b1, 0x86)) != 975 {
			t.Fatal("離店殘留或售後金額不符", name)
		}
	}
	if name == "sold-inventory-empty" {
		found := false
		for _, stamp := range l.stamps {
			found = found || stamp.source == l.sources["zelres2.file.004ab0"]
		}
		if !found || o.Word(oracle.Far(0x01b1, 0x86)) != 975 {
			t.Fatal("售後無物品標籤或金額不符")
		}
	}
}

func TestItemServiceCatalogValidation(t *testing.T) {
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
	catalog, err := os.ReadFile("../../../text/dialogue-prototype.tsv")
	if err != nil {
		t.Fatal(err)
	}
	r := csv.NewReader(bytes.NewReader(catalog))
	r.Comma = '\t'
	records, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	o, err := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
	if err != nil {
		t.Fatal(err)
	}
	defer o.Close()
	for _, service := range []string{"sale", "explanation"} {
		newChain := newItemSale
		key, last := "zelres2.file.01ae60", "zelres2.file.01af6d"
		if service == "explanation" {
			newChain = newItemExplanation
			key, last = "zelres2.file.01b053", "zelres2.file.01b504"
		}
		for _, bad := range []string{"missing-item", "unknown", "incomplete", "font"} {
			t.Run(service+"/"+bad, func(t *testing.T) {
				texts := map[string]string{}
				for _, record := range records[1:] {
					texts[record[0]] = record[1]
				}
				switch bad {
				case "missing-item":
					texts[key] = strings.ReplaceAll(texts[key], "{item}", "")
				case "unknown":
					texts[key] += "{other}"
				case "incomplete":
					delete(texts, last)
				case "font":
					texts[last] = string(rune(0x10ffff))
				}
				if _, err := newChain(o, data, texts, font, &paintHistory{}); err == nil {
					t.Fatal("新道具服務不相容catalog未拒絕")
				}
			})
		}
		if q, err := newChain(o, data, map[string]string{}, font, &paintHistory{}); err != nil || q != nil {
			t.Fatal("舊catalog須保留原貌", service)
		}
	}
}

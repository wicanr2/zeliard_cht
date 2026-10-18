package overlay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

func TestWeaponShopPurchaseNormalPathReadOnly(t *testing.T) {
	testWeaponShopNormalPath(t, "../../../tools/weapon-shop.commands", 9)
}

func TestWeaponShopInsufficientGoldNormalPathReadOnly(t *testing.T) {
	if out := os.Getenv("ZELIARD_SHOP_OUTPUT"); out != "" {
		t.Setenv("ZELIARD_SHOP_OUTPUT", filepath.Join(out, "insufficient"))
	}
	testWeaponShopNormalPath(t, "../../../tools/weapon-insufficient.commands", 11)
}

func TestInventoryNormalPathReadOnly(t *testing.T) {
	if out := os.Getenv("ZELIARD_SHOP_OUTPUT"); out != "" {
		t.Setenv("ZELIARD_SHOP_OUTPUT", filepath.Join(out, "inventory"))
	}
	testWeaponShopNormalPath(t, "../../../tools/inventory.commands", 11)
}

func TestWeaponShopNoPurchaseDepartureReadOnly(t *testing.T) {
	if out := os.Getenv("ZELIARD_SHOP_OUTPUT"); out != "" {
		t.Setenv("ZELIARD_SHOP_OUTPUT", filepath.Join(out, "no-purchase"))
	}
	testWeaponShopNormalPath(t, "../../../tools/weapon-shop-no-purchase.commands", 3)
}

func testWeaponShopNormalPath(t *testing.T, commandPath string, wantCheckpoints int) {
	t.Helper()
	root, fontPath, out := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_SHOP_OUTPUT")
	if root == "" || fontPath == "" || out == "" {
		t.Skip("須掛載原版、字型與商店收據輸出目錄")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	var machines [2]*oracle.Oracle
	for i := range machines {
		machines[i], err = oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if err != nil {
			t.Fatal(err)
		}
		machines[i].SetDOSBoxCycles(3000)
		scratch := t.TempDir()
		if strings.Contains(commandPath, "bank") || strings.Contains(commandPath, "item-shop") {
			savePath := os.Getenv("ZELIARD_BANK_SAVE")
			if strings.Contains(commandPath, "item-shop") {
				savePath = os.Getenv("ZELIARD_ITEM_SAVE")
			}
			save, e := os.ReadFile(savePath)
			if e != nil || fmtHash(save) != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
				t.Fatal("須使用自己正常建立並核對身份的銀行起始存檔", e)
			}
			if e = os.WriteFile(filepath.Join(scratch, "SAVETEST.usr"), save, 0644); e != nil {
				t.Fatal(e)
			}
			info, e := os.Stat(savePath)
			if e != nil {
				t.Fatal(e)
			}
			if e = os.Chtimes(filepath.Join(scratch, "SAVETEST.usr"), info.ModTime(), info.ModTime()); e != nil {
				t.Fatal(e)
			}
		}
		machines[i].SetScratch(scratch)
		defer machines[i].Close()
	}
	d, err := NewDialogue(machines[1], filepath.Join(root, "ZELRES2.SAR"), "../../../text/dialogue-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	l, err := NewLabels(machines[1], filepath.Join(root, "ZELRES2.SAR"), "../../../text/labels-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	d.UseLabels(l)
	var hud *HUD
	if catalog := os.Getenv("ZELIARD_HUD_TEXT"); catalog != "" {
		hud, err = NewHUD(machines[1], root, catalog, font)
		if err != nil {
			t.Fatal(err)
		}
	}
	seenHUD := map[string]bool{}
	if err = os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	commands, err := os.ReadFile(commandPath)
	if err != nil {
		t.Fatal(err)
	}
	run := func(n uint64) {
		for _, o := range machines {
			if err := o.Run(n); err != nil {
				t.Fatal(err)
			}
		}
	}
	number := func(s string) uint64 {
		n, e := strconv.ParseUint(s, 0, 64)
		if e != nil {
			t.Fatal(e)
		}
		return n
	}
	type event struct {
		step uint64
		scan uint8
		down bool
	}
	var receipts []map[string]any
	for _, line := range strings.Split(string(commands), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "king":
			var events []event
			for _, s := range []uint64{12000000, 20000000, 28000000, 36000000, 44000000, 52000000, 86000000, 94000000, 102000000, 110000000, 118000000, 126000000} {
				events = append(events, event{s, 0x39, true}, event{s + 50000, 0x39, false})
			}
			events = append(events, event{60000000, 0x4d, true}, event{62000000, 0x4d, false}, event{64000000, 0x48, true}, event{64250000, 0x48, false})
			sort.SliceStable(events, func(i, j int) bool { return events[i].step < events[j].step })
			for _, e := range events {
				run(e.step - machines[0].Steps())
				for _, o := range machines {
					if e.down {
						o.KeyDown(e.scan)
					} else {
						o.KeyUp(e.scan)
					}
				}
			}
			run(130000000 - machines[0].Steps())
		case "run":
			run(number(f[1]))
		case "key":
			for _, o := range machines {
				o.KeyDown(uint8(number(f[1])))
			}
			run(number(f[2]))
			for _, o := range machines {
				o.KeyUp(uint8(number(f[1])))
			}
		case "down", "up":
			for _, o := range machines {
				if f[0] == "down" {
					o.KeyDown(uint8(number(f[1])))
				} else {
					o.KeyUp(uint8(number(f[1])))
				}
			}
		case "shot":
			a, b := machines[0], machines[1]
			raw, pal, mem := b.Indexed(), b.Palette(), b.Bytes(oracle.Far(0, 0), 1<<20)
			if !bytes.Equal(raw, a.Indexed()) || pal != a.Palette() || !bytes.Equal(mem, a.Bytes(oracle.Far(0, 0), 1<<20)) || a.IP() != b.IP() || a.Cycles() != b.Cycles() || a.Steps() != b.Steps() {
				t.Fatalf("%s：中文改寫原版狀態", f[1])
			}
			_, _, rgb := b.ScreenRGB()
			dst := image.NewRGBA(image.Rect(0, 0, 960, 600))
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := (y/3*320 + x/3) * 3
					dst.SetRGBA(x, y, color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255})
				}
			}
			before := append([]byte(nil), dst.Pix...)
			drawn := d.Draw(dst, raw, pal, image.Point{})
			paintSerial := d.paintSerial
			if q := d.quotation; q != nil && q.active {
				paintSerial = q.paintSerial
			}
			if d.bank != nil && d.bank.active {
				paintSerial = d.bank.paintSerial
			}
			if d.itemPurchase != nil && d.itemPurchase.active {
				paintSerial = d.itemPurchase.paintSerial
			}
			for _, q := range []*itemPurchase{d.itemSale, d.itemExplanation} {
				if q != nil && q.active {
					paintSerial = q.paintSerial
				}
			}
			dialogueMasks, _ := d.paint.masks(paintSerial, raw)
			if drawn {
				for y := 0; y < 600; y++ {
					for x := 0; x < 960; x++ {
						i := dst.PixOffset(x, y)
						if occluded(image.Pt(x/3, y/3), dialogueMasks) && !bytes.Equal(before[i:i+4], dst.Pix[i:i+4]) {
							t.Fatalf("%s：對話覆蓋原版彈窗", f[1])
						}
					}
				}
			}
			l.ObserveFrame(raw, pal)
			labels := l.Draw(dst, raw, pal, image.Point{})
			maskedLabels := 0
			for _, stamp := range l.stamps {
				masks, _ := l.paint.masks(stamp.paintSerial, raw)
				intersects := false
				for _, mask := range masks {
					intersects = intersects || mask.Overlaps(stamp.box)
				}
				if !intersects {
					continue
				}
				maskedLabels++
				isolated := *l
				isolated.stamps = []labelStamp{stamp}
				canvas := image.NewRGBA(dst.Bounds())
				copy(canvas.Pix, before)
				isolated.Draw(canvas, raw, pal, image.Point{})
				for y := 0; y < 600; y++ {
					for x := 0; x < 960; x++ {
						i := canvas.PixOffset(x, y)
						if occluded(image.Pt(x/3, y/3), masks) && !bytes.Equal(before[i:i+4], canvas.Pix[i:i+4]) {
							t.Fatalf("%s：舊選單覆蓋原版彈窗", f[1])
						}
					}
				}
			}
			var boxes []image.Rectangle
			var keys []string
			for _, stamp := range l.stamps {
				boxes = append(boxes, stamp.rows...)
				for key, source := range l.sources {
					if source == stamp.source {
						keys = append(keys, key)
					}
				}
			}
			sort.Strings(keys)
			if !strings.Contains(commandPath, "insufficient") && (f[1] == "inventory" || f[1] == "released" || f[1] == "reopened") {
				want := map[string]int{"zelres2.file.004a1d": 1, "zelres2.file.004a2e": 1, "zelres2.file.004a37": 1, "zelres2.file.004a3f": 1, "zelres2.file.004ab0": 3, "zelres2.file.004d03": 1, "zelres2.file.004d0c": 1}
				got := map[string]int{}
				inventorySources := map[*labelSource]bool{}
				for k := range want {
					inventorySources[l.sources[k]] = true
				}
				for _, k := range keys {
					if _, ok := want[k]; ok {
						got[k]++
					}
				}
				if labels < 9 || len(got) != len(want) {
					t.Fatalf("裝備中文標籤缺漏：%d %v", labels, got)
				}
				for k, n := range want {
					if got[k] != n {
						t.Fatalf("裝備中文來源數量不符：%s %d", k, got[k])
					}
				}
				for _, stamp := range l.stamps {
					if !inventorySources[stamp.source] {
						isolated := *l
						isolated.stamps = []labelStamp{stamp}
						canvas := image.NewRGBA(dst.Bounds())
						copy(canvas.Pix, before)
						isolated.Draw(canvas, raw, pal, image.Point{})
						if !bytes.Equal(canvas.Pix, before) {
							t.Fatal("裝備頁透出舊選單中文")
						}
						continue
					}
					if stamp.source.inventory {
						wantColor := uint8(27)
						if stamp.source.offset == 0x9ff {
							wantColor = 18
						}
						if stamp.source.offset == 0xa92 {
							wantColor = 9
						}
						if stamp.foreground != wantColor || (wantColor != 9 && (stamp.shadow != 45 || stamp.box.Dy() != 9)) {
							t.Fatal("裝備原版標題色彩或陰影遺失")
						}
					}
					unknown := append([]byte(nil), raw...)
					unknown[stamp.box.Min.Y*320+stamp.box.Min.X] ^= 0x80
					isolated := *l
					isolated.stamps = []labelStamp{stamp}
					if isolated.Draw(image.NewRGBA(dst.Bounds()), unknown, pal, image.Point{}) != 0 {
						t.Fatal("裝備未知像素沒有退回原版")
					}
				}
			}
			key := ""
			if d.current >= 0 {
				key = d.paragraphs[d.current].key
			}
			if f[1] == "greeting" && (!drawn || key != "zelres2.file.0185ea" || labels != 5) {
				t.Fatalf("初訪中文不完整：%v %s %v", drawn, key, keys)
			}
			if strings.Contains(commandPath, "bank") {
				assertBankCheckpoint(t, f[1], d, l, b, dst, drawn)
			}
			if strings.Contains(commandPath, "item-shop") {
				if strings.Contains(commandPath, "branches") {
					assertItemBranchCheckpoint(t, f[1], d, l, b, dst, drawn)
				} else {
					assertItemCheckpoint(t, f[1], d, l, b, dst, drawn)
				}
				if strings.Contains(commandPath, "-use.") {
					assertItemUseCheckpoint(t, f[1], l, b, dst, drawn)
				}
			}
			if strings.Contains(commandPath, "weapon-shop-no-purchase") && f[1] == "no-purchase-farewell" {
				if !drawn || key != "zelres2.file.018a17" || d.visible != len(d.paragraphs[d.current].text) || b.Word(oracle.Far(0x01b1, 0x86)) != 1000 || b.Word(oracle.Far(0x01b1, 0xff4c)) != 0xb23e {
					t.Fatalf("未購物離店來源、完整中文或原版金額不符：%v %s %d", drawn, key, d.visible)
				}
			}
			if f[1] == "goods" {
				if !drawn || key != "zelres2.file.01883e" || len(dialogueMasks) == 0 {
					t.Fatalf("商品彈窗上方詢問未保留中文：%v %s", drawn, key)
				}
				for _, want := range []string{"zelres2.file.018534", "zelres2.file.018543"} {
					found := false
					for _, key := range keys {
						found = found || key == want
					}
					if !found {
						t.Fatal("缺少中文商品", want, keys)
					}
				}
			}
			if f[1] == "quote" {
				q := d.quotation
				if !drawn || q == nil || q.ready != 10 || string(q.price.raw) != "400\xff" || string(q.trade.raw) != "200\xff" || q.Completed != 1 {
					t.Fatalf("報價中文或原版數字不符：%v %v", drawn, d.QuotationState())
				}
				positions, valid := quoteLayout(q.tokens())
				if !valid {
					t.Fatal("報價超出安全文字區")
				}
				copied := 0
				for i, token := range q.tokens() {
					if n := token.number; n != nil {
						at := positions[i].Add(image.Pt(dialogueBox.Min.X*3, dialogueBox.Min.Y*3+4))
						for y := 0; y < n.box.Dy()*3; y++ {
							for x := 0; x < n.box.Dx()*3; x++ {
								v := pal[n.pixels[y/3*n.box.Dx()+x/3]]
								if dst.RGBAAt(at.X+x, at.Y+y) != (color.RGBA{v[0], v[1], v[2], 255}) {
									t.Fatal("報價數字未保留原版字形")
								}
							}
						}
						copied++
					}
				}
				if copied != 2 {
					t.Fatal("未合成兩個獨立的原版數字")
				}
				found := false
				for _, k := range keys {
					found = found || k == "zelres2.file.0184ea"
				}
				if !found || maskedLabels == 0 {
					t.Fatal("報價彈窗旁的舊選單中文已失效", keys)
				}
			}
			if f[1] == "purchased" && (!drawn || key != "zelres2.file.018634" || b.Word(oracle.Far(0x01b1, 0x86)) != 800) {
				t.Fatalf("交易後中文或原版金額不符：%v %s %d", drawn, key, b.Word(oracle.Far(0x01b1, 0x86)))
			}
			if f[1] == "insufficient-page1" || f[1] == "insufficient-page2" || f[1] == "insufficient-settled" {
				if !drawn || key != "zelres2.file.01876b" || b.Word(oracle.Far(0x01b1, 0x86)) != 1000 || d.quotation.Completed != 0 {
					t.Fatalf("資金不足分支中文、資金或報價狀態不符：%v %s %v", drawn, key, d.QuotationState())
				}
			}
			inside, outside := 0, 0
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := dst.PixOffset(x, y)
					if bytes.Equal(before[i:i+4], dst.Pix[i:i+4]) {
						continue
					}
					p := image.Pt(x/3, y/3)
					allowed := drawn && p.In(dialogueBox)
					for _, box := range boxes {
						allowed = allowed || p.In(box)
					}
					if allowed {
						inside++
					} else {
						outside++
					}
				}
			}
			if outside != 0 || !bytes.Equal(mem, b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(raw, b.Indexed()) {
				t.Fatalf("%s：覆繪越界或改寫原版", f[1])
			}
			if f[1] == "closed" && inside+outside != 0 {
				t.Fatalf("關閉裝備後仍改動畫面：pixels=%d labels=%d keys=%v dialogue=%v", inside+outside, labels, keys, drawn)
			}
			if strings.Contains(commandPath, "no-purchase") && f[1] == "outside" && (inside+outside != 0 || b.Word(oracle.Far(0x01b1, 0x86)) != 1000) {
				t.Fatal("未購物離店後有中文殘留或金額改動")
			}
			if strings.Contains(commandPath, "item-shop") && (f[1] == "item-outside" || f[1] == "item-inventory-closed") && inside+outside != 0 {
				t.Fatal("離店或關閉物品頁後殘留中文", f[1])
			}
			var hudKeys []string
			hudPixels := 0
			if hud != nil {
				hudPixels, hudKeys = assertHUDCheckpoint(t, f[1], hud, b, dst)
				for _, key := range hudKeys {
					seenHUD[key] = true
				}
				inside += hudPixels
				if !bytes.Equal(mem, b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(raw, b.Indexed()) {
					t.Fatal("狀態列覆繪改寫原版")
				}
			}
			fp, e := os.Create(filepath.Join(out, f[1]+".png"))
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(fp, dst)
			fp.Close()
			if e != nil {
				t.Fatal(e)
			}
			receipts = append(receipts, map[string]any{"name": f[1], "steps": b.Steps(), "cycles": b.Cycles(), "state_equal": true, "raw_sha256": fmtHash(raw), "memory_sha256": fmtHash(mem), "gold": b.Word(oracle.Far(0x01b1, 0x86)), "dialogue_rendered": drawn, "dialogue_key": key, "quotation": d.QuotationState(), "bank_dialogue": d.BankState(), "item_purchase": d.ItemPurchaseState(), "item_sale": d.ItemSaleState(), "item_explanation": d.ItemExplanationState(), "dialogue_lead": string(d.lead), "dialogue_masks": dialogueMasks, "masked_labels": maskedLabels, "modal_pixels_changed": 0, "label_keys": keys, "hud_keys": hudKeys, "hud_pixels_changed": hudPixels, "inside_changed": inside, "outside_changed": outside, "held_keys": b.HeldKeys()})
		default:
			t.Fatalf("未知命令 %s", line)
		}
	}
	if len(receipts) != wantCheckpoints || len(machines[1].HeldKeys()) != 0 {
		t.Fatal("正常商店路徑未完整完成")
	}
	if hud != nil && len(seenHUD) != 9 {
		t.Fatal("狀態列九個來源未全部走到", seenHUD)
	}
	scope := "normal new game, reward, Chinese quotation with original price/trade glyphs, purchase, farewell, inventory"
	if strings.Contains(commandPath, "insufficient") {
		scope = "normal new game, reward, insufficient gold dialogue/pages, farewell and town; original gold remains1000"
	}
	if strings.Contains(commandPath, "no-purchase") {
		scope = "normal new game, reward, weapon shop visit without purchase, distinct departure and town; original gold remains1000"
	}
	if strings.Contains(commandPath, "bank") {
		scope = "normal cold F7 load of own save, bank query0, direction-key amount preview, deposit100, withdraw50, query50, no almas, farewell, town; full original state equal"
	}
	if strings.Contains(commandPath, "item-shop") {
		scope = "normal cold F7 own save, MAGIC first item purchase50, continuation No, farewell, town, inventory potion icon and close; full original state equal"
	}
	if strings.Contains(commandPath, "bank-no-purchase") {
		scope = "normal cold F7 own save, bank greeting, no transaction farewell and outside1000; full original state equal"
	}
	if strings.Contains(commandPath, "branches") {
		scope = "normal cold F7 own save, three item descriptions/pages, empty sell, prerequisite purchase50, sale confirmation25 and result975, departure and inventory removal; full original state equal"
	}
	if hud != nil {
		scope += "; independent HUD layer: four fields, three scene names and two source variants; original numeric/bar/icon pixels preserved"
	}
	data, err := json.MarshalIndent(map[string]any{"commands_sha256": fmtHash(commands), "checkpoints": receipts, "dialogue_rejections": d.Rejections, "scope": scope}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("武器店中文正常路徑配對通過，%d個抽樣點\n", len(receipts))
}

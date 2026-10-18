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

func TestBankNormalPathReadOnly(t *testing.T) {
	if os.Getenv("ZELIARD_BANK_SAVE") == "" {
		t.Skip("須掛載自己正常建立的起始存檔")
	}
	if out := os.Getenv("ZELIARD_SHOP_OUTPUT"); out != "" {
		t.Setenv("ZELIARD_SHOP_OUTPUT", filepath.Join(out, "bank"))
	}
	testWeaponShopNormalPath(t, "../../../tools/bank.commands", 35)
}

func assertBankCheckpoint(t *testing.T, name string, d *Dialogue, l *Labels, o *oracle.Oracle, dst *image.RGBA, drawn bool) {
	t.Helper()
	wants := map[string][]string{
		"bank-greeting":   {"zelres2.file.019def"},
		"initial-balance": {"zelres2.file.01a05a"},
		"deposit-prompt":  {"zelres2.file.019f2d"},
		"deposit-hundred": {"zelres2.file.019f2d"},
		"deposit-result":  {"zelres2.file.019f2d", "zelres2.file.019f57"},
		"withdraw-prompt": {"zelres2.file.019fe3"},
		"withdraw-fifty":  {"zelres2.file.019fe3"},
		"withdraw-result": {"zelres2.file.019fe3", "zelres2.file.01a006", "zelres2.file.019f57"},
		"final-balance":   {"zelres2.file.01a073"},
		"no-almas":        {"zelres2.file.019e15"},
		"bank-farewell":   {"zelres2.file.01a100"},
	}
	if want, ok := wants[name]; ok {
		b := d.bank
		if b == nil || !drawn || !b.active || b.current != nil || len(b.parts) != len(want) {
			t.Fatalf("%s銀行文字鏈未完成：%v", name, d.BankState())
		}
		for i, key := range want {
			if b.parts[i].source.key != key {
				t.Fatalf("%s銀行文字順序不符：%v", name, d.BankState())
			}
		}
		wantGold := uint16(1000)
		if name == "deposit-result" || name == "withdraw-prompt" || name == "withdraw-fifty" {
			wantGold = 900
		}
		if name == "withdraw-result" || name == "final-balance" || name == "no-almas" || name == "bank-farewell" {
			wantGold = 950
		}
		if o.Word(oracle.Far(0x01b1, 0x86)) != wantGold {
			t.Fatal("原版身上金額不符", name)
		}
		wantNumber := "50\xff"
		if name == "deposit-result" {
			wantNumber = "100\xff"
		}
		for _, p := range b.parts {
			if len(p.number.raw) > 0 && string(p.number.raw) != wantNumber {
				t.Fatal("銀行沿用了錯誤的原版數字", name)
			}
		}
		tokens := b.tokens()
		points, valid := quoteLayout(tokens)
		if !valid {
			t.Fatal("銀行文字超出安全區")
		}
		for i, token := range tokens {
			if n := token.number; n != nil {
				at := points[i].Add(image.Pt(dialogueBox.Min.X*3, dialogueBox.Min.Y*3+4))
				pal := o.Palette()
				for y := 0; y < n.box.Dy()*3; y++ {
					for x := 0; x < n.box.Dx()*3; x++ {
						v := pal[n.pixels[y/3*n.box.Dx()+x/3]]
						if dst.RGBAAt(at.X+x, at.Y+y) != (color.RGBA{v[0], v[1], v[2], 255}) {
							t.Fatal("銀行數字圖形被重畫", name)
						}
					}
				}
			}
		}
		unknown := append([]byte(nil), o.Indexed()...)
		unknown[100*320+55] ^= 128
		isolated := *b
		canvas := image.NewRGBA(dst.Bounds())
		before := append([]byte(nil), canvas.Pix...)
		if isolated.Draw(canvas, unknown, o.Palette(), image.Point{}) || isolated.active || !bytes.Equal(before, canvas.Pix) {
			t.Fatal("銀行未知像素未退回原版")
		}
	}
	if name == "bank-greeting" || name == "deposit-prompt" || name == "withdraw-prompt" {
		keys := []string{"zelres2.file.019d54", "zelres2.file.019d6e", "zelres2.file.019d79", "zelres2.file.019d88", "zelres2.file.019d96", "zelres2.file.019da5"}
		if name == "deposit-prompt" {
			keys = append(keys, "zelres2.file.019db3", "zelres2.file.019dc2")
		}
		if name == "withdraw-prompt" {
			keys = append(keys, "zelres2.file.019dcf", "zelres2.file.019dde")
		}
		for _, key := range keys {
			found := false
			for _, s := range l.stamps {
				found = found || s.source == l.sources[key]
			}
			if !found {
				t.Fatal("銀行標籤缺漏", key)
			}
		}
		for _, s := range l.stamps {
			if s.source.bankTitle {
				if s.background != 5 || s.foreground != 9 || s.shadow != 45 || s.shadowOffset == nil || *s.shadowOffset != image.Pt(1, 0) || s.box != image.Rect(98, 175, 138, 183) {
					t.Fatal("銀行標題顏色或範圍不符")
				}
				isolated := *l
				isolated.stamps = []labelStamp{s}
				unknown := append([]byte(nil), o.Indexed()...)
				unknown[175*320+98] ^= 128
				if isolated.Draw(image.NewRGBA(dst.Bounds()), unknown, o.Palette(), image.Point{}) != 0 {
					t.Fatal("銀行標題未知像素未退回")
				}
			}
		}
	}
	if name == "bank-outside" && (drawn || d.bank.active || o.Word(oracle.Far(0x01b1, 0x86)) != 950) {
		t.Fatal("銀行離店殘留文字或金額錯誤")
	}
}

func TestBankCatalogValidation(t *testing.T) {
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
	keys := []string{"019def", "019e15", "019f2d", "019fe3", "01a05a", "01a100", "019f57", "01a006", "01a073"}
	for _, bad := range []string{"missing", "duplicate", "unknown", "incomplete", "font", "overflow"} {
		t.Run(bad, func(t *testing.T) {
			texts := map[string]string{}
			for i, k := range keys {
				texts["zelres2.file."+k] = "測試"
				if i >= 6 {
					texts["zelres2.file."+k] = "測試{amount}金幣。"
				}
			}
			key := "zelres2.file.019f57"
			switch bad {
			case "missing":
				texts[key] = "餘額"
			case "duplicate":
				texts[key] = "{amount}{amount}"
			case "unknown":
				texts[key] = "{amount}{other}"
			case "incomplete":
				delete(texts, "zelres2.file.019def")
			case "font":
				texts[key] = "{amount}" + string(rune(0x10ffff))
			case "overflow":
				texts[key] = string(bytes.Repeat([]byte("金"), 70)) + "{amount}"
			}
			if _, err := newBankDialogue(o, data, texts, font, &paintHistory{}); err == nil {
				t.Fatal("銀行不相容catalog未拒絕")
			}
		})
	}
	if b, err := newBankDialogue(o, data, map[string]string{}, font, &paintHistory{}); err != nil || b != nil {
		t.Fatal("舊catalog應保留原版銀行")
	}
}

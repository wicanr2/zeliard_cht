package overlay

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

func TestHUDNormalPathReadOnly(t *testing.T) {
	if os.Getenv("ZELIARD_ITEM_SAVE") == "" {
		t.Skip("須掛載自己正常建立的存檔")
	}
	t.Setenv("ZELIARD_HUD_TEXT", "../../../text/hud-prototype.tsv")
	if out := os.Getenv("ZELIARD_SHOP_OUTPUT"); out != "" {
		t.Setenv("ZELIARD_SHOP_OUTPUT", filepath.Join(out, "hud"))
	}
	testWeaponShopNormalPath(t, "../../../tools/item-shop-use.commands", 48)
}

func TestHUDCatalogValidation(t *testing.T) {
	root, fontPath := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT")
	if root == "" || fontPath == "" {
		t.Skip("須掛載原版與字型")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	key := "zelres1.file.00db4c"
	for name, text := range map[string]string{
		"duplicate": "key\tzh-TW\n" + key + "\t生命\n" + key + "\t生命\n",
		"unknown":   "key\tzh-TW\nzelres1.file.000001\t生命\n",
		"overflow":  "key\tzh-TW\n" + key + "\t生命生命生命\n",
		"control":   "key\tzh-TW\n" + key + "\t\"生\n命\"\n",
		"empty":     "key\tzh-TW\n" + key + "\t\n",
		"columns":   "key\tzh-TW\n" + key + "\t生命\textra\n",
		"encoding":  "key\tzh-TW\n" + key + "\t\xff\n",
	} {
		t.Run(name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "hud.tsv")
			if err := os.WriteFile(p, []byte(text), 0644); err != nil {
				t.Fatal(err)
			}
			if _, err := NewHUD(nil, root, p, font); err == nil {
				t.Fatal("錯誤catalog未拒絕")
			}
		})
	}
	missing := bitmap.Font{}
	for ch, g := range font {
		if ch != '生' {
			missing[ch] = g
		}
	}
	if _, err = NewHUD(nil, root, "../../../text/hud-prototype.tsv", missing); err == nil {
		t.Fatal("缺字catalog未拒絕")
	}
	data, err := os.ReadFile("../../../text/hud-prototype.tsv")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	p := filepath.Join(t.TempDir(), "missing.tsv")
	if err = os.WriteFile(p, []byte(lines[0]+"\n"+strings.Join(lines[2:], "\n")), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err = NewHUD(nil, root, p, font); err == nil || !strings.Contains(err.Error(), "來源缺漏") {
		t.Fatal("漏譯檢查未拒絕不完整catalog", err)
	}
}

func assertHUDCheckpoint(t *testing.T, name string, h *HUD, o *oracle.Oracle, dst *image.RGBA) (int, []string) {
	t.Helper()
	before := append([]byte(nil), dst.Pix...)
	raw, pal := o.Indexed(), o.Palette()
	h.Draw(dst, raw, pal, image.Point{})
	keys := h.VisibleKeys()
	wants := map[string]string{"zelres1.file.00db4c": "生命", "zelres1.file.00db54": "寶石", "zelres1.file.00db5d": "金幣", "zelres1.file.00db65": "地點", "zelres2.file.000d34": "金幣", "zelres2.file.000d3c": "地點", "zelres2.file.038dce": "Felishika城堡", "zelres2.file.0399b8": "Muralla鎮", "zelres3.file.00c31a": "Malicia洞窟", "zelres2.file.017238": "Felishika國王", "zelres2.file.017809": "小屋", "zelres2.file.0184c9": "武器防具店", "zelres2.file.01c997": "Marid長老"}
	var boxes []image.Rectangle
	for _, key := range keys {
		s := h.sources[key]
		if string(s.text) != wants[key] {
			t.Fatal("狀態列譯文不符", key)
		}
		boxes = append(boxes, s.box)
		isolated := *h
		isolated.fields = map[string]*hudSource{s.kind: s}
		canvas := inventoryOriginalCanvas(o)
		original := append([]byte(nil), canvas.Pix...)
		if isolated.Draw(canvas, raw, pal, image.Point{}) != 1 || bytes.Equal(original, canvas.Pix) {
			t.Fatal("狀態列完整中文沒有實際像素", name, key)
		}
		// 每字的前景墨點數須等於該字型；只擦除英文背景不能通過。
		x, y := s.box.Min.X*3, s.box.Min.Y*3
		for _, ch := range s.text {
			g := h.font[ch]
			w := len(g) / 16 * 8
			dw := w * 3 / 2
			want, got := 0, 0
			for yy := 0; yy < 24; yy++ {
				for xx := 0; xx < dw; xx++ {
					if g[(yy*2/3)*(w/8)+(xx*2/3)/8]&(0x80>>uint((xx*2/3)%8)) != 0 {
						want++
					}
					c := pal[s.foreground]
					if canvas.RGBAAt(x+xx, y+yy) == (color.RGBA{c[0], c[1], c[2], 255}) {
						got++
					}
				}
			}
			if want == 0 || want != got {
				t.Fatal("狀態列中文字模墨點不符", name, key, ch, want, got)
			}
			x += dw
		}
		if name == "restored" {
			for _, bad := range []string{"pixels", "palette", "source", "prefix", "short"} {
				copyHUD := isolated
				copySource := *s
				copyHUD.fields = map[string]*hudSource{s.kind: &copySource}
				r, p := append([]byte(nil), raw...), pal
				switch bad {
				case "pixels":
					r[s.box.Min.Y*320+s.box.Min.X] ^= 128
				case "palette":
					p[9][0] ^= 1
				case "source":
					copySource.raw = append([]byte(nil), s.raw...)
					copySource.raw[4] ^= 1
				case "prefix":
					copySource.prefix = append([]byte(nil), s.prefix...)
					copySource.prefix[0] ^= 1
				case "short":
					r = r[:len(r)-1]
				}
				canvas := image.NewRGBA(dst.Bounds())
				old := append([]byte(nil), canvas.Pix...)
				if copyHUD.Draw(canvas, r, p, image.Point{}) != 0 || !bytes.Equal(old, canvas.Pix) {
					t.Fatal("狀態列未知欄位未原子退回", key, bad)
				}
				if bad == "pixels" && copyHUD.Draw(canvas, raw, pal, image.Point{}) != 1 {
					t.Fatal("原版還原同一完整欄位後未恢復中文")
				}
			}
		}
	}
	var expected []string
	switch name {
	case "cold-town":
		expected = []string{"zelres1.file.00db4c", "zelres1.file.00db54", "zelres1.file.00db5d", "zelres1.file.00db65", "zelres2.file.038dce"}
	case "restored", "item-outside", "item-inventory", "item-inventory-closed", "use-return-town":
		expected = []string{"zelres1.file.00db4c", "zelres1.file.00db54", "zelres1.file.00db5d", "zelres1.file.00db65", "zelres2.file.0399b8"}
	case "cave-ready", "cave-item-inventory", "cave-potion-selected", "cave-potion-used", "cave-use-closed", "cave-inventory-empty", "cave-inventory-closed":
		expected = []string{"zelres1.file.00db4c", "zelres1.file.00db54", "zelres2.file.000d34", "zelres2.file.000d3c", "zelres3.file.00c31a"}
	}
	if expected != nil && strings.Join(expected, "|") != strings.Join(keys, "|") {
		t.Fatal("狀態列或場景中文缺漏", name, keys)
	}
	if name == "cave-ready" {
		for _, key := range []string{"zelres1.file.00db4c", "zelres1.file.00db54"} {
			s := h.sources[key]
			if bytes.Equal(o.Bytes(oracle.Far(0x01b1, s.base), 64), s.prefix) {
				t.Fatal("未真正驗到原版卸載狀態列來源模組")
			}
		}
	}
	changed := 0
	for y := 0; y < 600; y++ {
		for x := 0; x < 960; x++ {
			i := dst.PixOffset(x, y)
			if bytes.Equal(before[i:i+4], dst.Pix[i:i+4]) {
				continue
			}
			changed++
			allowed := false
			for _, box := range boxes {
				allowed = allowed || image.Pt(x/3, y/3).In(box)
			}
			if !allowed {
				t.Fatal("狀態列覆繪超出安全區", name, x, y)
			}
		}
	}
	_, _, rgb := o.ScreenRGB()
	for _, box := range []image.Rectangle{image.Rect(84, 163, 187, 171), image.Rect(88, 187, 120, 195), image.Rect(160, 187, 187, 195)} {
		for y := box.Min.Y * 3; y < box.Max.Y*3; y++ {
			for x := box.Min.X * 3; x < box.Max.X*3; x++ {
				i := (y/3*320 + x/3) * 3
				if dst.RGBAAt(x, y) != (color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255}) {
					t.Fatal("狀態列改動原版生命條或數字", name, box)
				}
			}
		}
	}
	return changed, keys
}

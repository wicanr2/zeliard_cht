package hd

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

func sceneryTestInputs(t *testing.T) (*TownScenery, string) {
	t.Helper()
	project := os.Getenv("ZELIARD_PROJECT")
	root := os.Getenv("ZELIARD_ORIG")
	fontPath := os.Getenv("ZELIARD_FONT")
	if project == "" || root == "" || fontPath == "" {
		t.Skip("須掛載本地原版、字型與正常來源；skip不算驗收")
	}
	base := filepath.Join(project, "workplace/town-hd-v1/foreground-v1")
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewTownScenery(root, filepath.Join(base, "normal-v1/restored.memory"), filepath.Join(base, "normal-v1/restored.palette"), base, filepath.Join(project, "text/town-signs-prototype.tsv"), font)
	if err != nil {
		t.Fatal(err)
	}
	return s, base
}

func TestTownSceneryOriginalFramesAndActorIsolation(t *testing.T) {
	s, base := sceneryTestInputs(t)
	cases := []string{"normal-v1/cold-town", "normal-v1/restored", "normal-v1/citizen-near", "normal-v1/citizen-outside", "normal-v1/citizen-page", "normal-v1/citizen-last", "scene-cave-v1/cave-near", "scene-cave-v1/cave-entered"}
	for _, n := range []string{"1", "2", "3", "4", "5", "6", "7"} {
		cases = append(cases, "scene-walk-v1/walk-left-"+n)
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			memory, err := os.ReadFile(filepath.Join(base, name+".memory"))
			if err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(filepath.Join(base, name+".indexed"))
			if err != nil {
				t.Fatal(err)
			}
			rgb, dst := skyOriginal(raw, s.palette)
			before := append([]byte(nil), dst.Pix...)
			original := append([]byte(nil), memory...)
			p, r, e := memory[0x1b10:0x11b10], memory[0x11b10:0x21b10], memory[0x21b10:0x31b10]
			want := !strings.Contains(name, "cold-town") && !strings.Contains(name, "walk-left-7") && !strings.Contains(name, "cave-entered") && !strings.Contains(name, "citizen-page") && !strings.Contains(name, "citizen-last")
			if got := s.Draw(dst, 0x13, p, r, e, raw, rgb, image.Point{}, true); got != want {
				t.Fatalf("匹配 %v，預期 %v", got, want)
			}
			if !bytes.Equal(memory, original) {
				t.Fatal("覆繪改記憶體")
			}
			if !want {
				if !bytes.Equal(dst.Pix, before) {
					t.Fatal("拒絕卻寫畫布")
				}
				return
			}
			// Independent Python source decoder masks, already verified against original VGA.
			maskName := filepath.Base(name) + "-sprite-opacity.pgm"
			mask, err := os.ReadFile(filepath.Join(base, maskName))
			if err != nil {
				t.Fatal(err)
			}
			header := []byte("P5\n224 64\n255\n")
			if !bytes.HasPrefix(mask, header) || len(mask) != len(header)+14336 {
				t.Fatal("獨立遮罩形狀不符")
			}
			mask = mask[len(header):]
			changed := 0
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := dst.PixOffset(x, y)
					if bytes.Equal(before[i:i+4], dst.Pix[i:i+4]) {
						continue
					}
					changed++
					if x < 144 || x >= 816 || y < 234 || y >= 426 {
						t.Fatal("前景框外寫入", x, y)
					}
					if mask[((y-234)/3)*224+(x-144)/3] != 0 {
						t.Fatal("覆寫原版角色", x, y)
					}
				}
			}
			if changed < 60000 {
				t.Fatal("前景未完整接入", changed)
			}
		})
	}
}

func TestTownSceneryFallbackAndSignBounds(t *testing.T) {
	s, base := sceneryTestInputs(t)
	memory, err := os.ReadFile(filepath.Join(base, "normal-v1/restored.memory"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(base, "normal-v1/restored.indexed"))
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"mode", "primary", "resources", "extra", "indexed", "rgb", "graphics", "player-code", "atlas", "masks", "module", "camera", "actor-table", "player-phase", "display-cache", "native-pixel", "rgb-pixel", "canvas"} {
		t.Run(bad, func(t *testing.T) {
			p := append([]byte(nil), memory[0x1b10:0x11b10]...)
			r := append([]byte(nil), memory[0x11b10:0x21b10]...)
			e := append([]byte(nil), memory[0x21b10:0x31b10]...)
			indices := append([]byte(nil), raw...)
			rgb, dst := skyOriginal(indices, s.palette)
			mode := uint8(0x13)
			switch bad {
			case "mode":
				mode = 3
			case "primary":
				p = p[:100]
			case "resources":
				r = r[:100]
			case "extra":
				e = e[:100]
			case "indexed":
				indices = indices[:100]
			case "rgb":
				rgb = rgb[:100]
			case "graphics":
				p[0x3000] ^= 1
			case "player-code":
				p[0x6975] ^= 1
			case "atlas":
				r[0x4100] ^= 1
			case "masks":
				e[0x7000] ^= 1
			case "module":
				p[0xc000] ^= 1
			case "camera":
				p[0x80] = 255
				p[0x81] = 255
			case "actor-table":
				p[0xc00f] = 255
				p[0xc010] = 255
			case "player-phase":
				p[0xe7] = 255
			case "display-cache":
				p[0xe000] ^= 128
			case "native-pixel":
				indices[78*320+48] ^= 1
			case "rgb-pixel":
				rgb[(78*320+48)*3] ^= 1
			case "canvas":
				dst = image.NewRGBA(image.Rect(0, 0, 300, 200))
			}
			before := append([]byte(nil), dst.Pix...)
			if s.Draw(dst, mode, p, r, e, indices, rgb, image.Point{}, true) || !bytes.Equal(dst.Pix, before) {
				t.Fatal("失配未整層退回")
			}
		})
	}
	for _, sign := range s.signs {
		if sign.width > 88 || sign.width <= 0 {
			t.Fatal("招牌寬度不符")
		}
		for _, r := range sign.text {
			if len(s.font[r]) != 32 {
				t.Fatal("預期16pixel中文字模")
			}
		}
	}
	// Missing one panel rejects the entire group instead of a partial atlas.
	if _, err := NewTownScenery(os.Getenv("ZELIARD_ORIG"), filepath.Join(base, "normal-v1/restored.memory"), filepath.Join(base, "normal-v1/restored.palette"), t.TempDir(), "", s.font); err == nil {
		t.Fatal("缺圖未拒絕")
	}
	for _, label := range []string{"魔法物品魔法物品", "缺\t欄", "\xff", "\x01", " 銀行"} {
		tmp := filepath.Join(t.TempDir(), "signs.tsv")
		data := "key\tworld_x\tsource\tzh_TW\nweapon\t321\tWEAPON\t" + label + "\nmagic\t897\tMAGIC\t魔法物品\nbank\t1113\tBANK\t銀行\n"
		if err := os.WriteFile(tmp, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := NewTownScenery(os.Getenv("ZELIARD_ORIG"), filepath.Join(base, "normal-v1/restored.memory"), filepath.Join(base, "normal-v1/restored.palette"), base, tmp, s.font); err == nil {
			t.Fatal("招牌不符未拒絕")
		}
	}
}

package hd

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestItemShopFramesRegionsAndFallback(t *testing.T) {
	var refs, arts [4]string
	for i := range refs {
		refs[i], arts[i] = os.Getenv(fmt.Sprintf("ZELIARD_ITEM_REF_%d", i)), os.Getenv(fmt.Sprintf("ZELIARD_ITEM_ART_%d", i))
		if refs[i] == "" || arts[i] == "" {
			t.Skip("須掛載道具店四組原版畫格與圖面")
		}
	}
	palette := os.Getenv("ZELIARD_ITEM_PALETTE")
	m, err := NewItemShop(refs, palette, arts)
	if err != nil {
		t.Fatal(err)
	}
	pal, err := os.ReadFile(palette)
	if err != nil {
		t.Fatal(err)
	}
	for pose, path := range refs {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		rgb := make([]byte, len(raw)*3)
		for i, v := range raw {
			copy(rgb[i*3:i*3+3], pal[int(v)*3:int(v)*3+3])
		}
		beforeRaw, beforeRGB := append([]byte(nil), raw...), append([]byte(nil), rgb...)
		dst := image.NewRGBA(image.Rect(0, 0, 960, 680))
		for y := 48; y < 648; y++ {
			for x := 0; x < 960; x++ {
				i := ((y-48)/3*320 + x/3) * 3
				dst.SetRGBA(x, y, color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255})
			}
		}
		original := append([]byte(nil), dst.Pix...)
		matched, got := m.Draw(dst, raw, rgb, image.Pt(0, 48))
		if !matched || got != pose {
			t.Fatal("未匹配道具店原版畫格", pose, got)
		}
		inner := itemShopBox.Inset(1)
		box := image.Rect(inner.Min.X*scale, inner.Min.Y*scale+48, inner.Max.X*scale, inner.Max.Y*scale+48)
		changed := 0
		for y := 0; y < 680; y++ {
			for x := 0; x < 960; x++ {
				i := dst.PixOffset(x, y)
				if !bytes.Equal(original[i:i+4], dst.Pix[i:i+4]) {
					if !image.Pt(x, y).In(box) {
						t.Fatal("道具店HD改寫內框以外", x, y)
					}
					changed++
				}
			}
		}
		if changed < 50000 || !bytes.Equal(raw, beforeRaw) || !bytes.Equal(rgb, beforeRGB) {
			t.Fatal("HD未完整顯示或改寫原版", changed)
		}
		if out := os.Getenv("ZELIARD_ITEM_HD_OUTPUT"); out != "" {
			f, err := os.Create(filepath.Join(out, fmt.Sprintf("composed-%d.png", pose)))
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(f, dst)
			f.Close()
			if err != nil {
				t.Fatal(err)
			}
		}
		after := append([]byte(nil), dst.Pix...)
		for _, bad := range []string{"index", "rgb", "short"} {
			i, c := append([]byte(nil), raw...), append([]byte(nil), rgb...)
			switch bad {
			case "index":
				i[23*320+56] ^= 1
			case "rgb":
				c[(24*320+57)*3] ^= 1
			case "short":
				i = i[:10]
			}
			if yes, _ := m.Draw(dst, i, c, image.Pt(0, 48)); yes || !bytes.Equal(after, dst.Pix) {
				t.Fatal("未原子退回", bad)
			}
		}
	}
	// 兩個蒸氣相位不改人物，抬手相位不改右側蒸氣，其他背景四組一致。
	for _, pair := range [][2]int{{0, 1}, {0, 2}, {2, 3}} {
		allowed := itemShopSteam
		if pair == [2]int{0, 2} {
			allowed = itemShopCharacter
		}
		changes := 0
		for y := 0; y < 186; y++ {
			for x := 0; x < 282; x++ {
				if m.art[pair[0]].RGBAAt(x, y) != m.art[pair[1]].RGBAAt(x, y) {
					changes++
					if !image.Pt(x, y).In(allowed) {
						t.Fatal("動畫改動已觀察區域以外", pair, x, y)
					}
				}
			}
		}
		if changes == 0 {
			t.Fatal("動畫相位無差異", pair)
		}
	}
	badRefs, badArts := refs, arts
	badRefs[1], badArts[2] = refs[0], arts[2]+".missing"
	if _, err := NewItemShop(badRefs, palette, arts); err == nil {
		t.Fatal("錯誤來源未拒絕")
	}
	if _, err := NewItemShop(refs, palette, badArts); err == nil {
		t.Fatal("缺圖未拒絕")
	}
}

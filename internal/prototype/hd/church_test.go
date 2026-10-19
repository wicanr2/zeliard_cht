package hd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestChurchRegisteredFramesAndFallback(t *testing.T) {
	refs, assets, expected := os.Getenv("ZELIARD_CHURCH_REFERENCE"), os.Getenv("ZELIARD_CHURCH_ART"), os.Getenv("ZELIARD_CHURCH_EXPECTED")
	transients := os.Getenv("ZELIARD_CHURCH_TRANSIENT_DIR")
	if refs == "" || assets == "" || expected == "" || transients == "" {
		t.Skip("須掛載教堂15個原版畫格、七張手繪圖、獨立註冊預覽及繪製中來源")
	}
	m, err := NewChurch(refs, assets)
	if err != nil {
		t.Fatal(err)
	}
	pal, err := os.ReadFile(filepath.Join(refs, "church-entered.palette"))
	if err != nil {
		t.Fatal(err)
	}
	origin := image.Pt(0, 48)
	nativeInner := churchBox.Inset(1)
	inner := image.Rectangle{Min: nativeInner.Min.Mul(scale), Max: nativeInner.Max.Mul(scale)}.Add(origin)
	for frame, ref := range churchReferences {
		t.Run(fmt.Sprintf("frame%d", frame), func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(refs, ref.file))
			if err != nil {
				t.Fatal(err)
			}
			rgb := make([]byte, len(raw)*3)
			for i, value := range raw {
				copy(rgb[i*3:i*3+3], pal[int(value)*3:int(value)*3+3])
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
			if yes, got := m.Draw(dst, raw, rgb, origin); !yes || got != frame {
				t.Fatal("教堂原版畫格未匹配", yes, got)
			}
			f, err := os.Open(filepath.Join(expected, fmt.Sprintf("target-%d.png", frame)))
			if err != nil {
				t.Fatal(err)
			}
			want, err := png.Decode(f)
			f.Close()
			if err != nil || want.Bounds() != image.Rect(0, 0, 282, 186) {
				t.Fatal("獨立高清預覽形狀不符", err)
			}
			changed := 0
			for y := 0; y < 680; y++ {
				for x := 0; x < 960; x++ {
					i := dst.PixOffset(x, y)
					if !bytes.Equal(dst.Pix[i:i+4], original[i:i+4]) {
						if !image.Pt(x, y).In(inner) {
							t.Fatal("教堂HD改寫內框以外", x, y)
						}
						changed++
					}
					if image.Pt(x, y).In(inner) && dst.RGBAAt(x, y) != color.RGBAModel.Convert(want.At(x-inner.Min.X, y-inner.Min.Y)).(color.RGBA) {
						t.Fatal("HD與獨立Python註冊預覽不同", x, y)
					}
				}
			}
			if changed < 40000 || !bytes.Equal(raw, beforeRaw) || !bytes.Equal(rgb, beforeRGB) {
				t.Fatal("HD沒有完整顯示或改動原版輸入", changed)
			}
			after := append([]byte(nil), dst.Pix...)
			badRaw := append([]byte(nil), raw...)
			badRaw[23*320+112] ^= 1
			badRGB := append([]byte(nil), rgb...)
			badRGB[(24*320+113)*3] ^= 1
			for _, bad := range []struct{ indexed, rgb []byte }{{badRaw, rgb}, {raw, badRGB}, {raw[:10], rgb}, {raw, rgb[:10]}} {
				if yes, _ := m.Draw(dst, bad.indexed, bad.rgb, origin); yes || !bytes.Equal(after, dst.Pix) {
					t.Fatal("未知來源、RGB或短畫格未原子退回")
				}
			}
			if yes, _ := m.Draw(nil, raw, rgb, origin); yes {
				t.Fatal("接受空畫布")
			}
			small := image.NewRGBA(image.Rect(0, 0, inner.Max.X-1, inner.Max.Y))
			before := append([]byte(nil), small.Pix...)
			if yes, _ := m.Draw(small, raw, rgb, origin); yes || !bytes.Equal(before, small.Pix) {
				t.Fatal("不完整輸出畫布未原子退回")
			}
		})
	}
	var transientIndex struct {
		Frames []struct {
			File string `json:"file"`
		} `json:"paint_in_progress_candidates"`
	}
	indexData, err := os.ReadFile(filepath.Join(transients, "inventory.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(indexData, &transientIndex); err != nil {
		t.Fatal(err)
	}
	if len(transientIndex.Frames) != 5 {
		t.Fatal("繪製中來源數量不符")
	}
	for _, f := range transientIndex.Frames {
		raw, err := os.ReadFile(filepath.Join(transients, f.File))
		if err != nil {
			t.Fatal(err)
		}
		rgb := make([]byte, len(raw)*3)
		for i, value := range raw {
			copy(rgb[i*3:i*3+3], pal[int(value)*3:int(value)*3+3])
		}
		dst := image.NewRGBA(image.Rect(0, 0, 960, 680))
		before := append([]byte(nil), dst.Pix...)
		if yes, _ := m.Draw(dst, raw, rgb, origin); yes || !bytes.Equal(before, dst.Pix) {
			t.Fatal("繪製中來源冒充完整手勢", f.File)
		}
	}
	for _, dirs := range [][2]string{{"", assets}, {refs, ""}, {refs, t.TempDir()}} {
		if _, err := NewChurch(dirs[0], dirs[1]); err == nil {
			t.Fatal("缺少參考或素材未拒絕整組")
		}
	}
	badArt := t.TempDir()
	for i, a := range churchArtworks {
		source := a.file
		if i == 0 {
			source = churchArtworks[1].file
		}
		if err := os.Symlink(filepath.Join(assets, source), filepath.Join(badArt, a.file)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := NewChurch(refs, badArt); err == nil {
		t.Fatal("錯誤素材身份未拒絕整組")
	}
	badRefs := t.TempDir()
	if err := os.Symlink(filepath.Join(refs, "church-entered.palette"), filepath.Join(badRefs, "church-entered.palette")); err != nil {
		t.Fatal(err)
	}
	for i, ref := range churchReferences {
		source := ref.file
		if i == len(churchReferences)-1 {
			source = churchReferences[0].file
		}
		if err := os.Symlink(filepath.Join(refs, source), filepath.Join(badRefs, ref.file)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := NewChurch(badRefs, assets); err == nil {
		t.Fatal("錯誤最後來源未拒絕整組")
	}
}

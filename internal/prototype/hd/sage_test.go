package hd

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestSagePortraitPreservesOriginalAnimation(t *testing.T) {
	a, b, palette, art := os.Getenv("ZELIARD_SAGE_A"), os.Getenv("ZELIARD_SAGE_B"), os.Getenv("ZELIARD_SAGE_PALETTE"), os.Getenv("ZELIARD_SAGE_ART")
	if a == "" || b == "" || palette == "" || art == "" {
		t.Skip("須掛載長老原版兩畫格、色盤及本地手繪圖")
	}
	s, err := NewSage(a, b, palette, art)
	if err != nil {
		t.Fatal(err)
	}
	pal, err := os.ReadFile(palette)
	if err != nil {
		t.Fatal(err)
	}
	animated := 0
	for _, on := range s.animated {
		if on {
			animated++
		}
	}
	if animated != 4 {
		t.Fatal("原版動畫像素數量不符", animated)
	}
	for pose, path := range []string{a, b} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		beforeRaw := append([]byte(nil), raw...)
		rgb := make([]byte, len(raw)*3)
		for i, v := range raw {
			copy(rgb[i*3:i*3+3], pal[int(v)*3:int(v)*3+3])
		}
		beforeRGB := append([]byte(nil), rgb...)
		dst := image.NewRGBA(image.Rect(0, 0, 960, 680))
		for y := 0; y < 600; y++ {
			for x := 0; x < 960; x++ {
				i := (y/3*320 + x/3) * 3
				dst.SetRGBA(x, y+48, color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255})
			}
		}
		original := append([]byte(nil), dst.Pix...)
		matched, got := s.Draw(dst, raw, rgb, image.Pt(0, 48))
		if !matched || got != pose {
			t.Fatal("原版長老畫格未匹配", pose, got)
		}
		changed := 0
		for y := 0; y < 680; y++ {
			for x := 0; x < 960; x++ {
				i := dst.PixOffset(x, y)
				if bytes.Equal(original[i:i+4], dst.Pix[i:i+4]) {
					continue
				}
				p := image.Pt(x/3, (y-48)/3)
				if !p.In(sageBox.Inset(1)) {
					t.Fatal("改寫長老內框以外像素", x, y)
				}
				local := p.Sub(sageBox.Min)
				if s.animated[local.Y*sageBox.Dx()+local.X] {
					t.Fatal("改寫原版動畫像素", p)
				}
				changed++
			}
		}
		if changed < 50000 {
			t.Fatal("長老高清圖面沒有完整顯示", changed)
		}
		if !bytes.Equal(raw, beforeRaw) || !bytes.Equal(rgb, beforeRGB) {
			t.Fatal("改寫原版畫格或RGB")
		}
		copyAfter := append([]byte(nil), dst.Pix...)
		raw[23*320+56] ^= 1
		if yes, _ := s.Draw(dst, raw, rgb, image.Point{}); yes || !bytes.Equal(copyAfter, dst.Pix) {
			t.Fatal("未知畫格未完整退回")
		}
		raw[23*320+56] ^= 1
		rgb[(24*320+57)*3] ^= 1
		if yes, _ := s.Draw(dst, raw, rgb, image.Point{}); yes || !bytes.Equal(copyAfter, dst.Pix) {
			t.Fatal("色盤失配未完整退回")
		}
		if yes, _ := s.Draw(dst, raw[:10], rgb, image.Point{}); yes {
			t.Fatal("短畫格被接受")
		}
	}
	if _, err := NewSage(a, b, palette, art+".missing"); err == nil {
		t.Fatal("缺圖未拒絕整組素材")
	}
	if _, err := NewSage(a, a, palette, art); err == nil {
		t.Fatal("錯誤第二畫格未拒絕")
	}
}

func TestSageCenteredSourceFrames(t *testing.T) {
	project, study := os.Getenv("ZELIARD_PROJECT"), os.Getenv("ZELIARD_CAVE_STUDY")
	if project == "" || study == "" {
		t.Skip("須掛載兩個真實置中肖像來源")
	}
	palette := filepath.Join(study, "exploration-return-source-v1/10-revival-message-full.palette")
	read := func(path string) []byte {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	pal := read(palette)
	s, e := NewSage(filepath.Join(project, "workplace/save-path-verified/sage-menu.indexed"), filepath.Join(project, "workplace/save-path-verified/save-confirm-text.indexed"), palette, filepath.Join(project, "workplace/hd-sage-prototype.png"))
	if e != nil {
		t.Fatal(e)
	}
	art, e := png.Decode(bytes.NewReader(read(filepath.Join(project, "workplace/hd-sage-prototype.png"))))
	if e != nil {
		t.Fatal(e)
	}
	rgbOf := func(raw []byte) []byte {
		rgb := make([]byte, len(raw)*3)
		for i, v := range raw {
			copy(rgb[i*3:i*3+3], pal[int(v)*3:int(v)*3+3])
		}
		return rgb
	}
	for n := 0; n < 2; n++ {
		t.Run(string(rune('0'+n)), func(t *testing.T) {
			stem := filepath.Join(study, "exploration-return-source-v1", fmtSageSource(n))
			var record struct {
				SHA string `json:"portrait_indexed_sha256"`
			}
			if e := json.Unmarshal(read(stem+".json"), &record); e != nil {
				t.Fatal(e)
			}
			raw := read(stem + ".indexed")
			rgb := rgbOf(raw)
			beforeRaw, beforeRGB := bytes.Clone(raw), bytes.Clone(rgb)
			pose := -1
			switch record.SHA {
			case "bbc79b87bc9d9b3bb59a049d5acd8206a7b9dfb287d557b71437672544d3f16c":
				pose = 0
			case "0762364b30958db2e1948f850d449ff042a8f736ea5b62d8674be9296c90a66a":
				pose = 1
			}
			if pose < 0 {
				t.Fatal("未知原版置中來源")
			}
			origin := image.Pt(0, 48)
			dst := image.NewRGBA(image.Rect(0, 0, 960, 680))
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := (y/3*320 + x/3) * 3
					dst.SetRGBA(x, y+48, color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255})
				}
			}
			old := bytes.Clone(dst.Pix)
			if ok, p := s.Draw(dst, raw, rgb, origin); !ok || p != pose {
				t.Fatal("真實置中來源未匹配", p, pose)
			}
			for y := 0; y < 680; y++ {
				for x := 0; x < 960; x++ {
					px, py := x-112*3, y-23*3-48
					i := dst.PixOffset(x, y)
					inside := px >= 3 && px < 285 && py >= 3 && py < 189
					if !inside || s.animated[py/3*96+px/3] {
						if !bytes.Equal(dst.Pix[i:i+4], old[i:i+4]) {
							t.Fatal("覆寫外框、動畫或區外", x, y)
						}
						continue
					}
					want := color.RGBAModel.Convert(art.At(px*1536/288, py*1024/192)).(color.RGBA)
					if dst.RGBAAt(x, y) != want {
						t.Fatal("置中未按獨立直接取樣", x, y)
					}
				}
			}
			if !bytes.Equal(raw, beforeRaw) || !bytes.Equal(rgb, beforeRGB) {
				t.Fatal("改寫原版輸入")
			}
			for _, bad := range []string{"source", "RGB", "position", "canvas"} {
				rawBad, rgbBad := bytes.Clone(raw), bytes.Clone(rgb)
				canvas := image.NewRGBA(dst.Bounds())
				switch bad {
				case "source":
					rawBad[23*320+112] ^= 1
				case "RGB":
					rgbBad[(23*320+112)*3] ^= 1
				case "position":
					rawBad = make([]byte, 64000)
					for y := 0; y < 64; y++ {
						copy(rawBad[(23+y)*320+110:], raw[(23+y)*320+112:(23+y)*320+208])
					}
					rgbBad = rgbOf(rawBad)
				case "canvas":
					canvas = image.NewRGBA(image.Rect(0, 0, 630, 300))
				}
				before := bytes.Clone(canvas.Pix)
				if ok, _ := s.Draw(canvas, rawBad, rgbBad, origin); ok || !bytes.Equal(before, canvas.Pix) {
					t.Fatal("失配未原子退回", bad)
				}
			}
		})
	}
}

func fmtSageSource(n int) string {
	if n == 0 {
		return "portrait-00"
	}
	return "portrait-01"
}

package hd

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func caveOriginalCanvas(f *CavePresentation) *image.RGBA {
	c := image.NewRGBA(image.Rect(0, 0, 960, 600))
	for y := 0; y < 600; y++ {
		for x := 0; x < 960; x++ {
			v := int(f.Indexed[(y/3)*320+x/3])
			c.SetRGBA(x, y, color.RGBA{f.Palette[v*3], f.Palette[v*3+1], f.Palette[v*3+2], 255})
		}
	}
	return c
}
func cavePresentedFixture(t *testing.T, root string, p caveFixturePoint) *CavePresentation {
	t.Helper()
	in, body := caveFixture(t, root, p)
	f := &CaveFrames{body: &body}
	f.finish(in)
	if f.complete == nil {
		t.Fatal(f.lastRefusal)
	}
	return f.selectFrame(in, false)
}

func TestCaveTerrainIndependentOfflinePixels(t *testing.T) {
	root, points := caveFixtures(t)
	art, e := NewCaveTerrain(root)
	if e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"cave-idle-1", "advance-strike-2", "jump-key-500k", "down-key-250k"} {
		t.Run(name, func(t *testing.T) {
			var point caveFixturePoint
			for _, p := range points {
				if p.Name == name+"-pc-7114" {
					point = p
				}
			}
			if point.Name == "" {
				t.Fatal("來源不存在")
			}
			f := cavePresentedFixture(t, root, point)
			canvas := caveOriginalCanvas(f)
			before := cloneTownBytes(canvas.Pix)
			sourceMemory, bodyMemory := cloneTownBytes(f.Memory), cloneTownBytes(f.BodyMemory)
			if !art.Draw(canvas, f, image.Point{}) {
				t.Fatal("正常來源HD被拒絕")
			}
			file, e := os.Open(filepath.Join(root, "previews-v2", name+"-complete-terrain-hd-v5.png"))
			if e != nil {
				t.Fatal(e)
			}
			defer file.Close()
			want, e := png.Decode(file)
			if e != nil {
				t.Fatal(e)
			}
			changed := 0
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := canvas.PixOffset(x, y)
					if image.Pt(x, y).In(image.Rect(144, 42, 816, 474)) {
						expect := color.RGBAModel.Convert(want.At(x, y)).(color.RGBA)
						if canvas.RGBAAt(x, y) != expect {
							t.Fatalf("與獨立Python預覽不同 %d,%d: got%v want%v", x, y, canvas.RGBAAt(x, y), expect)
						}
						if !bytes.Equal(before[i:i+4], canvas.Pix[i:i+4]) {
							changed++
						}
					} else if !bytes.Equal(before[i:i+4], canvas.Pix[i:i+4]) {
						t.Fatal("區外被寫入", x, y)
					}
				}
			}
			if changed == 0 {
				t.Fatal("高清圖沒有可見改變")
			}
			for _, layer := range [][]cavePixel{f.Decoded.Dynamic, f.Decoded.Actor, f.Decoded.Weapon} {
				for _, v := range layer {
					for yy := 0; yy < 3; yy++ {
						for xx := 0; xx < 3; xx++ {
							i := canvas.PixOffset(v.X*3+xx, v.Y*3+yy)
							if !bytes.Equal(canvas.Pix[i:i+4], before[i:i+4]) {
								t.Fatal("原版物件像素改變", v)
							}
						}
					}
				}
			}
			if !bytes.Equal(f.Memory, sourceMemory) || !bytes.Equal(f.BodyMemory, bodyMemory) {
				t.Fatal("HD改寫來源")
			}
		})
	}
}

func TestCaveTerrainAtomicFallback(t *testing.T) {
	root, points := caveFixtures(t)
	art, e := NewCaveTerrain(root)
	if e != nil {
		t.Fatal(e)
	}
	cases := map[string]func(*CavePresentation){
		"memory-size":       func(f *CavePresentation) { f.Memory = f.Memory[:100] },
		"body-size":         func(f *CavePresentation) { f.BodyMemory = f.BodyMemory[:100] },
		"RGB-size":          func(f *CavePresentation) { f.RGB = f.RGB[:100] },
		"indexed-size":      func(f *CavePresentation) { f.Indexed = f.Indexed[:100] },
		"palette":           func(f *CavePresentation) { f.Palette[1] ^= 1 },
		"module":            func(f *CavePresentation) { f.Memory[0x1b10+0xc000] ^= 1 },
		"tile-identity":     func(f *CavePresentation) { f.Memory[0x11b10+0x8030] ^= 1 },
		"unknown-equipment": func(f *CavePresentation) { f.BodyMemory[0x1b10+0x93] = 1 },
		"source-pixel":      func(f *CavePresentation) { f.SourceIndexed[100*320+100] ^= 1 },
		"shown-pixel":       func(f *CavePresentation) { f.Indexed[100*320+100] ^= 1 },
		"RGB":               func(f *CavePresentation) { f.RGB[(100*320+100)*3] ^= 1 },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			f := cavePresentedFixture(t, root, points[20])
			canvas := caveOriginalCanvas(f)
			before := cloneTownBytes(canvas.Pix)
			change(f)
			if art.Draw(canvas, f, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
				t.Fatal("拒絕前已寫畫布")
			}
		})
	}
	f := cavePresentedFixture(t, root, points[20])
	canvas := caveOriginalCanvas(f)
	before := cloneTownBytes(canvas.Pix)
	var absent *CaveTerrain
	if absent.Draw(canvas, f, image.Point{}) || art.Draw(canvas, nil, image.Point{}) || art.Draw(nil, f, image.Point{}) || art.Draw(canvas, f, image.Pt(1000, 1000)) || !bytes.Equal(before, canvas.Pix) {
		t.Fatal("缺少輸入或越界沒有原子退回")
	}
	if x, e := NewCaveTerrain(t.TempDir()); e == nil || x != nil {
		t.Fatal("缺圖被接受")
	}
	temp := t.TempDir()
	if e = os.WriteFile(filepath.Join(temp, "terrain-handpaint-v2.png"), []byte("broken"), 0644); e != nil {
		t.Fatal(e)
	}
	if x, e := NewCaveTerrain(temp); e == nil || x != nil {
		t.Fatal("損壞圖被接受")
	}
}

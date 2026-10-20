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

func peligroArt(t *testing.T) *CaveTerrain {
	t.Helper()
	root, _ := caveFixtures(t)
	art, err := NewCaveTerrain(root)
	if err != nil {
		t.Fatal(err)
	}
	b := iceStudy(t)
	if err = art.LoadIce(filepath.Join(b, "cave-ice-complete-v1")); err != nil {
		t.Fatal(err)
	}
	if err = art.LoadPeligro(filepath.Join(b, "cave-peligro-complete-v1")); err != nil {
		t.Fatal(err)
	}
	return art
}

func TestCavePeligroIndependentStaticMaps(t *testing.T) {
	art := peligroArt(t)
	checkCaveStaticMaps(t, art, "MPP2", "cave-peligro-complete-v1", []string{"MP20", "MP21", "MP2D"})
	// 同一呈現器也須保持原冰窟的獨立像素，不能因先載入MPP2而換錯圖集。
	checkCaveStaticMaps(t, art, "MPP4", "cave-ice-complete-v1", []string{"MP40", "MP41", "MP4D"})
}

func TestCavePeligroAllStructuralFrames(t *testing.T) {
	art := peligroArt(t)
	grid := make([]byte, 504)
	for i := 1; i <= 34; i++ {
		grid[i-1] = byte(i)
	}
	f, _ := caveStructuralFrame(t, "MPP2", "MP20", grid)
	canvas := caveOriginalCanvas(f)
	if !art.Draw(canvas, f, image.Point{}) {
		t.Fatal("34格結構來源未全部接受")
	}
	atlas, err := png.Decode(bytes.NewReader(iceRead(t, filepath.Join(iceStudy(t), "cave-peligro-complete-v1/peligro-tiles-v1.png"))))
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []int{1, 2, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34} {
		n := id - 1
		changed := 0
		for y := 0; y < 24; y++ {
			for x := 0; x < 24; x++ {
				c := color.NRGBAModel.Convert(atlas.At((id-1)*24+x, y)).(color.NRGBA)
				want := color.RGBA{uint8((uint32(c.R)*uint32(c.A) + 127) / 255), uint8((uint32(c.G)*uint32(c.A) + 127) / 255), uint8((uint32(c.B)*uint32(c.A) + 127) / 255), 255}
				got := canvas.RGBAAt(144+n%28*24+x, 42+n/28*24+y)
				if got != want {
					t.Fatalf("圖格%d取樣不符 %d,%d", id, x, y)
				}
				if got.R != 0 || got.G != 0 || got.B != 0 {
					changed++
				}
			}
		}
		if changed == 0 {
			t.Fatal("圖格沒有可見像素", id)
		}
	}
}

func TestCavePeligroAtomicLoadAndRefusal(t *testing.T) {
	art := peligroArt(t)
	b := iceStudy(t)
	loaded, ice := art.peligro, art.ice
	broken := t.TempDir()
	if err := os.WriteFile(filepath.Join(broken, "peligro-tiles-v1.png"), iceRead(t, filepath.Join(b, "cave-peligro-complete-v1/peligro-tiles-v1.png")), 0644); err != nil {
		t.Fatal(err)
	}
	if err := art.LoadPeligro(broken); err == nil || art.peligro != loaded || art.ice != ice {
		t.Fatal("缺材質破壞已載圖集")
	}
	if err := os.WriteFile(filepath.Join(broken, "peligro-material-v1.png"), []byte("broken"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := art.LoadPeligro(broken); err == nil || art.peligro != loaded || art.ice != ice {
		t.Fatal("損壞圖面被接受")
	}
	for _, count := range []int{0, 128} {
		if _, err := loadCaveTerrainArt(broken, "peligro", "", "", count); err == nil {
			t.Fatal("圖格數越界未拒絕", count)
		}
	}
	grid := iceRead(t, filepath.Join(b, "cave-peligro-complete-v1/MP20-grid.bin"))
	cases := map[string]func(*CavePresentation){
		"different-map": func(f *CavePresentation) {
			copy(f.BodyMemory[0x1b10+0xc000:], iceRead(t, filepath.Join(b, "cave-catalog-fixtures-v1/MP21.header")))
		},
		"last-source-byte": func(f *CavePresentation) { f.Memory[0x11b10+0x8030+34*48-1] ^= 1 },
		"body-resource":    func(f *CavePresentation) { f.BodyMemory[0x11b10+0x8030] ^= 1 },
		"shown-pixel":      func(f *CavePresentation) { f.Indexed[100*320+100] ^= 1 },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			f, _ := caveStructuralFrame(t, "MPP2", "MP20", grid)
			c := caveOriginalCanvas(f)
			before := cloneTownBytes(c.Pix)
			change(f)
			if art.Draw(c, f, image.Point{}) || !bytes.Equal(before, c.Pix) {
				t.Fatal("拒絕前已寫畫布")
			}
		})
	}
	f, _ := caveStructuralFrame(t, "MPP2", "MP20", grid)
	c := caveOriginalCanvas(f)
	before := cloneTownBytes(c.Pix)
	art.peligro = nil
	if art.Draw(c, f, image.Point{}) || !bytes.Equal(before, c.Pix) {
		t.Fatal("誤用已載冰窟圖面")
	}
	art.peligro = loaded
	last := loaded.tiles[34]
	loaded.tiles[34] = nil
	if art.Draw(c, f, image.Point{}) || !bytes.Equal(before, c.Pix) {
		t.Fatal("缺最後圖格未原子拒絕")
	}
	loaded.tiles[34] = last
	grid[0] = 35
	f, _ = caveStructuralFrame(t, "MPP2", "MP20", grid)
	c = caveOriginalCanvas(f)
	before = cloneTownBytes(c.Pix)
	if art.Draw(c, f, image.Point{}) || !bytes.Equal(before, c.Pix) {
		t.Fatal("未知ID35未原子拒絕")
	}
}

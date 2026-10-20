package hd

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func mpp1Art(t *testing.T) *CaveTerrain {
	t.Helper()
	art := remainingCaveArt(t)
	if err := art.LoadMPP1(filepath.Join(iceStudy(t), "cave-mpp1-complete-v2")); err != nil {
		t.Fatal(err)
	}
	return art
}

func TestCaveMPP1IndependentStaticMaps(t *testing.T) {
	art := mpp1Art(t)
	b := filepath.Join(iceStudy(t), "cave-mpp1-complete-v2")
	for _, view := range []struct{ mapName, tag string }{{"MP10", "MP10-Satono"}, {"MP10", "MP10-Muralla"}, {"MP1D", "MP1D"}} {
		t.Run(view.tag, func(t *testing.T) {
			grid := iceRead(t, filepath.Join(b, view.tag+"-grid.bin"))
			f, actor := caveStructuralFrame(t, "MPP1", view.mapName, grid)
			canvas := caveOriginalCanvas(f)
			before := cloneTownBytes(canvas.Pix)
			memory, body := cloneTownBytes(f.Memory), cloneTownBytes(f.BodyMemory)
			preview, err := png.Decode(bytes.NewReader(iceRead(t, filepath.Join(b, view.tag+"-offline-hd.png"))))
			if err != nil {
				t.Fatal(err)
			}
			want := image.NewRGBA(canvas.Bounds())
			copy(want.Pix, before)
			draw.Draw(want, image.Rect(144, 42, 816, 474), preview, image.Point{}, draw.Src)
			for _, v := range actor {
				c := color.RGBA{f.Palette[int(v.Color)*3], f.Palette[int(v.Color)*3+1], f.Palette[int(v.Color)*3+2], 255}
				draw.Draw(want, image.Rect(v.X*3, v.Y*3, v.X*3+3, v.Y*3+3), &image.Uniform{C: c}, image.Point{}, draw.Src)
			}
			if !art.Draw(canvas, f, image.Point{}) {
				t.Fatal("結構來源被拒絕")
			}
			if !bytes.Equal(canvas.Pix, want.Pix) {
				for y := 0; y < 600; y++ {
					for x := 0; x < 960; x++ {
						if canvas.RGBAAt(x, y) != want.RGBAAt(x, y) {
							t.Fatalf("獨立像素不符%d,%d got%v want%v", x, y, canvas.RGBAAt(x, y), want.RGBAAt(x, y))
						}
					}
				}
			}
			if !bytes.Equal(memory, f.Memory) || !bytes.Equal(body, f.BodyMemory) {
				t.Fatal("HD回寫來源")
			}
		})
	}
	for _, c := range remainingCaveCases {
		checkCaveStaticMaps(t, art, c.family, "cave-"+strings.ToLower(c.family)+"-complete-v1", c.maps)
	}
	for _, old := range []struct {
		key, folder string
		maps        []string
	}{
		{"MPP2", "cave-peligro-complete-v1", []string{"MP20", "MP21", "MP2D"}},
		{"MPP3", "cave-mpp3-complete-v2", []string{"MP30", "MP31", "MP3D"}},
		{"MPP4", "cave-ice-complete-v1", []string{"MP40", "MP41", "MP4D"}},
		{"MPP5", "cave-mpp5-complete-v1", []string{"MP50", "MP51", "MP5D"}},
		{"MPP6", "cave-mpp6-complete-v2", []string{"MP60", "MP61", "MP62", "MP6D"}},
		{"MPP7", "cave-mpp7-complete-v1", []string{"MP70", "MP71", "MP72", "MP7D"}},
	} {
		checkCaveStaticMaps(t, art, old.key, old.folder, old.maps)
	}
}

func TestCaveMPP1LegacyCompatibilityAndSlots(t *testing.T) {
	art := mpp1Art(t)
	grid := make([]byte, 504)
	for id := 1; id <= 25; id++ {
		grid[id-1] = byte(id)
	}
	f, _ := caveStructuralFrame(t, "MPP1", "MP10", grid)
	canvas := caveOriginalCanvas(f)
	before := caveOriginalCanvas(f)
	legacy := *art
	legacy.mpp1 = nil
	if !legacy.Draw(before, f, image.Point{}) || !art.Draw(canvas, f, image.Point{}) {
		t.Fatal("25格來源未接受")
	}
	if art.mpp1.count != 25 || &art.mpp1.material[0] != &art.material[0] {
		t.Fatal("既有材質被替換")
	}
	for id := 1; id <= 14; id++ {
		if !slices.Equal(art.tiles[id], art.mpp1.tiles[id]) {
			t.Fatal("既有圖格被替換", id)
		}
		for y := 42; y < 66; y++ {
			for x := 144 + (id-1)*24; x < 144+id*24; x++ {
				if before.RGBAAt(x, y) != canvas.RGBAAt(x, y) {
					t.Fatal("舊圖格或材質混色改變", id, x, y)
				}
			}
		}
	}
	atlas, err := png.Decode(bytes.NewReader(iceRead(t, filepath.Join(iceStudy(t), "cave-mpp1-complete-v2/mpp1-extra-tiles-v1.png"))))
	if err != nil {
		t.Fatal(err)
	}
	for id := 15; id <= 25; id++ {
		raw, err := caveStatic(f.Memory[0x11b10:0x21b10], byte(id))
		if err != nil {
			t.Fatal(err)
		}
		changed := 0
		for y := 0; y < 24; y++ {
			for x := 0; x < 24; x++ {
				px, py := 144+(id-1)*24+x, 42+y
				old := before.RGBAAt(px, py)
				want := old
				c := color.NRGBAModel.Convert(atlas.At((id-15)*24+x, y)).(color.NRGBA)
				if id >= 16 || raw[y/3*8+x/3] != 0 {
					if id >= 16 {
						old = color.RGBA{A: 255}
					}
					a, inv := uint32(c.A), uint32(255-c.A)
					want = color.RGBA{uint8((uint32(c.R)*a + uint32(old.R)*inv + 127) / 255), uint8((uint32(c.G)*a + uint32(old.G)*inv + 127) / 255), uint8((uint32(c.B)*a + uint32(old.B)*inv + 127) / 255), 255}
				}
				if canvas.RGBAAt(px, py) != want {
					t.Fatalf("新增圖格%d不符%d,%d", id, x, y)
				}
				if canvas.RGBAAt(px, py) != before.RGBAAt(px, py) {
					changed++
				}
			}
		}
		if changed == 0 {
			t.Fatal("新增圖格未呈現", id)
		}
	}
}

func TestCaveMPP1AtomicLoadAndRefusal(t *testing.T) {
	art := mpp1Art(t)
	loaded, other := art.mpp1, art.mpp8
	broken := t.TempDir()
	if err := art.LoadMPP1(broken); err == nil || art.mpp1 != loaded || art.mpp8 != other {
		t.Fatal("缺圖未原子拒絕")
	}
	if err := os.WriteFile(filepath.Join(broken, "mpp1-extra-tiles-v1.png"), []byte("broken"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := art.LoadMPP1(broken); err == nil || art.mpp1 != loaded {
		t.Fatal("損壞圖面未拒絕")
	}
	grid := iceRead(t, filepath.Join(iceStudy(t), "cave-mpp1-complete-v2/MP10-Satono-grid.bin"))
	for name, change := range map[string]func(*CavePresentation){
		"last-resource": func(f *CavePresentation) { f.Memory[0x11b10+0x8030+25*48-1] ^= 1 },
		"body-resource": func(f *CavePresentation) { f.BodyMemory[0x11b10+0x8030] ^= 1 },
		"different-map": func(f *CavePresentation) { f.BodyMemory[0x1b10+0xc000] ^= 1 },
		"shown-pixel":   func(f *CavePresentation) { f.Indexed[100*320+100] ^= 1 },
	} {
		t.Run(name, func(t *testing.T) {
			f, _ := caveStructuralFrame(t, "MPP1", "MP10", grid)
			c := caveOriginalCanvas(f)
			before := cloneTownBytes(c.Pix)
			change(f)
			if art.Draw(c, f, image.Point{}) || !bytes.Equal(before, c.Pix) {
				t.Fatal("來源失配未原子拒絕")
			}
		})
	}
	f, _ := caveStructuralFrame(t, "MPP1", "MP10", grid)
	c := caveOriginalCanvas(f)
	before := cloneTownBytes(c.Pix)
	refuse := func() {
		t.Helper()
		if art.Draw(c, f, image.Point{}) || !bytes.Equal(before, c.Pix) {
			t.Fatal("缺格或錯配未原子拒絕")
		}
	}
	last := loaded.tiles[25]
	loaded.tiles[25] = nil
	refuse()
	loaded.tiles[25] = last
	art.mpp1 = art.mpp8
	refuse()
	art.mpp1 = loaded
	grid[0] = 26
	f, _ = caveStructuralFrame(t, "MPP1", "MP10", grid)
	c = caveOriginalCanvas(f)
	before = cloneTownBytes(c.Pix)
	refuse()
	grid[0] = 73
	f, _ = caveStructuralFrame(t, "MPP1", "MP10", grid)
	c = caveOriginalCanvas(f)
	before = cloneTownBytes(c.Pix)
	refuse() // 同一入口ID若沒有完整矩陣仍須拒絕。
	grid = make([]byte, 504)
	for y, row := range cavePortalMatrix {
		for x, id := range row {
			grid[y*28+x] = id
		}
	}
	f, _ = caveStructuralFrame(t, "MPP1", "MP10", grid)
	c = caveOriginalCanvas(f)
	before = cloneTownBytes(c.Pix)
	if !art.Draw(c, f, image.Point{}) {
		t.Fatal("完整已驗入口矩陣被拒絕")
	}
	for y := 66; y < 42+5*24; y++ {
		for x := 144; x < 144+5*24; x++ {
			i := c.PixOffset(x, y)
			if !bytes.Equal(before[i:i+4], c.Pix[i:i+4]) {
				t.Fatal("共用入口下層被改寫", x, y)
			}
		}
	}
	for _, body := range []bool{false, true} {
		f, _ = caveStructuralFrame(t, "MPP1", "MP10", grid)
		if body {
			f.BodyMemory[0x11b10+0x8030+(97-1)*48] ^= 1
		} else {
			f.Memory[0x11b10+0x8030+(97-1)*48] ^= 1
		}
		c = caveOriginalCanvas(f)
		before = cloneTownBytes(c.Pix)
		refuse()
	}
}

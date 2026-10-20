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

func mpp7Art(t *testing.T) *CaveTerrain {
	t.Helper()
	art := mpp6Art(t)
	if err := art.LoadMPP7(filepath.Join(iceStudy(t), "cave-mpp7-complete-v1")); err != nil {
		t.Fatal(err)
	}
	return art
}

func TestCaveMPP7IndependentStaticMaps(t *testing.T) {
	art := mpp7Art(t)
	checkCaveStaticMaps(t, art, "MPP7", "cave-mpp7-complete-v1", []string{"MP70", "MP71", "MP72", "MP7D"})
	checkCaveStaticMaps(t, art, "MPP6", "cave-mpp6-complete-v2", []string{"MP60", "MP61", "MP62", "MP6D"})
	checkCaveStaticMaps(t, art, "MPP5", "cave-mpp5-complete-v1", []string{"MP50", "MP51", "MP5D"})
	checkCaveStaticMaps(t, art, "MPP3", "cave-mpp3-complete-v2", []string{"MP30", "MP31", "MP3D"})
	checkCaveStaticMaps(t, art, "MPP2", "cave-peligro-complete-v1", []string{"MP20", "MP21", "MP2D"})
	checkCaveStaticMaps(t, art, "MPP4", "cave-ice-complete-v1", []string{"MP40", "MP41", "MP4D"})
}

func TestCaveMPP7AllStructuralFrames(t *testing.T) {
	art := mpp7Art(t)
	folder := filepath.Join(iceStudy(t), "cave-mpp7-complete-v1")
	var meta struct {
		MaterialIDs  []int `json:"material_ids"`
		SecondaryIDs []int `json:"secondary_material_ids"`
		TextIDs      []int `json:"text_ids"`
		BlankIDs     []int `json:"blank_ids"`
	}
	if err := json.Unmarshal(iceRead(t, filepath.Join(folder, "manifest.json")), &meta); err != nil {
		t.Fatal(err)
	}
	var material, secondary, text, blank [128]bool
	for _, id := range meta.MaterialIDs {
		material[id] = true
	}
	for _, id := range meta.SecondaryIDs {
		material[id] = true
		secondary[id] = true
	}
	for _, id := range meta.TextIDs {
		text[id] = true
	}
	for _, id := range meta.BlankIDs {
		blank[id] = true
	}
	for i := 0; i < 128; i++ {
		if art.mpp7.materialIDs[i] != material[i] || art.mpp7.secondaryIDs[i] != secondary[i] || art.mpp7.textIDs[i] != text[i] {
			t.Fatal("來源分組不符", i)
		}
	}
	if art.mpp7.count != 61 || art.mpp7.textFirst != 0 || art.mpp7.textLast != 0 {
		t.Fatal("圖格／字帶範圍不符")
	}
	grid := make([]byte, 504)
	for i := 1; i <= 61; i++ {
		grid[i-1] = byte(i)
	}
	f, _ := caveStructuralFrame(t, "MPP7", "MP70", grid)
	canvas := caveOriginalCanvas(f)
	if !art.Draw(canvas, f, image.Point{}) {
		t.Fatal("61槽結構來源未全部接受")
	}
	atlas, err := png.Decode(bytes.NewReader(iceRead(t, filepath.Join(folder, "mpp7-tiles-v1.png"))))
	if err != nil {
		t.Fatal(err)
	}
	for id := 1; id <= 61; id++ {
		if material[id] {
			continue
		}
		n, changed := id-1, 0
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
		if blank[id] {
			raw, err := caveStatic(f.Memory[0x11b10:0x21b10], byte(id))
			if err != nil || !bytes.Equal(raw, make([]byte, 64)) || changed != 0 {
				t.Fatal("原版空白槽或HD透明槽不符", id)
			}
		} else if changed == 0 {
			t.Fatal("圖格沒有可見像素", id)
		}
	}
}

func TestCaveMPP7AtomicLoadAndRefusal(t *testing.T) {
	art := mpp7Art(t)
	b := iceStudy(t)
	loaded, ice, peligro, mpp3, mpp5, mpp6 := art.mpp7, art.ice, art.peligro, art.mpp3, art.mpp5, art.mpp6
	broken := t.TempDir()
	if err := os.WriteFile(filepath.Join(broken, "mpp7-tiles-v1.png"), iceRead(t, filepath.Join(b, "cave-mpp7-complete-v1/mpp7-tiles-v1.png")), 0644); err != nil {
		t.Fatal(err)
	}
	if err := art.LoadMPP7(broken); err == nil || art.mpp7 != loaded || art.ice != ice || art.peligro != peligro || art.mpp3 != mpp3 || art.mpp5 != mpp5 || art.mpp6 != mpp6 {
		t.Fatal("缺材質破壞已載圖集")
	}
	if err := os.WriteFile(filepath.Join(broken, "mpp7-material-v1.png"), iceRead(t, filepath.Join(b, "cave-mpp7-complete-v1/mpp7-material-v1.png")), 0644); err != nil {
		t.Fatal(err)
	}
	if err := art.LoadMPP7(broken); err == nil || art.mpp7 != loaded || art.mpp5 != mpp5 || art.mpp6 != mpp6 {
		t.Fatal("缺第二材質破壞已載圖集")
	}
	if err := os.WriteFile(filepath.Join(broken, "mpp7-secondary-material-v1.png"), []byte("broken"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := art.LoadMPP7(broken); err == nil || art.mpp7 != loaded {
		t.Fatal("損壞第二材質被接受")
	}
	if err := os.WriteFile(filepath.Join(broken, "mpp7-material-v1.png"), []byte("broken"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := art.LoadMPP7(broken); err == nil || art.mpp7 != loaded {
		t.Fatal("損壞圖面被接受")
	}
	grid := iceRead(t, filepath.Join(b, "cave-mpp7-complete-v1/MP70-grid.bin"))
	cases := map[string]func(*CavePresentation){
		"different-map": func(f *CavePresentation) {
			copy(f.BodyMemory[0x1b10+0xc000:], iceRead(t, filepath.Join(b, "cave-catalog-fixtures-v1/MP71.header")))
		},
		"last-source-byte": func(f *CavePresentation) { f.Memory[0x11b10+0x8030+61*48-1] ^= 1 },
		"body-resource":    func(f *CavePresentation) { f.BodyMemory[0x11b10+0x8030] ^= 1 },
		"shown-pixel":      func(f *CavePresentation) { f.Indexed[100*320+100] ^= 1 },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			f, _ := caveStructuralFrame(t, "MPP7", "MP70", grid)
			c := caveOriginalCanvas(f)
			before := cloneTownBytes(c.Pix)
			change(f)
			if art.Draw(c, f, image.Point{}) || !bytes.Equal(before, c.Pix) {
				t.Fatal("拒絕前已寫畫布")
			}
		})
	}
	f, _ := caveStructuralFrame(t, "MPP7", "MP70", grid)
	c := caveOriginalCanvas(f)
	before := cloneTownBytes(c.Pix)
	art.mpp7 = nil
	if art.Draw(c, f, image.Point{}) || !bytes.Equal(before, c.Pix) {
		t.Fatal("誤用其他家族圖面")
	}
	art.mpp7 = peligro
	if art.Draw(c, f, image.Point{}) || !bytes.Equal(before, c.Pix) {
		t.Fatal("來源／圖格數錯配未拒絕")
	}
	art.mpp7 = loaded
	last := loaded.tiles[61]
	loaded.tiles[61] = nil
	if art.Draw(c, f, image.Point{}) || !bytes.Equal(before, c.Pix) {
		t.Fatal("缺最後圖格未原子拒絕")
	}
	loaded.tiles[61] = last
	secondaryPixels := loaded.secondaryMaterial
	loaded.secondaryMaterial = nil
	if art.Draw(c, f, image.Point{}) || !bytes.Equal(before, c.Pix) {
		t.Fatal("缺第二材質未原子拒絕")
	}
	loaded.secondaryMaterial = secondaryPixels
	grid[0] = 62
	f, _ = caveStructuralFrame(t, "MPP7", "MP70", grid)
	c = caveOriginalCanvas(f)
	before = cloneTownBytes(c.Pix)
	if art.Draw(c, f, image.Point{}) || !bytes.Equal(before, c.Pix) {
		t.Fatal("未知ID62未原子拒絕")
	}
}

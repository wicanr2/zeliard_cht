package hd

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type remainingCaveCase struct {
	family string
	count  int
	maps   []string
	load   func(*CaveTerrain, string) error
	field  func(*CaveTerrain) **caveTerrainArt
}

var remainingCaveCases = []remainingCaveCase{
	{"MPP8", 43, []string{"MP80", "MP81", "MP82", "MP83", "MP84", "MP8D"}, (*CaveTerrain).LoadMPP8, func(a *CaveTerrain) **caveTerrainArt { return &a.mpp8 }},
	{"MPP9", 38, []string{"MP90"}, (*CaveTerrain).LoadMPP9, func(a *CaveTerrain) **caveTerrainArt { return &a.mpp9 }},
	{"MPPA", 20, []string{"MPA0"}, (*CaveTerrain).LoadMPPA, func(a *CaveTerrain) **caveTerrainArt { return &a.mppa }},
	{"MPPB", 5, []string{"MP73"}, (*CaveTerrain).LoadMPPB, func(a *CaveTerrain) **caveTerrainArt { return &a.mppb }},
}

func remainingCaveArt(t *testing.T) *CaveTerrain {
	t.Helper()
	art := mpp7Art(t)
	for _, c := range remainingCaveCases {
		if err := c.load(art, filepath.Join(iceStudy(t), "cave-"+strings.ToLower(c.family)+"-complete-v1")); err != nil {
			t.Fatal(err)
		}
	}
	return art
}

// 正常MPP1回歸可同時載四個新家族，未指定環境時沿既有測試範圍。
func loadRemainingNormalTestArt(t *testing.T, art *CaveTerrain) {
	t.Helper()
	if dir := os.Getenv("ZELIARD_CAVE_DOOR_ART"); dir != "" {
		if err := art.LoadDoors(dir); err != nil {
			t.Fatal(err)
		}
	}
	if dir := os.Getenv("ZELIARD_CAVE_MAGIA_ART"); dir != "" {
		if err := art.LoadMagia(dir); err != nil {
			t.Fatal(err)
		}
	}
	if dir := os.Getenv("ZELIARD_CAVE_COMMON_ART"); dir != "" {
		if err := art.LoadCommon(dir); err != nil {
			t.Fatal(err)
		}
	}
	if dir := os.Getenv("ZELIARD_CAVE_MPP1_ART"); dir != "" {
		if err := art.LoadMPP1(dir); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range remainingCaveCases {
		if dir := os.Getenv("ZELIARD_CAVE_" + c.family + "_ART"); dir != "" {
			if err := c.load(art, dir); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestCaveRemainingFamiliesIndependentStaticMaps(t *testing.T) {
	art := remainingCaveArt(t)
	for _, c := range remainingCaveCases {
		checkCaveStaticMaps(t, art, c.family, "cave-"+strings.ToLower(c.family)+"-complete-v1", c.maps)
	}
	for _, old := range []struct {
		family, folder string
		maps           []string
	}{
		{"MPP7", "cave-mpp7-complete-v1", []string{"MP70", "MP71", "MP72", "MP7D"}},
		{"MPP6", "cave-mpp6-complete-v2", []string{"MP60", "MP61", "MP62", "MP6D"}},
		{"MPP5", "cave-mpp5-complete-v1", []string{"MP50", "MP51", "MP5D"}},
		{"MPP3", "cave-mpp3-complete-v2", []string{"MP30", "MP31", "MP3D"}},
		{"MPP2", "cave-peligro-complete-v1", []string{"MP20", "MP21", "MP2D"}},
		{"MPP4", "cave-ice-complete-v1", []string{"MP40", "MP41", "MP4D"}},
	} {
		checkCaveStaticMaps(t, art, old.family, old.folder, old.maps)
	}
}

func TestCaveRemainingFamiliesAllStructuralSlots(t *testing.T) {
	art := remainingCaveArt(t)
	for _, c := range remainingCaveCases {
		t.Run(c.family, func(t *testing.T) {
			folder := filepath.Join(iceStudy(t), "cave-"+strings.ToLower(c.family)+"-complete-v1")
			var meta struct {
				MaterialIDs  []int `json:"material_ids"`
				SecondaryIDs []int `json:"secondary_material_ids"`
				TextIDs      []int `json:"text_ids"`
				BlankIDs     []int `json:"blank_ids"`
			}
			if err := json.Unmarshal(iceRead(t, filepath.Join(folder, "manifest.json")), &meta); err != nil {
				t.Fatal(err)
			}
			var materials, secondary, text, blank [128]bool
			for _, id := range meta.MaterialIDs {
				materials[id] = true
			}
			for _, id := range meta.SecondaryIDs {
				materials[id] = true
				secondary[id] = true
			}
			for _, id := range meta.TextIDs {
				text[id] = true
			}
			for _, id := range meta.BlankIDs {
				blank[id] = true
			}
			loaded := *c.field(art)
			if loaded.count != c.count || loaded.materialIDs != materials || loaded.secondaryIDs != secondary || loaded.textIDs != text || loaded.textFirst != 0 || loaded.textLast != 0 {
				t.Fatal("來源分組不符")
			}
			grid := make([]byte, 504)
			for id := 1; id <= c.count; id++ {
				grid[id-1] = byte(id)
			}
			f, _ := caveStructuralFrame(t, c.family, c.maps[0], grid)
			canvas := caveOriginalCanvas(f)
			if !art.Draw(canvas, f, image.Point{}) {
				t.Fatal("全槽來源未接受")
			}
			atlas, err := png.Decode(bytes.NewReader(iceRead(t, filepath.Join(folder, strings.ToLower(c.family)+"-tiles-v1.png"))))
			if err != nil {
				t.Fatal(err)
			}
			for id := 1; id <= c.count; id++ {
				if materials[id] {
					continue
				}
				n, changed := id-1, 0
				for y := 0; y < 24; y++ {
					for x := 0; x < 24; x++ {
						pixel := color.NRGBAModel.Convert(atlas.At((id-1)*24+x, y)).(color.NRGBA)
						want := color.RGBA{uint8((uint32(pixel.R)*uint32(pixel.A) + 127) / 255), uint8((uint32(pixel.G)*uint32(pixel.A) + 127) / 255), uint8((uint32(pixel.B)*uint32(pixel.A) + 127) / 255), 255}
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
						t.Fatal("原版空白槽不符", id)
					}
				} else if changed == 0 {
					t.Fatal("缺少可見圖格", id)
				}
			}
		})
	}
}

func TestCaveRemainingFamiliesAtomicLoadAndRefusal(t *testing.T) {
	art := remainingCaveArt(t)
	for _, c := range remainingCaveCases {
		t.Run(c.family, func(t *testing.T) {
			prefix := strings.ToLower(c.family)
			folder := filepath.Join(iceStudy(t), "cave-"+prefix+"-complete-v1")
			loaded := *c.field(art)
			old := []*caveTerrainArt{art.ice, art.peligro, art.mpp3, art.mpp5, art.mpp6, art.mpp7, art.mpp8, art.mpp9, art.mppa, art.mppb}
			broken := t.TempDir()
			if err := os.WriteFile(filepath.Join(broken, prefix+"-tiles-v1.png"), iceRead(t, filepath.Join(folder, prefix+"-tiles-v1.png")), 0644); err != nil {
				t.Fatal(err)
			}
			if err := c.load(art, broken); err == nil || *c.field(art) != loaded {
				t.Fatal("缺材質未原子拒絕")
			}
			if c.family == "MPPA" {
				if err := os.WriteFile(filepath.Join(broken, prefix+"-material-v1.png"), iceRead(t, filepath.Join(folder, prefix+"-material-v1.png")), 0644); err != nil {
					t.Fatal(err)
				}
				if err := c.load(art, broken); err == nil || *c.field(art) != loaded {
					t.Fatal("缺第二材質未原子拒絕")
				}
			}
			if err := os.WriteFile(filepath.Join(broken, prefix+"-material-v1.png"), []byte("broken"), 0644); err != nil {
				t.Fatal(err)
			}
			if err := c.load(art, broken); err == nil || *c.field(art) != loaded {
				t.Fatal("損壞材質未拒絕")
			}
			now := []*caveTerrainArt{art.ice, art.peligro, art.mpp3, art.mpp5, art.mpp6, art.mpp7, art.mpp8, art.mpp9, art.mppa, art.mppb}
			for i := range old {
				if old[i] != now[i] {
					t.Fatal("失敗載入破壞其他家族")
				}
			}
			grid := iceRead(t, filepath.Join(folder, c.maps[0]+"-grid.bin"))
			for name, change := range map[string]func(*CavePresentation){
				"last-source-byte": func(f *CavePresentation) { f.Memory[0x11b10+0x8030+c.count*48-1] ^= 1 },
				"body-resource":    func(f *CavePresentation) { f.BodyMemory[0x11b10+0x8030] ^= 1 },
				"shown-pixel":      func(f *CavePresentation) { f.Indexed[100*320+100] ^= 1 },
				"different-map":    func(f *CavePresentation) { f.BodyMemory[0x1b10+0xc000] ^= 1 },
			} {
				t.Run(name, func(t *testing.T) {
					f, _ := caveStructuralFrame(t, c.family, c.maps[0], grid)
					canvas := caveOriginalCanvas(f)
					before := cloneTownBytes(canvas.Pix)
					change(f)
					if art.Draw(canvas, f, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
						t.Fatal("失配未原子拒絕")
					}
				})
			}
			f, _ := caveStructuralFrame(t, c.family, c.maps[0], grid)
			canvas := caveOriginalCanvas(f)
			before := cloneTownBytes(canvas.Pix)
			refuse := func() {
				t.Helper()
				if art.Draw(canvas, f, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
					t.Fatal("缺圖或錯圖未原子拒絕")
				}
			}
			*c.field(art) = nil
			refuse()
			*c.field(art) = art.peligro
			refuse()
			*c.field(art) = loaded
			last := loaded.tiles[c.count]
			loaded.tiles[c.count] = nil
			refuse()
			loaded.tiles[c.count] = last
			if c.family == "MPPA" {
				secondary := loaded.secondaryMaterial
				loaded.secondaryMaterial = nil
				refuse()
				loaded.secondaryMaterial = secondary
			}
			grid[0] = byte(c.count + 1)
			f, _ = caveStructuralFrame(t, c.family, c.maps[0], grid)
			canvas = caveOriginalCanvas(f)
			before = cloneTownBytes(canvas.Pix)
			refuse()
		})
	}
}

package hd

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func iceStudy(t *testing.T) string {
	t.Helper()
	project := os.Getenv("ZELIARD_PROJECT")
	if project == "" {
		t.Skip("須掛載冰窟本地圖面；skip不算驗收")
	}
	return filepath.Join(project, "workplace/full-hd-v1")
}

func iceRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// 合成結構快照。由已驗第一洞窟取程式及角色，換入真實MPP4與靜態格；不執行VM。
// 原生地形於此組合，故只驗資料／繪圖，不作冰窟原版行為或正常可達證據。
func iceStructuralFrame(t *testing.T, name string, grid []byte) (*CavePresentation, []cavePixel) {
	t.Helper()
	root, points := caveFixtures(t)
	f := cavePresentedFixture(t, root, points[20])
	actor := append([]cavePixel(nil), f.Decoded.Actor...)
	b := iceStudy(t)
	header := iceRead(t, filepath.Join(b, "cave-catalog-fixtures-v1", name+".header"))
	resource := iceRead(t, filepath.Join(b, "cave-catalog-fixtures-v1/MPP4.resource"))
	if len(grid) != 504 {
		t.Fatal("網格尺寸不符")
	}
	for _, memory := range [][]byte{f.Memory, f.BodyMemory} {
		copy(memory[0x1b10+0xc000:], header)
		copy(memory[0x11b10+0x8000:], resource)
		p := memory[0x1b10:0x11b10]
		clear(p[0xe000:0xe900])
		binary.LittleEndian.PutUint16(p[0xff31:], 0xe000)
		p[0xff44] = 0
		for n, id := range grid {
			p[0xe000+n/28*36+4+n%28] = id
		}
	}
	f.Paints = nil
	r := f.Memory[0x11b10:0x21b10]
	for n, id := range grid {
		tile, err := caveStatic(r, id)
		if err != nil {
			t.Fatal(err)
		}
		for y := 0; y < 8; y++ {
			copy(f.SourceIndexed[(14+n/28*8+y)*320+48+n%28*8:], tile[y*8:y*8+8])
		}
	}
	for _, v := range actor {
		f.SourceIndexed[v.Y*320+v.X] = v.Color
	}
	f.Indexed = cloneTownBytes(f.SourceIndexed)
	for i, v := range f.Indexed {
		copy(f.RGB[i*3:], f.Palette[int(v)*3:int(v)*3+3])
	}
	return f, actor
}

func TestCaveIceIndependentStaticMaps(t *testing.T) {
	b := iceStudy(t)
	root, _ := caveFixtures(t)
	art, err := NewCaveTerrain(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = art.LoadIce(filepath.Join(b, "cave-ice-complete-v1")); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"MP40", "MP41", "MP4D"} {
		t.Run(name, func(t *testing.T) {
			grid := iceRead(t, filepath.Join(b, "cave-ice-complete-v1", name+"-grid.bin"))
			f, actor := iceStructuralFrame(t, name, grid)
			canvas := caveOriginalCanvas(f)
			before, memory, body := cloneTownBytes(canvas.Pix), cloneTownBytes(f.Memory), cloneTownBytes(f.BodyMemory)
			wantImage, e := png.Decode(bytes.NewReader(iceRead(t, filepath.Join(b, "cave-ice-complete-v1", name+"-offline-hd.png"))))
			if e != nil {
				t.Fatal(e)
			}
			want := image.NewRGBA(canvas.Bounds())
			copy(want.Pix, before)
			draw.Draw(want, image.Rect(144, 42, 816, 474), wantImage, image.Point{}, draw.Src)
			for _, v := range actor {
				c := color.RGBA{f.Palette[int(v.Color)*3], f.Palette[int(v.Color)*3+1], f.Palette[int(v.Color)*3+2], 255}
				draw.Draw(want, image.Rect(v.X*3, v.Y*3, v.X*3+3, v.Y*3+3), &image.Uniform{C: c}, image.Point{}, draw.Src)
			}
			if !art.Draw(canvas, f, image.Point{}) {
				t.Fatal("完整結構來源被拒絕")
			}
			if !bytes.Equal(canvas.Pix, want.Pix) {
				for y := 0; y < 600; y++ {
					for x := 0; x < 960; x++ {
						if canvas.RGBAAt(x, y) != want.RGBAAt(x, y) {
							t.Fatalf("獨立離線像素不符 %d,%d got%v want%v", x, y, canvas.RGBAAt(x, y), want.RGBAAt(x, y))
						}
					}
				}
			}
			if !bytes.Equal(memory, f.Memory) || !bytes.Equal(body, f.BodyMemory) {
				t.Fatal("呈現回寫來源")
			}
			if bytes.Equal(before, canvas.Pix) {
				t.Fatal("HD沒有可見變化")
			}
		})
	}
}

func TestCaveIceStructuralOrbs(t *testing.T) {
	b := iceStudy(t)
	root, _ := caveFixtures(t)
	art, err := NewCaveTerrain(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = art.LoadIce(filepath.Join(b, "cave-ice-complete-v1")); err != nil {
		t.Fatal(err)
	}
	grid := make([]byte, 504)
	grid[0], grid[1] = 20, 21
	f, _ := iceStructuralFrame(t, "MP40", grid)
	canvas := caveOriginalCanvas(f)
	if !art.Draw(canvas, f, image.Point{}) {
		t.Fatal("兩個結構光球來源被拒絕")
	}
	atlas, err := png.Decode(bytes.NewReader(iceRead(t, filepath.Join(b, "cave-ice-complete-v1/ice-tiles-v1.png"))))
	if err != nil {
		t.Fatal(err)
	}
	for n, id := range []int{20, 21} {
		changed := 0
		for y := 0; y < 24; y++ {
			for x := 0; x < 24; x++ {
				c := color.NRGBAModel.Convert(atlas.At((id-1)*24+x, y)).(color.NRGBA)
				want := color.RGBA{uint8((uint32(c.R)*uint32(c.A) + 127) / 255), uint8((uint32(c.G)*uint32(c.A) + 127) / 255), uint8((uint32(c.B)*uint32(c.A) + 127) / 255), 255}
				got := canvas.RGBAAt(144+n*24+x, 42+y)
				if got != want {
					t.Fatalf("圖格%d取樣不符 %d,%d", id, x, y)
				}
				if got.R != 0 || got.G != 0 || got.B != 0 {
					changed++
				}
			}
		}
		if changed == 0 {
			t.Fatal("光球沒有可見像素")
		}
	}
}

func TestCaveIceAtomicLoadAndRefusal(t *testing.T) {
	b := iceStudy(t)
	root, _ := caveFixtures(t)
	art, err := NewCaveTerrain(root)
	if err != nil {
		t.Fatal(err)
	}
	good := filepath.Join(b, "cave-ice-complete-v1")
	if err = art.LoadIce(good); err != nil {
		t.Fatal(err)
	}
	loaded := art.ice
	broken := t.TempDir()
	if err = os.WriteFile(filepath.Join(broken, "ice-tiles-v1.png"), iceRead(t, filepath.Join(good, "ice-tiles-v1.png")), 0644); err != nil {
		t.Fatal(err)
	}
	if err = art.LoadIce(broken); err == nil || art.ice != loaded {
		t.Fatal("缺材質破壞原子載入")
	}
	if err = os.WriteFile(filepath.Join(broken, "ice-material-v1.png"), []byte("broken"), 0644); err != nil {
		t.Fatal(err)
	}
	if err = art.LoadIce(broken); err == nil || art.ice != loaded {
		t.Fatal("損壞圖面被接受")
	}
	grid := iceRead(t, filepath.Join(good, "MP40-grid.bin"))
	cases := map[string]func(*CavePresentation){
		"different-map": func(f *CavePresentation) {
			copy(f.BodyMemory[0x1b10+0xc000:], iceRead(t, filepath.Join(b, "cave-catalog-fixtures-v1/MP41.header")))
		},
		"last-tile-byte": func(f *CavePresentation) { f.Memory[0x11b10+0x8030+21*48-1] ^= 1 },
		"body-resource":  func(f *CavePresentation) { f.BodyMemory[0x11b10+0x8030] ^= 1 },
		"original-pixel": func(f *CavePresentation) { f.SourceIndexed[100*320+100] ^= 1 },
		"shown-pixel":    func(f *CavePresentation) { f.Indexed[100*320+100] ^= 1 },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			f, _ := iceStructuralFrame(t, "MP40", grid)
			canvas := caveOriginalCanvas(f)
			before := cloneTownBytes(canvas.Pix)
			change(f)
			if art.Draw(canvas, f, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
				t.Fatal("來源拒絕前已寫入")
			}
		})
	}
	f, _ := iceStructuralFrame(t, "MP40", grid)
	canvas := caveOriginalCanvas(f)
	before := cloneTownBytes(canvas.Pix)
	art.ice = nil
	if art.Draw(canvas, f, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
		t.Fatal("未載圖仍交換地形")
	}
	art.ice = loaded
	grid[0] = 22
	f, _ = iceStructuralFrame(t, "MP40", grid)
	canvas = caveOriginalCanvas(f)
	before = cloneTownBytes(canvas.Pix)
	if art.Draw(canvas, f, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
		t.Fatal("未登錄圖格未原子拒絕")
	}
}

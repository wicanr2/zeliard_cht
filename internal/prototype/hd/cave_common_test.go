package hd

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// 固定正常F7來源，沒有注入；局部修改只用於下方拒絕測試。
func commonNormalFrame(t *testing.T, name string) *CavePresentation {
	t.Helper()
	base := filepath.Join(iceStudy(t), "cave-platform-source-pair-v2/cave-platform")
	var receipt struct {
		Points []struct {
			Name         string
			Presentation struct {
				SourceSteps  uint64      `json:"source_steps"`
				SourceCycles uint64      `json:"source_cycles"`
				BodySteps    uint64      `json:"body_steps"`
				BodyCycles   uint64      `json:"body_cycles"`
				MemorySHA    string      `json:"source_memory_sha256"`
				BodySHA      string      `json:"body_memory_sha256"`
				IndexedSHA   string      `json:"source_indexed_sha256"`
				Paints       []CavePaint `json:"native_paints"`
			}
		}
	}
	if err := json.Unmarshal(iceRead(t, filepath.Join(base, "receipt.json")), &receipt); err != nil {
		t.Fatal(err)
	}
	for _, point := range receipt.Points {
		if point.Name != name {
			continue
		}
		read := func(suffix string) []byte { return iceRead(t, filepath.Join(base, name+".presented."+suffix)) }
		p := point.Presentation
		f := &CavePresentation{Memory: read("memory"), BodyMemory: read("body.memory"), Indexed: read("indexed"), RGB: read("rgb"), Palette: read("palette"), Steps: p.SourceSteps, Cycles: p.SourceCycles, BodySteps: p.BodySteps, BodyCycles: p.BodyCycles, Paints: p.Paints}
		f.SourceIndexed = cloneTownBytes(f.Indexed)
		if townBackgroundHash(f.Memory) != p.MemorySHA || townBackgroundHash(f.BodyMemory) != p.BodySHA || townBackgroundHash(f.SourceIndexed) != p.IndexedSHA {
			t.Fatal("正常來源雜湊不符", name)
		}
		var err error
		f.Decoded, err = decodeCaveFrameWithPaints(f.Memory, f.BodyMemory, f.SourceIndexed, f.Paints)
		if err != nil {
			t.Fatal(err)
		}
		return f
	}
	t.Fatal("缺少正常平台點", name)
	return nil
}

func commonTestArt(t *testing.T) *CaveTerrain {
	t.Helper()
	art := mpp1Art(t)
	if err := art.LoadCommon(filepath.Join(iceStudy(t), "cave-common-registration-v2")); err != nil {
		t.Fatal(err)
	}
	return art
}

func TestCaveCommonNormalIndependentPixels(t *testing.T) {
	art := commonTestArt(t)
	atlas, err := png.Decode(bytes.NewReader(iceRead(t, filepath.Join(iceStudy(t), "cave-common-registration-v2/green-platform-72x24.png"))))
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, name := range []string{"platform-board", "platform-raised", "platform-lowered", "platform-reraised", "platform-leave-edge"} {
		t.Run(name, func(t *testing.T) {
			f := commonNormalFrame(t, name)
			before, body := cloneTownBytes(f.Memory), cloneTownBytes(f.BodyMemory)
			canvas := caveOriginalCanvas(f)
			if !art.Draw(canvas, f, image.Point{}) {
				t.Fatal("正常平台HD被拒絕")
			}
			if !bytes.Equal(before, f.Memory) || !bytes.Equal(body, f.BodyMemory) {
				t.Fatal("HD回寫來源")
			}
			state := art.CommonState()
			positions, ok := state["placements"].([]image.Point)
			if state["matched"] != true || !ok || len(positions) != 1 {
				t.Fatal("平台位置未辨識", state)
			}
			// 獨立從解碼圖格與圖面取像素；正常動態、主角及武器在平台之後繪製。
			covered := map[image.Point]bool{}
			for _, layer := range [][]cavePixel{f.Decoded.Dynamic, f.Decoded.Actor, f.Decoded.Weapon} {
				for _, p := range layer {
					covered[image.Pt(p.X, p.Y)] = true
				}
			}
			checked := 0
			for _, cell := range f.Decoded.Cells {
				if cell.ID < 64 || cell.ID > 66 {
					continue
				}
				raw, e := caveStatic(f.Memory[0x11b10:0x21b10], cell.ID)
				if e != nil {
					t.Fatal(e)
				}
				for y := 0; y < 24; y++ {
					for x := 0; x < 24; x++ {
						if raw[y/3*8+x/3] == 0 || covered[image.Pt(cell.X+x/3, cell.Y+y/3)] {
							continue
						}
						c := color.NRGBAModel.Convert(atlas.At(int(cell.ID-64)*24+x, y)).(color.NRGBA)
						if c.A != 255 {
							t.Fatal("平台有效圖面透明")
						}
						want := color.RGBA{c.R, c.G, c.B, 255}
						if got := canvas.RGBAAt(cell.X*3+x, cell.Y*3+y); got != want {
							t.Fatalf("平台像素不符 %d,%d got%v want%v", x, y, got, want)
						}
						checked++
					}
				}
			}
			if checked < 1000 {
				t.Fatal("平台有效抽樣不足", checked)
			}
			total += checked
			// 診斷副本不得讓呼叫端改掉內部位置。
			positions[0].X = -999
			if art.CommonState()["placements"].([]image.Point)[0].X < 0 {
				t.Fatal("診斷洩漏可寫位置")
			}
		})
	}
	t.Logf("五個正常平台點獨立像素合計 %d", total)
}

func TestCaveCommonAtomicSourceAndArtRefusal(t *testing.T) {
	art := commonTestArt(t)
	loaded := art.common
	if err := art.LoadCommon(t.TempDir()); err == nil || art.common != loaded {
		t.Fatal("缺圖破壞原子載入")
	}
	broken := t.TempDir()
	if err := os.WriteFile(filepath.Join(broken, "green-platform-72x24.png"), []byte("broken"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := art.LoadCommon(broken); err == nil || art.common != loaded {
		t.Fatal("壞圖破壞原子載入")
	}
	for name, change := range map[string]func(*CavePresentation){
		"last-common-byte":    func(f *CavePresentation) { f.Memory[0x11b10+0x934f] ^= 1 },
		"body-common-byte":    func(f *CavePresentation) { f.BodyMemory[0x11b10+0x934f] ^= 1 },
		"platform-code":       func(f *CavePresentation) { f.Memory[0x1b10+0x7fdc] ^= 1 },
		"body-platform-code":  func(f *CavePresentation) { f.BodyMemory[0x1b10+0x8162] ^= 1 },
		"table-position":      func(f *CavePresentation) { p := f.Memory[0x1b10:0x11b10]; p[caveWord(p, 0xc004)+2] ^= 1 },
		"body-table-position": func(f *CavePresentation) { p := f.BodyMemory[0x1b10:0x11b10]; p[caveWord(p, 0xc004)+2] ^= 1 },
		"table-sentinel":      func(f *CavePresentation) { p := f.Memory[0x1b10:0x11b10]; p[caveWord(p, 0xc006)-1] = 0 },
		"table-stride":        func(f *CavePresentation) { f.Memory[0x1b10+0xc006]++ },
		"missing-triplet": func(f *CavePresentation) {
			for _, c := range f.Decoded.Cells {
				if c.ID == 65 {
					f.Memory[0x1b10+c.Address] = 0
					return
				}
			}
			t.Fatal("缺少平台中格")
		},
	} {
		t.Run(name, func(t *testing.T) {
			f := commonNormalFrame(t, "platform-raised")
			canvas := caveOriginalCanvas(f)
			before := cloneTownBytes(canvas.Pix)
			change(f)
			if _, ok := caveCommonPlacements(f.Memory, f.BodyMemory, f.Decoded.Cells); ok {
				t.Fatal("來源守衛接受失配")
			}
			if art.Draw(canvas, f, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
				t.Fatal("來源失配未原子退回")
			}
		})
	}
	for _, mode := range []string{"missing-art", "missing-tile"} {
		t.Run(mode, func(t *testing.T) {
			f := commonNormalFrame(t, "platform-raised")
			canvas := caveOriginalCanvas(f)
			before := cloneTownBytes(canvas.Pix)
			art.common = nil
			if mode == "missing-tile" {
				copyArt := *loaded
				copyArt.tiles[2] = nil
				art.common = &copyArt
			}
			if art.Draw(canvas, f, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
				t.Fatal("缺平台素材仍繪製")
			}
			art.common = loaded
		})
	}
}

// 結構邊界測試：只改鏡頭與快取起點，不當正常捲動到達證據。
func TestCaveCommonClippedTripletGuard(t *testing.T) {
	for _, edge := range []struct {
		name                  string
		camera, delta, anchor int
		cells                 []caveCell
	}{
		{"left", 45, 13, 40, []caveCell{{X: 48, Y: 118, ID: 65}, {X: 56, Y: 118, ID: 66}}},
		{"right", 17, -15, 264, []caveCell{{X: 264, Y: 118, ID: 64}}},
	} {
		t.Run(edge.name, func(t *testing.T) {
			f := commonNormalFrame(t, "platform-raised")
			for _, memory := range [][]byte{f.Memory, f.BodyMemory} {
				p := memory[0x1b10:0x11b10]
				binary.LittleEndian.PutUint16(p[0x80:], uint16(edge.camera))
				ring := (caveWord(p, 0xff31) - 0xe000 + edge.delta + 0x900) % 0x900
				binary.LittleEndian.PutUint16(p[0xff31:], uint16(0xe000+ring))
			}
			positions, ok := caveCommonPlacements(f.Memory, f.BodyMemory, edge.cells)
			if !ok || len(positions) != 1 || positions[0] != image.Pt(edge.anchor, 118) {
				t.Fatal("邊界完整快取未接受", positions, ok)
			}
			// 破壞畫面外的另一格；可見格仍在時，也必須整組拒絕。
			id := byte(64)
			if edge.name == "right" {
				id = 66
			}
			for _, cell := range f.Decoded.Cells {
				if cell.ID == id {
					f.Memory[0x1b10+cell.Address] = 0
					break
				}
			}
			if _, ok := caveCommonPlacements(f.Memory, f.BodyMemory, edge.cells); ok {
				t.Fatal("畫面外格失配仍放行")
			}
		})
	}
}

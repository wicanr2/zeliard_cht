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

func magiaFrames(t *testing.T) []*CavePresentation {
	t.Helper()
	base := filepath.Join(iceStudy(t), "cave-magia-source-v3")
	var receipt struct {
		Points []struct {
			File          string
			Steps, Cycles uint64
			BodySteps     uint64 `json:"body_steps"`
			BodyCycles    uint64 `json:"body_cycles"`
			MemorySHA     string `json:"memory_sha256"`
			BodySHA       string `json:"body_memory_sha256"`
			Paints        []CavePaint
		}
	}
	if err := json.Unmarshal(iceRead(t, filepath.Join(base, "receipt.json")), &receipt); err != nil {
		t.Fatal(err)
	}
	if len(receipt.Points) != 27 {
		t.Fatal("來源數不符")
	}
	var frames []*CavePresentation
	for _, p := range receipt.Points {
		read := func(suffix string) []byte { return iceRead(t, filepath.Join(base, p.File+"."+suffix)) }
		f := &CavePresentation{Memory: read("memory"), BodyMemory: read("body.memory"), SourceIndexed: read("indexed"), Palette: read("palette"), Steps: p.Steps, Cycles: p.Cycles, BodySteps: p.BodySteps, BodyCycles: p.BodyCycles, Paints: p.Paints}
		if townBackgroundHash(f.Memory) != p.MemorySHA || townBackgroundHash(f.BodyMemory) != p.BodySHA {
			t.Fatal("來源雜湊不符")
		}
		magiaFrameRGB(f)
		var err error
		f.Decoded, err = decodeCaveFrameWithPaints(f.Memory, f.BodyMemory, f.SourceIndexed, f.Paints)
		if err != nil {
			t.Fatal("已存正常來源拒絕", p.File, err)
		}
		frames = append(frames, f)
	}
	return frames
}

func magiaFrameRGB(f *CavePresentation) {
	f.Indexed = cloneTownBytes(f.SourceIndexed)
	f.RGB = make([]byte, 192000)
	for i, v := range f.Indexed {
		copy(f.RGB[i*3:], f.Palette[int(v)*3:int(v)*3+3])
	}
}

func magiaArt(t *testing.T) *CaveTerrain {
	t.Helper()
	art := commonTestArt(t)
	if err := art.LoadMagia(filepath.Join(iceStudy(t), "cave-magia-complete-v1")); err != nil {
		t.Fatal(err)
	}
	return art
}

func TestCaveMagiaNormalSourcesAndIndependentPixels(t *testing.T) {
	frames, art := magiaFrames(t), magiaArt(t)
	atlas, err := png.Decode(bytes.NewReader(iceRead(t, filepath.Join(iceStudy(t), "cave-magia-complete-v1/magia-shard-24x24.png"))))
	if err != nil {
		t.Fatal(err)
	}
	checked, cleared, duplicates := 0, 0, 0
	for n, f := range frames {
		memory, body := cloneTownBytes(f.Memory), cloneTownBytes(f.BodyMemory)
		canvas := caveOriginalCanvas(f)
		if !art.Draw(canvas, f, image.Point{}) {
			t.Fatal("魔法石正常來源HD拒絕", n)
		}
		if !bytes.Equal(memory, f.Memory) || !bytes.Equal(body, f.BodyMemory) {
			t.Fatal("覆繪回寫來源")
		}
		seen := map[image.Point]bool{}
		mask := map[image.Point]bool{}
		for _, r := range f.Decoded.MagiaRecords {
			if r.Visibility == "cleared" {
				cleared++
				continue
			}
			pos := image.Pt(r.X, r.Y)
			if seen[pos] {
				duplicates++
				continue
			}
			seen[pos] = true
			for y := 0; y < 24; y++ {
				for x := 0; x < 24; x++ {
					c := color.NRGBAModel.Convert(atlas.At(x, y)).(color.NRGBA)
					if c.A == 0 {
						continue
					}
					if c.A != 255 {
						t.Fatal("有效圖面透明")
					}
					pixel := image.Pt(r.X*3+x, r.Y*3+y)
					mask[pixel] = true
					if got := canvas.RGBAAt(pixel.X, pixel.Y); got != (color.RGBA{c.R, c.G, c.B, 255}) {
						t.Fatal("獨立高清像素不符", n, pixel, got, c)
					}
					checked++
				}
			}
		}
		// 結構對照：只從來源色號移去已解圖層，檢查HD寫入沒有越出102包絡。
		base := *f
		base.SourceIndexed = cloneTownBytes(f.Decoded.Background)
		for _, layer := range [][]cavePixel{f.Decoded.Actor, f.Decoded.Weapon} {
			for _, v := range layer {
				base.SourceIndexed[v.Y*320+v.X] = v.Color
			}
		}
		magiaFrameRGB(&base)
		without := caveOriginalCanvas(&base)
		if !art.Draw(without, &base, image.Point{}) {
			t.Fatal("結構清除來源拒絕", n)
		}
		for y := 0; y < 600; y++ {
			for x := 0; x < 960; x++ {
				if !mask[image.Pt(x, y)] && canvas.RGBAAt(x, y) != without.RGBAAt(x, y) {
					t.Fatal("HD越出102包絡", n, x, y)
				}
			}
		}
	}
	if cleared < 4 || duplicates == 0 || checked < 20000 {
		t.Fatal("正常分支抽樣不足", cleared, duplicates, checked)
	}
	t.Logf("27張原版來源；獨立HD像素%d，已清除槽%d，重疊槽%d", checked, cleared, duplicates)
}

func TestCaveMagiaAtomicRefusal(t *testing.T) {
	art := magiaArt(t)
	saved := art.magia
	if err := art.LoadMagia(t.TempDir()); err == nil || &art.magia[0] != &saved[0] {
		t.Fatal("缺圖載入非原子")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "magia-shard-24x24.png"), []byte("broken"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := art.LoadMagia(dir); err == nil || &art.magia[0] != &saved[0] {
		t.Fatal("壞圖載入非原子")
	}
	for name, change := range map[string]func(*CavePresentation){
		"resource-last":   func(f *CavePresentation) { f.Memory[0x11b10+0x934f] ^= 1 },
		"body-resource":   func(f *CavePresentation) { f.BodyMemory[0x11b10+0x934f] ^= 1 },
		"phase-code":      func(f *CavePresentation) { f.Memory[0x1b10+0x878f] ^= 1 },
		"body-phase-code": func(f *CavePresentation) { f.BodyMemory[0x1b10+0x8679] ^= 1 },
		"position-table":  func(f *CavePresentation) { f.Memory[0x1b10+0x8790] ^= 1 },
		"draw-pointer":    func(f *CavePresentation) { f.Memory[0x1b10+0x3008] ^= 1 },
		"phase":           func(f *CavePresentation) { f.Memory[0x1b10+0xeb60] = 16 },
		"direction":       func(f *CavePresentation) { f.Memory[0x1b10+0xeb61] = 0 },
		"life":            func(f *CavePresentation) { f.Memory[0x1b10+0xeb62] = 0 },
		"cache-position":  func(f *CavePresentation) { f.Memory[0x1b10+0xeb65]++ },
		"partial-pixel": func(f *CavePresentation) {
			for _, v := range f.Decoded.Magia {
				f.SourceIndexed[v.Y*320+v.X] ^= 1
				magiaFrameRGB(f)
				break
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			f := magiaFrames(t)[0]
			canvas := caveOriginalCanvas(f)
			before := cloneTownBytes(canvas.Pix)
			change(f)
			if art.Draw(canvas, f, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
				t.Fatal("來源失配未原子拒絕")
			}
		})
	}
	f := magiaFrames(t)[0]
	canvas := caveOriginalCanvas(f)
	before := cloneTownBytes(canvas.Pix)
	art.magia = nil
	if art.Draw(canvas, f, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
		t.Fatal("缺圖仍接受可見魔法石")
	}
	art.magia = saved
	if !art.Draw(canvas, f, image.Point{}) {
		t.Fatal("恢復圖面未恢復呈現")
	}
	state := art.MagiaState()
	records := state["records"].([]caveMagiaRecord)
	records[0].X = -100
	if art.MagiaState()["records"].([]caveMagiaRecord)[0].X < 0 {
		t.Fatal("診斷洩漏可寫來源")
	}
}

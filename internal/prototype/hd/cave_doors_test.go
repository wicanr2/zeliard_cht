package hd

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// 已存GUI日誌逐筆重播所得原始來源；不改其輸入或原版狀態。
func doorNormalFrame(t *testing.T) *CavePresentation {
	t.Helper()
	base := filepath.Join(iceStudy(t), "cave-station-source-v1")
	read := func(suffix string) []byte { return iceRead(t, filepath.Join(base, "source."+suffix)) }
	f := &CavePresentation{Memory: read("memory"), BodyMemory: read("body.memory"), SourceIndexed: read("indexed"), Palette: read("palette"), Steps: 230314104, Cycles: 236058109, BodySteps: 230206020, BodyCycles: 235949157}
	if townBackgroundHash(f.Memory) != "c38eb353662ebaaf176dfbb3e9a29bb0d094132715e307696575528edec04973" || townBackgroundHash(f.BodyMemory) != "97ae502c176007c961a41798aa44cd4d115d46a3ea324bccc83ab92e1e459998" || townBackgroundHash(f.SourceIndexed) != "f6add75605674388536c86e820bbbf2884f0bf20c566755cf38bb3306d03fa1e" {
		t.Fatal("正常GUI重播來源雜湊不符")
	}
	if err := json.Unmarshal(read("paints.json"), &f.Paints); err != nil {
		t.Fatal(err)
	}
	magiaFrameRGB(f)
	var err error
	f.Decoded, err = decodeCaveFrameWithPaints(f.Memory, f.BodyMemory, f.SourceIndexed, f.Paints)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func doorTestArt(t *testing.T) *CaveTerrain {
	t.Helper()
	art := magiaArt(t)
	if err := art.LoadDoors(filepath.Join(iceStudy(t), "cave-door-complete-v1")); err != nil {
		t.Fatal(err)
	}
	return art
}

func TestCaveDoorsNormalStationAndIndependentPixels(t *testing.T) {
	f, art := doorNormalFrame(t), doorTestArt(t)
	memory, body := cloneTownBytes(f.Memory), cloneTownBytes(f.BodyMemory)
	original, canvas := caveOriginalCanvas(f), caveOriginalCanvas(f)
	if !art.Draw(canvas, f, image.Point{}) {
		t.Fatal("已重生的正常門框站位仍退回原貌")
	}
	if !bytes.Equal(f.Memory, memory) || !bytes.Equal(f.BodyMemory, body) {
		t.Fatal("HD回寫原始來源")
	}
	state := art.DoorState()
	places, ok := state["placements"].([]caveDoorPlacement)
	if !ok || len(places) != 1 || places[0].Position != image.Pt(16, 38) || places[0].Marker != 98 || !places[0].Open || state["matched"] != true {
		t.Fatal("正常門框辨識錯誤", state)
	}
	atlas, err := png.Decode(bytes.NewReader(iceRead(t, filepath.Join(iceStudy(t), "cave-door-complete-v1/door-98-open-120x96.png"))))
	if err != nil {
		t.Fatal(err)
	}
	covered := map[image.Point]bool{}
	for _, layer := range [][]cavePixel{f.Decoded.Dynamic, f.Decoded.Actor, f.Decoded.Magia, f.Decoded.Weapon} {
		for _, p := range layer {
			covered[image.Pt(p.X, p.Y)] = true
		}
	}
	checked, changed := 0, 0
	for y := 0; y < 96; y++ {
		for x := 0; x < 120; x++ {
			px, py := 16*3+x, 38*3+y
			if px < 144 || px >= 816 || py < 42 || py >= 474 || covered[image.Pt(px/3, py/3)] {
				continue
			}
			c := color.NRGBAModel.Convert(atlas.At(x, y)).(color.NRGBA)
			want := color.RGBA{uint8((uint32(c.R)*uint32(c.A) + 127) / 255), uint8((uint32(c.G)*uint32(c.A) + 127) / 255), uint8((uint32(c.B)*uint32(c.A) + 127) / 255), 255}
			if got := canvas.RGBAAt(px, py); got != want {
				t.Fatal("正常門框獨立像素不符", px, py, got, want)
			}
			checked++
			if original.RGBAAt(px, py) != want {
				changed++
			}
		}
	}
	if checked < 2000 || changed < 500 {
		t.Fatal("正常圖面抽樣不足", checked, changed)
	}
	places[0].Position.X = -999
	if art.DoorState()["placements"].([]caveDoorPlacement)[0].Position.X != 16 {
		t.Fatal("診斷洩漏可寫狀態")
	}
	if out := os.Getenv("ZELIARD_CAVE_DOOR_OUTPUT"); out != "" {
		for name, im := range map[string]image.Image{"station-original": original, "station-hd": canvas} {
			file, e := os.Create(filepath.Join(out, name+".png"))
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(file, im)
			file.Close()
			if e != nil {
				t.Fatal(e)
			}
		}
	}
	t.Logf("正常98/open站位：獨立像素%d，改變%d，完整／主體來源未寫入", checked, changed)
}

// 局部改模板及記錄只驗結構；不執行VM，不當其餘九種門框正常到達證據。
func TestCaveDoorsTenStructuralVariantsAndClipping(t *testing.T) {
	art := doorTestArt(t)
	var sources struct {
		Variants []struct {
			Marker byte
			Kind   string
			Matrix [][]byte
		}
	}
	if err := json.Unmarshal(iceRead(t, filepath.Join(iceStudy(t), "cave-door-source-v2/receipt.json")), &sources); err != nil {
		t.Fatal(err)
	}
	if len(sources.Variants) != 10 {
		t.Fatal("缺原版圖面矩陣")
	}
	total := 0
	for _, v := range sources.Variants {
		f := doorNormalFrame(t)
		p := f.Memory[0x1b10:0x11b10]
		p[caveWord(p, 0xc00a)+3] = v.Marker - 97
		if v.Kind == "open" {
			p[caveWord(p, 0xc00a)+3] |= 128
		}
		for y, row := range v.Matrix {
			for x, value := range row {
				at := (caveWord(p, 0xff31) - 0xe000 + (3+y)*36 + x) % 0x900
				p[0xe000+at] = value
			}
		}
		f.BodyMemory = cloneTownBytes(f.Memory)
		places, ok := caveDoorPlacements(f.Memory, f.BodyMemory, nil)
		if !ok || len(places) != 1 || places[0].Marker != v.Marker || places[0].Open != (v.Kind == "open") {
			t.Fatal("結構變體辨識失敗", v.Marker, v.Kind, places, ok)
		}
		file := filepath.Join(iceStudy(t), "cave-door-complete-v1", "door-"+strconv.Itoa(int(v.Marker))+"-"+v.Kind+"-120x96.png")
		im, err := png.Decode(bytes.NewReader(iceRead(t, file)))
		if err != nil {
			t.Fatal(err)
		}
		for _, position := range []image.Point{{100, 80}, {16, 38}, {264, 150}} {
			places[0].Position = position
			canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
			art.doors.draw(canvas, places, image.Point{})
			box := image.Rect(position.X*3, position.Y*3, position.X*3+120, position.Y*3+96).Intersect(image.Rect(144, 42, 816, 474))
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					want := color.RGBA{}
					if image.Pt(x, y).In(box) {
						c := color.NRGBAModel.Convert(im.At(x-position.X*3, y-position.Y*3)).(color.NRGBA)
						want = color.RGBA{uint8((uint32(c.R)*uint32(c.A) + 127) / 255), uint8((uint32(c.G)*uint32(c.A) + 127) / 255), uint8((uint32(c.B)*uint32(c.A) + 127) / 255), 255}
						total++
					}
					if canvas.RGBAAt(x, y) != want {
						t.Fatal("結構圖面或裁切越界", v.Marker, v.Kind, position, x, y)
					}
				}
			}
		}
	}
	t.Logf("十變體三位置的獨立結構像素%d；框外全未寫入", total)
}

func TestCaveDoorsAtomicSourceAndArtRefusal(t *testing.T) {
	art := doorTestArt(t)
	loaded := art.doors
	if err := art.LoadDoors(t.TempDir()); err == nil || art.doors != loaded {
		t.Fatal("缺圖非原子載入")
	}
	broken := t.TempDir()
	files, err := os.ReadDir(filepath.Join(iceStudy(t), "cave-door-complete-v1"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".png" {
			continue
		}
		data := iceRead(t, filepath.Join(iceStudy(t), "cave-door-complete-v1", file.Name()))
		if file.Name() == "door-101-open-120x96.png" {
			data = []byte("broken final group member")
		}
		if err := os.WriteFile(filepath.Join(broken, file.Name()), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := art.LoadDoors(broken); err == nil || art.doors != loaded {
		t.Fatal("最後一張損壞仍交換整組圖面")
	}
	for name, change := range map[string]func(*CavePresentation){
		"code":            func(f *CavePresentation) { f.Memory[0x1b10+0x78dd] ^= 1 },
		"body-code":       func(f *CavePresentation) { f.BodyMemory[0x1b10+0x79b3] ^= 1 },
		"template":        func(f *CavePresentation) { f.Memory[0x1b10+0x79b4] ^= 1 },
		"template-marker": func(f *CavePresentation) { f.BodyMemory[0x1b10+0x79ca] = 102 },
		"resource":        func(f *CavePresentation) { f.Memory[0x11b10+0x8c00] ^= 1 },
		"body-resource":   func(f *CavePresentation) { f.BodyMemory[0x11b10+0x934f] ^= 1 },
		"record":          func(f *CavePresentation) { p := f.Memory[0x1b10:0x11b10]; p[caveWord(p, 0xc00a)]++ },
		"flags":           func(f *CavePresentation) { p := f.Memory[0x1b10:0x11b10]; p[caveWord(p, 0xc00a)+3] ^= 128 },
		"sentinel":        func(f *CavePresentation) { p := f.Memory[0x1b10:0x11b10]; p[caveWord(p, 0xc00c)-1] = 0 },
		"stride":          func(f *CavePresentation) { f.Memory[0x1b10+0xc00c]++ },
		"offscreen-cache": func(f *CavePresentation) { p := f.Memory[0x1b10:0x11b10]; p[caveWord(p, 0xff31)+3*36] = 0 },
	} {
		t.Run(name, func(t *testing.T) {
			f := doorNormalFrame(t)
			canvas := caveOriginalCanvas(f)
			before := cloneTownBytes(canvas.Pix)
			change(f)
			if art.Draw(canvas, f, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
				t.Fatal("來源失配未原子退回")
			}
		})
	}
	f := doorNormalFrame(t)
	canvas := caveOriginalCanvas(f)
	before := cloneTownBytes(canvas.Pix)
	art.doors = nil
	if art.Draw(canvas, f, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
		t.Fatal("缺圖仍接受新門框")
	}
	art.doors = loaded
	if !art.Draw(canvas, f, image.Point{}) {
		t.Fatal("恢復圖面未恢復呈現")
	}
}

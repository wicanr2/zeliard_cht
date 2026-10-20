package hd

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"testing"
)

func townActorsTestInputs(t *testing.T) (*TownActors, string) {
	t.Helper()
	scene, base := sceneryTestInputs(t)
	player, err := NewTownPlayer(scene, base)
	if err != nil {
		t.Fatal(err)
	}
	dir := os.Getenv("ZELIARD_TOWN_NPC_ART")
	if dir == "" {
		t.Skip("須掛載本地居民圖集；skip不算驗收")
	}
	actors, err := NewTownActors(player, dir)
	if err != nil {
		t.Fatal(err)
	}
	return actors, base
}

func townActorsFixture(t *testing.T, base, source string) ([]byte, []byte) {
	t.Helper()
	read := func(suffix string) []byte {
		b, err := os.ReadFile(filepath.Join(base, source+suffix))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	return read(".memory"), read(".indexed")
}

func TestTownActorsAllSources(t *testing.T) {
	s, base := townActorsTestInputs(t)
	memory, _ := townActorsFixture(t, base, "normal-v1/restored")
	r, e := memory[0x11b10:0x21b10], memory[0x21b10:0x31b10]
	seen := map[string][]byte{}
	for _, asset := range townNPCAssets {
		raw, err := townNPCRaster(r, e, s.player.scene.palette, asset.group, asset.phase)
		if err != nil || townBackgroundHash(raw) != asset.source {
			t.Fatalf("原版居民g%d-p%d與獨立來源不符：%v", asset.group, asset.phase, err)
		}
		art := s.art[asset.group][asset.phase]
		if prior, ok := seen[asset.source]; ok && !bytes.Equal(prior, art.Pix) {
			t.Fatal("同一原版姿態產生不同動畫")
		}
		seen[asset.source] = art.Pix
		for y := 0; y < 72; y++ {
			for x := 0; x < 48; x++ {
				if !image.Pt(x/3, y/3).In(asset.target) && art.RGBAAt(x, y).A != 0 {
					t.Fatal("居民HD超出原版姿態外框")
				}
			}
		}
	}
	if len(seen) != 34 {
		t.Fatal("獨立原圖分母改變", len(seen))
	}
}

func TestTownActorsAtomicFallback(t *testing.T) {
	s, base := townActorsTestInputs(t)
	memory, raw := townActorsFixture(t, base, "normal-v1/restored")
	for _, name := range []string{"accepted", "nil", "nil-canvas", "small-canvas", "wrong-mode", "short-primary", "short-resources", "short-extra", "short-indexed", "short-rgb", "short-palette", "wrong-palette", "wrong-code", "wrong-pixel", "unknown-phase", "missing-pose", "wrong-source-key"} {
		t.Run(name, func(t *testing.T) {
			p, r, e := cloneTownBytes(memory[0x1b10:0x11b10]), cloneTownBytes(memory[0x11b10:0x21b10]), cloneTownBytes(memory[0x21b10:0x31b10])
			indices, palette := cloneTownBytes(raw), cloneTownBytes(s.player.scene.palette)
			rgb, canvas := skyOriginal(indices, palette)
			candidate := *s
			renderer := &candidate
			mode := uint8(0x13)
			switch name {
			case "nil":
				renderer = nil
			case "nil-canvas":
				canvas = nil
			case "small-canvas":
				canvas = image.NewRGBA(image.Rect(0, 0, 100, 100))
			case "wrong-mode":
				mode = 3
			case "short-primary":
				p = p[:100]
			case "short-resources":
				r = r[:100]
			case "short-extra":
				e = e[:100]
			case "short-indexed":
				indices = indices[:100]
			case "short-rgb":
				rgb = rgb[:100]
			case "short-palette":
				palette = palette[:100]
			case "wrong-palette":
				palette[0] ^= 1
			case "wrong-code":
				p[0x6975] ^= 1
			case "wrong-pixel":
				indices[118*320+48] ^= 1
			case "unknown-phase":
				p[0xe7] = 4
			case "missing-pose":
				for g := range candidate.art {
					for phase := range candidate.art[g] {
						candidate.art[g][phase] = nil
					}
				}
			case "wrong-source-key":
				for g := range candidate.source {
					for phase := range candidate.source[g] {
						candidate.source[g][phase] = ""
					}
				}
			}
			var before []byte
			if canvas != nil {
				before = cloneTownBytes(canvas.Pix)
			}
			inputs := [][]byte{p, r, e, indices, rgb, palette}
			copies := make([][]byte, len(inputs))
			for i := range inputs {
				copies[i] = cloneTownBytes(inputs[i])
			}
			got := renderer.Draw(canvas, mode, p, r, e, indices, rgb, palette, image.Point{})
			if got != (name == "accepted") {
				t.Fatalf("案例%s匹配%v", name, got)
			}
			if !got && canvas != nil && !bytes.Equal(before, canvas.Pix) {
				t.Fatal("拒絕後改寫畫布")
			}
			for i := range inputs {
				if !bytes.Equal(inputs[i], copies[i]) {
					t.Fatal("改寫原版來源", i)
				}
			}
			if got {
				actors, err := townSceneSources(p, r, e)
				if err != nil {
					t.Fatal(err)
				}
				boxes := []image.Rectangle{image.Rect((48+int(p[0x83])*8)*3, 354, (48+int(p[0x83])*8)*3+48, 426)}
				for _, a := range actors {
					x := (a.x - townSceneWord(p, 0x80) - 4) * 8
					boxes = append(boxes, image.Rect((48+x)*3, 354, (48+x)*3+48, 426))
				}
				changes := 0
				for y := 0; y < 600; y++ {
					for x := 0; x < 960; x++ {
						at := canvas.PixOffset(x, y)
						if bytes.Equal(before[at:at+4], canvas.Pix[at:at+4]) {
							continue
						}
						changes++
						inside := false
						for _, box := range boxes {
							inside = inside || image.Pt(x, y).In(box)
						}
						if !inside || !image.Pt(x, y).In(image.Rect(144, 354, 816, 426)) {
							t.Fatal("角色聯集外寫入", x, y)
						}
					}
				}
				if changes == 0 {
					t.Fatal("合成未改變任何人物像素")
				}
				state := renderer.State()
				keys := state["keys"].([]string)
				keys[0] = "changed"
				if renderer.State()["keys"].([]string)[0] == "changed" {
					t.Fatal("診斷未深複製")
				}
				renderer.Clear()
				if renderer.State()["matched"] != false {
					t.Fatal("撤除後仍顯示匹配")
				}
			}
		})
	}
}

// 合成順序用已存在的正常重疊來源與合成色塊驗證，不把色塊當美術／原版證據。
func TestTownActorsLayerOrder(t *testing.T) {
	s, base := townActorsTestInputs(t)
	data, err := os.ReadFile(filepath.Join(base, "scene-prototype-v1/coherent-frame-v1/player-preview-v4.json"))
	if err != nil {
		t.Fatal(err)
	}
	var proof struct {
		Records []struct {
			Source  string
			Overlap int `json:"NPC_overlap"`
		}
	}
	if err = json.Unmarshal(data, &proof); err != nil {
		t.Fatal(err)
	}
	var source string
	for _, record := range proof.Records {
		if record.Overlap > 0 {
			source = record.Source
			break
		}
	}
	if source == "" {
		t.Fatal("既有正常資料缺少居民重疊來源")
	}
	memory, raw := townActorsFixture(t, base, source)
	p, r, e := memory[0x1b10:0x11b10], memory[0x11b10:0x21b10], memory[0x21b10:0x31b10]
	rgb, canvas := skyOriginal(raw, s.player.scene.palette)
	candidate := *s
	player := *s.player
	candidate.player = &player
	red := image.NewRGBA(image.Rect(0, 0, 48, 72))
	draw.Draw(red, red.Bounds(), image.NewUniform(color.RGBA{255, 0, 0, 255}), image.Point{}, draw.Src)
	blue := image.NewRGBA(red.Bounds())
	draw.Draw(blue, blue.Bounds(), image.NewUniform(color.RGBA{0, 0, 128, 128}), image.Point{}, draw.Src)
	for g := range candidate.art {
		for phase := range candidate.art[g] {
			candidate.art[g][phase] = red
		}
	}
	for phase := range player.art {
		player.art[phase] = blue
	}
	if !candidate.Draw(canvas, 0x13, p, r, e, raw, rgb, s.player.scene.palette, image.Point{}) {
		t.Fatal("正常重疊來源被拒絕")
	}
	actors, err := townSceneSources(p, r, e)
	if err != nil {
		t.Fatal(err)
	}
	px := int(p[0x83]) * 8
	checked := 0
	for _, actor := range actors {
		x := (actor.x - townSceneWord(p, 0x80) - 4) * 8
		for xx := max(0, x, px); xx < min(224, x+16, px+16); xx++ {
			if got := canvas.RGBAAt((48+xx)*3+1, 380); got != (color.RGBA{127, 0, 128, 255}) {
				t.Fatalf("主角未在居民上方：%v", got)
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("未驗到人物重疊")
	}
}

func TestTownActorsMissingAssets(t *testing.T) {
	s, _ := townActorsTestInputs(t)
	if got, err := NewTownActors(nil, t.TempDir()); err == nil || got != nil {
		t.Fatal("缺主角卻載入")
	}
	for _, broken := range []bool{false, true} {
		dir := t.TempDir()
		if broken {
			if err := os.WriteFile(filepath.Join(dir, "g0-handpaint-v1.png"), []byte("broken"), 0644); err != nil {
				t.Fatal(err)
			}
		}
		if got, err := NewTownActors(s.player, dir); err == nil || got != nil {
			t.Fatal("缺圖／壞圖未整組拒絕")
		}
	}
}

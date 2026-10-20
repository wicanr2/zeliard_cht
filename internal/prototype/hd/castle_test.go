package hd

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestCastleNormalPathsReadOnly(t *testing.T) {
	project, out := os.Getenv("ZELIARD_PROJECT"), os.Getenv("ZELIARD_CASTLE_PAIR_OUTPUT")
	if project == "" || out == "" {
		t.Skip("須掛載正常鍵序與原版資料；skip不算驗收")
	}
	t.Setenv("ZELIARD_ENABLE_CASTLE", "1")
	runTownFramesPairWithActors(t, filepath.Join(project, "tools/castle-hd.commands"), out, true)
	b, err := os.ReadFile(filepath.Join(out, "receipt.json"))
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		Points []struct {
			Name  string         `json:"name"`
			State map[string]any `json:"presentation"`
		} `json:"checkpoints"`
	}
	if err = json.Unmarshal(b, &receipt); err != nil {
		t.Fatal(err)
	}
	expected := map[string]string{"cold": "castle", "idle-1": "castle", "idle-2": "castle", "idle-3": "castle", "right-1": "castle", "left-1": "castle", "right-2": "castle", "right-4": "muralla", "right-5": "muralla", "left-return-1": "muralla", "left-return-3": "castle"}
	transitions := map[string]bool{"right-3": true, "left-return-2": true}
	for _, point := range receipt.Points {
		if want, ok := expected[point.Name]; ok {
			if point.State["profile"] != want || point.State["used"] != true {
				t.Fatal("穩定檢查點沒有當次完整profile", point.Name, point.State)
			}
			delete(expected, point.Name)
		}
		if transitions[point.Name] {
			if point.State["used"] != false {
				t.Fatal("轉場沿用舊場景畫格", point.Name)
			}
			delete(transitions, point.Name)
		}
	}
	if len(expected) != 0 || len(transitions) != 0 || len(receipt.Points) != 13 {
		t.Fatal("正常檢查點不足", expected, transitions)
	}
}

func castleTestInputs(t *testing.T) (*TownScenery, *TownScenery, *TownBackground, string) {
	t.Helper()
	scene, _ := sceneryTestInputs(t)
	project, root := os.Getenv("ZELIARD_PROJECT"), os.Getenv("ZELIARD_ORIG")
	data := filepath.Join(project, "workplace/music-release-v1/release-build/data/assets/town")
	background, err := NewTownBackground(root, filepath.Join(data, "sky-reference.indexed"), filepath.Join(data, "sky-reference.palette"), filepath.Join(data, "restored.memory"), filepath.Join(data, "background-handpaint-v2.png"))
	if err != nil {
		t.Fatal(err)
	}
	study := filepath.Join(project, "workplace/full-hd-v1")
	castle, err := NewCastleScenery(root, filepath.Join(project, "workplace/town-hd-v1/foreground-v1/normal-v1/cold-town.memory"), filepath.Join(study, "castle-registration-v1"), background)
	if err != nil {
		t.Fatal(err)
	}
	if err = scene.AttachCastle(castle); err != nil {
		t.Fatal(err)
	}
	return scene, castle, background, study
}

func castleFixture(t *testing.T, study, name string) ([]byte, []byte, []byte) {
	t.Helper()
	read := func(suffix string) []byte {
		b, e := os.ReadFile(filepath.Join(study, "castle-normal-v1", name+"-source"+suffix))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	return read(".memory"), read(".indexed"), read(".palette")
}

func TestCastleNormalSourcesAndLayers(t *testing.T) {
	scene, castle, background, study := castleTestInputs(t)
	project := os.Getenv("ZELIARD_PROJECT")
	base := filepath.Join(project, "workplace/town-hd-v1/foreground-v1")
	ground, err := NewTownGround(scene, filepath.Join(base, "normal-v1/restored.indexed"), filepath.Join(project, "workplace/music-release-v1/release-build/data/assets/town/ground-handpaint-v1.png"))
	if err != nil {
		t.Fatal(err)
	}
	player, err := NewTownPlayer(scene, base)
	if err != nil {
		t.Fatal(err)
	}
	actors, err := NewTownActors(player, filepath.Join(study, "town-npc-art"))
	if err != nil {
		t.Fatal(err)
	}
	seen := map[byte]bool{}
	var results []map[string]any
	for _, name := range []string{"cold", "idle-1", "idle-2", "idle-3", "right-1", "left-1", "right-2", "left-return-3"} {
		t.Run(name, func(t *testing.T) {
			memory, raw, palette := castleFixture(t, study, name)
			original := cloneTownBytes(memory)
			p, r, e := memory[0x1b10:0x11b10], memory[0x11b10:0x21b10], memory[0x21b10:0x31b10]
			rgb, canvas := skyOriginal(raw, palette)
			f, err := scene.decode(0x13, p, r, e, raw, rgb)
			if err != nil {
				t.Fatal(err)
			}
			if f.scene != castle {
				t.Fatal("錯誤profile")
			}
			for _, id := range f.flags {
				if id != 0 {
					seen[id] = true
				}
			}
			if !background.Draw(canvas, 0x13, p, r, raw, rgb, image.Point{}) {
				t.Fatal("城堡背景拒絕")
			}
			before := cloneTownBytes(canvas.Pix)
			if !scene.Draw(canvas, 0x13, p, r, e, raw, rgb, image.Point{}, true) {
				t.Fatal("城堡前景拒絕")
			}
			changed := 0
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := canvas.PixOffset(x, y)
					if bytes.Equal(before[i:i+4], canvas.Pix[i:i+4]) {
						continue
					}
					changed++
					if !image.Pt(x, y).In(image.Rect(144, 234, 816, 426)) {
						t.Fatal("城堡覆寫區外", x, y)
					}
					if f.opaque[(y/3-78)*224+x/3-48] {
						t.Fatal("城堡覆寫原版人物", x, y)
					}
				}
			}
			if changed < 10000 {
				t.Fatal("城堡HD沒有有效繪製", changed)
			}
			if !ground.Draw(canvas, 0x13, p, r, e, raw, rgb, palette, image.Point{}) {
				t.Fatal("共用地面來源拒絕")
			}
			matched := actors.Draw(canvas, 0x13, p, r, e, raw, rgb, palette, image.Point{})
			if p[0xe7] < 4 && !matched {
				t.Fatal("共用居民／主角拒絕", p[0xe7])
			}
			if !bytes.Equal(memory, original) {
				t.Fatal("改寫原版")
			}
			if out := os.Getenv("ZELIARD_CASTLE_OUTPUT"); out != "" {
				file, err := os.Create(filepath.Join(out, name+"-hd.png"))
				if err != nil {
					t.Fatal(err)
				}
				err = png.Encode(file, canvas)
				file.Close()
				if err != nil {
					t.Fatal(err)
				}
			}
			results = append(results, map[string]any{"name": name, "original_memory_sha256": townBackgroundHash(memory), "profile": "castle", "changed_foreground_pixels": changed, "actors_matched": matched, "player_phase": p[0xe7], "memory_unchanged": true})
		})
	}
	if len(seen) != 4 {
		t.Fatal("正常來源未涵蓋四旗幟姿態", seen)
	}
	if out := os.Getenv("ZELIARD_CASTLE_OUTPUT"); out != "" {
		b, e := json.MarshalIndent(map[string]any{"status": "PASS_CASTLE_NORMAL_SOURCE_RENDER", "points": results, "observed_flag_ids": seen, "scope": "八個正常完成來源的唯讀離線HD合成；不是實際視窗或全遊戲完成"}, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(out, "render-receipt.json"), append(b, 10), 0644); e != nil {
			t.Fatal(e)
		}
	}
}

func TestCastleRefusalIsAtomic(t *testing.T) {
	scene, castle, _, study := castleTestInputs(t)
	memory, raw, pal := castleFixture(t, study, "cold")
	for _, name := range []string{"wrong-mode", "short-primary", "wrong-terrain", "wrong-npc", "wrong-mask", "wrong-cache", "wrong-native", "wrong-width", "missing-flag", "small-canvas"} {
		t.Run(name, func(t *testing.T) {
			p, r, e := cloneTownBytes(memory[0x1b10:0x11b10]), cloneTownBytes(memory[0x11b10:0x21b10]), cloneTownBytes(memory[0x21b10:0x31b10])
			indices := cloneTownBytes(raw)
			rgb, canvas := skyOriginal(indices, pal)
			mode := uint8(0x13)
			switch name {
			case "wrong-mode":
				mode = 3
			case "short-primary":
				p = p[:32]
			case "wrong-terrain":
				r[0x8100] ^= 1
			case "wrong-npc":
				r[0x4100] ^= 1
			case "wrong-mask":
				r[0xd000+2*8] ^= 1
			case "wrong-cache":
				for col := 0; col < 28; col++ {
					for row := 0; row < 3; row++ {
						id := p[0xe000+col*8+row]
						if id >= 2 && id <= 5 {
							p[0xe000+col*8+row] = 6
						}
					}
				}
			case "wrong-native":
				indices[78*320+48] ^= 1
			case "wrong-width":
				p[0xc002] ^= 1
			case "missing-flag":
				old := castle.castle.flags
				castle.castle.flags = [4]*image.RGBA{}
				defer func() { castle.castle.flags = old }()
			case "small-canvas":
				canvas = image.NewRGBA(image.Rect(0, 0, 100, 100))
			}
			before := cloneTownBytes(canvas.Pix)
			if scene.Draw(canvas, mode, p, r, e, indices, rgb, image.Point{}, true) || !bytes.Equal(canvas.Pix, before) {
				t.Fatal("拒絕改寫畫布", name)
			}
		})
	}
}

func TestCastlePresentationCrossProfile(t *testing.T) {
	scene, _, _, study := castleTestInputs(t)
	memory, raw, palette := castleFixture(t, study, "cold")
	rgb, _ := skyOriginal(raw, palette)
	in := townFrameInput{mode: 0x13, primary: memory[0x1b10:0x11b10], resources: memory[0x11b10:0x21b10], extra: memory[0x21b10:0x31b10], indexed: raw, rgb: rgb, palette: palette, steps: 100, cycles: 100}
	frames := &TownFrames{scene: scene}
	frames.finish(in)
	if frames.complete == nil {
		t.Fatal("城堡完成來源未接受")
	}
	project := os.Getenv("ZELIARD_PROJECT")
	base := filepath.Join(project, "workplace/town-hd-v1/foreground-v1/normal-v1")
	m, err := os.ReadFile(filepath.Join(base, "restored.memory"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := os.ReadFile(filepath.Join(base, "restored.indexed"))
	if err != nil {
		t.Fatal(err)
	}
	c, _ := skyOriginal(r, palette)
	in.primary, in.resources, in.extra, in.indexed, in.rgb = m[0x1b10:0x11b10], m[0x11b10:0x21b10], m[0x21b10:0x31b10], r, c
	in.steps, in.cycles = 200, 200
	frames.drawing = true
	if frames.selectFrame(in, false) != nil || frames.complete != nil {
		t.Fatal("城堡快照洩漏到Muralla")
	}
	frames.finish(in)
	if frames.complete == nil {
		t.Fatal("已掛城堡profile後Muralla來源被拒絕")
	}
	_, canvas := skyOriginal(r, palette)
	if !scene.Draw(canvas, 0x13, in.primary, in.resources, in.extra, in.indexed, in.rgb, image.Point{}, true) {
		t.Fatal("已掛城堡profile後Muralla繪製被拒絕")
	}
}

func TestCastlePlayerHiddenAtRightBoundary(t *testing.T) {
	scene, _, _, study := castleTestInputs(t)
	read := func(suffix string) []byte {
		b, e := os.ReadFile(filepath.Join(study, "castle-pair-v2/uncovered"+suffix))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	memory, raw, palette := read(".memory"), read(".indexed"), read(".palette")
	if townBackgroundHash(memory) != "c6a54aa329a9483f7a6cbf03af49756e840242a94e80a247233c7f827f6799e8" {
		t.Fatal("邊界正常來源身份不符")
	}
	p, r, e := memory[0x1b10:0x11b10], memory[0x11b10:0x21b10], memory[0x21b10:0x31b10]
	if p[0x83] != 27 {
		t.Fatal("未到真實右邊界")
	}
	rgb, canvas := skyOriginal(raw, palette)
	if _, err := scene.decode(0x13, p, r, e, raw, rgb); err != nil {
		t.Fatal("正常邊界14336像素重建失配", err)
	}
	project := os.Getenv("ZELIARD_PROJECT")
	player, err := NewTownPlayer(scene, filepath.Join(project, "workplace/town-hd-v1/foreground-v1"))
	if err != nil {
		t.Fatal(err)
	}
	actors, err := NewTownActors(player, filepath.Join(study, "town-npc-art"))
	if err != nil {
		t.Fatal(err)
	}
	if !scene.Draw(canvas, 0x13, p, r, e, raw, rgb, image.Point{}, true) {
		t.Fatal("正常邊界前景HD拒絕")
	}
	before := cloneTownBytes(canvas.Pix)
	if player.Draw(canvas, 0x13, p, r, e, raw, rgb, palette, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
		t.Fatal("在主角隱藏狀態畫入半個主角")
	}
	if !actors.Draw(canvas, 0x13, p, r, e, raw, rgb, palette, image.Point{}) || actors.State()["player_visible"] != false {
		t.Fatal("主角隱藏導致居民撤除")
	}
	for y := 354; y < 426; y++ {
		for x := 792; x < 816; x++ {
			i := canvas.PixOffset(x, y)
			if !bytes.Equal(before[i:i+4], canvas.Pix[i:i+4]) {
				t.Fatal("右側出現虛構主角", x, y)
			}
		}
	}
}

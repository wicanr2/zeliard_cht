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

// 五份正常原版來源的圖層與邊界。資料變動只在單元輸入副本，不寫原版。
func TestTownReturnTerrainSourcesAndLayers(t *testing.T) {
	s, _ := sceneryTestInputs(t)
	project, study := os.Getenv("ZELIARD_PROJECT"), os.Getenv("ZELIARD_CAVE_STUDY")
	out := os.Getenv("ZELIARD_RETURN_TOWN_UNIT_OUTPUT")
	if study == "" || out == "" {
		t.Skip("須提供已驗原生讀取來源及新的圖面輸出；skip不計驗收")
	}
	read := func(p string) []byte {
		b, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	var footprint struct {
		Records []struct {
			Source            string
			Complete          bool
			NativeDifferences int `json:"native_differences"`
		}
	}
	if e := json.Unmarshal(read(filepath.Join(study, "return-town-footprint-v1.json")), &footprint); e != nil || len(footprint.Records) != 5 {
		t.Fatal("原版來源", e)
	}
	tools := filepath.Join(study, "exhaustion-tools-v6")
	bg, e := NewTownBackground(os.Getenv("ZELIARD_ORIG"), filepath.Join(tools, "sky-reference.indexed"), filepath.Join(tools, "sky-reference.palette"), filepath.Join(tools, "restored.memory"), filepath.Join(tools, "background-handpaint-v2.png"))
	if e != nil {
		t.Fatal(e)
	}
	ground, e := NewTownGround(s, filepath.Join(tools, "sky-reference.indexed"), filepath.Join(tools, "ground-handpaint-v1.png"))
	if e != nil {
		t.Fatal(e)
	}
	player, e := NewTownPlayer(s, filepath.Join(tools, "player-art"))
	if e != nil {
		t.Fatal(e)
	}
	if e = os.Mkdir(out, 0755); e != nil {
		t.Fatal(e)
	}
	var records []map[string]any
	var first townFrameInput
	drawAll := func(in townFrameInput) *image.RGBA {
		_, c := skyOriginal(in.indexed, in.palette)
		if !bg.Draw(c, in.mode, in.primary, in.resources, in.indexed, in.rgb, image.Point{}) || !s.Draw(c, in.mode, in.primary, in.resources, in.extra, in.indexed, in.rgb, image.Point{}, true) || !ground.Draw(c, in.mode, in.primary, in.resources, in.extra, in.indexed, in.rgb, in.palette, image.Point{}) || !player.Draw(c, in.mode, in.primary, in.resources, in.extra, in.indexed, in.rgb, in.palette, image.Point{}) {
			t.Fatal("正常送返圖層未恢復")
		}
		return c
	}
	for i, r := range footprint.Records {
		if !r.Complete || r.NativeDifferences != 0 {
			t.Fatal("不可用中途來源")
		}
		prefix := filepath.Join(project, r.Source)
		m, raw := read(prefix+".memory"), read(prefix+".indexed")
		palettePath := prefix + ".palette"
		if i == 4 {
			// GUI原擷取沒有獨立palette檔；使用已核同一CPU時刻的原版重播輸出。
			replay := filepath.Join(study, "exhaustion-gui-pair-v2/source-05")
			if !bytes.Equal(m, read(replay+".memory")) || !bytes.Equal(raw, read(replay+".indexed")) {
				t.Fatal("GUI與原版重播來源不相同")
			}
			palettePath = replay + ".palette"
		}
		pal := read(palettePath)
		before := cloneTownBytes(m)
		rgb, _ := skyOriginal(raw, pal)
		in := townFrameInput{mode: 0x13, primary: m[0x1b10:0x11b10], resources: m[0x11b10:0x21b10], extra: m[0x21b10:0x31b10], indexed: raw, rgb: rgb, palette: pal}
		if !townFrameSource(in) {
			t.Fatal("正常送返來源拒絕", r.Source)
		}
		frames := &TownFrames{scene: s}
		frames.finish(in)
		if frames.complete == nil {
			t.Fatal("正常完整畫格拒絕")
		}
		canvas := drawAll(in)
		if !bytes.Equal(m, before) {
			t.Fatal("來源被改寫")
		}
		name := filepath.Base(r.Source) + ".png"
		if i == 4 {
			name = "GUI-return.png"
		}
		f, e := os.Create(filepath.Join(out, name))
		if e != nil {
			t.Fatal(e)
		}
		if e = png.Encode(f, canvas); e != nil {
			t.Fatal(e)
		}
		if e = f.Close(); e != nil {
			t.Fatal(e)
		}
		records = append(records, map[string]any{"source": r.Source, "png": name, "full_source_unchanged": true, "background": true, "scenery": true, "ground": true, "player": true, "completed_frame": true})
		if i == 0 {
			first = copyTownInput(in)
		}
	}
	original := drawAll(first)
	unused := copyTownInput(first)
	for i := 0xae00; i < 0xd000; i++ {
		unused.resources[i] ^= 0xa5
	}
	if !bytes.Equal(original.Pix, drawAll(unused).Pix) {
		t.Fatal("排除區影響圖面")
	}
	seen239 := false
	for _, column := range s.tiles {
		for _, id := range column {
			if id >= 240 {
				t.Fatal("canonical上界")
			}
			seen239 = seen239 || id == 239
		}
	}
	if !seen239 {
		t.Fatal("未涵蓋239邊界")
	}
	for _, kind := range []string{"last-used-byte", "first-mask-byte", "last-mask-byte", "NPC", "player", "unknown-terrain-240", "unknown-terrain-255"} {
		t.Run(kind, func(t *testing.T) {
			in := copyTownInput(first)
			switch kind {
			case "last-used-byte":
				in.resources[0xadff] ^= 1
			case "first-mask-byte":
				in.resources[0xd000] ^= 1
			case "last-mask-byte":
				in.resources[0xd7ff] ^= 1
			case "NPC":
				in.resources[0x4100] ^= 1
			case "player":
				in.resources[0x6000] ^= 1
			case "unknown-terrain-240":
				in.primary[0xe000] = 240
			case "unknown-terrain-255":
				in.primary[0xe000] = 255
				in.primary[townSceneWord(in.primary, 0xff2a)+32] = 240
			}
			_, canvas := skyOriginal(in.indexed, in.palette)
			before := cloneTownBytes(canvas.Pix)
			if s.Draw(canvas, in.mode, in.primary, in.resources, in.extra, in.indexed, in.rgb, image.Point{}, true) || !bytes.Equal(before, canvas.Pix) {
				t.Fatal("來源失配未原子退回")
			}
			if kind == "last-used-byte" || kind == "first-mask-byte" || kind == "last-mask-byte" || kind == "unknown-terrain-240" {
				_, c := skyOriginal(in.indexed, in.palette)
				b := cloneTownBytes(c.Pix)
				if bg.Draw(c, in.mode, in.primary, in.resources, in.indexed, in.rgb, image.Point{}) || !bytes.Equal(b, c.Pix) {
					t.Fatal("遠山失配未原子退回")
				}
			}
		})
	}
	result := map[string]any{"status": "PASS_RETURN_TOWN_LAYERS_AND_BOUNDS", "records": records, "unused_atlas_tail_identical": true, "canonical_239_present": true, "unknown_ID_240_rejected": true, "scope": "五份正常原版快照的單元圖面；未執行或注入原版，真正GUI及原生日誌另驗"}
	b, e := json.MarshalIndent(result, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(b, 10), 0644); e != nil {
		t.Fatal(e)
	}
}

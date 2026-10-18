package hd

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"testing"
)

func townFrameFixture(t *testing.T) (*TownScenery, townFrameInput) {
	t.Helper()
	s, base := sceneryTestInputs(t)
	memory, err := os.ReadFile(filepath.Join(base, "normal-v1/restored.memory"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(base, "normal-v1/restored.indexed"))
	if err != nil {
		t.Fatal(err)
	}
	rgb, _ := skyOriginal(raw, s.palette)
	return s, townFrameInput{0x13, memory[0x1b10:0x11b10], memory[0x11b10:0x21b10], memory[0x21b10:0x31b10], raw, rgb, cloneTownBytes(s.palette), 100, 200}
}
func copyTownInput(in townFrameInput) townFrameInput {
	in.primary = cloneTownBytes(in.primary)
	in.resources = cloneTownBytes(in.resources)
	in.extra = cloneTownBytes(in.extra)
	in.indexed = cloneTownBytes(in.indexed)
	in.rgb = cloneTownBytes(in.rgb)
	in.palette = cloneTownBytes(in.palette)
	return in
}

func TestTownFramesCompleteAndPartialIsolation(t *testing.T) {
	s, in := townFrameFixture(t)
	f := &TownFrames{scene: s}
	if f.selectFrame(in, false) != nil {
		t.Fatal("無完成來源卻呈現")
	}
	f.finish(in)
	if f.complete == nil {
		t.Fatal("正常完成來源被拒絕")
	}
	original := copyTownInput(in)
	in = copyTownInput(in)
	in.steps++
	in.cycles++
	f.begin(in)
	// 模擬已驗批次中的VGA寫入；同時修改畫帶外HUD。
	in.indexed[120*320+160] ^= 1
	in.rgb[(120*320+160)*3] ^= 1
	in.indexed[0] ^= 1
	in.rgb[0] ^= 1
	got := f.selectFrame(in, false)
	if got == nil {
		t.Fatal("繪製中途未保留完成來源")
	}
	if !townFrameEqual(got.Indexed, original.indexed) || got.Indexed[0] != in.indexed[0] || got.RGB[0] != in.rgb[0] {
		t.Fatal("完整畫格或區外當次原版遺失")
	}
	if !bytes.Equal(got.Primary, original.primary) || !bytes.Equal(got.SourceIndexed, original.indexed) {
		t.Fatal("來源不是完成點")
	}
	canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
	if !s.Draw(canvas, 0x13, got.Primary, got.Resources, got.Extra, got.Indexed, got.RGB, image.Point{}, true) {
		t.Fatal("呈現無法通過原有完整像素守衛")
	}
	if in.indexed[120*320+160] == original.indexed[120*320+160] {
		t.Fatal("呈現回寫原始色號")
	}
	// 完成點資料須深複製；呼叫者後續輸入變更不影響已保存快照。
	in.primary[0x80] ^= 1
	if !bytes.Equal(f.complete.Primary, original.primary) {
		t.Fatal("來源未深複製")
	}
	got.Indexed[120*320+160] ^= 1
	if !townFrameEqual(f.complete.SourceIndexed, original.indexed) {
		t.Fatal("呈現輸出回寫快照")
	}
	if f.State()["lag_steps"] != uint64(1) {
		t.Fatal("未記來源時間差", f.State())
	}
}

func TestTownFramesRefuseUnknownAndModal(t *testing.T) {
	s, source := townFrameFixture(t)
	cases := map[string]func(*townFrameInput){
		"mode":                func(i *townFrameInput) { i.mode = 3 },
		"palette":             func(i *townFrameInput) { i.palette[0] ^= 1 },
		"scroll-code":         func(i *townFrameInput) { i.primary[0x36e6] ^= 1 },
		"flow-code":           func(i *townFrameInput) { i.primary[0x6852] ^= 1 },
		"module":              func(i *townFrameInput) { i.primary[0xc000] ^= 1 },
		"atlas":               func(i *townFrameInput) { i.resources[0x8100] ^= 1 },
		"mask":                func(i *townFrameInput) { i.extra[0x7000] ^= 1 },
		"unknown-actor-group": func(i *townFrameInput) { i.primary[townSceneWord(i.primary, 0xc00f)+2] = 5 },
		"short-indexed":       func(i *townFrameInput) { i.indexed = i.indexed[:100] },
		"short-rgb":           func(i *townFrameInput) { i.rgb = i.rgb[:100] },
		"short-source":        func(i *townFrameInput) { i.primary = i.primary[:100] },
		"steps-rewind":        func(i *townFrameInput) { i.steps-- },
		"cycles-rewind":       func(i *townFrameInput) { i.cycles-- },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			f := &TownFrames{scene: s}
			f.finish(source)
			f.begin(source)
			in := copyTownInput(source)
			mutate(&in)
			if f.selectFrame(in, false) != nil || f.complete != nil {
				t.Fatal("未知來源／短輸入仍保留快照")
			}
		})
	}
	t.Run("modal", func(t *testing.T) {
		f := &TownFrames{scene: s}
		f.finish(source)
		f.begin(source)
		if f.selectFrame(source, true) != nil || f.complete != nil {
			t.Fatal("城鎮蓋住對話")
		}
	})
	t.Run("outside-batch-write", func(t *testing.T) {
		f := &TownFrames{scene: s}
		f.finish(source)
		in := copyTownInput(source)
		in.indexed[120*320+160] ^= 1
		if f.selectFrame(in, false) != nil || f.complete != nil {
			t.Fatal("批次外未知寫入未撤除")
		}
	})
	t.Run("failed-completion", func(t *testing.T) {
		f := &TownFrames{scene: s}
		f.finish(source)
		in := copyTownInput(source)
		in.indexed[120*320+160] ^= 1
		f.finish(in)
		if f.complete != nil {
			t.Fatal("失敗完成點保留過期來源")
		}
	})
}

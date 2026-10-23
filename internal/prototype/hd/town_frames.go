package hd

import (
	"bytes"
	"image"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/glyphpatch"
)

// TownFrames只保留已驗原版完成點。位址、來源與時間差見規格008。
type TownFrames struct {
	original          *oracle.Oracle
	scene             *TownScenery
	complete          *TownPresentation
	wipeSource        *TownPresentation
	wipeErased        int
	drawing           bool
	captured, refused uint64
	state             map[string]any
	lastRefusal       map[string]any
	// Dialogue 回傳居民對話是否進行中及目前對話框。對話期間原版仍繼續繪製居民走動，
	// 完成點的畫面需用對話前的已驗畫面補上框內像素。規格066。
	Dialogue func() (bool, image.Rectangle)
	backdrop *TownPresentation
}

// TownPresentation的來源屬完成時刻，畫帶外仍為當次原版。
// 呼叫者只讀；此資料不可冒稱當次CPU的原版收據。
type TownPresentation struct {
	Indexed, RGB, Primary, Resources, Extra, SourceIndexed, Palette []byte
	Steps, Cycles                                                   uint64
	Wiping                                                          bool
	EraseMask                                                       []byte
	DialogueBox                                                     image.Rectangle
}

var townPresentationBox = image.Rect(48, 24, 272, 158)

type townFrameInput struct {
	mode                                             uint8
	primary, resources, extra, indexed, rgb, palette []byte
	steps, cycles                                    uint64
}

func NewTownFrames(o *oracle.Oracle, scene *TownScenery) *TownFrames {
	if o == nil || scene == nil {
		return nil
	}
	t := &TownFrames{original: o, scene: scene}
	for _, off := range []uint16{0x67d8, 0x6852, 0x68ac} {
		o.OnCall(oracle.Far(0x1b1, off), func(o *oracle.Oracle) {
			if o.DSReg() != 0x1b1 {
				t.clear()
				return
			}
			t.begin(t.read(nil, nil))
		})
	}
	o.OnCall(oracle.Far(0x1b1, 0x68ba), func(o *oracle.Oracle) {
		if o.DSReg() != 0x1b1 {
			t.clear()
			return
		}
		_, _, rgb := glyphpatch.ScreenRGB(o)
		t.finish(t.read(glyphpatch.Indexed(o), rgb))
	})
	return t
}

func (t *TownFrames) read(raw, rgb []byte) townFrameInput {
	o := t.original
	pal := o.Palette()
	p := make([]byte, 0, 768)
	for _, c := range pal {
		p = append(p, c[:]...)
	}
	return townFrameInput{o.VideoMode(), o.Bytes(oracle.Far(0x1b1, 0), 65536), o.Bytes(oracle.Far(0x11b1, 0), 65536), o.Bytes(oracle.Far(0x21b1, 0), 65536), raw, rgb, p, o.Steps(), o.Cycles()}
}

func townFrameSource(in townFrameInput) bool {
	if in.mode != 0x13 || len(in.primary) != 65536 || len(in.resources) != 65536 || len(in.extra) != 65536 || len(in.palette) != 768 {
		return false
	}
	for _, g := range []struct {
		b    []byte
		hash string
	}{
		{in.palette, "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c"},
		{in.primary[0x3000:0x371c], "cc52036cd4c7c43747d3235e3026b0e333c7eb692ed651dff569886b2233588d"},
		{in.primary[0x6026:0x6c72], "6569836474343da7fd58f63b4342b7e5a323192e194920a30753c36020583d53"},
	} {
		if townBackgroundHash(g.b) != g.hash {
			return false
		}
	}
	_, err := townSceneSources(in.primary, in.resources, in.extra)
	return err == nil
}

func (t *TownFrames) clear() { t.complete = nil; t.drawing = false }
func (t *TownFrames) begin(in townFrameInput) {
	if !t.scene.frameSource(in) {
		t.clear()
		return
	}
	t.drawing = true
}
func cloneTownBytes(b []byte) []byte { return append([]byte(nil), b...) }
func (t *TownFrames) finish(in townFrameInput) {
	t.drawing = false
	if t.Dialogue != nil {
		if active, box := t.Dialogue(); active {
			// 對話中的完成畫面含原版對話框。框已確定且有對話前的底圖時，框內換成底圖像素後照原本驗證，
			// 相機或場景改變時驗證會拒收；框未確定時不收這次完成，沿用上一張。
			if box.Empty() || !box.In(image.Rect(0, 0, 320, 200)) || t.backdrop == nil || len(in.indexed) != 64000 || len(in.rgb) != 192000 {
				return
			}
			masked := in
			masked.indexed, masked.rgb = cloneTownBytes(in.indexed), cloneTownBytes(in.rgb)
			for y := box.Min.Y; y < box.Max.Y; y++ {
				a, b := y*320+box.Min.X, y*320+box.Max.X
				copy(masked.indexed[a:b], t.backdrop.SourceIndexed[a:b])
				copy(masked.rgb[a*3:b*3], t.backdrop.RGB[a*3:b*3])
			}
			// 背景圖塊快取 E000 每欄8列，第r列對應畫面 y=78+8r；被框蓋住的格子原版改成框用編號，
			// 同樣取對話前底圖的值，讓像素與快取來自同一份狀態。
			masked.primary = cloneTownBytes(in.primary)
			for col := 0; col < 28; col++ {
				for row := 0; row < 8; row++ {
					if image.Rect(48+col*8, 78+row*8, 56+col*8, 86+row*8).Overlaps(box) {
						masked.primary[0xe000+col*8+row] = t.backdrop.Primary[0xe000+col*8+row]
					}
				}
			}
			in = masked
		}
	}
	if !t.scene.frameSource(in) {
		t.refused++
		t.lastRefusal = map[string]any{"steps": in.steps, "cycles": in.cycles, "reason": "source"}
		t.clear()
		return
	}
	if _, err := t.scene.decode(in.mode, in.primary, in.resources, in.extra, in.indexed, in.rgb); err != nil {
		t.refused++
		t.lastRefusal = map[string]any{"steps": in.steps, "cycles": in.cycles, "reason": err.Error()}
		t.clear()
		return
	}
	t.complete = &TownPresentation{Indexed: cloneTownBytes(in.indexed), RGB: cloneTownBytes(in.rgb), Primary: cloneTownBytes(in.primary), Resources: cloneTownBytes(in.resources), Extra: cloneTownBytes(in.extra), SourceIndexed: cloneTownBytes(in.indexed), Palette: cloneTownBytes(in.palette), Steps: in.steps, Cycles: in.cycles}
	t.wipeSource = t.complete
	t.backdrop = t.complete
	t.wipeErased = 0
	t.captured++
	t.lastRefusal = nil
}

func townFrameEqual(a, b []byte) bool {
	if len(a) != 64000 || len(b) != 64000 {
		return false
	}
	for y := townPresentationBox.Min.Y; y < townPresentationBox.Max.Y; y++ {
		start := y*320 + townPresentationBox.Min.X
		end := y*320 + townPresentationBox.Max.X
		if !bytes.Equal(a[start:end], b[start:end]) {
			return false
		}
	}
	return true
}

// Present不執行原版；未通過守衛時立即撤除，原貌與HD共用選出的基底。
func (t *TownFrames) Present(raw, rgb []byte, blocked bool) *TownPresentation {
	return t.selectFrame(t.read(raw, rgb), blocked)
}

// PresentDialogue只保留對話框之外完全相同的已驗城鎮；對話框由前端最後放回。
func (t *TownFrames) PresentDialogue(raw, rgb []byte, box image.Rectangle) *TownPresentation {
	if t == nil {
		return nil
	}
	return t.selectDialogueFrame(t.read(raw, rgb), box)
}

func (t *TownFrames) selectDialogueFrame(in townFrameInput, box image.Rectangle) *TownPresentation {
	t.state = map[string]any{"used": false, "dialogue_backdrop": false, "captured": t.captured, "refused": t.refused, "scope": "居民對話框外的已驗完成背景"}
	f := t.complete
	if f == nil {
		// 正常居民進字仍會經過68BA；被文字覆蓋的候選不能取代已驗背景。
		f = t.wipeSource
	}
	if f == nil {
		t.state["reason"] = "missing completed backdrop"
		return nil
	}
	if len(in.indexed) != 64000 || len(in.rgb) != 192000 || !t.scene.frameSource(in) || in.steps < f.Steps || in.cycles < f.Cycles || townSceneWord(in.primary, 0x80) != townSceneWord(f.Primary, 0x80) || t.scene.sourceProfile(in) != t.scene.sourceProfile(townFrameInput{mode: in.mode, primary: f.Primary, resources: f.Resources, extra: f.Extra, palette: f.Palette}) {
		t.clear()
		t.wipeSource = nil
		t.state["reason"] = "dialogue source, camera or timeline"
		return nil
	}
	if box.Empty() {
		t.state["reason"] = "dialogue box pending"
		return nil
	} // 框尚未建立，保留快照但不畫未知遮擋。
	if box.Dx() != 176 || box.Dy() < 16 || box.Dy() > 86 || (box.Dy()-6)%10 != 0 || (box.Min.X != 56 && box.Min.X != 88) || !box.In(image.Rect(48, 24, 272, 148)) {
		t.clear()
		t.wipeSource = nil
		t.state["reason"] = "dialogue geometry"
		return nil
	}
	// 原版正在重畫城鎮時（居民走動），下一個完成點會帶來新的已驗畫面；這段期間沿用上一張，不撤除 HD。
	for y := townPresentationBox.Min.Y; y < townPresentationBox.Max.Y && !t.drawing; y++ {
		for x := townPresentationBox.Min.X; x < townPresentationBox.Max.X; x++ {
			if image.Pt(x, y).In(box) {
				continue
			}
			i := y*320 + x
			if in.indexed[i] != f.SourceIndexed[i] || !bytes.Equal(in.rgb[i*3:i*3+3], f.RGB[i*3:i*3+3]) {
				t.clear()
				t.wipeSource = nil
				t.state["reason"] = "outside dialogue changed"
				t.state["outside_pixel"] = image.Pt(x, y)
				return nil
			}
		}
	}
	masked := in
	masked.indexed = cloneTownBytes(in.indexed)
	masked.rgb = cloneTownBytes(in.rgb)
	for y := box.Min.Y; y < box.Max.Y; y++ {
		a, b := y*320+box.Min.X, y*320+box.Max.X
		copy(masked.indexed[a:b], f.SourceIndexed[a:b])
		copy(masked.rgb[a*3:b*3], f.RGB[a*3:b*3])
	}
	t.complete = f
	got := t.selectFrame(masked, false)
	if got != nil {
		got.DialogueBox = box
		t.state["dialogue_backdrop"] = true
		t.state["dialogue_box"] = box
	}
	return got
}
func (t *TownFrames) selectFrame(in townFrameInput, blocked bool) *TownPresentation {
	t.state = map[string]any{"used": false, "drawing": t.drawing, "captured": t.captured, "refused": t.refused, "scope": "完成時刻的城鎮畫帶；畫帶外為當次原版"}
	t.state["last_refusal"] = t.lastRefusal
	f := t.complete
	sourceOK := f != nil && t.scene.frameSource(in)
	wipeCandidate := !sourceOK || (!t.drawing && f != nil && !townFrameEqual(in.indexed, f.SourceIndexed))
	if !blocked && wipeCandidate && t.wipeSource != nil {
		if wiping, erased := nativeTownWipe(in, t.wipeSource); wiping != nil && erased >= t.wipeErased {
			t.wipeErased = erased
			t.state["used"], t.state["wiping"], t.state["erased"] = true, true, erased
			t.state["source_steps"] = wiping.Steps
			t.state["source_indexed_sha256"] = townBackgroundHash(wiping.SourceIndexed)
			return wiping
		}
	}
	if blocked || f == nil || len(in.indexed) != 64000 || len(in.rgb) != 192000 || !sourceOK || in.steps < f.Steps || in.cycles < f.Cycles || (!t.drawing && !townFrameEqual(in.indexed, f.SourceIndexed)) {
		t.clear()
		t.wipeSource = nil
		return nil
	}
	current := t.scene.sourceProfile(in)
	saved := t.scene.sourceProfile(townFrameInput{mode: in.mode, primary: f.Primary, resources: f.Resources, extra: f.Extra, palette: f.Palette})
	if current != saved {
		t.clear()
		t.wipeSource = nil
		return nil
	}
	raw, rgb := cloneTownBytes(in.indexed), cloneTownBytes(in.rgb)
	for y := townPresentationBox.Min.Y; y < townPresentationBox.Max.Y; y++ {
		start := y*320 + townPresentationBox.Min.X
		end := y*320 + townPresentationBox.Max.X
		copy(raw[start:end], f.SourceIndexed[start:end])
		copy(rgb[start*3:end*3], f.RGB[start*3:end*3])
	}
	t.state["used"] = true
	if saved.castle != nil {
		t.state["profile"] = "castle"
		var flags []map[string]int
		for col := 0; col < 28; col++ {
			for row := 0; row < 3; row++ {
				id := int(f.Primary[0xe000+col*8+row])
				if id >= 2 && id <= 5 {
					flags = append(flags, map[string]int{"column": col, "row": row, "id": id})
				}
			}
		}
		t.state["flag_cells"] = flags
	} else {
		t.state["profile"] = "muralla"
	}
	t.state["source_steps"] = f.Steps
	t.state["source_cycles"] = f.Cycles
	t.state["lag_steps"] = in.steps - f.Steps
	t.state["lag_cycles"] = in.cycles - f.Cycles
	t.state["rect"] = townPresentationBox
	t.state["source_indexed_sha256"] = townBackgroundHash(f.SourceIndexed)
	t.state["source_primary_sha256"] = townBackgroundHash(f.Primary)
	t.state["source_resources_sha256"] = townBackgroundHash(f.Resources)
	t.state["source_extra_sha256"] = townBackgroundHash(f.Extra)
	t.state["source_palette_sha256"] = townBackgroundHash(f.Palette)
	t.state["raw_indexed_sha256"] = townBackgroundHash(in.indexed)
	t.state["presented_indexed_sha256"] = townBackgroundHash(raw)
	t.state["source_camera"] = townSceneWord(f.Primary, 0x80)
	t.state["source_background_cache"] = cloneTownBytes(f.Primary[0xe000:0xe0e0])
	t.state["source_module"] = cloneTownBytes(f.Primary[0xc000:0xc040])
	return &TownPresentation{Indexed: raw, RGB: rgb, Primary: f.Primary, Resources: f.Resources, Extra: f.Extra, SourceIndexed: f.SourceIndexed, Palette: f.Palette, Steps: f.Steps, Cycles: f.Cycles}
}
func (t *TownFrames) State() map[string]any {
	if t == nil {
		return nil
	}
	return t.state
}

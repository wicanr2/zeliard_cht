package hd

import (
	"bytes"
	"image"

	"github.com/wicanr2/dosgolem/oracle"
)

// TownFrames只保留已驗原版完成點。位址、來源與時間差見規格008。
type TownFrames struct {
	original          *oracle.Oracle
	scene             *TownScenery
	complete          *TownPresentation
	drawing           bool
	captured, refused uint64
	state             map[string]any
	lastRefusal       map[string]any
}

// TownPresentation的來源屬完成時刻，畫帶外仍為當次原版。
// 呼叫者只讀；此資料不可冒稱當次CPU的原版收據。
type TownPresentation struct {
	Indexed, RGB, Primary, Resources, Extra, SourceIndexed, Palette []byte
	Steps, Cycles                                                   uint64
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
		_, _, rgb := o.ScreenRGB()
		t.finish(t.read(o.Indexed(), rgb))
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
	if !townFrameSource(in) {
		t.clear()
		return
	}
	t.drawing = true
}
func cloneTownBytes(b []byte) []byte { return append([]byte(nil), b...) }
func (t *TownFrames) finish(in townFrameInput) {
	t.drawing = false
	if !townFrameSource(in) {
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
func (t *TownFrames) selectFrame(in townFrameInput, blocked bool) *TownPresentation {
	t.state = map[string]any{"used": false, "drawing": t.drawing, "captured": t.captured, "refused": t.refused, "scope": "完成時刻的城鎮畫帶；畫帶外為當次原版"}
	t.state["last_refusal"] = t.lastRefusal
	f := t.complete
	if blocked || f == nil || len(in.indexed) != 64000 || len(in.rgb) != 192000 || !townFrameSource(in) || in.steps < f.Steps || in.cycles < f.Cycles || (!t.drawing && !townFrameEqual(in.indexed, f.SourceIndexed)) {
		t.clear()
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

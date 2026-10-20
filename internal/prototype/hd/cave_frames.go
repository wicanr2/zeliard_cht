// 洞窟完成來源與圖層契約見docs/spec/012-cave-frame-sources.md。
package hd

import (
	"bytes"
	"fmt"
	"image"

	"github.com/wicanr2/dosgolem/oracle"
)

var cavePresentationBox = image.Rect(48, 14, 272, 158)

// CavePresentation分別保存完成時刻與主角繪製時刻。它不是當次CPU記憶體。
type CavePresentation struct {
	Indexed, RGB, Palette, SourceIndexed, Memory, BodyMemory []byte
	Steps, Cycles, BodySteps, BodyCycles                     uint64
	Decoded                                                  *caveDecoded
	Paints                                                   []CavePaint
}

type caveFrameInput struct {
	mode                          uint8
	memory, indexed, rgb, palette []byte
	steps, cycles                 uint64
	hook                          uint16
}

// CaveRejected是未被呈現的真實候選，僅供本地蒐證。
type CaveRejected struct {
	Memory, BodyMemory, Indexed, Palette []byte
	Steps, Cycles, BodySteps, BodyCycles uint64
	Hook                                 uint16
	Reason                               string
	Paints                               []CavePaint
}

// CaveFrames只觀察原版，不執行指令或改寫記憶體。
type CaveFrames struct {
	original           *oracle.Oracle
	complete           *CavePresentation
	body               *caveFrameInput
	paints             []CavePaint
	drawing            bool
	captured, refused  uint64
	state, lastRefusal map[string]any
	rejected           *CaveRejected
	rejectedState      map[string]any
}

func NewCaveFrames(o *oracle.Oracle) *CaveFrames {
	if o == nil {
		return nil
	}
	f := &CaveFrames{original: o}
	for _, off := range []uint16{0x703c, 0x7136} {
		o.OnCall(oracle.Far(0x1b1, off), func(o *oracle.Oracle) {
			if o.DSReg() != 0x1b1 {
				f.clear()
				return
			}
			f.begin(f.read(nil, nil))
		})
	}
	o.OnCall(oracle.Far(0x1b1, 0x4129), func(o *oracle.Oracle) {
		if o.DSReg() != 0x1b1 {
			f.clear()
			return
		}
		in := f.read(nil, nil)
		if !caveFrameSource(in) {
			f.clear()
			return
		}
		f.body = &in
		f.paints = nil
	})
	for _, off := range []uint16{0x3272, 0x3599} {
		o.OnCall(oracle.Far(0x1b1, off), func(o *oracle.Oracle) { f.observePaint(o, off) })
	}
	for _, off := range []uint16{0x7114, 0x714c} {
		o.OnCall(oracle.Far(0x1b1, off), func(o *oracle.Oracle) {
			if o.DSReg() != 0x1b1 {
				f.clear()
				return
			}
			_, _, rgb := o.ScreenRGB()
			f.finish(f.read(o.Indexed(), rgb))
		})
	}
	return f
}

func (f *CaveFrames) read(raw, rgb []byte) caveFrameInput {
	o := f.original
	var pal []byte
	for _, c := range o.Palette() {
		pal = append(pal, c[:]...)
	}
	return caveFrameInput{o.VideoMode(), o.Bytes(oracle.Far(0, 0), 1<<20), raw, rgb, pal, o.Steps(), o.Cycles(), o.IP().Off}
}

func caveFrameSource(in caveFrameInput) bool {
	if in.mode != 0x13 || len(in.memory) != 1<<20 || len(in.palette) != 768 {
		return false
	}
	p := in.memory[0x1b10:0x11b10]
	if caveWord(p, 0xff2c) != 0x11b1 {
		return false
	}
	// 規格027的兩個64byte前綴保持原契約；037另驗地圖及完整MPP身份。
	prefix := townBackgroundHash(p[0xc000:0xc040])
	if prefix != "3f5edbc8b51aa017211fb0ca808f807a1ade51d95e14b49dadf593e90ada24eb" && prefix != "6c853b5dab1da5682a959dd804b0f23e5acdc3aa67099596335e1293694f03b2" {
		// 037：新增地圖須同時匹配完整MPP資源；舊兩個來源沿原契約。
		if entry, _ := caveTerrainIdentity(in.memory); entry == nil {
			return false
		}
	}
	for _, guard := range []struct {
		source []byte
		hash   string
	}{
		{p[0x302c:0x4f8c], "018193b8bc08668248cd7b831c9b5be1b26743fadfbf534ea09b7066b7dd5310"},
		{p[0x6f9b:0x7183], "26dedabe338c9c2f6ebaeb8e26f5095486ce3c8863a02316420e72ed3e44aac5"},
		{in.palette, "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c"},
	} {
		if townBackgroundHash(guard.source) != guard.hash {
			return false
		}
	}
	return true
}

func (f *CaveFrames) clear() { f.complete = nil; f.body = nil; f.paints = nil; f.drawing = false }

func (f *CaveFrames) observePaint(o *oracle.Oracle, hook uint16) {
	if f.body == nil {
		return
	}
	position := int(o.DX())
	if hook == 0x3272 {
		position *= 2
	}
	x, y := position%320, position/320
	at := caveWord(f.body.memory[0x1b10:0x11b10], 0x4fed)
	bx, by := at%320, at/320
	if x >= bx+24 || x+8 <= bx || y >= by+24 || y+8 <= by {
		return
	}
	in := f.read(nil, nil)
	if o.DSReg() != 0x1b1 || !caveFrameSource(in) || len(f.paints) >= 64 {
		f.reject(in, "native paint source or limit")
		return
	}
	ax := o.Regs().AX
	p := CavePaint{X: x, Y: y, ID: byte(ax), Hook: hook, Steps: in.steps, Cycles: in.cycles}
	if hook == 0x3599 {
		p.ID, p.Background = byte(ax>>8), byte(ax)
		copy(p.SourceMask[:], o.Bytes(oracle.Far(0x11b1, 0xa000+uint16(p.ID)*8), 8))
	}
	mask, ok := cavePaintMask(p)
	if !ok {
		f.reject(in, "native paint mask")
		return
	}
	p.Mask = mask
	f.paints = append(f.paints, p)
}
func (f *CaveFrames) begin(in caveFrameInput) {
	if !caveFrameSource(in) {
		f.clear()
		return
	}
	f.drawing = true
}
func (f *CaveFrames) reject(in caveFrameInput, reason string) {
	f.refused++
	f.lastRefusal = map[string]any{"steps": in.steps, "cycles": in.cycles, "reason": reason}
	f.lastRefusal["primary_fields"] = caveDiagnosticFields(in.memory)
	if f.body != nil {
		f.lastRefusal["body_steps"] = f.body.steps
		f.lastRefusal["body_primary_fields"] = caveDiagnosticFields(f.body.memory)
		if caveFrameSource(in) && caveFrameSource(*f.body) && len(in.indexed) == 64000 {
			f.rejected = &CaveRejected{Memory: cloneTownBytes(in.memory), BodyMemory: cloneTownBytes(f.body.memory), Indexed: cloneTownBytes(in.indexed), Palette: cloneTownBytes(in.palette), Steps: in.steps, Cycles: in.cycles, BodySteps: f.body.steps, BodyCycles: f.body.cycles, Hook: in.hook, Reason: reason}
			f.rejectedState = map[string]any{"steps": in.steps, "cycles": in.cycles, "body_steps": f.body.steps, "body_cycles": f.body.cycles, "hook": fmt.Sprintf("01B1:%04X", in.hook), "reason": reason, "memory_sha256": townBackgroundHash(in.memory), "body_memory_sha256": townBackgroundHash(f.body.memory), "indexed_sha256": townBackgroundHash(in.indexed), "palette_sha256": townBackgroundHash(in.palette), "body_primary_fields": caveDiagnosticFields(f.body.memory), "scope": "rejected candidate; not presented or current CPU"}
		}
	}
	if f.rejected != nil && f.rejected.Steps == in.steps {
		f.rejected.Paints = append([]CavePaint(nil), f.paints...)
		f.rejectedState["native_paints"] = append([]CavePaint(nil), f.paints...)
	}
	f.clear()
}

func caveDiagnosticFields(memory []byte) map[string]uint8 {
	if len(memory) != 1<<20 {
		return nil
	}
	p := memory[0x1b10:0x11b10]
	fields := map[string]uint8{}
	for _, off := range []int{0x93, 0xe7, 0xe8, 0xff34, 0xff36, 0xff37, 0xff38, 0xff39, 0xff3a, 0xff3d, 0xff3f, 0xff40, 0xff41, 0xff42, 0xff43, 0xff44} {
		fields[fmt.Sprintf("01B1:%04X", off)] = p[off]
	}
	return fields
}

func (f *CaveFrames) RejectedSource() *CaveRejected {
	if f == nil || f.rejected == nil {
		return nil
	}
	r := *f.rejected
	r.Memory = cloneTownBytes(r.Memory)
	r.BodyMemory = cloneTownBytes(r.BodyMemory)
	r.Indexed = cloneTownBytes(r.Indexed)
	r.Palette = cloneTownBytes(r.Palette)
	r.Paints = append([]CavePaint(nil), r.Paints...)
	return &r
}
func (f *CaveFrames) finish(in caveFrameInput) {
	f.drawing = false
	if !caveFrameSource(in) || f.body == nil || !caveFrameSource(*f.body) || f.body.steps > in.steps || f.body.cycles > in.cycles {
		f.reject(in, "source or missing body")
		return
	}
	if caveCatalogMapsDiffer(in.memory, f.body.memory) {
		f.reject(in, "body and complete source belong to different maps")
		return
	}
	if len(in.indexed) != 64000 || len(in.rgb) != 192000 {
		f.reject(in, "frame size")
		return
	}
	for i, v := range in.indexed {
		if !bytes.Equal(in.rgb[i*3:i*3+3], in.palette[int(v)*3:int(v)*3+3]) {
			f.reject(in, "RGB")
			return
		}
	}
	if !cavePaintTimeline(f.paints, f.body.steps, f.body.cycles, in.steps, in.cycles) {
		f.reject(in, "native paint timeline")
		return
	}
	decoded, err := decodeCaveFrameWithPaints(in.memory, f.body.memory, in.indexed, f.paints)
	if err != nil {
		f.reject(in, err.Error())
		return
	}
	f.complete = &CavePresentation{Indexed: cloneTownBytes(in.indexed), RGB: cloneTownBytes(in.rgb), Palette: cloneTownBytes(in.palette), SourceIndexed: cloneTownBytes(in.indexed), Memory: cloneTownBytes(in.memory), BodyMemory: cloneTownBytes(f.body.memory), Steps: in.steps, Cycles: in.cycles, BodySteps: f.body.steps, BodyCycles: f.body.cycles, Decoded: decoded}
	f.captured++
	f.complete.Paints = append([]CavePaint(nil), f.paints...)
	f.lastRefusal = nil
}

func caveFrameEqual(a, b []byte) bool {
	if len(a) != 64000 || len(b) != 64000 {
		return false
	}
	for y := 14; y < 158; y++ {
		if !bytes.Equal(a[y*320+48:y*320+272], b[y*320+48:y*320+272]) {
			return false
		}
	}
	return true
}

// Present保留原版完成來源；畫面區外始終使用當次原版。
func (f *CaveFrames) Present(raw, rgb []byte, blocked bool) *CavePresentation {
	if f == nil {
		return nil
	}
	return f.selectFrame(f.read(raw, rgb), blocked)
}
func (f *CaveFrames) selectFrame(in caveFrameInput, blocked bool) *CavePresentation {
	f.state = map[string]any{"used": false, "drawing": f.drawing, "captured": f.captured, "refused": f.refused, "last_refusal": f.lastRefusal, "last_rejected_source": f.rejectedState, "scope": "完成時刻洞窟遊戲區；區外為當次原版"}
	last := f.complete
	if blocked || !caveFrameSource(in) || len(in.indexed) != 64000 || len(in.rgb) != 192000 {
		f.clear()
		return nil
	}
	if last == nil {
		return nil
	}
	if caveCatalogMapsDiffer(in.memory, last.Memory) {
		f.clear()
		return nil
	}
	if in.steps < last.Steps || in.cycles < last.Cycles {
		f.clear()
		return nil
	}
	if !f.drawing && !caveFrameEqual(in.indexed, last.SourceIndexed) {
		// 028：4129新主體可早於批次開始。撤除舊呈現，保留當次可信主體；
		// 未完成圖面仍不呈現，後續finish須重新核整個原版畫格。
		fresh := f.body != nil && f.body.steps > last.Steps && f.body.cycles >= last.Cycles && f.body.steps <= in.steps && f.body.cycles <= in.cycles && caveFrameSource(*f.body)
		if fresh {
			f.complete = nil
		} else {
			f.clear()
		}
		return nil
	}
	raw, rgb := cloneTownBytes(in.indexed), cloneTownBytes(in.rgb)
	for y := 14; y < 158; y++ {
		a, b := y*320+48, y*320+272
		copy(raw[a:b], last.SourceIndexed[a:b])
		copy(rgb[a*3:b*3], last.RGB[a*3:b*3])
	}
	f.state["used"] = true
	f.state["source_steps"] = last.Steps
	f.state["source_cycles"] = last.Cycles
	f.state["body_steps"] = last.BodySteps
	f.state["body_cycles"] = last.BodyCycles
	f.state["lag_steps"] = in.steps - last.Steps
	f.state["lag_cycles"] = in.cycles - last.Cycles
	f.state["rect"] = cavePresentationBox
	f.state["source_memory_sha256"] = townBackgroundHash(last.Memory)
	f.state["body_memory_sha256"] = townBackgroundHash(last.BodyMemory)
	f.state["source_indexed_sha256"] = townBackgroundHash(last.SourceIndexed)
	f.state["presented_indexed_sha256"] = townBackgroundHash(raw)
	f.state["objects"] = len(last.Decoded.Objects)
	f.state["native_paints"] = append([]CavePaint(nil), last.Paints...)
	p := last.Memory[0x1b10:0x11b10]
	bp := last.BodyMemory[0x1b10:0x11b10]
	actorPosition := caveWord(bp, 0x4fed)
	f.state["source_actor_x"] = actorPosition % 320
	f.state["source_actor_y"] = actorPosition / 320
	f.state["source_map_origin"] = caveWord(p, 0xff31)
	if entry, terrain := caveTerrainIdentity(last.Memory); entry != nil {
		f.state["source_map_key"] = entry.name
		f.state["source_terrain_key"] = terrain.key
	}
	// 深複製來源及圖層，觀測或繪圖端不能回寫保存的完成來源。
	d := *last.Decoded
	d.Background = cloneTownBytes(d.Background)
	d.Dynamic = append([]cavePixel(nil), d.Dynamic...)
	d.Actor = append([]cavePixel(nil), d.Actor...)
	d.Weapon = append([]cavePixel(nil), d.Weapon...)
	d.Cells = append([]caveCell(nil), d.Cells...)
	d.Objects = append([]caveObject(nil), d.Objects...)
	return &CavePresentation{Indexed: raw, RGB: rgb, Palette: cloneTownBytes(last.Palette), SourceIndexed: cloneTownBytes(last.SourceIndexed), Memory: cloneTownBytes(last.Memory), BodyMemory: cloneTownBytes(last.BodyMemory), Steps: last.Steps, Cycles: last.Cycles, BodySteps: last.BodySteps, BodyCycles: last.BodyCycles, Decoded: &d, Paints: append([]CavePaint(nil), last.Paints...)}
}

func (f *CaveFrames) State() map[string]any {
	if f == nil {
		return nil
	}
	return f.state
}

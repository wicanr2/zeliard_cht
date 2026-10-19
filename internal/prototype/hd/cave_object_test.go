package hd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type caveObjectProof struct {
	caveWeaponProof
	Stats struct {
		Known, Unknown int
		Clipped        int `json:"clipped_objects"`
	} `json:"object_stats"`
}

func caveObjectProofs(t *testing.T, root string) []caveObjectProof {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(root, "object-composition-preview-v2/receipt.json"))
	if e != nil {
		t.Fatal(e)
	}
	var j struct{ Records []caveObjectProof }
	if e = json.Unmarshal(b, &j); e != nil || len(j.Records) != 35 {
		t.Fatal("35份獨立物件圖面未齊", e)
	}
	return j.Records
}

func caveObjectFixture(t *testing.T, root string, p caveObjectProof) *CavePresentation {
	t.Helper()
	f := caveWeaponFixture(t, root, p.caveWeaponProof)
	b, e := os.ReadFile(filepath.Join(root, "object-source-v1/receipt.json"))
	if e != nil {
		t.Fatal(e)
	}
	var source struct {
		Sources []struct {
			File          string
			Steps, Cycles uint64
			BodySteps     uint64 `json:"body_steps"`
			BodyCycles    uint64 `json:"body_cycles"`
		}
	}
	if e = json.Unmarshal(b, &source); e != nil {
		t.Fatal(e)
	}
	for _, r := range source.Sources {
		if p.Prefix == "object-source-v1/"+r.File {
			f.Steps, f.Cycles, f.BodySteps, f.BodyCycles = r.Steps, r.Cycles, r.BodySteps, r.BodyCycles
		}
	}
	return f
}

func TestCaveObjectsDirectRegisteredPixels(t *testing.T) {
	root, _ := caveFixtures(t)
	objects, e := NewCaveObjects(root)
	if e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(filepath.Join(root, "object-art-preview-v2/registration.json"))
	if e != nil {
		t.Fatal(e)
	}
	var j struct {
		Poses []struct {
			Key  string
			Path string `json:"sprite_raw_path"`
		}
	}
	if e = json.Unmarshal(b, &j); e != nil || len(j.Poses) != 29 || len(objects.poses) != 29 {
		t.Fatal("29物件圖面未齊", e)
	}
	for _, r := range j.Poses {
		p := objects.poses[r.Key]
		if p == nil {
			t.Fatal("來源鍵不符")
		}
		want, e := os.ReadFile(filepath.Join(root, r.Path))
		if e != nil {
			t.Fatal(e)
		}
		var got []byte
		for _, c := range p.pixels {
			got = append(got, c.R, c.G, c.B, c.A)
		}
		if !bytes.Equal(got, want) {
			t.Fatal("直接高清RGBA與獨立取樣不同", r.Key)
		}
	}
}

func TestCaveObjectsNativeRasterSources(t *testing.T) {
	root, _ := caveFixtures(t)
	b, e := os.ReadFile(filepath.Join(root, "object-source-v1/receipt.json"))
	if e != nil {
		t.Fatal(e)
	}
	var j struct {
		Sources []struct {
			File, Key string
			Record    int
		}
	}
	if e = json.Unmarshal(b, &j); e != nil || len(j.Sources) != 28 {
		t.Fatal("正常原生來源未齊", e)
	}
	for _, source := range j.Sources {
		read := func(ext string) []byte {
			b, e := os.ReadFile(filepath.Join(root, "object-source-v1", source.File+"."+ext))
			if e != nil {
				t.Fatal(e)
			}
			return b
		}
		var paints []CavePaint
		if e = json.Unmarshal(read("paints.json"), &paints); e != nil {
			t.Fatal(e)
		}
		d, e := decodeCaveFrameWithPaints(read("memory"), read("body.memory"), read("indexed"), paints)
		if e != nil {
			t.Fatal(e)
		}
		found := false
		for _, object := range d.Objects {
			if object.Record == source.Record {
				if !bytes.Equal(object.Raster[:], read("object.indexed")) || townBackgroundHash(object.Raster[:]) != source.Key || object.Backgrounds != [4]byte{} {
					t.Fatal("物件256bytes、四底圖或來源鍵不同", source.File)
				}
				found = true
				break
			}
		}
		if !found {
			t.Fatal("遺漏原版物件", source.File)
		}
	}
}

func TestCaveObjectsIndependentCompletedPixels(t *testing.T) {
	root, _ := caveFixtures(t)
	out := os.Getenv("ZELIARD_CAVE_OBJECT_OUTPUT")
	if out == "" {
		t.Skip("須指定新的物件驗證輸出目錄")
	}
	if e := os.Mkdir(filepath.Join(out, "pixels"), 0755); e != nil {
		t.Fatal(e)
	}
	objects, e := NewCaveObjects(root)
	if e != nil {
		t.Fatal(e)
	}
	player, e := NewCavePlayer(root)
	if e != nil {
		t.Fatal(e)
	}
	weapon, e := NewCaveWeapon(root)
	if e != nil {
		t.Fatal(e)
	}
	art, e := NewCaveTerrain(root)
	if e != nil {
		t.Fatal(e)
	}
	baseline, e := NewCaveTerrain(root)
	if e != nil {
		t.Fatal(e)
	}
	art.SetObjects(objects)
	var results []map[string]any
	for _, p := range caveObjectProofs(t, root) {
		t.Run(p.Name, func(t *testing.T) {
			f := caveObjectFixture(t, root, p)
			art.SetPlayer(nil)
			baseline.SetPlayer(nil)
			art.SetWeapon(nil)
			baseline.SetWeapon(nil)
			if p.Player {
				art.SetPlayer(player)
				baseline.SetPlayer(player)
				art.SetWeapon(weapon)
				baseline.SetWeapon(weapon)
			}
			canvas, old := caveOriginalCanvas(f), caveOriginalCanvas(f)
			before := cloneTownBytes(canvas.Pix)
			memory, body, raw := cloneTownBytes(f.Memory), cloneTownBytes(f.BodyMemory), cloneTownBytes(f.SourceIndexed)
			if !art.Draw(canvas, f, image.Point{}) || !baseline.Draw(old, f, image.Point{}) {
				t.Fatal("已驗來源被拒絕")
			}
			state := art.ObjectState()
			if state["known"] != p.Stats.Known || state["unknown"] != p.Stats.Unknown || state["clipped"] != p.Stats.Clipped {
				t.Fatal("物件來源計數不同", state, p.Stats)
			}
			fp, e := os.Open(filepath.Join(root, "object-composition-preview-v2", p.Name+".png"))
			if e != nil {
				t.Fatal(e)
			}
			want, e := png.Decode(fp)
			fp.Close()
			if e != nil {
				t.Fatal(e)
			}
			changed := 0
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := canvas.PixOffset(x, y)
					pt := image.Pt(x, y)
					if pt.In(image.Rect(144, 42, 816, 474)) {
						if canvas.RGBAAt(x, y) != color.RGBAModel.Convert(want.At(x, y)).(color.RGBA) {
							t.Fatalf("獨立完整物件合成不符 %d,%d", x, y)
						}
					} else if !bytes.Equal(canvas.Pix[i:i+4], before[i:i+4]) {
						t.Fatal("改動遊戲區外")
					}
					if !bytes.Equal(canvas.Pix[i:i+4], old.Pix[i:i+4]) {
						changed++
						inside := false
						for _, object := range f.Decoded.Objects {
							_, pose := objects.selectPose(object)
							if pose != nil && pt.In(image.Rect(object.X*3, object.Y*3, (object.X+16)*3, (object.Y+16)*3)) {
								inside = true
							}
						}
						if !inside {
							t.Fatal("物件HD超出原版矩形", x, y)
						}
					}
				}
			}
			if !bytes.Equal(memory, f.Memory) || !bytes.Equal(body, f.BodyMemory) || !bytes.Equal(raw, f.SourceIndexed) {
				t.Fatal("覆繪回寫原版")
			}
			if p.Stats.Known > 0 && changed == 0 {
				t.Fatal("沒有可見物件HD")
			}
			fp, e = os.Create(filepath.Join(out, "pixels", p.Name+".png"))
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(fp, canvas)
			fp.Close()
			if e != nil {
				t.Fatal(e)
			}
			results = append(results, map[string]any{"name": p.Name, "pixels_compared": 672 * 432, "independent_pixels_equal": true, "changed_pixels": changed, "objects": state, "source_unchanged": true})
		})
	}
	if len(results) != 35 {
		t.Fatal("35圖未完成")
	}
	b, e := json.MarshalIndent(map[string]any{"status": "PASS_CAVE_OBJECT_INDEPENDENT_PIXELS", "records": results, "limits": "單元完整圖面；正常原版與實際GUI另驗"}, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(b, 10), 0644); e != nil {
		t.Fatal(e)
	}
}

func TestCaveObjectsFallbackAtomicAndState(t *testing.T) {
	root, _ := caveFixtures(t)
	f := caveObjectFixture(t, root, caveObjectProofs(t, root)[0])
	objects, e := NewCaveObjects(root)
	if e != nil {
		t.Fatal(e)
	}
	art, e := NewCaveTerrain(root)
	if e != nil {
		t.Fatal(e)
	}
	baseline, e := NewCaveTerrain(root)
	if e != nil {
		t.Fatal(e)
	}
	art.SetObjects(objects)
	known := objects.poses
	objects.poses = map[string]*caveObjectPose{}
	a, b := caveOriginalCanvas(f), caveOriginalCanvas(f)
	if !art.Draw(a, f, image.Point{}) || !baseline.Draw(b, f, image.Point{}) || !bytes.Equal(a.Pix, b.Pix) || art.ObjectState()["known"] != 0 || art.ObjectState()["unknown"] != len(f.Decoded.Objects) {
		t.Fatal("未知物件未完整退回原貌")
	}
	objects.poses = known
	object := f.Decoded.Objects[0]
	_, pose := objects.selectPose(object)
	if pose == nil {
		t.Fatal("已驗來源未匹配")
	}
	object.Backgrounds[0] = 1
	if _, p := objects.selectPose(object); p != nil {
		t.Fatal("非零底圖借圖")
	}
	object = f.Decoded.Objects[0]
	object.Raster[0] ^= 1
	if _, p := objects.selectPose(object); p != nil {
		t.Fatal("未知256來源借圖")
	}
	bad := *f
	bad.SourceIndexed = cloneTownBytes(f.SourceIndexed)
	bad.SourceIndexed[100*320+100] ^= 128
	a = caveOriginalCanvas(f)
	before := cloneTownBytes(a.Pix)
	if art.Draw(a, &bad, image.Point{}) || !bytes.Equal(before, a.Pix) {
		t.Fatal("來源不符未原子退回")
	}
	if !art.Draw(caveOriginalCanvas(f), f, image.Point{}) {
		t.Fatal("合法來源未恢復")
	}
	state := art.ObjectState()
	state["source_keys"].([]string)[0] = "changed"
	if art.ObjectState()["source_keys"].([]string)[0] == "changed" {
		t.Fatal("觀察狀態未深複製")
	}
	art.SetObjects(nil)
	if art.ObjectState() != nil {
		t.Fatal("撤除物件狀態殘留")
	}
	for _, missing := range caveObjectAtlases {
		t.Run(missing.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, atlas := range caveObjectAtlases {
				if atlas.name == missing.name {
					continue
				}
				data, e := os.ReadFile(filepath.Join(root, atlas.name))
				if e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(filepath.Join(dir, atlas.name), data, 0644); e != nil {
					t.Fatal(e)
				}
			}
			if group, e := NewCaveObjects(dir); group != nil || e == nil {
				t.Fatal("缺圖未整組拒絕")
			}
			if e = os.WriteFile(filepath.Join(dir, missing.name), []byte("broken"), 0644); e != nil {
				t.Fatal(e)
			}
			if group, e := NewCaveObjects(dir); group != nil || e == nil {
				t.Fatal("壞圖未整組拒絕")
			}
		})
	}
}

// 合成事件驗證順序與裁切；不作遊戲規則或原版時鐘證據。
func TestCaveObjectsPainterOrderBlackAndClipping(t *testing.T) {
	root, _ := caveFixtures(t)
	f := caveObjectFixture(t, root, caveObjectProofs(t, root)[0])
	objects, e := NewCaveObjects(root)
	if e != nil {
		t.Fatal(e)
	}
	known := f.Decoded.Objects[0]
	known.X, known.Y = 100, 80
	unknown := known
	for i := range unknown.Raster {
		unknown.Raster[i] = 7
	}
	unknown.X += 8
	d := &caveDecoded{Objects: []caveObject{known, unknown}}
	canvas := image.NewRGBA(image.Rect(0, 0, 1000, 700))
	origin := image.Pt(11, 17)
	objects.draw(canvas, d, f.Palette, origin)
	c := color.RGBA{f.Palette[21], f.Palette[22], f.Palette[23], 255}
	if canvas.RGBAAt((unknown.X+1)*3+origin.X, (unknown.Y+1)*3+origin.Y) != c {
		t.Fatal("後畫未知來源未覆蓋先畫HD")
	}
	d.Objects = []caveObject{unknown, known}
	objects.draw(canvas, d, f.Palette, origin)
	black := color.RGBA{f.Palette[0], f.Palette[1], f.Palette[2], 255}
	if canvas.RGBAAt(known.X*3+origin.X, known.Y*3+origin.Y) != black {
		t.Fatal("已知來源未保留原版清黑")
	}
	for _, point := range []image.Point{{40, 6}, {264, 6}, {40, 150}, {264, 150}} {
		t.Run(fmt.Sprint(point), func(t *testing.T) {
			object := known
			object.X, object.Y = point.X, point.Y
			dst := image.NewRGBA(image.Rect(0, 0, 1000, 700))
			for i := range dst.Pix {
				dst.Pix[i] = 93
			}
			before := cloneTownBytes(dst.Pix)
			state := objects.draw(dst, &caveDecoded{Objects: []caveObject{object}}, f.Palette, origin)
			if state["known"] != 1 || state["clipped"] != 1 {
				t.Fatal("部分出界來源錯誤拒絕")
			}
			box := image.Rect(144, 42, 816, 474).Add(origin)
			for y := 0; y < 700; y++ {
				for x := 0; x < 1000; x++ {
					i := dst.PixOffset(x, y)
					if !image.Pt(x, y).In(box) && !bytes.Equal(dst.Pix[i:i+4], before[i:i+4]) {
						t.Fatal("部分出界改動遊戲區外", x, y)
					}
				}
			}
		})
	}
	empty := objects.draw(canvas, &caveDecoded{}, f.Palette, origin)
	if empty["empty"] != true || empty["matched"] != false || empty["known"] != 0 {
		t.Fatal("空白仍沿用舊來源")
	}
}

type caveObjectCoverage struct {
	objects                        *CaveObjects
	frames                         map[uint64]bool
	keys                           map[string]int
	dumped                         map[string]bool
	known, unknown, empty, clipped int
	sources                        []map[string]any
	capture                        map[uint64]bool
	order                          []uint64
	captures                       []map[string]any
}

func newCaveObjectCoverage(objects *CaveObjects) *caveObjectCoverage {
	c := &caveObjectCoverage{objects: objects, frames: map[uint64]bool{}, keys: map[string]int{}, dumped: map[string]bool{}, capture: map[uint64]bool{}}
	for _, s := range strings.Split(os.Getenv("ZELIARD_CAVE_OBJECT_CAPTURE_STEPS"), ",") {
		if s == "" {
			continue
		}
		v, e := strconv.ParseUint(s, 10, 64)
		if e != nil {
			panic(e)
		}
		if !c.capture[v] {
			c.capture[v] = true
			c.order = append(c.order, v)
		}
	}
	return c
}
func (c *caveObjectCoverage) observe(t *testing.T, out string, f *CavePresentation) {
	t.Helper()
	if c.objects == nil || f == nil {
		return
	}
	write := func(stem string) {
		for ext, data := range map[string][]byte{"memory": f.Memory, "body.memory": f.BodyMemory, "indexed": f.SourceIndexed, "palette": f.Palette} {
			if e := os.WriteFile(filepath.Join(out, stem+"."+ext), data, 0644); e != nil {
				t.Fatal(e)
			}
		}
		b, e := json.Marshal(f.Paints)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(out, stem+".paints.json"), b, 0644); e != nil {
			t.Fatal(e)
		}
	}
	if c.capture[f.Steps] {
		stem := fmt.Sprintf("object-observed-%02d", len(c.captures))
		write(stem)
		c.captures = append(c.captures, map[string]any{"file": stem, "steps": f.Steps, "cycles": f.Cycles, "body_steps": f.BodySteps, "body_cycles": f.BodyCycles, "memory_sha256": townBackgroundHash(f.Memory), "body_memory_sha256": townBackgroundHash(f.BodyMemory), "indexed_sha256": townBackgroundHash(f.SourceIndexed), "palette_sha256": townBackgroundHash(f.Palette)})
		c.capture[f.Steps] = false
	}
	if c.frames[f.Steps] {
		return
	}
	c.frames[f.Steps] = true
	if len(f.Decoded.Objects) == 0 {
		c.empty++
	}
	for _, object := range f.Decoded.Objects {
		key, pose := c.objects.selectPose(object)
		c.keys[key]++
		if !image.Rect(object.X, object.Y, object.X+16, object.Y+16).In(image.Rect(48, 14, 272, 158)) {
			c.clipped++
		}
		if pose != nil {
			c.known++
			continue
		}
		c.unknown++
		if c.dumped[key] {
			continue
		}
		c.dumped[key] = true
		stem := fmt.Sprintf("object-missing-%02d", len(c.sources))
		write(stem)
		if e := os.WriteFile(filepath.Join(out, stem+".object.indexed"), object.Raster[:], 0644); e != nil {
			t.Fatal(e)
		}
		c.sources = append(c.sources, map[string]any{"key": key, "file": stem, "steps": f.Steps, "cycles": f.Cycles, "body_steps": f.BodySteps, "body_cycles": f.BodyCycles, "x": object.X, "y": object.Y, "record": object.Record, "template": object.Template, "backgrounds": object.Backgrounds, "memory_sha256": townBackgroundHash(f.Memory), "body_memory_sha256": townBackgroundHash(f.BodyMemory), "indexed_sha256": townBackgroundHash(f.SourceIndexed)})
	}
}
func (c *caveObjectCoverage) summary() map[string]any {
	var missing []uint64
	for _, step := range c.order {
		if c.capture[step] {
			missing = append(missing, step)
		}
	}
	return map[string]any{"source_frames": len(c.frames), "known_instances": c.known, "unknown_instances": c.unknown, "empty_frames": c.empty, "clipped_instances": c.clipped, "keys": c.keys, "unknown_sources": c.sources, "captured_sources": c.captures, "uncaptured_steps": missing, "scope": "有限選定完成來源；原版配色及未見來源仍原貌，不稱全遊戲"}
}

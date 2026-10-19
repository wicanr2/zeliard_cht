package hd

import (
	"bytes"
	"encoding/json"
	"image"
	"os"
	"path/filepath"
	"testing"
)

type caveFixturePoint struct {
	Name, Hook    string
	Steps, Cycles uint64
	MemorySHA     string `json:"memory_sha256"`
	Body          struct {
		Path          string
		Steps, Cycles uint64
		MemorySHA     string `json:"memory_sha256"`
	}
}

func caveFixtures(t *testing.T) (string, []caveFixturePoint) {
	t.Helper()
	root := os.Getenv("ZELIARD_CAVE_STUDY")
	if root == "" {
		t.Skip("須掛載已驗洞窟來源；skip不算驗收")
	}
	b, e := os.ReadFile(filepath.Join(root, "full-frame-v2/completed-receipt.json"))
	if e != nil {
		t.Fatal(e)
	}
	var r struct{ Checkpoints []caveFixturePoint }
	if e = json.Unmarshal(b, &r); e != nil {
		t.Fatal(e)
	}
	var points []caveFixturePoint
	for _, p := range r.Checkpoints {
		if p.Hook == "7114" || p.Hook == "714C" {
			points = append(points, p)
		}
	}
	if len(points) != 38 {
		t.Fatal("完成來源數不符", len(points))
	}
	return root, points
}
func caveFixture(t *testing.T, root string, p caveFixturePoint) (caveFrameInput, caveFrameInput) {
	t.Helper()
	base := filepath.Join(root, "full-frame-v2")
	read := func(name string) []byte {
		b, e := os.ReadFile(filepath.Join(base, name))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	m, bm := read(p.Name+".memory"), read(p.Body.Path)
	if townBackgroundHash(m) != p.MemorySHA || townBackgroundHash(bm) != p.Body.MemorySHA {
		t.Fatal("來源雜湊不符", p.Name)
	}
	in := caveFrameInput{mode: 0x13, memory: m, indexed: read(p.Name + ".indexed"), rgb: read(p.Name + ".rgb"), palette: read(p.Name + ".palette"), steps: p.Steps, cycles: p.Cycles}
	body := caveFrameInput{mode: 0x13, memory: bm, palette: cloneTownBytes(in.palette), steps: p.Body.Steps, cycles: p.Body.Cycles}
	return in, body
}
func cloneCaveInput(in caveFrameInput) caveFrameInput {
	in.memory = cloneTownBytes(in.memory)
	in.indexed = cloneTownBytes(in.indexed)
	in.rgb = cloneTownBytes(in.rgb)
	in.palette = cloneTownBytes(in.palette)
	return in
}

func TestCaveSourcesNormalCompletedFrames(t *testing.T) {
	root, points := caveFixtures(t)
	data, e := os.ReadFile(filepath.Join(root, "lifecycle-map-proof-v1.json"))
	if e != nil {
		t.Fatal(e)
	}
	var proof struct {
		Records []struct {
			Name        string
			DecodedSHA  string `json:"decoded_sha256"`
			Differences int
		}
	}
	if e = json.Unmarshal(data, &proof); e != nil {
		t.Fatal(e)
	}
	known := map[string]string{}
	for _, x := range proof.Records {
		if x.Differences != 0 {
			t.Fatal("獨立Python重建未通過")
		}
		known[x.Name] = x.DecodedSHA
	}
	for _, point := range points {
		t.Run(point.Name, func(t *testing.T) {
			in, body := caveFixture(t, root, point)
			before := cloneCaveInput(in)
			bb := cloneCaveInput(body)
			if !caveFrameSource(in) || !caveFrameSource(body) {
				t.Fatal("來源守衛失配")
			}
			d, e := decodeCaveFrame(in.memory, body.memory, in.indexed)
			if e != nil {
				t.Fatal(e)
			}
			var region []byte
			for y := 14; y < 158; y++ {
				region = append(region, d.Background[y*320+48:y*320+272]...)
			}
			if townBackgroundHash(region) != known[point.Name] || len(d.Cells) != 504 {
				t.Fatal("背景與獨立Python重建不同")
			}
			if !bytes.Equal(in.memory, before.memory) || !bytes.Equal(body.memory, bb.memory) || !bytes.Equal(in.indexed, before.indexed) {
				t.Fatal("解碼回寫來源")
			}
			f := &CaveFrames{body: &body}
			f.finish(in)
			if f.complete == nil || f.captured != 1 || f.refused != 0 {
				t.Fatal("完成來源被拒絕", f.lastRefusal)
			}
			if f.complete.Steps != in.steps || f.complete.BodySteps != body.steps {
				t.Fatal("來源時刻遺失")
			}
		})
	}
}

func TestCaveSourcesObservedFF39Branch(t *testing.T) {
	root, _ := caveFixtures(t)
	read := func(suffix string) []byte {
		b, err := os.ReadFile(filepath.Join(root, "input-replay-v4", "source-00."+suffix))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	memory, bodyMemory, raw, palette := read("memory"), read("body.memory"), read("indexed"), read("palette")
	if townBackgroundHash(memory) != "cb9e15dbbdd125c15dae7a49391b7dd6cd8c9c286d5776b8164d8a93be6bc446" || townBackgroundHash(bodyMemory) != "30c6c355bf7d64668eacc974a6603f6195201a8f1d809fbce4c0b5a9da471a40" {
		t.Fatal("FF39真實來源不符")
	}
	before, beforeBody := cloneTownBytes(memory), cloneTownBytes(bodyMemory)
	d, err := decodeCaveFrame(memory, bodyMemory, raw)
	if err != nil || d.Differences != 0 || len(d.Actor) != 171 {
		t.Fatal("與獨立Python／原版不符", err)
	}
	rgb := make([]byte, len(raw)*3)
	for i, value := range raw {
		copy(rgb[i*3:], palette[int(value)*3:int(value)*3+3])
	}
	// 此單元核圖層與呈現；未記錄的主角週期不推算，兩側用測試時鐘0。
	in := caveFrameInput{mode: 0x13, memory: memory, indexed: raw, rgb: rgb, palette: palette, steps: 161647511, hook: 0x7114}
	body := caveFrameInput{mode: 0x13, memory: bodyMemory, palette: palette, steps: 161635874}
	f := &CaveFrames{body: &body}
	f.finish(in)
	if f.selectFrame(in, false) == nil || f.captured != 1 || f.refused != 0 {
		t.Fatal("已驗FF39來源未呈現", f.lastRefusal)
	}
	for _, value := range []byte{1, 127, 254} {
		unknown := cloneTownBytes(bodyMemory)
		unknown[0x1b10+0xff39] = value
		if _, err := decodeCaveFrame(memory, unknown, raw); err == nil {
			t.Fatal("接受未驗FF39", value)
		}
	}
	if !bytes.Equal(memory, before) || !bytes.Equal(bodyMemory, beforeBody) {
		t.Fatal("解碼改寫來源")
	}
}

func TestCaveObservedNativePaints(t *testing.T) {
	root, _ := caveFixtures(t)
	data, err := os.ReadFile(filepath.Join(root, "native-paints-proof-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var proof struct {
		Records []struct {
			Source      string
			Steps       uint64
			MemorySHA   string      `json:"memory_sha256"`
			BodySHA     string      `json:"body_memory_sha256"`
			Paints      []CavePaint `json:"native_paints"`
			ActorPixels int         `json:"remaining_actor_pixels"`
		}
	}
	if err = json.Unmarshal(data, &proof); err != nil {
		t.Fatal(err)
	}
	if len(proof.Records) != 3 {
		t.Fatal("獨立增量來源數量不符")
	}
	art, err := NewCaveTerrain(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, point := range proof.Records {
		t.Run(point.Source, func(t *testing.T) {
			read := func(suffix string) []byte {
				b, e := os.ReadFile(filepath.Join(root, point.Source+"."+suffix))
				if e != nil {
					t.Fatal(e)
				}
				return b
			}
			memory, bodyMemory, raw, palette := read("memory"), read("body.memory"), read("indexed"), read("palette")
			if townBackgroundHash(memory) != point.MemorySHA || townBackgroundHash(bodyMemory) != point.BodySHA {
				t.Fatal("增量來源SHA不符")
			}
			if _, e := decodeCaveFrame(memory, bodyMemory, raw); e == nil {
				t.Fatal("未套增量覆寫應重現原失配")
			}
			rgb := make([]byte, len(raw)*3)
			for i, v := range raw {
				copy(rgb[i*3:], palette[int(v)*3:int(v)*3+3])
			}
			// 單元以0作主角測試時鐘，真實時鐘另由正常原版配對核對。
			endCycle := point.Paints[len(point.Paints)-1].Cycles
			in := caveFrameInput{mode: 0x13, memory: memory, indexed: raw, rgb: rgb, palette: palette, steps: point.Steps, cycles: endCycle}
			body := caveFrameInput{mode: 0x13, memory: bodyMemory, palette: palette}
			f := &CaveFrames{body: &body, paints: append([]CavePaint(nil), point.Paints...)}
			f.finish(in)
			got := f.selectFrame(in, false)
			if got == nil || got.Decoded.Differences != 0 || len(got.Decoded.Actor) != point.ActorPixels {
				t.Fatal("與Python原版全區重建不同", f.lastRefusal)
			}
			canvas := caveOriginalCanvas(got)
			if !art.Draw(canvas, got, image.Point{}) {
				t.Fatal("增量HD未繪製")
			}
			got.Paints[0].Mask ^= 1
			if f.complete.Paints[0] != point.Paints[0] {
				t.Fatal("呈現輸出回寫增量紀錄")
			}
			for name, change := range map[string]func(*CavePresentation){
				"mask":   func(p *CavePresentation) { p.Paints[0].Mask ^= 1 },
				"hook":   func(p *CavePresentation) { p.Paints[0].Hook = 0 },
				"bounds": func(p *CavePresentation) { p.Paints[0].X = 0 },
				"future": func(p *CavePresentation) { p.Paints[0].Steps = p.Steps + 1 },
				"limit": func(p *CavePresentation) {
					for len(p.Paints) < 65 {
						p.Paints = append(p.Paints, p.Paints[0])
					}
				},
			} {
				t.Run(name, func(t *testing.T) {
					bad := f.selectFrame(in, false)
					change(bad)
					dst := caveOriginalCanvas(bad)
					before := cloneTownBytes(dst.Pix)
					if art.Draw(dst, bad, image.Point{}) || !bytes.Equal(before, dst.Pix) {
						t.Fatal("增量失配未原子拒絕")
					}
				})
			}
			if townBackgroundHash(memory) != point.MemorySHA || townBackgroundHash(bodyMemory) != point.BodySHA {
				t.Fatal("覆繪改寫來源")
			}
		})
	}
}

func TestCaveNativeMaskedPaint(t *testing.T) {
	p := CavePaint{Hook: 0x3599, Background: 1, SourceMask: [8]byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe}}
	mask, ok := cavePaintMask(p)
	if !ok || mask != (uint64(1)|(uint64(1)<<63)) {
		t.Fatal("動態遮罩方向不符", mask)
	}
	p.Background = 0
	mask, ok = cavePaintMask(p)
	if !ok || mask != ^uint64(0) {
		t.Fatal("零底圖應整格寫入")
	}
	p.Background = 128
	if _, ok = cavePaintMask(p); ok {
		t.Fatal("接受未知底圖")
	}
}

func TestCaveSourcesObservedE8AndFF3ABranches(t *testing.T) {
	root, _ := caveFixtures(t)
	art, err := NewCaveTerrain(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, prefix, memorySHA, bodySHA string
		actorPixels                      int
	}{
		{"E8", "combat-replay-v5/source-00", "1156d4f08451c71427a15d17428404bd9ca30c3f349a778f48c1d79e30904a95", "e980cc99817c84b3267968b339c744679c560a3622160833e19a65cb5316fefb", 194},
		{"FF3A", "return-replay-v5/source-01", "eb53e8e1eb37da5783d36a31e6c744e3d8d1f40e668963f701cb3f64de124345", "ca4372cc82590b10fb4414c119c7d65b8c732336cec0e49b55547f400553c51c", 240},
	} {
		t.Run(tc.name, func(t *testing.T) {
			read := func(suffix string) []byte {
				b, err := os.ReadFile(filepath.Join(root, tc.prefix+"."+suffix))
				if err != nil {
					t.Fatal(err)
				}
				return b
			}
			memory, bodyMemory, raw, palette := read("memory"), read("body.memory"), read("indexed"), read("palette")
			if townBackgroundHash(memory) != tc.memorySHA || townBackgroundHash(bodyMemory) != tc.bodySHA {
				t.Fatal("真實分支來源不符")
			}
			rgb := make([]byte, len(raw)*3)
			for i, v := range raw {
				copy(rgb[i*3:], palette[int(v)*3:int(v)*3+3])
			}
			in := caveFrameInput{mode: 0x13, memory: memory, indexed: raw, rgb: rgb, palette: palette}
			body := caveFrameInput{mode: 0x13, memory: bodyMemory, palette: palette}
			f := &CaveFrames{body: &body}
			f.finish(in)
			got := f.selectFrame(in, false)
			if got == nil || got.Decoded.Differences != 0 || len(got.Decoded.Actor) != tc.actorPixels {
				t.Fatal("獨立Python／原版全區重建不符", f.lastRefusal)
			}
			canvas := caveOriginalCanvas(got)
			if !art.Draw(canvas, got, image.Point{}) {
				t.Fatal("已驗分支HD未繪製")
			}
			for _, px := range got.Decoded.Actor {
				for y := 0; y < 3; y++ {
					for x := 0; x < 3; x++ {
						c := canvas.RGBAAt(px.X*3+x, px.Y*3+y)
						if !bytes.Equal([]byte{c.R, c.G, c.B}, palette[int(px.Color)*3:int(px.Color)*3+3]) {
							t.Fatal("地形蓋住原貌主角")
						}
					}
				}
			}
			unknown := cloneTownBytes(bodyMemory)
			unknown[0x1b10+0xff39] = 255
			if _, err := decodeCaveFrame(memory, unknown, raw); err == nil {
				t.Fatal("接受未驗混合分支")
			}
			if townBackgroundHash(memory) != tc.memorySHA || townBackgroundHash(bodyMemory) != tc.bodySHA {
				t.Fatal("改寫原版來源")
			}
		})
	}
}

func TestCaveFramesPartialAndOwnership(t *testing.T) {
	root, points := caveFixtures(t)
	in, body := caveFixture(t, root, points[20])
	f := &CaveFrames{body: &body}
	f.begin(in)
	if f.selectFrame(in, false) != nil || f.body == nil || !f.drawing {
		t.Fatal("第一個完成前不應遺失主角來源")
	}
	f.finish(in)
	if f.complete == nil {
		t.Fatal("完成來源被拒絕", f.lastRefusal)
	}
	original := cloneCaveInput(in)
	live := cloneCaveInput(in)
	live.steps++
	live.cycles++
	f.begin(live)
	live.indexed[100*320+100] ^= 1
	live.rgb[(100*320+100)*3] ^= 1
	live.indexed[190*320+60] ^= 1
	live.rgb[(190*320+60)*3] ^= 1
	got := f.selectFrame(live, false)
	if got == nil {
		t.Fatal("未保留批次中的完成來源")
	}
	if !caveFrameEqual(got.Indexed, original.indexed) || got.Indexed[190*320+60] != live.indexed[190*320+60] || got.RGB[(190*320+60)*3] != live.rgb[(190*320+60)*3] {
		t.Fatal("遊戲區或區外來源錯誤")
	}
	if f.State()["lag_steps"] != uint64(1) {
		t.Fatal("來源時間差錯誤")
	}
	position := caveWord(body.memory[0x1b10:0x11b10], 0x4fed)
	if f.State()["source_actor_x"] != position%320 || f.State()["source_actor_y"] != position/320 {
		t.Fatal("觀測位置不是主角來源時刻")
	}
	got.Memory[0] ^= 1
	got.BodyMemory[0] ^= 1
	got.Indexed[0] ^= 1
	got.SourceIndexed[0] ^= 1
	got.Decoded.Background[0] ^= 1
	got.Decoded.Actor[0].Color ^= 1
	if !bytes.Equal(f.complete.Memory, original.memory) || !bytes.Equal(f.complete.BodyMemory, body.memory) || !bytes.Equal(f.complete.Indexed, original.indexed) || !bytes.Equal(f.complete.SourceIndexed, original.indexed) {
		t.Fatal("輸出改寫內部快照")
	}
	next := f.selectFrame(live, false)
	if next == nil || !bytes.Equal(next.Decoded.Background, f.complete.Decoded.Background) || next.Decoded.Actor[0] != f.complete.Decoded.Actor[0] {
		t.Fatal("圖層輸出改寫快照")
	}
	in.memory[0] ^= 1
	body.memory[0] ^= 1
	if !bytes.Equal(f.complete.Memory, original.memory) || bytes.Equal(f.complete.BodyMemory, body.memory) {
		t.Fatal("完成來源沒有深複製")
	}
}

func TestCaveFramesRejectSourceAndTransitions(t *testing.T) {
	root, points := caveFixtures(t)
	source, body := caveFixture(t, root, points[20])
	cases := map[string]func(*caveFrameInput){
		"mode":                func(in *caveFrameInput) { in.mode = 3 },
		"memory-size":         func(in *caveFrameInput) { in.memory = in.memory[:40000] },
		"palette":             func(in *caveFrameInput) { in.palette[3] ^= 1 },
		"renderer":            func(in *caveFrameInput) { in.memory[0x1b10+0x302c] ^= 1 },
		"loop":                func(in *caveFrameInput) { in.memory[0x1b10+0x710f] ^= 1 },
		"module":              func(in *caveFrameInput) { in.memory[0x1b10+0xc000] ^= 1 },
		"indexed-size":        func(in *caveFrameInput) { in.indexed = in.indexed[:500] },
		"rgb-size":            func(in *caveFrameInput) { in.rgb = in.rgb[:500] },
		"steps-backward":      func(in *caveFrameInput) { in.steps-- },
		"cycles-backward":     func(in *caveFrameInput) { in.cycles-- },
		"outside-batch-pixel": func(in *caveFrameInput) { in.indexed[100*320+100] ^= 1 },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			in := cloneCaveInput(source)
			f := &CaveFrames{body: &body}
			f.finish(in)
			if f.complete == nil {
				t.Fatal(f.lastRefusal)
			}
			change(&in)
			before := cloneCaveInput(in)
			if f.selectFrame(in, false) != nil || f.complete != nil || f.body != nil || f.drawing {
				t.Fatal("失配沒有撤除")
			}
			if !bytes.Equal(in.memory, before.memory) || !bytes.Equal(in.indexed, before.indexed) {
				t.Fatal("拒絕回寫來源")
			}
		})
	}
	f := &CaveFrames{body: &body}
	f.finish(source)
	if f.selectFrame(source, true) != nil || f.complete != nil {
		t.Fatal("模態未撤除")
	}
	var absent *CaveFrames
	if absent.State() != nil || absent.Present(nil, nil, false) != nil || NewCaveFrames(nil) != nil {
		t.Fatal("缺少機器未安全退回")
	}
}

func TestCaveSourcesRejectUnknownOrCorrupt(t *testing.T) {
	root, points := caveFixtures(t)
	source, body := caveFixture(t, root, points[20])
	for name, change := range map[string]func(*caveFrameInput, *caveFrameInput){
		"unknown-equipment": func(in, b *caveFrameInput) { b.memory[0x1b10+0x93] = 1 },
		"unknown-magic":     func(in, b *caveFrameInput) { b.memory[0x1b10+0xff41] = 1 },
		"map-bounds":        func(in, b *caveFrameInput) { in.memory[0x1b10+0xff31] = 0; in.memory[0x1b10+0xff32] = 0 },
		"body-outside":      func(in, b *caveFrameInput) { b.memory[0x1b10+0x4fed] = 0; b.memory[0x1b10+0x4fee] = 0 },
		"pixel-mismatch":    func(in, b *caveFrameInput) { in.indexed[100*320+100] ^= 1 },
	} {
		t.Run(name, func(t *testing.T) {
			in, b := cloneCaveInput(source), cloneCaveInput(body)
			change(&in, &b)
			if _, e := decodeCaveFrame(in.memory, b.memory, in.indexed); e == nil {
				t.Fatal("未知或損壞來源被接受")
			}
		})
	}
	for _, name := range []string{"missing-body", "rgb"} {
		t.Run(name, func(t *testing.T) {
			in := cloneCaveInput(source)
			f := &CaveFrames{}
			if name == "rgb" {
				f.body = &body
				in.rgb[(100*320+100)*3] ^= 1
			}
			f.finish(in)
			if f.complete != nil || f.refused != 1 {
				t.Fatal("不完整來源沒有原子拒絕")
			}
		})
	}
}

func TestCaveRejectedDiagnosticsStaySeparate(t *testing.T) {
	root, points := caveFixtures(t)
	in, body := caveFixture(t, root, points[20])
	before := cloneCaveInput(in)
	body.memory[0x1b10+0xe8] = 1
	f := &CaveFrames{body: &body}
	f.finish(in)
	if f.complete != nil || f.refused != 1 {
		t.Fatal("診斷改變拒絕條件")
	}
	r := f.RejectedSource()
	if r == nil || r.Reason != "unobserved actor profile" || r.Steps != in.steps || r.BodySteps != body.steps {
		t.Fatal("沒有保留拒絕候選")
	}
	if r.BodyMemory[0x1b10+0xe8] != 1 || !bytes.Equal(r.Memory, in.memory) || !bytes.Equal(r.Indexed, in.indexed) {
		t.Fatal("診斷來源錯誤")
	}
	if f.lastRefusal["body_primary_fields"].(map[string]uint8)["01B1:00E8"] != 1 {
		t.Fatal("未保存原始偏移與bytes")
	}
	r.BodyMemory[0] ^= 1
	r.Memory[0] ^= 1
	if !bytes.Equal(f.RejectedSource().Memory, in.memory) || f.RejectedSource().BodyMemory[0] != body.memory[0] {
		t.Fatal("診斷輸出可改寫快照")
	}
	if !bytes.Equal(in.memory, before.memory) {
		t.Fatal("診斷回寫原版")
	}
	var absent *CaveFrames
	if absent.RejectedSource() != nil {
		t.Fatal("缺少來源未安全退回")
	}
}

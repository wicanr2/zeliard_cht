package hd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
)

const caveTransitionKey = "a119ce2b9025164280afcca26ca61fb31cf53bfc5760c46c5ca570b1a8cfac94"

type caveTransitionSource struct {
	File          string
	Steps, Cycles uint64
	BodySteps     uint64 `json:"body_steps"`
	BodyCycles    uint64 `json:"body_cycles"`
	MemorySHA     string `json:"memory_sha256"`
	BodySHA       string `json:"body_memory_sha256"`
	RawSHA        string `json:"indexed_sha256"`
	PaletteSHA    string `json:"palette_sha256"`
	Paints        []CavePaint
}

func caveTransitionSources(t *testing.T) []caveTransitionSource {
	t.Helper()
	study := caveShieldStudy(t)
	var r struct {
		Sources []caveTransitionSource `json:"native_special_sources"`
	}
	if e := json.Unmarshal(caveShieldRead(t, filepath.Join(study, "shield-transition-source-v2/extra-receipt.json")), &r); e != nil {
		t.Fatal(e)
	}
	if len(r.Sources) != 4 {
		t.Fatal("雙朝向四份原生完成來源未齊")
	}
	return r.Sources
}

func caveTransitionFixture(t *testing.T, r caveTransitionSource) *CavePresentation {
	t.Helper()
	base := filepath.Join(caveShieldStudy(t), "shield-transition-source-v2", r.File)
	m, b, raw, pal := caveShieldRead(t, base+".memory"), caveShieldRead(t, base+".body.memory"), caveShieldRead(t, base+".indexed"), caveShieldRead(t, base+".palette")
	if townBackgroundHash(m) != r.MemorySHA || townBackgroundHash(b) != r.BodySHA || townBackgroundHash(raw) != r.RawSHA || townBackgroundHash(pal) != r.PaletteSHA {
		t.Fatal("028來源身份不符", r.File)
	}
	d, e := decodeCaveFrameWithPaints(m, b, raw, r.Paints)
	if e != nil {
		t.Fatal(e)
	}
	rgb := make([]byte, 192000)
	for i, v := range raw {
		copy(rgb[i*3:], pal[int(v)*3:int(v)*3+3])
	}
	return &CavePresentation{Memory: m, BodyMemory: b, SourceIndexed: raw, Indexed: cloneTownBytes(raw), RGB: rgb, Palette: pal, Decoded: d, Paints: r.Paints, Steps: r.Steps, Cycles: r.Cycles, BodySteps: r.BodySteps, BodyCycles: r.BodyCycles}
}

func TestCaveShieldTransitionIndependentPixels(t *testing.T) {
	study := caveShieldStudy(t)
	art := caveShieldArt(t)
	for _, r := range caveTransitionSources(t) {
		t.Run(r.File, func(t *testing.T) {
			f := caveTransitionFixture(t, r)
			if townBackgroundHash(f.Decoded.ActorRaster[:]) != caveTransitionKey || f.Decoded.Differences != 0 {
				t.Fatal("028完整角色或原圖不同")
			}
			p := art.player.poses[caveTransitionKey]
			if p == nil || p.number != 54 || len(art.player.poses) != 55 {
				t.Fatal("離洞姿態未登錄")
			}
			var sprite []byte
			for _, c := range p.pixels {
				sprite = append(sprite, c.R, c.G, c.B, c.A)
			}
			if !bytes.Equal(sprite, caveShieldRead(t, filepath.Join(study, "shield-transition-art-v1/pose-54.rgba"))) {
				t.Fatal("離洞直接高清取樣不同")
			}
			canvas := caveOriginalCanvas(f)
			beforeM, beforeB, beforeR := cloneTownBytes(f.Memory), cloneTownBytes(f.BodyMemory), cloneTownBytes(f.SourceIndexed)
			if !art.Draw(canvas, f, image.Point{}) || art.PlayerState()["pose"] != 54 || art.PlayerState()["matched"] != true {
				t.Fatal("離洞HD未匹配", art.PlayerState())
			}
			fp, e := os.Open(filepath.Join(study, "shield-transition-composition-v1", r.File+"-player-HD.png"))
			if e != nil {
				t.Fatal(e)
			}
			expected, e := png.Decode(fp)
			fp.Close()
			if e != nil {
				t.Fatal(e)
			}
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					if canvas.RGBAAt(x, y) != color.RGBAModel.Convert(expected.At(x, y)).(color.RGBA) {
						t.Fatalf("離洞整幅高清不同 %d,%d", x, y)
					}
				}
			}
			if !bytes.Equal(beforeM, f.Memory) || !bytes.Equal(beforeB, f.BodyMemory) || !bytes.Equal(beforeR, f.SourceIndexed) {
				t.Fatal("離洞HD回寫原版")
			}
		})
	}
}

// 真實來源的呈現器單元邊界；原版正常執行由下一測試獨立驗證。
func TestCaveShieldTransitionRetainsActualBody(t *testing.T) {
	rows := caveTransitionSources(t)
	for _, n := range []int{0, 2} {
		a, b := caveTransitionFixture(t, rows[n]), caveTransitionFixture(t, rows[n+1])
		if a.BodySteps != b.BodySteps || !bytes.Equal(a.BodyMemory, b.BodyMemory) {
			t.Fatal("原版未重用相同主體")
		}
		body := &caveFrameInput{mode: 0x13, memory: a.BodyMemory, palette: a.Palette, steps: a.BodySteps, cycles: a.BodyCycles}
		frames := &CaveFrames{body: body}
		for _, f := range []*CavePresentation{a, b} {
			frames.finish(caveFrameInput{mode: 0x13, memory: f.Memory, indexed: f.SourceIndexed, rgb: f.RGB, palette: f.Palette, steps: f.Steps, cycles: f.Cycles})
			if frames.complete == nil || frames.body != body || frames.refused != 0 {
				t.Fatal("已驗離洞來源丟失真實主體")
			}
		}
	}
}

func TestCaveShieldTransitionRejectsUnobservedCombinations(t *testing.T) {
	f := caveTransitionFixture(t, caveTransitionSources(t)[0])
	for _, off := range []int{0xe8, 0xff39, 0xff41, 0xff38, 0xff40} {
		for _, v := range []byte{1, 255} {
			b := cloneTownBytes(f.BodyMemory)
			b[0x1b10+off] = v
			if _, e := decodeCaveFrameWithPaints(f.Memory, b, f.SourceIndexed, f.Paints); e == nil {
				t.Fatal("未驗組合被接納", off, v)
			}
		}
	}
	art := caveShieldArt(t)
	delete(art.player.poses, caveTransitionKey)
	canvas := caveOriginalCanvas(f)
	if !art.Draw(canvas, f, image.Point{}) || art.PlayerState()["matched"] != false {
		t.Fatal("離洞缺圖未保留原貌")
	}
}

type caveTransitionInput struct {
	Kind         string
	Scan         uint8
	Pressed      bool
	Budget       uint64      `json:"cycle_budget"`
	Instructions uint64      `json:"instruction_budget"`
	BeforeSteps  uint64      `json:"before_steps"`
	AfterSteps   uint64      `json:"after_steps"`
	BeforeCycles uint64      `json:"before_cycles"`
	AfterCycles  uint64      `json:"after_cycles"`
	AfterRegs    oracle.Regs `json:"after_regs"`
	Audio        bool        `json:"audio_render"`
	Error        string
}

func TestCaveShieldTransitionNormalOriginalPair(t *testing.T) {
	study := caveShieldStudy(t)
	root := os.Getenv("ZELIARD_ORIG")
	out := os.Getenv("ZELIARD_SHIELD_TRANSITION_OUTPUT")
	if root == "" || out == "" {
		t.Skip("須提供正常原版與新的配對輸出；skip不算驗收")
	}
	read := func(p string) []byte { return caveShieldRead(t, p) }
	save := read(filepath.Join(root, "SAVEM1.USR"))
	if townBackgroundHash(save) != "03df9abcf62b1712d6b46dd1fc5b23335a4b8711b3dc84875a1498260d37735c" {
		t.Fatal("來源存檔身份")
	}
	info, e := os.Stat(filepath.Join(root, "SAVEM1.USR"))
	if e != nil {
		t.Fatal(e)
	}
	journal := read(filepath.Join(study, "cave-shield-gui-v2/SAVEM1/guest-inputs.jsonl"))
	scanner := bufio.NewScanner(bytes.NewReader(journal))
	scanner.Buffer(make([]byte, 4096), 1<<20)
	if !scanner.Scan() {
		t.Fatal("日誌缺標頭")
	}
	var header struct {
		Kind        string
		CyclesPerMS uint64      `json:"cycles_per_ms"`
		SampleRate  int         `json:"audio_sample_rate"`
		Memory      string      `json:"initial_memory_sha256"`
		Regs        oracle.Regs `json:"initial_regs"`
	}
	if e = json.Unmarshal(scanner.Bytes(), &header); e != nil || header.Kind != "header" {
		t.Fatal("日誌標頭", e)
	}
	var machines [2]*oracle.Oracle
	var audio [2]*oracle.Audio
	var saves []string
	for i := range machines {
		o, e := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if e != nil {
			t.Fatal(e)
		}
		machines[i] = o
		defer o.Close()
		scratch := t.TempDir()
		dst := filepath.Join(scratch, "SAVEM1.USR")
		if e = os.WriteFile(dst, save, 0644); e != nil {
			t.Fatal(e)
		}
		if e = os.Chtimes(dst, info.ModTime(), info.ModTime()); e != nil {
			t.Fatal(e)
		}
		saves = append(saves, dst)
		o.SetScratch(scratch)
		o.SetDOSBoxCycles(header.CyclesPerMS)
		o.TraceFiles()
		audio[i] = o.NewAudio(header.SampleRate)
		if townBackgroundHash(o.Bytes(oracle.Far(0, 0), 1<<20)) != header.Memory || o.Regs() != header.Regs {
			t.Fatal("冷啟動身份")
		}
	}
	a, b := machines[0], machines[1]
	frames := NewCaveFrames(b)
	art := caveShieldArt(t)
	wanted := map[uint64]caveTransitionSource{}
	for _, r := range caveTransitionSources(t) {
		wanted[r.Steps] = r
	}
	seen := map[uint64]bool{}
	for _, off := range []uint16{0x7114, 0x714c} {
		b.OnCall(oracle.Far(0x1b1, off), func(o *oracle.Oracle) {
			r, ok := wanted[o.Steps()]
			if !ok {
				return
			}
			_, _, rgb := o.ScreenRGB()
			f := frames.Present(o.Indexed(), rgb, false)
			if f == nil || f.Steps != r.Steps || f.BodySteps != r.BodySteps || townBackgroundHash(f.Memory) != r.MemorySHA || townBackgroundHash(f.BodyMemory) != r.BodySHA || townBackgroundHash(f.SourceIndexed) != r.RawSHA {
				t.Fatal("正常原版完成來源不同", r.File)
			}
			canvas := caveOriginalCanvas(f)
			if !art.Draw(canvas, f, image.Point{}) || art.PlayerState()["pose"] != 54 {
				t.Fatal("正常離洞HD未匹配")
			}
			seen[o.Steps()] = true
		})
	}
	checks := 0
	equal := func() {
		checks++
		if !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) || a.Regs() != b.Regs() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) {
			t.Fatal("028適配層改動完整原版", checks)
		}
	}
	equal()
	execute := func(r caveTransitionInput) {
		if a.Steps() != r.BeforeSteps || a.Cycles() != r.BeforeCycles {
			t.Fatal("日誌前置時間")
		}
		for i, o := range machines {
			switch r.Kind {
			case "key":
				if r.Pressed {
					o.KeyDown(r.Scan)
				} else {
					o.KeyUp(r.Scan)
				}
			case "run":
				if r.Budget > 5000000 || r.Instructions > 20000000 || o.Steps() > 220000000 {
					t.Fatal("原版預算")
				}
				var e error
				if r.Instructions != 0 {
					e = o.Run(r.Instructions)
				} else {
					e = o.RunCycles(r.Budget)
				}
				if r.Audio {
					audio[i].Render()
				}
				message := ""
				if e != nil {
					message = e.Error()
				}
				if message != r.Error || (r.Instructions == 0 && o.Regs() != r.AfterRegs) {
					t.Fatal("原版暫存器或錯誤", checks)
				}
			default:
				t.Fatal("未知輸入")
			}
			if o.Steps() != r.AfterSteps || o.Cycles() != r.AfterCycles {
				t.Fatal("日誌完成時間")
			}
		}
		equal()
	}
	records := 1
	for scanner.Scan() {
		var r caveTransitionInput
		if e = json.Unmarshal(scanner.Bytes(), &r); e != nil {
			t.Fatal(e)
		}
		execute(r)
		records++
	}
	if e = scanner.Err(); e != nil {
		t.Fatal(e)
	}
	var extra struct {
		Inputs []caveTransitionInput `json:"input_sequence"`
	}
	if e = json.Unmarshal(read(filepath.Join(study, "shield-transition-source-v2/extra-receipt.json")), &extra); e != nil {
		t.Fatal(e)
	}
	for _, r := range extra.Inputs {
		execute(r)
	}
	if len(seen) != 4 || frames.refused != 0 {
		t.Fatal("正常雙朝向離洞來源未齊", len(seen), frames.refused, frames.lastRefusal)
	}
	for _, o := range machines {
		if len(o.HeldKeys()) != 0 || len(o.Wrote()) != 0 || len(o.Unimplemented()) != 0 {
			t.Fatal("原版輸入／寫入未齊")
		}
	}
	for _, dst := range saves {
		after, e := os.Stat(dst)
		if e != nil || !after.ModTime().Equal(info.ModTime()) || !bytes.Equal(read(dst), save) {
			t.Fatal("來源存檔改變")
		}
	}
	report := map[string]any{"status": "PASS_SHIELD_TRANSITION_NORMAL_FULL_ORIGINAL_PAIR", "journal_records_including_header": records, "extra_inputs": len(extra.Inputs), "full_original_comparisons": checks, "native_special_HD_completions": len(seen), "refused": frames.refused, "save_unchanged": true, "state_injection": false, "snapshot_restore": false, "final_steps": b.Steps(), "final_cycles": b.Cycles(), "journal_sha256": townBackgroundHash(journal), "scope": "正常冷F7及實際日誌後Left／Up，雙朝向各首次及重用主體完成；實際視窗另驗"}
	data, e := json.MarshalIndent(report, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "normal-pair.json"), append(data, 10), 0644); e != nil {
		t.Fatal(e)
	}
	t.Log(fmt.Sprint(report))
}

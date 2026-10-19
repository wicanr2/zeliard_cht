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

type caveWeaponProof struct {
	Name       string
	Prefix     string `json:"source_prefix"`
	Player     bool   `json:"player_enabled"`
	Enabled    bool   `json:"weapon_enabled"`
	HD         bool   `json:"weapon_HD"`
	Empty      bool   `json:"weapon_empty"`
	Key        string `json:"weapon_key"`
	MemorySHA  string `json:"memory_sha256"`
	BodySHA    string `json:"body_memory_sha256"`
	IndexedSHA string `json:"indexed_sha256"`
	PaletteSHA string `json:"palette_sha256"`
	Paints     []CavePaint
}

func caveWeaponProofs(t *testing.T, root string) []caveWeaponProof {
	t.Helper()
	data, e := os.ReadFile(filepath.Join(root, "weapon-composition-preview-v3/receipt.json"))
	if e != nil {
		t.Fatal(e)
	}
	var j struct{ Records []caveWeaponProof }
	if e = json.Unmarshal(data, &j); e != nil {
		t.Fatal(e)
	}
	if len(j.Records) != 29 {
		t.Fatal("29份獨立來源未齊")
	}
	return j.Records
}

// 正常十來源保留真實兩個時鐘。舊後畫來源未保存完整時鐘時，只設單元邊界；
// 不作原版同時刻證據，正常與GUI原版配對另驗。
func caveWeaponFixture(t *testing.T, root string, p caveWeaponProof) *CavePresentation {
	t.Helper()
	read := func(ext string) []byte {
		b, e := os.ReadFile(filepath.Join(root, p.Prefix+"."+ext))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	memory, body, raw, pal := read("memory"), read("body.memory"), read("indexed"), read("palette")
	if townBackgroundHash(memory) != p.MemorySHA || townBackgroundHash(body) != p.BodySHA || townBackgroundHash(raw) != p.IndexedSHA || townBackgroundHash(pal) != p.PaletteSHA {
		t.Fatal("獨立武器來源身份不符", p.Name)
	}
	d, e := decodeCaveFrameWithPaints(memory, body, raw, p.Paints)
	if e != nil {
		t.Fatal(e)
	}
	if townBackgroundHash(d.WeaponRaster[:]) != p.Key {
		t.Fatal("1024武器色號／透明來源不符")
	}
	rgb := make([]byte, 192000)
	for i, v := range raw {
		copy(rgb[i*3:], pal[int(v)*3:int(v)*3+3])
	}
	var steps, cycles, bodySteps, bodyCycles uint64
	data, e := os.ReadFile(filepath.Join(root, "weapon-source-v1/receipt.json"))
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
	if e = json.Unmarshal(data, &source); e != nil {
		t.Fatal(e)
	}
	for _, r := range source.Sources {
		if p.Prefix == "weapon-source-v1/"+r.File {
			steps, cycles, bodySteps, bodyCycles = r.Steps, r.Cycles, r.BodySteps, r.BodyCycles
		}
	}
	for _, q := range p.Paints {
		steps = max(steps, q.Steps)
		cycles = max(cycles, q.Cycles)
	}
	return &CavePresentation{Memory: memory, BodyMemory: body, SourceIndexed: raw, Indexed: cloneTownBytes(raw), RGB: rgb, Palette: pal, Decoded: d, Paints: p.Paints, Steps: steps, Cycles: cycles, BodySteps: bodySteps, BodyCycles: bodyCycles}
}

func TestCaveWeaponDirectRegisteredPixels(t *testing.T) {
	root, _ := caveFixtures(t)
	weapon, e := NewCaveWeapon(root)
	if e != nil {
		t.Fatal(e)
	}
	data, e := os.ReadFile(filepath.Join(root, "weapon-art-preview-v1/registration.json"))
	if e != nil {
		t.Fatal(e)
	}
	var review struct {
		Poses []struct {
			Key  string
			Path string `json:"sprite_raw_path"`
		}
	}
	if e = json.Unmarshal(data, &review); e != nil {
		t.Fatal(e)
	}
	if len(review.Poses) != 10 || len(weapon.poses) != 10 {
		t.Fatal("武器十圖不完整")
	}
	for _, r := range review.Poses {
		p := weapon.poses[r.Key]
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
			t.Fatal("未直接按獨立Python取樣", r.Key)
		}
	}
}

func TestCaveWeaponIndependentCompletedPixels(t *testing.T) {
	root, _ := caveFixtures(t)
	out := os.Getenv("ZELIARD_CAVE_WEAPON_OUTPUT")
	if out == "" {
		t.Skip("須指定新的武器驗證輸出目錄")
	}
	if e := os.Mkdir(filepath.Join(out, "pixels"), 0755); e != nil {
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
	var results []map[string]any
	art, e := NewCaveTerrain(root)
	if e != nil {
		t.Fatal(e)
	}
	baseline, e := NewCaveTerrain(root)
	if e != nil {
		t.Fatal(e)
	}
	for _, p := range caveWeaponProofs(t, root) {
		t.Run(p.Name, func(t *testing.T) {
			f := caveWeaponFixture(t, root, p)
			art.SetPlayer(nil)
			baseline.SetPlayer(nil)
			art.SetWeapon(nil)
			if p.Player {
				art.SetPlayer(player)
				baseline.SetPlayer(player)
			}
			if p.Enabled {
				art.SetWeapon(weapon)
			}
			canvas, old := caveOriginalCanvas(f), caveOriginalCanvas(f)
			before := cloneTownBytes(canvas.Pix)
			memory, body, raw := cloneTownBytes(f.Memory), cloneTownBytes(f.BodyMemory), cloneTownBytes(f.SourceIndexed)
			if !art.Draw(canvas, f, image.Point{}) || !baseline.Draw(old, f, image.Point{}) {
				t.Fatal("已驗武器來源被拒絕")
			}
			state := art.WeaponState()
			if p.Enabled {
				if state["matched"] != p.HD || state["empty"] != p.Empty {
					t.Fatal("空白／武器匹配狀態不符", state)
				}
			} else if state != nil {
				t.Fatal("未開武器HD卻保留旗標")
			}
			fp, e := os.Open(filepath.Join(root, "weapon-composition-preview-v3", p.Name+".png"))
			if e != nil {
				t.Fatal(e)
			}
			want, e := png.Decode(fp)
			fp.Close()
			if e != nil {
				t.Fatal(e)
			}
			bounds := image.Rectangle{}
			if pose := weapon.poses[p.Key]; p.HD && pose != nil {
				bounds = pose.bounds.Add(image.Pt(f.Decoded.WeaponX*3, f.Decoded.WeaponY*3))
			}
			changed := 0
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := canvas.PixOffset(x, y)
					pt := image.Pt(x, y)
					if pt.In(image.Rect(144, 42, 816, 474)) {
						if canvas.RGBAAt(x, y) != color.RGBAModel.Convert(want.At(x, y)).(color.RGBA) {
							t.Fatalf("獨立武器合成不符 %d,%d", x, y)
						}
					} else if !bytes.Equal(canvas.Pix[i:i+4], before[i:i+4]) {
						t.Fatal("覆寫遊戲區外")
					}
					if !bytes.Equal(canvas.Pix[i:i+4], old.Pix[i:i+4]) {
						changed++
						if !pt.In(bounds) {
							t.Fatal("武器HD改動原版外包絡之外", x, y, bounds)
						}
					}
				}
			}
			if p.HD && changed == 0 {
				t.Fatal("武器HD沒有可見改動")
			}
			if !p.HD && changed != 0 {
				t.Fatal("空白或未開武器HD改動圖面")
			}
			if !bytes.Equal(memory, f.Memory) || !bytes.Equal(body, f.BodyMemory) || !bytes.Equal(raw, f.SourceIndexed) {
				t.Fatal("武器高清回寫原版")
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
			results = append(results, map[string]any{"name": p.Name, "independent_pixels_equal": true, "pixels_compared": 672 * 432, "weapon_HD": p.HD, "weapon_empty": p.Empty, "weapon_changed_pixels": changed, "player_enabled": p.Player, "source_unchanged": true, "weapon": state})
		})
	}
	if len(results) != 29 {
		t.Fatal("29份武器圖面未完成")
	}
	data, e := json.MarshalIndent(map[string]any{"status": "PASS_CAVE_WEAPON_INDEPENDENT_PIXELS", "records": results, "limits": "單元圖面，正常原版時鐘及實際視窗另驗"}, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, 10), 0644); e != nil {
		t.Fatal(e)
	}
}

func TestCaveWeaponFallbackAndLifecycle(t *testing.T) {
	root, _ := caveFixtures(t)
	proofs := caveWeaponProofs(t, root)
	f := caveWeaponFixture(t, root, proofs[0])
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
	art.SetWeapon(weapon)
	key := townBackgroundHash(f.Decoded.WeaponRaster[:])
	pose := weapon.poses[key]
	delete(weapon.poses, key)
	got, want := caveOriginalCanvas(f), caveOriginalCanvas(f)
	if !art.Draw(got, f, image.Point{}) || !baseline.Draw(want, f, image.Point{}) || !bytes.Equal(got.Pix, want.Pix) || art.WeaponState()["matched"] != false {
		t.Fatal("缺少姿態未保留原版武器")
	}
	weapon.poses[key] = pose
	unknown := *f.Decoded
	unknown.WeaponRaster[0] ^= 1
	if _, p, empty := weapon.selectPose(&unknown); p != nil || empty {
		t.Fatal("未知武器借用相似素材")
	}
	bad := *f
	bad.SourceIndexed = cloneTownBytes(f.SourceIndexed)
	bad.SourceIndexed[100*320+100] ^= 128
	got = caveOriginalCanvas(f)
	before := cloneTownBytes(got.Pix)
	if art.Draw(got, &bad, image.Point{}) || !bytes.Equal(got.Pix, before) {
		t.Fatal("來源不符未原子退回")
	}
	originalBounds := pose.bounds
	pose.bounds = image.Rect(-1, 0, 96, 96)
	if _, p, _ := weapon.selectPose(f.Decoded); p != nil {
		t.Fatal("壞包絡仍被選用")
	}
	pose.bounds = originalBounds
	for _, corrupt := range []bool{false, true} {
		dir := t.TempDir()
		if corrupt {
			if e = os.WriteFile(filepath.Join(dir, "weapon-atlas-handpaint-v1.png"), []byte("broken"), 0644); e != nil {
				t.Fatal(e)
			}
		}
		if w, e := NewCaveWeapon(dir); w != nil || e == nil {
			t.Fatal("缺圖或壞圖未整組拒絕")
		}
	}
	if !art.Draw(caveOriginalCanvas(f), f, image.Point{}) || art.WeaponState()["matched"] != true {
		t.Fatal("已知武器未恢復")
	}
	var emptyProof *caveWeaponProof
	for i := range proofs {
		if proofs[i].Empty {
			emptyProof = &proofs[i]
			break
		}
	}
	if emptyProof == nil {
		t.Fatal("缺少真實無武器來源")
	}
	empty := caveWeaponFixture(t, root, *emptyProof)
	if !art.Draw(caveOriginalCanvas(empty), empty, image.Point{}) || art.WeaponState()["empty"] != true || art.WeaponState()["matched"] != false {
		t.Fatal("空白武器仍沿用上一姿態")
	}
	art.SetWeapon(nil)
	if art.WeaponState() != nil {
		t.Fatal("撤除後旗標殘留")
	}
}

// 正常路徑及固定GUI日誌逐個選定完成來源盤點，不把缺圖當已匹配。
type caveWeaponCoverage struct {
	weapon                *CaveWeapon
	frames                map[uint64]bool
	keys                  map[string]int
	dumped                map[string]bool
	known, empty, missing int
	sources               []map[string]any
	captureSteps          map[uint64]bool
	captureOrder          []uint64
	capturedSources       []map[string]any
}

func newCaveWeaponCoverage(w *CaveWeapon) *caveWeaponCoverage {
	c := &caveWeaponCoverage{weapon: w, frames: map[uint64]bool{}, keys: map[string]int{}, dumped: map[string]bool{}, captureSteps: map[uint64]bool{}}
	for _, text := range strings.Split(os.Getenv("ZELIARD_CAVE_WEAPON_CAPTURE_STEPS"), ",") {
		if text == "" {
			continue
		}
		step, e := strconv.ParseUint(text, 10, 64)
		if e != nil {
			panic(e)
		}
		if !c.captureSteps[step] {
			c.captureSteps[step] = true
			c.captureOrder = append(c.captureOrder, step)
		}
	}
	return c
}
func (c *caveWeaponCoverage) observe(t *testing.T, out string, f *CavePresentation) {
	t.Helper()
	if c.weapon == nil || f == nil {
		return
	}
	if c.captureSteps[f.Steps] {
		stem := fmt.Sprintf("weapon-observed-%02d", len(c.capturedSources))
		for ext, data := range map[string][]byte{"memory": f.Memory, "body.memory": f.BodyMemory, "indexed": f.SourceIndexed, "palette": f.Palette, "weapon.indexed": f.Decoded.WeaponRaster[:]} {
			if e := os.WriteFile(filepath.Join(out, stem+"."+ext), data, 0644); e != nil {
				t.Fatal(e)
			}
		}
		data, e := json.Marshal(f.Paints)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(out, stem+".paints.json"), data, 0644); e != nil {
			t.Fatal(e)
		}
		r := map[string]any{"file": stem, "steps": f.Steps, "cycles": f.Cycles, "body_steps": f.BodySteps, "body_cycles": f.BodyCycles, "memory_sha256": townBackgroundHash(f.Memory), "body_memory_sha256": townBackgroundHash(f.BodyMemory), "indexed_sha256": townBackgroundHash(f.SourceIndexed), "palette_sha256": townBackgroundHash(f.Palette), "weapon_key": townBackgroundHash(f.Decoded.WeaponRaster[:]), "weapon_visible_pixels": len(f.Decoded.Weapon), "scope": "GUI觀察器指定步數的真實完成來源，非當次CPU記憶體"}
		data, e = json.MarshalIndent(r, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(out, stem+".json"), append(data, 10), 0644); e != nil {
			t.Fatal(e)
		}
		c.capturedSources = append(c.capturedSources, r)
		c.captureSteps[f.Steps] = false
	}
	if c.frames[f.Steps] {
		return
	}
	c.frames[f.Steps] = true
	key, pose, empty := c.weapon.selectPose(f.Decoded)
	c.keys[key]++
	if empty {
		c.empty++
		return
	}
	if pose != nil {
		c.known++
		return
	}
	c.missing++
	if c.dumped[key] {
		return
	}
	c.dumped[key] = true
	stem := fmt.Sprintf("weapon-missing-%02d", len(c.sources))
	for suffix, data := range map[string][]byte{"memory": f.Memory, "body.memory": f.BodyMemory, "indexed": f.SourceIndexed, "palette": f.Palette} {
		if e := os.WriteFile(filepath.Join(out, stem+"."+suffix), data, 0644); e != nil {
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
	c.sources = append(c.sources, map[string]any{"key": key, "file": stem, "steps": f.Steps, "cycles": f.Cycles, "body_steps": f.BodySteps, "body_cycles": f.BodyCycles, "memory_sha256": townBackgroundHash(f.Memory), "body_memory_sha256": townBackgroundHash(f.BodyMemory), "indexed_sha256": townBackgroundHash(f.SourceIndexed)})
}
func (c *caveWeaponCoverage) summary() map[string]any {
	var pending []uint64
	for _, step := range c.captureOrder {
		if c.captureSteps[step] {
			pending = append(pending, step)
		}
	}
	return map[string]any{"source_frames": len(c.frames), "known_frames": c.known, "empty_frames": c.empty, "missing_visible_frames": c.missing, "keys": c.keys, "missing_sources": c.sources, "requested_source_captures": c.captureOrder, "captured_sources": c.capturedSources, "uncaptured_steps": pending, "scope": "逐個已選定完成來源；缺圖回退，未宣稱未觀察姿態"}
}

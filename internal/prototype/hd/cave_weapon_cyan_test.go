package hd

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func loadCyanNormalArt(t *testing.T, weapon *CaveWeapon) {
	t.Helper()
	if dir := os.Getenv("ZELIARD_CAVE_CYAN_WEAPON_ART"); dir != "" {
		if err := weapon.LoadCyan(dir); err != nil {
			t.Fatal(err)
		}
	}
	if dir := os.Getenv("ZELIARD_CAVE_WEAPON_EXTRA_ART"); dir != "" {
		if err := weapon.LoadExtra(dir); err != nil {
			t.Fatal(err)
		}
	}
}

func cyanWeaponArt(t *testing.T) *CaveWeapon {
	t.Helper()
	root, _ := caveFixtures(t)
	w, err := NewCaveWeapon(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.LoadCyan(filepath.Join(iceStudy(t), "cave-cyan-weapon-complete-v1")); err != nil {
		t.Fatal(err)
	}
	return w
}

func cyanSavedSources(t *testing.T) []struct {
	Kind   string
	Prefix string         `json:"source_prefix"`
	Point  savedGearPoint `json:"source"`
} {
	t.Helper()
	var receipt struct {
		Sources []struct {
			Kind   string
			Prefix string         `json:"source_prefix"`
			Point  savedGearPoint `json:"source"`
		}
	}
	if err := json.Unmarshal(iceRead(t, filepath.Join(iceStudy(t), "cave-saved-gear-source-v1/receipt.json")), &receipt); err != nil {
		t.Fatal(err)
	}
	return receipt.Sources
}

func TestCaveCyanWeaponSavedSourcesAndIndependentPixels(t *testing.T) {
	w, art := cyanWeaponArt(t), doorTestArt(t)
	art.SetWeapon(w)
	project := filepath.Dir(filepath.Dir(iceStudy(t)))
	count, checked := 0, 0
	for _, row := range cyanSavedSources(t) {
		if row.Kind != "weapon" {
			continue
		}
		read := func(ext string) []byte { return iceRead(t, filepath.Join(project, row.Prefix+"."+ext)) }
		p := row.Point
		f := &CavePresentation{Memory: read("memory"), BodyMemory: read("body.memory"), SourceIndexed: read("indexed"), Palette: read("palette"), Steps: p.Steps, Cycles: p.Cycles, BodySteps: p.BodySteps, BodyCycles: p.BodyCycles, Paints: p.Paints}
		if data, err := os.ReadFile(filepath.Join(project, row.Prefix+".paints.json")); err == nil {
			if err := json.Unmarshal(data, &f.Paints); err != nil {
				t.Fatal(err)
			}
		}
		if townBackgroundHash(f.Memory) != p.MemorySHA || townBackgroundHash(f.BodyMemory) != p.BodySHA || townBackgroundHash(f.SourceIndexed) != p.IndexedSHA {
			t.Fatal("已存正常來源改變")
		}
		magiaFrameRGB(f)
		var err error
		f.Decoded, err = decodeCaveFrameWithPaints(f.Memory, f.BodyMemory, f.SourceIndexed, f.Paints)
		if err != nil {
			t.Fatal(err)
		}
		memory, body := cloneTownBytes(f.Memory), cloneTownBytes(f.BodyMemory)
		canvas := caveOriginalCanvas(f)
		if !art.Draw(canvas, f, image.Point{}) || art.WeaponState()["matched"] != true || art.WeaponState()["source_key"] != p.Key {
			t.Fatal("已存青藍來源仍退回", p.Key, art.WeaponState())
		}
		if !bytes.Equal(memory, f.Memory) || !bytes.Equal(body, f.BodyMemory) {
			t.Fatal("武器HD回寫來源")
		}
		im, err := png.Decode(bytes.NewReader(iceRead(t, filepath.Join(iceStudy(t), "cave-cyan-weapon-complete-v1", "weapon-"+p.Key[:12]+"-96x96.png"))))
		if err != nil {
			t.Fatal(err)
		}
		pose := w.poses[p.Key]
		// 透明替身只用於隔離下層合成，不作原版行為證據。
		// 仍走同一個來源及非nil武器分支，避免把原生刀光當混色下層。
		transparent := *pose
		transparent.pixels = make([]color.NRGBA, 96*96)
		w.poses[p.Key] = &transparent
		under := caveOriginalCanvas(f)
		if !art.Draw(under, f, image.Point{}) {
			t.Fatal("無色刀光的下層合成被拒絕")
		}
		w.poses[p.Key] = pose
		for y := 0; y < 96; y++ {
			for x := 0; x < 96; x++ {
				c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
				if c != pose.pixels[y*96+x] {
					t.Fatal("獨立註冊RGBA不符")
				}
				if c.A == 0 {
					continue
				}
				px, py := f.Decoded.WeaponX*3+x, f.Decoded.WeaponY*3+y
				old := under.RGBAAt(px, py)
				a, inv := uint32(c.A), uint32(255-c.A)
				want := color.RGBA{uint8((uint32(c.R)*a + uint32(old.R)*inv + 127) / 255), uint8((uint32(c.G)*a + uint32(old.G)*inv + 127) / 255), uint8((uint32(c.B)*a + uint32(old.B)*inv + 127) / 255), 255}
				if canvas.RGBAAt(px, py) != want {
					t.Fatal("正常刀光獨立像素不符", p.Key, x, y)
				}
				checked++
			}
		}
		unknown := *f.Decoded
		unknown.WeaponRaster[0] ^= 1
		if _, pose, _ := w.selectPose(&unknown); pose != nil {
			t.Fatal("未知來源借用相似刀光")
		}
		count++
	}
	if count != 10 || len(w.poses) != 20 || checked < 100 {
		t.Fatal("十姿態覆蓋不足", count, len(w.poses), checked)
	}
	t.Logf("十個正常保存來源已匹配，92160註冊RGBA與%d個透明度合成像素獨立相同，原始來源未改", checked)
}

func TestCaveCyanWeaponAtomicLoadAndLegacy(t *testing.T) {
	w := cyanWeaponArt(t)
	old := w.poses[caveWeaponRegistrations[0].key]
	before := make(map[string]*caveWeaponPose, len(w.poses))
	for key, pose := range w.poses {
		before[key] = pose
	}
	if err := w.LoadCyan(t.TempDir()); err == nil || len(w.poses) != 20 || w.poses[caveWeaponRegistrations[0].key] != old {
		t.Fatal("缺圖破壞既有圖組")
	}
	dir := t.TempDir()
	for i, r := range caveCyanWeapons {
		name := "weapon-" + r.key[:12] + "-96x96.png"
		data := iceRead(t, filepath.Join(iceStudy(t), "cave-cyan-weapon-complete-v1", name))
		if i == 9 {
			data = []byte("broken")
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.LoadCyan(dir); err == nil || w.poses[caveWeaponRegistrations[0].key] != old || len(w.poses) != 20 {
		t.Fatal("最後一圖失敗未原子拒絕")
	}
	for key, pose := range before {
		if w.poses[key] != pose {
			t.Fatal("失敗載入替換了既有圖面", key)
		}
	}
	for _, r := range caveWeaponRegistrations {
		pose := w.poses[r.key]
		var raw []byte
		for _, c := range pose.pixels {
			raw = append(raw, c.R, c.G, c.B, c.A)
		}
		if townBackgroundHash(raw) != r.pixelsSHA {
			t.Fatal("舊十圖被改變")
		}
	}
}

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

func loadExtraPlayerNormalArt(t *testing.T, player *CavePlayer) {
	t.Helper()
	if dir := os.Getenv("ZELIARD_CAVE_PLAYER_EXTRA_ART"); dir != "" {
		if err := player.LoadExtra(dir); err != nil {
			t.Fatal(err)
		}
	}
	if dir := os.Getenv("ZELIARD_CAVE_PLAYER_PROFILE_ART"); dir != "" {
		if err := player.LoadProfiles(dir); err != nil {
			t.Fatal(err)
		}
	}
}

func savedPlayerArt(t *testing.T) *CavePlayer {
	t.Helper()
	root, _ := caveFixtures(t)
	p, err := NewCavePlayer(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.LoadExtra(filepath.Join(iceStudy(t), "cave-saved-player-complete-v1")); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCaveExtraPlayerSavedSourcesAndIndependentPixels(t *testing.T) {
	p, w, art := savedPlayerArt(t), cyanWeaponArt(t), doorTestArt(t)
	art.SetPlayer(p)
	art.SetWeapon(w)
	project := filepath.Dir(filepath.Dir(iceStudy(t)))
	count, checked, variants := 0, 0, 0
	for _, row := range cyanSavedSources(t) {
		if row.Kind != "player" {
			continue
		}
		read := func(ext string) []byte { return iceRead(t, filepath.Join(project, row.Prefix+"."+ext)) }
		r := row.Point
		f := &CavePresentation{Memory: read("memory"), BodyMemory: read("body.memory"), SourceIndexed: read("indexed"), Palette: read("palette"), Steps: r.Steps, Cycles: r.Cycles, BodySteps: r.BodySteps, BodyCycles: r.BodyCycles, Paints: r.Paints}
		if data, err := os.ReadFile(filepath.Join(project, row.Prefix+".paints.json")); err == nil {
			if err := json.Unmarshal(data, &f.Paints); err != nil {
				t.Fatal(err)
			}
		}
		if townBackgroundHash(f.Memory) != r.MemorySHA || townBackgroundHash(f.BodyMemory) != r.BodySHA || townBackgroundHash(f.SourceIndexed) != r.IndexedSHA {
			t.Fatal("保存來源SHA改變")
		}
		magiaFrameRGB(f)
		var err error
		f.Decoded, err = decodeCaveFrameWithPaints(f.Memory, f.BodyMemory, f.SourceIndexed, f.Paints)
		if err != nil || townBackgroundHash(f.Decoded.ActorRaster[:]) != r.Key {
			t.Fatal("主角原生來源不符", err)
		}
		before, body := cloneTownBytes(f.Memory), cloneTownBytes(f.BodyMemory)
		canvas := caveOriginalCanvas(f)
		if !art.Draw(canvas, f, image.Point{}) || art.PlayerState()["matched"] != true || art.PlayerState()["source_key"] != r.Key {
			t.Fatal("新主角來源仍退回", r.Key, art.PlayerState())
		}
		if !bytes.Equal(before, f.Memory) || !bytes.Equal(body, f.BodyMemory) {
			t.Fatal("主角HD回寫來源")
		}
		pose := p.poses[r.Key]
		if pose.number >= 61 {
			variants++
		}
		im, err := png.Decode(bytes.NewReader(iceRead(t, filepath.Join(iceStudy(t), "cave-saved-player-complete-v1", "player-"+r.Key[:12]+"-72x72.png"))))
		if err != nil {
			t.Fatal(err)
		}
		// 無色替身只隔離下層，不作原版行為證據；兩側仍走同一個非nil主角分支。
		transparent := *pose
		transparent.pixels = make([]color.NRGBA, 72*72)
		p.poses[r.Key] = &transparent
		under := caveOriginalCanvas(f)
		if !art.Draw(under, f, image.Point{}) {
			t.Fatal("主角下層合成被拒絕")
		}
		p.poses[r.Key] = pose
		blocked := map[image.Point]bool{}
		for _, paint := range f.Paints {
			for y := 0; y < 8; y++ {
				for x := 0; x < 8; x++ {
					if paint.Mask&(uint64(1)<<uint(y*8+x)) != 0 {
						for dy := 0; dy < 3; dy++ {
							for dx := 0; dx < 3; dx++ {
								blocked[image.Pt((paint.X+x)*3+dx, (paint.Y+y)*3+dy)] = true
							}
						}
					}
				}
			}
		}
		for _, v := range f.Decoded.Magia {
			for dy := 0; dy < 3; dy++ {
				for dx := 0; dx < 3; dx++ {
					blocked[image.Pt(v.X*3+dx, v.Y*3+dy)] = true
				}
			}
		}
		_, weapon, _ := w.selectPose(f.Decoded)
		if weapon != nil {
			for y := weapon.bounds.Min.Y; y < weapon.bounds.Max.Y; y++ {
				for x := weapon.bounds.Min.X; x < weapon.bounds.Max.X; x++ {
					if weapon.pixels[y*96+x].A != 0 {
						blocked[image.Pt(f.Decoded.WeaponX*3+x, f.Decoded.WeaponY*3+y)] = true
					}
				}
			}
		} else {
			for _, v := range f.Decoded.Weapon {
				for dy := 0; dy < 3; dy++ {
					for dx := 0; dx < 3; dx++ {
						blocked[image.Pt(v.X*3+dx, v.Y*3+dy)] = true
					}
				}
			}
		}
		local := 0
		for y := 0; y < 72; y++ {
			for x := 0; x < 72; x++ {
				c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
				if c != pose.pixels[y*72+x] {
					t.Fatal("主角獨立註冊RGBA不符")
				}
				px, py := f.Decoded.ActorX*3+x, f.Decoded.ActorY*3+y
				if c.A == 0 || blocked[image.Pt(px, py)] {
					continue
				}
				old := under.RGBAAt(px, py)
				a, inv := uint32(c.A), uint32(255-c.A)
				want := color.RGBA{uint8((uint32(c.R)*a + uint32(old.R)*inv + 127) / 255), uint8((uint32(c.G)*a + uint32(old.G)*inv + 127) / 255), uint8((uint32(c.B)*a + uint32(old.B)*inv + 127) / 255), 255}
				if canvas.RGBAAt(px, py) != want {
					t.Fatal("主角獨立混色不符", r.Key, x, y)
				}
				checked++
				local++
			}
		}
		if local < 100 {
			t.Fatal("主角未遮擋像素不足", r.Key, local)
		}
		unknown := *f.Decoded
		unknown.ActorRaster[0] ^= 1
		if _, pose, _ := p.selectPose(&unknown); pose != nil {
			t.Fatal("未知主角借用相似姿態")
		}
		if out := os.Getenv("ZELIARD_SAVED_PLAYER_OUTPUT"); out != "" && (count == 0 || count == 6 || count == 11) {
			file, err := os.Create(filepath.Join(out, "player-"+r.Key[:12]+"-hd.png"))
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(file, canvas)
			file.Close()
			if err != nil {
				t.Fatal(err)
			}
		}
		count++
	}
	if count != 12 || variants != 6 || len(p.poses) != 67 {
		t.Fatal("主角覆蓋不足", count, variants, len(p.poses))
	}
	t.Logf("十二個正常保存來源、六個色彩變體：62208註冊RGBA與%d個未遮擋混色像素獨立相同，來源未改", checked)
}

func TestCaveExtraPlayerAtomicLoadAndLegacy(t *testing.T) {
	p := savedPlayerArt(t)
	before := map[string]*cavePlayerPose{}
	for key, pose := range p.poses {
		before[key] = pose
	}
	if err := p.LoadExtra(t.TempDir()); err == nil {
		t.Fatal("缺圖未拒絕")
	}
	dir := t.TempDir()
	for i, r := range caveSavedPlayerPoses {
		name := "player-" + r.key[:12] + "-72x72.png"
		data := iceRead(t, filepath.Join(iceStudy(t), "cave-saved-player-complete-v1", name))
		if i == 11 {
			data = []byte("broken last image")
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.LoadExtra(dir); err == nil || len(p.poses) != 67 {
		t.Fatal("壞圖未原子拒絕")
	}
	for key, pose := range before {
		if p.poses[key] != pose {
			t.Fatal("失敗載入替換既有姿態", key)
		}
	}
	for _, r := range cavePlayerRegistrations {
		var raw []byte
		for _, c := range p.poses[r.key].pixels {
			raw = append(raw, c.R, c.G, c.B, c.A)
		}
		if townBackgroundHash(raw) != r.pixelsSHA {
			t.Fatal("原55圖被改變", r.key)
		}
	}
}

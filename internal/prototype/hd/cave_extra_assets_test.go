package hd

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func loadExtraNormalObjects(t *testing.T, objects *CaveObjects) {
	t.Helper()
	if dir := os.Getenv("ZELIARD_CAVE_OBJECT_EXTRA_ART"); dir != "" {
		if err := objects.LoadExtra(dir); err != nil {
			t.Fatal(err)
		}
	}
}

func unmatchedArtDir(t *testing.T) string {
	return filepath.Join(iceStudy(t), "cave-unmatched-complete-v1")
}

func extraObjects(t *testing.T) *CaveObjects {
	t.Helper()
	root, _ := caveFixtures(t)
	s, err := NewCaveObjects(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.LoadExtra(unmatchedArtDir(t)); err != nil {
		t.Fatal(err)
	}
	return s
}

func unmatchedSources(t *testing.T) []struct {
	Kind, Key string
	Prefix    string         `json:"source_prefix"`
	Point     savedGearPoint `json:"source"`
	Object    *caveObject    `json:"object"`
} {
	t.Helper()
	var j struct {
		Sources []struct {
			Kind, Key string
			Prefix    string         `json:"source_prefix"`
			Point     savedGearPoint `json:"source"`
			Object    *caveObject    `json:"object"`
		}
	}
	if err := json.Unmarshal(iceRead(t, filepath.Join(iceStudy(t), "cave-unmatched-assets-source-v1/receipt.json")), &j); err != nil {
		t.Fatal(err)
	}
	if len(j.Sources) != 17 {
		t.Fatal("17來源不齊")
	}
	return j.Sources
}

func TestCaveExtraAssetsSavedSources(t *testing.T) {
	art := doorTestArt(t)
	w, s := cyanWeaponArt(t), extraObjects(t)
	if err := w.LoadExtra(unmatchedArtDir(t)); err != nil {
		t.Fatal(err)
	}
	art.SetPlayer(profilePlayerArt(t))
	art.SetWeapon(w)
	art.SetObjects(s)
	project := filepath.Dir(filepath.Dir(iceStudy(t)))
	count, masked, weaponBlends := 0, 0, 0
	for _, row := range unmatchedSources(t) {
		read := func(ext string) []byte { return iceRead(t, filepath.Join(project, row.Prefix+"."+ext)) }
		p := row.Point
		f := &CavePresentation{Memory: read("memory"), BodyMemory: read("body.memory"), SourceIndexed: read("indexed"), Palette: read("palette"), Steps: p.Steps, Cycles: p.Cycles, BodySteps: p.BodySteps, BodyCycles: p.BodyCycles}
		if err := json.Unmarshal(read("paints.json"), &f.Paints); err != nil {
			t.Fatal(err)
		}
		if townBackgroundHash(f.Memory) != p.MemorySHA || townBackgroundHash(f.BodyMemory) != p.BodySHA || townBackgroundHash(f.SourceIndexed) != p.IndexedSHA {
			t.Fatal("未改來源SHA不符")
		}
		var err error
		f.Decoded, err = decodeCaveFrameWithPaints(f.Memory, f.BodyMemory, f.SourceIndexed, f.Paints)
		if err != nil || f.Decoded.Differences != 0 {
			t.Fatal("原生完整重建失配", err)
		}
		magiaFrameRGB(f)
		before, body := cloneTownBytes(f.Memory), cloneTownBytes(f.BodyMemory)
		canvas := caveOriginalCanvas(f)
		original := append([]byte{}, canvas.Pix...)
		if !art.Draw(canvas, f, image.Point{}) {
			t.Fatal("來源仍退回", row.Key)
		}
		if row.Kind == "weapon" {
			if art.WeaponState()["matched"] != true || art.WeaponState()["source_key"] != row.Key {
				t.Fatal("刀光仍缺圖", art.WeaponState())
			}
			pose := w.poses[row.Key]
			transparent := *pose
			transparent.pixels = make([]color.NRGBA, 96*96)
			w.poses[row.Key] = &transparent
			under := caveOriginalCanvas(f)
			if !art.Draw(under, f, image.Point{}) {
				t.Fatal("刀光下層拒絕")
			}
			w.poses[row.Key] = pose
			for y := 0; y < 96; y++ {
				for x := 0; x < 96; x++ {
					c := pose.pixels[y*96+x]
					if c.A == 0 {
						continue
					}
					px, py := f.Decoded.WeaponX*3+x, f.Decoded.WeaponY*3+y
					old := under.RGBAAt(px, py)
					a, b := uint32(c.A), uint32(255-c.A)
					want := color.RGBA{uint8((uint32(c.R)*a + uint32(old.R)*b + 127) / 255), uint8((uint32(c.G)*a + uint32(old.G)*b + 127) / 255), uint8((uint32(c.B)*a + uint32(old.B)*b + 127) / 255), 255}
					if canvas.RGBAAt(px, py) != want {
						t.Fatal("四個新刀光的獨立混色失配", row.Key, x, y)
					}
					weaponBlends++
				}
			}
		} else {
			key, pose := s.selectPose(*row.Object)
			if key != row.Key || pose == nil {
				t.Fatal("物件仍缺圖", row.Key)
			}
			if row.Object.Backgrounds != [4]byte{} {
				masked++
				r := s.masked[key]
				if pose != s.poses[r.target] || art.ObjectState()["terrain_underlay"] != true {
					t.Fatal("未共用目標或沒有純地形下層")
				}
				bad := *row.Object
				bad.Backgrounds[0] ^= 1
				if _, p := s.selectPose(bad); p != nil {
					t.Fatal("底圖失配借圖")
				}
				bad = *row.Object
				bad.Tiles[0] ^= 1
				if _, p := s.selectPose(bad); p != nil {
					t.Fatal("圖格失配借圖")
				}
				bad = *row.Object
				bad.Raster[0] ^= 1
				if _, p := s.selectPose(bad); p != nil {
					t.Fatal("未知Raster借圖")
				}
			}
		}
		if !bytes.Equal(before, f.Memory) || !bytes.Equal(body, f.BodyMemory) {
			t.Fatal("HD改寫來源")
		}
		for y := 0; y < canvas.Bounds().Dy(); y++ {
			for x := 0; x < canvas.Bounds().Dx(); x++ {
				if !image.Pt(x, y).In(image.Rect(144, 42, 816, 474)) {
					i := canvas.PixOffset(x, y)
					if !bytes.Equal(canvas.Pix[i:i+4], original[i:i+4]) {
						t.Fatal("寫出遊戲區")
					}
				}
			}
		}
		count++
	}
	if count != 17 || masked != 9 || len(w.poses) != 24 || len(s.poses) != 33 {
		t.Fatal("缺少來源或圖面", count, masked, len(w.poses), len(s.poses))
	}
	t.Log("17份來源548352原生像素零差、九個精確遮罩來源通過，RAM與區外不變")
	t.Log("四個刀光獨立混色像素", weaponBlends)
}

func TestCaveExtraAssetsAtomicAndDirectPixels(t *testing.T) {
	w := cyanWeaponArt(t)
	root, _ := caveFixtures(t)
	s, err := NewCaveObjects(root)
	if err != nil {
		t.Fatal(err)
	}
	oldW, oldO := map[string]*caveWeaponPose{}, map[string]*caveObjectPose{}
	for k, p := range w.poses {
		oldW[k] = p
	}
	for k, p := range s.poses {
		oldO[k] = p
	}
	for _, kind := range []string{"weapon", "object"} {
		regs, size := caveExtraWeapons, 96
		if kind == "object" {
			regs, size = caveExtraObjects, 48
		}
		dir := t.TempDir()
		for n, r := range regs {
			name := kind + "-" + r.key[:12] + "-"
			if size == 96 {
				name += "96x96.png"
			} else {
				name += "48x48.png"
			}
			data := iceRead(t, filepath.Join(unmatchedArtDir(t), name))
			if n == len(regs)-1 {
				data = []byte("broken")
			}
			if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
				t.Fatal(err)
			}
		}
		if kind == "weapon" {
			if w.LoadExtra(t.TempDir()) == nil || w.LoadExtra(dir) == nil || len(w.poses) != 20 {
				t.Fatal("刀光載入不原子")
			}
		} else {
			if s.LoadExtra(t.TempDir()) == nil || s.LoadExtra(dir) == nil || len(s.poses) != 29 || s.maskedEnabled() {
				t.Fatal("物件載入不原子")
			}
		}
	}
	if err := w.LoadExtra(unmatchedArtDir(t)); err != nil {
		t.Fatal(err)
	}
	if err := s.LoadExtra(unmatchedArtDir(t)); err != nil {
		t.Fatal(err)
	}
	for k, p := range oldW {
		if w.poses[k] != p {
			t.Fatal("原20刀光被替換")
		}
	}
	for k, p := range oldO {
		if s.poses[k] != p {
			t.Fatal("原29物件被替換")
		}
	}
	checked := 0
	for _, kind := range []string{"weapon", "object"} {
		regs, size := caveExtraWeapons, 96
		if kind == "object" {
			regs, size = caveExtraObjects, 48
		}
		for _, r := range regs {
			name := kind + "-" + r.key[:12] + "-"
			if size == 96 {
				name += "96x96.png"
			} else {
				name += "48x48.png"
			}
			im, err := png.Decode(bytes.NewReader(iceRead(t, filepath.Join(unmatchedArtDir(t), name))))
			if err != nil {
				t.Fatal(err)
			}
			pixels := w.poses[caveExtraWeapons[0].key].pixels
			if kind == "weapon" {
				pixels = w.poses[r.key].pixels
			} else {
				pixels = s.poses[r.key].pixels
			}
			for y := 0; y < size; y++ {
				for x := 0; x < size; x++ {
					c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
					if pixels[y*size+x] != c {
						t.Fatal("直接註冊RGBA失配")
					}
					if !image.Pt(x, y).In(image.Rectangle{Min: r.bounds.Min.Mul(3), Max: r.bounds.Max.Mul(3)}) && c.A != 0 {
						t.Fatal("圖面越過原生包絡")
					}
					checked++
				}
			}
		}
	}
	t.Log("八張直接註冊", checked, "RGBA像素相同，原20／29圖不變，缺圖與最後一圖損壞原子拒絕")
}

func TestCaveExtraObjectsTerrainOrderClipAndFallback(t *testing.T) {
	s := extraObjects(t)
	palette := make([]byte, 768)
	for i := 0; i < 256; i++ {
		palette[i*3] = byte(i)
		palette[i*3+1] = byte(i / 2)
		palette[i*3+2] = byte(255 - i)
	}
	origin := image.Pt(7, 11)
	box := image.Rect(144, 42, 816, 474).Add(origin)
	base := image.NewRGBA(box)
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			base.SetRGBA(x, y, color.RGBA{byte(x), byte(y), byte(x + y), 255})
		}
	}
	makeCanvas := func() *image.RGBA {
		v := image.NewRGBA(image.Rect(0, 0, 980, 620))
		draw.Draw(v, v.Bounds(), image.NewUniform(color.RGBA{17, 29, 31, 255}), image.Point{}, draw.Src)
		draw.Draw(v, box, base, box.Min, draw.Src)
		return v
	}
	blended, tests := 0, 0
	for _, row := range unmatchedSources(t) {
		if row.Object == nil {
			continue
		}
		obj := *row.Object
		_, pose := s.selectPose(obj)
		if pose == nil {
			t.Fatal("來源未匹配")
		}
		for _, at := range []image.Point{image.Pt(61, 29), image.Pt(45, 10), image.Pt(267, 152)} {
			obj.X, obj.Y = at.X, at.Y
			// 首物件故意覆蓋同位置。後一物件須還原純地形，不能透出首物件。
			prior := obj
			for i := range prior.Raster {
				prior.Raster[i] = 7
			}
			prior.Backgrounds = [4]byte{}
			d := &caveDecoded{Objects: []caveObject{prior, obj}}
			got, want := makeCanvas(), makeCanvas()
			state := s.drawWithTerrain(got, d, palette, origin, base)
			if state["known"] != 1 || state["unknown"] != 1 {
				t.Fatal("覆蓋來源計數不符", state)
			}
			for y := 0; y < 48; y++ {
				for x := 0; x < 48; x++ {
					px, py := obj.X*3+x+origin.X, obj.Y*3+y+origin.Y
					if !image.Pt(px, py).In(box) {
						continue
					}
					q := y/24*2 + x/24
					old := base.RGBAAt(px, py)
					if obj.Backgrounds[q] == 0 {
						old = color.RGBA{palette[0], palette[1], palette[2], 255}
					}
					c := pose.pixels[y*48+x]
					a, b := uint32(c.A), uint32(255-c.A)
					want.SetRGBA(px, py, color.RGBA{uint8((uint32(c.R)*a + uint32(old.R)*b + 127) / 255), uint8((uint32(c.G)*a + uint32(old.G)*b + 127) / 255), uint8((uint32(c.B)*a + uint32(old.B)*b + 127) / 255), 255})
					if c.A != 0 {
						blended++
					}
				}
			}
			if !bytes.Equal(got.Pix, want.Pix) {
				t.Fatal("背景／黑色／物件順序／裁切／平移像素失配", row.Key, at)
			}
			if obj.Backgrounds != [4]byte{} {
				got, want = makeCanvas(), makeCanvas()
				state = s.draw(got, &caveDecoded{Objects: []caveObject{obj}}, palette, origin)
				caveObjectNative(want, obj, palette, origin, false)
				if state["known"] != 0 || !bytes.Equal(got.Pix, want.Pix) {
					t.Fatal("沒有下層時未原版退回")
				}
			}
			tests++
		}
	}
	t.Log("39組真實物件獨立下層／黑色／重疊／裁切／平移", tests, "混色像素", blended)
}

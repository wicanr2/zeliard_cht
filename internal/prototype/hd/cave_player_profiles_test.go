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

func profileNormalFrames(t *testing.T) []*CavePresentation {
	t.Helper()
	var study struct {
		Sources []struct {
			Directory string `json:"source_directory"`
			Key       string `json:"actor_key"`
			Source    struct {
				File          string
				Steps, Cycles uint64
				BodySteps     uint64 `json:"body_steps"`
				BodyCycles    uint64 `json:"body_cycles"`
				MemorySHA     string `json:"memory_sha256"`
				BodySHA       string `json:"body_memory_sha256"`
				IndexedSHA    string `json:"indexed_sha256"`
				PaletteSHA    string `json:"palette_sha256"`
				Paints        []CavePaint
			}
		}
	}
	if err := json.Unmarshal(iceRead(t, filepath.Join(iceStudy(t), "cave-combined-profile-study-v1/receipt.json")), &study); err != nil {
		t.Fatal(err)
	}
	var frames []*CavePresentation
	for _, row := range study.Sources {
		point := row.Source
		read := func(ext string) []byte {
			return iceRead(t, filepath.Join(iceStudy(t), row.Directory, point.File+"."+ext))
		}
		f := &CavePresentation{Memory: read("memory"), BodyMemory: read("body.memory"), SourceIndexed: read("indexed"), Palette: read("palette"), Steps: point.Steps, Cycles: point.Cycles, BodySteps: point.BodySteps, BodyCycles: point.BodyCycles, Paints: point.Paints}
		if townBackgroundHash(f.Memory) != point.MemorySHA || townBackgroundHash(f.BodyMemory) != point.BodySHA || townBackgroundHash(f.SourceIndexed) != point.IndexedSHA || townBackgroundHash(f.Palette) != point.PaletteSHA {
			t.Fatal("組合來源SHA改變")
		}
		var err error
		f.Decoded, err = decodeCaveFrameWithPaints(f.Memory, f.BodyMemory, f.SourceIndexed, f.Paints)
		if err != nil || f.Decoded.Differences != 0 || townBackgroundHash(f.Decoded.ActorRaster[:]) != row.Key {
			t.Fatal("真實組合來源整幅重建失配", row.Key, err)
		}
		magiaFrameRGB(f)
		frames = append(frames, f)
	}
	if len(frames) != 46 {
		t.Fatal("來源數不符", len(frames))
	}
	return frames
}

func profilePlayerArt(t *testing.T) *CavePlayer {
	t.Helper()
	p := savedPlayerArt(t)
	if err := p.LoadProfiles(filepath.Join(iceStudy(t), "cave-combined-profile-complete-v1")); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCaveProfilePlayerNormalSourcesAndIndependentPixels(t *testing.T) {
	p, art := profilePlayerArt(t), doorTestArt(t)
	art.SetPlayer(p)
	art.SetWeapon(cyanWeaponArt(t))
	seen := map[string]bool{}
	checked, rgba, native := 0, 0, 0
	for _, f := range profileNormalFrames(t) {
		memory, body := townBackgroundHash(f.Memory), townBackgroundHash(f.BodyMemory)
		if !art.Draw(caveOriginalCanvas(f), f, image.Point{}) || art.PlayerState()["matched"] != true {
			t.Fatal("已驗組合來源仍缺主角HD", art.PlayerState())
		}
		if townBackgroundHash(f.Memory) != memory || townBackgroundHash(f.BodyMemory) != body {
			t.Fatal("HD改寫原版來源")
		}
		native += 224 * 144
		key, pose, hidden := p.selectPose(f.Decoded)
		if hidden || pose == nil {
			t.Fatal("來源未選到姿態", key)
		}
		if pose.number < 67 || seen[key] {
			continue
		}
		seen[key] = true
		im, err := png.Decode(bytes.NewReader(iceRead(t, filepath.Join(iceStudy(t), "cave-combined-profile-complete-v1", "player-"+key[:12]+"-72x72.png"))))
		if err != nil {
			t.Fatal(err)
		}
		canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
		old := color.RGBA{163, 23, 91, 255}
		draw.Draw(canvas, canvas.Bounds(), image.NewUniform(old), image.Point{}, draw.Src)
		pose.draw(canvas, f.Decoded, f.Paints, image.Point{})
		for y := 0; y < 72; y++ {
			for x := 0; x < 72; x++ {
				c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
				if pose.pixels[y*72+x] != c {
					t.Fatal("組合圖獨立RGBA失配", key, x, y)
				}
				rgba++
				want := old
				if c.A != 0 && !cavePlayerPixelBlocked(f.Decoded.ActorX+x/3, f.Decoded.ActorY+y/3, f.Paints) {
					a, inv := uint32(c.A), uint32(255-c.A)
					want = color.RGBA{uint8((uint32(c.R)*a + uint32(old.R)*inv + 127) / 255), uint8((uint32(c.G)*a + uint32(old.G)*inv + 127) / 255), uint8((uint32(c.B)*a + uint32(old.B)*inv + 127) / 255), 255}
					checked++
				}
				if got := canvas.RGBAAt(f.Decoded.ActorX*3+x, f.Decoded.ActorY*3+y); got != want {
					t.Fatal("組合圖獨立混色失配", key, x, y, got, want)
				}
			}
		}
		unknown := *f.Decoded
		unknown.ActorRaster[0] ^= 1
		if _, pose, _ := p.selectPose(&unknown); pose != nil {
			t.Fatal("未知鍵借用相似姿態")
		}
	}
	if len(seen) != 5 || len(p.poses) != 72 || rgba != 25920 || checked < 5000 {
		t.Fatal("組合主角覆蓋不足", len(seen), len(p.poses), rgba, checked)
	}
	t.Logf("46真實來源原生%d像素零差；五新圖%d註冊RGBA、%d混色像素相同，原版來源未變", native, rgba, checked)
}

func TestCaveProfilePlayerAtomicLoadAndLegacy(t *testing.T) {
	p := savedPlayerArt(t)
	before := map[string]*cavePlayerPose{}
	for key, pose := range p.poses {
		before[key] = pose
	}
	if err := p.LoadProfiles(t.TempDir()); err == nil || len(p.poses) != 67 {
		t.Fatal("缺圖沒有原子拒絕")
	}
	dir := t.TempDir()
	for n, r := range caveProfilePlayerPoses {
		name := "player-" + r.key[:12] + "-72x72.png"
		data := iceRead(t, filepath.Join(iceStudy(t), "cave-combined-profile-complete-v1", name))
		if n == 4 {
			data = []byte("broken final image")
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.LoadProfiles(dir); err == nil {
		t.Fatal("損壞最後一圖未拒絕")
	}
	for key, pose := range before {
		if p.poses[key] != pose {
			t.Fatal("失敗載入改變原67圖", key)
		}
	}
	if err := p.LoadProfiles(filepath.Join(iceStudy(t), "cave-combined-profile-complete-v1")); err != nil {
		t.Fatal(err)
	}
	for key, pose := range before {
		if p.poses[key] != pose {
			t.Fatal("載入替換原67圖", key)
		}
	}
}

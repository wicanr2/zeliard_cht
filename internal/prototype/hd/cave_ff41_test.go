package hd

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func ff41NormalFrame(t *testing.T) *CavePresentation {
	t.Helper()
	base := filepath.Join(iceStudy(t), "cave-ff41-source-v1")
	read := func(ext string) []byte { return iceRead(t, filepath.Join(base, "source."+ext)) }
	f := &CavePresentation{Memory: read("memory"), BodyMemory: read("body.memory"), SourceIndexed: read("indexed"), Palette: read("palette"), Steps: 202440421, Cycles: 207849857, BodySteps: 202428860, BodyCycles: 207837620}
	if townBackgroundHash(f.Memory) != "c43f08361e64e6abaa0b9f0f1818d42305556a89c5652937e28e60cbc7b52a6c" || townBackgroundHash(f.BodyMemory) != "0fa4da56ec7b063d0185e449410f04a562d224283e3a29ba5f4aa2856472dffc" || townBackgroundHash(f.SourceIndexed) != "b9c80da87e0095ebf2488acdc3e9672b9c2b7c841c8c1151820b486c6ba57988" {
		t.Fatal("真實FF41來源雜湊不符")
	}
	if err := json.Unmarshal(read("paints.json"), &f.Paints); err != nil {
		t.Fatal(err)
	}
	magiaFrameRGB(f)
	var err error
	f.Decoded, err = decodeCaveFrameWithPaints(f.Memory, f.BodyMemory, f.SourceIndexed, f.Paints)
	if err != nil || f.Decoded.Differences != 0 {
		t.Fatal("未改動的原生整幅來源重建失配", err)
	}
	return f
}

func TestCaveFF41InactiveNormalSourceAndHD(t *testing.T) {
	f := ff41NormalFrame(t)
	memory, body := cloneTownBytes(f.Memory), cloneTownBytes(f.BodyMemory)
	if f.BodyMemory[0x1b10+0xff41] != 1 || f.BodyMemory[0x1b10+0xff40] != 0 {
		t.Fatal("正常來源條件改變")
	}
	art := doorTestArt(t)
	art.SetPlayer(savedPlayerArt(t))
	art.SetWeapon(cyanWeaponArt(t))
	root, _ := caveFixtures(t)
	objects, err := NewCaveObjects(root)
	if err != nil {
		t.Fatal(err)
	}
	art.SetObjects(objects)
	canvas := caveOriginalCanvas(f)
	if !art.Draw(canvas, f, image.Point{}) || art.PlayerState()["matched"] != true {
		t.Fatal("真實FF41來源仍未接通既有HD", art.PlayerState())
	}
	// 僅解碼邊界驗證：原版不執行，FF41=0對照不當新增玩家證據。
	zero := *f
	zero.BodyMemory = cloneTownBytes(f.BodyMemory)
	zero.BodyMemory[0x1b10+0xff41] = 0
	zero.Decoded, err = decodeCaveFrameWithPaints(zero.Memory, zero.BodyMemory, zero.SourceIndexed, zero.Paints)
	if err != nil || !reflect.DeepEqual(f.Decoded, zero.Decoded) {
		t.Fatal("FF40=0的FF41值改變原生圖層", err)
	}
	baseline := caveOriginalCanvas(&zero)
	if !art.Draw(baseline, &zero, image.Point{}) || !bytes.Equal(canvas.Pix, baseline.Pix) {
		t.Fatal("未使用的FF41值改變HD像素")
	}
	if !bytes.Equal(memory, f.Memory) || !bytes.Equal(body, f.BodyMemory) {
		t.Fatal("解碼或HD改寫真實來源")
	}
	if out := os.Getenv("ZELIARD_FF41_HD_OUTPUT"); out != "" {
		for name, im := range map[string]image.Image{"original": caveOriginalCanvas(f), "hd": canvas} {
			file, err := os.Create(filepath.Join(out, name+".png"))
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(file, im)
			file.Close()
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Log("原生32256像素零差；未使用值的完整HD畫布576000像素相同；真實來源不變", art.ObjectState())
}

func TestCaveFF41UnsupportedProfilesRemainRejected(t *testing.T) {
	f := ff41NormalFrame(t)
	for _, test := range []struct{ ff41, ff40, shield, e8 byte }{
		{2, 0, 2, 0}, {255, 0, 2, 0}, {1, 255, 2, 0}, {1, 1, 2, 0},
		{1, 0, 4, 0}, {1, 0, 2, 255},
	} {
		body := cloneTownBytes(f.BodyMemory)
		bp := body[0x1b10:0x11b10]
		bp[0xff41], bp[0xff40], bp[0x93], bp[0xe8] = test.ff41, test.ff40, test.shield, test.e8
		if _, err := decodeCaveFrameWithPaints(f.Memory, body, f.SourceIndexed, f.Paints); err == nil {
			t.Fatal("未核分支被接受", test)
		}
	}
}

package hd

import (
	"bytes"
	"encoding/json"
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

type runtimeArtFixture struct{ raw, rgb, resources, extra []byte }

func runtimeArtFixtures(tb testing.TB) (*ManualArtwork, []runtimeArtFixture) {
	tb.Helper()
	root := os.Getenv("ZELIARD_PROJECT")
	if root == "" {
		tb.Skip("需要本地原版與HD來源；skip不算驗收")
	}
	art, err := NewManualArtwork(filepath.Join(root, "workplace/orig"), filepath.Join(root, "workplace/full-hd-v1/manual-hd-v1/registered"))
	if err != nil {
		tb.Fatal(err)
	}
	read := func(p string) []byte {
		data, e := os.ReadFile(p)
		if e != nil {
			tb.Fatal(e)
		}
		return data
	}
	var fixtures []runtimeArtFixture
	for _, name := range []string{"point-000", "point-014"} {
		base := filepath.Join(root, "workplace/full-hd-v1/cave-profile-source-v1", name)
		memory, raw, palette := read(base+".memory"), read(base+".indexed"), read(base+".palette")
		if len(memory) != 1<<20 || len(raw) != 64000 || len(palette) != 768 {
			tb.Fatal("來源尺寸")
		}
		rgb := make([]byte, 192000)
		for i, v := range raw {
			copy(rgb[i*3:i*3+3], palette[int(v)*3:int(v)*3+3])
		}
		fixtures = append(fixtures, runtimeArtFixture{raw, rgb, memory[0x11b10:0x21b10], memory[0x21b10:0x31b10]})
	}
	return art, fixtures
}

// 真實來源的完整輸出與修改前收據比較，不以實作自身的預期值驗收。
func TestRuntimeArtworkPixels(t *testing.T) {
	art, fixtures := runtimeArtFixtures(t)
	for i, f := range fixtures {
		canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
		bitmap.DrawRGB3(canvas, f.rgb, image.Point{})
		art.SetResources(f.resources, f.extra)
		art.Draw(canvas, f.raw, f.rgb, image.Point{})
		name := []string{"profile", "crouch"}[i] + ".rgba"
		if out := os.Getenv("ZELIARD_HD_REFERENCE_OUT"); out != "" {
			if err := os.WriteFile(filepath.Join(out, name), canvas.Pix, 0644); err != nil {
				t.Fatal(err)
			}
		}
		if ref := os.Getenv("ZELIARD_HD_REFERENCE"); ref != "" {
			expected, err := os.ReadFile(filepath.Join(ref, name))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(expected, canvas.Pix) {
				t.Fatal("HD輸出改變", name)
			}
		}
	}
}

func BenchmarkRuntimeArtworkMoving(b *testing.B) {
	art, fixtures := runtimeArtFixtures(b)
	canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fixtures[i%len(fixtures)]
		bitmap.DrawRGB3(canvas, f.rgb, image.Point{})
		art.SetResources(f.resources, f.extra)
		art.Draw(canvas, f.raw, f.rgb, image.Point{})
	}
}

func TestRuntimeDamageSources(t *testing.T) {
	dir := os.Getenv("ZELIARD_DAMAGE_SOURCES")
	if dir == "" {
		t.Skip("需要正常受傷來源；skip不算驗收")
	}
	data, err := os.ReadFile(filepath.Join(dir, "receipt.json"))
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		Points []struct {
			File   string
			Paints []CavePaint
		}
	}
	if err = json.Unmarshal(data, &receipt); err != nil {
		t.Fatal(err)
	}
	if len(receipt.Points) < 3 {
		t.Fatal("受傷相位來源不足")
	}
	root := os.Getenv("ZELIARD_PROJECT")
	art, err := loadCompleteArtwork(filepath.Join(root, "workplace/full-hd-v1/manual-hd-v1/registered"))
	if err != nil {
		t.Fatal(err)
	}
	read := func(path string) []byte {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	for _, p := range receipt.Points {
		t.Run(p.File, func(t *testing.T) {
			memory, body, raw := read(filepath.Join(dir, p.File+".memory")), read(filepath.Join(dir, p.File+".body.memory")), read(filepath.Join(dir, p.File+".indexed"))
			d, e := decodeCaveFrameWithPaints(memory, body, raw, p.Paints)
			if e != nil {
				t.Fatal(e)
			}
			if d.Differences != 0 {
				t.Fatal("受傷原版畫面不符")
			}
			pixels, bounds := art.compose("fman-all", d.ActorRaster[:], 24)
			if pixels == nil || bounds.Empty() {
				t.Fatal("受傷HD未接通")
			}
			t.Logf("完整原版重建零差，受傷HD %v", bounds)
		})
	}
	// 僅此三組相位放行；合成反例只驗來源守衛，不能當成正常玩家證據。
	base := filepath.Join(dir, receipt.Points[0].File)
	for _, off := range []int{0xff38, 0xff40, 0xff39} {
		memory, body, raw := read(base+".memory"), read(base+".body.memory"), read(base+".indexed")
		body[0x1b10+off] = 255
		_, err := decodeCaveFrameWithPaints(memory, body, raw, receipt.Points[0].Paints)
		if err == nil || !strings.Contains(err.Error(), "unobserved combined") {
			t.Fatalf("未知受傷組合仍應拒絕 %04X: %v", off, err)
		}
	}
}

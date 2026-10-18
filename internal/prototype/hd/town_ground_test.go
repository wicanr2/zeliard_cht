package hd

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func groundFixture(t *testing.T) (*TownGround, string) {
	t.Helper()
	scene, base := sceneryTestInputs(t)
	s, err := NewTownGround(scene, filepath.Join(base, "normal-v1/restored.indexed"), filepath.Join(base, "ground-handpaint-v1.png"))
	if err != nil {
		t.Fatal(err)
	}
	return s, base
}
func TestTownGroundOriginalAndDirectSamples(t *testing.T) {
	s, base := groundFixture(t)
	data, err := os.ReadFile(filepath.Join(base, "ground-handpaint-v1.png"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	names := []string{"normal-v1/restored", "normal-v1/citizen-near", "normal-v1/citizen-outside", "scene-cave-v1/cave-near", "normal-v1/cold-town", "scene-cave-v1/cave-entered", "normal-v1/citizen-page", "normal-v1/citizen-last"}
	for _, n := range []string{"1", "2", "3", "4", "5", "6", "7"} {
		names = append(names, "scene-walk-v1/walk-left-"+n)
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			memory, err := os.ReadFile(filepath.Join(base, name+".memory"))
			if err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(filepath.Join(base, name+".indexed"))
			if err != nil {
				t.Fatal(err)
			}
			rgb, dst := skyOriginal(raw, s.scene.palette)
			before := cloneTownBytes(dst.Pix)
			initial := cloneTownBytes(memory)
			p, r, e := memory[0x1b10:0x11b10], memory[0x11b10:0x21b10], memory[0x21b10:0x31b10]
			valid := name != "normal-v1/cold-town" && name != "scene-cave-v1/cave-entered" && name != "normal-v1/citizen-page" && name != "normal-v1/citizen-last" && name != "scene-walk-v1/walk-left-7"
			if s.Draw(dst, 0x13, p, r, e, raw, rgb, s.scene.palette, image.Point{}) != valid {
				t.Fatal("source acceptance", name)
			}
			if !bytes.Equal(memory, initial) {
				t.Fatal("source mutation")
			}
			camera := townSceneWord(p, 0x80)
			changed := 0
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					offset := dst.PixOffset(x, y)
					inside := valid && x >= 144 && x < 816 && y >= 426 && y < 474
					if !inside {
						if !bytes.Equal(before[offset:offset+4], dst.Pix[offset:offset+4]) {
							t.Fatal("outside or rejected write", x, y)
						}
						continue
					}
					if !bytes.Equal(before[offset:offset+4], dst.Pix[offset:offset+4]) {
						changed++
					}
					factor := 1
					if y >= 450 {
						factor = 2
					}
					u := (x - 144 + (camera-153)*24*factor) % 336
					if u < 0 {
						u += 336
					}
					sampled := color.RGBAModel.Convert(source.At(u*2172/336, 273+(y-426)*194/48)).(color.RGBA)
					inv := uint32(255 - sampled.A)
					want := color.RGBA{uint8(uint32(sampled.R) + uint32(before[offset])*inv/255), uint8(uint32(sampled.G) + uint32(before[offset+1])*inv/255), uint8(uint32(sampled.B) + uint32(before[offset+2])*inv/255), 255}
					if dst.RGBAAt(x, y) != want {
						t.Fatal("direct art sampling", x, y)
					}
				}
			}
			if valid && changed == 0 {
				t.Fatal("no HD changes")
			}
		})
	}
}
func TestTownGroundRejectsAtomically(t *testing.T) {
	s, base := groundFixture(t)
	memory, err := os.ReadFile(filepath.Join(base, "normal-v1/restored.memory"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(base, "normal-v1/restored.indexed"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"mode", "primary-short", "resources-short", "extra-short", "indexed-short", "rgb-short", "palette-short", "palette", "scroll-code", "flow-code", "module", "atlas", "mask", "camera", "ground-index", "ground-rgb", "foreground-index", "unknown-group", "canvas"} {
		t.Run(name, func(t *testing.T) {
			p, r, e := cloneTownBytes(memory[0x1b10:0x11b10]), cloneTownBytes(memory[0x11b10:0x21b10]), cloneTownBytes(memory[0x21b10:0x31b10])
			indices := cloneTownBytes(raw)
			rgb, dst := skyOriginal(indices, s.scene.palette)
			palette := cloneTownBytes(s.scene.palette)
			mode := uint8(0x13)
			switch name {
			case "mode":
				mode = 3
			case "primary-short":
				p = p[:3]
			case "resources-short":
				r = r[:3]
			case "extra-short":
				e = e[:3]
			case "indexed-short":
				indices = indices[:3]
			case "rgb-short":
				rgb = rgb[:3]
			case "palette-short":
				palette = palette[:3]
			case "palette":
				palette[767] ^= 1
			case "scroll-code":
				p[0x36a4] ^= 1
			case "flow-code":
				p[0x68ba] ^= 1
			case "module":
				p[0xc000] ^= 1
			case "atlas":
				r[0x5d7f] ^= 1
			case "mask":
				e[0x74bf] ^= 1
			case "camera":
				p[0x80] = 255
				p[0x81] = 255
			case "ground-index":
				indices[142*320+48] ^= 1
			case "ground-rgb":
				rgb[(157*320+271)*3] ^= 1
			case "foreground-index":
				indices[78*320+48] ^= 1
			case "unknown-group":
				p[townSceneWord(p, 0xc00f)+2] = 5
			case "canvas":
				dst = image.NewRGBA(image.Rect(0, 0, 200, 200))
			}
			before := cloneTownBytes(dst.Pix)
			if s.Draw(dst, mode, p, r, e, indices, rgb, palette, image.Point{}) || !bytes.Equal(before, dst.Pix) {
				t.Fatal("non-atomic fallback", name)
			}
		})
	}
}
func TestTownGroundAssetIdentity(t *testing.T) {
	s, base := groundFixture(t)
	ref := filepath.Join(base, "normal-v1/restored.indexed")
	art := filepath.Join(base, "ground-handpaint-v1.png")
	if _, err := NewTownGround(nil, ref, art); err == nil {
		t.Fatal("nil scene")
	}
	if _, err := NewTownGround(s.scene, ref, filepath.Join(t.TempDir(), "absent.png")); err == nil {
		t.Fatal("missing asset")
	}
	for _, name := range []string{"reference", "art"} {
		input := ref
		if name == "art" {
			input = art
		}
		data, err := os.ReadFile(input)
		if err != nil {
			t.Fatal(err)
		}
		data[len(data)-1] ^= 1
		path := filepath.Join(t.TempDir(), "corrupt")
		if err = os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		a, b := ref, art
		if name == "reference" {
			a = path
		} else {
			b = path
		}
		if _, err = NewTownGround(s.scene, a, b); err == nil {
			t.Fatal("corrupt identity", name)
		}
	}
}

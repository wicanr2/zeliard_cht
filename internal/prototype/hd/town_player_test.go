package hd

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTownPlayerFallbackAtomic(t *testing.T) {
	scene, base := sceneryTestInputs(t)
	player, err := NewTownPlayer(scene, base)
	if err != nil {
		t.Fatal(err)
	}
	memory, err := os.ReadFile(filepath.Join(base, "normal-v1/restored.memory"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(base, "normal-v1/restored.indexed"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"accepted", "nil-player", "primary-short", "resources-short", "extra-short", "raw-short", "rgb-short", "palette-short", "palette-changed", "unknown-phase", "unobserved-phase4", "nil-canvas", "unknown-module", "wrong-native-pixel", "small-canvas", "missing-pose", "code-changed"} {
		t.Run(name, func(t *testing.T) {
			p, r, e := cloneTownBytes(memory[0x1b10:0x11b10]), cloneTownBytes(memory[0x11b10:0x21b10]), cloneTownBytes(memory[0x21b10:0x31b10])
			indices := cloneTownBytes(raw)
			pal := cloneTownBytes(scene.palette)
			rgb, dst := skyOriginal(indices, pal)
			candidate := *player
			var renderer *TownPlayer = &candidate
			switch name {
			case "nil-player":
				renderer = nil
			case "primary-short":
				p = p[:100]
			case "resources-short":
				r = r[:100]
			case "extra-short":
				e = e[:100]
			case "raw-short":
				indices = indices[:100]
			case "rgb-short":
				rgb = rgb[:100]
			case "palette-short":
				pal = pal[:100]
			case "palette-changed":
				pal[0] ^= 1
			case "unobserved-phase4":
				p[0xe7] = 4
			case "nil-canvas":
				dst = nil
			case "unknown-phase":
				p[0xe7] = 5
			case "unknown-module":
				p[0xc000] ^= 1
			case "wrong-native-pixel":
				indices[118*320+48] ^= 1
			case "small-canvas":
				dst = image.NewRGBA(image.Rect(0, 0, 100, 100))
			case "missing-pose":
				candidate.art[int(p[0xc2]&1)*5+int(p[0xe7])] = nil
			case "code-changed":
				p[0x6975] ^= 1
			}
			var before []byte
			if dst != nil {
				before = cloneTownBytes(dst.Pix)
			}
			inputs := [][]byte{p, r, e, indices, rgb, pal}
			copies := make([][]byte, len(inputs))
			for i := range inputs {
				copies[i] = cloneTownBytes(inputs[i])
			}
			got := renderer.Draw(dst, 0x13, p, r, e, indices, rgb, pal, image.Point{})
			if got != (name == "accepted") {
				t.Fatalf("匹配%v，案例%s", got, name)
			}
			if !got && dst != nil && !bytes.Equal(before, dst.Pix) {
				t.Fatal("拒絕卻寫畫布")
			}
			for i := range inputs {
				if !bytes.Equal(inputs[i], copies[i]) {
					t.Fatal("改寫來源", i)
				}
			}
			if got {
				box := image.Rect((48+int(p[0x83])*8)*3, 354, (48+int(p[0x83])*8)*3+48, 426)
				for y := 0; y < 600; y++ {
					for x := 0; x < 960; x++ {
						if !image.Pt(x, y).In(box) {
							off := dst.PixOffset(x, y)
							if !bytes.Equal(before[off:off+4], dst.Pix[off:off+4]) {
								t.Fatal("主角框外寫入", x, y)
							}
						}
					}
				}
			}
		})
	}
}

// 比對獨立Python已逐像素驗過的固定離線證據；不把這項測試當GUI。
func TestTownPlayerIndependentCompositions(t *testing.T) {
	scene, base := sceneryTestInputs(t)
	player, err := NewTownPlayer(scene, base)
	if err != nil {
		t.Fatal(err)
	}
	proofDir := filepath.Join(base, "scene-prototype-v1/coherent-frame-v1")
	bg, err := NewTownBackground(os.Getenv("ZELIARD_ORIG"), filepath.Join(base, "normal-v1/restored.indexed"), filepath.Join(base, "normal-v1/restored.palette"), filepath.Join(base, "normal-v1/restored.memory"), filepath.Join(base, "background-handpaint-v2.png"))
	if err != nil {
		t.Fatal(err)
	}
	ground, err := NewTownGround(scene, filepath.Join(base, "normal-v1/restored.indexed"), filepath.Join(base, "ground-handpaint-v1.png"))
	if err != nil {
		t.Fatal(err)
	}
	read := func(path string) []byte {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	checkedProof := func(file, digest string, dst any) {
		b := read(filepath.Join(proofDir, file))
		if townBackgroundHash(b) != digest {
			t.Fatal("獨立證據身份不符", file)
		}
		if err := json.Unmarshal(b, dst); err != nil {
			t.Fatal(err)
		}
	}
	for _, fixture := range []struct{ preview, previewHash, review, reviewHash string }{
		{"player-preview-v4.json", "facdd1bc8d1f92b410d3cff6b544f9b6ee0a4fac2480e61f9a9d1dcd8fe2822d", "player-pixel-review-v5.json", "dfbf1a12e5042c1f732f54d903d407c906bfdde5cbee2b2f8a09ca08f9e6d8f1"},
		{"player-presented-preview-v1.json", "c6cddc70276e7cd1ca313bafbeb4d64c6acb9dbd79be9791df5d88e2cc7cb1ec", "player-presented-pixel-review-v1.json", "e865f4aa4471c58e2030e07b75a0fb391041fbca3a0175253ca7d073531432ff"},
	} {
		var preview struct {
			Records []struct {
				Source, Prefix string
				Memory         string            `json:"memory_sha256"`
				SourceHashes   map[string]string `json:"source_hashes"`
			}
		}
		var review struct {
			Records []struct {
				Source  string
				PNGHash string `json:"after_PNG_sha256"`
			}
		}
		checkedProof(fixture.preview, fixture.previewHash, &preview)
		checkedProof(fixture.review, fixture.reviewHash, &review)
		pngHashes := map[string]string{}
		for _, r := range review.Records {
			pngHashes[r.Source] = r.PNGHash
		}
		for _, r := range preview.Records {
			t.Run(strings.ReplaceAll(r.Source, "/", "-"), func(t *testing.T) {
				var p, resources, e, raw, pal []byte
				if r.SourceHashes != nil {
					parts := map[string][]byte{}
					for suffix, digest := range r.SourceHashes {
						b := read(filepath.Join(proofDir, r.Source+"."+suffix))
						if townBackgroundHash(b) != digest {
							t.Fatal("完成來源身份不符", suffix)
						}
						parts[suffix] = b
					}
					p, resources, e, raw, pal = parts["primary"], parts["resources"], parts["extra"], parts["indexed"], parts["palette"]
				} else {
					memory := read(filepath.Join(base, r.Source+".memory"))
					if townBackgroundHash(memory) != r.Memory {
						t.Fatal("原版記憶體身份不符")
					}
					p, resources, e = memory[0x1b10:0x11b10], memory[0x11b10:0x21b10], memory[0x21b10:0x31b10]
					raw, pal = read(filepath.Join(base, r.Source+".indexed")), scene.palette
				}
				rgb, dst := skyOriginal(raw, pal)
				if !bg.Draw(dst, 0x13, p, resources, raw, rgb, image.Point{}) || !scene.Draw(dst, 0x13, p, resources, e, raw, rgb, image.Point{}, true) || !ground.Draw(dst, 0x13, p, resources, e, raw, rgb, pal, image.Point{}) || !player.Draw(dst, 0x13, p, resources, e, raw, rgb, pal, image.Point{}) {
					t.Fatal("已驗來源合成失配")
				}
				pngBytes := read(filepath.Join(proofDir, r.Prefix+"-after.png"))
				if townBackgroundHash(pngBytes) != pngHashes[r.Source] {
					t.Fatal("獨立圖面身份不符")
				}
				want, err := png.Decode(bytes.NewReader(pngBytes))
				if err != nil {
					t.Fatal(err)
				}
				if want.Bounds() != dst.Bounds() {
					t.Fatal("圖面尺寸不符")
				}
				for y := 0; y < 600; y++ {
					for x := 0; x < 960; x++ {
						a, b, c, d := want.At(x, y).RGBA()
						v := dst.RGBAAt(x, y)
						if a != uint32(v.R)*257 || b != uint32(v.G)*257 || c != uint32(v.B)*257 || d != uint32(v.A)*257 {
							t.Fatal("獨立合成像素不符", x, y)
						}
					}
				}
			})
		}
	}
}

func TestTownPlayerAssetIdentity(t *testing.T) {
	scene, base := sceneryTestInputs(t)
	if _, err := NewTownPlayer(nil, base); err == nil {
		t.Fatal("未拒絕nil前景")
	}
	for _, name := range []string{"missing", "wrong-content"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			files, err := filepath.Glob(filepath.Join(base, "player-c2bit*-phase*-handpaint-v*.png"))
			if err != nil {
				t.Fatal(err)
			}
			bad := "player-c2bit0-phase0-handpaint-v2.png"
			for _, file := range files {
				if filepath.Base(file) != bad {
					if err := os.Symlink(file, filepath.Join(dir, filepath.Base(file))); err != nil {
						t.Fatal(err)
					}
				}
			}
			if name == "wrong-content" {
				if err := os.WriteFile(filepath.Join(dir, bad), []byte("corrupt"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := NewTownPlayer(scene, dir); err == nil {
				t.Fatal("未拒絕圖面身份", name)
			}
		})
	}
}

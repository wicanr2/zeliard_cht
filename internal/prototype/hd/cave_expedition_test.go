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

// 023的正常原版來源，不還原成可執行狀態。此處只驗證唯讀圖面合成。
func TestCaveExpeditionIndependentPixels(t *testing.T) {
	root, _ := caveFixtures(t)
	read := func(path string) []byte {
		b, e := os.ReadFile(filepath.Join(root, path))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	var review struct{ Records []cavePlayerProof }
	if e := json.Unmarshal(read("expedition-composition-preview-v1/receipt.json"), &review); e != nil || len(review.Records) != 18 {
		t.Fatal("十八份獨立來源未齊", e)
	}
	art, e := NewCaveTerrain(root)
	if e != nil {
		t.Fatal(e)
	}
	player, e := NewCavePlayer(root)
	if e != nil {
		t.Fatal(e)
	}
	art.SetPlayer(player)
	visible, hidden, native := 0, 0, 0
	var results []map[string]any
	for _, proof := range review.Records {
		t.Run(proof.Name, func(t *testing.T) {
			prefix := "expedition-source-v2/" + proof.Name
			memory, body, raw, pal := read(prefix+".source.memory"), read(prefix+".body.memory"), read(prefix+".source.indexed"), read(prefix+".source.palette")
			if townBackgroundHash(memory) != proof.MemorySHA || townBackgroundHash(body) != proof.BodySHA || townBackgroundHash(raw) != proof.IndexedSHA {
				t.Fatal("來源身份失配")
			}
			var clocks struct {
				Steps      uint64 `json:"source_steps"`
				Cycles     uint64 `json:"source_cycles"`
				BodySteps  uint64 `json:"body_steps"`
				BodyCycles uint64 `json:"body_cycles"`
			}
			if e := json.Unmarshal(read(prefix+".json"), &clocks); e != nil {
				t.Fatal(e)
			}
			var paints []CavePaint
			if e := json.Unmarshal(read(prefix+".paints.json"), &paints); e != nil {
				t.Fatal(e)
			}
			d, e := decodeCaveFrameWithPaints(memory, body, raw, paints)
			if e != nil {
				t.Fatal(e)
			}
			rgb := make([]byte, 192000)
			for i, v := range raw {
				copy(rgb[i*3:], pal[int(v)*3:int(v)*3+3])
			}
			f := &CavePresentation{Memory: memory, BodyMemory: body, SourceIndexed: raw, Indexed: cloneTownBytes(raw), RGB: rgb, Palette: pal, Decoded: d, Paints: paints, Steps: clocks.Steps, Cycles: clocks.Cycles, BodySteps: clocks.BodySteps, BodyCycles: clocks.BodyCycles}
			beforeMemory, beforeBody, beforeRaw := cloneTownBytes(memory), cloneTownBytes(body), cloneTownBytes(raw)
			canvas := caveOriginalCanvas(f)
			before := cloneTownBytes(canvas.Pix)
			if !art.Draw(canvas, f, image.Point{}) {
				t.Fatal("正常原版來源遭拒絕")
			}
			state := art.PlayerState()
			if state["matched"] != proof.Visible || state["occluded"] != (len(d.Actor) == 0) {
				t.Fatal("可見、全遮擋或原貌退回不同", state)
			}
			if proof.Visible {
				visible++
			} else if len(d.Actor) == 0 {
				hidden++
			} else {
				native++
			}
			expected, e := png.Decode(bytes.NewReader(read("expedition-composition-preview-v1/" + proof.Name + ".png")))
			if e != nil {
				t.Fatal(e)
			}
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := canvas.PixOffset(x, y)
					if image.Pt(x, y).In(image.Rect(144, 42, 816, 474)) {
						if canvas.RGBAAt(x, y) != color.RGBAModel.Convert(expected.At(x, y)).(color.RGBA) {
							t.Fatalf("獨立高清合成不同 %d,%d", x, y)
						}
					} else if !bytes.Equal(canvas.Pix[i:i+4], before[i:i+4]) {
						t.Fatal("寫出遊戲區")
					}
				}
			}
			if !bytes.Equal(memory, beforeMemory) || !bytes.Equal(body, beforeBody) || !bytes.Equal(raw, beforeRaw) {
				t.Fatal("高清回寫原版")
			}
			if proof.Name == "08-lower-vine-2" {
				key := townBackgroundHash(d.ActorRaster[:])
				if key != "39ca530247f1bc39cf0a2f1ac03aa080fc351d31d6a38c9dedd61f0ac09dd9a4" || player.poses[key] != nil || proof.Visible {
					t.Fatal("配色變化誤用正常HD")
				}
			}
			results = append(results, map[string]any{"name": proof.Name, "pixels_compared": 672 * 432, "differences": 0, "sources_unchanged": true, "player": state})
		})
	}
	if visible != 16 || hidden != 1 || native != 1 || len(results) != 18 {
		t.Fatal("正常來源分類未齊", visible, hidden, native)
	}
	if out := os.Getenv("ZELIARD_EXPEDITION_OUTPUT"); out != "" {
		b, e := json.MarshalIndent(map[string]any{"status": "PASS_EXPEDITION_INDEPENDENT_PIXELS", "records": results, "visible_HD": visible, "fully_occluded": hidden, "colour_preserved": native}, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(out, "pixels-receipt.json"), append(b, 10), 0644); e != nil {
			t.Fatal(e)
		}
	}
}

func TestCaveExpeditionMissingAtlasIsAtomic(t *testing.T) {
	root, _ := caveFixtures(t)
	for _, corrupt := range []bool{false, true} {
		dir := t.TempDir()
		for _, name := range []string{"player-atlas-handpaint-v1.png", "player-motion-atlas-handpaint-v1.png", "player-gui-atlas-handpaint-v1.png", "player-left-atlas-handpaint-v2.png", "player-motion-extra-handpaint-v1.png"} {
			if e := os.Symlink(filepath.Join(root, name), filepath.Join(dir, name)); e != nil {
				t.Fatal(e)
			}
		}
		if corrupt {
			if e := os.WriteFile(filepath.Join(dir, "player-expedition-atlas-handpaint-v1.png"), []byte("broken"), 0644); e != nil {
				t.Fatal(e)
			}
		}
		if p, e := NewCavePlayer(dir); p != nil || e == nil {
			t.Fatal("第六張缺失或損壞仍交出部分圖組")
		}
	}
}

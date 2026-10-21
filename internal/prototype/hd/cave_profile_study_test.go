package hd

import (
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// 只供048研究編譯覆蓋使用；不修改原版狀態，不作正式前端或GUI驗收。
func TestCaveCombinedProfileStudy(t *testing.T) {
	out := os.Getenv("ZELIARD_PROFILE_STUDY_OUTPUT")
	if out == "" {
		t.Skip("須指定隔離研究輸出；skip不算驗收")
	}
	player, weapon, art := savedPlayerArt(t), cyanWeaponArt(t), doorTestArt(t)
	art.SetPlayer(player)
	art.SetWeapon(weapon)
	root, _ := caveFixtures(t)
	objects, err := NewCaveObjects(root)
	if err != nil {
		t.Fatal(err)
	}
	art.SetObjects(objects)
	var rows []map[string]any
	seen := map[string]bool{}
	matched, failed := 0, 0
	for _, directory := range []string{"cave-profile-source-v1", "cave-profile-magia-source-v1"} {
		base := filepath.Join(iceStudy(t), directory)
		var receipt struct {
			Points []struct {
				File, Reason  string
				Steps, Cycles uint64
				BodySteps     uint64 `json:"body_steps"`
				BodyCycles    uint64 `json:"body_cycles"`
				MemorySHA     string `json:"memory_sha256"`
				BodySHA       string `json:"body_memory_sha256"`
				IndexedSHA    string `json:"indexed_sha256"`
				PaletteSHA    string `json:"palette_sha256"`
				Paints        []CavePaint
				Signature     string `json:"profile_signature"`
			}
		}
		if err := json.Unmarshal(iceRead(t, filepath.Join(base, "receipt.json")), &receipt); err != nil {
			t.Fatal(err)
		}
		for _, point := range receipt.Points {
			if point.Signature == "" {
				continue
			}
			read := func(ext string) []byte { return iceRead(t, filepath.Join(base, point.File+"."+ext)) }
			f := &CavePresentation{Memory: read("memory"), BodyMemory: read("body.memory"), SourceIndexed: read("indexed"), Palette: read("palette"), Steps: point.Steps, Cycles: point.Cycles, BodySteps: point.BodySteps, BodyCycles: point.BodyCycles, Paints: point.Paints}
			if townBackgroundHash(f.Memory) != point.MemorySHA || townBackgroundHash(f.BodyMemory) != point.BodySHA || townBackgroundHash(f.SourceIndexed) != point.IndexedSHA || townBackgroundHash(f.Palette) != point.PaletteSHA {
				t.Fatal("真實保存來源SHA不符")
			}
			row := map[string]any{"source_directory": directory, "source": point, "decode_error": nil}
			f.Decoded, err = decodeCaveFrameWithPaints(f.Memory, f.BodyMemory, f.SourceIndexed, f.Paints)
			if err != nil {
				row["decode_error"] = err.Error()
				failed++
			} else {
				matched++
				magiaFrameRGB(f)
				canvas := caveOriginalCanvas(f)
				row["terrain_hd"] = art.Draw(canvas, f, image.Point{})
				row["player_hd"], row["weapon_hd"], row["objects_hd"] = art.PlayerState(), art.WeaponState(), art.ObjectState()
			}
			if f.Decoded != nil {
				d := f.Decoded
				key := townBackgroundHash(d.ActorRaster[:])
				row["actor_key"], row["viewport_differences"] = key, d.Differences
				if !seen[key] {
					seen[key] = true
					im := image.NewNRGBA(image.Rect(0, 0, 24, 24))
					for n, value := range d.ActorRaster {
						if value != 255 {
							at := int(value) * 3
							im.SetNRGBA(n%24, n/24, color.NRGBA{f.Palette[at], f.Palette[at+1], f.Palette[at+2], 255})
						}
					}
					if err := os.WriteFile(filepath.Join(out, "actor-"+key[:12]+".raster"), d.ActorRaster[:], 0644); err != nil {
						t.Fatal(err)
					}
					file, err := os.Create(filepath.Join(out, "actor-"+key[:12]+".png"))
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
			rows = append(rows, row)
		}
	}
	report := map[string]any{"schema": 1, "status": "DRAFT_COMBINED_PROFILE_SOURCE_STUDY", "native_matched": matched, "native_failed": failed, "unique_actor_keys": len(seen), "sources": rows, "limits": "原版來源完全未改；研究decoder以編譯覆蓋啟用，不是正式HD或GUI通過。"}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "receipt.json"), append(data, 10), 0644); err != nil {
		t.Fatal(err)
	}
	t.Log("DRAFT_COMBINED_PROFILE_SOURCE_STUDY", matched, "native matched", failed, "native failed", len(seen), "actor keys")
}

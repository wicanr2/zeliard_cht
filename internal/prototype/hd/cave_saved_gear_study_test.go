package hd

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

type savedGearPoint struct {
	File, Key     string
	MemorySHA     string `json:"memory_sha256"`
	BodySHA       string `json:"body_memory_sha256"`
	IndexedSHA    string `json:"indexed_sha256"`
	Steps, Cycles uint64
	BodySteps     uint64      `json:"body_steps"`
	BodyCycles    uint64      `json:"body_cycles"`
	Paints        []CavePaint `json:"native_paints"`
}

// 只重解已存來源，供圖面研究；不是新增玩家路徑或正式呈現驗收。
func TestCaveSavedGearStudy(t *testing.T) {
	out := os.Getenv("ZELIARD_SAVED_GEAR_STUDY")
	if out == "" {
		t.Skip("須指定新的本機研究輸出；skip不算驗收")
	}
	files, err := os.ReadDir(out)
	if err != nil || len(files) != 0 {
		t.Fatal("須使用既有空研究目錄", err)
	}
	project := filepath.Dir(filepath.Dir(iceStudy(t)))
	write := func(name string, data []byte) {
		if err := os.WriteFile(filepath.Join(out, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	export := func(stem string, raw, palette []byte, size int) map[string]any {
		im := image.NewNRGBA(image.Rect(0, 0, size, size))
		box := image.Rectangle{}
		for n, value := range raw {
			if value == 255 {
				continue
			}
			x, y := n%size, n/size
			im.SetNRGBA(x, y, color.NRGBA{palette[int(value)*3], palette[int(value)*3+1], palette[int(value)*3+2], 255})
			box = box.Union(image.Rect(x, y, x+1, y+1))
		}
		write(stem+".raster", raw)
		file, e := os.Create(filepath.Join(out, stem+".png"))
		if e != nil {
			t.Fatal(e)
		}
		e = png.Encode(file, im)
		file.Close()
		if e != nil {
			t.Fatal(e)
		}
		return map[string]any{"file": stem, "key": townBackgroundHash(raw), "rgba_sha256": townBackgroundHash(im.Pix), "size": size, "bounds": box, "palette_sha256": townBackgroundHash(palette)}
	}
	var rows []map[string]any
	seen := map[string]bool{}
	for _, pathName := range []string{"cave-platform", "cave-platform-magia"} {
		base := filepath.Join(iceStudy(t), "cave-door-normal-pair-v1", pathName)
		var receipt struct {
			Player struct {
				Sources []savedGearPoint `json:"missing_sources"`
			} `json:"player_coverage"`
			Weapon struct {
				Sources []savedGearPoint `json:"missing_sources"`
			} `json:"weapon_coverage"`
		}
		data := iceRead(t, filepath.Join(base, "receipt.json"))
		if err := json.Unmarshal(data, &receipt); err != nil {
			t.Fatal(err)
		}
		for _, group := range []struct {
			kind string
			rows []savedGearPoint
		}{{"player", receipt.Player.Sources}, {"weapon", receipt.Weapon.Sources}} {
			for _, point := range group.rows {
				if seen[group.kind+point.Key] {
					continue
				}
				seen[group.kind+point.Key] = true
				read := func(ext string) []byte { return iceRead(t, filepath.Join(base, point.File+"."+ext)) }
				f := &CavePresentation{Memory: read("memory"), BodyMemory: read("body.memory"), SourceIndexed: read("indexed"), Palette: read("palette"), Steps: point.Steps, Cycles: point.Cycles, BodySteps: point.BodySteps, BodyCycles: point.BodyCycles, Paints: point.Paints}
				if townBackgroundHash(f.Memory) != point.MemorySHA || townBackgroundHash(f.BodyMemory) != point.BodySHA || townBackgroundHash(f.SourceIndexed) != point.IndexedSHA {
					t.Fatal("已存來源SHA不符", pathName, point.File)
				}
				if paint, e := os.ReadFile(filepath.Join(base, point.File+".paints.json")); e == nil {
					if err := json.Unmarshal(paint, &f.Paints); err != nil {
						t.Fatal(err)
					}
				}
				d, err := decodeCaveFrameWithPaints(f.Memory, f.BodyMemory, f.SourceIndexed, f.Paints)
				if err != nil || d.Differences != 0 {
					t.Fatal("原生整幅重建失配", pathName, point.File, err)
				}
				raw, size := d.ActorRaster[:], 24
				if group.kind == "weapon" {
					raw, size = d.WeaponRaster[:], 32
				}
				if townBackgroundHash(raw) != point.Key {
					t.Fatal("已存來源鍵改變", point.Key)
				}
				row := export(group.kind+"-"+point.Key[:12], raw, f.Palette, size)
				row["kind"], row["source"] = group.kind, point
				row["source_receipt_sha256"] = townBackgroundHash(data)
				rel, err := filepath.Rel(project, filepath.Join(base, point.File))
				if err != nil {
					t.Fatal(err)
				}
				row["source_prefix"] = rel
				rows = append(rows, row)
			}
		}
	}
	root, _ := caveFixtures(t)
	var known []map[string]any
	seenKnown := map[string]bool{}
	for _, proof := range caveWeaponProofs(t, root) {
		if seenKnown[proof.Key] {
			continue
		}
		seenKnown[proof.Key] = true
		f := caveWeaponFixture(t, root, proof)
		known = append(known, export("known-weapon-"+proof.Key[:12], f.Decoded.WeaponRaster[:], f.Palette, 32))
	}
	report := map[string]any{"schema": 1, "status": "SAVED_GEAR_SOURCE_ONLY", "sources": rows, "known_weapons": known,
		"tool": "Go/dosgolem既有原生來源解碼器", "address_space": "已存dosgolem 1MiB實模式來源；24x24主角與32x32武器的色號／透明Raster，非EXE檔案偏移。",
		"script_sha256": townBackgroundHash(iceRead(t, filepath.Join(project, "internal/prototype/hd/cave_saved_gear_study_test.go"))),
		"limits":        "只重解既有正常來源；不當新正常路徑、HD接入或GUI驗收。"}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	write("receipt.json", append(data, 10))
	fmt.Println("SAVED_GEAR_SOURCE_ONLY", len(rows), "unique saved sources", len(known), "known weapon references")
}

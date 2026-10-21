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

// 只研究已存正常來源，未改正式選圖、遮擋或原版狀態。
func TestCaveUnmatchedAssetsStudy(t *testing.T) {
	out := os.Getenv("ZELIARD_UNMATCHED_ASSETS_STUDY")
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
	export := func(kind, key string, raw, palette []byte, size int) map[string]any {
		if len(raw) != size*size || townBackgroundHash(raw) != key {
			t.Fatal("匯出來源鍵不符", kind, key)
		}
		stem := kind + "-" + key[:12]
		foreground := cloneTownBytes(raw)
		for n, value := range foreground {
			if value == 0 {
				foreground[n] = 255
			}
		}
		bounds := image.Rectangle{}
		for suffix, data := range map[string][]byte{"": raw, ".foreground": foreground} {
			im := image.NewNRGBA(image.Rect(0, 0, size, size))
			for n, value := range data {
				if value != 255 {
					at := int(value) * 3
					im.SetNRGBA(n%size, n/size, color.NRGBA{palette[at], palette[at+1], palette[at+2], 255})
					if suffix == ".foreground" {
						bounds = bounds.Union(image.Rect(n%size, n/size, n%size+1, n/size+1))
					}
				}
			}
			file, err := os.Create(filepath.Join(out, stem+suffix+".png"))
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(file, im)
			file.Close()
			if err != nil {
				t.Fatal(err)
			}
		}
		write(stem+".raster", raw)
		return map[string]any{"kind": kind, "key": key, "file": stem, "size": size, "foreground_key": townBackgroundHash(foreground), "foreground_bounds": bounds, "palette_sha256": townBackgroundHash(palette)}
	}
	type point struct {
		savedGearPoint
		Record, Template int
	}
	seen := map[string]bool{}
	var rows []map[string]any
	for _, name := range []string{"cave-platform", "cave-platform-magia"} {
		base := filepath.Join(iceStudy(t), "cave-profile-normal-pair-v2", name)
		var receipt struct {
			Weapons struct {
				Sources []point `json:"missing_sources"`
			} `json:"weapon_coverage"`
			Objects struct {
				Sources []point `json:"unknown_sources"`
			} `json:"object_coverage"`
		}
		data := iceRead(t, filepath.Join(base, "receipt.json"))
		if err := json.Unmarshal(data, &receipt); err != nil {
			t.Fatal(err)
		}
		for _, group := range []struct {
			kind   string
			points []point
		}{{"weapon", receipt.Weapons.Sources}, {"object", receipt.Objects.Sources}} {
			for _, source := range group.points {
				if seen[group.kind+source.Key] {
					continue
				}
				seen[group.kind+source.Key] = true
				read := func(ext string) []byte { return iceRead(t, filepath.Join(base, source.File+"."+ext)) }
				f := &CavePresentation{Memory: read("memory"), BodyMemory: read("body.memory"), SourceIndexed: read("indexed"), Palette: read("palette"), Steps: source.Steps, Cycles: source.Cycles, BodySteps: source.BodySteps, BodyCycles: source.BodyCycles, Paints: source.Paints}
				if townBackgroundHash(f.Memory) != source.MemorySHA || townBackgroundHash(f.BodyMemory) != source.BodySHA || townBackgroundHash(f.SourceIndexed) != source.IndexedSHA {
					t.Fatal("保存來源SHA不符", name, source.File)
				}
				if err := json.Unmarshal(read("paints.json"), &f.Paints); err != nil {
					t.Fatal(err)
				}
				d, err := decodeCaveFrameWithPaints(f.Memory, f.BodyMemory, f.SourceIndexed, f.Paints)
				if err != nil || d.Differences != 0 {
					t.Fatal("完整原生重建失配", err)
				}
				raw, size := d.WeaponRaster[:], 32
				var object *caveObject
				if group.kind == "object" {
					size = 16
					for n := range d.Objects {
						v := &d.Objects[n]
						if v.Record == source.Record && v.Template == source.Template && townBackgroundHash(v.Raster[:]) == source.Key {
							object = v
							raw = v.Raster[:]
							break
						}
					}
					if object == nil {
						t.Fatal("未找到指定物件", source.Key)
					}
				}
				row := export(group.kind, source.Key, raw, f.Palette, size)
				row["source"], row["source_receipt_sha256"], row["object"] = source, townBackgroundHash(data), object
				prefix, err := filepath.Rel(project, filepath.Join(base, source.File))
				if err != nil {
					t.Fatal(err)
				}
				row["source_prefix"] = prefix
				rows = append(rows, row)
			}
		}
	}
	root, _ := caveFixtures(t)
	knownSeen := map[string]bool{}
	var known []map[string]any
	objects, err := NewCaveObjects(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, proof := range caveObjectProofs(t, root) {
		f := caveObjectFixture(t, root, proof)
		for _, object := range f.Decoded.Objects {
			key, pose := objects.selectPose(object)
			if pose == nil || knownSeen["object"+key] {
				continue
			}
			knownSeen["object"+key] = true
			row := export("known-object", key, object.Raster[:], f.Palette, 16)
			row["object"] = object
			known = append(known, row)
		}
	}
	var gear struct {
		Sources []struct{ Kind, File, Key string }
		Known   []struct{ File, Key string } `json:"known_weapons"`
	}
	gearDir := filepath.Join(iceStudy(t), "cave-saved-gear-source-v1")
	if err := json.Unmarshal(iceRead(t, filepath.Join(gearDir, "receipt.json")), &gear); err != nil {
		t.Fatal(err)
	}
	for _, source := range gear.Sources {
		if source.Kind == "weapon" {
			gear.Known = append(gear.Known, struct{ File, Key string }{source.File, source.Key})
		}
	}
	palette := iceRead(t, filepath.Join(iceStudy(t), "cave-profile-normal-pair-v2/cave-platform/weapon-missing-00.palette"))
	for _, source := range gear.Known {
		if knownSeen["weapon"+source.Key] {
			continue
		}
		knownSeen["weapon"+source.Key] = true
		known = append(known, export("known-weapon", source.Key, iceRead(t, filepath.Join(gearDir, source.File+".raster")), palette, 32))
	}
	report := map[string]any{"schema": 1, "status": "UNMATCHED_ASSETS_SOURCE_ONLY", "sources": rows, "known": known, "script_sha256": townBackgroundHash(iceRead(t, filepath.Join(project, "internal/prototype/hd/cave_unmatched_assets_study_test.go"))), "limits": "只重解保存正常來源，不當新GUI或HD接入。去除0／255的前景鍵只作分類候選，不當透明或別名證據。"}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	write("receipt.json", append(data, 10))
	t.Log("UNMATCHED_ASSETS_SOURCE_ONLY", len(rows), "unknown keys", len(known), "known references")
}

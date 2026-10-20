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

// DRAFT診斷：記錄現有解碼器的真實差分，不把輸出視為驗收通過。
func TestCaveRejectedSourceStudy(t *testing.T) {
	base, out := os.Getenv("ZELIARD_REJECTED_STUDY"), os.Getenv("ZELIARD_REJECTED_OUTPUT")
	if base == "" || out == "" {
		t.Skip("須掛載被拒絕來源及研究輸出；skip不算驗收")
	}
	var receipt struct {
		Points []struct {
			Name, File, Reason string
			MemorySHA          string `json:"memory_sha256"`
			BodySHA            string `json:"body_memory_sha256"`
			Paints             []CavePaint
		}
	}
	if e := json.Unmarshal(iceRead(t, filepath.Join(base, "receipt.json")), &receipt); e != nil {
		t.Fatal(e)
	}
	var points []map[string]any
	for _, point := range receipt.Points {
		read := func(suffix string) []byte { return iceRead(t, filepath.Join(base, point.File+"."+suffix)) }
		memory, body, raw, palette := read("memory"), read("body.memory"), read("indexed"), read("palette")
		if townBackgroundHash(memory) != point.MemorySHA || townBackgroundHash(body) != point.BodySHA {
			t.Fatal("來源不符")
		}
		d, err := decodeCaveFrameWithPaints(memory, body, raw, point.Paints)
		if d == nil {
			t.Fatal("來源未到完整差分階段", err)
		}
		expected := cloneTownBytes(d.Background)
		for _, layer := range [][]cavePixel{d.Actor, d.Weapon} {
			for _, v := range layer {
				expected[v.Y*320+v.X] = v.Color
			}
		}
		bounds := image.Rectangle{}
		var diffs []map[string]int
		colors := map[string]int{}
		for y := 14; y < 158; y++ {
			for x := 48; x < 272; x++ {
				i := y*320 + x
				if expected[i] != raw[i] {
					diffs = append(diffs, map[string]int{"x": x, "y": y, "expected": int(expected[i]), "original": int(raw[i])})
					bounds = bounds.Union(image.Rect(x, y, x+1, y+1))
					colors[fmtPair(expected[i], raw[i])]++
				}
			}
		}
		for label, data := range map[string][]byte{"original": raw, "expected": expected} {
			im := image.NewRGBA(image.Rect(0, 0, 320, 200))
			for i, v := range data {
				at := int(v) * 3
				im.SetRGBA(i%320, i/320, color.RGBA{palette[at], palette[at+1], palette[at+2], 255})
			}
			f, e := os.Create(filepath.Join(out, point.Name+"-"+label+".png"))
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(f, im)
			f.Close()
			if e != nil {
				t.Fatal(e)
			}
		}
		// DRAFT：依IDA的EB60四個七byte記錄與412F非零色號契約試解。
		// 只驗候選重建，不接正式呈現器；仍完整列出修補前差分。
		candidate := cloneTownBytes(expected)
		primary := memory[0x1b10:0x11b10]
		tile, tileErr := caveStatic(memory[0x11b10:0x21b10], 102)
		if tileErr != nil {
			t.Fatal(tileErr)
		}
		var placements []map[string]int
		for slot := 0; slot < 4; slot++ {
			record := primary[0xeb60+slot*7 : 0xeb60+(slot+1)*7]
			position := caveWord(record, 3)
			if record[0] == 255 || record[2] == 0 || position&0x8000 == 0 {
				continue
			}
			x, y := (position&0x7fff)*2%320, (position&0x7fff)*2/320
			if x < 48 || x+8 > 272 || y < 14 || y+8 > 158 {
				t.Fatal("候選位置越界")
			}
			visible, cleared := true, true
			for n, color := range tile {
				if color != 0 {
					i := (y+n/8)*320 + x + n%8
					visible = visible && raw[i] == color
					cleared = cleared && raw[i] == expected[i]
				}
			}
			visibility := -1
			if visible && !cleared {
				visibility = 1
			} else if cleared && !visible {
				visibility = 0
			}
			placements = append(placements, map[string]int{"slot": slot, "x": x, "y": y, "phase": int(record[0]), "life": int(record[2]), "visibility": visibility})
			if visibility != 1 {
				continue
			}
			for n, color := range tile {
				if color != 0 {
					candidate[(y+n/8)*320+x+n%8] = color
				}
			}
		}
		candidateDiff := 0
		for y := 14; y < 158; y++ {
			for x := 48; x < 272; x++ {
				if candidate[y*320+x] != raw[y*320+x] {
					candidateDiff++
				}
			}
		}
		reason := ""
		if err != nil {
			reason = err.Error()
		}
		points = append(points, map[string]any{"name": point.Name, "reason": reason, "differences": diffs, "bounds": bounds, "color_pairs": colors, "actor_x": d.ActorX, "actor_y": d.ActorY, "dynamic_pixels": len(d.Dynamic), "candidate_mapped_difference_count": candidateDiff, "candidate_placements": placements})
		t.Log(point.Name, reason, bounds, colors, "candidate_difference", candidateDiff)
	}
	source := iceRead(t, "cave_rejected_study_test.go")
	if e := os.WriteFile(filepath.Join(out, "study-source.go"), source, 0644); e != nil {
		t.Fatal(e)
	}
	data, e := json.MarshalIndent(map[string]any{"status": "DRAFT_REJECTED_DIFFERENCE_STUDY", "points": points, "study_source_sha256": townBackgroundHash(source)}, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, 10), 0644); e != nil {
		t.Fatal(e)
	}
}

func fmtPair(a, b byte) string {
	return fmt.Sprintf("%d>%d", a, b)
}

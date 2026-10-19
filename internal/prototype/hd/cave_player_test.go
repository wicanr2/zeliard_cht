package hd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestCavePlayerDirectRegisteredPixels(t *testing.T) {
	root, _ := caveFixtures(t)
	player, err := NewCavePlayer(root)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "expedition-art-preview-v2/registration.json"))
	if err != nil {
		t.Fatal(err)
	}
	var review struct {
		Poses []struct {
			Number  int
			Key     string
			Native  []int  `json:"native_bounds"`
			RawPath string `json:"sprite_raw_path"`
		}
	}
	if err = json.Unmarshal(data, &review); err != nil {
		t.Fatal(err)
	}
	if len(review.Poses) != 37 || len(player.poses) != 37 {
		t.Fatal("37姿態圖組未完整載入")
	}
	for _, entry := range review.Poses {
		pose := player.poses[entry.Key]
		if pose == nil {
			t.Fatal("來源鍵不符", entry.Number)
		}
		want, err := os.ReadFile(filepath.Join(root, entry.RawPath))
		if err != nil {
			t.Fatal(err)
		}
		var got []byte
		for _, c := range pose.pixels {
			got = append(got, c.R, c.G, c.B, c.A)
		}
		if !bytes.Equal(got, want) {
			t.Fatal("未直接依獨立Python取樣", entry.Number)
		}
	}
}

type cavePlayerProof struct {
	Name       string
	MemorySHA  string `json:"memory_sha256"`
	BodySHA    string `json:"body_memory_sha256"`
	IndexedSHA string `json:"source_indexed_sha256"`
	Visible    bool   `json:"player_HD"`
}

// 此處比較真實來源圖面；未保存完整時鐘的舊圖面只設定單元測試邊界，
// 不把它稱為原版同時刻配對。正常玩家與完整時鐘由另一路徑驗證。
func cavePlayerFixture(t *testing.T, root string, proof cavePlayerProof) *CavePresentation {
	t.Helper()
	_, points := caveFixtures(t)
	for _, point := range points {
		if point.Name == proof.Name+"-pc-7114" {
			f := cavePresentedFixture(t, root, point)
			if townBackgroundHash(f.Memory) != proof.MemorySHA || townBackgroundHash(f.BodyMemory) != proof.BodySHA || townBackgroundHash(f.SourceIndexed) != proof.IndexedSHA {
				t.Fatal("獨立來源不符", proof.Name)
			}
			return f
		}
	}
	prefixes := map[string]string{"input-replay-v4": "input-replay-v4/source-00", "combat-replay-v5": "combat-replay-v5/source-00", "return-replay-v5": "return-replay-v5/source-01"}
	var paints []CavePaint
	if len(proof.Name) > 11 && proof.Name[:11] == "gui-motion-" {
		data, e := os.ReadFile(filepath.Join(root, "player-gui-missing-v1/index.json"))
		if e != nil {
			t.Fatal(e)
		}
		var index struct {
			Records []struct {
				Prefix string `json:"source_prefix"`
			}
		}
		if e = json.Unmarshal(data, &index); e != nil {
			t.Fatal(e)
		}
		var n int
		if _, e = fmt.Sscanf(proof.Name, "gui-motion-%d", &n); e != nil || n < 0 || n >= len(index.Records) {
			t.Fatal("GUI動作來源索引錯誤")
		}
		prefixes[proof.Name] = index.Records[n].Prefix
		data, e = os.ReadFile(filepath.Join(root, index.Records[n].Prefix+".paints.json"))
		if e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(data, &paints); e != nil {
			t.Fatal(e)
		}
	}
	if len(proof.Name) > 7 && proof.Name[:7] == "motion-" {
		data, e := os.ReadFile(filepath.Join(root, "player-missing-v1/index.json"))
		if e != nil {
			t.Fatal(e)
		}
		var index struct {
			Records []struct {
				Number int
				Prefix string `json:"source_prefix"`
			}
		}
		if e = json.Unmarshal(data, &index); e != nil {
			t.Fatal(e)
		}
		var n int
		if _, e = fmt.Sscanf(proof.Name, "motion-%d", &n); e != nil || n < 0 || n >= len(index.Records) {
			t.Fatal("正常動作來源索引錯誤")
		}
		prefixes[proof.Name] = index.Records[n].Prefix
		data, e = os.ReadFile(filepath.Join(root, index.Records[n].Prefix+".paints.json"))
		if e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(data, &paints); e != nil {
			t.Fatal(e)
		}
	}
	if len(proof.Name) > 13 && proof.Name[:13] == "native-paint-" {
		data, e := os.ReadFile(filepath.Join(root, "native-paints-proof-v1.json"))
		if e != nil {
			t.Fatal(e)
		}
		var native struct {
			Records []struct {
				Source string
				Paints []CavePaint `json:"native_paints"`
			}
		}
		if e = json.Unmarshal(data, &native); e != nil {
			t.Fatal(e)
		}
		var n int
		if _, e = fmt.Sscanf(proof.Name, "native-paint-%d", &n); e != nil || n < 0 || n >= len(native.Records) {
			t.Fatal("原版後畫索引錯誤")
		}
		prefixes[proof.Name] = native.Records[n].Source
		paints = native.Records[n].Paints
	}
	prefix, ok := prefixes[proof.Name]
	if !ok {
		t.Fatal("未列的原版圖面", proof.Name)
	}
	read := func(ext string) []byte {
		b, e := os.ReadFile(filepath.Join(root, prefix+"."+ext))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	memory, body, raw, pal := read("memory"), read("body.memory"), read("indexed"), read("palette")
	if townBackgroundHash(memory) != proof.MemorySHA || townBackgroundHash(body) != proof.BodySHA || townBackgroundHash(raw) != proof.IndexedSHA {
		t.Fatal("原版圖面身份不符", proof.Name)
	}
	d, err := decodeCaveFrameWithPaints(memory, body, raw, paints)
	if err != nil {
		t.Fatal(err)
	}
	rgb := make([]byte, 192000)
	for i, v := range raw {
		copy(rgb[i*3:], pal[int(v)*3:int(v)*3+3])
	}
	var steps, cycles uint64
	for _, p := range paints {
		steps = max(steps, p.Steps)
		cycles = max(cycles, p.Cycles)
	}
	return &CavePresentation{Memory: memory, BodyMemory: body, SourceIndexed: raw, Indexed: cloneTownBytes(raw), RGB: rgb, Palette: pal, Decoded: d, Paints: paints, Steps: steps, Cycles: cycles}
}

func cavePlayerProofs(t *testing.T, root string) []cavePlayerProof {
	t.Helper()
	data, e := os.ReadFile(filepath.Join(root, "player-composition-preview-v4/receipt.json"))
	if e != nil {
		t.Fatal(e)
	}
	var receipt struct{ Records []cavePlayerProof }
	if e = json.Unmarshal(data, &receipt); e != nil {
		t.Fatal(e)
	}
	if len(receipt.Records) != 22 {
		t.Fatal("二十二個獨立來源未齊")
	}
	return receipt.Records
}

func TestCavePlayerIndependentCompletedPixels(t *testing.T) {
	root, _ := caveFixtures(t)
	out := os.Getenv("ZELIARD_CAVE_PLAYER_OUTPUT")
	if out == "" {
		t.Skip("須提供新的主角驗證輸出目錄")
	}
	if err := os.Mkdir(filepath.Join(out, "pixels"), 0755); err != nil {
		t.Fatal(err)
	}
	art, err := NewCaveTerrain(root)
	if err != nil {
		t.Fatal(err)
	}
	player, err := NewCavePlayer(root)
	if err != nil {
		t.Fatal(err)
	}
	art.SetPlayer(player)
	var results []map[string]any
	for _, proof := range cavePlayerProofs(t, root) {
		t.Run(proof.Name, func(t *testing.T) {
			f := cavePlayerFixture(t, root, proof)
			memory, body, raw := cloneTownBytes(f.Memory), cloneTownBytes(f.BodyMemory), cloneTownBytes(f.SourceIndexed)
			canvas := caveOriginalCanvas(f)
			before := cloneTownBytes(canvas.Pix)
			if !art.Draw(canvas, f, image.Point{}) {
				t.Fatal("已驗來源被拒絕")
			}
			state := art.PlayerState()
			wantOccluded := len(f.Decoded.Actor) == 0
			if state["matched"] != proof.Visible || state["occluded"] != wantOccluded {
				t.Fatal("可見與完全遮擋混淆", state)
			}
			fp, e := os.Open(filepath.Join(root, "player-composition-preview-v4", proof.Name+".png"))
			if e != nil {
				t.Fatal(e)
			}
			expected, e := png.Decode(fp)
			fp.Close()
			if e != nil {
				t.Fatal(e)
			}
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := canvas.PixOffset(x, y)
					if image.Pt(x, y).In(image.Rect(144, 42, 816, 474)) {
						if canvas.RGBAAt(x, y) != color.RGBAModel.Convert(expected.At(x, y)).(color.RGBA) {
							t.Fatalf("獨立主角合成不同 %d,%d", x, y)
						}
					} else if !bytes.Equal(canvas.Pix[i:i+4], before[i:i+4]) {
						t.Fatal("主角或地形寫出遊戲區")
					}
				}
			}
			for _, v := range f.Decoded.Weapon {
				for y := 0; y < 3; y++ {
					for x := 0; x < 3; x++ {
						i := canvas.PixOffset(v.X*3+x, v.Y*3+y)
						if !bytes.Equal(canvas.Pix[i:i+4], before[i:i+4]) {
							t.Fatal("原版最後武器被角色覆蓋")
						}
					}
				}
			}
			if !bytes.Equal(memory, f.Memory) || !bytes.Equal(body, f.BodyMemory) || !bytes.Equal(raw, f.SourceIndexed) {
				t.Fatal("高清回寫原版")
			}
			fp, e = os.Create(filepath.Join(out, "pixels", proof.Name+".png"))
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(fp, canvas)
			fp.Close()
			if e != nil {
				t.Fatal(e)
			}
			results = append(results, map[string]any{"name": proof.Name, "independent_pixels_equal": true, "pixels_compared": 672 * 432, "visible_HD": proof.Visible, "player": state, "weapon_preserved": true, "sources_unchanged": true})
		})
	}
	if len(results) != 22 {
		t.Fatal("二十二份主角圖面未完成")
	}
	data, e := json.MarshalIndent(map[string]any{"status": "PASS_CAVE_PLAYER_INDEPENDENT_PIXELS", "records": results, "scope": "十八可見HD、三完全遮擋、一原版配色保留，獨立圖面；非正常原版時鐘配對"}, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, 10), 0644); e != nil {
		t.Fatal(e)
	}
}

func TestCavePlayerFallbackAndMask(t *testing.T) {
	root, _ := caveFixtures(t)
	proofs := cavePlayerProofs(t, root)
	f := cavePlayerFixture(t, root, proofs[0])
	baseline, err := NewCaveTerrain(root)
	if err != nil {
		t.Fatal(err)
	}
	art, err := NewCaveTerrain(root)
	if err != nil {
		t.Fatal(err)
	}
	player, err := NewCavePlayer(root)
	if err != nil {
		t.Fatal(err)
	}
	art.SetPlayer(player)
	want := caveOriginalCanvas(f)
	if !baseline.Draw(want, f, image.Point{}) {
		t.Fatal("舊地形來源失敗")
	}
	key := townBackgroundHash(f.Decoded.ActorRaster[:])
	pose := player.poses[key]
	delete(player.poses, key)
	canvas := caveOriginalCanvas(f)
	if !art.Draw(canvas, f, image.Point{}) || !bytes.Equal(want.Pix, canvas.Pix) || art.PlayerState()["matched"] != false {
		t.Fatal("缺少姿態未完整保留原角色")
	}
	player.poses[key] = pose
	unknown := *f.Decoded
	unknown.ActorRaster[0] ^= 1
	if _, p, hidden := player.selectPose(&unknown); p != nil || hidden {
		t.Fatal("未驗來源借用相似姿態")
	}
	bad := *f
	bad.SourceIndexed = cloneTownBytes(f.SourceIndexed)
	bad.SourceIndexed[100*320+100] ^= 128
	canvas = caveOriginalCanvas(f)
	before := cloneTownBytes(canvas.Pix)
	if art.Draw(canvas, &bad, image.Point{}) || !bytes.Equal(canvas.Pix, before) {
		t.Fatal("來源失配未原子退回")
	}
	for _, corrupt := range []bool{false, true} {
		dir := t.TempDir()
		if corrupt {
			if e := os.WriteFile(filepath.Join(dir, "player-atlas-handpaint-v1.png"), []byte("broken"), 0644); e != nil {
				t.Fatal(e)
			}
		}
		if p, e := NewCavePlayer(dir); e == nil || p != nil {
			t.Fatal("缺圖／壞圖未整組拒絕")
		}
	}
	// 純合成邊界測試：用真實Mask形狀裁另一已驗姿態，不作原版此姿態證據。
	hidden := cavePlayerFixture(t, root, proofs[7])
	if len(hidden.Decoded.Actor) != 0 || len(hidden.Paints) == 0 {
		t.Fatal("完全遮擋原版來源不符")
	}
	// 九次原版呼叫覆蓋整個24×24；純局部合成只取已證實的中央8×8呼叫。
	unitPaints := hidden.Paints[3:4]
	unit := image.NewRGBA(image.Rect(0, 0, 960, 600))
	for i := range unit.Pix {
		unit.Pix[i] = 91
	}
	initial := cloneTownBytes(unit.Pix)
	drawn := 0
	blocked := 0
	pose.draw(unit, hidden.Decoded, unitPaints, image.Point{})
	for y := 0; y < 72; y++ {
		for x := 0; x < 72; x++ {
			px, py := hidden.Decoded.ActorX*3+x, hidden.Decoded.ActorY*3+y
			i := unit.PixOffset(px, py)
			if cavePlayerPixelBlocked(px/3, py/3, unitPaints) {
				blocked++
				if !bytes.Equal(initial[i:i+4], unit.Pix[i:i+4]) {
					t.Fatal("高清越過原版後畫遮罩")
				}
			} else if !bytes.Equal(initial[i:i+4], unit.Pix[i:i+4]) {
				drawn++
			}
		}
	}
	if drawn == 0 || blocked == 0 {
		t.Fatal("局部遮擋合成未覆蓋可見與被遮兩側")
	}
}

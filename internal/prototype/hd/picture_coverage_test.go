package hd

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

// 規格068：以實機截圖（色號＋色盤）直接合成，插畫範圍內兩色以上的8×8格都要畫上HD。
func TestPictureCoverageRuntimeDumps(t *testing.T) {
	root := os.Getenv("ZELIARD_PROJECT")
	if root == "" {
		t.Skip("需要本地素材及實機截圖；skip不算驗收")
	}
	art, err := NewManualArtwork(filepath.Join(root, "workplace/orig"), filepath.Join(root, "workplace/full-hd-v1/manual-hd-v1/registered"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		dump    string
		box     image.Rectangle
		exposed int
		keep    byte // 刻意露出原版的外圍背景色號，0表示沒有
	}{
		{"res-story-long_probe.shot-260010000", image.Rect(16, 16, 304, 120), 0, 0},    // 片頭陽台
		{"res-story-long_probe.shot-412010000", image.Rect(16, 16, 304, 120), 0, 0},    // 片頭公主特寫
		{"res-story-long_probe.shot-604010000", image.Rect(16, 16, 304, 120), 0, 0},    // 片頭公主石像
		{"res-story-long_probe.shot-668010000", image.Rect(16, 16, 304, 120), 0, 0},    // 片頭國王
		{"r10-ending-poke_probe.shot-656010000", image.Rect(44, 18, 148, 118), 0, 0},   // 結局Garland離城
		{"r10-ending-poke_probe.shot-266010000", image.Rect(116, 18, 228, 118), 0, 51}, // 結局國王擁抱公主，洋紅背景維持原版
		{"r10-ending-poke_probe.shot-1256010000", image.Rect(8, 9, 312, 128), 0, 0},    // FIN背景；外框是純色線條，HD與原版同色，只量內部
		{"r10-ending-poke_probe.shot-1286010000", image.Rect(8, 9, 312, 128), 0, 0},    // FIN
		// 正確色序重畫後新增的覆蓋（規格068）。
		{"r10-ending-poke_probe.shot-386010000", image.Rect(88, 16, 232, 120), 0, 0},        // 守護精靈sei，點陣以實機為準
		{"r10-ending-poke_probe.shot-116010000", image.Rect(180, 24, 276, 112), 0, 0},       // 結局第一段公主new1
		{"r10-ending-poke_probe.shot-116010000", image.Rect(44, 24, 140, 112), 0, 0},        // 同畫面主角人像
		{"d-ending-poke_probe.shot-206050000", image.Rect(180, 24, 276, 112), 0, 0},         // new1說話嘴型
		{"stopped-a-ending-poke_probe.shot-236010000", image.Rect(180, 24, 276, 112), 0, 0}, // new1另一嘴型
		{"r10-ending-poke_probe.shot-566010000", image.Rect(44, 24, 140, 112), 0, 0},        // 主角說話中
		{"r10-ending-poke_probe.shot-566010000", image.Rect(180, 24, 276, 112), 0, 0},       // 公主人像himp
		{"t-story-tail_probe.shot-1190050000", image.Rect(180, 24, 276, 112), 0, 0},         // 國王說話中
		{"t-story-tail_probe.shot-1100050000", image.Rect(16, 16, 304, 120), 0, 0},          // yuu1，288×104
		{"t-story-tail_probe.shot-1680050000", image.Rect(64, 16, 260, 112), 0, 0},          // yuu2，196×96
		{"w-story-title_probe.shot-1704650000", image.Rect(32, 8, 288, 200), 0, 0},          // yuu3-base
		{"w-story-title_probe.shot-1710050000", image.Rect(32, 8, 288, 200), 0, 0},          // yuu3-final
		{"w-story-title_probe.shot-1714050000", image.Rect(32, 8, 288, 200), 0, 0},          // 灰光剪影畫格
		{"v-boot-title_probe.shot-3000000", image.Rect(28, 15, 288, 127), 0, 0},             // 標題ttl3
		{"x-boot-threat_probe.shot-167600000", image.Rect(92, 32, 228, 144), 0, 0},          // 魔王宣言，點陣以實機為準
		{"x-boot-threat_probe.shot-172000000", image.Rect(92, 32, 228, 144), 0, 0},          // 魔王說話嘴型
	}
	for _, c := range cases {
		data, err := os.ReadFile(filepath.Join(root, "workplace/picture-coverage-v1/evidence", c.dump+".indexed"))
		if err != nil {
			t.Fatal(err)
		}
		if len(data) != 64768 {
			t.Fatal("截圖格式不符", c.dump)
		}
		raw, palette := data[:64000], data[64000:]
		rgb := make([]byte, 192000)
		for i, v := range raw {
			copy(rgb[i*3:i*3+3], palette[int(v)*3:int(v)*3+3])
		}
		canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
		bitmap.DrawRGB3(canvas, rgb, image.Point{})
		art.SetResources(nil, nil)
		art.SetIntroMode(true)
		art.Draw(canvas, raw, rgb, image.Point{})
		exposed := 0
		for cy := c.box.Min.Y; cy+8 <= c.box.Max.Y; cy += 8 {
			for cx := c.box.Min.X; cx+8 <= c.box.Max.X; cx += 8 {
				colors := map[byte]bool{}
				// 只計非黑色點：HD圖面在黑色處通常也是黑色，與原版無法分辨。
				native, inked := 0, 0
				for y := cy; y < cy+8; y++ {
					for x := cx; x < cx+8; x++ {
						colors[raw[y*320+x]] = true
						if raw[y*320+x] == 0 || c.keep != 0 && raw[y*320+x] == c.keep {
							continue
						}
						inked++
						i := (y*320 + x) * 3
						p := canvas.RGBAAt(x*3+1, y*3+1)
						if p.R == rgb[i] && p.G == rgb[i+1] && p.B == rgb[i+2] {
							native++
						}
					}
				}
				if len(colors) >= 2 && native*2 > inked {
					exposed++
					t.Logf("%s：外露格 (%d,%d)", c.dump, cx, cy)
				}
			}
		}
		if exposed != c.exposed {
			t.Errorf("%s：原版外露 %d 格，預期 %d", c.dump, exposed, c.exposed)
		}
	}
	// 騎士圖蓋住公主日景右下：公主圖可見部分過半相符，由局部母圖補畫可見範圍的每一格。
	// 這區大片黑底與純色框線，HD與原版同色，改查格標記。同一畫面在遊戲中（非片頭模式）不得出現局部母圖。
	data, err := os.ReadFile(filepath.Join(root, "workplace/picture-coverage-v1/evidence/r10-ending-poke_probe.shot-1016010000.indexed"))
	if err != nil {
		t.Fatal(err)
	}
	raw, palette := data[:64000], data[64000:]
	rgb := make([]byte, 192000)
	for i, v := range raw {
		copy(rgb[i*3:i*3+3], palette[int(v)*3:int(v)*3+3])
	}
	canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
	bitmap.DrawRGB3(canvas, rgb, image.Point{})
	art.SetIntroMode(true)
	art.Draw(canvas, raw, rgb, image.Point{})
	var end6 *completePlacement
	for i, p := range art.complete.placements {
		if p.entry.Source == "end6" && p.cells != nil {
			end6 = &art.complete.placements[i]
		}
	}
	if end6 == nil || end6.x != 0 || end6.y != 0 {
		t.Fatal("騎士圖蓋住時沒有局部公主圖")
	}
	for y := 0; y < 112; y++ {
		for x := 0; x < 128; x++ {
			if !end6.cells[y*end6.entry.Width+x] {
				t.Fatal("公主圖可見範圍有格未補", x, y)
			}
		}
	}
	// 騎士圖在(132,20)：左緣與上緣外的交界格只畫與原版相同的非黑色點。
	for y := 16; y < 112; y++ {
		for x := 128; x < 188; x++ {
			if (x < 132 || y < 20) && raw[y*320+x] != 0 && raw[y*320+x] == end6.entry.Raster[y*end6.entry.Width+x] && !end6.cells[y*end6.entry.Width+x] {
				t.Fatal("交界格的可見點未補", x, y)
			}
		}
	}
	if end6.cells[60*end6.entry.Width+150] {
		t.Fatal("被騎士圖蓋住的格仍標記")
	}
	canvas = image.NewRGBA(image.Rect(0, 0, 960, 600))
	bitmap.DrawRGB3(canvas, rgb, image.Point{})
	art.SetIntroMode(false)
	art.Draw(canvas, raw, rgb, image.Point{})
	for _, p := range art.complete.placements {
		if p.cells != nil {
			t.Fatal("遊戲中出現局部母圖", p.entry.File)
		}
	}
}

// 不該出現局部母圖的畫面：幾乎全黑的守護精靈與魔王雙眼（大片黑色的母圖只有黑點相同），
// 以及Garland離開後的空背景（只剩背景抖色相同，HD背景格會帶出人物殘影）。規格068。
func TestPictureCoverageNoFalsePartial(t *testing.T) {
	root := os.Getenv("ZELIARD_PROJECT")
	if root == "" {
		t.Skip("需要本地素材及實機截圖；skip不算驗收")
	}
	art, err := NewManualArtwork(filepath.Join(root, "workplace/orig"), filepath.Join(root, "workplace/full-hd-v1/manual-hd-v1/registered"))
	if err != nil {
		t.Fatal(err)
	}
	for _, dump := range []string{"r10-ending-poke_probe.shot-386010000", "res-story-d_probe.shot-740050000", "res-story-d_probe.shot-540050000", "r10-ending-poke_probe.shot-716010000"} {
		data, err := os.ReadFile(filepath.Join(root, "workplace/picture-coverage-v1/evidence", dump+".indexed"))
		if err != nil {
			t.Fatal(err)
		}
		raw, palette := data[:64000], data[64000:]
		rgb := make([]byte, 192000)
		for i, v := range raw {
			copy(rgb[i*3:i*3+3], palette[int(v)*3:int(v)*3+3])
		}
		canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
		bitmap.DrawRGB3(canvas, rgb, image.Point{})
		art.SetResources(nil, nil)
		art.SetIntroMode(true)
		art.Draw(canvas, raw, rgb, image.Point{})
		for _, p := range art.complete.placements {
			if p.cells != nil {
				t.Errorf("%s：無關母圖被當成局部母圖 %s (%d,%d)", dump, p.entry.File, p.x, p.y)
			}
		}
	}
}

// 規格068補登錄與局部母圖的合成成本：同一張實機截圖，每次都重新比對。
// nostory 移除story-群組並關閉片頭模式，近似補登錄前的比對量；兩者在同一場景交錯執行，減少主機負載變化的影響。
func BenchmarkPictureCoverageDraw(b *testing.B) {
	root := os.Getenv("ZELIARD_PROJECT")
	if root == "" {
		b.Skip("需要本地素材及實機截圖")
	}
	read := func(name string) []byte {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			b.Fatal(err)
		}
		return data
	}
	story := read("workplace/picture-coverage-v1/evidence/res-story-long_probe.shot-412010000.indexed")
	talking := read("workplace/picture-coverage-v1/probe/out/g-story-long/probe.shot-468050000.indexed")
	caveRaw := read("workplace/hd-cave-artifact-v1/replay-source/source.indexed")
	cavePalette := read("workplace/hd-cave-artifact-v1/replay-source/source.palette")
	memory := read("workplace/hd-cave-artifact-v1/replay-source/source.memory")
	scenes := []struct {
		name         string
		raw, palette []byte
		intro, cave  bool
	}{
		{"intro-412M", story[:64000], story[64000:], true, false},
		{"intro-468M-talking", talking[:64000], talking[64000:], true, false},
		{"cave-ada", caveRaw, cavePalette, false, true},
	}
	arts := map[string]*ManualArtwork{}
	for _, variant := range []string{"after", "nostory"} {
		art, err := NewManualArtwork(filepath.Join(root, "workplace/orig"), filepath.Join(root, "workplace/full-hd-v1/manual-hd-v1/registered"))
		if err != nil {
			b.Fatal(err)
		}
		if variant == "nostory" {
			kept := art.complete.entries[:0]
			for _, e := range art.complete.entries {
				if !strings.HasPrefix(e.Source, "story-") {
					kept = append(kept, e)
				}
			}
			art.complete.entries = kept
			art.complete.reindex()
		}
		arts[variant] = art
	}
	for _, sc := range scenes {
		rgb := make([]byte, 192000)
		for i, v := range sc.raw {
			copy(rgb[i*3:i*3+3], sc.palette[int(v)*3:int(v)*3+3])
		}
		native := image.NewRGBA(image.Rect(0, 0, 960, 600))
		bitmap.DrawRGB3(native, rgb, image.Point{})
		canvas := image.NewRGBA(native.Rect)
		for _, variant := range []string{"after", "nostory"} {
			art := arts[variant]
			b.Run(sc.name+"/"+variant, func(b *testing.B) {
				if sc.cave {
					art.SetResources(memory[0x11b10:0x21b10], memory[0x21b10:0x31b10])
				} else {
					art.SetResources(nil, nil)
				}
				art.SetIntroMode(sc.intro && variant == "after")
				for i := 0; i < b.N; i++ {
					copy(canvas.Pix, native.Pix)
					art.complete.lastRaw, art.lastRaw = nil, nil
					art.Draw(canvas, sc.raw, rgb, image.Point{})
				}
			})
		}
	}
}

// 直接Pix繪製與RGBAAt／SetRGBA路徑整幅逐點相同。規格068。
func TestCompletePaintDirectEquivalent(t *testing.T) {
	root := os.Getenv("ZELIARD_PROJECT")
	if root == "" {
		t.Skip("需要本地素材及實機截圖；skip不算驗收")
	}
	art, err := NewManualArtwork(filepath.Join(root, "workplace/orig"), filepath.Join(root, "workplace/full-hd-v1/manual-hd-v1/registered"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { paintDirect = true }()
	for _, dump := range []string{"res-story-long_probe.shot-412010000", "r10-ending-poke_probe.shot-1016010000", "r10-ending-poke_probe.shot-266010000"} {
		data, err := os.ReadFile(filepath.Join(root, "workplace/picture-coverage-v1/evidence", dump+".indexed"))
		if err != nil {
			t.Fatal(err)
		}
		raw, palette := data[:64000], data[64000:]
		rgb := make([]byte, 192000)
		for i, v := range raw {
			copy(rgb[i*3:i*3+3], palette[int(v)*3:int(v)*3+3])
		}
		var out [2]*image.RGBA
		for n, direct := range []bool{true, false} {
			paintDirect = direct
			canvas := image.NewRGBA(image.Rect(0, 0, 960, 680))
			bitmap.DrawRGB3(canvas, rgb, image.Pt(0, 48))
			art.SetResources(nil, nil)
			art.SetIntroMode(true)
			art.complete.lastRaw, art.lastRaw = nil, nil
			art.Draw(canvas, raw, rgb, image.Pt(0, 48))
			out[n] = canvas
		}
		if !bytes.Equal(out[0].Pix, out[1].Pix) {
			t.Fatal("兩條繪製路徑結果不同", dump)
		}
	}
}

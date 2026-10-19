package hd

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

func cavePortalFont(t *testing.T) bitmap.Font {
	t.Helper()
	f, e := bitmap.Load(os.Getenv("ZELIARD_FONT"))
	if e != nil {
		t.Fatal(e)
	}
	return f
}

func configureCavePortal(t *testing.T, art *CaveTerrain) {
	t.Helper()
	if path := os.Getenv("ZELIARD_CAVE_PORTAL_ART"); path != "" {
		if art == nil {
			t.Fatal("入口HD須搭配地形")
		}
		portal, e := NewCavePortal(path, cavePortalFont(t), "Muralla鎮")
		if e != nil {
			t.Fatal(e)
		}
		art.SetPortal(portal)
	}
}

func TestCavePortalIndependentPixels(t *testing.T) {
	root, _ := caveFixtures(t)
	font := cavePortalFont(t)
	portals := map[string]*CavePortal{}
	for name, title := range map[string]string{"original": "MURALLA", "zh": "Muralla鎮"} {
		p, e := NewCavePortal(filepath.Join(root, "portal-handpaint-v1.png"), font, title)
		if e != nil {
			t.Fatal(e)
		}
		var raw []byte
		for _, c := range p.pixels {
			raw = append(raw, c.R, c.G, c.B, c.A)
		}
		want, e := os.ReadFile(filepath.Join(root, "portal-art-preview-v1/portal-"+name+".rgba"))
		if e != nil || !bytes.Equal(raw, want) {
			t.Fatal("獨立圖牌RGBA不同", name, e)
		}
		portals[name] = p
	}
	art, e := NewCaveTerrain(root)
	if e != nil {
		t.Fatal(e)
	}
	objects, e := NewCaveObjects(root)
	if e != nil {
		t.Fatal(e)
	}
	art.SetObjects(objects)
	player, e := NewCavePlayer(root)
	if e != nil {
		t.Fatal(e)
	}
	weapon, e := NewCaveWeapon(root)
	if e != nil {
		t.Fatal(e)
	}
	var j struct {
		Records []struct {
			caveWeaponProof
			Caption   string
			Locations [][]int `json:"portal_placements"`
		}
	}
	b, e := os.ReadFile(filepath.Join(root, "portal-composition-preview-v1/receipt.json"))
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &j); e != nil || len(j.Records) != 57 {
		t.Fatal("57獨立圖面未齊", e)
	}
	var results []map[string]any
	for _, p := range j.Records {
		t.Run(p.Name, func(t *testing.T) {
			f := caveObjectFixture(t, root, caveObjectProof{caveWeaponProof: p.caveWeaponProof})
			art.SetPortal(portals[p.Caption])
			art.SetPlayer(nil)
			art.SetWeapon(weapon)
			if p.Player {
				art.SetPlayer(player)
			}
			canvas := caveOriginalCanvas(f)
			before := bytes.Clone(canvas.Pix)
			memory, body, raw := bytes.Clone(f.Memory), bytes.Clone(f.BodyMemory), bytes.Clone(f.SourceIndexed)
			if !art.Draw(canvas, f, image.Point{}) {
				t.Fatal("已驗來源拒絕")
			}
			state := art.PortalState()
			locations := state["placements"].([]image.Point)
			if state["source_valid"] != true || len(locations) != len(p.Locations) || state["matched"] != (len(p.Locations) > 0) {
				t.Fatal("入口來源狀態不符", state, p.Locations)
			}
			for i, pt := range locations {
				if pt != image.Pt(p.Locations[i][0], p.Locations[i][1]) {
					t.Fatal("入口定位不符", pt, p.Locations)
				}
			}
			fp, e := os.Open(filepath.Join(root, "portal-composition-preview-v1", p.Name+".png"))
			if e != nil {
				t.Fatal(e)
			}
			want, e := png.Decode(fp)
			fp.Close()
			if e != nil {
				t.Fatal(e)
			}
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					if image.Pt(x, y).In(image.Rect(144, 42, 816, 474)) {
						if canvas.RGBAAt(x, y) != color.RGBAModel.Convert(want.At(x, y)).(color.RGBA) {
							t.Fatalf("完整獨立合成不符 %d,%d", x, y)
						}
					} else {
						i := canvas.PixOffset(x, y)
						if !bytes.Equal(canvas.Pix[i:i+4], before[i:i+4]) {
							t.Fatal("區外寫入")
						}
					}
				}
			}
			if !bytes.Equal(memory, f.Memory) || !bytes.Equal(body, f.BodyMemory) || !bytes.Equal(raw, f.SourceIndexed) {
				t.Fatal("覆繪改寫原版來源")
			}
			if len(locations) > 0 {
				locations[0].X++
				if reflect.DeepEqual(locations, art.PortalState()["placements"]) {
					t.Fatal("外部可改寫狀態")
				}
			}
			results = append(results, map[string]any{"name": p.Name, "pixels_compared": 672 * 432, "independent_pixels_equal": true, "portal": art.PortalState(), "original_unchanged": true})
		})
	}
	art.SetPortal(nil)
	if art.PortalState() != nil {
		t.Fatal("撤除後殘留")
	}
	if out := os.Getenv("ZELIARD_CAVE_PORTAL_OUTPUT"); out != "" {
		b, e = json.MarshalIndent(map[string]any{"status": "PASS_PORTAL_INDEPENDENT_PIXELS", "records": results, "scope": "獨立單元圖面；不是新GUI或原版時鐘配對"}, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(b, 10), 0644); e != nil {
			t.Fatal(e)
		}
	}
}

func TestCavePortalSourceAndCaptionFallback(t *testing.T) {
	root, _ := caveFixtures(t)
	font := cavePortalFont(t)
	path := filepath.Join(root, "portal-handpaint-v1.png")
	for _, title := range []string{"", " Muralla鎮", "Muralla鎮\n", "MURALLAMURALLA", "\x00", string([]byte{255})} {
		if p, e := NewCavePortal(path, font, title); p != nil || e == nil {
			t.Fatal("壞圖牌未拒絕", title)
		}
	}
	missing := bitmap.Font{}
	if p, e := NewCavePortal(path, missing, "MURALLA"); p != nil || e == nil {
		t.Fatal("缺字未拒絕")
	}
	if p, e := NewCavePortal(filepath.Join(t.TempDir(), "missing.png"), font, "MURALLA"); p != nil || e == nil {
		t.Fatal("缺圖未拒絕")
	}
	bad := filepath.Join(t.TempDir(), "bad.png")
	if e := os.WriteFile(bad, []byte("broken"), 0644); e != nil {
		t.Fatal(e)
	}
	if p, e := NewCavePortal(bad, font, "MURALLA"); p != nil || e == nil {
		t.Fatal("壞圖未拒絕")
	}
	p := caveObjectProofs(t, root)[0]
	f := caveObjectFixture(t, root, p)
	locations, valid := cavePortalLocations(f.Memory, f.Decoded.Cells)
	if !valid || len(locations) != 1 {
		t.Fatal("固定來源未匹配")
	}
	before := bytes.Clone(f.Memory)
	f.Memory[0x11b10+0x8030+20*48] ^= 1
	if pts, ok := cavePortalLocations(f.Memory, f.Decoded.Cells); ok || len(pts) != 0 {
		t.Fatal("未知圖塊借用素材")
	}
	copy(f.Memory, before)
	anchor := locations[0]
	x, y := (anchor.X-48)/8, (anchor.Y-14)/8
	pseg := f.Memory[0x1b10:0x11b10]
	at := 0xe000 + (caveWord(pseg, 0xff31)-0xe000+y*36+4+x)%0x900
	pseg[at] ^= 1
	if pts, ok := cavePortalLocations(f.Memory, f.Decoded.Cells); !ok || len(pts) != 0 {
		t.Fatal("不完整矩陣仍匹配")
	}
	copy(f.Memory, before)
	portal, e := NewCavePortal(path, font, "MURALLA")
	if e != nil {
		t.Fatal(e)
	}
	art, e := NewCaveTerrain(root)
	if e != nil {
		t.Fatal(e)
	}
	art.SetPortal(portal)
	if !art.Draw(caveOriginalCanvas(f), f, image.Point{}) || art.PortalState()["matched"] != true {
		t.Fatal("來源未恢復")
	}
	badFrame := *f
	badFrame.SourceIndexed = bytes.Clone(f.SourceIndexed)
	badFrame.SourceIndexed[100*320+100] ^= 1
	canvas := caveOriginalCanvas(f)
	pixels := bytes.Clone(canvas.Pix)
	if art.Draw(canvas, &badFrame, image.Point{}) || !bytes.Equal(pixels, canvas.Pix) || art.PortalState() != nil {
		t.Fatal("來源失配未原子退回及撤除")
	}
}

// Malicia至Muralla的固定入口覆繪，來源及文字安全區見規格020。
package hd

import (
	"fmt"
	"image"
	"image/color"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

var cavePortalMatrix = [5][5]byte{
	{21, 22, 23, 24, 25}, {73, 74, 97, 75, 76}, {77, 88, 0, 89, 78},
	{95, 90, 0, 91, 96}, {95, 92, 93, 94, 96},
}

type CavePortal struct {
	pixels []color.NRGBA
	title  string
	width  int
}

func NewCavePortal(path string, font bitmap.Font, title string) (*CavePortal, error) {
	if !utf8.ValidString(title) || title == "" || strings.TrimSpace(title) != title {
		return nil, fmt.Errorf("入口圖牌文字不符")
	}
	width := 0
	for _, r := range title {
		g := font[r]
		if unicode.IsControl(r) || len(g) != 16 && len(g) != 32 {
			return nil, fmt.Errorf("入口圖牌字型缺字或控制碼：%U", r)
		}
		width += len(g) / 16 * 8
		if width+1 > 96 {
			return nil, fmt.Errorf("入口圖牌超出安全區")
		}
	}
	atlas, err := caveArtwork(path, "3a2e060f2726b7acc71ee0a29968cf6586900ed683bf79dffef0d6e6d645e843")
	if err != nil {
		return nil, err
	}
	crop := cavePlayerCrop{Image: atlas, box: image.Rect(18, 40, 1238, 1217)}
	s := &CavePortal{pixels: make([]color.NRGBA, 120*120), title: title, width: width}
	raw := make([]byte, 0, 120*120*4)
	for y := 0; y < 120; y++ {
		for x := 0; x < 120; x++ {
			c := caveArea(crop, x, y, 120, 120)
			s.pixels[y*120+x] = c
			raw = append(raw, c.R, c.G, c.B, c.A)
		}
	}
	if townBackgroundHash(raw) != "005efae3f819da0b21893d56c36c452bfa751d6d3070e797a3a36f696e37ff38" {
		return nil, fmt.Errorf("入口直接高清取樣不符")
	}
	for pass, ink := range []color.NRGBA{{24, 10, 24, 255}, {253, 232, 167, 255}} {
		x, y := (120-width)/2+1-pass, 6+1-pass
		for _, r := range title {
			g := font[r]
			w := len(g) / 16 * 8
			for row := 0; row < 16; row++ {
				for col := 0; col < w; col++ {
					if g[row*(w/8)+col/8]&(0x80>>uint(col%8)) != 0 {
						s.pixels[(y+row)*120+x+col] = ink
					}
				}
			}
			x += w
		}
	}
	return s, nil
}

// 完整矩陣先驗；可見格只提供候選定位，不以單格或顏色決定入口。
func cavePortalLocations(memory []byte, cells []caveCell) ([]image.Point, bool) {
	if len(memory) < 0x21b10 {
		return nil, false
	}
	p, r := memory[0x1b10:0x11b10], memory[0x11b10:0x21b10]
	var pack []byte
	seen := map[byte]bool{}
	var ids []int
	for _, row := range cavePortalMatrix {
		for _, id := range row {
			if id != 0 && !seen[id] {
				seen[id] = true
				ids = append(ids, int(id))
			}
		}
	}
	sort.Ints(ids)
	for _, id := range ids {
		pack = append(pack, byte(id))
		pack = append(pack, r[0x8030+(id-1)*48:0x8030+id*48]...)
	}
	if townBackgroundHash(pack) != "3fd046b8c1281a013bfe1c8dfa627a663fec2eac5cebe7e3df025a63c1db6633" {
		return nil, false
	}
	cell := func(x, y int) byte {
		at := (caveWord(p, 0xff31) - 0xe000 + y*36 + 4 + x) % 0x900
		if at < 0 {
			at += 0x900
		}
		id := p[0xe000+at]
		if id&128 != 0 {
			id = p[0xed20+int(id&127)]
		}
		return id
	}
	candidates := map[image.Point]bool{}
	for _, c := range cells {
		if !seen[c.ID] {
			continue
		}
		for y, row := range cavePortalMatrix {
			for x, id := range row {
				if id == c.ID {
					candidates[image.Pt((c.X-48)/8-x, (c.Y-14)/8-y)] = true
				}
			}
		}
	}
	locations := []image.Point{}
	for anchor := range candidates {
		ok := true
		for y, row := range cavePortalMatrix {
			for x, id := range row {
				if cell(anchor.X+x, anchor.Y+y) != id {
					ok = false
				}
			}
		}
		position := image.Pt(48+anchor.X*8, 14+anchor.Y*8)
		if ok && !(image.Rectangle{Min: position, Max: position.Add(image.Pt(40, 40))}).Intersect(image.Rect(48, 14, 272, 158)).Empty() {
			locations = append(locations, position)
		}
	}
	sort.Slice(locations, func(i, j int) bool {
		if locations[i].Y != locations[j].Y {
			return locations[i].Y < locations[j].Y
		}
		return locations[i].X < locations[j].X
	})
	return locations, true
}

func (s *CavePortal) draw(dst *image.RGBA, f *CavePresentation, d *caveDecoded, origin image.Point) map[string]any {
	locations, valid := cavePortalLocations(f.Memory, d.Cells)
	if len(s.pixels) != 120*120 {
		locations = nil
		valid = false
	}
	for _, p := range locations {
		for y := 0; y < 120; y++ {
			for x := 0; x < 120; x++ {
				px, py := p.X*3+x, p.Y*3+y
				if px < 144 || px >= 816 || py < 42 || py >= 474 {
					continue
				}
				c := s.pixels[y*120+x]
				a := uint32(c.A)
				// 原入口矩形先清原版黑底；物件、人物及武器由呼叫端後畫。
				dst.SetRGBA(px+origin.X, py+origin.Y, color.RGBA{uint8((uint32(c.R)*a + 127) / 255), uint8((uint32(c.G)*a + 127) / 255), uint8((uint32(c.B)*a + 127) / 255), 255})
			}
		}
	}
	return map[string]any{"loaded": true, "matched": len(locations) > 0, "empty": valid && len(locations) == 0, "source_valid": valid, "placements": locations, "caption": s.title, "caption_width": s.width, "source_steps": f.Steps, "body_steps": f.BodySteps}
}

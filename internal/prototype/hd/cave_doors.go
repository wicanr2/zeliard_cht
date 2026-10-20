// 無招牌的共用門框。僅呈現原版已畫出的5×4矩陣，來源與邊界見規格045。
package hd

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"path/filepath"
	"sort"
)

var caveDoorTemplates = [2][20]byte{
	{73, 74, 97, 75, 76, 77, 79, 80, 81, 78, 95, 82, 83, 84, 96, 95, 85, 86, 87, 96},
	{73, 74, 97, 75, 76, 77, 88, 0, 89, 78, 95, 90, 0, 91, 96, 95, 92, 93, 94, 96},
}

type caveDoorPlacement struct {
	Position image.Point
	Marker   byte
	Open     bool
	Record   int
}

type caveDoorArt struct {
	pixels [10][]color.NRGBA
}

func (s *CaveTerrain) LoadDoors(dir string) error {
	if s == nil || len(s.material) != 432*768 {
		return fmt.Errorf("缺少洞窟呈現器")
	}
	hashes := [10]string{
		"829dedb465f86d887ef9abf53759ea265781dcc2e87e9699d13e572006f76a6e", "2e45fd7e5c3097e9d9c94983470953afed77013ab38a513eef0b64abb2ab678f",
		"7a9d6672e69b2008f594f5a454190ab100d2188472207763f16992b3c912e783", "bac1ae10929d9b5fed274b16b791253bc69cb718bd545a9facf27ba52cf12843",
		"d8cfecf421e836b783129a82b2e4a8e592c0960dd8237b3f7ea096d7554a31e5", "ca8888db2a7f62f2027144188e91c668c4dcd46b1ca59de7c9b8aebec8dd4111",
		"4fd0f4157c8309ead87019d722ba35dacaf4847970479abdf095331502d8ce1e", "630c0ab9e63596c748f50b1079b8a2db9b758495d66bacbdc08bda6fbeb8070e",
		"5fc63929f3f7fb4b9f6cc60d9aa59a5226277adac5f9baa1150c4ab6ed283a12", "65bd7beb7122d8328cd70b97f2f1813257bd7268359f04eccaf1b184057b9a08",
	}
	art := &caveDoorArt{}
	for i, hash := range hashes {
		kind := "filled"
		if i%2 != 0 {
			kind = "open"
		}
		file := fmt.Sprintf("door-%d-%s-120x96.png", 97+i/2, kind)
		im, err := checkedTerrainImage(filepath.Join(dir, file), hash, 120, 96)
		if err != nil {
			return err
		}
		for y := 0; y < 96; y++ {
			for x := 0; x < 120; x++ {
				art.pixels[i] = append(art.pixels[i], color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA))
			}
		}
	}
	s.doors, s.doorState = art, nil
	return nil
}

func (s *CaveTerrain) DoorState() map[string]any {
	if s == nil || s.doorState == nil {
		return nil
	}
	state := make(map[string]any, len(s.doorState))
	for key, value := range s.doorState {
		if places, ok := value.([]caveDoorPlacement); ok {
			value = append([]caveDoorPlacement{}, places...)
		}
		state[key] = value
	}
	return state
}

func caveDoorSource(memory []byte) (p, table []byte, ok bool) {
	if len(memory) != 1<<20 {
		return nil, nil, false
	}
	entry, _ := caveTerrainIdentity(memory)
	if entry == nil {
		return nil, nil, false
	}
	p, r := memory[0x1b10:0x11b10], memory[0x11b10:0x21b10]
	if townBackgroundHash(r[0x8c00:0x9350]) != "7b4628728e0b6ce81a25b7bc0bbb3b4eedd640b0da102924b4384da0d38f690b" ||
		townBackgroundHash(p[0x78dd:0x79b4]) != "66229dab219f9becfb4fffe007f408c128afe7b71c5d4ae2da494329b0b6732b" {
		return nil, nil, false
	}
	for n, expected := range caveDoorTemplates {
		actual := p[0x79b4+n*20 : 0x79b4+(n+1)*20]
		for i, value := range actual {
			if i == 2 {
				if value < 97 || value > 101 {
					return nil, nil, false
				}
			} else if value != expected[i] {
				return nil, nil, false
			}
		}
	}
	start, end, width := caveWord(p, 0xc00a), caveWord(p, 0xc00c), caveWord(p, 0xc002)
	if start < 0xc01b || end < start+2 || end > 0xe000 || (end-start-2)%12 != 0 || end-start > 64*12+2 || width == 0 || caveWord(p, 0x80) >= width || caveWord(p, 0xff31) < 0xe000 || caveWord(p, 0xff31) >= 0xe900 {
		return nil, nil, false
	}
	table = p[start:end]
	if !bytes.Equal(table[len(table)-2:], []byte{255, 255}) {
		return nil, nil, false
	}
	for at := 0; at < len(table)-2; at += 12 {
		if caveWord(table, at) >= width || table[at+2] >= 64 || table[at+3]&7 > 4 {
			return nil, nil, false
		}
	}
	return p, table, true
}

func (p caveDoorPlacement) tile(x, y int) byte {
	i := 0
	if p.Open {
		i = 1
	}
	if x == 2 && y == 0 {
		return p.Marker
	}
	return caveDoorTemplates[i][y*5+x]
}

// 表項、完整快取與來源同時符合才接受。原版處理開門與通行，此處不改旗標。
func caveDoorPlacements(memory, body []byte, named []image.Point) ([]caveDoorPlacement, bool) {
	p, table, ok := caveDoorSource(memory)
	bp, bt, bodyOK := caveDoorSource(body)
	if !ok || !bodyOK || len(table) != len(bt) || caveWord(p, 0xc002) != caveWord(bp, 0xc002) || caveCatalogMapsDiffer(memory, body) {
		return nil, false
	}
	for at := 0; at < len(table)-2; at += 12 {
		if !bytes.Equal(table[at:at+4], bt[at:at+4]) {
			return nil, false
		}
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
	width := caveWord(p, 0xc002)
	var found []caveDoorPlacement
	for at := 0; at < len(table)-2; at += 12 {
		wx := (caveWord(table, at) - 1 + width) % width
		x0 := wx - caveWord(p, 0x80) - 4
		y0 := int(table[at+2]) - int(p[0x82]&63)
		for _, x := range []int{x0 - width, x0, x0 + width} {
			for _, y := range []int{y0 - 64, y0, y0 + 64} {
				if x < -4 || x > 27 || y < -3 || y > 17 {
					continue
				}
				place := caveDoorPlacement{Position: image.Pt(48+x*8, 14+y*8), Marker: 97 + table[at+3]&7, Open: table[at+3]&128 != 0, Record: at / 12}
				match := true
				for _, old := range named {
					if place.Position == old.Add(image.Pt(0, 8)) {
						match = false
					}
				}
				for row := 0; row < 4 && match; row++ {
					for column := 0; column < 5; column++ {
						if cell(x+column, y+row) != place.tile(column, row) {
							match = false
							break
						}
					}
				}
				if match {
					found = append(found, place)
				}
			}
		}
	}
	sort.Slice(found, func(i, j int) bool {
		if found[i].Position.Y != found[j].Position.Y {
			return found[i].Position.Y < found[j].Position.Y
		}
		return found[i].Position.X < found[j].Position.X
	})
	return found, true
}

func (s *caveDoorArt) draw(dst *image.RGBA, places []caveDoorPlacement, origin image.Point) {
	for _, p := range places {
		i := int(p.Marker-97) * 2
		if p.Open {
			i++
		}
		for y := 0; y < 96; y++ {
			for x := 0; x < 120; x++ {
				px, py := p.Position.X*3+x, p.Position.Y*3+y
				if px < 144 || px >= 816 || py < 42 || py >= 474 {
					continue
				}
				c := s.pixels[i][y*120+x]
				a := uint32(c.A)
				dst.SetRGBA(px+origin.X, py+origin.Y, color.RGBA{uint8((uint32(c.R)*a + 127) / 255), uint8((uint32(c.G)*a + 127) / 255), uint8((uint32(c.B)*a + 127) / 255), 255})
			}
		}
	}
}

// 共用綠色平台64～66。來源、三byte表與拒絕契約見規格043。
package hd

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"path/filepath"
	"sort"
)

type caveCommonArt struct {
	tiles [3][]color.NRGBA
}

func (s *CaveTerrain) LoadCommon(dir string) error {
	if s == nil || len(s.material) != 432*768 {
		return fmt.Errorf("缺少洞窟呈現器")
	}
	im, err := checkedTerrainImage(filepath.Join(dir, "green-platform-72x24.png"), "1fbae4337206aad65b4176177f21696a7859ecb11b5af6ba5587ec5f89936ec6", 72, 24)
	if err != nil {
		return err
	}
	art := &caveCommonArt{}
	for id := range art.tiles {
		for y := 0; y < 24; y++ {
			for x := 0; x < 24; x++ {
				art.tiles[id] = append(art.tiles[id], color.NRGBAModel.Convert(im.At(id*24+x, y)).(color.NRGBA))
			}
		}
	}
	s.common, s.commonState = art, nil
	return nil
}

func (s *CaveTerrain) CommonState() map[string]any {
	if s == nil || s.commonState == nil {
		return nil
	}
	state := make(map[string]any, len(s.commonState))
	for key, value := range s.commonState {
		if points, ok := value.([]image.Point); ok {
			value = append([]image.Point{}, points...)
		}
		state[key] = value
	}
	return state
}

func caveCommonSource(memory []byte) (primary, table []byte, ok bool) {
	if len(memory) != 1<<20 {
		return nil, nil, false
	}
	entry, _ := caveTerrainIdentity(memory)
	if entry == nil {
		return nil, nil, false
	}
	p, r := memory[0x1b10:0x11b10], memory[0x11b10:0x21b10]
	if townBackgroundHash(r[0x8c00:0x9350]) != "7b4628728e0b6ce81a25b7bc0bbb3b4eedd640b0da102924b4384da0d38f690b" ||
		townBackgroundHash(p[0x7fdc:0x8163]) != "494c0a9b65acd0d48ef6ede08f336d06b9ebdab9e6c418865d5f4c392437ebff" {
		return nil, nil, false
	}
	start, end, width := caveWord(p, 0xc004), caveWord(p, 0xc006), caveWord(p, 0xc002)
	if start < 0xc01b || end <= start || end > 0xe000 || (end-start-2)%3 != 0 || end-start > 64*3+2 || width == 0 || caveWord(p, 0x80) >= width {
		return nil, nil, false
	}
	table = p[start:end]
	if table[len(table)-2] != 255 || table[len(table)-1] != 255 {
		return nil, nil, false
	}
	for i := 0; i < len(table)-2; i += 3 {
		if caveWord(table, i) >= width || table[i+2] >= 64 {
			return nil, nil, false
		}
	}
	return p, table, true
}

// 必須同時有原版表項及快取內完整三格；畫面外的格也核對，不能單格放行。
func caveCommonPlacements(memory, body []byte, cells []caveCell) ([]image.Point, bool) {
	p, table, ok := caveCommonSource(memory)
	_, bodyTable, bodyOK := caveCommonSource(body)
	if !ok || !bodyOK || !bytes.Equal(table, bodyTable) {
		return nil, false
	}
	width := caveWord(p, 0xc002)
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
	seen := map[image.Point]bool{}
	for _, c := range cells {
		if c.ID < 64 || c.ID > 66 {
			continue
		}
		x, y := (c.X-48)/8-int(c.ID-64), (c.Y-14)/8
		for i := 0; i < 3; i++ {
			if cell(x+i, y) != byte(64+i) {
				return nil, false
			}
		}
		worldX := (caveWord(p, 0x80) + 4 + x) % width
		worldY := (int(p[0x82]) + y) & 63
		found := false
		for i := 0; i < len(table)-2; i += 3 {
			if caveWord(table, i) == worldX && int(table[i+2]) == worldY {
				found = true
				break
			}
		}
		if !found {
			return nil, false
		}
		seen[image.Pt(48+x*8, 14+y*8)] = true
	}
	positions := make([]image.Point, 0, len(seen))
	for position := range seen {
		positions = append(positions, position)
	}
	sort.Slice(positions, func(i, j int) bool {
		if positions[i].Y == positions[j].Y {
			return positions[i].X < positions[j].X
		}
		return positions[i].Y < positions[j].Y
	})
	return positions, true
}

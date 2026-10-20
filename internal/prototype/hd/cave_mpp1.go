package hd

import (
	"fmt"
	"image/color"
	"path/filepath"
)

// LoadMPP1只補15..25；既有14格與岩面的圖面及混色方式保持不變。
func (s *CaveTerrain) LoadMPP1(dir string) error {
	if s == nil || len(s.material) != 432*768 {
		return fmt.Errorf("缺少原有洞窟呈現器")
	}
	image, err := checkedTerrainImage(filepath.Join(dir, "mpp1-extra-tiles-v1.png"), "eeb7ce0286a08aafdca396379109a37dd6e03e163680fd58b2714ddfedcbe7cd", 264, 24)
	if err != nil {
		return err
	}
	art := &caveTerrainArt{count: 25, tiles: s.tiles, material: s.material}
	for i := 1; i <= 14; i++ {
		if len(art.tiles[i]) != 576 {
			return fmt.Errorf("缺少原有圖格%d", i)
		}
	}
	for id := 15; id <= 25; id++ {
		for y := 0; y < 24; y++ {
			for x := 0; x < 24; x++ {
				art.tiles[id] = append(art.tiles[id], color.NRGBAModel.Convert(image.At((id-15)*24+x, y)).(color.NRGBA))
			}
		}
	}
	for id := 3; id <= 8; id++ {
		art.materialIDs[id] = true
	}
	for id := 16; id <= 25; id++ {
		art.textIDs[id] = true
	}
	s.mpp1 = art
	return nil
}

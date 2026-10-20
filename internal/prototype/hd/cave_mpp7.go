package hd

import (
	"fmt"
	"image/color"
	"path/filepath"
)

// LoadMPP7依037載61槽；原版空白槽保留透明，白熱尾端圖樣逐格保存。
func (s *CaveTerrain) LoadMPP7(dir string) error {
	if s == nil {
		return fmt.Errorf("缺少洞窟呈現器")
	}
	art, err := loadCaveTerrainArt(dir, "mpp7", "74e79828e3f1921388118ac82de7b1550670de109df4ab0954d7a9dfce5adea7", "b593a591f787a075fb1a9b37bfa458048b744b0cca18992c462ae4a69b8eaf47", 61)
	if err != nil {
		return err
	}
	secondary, err := checkedTerrainImage(filepath.Join(dir, "mpp7-secondary-material-v1.png"), "297dd13c6bcc41807856f979aaaf455fd708f5f99a3a7ee160b758e91b07baa0", 432, 768)
	if err != nil {
		return err
	}
	for y := 0; y < 768; y++ {
		for x := 0; x < 432; x++ {
			art.secondaryMaterial = append(art.secondaryMaterial, color.NRGBAModel.Convert(secondary.At(x, y)).(color.NRGBA))
		}
	}
	for _, id := range []int{4, 5, 6, 7, 12, 13, 14, 15, 16, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29} {
		art.materialIDs[id] = true
	}
	for _, id := range []int{12, 13, 14, 15, 16} {
		art.secondaryIDs[id] = true
	}
	for _, id := range []int{36, 37, 38, 39, 46} {
		art.textIDs[id] = true
	}
	s.mpp7 = art
	return nil
}

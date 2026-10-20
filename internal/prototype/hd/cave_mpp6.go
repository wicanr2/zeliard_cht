package hd

import (
	"fmt"
	"image/color"
	"path/filepath"
)

// LoadMPP6依037分開金黃磚牆、紫色岩面與非連號DORADO字帶。
func (s *CaveTerrain) LoadMPP6(dir string) error {
	if s == nil {
		return fmt.Errorf("缺少洞窟呈現器")
	}
	art, err := loadCaveTerrainArt(dir, "mpp6", "059853dce8d844f218194172e1d6bc1a586371750191f729993e849a4d2e6703", "d96f553170cf4870f68ec20811dfeb424ab4bed935e881dbea298b567ec2f59b", 36)
	if err != nil {
		return err
	}
	secondary, err := checkedTerrainImage(filepath.Join(dir, "mpp6-secondary-material-v1.png"), "5916121f2097de409fc8e744f7af64c8ac8f3764b08c881c44b87bd4b7839c6a", 432, 768)
	if err != nil {
		return err
	}
	for y := 0; y < 768; y++ {
		for x := 0; x < 432; x++ {
			art.secondaryMaterial = append(art.secondaryMaterial, color.NRGBAModel.Convert(secondary.At(x, y)).(color.NRGBA))
		}
	}
	for _, id := range []int{3, 4, 5, 7, 8, 9, 10, 11, 12, 13} {
		art.materialIDs[id] = true
	}
	for _, id := range []int{9, 10, 11, 12, 13} {
		art.secondaryIDs[id] = true
	}
	for _, id := range []int{26, 27, 28, 23, 6} {
		art.textIDs[id] = true
	}
	s.mpp6 = art
	return nil
}

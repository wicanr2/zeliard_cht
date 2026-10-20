package hd

import "fmt"

// LoadMPP3只載037已核的43格；圖格5及紅褐結構不能當成綠色連續材質。
func (s *CaveTerrain) LoadMPP3(dir string) error {
	if s == nil {
		return fmt.Errorf("缺少洞窟呈現器")
	}
	art, err := loadCaveTerrainArt(dir, "mpp3", "40f9be56644f647d6f5b7247aa3339de6e890f8a5a8d5875fd490cfc319aa678", "4ceb8428d6d7231e764bdcccea20520d09bf079424fd645fa58ffc612cc555a6", 43)
	if err != nil {
		return err
	}
	for _, id := range []int{3, 4, 6, 7, 8, 9, 10, 32, 33} {
		art.materialIDs[id] = true
	}
	art.textFirst, art.textLast = 34, 38
	s.mpp3 = art
	return nil
}

package hd

import "fmt"

// LoadMPP5只載037已核的52格，水面、晶柱及尾端圖樣沿各自輪廓呈現。
func (s *CaveTerrain) LoadMPP5(dir string) error {
	if s == nil {
		return fmt.Errorf("缺少洞窟呈現器")
	}
	art, err := loadCaveTerrainArt(dir, "mpp5", "58e76f6b1ef046960a2fabcf4f3268bb1df514d23d9e500ae6a80e2a49b7e02b", "d640e43e0b4dec2aa3235a2b038a64cbb7b18cac9fe139c09ed19b1229633c8b", 52)
	if err != nil {
		return err
	}
	for _, id := range []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 25, 35, 36, 37, 38, 41, 42, 43, 44} {
		art.materialIDs[id] = true
	}
	s.mpp5 = art
	return nil
}

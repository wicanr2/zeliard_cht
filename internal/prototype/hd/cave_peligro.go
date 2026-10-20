package hd

import "fmt"

// LoadPeligro只載037已註冊的MPP2圖面，不推定其他家族或事件用途。
func (s *CaveTerrain) LoadPeligro(dir string) error {
	if s == nil {
		return fmt.Errorf("缺少洞窟呈現器")
	}
	art, err := loadCaveTerrainArt(dir, "peligro", "d5b5b4addc544e008d1b37ea9e2ee24926bf7ee72a1ed346d42410cf134c9e1f", "35d1aa52b716b925bd75f34f85d682a69e9f8aa5729e4e868f2aafc54f91d6d9", 34)
	if err != nil {
		return err
	}
	s.peligro = art
	return nil
}

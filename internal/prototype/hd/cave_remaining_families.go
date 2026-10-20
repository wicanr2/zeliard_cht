package hd

import (
	"fmt"
	"image/color"
	"path/filepath"
	"strings"
)

// 規格037的四組固定圖面；分組只控制呈現，不推定地形或事件語意。
type remainingCaveArtSpec struct {
	count                           int
	atlasSHA, materialSHA, extraSHA string
	materialIDs, extraIDs, textIDs  []int
}

var remainingCaveArtSpecs = map[string]remainingCaveArtSpec{
	"MPP8": {count: 43,
		atlasSHA:    "792e1da6cbbf2a7e4b5ae5bce74d29acde6c3bee260160c5d912dfd4d751a745",
		materialSHA: "2f1582610aff83f95bdbd0cae51175dd4da08552601457cfdce1f30544bfd59e",
		materialIDs: []int{7, 8, 21, 23, 26},
		textIDs:     []int{29, 30, 31, 32, 33, 34, 35, 36, 13, 14}},
	"MPP9": {count: 38,
		atlasSHA:    "235abc0d25a62bd9cdf345b30e74d699d3d86993d64cb33f90f58d21112ce212",
		materialSHA: "cde46e75fdfd0d18b9f4682e54a778527f1dfe8e5b39700442934b3740065ac7",
		materialIDs: []int{19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38}},
	"MPPA": {count: 20,
		atlasSHA:    "21366e96cab4e9b56d1e2235cd37b5be798610c4f1f6f8978cf389554e0af1c8",
		materialSHA: "4e86784dcac7be673d83df2cd60d4453a6bf8e202e729c9c30eda76b2716f267",
		extraSHA:    "2db572fdcb06c866514921659550264aa406bc029a5bb6c6b2ff2972868cf9e3",
		materialIDs: []int{3, 4}, extraIDs: []int{17, 18, 19}},
	"MPPB": {count: 5,
		atlasSHA:    "6e866a7c88551b5a6959a176c4c633a8e58555121b5ad4fb82818ec2411faafe",
		materialSHA: "f3bd7e1c7729fc8f3fb4fb9350e6d7e164c0c72475b2ad058f354672b387921f",
		materialIDs: []int{1, 2, 3}},
}

func (s *CaveTerrain) LoadMPP8(dir string) error { return s.loadRemainingCaveArt(dir, "MPP8") }
func (s *CaveTerrain) LoadMPP9(dir string) error { return s.loadRemainingCaveArt(dir, "MPP9") }
func (s *CaveTerrain) LoadMPPA(dir string) error { return s.loadRemainingCaveArt(dir, "MPPA") }
func (s *CaveTerrain) LoadMPPB(dir string) error { return s.loadRemainingCaveArt(dir, "MPPB") }

func (s *CaveTerrain) loadRemainingCaveArt(dir, family string) error {
	if s == nil {
		return fmt.Errorf("缺少洞窟呈現器")
	}
	cfg, ok := remainingCaveArtSpecs[family]
	if !ok {
		return fmt.Errorf("未知洞窟圖面家族")
	}
	prefix := strings.ToLower(family)
	art, err := loadCaveTerrainArt(dir, prefix, cfg.atlasSHA, cfg.materialSHA, cfg.count)
	if err != nil {
		return err
	}
	if cfg.extraSHA != "" {
		secondary, err := checkedTerrainImage(filepath.Join(dir, prefix+"-secondary-material-v1.png"), cfg.extraSHA, 432, 768)
		if err != nil {
			return err
		}
		for y := 0; y < 768; y++ {
			for x := 0; x < 432; x++ {
				art.secondaryMaterial = append(art.secondaryMaterial, color.NRGBAModel.Convert(secondary.At(x, y)).(color.NRGBA))
			}
		}
	}
	for _, id := range cfg.materialIDs {
		art.materialIDs[id] = true
	}
	for _, id := range cfg.extraIDs {
		art.materialIDs[id], art.secondaryIDs[id] = true, true
	}
	for _, id := range cfg.textIDs {
		art.textIDs[id] = true
	}
	// 完整驗證後才交換一個家族。錯圖不得破壞已載圖面。
	switch family {
	case "MPP8":
		s.mpp8 = art
	case "MPP9":
		s.mpp9 = art
	case "MPPA":
		s.mppa = art
	case "MPPB":
		s.mppb = art
	}
	return nil
}

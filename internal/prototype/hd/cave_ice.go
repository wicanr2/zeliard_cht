package hd

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"path/filepath"
)

// caveTerrainArt只保存目標3倍圖面；原始資源身份由catalog另行核對。
type caveTerrainArt struct {
	tiles               [128][]color.NRGBA
	material            []color.NRGBA
	count               int
	materialIDs         [128]bool
	textFirst, textLast int
}

func checkedTerrainImage(path, sha string, width, height int) (image.Image, error) {
	b, err := checked(path, sha, 8<<20)
	if err != nil {
		return nil, err
	}
	c, err := png.DecodeConfig(bytes.NewReader(b))
	if err != nil || c.Width != width || c.Height != height {
		return nil, fmt.Errorf("洞窟圖面尺寸不符")
	}
	return png.Decode(bytes.NewReader(b))
}

func loadCaveTerrainArt(dir, prefix, atlasSHA, materialSHA string, count int) (*caveTerrainArt, error) {
	if count < 1 || count > 127 {
		return nil, fmt.Errorf("洞窟圖格數越界")
	}
	atlas, err := checkedTerrainImage(filepath.Join(dir, prefix+"-tiles-v1.png"), atlasSHA, count*24, 24)
	if err != nil {
		return nil, err
	}
	material, err := checkedTerrainImage(filepath.Join(dir, prefix+"-material-v1.png"), materialSHA, 432, 768)
	if err != nil {
		return nil, err
	}
	art := &caveTerrainArt{count: count}
	for id := 1; id <= count; id++ {
		for y := 0; y < 24; y++ {
			for x := 0; x < 24; x++ {
				art.tiles[id] = append(art.tiles[id], color.NRGBAModel.Convert(atlas.At((id-1)*24+x, y)).(color.NRGBA))
			}
		}
	}
	for y := 0; y < 768; y++ {
		for x := 0; x < 432; x++ {
			art.material = append(art.material, color.NRGBAModel.Convert(material.At(x, y)).(color.NRGBA))
		}
	}
	return art, nil
}

// LoadIce先驗完整圖帶與材質，全部成功才交換；失敗保留原有圖集。
func (s *CaveTerrain) LoadIce(dir string) error {
	if s == nil {
		return fmt.Errorf("缺少洞窟呈現器")
	}
	art, err := loadCaveTerrainArt(dir, "ice", "4305348c37fa0ff0c5bf0f9e15546f7e5b4bcffdb3416ec16f4beea75d816fee", "9f2dc6eed83d9582edeef8ac4f64124c382c93550b5f6fe295fbdec064d9e5ac", 21)
	if err != nil {
		return err
	}
	for id := 3; id <= 12; id++ {
		art.materialIDs[id] = true
	}
	art.textFirst, art.textLast = 15, 19
	s.ice = art
	return nil
}

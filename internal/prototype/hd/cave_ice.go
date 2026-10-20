package hd

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"path/filepath"
)

// caveIceArt只保存目標3倍畫布的圖面；原始MPP4來源由catalog另行核對。
type caveIceArt struct {
	tiles    [22][]color.NRGBA
	material []color.NRGBA
}

func checkedIceImage(path, sha string, width, height int) (image.Image, error) {
	b, err := checked(path, sha, 8<<20)
	if err != nil {
		return nil, err
	}
	c, err := png.DecodeConfig(bytes.NewReader(b))
	if err != nil || c.Width != width || c.Height != height {
		return nil, fmt.Errorf("冰窟圖面尺寸不符")
	}
	return png.Decode(bytes.NewReader(b))
}

// LoadIce先驗完整圖帶與材質，全部成功才交換；失敗保留原有圖集。
func (s *CaveTerrain) LoadIce(dir string) error {
	if s == nil {
		return fmt.Errorf("缺少洞窟呈現器")
	}
	atlas, err := checkedIceImage(filepath.Join(dir, "ice-tiles-v1.png"), "4305348c37fa0ff0c5bf0f9e15546f7e5b4bcffdb3416ec16f4beea75d816fee", 504, 24)
	if err != nil {
		return err
	}
	material, err := checkedIceImage(filepath.Join(dir, "ice-material-v1.png"), "9f2dc6eed83d9582edeef8ac4f64124c382c93550b5f6fe295fbdec064d9e5ac", 432, 768)
	if err != nil {
		return err
	}
	art := &caveIceArt{}
	for id := 1; id <= 21; id++ {
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
	s.ice = art
	return nil
}

package hd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

var townSkyBox = image.Rect(48, 24, 272, 78)

// TownSky只覆繪兩個已證實模組的相同遠景平面，未知變化保留原貌。
// 來源及座標空間見docs/spec/005-town-skyline-hd.md。
type TownSky struct {
	indices, rgb []byte
	prefixes     [2][]byte
	art          *image.RGBA
}

func skyCrop(data []byte, channels int) []byte {
	if len(data) != nativeWidth*nativeHeight*channels {
		return nil
	}
	out := make([]byte, townSkyBox.Dx()*townSkyBox.Dy()*channels)
	for y := townSkyBox.Min.Y; y < townSkyBox.Max.Y; y++ {
		copy(out[(y-townSkyBox.Min.Y)*townSkyBox.Dx()*channels:], data[(y*nativeWidth+townSkyBox.Min.X)*channels:(y*nativeWidth+townSkyBox.Max.X)*channels])
	}
	return out
}

func NewTownSky(root, reference, palette, artwork string) (*TownSky, error) {
	sar, err := checked(filepath.Join(root, "ZELRES2.SAR"), "bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1", 1<<20)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(reference)
	if err != nil {
		return nil, err
	}
	crop := skyCrop(raw, 1)
	if len(crop) == 0 || fmt.Sprintf("%x", sha256.Sum256(crop)) != "22fb614b02b3d9ac436c85b7d56c62ce9004d1254f6c52160de2cf62edeb7cf9" {
		return nil, fmt.Errorf("城鎮遠山原版畫格不符")
	}
	pal, err := checked(palette, "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c", 768)
	if err != nil {
		return nil, err
	}
	data, err := checked(artwork, "275f14de1b18ea6dea297999cdf76e2574bcdc9a4a975782d5c01901480058de", 16<<20)
	if err != nil {
		return nil, err
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width != 1973 || cfg.Height != 797 {
		return nil, fmt.Errorf("城鎮遠山HD圖面尺寸不符")
	}
	art, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	s := &TownSky{indices: crop, rgb: make([]byte, len(crop)*3)}
	for i, v := range crop {
		copy(s.rgb[i*3:i*3+3], pal[int(v)*3:int(v)*3+3])
	}
	for i, offset := range []int{0x38a1a, 0x392dc} {
		s.prefixes[i] = append([]byte(nil), sar[offset:offset+64]...)
	}
	s.art = image.NewRGBA(image.Rect(0, 0, townSkyBox.Dx()*scale, townSkyBox.Dy()*scale))
	for y := 0; y < s.art.Bounds().Dy(); y++ {
		for x := 0; x < s.art.Bounds().Dx(); x++ {
			s.art.Set(x, y, art.At(x*1973/s.art.Bounds().Dx(), 207+y*382/s.art.Bounds().Dy()))
		}
	}
	return s, nil
}

func (s *TownSky) Match(mode uint8, prefix, indexed, rgb []byte) bool {
	known := false
	for _, p := range s.prefixes {
		known = known || bytes.Equal(prefix, p)
	}
	return mode == 0x13 && known && bytes.Equal(skyCrop(indexed, 1), s.indices) && bytes.Equal(skyCrop(rgb, 3), s.rgb)
}

func (s *TownSky) Draw(dst *image.RGBA, mode uint8, prefix, indexed, rgb []byte, origin image.Point) bool {
	box := s.art.Bounds().Add(origin.Add(townSkyBox.Min.Mul(scale)))
	if !box.In(dst.Bounds()) || !s.Match(mode, prefix, indexed, rgb) {
		return false
	}
	draw.Draw(dst, box, s.art, image.Point{}, draw.Src)
	return true
}

package hd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/png"
	"os"
)

var sageBox = image.Rect(56, 23, 152, 87)

const sageArtSHA = "a1fa3b745422f59153239b903591db104c5c3fb320c18ee7f5181a247c33a5aa"

// Sage 只替換已觀察的兩個原版畫格；四個變動像素保留原版動畫。
// 不替這四個像素猜測嘴部、眼睛或魔法等語意。
type Sage struct {
	indices  [2][]byte
	rgb      [2][]byte
	animated []bool
	art      *image.RGBA
}

func sageCrop(raw []byte, channels int) []byte {
	if len(raw) != nativeWidth*nativeHeight*channels {
		return nil
	}
	out := make([]byte, sageBox.Dx()*sageBox.Dy()*channels)
	for y := sageBox.Min.Y; y < sageBox.Max.Y; y++ {
		copy(out[(y-sageBox.Min.Y)*sageBox.Dx()*channels:], raw[(y*nativeWidth+sageBox.Min.X)*channels:(y*nativeWidth+sageBox.Max.X)*channels])
	}
	return out
}

func NewSage(referenceA, referenceB, palette, artwork string) (*Sage, error) {
	pal, err := checked(palette, "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c", 768)
	if err != nil {
		return nil, err
	}
	s := &Sage{}
	want := []string{"bbc79b87bc9d9b3bb59a049d5acd8206a7b9dfb287d557b71437672544d3f16c", "0762364b30958db2e1948f850d449ff042a8f736ea5b62d8674be9296c90a66a"}
	for i, path := range []string{referenceA, referenceB} {
		raw, e := os.ReadFile(path)
		if e != nil {
			return nil, e
		}
		crop := sageCrop(raw, 1)
		if len(crop) == 0 || fmt.Sprintf("%x", sha256.Sum256(crop)) != want[i] {
			return nil, fmt.Errorf("長老原版畫格不符")
		}
		s.indices[i] = crop
		s.rgb[i] = make([]byte, len(crop)*3)
		for n, v := range crop {
			copy(s.rgb[i][n*3:n*3+3], pal[int(v)*3:int(v)*3+3])
		}
	}
	s.animated = make([]bool, len(s.indices[0]))
	for i, v := range s.indices[0] {
		s.animated[i] = v != s.indices[1][i]
	}
	data, err := checked(artwork, sageArtSHA, 16<<20)
	if err != nil {
		return nil, err
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width != 1536 || cfg.Height != 1024 {
		return nil, fmt.Errorf("長老 HD 圖面須為本輪 1536×1024 PNG")
	}
	art, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	s.art = image.NewRGBA(image.Rect(0, 0, sageBox.Dx()*scale, sageBox.Dy()*scale))
	for y := 0; y < s.art.Bounds().Dy(); y++ {
		for x := 0; x < s.art.Bounds().Dx(); x++ {
			s.art.Set(x, y, art.At(x*cfg.Width/s.art.Bounds().Dx(), y*cfg.Height/s.art.Bounds().Dy()))
		}
	}
	return s, nil
}

func (s *Sage) Draw(dst *image.RGBA, indexed, rgb []byte, origin image.Point) (bool, int) {
	idx, col := sageCrop(indexed, 1), sageCrop(rgb, 3)
	for pose := range s.indices {
		if !bytes.Equal(idx, s.indices[pose]) || !bytes.Equal(col, s.rgb[pose]) {
			continue
		}
		at := origin.Add(sageBox.Min.Mul(scale))
		for y := scale; y < (sageBox.Dy()-1)*scale; y++ {
			for x := scale; x < (sageBox.Dx()-1)*scale; x++ {
				if s.animated[y/scale*sageBox.Dx()+x/scale] {
					continue
				}
				dst.Set(at.X+x, at.Y+y, s.art.At(x, y))
			}
		}
		return true, pose
	}
	return false, -1
}

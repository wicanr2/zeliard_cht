package hd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/png"
	"os"
)

var merchantBox = image.Rect(56, 23, 152, 87)

// Merchant 只辨識正常商店路徑已觀察的三個完整肖像畫格。
// 姿態由原版畫格選擇，不新增計時器或改動原版動畫。
type Merchant struct {
	indices [3][]byte
	rgb     [3][]byte
	art     [3]*image.RGBA
}

func merchantCrop(raw []byte, channels int) []byte {
	if len(raw) != nativeWidth*nativeHeight*channels {
		return nil
	}
	out := make([]byte, merchantBox.Dx()*merchantBox.Dy()*channels)
	for y := merchantBox.Min.Y; y < merchantBox.Max.Y; y++ {
		copy(out[(y-merchantBox.Min.Y)*merchantBox.Dx()*channels:], raw[(y*nativeWidth+merchantBox.Min.X)*channels:(y*nativeWidth+merchantBox.Max.X)*channels])
	}
	return out
}

func NewMerchant(referenceA, referenceB, referenceC, palette, artworkA, artworkB, artworkC string) (*Merchant, error) {
	pal, err := checked(palette, "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c", 768)
	if err != nil {
		return nil, err
	}
	m := &Merchant{}
	want := [3]string{"77dcc8e94e5125990656f7b33cc04ed7fa9a9719c20487b37424e3538a8a6baf", "bc69c9d588c150a9ecc491afba4c064b327d2b8306eb0dd6f8eecaaedb52d4d8", "cc24c5722e583f61aa229b2c3a18cba1fe365c170da7a42a8655ed761859a8e6"}
	for pose, path := range []string{referenceA, referenceB, referenceC} {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		crop := merchantCrop(raw, 1)
		if len(crop) == 0 || fmt.Sprintf("%x", sha256.Sum256(crop)) != want[pose] {
			return nil, fmt.Errorf("武器商人原版肖像畫格不符")
		}
		m.indices[pose] = crop
		m.rgb[pose] = make([]byte, len(crop)*3)
		for i, value := range crop {
			copy(m.rgb[pose][i*3:i*3+3], pal[int(value)*3:int(value)*3+3])
		}
	}
	wantArt := [3]string{"2660fc7892f9a5229d4cee3723e992146992a460a1678197793cc890b0dfb59b", "90243f65aef8532303b12dbbc3e1ddd9bd06c03eb58e14548a2d15419892946f", "48f567a000f2d6c413f8090b72417763c7ff96a7a7a24b1dea059dc5fb2a0ec8"}
	for pose, path := range []string{artworkA, artworkB, artworkC} {
		data, err := checked(path, wantArt[pose], 16<<20)
		if err != nil {
			return nil, err
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil || cfg.Width != 1536 || cfg.Height != 1024 {
			return nil, fmt.Errorf("武器商人 HD 圖面須為本輪1536×1024 PNG")
		}
		art, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		// 直接合成到內框282×186，保留原版黃色外框。
		target := image.NewRGBA(image.Rect(0, 0, merchantBox.Inset(1).Dx()*scale, merchantBox.Inset(1).Dy()*scale))
		for y := 0; y < target.Bounds().Dy(); y++ {
			for x := 0; x < target.Bounds().Dx(); x++ {
				target.Set(x, y, art.At(x*cfg.Width/target.Bounds().Dx(), y*cfg.Height/target.Bounds().Dy()))
			}
		}
		m.art[pose] = target
	}
	return m, nil
}

func (m *Merchant) Draw(dst *image.RGBA, indexed, rgb []byte, origin image.Point) (bool, int) {
	idx, col := merchantCrop(indexed, 1), merchantCrop(rgb, 3)
	for pose := range m.indices {
		if !bytes.Equal(idx, m.indices[pose]) || !bytes.Equal(col, m.rgb[pose]) {
			continue
		}
		at := origin.Add(merchantBox.Inset(1).Min.Mul(scale))
		for y := 0; y < m.art[pose].Bounds().Dy(); y++ {
			for x := 0; x < m.art[pose].Bounds().Dx(); x++ {
				dst.Set(at.X+x, at.Y+y, m.art[pose].At(x, y))
			}
		}
		return true, pose
	}
	return false, -1
}

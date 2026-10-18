package hd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/png"
	"os"
)

var bankBox = image.Rect(56, 23, 152, 87)

// Bank 只辨識正常銀行路徑已觀察的三個完整肖像畫格。
// 姿態由原版畫格選擇，不新增計時器或改動原版動畫。
type Bank struct {
	indices [3][]byte
	rgb     [3][]byte
	art     [3]*image.RGBA
}

func bankCrop(raw []byte, channels int) []byte {
	if len(raw) != nativeWidth*nativeHeight*channels {
		return nil
	}
	out := make([]byte, bankBox.Dx()*bankBox.Dy()*channels)
	for y := bankBox.Min.Y; y < bankBox.Max.Y; y++ {
		copy(out[(y-bankBox.Min.Y)*bankBox.Dx()*channels:], raw[(y*nativeWidth+bankBox.Min.X)*channels:(y*nativeWidth+bankBox.Max.X)*channels])
	}
	return out
}

func NewBank(referenceA, referenceB, referenceC, palette, artworkA, artworkB, artworkC string) (*Bank, error) {
	pal, err := checked(palette, "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c", 768)
	if err != nil {
		return nil, err
	}
	m := &Bank{}
	want := [3]string{"16e3e8fba2ff62b9432929b5b13182f4e47bf240bcdd80bbebe5256c806f3d1e", "a9f97fd254998a19e9394acfcfc5d8b098615d96cbad99c3d69f984975d722f3", "e7412876a41b599f8e260c66b9745255cffcab2b1d08b318429908ab139cd3be"}
	for pose, path := range []string{referenceA, referenceB, referenceC} {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		crop := bankCrop(raw, 1)
		if len(crop) == 0 || fmt.Sprintf("%x", sha256.Sum256(crop)) != want[pose] {
			return nil, fmt.Errorf("銀行原版肖像畫格不符")
		}
		m.indices[pose] = crop
		m.rgb[pose] = make([]byte, len(crop)*3)
		for i, value := range crop {
			copy(m.rgb[pose][i*3:i*3+3], pal[int(value)*3:int(value)*3+3])
		}
	}
	wantArt := [3]string{"99bbac749a11b9ca382087ba2bfd887b9c73ccb5200eddffe23cb6c1a336a9eb", "3046c17cac758ed65b664c1a02af7343cd992f8e0613a931eb6bdc57d017823b", "89ff8e0d13fcc4b7a0cc1a56c8103f89d3cdf588a076ee4ee2c77ece998f8e7c"}
	for pose, path := range []string{artworkA, artworkB, artworkC} {
		data, err := checked(path, wantArt[pose], 16<<20)
		if err != nil {
			return nil, err
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil || cfg.Width != 1536 || cfg.Height != 1024 {
			return nil, fmt.Errorf("銀行 HD 圖面須為本輪1536×1024 PNG")
		}
		art, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		// 直接合成到內框282×186，保留原版黃色外框。
		target := image.NewRGBA(image.Rect(0, 0, bankBox.Inset(1).Dx()*scale, bankBox.Inset(1).Dy()*scale))
		for y := 0; y < target.Bounds().Dy(); y++ {
			for x := 0; x < target.Bounds().Dx(); x++ {
				target.Set(x, y, art.At(x*cfg.Width/target.Bounds().Dx(), y*cfg.Height/target.Bounds().Dy()))
			}
		}
		m.art[pose] = target
	}
	return m, nil
}

func (m *Bank) Draw(dst *image.RGBA, indexed, rgb []byte, origin image.Point) (bool, int) {
	idx, col := bankCrop(indexed, 1), bankCrop(rgb, 3)
	for pose := range m.indices {
		if !bytes.Equal(idx, m.indices[pose]) || !bytes.Equal(col, m.rgb[pose]) {
			continue
		}
		at := origin.Add(bankBox.Inset(1).Min.Mul(scale))
		for y := 0; y < m.art[pose].Bounds().Dy(); y++ {
			for x := 0; x < m.art[pose].Bounds().Dx(); x++ {
				dst.Set(at.X+x, at.Y+y, m.art[pose].At(x, y))
			}
		}
		return true, pose
	}
	return false, -1
}

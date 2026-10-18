package hd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

var kingBox = image.Rect(112, 23, 208, 87)

// King 只接入已觀察的兩個口部圖樣，原版色號決定畫格，未知姿態保留原貌。
type King struct {
	indices [2][]byte
	rgb     [2][]byte
	art     [2]*image.RGBA
}

func kingCrop(raw []byte, channels int) []byte {
	if len(raw) != 320*200*channels {
		return nil
	}
	out := make([]byte, kingBox.Dx()*kingBox.Dy()*channels)
	for y := kingBox.Min.Y; y < kingBox.Max.Y; y++ {
		copy(out[(y-kingBox.Min.Y)*kingBox.Dx()*channels:], raw[(y*320+kingBox.Min.X)*channels:(y*320+kingBox.Max.X)*channels])
	}
	return out
}

func NewKing(referenceA, referenceB, palette, closed, speaking string) (*King, error) {
	pal, err := checked(palette, "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c", 768)
	if err != nil {
		return nil, err
	}
	k := &King{}
	want := []string{"a9dc4a05952f31e0c24f83e05e41478c960446b22c29193404a39e35d5149670", "7bf3b369d882ae533208ad496d71de58051325f37ad1855f924c8a204f235eeb"}
	for i, path := range []string{referenceA, referenceB} {
		raw, e := os.ReadFile(path)
		if e != nil {
			return nil, e
		}
		crop := kingCrop(raw, 1)
		if len(crop) == 0 || fmt.Sprintf("%x", sha256.Sum256(crop)) != want[i] {
			return nil, fmt.Errorf("國王原版姿態不符")
		}
		k.indices[i] = crop
		k.rgb[i] = make([]byte, len(crop)*3)
		for n, v := range crop {
			copy(k.rgb[i][n*3:n*3+3], pal[int(v)*3:int(v)*3+3])
		}
	}
	images := make([]image.Image, 2)
	for i, path := range []string{closed, speaking} {
		hash := []string{"d6be3a9fa2c93da491485e4167bc941f75367b8142ae1b3b6131dfffe6e78a70", "06ca7fbe1dd91ebdf0c602d295019b8d6eb9cead56d3ae91d4b2bc308b52ff00"}[i]
		data, e := checked(path, hash, 16<<20)
		if e != nil {
			return nil, e
		}
		cfg, e := png.DecodeConfig(bytes.NewReader(data))
		if e != nil || cfg.Width != 1536 || cfg.Height != 1024 {
			return nil, fmt.Errorf("國王 HD 圖面須為本輪 1536×1024 PNG")
		}
		images[i], e = png.Decode(bytes.NewReader(data))
		if e != nil {
			return nil, e
		}
	}
	base := image.NewRGBA(image.Rect(0, 0, 288, 192))
	for y := 0; y < 192; y++ {
		for x := 0; x < 288; x++ {
			base.Set(x, y, images[0].At(x*1536/288, y*1024/192))
		}
	}
	open := image.NewRGBA(base.Bounds())
	draw.Draw(open, open.Bounds(), base, image.Point{}, draw.Src)
	patch := image.Rect(104, 86, 184, 125)
	for y := patch.Min.Y; y < patch.Max.Y; y++ {
		for x := patch.Min.X; x < patch.Max.X; x++ {
			// 只混合口部，固定外圍所有像素，避免生成圖細微差異造成整張肖像閃動。
			edge := min(x-patch.Min.X, patch.Max.X-1-x, y-patch.Min.Y, patch.Max.Y-1-y)
			alpha := min(edge, 6)
			a := base.RGBAAt(x, y)
			r, g, b, _ := images[1].At(x*1536/288, y*1024/192).RGBA()
			mix := func(a uint8, b uint32) uint8 { return uint8((int(a)*(6-alpha) + int(b>>8)*alpha) / 6) }
			open.SetRGBA(x, y, color.RGBA{mix(a.R, r), mix(a.G, g), mix(a.B, b), 255})
		}
	}
	k.art[0], k.art[1] = open, base
	return k, nil
}

func (k *King) Draw(dst *image.RGBA, indexed, rgb []byte, origin image.Point) (bool, int) {
	idx, col := kingCrop(indexed, 1), kingCrop(rgb, 3)
	for i := range k.indices {
		if bytes.Equal(idx, k.indices[i]) && bytes.Equal(col, k.rgb[i]) {
			at := origin.Add(image.Pt(kingBox.Min.X*3, kingBox.Min.Y*3))
			// 保留原版肖像的黃色外框，只替換內部圖面。
			draw.Draw(dst, k.art[i].Bounds().Inset(3).Add(at), k.art[i], image.Pt(3, 3), draw.Src)
			return true, i
		}
	}
	return false, -1
}

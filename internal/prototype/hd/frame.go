// Package hd 試作手繪邊框；只讀原版畫格，未辨識畫面保留原貌。
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

const scale, nativeWidth, nativeHeight, sideWidth = 3, 320, 200, 48

type Frame struct {
	indices, rgb []byte
	panels       [2]*image.RGBA
}

func checked(path, expected string, maximum int64) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maximum {
		return nil, fmt.Errorf("HD 試作輸入類型或大小不符")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != expected {
		return nil, fmt.Errorf("HD 試作輸入雜湊不符：%s", path)
	}
	return data, nil
}

// NewFrame 只接受已記錄的城鎮基線與本輪手繪 atlas，不是正式資產格式。
func NewFrame(reference, palette, artwork string) (*Frame, error) {
	idx, err := checked(reference, "864fb0539ba53ac15592344d5c63723e7d8c46754d5ed7cb9570f21bf3bdd57d", 64000)
	if err != nil {
		return nil, err
	}
	pal, err := checked(palette, "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c", 768)
	if err != nil {
		return nil, err
	}
	art, err := checked(artwork, "bb9cc2e211b1d2223e3baf0e185bab25bc484bb26477078e9d769e67d5f9582d", 16<<20)
	if err != nil {
		return nil, err
	}
	config, err := png.DecodeConfig(bytes.NewReader(art))
	if err != nil || config.Width < 288 || config.Height < 600 || config.Width > 4096 || config.Height > 4096 {
		return nil, fmt.Errorf("HD 試作 PNG 尺寸不符")
	}
	img, err := png.Decode(bytes.NewReader(art))
	if err != nil {
		return nil, err
	}
	f := &Frame{indices: idx, rgb: make([]byte, 320*200*3)}
	for i, value := range idx {
		copy(f.rgb[i*3:i*3+3], pal[int(value)*3:int(value)*3+3])
	}
	for side := range f.panels {
		bounds := image.Rect(side*config.Width/2, 0, (side+1)*config.Width/2, config.Height)
		ink := image.Rectangle{}
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				_, _, _, alpha := img.At(x, y).RGBA()
				if alpha != 0 {
					ink = ink.Union(image.Rect(x, y, x+1, y+1))
				}
			}
		}
		if ink.Empty() {
			return nil, fmt.Errorf("HD 試作缺少一側邊框")
		}
		panel := image.NewRGBA(image.Rect(0, 0, sideWidth*scale, nativeHeight*scale))
		draw.Draw(panel, panel.Bounds(), &image.Uniform{color.RGBA{15, 22, 19, 255}}, image.Point{}, draw.Src)
		// 直接在 144×600 目標畫布合成，不先降成原版 48×200。
		foreground := image.NewRGBA(panel.Bounds())
		for y := 0; y < panel.Bounds().Dy(); y++ {
			for x := 0; x < panel.Bounds().Dx(); x++ {
				foreground.Set(x, y, img.At(ink.Min.X+x*ink.Dx()/panel.Bounds().Dx(), ink.Min.Y+y*ink.Dy()/panel.Bounds().Dy()))
			}
		}
		draw.Draw(panel, panel.Bounds(), foreground, image.Point{}, draw.Over)
		f.panels[side] = panel
	}
	return f, nil
}

// Match 同時核對左右兩側全部色號及 RGB。色盤淡入、換場或任一側失配時整組退回。
func (f *Frame) Match(indexed, rgb []byte) bool {
	if len(indexed) != 320*200 || len(rgb) != 320*200*3 {
		return false
	}
	for y := 0; y < nativeHeight; y++ {
		for _, start := range []int{0, nativeWidth - sideWidth} {
			lo, hi := y*nativeWidth+start, y*nativeWidth+start+sideWidth
			if !bytes.Equal(indexed[lo:hi], f.indices[lo:hi]) || !bytes.Equal(rgb[lo*3:hi*3], f.rgb[lo*3:hi*3]) {
				return false
			}
		}
	}
	return true
}

func (f *Frame) Draw(dst *image.RGBA, indexed, rgb []byte, origin image.Point) bool {
	if !f.Match(indexed, rgb) {
		return false
	}
	for side, panel := range f.panels {
		position := origin.Add(image.Pt(side*(nativeWidth-sideWidth)*scale, 0))
		draw.Draw(dst, panel.Bounds().Add(position), panel, image.Point{}, draw.Src)
	}
	return true
}

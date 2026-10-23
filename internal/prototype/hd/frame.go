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
	indices, rgb, palette []byte
	panels                [2]*image.RGBA
}

// frameEyes 是開局逐格揭開時兩側雕像眼睛的暗色畫格（色號）。原版畫完柱子後才把這30點換成亮色；
// 實測於冷新局48.31M指令，規格067。
var frameEyes = map[int]byte{
	47*320 + 9: 24, 47*320 + 309: 29, 47*320 + 310: 24, 48*320 + 9: 43, 48*320 + 11: 24,
	48*320 + 309: 29, 48*320 + 310: 24, 48*320 + 311: 24, 49*320 + 8: 29, 49*320 + 9: 29,
	49*320 + 10: 24, 49*320 + 11: 27, 49*320 + 307: 29, 49*320 + 308: 29, 49*320 + 309: 24,
	49*320 + 310: 29, 49*320 + 311: 27, 50*320 + 8: 29, 50*320 + 9: 24, 50*320 + 10: 27,
	50*320 + 11: 27, 50*320 + 307: 24, 50*320 + 308: 24, 50*320 + 309: 27, 50*320 + 310: 27,
	50*320 + 311: 27, 51*320 + 8: 24, 51*320 + 9: 24, 51*320 + 310: 27, 51*320 + 311: 27,
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
	f := &Frame{indices: idx, rgb: make([]byte, 320*200*3), palette: pal}
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

// revealing 回報原版是否正在逐格畫出同一組邊框：兩側每一點不是尚未畫的色號0，就是參考色號與RGB，
// 或眼睛的暗色畫格；且至少已畫一點。任何其他色號或色盤不同都不算。規格067。
func (f *Frame) revealing(indexed, rgb []byte) bool {
	if len(indexed) != 320*200 || len(rgb) != 320*200*3 {
		return false
	}
	drawn := false
	for y := 0; y < nativeHeight; y++ {
		for _, start := range []int{0, nativeWidth - sideWidth} {
			for i := y*nativeWidth + start; i < y*nativeWidth+start+sideWidth; i++ {
				v := indexed[i]
				switch {
				case v == 0:
					continue
				case v == f.indices[i]:
					if !bytes.Equal(rgb[i*3:i*3+3], f.rgb[i*3:i*3+3]) {
						return false
					}
				case frameEyes[i] == v:
					if !bytes.Equal(rgb[i*3:i*3+3], f.palette[int(v)*3:int(v)*3+3]) {
						return false
					}
				default:
					return false
				}
				drawn = true
			}
		}
	}
	return drawn
}

// Draw 完整吻合時畫兩側 HD 邊框；原版逐格揭開期間只畫已有像素的8×8格，尚未畫到的格維持原版的黑色，
// 不先露出原版圖格。規格067。
func (f *Frame) Draw(dst *image.RGBA, indexed, rgb []byte, origin image.Point) bool {
	if f.Match(indexed, rgb) {
		for side, panel := range f.panels {
			position := origin.Add(image.Pt(side*(nativeWidth-sideWidth)*scale, 0))
			draw.Draw(dst, panel.Bounds().Add(position), panel, image.Point{}, draw.Src)
		}
		return true
	}
	if !f.revealing(indexed, rgb) {
		return false
	}
	for side, panel := range f.panels {
		start := side * (nativeWidth - sideWidth)
		for cy := 0; cy < nativeHeight; cy += 8 {
			for cx := 0; cx < sideWidth; cx += 8 {
				if !cellInked(indexed, start+cx, cy) {
					continue
				}
				cell := image.Rect(cx*scale, cy*scale, (cx+8)*scale, (cy+8)*scale)
				draw.Draw(dst, cell.Add(origin).Add(image.Pt(start*scale, 0)), panel, cell.Min, draw.Src)
			}
		}
	}
	return true
}

func cellInked(indexed []byte, x, y int) bool {
	for dy := 0; dy < 8; dy++ {
		for dx := 0; dx < 8; dx++ {
			if indexed[(y+dy)*nativeWidth+x+dx] != 0 {
				return true
			}
		}
	}
	return false
}

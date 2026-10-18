package hd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
)

var itemShopBox = image.Rect(56, 23, 152, 87)
var itemShopCharacter = image.Rect(51, 21, 141, 186)
var itemShopSteam = image.Rect(141, 0, 282, 141)

// ItemShop 只識別十個正常路徑抽樣中的四組完整畫格。
// 同一底圖加上原版差異區域，避免動畫切換時整間店跳動。
type ItemShop struct {
	indices [4][]byte
	rgb     [4][]byte
	art     [4]*image.RGBA
}

func itemShopCrop(raw []byte, channels int) []byte {
	if len(raw) != nativeWidth*nativeHeight*channels {
		return nil
	}
	out := make([]byte, itemShopBox.Dx()*itemShopBox.Dy()*channels)
	for y := itemShopBox.Min.Y; y < itemShopBox.Max.Y; y++ {
		copy(out[(y-itemShopBox.Min.Y)*itemShopBox.Dx()*channels:], raw[(y*nativeWidth+itemShopBox.Min.X)*channels:(y*nativeWidth+itemShopBox.Max.X)*channels])
	}
	return out
}

// artworks 順序為基準、蒸氣B、抬手、蒸氣C。來源與提示詞見 CONTEXT 的 item-hd-v1 入口。
func NewItemShop(references [4]string, palette string, artworks [4]string) (*ItemShop, error) {
	pal, err := checked(palette, "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c", 768)
	if err != nil {
		return nil, err
	}
	m := &ItemShop{}
	want := [4]string{"c7454368eb514d561aecebf2f6643c13d3388b5562c04929cf926ad80dd3664a", "4573f001d045608b6002c169000280b9abda73d8ac02602da73418cf44bc12c8", "92b18227daf0ff19b924e7cd4dda23a8f4c61ad169c337ad31fcf12b3ab0c87d", "41443a23827af172d794ec5f9cef0deae93121ff746216fd5f8aedfb75862a7e"}
	for pose, path := range references {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		crop := itemShopCrop(raw, 1)
		if len(crop) == 0 || fmt.Sprintf("%x", sha256.Sum256(crop)) != want[pose] {
			return nil, fmt.Errorf("道具店原版畫格不符")
		}
		m.indices[pose] = crop
		m.rgb[pose] = make([]byte, len(crop)*3)
		for i, value := range crop {
			copy(m.rgb[pose][i*3:i*3+3], pal[int(value)*3:int(value)*3+3])
		}
	}
	wantArt := [4]string{"8f62986f755db681a9b1e036a3349102aa8d35952f32805249ada836e774b3fa", "23c65e3d0acee070176bdb70ed906f83c0f523663ee90b9c9658907d7e95a3ea", "b3b4df05a4fe694440f137710d696f9b1c8fff537e5a4f6140ab94211a0aff8d", "b9481571a67effb66bd7d9404ed06bf419ddd307625cba7d2e7b23263809471d"}
	var variants [4]*image.RGBA
	for i, path := range artworks {
		data, err := checked(path, wantArt[i], 16<<20)
		if err != nil {
			return nil, err
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil || cfg.Width != 1536 || cfg.Height != 1024 {
			return nil, fmt.Errorf("道具店 HD 須為已核對的1536×1024 PNG")
		}
		art, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		target := image.NewRGBA(image.Rect(0, 0, itemShopBox.Inset(1).Dx()*scale, itemShopBox.Inset(1).Dy()*scale))
		for y := 0; y < target.Bounds().Dy(); y++ {
			for x := 0; x < target.Bounds().Dx(); x++ {
				target.Set(x, y, art.At(x*cfg.Width/target.Bounds().Dx(), y*cfg.Height/target.Bounds().Dy()))
			}
		}
		variants[i] = target
	}
	for pose := range m.art {
		target := image.NewRGBA(variants[0].Bounds())
		draw.Draw(target, target.Bounds(), variants[0], image.Point{}, draw.Src)
		if pose >= 2 {
			draw.Draw(target, itemShopCharacter, variants[2], itemShopCharacter.Min, draw.Src)
		}
		if pose == 1 || pose == 3 {
			variant := 1
			if pose == 3 {
				variant = 3
			}
			draw.Draw(target, itemShopSteam, variants[variant], itemShopSteam.Min, draw.Src)
		}
		m.art[pose] = target
	}
	return m, nil
}

func (m *ItemShop) Draw(dst *image.RGBA, indexed, rgb []byte, origin image.Point) (bool, int) {
	idx, col := itemShopCrop(indexed, 1), itemShopCrop(rgb, 3)
	for pose := range m.indices {
		if !bytes.Equal(idx, m.indices[pose]) || !bytes.Equal(col, m.rgb[pose]) {
			continue
		}
		at := origin.Add(itemShopBox.Inset(1).Min.Mul(scale))
		draw.Draw(dst, m.art[pose].Bounds().Add(at), m.art[pose], image.Point{}, draw.Src)
		return true, pose
	}
	return false, -1
}

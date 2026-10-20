// Muralla地面覆繪的來源、取樣及退回契約見docs/spec/009-town-ground-hd.md。
package hd

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
)

var townGroundBox = image.Rect(48, 142, 272, 158)

type TownGround struct {
	scene     *TownScenery
	reference []byte
	art       *image.RGBA
}

func NewTownGround(scene *TownScenery, reference, artwork string) (*TownGround, error) {
	if scene == nil {
		return nil, fmt.Errorf("地面HD需要已驗城鎮前景來源")
	}
	raw, err := checked(reference, "dd777f4940d2ffe32d5fae6a8274c0f3a0878fb84e3f7fd3f425a4af92d6d224", 64000)
	if err != nil || len(raw) != 64000 {
		return nil, fmt.Errorf("地面原版參考不符：%v", err)
	}
	data, err := checked(artwork, "ea2b08001d76d6ed739e48a6ec6695db2bda05a4c843c9e573977b8ffcc45074", 16<<20)
	if err != nil {
		return nil, err
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width != 2172 || cfg.Height != 724 {
		return nil, fmt.Errorf("地面圖面尺寸不符")
	}
	art, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	s := &TownGround{scene: scene, art: image.NewRGBA(image.Rect(0, 0, 336, 48))}
	for y := 142; y < 158; y++ {
		s.reference = append(s.reference, raw[y*320+48:y*320+160]...)
	}
	for y := 0; y < 48; y++ {
		for x := 0; x < 336; x++ {
			s.art.Set(x, y, art.At(x*2172/336, 273+y*194/48))
		}
	}
	return s, nil
}

// Draw先驗完整原版前景、地面色號及RGB；任何失配都不寫畫布。
func (s *TownGround) Draw(dst *image.RGBA, mode uint8, p, r, e, raw, rgb, palette []byte, origin image.Point) bool {
	box := image.Rect(144, 426, 816, 474).Add(origin)
	if s == nil || !box.In(dst.Bounds()) || !s.scene.frameSource(townFrameInput{mode: mode, primary: p, resources: r, extra: e, palette: palette}) {
		return false
	}
	_, err := s.scene.decode(mode, p, r, e, raw, rgb)
	if err != nil {
		return false
	}
	phase, ok := s.phases(raw, rgb)
	if !ok {
		return false
	}
	for y := 0; y < 48; y++ {

		for x := 0; x < 672; x++ {
			u := (x + phase[y/24]*3) % 336
			if u < 0 {
				u += 336
			}
			c := s.art.RGBAAt(u, y)
			px, py := box.Min.X+x, box.Min.Y+y
			old := dst.RGBAAt(px, py)
			inv := uint32(255 - c.A)
			dst.SetRGBA(px, py, color.RGBA{uint8(uint32(c.R) + uint32(old.R)*inv/255), uint8(uint32(c.G) + uint32(old.G)*inv/255), uint8(uint32(c.B) + uint32(old.B)*inv/255), 255})
		}
	}
	return true
}

// phases只接受完整圖樣與RGB的唯一循環位移，契約見009回城勘誤。
func (s *TownGround) phases(raw, rgb []byte) ([2]int, bool) {
	var phases [2]int
	if len(raw) != 64000 || len(rgb) != 192000 || len(s.reference) != 1792 {
		return phases, false
	}
	for band := 0; band < 2; band++ {
		found := -1
		for shift := 0; shift < 112; shift++ {
			matched := true
			for y := band * 8; y < band*8+8 && matched; y++ {
				for x := 0; x < 224; x++ {
					value := s.reference[y*112+(x+shift)%112]
					i := (142+y)*320 + 48 + x
					if raw[i] != value || !bytes.Equal(rgb[i*3:i*3+3], s.scene.palette[int(value)*3:int(value)*3+3]) {
						matched = false
						break
					}
				}
			}
			if matched {
				if found >= 0 {
					return phases, false
				}
				found = shift
			}
		}
		if found < 0 {
			return phases, false
		}
		phases[band] = found
	}
	return phases, true
}

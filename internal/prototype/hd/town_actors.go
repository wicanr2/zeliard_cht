// 居民與主角共同合成，保留原版繪製順序；來源與邊界見規格035。
package hd

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"path/filepath"
)

type townNPCAsset struct {
	group, phase       int
	file, hash, source string
	crop, target       image.Rectangle
}

// TownActors 只持有唯讀圖面，不持有執行器、輸入或存檔。
type TownActors struct {
	player *TownPlayer
	art    [5][8]*image.RGBA
	source [5][8]string
	state  map[string]any
}

func NewTownActors(player *TownPlayer, directory string) (*TownActors, error) {
	if player == nil || player.scene == nil {
		return nil, fmt.Errorf("居民HD需要已載入的城鎮主角及前景")
	}
	s := &TownActors{player: player}
	cache := map[string]image.Image{}
	for _, asset := range townNPCAssets {
		if asset.group < 0 || asset.group >= 5 || asset.phase < 0 || asset.phase >= 8 ||
			s.art[asset.group][asset.phase] != nil || asset.target.Empty() || !asset.target.In(image.Rect(0, 0, 16, 24)) {
			return nil, fmt.Errorf("居民資產表不完整或重複")
		}
		im := cache[asset.file]
		if im == nil {
			data, err := checked(filepath.Join(directory, asset.file), asset.hash, 16<<20)
			if err != nil {
				return nil, err
			}
			cfg, err := png.DecodeConfig(bytes.NewReader(data))
			if err != nil || cfg.Width != 1448 || cfg.Height != 1086 {
				return nil, fmt.Errorf("居民圖集尺寸不符")
			}
			im, err = png.Decode(bytes.NewReader(data))
			if err != nil {
				return nil, err
			}
			cache[asset.file] = im
		}
		if asset.crop.Empty() || !asset.crop.In(im.Bounds()) {
			return nil, fmt.Errorf("居民圖集裁切越界")
		}
		cropped := im.(interface {
			SubImage(image.Rectangle) image.Image
		}).SubImage(asset.crop)
		sprite := image.NewRGBA(image.Rect(0, 0, 48, 72))
		w, h := asset.target.Dx()*3, asset.target.Dy()*3
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				sprite.Set(asset.target.Min.X*3+x, asset.target.Min.Y*3+y, caveArea(cropped, x, y, w, h))
			}
		}
		s.art[asset.group][asset.phase] = sprite
		s.source[asset.group][asset.phase] = asset.source
	}
	for group := range s.art {
		for phase := range s.art[group] {
			if s.art[group][phase] == nil || len(s.source[group][phase]) != 64 {
				return nil, fmt.Errorf("居民姿態缺漏")
			}
		}
	}
	return s, nil
}

func (s *TownActors) State() map[string]any {
	if s == nil || s.state == nil {
		return nil
	}
	state := make(map[string]any, len(s.state))
	for k, v := range s.state {
		if keys, ok := v.([]string); ok {
			v = append([]string(nil), keys...)
		}
		state[k] = v
	}
	return state
}

func (s *TownActors) Clear() {
	if s != nil {
		s.state = map[string]any{"loaded": true, "matched": false}
	}
}

// townNPCRaster沿007的圖格與遮罩解出原版RGBA，不以顯示顏色猜人物。
func townNPCRaster(r, e, palette []byte, group, phase int) ([]byte, error) {
	if len(r) != 65536 || len(e) != 65536 || len(palette) != 768 || group < 0 || group >= 5 || phase < 0 || phase >= 8 {
		return nil, fmt.Errorf("居民來源尺寸或姿態不符")
	}
	result := make([]byte, 16*24*4)
	for q, one := range r[0x4000+group*48+phase*6 : 0x4000+group*48+phase*6+6] {
		id := int(one) - 1
		if id < 0 || id >= 152 {
			return nil, fmt.Errorf("居民圖格越界")
		}
		pixels := townSceneTile(r, 0x4100+id*48)
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				keep := e[0x7000+id*8+y]&(128>>x) != 0
				c := int(pixels[y*8+x])
				if keep {
					if c != 0 {
						return nil, fmt.Errorf("未驗居民OR混合")
					}
					continue
				}
				i := ((q%3*8+y)*16 + q/3*8 + x) * 4
				copy(result[i:i+3], palette[c*3:c*3+3])
				result[i+3] = 255
			}
		}
	}
	return result, nil
}

func (s *TownActors) Draw(dst *image.RGBA, mode uint8, p, r, e, raw, rgb, palette []byte, origin image.Point) bool {
	if s == nil {
		return false
	}
	s.Clear()
	if s.player == nil {
		return false
	}
	v, err := s.player.prepare(mode, p, r, e, raw, rgb, palette)
	row := image.Rect(144, 354, 816, 426).Add(origin)
	if err != nil || dst == nil || !row.In(dst.Bounds()) || s.player.art[v.bit*5+v.phase] == nil {
		return false
	}
	actors, err := v.f.scene.ownSources(p, r, e)
	if err != nil {
		return false
	}
	actors, err = townScenePaintActors(p, actors)
	if err != nil {
		return false
	}
	type placed struct {
		x   int
		art *image.RGBA
	}
	var sprites []placed
	var keys []string
	covered := make([]bool, 224)
	cover := func(x int) {
		for xx := max(0, x); xx < min(224, x+16); xx++ {
			covered[xx] = true
		}
	}
	for _, actor := range actors {
		x := (actor.x - v.f.camera - 4) * 8
		if x <= -16 || x >= 224 {
			continue
		}
		phase := actor.phase
		if !actor.direction {
			phase += 4
		}
		art := s.art[actor.group][phase]
		source, err := townNPCRaster(r, e, palette, actor.group, phase)
		if err != nil || art == nil || art.Bounds() != image.Rect(0, 0, 48, 72) || townBackgroundHash(source) != s.source[actor.group][phase] {
			return false
		}
		sprites = append(sprites, placed{x, art})
		keys = append(keys, fmt.Sprintf("g%d-p%d", actor.group, phase))
		cover(x)
	}
	if v.visible {
		cover(v.x)
	}
	// 所有守衛完成才建立補丁。補丁內先還原靜態層，再按原版順序畫人物。
	patch := image.NewRGBA(row)
	draw.Draw(patch, row, dst, row.Min, draw.Src)
	for y := 0; y < 72; y++ {
		for x := 0; x < 672; x++ {
			if !covered[x/3] {
				continue
			}
			i := (40+y/3)*224 + x/3
			if !v.f.foreground[i] {
				return false
			}
			at := int(v.f.base[i]) * 3
			base := color.RGBA{palette[at], palette[at+1], palette[at+2], 255}
			worldX := (v.f.camera+4)*24 + x
			c := townPlayerOver(v.f.scene.art.RGBAAt(worldX, 120+y), base)
			patch.SetRGBA(row.Min.X+x, row.Min.Y+y, c)
		}
	}
	for _, sprite := range sprites {
		box := image.Rect(row.Min.X+sprite.x*3, row.Min.Y, row.Min.X+sprite.x*3+48, row.Max.Y)
		draw.Draw(patch, box, sprite.art, image.Point{}, draw.Over)
	}
	if v.visible {
		box := image.Rect(row.Min.X+v.x*3, row.Min.Y, row.Min.X+v.x*3+48, row.Max.Y)
		draw.Draw(patch, box, s.player.art[v.bit*5+v.phase], image.Point{}, draw.Over)
	}
	draw.Draw(dst, row, patch, row.Min, draw.Src)
	s.state = map[string]any{"loaded": true, "matched": true, "residents": len(sprites), "keys": keys, "player_visible": v.visible, "player_phase": v.phase, "player_direction_bit": v.bit}
	return true
}

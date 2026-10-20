// Muralla主角八個行走姿態覆繪契約見docs/spec/010-town-player-hd.md。
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

type TownPlayer struct {
	scene *TownScenery
	art   [10]*image.RGBA
}

func NewTownPlayer(scene *TownScenery, artwork string) (*TownPlayer, error) {
	if scene == nil {
		return nil, fmt.Errorf("主角HD需要已驗前景")
	}

	s := &TownPlayer{scene: scene}
	assets := []struct {
		key        int
		file, hash string
		crop       image.Rectangle
	}{
		{0, "player-c2bit0-phase0-handpaint-v2.png", "974da640846589a4e8a4fd9aac48f93b14bf3978a454d8e0dffbd0a9485adb1d", image.Rect(12, 2, 1007, 1495)},
		{1, "player-c2bit0-phase1-handpaint-v2.png", "ea49f2489f990bd774fe98f1430fc2786b28b2bc22e7e2f2142b0b4ca8303ef0", image.Rect(12, -12, 1007, 1481)},
		{2, "player-c2bit0-phase2-handpaint-v2.png", "9ba3a40880add6c314fb74b269269a3d103f24a8eb00d3f4bcf83866122120dc", image.Rect(12, -21, 1007, 1472)},
		{3, "player-c2bit0-phase3-handpaint-v1.png", "a825ba49535cc182602e9d257d0576a9f8501354be5e74e83e60ef8f17fdec02", image.Rect(12, -12, 1007, 1481)},
		{5, "player-c2bit1-phase0-handpaint-v2.png", "14f2687dfea4b978b80567eecf3020a4c243408cb783ce27de1306b654f7fd5d", image.Rect(12, -20, 1007, 1473)},
		{6, "player-c2bit1-phase1-handpaint-v2.png", "0a4b06a1eccf89eba0d6a3524a8f327544f999e0f03e14ca88bf0c4d9855f175", image.Rect(12, -18, 1007, 1475)},
		{7, "player-c2bit1-phase2-handpaint-v2.png", "3e45e5754ec131dd71627b4d2497ac73630daf6bab91ed3c974a3b9ba9cb800a", image.Rect(12, -30, 1007, 1463)},
		{8, "player-c2bit1-phase3-handpaint-v1.png", "d13a38ab3f7e326a57bad421e30dbb03cc21092bcf3f48d15e7168898eca64c0", image.Rect(12, -44, 1007, 1449)},
	}
	for _, a := range assets {
		b, err := checked(filepath.Join(artwork, a.file), a.hash, 16<<20)
		if err != nil {
			return nil, err
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(b))
		if err != nil || cfg.Width != 1024 || cfg.Height != 1536 {
			return nil, fmt.Errorf("主角HD圖尺寸不符")
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		sprite := image.NewRGBA(image.Rect(0, 0, 48, 72))
		for y := 0; y < 72; y++ {
			for x := 0; x < 48; x++ {
				x0, x1 := a.crop.Min.X+x*a.crop.Dx()/48, a.crop.Min.X+(x+1)*a.crop.Dx()/48
				y0, y1 := a.crop.Min.Y+y*a.crop.Dy()/72, a.crop.Min.Y+(y+1)*a.crop.Dy()/72
				var sum [4]uint64
				for sy := y0; sy < y1; sy++ {
					for sx := x0; sx < x1; sx++ {
						r, g, b, a := im.At(sx, sy).RGBA()
						sum[0] += uint64(r)
						sum[1] += uint64(g)
						sum[2] += uint64(b)
						sum[3] += uint64(a)
					}
				}
				n := uint64((x1 - x0) * (y1 - y0))
				sprite.SetRGBA(x, y, color.RGBA{uint8((sum[0] / n) >> 8), uint8((sum[1] / n) >> 8), uint8((sum[2] / n) >> 8), uint8((sum[3] / n) >> 8)})
			}
		}
		s.art[a.key] = sprite
	}

	return s, nil
}

type townPlayerFrame struct {
	f             *townSceneFrame
	underlay      []byte
	npc, player   []byte
	x, bit, phase int
	visible       bool
}

func (s *TownPlayer) prepare(mode uint8, p, r, e, raw, rgb, palette []byte) (*townPlayerFrame, error) {
	if s == nil || len(p) != 65536 || p[0xe7] > 3 || !s.scene.frameSource(townFrameInput{mode: mode, primary: p, resources: r, extra: e, palette: palette}) {
		return nil, fmt.Errorf("主角來源拒絕")
	}
	f, err := s.scene.decode(mode, p, r, e, raw, rgb)
	if err != nil {
		return nil, err
	}
	actors, err := f.scene.ownSources(p, r, e)
	if err != nil {
		return nil, err
	}
	v := &townPlayerFrame{f: f, underlay: append([]byte(nil), f.base...), npc: make([]byte, 14336), player: make([]byte, 14336), x: int(p[0x83]) * 8, bit: int(p[0xc2] & 1), phase: int(p[0xe7])}
	v.visible = p[0x83] < 27
	paint := func(x int, ids []byte, atlas, masks int, mask []byte, apply bool) error {
		for col := 0; col < 2; col++ {
			for row := 0; row < 3; row++ {
				id := int(ids[col*3+row])
				pixels := townSceneTile(r, atlas+id*48)
				for dy := 0; dy < 8; dy++ {
					for dx := 0; dx < 8; dx++ {
						px, py := x+col*8+dx, 40+row*8+dy
						if px < 0 || px >= 224 {
							continue
						}
						i := py*224 + px
						keep := e[masks+id*8+dy]&(0x80>>dx) != 0
						c := pixels[dy*8+dx]
						if keep && c != 0 {
							return fmt.Errorf("未驗OR混合")
						}
						if !keep {
							mask[i] = 1
							if apply {
								v.underlay[i] = c
							}
						}
					}
				}
			}
		}
		return nil
	}
	for _, a := range actors {
		x := (a.x - f.camera - 4) * 8
		if x <= -16 || x >= 224 {
			continue
		}
		phase := a.phase
		if !a.direction {
			phase += 4
		}
		off := 0x4000 + a.group*48 + phase*6
		ids := append([]byte(nil), r[off:off+6]...)
		for i := range ids {
			ids[i]--
		}
		if err := paint(x, ids, 0x4100, 0x7000, v.npc, true); err != nil {
			return nil, err
		}
	}
	table := 0x6a59
	if v.bit != 0 {
		table = 0x6a3b
	}
	table += v.phase * 6
	if v.visible {
		if err := paint(v.x, p[table:table+6], 0x6000, 0x8000, v.player, false); err != nil {
			return nil, err
		}
	}
	for i, opaque := range v.player {
		if opaque != 0 && !f.foreground[i] {
			return nil, fmt.Errorf("主角下方非已驗前景")
		}
	}
	return v, nil
}

func townPlayerOver(c, old color.RGBA) color.RGBA {
	inv := uint32(255 - c.A)
	return color.RGBA{uint8(uint32(c.R) + uint32(old.R)*inv/255), uint8(uint32(c.G) + uint32(old.G)*inv/255), uint8(uint32(c.B) + uint32(old.B)*inv/255), 255}
}

func (s *TownPlayer) Draw(dst *image.RGBA, mode uint8, p, r, e, raw, rgb, palette []byte, origin image.Point) bool {
	v, err := s.prepare(mode, p, r, e, raw, rgb, palette)
	if err != nil || !v.visible || s.art[v.bit*5+v.phase] == nil {
		return false
	}
	return s.drawPrepared(dst, v, origin)
}

func (s *TownPlayer) drawPrepared(dst *image.RGBA, v *townPlayerFrame, origin image.Point) bool {
	game := image.Rect(144, 354, 816, 426).Add(origin)
	box := image.Rect((48+v.x)*3, 354, (48+v.x)*3+48, 426).Add(origin)
	safe := box.Intersect(game)
	if dst == nil || safe.Empty() || !game.In(dst.Bounds()) {
		return false
	}
	patch := image.NewRGBA(safe)
	draw.Draw(patch, safe, dst, safe.Min, draw.Src)
	for y := 0; y < 24; y++ {
		for x := 0; x < 16; x++ {
			px := v.x + x
			if px < 0 || px >= 224 {
				continue
			}
			i := (40+y)*224 + px
			if v.player[i] == 0 {
				continue
			}
			index := int(v.underlay[i]) * 3
			base := color.RGBA{s.scene.palette[index], s.scene.palette[index+1], s.scene.palette[index+2], 255}
			for dy := 0; dy < 3; dy++ {
				for dx := 0; dx < 3; dx++ {
					c := base
					if v.npc[i] == 0 {
						worldX := (v.f.camera+4)*24 + px*3 + dx
						c = townPlayerOver(v.f.scene.art.RGBAAt(worldX, (40+y)*3+dy), base)
					}
					patch.SetRGBA(box.Min.X+x*3+dx, box.Min.Y+y*3+dy, c)
				}
			}
		}
	}
	for y := safe.Min.Y; y < safe.Max.Y; y++ {
		for x := safe.Min.X; x < safe.Max.X; x++ {
			c := s.art[v.bit*5+v.phase].RGBAAt(x-box.Min.X, y-box.Min.Y)
			patch.SetRGBA(x, y, townPlayerOver(c, patch.RGBAAt(x, y)))
		}
	}
	draw.Draw(dst, safe, patch, safe.Min, draw.Src)
	return true
}

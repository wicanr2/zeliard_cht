// 規格058：完整首領構圖優先於逐元件覆繪；只讀原版來源。
package hd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
)

type coherentBoss struct {
	Key, File      string
	Hash           string `json:"sha256"`
	Width, Height  int
	Raster, Colors []byte
	art            image.Image
	anchor         []byte
	ax, ay         int
}
type coherentBosses struct {
	entries []*coherentBoss
	matched []string
}

func loadCoherentBosses(dir string) (*coherentBosses, error) {
	data, e := os.ReadFile(filepath.Join(dir, "coherent-bosses.json"))
	if os.IsNotExist(e) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	var m struct {
		Schema, Scale int
		Figures       []*coherentBoss
	}
	if e = json.Unmarshal(data, &m); e != nil {
		return nil, e
	}
	if m.Schema != 1 || m.Scale != 3 || len(m.Figures) == 0 || len(m.Figures) > 32 {
		return nil, fmt.Errorf("完整首領清單不符")
	}
	s := &coherentBosses{}
	for _, r := range m.Figures {
		if r.Width < 8 || r.Width > 224 || r.Height < 1 || r.Height > 144 || len(r.Raster) != r.Width*r.Height || len(r.Colors) != 768 || filepath.Base(r.File) != r.File {
			return nil, fmt.Errorf("完整首領來源不符")
		}
		r.art, e = checkedTerrainImage(filepath.Join(dir, r.File), r.Hash, r.Width*3, r.Height*3)
		if e != nil {
			return nil, e
		}
		best := -1
		for y := 0; y < r.Height; y++ {
			for x := 0; x <= r.Width-8; x++ {
				a := r.Raster[y*r.Width+x : y*r.Width+x+8]
				var seen [256]bool
				score := 0
				for _, v := range a {
					if v != 0 {
						score++
						if !seen[v] {
							score += 8
							seen[v] = true
						}
					}
				}
				if score > best {
					best = score
					r.ax = x
					r.ay = y
					r.anchor = append([]byte(nil), a...)
				}
			}
		}
		if best <= 0 {
			return nil, fmt.Errorf("完整首領沒有可見來源")
		}
		s.entries = append(s.entries, r)
	}
	return s, nil
}

// DrawCoherent以同一份已選出的呈現來源匹配，不能混用當次CPU與上一畫格。
func (s *ManualArtwork) DrawCoherent(dst *image.RGBA, raw, rgb []byte, origin image.Point) bool {
	if s == nil || s.coherent == nil || dst == nil || len(raw) != 64000 || len(rgb) != 192000 {
		return false
	}
	c := s.coherent
	c.matched = nil
	for _, r := range c.entries {
		for y := 14; y < 158; y++ {
			row := raw[y*320+48 : y*320+272]
			start := 0
			for start <= len(row)-8 {
				at := bytes.Index(row[start:], r.anchor)
				if at < 0 {
					break
				}
				at += start
				start = at + 1
				left, top := 48+at-r.ax, y-r.ay
				box := image.Rect(left, top, left+r.Width, top+r.Height)
				if !box.In(cavePresentationBox) {
					continue
				}
				match := true
				for sy := 0; sy < r.Height && match; sy++ {
					for sx := 0; sx < r.Width; sx++ {
						v := r.Raster[sy*r.Width+sx]
						if v == 0 {
							continue
						}
						i := (top+sy)*320 + left + sx
						if raw[i] != v || !bytes.Equal(rgb[i*3:i*3+3], r.Colors[int(v)*3:int(v)*3+3]) {
							match = false
							break
						}
					}
				}
				if !match {
					continue
				}
				target := image.Rect(left*3, top*3, (left+r.Width)*3, (top+r.Height)*3).Add(origin)
				if !target.In(dst.Bounds()) {
					continue
				}
				// 原生透明幾何限定覆繪範圍；其他角色與背景不被整個矩形蓋住。
				for sy := 0; sy < r.Height; sy++ {
					for sx := 0; sx < r.Width; sx++ {
						v := r.Raster[sy*r.Width+sx]
						if v == 0 {
							continue
						}
						for dy := 0; dy < 3; dy++ {
							for dx := 0; dx < 3; dx++ {
								art := color.NRGBAModel.Convert(r.art.At(sx*3+dx, sy*3+dy)).(color.NRGBA)
								a, inv := uint32(art.A), uint32(255-art.A)
								i := int(v) * 3
								px, py := (left+sx)*3+dx+origin.X, (top+sy)*3+dy+origin.Y
								dst.SetRGBA(px, py, color.RGBA{uint8((uint32(art.R)*a + uint32(r.Colors[i])*inv + 127) / 255), uint8((uint32(art.G)*a + uint32(r.Colors[i+1])*inv + 127) / 255), uint8((uint32(art.B)*a + uint32(r.Colors[i+2])*inv + 127) / 255), 255})
							}
						}
					}
				}
				c.matched = append(c.matched, r.Key)
			}
		}
	}
	return len(c.matched) > 0
}

func (s *ManualArtwork) CoherentState() map[string]any {
	if s == nil || s.coherent == nil {
		return nil
	}
	return map[string]any{"loaded": len(s.coherent.entries), "matched": append([]string{}, s.coherent.matched...)}
}

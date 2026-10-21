// 050人工驗證候選：以完整原生8×8圖樣及色彩匹配補圖，不改原版狀態。
package hd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"os"
	"path/filepath"
)

type manualTile struct {
	raw      [64]byte
	rgb      [192]byte
	art      image.Image
	matchRGB bool
}
type manualPlacement struct {
	tile *manualTile
	x, y int
}
type ManualArtwork struct {
	index            map[[16]byte][]*manualTile
	zero             map[[64]byte][]*manualTile
	lastRaw, lastRGB []byte
	placements       []manualPlacement
}

func NewManualArtwork(root, dir string) (*ManualArtwork, error) {
	if _, err := checked(filepath.Join(root, "ZELIARD.EXE"), "fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341", 1<<20); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(dir, "manual-tiles.json"))
	if err != nil {
		return nil, err
	}
	var m struct {
		Tiles []struct {
			Raster, RGB []byte
			File        string
			Hash        string `json:"sha256"`
		}
	}
	if err = json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if len(m.Tiles) == 0 || len(m.Tiles) > 32768 {
		return nil, fmt.Errorf("人工圖格數量不符")
	}
	s := &ManualArtwork{index: make(map[[16]byte][]*manualTile), zero: make(map[[64]byte][]*manualTile)}
	for _, r := range m.Tiles {
		if len(r.Raster) != 64 || len(r.RGB) != 0 && len(r.RGB) != 192 || filepath.Base(r.File) != r.File {
			return nil, fmt.Errorf("人工圖格來源不符")
		}
		im, e := checkedTerrainImage(filepath.Join(dir, r.File), r.Hash, 24, 24)
		if e != nil {
			return nil, e
		}
		t := &manualTile{art: im, matchRGB: len(r.RGB) == 192}
		copy(t.raw[:], r.Raster)
		copy(t.rgb[:], r.RGB)
		var signature [16]byte
		copy(signature[:], t.raw[24:40])
		if signature == [16]byte{} {
			s.zero[t.raw] = append(s.zero[t.raw], t)
		} else {
			s.index[signature] = append(s.index[signature], t)
		}
	}
	return s, nil
}

func (s *ManualArtwork) Draw(dst *image.RGBA, raw, rgb []byte, origin image.Point) bool {
	if s == nil || len(raw) != 64000 || len(rgb) != 192000 {
		return false
	}
	if !bytes.Equal(raw, s.lastRaw) || !bytes.Equal(rgb, s.lastRGB) {
		s.placements = s.placements[:0]
		var used [64000]bool
		for y := 0; y <= 192; y++ {
			for x := 0; x <= 312; x++ {
				if used[y*320+x] {
					continue
				}
				var signature [16]byte
				copy(signature[:8], raw[(y+3)*320+x:(y+3)*320+x+8])
				copy(signature[8:], raw[(y+4)*320+x:(y+4)*320+x+8])
				candidates := s.index[signature]
				if signature == [16]byte{} {
					var source [64]byte
					for row := 0; row < 8; row++ {
						copy(source[row*8:row*8+8], raw[(y+row)*320+x:(y+row)*320+x+8])
					}
					candidates = s.zero[source]
				}
				for _, tile := range candidates {
					match := true
					for row := 0; row < 8 && match; row++ {
						i := (y+row)*320 + x
						match = bytes.Equal(raw[i:i+8], tile.raw[row*8:row*8+8]) && (!tile.matchRGB || bytes.Equal(rgb[i*3:(i+8)*3], tile.rgb[row*24:row*24+24]))
						for col := 0; col < 8 && match; col++ {
							match = !used[i+col]
						}
					}
					if !match {
						continue
					}
					s.placements = append(s.placements, manualPlacement{tile, x, y})
					for row := 0; row < 8; row++ {
						for col := 0; col < 8; col++ {
							used[(y+row)*320+x+col] = true
						}
					}
					break
				}
			}
		}
		s.lastRaw = append(s.lastRaw[:0], raw...)
		s.lastRGB = append(s.lastRGB[:0], rgb...)
	}
	for _, p := range s.placements {
		// 已由專用HD圖層畫好的區域不再覆寫。
		native := true
		for y := 0; y < 24 && native; y++ {
			for x := 0; x < 24; x++ {
				i := ((p.y+y/3)*320 + p.x + x/3) * 3
				c := dst.RGBAAt(p.x*3+x+origin.X, p.y*3+y+origin.Y)
				if c.R != rgb[i] || c.G != rgb[i+1] || c.B != rgb[i+2] {
					native = false
					break
				}
			}
		}
		if native {
			draw.Draw(dst, image.Rect(p.x*3, p.y*3, p.x*3+24, p.y*3+24).Add(origin), p.tile.art, image.Point{}, draw.Over)
		}
	}
	return len(s.placements) > 0
}

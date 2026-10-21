// 全圖補繪只讀取來源及畫布；完整圖樣、資源段與透明像素同時決定是否覆繪。
package hd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"sort"
)

type completeBinding struct {
	Segment string
	Offset  int
	Data    []byte
}
type completeGroup struct {
	Key      string
	Bindings []completeBinding
}
type completeEntry struct {
	Group, Width, Height int
	Source, File         string
	Hash                 string `json:"sha256"`
	Raster, Mask, Colors []byte
	art                  image.Image
	ax, ay, anchorLen    int
	baseline             byte
}
type completePlacement struct {
	entry *completeEntry
	x, y  int
}
type completeSignature struct {
	length byte
	value  [8]byte
}
type CompleteArtwork struct {
	groups             []completeGroup
	entries            []*completeEntry
	components         map[string]map[string]image.Image
	index              map[completeSignature][]*completeEntry
	active, lastActive []bool
	lastRaw            []byte
	placements         []completePlacement
}

func loadCompleteArtwork(dir string) (*CompleteArtwork, error) {
	data, err := os.ReadFile(filepath.Join(dir, "complete-art.json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var m struct {
		Schema, Scale       int
		Groups              []completeGroup
		Entries, Components []completeEntry
	}
	if err = json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.Schema != 1 || m.Scale != 3 || len(m.Entries) == 0 || len(m.Entries) > 16384 || len(m.Groups) > 128 {
		return nil, fmt.Errorf("全圖補繪清單不符")
	}
	s := &CompleteArtwork{groups: m.Groups, index: map[completeSignature][]*completeEntry{}, components: map[string]map[string]image.Image{}, active: make([]bool, len(m.Groups))}
	for _, g := range m.Groups {
		for _, b := range g.Bindings {
			if (b.Segment != "r" && b.Segment != "e") || b.Offset < 0 || len(b.Data) == 0 || b.Offset+len(b.Data) > 65536 {
				return nil, fmt.Errorf("全圖資源守衛不符")
			}
		}
	}
	for i := range m.Entries {
		e := &m.Entries[i]
		if e.Group < 0 || e.Group >= len(s.groups) || e.Width < 1 || e.Height < 1 || e.Width > 320 || e.Height > 200 || len(e.Raster) != e.Width*e.Height || (len(e.Mask) != 0 && len(e.Mask) != len(e.Raster)) || len(e.Colors) != 768 || filepath.Base(e.File) != e.File {
			return nil, fmt.Errorf("全圖來源幾何不符")
		}
		im, err := checkedTerrainImage(filepath.Join(dir, e.File), e.Hash, e.Width*3, e.Height*3)
		if err != nil {
			return nil, err
		}
		e.art = im
		for j, v := range e.Raster {
			if len(e.Mask) != 0 && e.Mask[j] == 0 {
				continue
			}
			for _, c := range e.Colors[int(v)*3 : int(v)*3+3] {
				if c > e.baseline {
					e.baseline = c
				}
			}
		}
		best := -1
		for _, n := range []int{8, 4, 2, 1} {
			for y := 0; y < e.Height; y++ {
				for x := 0; x+n <= e.Width; x++ {
					var seen [256]bool
					score := 0
					valid := true
					for k := 0; k < n; k++ {
						at := y*e.Width + x + k
						if len(e.Mask) != 0 && e.Mask[at] == 0 {
							valid = false
							break
						}
						if !seen[e.Raster[at]] {
							score++
							seen[e.Raster[at]] = true
						}
					}
					if valid && score > best {
						best = score
						e.ax = x
						e.ay = y
						e.anchorLen = n
					}
				}
			}
			if best >= 2 {
				break
			}
		}
		// 純色空白仍保留在清單及人物組裝器，避免以大面積單色誤認場景。
		if best < 2 {
			continue
		}
		var signature completeSignature
		signature.length = byte(e.anchorLen)
		at := e.ay*e.Width + e.ax
		copy(signature.value[:], e.Raster[at:at+e.anchorLen])
		s.index[signature] = append(s.index[signature], e)
		s.entries = append(s.entries, e)
	}
	for _, e := range m.Components {
		if e.Width != 8 || e.Height != 8 || len(e.Raster) != 64 || filepath.Base(e.File) != e.File {
			return nil, fmt.Errorf("人物補繪元件不符")
		}
		im, err := checkedTerrainImage(filepath.Join(dir, e.File), e.Hash, 24, 24)
		if err != nil {
			return nil, err
		}
		if s.components[e.Source] == nil {
			s.components[e.Source] = map[string]image.Image{}
		}
		key := townBackgroundHash(e.Raster)
		if s.components[e.Source][key] == nil {
			s.components[e.Source][key] = im
		}
	}
	return s, nil
}

func (s *CompleteArtwork) resources(r, e []byte) {
	if s == nil {
		return
	}
	for n, g := range s.groups {
		active := true
		for _, b := range g.Bindings {
			mem := r
			if b.Segment == "e" {
				mem = e
			}
			if len(mem) < b.Offset+len(b.Data) || !bytes.Equal(mem[b.Offset:b.Offset+len(b.Data)], b.Data) {
				active = false
				break
			}
		}
		s.active[n] = active
	}
}

func (s *CompleteArtwork) draw(dst *image.RGBA, raw, rgb []byte, origin image.Point) bool {
	if s == nil {
		return false
	}
	changed := len(s.lastActive) != len(s.active)
	if !changed {
		for i, v := range s.active {
			if s.lastActive[i] != v {
				changed = true
				break
			}
		}
	}
	if changed || !bytes.Equal(raw, s.lastRaw) {
		s.placements = s.placements[:0]
		for y := 0; y < 200; y++ {
			for x := 0; x < 320; x++ {
				for _, n := range []int{8, 4, 2, 1} {
					if x+n > 320 {
						continue
					}
					signature := completeSignature{length: byte(n)}
					copy(signature.value[:], raw[y*320+x:y*320+x+n])
					for _, e := range s.index[signature] {
						if !s.active[e.Group] {
							continue
						}
						px, py := x-e.ax, y-e.ay
						if px < 0 || py < 0 || px+e.Width > 320 || py+e.Height > 200 {
							continue
						}
						match := true
						for row := 0; row < e.Height && match; row++ {
							for col := 0; col < e.Width; col++ {
								at := row*e.Width + col
								if len(e.Mask) != 0 && e.Mask[at] == 0 {
									continue
								}
								if raw[(py+row)*320+px+col] != e.Raster[at] {
									match = false
									break
								}
							}
						}
						if match {
							s.placements = append(s.placements, completePlacement{e, px, py})
						}
					}
				}
			}
		}
		sort.SliceStable(s.placements, func(i, j int) bool {
			a, b := s.placements[i].entry, s.placements[j].entry
			return a.Width*a.Height > b.Width*b.Height
		})
		s.lastRaw = append(s.lastRaw[:0], raw...)
		s.lastActive = append(s.lastActive[:0], s.active...)
	}
	painted := false
	for _, p := range s.placements {
		e := p.entry
		brightness := byte(0)
		for y := 0; y < e.Height; y++ {
			for x := 0; x < e.Width; x++ {
				at := y*e.Width + x
				if len(e.Mask) != 0 && e.Mask[at] == 0 {
					continue
				}
				i := ((p.y+y)*320 + p.x + x) * 3
				for _, v := range rgb[i : i+3] {
					if v > brightness {
						brightness = v
					}
				}
			}
		}
		for y := 0; y < e.Height; y++ {
			for x := 0; x < e.Width; x++ {
				at := y*e.Width + x
				if len(e.Mask) != 0 && e.Mask[at] == 0 {
					continue
				}
				i := ((p.y+y)*320 + p.x + x) * 3
				native := true
				for dy := 0; dy < 3 && native; dy++ {
					for dx := 0; dx < 3; dx++ {
						c := dst.RGBAAt((p.x+x)*3+dx+origin.X, (p.y+y)*3+dy+origin.Y)
						if c.R != rgb[i] || c.G != rgb[i+1] || c.B != rgb[i+2] {
							native = false
							break
						}
					}
				}
				if !native {
					continue
				}
				for dy := 0; dy < 3; dy++ {
					for dx := 0; dx < 3; dx++ {
						c := color.NRGBAModel.Convert(e.art.At(x*3+dx, y*3+dy)).(color.NRGBA)
						if e.baseline > 0 && brightness < e.baseline {
							c.R = uint8(uint32(c.R) * uint32(brightness) / uint32(e.baseline))
							c.G = uint8(uint32(c.G) * uint32(brightness) / uint32(e.baseline))
							c.B = uint8(uint32(c.B) * uint32(brightness) / uint32(e.baseline))
						}
						inv, a := uint32(255-c.A), uint32(c.A)
						dst.SetRGBA((p.x+x)*3+dx+origin.X, (p.y+y)*3+dy+origin.Y, color.RGBA{uint8((uint32(c.R)*a + uint32(rgb[i])*inv + 127) / 255), uint8((uint32(c.G)*a + uint32(rgb[i+1])*inv + 127) / 255), uint8((uint32(c.B)*a + uint32(rgb[i+2])*inv + 127) / 255), 255})
						painted = true
					}
				}
			}
		}
	}
	return painted
}

// 以原版九／十六個8×8元件重組新姿態，不建立新的動畫或座標。
func (s *CompleteArtwork) compose(source string, raster []byte, width int) ([]color.NRGBA, image.Rectangle) {
	if s == nil || len(raster) != width*width || width%8 != 0 {
		return nil, image.Rectangle{}
	}
	parts := s.components[source]
	if len(parts) == 0 {
		return nil, image.Rectangle{}
	}
	canvas := image.NewNRGBA(image.Rect(0, 0, width*3, width*3))
	bounds := image.Rectangle{}
	for y := 0; y < width; y += 8 {
		for x := 0; x < width; x += 8 {
			var tile [64]byte
			empty := true
			for row := 0; row < 8; row++ {
				copy(tile[row*8:row*8+8], raster[(y+row)*width+x:(y+row)*width+x+8])
				for _, v := range tile[row*8 : row*8+8] {
					if v != 255 {
						empty = false
					}
				}
			}
			if empty {
				continue
			}
			im := parts[townBackgroundHash(tile[:])]
			if im == nil {
				return nil, image.Rectangle{}
			}
			for dy := 0; dy < 24; dy++ {
				for dx := 0; dx < 24; dx++ {
					c := color.NRGBAModel.Convert(im.At(dx, dy)).(color.NRGBA)
					canvas.SetNRGBA(x*3+dx, y*3+dy, c)
					if c.A != 0 {
						r := image.Rect(x*3+dx, y*3+dy, x*3+dx+1, y*3+dy+1)
						if bounds.Empty() {
							bounds = r
						} else {
							bounds = bounds.Union(r)
						}
					}
				}
			}
		}
	}
	pixels := make([]color.NRGBA, width*width*9)
	for y := 0; y < width*3; y++ {
		for x := 0; x < width*3; x++ {
			pixels[y*width*3+x] = canvas.NRGBAAt(x, y)
		}
	}
	return pixels, bounds
}

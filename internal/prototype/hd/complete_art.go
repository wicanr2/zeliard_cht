// 全圖補繪只讀取來源及畫布；完整圖樣、資源段與透明像素同時決定是否覆繪。
package hd

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	DerivativeOf         string `json:"derivative_of"`
	Raster, Mask, Colors []byte
	art                  image.Image
	pixels               []color.NRGBA
	ax, ay, anchorLen    int
	baseline             byte
	context              *completeEntry
	contextX, contextY   int
}
type completePlacement struct {
	entry *completeEntry
	x, y  int
}
type CompleteArtwork struct {
	groups             []completeGroup
	entries            []*completeEntry
	components         map[string]map[string]image.Image
	index              [4]map[uint64][]*completeEntry
	prefix             [65536]uint8
	singles            [256][]*completeEntry
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
	s := &CompleteArtwork{groups: m.Groups, components: map[string]map[string]image.Image{}, active: make([]bool, len(m.Groups))}
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
		e.pixels = make([]color.NRGBA, e.Width*e.Height*9)
		for y := 0; y < e.Height*3; y++ {
			for x := 0; x < e.Width*3; x++ {
				e.pixels[y*e.Width*3+x] = color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
			}
		}
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
		s.entries = append(s.entries, e)
	}
	parents := make(map[string]*completeEntry, len(m.Entries))
	for i := range m.Entries {
		e := &m.Entries[i]
		if parents[e.File] != nil {
			return nil, fmt.Errorf("全圖來源檔名重複")
		}
		parents[e.File] = e
	}
	for _, e := range s.entries {
		if e.DerivativeOf == "" || len(s.groups[e.Group].Bindings) != 0 {
			continue
		}
		parent := parents[e.DerivativeOf]
		if parent == nil || parent.Group != e.Group || parent.Source != e.Source || parent.DerivativeOf != "" {
			return nil, fmt.Errorf("插畫片段母圖不符：%s", e.File)
		}
		// schema1既有登錄器以母圖名及-xNNN-yNNN保存原生裁切座標。
		prefix := strings.TrimSuffix(parent.File, ".png")
		var x, y int
		_, err := fmt.Sscanf(strings.TrimPrefix(e.File, prefix), "-x%d-y%d.png", &x, &y)
		if err != nil || e.File != fmt.Sprintf("%s-x%03d-y%03d.png", prefix, x, y) || x < 0 || y < 0 || x+e.Width > parent.Width || y+e.Height > parent.Height {
			return nil, fmt.Errorf("插畫片段座標不符：%s", e.File)
		}
		for row := 0; row < e.Height; row++ {
			if !bytes.Equal(e.Raster[row*e.Width:(row+1)*e.Width], parent.Raster[(y+row)*parent.Width+x:(y+row)*parent.Width+x+e.Width]) {
				return nil, fmt.Errorf("插畫片段來源不符：%s", e.File)
			}
		}
		e.context, e.contextX, e.contextY = parent, x, y
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
	changed := false
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
		changed = changed || s.active[n] != active
		s.active[n] = active
	}
	if changed {
		s.reindex()
	}
}

// 索引只含當前資源；兩byte前綴先排除無候選位置，避免每像素四次大型雜湊查詢。
// 完整錨點、遮罩、整幅來源及原繪製順序仍逐項核對。
func (s *CompleteArtwork) reindex() {
	s.prefix = [65536]uint8{}
	s.singles = [256][]*completeEntry{}
	for i := range s.index {
		s.index[i] = make(map[uint64][]*completeEntry)
	}
	for _, e := range s.entries {
		if !s.active[e.Group] {
			continue
		}
		at := e.ay*e.Width + e.ax
		anchor := e.Raster[at : at+e.anchorLen]
		if e.anchorLen == 1 {
			s.singles[anchor[0]] = append(s.singles[anchor[0]], e)
			continue
		}
		level := 1
		if e.anchorLen == 4 {
			level = 2
		} else if e.anchorLen == 8 {
			level = 3
		}
		key := completeAnchor(anchor)
		s.index[level][key] = append(s.index[level][key], e)
		s.prefix[binary.LittleEndian.Uint16(anchor)] |= 1 << uint(level)
	}
}

func completeAnchor(b []byte) uint64 {
	switch len(b) {
	case 8:
		return binary.LittleEndian.Uint64(b)
	case 4:
		return uint64(binary.LittleEndian.Uint32(b))
	case 2:
		return uint64(binary.LittleEndian.Uint16(b))
	}
	return uint64(b[0])
}

// 無記憶體綁定的插畫片段須同時符合母圖可見上下文，防止單一色點跨場景誤貼。
func (e *completeEntry) contextMatches(raw []byte, x, y int) bool {
	parent := e.context
	if parent == nil {
		return true
	}
	left, top := x-e.contextX, y-e.contextY
	visible := image.Rect(left, top, left+parent.Width, top+parent.Height).Intersect(image.Rect(0, 0, 320, 200))
	if visible.Empty() {
		return false
	}
	for py := visible.Min.Y; py < visible.Max.Y; py++ {
		for px := visible.Min.X; px < visible.Max.X; px++ {
			at := (py-top)*parent.Width + px - left
			if len(parent.Mask) != 0 && parent.Mask[at] == 0 {
				continue
			}
			if raw[py*320+px] != parent.Raster[at] {
				return false
			}
		}
	}
	return true
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
				var bits uint8
				if x+2 <= 320 {
					bits = s.prefix[binary.LittleEndian.Uint16(raw[y*320+x:y*320+x+2])]
				}
				for level := 3; level >= 0; level-- {
					n := 1 << uint(level)
					if x+n > 320 {
						continue
					}
					var candidates []*completeEntry
					if level == 0 {
						candidates = s.singles[raw[y*320+x]]
					} else {
						if bits&(1<<uint(level)) == 0 {
							continue
						}
						candidates = s.index[level][completeAnchor(raw[y*320+x:y*320+x+n])]
					}
					for _, e := range candidates {
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
						if match && e.contextMatches(raw, px, py) {
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
						c := e.pixels[(y*3+dy)*e.Width*3+x*3+dx]
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

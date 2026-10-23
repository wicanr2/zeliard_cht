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
	// background 是母圖最常見的兩種非黑色（多半是背景）；局部母圖只以其餘顏色當證據。規格068。
	background [2]int
	// keySamples 是均勻取出的最多512個特徵點位置，供局部母圖先抽樣篩選。
	keySamples []int
	// distinct 是片段未遮罩點的色數；局部母圖只由三色以上的片段定位，兩色抖色片段在畫面上到處都對得上。
	distinct int
}
type completePlacement struct {
	entry *completeEntry
	x, y  int
	// cells 非空時只畫標記的點（被局部蓋住的母圖），其餘維持原版。
	cells []bool
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
	// held 是上一次比對到的大插畫；Held 是沿用次數。規格067。
	held    []completePlacement
	holding bool
	Held    int
	// introMode 是原版 FF77 非零的片頭／結局模式；只有這時沿用說話中的大插畫，片段上下文也只在這時放寬。
	introMode, lastIntro bool
}

func loadCompleteArtwork(dir string) (*CompleteArtwork, error) {
	return loadCompleteArtworkMode(dir, false)
}

func loadCompleteArtworkMode(dir string, lazy bool) (*CompleteArtwork, error) {
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
		im, err := deferredTerrainImage(filepath.Join(dir, e.File), e.Hash, e.Width*3, e.Height*3, lazy)
		if err != nil {
			return nil, err
		}
		e.art = im
		if !lazy {
			e.ensurePixels()
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
		var used [256]bool
		for j, v := range e.Raster {
			if (len(e.Mask) == 0 || e.Mask[j] != 0) && !used[v] {
				used[v] = true
				e.distinct++
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
		if parent.background == [2]int{} {
			parent.background = dominantColors(parent)
			parent.keySamples = keySamples(parent)
		}
	}
	for _, e := range m.Components {
		if e.Width != 8 || e.Height != 8 || len(e.Raster) != 64 || filepath.Base(e.File) != e.File {
			return nil, fmt.Errorf("人物補繪元件不符")
		}
		im, err := deferredTerrainImage(filepath.Join(dir, e.File), e.Hash, 24, 24, lazy)
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

func (e *completeEntry) ensurePixels() bool {
	if len(e.pixels) != 0 {
		return true
	}
	im := resolvedImage(e.art)
	if im == nil {
		return false
	}
	e.pixels = make([]color.NRGBA, e.Width*e.Height*9)
	for y := 0; y < e.Height*3; y++ {
		for x := 0; x < e.Width*3; x++ {
			e.pixels[y*e.Width*3+x] = color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
		}
	}
	return true
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
// relaxed 用於片頭／結局：原版會以另一張插畫局部蓋住（製作名單騎士圖蓋住公主圖），
// 母圖可見部分的特徵點過半相符即接受。特徵點是黑色與母圖兩種主要背景色以外的點：
// 大片黑色的插畫在全黑畫面上、人物被擦掉後只剩背景抖色的插畫，都只有背景相符。規格068。
func (e *completeEntry) contextMatches(raw []byte, x, y int, relaxed bool) bool {
	diff, keyDiff, keyTotal, ok := e.contextDiff(raw, x, y, !relaxed)
	return ok && (diff == 0 || relaxed && partialContext(keyDiff, keyTotal))
}

// partialMinKey 是局部母圖至少要相符的特徵點數。
const partialMinKey = 256

func partialContext(keyDiff, keyTotal int) bool {
	return keyDiff*2 < keyTotal && keyTotal-keyDiff >= partialMinKey
}

// dominantColors 回報母圖最常見的兩種非黑色；沒有就是-1。
func dominantColors(e *completeEntry) [2]int {
	var count [256]int
	for i, v := range e.Raster {
		if v != 0 && (len(e.Mask) == 0 || e.Mask[i] != 0) {
			count[v]++
		}
	}
	top := [2]int{-1, -1}
	for c := 1; c < 256; c++ {
		switch {
		case count[c] == 0:
		case top[0] < 0 || count[c] > count[top[0]]:
			top[1], top[0] = top[0], c
		case top[1] < 0 || count[c] > count[top[1]]:
			top[1] = c
		}
	}
	return top
}

// contextDiff 回報母圖可見部分與畫面不同的點數，以及其中特徵點的不同數與總數；strict 時遇到第一個不同就停。
func (e *completeEntry) contextDiff(raw []byte, x, y int, strict bool) (diff, keyDiff, keyTotal int, ok bool) {
	parent := e.context
	if parent == nil {
		return 0, 0, 0, true
	}
	left, top := x-e.contextX, y-e.contextY
	visible := image.Rect(left, top, left+parent.Width, top+parent.Height).Intersect(image.Rect(0, 0, 320, 200))
	if visible.Empty() {
		return 0, 0, 0, false
	}
	for py := visible.Min.Y; py < visible.Max.Y; py++ {
		for px := visible.Min.X; px < visible.Max.X; px++ {
			at := (py-top)*parent.Width + px - left
			if len(parent.Mask) != 0 && parent.Mask[at] == 0 {
				continue
			}
			v := int(parent.Raster[at])
			key := v != 0 && v != parent.background[0] && v != parent.background[1]
			if key {
				keyTotal++
			}
			if raw[py*320+px] != parent.Raster[at] {
				diff++
				if key {
					keyDiff++
				}
				if strict {
					return diff, keyDiff, keyTotal, true
				}
			}
		}
	}
	return diff, keyDiff, keyTotal, true
}

// keySamples 均勻取出最多512個特徵點的位置。
func keySamples(e *completeEntry) []int {
	var keys []int
	for i, v := range e.Raster {
		if v != 0 && int(v) != e.background[0] && int(v) != e.background[1] && (len(e.Mask) == 0 || e.Mask[i] != 0) {
			keys = append(keys, i)
		}
	}
	if len(keys) <= 512 {
		return keys
	}
	out := make([]int, 512)
	for i := range out {
		out[i] = keys[i*len(keys)/512]
	}
	return out
}

// partialLikely 以抽樣特徵點先篩選局部母圖：兩色抖色片段在畫面上到處都對得上，每個巧合位置都完整掃描
// 整張母圖，片頭一次比對會由約22毫秒升到約0.4秒。回報抽樣相符比例；可見樣本不足32點回報0。規格068。
func (e *completeEntry) partialLikely(raw []byte, left, top int) float64 {
	seen, same := 0, 0
	for _, at := range e.keySamples {
		x, y := left+at%e.Width, top+at/e.Width
		if x < 0 || y < 0 || x >= 320 || y >= 200 {
			continue
		}
		seen++
		if raw[y*320+x] == e.Raster[at] {
			same++
		}
	}
	if seen < 32 {
		return 0
	}
	return float64(same) / float64(seen)
}

// completeContext 是同一次比對中母圖與其原生左上角；同一畫面每個位置只核對一次上下文。
type completeContext struct {
	parent    *completeEntry
	left, top int
}

// partialCells 標出母圖在畫面上可畫的點（依母圖8×8格線，畫面外不計）。整格逐點相同就整格畫；
// 與這類格相鄰、被遮擋物切過的交界格，只畫其中逐點相同的非黑色點。遮擋物內部的格不畫，避免巧合同色點散落。規格068。
func partialCells(e *completeEntry, raw []byte, left, top int) []bool {
	cells := make([]bool, e.Width*e.Height)
	cw, ch := (e.Width+7)/8, (e.Height+7)/8
	full := make([]bool, cw*ch)
	box := func(cx, cy int) image.Rectangle {
		return image.Rect(left+cx*8, top+cy*8, left+min(cx*8+8, e.Width), top+min(cy*8+8, e.Height)).Intersect(image.Rect(0, 0, 320, 200))
	}
	same := func(px, py int) bool {
		at := (py-top)*e.Width + px - left
		return len(e.Mask) != 0 && e.Mask[at] == 0 || raw[py*320+px] == e.Raster[at]
	}
	for cy := 0; cy < ch; cy++ {
		for cx := 0; cx < cw; cx++ {
			r := box(cx, cy)
			if r.Empty() {
				continue
			}
			ok := true
			for py := r.Min.Y; py < r.Max.Y && ok; py++ {
				for px := r.Min.X; px < r.Max.X && ok; px++ {
					ok = same(px, py)
				}
			}
			full[cy*cw+cx] = ok
		}
	}
	for cy := 0; cy < ch; cy++ {
		for cx := 0; cx < cw; cx++ {
			edge := !full[cy*cw+cx] && (cx > 0 && full[cy*cw+cx-1] || cx+1 < cw && full[cy*cw+cx+1] || cy > 0 && full[(cy-1)*cw+cx] || cy+1 < ch && full[(cy+1)*cw+cx])
			if !full[cy*cw+cx] && !edge {
				continue
			}
			r := box(cx, cy)
			for py := r.Min.Y; py < r.Max.Y; py++ {
				for px := r.Min.X; px < r.Max.X; px++ {
					cells[(py-top)*e.Width+px-left] = full[cy*cw+cx] || e.Raster[(py-top)*e.Width+px-left] != 0 && same(px, py)
				}
			}
		}
	}
	return cells
}

func (s *CompleteArtwork) draw(dst *image.RGBA, raw, rgb []byte, origin image.Point) bool {
	if s == nil {
		return false
	}
	changed := len(s.lastActive) != len(s.active) || s.lastIntro != s.introMode
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
		// 0拒絕、1可見部分全部相符、2片頭／結局模式下過半相符（只畫逐點相同的格）。
		contexts := map[completeContext]int{}
		// 每張母圖最多一個局部位置：特徵點相符比例最高者。大片平滑區錯開幾點也會過半相同，
		// 不篩掉會在一百多個錯位位置重複畫。母圖在某處完全相符時不找局部。
		type partialCandidate struct {
			key         completeContext
			same, total int
			sample      float64
		}
		partial := map[*completeEntry]partialCandidate{}
		exact := map[*completeEntry]bool{}
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
						if match && e.context != nil {
							key := completeContext{e.context, px - e.contextX, py - e.contextY}
							state, seen := contexts[key]
							if !seen {
								// 先嚴格核對（第一個不同就停）；片頭／結局模式才以抽樣篩選後完整核對局部母圖。
								diff, _, _, ok := e.contextDiff(raw, px, py, true)
								if ok && diff == 0 {
									state = 1
									exact[e.context] = true
								} else if ok && s.introMode && e.distinct >= 3 && !exact[e.context] {
									// 抽樣過半，且不比目前最佳位置低0.1以上，才完整核對。
									best, has := partial[e.context]
									if sample := e.context.partialLikely(raw, key.left, key.top); sample >= 0.5 && (!has || sample >= best.sample-0.1) {
										if _, keyDiff, keyTotal, _ := e.contextDiff(raw, px, py, false); partialContext(keyDiff, keyTotal) {
											state = 2
											if same := keyTotal - keyDiff; !has || same*best.total > best.same*keyTotal {
												partial[e.context] = partialCandidate{key, same, keyTotal, sample}
											}
										}
									}
								}
								contexts[key] = state
							}
							// 過半相符時由母圖逐格畫出，片段本身不另放；片段去重後只登錄首見位置，不足以補滿。
							match = state == 1
						}
						if match {
							s.placements = append(s.placements, completePlacement{entry: e, x: px, y: py})
						}
					}
				}
			}
		}
		s.holdLarge(raw)
		sort.SliceStable(s.placements, func(i, j int) bool {
			a, b := s.placements[i].entry, s.placements[j].entry
			return a.Width*a.Height > b.Width*b.Height
		})
		// 局部母圖排在最後，只補完整插畫與沿用都沒畫到的格。
		// 依母圖在清單中的順序加入，結果與map走訪順序無關。
		for _, e := range s.entries {
			c, has := partial[e]
			if !has || exact[e] {
				continue
			}
			s.placements = append(s.placements, completePlacement{entry: e, x: c.key.left, y: c.key.top, cells: partialCells(e, raw, c.key.left, c.key.top)})
		}
		s.lastRaw = append(s.lastRaw[:0], raw...)
		s.lastActive = append(s.lastActive[:0], s.active...)
		s.lastIntro = s.introMode
	}
	painted := false
	for _, p := range s.placements {
		e := p.entry
		if !e.ensurePixels() {
			continue
		}
		// 放置範圍完全在畫布內時直接讀寫Pix，結果與RGBAAt／SetRGBA逐點相同；片頭整張插畫每次比對約省12毫秒。規格068。
		fast := paintDirect && image.Rect(p.x*3+origin.X, p.y*3+origin.Y, (p.x+e.Width)*3+origin.X, (p.y+e.Height)*3+origin.Y).In(dst.Rect)
		brightness := byte(0)
		for y := 0; y < e.Height; y++ {
			for x := 0; x < e.Width; x++ {
				at := y*e.Width + x
				if len(e.Mask) != 0 && e.Mask[at] == 0 || p.cells != nil && !p.cells[at] {
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
				if len(e.Mask) != 0 && e.Mask[at] == 0 || p.cells != nil && !p.cells[at] {
					continue
				}
				i := ((p.y+y)*320 + p.x + x) * 3
				if fast {
					painted = paintFast(dst, e, x, y, (p.x+x)*3+origin.X, (p.y+y)*3+origin.Y, rgb[i:i+3], brightness) || painted
					continue
				}
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
			im := resolvedImage(parts[townBackgroundHash(tile[:])])
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

// completeHoldArea 是可沿用的大插畫最小面積（64×64）。
const completeHoldArea = 64 * 64

// holdLarge 處理結局人物說話：原版在臉上貼小圖（enddemo 8x..Cx 控制碼，最大約36×32），整張插畫
// 比對不到，只剩局部碎片，畫面成為 HD 與原版交錯。上一次已比對到的大插畫，同一位置只有八分之一以內的點
// 不同時沿用整張，原版的說話小圖不顯示。畫面換掉或逐點清除超過八分之一就停止。規格067。
func (s *CompleteArtwork) holdLarge(raw []byte) {
	s.holding = false
	var large []completePlacement
	for _, p := range s.placements {
		if p.entry.Width*p.entry.Height >= completeHoldArea {
			large = append(large, p)
		}
	}
	for _, p := range s.held {
		if !s.introMode {
			// 遊戲中的大背景（如 ckpd、ympd）上會有角色移動，沿用會把角色蓋掉。
			break
		}
		found := false
		for _, q := range large {
			if q.entry == p.entry && q.x == p.x && q.y == p.y {
				found = true
				break
			}
		}
		if found || !s.active[p.entry.Group] {
			continue
		}
		diff, total := 0, 0
		for row := 0; row < p.entry.Height; row++ {
			for col := 0; col < p.entry.Width; col++ {
				at := row*p.entry.Width + col
				if len(p.entry.Mask) != 0 && p.entry.Mask[at] == 0 {
					continue
				}
				total++
				if raw[(p.y+row)*320+p.x+col] != p.entry.Raster[at] {
					diff++
				}
			}
		}
		if total > 0 && diff*8 <= total {
			s.placements = append(s.placements, p)
			large = append(large, p)
			s.Held++
			s.holding = true
		}
	}
	s.held = large
}

// largeCovers 回報這次比對（含沿用）有大插畫完整蓋住 r。
func (s *CompleteArtwork) largeCovers(r image.Rectangle) bool {
	if s == nil {
		return false
	}
	for _, p := range s.held {
		if r.In(image.Rect(p.x, p.y, p.x+p.entry.Width, p.y+p.entry.Height)) {
			return true
		}
	}
	return false
}

// paintDirect 只供測試比對兩條路徑。
var paintDirect = true

// paintFast 是draw的直接Pix版本：(dx0,dy0)起的3×3點都還是原版色時，以HD點覆蓋，否則不畫。
func paintFast(dst *image.RGBA, e *completeEntry, x, y, dx0, dy0 int, native []byte, brightness byte) bool {
	for dy := 0; dy < 3; dy++ {
		o := dst.PixOffset(dx0, dy0+dy)
		for dx := 0; dx < 3; dx++ {
			if dst.Pix[o] != native[0] || dst.Pix[o+1] != native[1] || dst.Pix[o+2] != native[2] {
				return false
			}
			o += 4
		}
	}
	dim := e.baseline > 0 && brightness < e.baseline
	for dy := 0; dy < 3; dy++ {
		o := dst.PixOffset(dx0, dy0+dy)
		row := (y*3+dy)*e.Width*3 + x*3
		for dx := 0; dx < 3; dx++ {
			c := e.pixels[row+dx]
			if dim {
				c.R = uint8(uint32(c.R) * uint32(brightness) / uint32(e.baseline))
				c.G = uint8(uint32(c.G) * uint32(brightness) / uint32(e.baseline))
				c.B = uint8(uint32(c.B) * uint32(brightness) / uint32(e.baseline))
			}
			inv, a := uint32(255-c.A), uint32(c.A)
			dst.Pix[o] = uint8((uint32(c.R)*a + uint32(native[0])*inv + 127) / 255)
			dst.Pix[o+1] = uint8((uint32(c.G)*a + uint32(native[1])*inv + 127) / 255)
			dst.Pix[o+2] = uint8((uint32(c.B)*a + uint32(native[2])*inv + 127) / 255)
			dst.Pix[o+3] = 255
			o += 4
		}
	}
	return true
}

// 規格057：完整背景錨點確認場景，每個原生8×8圖格獨立核對。
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

type transitionPose struct {
	Width, Height int
	Raster        []byte
	File          string
	Hash          string `json:"sha256"`
	art           image.Image
}
type caveTransition struct {
	Schema, Scale, Width, Height int
	File                         string
	Hash                         string `json:"sha256"`
	Raster, Colors               []byte
	Anchors                      [][4]int
	Poses                        []transitionPose
	art                          image.Image
	patterns                     map[string]*transitionPose
	matched                      bool
	matchedTiles, originalTiles  int
}

func transitionPoseKey(w, h int, raw []byte) string { return fmt.Sprintf("%d:%d:", w, h) + string(raw) }

func loadCaveTransition(dir string) (*caveTransition, error) {
	data, e := os.ReadFile(filepath.Join(dir, "cave-transition.json"))
	if os.IsNotExist(e) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	var s caveTransition
	if e = json.Unmarshal(data, &s); e != nil {
		return nil, e
	}
	if s.Schema != 1 || s.Scale != 3 || s.Width != 224 || s.Height != 144 || len(s.Raster) != 224*144 || len(s.Colors) != 768 || len(s.Poses) == 0 || len(s.Poses) > 1024 || len(s.Anchors) != 2 || filepath.Base(s.File) != s.File {
		return nil, fmt.Errorf("入洞過場來源不符")
	}
	for _, a := range s.Anchors {
		if a[0] < 0 || a[1] < 0 || a[2] > 224 || a[3] > 144 || a[2]-a[0] < 32 || a[3]-a[1] < 8 {
			return nil, fmt.Errorf("過場背景錨點不符")
		}
	}
	s.art, e = checkedTerrainImage(filepath.Join(dir, s.File), s.Hash, 672, 432)
	if e != nil {
		return nil, e
	}
	s.patterns = map[string]*transitionPose{}
	for i := range s.Poses {
		p := &s.Poses[i]
		if p.Width < 1 || p.Width > 24 || p.Height < 1 || p.Height > 24 || len(p.Raster) != p.Width*p.Height {
			return nil, fmt.Errorf("過場角色圖樣不符")
		}
		if p.File != "" {
			if filepath.Base(p.File) != p.File {
				return nil, fmt.Errorf("過場角色圖檔不符")
			}
			p.art, e = checkedTerrainImage(filepath.Join(dir, p.File), p.Hash, p.Width*3, p.Height*3)
			if e != nil {
				return nil, e
			}
		}
		s.patterns[transitionPoseKey(p.Width, p.Height, p.Raster)] = p
	}
	return &s, nil
}

func (t *caveTransition) exactBlock(raw, rgb []byte, box image.Rectangle) bool {
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			v := t.Raster[y*224+x]
			i := (y+14)*320 + x + 48
			if raw[i] != v || !bytes.Equal(rgb[i*3:i*3+3], t.Colors[int(v)*3:int(v)*3+3]) {
				return false
			}
		}
	}
	return true
}

func (s *ManualArtwork) DrawCaveTransition(dst *image.RGBA, raw, rgb []byte, origin image.Point) bool {
	if s == nil || s.transition == nil || dst == nil || len(raw) != 64000 || len(rgb) != 192000 {
		return false
	}
	t := s.transition
	t.matched = false
	t.matchedTiles = 0
	t.originalTiles = 0
	if !image.Rect(144, 42, 816, 474).Add(origin).In(dst.Bounds()) {
		return false
	}
	// 兩側完整複雜背景是場景身份；黑色空格不能當身份依據。
	for _, a := range t.Anchors {
		if !t.exactBlock(raw, rgb, image.Rect(a[0], a[1], a[2], a[3])) {
			return false
		}
	}
	for y := 0; y < 144; y += 8 {
		for x := 0; x < 224; x += 8 {
			if !t.exactBlock(raw, rgb, image.Rect(x, y, x+8, y+8)) {
				t.originalTiles++
				continue
			}
			at := image.Pt((x+48)*3, (y+14)*3).Add(origin)
			draw.Draw(dst, image.Rectangle{Min: at, Max: at.Add(image.Pt(24, 24))}, t.art, image.Pt(x*3, y*3), draw.Src)
			t.matchedTiles++
		}
	}
	// 未識別前景維持當次原版；只有整個已觀察角色圖樣吻合才套人物HD。
	difference := image.Rectangle{}
	for y := 0; y < 144; y++ {
		for x := 0; x < 224; x++ {
			v := raw[(y+14)*320+x+48]
			if v != 0 && v != t.Raster[y*224+x] {
				p := image.Rect(x, y, x+1, y+1)
				if difference.Empty() {
					difference = p
				} else {
					difference = difference.Union(p)
				}
			}
		}
	}
	if !difference.Empty() && difference.Min.Y >= 96 && difference.Max.Y <= 120 && difference.Dx() <= 24 && difference.Dy() <= 24 {
		actor := make([]byte, 0, difference.Dx()*difference.Dy())
		valid := true
		for y := difference.Min.Y; y < difference.Max.Y; y++ {
			for x := difference.Min.X; x < difference.Max.X; x++ {
				i := (y+14)*320 + x + 48
				v := raw[i]
				if !bytes.Equal(rgb[i*3:i*3+3], t.Colors[int(v)*3:int(v)*3+3]) {
					valid = false
				}
				actor = append(actor, v)
			}
		}
		p := t.patterns[transitionPoseKey(difference.Dx(), difference.Dy(), actor)]
		if valid && p != nil && p.art != nil {
			at := image.Pt((difference.Min.X+48)*3, (difference.Min.Y+14)*3).Add(origin)
			draw.Draw(dst, image.Rectangle{Min: at, Max: at.Add(p.art.Bounds().Size())}, p.art, p.art.Bounds().Min, draw.Over)
		}
	}
	t.matched = true
	return true
}

func (s *ManualArtwork) TransitionState() map[string]any {
	if s == nil || s.transition == nil {
		return nil
	}
	return map[string]any{"cave_loaded": true, "cave_matched": s.transition.matched, "hd_background_tiles": s.transition.matchedTiles, "original_background_tiles": s.transition.originalTiles}
}

func (s *ManualArtwork) BeginFrame() {
	if s == nil {
		return
	}
	if s.transition != nil {
		s.transition.matched = false
	}
	if s.coherent != nil {
		s.coherent.matched = nil
	}
}

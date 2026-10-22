// 050後續八城鎮候選。完整地圖與地形來源匹配後，只覆繪原版底圖像素。
package hd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
)

type worldTown struct {
	width, townID    int
	columns, terrain []byte
	art, native      image.Image
}
type worldFigure struct {
	raw, mask []byte
	art       image.Image
}
type WorldTowns struct {
	towns                         []worldTown
	figures                       [5][8]worldFigure
	figureResource, figurePalette []byte
}

func NewWorldTowns(root, dir string) (*WorldTowns, error) { return newWorldTowns(root, dir, false) }

func NewWorldTownsLazy(root, dir string) (*WorldTowns, error) { return newWorldTowns(root, dir, true) }

func newWorldTowns(root, dir string, lazy bool) (*WorldTowns, error) {
	if _, err := checked(filepath.Join(root, "ZELRES2.SAR"), "bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1", 1<<20); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(dir, "world-towns.json"))
	if err != nil {
		return nil, err
	}
	var m struct {
		Towns []struct {
			Width            int
			TownID           int `json:"town_id"`
			Columns, Terrain []byte
			File             string
			Hash             string `json:"sha256"`
			NativeFile       string `json:"native_file"`
			NativeHash       string `json:"native_sha256"`
		}
	}
	if err = json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if len(m.Towns) != 8 {
		return nil, fmt.Errorf("後續城鎮圖組數量不符")
	}
	s := &WorldTowns{}
	for _, r := range m.Towns {
		if r.Width < 32 || r.Width > 1024 || len(r.Columns) != r.Width*8 || len(r.Terrain)%48 != 0 || len(r.Terrain) == 0 || len(r.Terrain) > 65536-0x8100 || filepath.Base(r.File) != r.File || filepath.Base(r.NativeFile) != r.NativeFile {
			return nil, fmt.Errorf("城鎮來源不符")
		}
		art, e := deferredTerrainImage(filepath.Join(dir, r.File), r.Hash, r.Width*24, 192, lazy)
		if e != nil {
			return nil, e
		}
		native, e := deferredTerrainImage(filepath.Join(dir, r.NativeFile), r.NativeHash, r.Width*8, 64, lazy)
		if e != nil {
			return nil, e
		}
		s.towns = append(s.towns, worldTown{r.Width, r.TownID, r.Columns, r.Terrain, art, native})
	}
	data, err = os.ReadFile(filepath.Join(dir, "manual-figures.json"))
	if err != nil {
		return nil, err
	}
	var figures struct {
		Resource, Palette []byte
		Figures           []struct {
			Group, Phase int
			Raster, Mask []byte
			File         string
			Hash         string `json:"sha256"`
		}
	}
	if err = json.Unmarshal(data, &figures); err != nil {
		return nil, err
	}
	if len(figures.Figures) != 40 || len(figures.Resource) != 256+164*48 || len(figures.Palette) != 768 {
		return nil, fmt.Errorf("居民候選來源不符")
	}
	s.figureResource, s.figurePalette = figures.Resource, figures.Palette
	for _, f := range figures.Figures {
		if f.Group < 0 || f.Group >= 5 || f.Phase < 0 || f.Phase >= 8 || len(f.Raster) != 384 || len(f.Mask) != 384 || filepath.Base(f.File) != f.File {
			return nil, fmt.Errorf("居民姿態不符")
		}
		im, e := deferredTerrainImage(filepath.Join(dir, f.File), f.Hash, 48, 72, lazy)
		if e != nil {
			return nil, e
		}
		s.figures[f.Group][f.Phase] = worldFigure{f.Raster, f.Mask, im}
	}
	return s, nil
}

func (s *WorldTowns) Draw(dst *image.RGBA, mode uint8, p, r, raw, rgb []byte, origin image.Point) bool {
	if s == nil || mode != 0x13 || len(p) != 65536 || len(r) != 65536 || len(raw) != 64000 || len(rgb) != 192000 {
		return false
	}
	camera, mp := townSceneWord(p, 0x80), townSceneWord(p, 0xff2a)
	if mp-camera*8 != 0xc017 || camera < 0 {
		return false
	}
	var town *worldTown
	for i := range s.towns {
		t := &s.towns[i]
		if townSceneWord(p, 0xc002) != t.width || int(p[0xc006]) != t.townID || camera+32 > t.width || !bytes.Equal(r[0x8100:0x8100+len(t.terrain)], t.terrain) {
			continue
		}
		match := true
		for n, v := range t.columns {
			if v != 253 && p[0xc017+n] != v {
				match = false
				break
			}
		}
		if match {
			town = t
			break
		}
	}
	if town == nil {
		return false
	}
	town.art, town.native = resolvedImage(town.art), resolvedImage(town.native)
	if town.art == nil || town.native == nil {
		return false
	}
	for y := 0; y < 64; y++ {
		for x := 0; x < 224; x++ {
			wx := (camera+4)*8 + x
			base := color.NRGBAModel.Convert(town.native.At(wx, y)).(color.NRGBA)
			i := (78+y)*320 + 48 + x
			if base.A == 0 || rgb[i*3] != base.R || rgb[i*3+1] != base.G || rgb[i*3+2] != base.B {
				continue
			}
			for yy := 0; yy < 3; yy++ {
				for xx := 0; xx < 3; xx++ {
					c := color.NRGBAModel.Convert(town.art.At(wx*3+xx, y*3+yy)).(color.NRGBA)
					if c.A == 0 {
						continue
					}
					px, py := (48+x)*3+xx+origin.X, (78+y)*3+yy+origin.Y
					old := dst.RGBAAt(px, py)
					a := uint32(c.A)
					inv := 255 - a
					dst.SetRGBA(px, py, color.RGBA{uint8((uint32(c.R)*a + uint32(old.R)*inv) / 255), uint8((uint32(c.G)*a + uint32(old.G)*inv) / 255), uint8((uint32(c.B)*a + uint32(old.B)*inv) / 255), 255})
				}
			}
		}
	}
	// 原版居民仍在最上層。只交換本次來源與畫面同時匹配的姿態。
	if p[0xe7] != 4 && bytes.Equal(r[0x4000:0x4000+len(s.figureResource)], s.figureResource) {
		for at, count := townSceneWord(p, 0xc00f), 0; at >= 0 && at+8 <= len(p) && count < 256; at, count = at+8, count+1 {
			column := townSceneWord(p, at)
			if column == 65535 {
				break
			}
			group, phase := int(p[at+2]&127), int(p[at+4]&3)
			if p[at+2]&128 == 0 {
				phase += 4
			}
			if group >= 5 {
				continue
			}
			f := s.figures[group][phase]
			x := (column-camera)*8 + 16
			if x < 48 || x+16 > 272 || f.art == nil {
				continue
			}
			match := true
			for y := 0; y < 24 && match; y++ {
				for xx := 0; xx < 16; xx++ {
					j := y*16 + xx
					if f.mask[j] == 0 {
						continue
					}
					i := (118+y)*320 + x + xx
					c := int(f.raw[j]) * 3
					if raw[i] != f.raw[j] || rgb[i*3] != s.figurePalette[c] || rgb[i*3+1] != s.figurePalette[c+1] || rgb[i*3+2] != s.figurePalette[c+2] {
						match = false
						break
					}
				}
			}
			if match {
				art := resolvedImage(f.art)
				if art == nil {
					continue
				}
				draw.Draw(dst, image.Rect(x*3, 354, x*3+48, 426).Add(origin), art, image.Point{}, draw.Over)
			}
		}
	}
	return true
}

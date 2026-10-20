// 魔法石圖樣來源與高清覆繪，契約見規格044；不處理原版相位或碰撞。
package hd

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"
)

type caveMagiaRecord struct {
	Slot, Phase, Life, Direction int
	Position, Column, Row        int
	X, Y                         int
	Visibility                   string
}

func caveMagiaSource(memory []byte) bool {
	if len(memory) != 1<<20 {
		return false
	}
	p, r := memory[0x1b10:0x11b10], memory[0x11b10:0x21b10]
	return caveWord(p, 0x3008) == 0x412f &&
		townBackgroundHash(p[0x8679:0x8790]) == "ffb5feb38b13196eca83be6e5d3a54cdda73b7568a973ead38c14992cb0b4c4f" &&
		townBackgroundHash(p[0x8790:0x87b0]) == "0401f4ed729be111cfefb40506d3616794fcd9305eceffbfd8e72bae7a2c7900" &&
		townBackgroundHash(p[0x412f:0x4192]) == "95adedfd5ad1db18f43d0c7c1f18ba3887cebdcc2e5630f400d9f0de3755766c" &&
		townBackgroundHash(r[0x8c00:0x9350]) == "7b4628728e0b6ce81a25b7bc0bbb3b4eedd640b0da102924b4384da0d38f690b"
}

func decodeCaveMagia(memory, body, base, original []byte) ([]caveMagiaRecord, []cavePixel, error) {
	if len(memory) != 1<<20 || len(body) != 1<<20 || len(base) != 64000 || len(original) != 64000 {
		return nil, nil, fmt.Errorf("magia source sizes")
	}
	p := memory[0x1b10:0x11b10]
	var records []caveMagiaRecord
	for slot := 0; slot < 4; slot++ {
		r := p[0xeb60+slot*7 : 0xeb60+(slot+1)*7]
		if r[0] == 255 {
			continue
		}
		position := caveWord(r, 3)
		if r[0] > 15 || r[2] == 0 || (r[1] != 1 && r[1] != 255) || position&0x8000 == 0 {
			return nil, nil, fmt.Errorf("magia unobserved record")
		}
		x, y := (position&0x7fff)*2%320, (position&0x7fff)*2/320
		if x < 48 || x > 264 || y < 14 || y > 150 || x != 48+(int(r[5])-4)*8 || y != 14+int(r[6])*8 {
			return nil, nil, fmt.Errorf("magia cached position")
		}
		records = append(records, caveMagiaRecord{Slot: slot, Phase: int(r[0]), Life: int(r[2]), Direction: int(int8(r[1])), Position: position, Column: int(r[5]), Row: int(r[6]), X: x, Y: y})
	}
	if len(records) == 0 {
		return nil, nil, nil
	}
	if !caveMagiaSource(memory) || !caveMagiaSource(body) {
		return nil, nil, fmt.Errorf("magia source identity")
	}
	tile, err := caveStatic(memory[0x11b10:0x21b10], 102)
	if err != nil {
		return nil, nil, err
	}
	var pixels []cavePixel
	seen := map[image.Point]bool{}
	for n := range records {
		r := &records[n]
		visible, cleared := true, true
		for at, color := range tile {
			if color != 0 {
				i := (r.Y+at/8)*320 + r.X + at%8
				visible = visible && original[i] == color
				cleared = cleared && original[i] == base[i]
			}
		}
		if visible == cleared {
			return nil, nil, fmt.Errorf("magia partial or ambiguous pixels")
		}
		if cleared {
			r.Visibility = "cleared"
			continue
		}
		r.Visibility = "visible"
		position := image.Pt(r.X, r.Y)
		if seen[position] {
			continue
		}
		seen[position] = true
		for at, color := range tile {
			if color != 0 {
				pixels = append(pixels, cavePixel{r.X + at%8, r.Y + at/8, color})
			}
		}
	}
	return records, pixels, nil
}

func (s *CaveTerrain) LoadMagia(dir string) error {
	if s == nil || len(s.material) != 432*768 {
		return fmt.Errorf("缺少洞窟呈現器")
	}
	im, err := checkedTerrainImage(filepath.Join(dir, "magia-shard-24x24.png"), "5fc434eca775c2078edf4d38aa07a92728c8d4755f7e06f10c2cccb70eb59a5e", 24, 24)
	if err != nil {
		return err
	}
	pixels := make([]color.NRGBA, 0, 576)
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			pixels = append(pixels, color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA))
		}
	}
	s.magia, s.magiaState = pixels, nil
	return nil
}

func (s *CaveTerrain) MagiaState() map[string]any {
	if s == nil || s.magiaState == nil {
		return nil
	}
	out := make(map[string]any, len(s.magiaState))
	for k, v := range s.magiaState {
		if records, ok := v.([]caveMagiaRecord); ok {
			v = append([]caveMagiaRecord{}, records...)
		}
		out[k] = v
	}
	return out
}

func (s *CaveTerrain) drawMagia(dst *image.RGBA, d *caveDecoded, origin image.Point) {
	visible := 0
	seen := map[image.Point]bool{}
	for _, r := range d.MagiaRecords {
		point := image.Pt(r.X, r.Y)
		if r.Visibility != "visible" || seen[point] {
			continue
		}
		seen[point] = true
		visible++
		for y := 0; y < 24; y++ {
			for x := 0; x < 24; x++ {
				c := s.magia[y*24+x]
				if c.A == 0 {
					continue
				}
				dst.SetRGBA(r.X*3+x+origin.X, r.Y*3+y+origin.Y, color.RGBA{c.R, c.G, c.B, 255})
			}
		}
	}
	if len(s.magia) != 0 {
		s.magiaState = map[string]any{"loaded": true, "matched": len(d.MagiaRecords) > 0, "empty": len(d.MagiaRecords) == 0, "visible_positions": visible, "records": append([]caveMagiaRecord{}, d.MagiaRecords...)}
	}
}

// 長老完整手勢的唯讀顯示；來源與限制見規格040。
package hd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"path/filepath"
)

const sageFramesManifestSHA = "41bd587485d0e54b48063bce43b26433c560b064cd9c339ba84f6d61df1e470c"

type sageFrameCandidate struct {
	pixels                 []byte
	pose, phase, animation int
}

// SageFrames不持有執行器、輸入或存檔，只讀已完成來源。
type SageFrames struct {
	art    [8][2]*image.RGBA
	frames []sageFrameCandidate
	state  map[string]any
}

func NewSageFrames(original, directory string) (*SageFrames, error) {
	if _, err := checked(filepath.Join(original, "ZELRES2.SAR"), "bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1", 1<<20); err != nil {
		return nil, err
	}
	data, err := checked(filepath.Join(directory, "manifest.json"), sageFramesManifestSHA, 1<<20)
	if err != nil {
		return nil, err
	}
	// 明列JSON欄位，來源manifest不能覆蓋圖組索引或路徑。
	var input struct {
		Assets []struct {
			Pose  int    `json:"pose"`
			Phase int    `json:"phase"`
			File  string `json:"file"`
			SHA   string `json:"sha256"`
		} `json:"assets"`
	}
	if err = json.Unmarshal(data, &input); err != nil {
		return nil, err
	}
	if len(input.Assets) != 16 {
		return nil, fmt.Errorf("長老手勢圖組須16張")
	}
	s := &SageFrames{}
	for _, row := range input.Assets {
		if row.Pose < 0 || row.Pose >= 8 || row.Phase < 0 || row.Phase >= 2 || row.File != fmt.Sprintf("sage-pose-%d-phase-%d.png", row.Pose, row.Phase) || s.art[row.Pose][row.Phase] != nil {
			return nil, fmt.Errorf("長老圖組索引或檔名不符")
		}
		s.art[row.Pose][row.Phase], err = castleImage(filepath.Join(directory, row.File), row.SHA, 288, 192)
		if err != nil {
			return nil, err
		}
	}
	s.Clear()
	return s, nil
}

func (s *SageFrames) Clear() {
	if s != nil {
		s.state = map[string]any{"loaded": true, "matched": false}
	}
}
func (s *SageFrames) State() map[string]any {
	if s == nil {
		return nil
	}
	state := map[string]any{}
	for k, v := range s.state {
		state[k] = v
	}
	return state
}

func sageFrameCandidates(p, r []byte) ([]sageFrameCandidate, bool) {
	base, ok := interiorRaster(p, r, 0xa9b6, 12, 8, 166)
	if !ok {
		return nil, false
	}
	var result []sageFrameCandidate
	seen := map[string]bool{}
	for pose := 0; pose < 8; pose++ {
		patch, ok := interiorRaster(p, r, 0xaa47+pose*32, 8, 4, 166)
		if !ok {
			return nil, false
		}
		frame := bytes.Clone(base)
		for y := 0; y < 32; y++ {
			copy(frame[(16+y)*96+16:(16+y)*96+80], patch[y*64:(y+1)*64])
		}
		for choice := -1; choice < 4; choice++ {
			pixels := bytes.Clone(frame)
			animation, phase := -1, 0
			if choice >= 0 {
				animation = int(p[0xabfb+choice])
				if animation >= 166 {
					return nil, false
				}
				tile := townSceneTile(r, 0x8000+animation*48)
				for y := 0; y < 8; y++ {
					copy(pixels[(24+y)*96+56:(24+y)*96+64], tile[y*8:(y+1)*8])
				}
				phase = choice & 1
			}
			key := string(pixels)
			if seen[key] {
				continue
			}
			seen[key] = true
			result = append(result, sageFrameCandidate{pixels, pose, phase, animation})
		}
	}
	return result, len(result) == 38
}

func (s *SageFrames) Draw(dst *image.RGBA, mode uint8, p, r, raw, rgb, palette []byte, origin image.Point) bool {
	if s == nil {
		return false
	}
	s.Clear()
	if dst == nil || mode != 0x13 || len(p) != 65536 || len(r) != 65536 || len(raw) != 64000 || len(rgb) != 192000 || len(palette) != 768 ||
		townBackgroundHash(p[0xa000:0xbb12]) != "8da1b499ace6d6c5dcfa82921a327dedff8381aec17fe79a190cfb711e9c1196" ||
		townBackgroundHash(r[0x8000:0x9f20]) != "98fa4a0855443baf6e44bd52ae24e393ce6e147d2fce47d9a75713e60797eae8" ||
		townBackgroundHash(palette) != "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c" {
		return false
	}
	x, y := int(p[0xbb13])*8, int(p[0xbb12])
	if (x != 56 && x != 112) || y != 23 {
		return false
	}
	if s.frames == nil {
		var ok bool
		s.frames, ok = sageFrameCandidates(p, r)
		if !ok {
			s.frames = nil
			return false
		}
	}
	native := image.Rect(x, y, x+96, y+64)
	selected := -1
	for i, frame := range s.frames {
		if interiorNativeMatch(raw, rgb, palette, frame.pixels, native) {
			if selected >= 0 {
				return false
			}
			selected = i
		}
	}
	if selected < 0 {
		return false
	}
	frame := s.frames[selected]
	art := s.art[frame.pose][frame.phase]
	if art == nil || art.Bounds() != image.Rect(0, 0, 288, 192) {
		return false
	}
	box := image.Rectangle{Min: native.Inset(1).Min.Mul(3), Max: native.Inset(1).Max.Mul(3)}.Add(origin)
	if !box.In(dst.Bounds()) {
		return false
	}
	draw.Draw(dst, box, art, image.Pt(3, 3), draw.Src)
	s.state = map[string]any{"loaded": true, "matched": true, "pose": frame.pose, "phase": frame.phase, "animation_tile": frame.animation, "native_origin": []int{x, y}, "scope": "完整來源匹配，正常手勢0..7已另核；置中手勢正常路徑待驗"}
	return true
}

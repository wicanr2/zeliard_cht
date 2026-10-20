// 武器店八姿態與局部動畫的唯讀顯示；來源及限制見規格041。
package hd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"path/filepath"
)

const merchantFramesManifestSHA = "3b5f4da6b39ab356c06e8512d8d0c9023fecc5604db6f9a3b4ab85512aa3f064"

type merchantFrameCandidate struct {
	key    string
	pose   int
	pixels []byte
}

// MerchantFrames不持有執行器或輸入，只讀當次完整來源。
type MerchantFrames struct {
	art    map[string]*image.RGBA
	frames []merchantFrameCandidate
	state  map[string]any
}

func NewMerchantFrames(original, directory string) (*MerchantFrames, error) {
	if _, e := checked(filepath.Join(original, "ZELRES2.SAR"), "bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1", 1<<20); e != nil {
		return nil, e
	}
	data, e := checked(filepath.Join(directory, "manifest.json"), merchantFramesManifestSHA, 1<<20)
	if e != nil {
		return nil, e
	}
	var manifest struct {
		Assets []struct {
			Key  string `json:"key"`
			File string `json:"file"`
			SHA  string `json:"sha256"`
		} `json:"assets"`
	}
	if e = json.Unmarshal(data, &manifest); e != nil {
		return nil, e
	}
	if len(manifest.Assets) != 10 {
		return nil, fmt.Errorf("武器店圖組須10張")
	}
	allowed := map[string]bool{}
	for pose := 0; pose < 8; pose++ {
		allowed[fmt.Sprintf("body-%d", pose)] = true
	}
	allowed["body-0-animation-3"] = true
	allowed["body-2-animation-3"] = true
	s := &MerchantFrames{art: map[string]*image.RGBA{}}
	for _, row := range manifest.Assets {
		if !allowed[row.Key] || row.File != row.Key+"-hd.png" || s.art[row.Key] != nil {
			return nil, fmt.Errorf("武器店圖組索引不符")
		}
		s.art[row.Key], e = castleImage(filepath.Join(directory, row.File), row.SHA, 282, 186)
		if e != nil {
			return nil, e
		}
	}
	s.Clear()
	return s, nil
}

func (s *MerchantFrames) Clear() {
	if s != nil {
		s.state = map[string]any{"loaded": true, "matched": false}
	}
}
func (s *MerchantFrames) State() map[string]any {
	if s == nil {
		return nil
	}
	out := map[string]any{}
	for k, v := range s.state {
		out[k] = v
	}
	return out
}

func merchantBodyRaster(p, r []byte, pose int) ([]byte, bool) {
	if pose < 0 || pose > 7 {
		return nil, false
	}
	pixels := make([]byte, 0, 6144)
	rows := 0
	for group := 0; group < 2; group++ {
		at := 0xaa10 + pose*6 + group*3
		h := int(p[at])
		if h == 0 {
			break
		}
		if rows+h > 8 {
			return nil, false
		}
		table := int(p[at+1]) | int(p[at+2])<<8
		patch, ok := interiorRaster(p, r, table, 12, h, 256)
		if !ok {
			return nil, false
		}
		pixels = append(pixels, patch...)
		rows += h
	}
	return pixels, rows == 8 && len(pixels) == 6144
}

func merchantFrameCandidates(p, r []byte) ([]merchantFrameCandidate, bool) {
	var frames []merchantFrameCandidate
	seen := map[string]bool{}
	for pose := 0; pose < 8; pose++ {
		base, ok := merchantBodyRaster(p, r, pose)
		if !ok {
			return nil, false
		}
		for phase := -1; phase < 4; phase++ {
			if phase >= 0 && pose != 0 && pose != 2 {
				break
			}
			pixels := bytes.Clone(base)
			key := fmt.Sprintf("body-%d", pose)
			if phase >= 0 {
				table, w, h, x, y := 0xaad0, 2, 1, 72, 56
				if pose == 2 {
					table, w, h, x, y = 0xab68, 1, 2, 32, 32
				}
				patch, valid := interiorRaster(p, r, table+phase*2, w, h, 256)
				if !valid {
					return nil, false
				}
				for row := 0; row < h*8; row++ {
					copy(pixels[(y+row)*96+x:(y+row)*96+x+w*8], patch[row*w*8:(row+1)*w*8])
				}
				key = fmt.Sprintf("body-%d-animation-%d", pose, phase)
			}
			identity := string(pixels)
			if seen[identity] {
				continue
			}
			seen[identity] = true
			frames = append(frames, merchantFrameCandidate{key, pose, pixels})
		}
	}
	return frames, len(frames) == 10
}

func (s *MerchantFrames) Draw(dst *image.RGBA, mode uint8, p, r, raw, rgb, palette []byte, origin image.Point) bool {
	if s == nil {
		return false
	}
	s.Clear()
	if dst == nil || mode != 0x13 || len(p) != 65536 || len(r) != 65536 || len(raw) != 64000 || len(rgb) != 192000 || len(palette) != 768 ||
		townBackgroundHash(p[0xa000:0xbbfd]) != "9d35fd3f1f35b7b02af8af17b2cac6a0098c5d3ab7055a677686354ddac041b4" ||
		townBackgroundHash(r[0x8000:0xb000]) != "b908b4b98dece29dfeba950cc8b996dcf89412bfa5a3286254d8ea5a5ca3a6ed" ||
		townBackgroundHash(palette) != "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c" {
		return false
	}
	if s.frames == nil {
		var ok bool
		s.frames, ok = merchantFrameCandidates(p, r)
		if !ok {
			s.frames = nil
			return false
		}
	}
	selected := -1
	for i, frame := range s.frames {
		if interiorNativeMatch(raw, rgb, palette, frame.pixels, merchantBox) {
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
	art := s.art[frame.key]
	if art == nil || art.Bounds() != image.Rect(0, 0, 282, 186) {
		return false
	}
	native := merchantBox.Inset(1)
	box := image.Rectangle{Min: native.Min.Mul(3), Max: native.Max.Mul(3)}.Add(origin)
	if !box.In(dst.Bounds()) {
		return false
	}
	draw.Draw(dst, box, art, image.Point{}, draw.Src)
	s.state = map[string]any{"loaded": true, "matched": true, "key": frame.key, "pose": frame.pose, "scope": "八姿態及兩局部動畫；未引用圖格另列用途限制"}
	return true
}

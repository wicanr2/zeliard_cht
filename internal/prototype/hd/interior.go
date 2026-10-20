// 小屋石像與旅館的唯讀覆繪。來源、別名及正常路徑限制見規格039。
package hd

import (
	"fmt"
	"image"
	"image/draw"
	"path/filepath"
)

type interiorPose struct {
	key                    string
	art, table, cols, rows int
	x, y                   int
}

type interiorProfile struct {
	key                        string
	moduleBytes, graphicsBytes int
	moduleSHA, graphicsSHA     string
	box                        image.Rectangle
	table, cols, rows, tiles   int
	inset                      int
	poses                      []interiorPose
}

var interiorProfiles = []interiorProfile{
	{"hut", 595, 6480, "738f6ff33e36685ae0908605d4db50e4f5df7486e098401b0bc5e033972aee37", "503a953169c4045db3ddc4a9372ee994712bf04842bc44321d7b7ac5f057d350",
		image.Rect(96, 30, 232, 158), 0xa129, 17, 16, 135, 0, []interiorPose{{key: "base", art: 0}}},
	{"inn", 0x505, 5376, "f648352ad1d30d886526a964343d0cfe5f9380be8c391854b327f4aa4cc27ef3", "be13146a72b5689cec882347fe3dbadd4fa84f56832320c0408db8f234bd5a37",
		image.Rect(56, 23, 152, 87), 0xa1cf, 12, 8, 112, 1, []interiorPose{
			{key: "base-aliases", art: 1},
			{"talk", 2, 0xa27d, 2, 2, 8, 16},
			{"hand-near", 3, 0xa2a9, 5, 4, 8, 16},
			{"hand-open", 4, 0xa2bd, 5, 4, 8, 16},
		}},
}

// Interiors不持有執行器、輸入或存檔，只讀當次來源與五張已核圖。
type Interiors struct {
	art   [5]*image.RGBA
	state map[string]any
}

func NewInteriors(original, directory string) (*Interiors, error) {
	if _, err := checked(filepath.Join(original, "ZELRES2.SAR"), "bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1", 1<<20); err != nil {
		return nil, err
	}
	s := &Interiors{}
	for i, asset := range []struct {
		file, sha string
		w, h      int
	}{
		{"hut-hd-v1.png", "0397f04f4cd7ba958c7c67d12c7eddf169b1797e332a1dc5e78283319a83d636", 408, 384},
		{"inn-base-hd-v1.png", "c2c69d6f109070213c2c1287f07cccf5a7873dcf728b28c74c19457224575931", 288, 192},
		{"inn-ambient-1-hd-v1.png", "6d83769b4d1c68dac5be8c5562cbc73a1df546cf0c2c544d475b394b97ee0163", 288, 192},
		{"inn-rest-2-hd-v1.png", "99638a3f14bede7df6f96a58b8f6c4cfa2c0444ee6d6d46a8eac7efb0ae853e5", 288, 192},
		{"inn-rest-3-hd-v1.png", "d6afe32e2b72e01c51d967f9a16ec16dec5f921301cd7e92eff06ef1fc2e6e1f", 288, 192},
	} {
		var err error
		s.art[i], err = castleImage(filepath.Join(directory, asset.file), asset.sha, asset.w, asset.h)
		if err != nil {
			return nil, fmt.Errorf("室內HD圖組拒絕：%w", err)
		}
	}
	s.Clear()
	return s, nil
}

func (s *Interiors) Clear() {
	if s != nil {
		s.state = map[string]any{"loaded": true, "matched": false}
	}
}

func (s *Interiors) State() map[string]any {
	if s == nil {
		return nil
	}
	state := make(map[string]any, len(s.state))
	for key, value := range s.state {
		state[key] = value
	}
	return state
}

func interiorRaster(p, r []byte, table, cols, rows, tiles int) ([]byte, bool) {
	if table < 0 || cols <= 0 || rows <= 0 || table+cols*rows > len(p) || len(r) != 65536 {
		return nil, false
	}
	width := cols * 8
	pixels := make([]byte, width*rows*8)
	for n, id := range p[table : table+cols*rows] {
		if int(id) >= tiles {
			return nil, false
		}
		data := townSceneTile(r, 0x8000+int(id)*48)
		for y := 0; y < 8; y++ {
			at := (n/cols*8+y)*width + n%cols*8
			copy(pixels[at:at+8], data[y*8:y*8+8])
		}
	}
	return pixels, true
}

func interiorNativeMatch(raw, rgb, palette, pixels []byte, box image.Rectangle) bool {
	if len(pixels) != box.Dx()*box.Dy() {
		return false
	}
	for y := 0; y < box.Dy(); y++ {
		for x := 0; x < box.Dx(); x++ {
			v := int(pixels[y*box.Dx()+x])
			at := (y+box.Min.Y)*320 + x + box.Min.X
			if raw[at] != byte(v) || rgb[at*3] != palette[v*3] || rgb[at*3+1] != palette[v*3+1] || rgb[at*3+2] != palette[v*3+2] {
				return false
			}
		}
	}
	return true
}

func (s *Interiors) Draw(dst *image.RGBA, mode uint8, p, r, raw, rgb, palette []byte, origin image.Point) bool {
	if s == nil {
		return false
	}
	s.Clear()
	if dst == nil || mode != 0x13 || len(p) != 65536 || len(r) != 65536 || len(raw) != 64000 || len(rgb) != 192000 || len(palette) != 768 ||
		townBackgroundHash(palette) != "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c" {
		return false
	}
	for _, profile := range interiorProfiles {
		if townBackgroundHash(p[0xa000:0xa000+profile.moduleBytes]) != profile.moduleSHA || townBackgroundHash(r[0x8000:0x8000+profile.graphicsBytes]) != profile.graphicsSHA {
			continue
		}
		base, ok := interiorRaster(p, r, profile.table, profile.cols, profile.rows, profile.tiles)
		if !ok {
			return false
		}
		selected := -1
		for n, pose := range profile.poses {
			pixels := base
			if pose.table != 0 {
				patch, valid := interiorRaster(p, r, pose.table, pose.cols, pose.rows, profile.tiles)
				if !valid {
					return false
				}
				pixels = cloneTownBytes(base)
				for y := 0; y < pose.rows*8; y++ {
					at := (pose.y+y)*profile.box.Dx() + pose.x
					copy(pixels[at:at+pose.cols*8], patch[y*pose.cols*8:(y+1)*pose.cols*8])
				}
			}
			if interiorNativeMatch(raw, rgb, palette, pixels, profile.box) {
				if selected != -1 {
					return false
				}
				selected = n
			}
		}
		if selected < 0 {
			return false
		}
		pose := profile.poses[selected]
		art := s.art[pose.art]
		if art == nil || art.Bounds() != image.Rect(0, 0, profile.box.Dx()*3, profile.box.Dy()*3) {
			return false
		}
		native := profile.box.Inset(profile.inset)
		box := image.Rectangle{Min: native.Min.Mul(3), Max: native.Max.Mul(3)}.Add(origin)
		if !box.In(dst.Bounds()) {
			return false
		}
		draw.Draw(dst, box, art, native.Min.Sub(profile.box.Min).Mul(3), draw.Src)
		s.state = map[string]any{"loaded": true, "matched": true, "profile": profile.key, "frame": pose.key,
			"native_pixels": profile.box.Dx() * profile.box.Dy(), "scope": "完整原版來源及像素匹配；正常到達另驗"}
		return true
	}
	return false
}

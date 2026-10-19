// 洞窟武器十個已驗色號／透明來源。規格018保留原版位置與最後圖層。
package hd

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"
)

type caveWeaponRegistration struct {
	key          string
	crop, native image.Rectangle
	pixelsSHA    string
}

var caveWeaponRegistrations = []caveWeaponRegistration{
	{"5ec9fb65b9d8384c2655389c389d4cd0c02461b6d7dec8d55a602fd02c824194", image.Rect(66, 111, 236, 361), image.Rect(12, 8, 23, 24), "fc08c5bca8fb0c62fbce3f54418f6dba8ea1d72e6d2ddbf8de237bfc2ae7e231"},
	{"e2fc1aa201ceb108463d92b1021320824926de8c1b3dc0c64adde629e6f560b7", image.Rect(364, 143, 591, 351), image.Rect(9, 12, 25, 26), "93bd0dcbc39d5ef346fac06661a33bbf549deb68f6c0918ffaa006377370def1"},
	{"58c3e6597729ef0e0cfc4aece52bf0abb4b85b3d483a61f4397029f62038c596", image.Rect(667, 299, 917, 356), image.Rect(8, 24, 29, 28), "6b6fdfbe11a1df9c87d8f0c5b7bee35906df43f576b8b552a7353d274269544b"},
	{"7fb7c945a3439490c87d950f38b9686cdde254e8cd36c06f9fbc854d1ebc5f43", image.Rect(961, 301, 1216, 366), image.Rect(1, 24, 20, 29), "6267041d01eeb20b3a2c284b9dc19c4fbe5f8143d1d0c6cab3158f361c8c32e8"},
	{"1185bbea77bc10a9b7d69d47d16bd4c042ade8e6ba845b9cf3d5dac3102d5bce", image.Rect(39, 713, 245, 751), image.Rect(5, 24, 20, 26), "d7cc81705632cfd174caf8a0c1ef8a7798990dc652d7db6d8e5f89c41ea0fdf2"},
	{"e3e90e5a32eb3b709a411966432870f26c5ba2b3c308af037d0ea2ae1a110535", image.Rect(384, 510, 558, 761), image.Rect(9, 8, 20, 24), "8a3f661dec3b35c27ee2b97a737e5e59c1687e219f22ded72feaee76cb626509"},
	{"b373699bab92c142ac4830bd41f762e9f99dfc9b83b664435b9a0c718a8c3115", image.Rect(664, 562, 891, 751), image.Rect(8, 12, 23, 26), "28cd2b16a038e413e620b8b4fe48fa8fa7e2df2636fceff029fa20b5232343c4"},
	{"f9564895676cce35fea146a6c0cf156abee10d14a98c5a6ce29dfd540458909a", image.Rect(989, 707, 1216, 761), image.Rect(3, 24, 24, 28), "db946cbcbb3bba51fb5bb185229806abac5d139565a149c2854dbd539e8961d3"},
	{"895aa5525b602a52b6c74fc0d65c084830de7e50c940610d8519be32ff6058f3", image.Rect(37, 1058, 299, 1125), image.Rect(12, 24, 31, 29), "843dcf27d8cb1a8d725c0445170973261efa0248c86117e48a8adc0aaa3c98bf"},
	{"b1079475877c2f451758113a197ccda8c4fa9d38341152c08f2882e341385108", image.Rect(367, 1062, 599, 1099), image.Rect(12, 24, 27, 26), "afbcd94cd0b5e2b59d7a414c98bf36df3c941261fb356cd61eb603d5b578bb5c"},
}

type caveWeaponPose struct {
	number int
	bounds image.Rectangle
	pixels []color.NRGBA
}
type CaveWeapon struct{ poses map[string]*caveWeaponPose }

func NewCaveWeapon(dir string) (*CaveWeapon, error) {
	atlas, e := caveArtwork(filepath.Join(dir, "weapon-atlas-handpaint-v1.png"), "8ea1753a36e44f51dd9ba4732613095479a4ceabddd96c43a9336b1bc7b68e66")
	if e != nil {
		return nil, e
	}
	weapon := &CaveWeapon{poses: map[string]*caveWeaponPose{}}
	for n, r := range caveWeaponRegistrations {
		if r.crop.Empty() || !r.crop.In(atlas.Bounds()) || r.native.Empty() || !r.native.In(image.Rect(0, 0, 32, 32)) || weapon.poses[r.key] != nil {
			return nil, fmt.Errorf("武器美術登錄不符")
		}
		box := image.Rectangle{Min: r.native.Min.Mul(3), Max: r.native.Max.Mul(3)}
		pose := &caveWeaponPose{number: n, bounds: box, pixels: make([]color.NRGBA, 96*96)}
		crop := cavePlayerCrop{Image: atlas, box: r.crop}
		for y := 0; y < box.Dy(); y++ {
			for x := 0; x < box.Dx(); x++ {
				pose.pixels[(box.Min.Y+y)*96+box.Min.X+x] = caveArea(crop, x, y, box.Dx(), box.Dy())
			}
		}
		raw := make([]byte, 0, 96*96*4)
		for _, c := range pose.pixels {
			raw = append(raw, c.R, c.G, c.B, c.A)
		}
		if townBackgroundHash(raw) != r.pixelsSHA {
			return nil, fmt.Errorf("武器直接高清取樣不符：%d", n)
		}
		weapon.poses[r.key] = pose
	}
	return weapon, nil
}

func (w *CaveWeapon) selectPose(d *caveDecoded) (string, *caveWeaponPose, bool) {
	if w == nil || d == nil {
		return "", nil, false
	}
	key := townBackgroundHash(d.WeaponRaster[:])
	if len(d.Weapon) == 0 {
		return key, nil, true
	}
	p := w.poses[key]
	if p == nil || len(p.pixels) != 96*96 || p.bounds.Empty() || !p.bounds.In(image.Rect(0, 0, 96, 96)) || !p.bounds.Add(image.Pt(d.WeaponX*3, d.WeaponY*3)).In(image.Rect(144, 42, 816, 474)) {
		return key, nil, false
	}
	return key, p, false
}

func (p *caveWeaponPose) draw(dst *image.RGBA, d *caveDecoded, origin image.Point) {
	for y := p.bounds.Min.Y; y < p.bounds.Max.Y; y++ {
		for x := p.bounds.Min.X; x < p.bounds.Max.X; x++ {
			c := p.pixels[y*96+x]
			if c.A == 0 {
				continue
			}
			px, py := d.WeaponX*3+x+origin.X, d.WeaponY*3+y+origin.Y
			old := dst.RGBAAt(px, py)
			a, inv := uint32(c.A), uint32(255-c.A)
			dst.SetRGBA(px, py, color.RGBA{uint8((uint32(c.R)*a + uint32(old.R)*inv + 127) / 255), uint8((uint32(c.G)*a + uint32(old.G)*inv + 127) / 255), uint8((uint32(c.B)*a + uint32(old.B)*inv + 127) / 255), 255})
		}
	}
}

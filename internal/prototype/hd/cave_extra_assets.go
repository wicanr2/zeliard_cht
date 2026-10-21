// 已存刀光、物件及九個精確遮罩對應，來源與下層契約見規格049。
package hd

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"path/filepath"
)

type caveExtraAsset struct {
	key, hash string
	bounds    image.Rectangle
}

type caveMaskedObject struct {
	key, target        string
	backgrounds, tiles [4]byte
}

func extraAssetPixels(dir, kind string, size int, r caveExtraAsset) ([]color.NRGBA, error) {
	im, err := checkedTerrainImage(filepath.Join(dir, fmt.Sprintf("%s-%s-%dx%d.png", kind, r.key[:12], size, size)), r.hash, size, size)
	if err != nil {
		return nil, err
	}
	pixels := make([]color.NRGBA, size*size)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			pixels[y*size+x] = color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
		}
	}
	return pixels, nil
}

func (w *CaveWeapon) LoadExtra(dir string) error {
	if w == nil || len(w.poses) < len(caveWeaponRegistrations) {
		return fmt.Errorf("須先載入洞窟武器圖組")
	}
	poses := make(map[string]*caveWeaponPose, len(w.poses)+len(caveExtraWeapons))
	for key, pose := range w.poses {
		poses[key] = pose
	}
	for n, r := range caveExtraWeapons {
		pixels, err := extraAssetPixels(dir, "weapon", 96, r)
		if err != nil {
			return err
		}
		poses[r.key] = &caveWeaponPose{number: len(caveWeaponRegistrations) + len(caveCyanWeapons) + n, bounds: image.Rectangle{Min: r.bounds.Min.Mul(3), Max: r.bounds.Max.Mul(3)}, pixels: pixels}
	}
	w.poses = poses
	return nil
}

func (s *CaveObjects) LoadExtra(dir string) error {
	if s == nil || len(s.poses) < len(caveObjectRegistrations) {
		return fmt.Errorf("須先載入洞窟物件圖組")
	}
	poses := make(map[string]*caveObjectPose, len(s.poses)+len(caveExtraObjects))
	for key, pose := range s.poses {
		poses[key] = pose
	}
	for n, r := range caveExtraObjects {
		pixels, err := extraAssetPixels(dir, "object", 48, r)
		if err != nil {
			return err
		}
		poses[r.key] = &caveObjectPose{number: len(caveObjectRegistrations) + n, bounds: image.Rectangle{Min: r.bounds.Min.Mul(3), Max: r.bounds.Max.Mul(3)}, pixels: pixels}
	}
	aliases := make(map[string]caveMaskedObject, len(caveMaskedObjects))
	for _, r := range caveMaskedObjects {
		if poses[r.target] == nil {
			return fmt.Errorf("遮罩目標圖面缺失")
		}
		aliases[r.key] = r
	}
	s.poses, s.masked = poses, aliases
	return nil
}

func (s *CaveObjects) maskedEnabled() bool {
	return s != nil && (len(s.masked) == len(caveMaskedObjects) || len(s.manual) != 0)
}

// 每格先恢復原版列舉端的背景。0仍是黑色，非零取沒有物件的HD下層。
func caveObjectRestore(dst, base *image.RGBA, object caveObject, palette []byte, origin image.Point) {
	viewport := image.Rect(144, 42, 816, 474).Add(origin)
	for q, bg := range object.Backgrounds {
		at := image.Pt(object.X*3+(q%2)*24, object.Y*3+(q/2)*24).Add(origin)
		box := image.Rectangle{Min: at, Max: at.Add(image.Pt(24, 24))}.Intersect(viewport)
		if bg == 0 {
			draw.Draw(dst, box, image.NewUniform(color.RGBA{palette[0], palette[1], palette[2], 255}), image.Point{}, draw.Src)
		} else {
			draw.Draw(dst, box, base, box.Min, draw.Src)
		}
	}
}

func (s *CaveObjects) drawWithTerrain(dst *image.RGBA, d *caveDecoded, palette []byte, origin image.Point, base *image.RGBA) map[string]any {
	trusted := s.maskedEnabled() && base != nil && image.Rect(144, 42, 816, 474).Add(origin).In(base.Bounds())
	known, unknown, clipped, masked := 0, 0, 0, 0
	keys, unknownKeys := []string{}, []string{}
	for _, object := range d.Objects {
		key, pose := s.selectPose(object)
		keys = append(keys, key)
		if !image.Rect(object.X, object.Y, object.X+16, object.Y+16).In(image.Rect(48, 14, 272, 158)) {
			clipped++
		}
		if object.Backgrounds != [4]byte{} && !trusted {
			pose = nil
		}
		if trusted {
			caveObjectRestore(dst, base, object, palette, origin)
		}
		if pose == nil {
			unknown++
			unknownKeys = append(unknownKeys, key)
			caveObjectNative(dst, object, palette, origin, false)
		} else {
			known++
			if object.Backgrounds != [4]byte{} {
				masked++
			}
			if !trusted {
				caveObjectNative(dst, object, palette, origin, true)
			}
			pose.draw(dst, object, origin)
		}
	}
	return map[string]any{"loaded": true, "matched": known > 0, "known": known, "unknown": unknown, "empty": len(d.Objects) == 0, "clipped": clipped, "source_keys": keys, "unknown_keys": unknownKeys, "masked_known": masked, "terrain_underlay": trusted}
}

var caveExtraWeapons = []caveExtraAsset{
	{"2b385c0118b0d5c6bb6a3f108f0f6ab117f68e4b9cd1b41becc7ee60110ba99f", "2d339eed9a08cdcceedc9deedd21150d593133909bf557eb91a4c5a1e165972b", image.Rect(7, 18, 23, 24)},
	{"c54d531c4d90984aaedc844007a266ca04b87e8b83ad480ac8ff398ced66ae7f", "06bc26a61815857f3529e243a512c45ee4af0acf944b63febbab579e2f01bf05", image.Rect(6, 8, 23, 23)},
	{"b7dc34f56ac88448cc13acc7658fb9b6baada6b6e9656028741bc4ce8e02de22", "d7b7b05edbf513245f41eaf7fcd35c59111bb0f03d506061139fc66b61fad148", image.Rect(10, 9, 24, 28)},
	{"605aefcb7189d758ff3f81c620e66fbc1ff972d79b5de3d8d5e06e118591febb", "d741a097d73d8cd85f3e3ac7b9020ff7cbe319f66c667c119bd1608e8a220851", image.Rect(7, 9, 24, 18)},
}

var caveExtraObjects = []caveExtraAsset{
	{"628c5e9f9612a4fe7ea1169097bfc45a05b7b9c83717dfea645ac071f34a6594", "4cc8eb088b32a0feddbff9942b9d3d023e209b3465cee98cefd367f98edb4d0c", image.Rect(0, 0, 16, 16)},
	{"ce339bed11273c7c68e81b5780604c7b2e40e6b475cdd0231060d866660baa32", "526db4488777b06fc522cd1f2702ab305a97fd874d2b49e2224b7bbb37ad90f2", image.Rect(0, 0, 16, 16)},
	{"32a04f9bfb31e9769c1628129bfecb346f5a65311381bba73ec7a42919a8f506", "d8973b0859b2ed04702f89aabf8f258d63f3ed9a27ef8e0edb38297f8e055e7b", image.Rect(0, 2, 16, 16)},
	{"ccd697fd1eee9f9255481b95c0b4939c3bde6b898b3b1a82fcf5f03ad4d86e46", "d2f680612e023a1870a83b2185aa047fd38313d2ae30f493b4da1b76658fb295", image.Rect(0, 0, 16, 16)},
}

var caveMaskedObjects = []caveMaskedObject{
	{"91a56be86f31d6e79a230b376384ef880cdfd1c7860e5c834e12f3f6a8b3b2eb", "dcaa4740634ab751c1709475b4fcd761226f3d220bdb50e859900084d63f92e6", [4]byte{2, 0, 2, 0}, [4]byte{126, 123, 127, 128}},
	{"f768eb39eb5a68b66fef8cc4ac0072e7671dca688d061dace9e0a5d0ce35543b", "5eba8caa3e008cc886f70d7c552006afaaba7ec18be48d4f196f8a4375fca107", [4]byte{0, 2, 0, 2}, [4]byte{129, 130, 131, 132}},
	{"6ca427c4a94ef301a68ee64a487705e0431fa10a568d61702f5f02736ebfbb8d", "02c9e83d897bf1e590354616bd3727b1f184fa52d1e7d16eff39725d89cd0764", [4]byte{0, 2, 0, 2}, [4]byte{189, 190, 191, 192}},
	{"1eaa69a28c4c5082e305716c2fcda0bebd1519554b61f5b5341d044cb10a111b", "e55b49761767127ec263f07e6507b176f508aaf80c96afdc7fa4761c747469b9", [4]byte{0, 2, 0, 2}, [4]byte{193, 194, 195, 196}},
	{"214d91ce4c2bbda61275ca69361123b1cb7665ff239ceec5b3206d2d98742e86", "4c84e73e4189aa7b80ee75f6b9257fcc45a91886db2335c341fd54a46b13f3fe", [4]byte{0, 2, 0, 2}, [4]byte{0, 0, 199, 200}},
	{"6ba172576991442cd35f63e3822aed36f171f100300d07b3678f3d5542ad8818", "c82559904139aeed2e33844422c448f5be97f60e821d45f7875e3bce66ea474c", [4]byte{0, 2, 0, 2}, [4]byte{13, 14, 15, 16}},
	{"9b51aaaa22776c1b2c335ed0d0a95920f0bf4b7bbeea9db32ebf03ea508b8ef0", "53db8ad10c69de9a50e9c07e015e8770f4801ea9ad35fac31582a1ebb6bbc145", [4]byte{0, 2, 0, 0}, [4]byte{17, 18, 19, 20}},
	{"5994da5d0ce7d484e81ae069010efd14c6e5c905dbf3e860e8e86b952a186eb2", "ae50aace23626db209bd43322a9d071155839a6d1593427cc8be05bd0f7c94b5", [4]byte{0, 2, 0, 0}, [4]byte{153, 154, 155, 156}},
	{"e7c0f0499c62fddf4ec6b1cc4b43735d242f15582892422b8b4554d74934496d", "9b9c5261dc4449cfe02799784bcb3deaa154df55c6e3c241c118c969d46d3637", [4]byte{2, 0, 0, 0}, [4]byte{157, 158, 159, 160}},
}

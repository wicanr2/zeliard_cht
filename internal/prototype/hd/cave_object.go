// 洞窟動態物件的固定圖組與原版順序。來源、清黑及裁切契約見規格019。
package hd

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"path/filepath"
)

type caveObjectAtlas struct {
	name, sha     string
	width, height int
}

var caveObjectAtlases = []caveObjectAtlas{
	{"object-red-handpaint-v1.png", "a515202a34cb7bd2f42f95d8caf19ac3b13a1aa0e7208b37fb9cd709990af55f", 1774, 887},
	{"object-green-handpaint-v1.png", "314991d63cb3e6f04cace19929919183f497982660a638c464015d8d91367172", 1448, 1086},
	{"object-effects-handpaint-v1.png", "aec83e6729ab077e23610d972a8790198b4837814b0986f99f2b98a0cd65d4f6", 1774, 887},
	{"object-green-extra-handpaint-v1.png", "0b0661cc5ceff38bb9de4dd5b5e2952565df959f03403a1897d525d989556b30", 1463, 1075},
}

type caveObjectRegistration struct {
	key          string
	atlas        int
	crop, native image.Rectangle
	pixelsSHA    string
}

var caveObjectRegistrations = []caveObjectRegistration{
	{"cb16b1668b68ff263aa496670dce7508bfd3e09a33eab6740069f072f9e28d02", 0, image.Rect(48, 49, 425, 412), image.Rect(1, 1, 16, 16), "eff9afe6cba99b184f2c42307c5ddb4f950381220208e4ba8ecd4ecbb76f4817"},
	{"bd1204b2d4a87d136b4d2dba93bb9dbc22755df1f2ba62d6787934842e54b14c", 0, image.Rect(477, 62, 880, 412), image.Rect(0, 2, 15, 16), "06741f0aaa4468ecf0bac313eefc48c9c0417002433d19d821d65fc75d5ddca0"},
	{"d0f83d881a930d820ad4b1fcc60e5756ad2fc57dc84146219c5e1b808a2fb330", 1, image.Rect(23, 59, 346, 344), image.Rect(0, 2, 16, 16), "28cf913d164466202613ccc3e86bdff79244c3013d002ee2bd231fa1498e97c9"},
	{"300735537bc84de977df7be5e7a878365ad308356f7bcf597210909e7d9d6aec", 0, image.Rect(941, 75, 1322, 413), image.Rect(1, 2, 16, 16), "0c1910d284da62d963051ab696294e65368b5d53e93fecd08f52c997aef08813"},
	{"cd718211208e9f9873de5010aef9aff1f7ccfca8de397e6365adee5210cbafea", 1, image.Rect(381, 84, 702, 342), image.Rect(0, 3, 16, 16), "62e35f2ef36da77167a90f4a58693bed22bd5b765ebf2b9af98da78396ef4e81"},
	{"59fad97186848e7e9a29c6e2661d37d72d11ee8077aaa3cc78bb96f5efd5d9b7", 0, image.Rect(1398, 45, 1752, 415), image.Rect(1, 1, 16, 16), "c1230272e8c834b3ea4aa978a917a794c95b361da7f0dadfc8e7e619b91303ff"},
	{"eeeda48013f30299365d272e66ff46f1e4f2626ab7aa328984a9ac6fc2ae4dcb", 1, image.Rect(732, 97, 1067, 344), image.Rect(0, 4, 16, 16), "99bd5fc098982ceb8e373a359cb5a6c05431374798ad30a6017601cef85f4ecf"},
	{"dcaa4740634ab751c1709475b4fcd761226f3d220bdb50e859900084d63f92e6", 1, image.Rect(1107, 25, 1425, 344), image.Rect(0, 0, 16, 16), "1d9645c551822277c54e734e615b030255a47986ad6694f5008108992ccab183"},
	{"5eba8caa3e008cc886f70d7c552006afaaba7ec18be48d4f196f8a4375fca107", 1, image.Rect(26, 398, 346, 678), image.Rect(0, 1, 16, 15), "9f73372d65b6aabfb53734c467f2984c4546b59858b6c9da5fb5e6e7c221dd99"},
	{"02c9e83d897bf1e590354616bd3727b1f184fa52d1e7d16eff39725d89cd0764", 2, image.Rect(41, 65, 419, 419), image.Rect(0, 1, 16, 16), "0d0e23e8a97e1c5464dd0b2bd15054ce2225686a48575e48241e7f6f19376b2a"},
	{"e55b49761767127ec263f07e6507b176f508aaf80c96afdc7fa4761c747469b9", 2, image.Rect(494, 39, 854, 420), image.Rect(0, 0, 16, 16), "490fc5b6d98062c869db043e9b14ebe59367ee85e3333238cd54c8a10ab6fa48"},
	{"4c84e73e4189aa7b80ee75f6b9257fcc45a91886db2335c341fd54a46b13f3fe", 2, image.Rect(925, 248, 1308, 423), image.Rect(0, 8, 16, 16), "9d6170b09800a6be4378859aa1896a2625372761c10ab52853582fbae121465e"},
	{"6abffaaccedc85a4f68b257858615b6317f3180b4518ef005c80a1bb3758fa85", 2, image.Rect(1418, 95, 1730, 375), image.Rect(1, 2, 16, 14), "31a9e40f6a63018d80802cff6dc78459774a91c4796881261cfc2f07e82d705f"},
	{"56bca17b7a8cd8c4b18b0f110dfdab3ac391c28fa96da40afffd743beecae628", 1, image.Rect(380, 451, 711, 698), image.Rect(0, 4, 16, 16), "ff097e2eb0108387f2144fe5da9cdff3e53741485deb87dea28e46a58012cd94"},
	{"9bbdf0afd75237be2b432b837a60277f9f0598a674ed37067ee6be43ed37e4a9", 2, image.Rect(41, 469, 412, 859), image.Rect(0, 0, 16, 16), "004e721c4aee7a8c3905aed1e92a5e92718a4a4bda857d1140fbcad34620dba6"},
	{"4569e9b4fde700b519cb9b4ce570da39cc70464e85c3f19aac3f1c94e2da079c", 2, image.Rect(488, 496, 852, 850), image.Rect(0, 1, 16, 16), "d91e90f8eda575fe5bf61355d8897d2b64de42e48962ce0cb65f30d45ea45847"},
	{"7f1105802e295d998d5aa1d7e999234315249a1063a5ea110875ea13152a1282", 0, image.Rect(22, 483, 390, 852), image.Rect(0, 1, 15, 16), "e92546eecf3e5dee128e993c9002205e5474df1711b7d8b50d242510e4f8a1ef"},
	{"8c4443b7b6a508a48cde139dbb8187caa6167dfe28f09ba6b5397c727396f453", 0, image.Rect(464, 503, 864, 850), image.Rect(1, 2, 16, 16), "c610adc29da2b4f1991367413a03007bd70f852b2fa2c30712508dc13226e5b2"},
	{"c543c398267676eecf7da1fb7f66598c671e9716a06ff71289d2f5b6eeb6f706", 0, image.Rect(908, 517, 1277, 850), image.Rect(0, 2, 15, 16), "128f5e139c449e84550ff2acb3d5a0fba2830e35ffc84bd60018275b7e91853b"},
	{"2347a0f31c574581639dbae5d823fccdd869a6ddd87902277ccea9cb86bd21ba", 0, image.Rect(1349, 481, 1726, 851), image.Rect(0, 1, 15, 16), "6de9fd3b68b2245038a8d3603c1608da060ccee49b02465a976b927a72cfc48f"},
	{"aec99de8f5e412c3a414aee1c783b6c48ff029c14dc320add0fd4a389892b553", 1, image.Rect(737, 454, 1054, 703), image.Rect(0, 4, 16, 16), "b4b8532fced96c97c0b2ce34620343973ca4e3e95233d65b6b099aec66e995e8"},
	{"abef0dfac4d9c9c23938286412fb29d4db86631925674abf6d6f0e727eccb40f", 1, image.Rect(1105, 456, 1419, 703), image.Rect(0, 4, 16, 16), "1e6b77fd7b93c19856bb681558ea07de66c67cf7a280bebdc4410f18b76ef66d"},
	{"ba1b2b27163cea7114ad3a5ad61823767ee0d87e8f089a24e26da69faa56cafe", 1, image.Rect(22, 794, 343, 1050), image.Rect(0, 3, 16, 16), "1665a445f40f1e2133c2954cb28322b39899a3b7d18d68affeedc8854dd6ba7f"},
	{"ad943607d9f1ef26256bd046fb4b1ab35ca78b6cbe5ed11dd9b2d94ed2a9f5f7", 1, image.Rect(386, 737, 708, 1057), image.Rect(0, 0, 16, 16), "ef948ad749aa5a988e67d5e007fbfdfc5f4902f27da13bf82745e1b1416b27c4"},
	{"c82559904139aeed2e33844422c448f5be97f60e821d45f7875e3bce66ea474c", 2, image.Rect(968, 546, 1257, 812), image.Rect(2, 3, 14, 14), "8848015c1746770b5ce70252db663dd581a06c4ea1271b1609c64e68281f7438"},
	{"ae50aace23626db209bd43322a9d071155839a6d1593427cc8be05bd0f7c94b5", 1, image.Rect(744, 766, 1064, 1036), image.Rect(0, 1, 16, 15), "001b370976a31423e82442c455f1e684c229fe72aba37eafa6eeaf051197d46d"},
	{"53db8ad10c69de9a50e9c07e015e8770f4801ea9ad35fac31582a1ebb6bbc145", 2, image.Rect(1411, 546, 1702, 812), image.Rect(2, 3, 14, 14), "c116d941265311ed737c21a44145cbfeed155bbacb404740dbf0f86a8a5c75e7"},
	{"9b9c5261dc4449cfe02799784bcb3deaa154df55c6e3c241c118c969d46d3637", 1, image.Rect(1102, 774, 1425, 1051), image.Rect(0, 2, 16, 16), "bb7fa9fddfad10c232c3a4ab81eb35925f98a8e0a5791e2226311bf270b248e9"},
	{"fc4bd4ad202ffbeb2aed06ce3ac5ecaa62d1ff2264014964a674134c0a2a3e55", 3, image.Rect(158, 96, 1258, 986), image.Rect(0, 0, 16, 16), "9a332f490f84fd5c85aa20cfd27a28f446ed48cc07c4d05f192b4a62b6311b56"},
}

type caveObjectPose struct {
	art    image.Image
	number int
	bounds image.Rectangle
	pixels []color.NRGBA
}
type CaveObjects struct {
	poses  map[string]*caveObjectPose
	masked map[string]caveMaskedObject
	manual map[string]*caveObjectPose
}

func NewCaveObjects(dir string) (*CaveObjects, error) {
	atlases := make([]image.Image, len(caveObjectAtlases))
	for i, a := range caveObjectAtlases {
		raw, e := checked(filepath.Join(dir, a.name), a.sha, 16<<20)
		if e != nil {
			return nil, e
		}
		cfg, e := png.DecodeConfig(bytes.NewReader(raw))
		if e != nil || cfg.Width != a.width || cfg.Height != a.height {
			return nil, fmt.Errorf("洞窟物件圖面尺寸不符：%s", a.name)
		}
		atlases[i], e = png.Decode(bytes.NewReader(raw))
		if e != nil {
			return nil, e
		}
	}
	result := &CaveObjects{poses: map[string]*caveObjectPose{}}
	for n, r := range caveObjectRegistrations {
		if r.atlas < 0 || r.atlas >= len(atlases) || r.crop.Empty() || !r.crop.In(atlases[r.atlas].Bounds()) || r.native.Empty() || !r.native.In(image.Rect(0, 0, 16, 16)) || result.poses[r.key] != nil {
			return nil, fmt.Errorf("物件美術登錄不符")
		}
		box := image.Rectangle{Min: r.native.Min.Mul(3), Max: r.native.Max.Mul(3)}
		pose := &caveObjectPose{number: n, bounds: box, pixels: make([]color.NRGBA, 48*48)}
		crop := cavePlayerCrop{Image: atlases[r.atlas], box: r.crop}
		for y := 0; y < box.Dy(); y++ {
			for x := 0; x < box.Dx(); x++ {
				pose.pixels[(box.Min.Y+y)*48+box.Min.X+x] = caveArea(crop, x, y, box.Dx(), box.Dy())
			}
		}
		raw := make([]byte, 0, 48*48*4)
		for _, c := range pose.pixels {
			raw = append(raw, c.R, c.G, c.B, c.A)
		}
		if townBackgroundHash(raw) != r.pixelsSHA {
			return nil, fmt.Errorf("物件直接高清取樣不符：%d", n)
		}
		result.poses[r.key] = pose
	}
	return result, nil
}
func (s *CaveObjects) selectPose(object caveObject) (string, *caveObjectPose) {
	key := townBackgroundHash(object.Raster[:])
	if s == nil {
		return key, nil
	}
	target := key
	if object.Backgrounds != [4]byte{} {
		r, ok := s.masked[key]
		if !ok || r.backgrounds != object.Backgrounds || r.tiles != object.Tiles {
			return key, s.manualPose(object)
		}
		target = r.target
	}
	pose := s.poses[target]
	if pose == nil {
		pose = s.manualPose(object)
	}
	if pose == nil || len(pose.pixels) != 48*48 || pose.bounds.Empty() || !pose.bounds.In(image.Rect(0, 0, 48, 48)) {
		return key, nil
	}
	return key, pose
}
func caveObjectNative(dst *image.RGBA, object caveObject, palette []byte, origin image.Point, clear bool) {
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			px, py := object.X+x, object.Y+y
			v := object.Raster[y*16+x]
			if clear {
				v = 0
			}
			if v == 255 || px < 48 || px >= 272 || py < 14 || py >= 158 {
				continue
			}
			i := int(v) * 3
			c := color.RGBA{palette[i], palette[i+1], palette[i+2], 255}
			for yy := 0; yy < 3; yy++ {
				for xx := 0; xx < 3; xx++ {
					dst.SetRGBA(px*3+xx+origin.X, py*3+yy+origin.Y, c)
				}
			}
		}
	}
}
func (p *caveObjectPose) draw(dst *image.RGBA, object caveObject, origin image.Point) {
	for y := p.bounds.Min.Y; y < p.bounds.Max.Y; y++ {
		for x := p.bounds.Min.X; x < p.bounds.Max.X; x++ {
			px, py := object.X*3+x, object.Y*3+y
			c := p.pixels[y*48+x]
			if c.A == 0 || px < 144 || px >= 816 || py < 42 || py >= 474 {
				continue
			}
			px += origin.X
			py += origin.Y
			old := dst.RGBAAt(px, py)
			a, inv := uint32(c.A), uint32(255-c.A)
			dst.SetRGBA(px, py, color.RGBA{uint8((uint32(c.R)*a + uint32(old.R)*inv + 127) / 255), uint8((uint32(c.G)*a + uint32(old.G)*inv + 127) / 255), uint8((uint32(c.B)*a + uint32(old.B)*inv + 127) / 255), 255})
		}
	}
}
func (s *CaveObjects) draw(dst *image.RGBA, d *caveDecoded, palette []byte, origin image.Point) map[string]any {
	return s.drawWithTerrain(dst, d, palette, origin, nil)
}

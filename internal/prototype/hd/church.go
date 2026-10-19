package hd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

var churchBox = image.Rect(112, 23, 208, 87)
var churchHandPatch = image.Rect(165, 108, 249, 186)
var churchFlamePatches = [...]image.Rectangle{
	image.Rect(61, 89, 74, 114), image.Rect(80, 80, 92, 107),
	image.Rect(96, 89, 108, 114), image.Rect(172, 89, 184, 114), image.Rect(188, 80, 200, 107),
}

var churchReferences = [...]struct{ file, hash string }{
	{"church-anim00.indexed", "7f204a869102d8799da745b9421dc2c6592656a213fdeb4b6d37121042a2d7b1"},
	{"church-entered.indexed", "fd3e2e8326bdbc358c54cfc3390e6e71384d5886d4ff68614f49a45dc44233d1"},
	{"church-anim01.indexed", "062511f22958c58bce1a7c015f551a737c65894fccea497460e2669c7d3ba709"},
	{"church-wounded-page2.indexed", "f1da25c3c574767c68cf572ed097020bd5b42191d95ab204c86a39895aa7efb4"},
	{"church-healing-page2.indexed", "6ea27982fed67b55bc1e9a9a15d9939a86ef220ea2d5282064d47646710fe11e"},
	{"church-second-final.indexed", "98e8737cea2092f417dc4dea0f77617bff7e0fd6128ec9154d52d98fc020d1cd"},
	{"church-gesture-1-phase-0.indexed", "8f468ec72be594ed66a2f801be450f83a005b9529490c6595c547dec8bfe0cb2"},
	{"church-gesture-1-phase-1.indexed", "4955f7a6e547836495c983273c7a5c5a4c39c9f8df34d62a4dbcb7d01e15b416"},
	{"church-gesture-1-phase-2.indexed", "6c9252fa6af2c84e2d176edd606ae8feba47af5be5bffc34b7cfd39a6545d003"},
	{"church-gesture-2-phase-0.indexed", "aee2525d4935cc6fe2537750d7f61816d5d437587d222dbe258c2051bb6a35b6"},
	{"church-gesture-2-phase-1.indexed", "06b484946bb3ffaee81034950bf125a2f7a32a3d185146edf24c90df9dccae7e"},
	{"church-gesture-2-phase-2.indexed", "830ed32ec05f133bdc4c14fa3c27c39b4277bd82a0dc53ff9056df87a68b939c"},
	{"church-gesture-3-phase-0.indexed", "d157715324e177b54093c70a06fed7faa5ed6bbe16180e4c55d3d8920673633a"},
	{"church-gesture-3-phase-1.indexed", "41dec0baa1ab5f350e8c323da6a03d6f20f1a9728debbc16c16898feecc58a1d"},
	{"church-gesture-3-phase-2.indexed", "1dcf7b840b0f0ed41d590d15a704e16d669be03caa7c183c281443e169975a11"},
}

var churchArtworks = [...]struct{ file, hash string }{
	{"church-rest-handpaint-v1.png", "bf1f010947176a746edfc558c7bb4587058241357e671b6ee13c4b3bdcb741d7"},
	{"church-gesture-4-handpaint-v1.png", "178ac73699a504ec08ed270ffd4bf7fed8f50dcc0d0e4b09ff6c97937ed3708c"},
	{"church-flames-0-handpaint-v1.png", "81c90b5c5fe684fd4be45dd311fddfec3dd13d20f37dc2015fdbbe1ae3dec193"},
	{"church-flames-2-handpaint-v1.png", "865ee18a9374e92a01027d63491e317faa7b5a83aef7ec1cedbe8122464c7546"},
	{"church-gesture-1-handpaint-v1.png", "2fca4384e168cee0372dd61f7c1964bcf5d81b7bfedc00617a890db12793aa22"},
	{"church-gesture-2-handpaint-v1.png", "8f9be37e54a318d232f978040546ff1e5d0eefb3eec4c8524b408c47776408ba"},
	{"church-gesture-3-handpaint-v1.png", "53db22d4e8c8e568d9d24c8099c98e658c765f9e3d7adab1463b0d954d19c07b"},
}

// Church以15個已觀察的完整原版房間選圖，不增加動畫時間或接觸oracle。
// 畫格順序為休息、祝福、低位、胸前與下顎旁，各含三個燭光；見規格001。
type Church struct {
	indices [len(churchReferences)][]byte
	rgb     [len(churchReferences)][]byte
	art     [len(churchReferences)]*image.RGBA
}

func churchCrop(raw []byte, channels int) []byte {
	if len(raw) != nativeWidth*nativeHeight*channels {
		return nil
	}
	out := make([]byte, churchBox.Dx()*churchBox.Dy()*channels)
	for y := churchBox.Min.Y; y < churchBox.Max.Y; y++ {
		copy(out[(y-churchBox.Min.Y)*churchBox.Dx()*channels:], raw[(y*nativeWidth+churchBox.Min.X)*channels:(y*nativeWidth+churchBox.Max.X)*channels])
	}
	return out
}

func NewChurch(referenceDir, artworkDir string) (*Church, error) {
	if referenceDir == "" || artworkDir == "" {
		return nil, fmt.Errorf("教堂HD須同時提供原版參考與素材目錄")
	}
	pal, err := checked(filepath.Join(referenceDir, "church-entered.palette"), "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c", 768)
	if err != nil {
		return nil, err
	}
	m := &Church{}
	for i, ref := range churchReferences {
		raw, err := os.ReadFile(filepath.Join(referenceDir, ref.file))
		if err != nil {
			return nil, err
		}
		crop := churchCrop(raw, 1)
		if len(crop) == 0 || fmt.Sprintf("%x", sha256.Sum256(crop)) != ref.hash {
			return nil, fmt.Errorf("教堂原版房間畫格不符：%s", ref.file)
		}
		m.indices[i], m.rgb[i] = crop, make([]byte, len(crop)*3)
		for j, value := range crop {
			copy(m.rgb[i][j*3:j*3+3], pal[int(value)*3:int(value)*3+3])
		}
	}
	var variants [len(churchArtworks)]*image.RGBA
	for i, asset := range churchArtworks {
		data, err := checked(filepath.Join(artworkDir, asset.file), asset.hash, 16<<20)
		if err != nil {
			return nil, err
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil || cfg.Width != 1536 || cfg.Height != 1024 {
			return nil, fmt.Errorf("教堂HD須為已核對的1536×1024 PNG")
		}
		art, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		target := image.NewRGBA(image.Rect(0, 0, 282, 186))
		for y := 0; y < target.Bounds().Dy(); y++ {
			for x := 0; x < target.Bounds().Dx(); x++ {
				target.Set(x, y, art.At(x*cfg.Width/target.Bounds().Dx(), y*cfg.Height/target.Bounds().Dy()))
			}
		}
		variants[i] = target
	}
	for i := range m.art {
		target := image.NewRGBA(variants[0].Bounds())
		draw.Draw(target, target.Bounds(), variants[0], image.Point{}, draw.Src)
		gesture := 0
		if i >= 6 {
			gesture = 4 + (i-6)/3
		} else if i >= 3 {
			gesture = 1
		}
		if gesture != 0 {
			draw.Draw(target, churchHandPatch, variants[gesture], churchHandPatch.Min, draw.Src)
		}
		if phase := i % 3; phase != 1 {
			variant := 2
			if phase == 2 {
				variant = 3
			}
			for _, patch := range churchFlamePatches {
				draw.Draw(target, patch, variants[variant], patch.Min, draw.Src)
			}
		}
		m.art[i] = target
	}
	return m, nil
}

func (m *Church) Draw(dst *image.RGBA, indexed, rgb []byte, origin image.Point) (bool, int) {
	at := origin.Add(churchBox.Inset(1).Min.Mul(scale))
	box := image.Rect(0, 0, 282, 186).Add(at)
	if m == nil || dst == nil || !box.In(dst.Bounds()) {
		return false, -1
	}
	matched, frame := m.Match(indexed, rgb)
	if !matched {
		return false, -1
	}
	draw.Draw(dst, box, m.art[frame], image.Point{}, draw.Src)
	return true, frame
}

// Match使用同一原版快照辨識完整房間，不繪圖或記錄上一次狀態。
func (m *Church) Match(indexed, rgb []byte) (bool, int) {
	if m == nil {
		return false, -1
	}
	idx, col := churchCrop(indexed, 1), churchCrop(rgb, 3)
	if len(idx) == 0 || len(col) == 0 {
		return false, -1
	}
	for frame := range m.indices {
		if !bytes.Equal(idx, m.indices[frame]) || !bytes.Equal(col, m.rgb[frame]) {
			continue
		}
		return true, frame
	}
	return false, -1
}

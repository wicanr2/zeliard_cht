package hd

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

var townBackgroundBox = image.Rect(48, 24, 272, 102)

type townBackgroundProfile struct {
	module, atlas, mask, flags string
	tiles                      int
}

var townBackgroundProfiles = []townBackgroundProfile{
	{"968819bb6c4906514b4fce3e2ce99764b73cfa003f48d7d57d4cf8b12da8ffb6", "4498209b563897a0748e566ab3f6092568f0dc715d3a06444702a1a16b4c8229", "1d37729d004341fd0752fd6c2eecac02f02e7a8281d9cbe6c69fb31870866fa1", "39d0bc2e7a1c1ac63a45ce37be5a08d075a39d79607816518512a8c503a0cda7", 256},
	{"025bf2982660c942cbefed303ccdd0523ff824756a0fe0bb2587b99e0931e7ec", "889318f3979aaa81803facbf6d14ed4f7f45c78b4a7c2fe2881d64ea1e5da8c2", "0bdb78b2702490f3309f72ea3046fda2417bd088a9a6eec1c3f179fcad4cef71", "3c99e91e511b40beb89725b811da825b592d72b791cd936ef5682260e40ac677", townSceneTerrainTiles},
}

// TownBackground使用原版已畫圖塊快取與遮罩延伸遠山，不改前景。
// 原始位址、來源與失配政策見docs/spec/006-town-background-mask.md。
type TownBackground struct {
	indices, palette []byte
	art              *image.RGBA
}

func townBackgroundHash(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }

func NewTownBackground(root, reference, palette, memoryReference, artwork string) (*TownBackground, error) {
	for name, digest := range map[string]string{
		"ZELRES1.SAR": "56a73b59e952dab47fb0171a3fb1cdb7d641baceafa97e2826057f877a590eb0",
		"ZELRES2.SAR": "bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1",
	} {
		if _, err := checked(filepath.Join(root, name), digest, 1<<20); err != nil {
			return nil, err
		}
	}
	raw, err := os.ReadFile(reference)
	if err != nil {
		return nil, err
	}
	upper := skyCrop(raw, 1)
	if townBackgroundHash(upper) != "22fb614b02b3d9ac436c85b7d56c62ce9004d1254f6c52160de2cf62edeb7cf9" {
		return nil, fmt.Errorf("完整遠山原版畫格不符")
	}
	pal, err := checked(palette, "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c", 768)
	if err != nil {
		return nil, err
	}
	memory, err := checked(memoryReference, "9dd3657107f25c097debf5f6be68137d055d98bda2e22a59f6b854d88be5fdd9", 1<<20)
	if err != nil || len(memory) != 1<<20 {
		return nil, fmt.Errorf("完整遠山原版記憶體不符：%v", err)
	}
	s := &TownBackground{indices: append([]byte(nil), upper...), palette: pal}
	primary := memory[0x1b10:0x11b10]
	for y := 0; y < 24; y++ {
		for x := 0; x < 224; x++ {
			s.indices = append(s.indices, primary[0xa000+(x/8)*192+y*8+x%8])
		}
	}
	if townBackgroundHash(s.indices) != "17a6bbaa433a3dc7123dcc20fab641eab8eab2d743c7a6f70dafee08a656c63b" {
		return nil, fmt.Errorf("完整遠山背景快取不符")
	}
	data, err := checked(artwork, "3a6791169fb9fa28f3fac63310155249654ce3e80da21e6fec5b9d80bdf0d160", 16<<20)
	if err != nil {
		return nil, err
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width != 2126 || cfg.Height != 740 {
		return nil, fmt.Errorf("完整遠山HD尺寸不符")
	}
	art, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	s.art = image.NewRGBA(image.Rect(0, 0, 672, 234))
	for y := 0; y < 234; y++ {
		for x := 0; x < 672; x++ {
			s.art.Set(x, y, art.At(x*2126/672, y*740/234))
		}
	}
	return s, nil
}

func (s *TownBackground) backgroundMask(mode uint8, primary, resources, indexed, rgb []byte) ([]bool, bool) {
	if mode != 0x13 || len(primary) != 65536 || len(resources) != 65536 || len(indexed) != 64000 || len(rgb) != 192000 {
		return nil, false
	}
	if binary.LittleEndian.Uint16(primary[0xff2c:]) != 0x11b1 || binary.LittleEndian.Uint16(resources[0x8000:]) != 0x8006 ||
		townBackgroundHash(primary[0x3000:0x32fc]) != "9d263b7441ef4edc7ba21af5f7c8d66806ce5d4e993d0a2ee69cb3cbeac303f5" ||
		townBackgroundHash(primary[0xa000:0xb500]) != "1c2a41755ddff41db37e45b22517b1d9e3bfb8a658afa9d7a48fee2b05b12a5c" {
		return nil, false
	}
	tileLimit := 0
	for _, p := range townBackgroundProfiles {
		if townBackgroundHash(primary[0xc000:0xc040]) == p.module && townBackgroundHash(resources[0x8100:0x8100+p.tiles*48]) == p.atlas && townBackgroundHash(resources[0xd000:0xd800]) == p.mask && townBackgroundHash(resources[0x8006:0x8106]) == p.flags {
			tileLimit = p.tiles
			break
		}
	}
	if tileLimit == 0 {
		return nil, false
	}
	mask := make([]bool, 224*78)
	matches := func(x, y int, want byte) bool {
		i := (24+y)*320 + 48 + x
		return indexed[i] == want && bytes.Equal(rgb[i*3:i*3+3], s.palette[int(want)*3:int(want)*3+3])
	}
	for y := 0; y < 54; y++ {
		for x := 0; x < 224; x++ {
			i := y*224 + x
			if !matches(x, y, s.indices[i]) {
				return nil, false
			}
			mask[i] = true
		}
	}
	for col := 0; col < 28; col++ {
		for row := 0; row < 3; row++ {
			tile := int(primary[0xe000+col*8+row])
			if tile >= tileLimit {
				return nil, false
			}
			transparent := resources[0x8006+tile] != 0
			for dy := 0; dy < 8; dy++ {
				for dx := 0; dx < 8; dx++ {
					offset := 0x8100 + tile*48 + dy*6 + (dx/4)*3
					a, b, c := resources[offset], resources[offset+1], resources[offset+2]
					pixels := [4]byte{b >> 2, ((b & 3) << 4) | (a >> 4), ((a & 15) << 2) | (c >> 6), c & 63}
					x, y := col*8+dx, 54+row*8+dy
					i := y*224 + x
					useBackground := transparent && resources[0xd000+tile*8+dy]&(0x80>>dx) != 0
					want := pixels[dx%4]
					if useBackground {
						want = s.indices[i]
					}
					if !matches(x, y, want) {
						return nil, false
					}
					mask[i] = useBackground
				}
			}
		}
	}
	return mask, true
}

func (s *TownBackground) Draw(dst *image.RGBA, mode uint8, primary, resources, indexed, rgb []byte, origin image.Point) bool {
	box := s.art.Bounds().Add(origin.Add(townBackgroundBox.Min.Mul(scale)))
	if !box.In(dst.Bounds()) {
		return false
	}
	mask, ok := s.backgroundMask(mode, primary, resources, indexed, rgb)
	if !ok {
		return false
	}
	upper := image.Rect(0, 0, 672, 162).Add(box.Min)
	draw.Draw(dst, upper, s.art, image.Point{}, draw.Src)
	for y := 54; y < 78; y++ {
		for x := 0; x < 224; x++ {
			if !mask[y*224+x] {
				continue
			}
			for dy := 0; dy < scale; dy++ {
				a := s.art.PixOffset(x*scale, y*scale+dy)
				b := dst.PixOffset(box.Min.X+x*scale, box.Min.Y+y*scale+dy)
				copy(dst.Pix[b:b+scale*4], s.art.Pix[a:a+scale*4])
			}
		}
	}
	return true
}

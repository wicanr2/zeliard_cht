// 城堡來源、旗幟快取與3倍圖面契約見docs/spec/036-castle-hd.md。
package hd

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"path/filepath"
)

type castleProfile struct {
	background *TownBackground
	flags      [4]*image.RGBA
}

// castleResources只讀CPAT實際使用區，不把157格之後的其他資源當成城堡資料。
func castleResources(p, r []byte) bool {
	if len(p) != 65536 || len(r) != 65536 || townBackgroundHash(p[0xc000:0xc040]) != "968819bb6c4906514b4fce3e2ce99764b73cfa003f48d7d57d4cf8b12da8ffb6" || townBackgroundHash(r[0x8000:0x9e70]) != "88b751ec2afc08e4b93e92405c91c9c1dfc4f2f95aedff5c105e942878ef5dc6" {
		return false
	}
	var masks []byte
	for id := 0; id < 157; id++ {
		if r[0x8006+id] != 0 {
			masks = append(masks, r[0xd000+id*8:0xd008+id*8]...)
		}
	}
	return len(masks) == 232 && townBackgroundHash(masks) == "88ca58e07df6ddbef8b6c66bcccbacede13b94612c720191f2cd1728434a7c73"
}

func castleSources(p, r, e []byte) ([]townSceneActor, error) {
	if !castleResources(p, r) || len(e) != 65536 {
		return nil, fmt.Errorf("城堡地形來源拒絕")
	}
	for _, g := range []struct {
		b []byte
		h string
	}{
		{p[0x3000:0x371c], "cc52036cd4c7c43747d3235e3026b0e333c7eb692ed651dff569886b2233588d"},
		{p[0x6026:0x6c72], "6569836474343da7fd58f63b4342b7e5a323192e194920a30753c36020583d53"},
		{p[0xa000:0xb500], "1c2a41755ddff41db37e45b22517b1d9e3bfb8a658afa9d7a48fee2b05b12a5c"},
		{r[0x4000:0x5d80], "396a3b47dd240301fae669a494230194944987e2c46e32eaf35baa8e50d36825"},
		{r[0x6000:0x68a0], "045d4246ae0be72df3ac0201bc9ac8e5e1b50829783d6c79727fce7af0ef66c7"},
		{e[0x7000:0x74c0], "90de4921fb15fa26b9214b33e5c89daaa65106984db0f7417b3bd09f384392ce"},
		{e[0x8000:0x8170], "15f70e4b15a5887babfdf1c9154d788e6a37cd2e6a196e24cece90b1d6ac98d0"},
	} {
		if townBackgroundHash(g.b) != g.h {
			return nil, fmt.Errorf("城堡人物或程式來源拒絕")
		}
	}
	return townSceneActorRecords(p)
}

func (s *TownScenery) ownSources(p, r, e []byte) ([]townSceneActor, error) {
	if s.castle != nil {
		return castleSources(p, r, e)
	}
	return townSceneSources(p, r, e)
}

func (s *TownScenery) sourceProfile(in townFrameInput) *TownScenery {
	if s == nil {
		return nil
	}
	ok := false
	if s.castle == nil {
		ok = townFrameSource(in)
	} else if in.mode == 0x13 && len(in.palette) == 768 && townBackgroundHash(in.palette) == "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c" {
		_, err := s.ownSources(in.primary, in.resources, in.extra)
		ok = err == nil
	}
	if ok {
		return s
	}
	if s.castleVariant != nil {
		return s.castleVariant.sourceProfile(in)
	}
	return nil
}

func (s *TownScenery) frameSource(in townFrameInput) bool { return s.sourceProfile(in) != nil }

// AttachCastle僅由啟動流程組裝明示profile；原有Muralla來源守衛不變。
func (s *TownScenery) AttachCastle(profile *TownScenery) error {
	if s == nil || profile == nil || profile.castle == nil || s.castle != nil || s.castleVariant != nil || profile.castleVariant != nil {
		return fmt.Errorf("城堡profile組裝不符")
	}
	s.castleVariant = profile
	return nil
}

func castleImage(path, hash string, w, h int) (*image.RGBA, error) {
	b, err := checked(path, hash, 16<<20)
	if err != nil {
		return nil, err
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(b))
	if err != nil || cfg.Width != w || cfg.Height != h {
		return nil, fmt.Errorf("城堡HD尺寸不符")
	}
	im, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(out, out.Bounds(), im, image.Point{}, draw.Src)
	return out, nil
}

func NewCastleScenery(root, memoryReference, artDirectory string, background *TownBackground) (*TownScenery, error) {
	if background == nil {
		return nil, fmt.Errorf("城堡HD需要已驗完整遠山")
	}
	for name, h := range map[string]string{"ZELIARD.EXE": "fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341", "ZELRES2.SAR": "bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1"} {
		if _, err := checked(filepath.Join(root, name), h, 1<<20); err != nil {
			return nil, err
		}
	}
	memory, err := checked(memoryReference, "5c5b610c97bd2cd5266ae25b3da3764d0b46f02b30e8392a3b09c286a0db45a2", 1<<20)
	if err != nil || len(memory) != 1<<20 {
		return nil, fmt.Errorf("城堡正常參考不符：%v", err)
	}
	p, r, e := memory[0x1b10:0x11b10], memory[0x11b10:0x21b10], memory[0x21b10:0x31b10]
	actors, err := castleSources(p, r, e)
	if err != nil {
		return nil, err
	}
	s := &TownScenery{palette: cloneTownBytes(background.palette), castle: &castleProfile{background: background}}
	s.art, err = castleImage(filepath.Join(artDirectory, "castle-world-hd-v1.png"), "288ac2c28d625582a401343dedbc108ad8e50a7c78e1956cbe92b32ae9c63bb5", 2736, 192)
	if err != nil {
		return nil, err
	}
	for i, h := range []string{"580256402c97125098158c1e1621f9421b703fd1fdd635961c55f6c0a8a09d07", "98f19ee008179c070a43bd2d856089cf16b414fe80194d252bf2184241477bbf", "9c0a04a1f4df1ac6d720b8b8b132aa600c2145b52a832f36feaf5af8bf4e3bda", "b3b11950a3f689da88fe0e2c60c22625a221ea305e6ebf435fbda5ef26b5c1ac"} {
		s.castle.flags[i], err = castleImage(filepath.Join(artDirectory, fmt.Sprintf("castle-flag-%d-hd-v1.png", i+2)), h, 24, 24)
		if err != nil {
			return nil, err
		}
	}
	for x := 0; x < 114; x++ {
		for y := 0; y < 8; y++ {
			id := p[0xc017+x*8+y]
			if id == 253 {
				found := false
				for _, actor := range actors {
					if actor.x == x {
						id = byte(actor.under)
						found = true
						break
					}
				}
				if !found {
					return nil, fmt.Errorf("城堡底圖缺項")
				}
			}
			if id >= 157 {
				return nil, fmt.Errorf("城堡圖格越界")
			}
			if id >= 2 && id <= 5 {
				id = 2
			}
			s.tiles[x][y] = id
		}
	}
	return s, nil
}

type castleFlagPaint struct {
	box        image.Rectangle
	art        *image.RGBA
	background image.Point
}

func (c *castleProfile) prepareFlags(dst *image.RGBA, mode uint8, p, r, raw, rgb []byte, f *townSceneFrame, origin image.Point) ([]castleFlagPaint, bool) {
	if c == nil || c.background == nil || len(f.flags) != 14336 {
		return nil, false
	}
	if _, ok := c.background.backgroundMask(mode, p, r, raw, rgb); !ok {
		return nil, false
	}
	var flags []castleFlagPaint
	for col := 0; col < 28; col++ {
		for row := 0; row < 3; row++ {
			id := f.flags[row*8*224+col*8]
			if id == 0 {
				continue
			}
			if id < 2 || id > 5 || c.flags[id-2] == nil || c.flags[id-2].Bounds() != image.Rect(0, 0, 24, 24) {
				return nil, false
			}
			box := image.Rect(144+col*24, 234+row*24, 168+col*24, 258+row*24).Add(origin)
			if !box.In(dst.Bounds()) {
				return nil, false
			}
			flags = append(flags, castleFlagPaint{box, c.flags[id-2], image.Pt(col*24, (54+row*8)*3)})
		}
	}
	return flags, true
}

func (c *castleProfile) paintBackground(dst *image.RGBA, flags []castleFlagPaint) {
	for _, flag := range flags {
		draw.Draw(dst, flag.box, c.background.art, flag.background, draw.Src)
	}
}

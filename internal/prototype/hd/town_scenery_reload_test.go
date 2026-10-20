package hd

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"testing"
)

func TestTownSceneryReloadedOriginalSources(t *testing.T) {
	root := os.Getenv("ZELIARD_SCENERY_RELOAD_ROOT")
	if root == "" {
		t.Skip("須掛載正常重讀來源及獨立opacity證據；skip不算驗收")
	}
	s, _ := sceneryTestInputs(t)
	for _, name := range []string{"restored", "second-restored", "second-cave-near"} {
		t.Run(name, func(t *testing.T) {
			mem, e := os.ReadFile(filepath.Join(root, "reload-v1-"+name+".memory"))
			if e != nil {
				t.Fatal(e)
			}
			raw, e := os.ReadFile(filepath.Join(root, "reload-v1-"+name+".indexed"))
			if e != nil {
				t.Fatal(e)
			}
			mask, e := os.ReadFile(filepath.Join(root, "reload-"+name+"-sprite-opacity.pgm"))
			if e != nil {
				t.Fatal(e)
			}
			header := []byte("P5\n224 64\n255\n")
			if !bytes.HasPrefix(mask, header) || len(mask) != len(header)+14336 {
				t.Fatal("獨立opacity不符")
			}
			mask = mask[len(header):]
			original := cloneTownBytes(mem)
			rgb, canvas := skyOriginal(raw, s.palette)
			before := cloneTownBytes(canvas.Pix)
			if !s.Draw(canvas, 0x13, mem[0x1b10:0x11b10], mem[0x11b10:0x21b10], mem[0x21b10:0x31b10], raw, rgb, image.Point{}, true) {
				t.Fatal("正常重讀來源被拒絕", name)
			}
			changed := 0
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := canvas.PixOffset(x, y)
					if bytes.Equal(canvas.Pix[i:i+4], before[i:i+4]) {
						continue
					}
					changed++
					nx, ny := x/3, y/3
					if !image.Pt(nx, ny).In(image.Rect(48, 78, 272, 142)) || mask[(ny-78)*224+nx-48] != 0 {
						t.Fatal("重讀HD覆寫角色或區外", x, y)
					}
				}
			}
			if changed == 0 || !bytes.Equal(mem, original) {
				t.Fatal("未畫HD或改寫來源")
			}
		})
	}
}

func TestTownSceneryUnusedBytesAndSourceBoundaries(t *testing.T) {
	s, in := townFrameFixture(t)
	original := copyTownInput(in)
	for i := 0x5d80; i < 0x6000; i++ {
		in.resources[i] ^= 0xa5
	}
	for i := 0x74c0; i < 0x8000; i++ {
		in.extra[i] ^= 0xa5
	}
	// 規格041：Muralla只讀地形0..239；離店會改到第242格遮罩。
	for i := 0xd000 + townSceneTerrainTiles*8; i < 0xd800; i++ {
		in.resources[i] ^= 0xa5
	}
	canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
	if !s.Draw(canvas, in.mode, in.primary, in.resources, in.extra, in.indexed, in.rgb, image.Point{}, true) {
		t.Fatal("保留區改變仍誤拒絕")
	}
	frames := &TownFrames{scene: s}
	frames.finish(in)
	if frames.complete == nil {
		t.Fatal("完整快照仍誤拒絕保留區")
	}
	for _, name := range []string{"npc-last-pixel", "npc-last-mask", "player-image", "player-mask", "terrain-last-mask", "unknown-group"} {
		t.Run(name, func(t *testing.T) {
			input := copyTownInput(original)
			switch name {
			case "npc-last-pixel":
				input.resources[0x5d7f] ^= 1
			case "npc-last-mask":
				input.extra[0x74bf] ^= 1
			case "player-image":
				input.resources[0x6000] ^= 1
			case "player-mask":
				input.extra[0x8000] ^= 1
			case "terrain-last-mask":
				input.resources[0xd000+townSceneTerrainTiles*8-1] ^= 1
			case "unknown-group":
				input.primary[townSceneWord(input.primary, 0xc00f)+2] = 5
			}
			before := cloneTownBytes(canvas.Pix)
			if s.Draw(canvas, input.mode, input.primary, input.resources, input.extra, input.indexed, input.rgb, image.Point{}, true) || !bytes.Equal(before, canvas.Pix) {
				t.Fatal("已使用來源或未知group未整層拒絕")
			}
			if townFrameSource(input) {
				t.Fatal("完整畫格與前景的來源守衛不一致")
			}
		})
	}
}

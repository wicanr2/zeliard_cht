package hd

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestTransitionAndCoherentRegisteredSources(t *testing.T) {
	root := os.Getenv("ZELIARD_PROJECT")
	out := os.Getenv("ZELIARD_TRANSITION_OUTPUT")
	if root == "" || out == "" {
		t.Skip("需明確掛載本地原版及新素材；skip不作驗收")
	}
	dir := filepath.Join(root, "workplace/full-hd-v1/manual-hd-v1/registered")
	art, e := NewManualArtwork(filepath.Join(root, "workplace/orig"), dir)
	if e != nil {
		t.Fatal(e)
	}
	for _, pose := range art.transition.Poses {
		if pose.art == nil {
			t.Fatal("已登錄過場角色未載入HD")
		}
	}
	read := func(p string) []byte {
		b, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	need := func(e error) {
		if e != nil {
			t.Fatal(e)
		}
	}
	save := func(name string, im image.Image) {
		f, e := os.Create(filepath.Join(out, name+".png"))
		need(e)
		need(png.Encode(f, im))
		need(f.Close())
	}
	canvas := func() *image.RGBA {
		im := image.NewRGBA(image.Rect(0, 0, 960, 680))
		draw.Draw(im, im.Bounds(), image.NewUniform(color.RGBA{31, 41, 59, 255}), image.Point{}, draw.Src)
		return im
	}
	rgbOf := func(raw, pal []byte) []byte {
		rgb := make([]byte, len(raw)*3)
		for i, v := range raw {
			copy(rgb[i*3:i*3+3], pal[int(v)*3:int(v)*3+3])
		}
		return rgb
	}
	nativeScene := func(rgb []byte) *image.RGBA {
		im := canvas()
		for y := 14; y < 158; y++ {
			for x := 48; x < 272; x++ {
				i := (y*320 + x) * 3
				c := color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255}
				for dy := 0; dy < 3; dy++ {
					for dx := 0; dx < 3; dx++ {
						im.SetRGBA(x*3+dx, y*3+dy+48, c)
					}
				}
			}
		}
		return im
	}
	oracle := filepath.Join(root, "workplace/hd-cave-transition-v1/oracle-dense")
	var receipt struct{ Shots []struct{ File string } }
	need(json.Unmarshal(read(filepath.Join(oracle, "receipt.json")), &receipt))
	matched := 0
	for _, shot := range receipt.Shots {
		raw := read(filepath.Join(oracle, shot.File+".indexed"))
		pal := read(filepath.Join(oracle, shot.File+".palette"))
		rgb := rgbOf(raw, pal)
		before := append([]byte(nil), raw...)
		im := canvas()
		if art.DrawCaveTransition(im, raw, rgb, image.Pt(0, 48)) {
			matched++
			if matched == 1 || matched == 50 || matched == 100 || matched == 160 {
				save("transition-"+shot.File, im)
			}
			if im.RGBAAt(10, 60) != (color.RGBA{31, 41, 59, 255}) {
				t.Fatal("過場越界")
			}
			invalid := append([]byte(nil), raw...)
			invalid[14*320+48] ^= 1
			reject := canvas()
			expected := append([]byte(nil), reject.Pix...)
			if art.DrawCaveTransition(reject, invalid, rgbOf(invalid, pal), image.Pt(0, 48)) || !bytes.Equal(reject.Pix, expected) {
				t.Fatal("未知來源覆寫畫布")
			}
			badRGB := append([]byte(nil), rgb...)
			badRGB[(14*320+48)*3] ^= 1
			if art.DrawCaveTransition(canvas(), raw, badRGB, image.Pt(0, 48)) {
				t.Fatal("異常色彩未拒絕")
			}
		}
		if !bytes.Equal(raw, before) {
			t.Fatal("過場改寫來源")
		}
	}
	if matched < 150 {
		t.Fatalf("已觀察過場缺漏：%d", matched)
	}
	t.Logf("normal journal transition samples matched=%d/%d", matched, len(receipt.Shots))
	var gui struct {
		Shots []struct {
			File  string
			Steps uint64
			Regs  struct{ DI uint16 }
		}
	}
	guiDir := filepath.Join(root, "workplace/hd-cave-transition-v1/oracle-gui-v2")
	need(json.Unmarshal(read(filepath.Join(guiDir, "receipt.json")), &gui))
	for _, shot := range gui.Shots {
		if shot.Steps != 125427051 && shot.Steps != 125587272 {
			continue
		}
		raw := read(filepath.Join(guiDir, shot.File+".indexed"))
		pal := read(filepath.Join(guiDir, shot.File+".palette"))
		rgb := rgbOf(raw, pal)
		im := nativeScene(rgb)
		if !art.DrawCaveTransition(im, raw, rgb, image.Pt(0, 48)) {
			t.Fatal("原版未完成單格造成整幅HD撤除", shot.Steps)
		}
		for offset := 0; offset < 2; offset++ {
			i := int(shot.Regs.DI) + offset
			x, y := i%320, i/320
			c := im.RGBAAt(x*3, y*3+48)
			if c.R != rgb[i*3] || c.G != rgb[i*3+1] || c.B != rgb[i*3+2] {
				t.Fatal("未完成原生圖格被猜補")
			}
		}
	}
	// 實際原版城鎮到黑屏的遮罩，逐點只減不增。
	old := filepath.Join(root, "workplace/hd-cave-transition-v1/oracle")
	base := read(filepath.Join(old, "frame-0016.indexed"))
	pal := read(filepath.Join(old, "frame-0016.palette"))
	im := canvas()
	frozen := &TownPresentation{SourceIndexed: base, RGB: rgbOf(base, pal), Palette: pal}
	previous := 0
	for _, name := range []string{"frame-0017", "frame-0018", "frame-0019", "frame-0020"} {
		raw := read(filepath.Join(old, name+".indexed"))
		wiping, erased := nativeTownWipe(townFrameInput{mode: 0x13, indexed: raw, rgb: rgbOf(raw, pal), palette: pal}, frozen)
		if wiping == nil || erased < previous || !wiping.ApplyErase(im, image.Pt(0, 48)) {
			t.Fatal("正常原版擦除拒絕", name)
		}
		previous = erased
		save("wipe-"+name, im)
	}
	unknown := append([]byte(nil), base...)
	unknown[30*320+80] = 3
	if f, _ := nativeTownWipe(townFrameInput{mode: 0x13, indexed: unknown, rgb: rgbOf(unknown, pal), palette: pal}, frozen); f != nil {
		t.Fatal("異常擦除未拒絕")
	}
	// 首領為已知原生完整姿態的靜態接入檢查，不宣稱後段正常遊玩。
	for _, boss := range art.coherent.entries {
		raw := make([]byte, 64000)
		left, top := 60, 20
		for y := 0; y < boss.Height; y++ {
			copy(raw[(top+y)*320+left:], boss.Raster[y*boss.Width:(y+1)*boss.Width])
		}
		rgb := rgbOf(raw, boss.Colors)
		before := append([]byte(nil), raw...)
		im := canvas()
		if !art.DrawCoherent(im, raw, rgb, image.Pt(0, 48)) {
			t.Fatal("完整首領未接入", boss.Key)
		}
		save("boss-"+boss.Key, im)
		if im.RGBAAt(10, 60) != (color.RGBA{31, 41, 59, 255}) || !bytes.Equal(before, raw) {
			t.Fatal("首領來源或區外被寫入")
		}
		for i, v := range boss.Raster {
			if v != 0 {
				at := (top+i/boss.Width)*320 + left + i%boss.Width
				raw[at] ^= 1
				break
			}
		}
		reject := canvas()
		expected := append([]byte(nil), reject.Pix...)
		if art.DrawCoherent(reject, raw, rgbOf(raw, boss.Colors), image.Pt(0, 48)) || !bytes.Equal(expected, reject.Pix) {
			t.Fatal("不完整首領未拒絕", boss.Key)
		}
	}
}

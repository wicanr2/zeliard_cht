package hd

import (
	"image"
	"image/color"
	"testing"
)

// 規格067：洞窟開 F7 選單時，畫面只在開著的方框內改變才沿用完成畫面。
func TestCaveDiffInside(t *testing.T) {
	last := &CavePresentation{SourceIndexed: make([]byte, 64000), RGB: make([]byte, 192000), Palette: make([]byte, 768)}
	in := caveFrameInput{indexed: make([]byte, 64000), rgb: make([]byte, 192000), palette: make([]byte, 768)}
	box := image.Rect(104, 70, 224, 110)
	if got := caveDiffInside(in, last, []image.Rectangle{box}); !got.Empty() {
		t.Fatal("畫面相同卻回報差異", got)
	}
	in.indexed[80*320+120], in.indexed[100*320+200] = 7, 7
	if got := caveDiffInside(in, last, []image.Rectangle{box}); got != image.Rect(120, 80, 201, 101) {
		t.Fatal("框內差異範圍不符", got)
	}
	in.indexed[130*320+60] = 7
	if got := caveDiffInside(in, last, []image.Rectangle{box}); !got.Empty() {
		t.Fatal("框外也有變化仍沿用")
	}
	in.indexed[130*320+60] = 0
	in.palette[3] = 1
	if got := caveDiffInside(in, last, []image.Rectangle{box}); !got.Empty() {
		t.Fatal("色盤改變仍沿用")
	}
}

// 規格067：結局人物說話時，大插畫只有少量點不同就沿用整張 HD；換圖或大量清除就停止。
func TestCompleteArtworkHoldsTalkingPicture(t *testing.T) {
	art := image.NewNRGBA(image.Rect(0, 0, 64*3, 64*3))
	for i := range art.Pix {
		art.Pix[i] = 200
		if i%4 == 3 {
			art.Pix[i] = 255
		}
	}
	e := &completeEntry{Group: 0, Width: 64, Height: 64, Raster: make([]byte, 64*64), art: art, anchorLen: 8}
	for i := range e.Raster {
		e.Raster[i] = byte(1 + i%7)
	}
	s := &CompleteArtwork{groups: []completeGroup{{}}, entries: []*completeEntry{e}, active: []bool{true}, introMode: true}
	s.reindex()
	raw, rgb := make([]byte, 64000), make([]byte, 192000)
	place := func() {
		for y := 0; y < 64; y++ {
			copy(raw[(50+y)*320+100:(50+y)*320+164], e.Raster[y*64:y*64+64])
		}
	}
	painted := func() bool {
		dst := image.NewRGBA(image.Rect(0, 0, 960, 600))
		s.draw(dst, raw, rgb, image.Point{})
		return dst.RGBAAt(100*3+90, 50*3+90) == (color.RGBA{200, 200, 200, 255})
	}
	place()
	if !painted() {
		t.Fatal("完整插畫未覆繪")
	}
	for y := 20; y < 36; y++ { // 說話小圖：16×16，不到八分之一
		for x := 20; x < 36; x++ {
			raw[(50+y)*320+100+x] = 0
		}
	}
	if !painted() || !s.holding {
		t.Fatal("說話小圖讓整張 HD 撤除")
	}
	s.introMode = false
	for y := 20; y < 36; y++ {
		for x := 20; x < 36; x++ {
			raw[(50+y)*320+100+x] = 1
		}
	}
	if painted() {
		t.Fatal("遊戲模式不該沿用大插畫")
	}
	s.introMode = true
	place()
	if !painted() {
		t.Fatal("恢復完整後未重新比對到")
	}
	for y := 20; y < 36; y++ {
		for x := 20; x < 36; x++ {
			raw[(50+y)*320+100+x] = 0
		}
	}
	if !painted() {
		t.Fatal("片頭模式說話小圖未沿用")
	}
	for y := 0; y < 40; y++ { // 大量清除：超過八分之一
		for x := 0; x < 64; x++ {
			raw[(50+y)*320+100+x] = 0
		}
	}
	if painted() {
		t.Fatal("大量清除仍沿用")
	}
	place()
	raw[(50+30)*320+130] = 0
	if painted() {
		t.Fatal("停止後不該再沿用，須重新整張比對")
	}
}

package main

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

// dissolveFixture 以上半部非0、下半部0的原版畫面，配上整幅紅色的合成畫面（下半部代表畫在黑底上的中文）。
func dissolveFixture(added int, region image.Rectangle, exclude []image.Rectangle) (*dissolveHold, []byte) {
	raw := make([]byte, 64000)
	for i := 0; i < 32000; i++ {
		raw[i] = 5
	}
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(canvas, dissolveArea, &image.Uniform{color.RGBA{200, 0, 0, 255}}, image.Point{}, draw.Src)
	var pal [256][3]uint8
	return newDissolveHold("s", pal, raw, canvas, added, region, exclude), raw
}

func blockColor(c *image.RGBA, x, y int) color.RGBA {
	return c.RGBAAt(x*scale+1, barHeight+y*scale+1)
}

func TestDissolveFollowsClearedPixels(t *testing.T) {
	hold, raw := dissolveFixture(introDissolveAdded, introStoryRegion, nil)
	next := append([]byte(nil), raw...)
	for x := 0; x < 64; x++ {
		next[10*320+x] = 0
	}
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	applied, remaining := hold.dissolve(canvas, next)
	if !applied || remaining != 32000-64 {
		t.Fatal("逐點清除未介入", applied, remaining)
	}
	if blockColor(canvas, 5, 10).R != 0 || blockColor(canvas, 100, 10).R != 200 {
		t.Fatal("清掉的點未塗黑或其餘點未沿用保存畫面")
	}
	kept := 0
	for x := 0; x < 320; x++ {
		if blockColor(canvas, x, 120).R == 200 {
			kept++
		}
	}
	if kept < 300 {
		t.Fatal("剩餘比例接近1時，黑底上的中文不該大量消失", kept)
	}
}

func TestDissolveWholeClearLeavesNothing(t *testing.T) {
	hold, _ := dissolveFixture(introDissolveAdded, introStoryRegion, nil)
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	applied, remaining := hold.dissolve(canvas, make([]byte, 64000))
	if !applied || remaining != 0 {
		t.Fatal("整區清空未介入", applied, remaining)
	}
	for y := 0; y < 200; y += 7 {
		for x := 0; x < 320; x += 5 {
			if blockColor(canvas, x, y).R != 0 {
				t.Fatal("整區一次清空後仍殘留保存畫面", x, y)
			}
		}
	}
}

func TestDissolveRefusesOtherWrites(t *testing.T) {
	hold, raw := dissolveFixture(introDissolveAdded, introStoryRegion, nil)
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	if applied, _ := hold.dissolve(canvas, raw); applied {
		t.Fatal("畫面未變也介入")
	}
	next := append([]byte(nil), raw...)
	next[0] = 0
	for i := 0; i <= introDissolveAdded; i++ {
		next[40000+i] = 9
	}
	if applied, _ := hold.dissolve(canvas, next); applied {
		t.Fatal("新增寫入過多仍介入")
	}
	// 片頭實測：清黑開始的同一段有68點清除與最後一個打字字形的18點新增。
	next = append([]byte(nil), raw...)
	for i := 0; i < 68; i++ {
		next[i] = 0
	}
	for i := 0; i < 18; i++ {
		next[40000+i] = 9
	}
	if applied, _ := hold.dissolve(canvas, next); !applied || blockColor(canvas, 0, 125).R != 200 {
		t.Fatal("少量新增寫入應沿用保存畫面，不露出原版")
	}
	// 捲動：新增與清除相當，不介入。
	scroll, raw := dissolveFixture(introDissolveAdded, introStoryRegion, nil)
	next = append([]byte(nil), raw...)
	for i := 0; i < 40; i++ {
		next[i] = 0
		next[40000+i] = 9
	}
	if applied, _ := scroll.dissolve(canvas, next); applied {
		t.Fatal("新增與清除相當的捲動也介入")
	}
	room, raw := dissolveFixture(0, roomTextBox, nil)
	next = append([]byte(nil), raw...)
	next[0], next[40000] = 0, 9
	if applied, _ := room.dissolve(canvas, next); applied {
		t.Fatal("房間不容許任何新增寫入")
	}
}

// 句子之間只清對話框：框內原本非0的點全清，框內黑底上的中文須一起消失，不殘留舊句。
func TestDissolveRegionClearsTextBox(t *testing.T) {
	hold, raw := dissolveFixture(0, roomTextBox, nil)
	box := image.Rect(54, 98, 264, 150)
	for y := 100; y < 104; y++ {
		for x := 60; x < 200; x++ {
			raw[y*320+x] = 9
		}
	}
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(canvas, dissolveArea, &image.Uniform{color.RGBA{200, 0, 0, 255}}, image.Point{}, draw.Src)
	var pal [256][3]uint8
	hold = newDissolveHold("s", pal, raw, canvas, 0, roomTextBox, nil)
	next := append([]byte(nil), raw...)
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			next[y*320+x] = 0
		}
	}
	out := image.NewRGBA(image.Rect(0, 0, width, height))
	if applied, _ := hold.dissolve(out, next); !applied {
		t.Fatal("清框未介入")
	}
	for y := box.Min.Y; y < box.Max.Y; y += 3 {
		for x := box.Min.X; x < box.Max.X; x += 3 {
			if blockColor(out, x, y).R != 0 {
				t.Fatal("清框後框內仍殘留舊中文", x, y)
			}
		}
	}
	if blockColor(out, 10, 10).R != 200 {
		t.Fatal("框外未沿用保存畫面")
	}
	for y := 160; y < 200; y++ {
		for x := 0; x < 320; x++ {
			if blockColor(out, x, y).R != 200 {
				t.Fatal("只清對話框時，框外黑底上的高清細節被塗黑", x, y)
			}
		}
	}
}

// 人像動畫只清人像框內的點時不介入；轉場開始後才放寬。
func TestDissolveIgnoresPortraitOnlyClears(t *testing.T) {
	portrait := image.Rect(20, 10, 100, 70)
	hold, raw := dissolveFixture(0, roomTextBox, []image.Rectangle{portrait})
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	next := append([]byte(nil), raw...)
	next[30*320+40] = 0
	if applied, _ := hold.dissolve(canvas, next); applied {
		t.Fatal("只清人像框內的點也介入")
	}
	next[30*320+200] = 0
	if applied, _ := hold.dissolve(canvas, next); !applied {
		t.Fatal("人像框外也清除時未介入")
	}
	next[30*320+200] = 5
	if applied, _ := hold.dissolve(canvas, next); !applied {
		t.Fatal("轉場開始後只剩框內清除時中斷")
	}
}

package hd

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

// 真實Ada日誌重生的洞窟畫面，曾被只有一個色點的結局片段誤認。
func TestCompleteArtworkCaveContext(t *testing.T) {
	root := os.Getenv("ZELIARD_PROJECT")
	if root == "" {
		t.Skip("需要本地素材及Ada重生來源；skip不算驗收")
	}
	read := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(root, "workplace/hd-cave-artifact-v1/replay-source", name))
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	raw, palette, memory := read("source.indexed"), read("source.palette"), read("source.memory")
	if townBackgroundHash(raw) != "09731a92bdbab893c8d3a82bb0a7f183c8123e0a2be2a9d2760f75ee5a4768f9" {
		t.Fatal("Ada完成來源身份不符")
	}
	original := append([]byte(nil), memory...)
	art, err := NewManualArtwork(filepath.Join(root, "workplace/orig"), filepath.Join(root, "workplace/full-hd-v1/manual-hd-v1/registered"))
	if err != nil {
		t.Fatal(err)
	}
	rgb := make([]byte, 192000)
	for i, v := range raw {
		copy(rgb[i*3:i*3+3], palette[int(v)*3:int(v)*3+3])
	}
	canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
	bitmap.DrawRGB3(canvas, rgb, image.Point{})
	art.SetResources(memory[0x11b10:0x21b10], memory[0x21b10:0x31b10])
	art.Draw(canvas, raw, rgb, image.Point{})
	for _, placement := range art.complete.placements {
		if placement.x == 159 && placement.y == 90 && placement.entry.Source == "yuu3-final" {
			t.Fatal("結局片段誤貼洞窟")
		}
	}
	if !bytes.Equal(memory, original) {
		t.Fatal("原版來源被修改")
	}
	var fragment *completeEntry
	for _, entry := range art.complete.entries {
		if entry.File == "complete-yuu3-final-000-x120-y072.png" {
			fragment = entry
			break
		}
	}
	if fragment == nil || fragment.context == nil || fragment.contextMatches(raw, 159, 90) {
		t.Fatal("錯誤上下文仍被接受")
	}
	parent := fragment.context
	for _, position := range []image.Point{image.Pt(32, 4), image.Pt(-100, -60)} {
		frame := bytes.Repeat([]byte{255}, 64000)
		visible := image.Rect(position.X, position.Y, position.X+parent.Width, position.Y+parent.Height).Intersect(image.Rect(0, 0, 320, 200))
		for y := visible.Min.Y; y < visible.Max.Y; y++ {
			for x := visible.Min.X; x < visible.Max.X; x++ {
				frame[y*320+x] = parent.Raster[(y-position.Y)*parent.Width+x-position.X]
			}
		}
		x, y := position.X+fragment.contextX, position.Y+fragment.contextY
		if !fragment.contextMatches(frame, x, y) {
			t.Fatal("正確母圖或邊界裁切被拒絕", position)
		}
		// 改動片段之外、仍在可見母圖內的像素，局部圖樣維持原狀。
		frame[visible.Min.Y*320+visible.Min.X] ^= 1
		if fragment.contextMatches(frame, x, y) {
			t.Fatal("只看子圖而忽略母圖")
		}
	}
	if output := os.Getenv("ZELIARD_ARTIFACT_OUTPUT"); output != "" {
		file, err := os.Create(filepath.Join(output, "manual-after.png"))
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		if err = png.Encode(file, canvas); err != nil {
			t.Fatal(err)
		}
	}
}

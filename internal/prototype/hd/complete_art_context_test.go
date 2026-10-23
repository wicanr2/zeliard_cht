package hd

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
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
		// 規格068補登錄的片頭／結局插畫只能在原版畫出整張母圖時出現。
		if strings.HasPrefix(placement.entry.Source, "story-") {
			t.Fatal("片頭插畫片段誤貼洞窟", placement.entry.File, placement.x, placement.y)
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
	if fragment == nil || fragment.context == nil || fragment.contextMatches(raw, 159, 90, false) || fragment.contextMatches(raw, 159, 90, true) {
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
		if !fragment.contextMatches(frame, x, y, false) {
			t.Fatal("正確母圖或邊界裁切被拒絕", position)
		}
		// 改動片段之外、仍在可見母圖內的像素，局部圖樣維持原狀。
		frame[visible.Min.Y*320+visible.Min.X] ^= 1
		if fragment.contextMatches(frame, x, y, false) {
			t.Fatal("只看子圖而忽略母圖")
		}
		// 片頭／結局模式：母圖可見部分過半相符仍接受，過半不符就拒絕（規格068）。
		if !fragment.contextMatches(frame, x, y, true) {
			t.Fatal("片頭模式拒絕僅一點不同的母圖", position)
		}
		// 先還原上面改動的點，再改動片段以外、超過可見母圖一半的特徵點（黑色與兩種主要背景色不算上下文證據）。
		frame[visible.Min.Y*320+visible.Min.X] ^= 1
		key := func(v byte) bool {
			return v != 0 && int(v) != parent.background[0] && int(v) != parent.background[1]
		}
		inked := 0
		for py := visible.Min.Y; py < visible.Max.Y; py++ {
			for px := visible.Min.X; px < visible.Max.X; px++ {
				if key(parent.Raster[(py-position.Y)*parent.Width+px-position.X]) {
					inked++
				}
			}
		}
		need := inked/2 + 1
		for py := visible.Min.Y; py < visible.Max.Y && need > 0; py++ {
			for px := visible.Min.X; px < visible.Max.X && need > 0; px++ {
				if key(parent.Raster[(py-position.Y)*parent.Width+px-position.X]) && !image.Pt(px, py).In(image.Rect(x, y, x+fragment.Width, y+fragment.Height)) {
					frame[py*320+px] ^= 1
					need--
				}
			}
		}
		if need != 0 || len(parent.Mask) != 0 {
			t.Fatal("過半改動的測試前提不成立", position)
		}
		if fragment.contextMatches(frame, x, y, true) {
			t.Fatal("片頭模式接受過半不符的母圖", position)
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

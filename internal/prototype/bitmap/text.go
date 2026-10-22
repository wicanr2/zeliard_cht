package bitmap

import (
	"image"
	"image/color"
)

// TextLines保留指定字級，依字模的實際前進寬度換行。
// 半形字、全形字與缺字沿DrawGlyph使用同一個字距。
func (f Font) TextLines(text string, size, width int) []string {
	if size <= 0 || width <= 0 || text == "" {
		return nil
	}
	var lines []string
	line := []rune{}
	x := 0
	for _, ch := range text {
		if ch == '\n' {
			lines = append(lines, string(line))
			line, x = nil, 0
			continue
		}
		ch = f.textRune(ch)
		w := len(f[ch]) / 16 * 8 * size / 16
		if len(line) > 0 && x+w > width {
			lines = append(lines, string(line))
			line, x = nil, 0
		}
		line = append(line, ch)
		x += w
	}
	return append(lines, string(line))
}

func (f Font) textRune(ch rune) rune {
	if len(f[ch]) == 0 {
		return '?'
	}
	return ch
}

// DrawText直接使用目標字級的灰階遮罩；空間不足時省略並回傳false。
// 回傳值讓設定頁能檢查固定說明是否被裁切，不把縮字當作排版修復。
func (f Font) DrawText(dst *image.RGBA, box image.Rectangle, text string, size int, ink color.RGBA) bool {
	if text == "" {
		return true
	}
	box = box.Intersect(dst.Bounds())
	lines := f.TextLines(text, size, box.Dx())
	if size <= 0 || box.Dy() < size || len(lines) == 0 {
		return false
	}
	lineHeight := size + 4
	capacity := (box.Dy() + 4) / lineHeight
	fits := len(lines) <= capacity
	if !fits {
		lines = lines[:capacity]
		last := []rune(lines[len(lines)-1])
		mark := f.textRune('…')
		for len(last) > 0 && f.textWidth(string(last)+string(mark), size) > box.Dx() {
			last = last[:len(last)-1]
		}
		lines[len(lines)-1] = string(last) + string(mark)
	}
	y := box.Min.Y + (box.Dy()-(len(lines)*lineHeight-4))/2
	for _, line := range lines {
		x := box.Min.X
		for _, ch := range line {
			w := len(f[ch]) / 16 * 8 * size / 16
			if x+w > box.Max.X {
				fits = false
			}
			f.DrawGlyph(dst, image.Pt(x, y), ch, size, ink, func(p image.Point) bool { return p.In(box) })
			x += w
		}
		y += lineHeight
	}
	return fits
}

func (f Font) textWidth(text string, size int) int {
	w := 0
	for _, ch := range text {
		w += len(f[f.textRune(ch)]) / 16 * 8 * size / 16
	}
	return w
}

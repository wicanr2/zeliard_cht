// Package bitmap 載入使用者提供的 Unifont HEX 字型供前端雛形顯示。
package bitmap

import (
	"bufio"
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"
	"strings"
)

type Font map[rune][]byte

func Load(path string) (Font, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	out := Font{}
	scanner := bufio.NewScanner(z)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), ":", 2)
		if len(parts) != 2 {
			continue
		}
		cp, e := strconv.ParseUint(parts[0], 16, 32)
		if e != nil {
			return nil, e
		}
		data, e := hex.DecodeString(parts[1])
		if e != nil {
			return nil, e
		}
		if len(data) == 16 || len(data) == 32 {
			out[rune(cp)] = data
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("字型沒有可用字模")
	}
	return out, nil
}

func (f Font) Draw(dst *image.RGBA, box image.Rectangle, text string, ink color.RGBA) {
	if box.Dy() < 16 {
		return
	}
	box = box.Intersect(dst.Bounds())
	x, y := box.Min.X, box.Min.Y+(box.Dy()-16)/2
	runes := []rune(text)
	for i, r := range runes {
		data := f[r]
		if len(data) == 0 {
			data = f['?']
		}
		width := len(data) / 16 * 8
		if width == 0 {
			continue
		}
		if x+width > box.Max.X {
			break
		}
		if i < len(runes)-1 && x+width+16 > box.Max.X {
			data = f['…']
			if len(data) == 0 {
				data = f['?']
			}
			width = len(data) / 16 * 8
			i = len(runes) - 1
		}
		if width == 0 || x+width > box.Max.X {
			break
		}
		for row := 0; row < 16; row++ {
			for col := 0; col < width; col++ {
				if data[row*(width/8)+col/8]&(0x80>>uint(col%8)) != 0 {
					dst.SetRGBA(x+col, y+row, ink)
				}
			}
		}
		x += width
		if i == len(runes)-1 {
			break
		}
	}
}

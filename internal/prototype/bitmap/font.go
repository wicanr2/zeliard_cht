// Package bitmap 載入固定字距的HEX字型與可選灰階圓體遮罩。
package bitmap

import (
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	if err := out.loadSmooth(path); err != nil {
		return nil, err
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
			r = '?'
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
			r = '…'
			data = f['…']
			if len(data) == 0 {
				r = '?'
				data = f['?']
			}
			width = len(data) / 16 * 8
			i = len(runes) - 1
		}
		if width == 0 || x+width > box.Max.X {
			break
		}
		f.DrawGlyph(dst, image.Pt(x, y), r, 16, ink, func(p image.Point) bool { return p.In(box) })
		x += width
		if i == len(runes)-1 {
			break
		}
	}
}

// 灰階遮罩存在同一份Font的負鍵；Unicode字元均為非負，排版仍只讀原鍵。
// 每個字依序保存16、18、24、32像素高度的遮罩，字寬沿用HEX字模。
func (f Font) loadSmooth(path string) error {
	file, err := os.Open(path + ".smooth.json.gz")
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()
	z, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer z.Close()
	var sidecar struct {
		Schema     int               `json:"schema"`
		FontSHA256 string            `json:"font_sha256"`
		Sizes      []int             `json:"sizes"`
		Glyphs     map[string][]byte `json:"glyphs"`
	}
	if err = json.NewDecoder(z).Decode(&sidecar); err != nil {
		return err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if sidecar.Schema != 1 || sidecar.FontSHA256 != fmt.Sprintf("%x", sha256.Sum256(raw)) || fmt.Sprint(sidecar.Sizes) != "[16 18 24 32]" || len(sidecar.Glyphs) > 65536 {
		return fmt.Errorf("圓體遮罩來源不符")
	}
	for key, mask := range sidecar.Glyphs {
		cp, err := strconv.ParseInt(key, 10, 32)
		if err != nil || cp < 0 || cp > 0x10ffff {
			return fmt.Errorf("圓體字碼不符")
		}
		width := len(f[rune(cp)]) / 16 * 8
		if (width != 8 && width != 16) || len(mask) != width*(16*16+18*18+24*24+32*32)/16 {
			return fmt.Errorf("圓體遮罩尺寸不符：%U", cp)
		}
		// 舊式消費端在記憶體使用同字寬圓體；基底HEX檔案維持原樣。
		glyph := make([]byte, width*2)
		for y := 0; y < 16; y++ {
			for x := 0; x < width; x++ {
				if mask[y*width+x] >= 128 {
					glyph[y*(width/8)+x/8] |= 128 >> uint(x%8)
				}
			}
		}
		f[rune(cp)] = glyph
		f[-rune(cp)-1] = mask
	}
	return nil
}

// Coverage返回正式顯示尺寸的灰階覆蓋率；沒有圓體時沿用原字模取樣。
func (f Font) Coverage(ch rune, x, y, size int) byte {
	data := f[ch]
	width := len(data) / 16 * 8
	if size <= 0 || width == 0 || x < 0 || y < 0 || x >= width*size/16 || y >= size {
		return 0
	}
	if mask := f[-ch-1]; len(mask) > 0 {
		offset := 0
		for _, h := range []int{16, 18, 24, 32} {
			if h == size {
				return mask[offset+y*(width*h/16)+x]
			}
			offset += width * h * h / 16
		}
	}
	sx, sy := x*16/size, y*16/size
	if data[sy*(width/8)+sx/8]&(128>>uint(sx%8)) != 0 {
		return 255
	}
	return 0
}

// DrawGlyph保留呼叫端的裁切與遮擋條件，直接在目標解析度合成灰階筆畫。
func (f Font) DrawGlyph(dst *image.RGBA, at image.Point, ch rune, size int, ink color.RGBA, visible func(image.Point) bool) {
	width := len(f[ch]) / 16 * 8 * size / 16
	for y := 0; y < size; y++ {
		for x := 0; x < width; x++ {
			p := at.Add(image.Pt(x, y))
			if !p.In(dst.Bounds()) || (visible != nil && !visible(p)) {
				continue
			}
			a := uint32(f.Coverage(ch, x, y, size)) * uint32(ink.A) / 255
			if a == 0 {
				continue
			}
			base := dst.RGBAAt(p.X, p.Y)
			dst.SetRGBA(p.X, p.Y, color.RGBA{uint8((uint32(ink.R)*a + uint32(base.R)*(255-a) + 127) / 255), uint8((uint32(ink.G)*a + uint32(base.G)*(255-a) + 127) / 255), uint8((uint32(ink.B)*a + uint32(base.B)*(255-a) + 127) / 255), uint8(a + uint32(base.A)*(255-a)/255)})
		}
	}
}

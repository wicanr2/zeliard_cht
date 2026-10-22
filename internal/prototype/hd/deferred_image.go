package hd

import (
	"image"
	"image/color"
	"log"
	"sync"
	"sync/atomic"
)

var deferredRegistered, deferredLoaded, deferredFailed atomic.Uint64

func DeferredImageStats() map[string]uint64 {
	return map[string]uint64{"registered": deferredRegistered.Load(), "loaded": deferredLoaded.Load(), "failed": deferredFailed.Load()}
}

// deferredImage只記錄已核結構的清單項；首次使用仍須核對圖檔SHA及尺寸。
// 錯誤只記一次，呼叫端必須用resolvedImage取得整張有效圖後才開始繪製。
type deferredImage struct {
	path, hash    string
	width, height int
	once          sync.Once
	image         image.Image
}

func (s *deferredImage) resolve() image.Image {
	s.once.Do(func() {
		im, err := checkedTerrainImage(s.path, s.hash, s.width, s.height)
		if err != nil {
			deferredFailed.Add(1)
			log.Printf("HD圖檔未載入，保留原貌：%s：%v", s.path, err)
			return
		}
		s.image = im
		deferredLoaded.Add(1)
	})
	return s.image
}

func (s *deferredImage) Bounds() image.Rectangle { return image.Rect(0, 0, s.width, s.height) }
func (s *deferredImage) ColorModel() color.Model { return color.NRGBAModel }
func (s *deferredImage) At(x, y int) color.Color {
	if im := s.resolve(); im != nil {
		return im.At(x, y)
	}
	return color.NRGBA{}
}

func deferredTerrainImage(path, hash string, width, height int, lazy bool) (image.Image, error) {
	if lazy {
		deferredRegistered.Add(1)
		return &deferredImage{path: path, hash: hash, width: width, height: height}, nil
	}
	return checkedTerrainImage(path, hash, width, height)
}

func resolvedImage(im image.Image) image.Image {
	if d, ok := im.(*deferredImage); ok {
		return d.resolve()
	}
	return im
}

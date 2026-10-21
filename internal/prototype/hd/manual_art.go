// 使用者指定的人工驗證圖組。只交換顯示，不寫入原版狀態，見規格050。
package hd

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
)

func (s *CaveObjects) LoadManual(dir string) error {
	if s == nil {
		return fmt.Errorf("須先載入洞窟物件圖組")
	}
	data, err := os.ReadFile(filepath.Join(dir, "manual-art.json"))
	if err != nil {
		return err
	}
	var manifest struct {
		Objects []struct {
			Key, File, SHA string
			Hash           string `json:"sha256"`
			Raster         []byte
		}
	}
	if err = json.Unmarshal(data, &manifest); err != nil {
		return err
	}
	if len(manifest.Objects) == 0 || len(manifest.Objects) > 8192 {
		return fmt.Errorf("人工圖組數量不符")
	}
	poses := make(map[string]*caveObjectPose, len(manifest.Objects))
	for n, r := range manifest.Objects {
		if len(r.Raster) != 256 || townBackgroundHash(r.Raster) != r.Key || filepath.Base(r.File) != r.File {
			return fmt.Errorf("人工圖組來源不符")
		}
		im, err := checkedTerrainImage(filepath.Join(dir, r.File), r.Hash, 48, 48)
		if err != nil {
			return err
		}
		pose := &caveObjectPose{number: 1000 + n, bounds: image.Rect(0, 0, 48, 48), pixels: make([]color.NRGBA, 48*48)}
		for y := 0; y < 48; y++ {
			for x := 0; x < 48; x++ {
				pose.pixels[y*48+x] = color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
			}
		}
		poses[r.Key] = pose
	}
	s.manual = poses
	return nil
}

func (s *CaveObjects) manualPose(object caveObject) *caveObjectPose {
	if s == nil || len(s.manual) == 0 {
		return nil
	}
	raw := object.Raster
	for i, v := range raw {
		if v == 255 {
			raw[i] = 0
		}
	}
	return s.manual[townBackgroundHash(raw[:])]
}

// Package overlay 驗證 psychic-war 式的顯示覆繪，不修改原版資料或記憶體。
package overlay

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/dosgolem/xlate"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

type Entry struct{ Key, Source, Translation string }

// NewIntro 只接受本輪原版收據。原型不內嵌原版像素，參考畫格由本地 probe 產生。
func NewIntro(reference, textPath string, font bitmap.Font) (*xlate.Layer, error) {
	raw, err := os.ReadFile(reference)
	if err != nil {
		return nil, err
	}
	h := sha256.Sum256(raw)
	if len(raw) != 320*200 || hex.EncodeToString(h[:]) != "de520c43e8b0098258cf5938f9afb42b60e7f04a8742d3eee5e897fcbcbcad5d" {
		return nil, fmt.Errorf("片頭原型參考畫格不符，不套用覆繪")
	}
	data, err := os.ReadFile(textPath)
	if err != nil {
		return nil, err
	}
	var e Entry
	if err = json.Unmarshal(data, &e); err != nil {
		return nil, err
	}
	if e.Key == "" || e.Source != "Two thousand years," || len([]rune(e.Translation)) != 5 {
		return nil, fmt.Errorf("片頭原型只審查過此句與五字版面")
	}
	face := &xlate.Font{Name: "external-unifont16", W: 16, H: 16, Glyphs: map[rune][]byte{}}
	for _, r := range []rune(e.Translation) {
		if len(font[r]) != 32 {
			return nil, fmt.Errorf("片頭字型缺字或字寬不符：%U", r)
		}
		face.Glyphs[r] = append([]byte(nil), font[r]...)
	}
	want := make([]byte, 160*16)
	for y := 0; y < 16; y++ {
		copy(want[y*160:(y+1)*160], raw[(136+y)*320+88:(136+y)*320+248])
	}
	layer := &xlate.Layer{W: 320, H: 200}
	layer.Watch(&xlate.Watcher{Key: e.Key, X: 88, Y: 136, W: 160, H: 16, Want: want, Make: func() []*xlate.Stamp {
		return []*xlate.Stamp{{Key: e.Key, X: 88, Y: 136, Cells: 10, CellW: 16, CellH: 16, Font: face, GlyphX: 8, GlyphY: 8, GlyphScale: 2, Text: []rune("  " + e.Translation), State: xlate.Pending}}
	}})
	return layer, nil
}

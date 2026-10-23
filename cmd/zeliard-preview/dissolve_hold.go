package main

import (
	"image"
	"image/draw"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/hd"
)

// dissolveHold 是轉場前最後一張完整合成畫面，以及當時的原版色號與色盤。
type dissolveHold struct {
	scene   string
	palette [256][3]uint8
	raw     []byte
	pixels  *image.RGBA
	// added 是溶解期間容許的新增寫入點數；region 內原本為0的點依該區剩餘比例溶解。
	added  int
	region image.Rectangle
	// exclude 內的清除不算轉場開始：房間人像動畫可能只清掉人像框內的點。開始後不再要求。
	exclude []image.Rectangle
	engaged bool
}

// introDissolveAdded：片頭故事被 Space 打斷時，最後一個打字字形約18點與清黑落在同一段。
const introDissolveAdded = 256

// introStoryRegion 是片頭故事文字列，roomTextBox 是房間對話框（與 overlay 的 dialogueBox 相同）。
var (
	introStoryRegion = image.Rect(0, 143, 320, 200)
	roomTextBox      = image.Rect(54, 98, 264, 150)
	dissolveArea     = image.Rect(0, barHeight, width, barHeight+200*scale)
)

func newDissolveHold(scene string, palette [256][3]uint8, raw []byte, canvas *image.RGBA, added int, region image.Rectangle, exclude []image.Rectangle) *dissolveHold {
	h := &dissolveHold{scene: scene, palette: palette, raw: append([]byte(nil), raw...), pixels: image.NewRGBA(dissolveArea), added: added, region: region, exclude: exclude}
	draw.Draw(h.pixels, dissolveArea, canvas, dissolveArea.Min, draw.Src)
	return h
}

// dissolveCell 是區域外計算剩餘比例的格子大小（原版像素）。
const dissolveCell = 8

// dissolve 在原版相對保存畫面只有清成0的改寫時，把保存的合成畫面畫到 canvas：原版已清掉的點
// 塗黑；原本就是0的點（黑底上的中文與高清細節）依剩餘比例以原版像素位置抖動同步溶解。
// region 內用整區比例（中文排版與原版英文位置不同），其餘以8×8格計算，原版整格為0的格子
// 沿用保存畫面，全部清完才塗黑；其餘點與容許的少量新增寫入沿用保存畫面。
// 沒有清除或新增過多時不畫，回報 false。remaining 是保存時非0、現在仍非0的點數。
func (h *dissolveHold) dissolve(canvas *image.RGBA, raw []byte) (applied bool, remaining int) {
	removed, outside, added := 0, 0, 0
	regionInked, regionLeft := 0, 0
	var cellInked, cellLeft [200 / dissolveCell][320 / dissolveCell]int
	for i, v := range raw {
		old := h.raw[i]
		p := image.Pt(i%320, i/320)
		switch {
		case v == old:
		case v == 0:
			removed++
			if !inAny(p, h.exclude) {
				outside++
			}
		default:
			added++
		}
		if old == 0 {
			continue
		}
		in := p.In(h.region)
		cellInked[p.Y/dissolveCell][p.X/dissolveCell]++
		if in {
			regionInked++
		}
		if v != 0 {
			remaining++
			cellLeft[p.Y/dissolveCell][p.X/dissolveCell]++
			if in {
				regionLeft++
			}
		}
	}
	// 清黑轉場以清除為主；捲動或動畫的新增與清除大致相當，不介入。
	if removed == 0 || added > h.added || added*2 > removed || !h.engaged && outside == 0 {
		return false, remaining
	}
	h.engaged = true
	regionRatio := 0.0
	if regionInked > 0 {
		regionRatio = float64(regionLeft) / float64(regionInked)
	}
	black := []byte{h.palette[0][0], h.palette[0][1], h.palette[0][2], 255}
	for y := 0; y < 200; y++ {
		for x := 0; x < 320; x++ {
			v, old := raw[y*320+x], h.raw[y*320+x]
			clean := old != 0 && v == 0 || remaining == 0
			if old == 0 && v == 0 && remaining > 0 {
				keep := 1.0
				switch inked := cellInked[y/dissolveCell][x/dissolveCell]; {
				case image.Pt(x, y).In(h.region):
					keep = regionRatio
				case inked > 0:
					keep = float64(cellLeft[y/dissolveCell][x/dissolveCell]) / float64(inked)
				}
				hash := uint32(x)*2654435761 ^ uint32(y)*40503
				clean = float64(hash%1024) >= keep*1024
			}
			for dy := 0; dy < scale; dy++ {
				at := canvas.PixOffset(x*scale, barHeight+y*scale+dy)
				if clean {
					for dx := 0; dx < scale; dx++ {
						copy(canvas.Pix[at+dx*4:at+dx*4+4], black)
					}
					continue
				}
				from := h.pixels.PixOffset(x*scale, barHeight+y*scale+dy)
				copy(canvas.Pix[at:at+scale*4], h.pixels.Pix[from:from+scale*4])
			}
		}
	}
	return true, remaining
}

// dissolveFrame 處理原版逐點清黑的轉場（規格066）：片頭被 Space 打斷，以及離開房間。
// 原版只把像素改成0時，覆繪層因畫面改變而撤除，人像退回原貌、壓下的英文被寫回；
// 改由保存的合成畫面同步溶解。片頭模式（FF77非0）容許少量新增寫入；房間只在房間模組
// 已載入、城鎮與洞窟畫面都不在時介入，且不容許任何新增寫入。出現其他改寫、換場景、
// 換色盤或清完就停止，回到一般合成。
func (g *game) dissolveFrame(scene string, raw []byte, w, h int, outdoor bool) {
	if w != 320 || h != 200 || len(raw) != 64000 {
		g.dissolveHeld = nil
		return
	}
	// 片頭捲字自行處理溶解（只清字層），不由這裡接手。
	intro := g.o.Byte(oracle.Far(0x1b1, 0xff77)) != 0 && !g.layers["text_scroll"]
	room := g.o.Byte(oracle.Far(0x1b1, 0xff77)) == 0 && !outdoor && g.inRoom()
	pal := g.o.Palette()
	if hold := g.dissolveHeld; hold != nil && (intro || room) && hold.scene == scene && hold.palette == pal {
		if applied, remaining := hold.dissolve(g.canvas, raw); applied {
			g.layers["dissolve_hold"] = true
			if remaining == 0 {
				g.dissolveHeld = nil
			}
			return
		}
	}
	g.dissolveHeld = nil
	drawn := false
	for _, on := range g.layers {
		drawn = drawn || on
	}
	switch {
	case intro && drawn:
		g.dissolveHeld = newDissolveHold(scene, pal, raw, g.canvas, introDissolveAdded, introStoryRegion, nil)
	case room && drawn:
		// 只排除人像框內部；原版清到人像外框時人像層已退回原貌，必須算轉場開始。
		var portraits []image.Rectangle
		for _, p := range hd.RoomPortraits {
			for _, box := range p.Boxes {
				portraits = append(portraits, box.Inset(1))
			}
		}
		g.dissolveHeld = newDissolveHold(scene, pal, raw, g.canvas, 0, roomTextBox, portraits)
	}
}

func inAny(p image.Point, rects []image.Rectangle) bool {
	for _, r := range rects {
		if p.In(r) {
			return true
		}
	}
	return false
}

// inRoom 回報 01B1:A000 是否載著房間模組（國王、小屋、商店、銀行、教堂、旅館、長老）。
func (g *game) inRoom() bool {
	for _, module := range roomModulePayloads {
		if g.roomModuleLoaded(module) {
			return true
		}
	}
	return false
}

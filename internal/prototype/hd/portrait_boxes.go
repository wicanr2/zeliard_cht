package hd

import (
	"image"
	"image/draw"
)

// PortraitBoxes 是固定姿勢肖像模組的原生視窗。外圈一像素是原版外框，內框為 HD 圖面；
// 前端以外框是否改變判斷肖像視窗仍在。規格066。
var PortraitBoxes = map[string]image.Rectangle{
	"king":      kingBox,
	"item_shop": itemShopBox,
	"bank":      bankBox,
	"church":    churchBox,
}

// RoomPortrait 是載在 01B1:A000 的房間模組與它的人像層。規格066：模組已載入而人像層
// 尚未比對到時（原版逐格揭開人像），前端在人像框內只顯示原版，不讓逐格覆繪拼成馬賽克。
type RoomPortrait struct {
	Module int // SAR2 payload 起點
	Layer  string
	Boxes  []image.Rectangle
}

var RoomPortraits = []RoomPortrait{
	{0x16e1a, "king", []image.Rectangle{kingBox}},
	{0x17817, "merchant", []image.Rectangle{merchantBox}},
	{0x19462, "bank", []image.Rectangle{bankBox}},
	{0x1a19e, "church", []image.Rectangle{churchBox}},
	{0x1a588, "item_shop", []image.Rectangle{itemShopBox}},
	{0x1bcc6, "sage", []image.Rectangle{sageBox, sageReturnBox}},
}

// framedLike 比對人像框外圈一像素的色號與 RGB 是否與參考姿態相同。
// 外圈是原版外框；相同表示人像視窗仍在，只是內部為未登錄的姿態或揭開中的畫面。
func framedLike(idx, col, refIdx, refCol []byte, box image.Rectangle) bool {
	w, h := box.Dx(), box.Dy()
	if len(idx) != w*h || len(refIdx) != w*h || len(col) != w*h*3 || len(refCol) != w*h*3 {
		return false
	}
	same := func(i int) bool {
		return idx[i] == refIdx[i] && col[i*3] == refCol[i*3] && col[i*3+1] == refCol[i*3+1] && col[i*3+2] == refCol[i*3+2]
	}
	for x := 0; x < w; x++ {
		if !same(x) || !same((h-1)*w+x) {
			return false
		}
	}
	for y := 1; y < h-1; y++ {
		if !same(y*w) || !same(y*w+w-1) {
			return false
		}
	}
	return true
}

// DrawDefault 在外框與參考姿態相同、內部卻不是已登錄姿態時，畫第一個 HD 姿態。
// 只能在呼叫端已確認國王模組載入時使用；同一外框不能用來辨認角色。規格066。
func (k *King) DrawDefault(dst *image.RGBA, indexed, rgb []byte, origin image.Point) bool {
	if k == nil || !framedLike(kingCrop(indexed, 1), kingCrop(rgb, 3), k.indices[0], k.rgb[0], kingBox) {
		return false
	}
	at := origin.Add(image.Pt(kingBox.Min.X*3, kingBox.Min.Y*3))
	draw.Draw(dst, k.art[0].Bounds().Inset(3).Add(at), k.art[0], image.Pt(3, 3), draw.Src)
	return true
}

// DrawDefault 見 King.DrawDefault；呼叫端須先確認銀行模組載入。
func (m *Bank) DrawDefault(dst *image.RGBA, indexed, rgb []byte, origin image.Point) bool {
	if m == nil || !framedLike(bankCrop(indexed, 1), bankCrop(rgb, 3), m.indices[0], m.rgb[0], bankBox) {
		return false
	}
	at := origin.Add(bankBox.Inset(1).Min.Mul(scale))
	draw.Draw(dst, m.art[0].Bounds().Add(at), m.art[0], image.Point{}, draw.Src)
	return true
}

// DrawDefault 見 King.DrawDefault；呼叫端須先確認道具店模組載入。
func (m *ItemShop) DrawDefault(dst *image.RGBA, indexed, rgb []byte, origin image.Point) bool {
	if m == nil || !framedLike(itemShopCrop(indexed, 1), itemShopCrop(rgb, 3), m.indices[0], m.rgb[0], itemShopBox) {
		return false
	}
	at := origin.Add(itemShopBox.Inset(1).Min.Mul(scale))
	draw.Draw(dst, m.art[0].Bounds().Add(at), m.art[0], image.Point{}, draw.Src)
	return true
}

// DrawDefault 見 King.DrawDefault；呼叫端須先確認教堂模組載入。
func (m *Church) DrawDefault(dst *image.RGBA, indexed, rgb []byte, origin image.Point) bool {
	if m == nil || !framedLike(churchCrop(indexed, 1), churchCrop(rgb, 3), m.indices[0], m.rgb[0], churchBox) {
		return false
	}
	at := origin.Add(churchBox.Inset(1).Min.Mul(scale))
	box := image.Rect(0, 0, 282, 186).Add(at)
	if !box.In(dst.Bounds()) {
		return false
	}
	draw.Draw(dst, box, m.art[0], image.Point{}, draw.Src)
	return true
}

// MerchantBox 是武器店完整姿態的原生視窗，供前端在原版重畫肖像時沿用上一張 HD。
var MerchantBox = merchantBox

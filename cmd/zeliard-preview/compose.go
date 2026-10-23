package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
	"github.com/wicanr2/zeliard_cht/internal/prototype/glyphpatch"
	"github.com/wicanr2/zeliard_cht/internal/prototype/hd"
)

// portraitHold 是同一場景最後一張已匹配的 HD 肖像。
type portraitHold struct {
	scene   string
	box     image.Rectangle
	ring    []byte
	palette [256][3]uint8
	pixels  *image.RGBA
}

// sceneKey 以影像模式與三個已載入模組前綴辨識場景；與 caveFamily 用同一組 bytes。
func (g *game) sceneKey() string {
	key := []byte{g.o.VideoMode()}
	key = append(key, g.o.Bytes(oracle.Far(0x1b1, 0xa000), 32)...)
	key = append(key, g.o.Bytes(oracle.Far(0x1b1, 0xc000), 32)...)
	key = append(key, g.o.Bytes(oracle.Far(0x11b1, 0x8000), 48)...)
	return fmt.Sprintf("%x", sha256.Sum256(key))[:12]
}

func portraitRing(raw []byte, box image.Rectangle) []byte {
	var ring []byte
	for x := box.Min.X; x < box.Max.X; x++ {
		ring = append(ring, raw[box.Min.Y*320+x], raw[(box.Max.Y-1)*320+x])
	}
	for y := box.Min.Y + 1; y < box.Max.Y-1; y++ {
		ring = append(ring, raw[y*320+box.Min.X], raw[y*320+box.Max.X-1])
	}
	return ring
}

// holdPortrait 在同一場景沿用上一張已匹配的 HD 肖像。原版說話等動畫有未登錄畫格時，
// 肖像不退回原貌；場景、色盤或肖像外框一變就停止。規格066。
func (g *game) holdPortrait(name string, matched bool, raw []byte, scene string) bool {
	box, ok := hd.PortraitBoxes[name]
	if !ok {
		return matched
	}
	return g.holdPortraitBox(name, box, matched, raw, scene)
}

// holdPortraitBox 同 holdPortrait，肖像視窗由呼叫端給定；未匹配時沿用上次匹配的視窗。
// 長老完整手勢依原版畫格位置選視窗，原版重畫肖像的單一合成點也維持 HD。
func (g *game) holdPortraitBox(name string, box image.Rectangle, matched bool, raw []byte, scene string) bool {
	if len(raw) != 64000 {
		return matched
	}
	if h := g.portraits[name]; !matched && h != nil {
		box = h.box
	}
	if box.Empty() || !box.In(image.Rect(0, 0, 320, 200)) {
		return matched
	}
	inner := image.Rectangle{box.Inset(1).Min.Mul(scale), box.Inset(1).Max.Mul(scale)}.Add(image.Pt(0, barHeight))
	ring, pal := portraitRing(raw, box), g.o.Palette()
	if matched {
		h := &portraitHold{scene: scene, box: box, ring: ring, palette: pal, pixels: image.NewRGBA(inner)}
		draw.Draw(h.pixels, inner, g.canvas, inner.Min, draw.Src)
		g.portraits[name] = h
		return true
	}
	h := g.portraits[name]
	if h == nil || h.scene != scene || h.palette != pal || !bytes.Equal(h.ring, ring) {
		delete(g.portraits, name)
		return false
	}
	draw.Draw(g.canvas, inner, h.pixels, inner.Min, draw.Src)
	g.layers[name+"_held"] = true
	return true
}

// compose 只合成遊戲畫面，不碰 ebiten；GUI 的 Draw 與無頭閃爍探針共用同一路徑。
// layers 記錄本次各 HD 層與中文層是否實際覆繪，供閃爍判定。
func (g *game) compose() (presented *hd.TownPresentation, cavePresented *hd.CavePresentation, w, h int, raw, rawRGB []byte) {
	g.layers = map[string]bool{}
	if g.portraits == nil {
		g.portraits = map[string]*portraitHold{}
	}
	scene := g.sceneKey()
	g.manualArtwork.BeginFrame()
	w, h, rawRGB = glyphpatch.ScreenRGB(g.o)
	raw = glyphpatch.Indexed(g.o)
	rgb, townRaw := rawRGB, raw
	if g.townFrames != nil {
		blocked := g.townDialogue != nil && g.townDialogue.State()["key"] != ""
		if blocked {
			presented = g.townFrames.PresentDialogue(raw, rawRGB, g.townDialogue.BackdropBox())
		} else {
			presented = g.townFrames.Present(raw, rawRGB, false)
		}
		if presented == nil {
			// 規格067：F7 讀檔選單暫時取代城鎮模組，沿用最後一張已驗城鎮畫面，選單框內放回原版。
			presented = g.townFrames.PresentMenu(raw, rawRGB)
			g.layers["town_menu_hold"] = presented != nil
		}
		if presented != nil {
			rgb, townRaw = presented.RGB, presented.Indexed
		}
	}
	g.layers["town_frames"] = presented != nil
	if g.caveFrames != nil {
		blocked := g.townDialogue != nil && g.townDialogue.State()["key"] != ""
		cavePresented = g.caveFrames.Present(raw, rawRGB, blocked)
		g.layers["cave_menu_hold"] = cavePresented != nil && !cavePresented.MenuBox.Empty()
		if cavePresented != nil && presented == nil {
			rgb, townRaw = cavePresented.RGB, cavePresented.Indexed
		}
	}
	g.layers["cave_frames"] = cavePresented != nil

	draw.Draw(g.canvas, g.canvas.Bounds(), &image.Uniform{color.RGBA{18, 23, 32, 255}}, image.Point{}, draw.Src)
	if w == 320 && h == 200 && len(rgb) == 320*200*3 {
		bitmap.DrawRGB3(g.canvas, rgb, image.Pt(0, barHeight))
	}
	g.hdMatched = false
	if g.hdEnabled && g.hd != nil && w == 320 && h == 200 {
		g.hdMatched = g.hd.Draw(g.canvas, raw, rawRGB, image.Pt(0, barHeight))
		g.layers["hd_frame"] = g.hdMatched
	}
	g.townSkyHDMatched = false
	if g.hdEnabled && g.townSkyHD != nil && w == 320 && h == 200 {
		g.townSkyHDMatched = g.townSkyHD.Draw(g.canvas, g.o.VideoMode(), g.o.Bytes(oracle.Far(0x1b1, 0xc000), 64), raw, rawRGB, image.Pt(0, barHeight))
		g.hdMatched = g.hdMatched || g.townSkyHDMatched
		g.layers["town_sky"] = g.townSkyHDMatched
	}
	g.townBackgroundMatch = false
	var townPrimary, townResources, townExtra, townPalette []byte
	if g.townBackgroundHD != nil || g.townSceneryHD != nil || g.worldTowns != nil || g.manualArtwork != nil {
		townPrimary = g.o.Bytes(oracle.Far(0x1b1, 0), 65536)
		townResources = g.o.Bytes(oracle.Far(0x11b1, 0), 65536)
		townExtra = g.o.Bytes(oracle.Far(0x21b1, 0), 65536)
		if presented != nil {
			townPrimary, townResources, townExtra = presented.Primary, presented.Resources, presented.Extra
		}
		if cavePresented != nil && presented == nil {
			townPrimary = cavePresented.Memory[0x1b10:0x11b10]
			townResources = cavePresented.Memory[0x11b10:0x21b10]
			townExtra = cavePresented.Memory[0x21b10:0x31b10]
		}
	}
	if g.townGroundHD != nil || g.townPlayerHD != nil {
		if presented != nil {
			townPalette = presented.Palette
		} else {
			for _, c := range g.o.Palette() {
				townPalette = append(townPalette, c[:]...)
			}
		}
	}
	if g.hdEnabled && g.townBackgroundHD != nil && w == 320 && h == 200 {
		g.townBackgroundMatch = g.townBackgroundHD.Draw(g.canvas, g.o.VideoMode(), townPrimary, townResources, townRaw, rgb, image.Pt(0, barHeight))
		g.hdMatched = g.hdMatched || g.townBackgroundMatch
		g.layers["town_background"] = g.townBackgroundMatch
	}
	g.townSceneryMatch = false
	if g.hdEnabled && g.worldTowns != nil && w == 320 && h == 200 {
		matched := g.worldTowns.Draw(g.canvas, g.o.VideoMode(), townPrimary, townResources, townRaw, rgb, image.Pt(0, barHeight))
		g.hdMatched = matched || g.hdMatched
		g.layers["world_towns"] = matched
	}
	if g.hdEnabled && g.townSceneryHD != nil && w == 320 && h == 200 {
		g.townSceneryMatch = g.townSceneryHD.Draw(g.canvas, g.o.VideoMode(), townPrimary, townResources, townExtra, townRaw, rgb, image.Pt(0, barHeight), true)
		g.hdMatched = g.hdMatched || g.townSceneryMatch
		g.layers["town_scenery"] = g.townSceneryMatch
	}
	g.townGroundMatch = false
	if g.hdEnabled && g.townGroundHD != nil && w == 320 && h == 200 {
		g.townGroundMatch = g.townGroundHD.Draw(g.canvas, g.o.VideoMode(), townPrimary, townResources, townExtra, townRaw, rgb, townPalette, image.Pt(0, barHeight))
		g.hdMatched = g.hdMatched || g.townGroundMatch
		g.layers["town_ground"] = g.townGroundMatch
	}
	g.townActorsMatch = false
	g.townActorsHD.Clear()
	if g.hdEnabled && g.townActorsHD != nil && g.townSceneryMatch && w == 320 && h == 200 {
		g.townActorsMatch = g.townActorsHD.Draw(g.canvas, g.o.VideoMode(), townPrimary, townResources, townExtra, townRaw, rgb, townPalette, image.Pt(0, barHeight))
		g.layers["town_actors"] = g.townActorsMatch
	}
	g.townPlayerMatch = g.townActorsMatch
	if g.hdEnabled && !g.townActorsMatch && g.townPlayerHD != nil && g.townSceneryMatch && w == 320 && h == 200 {
		g.townPlayerMatch = g.townPlayerHD.Draw(g.canvas, g.o.VideoMode(), townPrimary, townResources, townExtra, townRaw, rgb, townPalette, image.Pt(0, barHeight))
		g.hdMatched = g.hdMatched || g.townPlayerMatch
		g.layers["town_player"] = g.townPlayerMatch
	}
	g.kingHDMatched, g.kingHDPose = false, -1
	g.caveTerrainMatch = false
	if g.hdEnabled && g.caveTerrainHD != nil && cavePresented != nil && w == 320 && h == 200 {
		g.caveTerrainMatch = g.caveTerrainHD.Draw(g.canvas, cavePresented, image.Pt(0, barHeight))
		g.hdMatched = g.hdMatched || g.caveTerrainMatch
		g.layers["cave_terrain"] = g.caveTerrainMatch
	}
	if g.hdEnabled && g.kingHD != nil && w == 320 && h == 200 {
		g.kingHDMatched, g.kingHDPose = g.kingHD.Draw(g.canvas, raw, rawRGB, image.Pt(0, barHeight))
		g.layers["king"] = g.holdPortrait("king", g.kingHDMatched, raw, scene)
		g.hdMatched = g.hdMatched || g.layers["king"]
	}
	g.sageHDMatched, g.sageHDPose = false, -1
	g.sageFramesHD.Clear()
	if g.hdEnabled && g.sageFramesHD != nil && w == 320 && h == 200 {
		p := g.o.Bytes(oracle.Far(0x1b1, 0), 65536)
		r := g.o.Bytes(oracle.Far(0x11b1, 0), 65536)
		var palette []byte
		for _, color := range g.o.Palette() {
			palette = append(palette, color[:]...)
		}
		g.sageHDMatched = g.sageFramesHD.Draw(g.canvas, g.o.VideoMode(), p, r, raw, rawRGB, palette, image.Pt(0, barHeight))
		if g.sageHDMatched {
			g.sageHDPose = g.sageFramesHD.State()["pose"].(int)
		}
		var box image.Rectangle
		if origin, ok := g.sageFramesHD.State()["native_origin"].([]int); ok && len(origin) == 2 {
			box = image.Rect(origin[0], origin[1], origin[0]+96, origin[1]+64)
		}
		g.layers["sage"] = g.holdPortraitBox("sage", box, g.sageHDMatched, raw, scene)
		g.hdMatched = g.hdMatched || g.layers["sage"]
	} else if g.hdEnabled && g.sageHD != nil && w == 320 && h == 200 {
		g.sageHDMatched, g.sageHDPose = g.sageHD.Draw(g.canvas, raw, rawRGB, image.Pt(0, barHeight))
		g.hdMatched = g.hdMatched || g.sageHDMatched
		g.layers["sage"] = g.sageHDMatched
	}
	g.merchantHDMatched, g.merchantHDPose = false, -1
	g.merchantFramesHD.Clear()
	if g.hdEnabled && g.merchantFramesHD != nil && w == 320 && h == 200 {
		p := g.o.Bytes(oracle.Far(0x1b1, 0), 65536)
		r := g.o.Bytes(oracle.Far(0x11b1, 0), 65536)
		var palette []byte
		for _, color := range g.o.Palette() {
			palette = append(palette, color[:]...)
		}
		g.merchantHDMatched = g.merchantFramesHD.Draw(g.canvas, g.o.VideoMode(), p, r, raw, rawRGB, palette, image.Pt(0, barHeight))
		if g.merchantHDMatched {
			g.merchantHDPose = g.merchantFramesHD.State()["pose"].(int)
		}
		g.layers["merchant"] = g.holdPortraitBox("merchant", hd.MerchantBox, g.merchantHDMatched, raw, scene)
		g.hdMatched = g.hdMatched || g.layers["merchant"]
	} else if g.hdEnabled && g.merchantHD != nil && w == 320 && h == 200 {
		g.merchantHDMatched, g.merchantHDPose = g.merchantHD.Draw(g.canvas, raw, rawRGB, image.Pt(0, barHeight))
		g.hdMatched = g.hdMatched || g.merchantHDMatched
		g.layers["merchant"] = g.merchantHDMatched
	}
	g.bankHDMatched, g.bankHDPose = false, -1
	if g.hdEnabled && g.bankHD != nil && w == 320 && h == 200 {
		g.bankHDMatched, g.bankHDPose = g.bankHD.Draw(g.canvas, raw, rawRGB, image.Pt(0, barHeight))
		g.layers["bank"] = g.holdPortrait("bank", g.bankHDMatched, raw, scene)
		g.hdMatched = g.hdMatched || g.layers["bank"]
	}
	g.itemShopHDMatched, g.itemShopHDPose = false, -1
	if g.hdEnabled && g.itemShopHD != nil && w == 320 && h == 200 {
		g.itemShopHDMatched, g.itemShopHDPose = g.itemShopHD.Draw(g.canvas, raw, rawRGB, image.Pt(0, barHeight))
		g.layers["item_shop"] = g.holdPortrait("item_shop", g.itemShopHDMatched, raw, scene)
		g.hdMatched = g.hdMatched || g.layers["item_shop"]
	}
	g.churchHDMatched, g.churchHDFrame = false, -1
	if g.hdEnabled && g.churchHD != nil && w == 320 && h == 200 {
		g.churchHDMatched, g.churchHDFrame = g.churchHD.Draw(g.canvas, raw, rawRGB, image.Pt(0, barHeight))
		g.layers["church"] = g.holdPortrait("church", g.churchHDMatched, raw, scene)
		g.hdMatched = g.hdMatched || g.layers["church"]
	}
	g.interiorHDMatch = false
	g.interiorHD.Clear()
	if g.hdEnabled && g.interiorHD != nil && w == 320 && h == 200 {
		var palette []byte
		for _, color := range g.o.Palette() {
			palette = append(palette, color[:]...)
		}
		g.interiorHDMatch = g.interiorHD.Draw(g.canvas, g.o.VideoMode(), g.o.Bytes(oracle.Far(0x1b1, 0), 65536), g.o.Bytes(oracle.Far(0x11b1, 0), 65536), raw, rawRGB, palette, image.Pt(0, barHeight))
		g.hdMatched = g.hdMatched || g.interiorHDMatch
		g.layers["interior"] = g.interiorHDMatch
	}
	if g.hdEnabled && g.manualArtwork != nil && w == 320 && h == 200 {
		transition := g.manualArtwork.DrawCaveTransition(g.canvas, raw, rawRGB, image.Pt(0, barHeight))
		g.manualArtwork.SetResources(townResources, townExtra)
		g.manualArtwork.SetIntroMode(g.o.Byte(oracle.Far(0x1b1, 0xff77)) != 0)
		manual := g.manualArtwork.Draw(g.canvas, townRaw, rgb, image.Pt(0, barHeight))
		coherent := g.manualArtwork.DrawCoherent(g.canvas, townRaw, rgb, image.Pt(0, barHeight))
		g.hdMatched = transition || manual || coherent || g.hdMatched
		g.layers["cave_transition"], g.layers["manual_art"], g.layers["coherent_art"] = transition, manual, coherent
		g.layers["manual_art_held"] = g.manualArtwork.CompleteHeld()
		// 規格067：結局對話的肖像框沒有整張肖像吻合（含說話沿用）時，整框顯示原版，
		// 不讓其他圖組的同色小格拼成 HD 與原版交錯的格子。
		if g.endingLoaded() {
			for _, p := range endingPortraits {
				if endingFramed(raw, p.frame) && !g.manualArtwork.LargeCovers(p.portrait) {
					restoreNativeBox(g.canvas, rawRGB, p.portrait)
					g.layers["ending_portrait_native"] = true
				}
			}
		}
	}
	// 規格066：房間模組已載入而人像層尚未比對到時（進房間的未登錄姿態或揭開中），外框仍在就畫
	// 第一個 HD 姿態；外框不在才只顯示原版。不讓逐格覆繪拼出 HD 與原版交錯的馬賽克。
	// 城鎮或洞窟畫面仍在時模組可能已先載入，不套用。
	if g.hdEnabled && presented == nil && cavePresented == nil && w == 320 && h == 200 {
		for _, p := range hd.RoomPortraits {
			if g.layers[p.Layer] || !g.roomModuleLoaded(p.Module) {
				continue
			}
			if paint := g.portraitDefault(p.Layer); paint != nil && paint(g.canvas, raw, rawRGB, image.Pt(0, barHeight)) {
				g.layers[p.Layer], g.layers[p.Layer+"_default"] = true, true
				g.hdMatched = true
				continue
			}
			for _, box := range p.Boxes {
				restoreNativeBox(g.canvas, rawRGB, box.Inset(1))
			}
			g.layers[p.Layer+"_native"] = true
		}
	}
	// 規格067：原版第一次畫一個城鎮場景（如開局進城堡）時，畫完前 HD 層比對不到；HD 模式沿用上一張
	// 合成的城鎮畫帶，到完成點才換成 HD，不顯示原版逐欄畫出的過程。
	holding := g.hdEnabled && presented == nil && cavePresented == nil && g.townHeld != nil && w == 320 && h == 200 && g.townFrames.FirstDraw(g.o.Steps())
	if holding {
		draw.Draw(g.canvas, g.townHeld.Bounds(), g.townHeld, g.townHeld.Bounds().Min, draw.Src)
		g.layers["town_first_draw_hold"] = true
	}
	if presented != nil && !presented.DialogueBox.Empty() {
		restoreChangedRows(g.canvas, raw, presented.SourceIndexed, rawRGB, presented.DialogueBox)
	}
	if cavePresented != nil && presented == nil && !cavePresented.MenuBox.Empty() {
		restoreChangedRows(g.canvas, raw, cavePresented.SourceIndexed, rawRGB, cavePresented.MenuBox)
	}
	if g.beforeText != nil {
		g.beforeText()
	}
	g.hideClaimedGlyphs(w, h, rawRGB)
	if g.intro != nil && w == 320 && h == 200 {
		g.intro.Frame(raw, rawRGB)
		clear(g.overlayPixels)
		if g.intro.Draw(g.overlayPixels, scale, nil) {
			src := &image.RGBA{Pix: g.overlayPixels, Stride: width * 4, Rect: image.Rect(0, 0, width, 600)}
			draw.Draw(g.canvas, image.Rect(0, barHeight, width, barHeight+600), src, image.Point{}, draw.Over)
			g.layers["text_intro"] = true
		}
	}
	if g.scroll != nil && w == 320 && h == 200 {
		g.layers["text_scroll"] = g.scroll.Draw(g.canvas, raw, g.o.Palette(), image.Pt(0, barHeight))
	}
	if g.introThreat != nil && w == 320 && h == 200 {
		g.layers["text_intro_threat"] = g.introThreat.Draw(g.canvas, raw, g.o.Palette(), image.Pt(0, barHeight))
	}
	if g.introStory != nil && w == 320 && h == 200 {
		g.layers["text_intro_story"] = g.introStory.Draw(g.canvas, raw, g.o.Palette(), image.Pt(0, barHeight))
	}
	if g.endingStory != nil && w == 320 && h == 200 {
		g.layers["text_ending_story"] = g.endingStory.Draw(g.canvas, raw, g.o.Palette(), image.Pt(0, barHeight))
	}
	if g.endingCredits != nil && w == 320 && h == 200 {
		g.layers["text_ending_credits"] = g.endingCredits.Draw(g.canvas, raw, g.o.Palette(), image.Pt(0, barHeight))
	}
	if g.dialogue != nil && w == 320 && h == 200 {
		g.layers["text_dialogue"] = g.dialogue.Draw(g.canvas, raw, g.o.Palette(), image.Pt(0, barHeight))
	}
	if g.menuLabels != nil && w == 320 && h == 200 {
		g.menuLabels.ObserveFrame(raw, g.o.Palette())
		g.layers["text_labels"] = g.menuLabels.Draw(g.canvas, raw, g.o.Palette(), image.Pt(0, barHeight)) > 0
	}
	if g.hud != nil && w == 320 && h == 200 {
		g.layers["text_hud"] = g.hud.Draw(g.canvas, raw, g.o.Palette(), image.Pt(0, barHeight)) > 0
	}
	if g.townDialogue != nil && w == 320 && h == 200 {
		g.layers["text_town"] = g.townDialogue.Draw(g.canvas, raw, g.o.Palette(), image.Pt(0, barHeight))
	}
	g.dissolveFrame(scene, raw, w, h, presented != nil || cavePresented != nil)
	presented.ApplyErase(g.canvas, image.Pt(0, barHeight))
	if !holding {
		box := image.Rect(townBox.Min.X*scale, barHeight+townBox.Min.Y*scale, townBox.Max.X*scale, barHeight+townBox.Max.Y*scale)
		if g.townHeld == nil {
			g.townHeld = image.NewRGBA(box)
		}
		draw.Draw(g.townHeld, box, g.canvas, box.Min, draw.Src)
	}
	return
}

// townBox 是城鎮畫帶（與 hd.TownFrames 的呈現範圍相同）。
var townBox = image.Rect(48, 24, 272, 158)

// hideClaimedGlyphs 在中文層之前清掉底圖上被壓下的原版字形。字形只存在守衛輸入；
// 底圖某點若整個 3×3 區塊仍是守衛色，換回 VRAM 實際的像素。中文層之後才畫，筆畫不受影響。
func (g *game) hideClaimedGlyphs(w, h int, guardRGB []byte) {
	gate := glyphpatch.For(g.o)
	if gate == nil || w != 320 || h != 200 {
		return
	}
	_, _, shown := g.o.ScreenRGB()
	gate.Hidden(func(p int, _ uint8) {
		x, y := p%320*3, barHeight+p/320*3
		if !blockIs(g.canvas, x, y, guardRGB[p*3:p*3+3]) {
			return
		}
		for dy := 0; dy < 3; dy++ {
			for dx := 0; dx < 3; dx++ {
				i := g.canvas.PixOffset(x+dx, y+dy)
				copy(g.canvas.Pix[i:i+3], shown[p*3:p*3+3])
			}
		}
	})
}

// blockIs 回報畫布上以 (x, y) 為左上角的 3×3 區塊是否全為指定顏色。
func blockIs(canvas *image.RGBA, x, y int, c []byte) bool {
	for dy := 0; dy < 3; dy++ {
		for dx := 0; dx < 3; dx++ {
			i := canvas.PixOffset(x+dx, y+dy)
			if canvas.Pix[i] != c[0] || canvas.Pix[i+1] != c[1] || canvas.Pix[i+2] != c[2] {
				return false
			}
		}
	}
	return true
}

// portraitDefault 回傳人像模組的預設姿態繪製；尚未載入或沒有預設姿態的模組回傳 nil。
func (g *game) portraitDefault(layer string) func(*image.RGBA, []byte, []byte, image.Point) bool {
	switch {
	case layer == "king" && g.kingHD != nil:
		return g.kingHD.DrawDefault
	case layer == "bank" && g.bankHD != nil:
		return g.bankHD.DrawDefault
	case layer == "item_shop" && g.itemShopHD != nil:
		return g.itemShopHD.DrawDefault
	case layer == "church" && g.churchHD != nil:
		return g.churchHD.DrawDefault
	}
	return nil
}

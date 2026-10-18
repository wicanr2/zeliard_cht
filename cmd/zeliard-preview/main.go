// zeliard-preview 直接執行原版的可撤回前端。規格見 docs/spec/001-runtime-overlay.md。
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/dosgolem/xlate"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
	"github.com/wicanr2/zeliard_cht/internal/prototype/controls"
	"github.com/wicanr2/zeliard_cht/internal/prototype/hd"
	"github.com/wicanr2/zeliard_cht/internal/prototype/overlay"
)

const width, height, scale, barHeight = 960, 680, 3, 48

type device struct {
	id        ebiten.GamepadID
	key, name string
}
type game struct {
	o                    *oracle.Oracle
	audio                *oracle.Audio
	font                 bitmap.Font
	labels               map[string]string
	settings             controls.Settings
	settingsPath         string
	selected             ebiten.GamepadID
	devices              []device
	edges                controls.Edges
	panel, paused        bool
	focus                int
	listening            string
	neutral              []float64
	previousButtons      []bool
	previousActions      map[string]bool
	notice               string
	cycles               uint64
	frames, captureFrame int
	capturePath          string
	captureRequest       string
	observePath          string
	captured             bool
	captureTail          int
	intro                *xlate.Layer
	scroll               *overlay.Scroll
	dialogue             *overlay.Dialogue
	townDialogue         *overlay.TownDialogue
	menuLabels           *overlay.Labels
	hud                  *overlay.HUD
	hd                   *hd.Frame
	townSkyHD            *hd.TownSky
	townSkyHDMatched     bool
	townBackgroundHD     *hd.TownBackground
	townBackgroundMatch  bool
	townSceneryHD        *hd.TownScenery
	townSceneryMatch     bool
	townFrames           *hd.TownFrames
	kingHD               *hd.King
	kingHDMatched        bool
	kingHDPose           int
	sageHD               *hd.Sage
	sageHDMatched        bool
	sageHDPose           int
	merchantHD           *hd.Merchant
	merchantHDMatched    bool
	merchantHDPose       int
	bankHD               *hd.Bank
	bankHDMatched        bool
	bankHDPose           int
	itemShopHD           *hd.ItemShop
	itemShopHDMatched    bool
	itemShopHDPose       int
	hdEnabled, hdMatched bool
	overlayPixels        []byte
	texture              *ebiten.Image
	canvas               *image.RGBA
}

func (g *game) label(key string) string {
	if s := g.labels[key]; s != "" {
		return s
	}
	return key
}
func (g *game) save() {
	if err := g.settings.Save(g.settingsPath); err != nil {
		g.notice = err.Error()
	} else {
		g.notice = g.label("saved")
	}
}
func (g *game) profile() map[string]controls.Binding {
	p := g.settings.Profiles[g.settings.Selected]
	if p == nil {
		p = map[string]controls.Binding{}
		g.settings.Profiles[g.settings.Selected] = p
	}
	return p
}

func (g *game) scanDevices() {
	g.devices = nil
	connected := false
	matches := []ebiten.GamepadID{}
	for _, id := range ebiten.AppendGamepadIDs(nil) {
		name := ebiten.GamepadName(id)
		key := ebiten.GamepadSDLID(id) + "/" + name
		g.devices = append(g.devices, device{id, key, name})
		if id == g.selected && key == g.settings.Selected {
			connected = true
		}
		if key == g.settings.Selected {
			matches = append(matches, id)
		}
	}
	if !connected && g.selected >= 0 {
		g.edges.Sync(g.o, nil)
		g.selected = -1
		g.previousButtons = nil
		g.listening = ""
		g.notice = g.label("disconnected")
	}
	if g.selected < 0 && len(matches) == 1 {
		g.selected = matches[0]
		g.previousButtons = nil
	}
}

func (g *game) nextDevice() {
	if len(g.devices) == 0 {
		return
	}
	idx := -1
	for i, d := range g.devices {
		if d.id == g.selected {
			idx = i
		}
	}
	d := g.devices[(idx+1)%len(g.devices)]
	g.edges.Sync(g.o, nil)
	g.selected = d.id
	g.settings.Selected = d.key
	g.previousButtons = nil
	g.listening = ""
	g.save()
}

func (g *game) samples() ([]bool, []float64) {
	if g.selected < 0 {
		return nil, nil
	}
	buttons := make([]bool, ebiten.GamepadButtonCount(g.selected))
	axes := make([]float64, ebiten.GamepadAxisCount(g.selected))
	for i := range buttons {
		buttons[i] = ebiten.IsGamepadButtonPressed(g.selected, ebiten.GamepadButton(i))
	}
	for i := range axes {
		axes[i] = ebiten.GamepadAxisValue(g.selected, i)
	}
	return buttons, axes
}

func (g *game) beginBind(action string, axes []float64) {
	if g.selected < 0 {
		g.notice = g.label("choose")
		return
	}
	g.listening = action
	g.neutral = append([]float64(nil), axes...)
	g.notice = g.label("listening")
}

func actionRect(i int) image.Rectangle {
	col, row := i/6, i%6
	return image.Rect(60+col*430, 270+row*36, 465+col*430, 302+row*36)
}
func (g *game) captureBinding(buttons []bool, axes []float64) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.listening = ""
		g.notice = ""
		return
	}
	var binding controls.Binding
	found := false
	for i, down := range buttons {
		if down && (i >= len(g.previousButtons) || !g.previousButtons[i]) {
			binding = controls.Binding{Kind: "button", Index: i}
			found = true
			break
		}
	}
	if !found {
		for i, value := range axes {
			if i < len(g.neutral) {
				delta := value - g.neutral[i]
				if delta > .6 || delta < -.6 {
					sign := 1
					if delta < 0 {
						sign = -1
					}
					binding = controls.Binding{Kind: "axis", Index: i, Sign: sign, Neutral: g.neutral[i]}
					found = true
					break
				}
			}
		}
	}
	if found {
		g.profile()[g.listening] = binding
		g.listening = ""
		g.save()
	}
}

func (g *game) Update() error {
	if g.captured {
		g.captureTail++
		if g.captureTail >= 60 {
			return ebiten.Termination
		}
		return nil
	}
	g.frames++
	g.scanDevices()
	buttons, axes := g.samples()
	actions := map[string]bool{}
	if g.selected >= 0 {
		for action, b := range g.profile() {
			actions[action] = b.Active(buttons, axes, g.settings.Deadzone)
		}
	}
	defer func() { g.previousButtons = buttons; g.previousActions = actions }()
	if !ebiten.IsFocused() {
		g.edges.Sync(g.o, nil)
		return nil
	}
	just := func(a string) bool { return actions[a] && !g.previousActions[a] }
	if !g.panel && just("toolbar") {
		g.panel = true
		g.edges.Sync(g.o, nil)
	}
	x, y := ebiten.CursorPosition()
	point := image.Pt(x, y)
	click := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if click && point.In(image.Rect(8, 8, 150, 40)) {
		g.hdEnabled = false
	}
	if click && point.In(image.Rect(158, 8, 340, 40)) && (g.hd != nil || g.townSkyHD != nil || g.townBackgroundHD != nil || g.townSceneryHD != nil || g.kingHD != nil || g.sageHD != nil || g.merchantHD != nil || g.bankHD != nil || g.itemShopHD != nil) {
		g.hdEnabled = !g.hdEnabled
	}
	if click && image.Pt(x, y).In(image.Rect(348, 8, 540, 40)) {
		g.panel = !g.panel
		g.listening = ""
		g.edges.Sync(g.o, nil)
	}
	if click && point.In(image.Rect(548, 8, 740, 40)) {
		g.paused = !g.paused
		g.edges.Sync(g.o, nil)
	}
	if g.panel {
		g.edges.Sync(g.o, nil)
		if g.listening != "" {
			g.captureBinding(buttons, axes)
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || click && point.In(image.Rect(740, 560, 900, 600)) {
			g.panel = false
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyTab) || click && point.In(image.Rect(60, 135, 900, 175)) {
			g.nextDevice()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || just("down") {
			g.focus = (g.focus + 1) % len(controls.Actions)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || just("jump") {
			g.focus = (g.focus + len(controls.Actions) - 1) % len(controls.Actions)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || just("attack") {
			g.beginBind(controls.Actions[g.focus], axes)
		}
		if click {
			for i, a := range controls.Actions {
				if point.In(actionRect(i)) {
					g.focus = i
					g.beginBind(a, axes)
				}
			}
		}
		if click && point.In(image.Rect(60, 560, 360, 600)) && g.selected >= 0 {
			g.settings.Profiles[g.settings.Selected] = map[string]controls.Binding{}
			g.save()
		}
		if click && point.In(image.Rect(400, 560, 690, 600)) {
			g.settings.Deadzone += .05
			if g.settings.Deadzone > .51 {
				g.settings.Deadzone = .10
			}
			g.save()
		}
		return nil
	}
	if g.paused {
		g.edges.Sync(g.o, nil)
		return nil
	}
	desired := map[uint8]bool{}
	for key, scan := range keyboard {
		if ebiten.IsKeyPressed(key) {
			desired[scan] = true
		}
	}
	for action, on := range actions {
		if scan := controls.Scans[action]; on && scan != 0 {
			desired[scan] = true
		}
	}
	g.edges.Sync(g.o, desired)
	if err := g.o.RunCycles(g.cycles * 1000 / 60); err != nil {
		var exited *oracle.ExitError
		if errors.As(err, &exited) {
			return ebiten.Termination
		}
		return err
	}
	// 雛形尚未接音效裝置；仍取出音訊，避免執行器埠紀錄持續成長。
	g.audio.Render()
	return nil
}

var keyboard = map[ebiten.Key]uint8{
	ebiten.KeyArrowLeft: 0x4b, ebiten.KeyArrowRight: 0x4d, ebiten.KeyArrowUp: 0x48, ebiten.KeyArrowDown: 0x50,
	ebiten.KeyNumpad1: 0x4f, ebiten.KeyNumpad2: 0x50, ebiten.KeyNumpad3: 0x51, ebiten.KeyNumpad4: 0x4b, ebiten.KeyNumpad5: 0x4c, ebiten.KeyNumpad6: 0x4d, ebiten.KeyNumpad7: 0x47, ebiten.KeyNumpad8: 0x48, ebiten.KeyNumpad9: 0x49,
	ebiten.KeySpace: 0x39, ebiten.KeyAltLeft: 0x38, ebiten.KeyEnter: 0x1c, ebiten.KeyEscape: 0x01, ebiten.KeyControlLeft: 0x1d, ebiten.KeyShiftLeft: 0x2a, ebiten.KeyShiftRight: 0x36,
	ebiten.KeyF1: 0x3b, ebiten.KeyF2: 0x3c, ebiten.KeyF3: 0x3d, ebiten.KeyF4: 0x3e, ebiten.KeyF5: 0x3f, ebiten.KeyF6: 0x40, ebiten.KeyF7: 0x41, ebiten.KeyF8: 0x42, ebiten.KeyF9: 0x43, ebiten.KeyF10: 0x44,
	ebiten.KeyBackspace: 0x0e, ebiten.KeyTab: 0x0f, ebiten.KeyA: 0x1e, ebiten.KeyB: 0x30, ebiten.KeyC: 0x2e, ebiten.KeyD: 0x20, ebiten.KeyE: 0x12, ebiten.KeyF: 0x21, ebiten.KeyG: 0x22, ebiten.KeyH: 0x23, ebiten.KeyI: 0x17, ebiten.KeyJ: 0x24, ebiten.KeyK: 0x25, ebiten.KeyL: 0x26, ebiten.KeyM: 0x32, ebiten.KeyN: 0x31, ebiten.KeyO: 0x18, ebiten.KeyP: 0x19, ebiten.KeyQ: 0x10, ebiten.KeyR: 0x13, ebiten.KeyS: 0x1f, ebiten.KeyT: 0x14, ebiten.KeyU: 0x16, ebiten.KeyV: 0x2f, ebiten.KeyW: 0x11, ebiten.KeyX: 0x2d, ebiten.KeyY: 0x15, ebiten.KeyZ: 0x2c,
	ebiten.KeyDigit0: 0x0b, ebiten.KeyDigit1: 0x02, ebiten.KeyDigit2: 0x03, ebiten.KeyDigit3: 0x04, ebiten.KeyDigit4: 0x05, ebiten.KeyDigit5: 0x06, ebiten.KeyDigit6: 0x07, ebiten.KeyDigit7: 0x08, ebiten.KeyDigit8: 0x09, ebiten.KeyDigit9: 0x0a,
}

var white = color.RGBA{230, 235, 245, 255}

func (g *game) text(box image.Rectangle, s string) { g.font.Draw(g.canvas, box, s, white) }
func (g *game) button(box image.Rectangle, s string, active bool) {
	c := color.RGBA{42, 52, 70, 255}
	if active {
		c = color.RGBA{40, 90, 115, 255}
	}
	draw.Draw(g.canvas, box, &image.Uniform{c}, image.Point{}, draw.Src)
	g.text(image.Rect(box.Min.X+10, box.Min.Y, box.Max.X-10, box.Max.Y), s)
}

func (g *game) Draw(screen *ebiten.Image) {
	w, h, rawRGB := g.o.ScreenRGB()
	raw := g.o.Indexed()
	rgb, townRaw := rawRGB, raw
	var presented *hd.TownPresentation
	if g.townFrames != nil {
		blocked := g.townDialogue != nil && g.townDialogue.State()["key"] != ""
		presented = g.townFrames.Present(raw, rawRGB, blocked)
		if presented != nil {
			rgb, townRaw = presented.RGB, presented.Indexed
		}
	}
	if g.observePath != "" {
		held := g.o.HeldKeys()
		var rejections []map[string]int
		var quotation, bankDialogue, itemPurchase, itemSale, itemExplanation map[string]any
		var townDialogue map[string]any
		if g.townDialogue != nil {
			townDialogue = g.townDialogue.State()
		}
		if g.dialogue != nil {
			rejections = g.dialogue.Rejections
			quotation = g.dialogue.QuotationState()
			bankDialogue = g.dialogue.BankState()
			itemPurchase = g.dialogue.ItemPurchaseState()
			itemSale = g.dialogue.ItemSaleState()
			itemExplanation = g.dialogue.ItemExplanationState()
		}
		data, _ := json.Marshal(map[string]any{"steps": g.o.Steps(), "cycles": g.o.Cycles(), "focused": ebiten.IsFocused(), "paused": g.paused, "panel": g.panel, "mode": g.o.VideoMode(), "ip": fmt.Sprint(g.o.IP()), "dialogue_rejections": rejections, "quotation": quotation, "bank_dialogue": bankDialogue, "item_purchase": itemPurchase, "item_sale": itemSale, "item_explanation": itemExplanation, "held_keys": held, "frames": g.frames,
			"town_dialogue": townDialogue,
			"hd_enabled":    g.hdEnabled, "town_sky_hd_matched": g.townSkyHDMatched,
			"town_background_hd_matched": g.townBackgroundMatch,
			"town_scenery_hd_matched":    g.townSceneryMatch,
			"town_frame_presentation":    g.townFrames.State(),
			"town_background_cache":      g.o.Bytes(oracle.Far(0x1b1, 0xe000), 224),
			"town_background_module":     g.o.Bytes(oracle.Far(0x1b1, 0xc000), 64),
			"text_cursor":                g.o.Word(oracle.Far(0x01b1, 0xff4c)),
			"advance_input_flags":        int(g.o.Byte(oracle.Far(0x01b1, 0xff1d))) | int(g.o.Byte(oracle.Far(0x01b1, 0xff1e)))<<8})
		// Draw 可能跨過多個 Update，不能用幀數整除判斷心跳。
		// 先寫同目錄暫存檔再換名，觀察者只會讀到完整快照。
		if err := os.WriteFile(g.observePath+".tmp", append(data, '\n'), 0644); err == nil {
			_ = os.Rename(g.observePath+".tmp", g.observePath)
		}
	}
	draw.Draw(g.canvas, g.canvas.Bounds(), &image.Uniform{color.RGBA{18, 23, 32, 255}}, image.Point{}, draw.Src)
	if w == 320 && h == 200 && len(rgb) == 320*200*3 {
		for y := 0; y < 200; y++ {
			for x := 0; x < 320; x++ {
				i := (y*320 + x) * 3
				c := color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255}
				for yy := 0; yy < scale; yy++ {
					for xx := 0; xx < scale; xx++ {
						g.canvas.SetRGBA(x*scale+xx, barHeight+y*scale+yy, c)
					}
				}
			}
		}
	}
	g.hdMatched = false
	if g.hdEnabled && g.hd != nil && w == 320 && h == 200 {
		g.hdMatched = g.hd.Draw(g.canvas, raw, rawRGB, image.Pt(0, barHeight))
	}
	g.townSkyHDMatched = false
	if g.hdEnabled && g.townSkyHD != nil && w == 320 && h == 200 {
		g.townSkyHDMatched = g.townSkyHD.Draw(g.canvas, g.o.VideoMode(), g.o.Bytes(oracle.Far(0x1b1, 0xc000), 64), raw, rawRGB, image.Pt(0, barHeight))
		g.hdMatched = g.hdMatched || g.townSkyHDMatched
	}
	g.townBackgroundMatch = false
	var townPrimary, townResources, townExtra []byte
	if g.townBackgroundHD != nil || g.townSceneryHD != nil {
		townPrimary = g.o.Bytes(oracle.Far(0x1b1, 0), 65536)
		townResources = g.o.Bytes(oracle.Far(0x11b1, 0), 65536)
		townExtra = g.o.Bytes(oracle.Far(0x21b1, 0), 65536)
		if presented != nil {
			townPrimary, townResources, townExtra = presented.Primary, presented.Resources, presented.Extra
		}
	}
	if g.hdEnabled && g.townBackgroundHD != nil && w == 320 && h == 200 {
		g.townBackgroundMatch = g.townBackgroundHD.Draw(g.canvas, g.o.VideoMode(), townPrimary, townResources, townRaw, rgb, image.Pt(0, barHeight))
		g.hdMatched = g.hdMatched || g.townBackgroundMatch
	}
	g.townSceneryMatch = false
	if g.hdEnabled && g.townSceneryHD != nil && w == 320 && h == 200 {
		g.townSceneryMatch = g.townSceneryHD.Draw(g.canvas, g.o.VideoMode(), townPrimary, townResources, townExtra, townRaw, rgb, image.Pt(0, barHeight), true)
		g.hdMatched = g.hdMatched || g.townSceneryMatch
	}
	g.kingHDMatched, g.kingHDPose = false, -1
	if g.hdEnabled && g.kingHD != nil && w == 320 && h == 200 {
		g.kingHDMatched, g.kingHDPose = g.kingHD.Draw(g.canvas, raw, rawRGB, image.Pt(0, barHeight))
		g.hdMatched = g.hdMatched || g.kingHDMatched
	}
	g.sageHDMatched, g.sageHDPose = false, -1
	if g.hdEnabled && g.sageHD != nil && w == 320 && h == 200 {
		g.sageHDMatched, g.sageHDPose = g.sageHD.Draw(g.canvas, raw, rawRGB, image.Pt(0, barHeight))
		g.hdMatched = g.hdMatched || g.sageHDMatched
	}
	g.merchantHDMatched, g.merchantHDPose = false, -1
	if g.hdEnabled && g.merchantHD != nil && w == 320 && h == 200 {
		g.merchantHDMatched, g.merchantHDPose = g.merchantHD.Draw(g.canvas, raw, rawRGB, image.Pt(0, barHeight))
		g.hdMatched = g.hdMatched || g.merchantHDMatched
	}
	g.bankHDMatched, g.bankHDPose = false, -1
	if g.hdEnabled && g.bankHD != nil && w == 320 && h == 200 {
		g.bankHDMatched, g.bankHDPose = g.bankHD.Draw(g.canvas, raw, rawRGB, image.Pt(0, barHeight))
		g.hdMatched = g.hdMatched || g.bankHDMatched
	}
	g.itemShopHDMatched, g.itemShopHDPose = false, -1
	if g.hdEnabled && g.itemShopHD != nil && w == 320 && h == 200 {
		g.itemShopHDMatched, g.itemShopHDPose = g.itemShopHD.Draw(g.canvas, raw, rawRGB, image.Pt(0, barHeight))
		g.hdMatched = g.hdMatched || g.itemShopHDMatched
	}
	if g.intro != nil && w == 320 && h == 200 {
		g.intro.Frame(raw, rawRGB)
		clear(g.overlayPixels)
		if g.intro.Draw(g.overlayPixels, scale, nil) {
			src := &image.RGBA{Pix: g.overlayPixels, Stride: width * 4, Rect: image.Rect(0, 0, width, 600)}
			draw.Draw(g.canvas, image.Rect(0, barHeight, width, barHeight+600), src, image.Point{}, draw.Over)
		}
	}
	if g.scroll != nil && w == 320 && h == 200 {
		g.scroll.Draw(g.canvas, g.o.Indexed(), g.o.Palette(), image.Pt(0, barHeight))
	}
	if g.dialogue != nil && w == 320 && h == 200 {
		g.dialogue.Draw(g.canvas, g.o.Indexed(), g.o.Palette(), image.Pt(0, barHeight))
	}
	if g.menuLabels != nil && w == 320 && h == 200 {
		g.menuLabels.ObserveFrame(g.o.Indexed(), g.o.Palette())
		g.menuLabels.Draw(g.canvas, g.o.Indexed(), g.o.Palette(), image.Pt(0, barHeight))
	}
	if g.hud != nil && w == 320 && h == 200 {
		g.hud.Draw(g.canvas, g.o.Indexed(), g.o.Palette(), image.Pt(0, barHeight))
	}
	if g.townDialogue != nil && w == 320 && h == 200 {
		g.townDialogue.Draw(g.canvas, g.o.Indexed(), g.o.Palette(), image.Pt(0, barHeight))
	}
	g.button(image.Rect(8, 8, 150, 40), g.label("original"), !g.hdEnabled)
	hdLabel := "hd"
	if g.hd != nil {
		hdLabel = "hd_ready"
	}
	if g.townSkyHD != nil || g.townBackgroundHD != nil || g.townSceneryHD != nil || g.kingHD != nil || g.sageHD != nil || g.merchantHD != nil || g.bankHD != nil || g.itemShopHD != nil {
		hdLabel = "hd_art_ready"
	}
	g.button(image.Rect(158, 8, 340, 40), g.label(hdLabel), g.hdEnabled)
	g.button(image.Rect(348, 8, 540, 40), g.label("controller"), g.panel)
	pause := "pause"
	if g.paused {
		pause = "resume"
	}
	g.button(image.Rect(548, 8, 740, 40), g.label(pause), g.paused)
	g.text(image.Rect(12, 650, 948, 678), g.label("prototype"))
	if g.hdEnabled && !g.hdMatched {
		g.text(image.Rect(748, 8, 948, 40), g.label("hd_fallback"))
	}
	if g.panel {
		draw.Draw(g.canvas, image.Rect(40, 80, 920, 630), &image.Uniform{color.RGBA{25, 32, 45, 255}}, image.Point{}, draw.Src)
		g.text(image.Rect(60, 90, 900, 120), g.label("panel_title"))
		deviceName := g.label("none")
		if len(g.devices) > 0 {
			deviceName = g.label("choose")
			for _, d := range g.devices {
				if d.id == g.selected {
					deviceName = d.name
				}
			}
		}
		g.button(image.Rect(60, 135, 900, 175), g.label("devices")+"："+deviceName, false)
		g.text(image.Rect(60, 180, 900, 205), g.label("pair"))
		g.text(image.Rect(60, 210, 900, 235), g.label("bind"))
		g.text(image.Rect(60, 238, 900, 263), g.notice)
		for i, a := range controls.Actions {
			value := g.label("unbound")
			if g.selected >= 0 {
				if b, ok := g.profile()[a]; ok {
					value = fmt.Sprintf("%s %d", g.label(b.Kind), b.Index)
					if b.Kind == "axis" {
						value = fmt.Sprintf("%s %d %+.0f", g.label("axis"), b.Index, float64(b.Sign))
					}
				}
			}
			if g.listening == a {
				value = g.label("listening")
			}
			g.button(actionRect(i), g.label(a)+"："+value, g.focus == i)
		}
		g.text(image.Rect(60, 520, 900, 548), g.label("keyboard"))
		g.button(image.Rect(60, 560, 360, 600), g.label("clear"), false)
		g.button(image.Rect(400, 560, 690, 600), fmt.Sprintf("%s %.2f", g.label("deadzone"), g.settings.Deadzone), false)
		g.button(image.Rect(740, 560, 900, 600), g.label("close"), false)
	}
	g.texture.WritePixels(g.canvas.Pix)
	screen.DrawImage(g.texture, nil)
	captureReady := g.frames >= g.captureFrame
	if g.captureRequest != "" {
		_, err := os.Stat(g.captureRequest)
		captureReady = err == nil
	}
	if g.capturePath != "" && captureReady && !g.captured {
		if err := savePNG(g.capturePath, g.canvas); err != nil {
			log.Printf("擷取失敗：%v", err)
		} else {
			g.captured = true
			if err := os.WriteFile(g.capturePath+".indexed", g.o.Indexed(), 0644); err != nil {
				log.Printf("色號擷取失敗：%v", err)
			}
			if presented != nil {
				for suffix, data := range map[string][]byte{".presented.indexed": presented.Indexed, ".presented.source.indexed": presented.SourceIndexed, ".presented.primary": presented.Primary, ".presented.resources": presented.Resources, ".presented.extra": presented.Extra, ".presented.palette": presented.Palette} {
					if err := os.WriteFile(g.capturePath+suffix, data, 0644); err != nil {
						log.Printf("完整畫格來源擷取失敗：%v", err)
					}
				}
			}
			stamps := 0
			if g.intro != nil {
				stamps = len(g.intro.Stamps)
			}
			rows, sources := 0, 0
			if g.scroll != nil {
				rows, sources = g.scroll.Rows, g.scroll.Sources
			}
			dialogueGlyphs, dialogueParagraphs := 0, 0
			var dialogueRejections []map[string]int
			var quotation, bankDialogue, itemPurchase, itemSale, itemExplanation map[string]any
			if g.dialogue != nil {
				dialogueGlyphs, dialogueParagraphs = g.dialogue.Glyphs, g.dialogue.Paragraphs
				dialogueRejections = g.dialogue.Rejections
				quotation = g.dialogue.QuotationState()
				bankDialogue = g.dialogue.BankState()
				itemPurchase = g.dialogue.ItemPurchaseState()
				itemSale = g.dialogue.ItemSaleState()
				itemExplanation = g.dialogue.ItemExplanationState()
			}
			meta, _ := json.Marshal(map[string]any{"steps": g.o.Steps(), "elapsed_cycles": g.o.Cycles(), "held_keys": g.o.HeldKeys(), "panel": g.panel, "mode": g.o.VideoMode(), "intro_stamps": stamps, "intro_scroll_rows": rows, "intro_scroll_sources": sources, "dialogue_glyphs": dialogueGlyphs, "dialogue_paragraphs": dialogueParagraphs, "dialogue_rejections": dialogueRejections, "quotation": quotation, "bank_dialogue": bankDialogue, "item_purchase": itemPurchase, "item_sale": itemSale, "item_explanation": itemExplanation, "hd_enabled": g.hdEnabled, "hd_matched": g.hdMatched, "king_hd_matched": g.kingHDMatched, "king_hd_pose": g.kingHDPose, "sage_hd_matched": g.sageHDMatched, "sage_hd_pose": g.sageHDPose,
				"town_dialogue": func() map[string]any {
					if g.townDialogue != nil {
						return g.townDialogue.State()
					}
					return nil
				}(),
				"merchant_hd_matched": g.merchantHDMatched, "merchant_hd_pose": g.merchantHDPose,
				"bank_hd_matched": g.bankHDMatched, "bank_hd_pose": g.bankHDPose,
				"item_shop_hd_matched": g.itemShopHDMatched, "item_shop_hd_pose": g.itemShopHDPose,
				"town_sky_hd_matched": g.townSkyHDMatched, "town_background_hd_matched": g.townBackgroundMatch,
				"town_scenery_hd_matched": g.townSceneryMatch,
				"town_frame_presentation": g.townFrames.State(),
				"town_background_cache":   g.o.Bytes(oracle.Far(0x1b1, 0xe000), 224), "town_background_module": g.o.Bytes(oracle.Far(0x1b1, 0xc000), 64)})
			if err := os.WriteFile(g.capturePath+".json", append(meta, '\n'), 0644); err != nil {
				log.Printf("擷取 metadata 失敗：%v", err)
			}
			log.Printf("GUI 擷取：%s；原版指令數 %d", g.capturePath, g.o.Steps())
		}
	}
}

func savePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	err = png.Encode(f, img)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func (g *game) Layout(int, int) (int, int) { return width, height }

func main() {
	orig := flag.String("orig", "", "合法原版資料目錄，須含 ZELIARD.EXE")
	fontPath := flag.String("font", "", "使用者提供的 Unifont HEX.gz")
	textPath := flag.String("text", "text/host.json", "前端中文資料")
	userDir := flag.String("user-dir", "", "原版存檔與前端設定的可寫目錄")
	cycles := flag.Uint64("cycles", 3000, "雛形 DOSBox 相容 cycles；尚非實機速度驗收值")
	capture := flag.String("capture", "", "GUI 測試擷取路徑，完成後結束")
	captureRequest := flag.String("capture-request", "", "GUI 測試：等待指定訊號檔存在再擷取，取代更新次數條件")
	observePath := flag.String("observe", "", "可選：每秒寫入唯讀執行進度，供 GUI 驗證診斷")
	frames := flag.Int("capture-frames", 120, "GUI 測試擷取前更新次數")
	panel := flag.Bool("show-controller", false, "GUI 測試啟動時展開手把頁")
	introRef := flag.String("intro-reference", "", "可選：本地 probe 的片頭原文參考畫格，只覆繪一句")
	introText := flag.String("intro-text", "text/intro-prototype.json", "片頭覆繪原型譯文")
	introScroll := flag.String("intro-scroll", "", "可選：首章捲動中文試譯 TSV，取代單句 prototype")
	dialogueText := flag.String("dialogue-text", "", "可選：對話與報價試譯 TSV")
	labelsText := flag.String("labels-text", "", "可選：城鎮選單標籤試譯 TSV")
	hudText := flag.String("hud-text", "", "可選：狀態列與場景名稱試譯 TSV")
	townText := flag.String("town-text", "", "可選：城鎮居民純文字試譯 TSV")
	hdRef := flag.String("hd-reference", "", "可選：本地城鎮 probe 的 screen.bin，同目錄須有 palette.bin")
	hdArt := flag.String("hd-art", "", "可選：本輪手繪邊框 PNG，僅供 HD prototype")
	townSkyRef := flag.String("town-sky-reference", "", "可選：城鎮遠山原版indexed畫格")
	townSkyPalette := flag.String("town-sky-palette", "", "可選：城鎮遠山原版色盤")
	townSkyArt := flag.String("town-sky-art", "", "可選：本地手繪遠山PNG")
	townBackgroundMemory := flag.String("town-background-memory", "", "可選：完整遠山正常原版記憶體參考")
	townBackgroundArt := flag.String("town-background-art", "", "可選：完整手繪遠山PNG，搭配原版樹梢遮罩")
	townSceneryArt := flag.String("town-scenery-art-dir", "", "可選：Muralla三段手繪前景圖目錄，搭配town-background-memory")
	townSignsText := flag.String("town-signs-text", "", "可選：Muralla招牌中文TSV，搭配前景HD")
	kingRefA := flag.String("king-reference-a", "", "可選：本地原版國王姿態 A 的 screen.bin")
	kingRefB := flag.String("king-reference-b", "", "可選：本地原版國王姿態 B 的 screen.bin")
	kingArt := flag.String("king-art", "", "可選：本地手繪國王閉口 PNG")
	kingSpeakingArt := flag.String("king-speaking-art", "", "可選：本地手繪國王說話 PNG")
	sageRefA := flag.String("sage-reference-a", "", "可選：本地原版長老畫格 A")
	sageRefB := flag.String("sage-reference-b", "", "可選：本地原版長老畫格 B")
	sagePalette := flag.String("sage-palette", "", "可選：本地原版長老色盤")
	sageArt := flag.String("sage-art", "", "可選：本地手繪長老 PNG")
	merchantRefA := flag.String("merchant-reference-a", "", "可選：武器商人原版迎客畫格")
	merchantRefB := flag.String("merchant-reference-b", "", "可選：武器商人原版告別畫格")
	merchantRefC := flag.String("merchant-reference-c", "", "可選：武器商人未購物告別原版畫格")
	merchantPalette := flag.String("merchant-palette", "", "可選：武器商人原版色盤")
	merchantArtA := flag.String("merchant-art-a", "", "可選：本地手繪武器商人迎客 PNG")
	merchantArtB := flag.String("merchant-art-b", "", "可選：本地手繪武器商人告別 PNG")
	merchantArtC := flag.String("merchant-art-c", "", "可選：本地手繪武器商人未購物告別 PNG")
	bankRefA := flag.String("bank-reference-a", "", "可選：銀行原版畫格 A")
	bankRefB := flag.String("bank-reference-b", "", "可選：銀行原版畫格 B")
	bankRefC := flag.String("bank-reference-c", "", "可選：銀行原版畫格 C")
	bankPalette := flag.String("bank-palette", "", "可選：銀行原版色盤")
	bankArtA := flag.String("bank-art-a", "", "可選：本地手繪銀行畫格 A PNG")
	bankArtB := flag.String("bank-art-b", "", "可選：本地手繪銀行畫格 B PNG")
	bankArtC := flag.String("bank-art-c", "", "可選：本地手繪銀行畫格 C PNG")
	var itemRefs, itemArts [4]*string
	for i := range itemRefs {
		itemRefs[i] = flag.String(fmt.Sprintf("item-reference-%d", i), "", "可選：道具店已核對原版畫格")
		itemArts[i] = flag.String(fmt.Sprintf("item-art-%d", i), "", "可選：道具店圖組，依序基準／蒸氣B／抬手／蒸氣C")
	}
	itemPalette := flag.String("item-palette", "", "可選：道具店原版色盤")
	flag.Parse()
	if *orig == "" || *fontPath == "" || *userDir == "" || *cycles < 60 || *cycles > 10000 || *frames < 1 {
		log.Fatal("須指定 -orig、-font、-user-dir；cycles 範圍 60–10000")
	}
	root, err := filepath.Abs(*orig)
	if err != nil {
		log.Fatal(err)
	}
	userRoot, err := filepath.Abs(*userDir)
	if err != nil {
		log.Fatal(err)
	}
	rel, err := filepath.Rel(root, userRoot)
	if err != nil || !(rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
		log.Fatal("使用者目錄不可位於原版資料目錄內")
	}
	if err := os.MkdirAll(filepath.Join(userRoot, "game"), 0700); err != nil {
		log.Fatal(err)
	}
	settings, err := controls.Load(filepath.Join(userRoot, "controller.json"))
	if err != nil {
		log.Fatal(err)
	}
	font, err := bitmap.Load(*fontPath)
	if err != nil {
		log.Fatal(err)
	}
	data, err := os.ReadFile(*textPath)
	if err != nil {
		log.Fatal(err)
	}
	labels := map[string]string{}
	if err = json.Unmarshal(data, &labels); err != nil {
		log.Fatal(err)
	}
	for _, s := range labels {
		for _, r := range s {
			if font[r] == nil {
				log.Fatalf("字型缺字 U+%04X", r)
			}
		}
	}
	o, err := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
	if err != nil {
		log.Fatal(err)
	}
	defer o.Close()
	o.SetScratch(filepath.Join(userRoot, "game"))
	o.SetDOSBoxCycles(*cycles)
	g := &game{o: o, audio: o.NewAudio(44100), font: font, labels: labels, settings: settings, settingsPath: filepath.Join(userRoot, "controller.json"), selected: -1, cycles: *cycles, panel: *panel, capturePath: *capture, captureRequest: *captureRequest, captureFrame: *frames, canvas: image.NewRGBA(image.Rect(0, 0, width, height)), texture: ebiten.NewImage(width, height)}
	g.observePath = *observePath
	if *hdRef != "" || *hdArt != "" {
		g.hd, err = hd.NewFrame(*hdRef, filepath.Join(filepath.Dir(*hdRef), "palette.bin"), *hdArt)
		if err != nil {
			log.Printf("HD 試作未載入，保留原貌：%v", err)
		}
	}
	if *townBackgroundMemory != "" || *townBackgroundArt != "" {
		g.townBackgroundHD, err = hd.NewTownBackground(root, *townSkyRef, *townSkyPalette, *townBackgroundMemory, *townBackgroundArt)
		if err != nil {
			log.Printf("完整遠山HD未載入，保留原貌：%v", err)
		}
	} else if *townSkyRef != "" || *townSkyPalette != "" || *townSkyArt != "" {
		g.townSkyHD, err = hd.NewTownSky(root, *townSkyRef, *townSkyPalette, *townSkyArt)
		if err != nil {
			log.Printf("城鎮遠山HD未載入，保留原貌：%v", err)
		}
	}
	if *townSceneryArt != "" || *townSignsText != "" {
		g.townSceneryHD, err = hd.NewTownScenery(root, *townBackgroundMemory, *townSkyPalette, *townSceneryArt, *townSignsText, font)
		if err != nil {
			log.Printf("城鎮前景HD未載入，保留原貌：%v", err)
		}
		g.townFrames = hd.NewTownFrames(g.o, g.townSceneryHD)
	}
	if *kingRefA != "" || *kingRefB != "" || *kingArt != "" || *kingSpeakingArt != "" {
		g.kingHD, err = hd.NewKing(*kingRefA, *kingRefB, filepath.Join(filepath.Dir(*kingRefA), "palette.bin"), *kingArt, *kingSpeakingArt)
		if err != nil {
			log.Printf("國王 HD 圖組未載入，保留原貌：%v", err)
		}
	}
	if *sageRefA != "" || *sageRefB != "" || *sagePalette != "" || *sageArt != "" {
		g.sageHD, err = hd.NewSage(*sageRefA, *sageRefB, *sagePalette, *sageArt)
		if err != nil {
			log.Printf("長老 HD 圖面未載入，保留原貌：%v", err)
		}
	}
	if *merchantRefA != "" || *merchantRefB != "" || *merchantRefC != "" || *merchantPalette != "" || *merchantArtA != "" || *merchantArtB != "" || *merchantArtC != "" {
		g.merchantHD, err = hd.NewMerchant(*merchantRefA, *merchantRefB, *merchantRefC, *merchantPalette, *merchantArtA, *merchantArtB, *merchantArtC)
		if err != nil {
			log.Printf("武器商人 HD 圖組未載入，保留原貌：%v", err)
		}
	}
	if *bankRefA != "" || *bankRefB != "" || *bankRefC != "" || *bankPalette != "" || *bankArtA != "" || *bankArtB != "" || *bankArtC != "" {
		g.bankHD, err = hd.NewBank(*bankRefA, *bankRefB, *bankRefC, *bankPalette, *bankArtA, *bankArtB, *bankArtC)
		if err != nil {
			log.Printf("銀行 HD 圖組未載入，保留原貌：%v", err)
		}
	}
	var itemReferencePaths, itemArtworkPaths [4]string
	itemRequested := *itemPalette != ""
	for i := range itemRefs {
		itemReferencePaths[i], itemArtworkPaths[i] = *itemRefs[i], *itemArts[i]
		itemRequested = itemRequested || *itemRefs[i] != "" || *itemArts[i] != ""
	}
	if itemRequested {
		g.itemShopHD, err = hd.NewItemShop(itemReferencePaths, *itemPalette, itemArtworkPaths)
		if err != nil {
			log.Printf("道具店 HD 圖組未載入，保留原貌：%v", err)
		}
	}
	if *introScroll != "" {
		g.scroll, err = overlay.NewScroll(o, filepath.Join(root, "ZELRES1.SAR"), *introScroll, font)
		if err != nil {
			log.Printf("片頭捲動中文未載入，保留原文：%v", err)
		}
	} else if *introRef != "" {
		g.intro, err = overlay.NewIntro(*introRef, *introText, font)
		if err != nil {
			log.Fatal(err)
		}
		g.overlayPixels = make([]byte, width*600*4)
	}
	if *dialogueText != "" {
		g.dialogue, err = overlay.NewDialogue(o, filepath.Join(root, "ZELRES2.SAR"), *dialogueText, font)
		if err != nil {
			log.Printf("對話試譯未載入，保留原文：%v", err)
		}
	}
	if *labelsText != "" {
		g.menuLabels, err = overlay.NewLabels(o, filepath.Join(root, "ZELRES2.SAR"), *labelsText, font)
		if err != nil {
			log.Printf("選單試譯未載入，保留原文：%v", err)
		}
	}
	if g.dialogue != nil {
		g.dialogue.UseLabels(g.menuLabels)
	}
	if *hudText != "" {
		g.hud, err = overlay.NewHUD(o, root, *hudText, font)
		if err != nil {
			log.Printf("狀態列試譯未載入，保留原文：%v", err)
		}
	}
	ebiten.SetWindowSize(width, height)
	if *townText != "" {
		g.townDialogue, err = overlay.NewTownDialogue(o, filepath.Join(root, "ZELRES2.SAR"), *townText, font)
		if err != nil {
			log.Printf("居民對話試譯未載入，保留原文：%v", err)
		}
	}
	ebiten.SetWindowTitle(g.label("title"))
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetRunnableOnUnfocused(true)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

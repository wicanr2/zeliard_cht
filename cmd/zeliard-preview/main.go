// zeliard-preview 直接執行原版的可撤回前端。規格見 docs/spec/001-runtime-overlay.md。
package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/dosgolem/xlate"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
	"github.com/wicanr2/zeliard_cht/internal/prototype/cheats"
	"github.com/wicanr2/zeliard_cht/internal/prototype/controls"
	"github.com/wicanr2/zeliard_cht/internal/prototype/fm"
	"github.com/wicanr2/zeliard_cht/internal/prototype/glyphpatch"
	"github.com/wicanr2/zeliard_cht/internal/prototype/hd"
	"github.com/wicanr2/zeliard_cht/internal/prototype/overlay"
	"github.com/wicanr2/zeliard_cht/internal/prototype/sound"
)

const width, height, scale, barHeight = 960, 680, 3, 48

var version = "development"

type device struct {
	id        ebiten.GamepadID
	key, name string
}
type game struct {
	textSizePanel               bool
	textSizeFocus               int
	uiTextOverflow              []string
	hdJobs                      []*hdLoadJob
	roomPrefixes                map[int][]byte
	selectPrefix                []byte
	endingPrefix                []byte
	hdLoading                   *hdLoadJob
	hdFamilySignature, hdFamily string

	lastRenderedAt     time.Time
	lastRenderedUpdate int

	cheats               *cheats.Controller
	cheatPanel           bool
	cheatFocus           int
	closeResources       func()
	audioMode, audioPath string

	cpuFrameUsed, lastDrawCost, sliceElapsed time.Duration
	sliceRan                                 uint64
	lastRateChange                           time.Time
	baseCycles                               uint64
	speedPanel                               bool
	speedFocus                               int
	perf                                     performance

	realtime             bool
	clock                sound.Clock
	o                    *oracle.Oracle
	audio                *oracle.Audio
	output               *audioOutput
	font                 bitmap.Font
	labels               map[string]string
	settings             controls.Settings
	settingsPath         string
	selected             ebiten.GamepadID
	devices              []device
	edges                controls.Edges
	panel, paused        bool
	keyboardPanel        bool
	releaseGate          bool
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
	inputTrace           string
	inputTraceError      string
	captured             bool
	captureTail          int
	intro                *xlate.Layer
	scroll               *overlay.Scroll
	introThreat          *overlay.IntroThreat
	introStory           *overlay.IntroStory
	endingStory          *overlay.IntroStory
	endingCredits        *overlay.EndingCredits
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
	manualArtwork        *hd.ManualArtwork
	worldTowns           *hd.WorldTowns
	townSceneryMatch     bool
	townFrames           *hd.TownFrames
	caveFrames           *hd.CaveFrames
	caveTerrainHD        *hd.CaveTerrain
	caveTerrainMatch     bool
	townGroundHD         *hd.TownGround
	townGroundMatch      bool
	townPlayerHD         *hd.TownPlayer
	townPlayerMatch      bool
	townActorsHD         *hd.TownActors
	townActorsMatch      bool
	interiorHD           *hd.Interiors
	interiorHDMatch      bool
	kingHD               *hd.King
	kingHDMatched        bool
	kingHDPose           int
	sageHD               *hd.Sage
	sageFramesHD         *hd.SageFrames
	sageHDMatched        bool
	sageHDPose           int
	merchantHD           *hd.Merchant
	merchantFramesHD     *hd.MerchantFrames
	merchantHDMatched    bool
	merchantHDPose       int
	bankHD               *hd.Bank
	bankHDMatched        bool
	bankHDPose           int
	itemShopHD           *hd.ItemShop
	itemShopHDMatched    bool
	itemShopHDPose       int
	churchHD             *hd.Church
	churchHDMatched      bool
	churchHDFrame        int
	hdEnabled, hdMatched bool
	layers               map[string]bool
	beforeText           func()
	portraits            map[string]*portraitHold
	dissolveHeld         *dissolveHold
	// townHeld 是上一次合成的城鎮畫帶，原版首次畫城鎮場景期間沿用。規格067。
	townHeld      *image.RGBA
	overlayPixels []byte
	texture       *ebiten.Image
	canvas        *image.RGBA
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
		g.edges.Sync(g, nil)
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
	g.edges.Sync(g, nil)
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
	return image.Rect(60+col*430, 270+row*40, 465+col*430, 308+row*40)
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
	begin := time.Now()
	defer func() { g.perf.update(begin) }()
	defer func() {
		if g.realtime && (g.captured || g.panel || g.paused || g.hdLoading != nil || g.releaseGate || !ebiten.IsFocused()) {
			g.clock.Reset()
		}
	}()
	defer g.audioLifecycle()
	if g.captured {
		g.captureTail++
		if g.captureTail >= 60 {
			return ebiten.Termination
		}
		return nil
	}
	g.frames++
	if g.updateHDLoading() {
		return nil
	}
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
		g.edges.Sync(g, nil)
		g.releaseGate = true
		return nil
	}
	just := func(a string) bool { return actions[a] && !g.previousActions[a] }
	if !g.panel && (just("toolbar") || g.keyboardToolbarPressed()) {
		g.panel = true
		g.keyboardPanel = g.keyboardToolbarPressed()
		g.speedPanel = false
		g.cheatPanel, g.textSizePanel = false, false
		g.releaseGate = true
		g.edges.Sync(g, nil)
	}
	x, y := ebiten.CursorPosition()
	point := image.Pt(x, y)
	click := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if click && point.In(toolbarOriginal) {
		g.hdEnabled = false
	}
	if click && point.In(toolbarHD) && g.hasHDArtwork() {
		g.hdEnabled = !g.hdEnabled
	}
	if click && image.Pt(x, y).In(toolbarController) {
		g.panel = !g.panel || g.keyboardPanel || g.speedPanel || g.cheatPanel || g.textSizePanel
		g.keyboardPanel = false
		g.cheatPanel = false
		g.speedPanel = false
		g.textSizePanel = false
		g.listening = ""
		g.releaseGate = true
		g.edges.Sync(g, nil)
	}
	if click && point.In(toolbarKeyboard) {
		g.panel = !g.panel || !g.keyboardPanel
		g.keyboardPanel = true
		g.cheatPanel = false
		g.speedPanel = false
		g.textSizePanel = false
		g.listening = ""
		g.releaseGate = true
		g.edges.Sync(g, nil)
	}
	if click && point.In(toolbarPause) {
		g.paused = !g.paused
		g.edges.Sync(g, nil)
	}
	if click && point.In(toolbarSpeed) {
		g.openSpeedPanel()
	}
	if click && point.In(toolbarCheats) {
		g.openCheatPanel()
	}
	if click && point.In(toolbarTextSize) {
		g.openTextSizePanel()
	}
	if g.panel && g.textSizePanel {
		return g.updateTextSizePanel(point, click, just)
	}
	if g.panel && g.cheatPanel {
		return g.updateCheatPanel(point, click, just)
	}
	if g.panel && g.speedPanel {
		return g.updateSpeedPanel(point, click, just)
	}
	if g.panel {
		g.edges.Sync(g, nil)
		if click && point.In(image.Rect(720, 88, 900, 128)) {
			g.keyboardPanel = !g.keyboardPanel
			g.listening, g.notice = "", ""
		}
		if g.listening != "" {
			if g.keyboardPanel {
				g.captureKeyboard()
			} else {
				g.captureBinding(buttons, axes)
			}
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || click && point.In(image.Rect(740, 560, 900, 600)) {
			g.panel = false
			g.releaseGate = true
			return nil
		}
		if !g.keyboardPanel && (inpututil.IsKeyJustPressed(ebiten.KeyTab) || click && point.In(image.Rect(60, 135, 900, 175))) {
			g.nextDevice()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || just("down") {
			g.focus = (g.focus + 1) % len(controls.Actions)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || just("jump") {
			g.focus = (g.focus + len(controls.Actions) - 1) % len(controls.Actions)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || just("attack") {
			if g.keyboardPanel {
				g.beginKeyboardBind(controls.Actions[g.focus])
			} else {
				g.beginBind(controls.Actions[g.focus], axes)
			}
		}
		if click {
			for i, a := range controls.Actions {
				if point.In(actionRect(i)) {
					g.focus = i
					if g.keyboardPanel {
						g.beginKeyboardBind(a)
					} else {
						g.beginBind(a, axes)
					}
				}
			}
		}
		if click && point.In(image.Rect(60, 560, 360, 600)) {
			if g.keyboardPanel {
				g.settings.Keyboard = controls.DefaultKeyboard()
				g.save()
			} else if g.selected >= 0 {
				g.settings.Profiles[g.settings.Selected] = map[string]controls.Binding{}
				g.save()
			}
		}
		if click && point.In(image.Rect(400, 560, 690, 600)) {
			if g.keyboardPanel {
				g.bindKeyboard("")
			} else {
				g.settings.Deadzone += .05
				if g.settings.Deadzone > .51 {
					g.settings.Deadzone = .10
				}
				g.save()
			}
		}
		return nil
	}
	if g.paused {
		g.edges.Sync(g, nil)
		return nil
	}
	if g.releaseGate {
		g.edges.Sync(g, nil)
		if len(inpututil.AppendPressedKeys(nil)) != 0 {
			return nil
		}
		for _, active := range actions {
			if active {
				return nil
			}
		}
		g.releaseGate = false
		return nil
	}
	desired := map[uint8]bool{}
	legacy := map[string]uint8{}
	for key, scan := range keyboard {
		legacy[key.String()] = scan
	}
	mapping := g.settings.KeyboardMap(legacy)
	for key := range keyboard {
		if scan := mapping[key.String()]; scan != 0 && ebiten.IsKeyPressed(key) {
			desired[scan] = true
		}
	}
	for action, on := range actions {
		if scan := controls.Scans[action]; on && scan != 0 {
			desired[scan] = true
		}
	}
	g.edges.Sync(g, desired)
	budget := g.cycles * 1000 / 60
	if g.realtime {
		budget = g.clock.Budget(time.Now(), g.o.Cycles(), g.cycles)
		if budget == 0 {
			return nil
		}
	}
	g.cheats.Apply()
	if err := g.runCPUSlice(budget); err != nil {
		var exited *oracle.ExitError
		if errors.As(err, &exited) {
			return ebiten.Termination
		}
		return err
	}
	// 原版取樣只在更新執行緒取得；播放端不持有原版執行器。
	pcm := g.audio.Render()
	g.trace(map[string]any{"kind": "audio"})
	if err := g.output.submit(pcm); err != nil {
		return err
	}
	g.adaptCPU()
	return nil
}

func (g *game) trace(event map[string]any) {
	if g.inputTrace == "" {
		return
	}
	if event["kind"] == "cpu_speed" || event["kind"] == "audio" || event["kind"] == "cheat_settings" {
		event["before_steps"], event["after_steps"] = g.o.Steps(), g.o.Steps()
		event["before_cycles"], event["after_cycles"] = g.o.Cycles(), g.o.Cycles()
		event["after_regs"] = g.o.Regs()
	}
	data, err := json.Marshal(event)
	if err == nil {
		var f *os.File
		f, err = os.OpenFile(g.inputTrace, os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			_, err = f.Write(append(data, '\n'))
			closeErr := f.Close()
			if err == nil {
				err = closeErr
			}
		}
	}
	if err != nil {
		g.inputTraceError = err.Error()
		log.Printf("輸入日誌寫入失敗：%v", err)
	}
}

func saveCavePaints(path string, paints []hd.CavePaint) error {
	data, err := json.MarshalIndent(paints, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

func (g *game) KeyDown(scan uint8) { g.traceKey(scan, true) }
func (g *game) KeyUp(scan uint8)   { g.traceKey(scan, false) }
func (g *game) traceKey(scan uint8, pressed bool) {
	steps, cycles := g.o.Steps(), g.o.Cycles()
	if pressed {
		g.o.KeyDown(scan)
	} else {
		g.o.KeyUp(scan)
	}
	if g.inputTrace != "" {
		g.trace(map[string]any{"kind": "key", "frame": g.frames, "scan": scan, "pressed": pressed, "before_steps": steps, "before_cycles": cycles, "after_steps": g.o.Steps(), "after_cycles": g.o.Cycles()})
	}
}
func (g *game) traceRun(steps, cycles, budget uint64, audio bool, err error) {
	if g.inputTrace == "" {
		return
	}
	message := ""
	if err != nil {
		message = err.Error()
	}
	g.trace(map[string]any{"kind": "run", "frame": g.frames, "cycle_budget": budget, "before_steps": steps, "before_cycles": cycles, "after_steps": g.o.Steps(), "after_cycles": g.o.Cycles(), "after_regs": g.o.Regs(), "audio_render": audio, "error": message})
}

var keyboard = map[ebiten.Key]uint8{
	ebiten.KeyF11: 0x57, ebiten.KeyF12: 0x58, ebiten.KeyNumpad0: 0x52,
	ebiten.KeyArrowLeft: 0x4b, ebiten.KeyArrowRight: 0x4d, ebiten.KeyArrowUp: 0x48, ebiten.KeyArrowDown: 0x50,
	ebiten.KeyNumpad1: 0x4f, ebiten.KeyNumpad2: 0x50, ebiten.KeyNumpad3: 0x51, ebiten.KeyNumpad4: 0x4b, ebiten.KeyNumpad5: 0x4c, ebiten.KeyNumpad6: 0x4d, ebiten.KeyNumpad7: 0x47, ebiten.KeyNumpad8: 0x48, ebiten.KeyNumpad9: 0x49,
	ebiten.KeySpace: 0x39, ebiten.KeyAltLeft: 0x38, ebiten.KeyEnter: 0x1c, ebiten.KeyEscape: 0x01, ebiten.KeyControlLeft: 0x1d, ebiten.KeyShiftLeft: 0x2a, ebiten.KeyShiftRight: 0x36,
	ebiten.KeyF1: 0x3b, ebiten.KeyF2: 0x3c, ebiten.KeyF3: 0x3d, ebiten.KeyF4: 0x3e, ebiten.KeyF5: 0x3f, ebiten.KeyF6: 0x40, ebiten.KeyF7: 0x41, ebiten.KeyF8: 0x42, ebiten.KeyF9: 0x43, ebiten.KeyF10: 0x44,
	ebiten.KeyBackspace: 0x0e, ebiten.KeyTab: 0x0f, ebiten.KeyA: 0x1e, ebiten.KeyB: 0x30, ebiten.KeyC: 0x2e, ebiten.KeyD: 0x20, ebiten.KeyE: 0x12, ebiten.KeyF: 0x21, ebiten.KeyG: 0x22, ebiten.KeyH: 0x23, ebiten.KeyI: 0x17, ebiten.KeyJ: 0x24, ebiten.KeyK: 0x25, ebiten.KeyL: 0x26, ebiten.KeyM: 0x32, ebiten.KeyN: 0x31, ebiten.KeyO: 0x18, ebiten.KeyP: 0x19, ebiten.KeyQ: 0x10, ebiten.KeyR: 0x13, ebiten.KeyS: 0x1f, ebiten.KeyT: 0x14, ebiten.KeyU: 0x16, ebiten.KeyV: 0x2f, ebiten.KeyW: 0x11, ebiten.KeyX: 0x2d, ebiten.KeyY: 0x15, ebiten.KeyZ: 0x2c,
	ebiten.KeyDigit0: 0x0b, ebiten.KeyDigit1: 0x02, ebiten.KeyDigit2: 0x03, ebiten.KeyDigit3: 0x04, ebiten.KeyDigit4: 0x05, ebiten.KeyDigit5: 0x06, ebiten.KeyDigit6: 0x07, ebiten.KeyDigit7: 0x08, ebiten.KeyDigit8: 0x09, ebiten.KeyDigit9: 0x0a,
}

var white = color.RGBA{230, 235, 245, 255}

func (g *game) text(box image.Rectangle, s string) {
	size := g.systemTextSize()
	// 工具列與狀態列的高度固定，最大24像素不侵入原版畫面。
	if (box.Max.Y <= barHeight || box.Min.Y >= 648) && size > 24 {
		size = 24
	}
	if !g.font.DrawText(g.canvas, box, s, size, white) {
		g.uiTextOverflow = append(g.uiTextOverflow, s)
	}
}
func (g *game) button(box image.Rectangle, s string, active bool) {
	c := color.RGBA{42, 52, 70, 255}
	if active {
		c = color.RGBA{40, 90, 115, 255}
	}
	draw.Draw(g.canvas, box, &image.Uniform{c}, image.Point{}, draw.Src)
	g.text(image.Rect(box.Min.X+6, box.Min.Y, box.Max.X-6, box.Max.Y), s)
}

func (g *game) cavePlayerState() map[string]any {
	if !g.hdEnabled || !g.caveTerrainMatch {
		return nil
	}
	return g.caveTerrainHD.PlayerState()
}

func (g *game) caveWeaponState() map[string]any {
	if !g.hdEnabled || !g.caveTerrainMatch {
		return nil
	}
	return g.caveTerrainHD.WeaponState()
}

func (g *game) caveObjectState() map[string]any {
	if !g.hdEnabled || !g.caveTerrainMatch {
		return nil
	}
	return g.caveTerrainHD.ObjectState()
}

func (g *game) cavePortalState() map[string]any {
	if !g.hdEnabled || !g.caveTerrainMatch {
		return nil
	}
	return g.caveTerrainHD.PortalState()
}

func (g *game) caveCommonState() map[string]any {
	if !g.hdEnabled || !g.caveTerrainMatch {
		return nil
	}
	return g.caveTerrainHD.CommonState()
}

func (g *game) caveMagiaState() map[string]any {
	if !g.hdEnabled || !g.caveTerrainMatch {
		return nil
	}
	return g.caveTerrainHD.MagiaState()
}

func (g *game) caveDoorState() map[string]any {
	if g.caveTerrainHD == nil {
		return nil
	}
	return g.caveTerrainHD.DoorState()
}

func (g *game) Draw(screen *ebiten.Image) {
	if g.hdLoading != nil {
		g.drawHDLoading(screen)
		return
	}
	if !g.lastRenderedAt.IsZero() && g.frames-g.lastRenderedUpdate < 2 && time.Since(g.lastRenderedAt) < time.Second/30 {
		screen.DrawImage(g.texture, nil)
		return
	}
	g.lastRenderedAt = time.Now()
	g.lastRenderedUpdate = g.frames
	g.uiTextOverflow = nil
	begin := time.Now()
	defer func() { g.lastDrawCost = time.Since(begin); g.cpuFrameUsed = 0; g.perf.draw(begin) }()
	presented, cavePresented, w, h, raw, rawRGB := g.compose()
	g.button(toolbarOriginal, g.label("original"), !g.hdEnabled)
	hdLabel := "hd"
	if g.hd != nil {
		hdLabel = "hd_ready"
	}
	if g.hasHDArtwork() {
		hdLabel = "hd_art_ready"
	}
	g.button(toolbarHD, g.label(hdLabel), g.hdEnabled)
	g.button(toolbarSpeed, fmt.Sprintf("%s %dx", g.label("cpu_speed"), g.settings.CPUSpeed), g.speedPanel)
	g.button(toolbarController, g.label("controller_short"), g.panel && !g.keyboardPanel && !g.speedPanel && !g.cheatPanel && !g.textSizePanel)
	g.button(toolbarCheats, g.label("cheat_toolbar"), g.cheatPanel || g.cheats != nil && (g.cheats.LockHP || g.cheats.LockMP || g.cheats.DisableRegeneration))
	g.button(toolbarKeyboard, g.label("keyboard_short"), g.panel && g.keyboardPanel)
	g.button(toolbarTextSize, g.label("ui_text_toolbar"), g.textSizePanel)
	pause := "pause"
	if g.paused {
		pause = "resume"
	}
	g.button(toolbarPause, g.label(pause), g.paused)
	footer := g.label("prototype")
	if g.hdEnabled && !g.hdMatched {
		footer = g.label("hd_fallback")
	}
	g.text(image.Rect(12, 650, 740, 678), footer)
	if g.panel && !g.speedPanel && !g.cheatPanel && !g.textSizePanel {
		draw.Draw(g.canvas, image.Rect(40, 80, 920, 630), &image.Uniform{color.RGBA{25, 32, 45, 255}}, image.Point{}, draw.Src)
		if g.keyboardPanel {
			g.drawKeyboardPanel()
		} else {
			g.text(image.Rect(60, 88, 700, 128), g.label("panel_title"))
			g.button(image.Rect(720, 88, 900, 128), g.label("keyboard_settings"), false)
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
			g.text(image.Rect(60, 180, 900, 216), g.label("pair"))
			g.text(image.Rect(60, 218, 900, 254), g.label("bind"))
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
					value = g.label("binding_waiting")
				}
				g.button(actionRect(i), g.label(a)+"："+value, g.focus == i)
			}
			g.text(image.Rect(60, 518, 900, 554), g.notice)
			g.button(image.Rect(60, 560, 360, 600), g.label("clear"), false)
			g.button(image.Rect(400, 560, 690, 600), fmt.Sprintf("%s %.2f", g.label("deadzone"), g.settings.Deadzone), false)
			g.button(image.Rect(740, 560, 900, 600), g.label("close"), false)
		}
	}
	if g.speedPanel {
		g.drawSpeedPanel()
	}
	if g.cheatPanel {
		g.drawCheatPanel()
	}
	if g.textSizePanel {
		g.drawTextSizePanel()
	}
	g.text(image.Rect(752, 650, 948, 678), fmt.Sprintf("%.2fx %.1f fps", float64(g.cycles)/float64(g.baseCycles), g.perf.FPS))
	// 診斷快照在本次合成後寫入，避免短暫姿態混到上一張HD狀態。
	if g.observePath != "" {
		churchMatched, churchFrame := false, -1
		if g.hdEnabled && g.churchHD != nil && w == 320 && h == 200 {
			churchMatched, churchFrame = g.churchHD.Match(raw, rawRGB)
		}
		observedPalette := make([]byte, 0, 768)
		for _, c := range g.o.Palette() {
			observedPalette = append(observedPalette, c[:]...)
		}
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
		data, _ := json.Marshal(map[string]any{"steps": g.o.Steps(), "cycles": g.o.Cycles(), "focused": ebiten.IsFocused(), "paused": g.paused, "panel": g.panel, "keyboard_panel": g.keyboardPanel, "keyboard_bindings": g.settings.Keyboard, "input_gate": g.releaseGate, "mode": g.o.VideoMode(), "ip": fmt.Sprint(g.o.IP()), "dialogue_rejections": rejections, "quotation": quotation, "bank_dialogue": bankDialogue, "item_purchase": itemPurchase, "item_sale": itemSale, "item_explanation": itemExplanation, "held_keys": held, "frames": g.frames,
			"ui_text_size": g.systemTextSize(), "ui_text_panel": g.textSizePanel, "ui_text_focus": g.textSizeFocus, "ui_listening": g.listening, "ui_text_overflow": g.uiTextOverflow,
			"town_dialogue": townDialogue,
			"hd_enabled":    g.hdEnabled, "town_sky_hd_matched": g.townSkyHDMatched,
			"town_background_hd_matched": g.townBackgroundMatch,
			"town_scenery_hd_matched":    g.townSceneryMatch,
			"town_frame_presentation":    g.townFrames.State(),
			"intro_threat":               g.introThreat.State(),
			"intro_story":                g.introStory.State(),
			"ending_story":               g.endingStory.State(),
			"ending_credits":             g.endingCredits.State(),
			"intro_scroll":               g.scroll.State(),
			"cave_frame_presentation":    g.caveFrames.State(),
			"cave_terrain_hd_matched":    g.caveTerrainMatch,
			"cave_player_hd":             g.cavePlayerState(),
			"cave_weapon_hd":             g.caveWeaponState(),
			"cave_objects_hd":            g.caveObjectState(),
			"cave_portal_hd":             g.cavePortalState(),
			"cave_common_hd":             g.caveCommonState(),
			"cave_magia_hd":              g.caveMagiaState(),
			"cave_door_hd":               g.caveDoorState(),
			"dialogue":                   g.dialogue.State(),
			"oracle_memory_sha256":       fmt.Sprintf("%x", sha256.Sum256(g.o.Bytes(oracle.Far(0, 0), 1<<20))),
			"oracle_indexed_sha256":      fmt.Sprintf("%x", sha256.Sum256(raw)),
			"oracle_palette_sha256":      fmt.Sprintf("%x", sha256.Sum256(observedPalette)),
			"cave_transition_hd":         g.manualArtwork.TransitionState(),
			"coherent_boss_hd":           g.manualArtwork.CoherentState(),
			"sage_hd_matched":            g.sageHDMatched,
			"sage_hd_pose":               g.sageHDPose,
			"sage_frames_hd":             g.sageFramesHD.State(),
			"merchant_frames_hd":         g.merchantFramesHD.State(),
			"input_trace_error":          g.inputTraceError,
			"audio_output":               g.output.state(),
			"cpu_effective_multiplier":   float64(g.cycles) / float64(g.baseCycles),
			"cpu_multiplier":             g.settings.CPUSpeed, "cycles_per_ms": g.cycles, "speed_panel": g.speedPanel, "cheat_panel": g.cheatPanel, "cheats": g.cheats.State(), "hd_deferred_images": hd.DeferredImageStats(),
			"performance":               map[string]any{"fps": g.perf.FPS, "draw_ms": g.perf.DrawMS, "update_ms": g.perf.UpdateMS},
			"town_ground_hd_matched":    g.townGroundMatch,
			"town_player_hd_matched":    g.townPlayerMatch,
			"town_residents_hd_matched": g.townActorsMatch,
			"town_residents_hd":         g.townActorsHD.State(),
			"town_background_cache":     g.o.Bytes(oracle.Far(0x1b1, 0xe000), 224),
			"town_background_module":    g.o.Bytes(oracle.Far(0x1b1, 0xc000), 64),
			"native_primary_0083":       g.o.Byte(oracle.Far(0x1b1, 0x0083)),
			"native_primary_0080":       g.o.Word(oracle.Far(0x1b1, 0x0080)),
			"church_hd_matched":         churchMatched,
			"church_hd_frame":           churchFrame,
			"interior_hd_matched":       g.interiorHDMatch,
			"interior_hd":               g.interiorHD.State(),
			"native_primary_0082":       g.o.Byte(oracle.Far(0x1b1, 0x0082)),
			"native_primary_0084":       g.o.Byte(oracle.Far(0x1b1, 0x0084)),
			"native_primary_0090":       g.o.Word(oracle.Far(0x1b1, 0x0090)),
			"native_primary_00b2":       g.o.Word(oracle.Far(0x1b1, 0x00b2)),
			"native_primary_4fed":       g.o.Word(oracle.Far(0x1b1, 0x4fed)),
			"text_cursor":               g.o.Word(oracle.Far(0x01b1, 0xff4c)),
			"advance_input_flags":       int(g.o.Byte(oracle.Far(0x01b1, 0xff1d))) | int(g.o.Byte(oracle.Far(0x01b1, 0xff1e)))<<8})
		// Draw 可能跨過多個 Update，不能用幀數整除判斷心跳。
		// 先寫同目錄暫存檔再換名，觀察者只會讀到完整快照。
		if err := os.WriteFile(g.observePath+".tmp", append(data, '\n'), 0644); err == nil {
			_ = os.Rename(g.observePath+".tmp", g.observePath)
		}
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
			originalMemory := g.o.Bytes(oracle.Far(0, 0), 1<<20)
			if err := os.WriteFile(g.capturePath+".memory", originalMemory, 0644); err != nil {
				log.Printf("當次原版記憶體擷取失敗：%v", err)
			}
			if presented != nil {
				for suffix, data := range map[string][]byte{".presented.indexed": presented.Indexed, ".presented.source.indexed": presented.SourceIndexed, ".presented.primary": presented.Primary, ".presented.resources": presented.Resources, ".presented.extra": presented.Extra, ".presented.palette": presented.Palette} {
					if err := os.WriteFile(g.capturePath+suffix, data, 0644); err != nil {
						log.Printf("完整畫格來源擷取失敗：%v", err)
					}
				}
			}
			if cavePresented != nil {
				if err := saveCavePaints(g.capturePath+".cave.paints.json", cavePresented.Paints); err != nil {
					log.Printf("洞窟增量繪製紀錄失敗：%v", err)
				}
				for suffix, data := range map[string][]byte{".cave.presented.indexed": cavePresented.Indexed, ".cave.source.indexed": cavePresented.SourceIndexed, ".cave.source.memory": cavePresented.Memory, ".cave.body.memory": cavePresented.BodyMemory, ".cave.palette": cavePresented.Palette} {
					if err := os.WriteFile(g.capturePath+suffix, data, 0644); err != nil {
						log.Printf("洞窟完成來源擷取失敗：%v", err)
					}
				}
			}
			if rejected := g.caveFrames.RejectedSource(); rejected != nil {
				if err := saveCavePaints(g.capturePath+".cave.rejected.paints.json", rejected.Paints); err != nil {
					log.Printf("洞窟拒絕增量紀錄失敗：%v", err)
				}
				for suffix, data := range map[string][]byte{".cave.rejected.memory": rejected.Memory, ".cave.rejected.body.memory": rejected.BodyMemory, ".cave.rejected.indexed": rejected.Indexed, ".cave.rejected.palette": rejected.Palette} {
					if err := os.WriteFile(g.capturePath+suffix, data, 0644); err != nil {
						log.Printf("洞窟拒絕來源擷取失敗：%v", err)
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
			meta, _ := json.Marshal(map[string]any{"steps": g.o.Steps(), "elapsed_cycles": g.o.Cycles(), "held_keys": g.o.HeldKeys(), "panel": g.panel, "keyboard_panel": g.keyboardPanel, "keyboard_bindings": g.settings.Keyboard, "input_gate": g.releaseGate, "mode": g.o.VideoMode(), "intro_stamps": stamps, "intro_scroll_rows": rows, "intro_scroll_sources": sources, "dialogue_glyphs": dialogueGlyphs, "dialogue_paragraphs": dialogueParagraphs, "dialogue_rejections": dialogueRejections, "quotation": quotation, "bank_dialogue": bankDialogue, "item_purchase": itemPurchase, "item_sale": itemSale, "item_explanation": itemExplanation, "hd_enabled": g.hdEnabled, "hd_matched": g.hdMatched, "king_hd_matched": g.kingHDMatched, "king_hd_pose": g.kingHDPose, "sage_hd_matched": g.sageHDMatched, "sage_hd_pose": g.sageHDPose,
				"town_dialogue": func() map[string]any {
					if g.townDialogue != nil {
						return g.townDialogue.State()
					}
					return nil
				}(),
				"original_memory_sha256": fmt.Sprintf("%x", sha256.Sum256(originalMemory)),
				"original_regs":          g.o.Regs(),
				"intro_threat":           g.introThreat.State(),
				"intro_story":            g.introStory.State(),
				"intro_scroll":           g.scroll.State(),
				"original_source_kind":   "live CPU/VGA; may be within drawing batch; distinct from cave completed/rejected source",
				"merchant_hd_matched":    g.merchantHDMatched, "merchant_hd_pose": g.merchantHDPose,
				"bank_hd_matched": g.bankHDMatched, "bank_hd_pose": g.bankHDPose,
				"item_shop_hd_matched": g.itemShopHDMatched, "item_shop_hd_pose": g.itemShopHDPose,
				"church_hd_matched": g.churchHDMatched, "church_hd_frame": g.churchHDFrame,
				"town_sky_hd_matched": g.townSkyHDMatched, "town_background_hd_matched": g.townBackgroundMatch,
				"town_scenery_hd_matched":   g.townSceneryMatch,
				"town_frame_presentation":   g.townFrames.State(),
				"cave_frame_presentation":   g.caveFrames.State(),
				"cave_terrain_hd_matched":   g.caveTerrainMatch,
				"cave_player_hd":            g.cavePlayerState(),
				"cave_weapon_hd":            g.caveWeaponState(),
				"cave_objects_hd":           g.caveObjectState(),
				"cave_portal_hd":            g.cavePortalState(),
				"cave_common_hd":            g.caveCommonState(),
				"cave_magia_hd":             g.caveMagiaState(),
				"cave_door_hd":              g.caveDoorState(),
				"dialogue":                  g.dialogue.State(),
				"input_trace_error":         g.inputTraceError,
				"audio_output":              g.output.state(),
				"town_ground_hd_matched":    g.townGroundMatch,
				"town_player_hd_matched":    g.townPlayerMatch,
				"town_residents_hd_matched": g.townActorsMatch,
				"town_residents_hd":         g.townActorsHD.State(),
				"town_background_cache":     g.o.Bytes(oracle.Far(0x1b1, 0xe000), 224), "town_background_module": g.o.Bytes(oracle.Far(0x1b1, 0xc000), 64)})
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

func (g *game) hasHDArtwork() bool {
	return len(g.hdJobs) > 0 || g.hd != nil || g.townSkyHD != nil || g.townBackgroundHD != nil || g.townSceneryHD != nil || g.caveTerrainHD != nil || g.kingHD != nil || g.sageHD != nil || g.sageFramesHD != nil || g.merchantHD != nil || g.merchantFramesHD != nil || g.bankHD != nil || g.itemShopHD != nil || g.churchHD != nil || g.interiorHD != nil || g.manualArtwork != nil || g.worldTowns != nil
}

func main() {
	if err := bundleArguments(); err != nil {
		showStartupError(fmt.Errorf("%v", err))
		return
	}
	showVersion := flag.Bool("version", false, "顯示版本")
	orig := flag.String("orig", "", "合法原版資料目錄，須含 ZELIARD.EXE")
	fontPath := flag.String("font", "", "固定字距HEX.gz字型，可附圓體灰階遮罩")
	textPath := flag.String("text", "text/host.json", "前端中文資料")
	userDir := flag.String("user-dir", "", "原版存檔與前端設定的可寫目錄")
	cpuSpeed := flag.Int("cpu-speed", 0, "CPU倍率1/2/4/8/16/32；0沿用保存設定")
	cycles := flag.Uint64("cycles", 3000, "雛形 DOSBox 相容 cycles；尚非實機速度驗收值")
	capture := flag.String("capture", "", "GUI 測試擷取路徑，完成後結束")
	captureRequest := flag.String("capture-request", "", "GUI 測試：等待指定訊號檔存在再擷取，取代更新次數條件")
	observePath := flag.String("observe", "", "可選：每秒寫入唯讀執行進度，供 GUI 驗證診斷")
	flickerProbe := flag.String("flicker-probe", "", "可選：不開視窗，以鍵序檔重播並在每個切片點合成，記錄閃爍證據")
	flickerOut := flag.String("flicker-out", "", "閃爍探針輸出前綴；檔案須不存在")
	flickerSlice := flag.Uint64("flicker-slice", 20000, "閃爍探針每次合成之間的原版指令數")
	flickerCensus := flag.Bool("flicker-census", false, "閃爍探針另外統計所有VRAM寫入者")
	flickerShots := flag.String("flicker-shots", "", "閃爍探針在這些指令數（逗號分隔、遞增）之後存下合成畫面")
	glyphPatch := flag.Bool("glyph-patch", false, "於可寫層放入patch版GMMCGA.BIN，已認領的原版字形不進畫面；規格066")
	inputTrace := flag.String("input-trace", "", "可選：本地原版輸入／週期日誌，不覆寫既有檔案")
	audioMode := flag.String("audio", "off", "原版音效試作：off 或 ebiten；正式預設待確認")
	audioWAV := flag.String("audio-wav", "", "可選：將原版生成取樣另存新WAV；不是裝置監聽錄音")
	adlib := flag.Bool("adlib", false, "使用原版 AdLib 配樂驅動；設定只寫入隔離使用者目錄")
	fmLibrary := flag.String("fm-library", "", "可替換的 Nuked-OPL3 動態函式庫")
	realtime := flag.Bool("realtime", false, "依實際時間推進，最多追趕100毫秒；暫停與失焦不追趕")
	startHD := flag.Bool("hd", false, "啟動時開啟已載入的HD圖庫，可由工具列切回原貌")
	frames := flag.Int("capture-frames", 120, "GUI 測試擷取前更新次數")
	panel := flag.Bool("show-controller", false, "GUI 測試啟動時展開手把頁")
	introRef := flag.String("intro-reference", "", "可選：本地 probe 的片頭原文參考畫格，只覆繪一句")
	introText := flag.String("intro-text", "text/intro-prototype.json", "片頭覆繪原型譯文")
	introScroll := flag.String("intro-scroll", "", "可選：首章捲動中文試譯 TSV，取代單句 prototype")
	introCredits := flag.String("intro-credits", "", "可選：製作名單中文試譯 TSV")
	introOutro := flag.String("intro-outro", "", "可選：故事後六行捲字中文試譯 TSV")
	introThreat := flag.String("intro-threat", "", "可選：魔王三行宣言中文試譯 TSV")
	introStory := flag.String("intro-story", "", "可選：完整故事中文試譯 TSV")
	endingStory := flag.String("ending-story", "", "可選：結局敘述中文 TSV（規格067）")
	endingCredits := flag.String("ending-credits", "", "可選：結局製作名單、登場魔物與版權中文 TSV（規格067）")
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
	townGroundArt := flag.String("town-ground-art", "", "可選：Muralla雙層手繪地面PNG，搭配已驗前景與town-sky-reference")
	townPlayerArt := flag.String("town-player-art-dir", "", "可選：Muralla主角八姿態手繪圖目錄，搭配已驗前景HD")
	townPlayerBackArt := flag.String("town-player-back-art", "", "可選：城鎮主角背向手繪圖目錄，搭配town-player-art-dir")
	townResidentsArt := flag.String("town-residents-art-dir", "", "可選：Muralla五類居民40姿態手繪圖集，搭配城鎮主角與前景HD")
	interiorArt := flag.String("interior-art-dir", "", "可選：小屋石像與旅館四種畫面的固定手繪圖組")
	castleMemory := flag.String("castle-memory", "", "可選：城堡正常冷新局固定1MiB參考，搭配castle-art-dir與完整遠山")
	castleArt := flag.String("castle-art-dir", "", "可選：城堡3倍世界圖與四姿態旗幟PNG目錄，搭配castle-memory")
	caveArt := flag.String("cave-art-dir", "", "可選：Malicia手繪地形圖目錄，角色與物件保留原貌")
	caveIceArt := flag.String("cave-ice-art", "", "可選：MPP4冰窟21格手繪圖集，需同時指定cave-art-dir")
	cavePeligroArt := flag.String("cave-peligro-art", "", "可選：MPP2的34格手繪圖集，需同時指定cave-art-dir")
	caveMPP3Art := flag.String("cave-mpp3-art", "", "可選：MPP3的43格手繪圖集，需同時指定cave-art-dir")
	caveMPP5Art := flag.String("cave-mpp5-art", "", "可選：MPP5的52格手繪圖集，需同時指定cave-art-dir")
	caveMPP6Art := flag.String("cave-mpp6-art", "", "可選：MPP6的36格雙材質手繪圖集，需同時指定cave-art-dir")
	caveMPP7Art := flag.String("cave-mpp7-art", "", "可選：MPP7的61槽雙材質手繪圖集，需同時指定cave-art-dir")
	caveMPP1Art := flag.String("cave-mpp1-art", "", "可選：MPP1的15至25格補充圖集，保留原有14格，需同時指定cave-art-dir")
	caveCommonArt := flag.String("cave-common-art", "", "可選：共用綠色平台64至66，需同時指定cave-art-dir")
	caveMagiaArt := flag.String("cave-magia-art", "", "可選：魔法石102手繪圖樣，需同時指定cave-art-dir")
	caveDoorArt := flag.String("cave-door-art", "", "可選：無招牌共用門框十種圖面，需同時指定cave-art-dir")
	caveMPP8Art := flag.String("cave-mpp8-art", "", "可選：MPP8的43槽雙字帶手繪圖集，需同時指定cave-art-dir")
	caveMPP9Art := flag.String("cave-mpp9-art", "", "可選：MPP9的38格手繪圖集，需同時指定cave-art-dir")
	caveMPPAArt := flag.String("cave-mppa-art", "", "可選：MPPA的20槽雙材質手繪圖集，需同時指定cave-art-dir")
	caveMPPBArt := flag.String("cave-mppb-art", "", "可選：MPPB的5格手繪圖集，需同時指定cave-art-dir")
	cavePlayerArt := flag.String("cave-player-art-dir", "", "可選：洞窟主角手繪透明圖集目錄，搭配洞窟地形HD")
	cavePlayerExtraArt := flag.String("cave-player-extra-art", "", "可選：十二個已存主角姿態及色彩變體，須同時指定cave-player-art-dir")
	cavePlayerProfileArt := flag.String("cave-player-profile-art", "", "可選：五個已驗組合姿態，須同時指定cave-player-art-dir")
	caveWeaponArt := flag.String("cave-weapon-art-dir", "", "可選：洞窟武器手繪圖集目錄，搭配洞窟地形HD")
	caveCyanWeaponArt := flag.String("cave-weapon-cyan-art", "", "可選：青藍刀光十姿態，須同時指定cave-weapon-art-dir")
	caveWeaponExtraArt := flag.String("cave-weapon-extra-art", "", "可選：四個已存刀光補圖，須同時指定cave-weapon-art-dir")
	caveObjectExtraArt := flag.String("cave-object-extra-art", "", "可選：四個物件補圖及九個遮罩對應，須同時指定cave-object-art-dir")
	manualArt := flag.String("manual-hd-art", "", "可選：050全圖像人工驗證候選圖組")
	worldTownArt := flag.String("world-town-art", "", "可選：050後續城鎮人工驗證候選圖組")
	caveObjectArt := flag.String("cave-object-art-dir", "", "可選：洞窟動態物件手繪圖集目錄，搭配洞窟地形HD")
	cavePortalArt := flag.String("cave-portal-art", "", "可選：Muralla洞窟入口手繪圖，搭配洞窟地形HD")
	townSignsText := flag.String("town-signs-text", "", "可選：Muralla招牌中文TSV，搭配前景HD")
	kingRefA := flag.String("king-reference-a", "", "可選：本地原版國王姿態 A 的 screen.bin")
	kingRefB := flag.String("king-reference-b", "", "可選：本地原版國王姿態 B 的 screen.bin")
	kingArt := flag.String("king-art", "", "可選：本地手繪國王閉口 PNG")
	kingSpeakingArt := flag.String("king-speaking-art", "", "可選：本地手繪國王說話 PNG")
	sageRefA := flag.String("sage-reference-a", "", "可選：本地原版長老畫格 A")
	sageRefB := flag.String("sage-reference-b", "", "可選：本地原版長老畫格 B")
	sagePalette := flag.String("sage-palette", "", "可選：本地原版長老色盤")
	sageArt := flag.String("sage-art", "", "可選：本地手繪長老 PNG")
	sageFramesArt := flag.String("sage-frames-art-dir", "", "可選：長老八手勢與高清動畫的固定圖組")
	merchantRefA := flag.String("merchant-reference-a", "", "可選：武器商人原版迎客畫格")
	merchantRefB := flag.String("merchant-reference-b", "", "可選：武器商人原版告別畫格")
	merchantRefC := flag.String("merchant-reference-c", "", "可選：武器商人未購物告別原版畫格")
	merchantPalette := flag.String("merchant-palette", "", "可選：武器商人原版色盤")
	merchantArtA := flag.String("merchant-art-a", "", "可選：本地手繪武器商人迎客 PNG")
	merchantArtB := flag.String("merchant-art-b", "", "可選：本地手繪武器商人告別 PNG")
	merchantArtC := flag.String("merchant-art-c", "", "可選：本地手繪武器商人未購物告別 PNG")
	merchantFramesArt := flag.String("merchant-frames-art-dir", "", "可選：武器店八姿態及兩局部動畫的固定圖組")
	bankRefA := flag.String("bank-reference-a", "", "可選：銀行原版畫格 A")
	bankRefB := flag.String("bank-reference-b", "", "可選：銀行原版畫格 B")
	bankRefC := flag.String("bank-reference-c", "", "可選：銀行原版畫格 C")
	bankPalette := flag.String("bank-palette", "", "可選：銀行原版色盤")
	bankArtA := flag.String("bank-art-a", "", "可選：本地手繪銀行畫格 A PNG")
	bankArtB := flag.String("bank-art-b", "", "可選：本地手繪銀行畫格 B PNG")
	bankArtC := flag.String("bank-art-c", "", "可選：本地手繪銀行畫格 C PNG")
	churchReferences := flag.String("church-reference-dir", "", "可選：教堂已核對原版房間畫格與色盤目錄")
	churchArtwork := flag.String("church-art-dir", "", "可選：教堂固定手繪素材目錄")
	var itemRefs, itemArts [4]*string
	for i := range itemRefs {
		itemRefs[i] = flag.String(fmt.Sprintf("item-reference-%d", i), "", "可選：道具店已核對原版畫格")
		itemArts[i] = flag.String(fmt.Sprintf("item-art-%d", i), "", "可選：道具店圖組，依序基準／蒸氣B／抬手／蒸氣C")
	}
	itemPalette := flag.String("item-palette", "", "可選：道具店原版色盤")
	flag.Parse()
	if *showVersion {
		fmt.Println(version)
		return
	}
	if *audioMode != "off" && *audioMode != "ebiten" {
		showStartupError(fmt.Errorf("%v", "-audio 須為 off 或 ebiten"))
		return
	}
	if *orig == "" || *fontPath == "" || *userDir == "" || *cycles < 60 || *cycles > 10000 || *frames < 1 {
		showStartupError(fmt.Errorf("%v", "須指定 -orig、-font、-user-dir；cycles 範圍 60–10000"))
		return
	}
	root, err := filepath.Abs(*orig)
	if err != nil {
		showStartupError(fmt.Errorf("%v", err))
		return
	}
	userRoot, err := filepath.Abs(*userDir)
	if err != nil {
		showStartupError(fmt.Errorf("%v", err))
		return
	}
	if insideDir(root, userRoot) {
		showStartupError(fmt.Errorf("使用者目錄不可位於原版資料目錄內：使用者目錄 %s；原版目錄 %s", userRoot, root))
		return
	}
	if *audioWAV != "" {
		path, e := filepath.Abs(*audioWAV)
		if e != nil {
			showStartupError(fmt.Errorf("%v", e))
			return
		}
		if insideDir(root, path) {
			showStartupError(fmt.Errorf("原版音訊記錄不可位於原版資料目錄內：%s；原版目錄 %s", path, root))
			return
		}
	}
	if err := os.MkdirAll(filepath.Join(userRoot, "game"), 0700); err != nil {
		showStartupError(fmt.Errorf("%v", err))
		return
	}

	logPath := filepath.Join(userRoot, "startup.log")
	logFile, logErr := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if logErr == nil {
		defer logFile.Close()
		log.SetOutput(io.MultiWriter(os.Stderr, logFile))
	}
	log.Printf("啟動 %s", version)
	log.Printf("原版目錄 %s；使用者目錄 %s", root, userRoot)
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("神劍封魔：啟動中")
	load := func(report func(string, bitmap.Font)) (*game, error) {
		var closers []func()
		ready := false
		defer func() {
			if !ready {
				for i := len(closers) - 1; i >= 0; i-- {
					closers[i]()
				}
			}
		}()
		report("操作設定", nil)
		settings, err := controls.Load(filepath.Join(userRoot, "controller.json"))
		if err != nil {
			return nil, err
		}
		report("字型與介面", nil)
		font, err := bitmap.Load(*fontPath)
		if err != nil {
			return nil, err
		}
		report("字型已載入", font)
		data, err := os.ReadFile(*textPath)
		if err != nil {
			return nil, err
		}
		labels := map[string]string{}
		if err = json.Unmarshal(data, &labels); err != nil {
			return nil, err
		}
		for _, s := range labels {
			for _, r := range s {
				if font[r] == nil {
					return nil, fmt.Errorf("字型缺字 U+%04X", r)
				}
			}
		}
		report("原版執行器", nil)
		o, err := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if err != nil {
			return nil, err
		}
		closers = append(closers, func() { o.Close() })
		o.SetScratch(filepath.Join(userRoot, "game"))
		// 規格066：patch版驅動只放在可寫層，原版目錄不動；關閉時只移除本工具產生的檔案。
		if *glyphPatch {
			sum, e := glyphpatch.Install(root, filepath.Join(userRoot, "game"))
			if e != nil {
				log.Printf("字形patch未啟用，保留原版繪字：%v", e)
			} else {
				glyphpatch.Enable(o)
				log.Printf("字形patch已啟用：%s %s", glyphpatch.DriverName, sum)
			}
		} else if e := glyphpatch.Remove(root, filepath.Join(userRoot, "game")); e != nil {
			return nil, e
		}
		if *adlib {
			config, e := os.ReadFile(filepath.Join(root, "RESOURCE.CFG"))
			if e != nil {
				return nil, e
			}
			config = []byte(strings.ReplaceAll(string(config), "MSCSTD.DRV", "MSCADLIB.DRV"))
			if e = os.WriteFile(filepath.Join(userRoot, "game/RESOURCE.CFG"), config, 0600); e != nil {
				return nil, e
			}
			o.SetAdLib(true)
		}
		if *cpuSpeed != 0 {
			if !controls.ValidCPUSpeed(*cpuSpeed) {
				return nil, fmt.Errorf("-cpu-speed只接受1/2/4/8/16/32")
			}
			settings.CPUSpeed = *cpuSpeed
		}
		effectiveCycles := *cycles
		o.SetDOSBoxCycles(effectiveCycles)
		g := &game{o: o, audio: o.NewAudio(44100), font: font, labels: labels, settings: settings, settingsPath: filepath.Join(userRoot, "controller.json"), selected: -1, baseCycles: *cycles, cycles: effectiveCycles, panel: *panel, capturePath: *capture, captureRequest: *captureRequest, captureFrame: *frames, canvas: image.NewRGBA(image.Rect(0, 0, width, height)), audioMode: *audioMode, audioPath: *audioWAV}
		g.roomPrefixes = loadRoomPrefixes(filepath.Join(root, "ZELRES2.SAR"))
		g.selectPrefix = loadSelectPrefix(filepath.Join(root, "ZELRES1.SAR"))
		g.endingPrefix = loadEndingPrefix(filepath.Join(root, "ZELRES2.SAR"))
		g.cheats, err = cheats.New(o, filepath.Join(root, "ZELIARD.EXE"))
		if err != nil {
			log.Printf("作弊功能未啟用：%v", err)
		} else {
			g.cheats.DisableRegeneration = settings.DisableStandingRegen
		}
		report("HD補圖索引", nil)
		if *manualArt != "" {
			g.manualArtwork, err = hd.NewManualArtworkLazy(root, *manualArt)
			if err != nil {
				log.Printf("人工候選圖格未載入，保留原貌：%v", err)
			}
		}
		if *worldTownArt != "" {
			g.worldTowns, err = hd.NewWorldTownsLazy(root, *worldTownArt)
			if err != nil {
				log.Printf("後續城鎮候選未載入，保留原貌：%v", err)
			}
		}
		if *fmLibrary != "" {
			synth, e := fm.Open(*fmLibrary, 44100)
			if e != nil {
				return nil, e
			}
			closers = append(closers, func() { synth.Close() })
			g.audio.SetFMSynth(synth)
		}
		g.observePath = *observePath
		g.realtime = *realtime
		if *inputTrace != "" {
			trace, err := os.OpenFile(*inputTrace, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
			if err != nil {
				return nil, err
			}
			if err = trace.Close(); err != nil {
				return nil, err
			}
			g.inputTrace = *inputTrace
			g.trace(map[string]any{"kind": "header", "cycles_per_ms": effectiveCycles, "audio_sample_rate": 44100, "initial_memory_sha256": fmt.Sprintf("%x", sha256.Sum256(o.Bytes(oracle.Far(0, 0), 1<<20))), "initial_regs": o.Regs(), "cheats": g.cheats.State(), "source_kind": "native key edges and RunCycles journal; active cheats are not original oracle evidence"})
		}
		report("登錄洞窟圖組", nil)
		if *caveArt != "" {
			g.deferHD("洞窟基本圖組", g.caveModule, func() {
				if *caveArt != "" {
					g.caveTerrainHD, err = hd.NewCaveTerrain(*caveArt)
					if err != nil {
						log.Printf("洞窟地形HD未載入，保留原貌：%v", err)
					} else {

						for _, extra := range []struct {
							name, dir string
							load      func(string) error
						}{
							{"MPP1", *caveMPP1Art, g.caveTerrainHD.LoadMPP1},
							{"共用平台", *caveCommonArt, g.caveTerrainHD.LoadCommon},
							{"魔法石", *caveMagiaArt, g.caveTerrainHD.LoadMagia},
							{"共用門框", *caveDoorArt, g.caveTerrainHD.LoadDoors},
						} {
							if extra.dir != "" {
								if e := extra.load(extra.dir); e != nil {
									log.Printf("%s地形HD未載入，保留該家族原貌：%v", extra.name, e)
								}
							}
						}
						if *caveObjectArt != "" {
							objects, e := hd.NewCaveObjects(*caveObjectArt)
							if e != nil {
								log.Printf("洞窟物件HD未載入，保留物件原貌：%v", e)
							} else {
								if *caveObjectExtraArt != "" {
									if err := objects.LoadExtra(*caveObjectExtraArt); err != nil {
										log.Printf("物件補圖未載入，保留原有圖組：%v", err)
									}
								}
								if *manualArt != "" {
									if e := objects.LoadManualLazy(*manualArt); e != nil {
										log.Printf("敵人與首領候選未載入：%v", e)
									}
								}
								g.caveTerrainHD.SetObjects(objects)
							}
						}
						if *caveWeaponArt != "" {
							weapon, e := hd.NewCaveWeapon(*caveWeaponArt)
							if e != nil {
								log.Printf("洞窟武器HD未載入，保留武器原貌：%v", e)
							} else {
								if *caveCyanWeaponArt != "" {
									if err := weapon.LoadCyan(*caveCyanWeaponArt); err != nil {
										log.Printf("青藍刀光未載入，保留原有武器圖組：%v", err)
									}
								}
								if *caveWeaponExtraArt != "" {
									if err := weapon.LoadExtra(*caveWeaponExtraArt); err != nil {
										log.Printf("刀光補圖未載入，保留原有圖組：%v", err)
									}
								}
								weapon.SetCompleteArtwork(g.manualArtwork)
								g.caveTerrainHD.SetWeapon(weapon)
							}
						}
						if *cavePlayerArt != "" {
							player, e := hd.NewCavePlayer(*cavePlayerArt)
							if e != nil {
								log.Printf("洞窟主角HD未載入，保留角色原貌：%v", e)
							} else {
								if *cavePlayerExtraArt != "" {
									if err := player.LoadExtra(*cavePlayerExtraArt); err != nil {
										log.Printf("主角擴充圖未載入，保留原有姿態：%v", err)
									}
								}
								if *cavePlayerProfileArt != "" {
									if err := player.LoadProfiles(*cavePlayerProfileArt); err != nil {
										log.Printf("主角組合圖未載入，保留原有姿態：%v", err)
									}
								}
								player.SetCompleteArtwork(g.manualArtwork)
								g.caveTerrainHD.SetPlayer(player)
							}
						}
					}
				}
				if *cavePortalArt != "" && g.caveTerrainHD != nil {
					title := "MURALLA"
					if translated, ok := g.hud.Translation("zelres2.file.0399b8"); ok {
						title = translated
					}
					portal, e := hd.NewCavePortal(*cavePortalArt, font, title)
					if e != nil {
						log.Printf("入口HD未載入，保留原圖：%v", e)
					} else {
						g.caveTerrainHD.SetPortal(portal)
					}
				}
			}, func() {
				if g.caveTerrainHD != nil {
					g.caveFrames = hd.NewCaveFrames(g.o)
					g.caveFrames.HoldFreshBody(true)
				}
			})
		}
		if *caveIceArt != "" {
			g.deferHD("冰窟地形", func() bool { return g.caveFamily() == "MPP4" }, func() {
				if e := g.caveTerrainHD.LoadIce(*caveIceArt); e != nil {
					log.Printf("冰窟圖組未載入：%v", e)
				}
			}, nil)
		}
		if *cavePeligroArt != "" {
			g.deferHD("Peligro地形", func() bool { return g.caveFamily() == "MPP2" }, func() {
				if e := g.caveTerrainHD.LoadPeligro(*cavePeligroArt); e != nil {
					log.Printf("Peligro圖組未載入：%v", e)
				}
			}, nil)
		}
		if *caveMPP3Art != "" {
			g.deferHD("MPP3地形", func() bool { return g.caveFamily() == "MPP3" }, func() {
				if e := g.caveTerrainHD.LoadMPP3(*caveMPP3Art); e != nil {
					log.Printf("MPP3圖組未載入：%v", e)
				}
			}, nil)
		}
		if *caveMPP5Art != "" {
			g.deferHD("MPP5地形", func() bool { return g.caveFamily() == "MPP5" }, func() {
				if e := g.caveTerrainHD.LoadMPP5(*caveMPP5Art); e != nil {
					log.Printf("MPP5圖組未載入：%v", e)
				}
			}, nil)
		}
		if *caveMPP6Art != "" {
			g.deferHD("MPP6地形", func() bool { return g.caveFamily() == "MPP6" }, func() {
				if e := g.caveTerrainHD.LoadMPP6(*caveMPP6Art); e != nil {
					log.Printf("MPP6圖組未載入：%v", e)
				}
			}, nil)
		}
		if *caveMPP7Art != "" {
			g.deferHD("MPP7地形", func() bool { return g.caveFamily() == "MPP7" }, func() {
				if e := g.caveTerrainHD.LoadMPP7(*caveMPP7Art); e != nil {
					log.Printf("MPP7圖組未載入：%v", e)
				}
			}, nil)
		}
		if *caveMPP8Art != "" {
			g.deferHD("MPP8地形", func() bool { return g.caveFamily() == "MPP8" }, func() {
				if e := g.caveTerrainHD.LoadMPP8(*caveMPP8Art); e != nil {
					log.Printf("MPP8圖組未載入：%v", e)
				}
			}, nil)
		}
		if *caveMPP9Art != "" {
			g.deferHD("MPP9地形", func() bool { return g.caveFamily() == "MPP9" }, func() {
				if e := g.caveTerrainHD.LoadMPP9(*caveMPP9Art); e != nil {
					log.Printf("MPP9圖組未載入：%v", e)
				}
			}, nil)
		}
		if *caveMPPAArt != "" {
			g.deferHD("MPPA地形", func() bool { return g.caveFamily() == "MPPA" }, func() {
				if e := g.caveTerrainHD.LoadMPPA(*caveMPPAArt); e != nil {
					log.Printf("MPPA圖組未載入：%v", e)
				}
			}, nil)
		}
		if *caveMPPBArt != "" {
			g.deferHD("MPPB地形", func() bool { return g.caveFamily() == "MPPB" }, func() {
				if e := g.caveTerrainHD.LoadMPPB(*caveMPPBArt); e != nil {
					log.Printf("MPPB圖組未載入：%v", e)
				}
			}, nil)
		}

		report("登錄城鎮圖組", nil)
		if *townBackgroundMemory != "" || *townSceneryArt != "" || *castleArt != "" || *hdArt != "" || *townSkyArt != "" {
			// 規格067：開局由選單模組逐格畫兩側邊框，邊框 HD 須在揭開前就緒。
			g.deferHD("城鎮與城堡HD", func() bool { return g.townModule() || g.selectModuleLoaded() }, func() {
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

				}
				if *castleMemory != "" || *castleArt != "" {
					var castle *hd.TownScenery
					castle, err = hd.NewCastleScenery(root, *castleMemory, *castleArt, g.townBackgroundHD)
					if err == nil {
						if g.townSceneryHD == nil {
							g.townSceneryHD = castle

						} else {
							err = g.townSceneryHD.AttachCastle(castle)
						}
					}
					if err != nil {
						log.Printf("城堡HD未載入，保留既有圖面：%v", err)
					}
				}
				if *townGroundArt != "" {
					g.townGroundHD, err = hd.NewTownGround(g.townSceneryHD, *townSkyRef, *townGroundArt)
					if err != nil {
						log.Printf("城鎮地面HD未載入，保留原貌：%v", err)
					}
				}
				if *townPlayerArt != "" {
					g.townPlayerHD, err = hd.NewTownPlayer(g.townSceneryHD, *townPlayerArt)
					if err != nil {
						log.Printf("城鎮主角HD未載入，保留原貌：%v", err)
					}
				}
				if *townPlayerBackArt != "" {
					if err = g.townPlayerHD.LoadBackPose(*townPlayerBackArt); err != nil {
						log.Printf("城鎮背向HD未載入，保留既有圖面：%v", err)
					}
				}
				if *townResidentsArt != "" {
					g.townActorsHD, err = hd.NewTownActors(g.townPlayerHD, *townResidentsArt)
					if err != nil {
						log.Printf("城鎮居民HD未載入，保留既有圖面：%v", err)
					}
				}
			}, func() {
				if g.townSceneryHD != nil {
					g.townFrames = hd.NewTownFrames(g.o, g.townSceneryHD)
					if g.townFrames != nil {
						g.townFrames.Menu = g.selectModuleLoaded
					}
					if g.townFrames != nil && g.townDialogue != nil {
						g.townFrames.Dialogue = func() (bool, image.Rectangle) {
							return g.townDialogue.State()["key"] != "", g.townDialogue.BackdropBox()
						}
					}
				}
			})
		}

		if *interiorArt != "" || *kingArt != "" || *sageFramesArt != "" || *sageArt != "" || *merchantFramesArt != "" || *merchantArtA != "" || *bankArtA != "" || *churchArtwork != "" || *itemPalette != "" {
			g.deferHD("人物與室內HD", g.roomEntered, func() {
				if *interiorArt != "" {
					g.interiorHD, err = hd.NewInteriors(root, *interiorArt)
					if err != nil {
						log.Printf("小屋與旅館HD未載入，保留原貌：%v", err)
					}
				}

				if *kingRefA != "" || *kingRefB != "" || *kingArt != "" || *kingSpeakingArt != "" {
					g.kingHD, err = hd.NewKing(*kingRefA, *kingRefB, filepath.Join(filepath.Dir(*kingRefA), "palette.bin"), *kingArt, *kingSpeakingArt)
					if err != nil {
						log.Printf("國王 HD 圖組未載入，保留原貌：%v", err)
					}
				}
				if *sageFramesArt != "" {
					g.sageFramesHD, err = hd.NewSageFrames(root, *sageFramesArt)
					if err != nil {
						log.Printf("長老完整手勢圖組未載入，保留原貌：%v", err)
					}
				} else if *sageRefA != "" || *sageRefB != "" || *sagePalette != "" || *sageArt != "" {
					g.sageHD, err = hd.NewSage(*sageRefA, *sageRefB, *sagePalette, *sageArt)
					if err != nil {
						log.Printf("長老 HD 圖面未載入，保留原貌：%v", err)
					}
				}
				if *merchantFramesArt != "" {
					g.merchantFramesHD, err = hd.NewMerchantFrames(root, *merchantFramesArt)
					if err != nil {
						log.Printf("武器店完整圖組未載入，保留原貌：%v", err)
					}
				} else if *merchantRefA != "" || *merchantRefB != "" || *merchantRefC != "" || *merchantPalette != "" || *merchantArtA != "" || *merchantArtB != "" || *merchantArtC != "" {
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
				if *churchReferences != "" || *churchArtwork != "" {
					g.churchHD, err = hd.NewChurch(*churchReferences, *churchArtwork)
					if err != nil {
						log.Printf("教堂HD圖組未載入，保留原貌：%v", err)
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
			}, nil)
		}

		report("中文翻譯", nil)
		if *introStory != "" {
			g.introStory, err = overlay.NewIntroStory(o, filepath.Join(root, "ZELRES1.SAR"), *introStory, font)
			if err != nil {
				log.Printf("故事中文未載入，保留原文：%v", err)
			}
		}
		if *endingStory != "" {
			g.endingStory, err = overlay.NewEndingStory(o, filepath.Join(root, "ZELRES2.SAR"), *endingStory, font)
			if err != nil {
				log.Printf("結局敘述中文未載入，保留原文：%v", err)
			}
		}
		if *endingCredits != "" {
			g.endingCredits, err = overlay.NewEndingCredits(o, filepath.Join(root, "ZELRES2.SAR"), *endingCredits, font)
			if err != nil {
				log.Printf("結局名單中文未載入，保留原文：%v", err)
			}
		}
		if *introThreat != "" {
			g.introThreat, err = overlay.NewIntroThreat(o, filepath.Join(root, "ZELRES1.SAR"), *introThreat, font)
			if err != nil {
				log.Printf("魔王宣言中文未載入，保留原文：%v", err)
			}
		}
		if *introScroll != "" || *introCredits != "" || *introOutro != "" {
			g.scroll, err = overlay.NewScrollBlocks(o, filepath.Join(root, "ZELRES1.SAR"), overlay.ScrollCatalogs{First: *introScroll, Credits: *introCredits, Outro: *introOutro}, font)
			if err != nil {
				log.Printf("片頭捲動中文未載入，保留原文：%v", err)
			}
		} else if *introRef != "" {
			g.intro, err = overlay.NewIntro(*introRef, *introText, font)
			if err != nil {
				return nil, err
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
		// 規格066：記錄實際載入的譯表路徑與雜湊，回報與探針先核這一行。
		for _, path := range []string{*dialogueText, *labelsText, *hudText, *townText} {
			if path != "" {
				log.Printf("譯表 %s", catalogIdentity(path))
			}
		}
		if *hudText != "" {
			g.hud, err = overlay.NewHUD(o, root, *hudText, font)
			if err != nil {
				log.Printf("狀態列試譯未載入，保留原文：%v", err)
			}
		}

		if *townText != "" {
			g.townDialogue, err = overlay.NewTownDialogue(o, filepath.Join(root, "ZELRES2.SAR"), *townText, font)
			if err != nil {
				log.Printf("居民對話試譯未載入，保留原文：%v", err)
			}
		}
		g.hdEnabled = *startHD && g.hasHDArtwork()
		ready = true
		g.closeResources = func() {
			for i := len(closers) - 1; i >= 0; i-- {
				closers[i]()
			}
		}
		return g, nil

	}
	if *flickerProbe != "" {
		g, err := load(func(string, bitmap.Font) {})
		if err != nil {
			log.Fatal(err)
		}
		var shots []uint64
		for _, f := range strings.Split(*flickerShots, ",") {
			if f = strings.TrimSpace(f); f != "" {
				v, e := strconv.ParseUint(f, 0, 64)
				if e != nil {
					log.Fatal(e)
				}
				shots = append(shots, v)
			}
		}
		err = g.runFlickerProbe(*flickerProbe, *flickerOut, *flickerSlice, *flickerCensus, shots)
		g.closeResources()
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	boot := newStartup(logPath, load)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetRunnableOnUnfocused(true)
	runErr := ebiten.RunGame(boot)
	closeErr := boot.Close()
	if err := errors.Join(runErr, closeErr); err != nil {
		log.Fatal(err)
	}
}

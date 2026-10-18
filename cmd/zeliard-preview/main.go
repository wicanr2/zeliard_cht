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
	captured             bool
	captureTail          int
	intro                *xlate.Layer
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
	draw.Draw(g.canvas, g.canvas.Bounds(), &image.Uniform{color.RGBA{18, 23, 32, 255}}, image.Point{}, draw.Src)
	w, h, rgb := g.o.ScreenRGB()
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
	if g.intro != nil && w == 320 && h == 200 {
		g.intro.Frame(g.o.Indexed(), rgb)
		clear(g.overlayPixels)
		if g.intro.Draw(g.overlayPixels, scale, nil) {
			src := &image.RGBA{Pix: g.overlayPixels, Stride: width * 4, Rect: image.Rect(0, 0, width, 600)}
			draw.Draw(g.canvas, image.Rect(0, barHeight, width, barHeight+600), src, image.Point{}, draw.Over)
		}
	}
	g.button(image.Rect(8, 8, 150, 40), g.label("original"), true)
	g.button(image.Rect(158, 8, 340, 40), g.label("hd"), false)
	g.button(image.Rect(348, 8, 540, 40), g.label("controller"), g.panel)
	pause := "pause"
	if g.paused {
		pause = "resume"
	}
	g.button(image.Rect(548, 8, 740, 40), g.label(pause), g.paused)
	g.text(image.Rect(12, 650, 948, 678), g.label("prototype"))
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
			stamps := 0
			if g.intro != nil {
				stamps = len(g.intro.Stamps)
			}
			meta, _ := json.Marshal(map[string]any{"steps": g.o.Steps(), "elapsed_cycles": g.o.Cycles(), "held_keys": g.o.HeldKeys(), "panel": g.panel, "mode": g.o.VideoMode(), "intro_stamps": stamps})
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
	frames := flag.Int("capture-frames", 120, "GUI 測試擷取前更新次數")
	panel := flag.Bool("show-controller", false, "GUI 測試啟動時展開手把頁")
	introRef := flag.String("intro-reference", "", "可選：本地 probe 的片頭原文參考畫格，只覆繪一句")
	introText := flag.String("intro-text", "text/intro-prototype.json", "片頭覆繪原型譯文")
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
	if *introRef != "" {
		g.intro, err = overlay.NewIntro(*introRef, *introText, font)
		if err != nil {
			log.Fatal(err)
		}
		g.overlayPixels = make([]byte, width*600*4)
	}
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle(g.label("title"))
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetRunnableOnUnfocused(true)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

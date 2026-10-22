package main

import (
	"fmt"
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/zeliard_cht/internal/prototype/controls"
)

var toolbarOriginal = image.Rect(8, 8, 112, 40)
var toolbarHD = image.Rect(120, 8, 256, 40)
var toolbarSpeed = image.Rect(264, 8, 400, 40)
var toolbarController = image.Rect(408, 8, 528, 40)
var toolbarPause = image.Rect(608, 8, 736, 40)
var toolbarKeyboard = image.Rect(744, 8, 856, 40)

func speedRect(i int) image.Rectangle { return image.Rect(300, 170+i*50, 660, 212+i*50) }

func (g *game) openSpeedPanel() {
	g.panel, g.speedPanel, g.keyboardPanel = true, true, false
	g.cheatPanel = false
	g.textSizePanel = false
	g.listening, g.notice = "", ""
	g.releaseGate = true
	g.edges.Sync(g, nil)
	for i, multiplier := range controls.CPUSpeeds {
		if multiplier == g.settings.CPUSpeed {
			g.speedFocus = i
		}
	}
}

func (g *game) setCPUSpeed(multiplier int) error {
	if !controls.ValidCPUSpeed(multiplier) {
		return fmt.Errorf("不支援的CPU倍率：%d", multiplier)
	}
	g.edges.Sync(g, nil)
	values := g.audio.Render()
	g.trace(map[string]any{"kind": "audio"})
	if err := g.output.submit(values); err != nil {
		return err
	}
	if err := g.output.suspend(); err != nil {
		return err
	}
	g.settings.CPUSpeed = multiplier
	// Start with the already sustainable rate; raise toward the new ceiling gradually.
	if ceiling := g.baseCycles * uint64(multiplier); g.cycles > ceiling {
		g.cycles = ceiling
	}
	g.o.SetDOSBoxCycles(g.cycles)
	g.clock.Reset()
	g.releaseGate = true
	g.lastRateChange = time.Now()
	g.sliceRan, g.sliceElapsed = 0, 0
	g.trace(map[string]any{"kind": "cpu_speed", "multiplier": multiplier, "cycles_per_ms": g.cycles, "steps": g.o.Steps(), "cycles": g.o.Cycles()})
	g.save()
	return nil
}

func (g *game) updateSpeedPanel(point image.Point, click bool, just func(string) bool) error {
	g.edges.Sync(g, nil)
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || click && point.In(image.Rect(740, 560, 900, 600)) {
		g.panel, g.speedPanel, g.releaseGate = false, false, true
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || just("down") {
		g.speedFocus = (g.speedFocus + 1) % len(controls.CPUSpeeds)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || just("jump") {
		g.speedFocus = (g.speedFocus + len(controls.CPUSpeeds) - 1) % len(controls.CPUSpeeds)
	}
	apply := inpututil.IsKeyJustPressed(ebiten.KeyEnter) || just("attack")
	for i := range controls.CPUSpeeds {
		if click && point.In(speedRect(i)) {
			g.speedFocus, apply = i, true
		}
	}
	if apply {
		if err := g.setCPUSpeed(controls.CPUSpeeds[g.speedFocus]); err != nil {
			return err
		}
		g.panel, g.speedPanel = false, false
	}
	return nil
}

func (g *game) drawSpeedPanel() {
	g.button(image.Rect(40, 80, 920, 630), "", false)
	g.text(image.Rect(60, 90, 900, 122), g.label("cpu_title"))
	g.text(image.Rect(60, 130, 900, 166), fmt.Sprintf("%s：%d cycles/ms　%s %.2fx", g.label("cpu_base"), g.baseCycles, g.label("cpu_effective"), float64(g.cycles)/float64(g.baseCycles)))
	for i, multiplier := range controls.CPUSpeeds {
		label := fmt.Sprintf("%dx", multiplier)
		if multiplier == 1 {
			label = g.label("cpu_original") + " 1x"
		}
		if multiplier == g.settings.CPUSpeed {
			label += "　" + g.label("cpu_current")
		}
		g.button(speedRect(i), label, i == g.speedFocus)
	}
	g.text(image.Rect(60, 478, 900, 514), g.label("cpu_help"))
	g.text(image.Rect(60, 518, 900, 554), g.label("cpu_scope"))
	g.button(image.Rect(740, 560, 900, 600), g.label("close"), false)
}

type performance struct {
	anchor                time.Time
	frames, updates       int
	drawTime, updateTime  time.Duration
	FPS, DrawMS, UpdateMS float64
}

func (p *performance) draw(begin time.Time) {
	p.drawTime += time.Since(begin)
	p.frames++
	if p.anchor.IsZero() {
		p.anchor = begin
	}
	if elapsed := time.Since(p.anchor); elapsed >= time.Second {
		p.FPS = float64(p.frames) / elapsed.Seconds()
		p.DrawMS = float64(p.drawTime.Nanoseconds()) / 1e6 / float64(p.frames)
		if p.updates > 0 {
			p.UpdateMS = float64(p.updateTime.Nanoseconds()) / 1e6 / float64(p.updates)
		}
		p.anchor = time.Now()
		p.frames, p.updates, p.drawTime, p.updateTime = 0, 0, 0, 0
	}
}

func (p *performance) update(begin time.Time) { p.updateTime += time.Since(begin); p.updates++ }

// A frame may contain several Update calls. Share its CPU allowance so catch-up
// cannot delay the next HD presentation with a chain of long RunCycles calls.
func (g *game) runCPUSlice(budget uint64) error {
	allowance := controls.CPUFrameAllowance(g.lastDrawCost)
	remaining := allowance - g.cpuFrameUsed
	if remaining <= 0 || budget == 0 {
		return nil
	}
	begin := time.Now()
	before := g.o.Cycles()
	deadline := begin.Add(remaining)
	for g.o.Cycles()-before < budget {
		n := budget - (g.o.Cycles() - before)
		if n > 2048 {
			n = 2048
		}
		steps, cycles := g.o.Steps(), g.o.Cycles()
		err := g.o.RunCycles(n)
		g.traceRun(steps, cycles, n, false, err)
		if err != nil {
			g.cpuFrameUsed += time.Since(begin)
			return err
		}
		if !time.Now().Before(deadline) {
			break
		}
	}
	elapsed := time.Since(begin)
	g.cpuFrameUsed += elapsed
	g.sliceRan += g.o.Cycles() - before
	g.sliceElapsed += elapsed
	return nil
}

// Audio.Render must run before this call, while cycles still has its old rate.
func (g *game) adaptCPU() {
	if g.sliceElapsed == 0 || time.Since(g.lastRateChange) < 200*time.Millisecond {
		return
	}
	allowance := controls.CPUFrameAllowance(g.lastDrawCost)
	next := controls.SustainableCPURate(g.cycles, g.baseCycles*uint64(g.settings.CPUSpeed), g.sliceRan, g.sliceElapsed, allowance)
	g.sliceRan, g.sliceElapsed = 0, 0
	if next == g.cycles {
		return
	}
	g.clock.RateChanged(g.o.Cycles(), g.cycles)
	g.cycles = next
	g.o.SetDOSBoxCycles(next)
	g.lastRateChange = time.Now()
	g.trace(map[string]any{"kind": "cpu_speed", "multiplier": g.settings.CPUSpeed, "cycles_per_ms": next, "steps": g.o.Steps(), "cycles": g.o.Cycles(), "automatic": true})
}

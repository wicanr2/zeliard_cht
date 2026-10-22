package main

import (
	"fmt"
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/zeliard_cht/internal/prototype/controls"
)

var toolbarTextSize = image.Rect(864, 8, 952, 40)

func textSizeRect(i int) image.Rectangle { return image.Rect(160, 225+i*82, 800, 293+i*82) }

func (g *game) systemTextSize() int {
	if controls.ValidUITextSize(g.settings.UITextSize) {
		return g.settings.UITextSize
	}
	return 24
}

func (g *game) openTextSizePanel() {
	g.panel, g.textSizePanel = true, true
	g.cheatPanel, g.keyboardPanel, g.speedPanel = false, false, false
	g.listening, g.notice = "", ""
	g.releaseGate = true
	g.edges.Sync(g, nil)
	for i, size := range controls.UITextSizes {
		if size == g.systemTextSize() {
			g.textSizeFocus = i
		}
	}
}

func (g *game) updateTextSizePanel(point image.Point, click bool, just func(string) bool) error {
	g.edges.Sync(g, nil)
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || click && point.In(image.Rect(740, 560, 900, 600)) {
		g.panel, g.textSizePanel, g.releaseGate = false, false, true
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || just("down") {
		g.textSizeFocus = (g.textSizeFocus + 1) % len(controls.UITextSizes)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || just("jump") {
		g.textSizeFocus = (g.textSizeFocus + len(controls.UITextSizes) - 1) % len(controls.UITextSizes)
	}
	apply := inpututil.IsKeyJustPressed(ebiten.KeyEnter) || just("attack")
	for i := range controls.UITextSizes {
		if click && point.In(textSizeRect(i)) {
			g.textSizeFocus, apply = i, true
		}
	}
	if apply {
		g.settings.UITextSize = controls.UITextSizes[g.textSizeFocus]
		g.save()
		g.lastRenderedAt = time.Time{}
		g.trace(map[string]any{"kind": "ui_text_size", "size": g.settings.UITextSize})
	}
	return nil
}

func (g *game) drawTextSizePanel() {
	g.button(image.Rect(40, 80, 920, 630), "", false)
	g.text(image.Rect(60, 90, 900, 134), g.label("ui_text_title"))
	g.text(image.Rect(60, 145, 900, 217), g.label("ui_text_help"))
	for i, key := range []string{"ui_text_standard", "ui_text_large", "ui_text_largest"} {
		label := fmt.Sprintf("%s　%d", g.label(key), controls.UITextSizes[i])
		if controls.UITextSizes[i] == g.systemTextSize() {
			label += "　" + g.label("ui_text_current")
		}
		g.button(textSizeRect(i), label, i == g.textSizeFocus)
	}
	g.text(image.Rect(60, 480, 900, 552), g.label("ui_text_scope"))
	g.text(image.Rect(60, 560, 720, 600), g.notice)
	g.button(image.Rect(740, 560, 900, 600), g.label("close"), false)
}

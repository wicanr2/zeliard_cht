package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var toolbarCheats = image.Rect(536, 8, 600, 40)

func cheatRect(i int) image.Rectangle { return image.Rect(100, 180+i*90, 860, 244+i*90) }

func (g *game) openCheatPanel() {
	g.panel, g.cheatPanel, g.keyboardPanel, g.speedPanel = true, true, false, false
	g.listening, g.notice = "", ""
	g.releaseGate = true
	g.edges.Sync(g, nil)
}

func (g *game) updateCheatPanel(point image.Point, click bool, just func(string) bool) error {
	g.edges.Sync(g, nil)
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || click && point.In(image.Rect(740, 560, 900, 600)) {
		g.panel, g.cheatPanel, g.releaseGate = false, false, true
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || just("down") {
		g.cheatFocus = (g.cheatFocus + 1) % 3
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || just("jump") {
		g.cheatFocus = (g.cheatFocus + 2) % 3
	}
	apply := inpututil.IsKeyJustPressed(ebiten.KeyEnter) || just("attack")
	for i := 0; i < 3; i++ {
		if click && point.In(cheatRect(i)) {
			g.cheatFocus, apply = i, true
		}
	}
	if !apply || g.cheats == nil {
		return nil
	}
	switch g.cheatFocus {
	case 0:
		g.cheats.LockHP = !g.cheats.LockHP
	case 1:
		g.cheats.LockMP = !g.cheats.LockMP
	case 2:
		g.cheats.DisableRegeneration = !g.cheats.DisableRegeneration
		g.settings.DisableStandingRegen = g.cheats.DisableRegeneration
		g.save()
	}
	g.cheats.Apply()
	g.trace(map[string]any{"kind": "cheat_settings", "cheats": g.cheats.State(), "source_kind": "explicit user cheat; not original oracle evidence"})
	return nil
}

func (g *game) drawCheatPanel() {
	g.button(image.Rect(40, 80, 920, 630), "", false)
	g.text(image.Rect(60, 90, 900, 125), g.label("cheat_title"))
	g.text(image.Rect(60, 130, 900, 162), g.label("cheat_scope"))
	values := [3]bool{}
	if g.cheats != nil {
		values = [3]bool{g.cheats.LockHP, g.cheats.LockMP, !g.cheats.DisableRegeneration}
	}
	keys := [3]string{"cheat_hp", "cheat_mp", "cheat_regen"}
	for i, k := range keys {
		state := "cheat_off"
		if values[i] {
			state = "cheat_on"
		}
		g.button(cheatRect(i), g.label(k)+"："+g.label(state), g.cheatFocus == i)
	}
	g.text(image.Rect(60, 465, 900, 502), g.label("cheat_mp_help"))
	g.text(image.Rect(60, 510, 900, 545), g.label("cheat_help"))
	g.button(image.Rect(740, 560, 900, 600), g.label("close"), false)
}

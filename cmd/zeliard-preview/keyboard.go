package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/zeliard_cht/internal/prototype/controls"
)

func (g *game) beginKeyboardBind(action string) {
	g.listening = action
	g.notice = g.label("keyboard_listening")
}

func (g *game) bindKeyboard(key string) {
	swapped, err := g.settings.BindKeyboard(controls.Actions[g.focus], key)
	if err != nil {
		g.notice = g.label("keyboard_unsupported")
		return
	}
	g.listening = ""
	g.releaseGate = true
	g.save()
	if swapped {
		g.notice = g.label("keyboard_conflict")
	}
}

func (g *game) captureKeyboard() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.listening, g.notice = "", ""
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		g.bindKeyboard("")
		return
	}
	for _, key := range inpututil.AppendJustPressedKeys(nil) {
		if _, ok := keyboard[key]; ok {
			g.bindKeyboard(key.String())
			return
		}
		g.notice = g.label("keyboard_unsupported")
	}
}

func (g *game) keyboardToolbarPressed() bool {
	for key := range keyboard {
		if key.String() == g.settings.Keyboard["toolbar"] && inpututil.IsKeyJustPressed(key) {
			return true
		}
	}
	return false
}

func (g *game) keyboardName(key string) string {
	if key == "" {
		return g.label("unbound")
	}
	if translated := g.labels["key_"+key]; translated != "" {
		return translated
	}
	return key
}

func (g *game) drawKeyboardPanel() {
	g.text(image.Rect(60, 90, 700, 120), g.label("keyboard_title"))
	g.button(image.Rect(720, 90, 900, 122), g.label("controller"), false)
	g.text(image.Rect(60, 135, 900, 175), g.label("keyboard_bind"))
	g.text(image.Rect(60, 180, 900, 205), g.label("keyboard_help"))
	g.text(image.Rect(60, 210, 900, 235), g.label("keyboard_scope"))
	g.text(image.Rect(60, 238, 900, 263), g.notice)
	for i, action := range controls.Actions {
		value := g.keyboardName(g.settings.Keyboard[action])
		if g.listening == action {
			value = g.label("keyboard_waiting")
		}
		g.button(actionRect(i), g.label(action)+"："+value, g.focus == i)
	}
	g.button(image.Rect(60, 560, 360, 600), g.label("keyboard_reset"), false)
	g.button(image.Rect(400, 560, 690, 600), g.label("keyboard_unbind"), false)
	g.button(image.Rect(740, 560, 900, 600), g.label("close"), false)
}

// Package controls 處理前端雛形的手把設定與掃描碼邊緣，不重作遊戲規則。
package controls

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
)

// 掃描碼依手冊第 6 頁及 PC 鍵盤對應；遊戲內動作仍待正常路徑驗收。
var Actions = []string{"left", "right", "jump", "down", "jump_left", "jump_right", "attack", "magic", "inventory", "game_pause", "toolbar"}
var Scans = map[string]uint8{"left": 0x4b, "right": 0x4d, "jump": 0x48, "down": 0x50, "jump_left": 0x47, "jump_right": 0x49, "attack": 0x39, "magic": 0x38, "inventory": 0x1c, "game_pause": 0x01}

type Binding struct {
	Kind    string  `json:"kind"`
	Index   int     `json:"index"`
	Sign    int     `json:"sign,omitempty"`
	Neutral float64 `json:"neutral,omitempty"`
}

func (b Binding) Active(buttons []bool, axes []float64, deadzone float64) bool {
	if b.Index < 0 {
		return false
	}
	switch b.Kind {
	case "button":
		return b.Index < len(buttons) && buttons[b.Index]
	case "axis":
		return b.Index < len(axes) && (axes[b.Index]-b.Neutral)*float64(b.Sign) > deadzone
	}
	return false
}

type Settings struct {
	UITextSize           int                           `json:"ui_text_size,omitempty"`
	DisableStandingRegen bool                          `json:"disable_standing_regen,omitempty"`
	CPUSpeed             int                           `json:"cpu_speed,omitempty"`
	Version              int                           `json:"version"`
	Selected             string                        `json:"selected"`
	Deadzone             float64                       `json:"deadzone"`
	Profiles             map[string]map[string]Binding `json:"profiles"`
	Keyboard             map[string]string             `json:"keyboard,omitempty"`
}

func Default() Settings {
	return Settings{Version: 1, UITextSize: 24, CPUSpeed: 1, Deadzone: .25, Profiles: map[string]map[string]Binding{}, Keyboard: DefaultKeyboard()}
}

func Load(path string) (Settings, error) {
	s := Default()
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if err = json.Unmarshal(b, &s); err != nil {
		return s, err
	}
	if !ValidCPUSpeed(s.CPUSpeed) {
		s.CPUSpeed = 1
	}
	if !ValidUITextSize(s.UITextSize) {
		s.UITextSize = 24
	}
	if s.Version != 1 || math.IsNaN(s.Deadzone) || s.Deadzone < .05 || s.Deadzone > .9 || s.Profiles == nil {
		return s, fmt.Errorf("手把設定格式或死區無效")
	}
	if s.Keyboard == nil {
		s.Keyboard = DefaultKeyboard()
	}
	if err := s.ValidateKeyboard(); err != nil {
		return s, err
	}
	for _, profile := range s.Profiles {
		for action, binding := range profile {
			known := action == "toolbar" || Scans[action] != 0
			if !known || binding.Index < 0 || binding.Index > 255 || (binding.Kind != "button" && binding.Kind != "axis") || (binding.Kind == "axis" && (binding.Sign != 1 && binding.Sign != -1 || math.IsNaN(binding.Neutral) || math.Abs(binding.Neutral) > 1)) {
				return s, fmt.Errorf("手把綁定無效：%s", action)
			}
		}
	}
	return s, nil
}

func (s Settings) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".controller-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(append(b, '\n')); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

type Sink interface {
	KeyDown(uint8)
	KeyUp(uint8)
}
type Edges struct{ held map[uint8]bool }

// Sync 接收已合併的鍵盤與手把集合，避免一個來源放開時中斷另一個來源。
func (e *Edges) Sync(s Sink, desired map[uint8]bool) {
	all := map[uint8]bool{}
	for k := range e.held {
		all[k] = true
	}
	for k, v := range desired {
		if v {
			all[k] = true
		}
	}
	keys := make([]int, 0, len(all))
	for k := range all {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, k := range keys {
		scan := uint8(k)
		if e.held[scan] && !desired[scan] {
			s.KeyUp(scan)
		}
		if !e.held[scan] && desired[scan] {
			s.KeyDown(scan)
		}
	}
	e.held = map[uint8]bool{}
	for k, v := range desired {
		if v {
			e.held[k] = true
		}
	}
}

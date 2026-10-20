package controls

import "fmt"

// 主機按鍵名稱是穩定識別值。原版掃描碼及存檔不隨譯文改變。
func DefaultKeyboard() map[string]string {
	return map[string]string{
		"left": "ArrowLeft", "right": "ArrowRight", "jump": "ArrowUp", "down": "ArrowDown",
		"jump_left": "Numpad7", "jump_right": "Numpad9", "attack": "Space", "magic": "AltLeft",
		"inventory": "Enter", "game_pause": "Escape", "toolbar": "F12",
	}
}

func KeyboardKeySupported(key string) bool {
	if len(key) == 1 && key[0] >= 'A' && key[0] <= 'Z' {
		return true
	}
	for _, prefix := range []string{"Digit", "Numpad"} {
		for i := 0; i <= 9; i++ {
			if key == fmt.Sprintf("%s%d", prefix, i) {
				return true
			}
		}
	}
	for i := 1; i <= 12; i++ {
		if key == fmt.Sprintf("F%d", i) {
			return true
		}
	}
	switch key {
	case "ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown", "Space", "AltLeft", "Enter", "Escape", "ControlLeft", "ShiftLeft", "ShiftRight", "Tab", "Backspace":
		return true
	}
	return false
}

func (s Settings) ValidateKeyboard() error {
	defaults, used := DefaultKeyboard(), map[string]bool{}
	for action, key := range s.Keyboard {
		if _, ok := defaults[action]; !ok || key != "" && (!KeyboardKeySupported(key) || used[key]) {
			return fmt.Errorf("鍵盤綁定無效：%s", action)
		}
		if key != "" {
			used[key] = true
		}
	}
	return nil
}

// 重複鍵位與目前動作交換，避免一次按鍵送出兩個遊戲動作。
func (s *Settings) BindKeyboard(action, key string) (bool, error) {
	if _, ok := DefaultKeyboard()[action]; !ok || key != "" && !KeyboardKeySupported(key) {
		return false, fmt.Errorf("鍵盤綁定無效")
	}
	if s.Keyboard == nil {
		s.Keyboard = DefaultKeyboard()
	}
	swapped := false
	if key != "" {
		for other, previous := range s.Keyboard {
			if other != action && previous == key {
				s.Keyboard[other] = s.Keyboard[action]
				swapped = true
			}
		}
	}
	s.Keyboard[action] = key
	return swapped, nil
}

// 未覆寫的DOS鍵盤仍可用；已重綁的來源鍵由新動作接管。
func (s Settings) KeyboardMap(legacy map[string]uint8) map[string]uint8 {
	result := map[string]uint8{}
	for key, scan := range legacy {
		result[key] = scan
	}
	for action, key := range DefaultKeyboard() {
		if selected, ok := s.Keyboard[action]; ok && selected != key {
			delete(result, key)
		}
	}
	for action, key := range s.Keyboard {
		if key == "" {
			continue
		}
		delete(result, key)
		if scan := Scans[action]; scan != 0 {
			result[key] = scan
		}
	}
	return result
}

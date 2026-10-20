package controls

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKeyboardRemapSwapAndRestore(t *testing.T) {
	s := Default()
	if swapped, err := s.BindKeyboard("left", "A"); err != nil || swapped {
		t.Fatal(swapped, err)
	}
	if swapped, err := s.BindKeyboard("right", "A"); err != nil || !swapped || s.Keyboard["left"] != "ArrowRight" {
		t.Fatal(s.Keyboard, swapped, err)
	}
	m := s.KeyboardMap(map[string]uint8{"A": 0x1e, "Q": 0x10, "ArrowLeft": 0x4b, "ArrowRight": 0x4d, "Numpad4": 0x4b, "F12": 0x58})
	if m["A"] != 0x4d || m["ArrowRight"] != 0x4b || m["ArrowLeft"] != 0 || m["Q"] != 0x10 || m["Numpad4"] != 0x4b || m["F12"] != 0 {
		t.Fatal(m)
	}
	if _, err := s.BindKeyboard("right", ""); err != nil || s.ValidateKeyboard() != nil {
		t.Fatal(err)
	}
	s.Keyboard = DefaultKeyboard()
	if s.Keyboard["left"] != "ArrowLeft" || s.ValidateKeyboard() != nil {
		t.Fatal(s.Keyboard)
	}
}

func TestKeyboardLegacyAndPersistence(t *testing.T) {
	p := filepath.Join(t.TempDir(), "controller.json")
	if err := os.WriteFile(p, []byte(`{"version":1,"selected":"usb/pad","deadzone":0.25,"profiles":{"usb/pad":{"attack":{"kind":"button","index":2}}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := Load(p)
	if err != nil || s.Keyboard["attack"] != "Space" || s.Profiles["usb/pad"]["attack"].Index != 2 {
		t.Fatal(s, err)
	}
	s.BindKeyboard("attack", "Z")
	if err = s.Save(p); err != nil {
		t.Fatal(err)
	}
	s, err = Load(p)
	if err != nil || s.Keyboard["attack"] != "Z" || s.Selected != "usb/pad" {
		t.Fatal(s, err)
	}
	s.Keyboard["jump"] = "Z"
	if s.ValidateKeyboard() == nil {
		t.Fatal("重複鍵位未拒絕")
	}
}

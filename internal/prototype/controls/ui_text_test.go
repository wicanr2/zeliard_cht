package controls

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUITextSettingsMigrationAndColdLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "controller.json")
	s := Default()
	s.Keyboard["attack"] = "A"
	s.CPUSpeed, s.DisableStandingRegen = 8, true
	s.Profiles["usb/device"] = map[string]Binding{"attack": {Kind: "button", Index: 3}}
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"", `,"ui_text_size":-999`, `,"ui_text_size":16`, `,"ui_text_size":24`, `,"ui_text_size":32`} {
		// 模擬舊設定沒有字級欄位，以及合法／損壞欄位。
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var legacy map[string]any
		if err := json.Unmarshal(raw, &legacy); err != nil {
			t.Fatal(err)
		}
		delete(legacy, "ui_text_size")
		raw, _ = json.Marshal(legacy)
		raw = append(append(raw[:len(raw)-1], []byte(value)...), '}')
		if err := os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
		got, err := Load(path)
		if err != nil {
			t.Fatal(err)
		}
		want := 24
		if value == `,"ui_text_size":16` {
			want = 16
		} else if value == `,"ui_text_size":32` {
			want = 32
		}
		if got.UITextSize != want || got.Keyboard["attack"] != "A" || got.CPUSpeed != 8 || !got.DisableStandingRegen || got.Profiles["usb/device"]["attack"].Index != 3 {
			t.Fatalf("migration %q: %+v", value, got)
		}
		if err := got.Save(path); err != nil {
			t.Fatal(err)
		}
		cold, err := Load(path)
		if err != nil || cold.UITextSize != want {
			t.Fatalf("cold load: %+v %v", cold, err)
		}
	}
}

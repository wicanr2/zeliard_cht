package controls

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type sink struct{ events []int }

func (s *sink) KeyDown(k uint8) { s.events = append(s.events, int(k)) }
func (s *sink) KeyUp(k uint8)   { s.events = append(s.events, -int(k)) }

func TestMergedInputAndDisconnect(t *testing.T) {
	s := &sink{}
	var e Edges
	e.Sync(s, map[uint8]bool{0x39: true})
	e.Sync(s, map[uint8]bool{0x39: true}) // 一個來源放開，另一個仍按著。
	e.Sync(s, nil)                        // 失焦或斷線。
	e.Sync(s, nil)
	if !reflect.DeepEqual(s.events, []int{0x39, -0x39}) {
		t.Fatal(s.events)
	}
}

func TestAxisNeutralAndBounds(t *testing.T) {
	b := Binding{Kind: "axis", Index: 0, Sign: 1, Neutral: -1}
	if b.Active(nil, []float64{-1}, .25) || b.Active(nil, nil, .25) {
		t.Fatal("中立或消失的軸不可觸發")
	}
	if !b.Active(nil, []float64{0}, .25) {
		t.Fatal("扳機離開中立位置應觸發")
	}
	if (Binding{Kind: "button", Index: 9}).Active([]bool{true}, nil, .25) {
		t.Fatal("不可讀取不存在的按鈕")
	}
}

func TestSettingsRoundTripAndCorruption(t *testing.T) {
	path := filepath.Join(t.TempDir(), "controller.json")
	s := Default()
	s.Selected = "guid/name"
	s.Profiles[s.Selected] = map[string]Binding{"attack": {Kind: "button", Index: 3}}
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil || !reflect.DeepEqual(got, s) {
		t.Fatal(got, err)
	}
	if err := os.WriteFile(path, []byte(`{"version":99,"profiles":{}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("未知版本不可靜默恢復預設")
	}
}

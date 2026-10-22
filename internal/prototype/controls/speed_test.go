package controls

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCPUSpeedPersistenceAndOlderSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "controller.json")
	for _, multiplier := range CPUSpeeds {
		s := Default()
		s.CPUSpeed = multiplier
		s.Keyboard["attack"] = "Z"
		if err := s.Save(path); err != nil {
			t.Fatal(err)
		}
		got, err := Load(path)
		if err != nil || got.CPUSpeed != multiplier || got.Keyboard["attack"] != "Z" {
			t.Fatalf("%d: %+v %v", multiplier, got, err)
		}
	}
	for _, value := range []string{"", `,"cpu_speed":0`, `,"cpu_speed":3`, `,"cpu_speed":64`, `,"cpu_speed":-2`} {
		if err := os.WriteFile(path, []byte(`{"version":1,"deadzone":0.25,"profiles":{}`+value+`}`), 0600); err != nil {
			t.Fatal(err)
		}
		got, err := Load(path)
		if err != nil || got.CPUSpeed != 1 {
			t.Fatalf("older/invalid %s: %+v %v", value, got, err)
		}
	}
}

func TestCPURateCeilingAndHeavyLoad(t *testing.T) {
	if got := SustainableCPURate(96000, 3000, 10000, time.Millisecond, 12*time.Millisecond); got != 3000 {
		t.Fatal("lower ceiling", got)
	}
	if got := SustainableCPURate(3000, 96000, 1000, 20*time.Millisecond, 12*time.Millisecond); got >= 3000 || got < 60 {
		t.Fatal("heavy load must lower rate", got)
	}
	rate := uint64(3000)
	for i := 0; i < 100; i++ {
		next := SustainableCPURate(rate, 96000, 1<<24, time.Millisecond, 12*time.Millisecond)
		if next > 96000 || next < 60 || next > rate+rate/8+1 {
			t.Fatal("unbounded increase", rate, next)
		}
		rate = next
	}
	if rate != 96000 {
		t.Fatal("idle capacity must approach ceiling", rate)
	}
}

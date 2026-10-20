package sound

import (
	"testing"
	"time"
)

func TestClockUsesActualCyclesAndBoundedCatchup(t *testing.T) {
	var c Clock
	now := time.Unix(100, 0)
	if n := c.Budget(now, 900000, 3000); n < 49999 || n > 50000 {
		t.Fatal(n)
	}
	if n := c.Budget(now.Add(time.Millisecond), 960000, 3000); n != 0 {
		t.Fatal("原版超前時不應再跑", n)
	}
	if n := c.Budget(now.Add(time.Second), 960000, 3000); n != 300000 || c.Dropped < 800*time.Millisecond {
		t.Fatal(n, c.Dropped)
	}
	if n := c.Budget(now.Add(time.Second), 1260000, 3000); n != 0 {
		t.Fatal("重複計算追趕", n)
	}
	dropped := c.Dropped
	c.Reset()
	if n := c.Budget(now.Add(time.Hour), 1260000, 3000); n < 49999 || n > 50000 || c.Dropped != dropped {
		t.Fatal("暫停期間不追趕", n, c.Dropped)
	}
	if n := c.Budget(now.Add(2*time.Hour), 0, 3000); n < 49999 || n > 50000 {
		t.Fatal("時鐘倒退應重新定錨", n)
	}
}

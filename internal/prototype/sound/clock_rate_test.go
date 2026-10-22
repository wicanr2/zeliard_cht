package sound

import (
	"testing"
	"time"
)

func TestClockRateChangePreservesElapsedTime(t *testing.T) {
	var c Clock
	now := time.Unix(100, 0)
	ran := c.Budget(now, 0, 3000)
	oldBudget := c.Budget(now.Add(10*time.Millisecond), ran, 3000)
	c.RateChanged(ran, 3000)
	newBudget := c.Budget(now.Add(10*time.Millisecond), ran, 6000)
	if newBudget+1 < oldBudget*2 || newBudget > oldBudget*2+1 {
		t.Fatalf("rate transition lost time: %d -> %d", oldBudget, newBudget)
	}
	c.Reset()
	if next := c.Budget(now.Add(time.Hour), ran, 6000); next < 99999 || next > 100000 {
		t.Fatal("pause must discard old rate segments", next)
	}
}

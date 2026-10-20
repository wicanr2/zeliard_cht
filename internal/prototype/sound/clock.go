package sound

import "time"

// Clock 只決定本次允許跑多少原版 cycles，不寫入原版狀態。
// 使用者已選定最多追趕100毫秒，暫停／失焦不累積落後。
type Clock struct {
	anchor  time.Time
	base    uint64
	skipped time.Duration
	Dropped time.Duration
}

func (c *Clock) Reset() { c.anchor = time.Time{}; c.skipped = 0 }

func (c *Clock) Budget(now time.Time, cycles, perMS uint64) uint64 {
	if perMS == 0 {
		return 0
	}
	if c.anchor.IsZero() || cycles < c.base {
		c.anchor = now.Add(-time.Second / 60)
		c.base = cycles
		c.skipped = 0
	}
	delta := cycles - c.base
	machine := time.Duration(delta/perMS)*time.Millisecond + time.Duration(delta%perMS)*time.Millisecond/time.Duration(perMS)
	lag := now.Sub(c.anchor) - c.skipped - machine
	if lag <= 0 {
		return 0
	}
	if lag > 100*time.Millisecond {
		discard := lag - 100*time.Millisecond
		c.skipped += discard
		c.Dropped += discard
		lag = 100 * time.Millisecond
	}
	return uint64(lag.Nanoseconds()) * perMS / 1000000
}

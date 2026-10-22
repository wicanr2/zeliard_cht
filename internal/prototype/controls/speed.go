package controls

import "time"

var CPUSpeeds = [...]int{1, 2, 4, 8, 16, 32}

func ValidCPUSpeed(value int) bool {
	for _, speed := range CPUSpeeds {
		if value == speed {
			return true
		}
	}
	return false
}

// SustainableCPURate treats the chosen speed as a ceiling and reserves a 30fps
// frame period for both CPU execution and presentation. It never exceeds the ceiling.
func SustainableCPURate(current, ceiling, ran uint64, elapsed, allowance time.Duration) uint64 {
	if current > ceiling {
		return ceiling
	}
	if ran < 256 || elapsed < 100*time.Microsecond || allowance <= 0 {
		return current
	}
	estimate := float64(ran) * float64(allowance) * 1e6 / float64(elapsed) / float64(time.Second/30)
	if estimate < 60 {
		estimate = 60
	}
	if estimate < float64(current)*.9 {
		next := uint64(estimate * .9)
		if next < 60 {
			next = 60
		}
		if next > ceiling {
			next = ceiling
		}
		return next
	}
	if estimate > float64(current)*1.1 {
		next := current + current/8 + 1
		if next > ceiling {
			next = ceiling
		}
		return next
	}
	return current
}

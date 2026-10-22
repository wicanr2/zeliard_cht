package controls

import (
	"testing"
	"time"
)

func TestCPUFrameAllowanceDoesNotStarveInput(t *testing.T) {
	for _, tc := range []struct{ draw, want time.Duration }{
		{0, 24 * time.Millisecond},
		{10 * time.Millisecond, time.Second/30 - 13*time.Millisecond},
		{40 * time.Millisecond, 4 * time.Millisecond},
		{5 * time.Second, 4 * time.Millisecond},
	} {
		if got := CPUFrameAllowance(tc.draw); got != tc.want {
			t.Fatalf("draw %v: got %v, want %v", tc.draw, got, tc.want)
		}
	}
}

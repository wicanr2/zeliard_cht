// Package sound實作規格029的單向播放緩衝，不持有原版執行器。
package sound

import (
	"encoding/binary"
	"io"
	"math"
	"sync"
)

const Rate = 44100
const Capacity = Rate / 10
const Prebuffer = Rate / 20

// Stats的單位為單聲道音框；Bytes是輸出浮點雙聲道位元組數。
type Stats struct {
	Capacity, Queued                                    int
	Pushed, Pulled, Underrun, Dropped, Discarded, Bytes uint64
	ReadCalls, Waits                                    uint64
	MaxReadBytes                                        int
	Closed                                              bool
}

// Stream由更新執行緒Push，裝置執行緒Read。來源是有號16位元單聲道，
// 輸出為32位元小端浮點雙聲道；任意長度Read均保留完整音框順序。
type Stream struct {
	mu           sync.Mutex
	ready        *sync.Cond
	buf          []int16
	start, count int
	pending      [8]byte
	offset       int
	stats        Stats
}

func NewStream(capacity int) *Stream {
	if capacity < 1 {
		panic("音訊容量必須大於零")
	}
	q := &Stream{buf: make([]int16, capacity), offset: 8, stats: Stats{Capacity: capacity}}
	q.ready = sync.NewCond(&q.mu)
	return q
}

func (q *Stream) Push(values []int16) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.stats.Closed {
		return
	}
	defer q.ready.Signal()
	q.stats.Pushed += uint64(len(values))
	if len(values) >= len(q.buf) {
		q.stats.Dropped += uint64(q.count + len(values) - len(q.buf))
		copy(q.buf, values[len(values)-len(q.buf):])
		q.start, q.count = 0, len(q.buf)
		return
	}
	if excess := q.count + len(values) - len(q.buf); excess > 0 {
		q.start = (q.start + excess) % len(q.buf)
		q.count -= excess
		q.stats.Dropped += uint64(excess)
	}
	tail := (q.start + q.count) % len(q.buf)
	n := copy(q.buf[tail:], values)
	copy(q.buf, values[n:])
	q.count += len(values)
}

func (q *Stream) Read(dst []byte) (int, error) {
	if len(dst) == 0 {
		return 0, nil
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.stats.Closed {
		return 0, io.EOF
	}
	q.stats.ReadCalls++
	if len(dst) > q.stats.MaxReadBytes {
		q.stats.MaxReadBytes = len(dst)
	}
	n := 0
	for n < len(dst) {
		if q.offset == 8 {
			if q.count == 0 && n > 0 {
				break
			}
			for q.count == 0 && !q.stats.Closed {
				q.stats.Waits++
				q.ready.Wait()
			}
			if q.stats.Closed {
				return 0, io.EOF
			}
			v := q.buf[q.start]
			q.start = (q.start + 1) % len(q.buf)
			q.count--
			q.stats.Pulled++
			bits := math.Float32bits(float32(v) / 32768)
			binary.LittleEndian.PutUint32(q.pending[:4], bits)
			binary.LittleEndian.PutUint32(q.pending[4:], bits)
			q.offset = 0
		}
		k := copy(dst[n:], q.pending[q.offset:])
		n += k
		q.offset += k
	}
	q.stats.Bytes += uint64(n)
	return n, nil
}

func (q *Stream) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.stats.Closed {
		q.stats.Discarded += uint64(q.count)
		q.start, q.count = 0, 0
		q.stats.Closed = true
		q.ready.Broadcast()
	}
	return nil
}

func (q *Stream) Snapshot() Stats {
	q.mu.Lock()
	defer q.mu.Unlock()
	s := q.stats
	s.Queued = q.count
	return s
}

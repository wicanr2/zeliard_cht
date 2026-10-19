package sound

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

func stereo(values []int16) []byte {
	b := make([]byte, 0, len(values)*8)
	for _, v := range values {
		bits := math.Float32bits(float32(v) / 32768)
		b = binary.LittleEndian.AppendUint32(b, bits)
		b = binary.LittleEndian.AppendUint32(b, bits)
	}
	return b
}

func TestStreamFragmentedFloatStereo(t *testing.T) {
	values := []int16{-32768, -8192, 0, 8192, 32767}
	q := NewStream(len(values))
	q.Push(values)
	values[0] = 123 // Push必須保有自己的資料。
	var got []byte
	for _, n := range []int{1, 7, 3, 5, 13, 11} {
		p := make([]byte, n)
		k, err := q.Read(p)
		if err != nil || k != n {
			t.Fatal(k, err)
		}
		got = append(got, p...)
	}
	if !bytes.Equal(got, stereo([]int16{-32768, -8192, 0, 8192, 32767})) {
		t.Fatal("部分讀取破壞聲道或音框")
	}
	s := q.Snapshot()
	if s.Pulled != 5 || s.Underrun != 0 || s.Dropped != 0 || s.Queued != 0 {
		t.Fatal(s)
	}
}

func awaitSourceWait(t *testing.T, q *Stream) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for q.Snapshot().Waits == 0 {
		if time.Now().After(deadline) {
			t.Fatal("來源未進入有界等待")
		}
		runtime.Gosched()
	}
}

func TestStreamShortReadThenWaitForRealFrames(t *testing.T) {
	q := NewStream(2)
	q.Push([]int16{8192})
	p := bytes.Repeat([]byte{255}, 24)
	n, err := q.Read(p)
	if err != nil || n != 8 || !bytes.Equal(p[:n], stereo([]int16{8192})) {
		t.Fatal("應先短讀已有的原版音框", n, err)
	}
	done := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 24)
		n, _ := q.Read(buf)
		done <- buf[:n]
	}()
	awaitSourceWait(t, q)
	select {
	case <-done:
		t.Fatal("空來源提前回傳")
	default:
	}
	q.Push([]int16{-8192, 32767})
	select {
	case got := <-done:
		if !bytes.Equal(got, stereo([]int16{-8192, 32767})) {
			t.Fatal("新資料喚醒後插入靜音或破壞順序")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Push未喚醒來源")
	}
	if s := q.Snapshot(); s.Pulled != 3 || s.Underrun != 0 || s.ReadCalls != 2 || s.Waits != 1 || s.MaxReadBytes != 24 {
		t.Fatal(s)
	}
}

func TestStreamCloseWakesEmptyReader(t *testing.T) {
	q := NewStream(2)
	done := make(chan error, 1)
	go func() {
		n, err := q.Read(make([]byte, 17640))
		if n != 0 {
			done <- fmt.Errorf("關閉後仍輸出%d", n)
			return
		}
		done <- err
	}()
	awaitSourceWait(t, q)
	_ = q.Close()
	select {
	case err := <-done:
		if err != io.EOF {
			t.Fatal("關閉未喚醒並結束來源", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("關閉仍卡住讀取")
	}
	if s := q.Snapshot(); s.Underrun != 0 || !s.Closed || s.Waits != 1 {
		t.Fatal(s)
	}
}

func TestStreamOverflowPreservesPartialFrame(t *testing.T) {
	q := NewStream(2)
	q.Push([]int16{1, 2})
	first := make([]byte, 1)
	_, _ = q.Read(first)
	q.Push([]int16{3, 4})
	rest := make([]byte, 23)
	_, _ = io.ReadFull(q, rest)
	if !bytes.Equal(append(first, rest...), stereo([]int16{1, 3, 4})) {
		t.Fatal("溢位改寫正在輸出的部分音框")
	}
	if s := q.Snapshot(); s.Dropped != 1 || s.Pulled != 3 {
		t.Fatal(s)
	}
}

func TestStreamWrapAndLargeOverflow(t *testing.T) {
	q := NewStream(3)
	q.Push([]int16{1, 2})
	p := make([]byte, 8)
	_, _ = io.ReadFull(q, p)
	q.Push([]int16{3, 4})
	p = make([]byte, 24)
	_, _ = io.ReadFull(q, p)
	if !bytes.Equal(p, stereo([]int16{2, 3, 4})) {
		t.Fatal("環形順序不符")
	}
	q.Push([]int16{5, 6, 7, 8, 9})
	p = make([]byte, 24)
	_, _ = io.ReadFull(q, p)
	if !bytes.Equal(p, stereo([]int16{7, 8, 9})) || q.Snapshot().Dropped != 2 {
		t.Fatal("大區塊未保留最新取樣")
	}
}

func TestStreamCloseDiscardsOldQueue(t *testing.T) {
	q := NewStream(3)
	q.Push([]int16{1, 2, 3})
	p := make([]byte, 1)
	_, _ = q.Read(p)
	if err := q.Close(); err != nil {
		t.Fatal(err)
	}
	q.Push([]int16{9})
	n, err := q.Read(p)
	if n != 0 || err != io.EOF {
		t.Fatal("舊來源未結束", n, err)
	}
	if s := q.Snapshot(); !s.Closed || s.Queued != 0 || s.Discarded != 2 || s.Pushed != 3 {
		t.Fatal(s)
	}
	fresh := NewStream(3)
	fresh.Push([]int16{7})
	p = make([]byte, 8)
	_, _ = io.ReadFull(fresh, p)
	if !bytes.Equal(p, stereo([]int16{7})) {
		t.Fatal("新來源帶入舊音框")
	}
}

func TestStreamConcurrentReadPushClose(t *testing.T) {
	q := NewStream(128)
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			q.Push([]int16{int16(i), int16(-i)})
		}
	}()
	go func() {
		defer wg.Done()
		p := make([]byte, 37)
		for i := 0; i < 1000; i++ {
			_, _ = q.Read(p)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			s := q.Snapshot()
			if s.Queued < 0 || s.Queued > s.Capacity {
				t.Error(s)
			}
		}
		_ = q.Close()
	}()
	wg.Wait()
	_ = q.Close()
	if !q.Snapshot().Closed {
		t.Fatal("來源仍開啟")
	}
}

func originalPCM(t *testing.T) ([]byte, map[string]any) {
	t.Helper()
	root := os.Getenv("ZELIARD_AUDIO_SOURCE")
	if root == "" {
		t.Skip("須提供真實原版音訊來源；skip不計驗收")
	}
	b, err := os.ReadFile(filepath.Join(root, "original-mono.pcm"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "receipt.json"))
	if err != nil {
		t.Fatal(err)
	}
	var r map[string]any
	if err = json.Unmarshal(data, &r); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != r["pcm_sha256"] || len(b) != 511929*2 || r["nonzero_samples"] != float64(18610) {
		t.Fatal("原版來源身份不符")
	}
	return b, r
}

func TestStreamOriginalPCMAllFrames(t *testing.T) {
	b, _ := originalPCM(t)
	q := NewStream(4096)
	input := make([]int16, len(b)/2)
	for i := range input {
		input[i] = int16(binary.LittleEndian.Uint16(b[2*i:]))
	}
	var got []byte
	for start := 0; start < len(input); {
		end := start + 1000
		if end > len(input) {
			end = len(input)
		}
		q.Push(input[start:end])
		remaining := (end - start) * 8
		for remaining > 0 {
			n := 37
			if n > remaining {
				n = remaining
			}
			p := make([]byte, n)
			k, err := q.Read(p)
			if err != nil || k != n {
				t.Fatal(k, err)
			}
			got = append(got, p...)
			remaining -= k
		}
		start = end
	}
	if !bytes.Equal(got, stereo(input)) {
		t.Fatal("真實原版取樣的格式轉換不同")
	}
	s := q.Snapshot()
	if s.Pushed != 511929 || s.Pulled != 511929 || s.Dropped != 0 || s.Underrun != 0 {
		t.Fatal(s)
	}
	t.Logf("511929個原版音框完整轉成4095432位元組雙聲道F32，零欠載／溢位")
}

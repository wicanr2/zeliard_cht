package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	ebaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/wicanr2/zeliard_cht/internal/prototype/sound"
)

type audioOutput struct {
	mode          string
	context       *ebaudio.Context
	player        *ebaudio.Player
	stream        *sound.Stream
	previous      sound.Stats
	recorder      *sound.WAV
	sourceSamples uint64
	epochs        uint64
	lastError     string
}

func newAudioOutput(mode, path string) (*audioOutput, error) {
	if mode != "off" && mode != "ebiten" {
		return nil, fmt.Errorf("-audio 須為 off 或 ebiten")
	}
	a := &audioOutput{mode: mode}
	if path != "" {
		w, err := sound.CreateWAV(path)
		if err != nil {
			return nil, err
		}
		a.recorder = w
	}
	if mode == "ebiten" {
		a.context = ebaudio.NewContext(sound.Rate)
	}
	return a, nil
}

func (a *audioOutput) submit(values []int16) error {
	if a == nil {
		return nil
	}
	a.sourceSamples += uint64(len(values))
	if a.recorder != nil {
		if err := a.recorder.Write(values); err != nil {
			return err
		}
	}
	if a.mode == "off" || len(values) == 0 {
		return nil
	}
	if a.stream == nil {
		a.stream = sound.NewStream(sound.Capacity)
		a.epochs++
	}
	a.stream.Push(values)
	// Oto可能同步預填。先備妥完整50毫秒，避免更新執行緒等待新取樣。
	if a.player == nil && a.stream.Snapshot().Queued >= sound.Prebuffer {
		p, err := a.context.NewPlayerF32(a.stream)
		if err != nil {
			return err
		}
		p.SetBufferSize(50 * time.Millisecond)
		a.player = p
		p.Play()
	}
	return nil
}

func addAudioStats(a, b sound.Stats) sound.Stats {
	a.Pushed += b.Pushed
	a.Pulled += b.Pulled
	a.Underrun += b.Underrun
	a.Dropped += b.Dropped
	a.Discarded += b.Discarded
	a.Bytes += b.Bytes
	a.ReadCalls += b.ReadCalls
	a.Waits += b.Waits
	if b.MaxReadBytes > a.MaxReadBytes {
		a.MaxReadBytes = b.MaxReadBytes
	}
	return a
}

func (a *audioOutput) suspend() error {
	if a == nil || a.stream == nil {
		return nil
	}
	_ = a.stream.Close()
	a.previous = addAudioStats(a.previous, a.stream.Snapshot())
	a.stream = nil
	var err error
	if a.player != nil {
		err = a.player.Close()
	}
	a.player = nil
	if err != nil {
		a.lastError = err.Error()
	}
	return err
}

func (a *audioOutput) state() map[string]any {
	if a == nil {
		return nil
	}
	s := a.previous
	if a.stream != nil {
		now := a.stream.Snapshot()
		s = addAudioStats(s, now)
		s.Queued = now.Queued
	}
	return map[string]any{"mode": a.mode, "active": a.player != nil, "prefilling": a.stream != nil && a.player == nil, "epochs": a.epochs, "rate": sound.Rate, "capacity_frames": sound.Capacity, "prebuffer_frames": sound.Prebuffer, "source_frames": a.sourceSamples, "queued_frames": s.Queued, "pushed_frames": s.Pushed, "pulled_frames": s.Pulled, "underrun_frames": s.Underrun, "source_waits": s.Waits, "source_read_calls": s.ReadCalls, "max_source_read_bytes": s.MaxReadBytes, "overflow_frames": s.Dropped, "suspend_discarded_frames": s.Discarded, "device_bytes": s.Bytes, "last_error": a.lastError, "source_recording": a.recorder.State(), "scope": "underrun僅計前端補零；來源等待及裝置缺音分開，不以補零0宣稱穩定播放"}
}

func (a *audioOutput) close() error {
	if a == nil {
		return nil
	}
	return errors.Join(a.suspend(), a.recorder.Close())
}

func (g *game) audioLifecycle() {
	if g.output != nil && (g.captured || g.panel || g.paused || !ebiten.IsFocused()) {
		if err := g.output.suspend(); err != nil {
			g.output.lastError = err.Error()
		}
	}
}

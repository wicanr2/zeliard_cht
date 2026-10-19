package main

import (
	"testing"

	"github.com/wicanr2/zeliard_cht/internal/prototype/sound"
)

func TestAudioSuspendBeforePlayerPrefill(t *testing.T) {
	// 尚未達預填，不需要建立任何裝置。模擬第一批之後立即開設定頁。
	a := &audioOutput{mode: "ebiten"}
	if err := a.submit(make([]int16, sound.Rate/60)); err != nil {
		t.Fatal(err)
	}
	if a.player != nil || a.stream == nil || a.epochs != 1 {
		t.Fatal("第一批便啟動播放器")
	}
	old := a.stream
	if err := a.suspend(); err != nil {
		t.Fatal(err)
	}
	if !old.Snapshot().Closed || old.Snapshot().Discarded != sound.Rate/60 || a.stream != nil {
		t.Fatal("未達預填的暫停未清空")
	}
	if err := a.submit([]int16{8192}); err != nil {
		t.Fatal(err)
	}
	if a.stream == old || a.epochs != 2 || a.stream.Snapshot().Queued != 1 {
		t.Fatal("恢復未建立乾淨來源")
	}
	if err := a.close(); err != nil {
		t.Fatal(err)
	}
	if a.previous.Discarded != sound.Rate/60+1 || a.previous.Underrun != 0 {
		t.Fatal(a.previous)
	}
}

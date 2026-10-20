// Package fm 載入獨立的 Nuked-OPL3 動態函式庫。原曲的暫存器事件由原版提供。
package fm

import (
	"fmt"
	"github.com/ebitengine/purego"
)

type Synth struct {
	handle uintptr
	chip   uintptr
	write  func(uintptr, uint16, uint8)
	sample func(uintptr) float64
	free   func(uintptr)
}

func Open(path string, rate int) (*Synth, error) {
	handle, err := openLibrary(path)
	if err != nil {
		return nil, err
	}
	s := &Synth{handle: handle}
	var create func(uint32) uintptr
	// 與封包內 bridge.c 的固定 C ABI 對應，缺少符號便拒絕使用。
	var failure any
	func() {
		defer func() { failure = recover() }()
		purego.RegisterLibFunc(&create, handle, "zopl_new")
		purego.RegisterLibFunc(&s.write, handle, "zopl_write")
		purego.RegisterLibFunc(&s.sample, handle, "zopl_sample")
		purego.RegisterLibFunc(&s.free, handle, "zopl_free")
	}()
	if failure != nil {
		closeLibrary(handle)
		return nil, fmt.Errorf("OPL 函式庫不相容：%v", failure)
	}
	s.chip = create(uint32(rate))
	if s.chip == 0 {
		closeLibrary(handle)
		return nil, fmt.Errorf("OPL 記憶體配置失敗")
	}
	return s, nil
}

func (s *Synth) Write(reg, value uint8) { s.write(s.chip, uint16(reg), value) }
func (s *Synth) Sample() float64        { return s.sample(s.chip) }
func (s *Synth) Close() {
	if s.chip != 0 {
		s.free(s.chip)
		s.chip = 0
		closeLibrary(s.handle)
	}
}

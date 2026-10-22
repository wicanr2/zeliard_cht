// Package cheats 實作使用者明示的原版規則差異，不能用作原版oracle證據。
package cheats

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/wicanr2/dosgolem/oracle"
)

type Controller struct {
	o                                       *oracle.Oracle
	LockHP, LockMP, DisableRegeneration     bool
	HPWrites, MPWrites, RegenerationBlocked uint64
}

func New(o *oracle.Oracle, executable string) (*Controller, error) {
	b, err := os.ReadFile(executable)
	if err != nil {
		return nil, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != "fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341" {
		return nil, fmt.Errorf("作弊來源不符，未啟用")
	}
	c := &Controller{o: o}
	c.attach()
	return c, nil
}

func (c *Controller) code(off uint16, want []byte) bool {
	return c.o.DSReg() == 0x1b1 && c.validCode(off, want)
}

func (c *Controller) validCode(off uint16, want []byte) bool {
	return c.o.VideoMode() == 0x13 && bytes.Equal(c.o.Bytes(oracle.Far(0x1b1, off), len(want)), want)
}

func (c *Controller) hp() {
	if !c.LockHP {
		return
	}
	max := c.o.Word(oracle.Far(0x1b1, 0xb2))
	if max != 0 && c.o.Word(oracle.Far(0x1b1, 0x90)) != max {
		c.o.SetWord(oracle.Far(0x1b1, 0x90), max)
		c.HPWrites++
	}
}

func (c *Controller) mp() {
	if !c.LockMP {
		return
	}
	for i := uint16(0); i < 7; i++ {
		max := c.o.Byte(oracle.Far(0x1b1, 0xb4+i))
		if max != 0 && c.o.Byte(oracle.Far(0x1b1, 0xab+i)) != max {
			c.o.SetByte(oracle.Far(0x1b1, 0xab+i), max)
			c.MPWrites++
		}
	}
}

// Apply只在已確認的原版顯示常式仍在時補滿，不猜未知程式的資料欄位。
func (c *Controller) Apply() {
	if c == nil || (!c.LockHP && !c.LockMP) {
		return
	}
	if c.validCode(0x2256, []byte{0xbf, 0x14, 0xcc, 0x2e, 0x8b, 0x1e, 0x90, 0x00}) {
		c.hp()
	}
	if c.validCode(0x23cc, []byte{0x1e, 0x33, 0xdb, 0x2e, 0x8a, 0x1e, 0x9d, 0x00}) {
		c.mp()
	}
}

func (c *Controller) attach() {
	// 受傷先完成原版的扣減／歸零，再補滿，保留原版受傷動畫與數字更新。
	c.o.OnCall(oracle.Far(0x1b1, 0x7691), func(*oracle.Oracle) {
		if c.LockHP && c.code(0x7685, []byte{0x29, 0x06, 0x90, 0x00, 0x73, 0x06, 0xc7, 0x06, 0x90, 0x00, 0, 0, 0x56, 0x2e, 0xff, 0x16, 0x08, 0x20, 0x5e, 0xc3}) {
			c.hp()
		}
	})
	for _, off := range []uint16{0x8813, 0x881f} {
		c.o.OnCall(oracle.Far(0x1b1, off), func(*oracle.Oracle) {
			if c.LockMP && c.code(0x8813, []byte{0xf6, 0x87, 0xab, 0, 0xff, 0x75, 1, 0xc3, 0xfe, 0x8f, 0xab, 0, 0x2e, 0xff, 0x16, 0x18, 0x20}) {
				c.mp()
			}
		})
	}
	c.o.OnCall(oracle.Far(0x1b1, 0x71ba), func(*oracle.Oracle) {
		if c.DisableRegeneration && !c.LockHP && c.code(0x71ae, []byte{0xa1, 0x90, 0, 0x3b, 0x06, 0xb2, 0, 0x73, 0x0b, 0x05, 2, 0, 0xa3, 0x90, 0, 0x2e, 0xff, 0x16, 8, 0x20}) {
			c.o.SetRegs(oracle.CallRegs{AX: c.o.Word(oracle.Far(0x1b1, 0x90)), SetAX: true})
			c.RegenerationBlocked++
		}
	})
}

func (c *Controller) State() map[string]any {
	if c == nil {
		return nil
	}
	return map[string]any{"lock_hp": c.LockHP, "lock_mp": c.LockMP, "standing_regeneration": !c.DisableRegeneration, "hp_writes": c.HPWrites, "mp_writes": c.MPWrites, "regeneration_blocked": c.RegenerationBlocked, "behavior_modified": c.LockHP || c.LockMP || c.DisableRegeneration || c.HPWrites != 0 || c.MPWrites != 0 || c.RegenerationBlocked != 0}
}

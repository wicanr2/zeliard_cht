package cheats

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
)

// 受控原版指令片段：只驗掛鉤及關閉分支，不當正常玩家路徑或原版parity。
func fixture(t *testing.T) (*oracle.Oracle, *Controller) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "TEST.EXE")
	fixture := make([]byte, 32+1024)
	copy(fixture, "MZ")
	put := func(off int, v uint16) { binary.LittleEndian.PutUint16(fixture[off:off+2], v) }
	put(2, uint16(len(fixture)%512))
	put(4, uint16((len(fixture)+511)/512))
	put(8, 2)
	put(0x0a, 0x10)
	put(0x0c, 0xffff)
	put(0x10, 0x400)
	put(0x18, 0x1c)
	copy(fixture[32:], []byte{0xeb, 0xfe})
	if err := os.WriteFile(path, fixture, 0600); err != nil {
		t.Fatal(err)
	}
	o, err := oracle.Load(path, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { o.Close() })
	o.SetWord(oracle.Far(0x1b1, 0x2008), 0xf100)
	o.SetWord(oracle.Far(0x1b1, 0x2018), 0xf100)
	o.WriteBytes(oracle.Far(0x1b1, 0xf100), []byte{0xc3}) // 替代顯示端，保持near返回。
	o.WriteBytes(oracle.Far(0x1b1, 0xf000), []byte{0xb8, 0x13, 0, 0xcd, 0x10, 0xc3})
	o.CallNear(oracle.Far(0x1b1, 0xf000), 0xff00, oracle.CallRegs{DS: 0x1b1, SetDS: true})
	if err = o.RunUntil(oracle.At(oracle.Far(0x1b1, 0xff00)), oracle.Budget(100)); err != nil {
		t.Fatal(err)
	}
	o.WriteBytes(oracle.Far(0x1b1, 0x2256), []byte{0xbf, 0x14, 0xcc, 0x2e, 0x8b, 0x1e, 0x90, 0})
	o.WriteBytes(oracle.Far(0x1b1, 0x23cc), []byte{0x1e, 0x33, 0xdb, 0x2e, 0x8a, 0x1e, 0x9d, 0})
	o.WriteBytes(oracle.Far(0x1b1, 0x7685), []byte{0x29, 0x06, 0x90, 0, 0x73, 0x06, 0xc7, 0x06, 0x90, 0, 0, 0, 0x56, 0x2e, 0xff, 0x16, 8, 0x20, 0x5e, 0xc3})
	o.WriteBytes(oracle.Far(0x1b1, 0x8813), []byte{0xf6, 0x87, 0xab, 0, 0xff, 0x75, 1, 0xc3, 0xfe, 0x8f, 0xab, 0, 0x2e, 0xff, 0x16, 0x18, 0x20, 0xc3})
	o.WriteBytes(oracle.Far(0x1b1, 0x71ae), []byte{0xa1, 0x90, 0, 0x3b, 0x06, 0xb2, 0, 0x73, 0x0b, 0x05, 2, 0, 0xa3, 0x90, 0, 0x2e, 0xff, 0x16, 8, 0x20, 0xc3})
	c := &Controller{o: o}
	c.attach()
	return o, c
}

func run(t *testing.T, o *oracle.Oracle, off uint16, r oracle.CallRegs) {
	t.Helper()
	r.DS = 0x1b1
	r.SetDS = true
	o.CallNear(oracle.Far(0x1b1, off), 0xff00, r)
	if err := o.RunUntil(oracle.At(oracle.Far(0x1b1, 0xff00)), oracle.Budget(100)); err != nil {
		t.Fatal(err)
	}
}

func TestDamageAndLethalDamage(t *testing.T) {
	for _, damage := range []uint16{20, 100} {
		for _, lock := range []bool{false, true} {
			o, c := fixture(t)
			c.LockHP = lock
			o.SetWord(oracle.Far(0x1b1, 0x90), 50)
			o.SetWord(oracle.Far(0x1b1, 0xb2), 80)
			run(t, o, 0x7685, oracle.CallRegs{AX: damage, SetAX: true})
			want := uint16(0)
			if damage < 50 {
				want = 50 - damage
			}
			if lock {
				want = 80
			}
			if got := o.Word(oracle.Far(0x1b1, 0x90)); got != want {
				t.Fatalf("damage=%d lock=%v got=%d want=%d", damage, lock, got, want)
			}
		}
	}
}

func TestMagicChargeAndIndependentSwitch(t *testing.T) {
	for _, lock := range []bool{false, true} {
		o, c := fixture(t)
		c.LockMP = lock
		o.SetByte(oracle.Far(0x1b1, 0xab), 1)
		o.SetByte(oracle.Far(0x1b1, 0xb4), 12)
		o.SetWord(oracle.Far(0x1b1, 0x90), 25)
		o.SetWord(oracle.Far(0x1b1, 0xb2), 80)
		run(t, o, 0x8813, oracle.CallRegs{BX: 0, SetBX: true})
		want := byte(0)
		if lock {
			want = 12
		}
		if got := o.Byte(oracle.Far(0x1b1, 0xab)); got != want {
			t.Fatalf("lock=%v got=%d want=%d", lock, got, want)
		}
		if o.Word(oracle.Far(0x1b1, 0x90)) != 25 {
			t.Fatal("鎖MP改了生命")
		}
	}
}

func TestOnlyNaturalRegenerationSuppressed(t *testing.T) {
	for _, disabled := range []bool{false, true} {
		o, c := fixture(t)
		c.DisableRegeneration = disabled
		o.SetWord(oracle.Far(0x1b1, 0x90), 50)
		o.SetWord(oracle.Far(0x1b1, 0xb2), 80)
		run(t, o, 0x71ae, oracle.CallRegs{})
		want := uint16(52)
		if disabled {
			want = 50
		}
		if got := o.Word(oracle.Far(0x1b1, 0x90)); got != want {
			t.Fatalf("disabled=%v got=%d", disabled, got)
		}
		// 藥水／教堂的直接回復不在此掛鉤，不被攔截。
		o.SetWord(oracle.Far(0x1b1, 0x90), 80)
		c.Apply()
		if o.Word(oracle.Far(0x1b1, 0x90)) != 80 {
			t.Fatal("阻擋了其他回復")
		}
	}
}

func TestDisabledAndWrongSourceDoNotWrite(t *testing.T) {
	o, c := fixture(t)
	o.SetWord(oracle.Far(0x1b1, 0x90), 23)
	o.SetWord(oracle.Far(0x1b1, 0xb2), 80)
	c.Apply()
	if o.Word(oracle.Far(0x1b1, 0x90)) != 23 {
		t.Fatal("預設作弊有寫入")
	}
	c.LockHP = true
	o.SetByte(oracle.Far(0x1b1, 0x2256), 0)
	c.Apply()
	if o.Word(oracle.Far(0x1b1, 0x90)) != 23 {
		t.Fatal("錯來源仍寫入")
	}
}

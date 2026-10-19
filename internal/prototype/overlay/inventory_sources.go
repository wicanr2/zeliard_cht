package overlay

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type inventoryTextKind uint8

const (
	inventorySmallWhite inventoryTextKind = iota + 1
	inventoryLargeWhite
	inventoryLargeColor
	inventoryShieldSecondLine
)

// inventorySourceKinds 只分類固定 SAR2 的顯示來源。指標表、讀取端及
// 字型證據見規格016；不解釋物品效果，也不改動原版指標或資料。
func inventorySourceKinds(data []byte) (map[int]inventoryTextKind, error) {
	const module, size = 0x401e, 3613
	if len(data) < module+size || int(binary.LittleEndian.Uint32(data[module-4:module])) != size {
		return nil, fmt.Errorf("裝備來源模組邊界不符")
	}
	kinds := map[int]inventoryTextKind{}
	add := func(pointer int, kind inventoryTextKind) (int, error) {
		off := module + pointer - 0xa000
		if off < module || off >= module+size {
			return 0, fmt.Errorf("裝備文字指標越界")
		}
		n := bytes.IndexByte(data[off:module+size], 0)
		if n < 0 {
			return 0, fmt.Errorf("裝備文字未結束")
		}
		for _, c := range data[off : off+n] {
			if c < 0x20 || c >= 0x7f {
				return 0, fmt.Errorf("裝備文字含未知控制碼")
			}
		}
		if n != 0 {
			if previous, exists := kinds[off]; exists && previous != kind {
				return 0, fmt.Errorf("裝備文字字型衝突")
			}
			kinds[off] = kind
		}
		return pointer + n + 1, nil
	}
	for _, table := range []struct {
		pointer, count, lines int
		kind                  inventoryTextKind
	}{
		{0xaab8, 7, 1, inventoryLargeWhite},
		{0xaaf3, 6, 2, inventoryLargeWhite},
		{0xab62, 8, 1, inventoryLargeWhite},
		{0xac32, 9, 2, inventoryLargeWhite},
		{0xacd9, 6, 2, inventorySmallWhite},
		{0xad67, 6, 2, inventorySmallWhite},
	} {
		off := module + table.pointer - 0xa000
		for i := 0; i < table.count; i++ {
			pointer := int(binary.LittleEndian.Uint16(data[off+i*2:]))
			for line := 0; line < table.lines; line++ {
				var err error
				kind := table.kind
				if table.pointer == 0xad67 && line == 1 {
					kind = inventoryShieldSecondLine
				}
				pointer, err = add(pointer, kind)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	for _, pointer := range []int{0xa9ff, 0xaa10, 0xaa19, 0xaa21, 0xaa92, 0xaa9a, 0xaaa2, 0xaaa8, 0xaaac} {
		kind := inventoryLargeWhite
		if pointer < 0xaa92 {
			kind = inventoryLargeColor
		}
		if _, err := add(pointer, kind); err != nil {
			return nil, err
		}
	}
	if len(kinds) != 72 {
		return nil, fmt.Errorf("裝備文字來源數量不符")
	}
	return kinds, nil
}

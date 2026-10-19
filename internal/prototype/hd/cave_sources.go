// 原版洞窟圖層解碼。只讀來源；完整色號重建與限制見規格012。
package hd

import (
	"encoding/binary"
	"fmt"
)

// CavePaint記錄4129之後的真實原版繪製，Mask以列優先記錄實際寫入位元。
type CavePaint struct {
	X, Y           int
	ID, Background byte
	SourceMask     [8]byte
	Mask           uint64
	Hook           uint16
	Steps, Cycles  uint64
}

func cavePaintMask(p CavePaint) (uint64, bool) {
	if p.Hook == 0x3272 {
		return ^uint64(0), p.ID < 128 && p.Background == 0 && p.SourceMask == [8]byte{}
	}
	if p.Hook != 0x3599 || p.Background >= 128 {
		return 0, false
	}
	if p.Background == 0 {
		return ^uint64(0), true
	}
	var mask uint64
	for y, bits := range p.SourceMask {
		for x := 0; x < 8; x++ {
			if bits&(128>>x) == 0 {
				mask |= uint64(1) << uint(y*8+x)
			}
		}
	}
	return mask, true
}

func cavePaintTimeline(paints []CavePaint, bodySteps, bodyCycles, steps, cycles uint64) bool {
	if len(paints) > 64 || bodySteps > steps || bodyCycles > cycles {
		return false
	}
	for _, p := range paints {
		if p.Steps < bodySteps || p.Steps > steps || p.Cycles < bodyCycles || p.Cycles > cycles {
			return false
		}
		bodySteps, bodyCycles = p.Steps, p.Cycles
	}
	return true
}

type cavePixel struct {
	X, Y  int
	Color byte
}
type caveCell struct {
	X, Y, Address int
	ID            byte
}
type caveObject struct {
	X, Y, Record, Template int
	ID, Palette            byte
	Tiles                  [4]byte
}
type caveDecoded struct {
	Background             []byte
	Actor, Weapon, Dynamic []cavePixel
	ActorRaster            [576]byte
	ActorX, ActorY         int
	WeaponRaster           [1024]byte
	WeaponX, WeaponY       int
	Cells                  []caveCell
	Objects                []caveObject
	Differences            int
}

func caveWord(b []byte, at int) int { return int(binary.LittleEndian.Uint16(b[at : at+2])) }
func cavePart(b []byte, at, n int) ([]byte, error) {
	if at < 0 || n < 0 || at > len(b)-n {
		return nil, fmt.Errorf("source bounds %x+%x/%x", at, n, len(b))
	}
	return b[at : at+n], nil
}

func caveStatic(r []byte, id byte) ([]byte, error) {
	v := make([]byte, 64)
	if id == 0 {
		return v, nil
	}
	if id >= 128 {
		return nil, fmt.Errorf("static id %x", id)
	}
	s, e := cavePart(r, 0x8030+(int(id)-1)*48, 48)
	if e != nil {
		return nil, e
	}
	for i := 0; i < 16; i++ {
		a, b, c := s[i*3], s[i*3+1], s[i*3+2]
		copy(v[i*4:], []byte{b >> 2, (b&3)<<4 | a>>4, (a&15)<<2 | c>>6, c & 63})
	}
	return v, nil
}

func caveQuad(r, p []byte, atlas, maskBase int, tile byte, pal []byte, background []byte, masked bool) ([]byte, []bool, error) {
	s, e := cavePart(r, atlas+int(tile)*32, 32)
	if e != nil {
		return nil, nil, e
	}
	m, e := cavePart(r, maskBase+int(tile)*8, 8)
	if e != nil {
		return nil, nil, e
	}
	if len(pal) != 16 || len(background) != 64 {
		return nil, nil, fmt.Errorf("quad input")
	}
	v := append([]byte(nil), background...)
	visible := make([]bool, 64)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			nibble := (caveWord(s, y*4+(x/4)*2) >> uint(12-(x%4)*4)) & 15
			if !masked || m[y]&(128>>x) == 0 {
				v[y*8+x] = pal[nibble]
				visible[y*8+x] = true
			}
		}
	}
	return v, visible, nil
}

func decodeCaveFrame(memory, bodyMemory, original []byte) (*caveDecoded, error) {
	return decodeCaveFrameWithPaints(memory, bodyMemory, original, nil)
}

func decodeCaveFrameWithPaints(memory, bodyMemory, original []byte, paints []CavePaint) (*caveDecoded, error) {
	if len(memory) != 1<<20 || len(bodyMemory) != 1<<20 || len(original) != 64000 {
		return nil, fmt.Errorf("frame sizes")
	}
	p := memory[0x1b10:0x11b10]
	bp := bodyMemory[0x1b10:0x11b10]
	if len(paints) > 64 {
		return nil, fmt.Errorf("native paint limit")
	}
	actorPos := caveWord(bp, 0x4fed)
	ax, ay := actorPos%320, actorPos/320
	for _, p := range paints {
		mask, ok := cavePaintMask(p)
		if !ok || mask != p.Mask || p.X < 48 || p.Y < 14 || p.X+8 > 272 || p.Y+8 > 158 || p.X >= ax+24 || p.X+8 <= ax || p.Y >= ay+24 || p.Y+8 <= ay {
			return nil, fmt.Errorf("native paint source")
		}
	}
	if caveWord(p, 0xff2c) != 0x11b1 || caveWord(bp, 0xff2c) != 0x11b1 {
		return nil, fmt.Errorf("resource segment")
	}
	r := memory[0x11b10:0x21b10]
	br := bodyMemory[0x11b10:0x21b10]
	// 主角及完成時刻各取真實來源，未消費區域不假定不可變。
	origin := caveWord(p, 0xff31)
	if origin < 0xe000 || origin >= 0xe900 {
		return nil, fmt.Errorf("map origin")
	}
	out := &caveDecoded{Background: append([]byte(nil), original...)}
	mapAddress := func(x, y int) int {
		n := (origin - 0xe000 + y*36 + 4 + x) % 0x900
		if n < 0 {
			n += 0x900
		}
		return 0xe000 + n
	}
	mapID := func(x, y int) byte { return p[mapAddress(x, y)] }
	backgroundID := func(id byte) byte {
		if id&128 != 0 {
			return p[0xed20+int(id&127)]
		}
		return id
	}
	inside := func(x, y int) bool { return x >= 48 && x < 272 && y >= 14 && y < 158 }
	put := func(x, y int, v []byte) {
		for yy := 0; yy < 8; yy++ {
			for xx := 0; xx < 8; xx++ {
				if inside(x+xx, y+yy) {
					out.Background[(y+yy)*320+x+xx] = v[yy*8+xx]
				}
			}
		}
	}
	for y := 0; y < 18; y++ {
		for x := 0; x < 28; x++ {
			id := backgroundID(mapID(x, y))
			tile, e := caveStatic(r, id)
			if e != nil {
				return nil, e
			}
			put(48+x*8, 14+y*8, tile)
			out.Cells = append(out.Cells, caveCell{48 + x*8, 14 + y*8, mapAddress(x, y), id})
		}
	}
	for y := -1; y < 18; y++ {
		for x := -1; x < 28; x++ {
			id := mapID(x, y)
			if id < 128 {
				continue
			}
			record := caveWord(p, 0xc010) + int(id&127)*16
			d, e := cavePart(p, record, 16)
			if e != nil {
				return nil, e
			}
			table := 0xa030
			if d[5]&128 != 0 {
				table = 0xa070
			}
			entry := caveWord(p, table+int(d[4]&31)*2) + int(d[6]&15)*5
			t, e := cavePart(p, entry, 5)
			if e != nil {
				return nil, e
			}
			pi := int(t[0])
			if p[0xff34] == 0 && d[5]&32 != 0 {
				pi += 3
			}
			if pi >= 8 {
				return nil, fmt.Errorf("unobserved object palette %d", pi)
			}
			pal, e := cavePart(p, caveWord(p, 0x4f8c+pi*2), 16)
			if e != nil {
				return nil, e
			}
			object := caveObject{X: 48 + x*8, Y: 14 + y*8, Record: record, Template: entry, ID: id, Palette: byte(pi)}
			copy(object.Tiles[:], t[1:])
			out.Objects = append(out.Objects, object)
			for q, tile := range t[1:] {
				gx, gy := x+q%2, y+q/2
				bg := backgroundID(mapID(gx, gy))
				base, e := caveStatic(r, bg)
				if e != nil {
					return nil, e
				}
				decoded, visible, e := caveQuad(r, p, 0x4000, 0xa000, tile, pal, base, bg != 0)
				if e != nil {
					return nil, e
				}
				px, py := 48+gx*8, 14+gy*8
				put(px, py, decoded)
				for yy := 0; yy < 8; yy++ {
					for xx := 0; xx < 8; xx++ {
						if visible[yy*8+xx] && inside(px+xx, py+yy) {
							out.Dynamic = append(out.Dynamic, cavePixel{px + xx, py + yy, decoded[yy*8+xx]})
						}
					}
				}
			}
		}
	}
	// 主角僅接受已觀測之裝備／姿態分支，未知來源安全拒絕。
	if bp[0x93] != 0 || bp[0xff41] != 0 {
		return nil, fmt.Errorf("unobserved actor profile")
	}
	profiles := 0
	for _, off := range []int{0xe8, 0xff39, 0xff3a} {
		if bp[off] == 255 {
			profiles++
		} else if bp[off] != 0 {
			return nil, fmt.Errorf("unobserved actor profile")
		}
	}
	if profiles > 1 {
		return nil, fmt.Errorf("unobserved combined actor profile")
	}
	if bp[0xff40] != 0 && bp[0xff40] != 255 {
		return nil, fmt.Errorf("actor attack flag")
	}
	pal, e := cavePart(bp, caveWord(bp, 0x4f8c+int(bp[0xff36]&3)*2), 16)
	if e != nil {
		return nil, e
	}
	actor := make([]byte, 576)
	opaque := make([]bool, 576)
	layer := func(table, count, start int) error {
		ids, e := cavePart(br, table, count)
		if e != nil {
			return e
		}
		for n, id := range ids {
			if id == 0 {
				continue
			}
			v, visible, e := caveQuad(br, bp, 0x6333, 0xd000, id, pal, make([]byte, 64), true)
			if e != nil {
				return e
			}
			q := n + start
			for yy := 0; yy < 8; yy++ {
				for xx := 0; xx < 8; xx++ {
					if visible[yy*8+xx] {
						i := (q/3*8+yy)*24 + q%3*8 + xx
						actor[i] = v[yy*8+xx]
						opaque[i] = true
					}
				}
			}
		}
		return nil
	}
	back, front := -1, 0x6117
	if bp[0xc2]&1 != 0 {
		front = 0x61b9
	}
	if bp[0xff40] != 0 {
		offset := int(bp[0xff3f]>>1) * 9
		if bp[0xc2]&1 != 0 {
			back = 0x61b9 + 0x24 + offset
			front = 0x625b + offset
		} else {
			back = 0x62c7 + offset
			front = 0x6117 + 0x24 + offset
		}
	} else {
		if bp[0xff38] == 0 && bp[0xe7] != 128 {
			phase := (bp[0xe7] + 2) & 3
			if phase&1 == 0 {
				back = front + int(phase)*9
			}
		}
		if bp[0xff38] != 0 || bp[0xe7] == 128 {
			front += 0x1b
		} else {
			front += int(bp[0xe7]&3) * 9
		}
	}
	count, start := 9, 0
	if bp[0xff38] != 0 {
		count, start = 6, 3
	}
	// 01B1:3AAE..3ABB：三個已驗非零分支均跳過後手。
	if profiles != 0 {
		back = -1
	}
	if back >= 0 {
		if e := layer(back, count, start); e != nil {
			return nil, e
		}
	}
	table := 0x6000
	if bp[0xc2]&1 != 0 {
		table = 0x6075
	}
	switch {
	case bp[0xff38] != 0:
		table += 0x2d
	case bp[0xff3d]&128 != 0:
		table += 0x3f
	case bp[0xff42] == 1:
		table += 0x48
	case bp[0xff42] == 2:
		table += 0x51
	case bp[0xff3d] == 127:
		table += 0x36
	case bp[0xe7] == 128:
		table += 0x24
	default:
		table += int(bp[0xe7]&3) * 9
	}
	// 01B1:3B94..3BAB、3BE7..3BF0：E8分支沿朝向及相位。
	if bp[0xe8] == 255 {
		table = 0x6000 + 0x5a + int(bp[0xe7]&3)*9
		if bp[0xc2]&1 != 0 {
			table += 0x75
		}
	}
	// 01B1:3B8A..3B92、3BE7..3BF0：FF39的九格模板。
	if bp[0xff39] == 255 {
		table = 0x60ea + int(bp[0xe7]&3)*9
	}
	// 01B1:3B80..3B88：FF3A選固定九格。
	if bp[0xff3a] == 255 {
		table = 0x610e
	}
	if e := layer(table, 9, 0); e != nil {
		return nil, e
	}
	// 01B1:3BFD..3C04及3C16..3C26：已驗分支與0093=0時無前手。
	if profiles == 0 {
		if e := layer(front, count, start); e != nil {
			return nil, e
		}
	}
	pos := caveWord(bp, 0x4fed)
	bx, by := pos%320, pos/320
	out.ActorX, out.ActorY = bx, by
	for i := range out.ActorRaster {
		out.ActorRaster[i] = 255
		if opaque[i] {
			out.ActorRaster[i] = actor[i]
		}
	}
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			erased := false
			for _, p := range paints {
				px, py := bx+x-p.X, by+y-p.Y
				if px >= 0 && px < 8 && py >= 0 && py < 8 && p.Mask&(uint64(1)<<uint(py*8+px)) != 0 {
					erased = true
					break
				}
			}
			if opaque[y*24+x] && !erased {
				out.Actor = append(out.Actor, cavePixel{bx + x, by + y, actor[y*24+x]})
			}
		}
	}
	for i := range out.WeaponRaster {
		out.WeaponRaster[i] = 255
	}
	weaponPosition := caveWord(p, 0x4ff7)
	out.WeaponX, out.WeaponY = weaponPosition%320, weaponPosition/320
	if p[0xff44] != 0 {
		template, e := cavePart(r, caveWord(p, 0x4ff9), 16)
		if e != nil {
			return nil, e
		}
		base := caveWord(r, 0xb000)
		pos := caveWord(p, 0x4ff7)
		wx, wy := pos%320, pos/320
		for n, tile := range template {
			if tile == 255 {
				continue
			}
			s, e := cavePart(r, base+int(tile)*16, 16)
			if e != nil {
				return nil, e
			}
			for y := 0; y < 8; y++ {
				bits := int(s[y*2])<<8 | int(s[y*2+1])
				for x := 0; x < 8; x++ {
					v := (bits >> uint(14-x*2)) & 3
					if v != 0 {
						color := p[0x4ff5]
						if v == 3 {
							color = p[0x4ff6]
						}
						out.Weapon = append(out.Weapon, cavePixel{wx + n/4*8 + x, wy + n%4*8 + y, color})
						out.WeaponRaster[(n%4*8+y)*32+n/4*8+x] = color
					}
				}
			}
		}
	}
	expected := append([]byte(nil), out.Background...)
	for _, layer := range [][]cavePixel{out.Actor, out.Weapon} {
		for _, v := range layer {
			if !inside(v.X, v.Y) {
				return nil, fmt.Errorf("actor or weapon outside viewport")
			}
			expected[v.Y*320+v.X] = v.Color
		}
	}
	for y := 14; y < 158; y++ {
		for x := 48; x < 272; x++ {
			i := y*320 + x
			if expected[i] != original[i] {
				out.Differences++
			}
		}
	}
	if out.Differences != 0 {
		return out, fmt.Errorf("full viewport differs: %d", out.Differences)
	}
	return out, nil
}

package glyphpatch

import (
	"image"

	"github.com/wicanr2/dosgolem/oracle"
)

const (
	vram   = 0xa0000
	hidden = Hidden << 4
	// buffer 是驅動 289A／28D9 存放方框的 CS+3000h 段，即 31B1:0000。
	buffer = 0x31b10
)

// Gate 決定每次原版字形是否進入畫面，並維護被壓下字形的虛擬平面。
// 平面依原版逐 byte 的實際寫入維護，疊回 VRAM 後在任一指令邊界都等於原版畫面，
// 供所有 HD 與中文守衛使用；顯示只用 VRAM。
type Gate struct {
	o *oracle.Oracle
	// expect 是已認領、尚未畫出的 27E9 字形返回位址。
	expect map[uint16]bool
	// lines 是 291A 多行字串的持續認領，只壓返回 294D 的字形。
	lines bool
	// linesRet 是 HoldGlyphs 指定的另一個返回位址，例如洞窟告示 740C 迴圈的 7465。
	linesRet uint16
	// text 是下一次小字串（22FE／2333）目的段載入的認領。
	text bool

	pix   [64000]uint8
	set   [64000]bool
	count int
	// moving 是驅動正在執行的方框搬移；期間 VRAM 與緩衝寫入依搬移方向更新平面。
	moving *move
	// saved 依 31B1 緩衝偏移保存 289A 存走的平面值。
	saved map[uint16]uint8
	// keep 是正在執行的讀改寫指令中保留原值的 VRAM 點，這些寫入不清平面。
	keep [2]int

	Suppressed, Passed, Anomalies, Revealed int
}

type move struct {
	kind          byte // 's' 捲動、'v' 存框、'r' 還原
	vram          int  // 方框左上角的 VRAM 偏移
	buffer        uint16
	width, height int
}

var gates = map[*oracle.Oracle]*Gate{}

// Enable 讓 o 使用字形閘。只有 patch 版驅動載入後掛鉤才會寫入；原版驅動時不改任何狀態。
func Enable(o *oracle.Oracle) *Gate {
	g := &Gate{o: o, saved: map[uint16]uint8{}, expect: map[uint16]bool{}, keep: [2]int{-1, -1}}
	gates[o] = g
	o.OnCall(oracle.Far(0x1b1, 0x2832), g.glyph8)
	o.OnCall(oracle.Far(0x1b1, 0x22fe), g.string5)
	o.OnCall(oracle.Far(0x1b1, 0x2333), g.string5)
	// 方框搬移的進入與返回前一點（pop ds）。
	o.OnCall(oracle.Far(0x1b1, 0x2857), func(o *oracle.Oracle) { g.begin(o, 's') })
	o.OnCall(oracle.Far(0x1b1, 0x289a), func(o *oracle.Oracle) { g.begin(o, 'v') })
	o.OnCall(oracle.Far(0x1b1, 0x28d9), func(o *oracle.Oracle) { g.begin(o, 'r') })
	for _, end := range []uint16{0x2898, 0x28d7, 0x2918} {
		o.OnCall(oracle.Far(0x1b1, end), func(*oracle.Oracle) { g.moving = nil })
	}
	// 驅動 20B6 畫框線時以遮罩對兩點做 and／or；遮罩 byte 使該點保持原值時，
	// 原版是把文字像素原樣寫回，平面要保留。四條指令後一條清除標記。
	for _, rmw := range []struct {
		at, next uint16
		bx       bool
		keepIf   uint8
	}{{0x20ba, 0x20bd, false, 0xff}, {0x20c1, 0x20c4, false, 0x00}, {0x20d7, 0x20da, true, 0xff}, {0x20de, 0x20e1, true, 0x00}} {
		rmw := rmw
		o.OnCall(oracle.Far(0x1b1, rmw.at), func(o *oracle.Oracle) {
			mask := o.AX()
			if rmw.bx {
				mask = o.BX()
			}
			g.keep = [2]int{-1, -1}
			if o.ES() == Visible {
				for i := 0; i < 2; i++ {
					if uint8(mask>>(8*i)) == rmw.keepIf {
						g.keep[i] = int(o.DI()) + i
					}
				}
			}
		})
		o.OnCall(oracle.Far(0x1b1, rmw.next), func(*oracle.Oracle) { g.keep = [2]int{-1, -1} })
	}
	// 被壓下的字形由原版自己寫到 B000；照實際寫入放進平面。
	o.WatchEveryWrite(hidden, hidden+63999, func(a uint32, _, v uint8) { g.put(int(a-hidden), v) })
	// 每一次 VRAM 寫入都更新該點；只在值改變時通知的監看會漏掉以底色清框。
	o.WatchEveryWrite(vram, vram+63999, func(a uint32, _, _ uint8) { g.vramWrite(int(a - vram)) })
	o.WatchEveryWrite(buffer, buffer+0xffff, func(a uint32, _, _ uint8) { g.bufferWrite(uint16(a - buffer)) })
	return g
}

// For 回傳 o 的字形閘；未啟用時為 nil。
func For(o *oracle.Oracle) *Gate { return gates[o] }

// Claim 讓下一個返回 ret 的 27E9 字形不進畫面。覆繪層只在已核對來源且會畫中文時呼叫。
func Claim(o *oracle.Oracle, rets ...uint16) {
	if g := gates[o]; g != nil {
		for _, ret := range rets {
			g.expect[ret] = true
		}
	}
}

// Unclaim 撤回尚未畫出的認領，例如字串結束時沒有用到的陰影呼叫。
func Unclaim(o *oracle.Oracle, rets ...uint16) {
	if g := gates[o]; g != nil {
		for _, ret := range rets {
			delete(g.expect, ret)
		}
	}
}

// ClaimText 讓下一次小字串（22DB、22CD、22BF 進入）整串不進畫面。
func ClaimText(o *oracle.Oracle) {
	if g := gates[o]; g != nil {
		g.text = true
	}
}

// HoldLines／ReleaseLines 涵蓋 291A 多行字串，從進入到 2939 返回。
func HoldLines(o *oracle.Oracle) {
	if g := gates[o]; g != nil {
		g.lines = true
	}
}

// HoldGlyphs 與 HoldLines 相同，另外壓返回 ret 的字形，給不經 291A 的逐字迴圈使用。
func HoldGlyphs(o *oracle.Oracle, ret uint16) {
	if g := gates[o]; g != nil {
		g.lines = true
		g.linesRet = ret
	}
}

func ReleaseLines(o *oracle.Oracle) {
	if g := gates[o]; g != nil {
		g.lines = false
		g.linesRet = 0
	}
}

func (g *Gate) setVar(hide bool) {
	seg := uint16(Visible)
	if hide {
		seg = Hidden
	}
	g.o.SetWord(oracle.Far(0x1b1, VarOffset), seg)
}

func (g *Gate) put(p int, c uint8) {
	if p < 0 || p >= 64000 {
		return
	}
	if !g.set[p] {
		g.set[p] = true
		g.count++
	}
	g.pix[p] = c
}

func (g *Gate) clear(p int) {
	if p >= 0 && p < 64000 && g.set[p] {
		g.set[p] = false
		g.count--
	}
}

// introMode 回報 FF77 非零的片頭 256 色模式。片頭轉場會讀出 VRAM、加遮罩再寫回，
// 被壓下的字形無法由寫入監看追蹤，因此片頭一律照原版畫字。
func introMode(o *oracle.Oracle) bool { return o.Byte(oracle.Far(0x1b1, 0xff77)) != 0 }

// glyph8 在 2832 目的段載入前決定；此時堆疊為 DS、返回位址。
func (g *Gate) glyph8(o *oracle.Oracle) {
	if !patchedAt(o, 0x2832) {
		return
	}
	ret := o.StackWord(1)
	hide := (g.expect[ret] || g.lines && (ret == 0x294d || g.linesRet != 0 && ret == g.linesRet)) && !introMode(o)
	delete(g.expect, ret)
	g.setVar(hide)
	if hide {
		g.Suppressed++
	} else {
		g.Passed++
	}
}

func (g *Gate) string5(o *oracle.Oracle) {
	if !patchedAt(o, o.IP().Off) {
		return
	}
	hide := g.text && !introMode(o)
	g.text = false
	if o.DSReg() != 0x1b1 {
		// patch 版以 DS 讀目的段變數；原版此處也以 DS 讀字型與色號，DS 不是 01B1 屬未知情形。
		g.Anomalies++
	}
	g.setVar(hide)
	if hide {
		g.Suppressed++
	} else {
		g.Passed++
	}
}

// begin 記錄驅動方框搬移的參數。2857 以 BL 列、BH 八點欄定位，由下一列往上複製；
// 289A／28D9 以 AL 列、AH 八點欄定位，DI 為 31B1 緩衝偏移。三者都是 CL 列、每列 CH*8 bytes。
func (g *Gate) begin(o *oracle.Oracle, kind byte) {
	cx := o.CX()
	m := &move{kind: kind, width: int(cx>>8) * 8, height: int(cx & 0xff), buffer: o.DI()}
	if kind == 's' {
		bx := o.BX()
		m.vram = int(bx&0xff)*320 + int(bx>>8)*8
	} else {
		ax := o.AX()
		m.vram = int(ax&0xff)*320 + int(ax>>8)*8
	}
	g.moving = m
}

// vramWrite 在每一次 VRAM 寫入時更新平面：一般寫入清除；捲動取下一列，還原取緩衝。
func (g *Gate) vramWrite(p int) {
	if p == g.keep[0] || p == g.keep[1] {
		return
	}
	m := g.moving
	switch {
	case m != nil && m.kind == 's':
		if src := p + 320; src < 64000 && g.set[src] {
			g.put(p, g.pix[src])
			return
		}
	case m != nil && m.kind == 'r':
		if d := p - m.vram; d >= 0 && d%320 < m.width {
			if v, ok := g.saved[m.buffer+uint16(d/320*m.width+d%320)]; ok {
				g.put(p, v)
				return
			}
		}
	}
	g.clear(p)
}

// bufferWrite 在 289A 存框時記下對應 VRAM 點的平面值；其他寫入讓該緩衝位置失效。
func (g *Gate) bufferWrite(b uint16) {
	if m := g.moving; m != nil && m.kind == 'v' && m.width > 0 {
		if d := int(b - m.buffer); d >= 0 && d < m.width*m.height {
			if p := m.vram + d/m.width*320 + d%m.width; p < 64000 && g.set[p] {
				g.saved[b] = g.pix[p]
				return
			}
		}
	}
	delete(g.saved, b)
}

// Reveal 把區域內被壓下的字形照原版寫回 VRAM。覆繪層拒收已認領的文字時呼叫，
// 畫面退回原版英文而不是留下空白；寫回的值即原版 VRAM 的值，平面隨寫入清除。
func Reveal(o *oracle.Oracle, rects ...image.Rectangle) {
	g := gates[o]
	if g == nil || g.count == 0 {
		return
	}
	for _, r := range rects {
		r = r.Intersect(image.Rect(0, 0, 320, 200))
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				if p := y*320 + x; g.set[p] {
					o.SetByte(oracle.Far(Visible, uint16(p)), g.pix[p])
					g.Revealed++
				}
			}
		}
	}
}

// Indexed 回傳守衛用的等同原版畫面：VRAM 疊上被壓下的字形。
func Indexed(o *oracle.Oracle) []byte {
	raw := o.Indexed()
	if g := gates[o]; g != nil && g.count > 0 && len(raw) == 64000 {
		for p, on := range g.set {
			if on {
				raw[p] = g.pix[p]
			}
		}
	}
	return raw
}

// ScreenRGB 是 Indexed 的 RGB 版本；平面點依目前色盤上色。
func ScreenRGB(o *oracle.Oracle) (int, int, []byte) {
	w, h, rgb := o.ScreenRGB()
	if g := gates[o]; g != nil && g.count > 0 && len(rgb) == 64000*3 {
		pal := o.Palette()
		for p, on := range g.set {
			if on {
				c := pal[g.pix[p]]
				rgb[p*3], rgb[p*3+1], rgb[p*3+2] = c[0], c[1], c[2]
			}
		}
	}
	return w, h, rgb
}

// Hidden 依序回報目前不在畫面上的字形點及其原版色號。
func (g *Gate) Hidden(fn func(p int, color uint8)) {
	if g == nil || g.count == 0 {
		return
	}
	for p, on := range g.set {
		if on {
			fn(p, g.pix[p])
		}
	}
}

// Buffer 回傳 289A 保存的平面值，供差分測試重建 31B1 緩衝。
func (g *Gate) Buffer() map[uint16]uint8 { return g.saved }

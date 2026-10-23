package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/glyphpatch"
)

// 閃爍探針以固定指令數切片重播正常鍵序，每片都走與 GUI 相同的 compose。
// 每個切片點都是 Draw 可能落下的位置，所以記錄的是「可能被呈現」的全部畫面，
// 不是實際時間下碰巧取樣到的幾張。不開視窗、不用實際時間，同一輸入重跑結果相同。

// probeGlyph 是一次原版 8×8 字形繪製（01B1:27E9）及其在畫面上的生命週期。
type probeGlyph struct {
	Step        uint64 `json:"step"`
	Scene       string `json:"scene"`
	Caller      string `json:"caller"`
	Char        uint8  `json:"char"`
	Color       uint8  `json:"color"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Visible     int    `json:"visible_composes"`
	TextCovered int    `json:"text_covered_composes"`
	HDCovered   int    `json:"hd_covered_composes"`
	// PatchHidden 是沒有任何中文層在畫、只靠字形 patch 不顯示的合成點數。
	PatchHidden  int    `json:"patch_hidden_composes"`
	FirstVisible uint64 `json:"first_visible_step,omitempty"`
	FirstText    uint64 `json:"first_text_covered_step,omitempty"`
	FirstHD      uint64 `json:"first_hd_covered_step,omitempty"`
	End          uint64 `json:"end_step"`
	// Coincident 是字形點同色、但同格空白點也同色而不計為可見的合成點數。
	Coincident int `json:"coincident_composes,omitempty"`
	pixels     []int
	blank      []int
	beforeText int
}

func anyText(layers map[string]bool) bool {
	for k, v := range layers {
		if v && strings.HasPrefix(k, "text_") {
			return true
		}
	}
	return false
}

// shown 計算字形仍在原版畫面、且畫布在該點正好顯示原版顏色的像素數。
func (g *game) shown(glyph *probeGlyph, raw, rawRGB []byte) (present, visible int, coincident bool) {
	var ink []byte
	for _, p := range glyph.pixels {
		if raw[p] != glyph.Color {
			continue
		}
		present++
		ink = rawRGB[p*3 : p*3+3]
		// 原版底圖以整塊 3×3 繪出；中文柔邊筆畫同色時不會剛好填滿一整塊。
		if blockIs(g.canvas, p%320*3, barHeight+p/320*3, ink) {
			visible++
		}
	}
	// 同格的空白點也呈現字色，或現存點只有少數可見時，畫面上的是同色的中文粗筆畫或保留的
	// 數字圖形，不是原版字形；真正露出的英文整個字都看得到。
	if visible > 0 && visible*2 < present {
		return present, 0, true
	}
	if visible > 0 {
		native := 0
		for _, p := range glyph.blank {
			if raw[p] == glyph.Color {
				continue
			}
			if blockIs(g.canvas, p%320*3, barHeight+p/320*3, ink) {
				return present, 0, true
			}
			if blockIs(g.canvas, p%320*3, barHeight+p/320*3, rawRGB[p*3:p*3+3]) {
				native++
			}
		}
		// 規格067：字形要看得出來，同格須有原版底色不同於字色的點照原版顯示。沒有的話，畫面上的字色
		// 是覆繪層的底色或筆畫，例如中文底色上的黑色陰影字，不算露出。
		if native == 0 {
			return present, 0, true
		}
	}
	return
}

type probeRow struct {
	Step    uint64          `json:"step"`
	Scene   string          `json:"scene"`
	HD      bool            `json:"hd"`
	Layers  map[string]bool `json:"layers"`
	English int             `json:"english_pixels"`
	Town    string          `json:"town,omitempty"`
	Pending bool            `json:"town_pending,omitempty"`
	// Refusal 是城鎮完成畫面未使用時的原因，供 HD 中斷歸因。
	Refusal  string `json:"town_refusal,omitempty"`
	Outside  string `json:"town_outside,omitempty"`
	Backdrop string `json:"town_background_reason,omitempty"`
	// Burst 是合成前 2000 指令內的 VRAM 寫入 byte 數。
	Burst int `json:"burst"`
	// Palette 是當下色盤的 SHA-256 前 12 碼，供淡入淡出期間的 HD 比對歸因。
	Palette string `json:"palette"`
	// Revealed 是上一個合成點之後，覆繪層拒收已認領文字而把原版字形寫回畫面的像素數。規格067。
	Revealed int `json:"revealed,omitempty"`
	// RevealedVisible 是寫回的點在這個合成點仍是寫回的色號、且畫布照原版顯示的數目，即寫回後真的露出的英文。
	RevealedVisible int `json:"revealed_visible,omitempty"`
	// HUDUndrawn 是原版欄位字形完整、卻沒有畫上中文的狀態列欄數。
	HUDUndrawn int `json:"hud_undrawn,omitempty"`
}

// loadHDNow 在探針內同步完成場景圖組載入；GUI 版在背景載入並暫停原版。
func (g *game) loadHDNow() {
	if !g.hdEnabled {
		return
	}
	for _, job := range g.hdJobs {
		if job.finished || !job.when() {
			continue
		}
		fmt.Fprintf(os.Stderr, "探針同步載入場景HD：%s（%d步）\n", job.name, g.o.Steps())
		job.load()
		if job.apply != nil {
			job.apply()
		}
		job.finished = true
	}
}

// glyphPixels 依 27E9 的實際計算重建本次字形像素：DS=CS=01B1，F500 為字型指標，
// FF77 非零時色號為 AH*17，否則查 24EA 表；DI=CL*320+BX，只寫字形位元為 1 的點。
func glyphPixels(o *oracle.Oracle) (pixels, blank []int, color uint8) {
	ah := uint8(o.AX() >> 8)
	if o.Byte(oracle.Far(0x1b1, 0xff77)) != 0 {
		color = ah * 17
	} else {
		color = o.Byte(oracle.Far(0x1b1, 0x24ea+uint16(ah)))
	}
	font := uint16(uint8(o.AX())-0x20)*8 + o.Word(oracle.Far(0x1b1, 0xf500))
	di := uint16(o.CX()&0xff)*320 + o.BX()
	for row, bits := range o.Bytes(oracle.Far(0x1b1, font), 8) {
		for col := 0; col < 8; col++ {
			p := int(di + uint16(row*320+col))
			switch {
			case p >= 64000:
			case bits&(0x80>>col) != 0:
				pixels = append(pixels, p)
			default:
				blank = append(blank, p)
			}
		}
	}
	return pixels, blank, color
}

func (g *game) runFlickerProbe(commandsPath, out string, slice uint64, census bool, shots []uint64) error {
	data, err := os.ReadFile(commandsPath)
	if err != nil {
		return err
	}
	create := func(suffix string) (*os.File, *bufio.Writer, error) {
		f, err := os.OpenFile(out+suffix, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return nil, nil, err
		}
		return f, bufio.NewWriter(f), nil
	}
	rowsFile, rows, err := create(".composes.jsonl")
	if err != nil {
		return err
	}
	defer rowsFile.Close()
	glyphFile, glyphs, err := create(".glyphs.jsonl")
	if err != nil {
		return err
	}
	defer glyphFile.Close()
	scene := g.sceneKey()
	burst := newVRAMBurst(g.o)
	revealed, blackGlyphs := 0, 0
	var active []*probeGlyph
	g.o.OnCall(oracle.Far(0x1b1, 0x27e9), func(o *oracle.Oracle) {
		pixels, blank, color := glyphPixels(o)
		ret := o.StackWord(0)
		// 規格067：色號0的字形只用於陰影（與彩色本字成對）與看不見的標記；黑底上畫完或擦掉都同色，
		// 追蹤不會結束，之後別的圖畫到同一格就誤判露出。不追蹤，另計數量。
		if color == 0 {
			blackGlyphs++
			return
		}
		active = append(active, &probeGlyph{Step: o.Steps(), Scene: scene, Caller: fmt.Sprintf("%04x", ret), Char: uint8(o.AX()), Color: color, X: int(o.BX()), Y: int(o.CX() & 0xff), pixels: pixels, blank: blank})
	})
	// 中文層開始前先記一次，分辨字形是被中文層還是 HD 層蓋掉。
	g.beforeText = func() {
		raw := glyphpatch.Indexed(g.o)
		_, _, rgb := glyphpatch.ScreenRGB(g.o)
		if len(raw) != 64000 {
			return
		}
		for _, glyph := range active {
			_, glyph.beforeText, _ = g.shown(glyph, raw, rgb)
		}
	}
	writers := map[string]map[string]int{}
	if census {
		g.o.OnWrite(0xa0000, 0xaf9ff, func(o *oracle.Oracle, w oracle.WriteHit) {
			if writers[scene] == nil {
				writers[scene] = map[string]int{}
			}
			writers[scene][fmt.Sprintf("%05x", w.At)]++
		})
	}
	encode := func(w *bufio.Writer, v any) error {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		_, err = w.Write(append(b, '\n'))
		return err
	}
	exited := false
	advance := func(n uint64) error {
		for n > 0 && !exited {
			s := min(slice, n)
			if err := g.o.Run(s); err != nil {
				var exit *oracle.ExitError
				if !errors.As(err, &exit) {
					return err
				}
				exited = true
			}
			n -= s
			g.loadHDNow()
			scene = g.sceneKey()
			presented, _, w, h, raw, rawRGB := g.compose()
			// 指定步數之後的第一個合成點存成圖片，供人工判讀中斷期間的實際畫面。
			// 另存守衛輸入的色號與色盤，供轉場遮罩的逐點判讀。
			for len(shots) > 0 && shots[0] <= g.o.Steps() {
				name := fmt.Sprintf("%s.shot-%d", out, g.o.Steps())
				if err := savePNG(name+".png", g.canvas); err != nil {
					return err
				}
				dump := append([]byte(nil), raw...)
				for _, c := range g.o.Palette() {
					dump = append(dump, c[:]...)
				}
				if err := os.WriteFile(name+".indexed", dump, 0644); err != nil {
					return err
				}
				shots = shots[1:]
			}
			row := probeRow{Step: g.o.Steps(), Scene: scene, HD: g.hdEnabled, Layers: g.layers, Burst: burst.Recent(4), Palette: paletteHash(g.o.Palette())}
			if gate := glyphpatch.For(g.o); gate != nil {
				row.Revealed, revealed = gate.Revealed-revealed, gate.Revealed
				for _, p := range gate.TakeRevealed() {
					if raw[p] != 0 && blockIs(g.canvas, p%320*3, barHeight+p/320*3, rawRGB[p*3:p*3+3]) {
						row.RevealedVisible++
					}
				}
			}
			if g.hud != nil {
				row.HUDUndrawn = g.hud.Undrawn
			}
			if g.townDialogue != nil {
				state := g.townDialogue.State()
				row.Town, _ = state["key"].(string)
				row.Pending, _ = state["pending"].(bool)
			}
			if g.townBackgroundHD != nil && !g.layers["town_background"] {
				row.Backdrop = g.townBackgroundHD.Reason
			}
			if presented == nil && g.townFrames != nil {
				if state := g.townFrames.State(); state != nil {
					row.Refusal, _ = state["reason"].(string)
					if p, ok := state["outside_pixel"]; ok {
						row.Outside = fmt.Sprint(p)
					}
					if row.Refusal == "" && state["used"] == false {
						row.Refusal = "frame guard"
					}
				}
			}
			kept := active[:0]
			for _, glyph := range active {
				present, visible, coincident := 0, 0, false
				if w == 320 && h == 200 {
					present, visible, coincident = g.shown(glyph, raw, rawRGB)
				}
				if coincident {
					glyph.Coincident++
				}
				// 字形剩不到一半即視為已擦除或移動。下一字的陰影會蓋到前一字邊緣，片頭轉場也逐點清字，
				// 只要有一點被改寫就停止追蹤會漏掉仍看得到的殘字；同色背景另由空白點檢查排除。
				if present == 0 || present*2 < len(glyph.pixels) {
					glyph.End = g.o.Steps()
					if err := encode(glyphs, glyph); err != nil {
						return err
					}
					continue
				}
				if visible > 0 {
					if glyph.Visible == 0 {
						glyph.FirstVisible = g.o.Steps()
					}
					glyph.Visible++
					row.English += visible
				} else if glyph.beforeText > 0 && !anyText(g.layers) {
					// 原版底圖本來看得到這個字，但沒有任何中文層在畫：只靠 patch 隱藏，畫面是空白。
					glyph.PatchHidden++
				} else if glyph.beforeText > 0 {
					if glyph.TextCovered == 0 {
						glyph.FirstText = g.o.Steps()
					}
					glyph.TextCovered++
				} else {
					if glyph.HDCovered == 0 {
						glyph.FirstHD = g.o.Steps()
					}
					glyph.HDCovered++
				}
				glyph.beforeText = 0
				kept = append(kept, glyph)
			}
			active = kept
			if err := encode(rows, row); err != nil {
				return err
			}
		}
		return nil
	}
	press := func(scans []uint8, n uint64) error {
		for _, s := range scans {
			g.o.KeyDown(s)
		}
		err := advance(n)
		for _, s := range scans {
			g.o.KeyUp(s)
		}
		return err
	}
	number := func(s string) (uint64, error) { return strconv.ParseUint(s, 0, 64) }
	for lineNo, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") || exited {
			continue
		}
		var err error
		switch f[0] {
		case "run":
			var n uint64
			if n, err = number(f[1]); err == nil {
				err = advance(n)
			}
		case "key", "keys":
			var scans []uint8
			for _, s := range strings.Split(f[1], ",") {
				v, e := number(s)
				if e != nil {
					return e
				}
				scans = append(scans, uint8(v))
			}
			var n uint64
			if n, err = number(f[2]); err == nil {
				err = press(scans, n)
			}
		case "down", "up":
			var v uint64
			if v, err = number(f[1]); err == nil {
				if f[0] == "down" {
					g.o.KeyDown(uint8(v))
				} else {
					g.o.KeyUp(uint8(v))
				}
			}
		case "king":
			// 與 save_load_test 的 king 相同：冷開機後固定 Space、右移與進門，到國王初訪。
			type event struct {
				at   uint64
				scan uint8
				down bool
			}
			var events []event
			for _, s := range []uint64{12000000, 20000000, 28000000, 36000000, 44000000, 52000000, 86000000, 94000000, 102000000, 110000000, 118000000, 126000000} {
				events = append(events, event{s, 0x39, true}, event{s + 50000, 0x39, false})
			}
			events = append(events, event{60000000, 0x4d, true}, event{62000000, 0x4d, false}, event{64000000, 0x48, true}, event{64250000, 0x48, false})
			sort.SliceStable(events, func(i, j int) bool { return events[i].at < events[j].at })
			for _, e := range events {
				if err = advance(e.at - g.o.Steps()); err != nil {
					break
				}
				if e.down {
					g.o.KeyDown(e.scan)
				} else {
					g.o.KeyUp(e.scan)
				}
			}
			if err == nil {
				err = advance(130000000 - g.o.Steps())
			}
		case "poke":
			// 結構驗證用：直接寫一個 byte 到原版記憶體，例如 poke 0x1b1:0x49 1 讓小屋載入結局。
			// 只供正常遊玩到不了的畫面，收據須標明不是正常玩家路徑。規格067。
			var seg, off, v uint64
			addr := strings.SplitN(f[1], ":", 2)
			if len(addr) != 2 || len(f) < 3 {
				err = fmt.Errorf("第 %d 行：poke 格式為 段:位移 值", lineNo+1)
			} else if seg, err = number(addr[0]); err == nil {
				if off, err = number(addr[1]); err == nil {
					if v, err = number(f[2]); err == nil {
						g.o.SetByte(oracle.Far(uint16(seg), uint16(off)), uint8(v))
					}
				}
			}
		case "shot":
			// 鍵序中的具名截圖存最後一次合成的畫面，與 GUI 同一路徑。
			if len(f) > 1 {
				err = savePNG(fmt.Sprintf("%s.shot-%s.png", out, f[1]), g.canvas)
				if err == nil {
					// 同時存 01B1 段記憶體，供辨識當下載入的模組。
					err = os.WriteFile(fmt.Sprintf("%s.shot-%s.mem", out, f[1]), g.o.Bytes(oracle.Far(0x1b1, 0), 65536), 0644)
				}
			}
		case "quit":
		default:
			err = fmt.Errorf("第 %d 行：探針不支援 %q", lineNo+1, f[0])
		}
		if err != nil {
			return err
		}
	}
	for _, glyph := range active {
		glyph.End = g.o.Steps()
		if err := encode(glyphs, glyph); err != nil {
			return err
		}
	}
	if err := rows.Flush(); err != nil {
		return err
	}
	if err := glyphs.Flush(); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "探針略過色號0字形（陰影與看不見的標記）：%d\n", blackGlyphs)
	if census {
		b, err := json.MarshalIndent(writers, "", " ")
		if err != nil {
			return err
		}
		if err = os.WriteFile(out+".writers.json", append(b, '\n'), 0644); err != nil {
			return err
		}
	}
	return nil
}

func paletteHash(p [256][3]uint8) string {
	b := make([]byte, 0, 768)
	for _, c := range p {
		b = append(b, c[:]...)
	}
	return fmt.Sprintf("%x", sha256.Sum256(b))[:12]
}

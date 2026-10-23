package overlay

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"image"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

const storyFileStart, storyFileEnd = 0x1a6a, 0x3104

// storyProfile 是一支示範程式（片頭 opdemo 或結局 enddemo）的故事來源與原版掛點。兩支程式用同一套
// 文字直譯器（列高10、F7/F3/F2/F1換列、FE清框、FA/FB/F9字色），只是位址不同。規格014、067。
type storyProfile struct {
	name, sar, sarSHA, keyPrefix string
	// module 是載在 01B1:6000 的模組在 SAR 檔內的起點；fileStart..fileEnd 是故事文字（含結尾FF）。
	module, fileStart, fileEnd int
	entries, chars, clears     int
	// startMarker 非0時，故事以這個字元的起始字形開頭（片頭的 P）。
	startMarker byte
	// firstPage 是第一段文字所在頁；結局第一頁前沒有 FE 清框，開始追蹤即可顯示。
	firstPage    int
	readyAtStart bool
	// segmentEnd 是結局每段結束返回主程式的 FD；片頭的 FD 不分段。
	segmentEnd                                           bool
	readByte, glyphBegin, glyphEnd, clearBegin, clearEnd uint16
	pointer, shadowVar, frontVar                         int
	guards                                               []storyGuard
}
type storyGuard struct {
	lo, hi int
	sha    string
}

// 片頭與結局共用的常駐字形常式及字型。
var storyResidentGuards = []storyGuard{{0x27e9, 0x2857, "e2e41fc888266454bb405a02919fee48682538bb741964383da854a83b559fca"}, {0xf506, 0xf806, "8ff0729de496f9d84d4cb3e4931a349a028981232507939f23af95a821d955d6"}}

var introStoryProfile = storyProfile{
	name: "片頭", sar: "ZELRES1.SAR", sarSHA: "56a73b59e952dab47fb0171a3fb1cdb7d641baceafa97e2826057f877a590eb0", keyPrefix: "zelres1.file",
	module: 0xa4, fileStart: storyFileStart, fileEnd: storyFileEnd, entries: 58, chars: 5082, clears: 52, startMarker: 'P',
	readByte: 0x6a8b, glyphBegin: 0x6ade, glyphEnd: 0x6aef, clearBegin: 0x6bfd, clearEnd: 0x6c0a, pointer: 0x6d56, shadowVar: 0x6d5b, frontVar: 0x6d5c,
	guards: append([]storyGuard{{0x6000, 0x6040, "0dc4cbac83b93284bdb4c90cb875a07f768bfd45e833dfd135a386f610a09e3e"}, {0x6a75, 0x6cc4, "e1c479c686b20282701e9928caf016841931a5876250e0fc4dd5659492563aa5"}, {0x947d, 0x953d, "15eda0237134c09a79b7a068929c292ef9161f96c9f46d3b79b5555b5bfe9cf9"}}, storyResidentGuards...),
}

// 結局 enddemo.bin（ZELRES2 第50項）由城堡小屋 OMOYPRO 在 01B1:0049 非零時載入 6000，FF77 設為 FF。
// 讀字 6329 lods 後 632E、畫字陰影 6381 至本色返回 6392、清框 64C0 至 64C5；指標 6630、字色 6635/6636。
var endingStoryProfile = storyProfile{
	name: "結局", sar: "ZELRES2.SAR", sarSHA: "bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1", keyPrefix: "zelres2.file",
	module: 0x4457c, fileStart: 0x45024, fileEnd: 0x459b3, entries: endingStoryEntries, chars: endingStoryChars, clears: endingStoryClears, firstPage: 1, readyAtStart: true, segmentEnd: true,
	readByte: 0x632e, glyphBegin: 0x6381, glyphEnd: 0x6392, clearBegin: 0x64c0, clearEnd: 0x64c5, pointer: 0x6630, shadowVar: 0x6635, frontVar: 0x6636,
	guards: append([]storyGuard{{0x6000, 0x6040, "ce1dfb1164e32867b51ccddd052d6dcd7128b0a67bd368a5b127cc23da7ec25d"}, {0x6318, 0x6630, "2e3f5bdea222c0afcbac52bcd43ff9e858f2bf8a2abf7ed0a8524c59669abc8b"}}, storyResidentGuards...),
}

// 結局故事的片段、字元與清框數，由 parseStoryProfile 對固定 SAR2 實算後固定。
const endingStoryEntries, endingStoryChars, endingStoryClears = 27, 1951, 19

// memory 回傳故事文字在 01B1 段內的位址範圍。
func (p *storyProfile) memory() (int, int) {
	return 0x6000 + p.fileStart - p.module, 0x6000 + p.fileEnd - p.module
}

// file 由原版指標（已讀過的下一個位址）換回 SAR 檔案偏移。
func (p *storyProfile) file(ptr uint16) int { return int(ptr) - 1 - 0x6000 + p.module }

type storyEntry struct {
	key       string
	files     []int
	page, row int
	style     byte
	newLine   bool
	text      []rune
	positions []storyPosition
}
type storyPosition struct {
	character rune
	ordinal   int
	at        image.Point
}
type storyChar struct {
	entry, index int
	value        byte
}

// parseStory只把固定原文分組供顯示。控制碼仍由原版執行。
func parseStory(data []byte) ([]storyEntry, map[int]storyChar, map[int]int, error) {
	return parseStoryProfile(data, &introStoryProfile)
}

func parseStoryProfile(data []byte, prof *storyProfile) ([]storyEntry, map[int]storyChar, map[int]int, error) {
	if len(data) < prof.fileEnd {
		return nil, nil, nil, fmt.Errorf("故事來源長度不足")
	}
	var entries []storyEntry
	chars := map[int]storyChar{}
	clears := map[int]int{}
	page, row := prof.firstPage, -1
	style := byte(0)
	newLine := false
	var files []int
	flush := func() error {
		if len(files) == 0 {
			return nil
		}
		if prof.startMarker != 0 && files[0] == prof.fileStart {
			if len(files) != 1 || data[files[0]] != prof.startMarker {
				return fmt.Errorf("故事起始來源不符")
			}
			files = nil
			newLine = false
			return nil
		}
		if page == 0 || row < 0 || (style != 0xfa && style != 0xfb && style != 0xf9) {
			return fmt.Errorf("故事片段缺少已驗字色或列")
		}
		e := storyEntry{key: fmt.Sprintf("%s.%06x", prof.keyPrefix, files[0]), files: append([]int(nil), files...), page: page, row: row, style: style, newLine: newLine}
		id := len(entries)
		for i, file := range files {
			chars[file] = storyChar{id, i, data[file]}
		}
		entries = append(entries, e)
		files = nil
		newLine = false
		return nil
	}
	for file := prof.fileStart; file < prof.fileEnd; file++ {
		b := data[file]
		switch b {
		case 0xfe, 0xfa, 0xfb, 0xf9, 0xf7, 0xf3, 0xf2, 0xf1, 0xff, 0xfd:
			if b == 0xfd && !prof.segmentEnd {
				continue
			}
			if err := flush(); err != nil {
				return nil, nil, nil, err
			}
			switch b {
			case 0xfe:
				page++
				row = 0
				newLine = true
				clears[file] = page
			case 0xfa, 0xfb, 0xf9:
				style = b
			case 0xf7:
				row = 0
				newLine = true
			case 0xf3:
				row = 1
				newLine = true
			case 0xf2:
				row = 2
				newLine = true
			case 0xf1:
				row = 3
				newLine = true
			case 0xff:
				if file != prof.fileEnd-1 {
					return nil, nil, nil, fmt.Errorf("故事結尾位置不符")
				}
			}
		default:
			// 其餘 80h 以上是等待、打字音效與人物表情小圖，不影響文字位置。
			if b >= 32 && b < 127 {
				files = append(files, file)
			}
		}
	}
	if len(entries) != prof.entries || len(chars) != prof.chars || len(clears) != prof.clears {
		return nil, nil, nil, fmt.Errorf("%s故事來源數量不符：%d片段、%d字、%d清框", prof.name, len(entries), len(chars), len(clears))
	}
	return entries, chars, clears, nil
}

func storyTokens(text []rune) [][]int {
	open := "（《「『【("
	close := "、。，：；！？）》」』】…,.!?;:)"
	latin := func(ch rune) bool {
		return ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' || ch == '\'' || ch == '-'
	}
	var out [][]int
	for i := 0; i < len(text); {
		a := i
		for i < len(text) && strings.ContainsRune(open, text[i]) {
			i++
		}
		if i < len(text) {
			if latin(text[i]) {
				for i < len(text) && latin(text[i]) {
					i++
				}
			} else {
				i++
			}
		}
		for i < len(text) && strings.ContainsRune(close, text[i]) {
			i++
		}
		var token []int
		for j := a; j < i; j++ {
			token = append(token, j)
		}
		out = append(out, token)
	}
	return out
}

func loadStoryText(entries []storyEntry, data []byte, font bitmap.Font) error {
	if !utf8.Valid(data) {
		return fmt.Errorf("故事譯表須為UTF-8")
	}
	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = '\t'
	r.FieldsPerRecord = 2
	rows, err := r.ReadAll()
	if err != nil {
		return err
	}
	if len(rows) != len(entries)+1 || rows[0][0] != "key" || rows[0][1] != "zh-TW" {
		return fmt.Errorf("故事須含%d片段及key／zh-TW欄位", len(entries))
	}
	x, row, page, lineUsed := 12, 0, 0, false
	for i := range entries {
		e := &entries[i]
		record := rows[i+1]
		if record[0] != e.key {
			return fmt.Errorf("故事鍵或順序不符：%s", record[0])
		}
		if strings.TrimSpace(record[1]) == "" {
			return fmt.Errorf("故事譯文不可空白")
		}
		e.text = []rune(record[1])
		for _, ch := range e.text {
			if unicode.IsControl(ch) || (len(font[ch]) != 16 && len(font[ch]) != 32) {
				return fmt.Errorf("故事缺字或控制碼：%U", ch)
			}
		}
		if e.page != page {
			page = e.page
			x = 12
			row = 0
			lineUsed = false
		}
		if e.newLine {
			if lineUsed {
				row++
			}
			row = max(row, e.row)
			x = 12
			lineUsed = false
		}
		for _, token := range storyTokens(e.text) {
			w := 0
			for _, j := range token {
				w += len(font[e.text[j]]) / 16 * 12
			}
			if w > 936 {
				return fmt.Errorf("故事詞組超出安全寬度：%s", e.key)
			}
			if x+w > 948 {
				row++
				x = 12
				lineUsed = false
			}
			for _, j := range token {
				ch := e.text[j]
				if ch == ' ' && !lineUsed {
					continue
				}
				cw := len(font[ch]) / 16 * 12
				y := 429 + row*30
				if row > 4 || x+cw+3 > 951 || y+27 > 600 {
					return fmt.Errorf("故事超出安全區：%s", e.key)
				}
				e.positions = append(e.positions, storyPosition{ch, j, image.Pt(x, y)})
				x += cw
				lineUsed = true
			}
		}
	}
	return nil
}

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
	if len(data) < storyFileEnd {
		return nil, nil, nil, fmt.Errorf("故事來源長度不足")
	}
	var entries []storyEntry
	chars := map[int]storyChar{}
	clears := map[int]int{}
	page, row := 0, -1
	style := byte(0)
	newLine := false
	var files []int
	flush := func() error {
		if len(files) == 0 {
			return nil
		}
		if files[0] == storyFileStart {
			if len(files) != 1 || data[files[0]] != 'P' {
				return fmt.Errorf("故事起始來源不符")
			}
			files = nil
			newLine = false
			return nil
		}
		if page == 0 || row < 0 || (style != 0xfa && style != 0xfb && style != 0xf9) {
			return fmt.Errorf("故事片段缺少已驗字色或列")
		}
		e := storyEntry{key: fmt.Sprintf("zelres1.file.%06x", files[0]), files: append([]int(nil), files...), page: page, row: row, style: style, newLine: newLine}
		id := len(entries)
		for i, file := range files {
			chars[file] = storyChar{id, i, data[file]}
		}
		entries = append(entries, e)
		files = nil
		newLine = false
		return nil
	}
	for file := storyFileStart; file < storyFileEnd; file++ {
		b := data[file]
		switch b {
		case 0xfe, 0xfa, 0xfb, 0xf9, 0xf7, 0xf3, 0xf2, 0xf1, 0xff:
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
				if file != storyFileEnd-1 {
					return nil, nil, nil, fmt.Errorf("故事結尾位置不符")
				}
			}
		default:
			if b >= 32 && b < 127 {
				files = append(files, file)
			}
		}
	}
	if len(entries) != 58 || len(chars) != 5082 || len(clears) != 52 {
		return nil, nil, nil, fmt.Errorf("故事來源數量不符")
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
	if len(rows) != 59 || rows[0][0] != "key" || rows[0][1] != "zh-TW" {
		return fmt.Errorf("故事須含58片段及key／zh-TW欄位")
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

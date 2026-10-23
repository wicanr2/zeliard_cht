package overlay

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

// 規格066補譯的靜態閘門：完整譯表以實際字模載入，逐段核對來源切法，
// 並以固定 SAR 的模組 bytes 模擬原版逐字呼叫 711B，確認每段都能從第一字走到全文。
// 這裡不跑遊戲，不能代替正常路徑；後段長老、旅館等只靠此測試與來源守衛。
func loadCatalogFixture(t *testing.T) (*oracle.Oracle, *Dialogue, []byte, bitmap.Font) {
	t.Helper()
	root, fontPath := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT")
	if root == "" || fontPath == "" {
		t.Skip("須掛載原版與字型")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	o, err := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { o.Close() })
	d, err := NewDialogue(o, filepath.Join(root, "ZELRES2.SAR"), "../../../text/dialogue-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "ZELRES2.SAR"))
	if err != nil {
		t.Fatal(err)
	}
	return o, d, data, font
}

func TestDialogueCatalogLayout(t *testing.T) {
	_, d, _, font := loadCatalogFixture(t)
	for _, p := range d.paragraphs {
		positions, err := dialogueLayout(font, p.text)
		if err != nil {
			t.Fatal(p.key, err)
		}
		last := positions[len(positions)-1]
		t.Logf("%s 行%d 末端x=%d 字數%d", p.key, last.Y/36+1, last.X, len(p.text))
	}
}

func TestDialogueSourcesStructure(t *testing.T) {
	o, d, data, _ := loadCatalogFixture(t)
	byKey := map[string]int{}
	for i, p := range d.paragraphs {
		byKey[p.key] = i
	}
	// 前36段為1.0.5已以正常路徑核對的來源，其中有刻意相鄰切分的片段，只檢查逐字對應。
	for i, source := range dialogueSources {
		strict := i >= 36
		key := fmt.Sprintf("zelres2.file.%06x", source[0])
		idx, ok := byKey[key]
		if !ok {
			continue
		}
		start, end, base := source[0], source[1], source[2]
		if c := data[start]; strict && (c < 0x20 || c >= 0x7f) {
			t.Fatalf("%s 起點不是可印字元：%02x", key, c)
		}
		if c := data[end]; strict && c >= 0x20 && c < 0x7f && c != '/' && data[end-2] != 0xff {
			t.Fatalf("%s 訖點切在文字中間：%02x", key, c)
		}
		for i := start; i < end; i++ {
			if data[i] == 0xff {
				i++
				continue
			}
			if strict && data[i] == 0x0c {
				t.Fatalf("%s 段內含清框碼", key)
			}
		}
		p := d.paragraphs[idx]
		if p.total == 0 || len(p.text) == 0 {
			t.Fatalf("%s 沒有可對應的字", key)
		}
		// 送返與教堂鏈的後續字段須接在前段之後，已由既有正常路徑配對測試涵蓋。
		if _, alone := d.retainedDialogueLead(idx); !alone {
			continue
		}
		// 模組載入 01B1:A000，711B 為原版呼叫字形的指令。
		o.SetBytes(oracle.Far(0x01b1, 0xa000), data[base:base+0x2000])
		o.SetBytes(oracle.Far(0x01b1, 0x711b), []byte{0x2e, 0xff, 0x16, 0x22, 0x20})
		d.reset()
		visible := 0
		for pos := int(p.start); pos < int(p.end); pos++ {
			n, counted := p.progress[uint16(pos)]
			if !counted {
				continue
			}
			ch := data[base+pos-0xa000]
			o.SetWord(oracle.Far(0x01b1, 0xff4c), uint16(pos+1))
			o.SetRegs(oracle.CallRegs{AX: 0x0100 | uint16(ch), BX: 60, CX: 100, DS: 0x01b1, SetAX: true, SetBX: true, SetCX: true, SetDS: true})
			d.beginGlyph(o)
			if !d.active || d.current != idx {
				t.Fatalf("%s 第%d字未被接受", key, n)
			}
			if d.visible < visible {
				t.Fatalf("%s 顯示字數倒退", key)
			}
			visible = d.visible
		}
		if visible != len(p.text) {
			t.Fatalf("%s 末字時顯示%d／%d", key, visible, len(p.text))
		}
		// 來源被改寫一個 byte 時整段拒收，不猜補。
		o.SetByte(oracle.Far(0x01b1, p.start), data[start]^0x20)
		d.reset()
		o.SetWord(oracle.Far(0x01b1, 0xff4c), p.start+1)
		o.SetRegs(oracle.CallRegs{AX: 0x0100 | uint16(data[start]^0x20), SetAX: true})
		d.beginGlyph(o)
		if d.active {
			t.Fatalf("%s 改寫來源仍被接受", key)
		}
	}
}

// TestItemExplanationCatalogLayout 核對八件商品「您想了解{item}」加兩段說明的完整版面。
// 說明在原版第二頁仍保留前文，三段合計須放進三行安全區，否則已認領的字形會留白。
func TestItemExplanationCatalogLayout(t *testing.T) {
	o, d, _, font := loadCatalogFixture(t)
	l, err := NewLabels(o, filepath.Join(os.Getenv("ZELIARD_ORIG"), "ZELRES2.SAR"), "../../../text/labels-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	d.UseLabels(l)
	q := d.itemExplanation
	if q == nil {
		t.Fatal("缺商品說明鏈")
	}
	for _, item := range itemDetails {
		var source *labelSource
		for _, s := range l.sources {
			if l.isItemListSource(s) && uint16(0xa000+s.offset) == item.name {
				source = s
			}
		}
		if source == nil {
			t.Fatalf("%04x 缺商品名稱譯文", item.name)
		}
		detail, ok := q.details[item.name]
		if !ok || detail[0] == "" || detail[1] == "" {
			t.Fatalf("%04x 缺說明譯文", item.name)
		}
		q.item = source
		q.texts[1], q.texts[2] = detail[0], detail[1]
		q.ready = q.terminal
		positions, ok := quoteLayout(q.tokens())
		if !ok {
			t.Fatalf("%s 說明超出三行安全區", string(source.text))
		}
		last := positions[len(positions)-1]
		t.Logf("%s 行%d 末端x=%d", string(source.text), last.Y/36+1, last.X)
	}
}

// TestNumberSentenceStructure 以固定模組 bytes 模擬原版逐字呼叫，確認數字句各段與數字暫存位址相接，
// 並以五位數字核對中文版面。數字暫存位址的來源見 number_chain.go。
func TestNumberSentenceStructure(t *testing.T) {
	o, d, data, _ := loadCatalogFixture(t)
	c := d.numbers
	if c == nil || len(c.sources) != len(numberSentences) {
		t.Fatal("數字句譯文不完整")
	}
	for i, s := range c.sources {
		spec := numberSentences[i]
		o.SetBytes(oracle.Far(0x01b1, 0xa000), data[spec.module:spec.module+0x2000])
		o.SetBytes(oracle.Far(0x01b1, 0x711b), []byte{0x2e, 0xff, 0x16, 0x22, 0x20})
		o.SetBytes(oracle.Far(0x01b1, s.number), []byte("12345\xff"))
		c.reset()
		x := 60
		call := func(pos uint16, y int) {
			o.SetWord(oracle.Far(0x01b1, 0xff4c), pos+1)
			o.SetRegs(oracle.CallRegs{AX: 0x0100 | uint16(o.Byte(oracle.Far(0x01b1, pos))), BX: uint16(x), CX: uint16(y), DS: 0x01b1, SetAX: true, SetBX: true, SetCX: true, SetDS: true})
			c.begin(o)
			x += 8
			if x >= 256 {
				x = 60
			}
		}
		for j, part := range s.parts {
			if j > 0 {
				x = 60
				for k := 0; k < 5; k++ {
					call(s.number+uint16(k), 120)
				}
			}
			for _, pos := range part.positions {
				call(pos, 110)
			}
			if !c.active || c.current != s {
				t.Fatalf("%s 第%d段未被接受", s.key, j)
			}
		}
		if c.part != 2*len(s.parts)-1 || len(c.numbers) != len(s.parts)-1 {
			t.Fatalf("%s 段落未走完：%d", s.key, c.part)
		}
		for k := range c.numbers {
			c.numbers[k].pixels = make([]byte, c.numbers[k].box.Dx()*c.numbers[k].box.Dy())
		}
		c.done = true
		if _, ok := quoteLayout(c.tokens()); !ok {
			t.Fatalf("%s 中文超出三行安全區", s.key)
		}
		// 數字不在固定暫存位址時拒收。
		c.reset()
		for _, pos := range s.parts[0].positions {
			call(pos, 110)
		}
		call(s.number+0x100, 120)
		if c.active {
			t.Fatalf("%s 接受了非暫存位址的數字", s.key)
		}
	}
}

// TestTownCatalogLayout 以實際字模載入居民譯表，任何一條超出原版對話框就失敗。
func TestTownCatalogLayout(t *testing.T) {
	root, fontPath := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT")
	if root == "" || fontPath == "" {
		t.Skip("須掛載原版與字型")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewTownDialogue(nil, filepath.Join(root, "ZELRES2.SAR"), "../../../text/town-prototype.tsv", font); err != nil {
		t.Fatal(err)
	}
}

// TestSignSources 核對兩個洞窟告示的原版字列：Y 加 19h、每列 12、X 為 58h 加位移、每字 8。
func TestSignSources(t *testing.T) {
	o, _, _, font := loadCatalogFixture(t)
	l, err := NewLabels(o, filepath.Join(os.Getenv("ZELIARD_ORIG"), "ZELRES2.SAR"), "../../../text/labels-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]image.Rectangle{
		"zelres3.file.00e001": {image.Rect(132, 26, 196, 34), image.Rect(120, 38, 200, 46), image.Rect(108, 50, 220, 58)},
		"zelres3.file.00ff16": {image.Rect(124, 28, 196, 36), image.Rect(110, 40, 214, 48)},
	}
	for key, rows := range want {
		s := l.sources[key]
		if s == nil || fmt.Sprint(s.signRows) != fmt.Sprint(rows) || s.base != 0xc000 || s.raw[len(s.raw)-1] != 0xff {
			t.Fatalf("%s 告示字列不符：%v", key, s)
		}
	}
}

package overlay

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

func townInputs(t *testing.T) (string, bitmap.Font) {
	t.Helper()
	root, fontPath := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT")
	if root == "" || fontPath == "" {
		t.Skip("須掛載固定原版及字型")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	return root, font
}

func TestTownCatalogAndOriginalGeometry(t *testing.T) {
	root, font := townInputs(t)
	d, err := NewTownDialogue(nil, filepath.Join(root, "ZELRES2.SAR"), "../../../text/town-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	fixture := os.Getenv("ZELIARD_TOWN_SEGMENT")
	if fixture == "" {
		t.Fatal("須掛載正常居民段作字寬表證據")
	}
	seg, err := os.ReadFile(fixture)
	if err != nil || len(seg) != 65536 || hudHash(seg[0x7b82:0x7c42]) != "15eda0237134c09a79b7a068929c292ef9161f96c9f46d3b79b5555b5bfe9cf9" {
		t.Fatal("字寬證據不符", err)
	}
	adv := seg[0x7be2:0x7c42]
	for _, s := range d.sources {
		// 獨立從原文與原版 advance 表算 6609 行數，不使用 runtime 覆繪框。
		rows, width := 0, 0
		for i, ch := range s.raw {
			if ch >= 128 {
				break
			}
			if ch == '/' {
				rows++
				width = 0
				continue
			}
			width += int(adv[ch-32])
			if ch == ' ' {
				next := 0
				for _, c := range s.raw[i+1:] {
					if c >= 128 || c == ' ' || c == '/' {
						break
					}
					if c >= 32 {
						next += int(adv[c-32])
					}
				}
				if width+next >= 168 {
					width = 0
					rows++
				}
			}
		}
		if width != 0 {
			rows++
		}
		if rows > 8 {
			rows = 8
		}
		if _, err = townLayout(font, s.text, image.Rect(0, 0, 168, rows*10-2)); err != nil {
			t.Fatal(s.key, rows, err)
		}
	}
	if len(d.sources) != 113 {
		t.Fatal("純文字來源數不符")
	}
	for _, key := range []string{"zelres2.file.03ba83", "zelres2.file.03f1e2", "zelres2.file.03f291", "zelres2.file.03f294", "zelres2.file.03f346"} {
		for _, s := range d.sources {
			if s.key == key {
				t.Fatal("混合控制碼誤入覆繪", key)
			}
		}
	}
	catalog, err := os.ReadFile("../../../text/town-prototype.tsv")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(catalog), "\n"), "\n")
	cases := map[string][]byte{
		"utf8":      append(append([]byte(nil), catalog...), 255),
		"columns":   []byte("key\tzh-TW\nkey\t字\t多欄\n"),
		"duplicate": []byte(string(catalog) + lines[1] + "\n"),
		"missing":   []byte(strings.Join(append(lines[:1:1], lines[2:]...), "\n") + "\n"),
		"orphan":    []byte(string(catalog) + "unknown\t未知\n"),
		"control":   []byte(strings.Replace(string(catalog), "\n", "\x01\n", 1)),
		"glyph":     []byte(strings.Replace(string(catalog), strings.Split(lines[1], "\t")[1], "\U0010ffff", 1)),
		"overflow":  []byte(strings.Replace(string(catalog), strings.Split(lines[1], "\t")[1], strings.Repeat("字", 500), 1)),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "bad.tsv")
			if err := os.WriteFile(p, data, 0644); err != nil {
				t.Fatal(err)
			}
			if _, err := NewTownDialogue(nil, filepath.Join(root, "ZELRES2.SAR"), p, font); err == nil {
				t.Fatal("損壞catalog未拒絕")
			}
		})
	}
}

func TestTownNormalDialogueReadOnly(t *testing.T) {
	root, font := townInputs(t)
	out, commandsPath, savePath := os.Getenv("ZELIARD_TOWN_OUTPUT"), os.Getenv("ZELIARD_TOWN_COMMANDS"), os.Getenv("ZELIARD_TOWN_SAVE")
	if out == "" || commandsPath == "" || savePath == "" {
		t.Fatal("須掛載正常鍵序、自己建立的存檔與輸出")
	}
	save, err := os.ReadFile(savePath)
	if err != nil || fmtHash(save) != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
		t.Fatal("存檔身份不符", err)
	}
	info, err := os.Stat(savePath)
	if err != nil {
		t.Fatal(err)
	}
	var machines [2]*oracle.Oracle
	for i := range machines {
		o, e := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if e != nil {
			t.Fatal(e)
		}
		machines[i] = o
		defer o.Close()
		o.SetDOSBoxCycles(3000)
		scratch := t.TempDir()
		p := filepath.Join(scratch, "SAVETEST.usr")
		if err = os.WriteFile(p, save, 0644); err != nil {
			t.Fatal(err)
		}
		if err = os.Chtimes(p, info.ModTime(), info.ModTime()); err != nil {
			t.Fatal(err)
		}
		o.SetScratch(scratch)
	}
	d, err := NewTownDialogue(machines[1], filepath.Join(root, "ZELRES2.SAR"), "../../../text/town-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	mutationFrames := installTownMutationProbe(t, d, machines[1])
	commands, err := os.ReadFile(commandsPath)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	var receipts []map[string]any
	run := func(n uint64) {
		for _, o := range machines {
			if err := o.Run(n); err != nil {
				t.Fatal(err)
			}
		}
	}
	number := func(s string) uint64 {
		n, e := strconv.ParseUint(s, 0, 64)
		if e != nil {
			t.Fatal(e)
		}
		return n
	}
	for _, line := range strings.Split(string(commands), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "run":
			run(number(f[1]))
		case "key":
			for _, o := range machines {
				o.KeyDown(uint8(number(f[1])))
			}
			run(number(f[2]))
			for _, o := range machines {
				o.KeyUp(uint8(number(f[1])))
			}
		case "shot":
			a, b := machines[0], machines[1]
			mem := a.Bytes(oracle.Far(0, 0), 1<<20)
			if !bytes.Equal(mem, b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.IP() != b.IP() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) {
				t.Fatal("中文改動原版狀態", f[1])
			}
			canvas := inventoryOriginalCanvas(b)
			before := append([]byte(nil), canvas.Pix...)
			drawn := d.Draw(canvas, b.Indexed(), b.Palette(), image.Point{})
			want := f[1] == "citizen-page" || f[1] == "citizen-last"
			if drawn != want {
				t.Fatal("正常對話覆繪狀態不符", f[1], d.State())
			}
			changed := 0
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := canvas.PixOffset(x, y)
					if bytes.Equal(before[i:i+4], canvas.Pix[i:i+4]) {
						continue
					}
					changed++
					p := image.Pt(x/3, y/3)
					if !p.In(d.safe) || d.prompt && p.In(d.marker) {
						t.Fatal("中文越界或蓋掉標記", f[1], p)
					}
				}
			}
			if want && (changed == 0 || d.current.key != "zelres2.file.039fff") {
				t.Fatal("正常來源沒有中文", f[1], d.State())
			}
			if f[1] == "citizen-page" && !d.prompt || f[1] == "citizen-last" && (d.prompt || d.visible != len(d.current.text)) {
				t.Fatal("原版續頁狀態不符", f[1])
			}
			if !bytes.Equal(mem, b.Bytes(oracle.Far(0, 0), 1<<20)) {
				t.Fatal("Draw改動原版")
			}
			fp, e := os.Create(filepath.Join(out, f[1]+".png"))
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(fp, canvas)
			fp.Close()
			if e != nil {
				t.Fatal(e)
			}
			receipts = append(receipts, map[string]any{"name": f[1], "state_equal": true, "memory_sha256": fmtHash(mem), "indexed_sha256": fmtHash(b.Indexed()), "steps": b.Steps(), "cycles": b.Cycles(), "drawn": drawn, "changed": changed, "outside_changed": 0, "town": d.State(), "held_keys": b.HeldKeys(), "fileops": b.FileOps()})
			if f[1] == "citizen-page" {
				assertTownFallback(t, d, b)
			}
		default:
			t.Fatal("未知正常鍵序", line)
		}
	}
	if mutationFrames["scroll-substep"] < 10 || mutationFrames["scroll-substep"]%10 != 0 {
		t.Fatal("未走完原版十步捲動", mutationFrames)
	}
	if len(receipts) != 9 || len(machines[1].HeldKeys()) != 0 || d.current != nil || d.Glyphs == 0 || d.Rejected != 0 {
		t.Fatal("居民路徑未完整完成", len(receipts), d.State())
	}
	data, err := json.MarshalIndent(map[string]any{"mutation_frames": mutationFrames, "commands_sha256": fmtHash(commands), "checkpoints": receipts, "scope": "一筆居民正常冷讀檔、續頁、結束；其餘112筆不計正常驗收"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
}

func assertTownFallback(t *testing.T, d *TownDialogue, o *oracle.Oracle) {
	t.Helper()
	mem := o.Bytes(oracle.Far(0x1b1, 0), 65536)
	defer o.SetBytes(oracle.Far(0x1b1, 0), mem)
	for _, bad := range []string{"pixel", "short", "palette", "source", "prefix", "code", "metrics", "table"} {
		copyD := *d
		raw, pal := append([]byte(nil), o.Indexed()...), o.Palette()
		o.SetBytes(oracle.Far(0x1b1, 0), mem)
		switch bad {
		case "pixel":
			raw[d.box.Min.Y*320+d.box.Min.X] ^= 1
		case "short":
			raw = raw[:100]
		case "palette":
			pal[9][0] ^= 1
		case "source":
			o.SetByte(oracle.Far(0x1b1, d.current.start), mem[d.current.start]^1)
		case "prefix":
			o.SetByte(oracle.Far(0x1b1, 0xc000), mem[0xc000]^1)
		case "code":
			o.SetByte(oracle.Far(0x1b1, 0x64b2), mem[0x64b2]^1)
		case "metrics":
			o.SetByte(oracle.Far(0x1b1, 0x7b82), mem[0x7b82]^1)
		case "table":
			o.SetByte(oracle.Far(0x1b1, d.current.tableAddress), mem[d.current.tableAddress]^1)
		}
		canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
		before := append([]byte(nil), canvas.Pix...)
		if copyD.Draw(canvas, raw, pal, image.Point{}) || !bytes.Equal(before, canvas.Pix) {
			t.Fatal("未知來源未原子退回", bad)
		}
	}
}

// 原版6513已做一像素捲動，6516才完成整列。呈現可以落在任意指令間。
func installTownMutationProbe(t *testing.T, d *TownDialogue, o *oracle.Oracle) map[string]int {
	counts := map[string]int{}
	o.OnCall(oracle.Far(0x1b1, 0x6513), func(o *oracle.Oracle) {
		if d.current == nil {
			return
		}
		key := d.current.key
		crop := townCrop(o.Indexed(), d.box)
		different := !bytes.Equal(crop, d.completed)
		counts["scroll-substep"]++
		canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
		drawn := d.Draw(canvas, o.Indexed(), o.Palette(), image.Point{})
		if drawn || d.current == nil || !d.pending {
			t.Fatalf("原版一像素捲動中途呈現失去來源：source=%s steps=%d DS=%04x bitmap_changed=%t drawn=%t state=%v", key, o.Steps(), o.DSReg(), different, drawn, d.State())
		}
	})
	for _, site := range []uint16{0x654c, 0x658a, 0x639e} {
		o.OnCall(oracle.Far(0x1b1, site), func(o *oracle.Oracle) {
			if d.current == nil {
				return
			}
			counts[strconv.Itoa(int(site))]++
			canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
			if d.Draw(canvas, o.Indexed(), o.Palette(), image.Point{}) || d.current == nil || !d.pending {
				t.Fatal("原版寫入前未保留pending來源", site, d.State())
			}
		})
	}
	return counts
}

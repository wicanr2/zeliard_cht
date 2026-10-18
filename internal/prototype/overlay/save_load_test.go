package overlay

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

// 由正常玩家操作取得的命令只驅動鍵盤與執行預算。
// 兩側各自建立存檔、退出、冷開機再讀取，不共享測試狀態或存檔內容。
func TestSaveLoadOverlayNormalPath(t *testing.T) {
	root, fontPath, out := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_SAVE_OUTPUT")
	if root == "" || fontPath == "" || out == "" {
		t.Skip("須掛載原版、字型與存讀檔收據輸出目錄")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	commands, err := os.ReadFile("../../../tools/save-load.commands")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	var originals = map[string]string{}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, e2 := os.ReadFile(filepath.Join(root, e.Name()))
		if e2 != nil {
			t.Fatal(e2)
		}
		originals[e.Name()] = fmtHash(b)
	}
	scratch := [2]string{t.TempDir(), t.TempDir()}
	var machines [2]*oracle.Oracle
	var d *Dialogue
	var l *Labels
	boot := 0
	load := func() {
		for i := range machines {
			if machines[i] != nil {
				machines[i].Close()
			}
			o, e := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
			if e != nil {
				t.Fatal(e)
			}
			o.SetDOSBoxCycles(3000)
			o.SetScratch(scratch[i])
			machines[i] = o
		}
		var e error
		d, e = NewDialogue(machines[1], filepath.Join(root, "ZELRES2.SAR"), "../../../text/dialogue-prototype.tsv", font)
		if e != nil {
			t.Fatal(e)
		}
		l, e = NewLabels(machines[1], filepath.Join(root, "ZELRES2.SAR"), "../../../text/labels-prototype.tsv", font)
		if e != nil {
			t.Fatal(e)
		}
		boot++
	}
	load()
	defer func() {
		for _, o := range machines {
			if o != nil {
				o.Close()
			}
		}
	}()
	run := func(n uint64) {
		for _, o := range machines {
			e := o.Run(n)
			var exited *oracle.ExitError
			if e != nil && !errors.As(e, &exited) {
				t.Fatal(e)
			}
		}
	}
	key := func(scan uint8, n uint64) {
		for _, o := range machines {
			o.KeyDown(scan)
		}
		run(n)
		for _, o := range machines {
			o.KeyUp(scan)
		}
	}
	king := func() {
		type event struct {
			s    uint64
			k    uint8
			down bool
		}
		var events []event
		for _, s := range []uint64{12000000, 20000000, 28000000, 36000000, 44000000, 52000000, 86000000, 94000000, 102000000, 110000000, 118000000, 126000000} {
			events = append(events, event{s, 0x39, true}, event{s + 50000, 0x39, false})
		}
		events = append(events, event{60000000, 0x4d, true}, event{62000000, 0x4d, false}, event{64000000, 0x48, true}, event{64250000, 0x48, false})
		sort.SliceStable(events, func(i, j int) bool { return events[i].s < events[j].s })
		for _, e := range events {
			run(e.s - machines[0].Steps())
			for _, o := range machines {
				if e.down {
					o.KeyDown(e.k)
				} else {
					o.KeyUp(e.k)
				}
			}
		}
		run(130000000 - machines[0].Steps())
	}
	number := func(s string) uint64 {
		n, e := strconv.ParseUint(s, 0, 64)
		if e != nil {
			t.Fatal(e)
		}
		return n
	}
	var receipts []map[string]any
	seen := map[string]bool{}
	for _, line := range strings.Split(string(commands), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "king":
			king()
		case "new":
			load()
		case "run":
			run(number(f[1]))
		case "key":
			key(uint8(number(f[1])), number(f[2]))
		case "down", "up":
			for _, o := range machines {
				if f[0] == "down" {
					o.KeyDown(uint8(number(f[1])))
				} else {
					o.KeyUp(uint8(number(f[1])))
				}
			}
		case "quit":
		case "shot":
			a, b := machines[0], machines[1]
			raw := b.Indexed()
			pal := b.Palette()
			mem := b.Bytes(oracle.Far(0, 0), 1<<20)
			if !bytes.Equal(a.Indexed(), raw) || a.Palette() != pal || a.Cycles() != b.Cycles() || a.Steps() != b.Steps() || a.IP() != b.IP() || !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), mem) {
				t.Fatalf("%s：覆繪改寫原版狀態", f[1])
			}
			_, _, rgb := b.ScreenRGB()
			dst := image.NewRGBA(image.Rect(0, 0, 960, 600))
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := (y/3*320 + x/3) * 3
					dst.SetRGBA(x, y, color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255})
				}
			}
			originalPixels := append([]byte(nil), dst.Pix...)
			drawn := d.Draw(dst, raw, pal, image.Point{})
			labels := l.Draw(dst, raw, pal, image.Point{})
			var labelKeys []string
			var labelBoxes []image.Rectangle
			for _, stamp := range l.stamps {
				labelBoxes = append(labelBoxes, stamp.rows...)
				for key, source := range l.sources {
					if source == stamp.source {
						labelKeys = append(labelKeys, key)
					}
				}
			}
			sort.Strings(labelKeys)
			sourceKey := ""
			if d.current >= 0 {
				sourceKey = d.paragraphs[d.current].key
			}
			if f[1] == "save-confirm-text" && (!drawn || sourceKey != "zelres2.file.01cc66" || d.visible != len(d.paragraphs[d.current].text)) {
				t.Fatalf("存檔成功中文未完整顯示：%v %s", drawn, sourceKey)
			}
			if (f[1] == "name-before-keyup" || f[1] == "load-name" || f[1] == "quit-dialog" || f[1] == "load-list") && labels != 1 {
				t.Fatalf("%s：提示中文未顯示，labels=%d", f[1], labels)
			}
			if f[1] == "save-confirm-text" && labels != 2 {
				t.Fatalf("是／否未顯示，labels=%d", labels)
			}
			for name, want := range map[string]string{"name-before-keyup": "zelres2.file.01c1e2", "load-name": "zelres1.file.00e663", "quit-dialog": "stick.file.00060a", "load-list": "stick.file.000883", "speed-prompt": "stick.file.000745", "speed-digit": "stick.file.000745"} {
				if f[1] == name && (len(labelKeys) != 1 || labelKeys[0] != want) {
					t.Fatalf("%s：介面來源錯誤 %v", name, labelKeys)
				}
			}
			inside, outside := 0, 0
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := dst.PixOffset(x, y)
					if bytes.Equal(originalPixels[i:i+4], dst.Pix[i:i+4]) {
						continue
					}
					p := image.Pt(x/3, y/3)
					allowed := drawn && p.In(image.Rect(54, 98, 264, 150))
					for _, box := range labelBoxes {
						allowed = allowed || p.In(box)
					}
					if allowed {
						inside++
					} else {
						outside++
					}
				}
			}
			if outside != 0 {
				t.Fatalf("%s：中文改寫文字區外 %d 像素", f[1], outside)
			}
			if !bytes.Equal(raw, b.Indexed()) || !bytes.Equal(mem, b.Bytes(oracle.Far(0, 0), 1<<20)) {
				t.Fatal("Draw 改寫原版")
			}
			fp, e := os.Create(filepath.Join(out, f[1]+".png"))
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(fp, dst)
			fp.Close()
			if e != nil {
				t.Fatal(e)
			}
			receipts = append(receipts, map[string]any{"name": f[1], "boot": boot, "steps": b.Steps(), "cycles": b.Cycles(), "raw_sha256": fmtHash(raw), "memory_sha256": fmtHash(mem), "state_equal": true, "dialogue_rendered": drawn, "dialogue_key": sourceKey, "dialogue_glyphs": d.Glyphs, "dialogue_rejections": d.Rejections, "labels": labels, "label_keys": labelKeys, "label_boxes": labelBoxes, "inside_changed": inside, "outside_changed": outside, "wrote": b.Wrote(), "held_keys": b.HeldKeys()})
			seen[f[1]] = true
		default:
			t.Fatalf("未知命令：%s", line)
		}
	}
	for _, name := range []string{"save-confirm-text", "dos-exited", "restored"} {
		if !seen[name] {
			t.Fatalf("缺少正常流程：%s", name)
		}
	}
	var saves [2][]byte
	for i := range scratch {
		saves[i], err = os.ReadFile(filepath.Join(scratch[i], "SAVETEST.usr"))
		if err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(saves[0], saves[1]) || len(saves[0]) != 256 || fmtHash(saves[0]) != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
		t.Fatal("新存檔內容與正常原版收據不同")
	}
	if boot != 2 || len(machines[1].HeldKeys()) != 0 || machines[1].Word(oracle.Far(0x01b1, 0x86)) != 1000 {
		t.Fatal("冷讀檔狀態或持鍵不符")
	}
	readNewSave := false
	for _, op := range machines[1].FileOps() {
		if strings.EqualFold(op.Name, "SAVETEST.USR") && op.Op == "read" && op.Len == 256 && !op.Failed {
			readNewSave = true
		}
	}
	if !readNewSave {
		t.Fatal("沒有成功讀回本次新存檔")
	}
	for name, want := range originals {
		b, e := os.ReadFile(filepath.Join(root, name))
		if e != nil || fmtHash(b) != want {
			t.Fatalf("來源被改寫：%s", name)
		}
	}
	data, _ := json.MarshalIndent(map[string]any{"commands_sha256": fmtHash(commands), "originals_unchanged": true, "save_sha256": fmtHash(saves[0]), "save_bytes": len(saves[0]), "cold_boots": boot, "new_save_read": readNewSave, "checkpoints": receipts}, "", "  ")
	if err = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	fmt.Println("正常新局、中文存檔確認、退出與冷讀檔完整狀態對照通過")
}

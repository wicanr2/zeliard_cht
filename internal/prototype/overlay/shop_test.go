package overlay

import (
	"bytes"
	"encoding/json"
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

func TestWeaponShopPurchaseNormalPathReadOnly(t *testing.T) {
	root, fontPath, out := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_SHOP_OUTPUT")
	if root == "" || fontPath == "" || out == "" {
		t.Skip("須掛載原版、字型與商店收據輸出目錄")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	var machines [2]*oracle.Oracle
	for i := range machines {
		machines[i], err = oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if err != nil {
			t.Fatal(err)
		}
		machines[i].SetDOSBoxCycles(3000)
		machines[i].SetScratch(t.TempDir())
		defer machines[i].Close()
	}
	d, err := NewDialogue(machines[1], filepath.Join(root, "ZELRES2.SAR"), "../../../text/dialogue-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	l, err := NewLabels(machines[1], filepath.Join(root, "ZELRES2.SAR"), "../../../text/labels-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	commands, err := os.ReadFile("../../../tools/weapon-shop.commands")
	if err != nil {
		t.Fatal(err)
	}
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
	type event struct {
		step uint64
		scan uint8
		down bool
	}
	var receipts []map[string]any
	for _, line := range strings.Split(string(commands), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "king":
			var events []event
			for _, s := range []uint64{12000000, 20000000, 28000000, 36000000, 44000000, 52000000, 86000000, 94000000, 102000000, 110000000, 118000000, 126000000} {
				events = append(events, event{s, 0x39, true}, event{s + 50000, 0x39, false})
			}
			events = append(events, event{60000000, 0x4d, true}, event{62000000, 0x4d, false}, event{64000000, 0x48, true}, event{64250000, 0x48, false})
			sort.SliceStable(events, func(i, j int) bool { return events[i].step < events[j].step })
			for _, e := range events {
				run(e.step - machines[0].Steps())
				for _, o := range machines {
					if e.down {
						o.KeyDown(e.scan)
					} else {
						o.KeyUp(e.scan)
					}
				}
			}
			run(130000000 - machines[0].Steps())
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
		case "down", "up":
			for _, o := range machines {
				if f[0] == "down" {
					o.KeyDown(uint8(number(f[1])))
				} else {
					o.KeyUp(uint8(number(f[1])))
				}
			}
		case "shot":
			a, b := machines[0], machines[1]
			raw, pal, mem := b.Indexed(), b.Palette(), b.Bytes(oracle.Far(0, 0), 1<<20)
			if !bytes.Equal(raw, a.Indexed()) || pal != a.Palette() || !bytes.Equal(mem, a.Bytes(oracle.Far(0, 0), 1<<20)) || a.IP() != b.IP() || a.Cycles() != b.Cycles() || a.Steps() != b.Steps() {
				t.Fatalf("%s：中文改寫原版狀態", f[1])
			}
			_, _, rgb := b.ScreenRGB()
			dst := image.NewRGBA(image.Rect(0, 0, 960, 600))
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := (y/3*320 + x/3) * 3
					dst.SetRGBA(x, y, color.RGBA{rgb[i], rgb[i+1], rgb[i+2], 255})
				}
			}
			before := append([]byte(nil), dst.Pix...)
			drawn := d.Draw(dst, raw, pal, image.Point{})
			labels := l.Draw(dst, raw, pal, image.Point{})
			var boxes []image.Rectangle
			var keys []string
			for _, stamp := range l.stamps {
				boxes = append(boxes, stamp.rows...)
				for key, source := range l.sources {
					if source == stamp.source {
						keys = append(keys, key)
					}
				}
			}
			sort.Strings(keys)
			key := ""
			if d.current >= 0 {
				key = d.paragraphs[d.current].key
			}
			if f[1] == "greeting" && (!drawn || key != "zelres2.file.0185ea" || labels != 5) {
				t.Fatalf("初訪中文不完整：%v %s %v", drawn, key, keys)
			}
			if f[1] == "goods" {
				for _, want := range []string{"zelres2.file.018534", "zelres2.file.018543"} {
					found := false
					for _, key := range keys {
						found = found || key == want
					}
					if !found {
						t.Fatal("缺少中文商品", want, keys)
					}
				}
			}
			if f[1] == "purchased" && (!drawn || key != "zelres2.file.018634" || b.Word(oracle.Far(0x01b1, 0x86)) != 800) {
				t.Fatalf("交易後中文或原版金額不符：%v %s %d", drawn, key, b.Word(oracle.Far(0x01b1, 0x86)))
			}
			inside, outside := 0, 0
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := dst.PixOffset(x, y)
					if bytes.Equal(before[i:i+4], dst.Pix[i:i+4]) {
						continue
					}
					p := image.Pt(x/3, y/3)
					allowed := drawn && p.In(dialogueBox)
					for _, box := range boxes {
						allowed = allowed || p.In(box)
					}
					if allowed {
						inside++
					} else {
						outside++
					}
				}
			}
			if outside != 0 || !bytes.Equal(mem, b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(raw, b.Indexed()) {
				t.Fatalf("%s：覆繪越界或改寫原版", f[1])
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
			receipts = append(receipts, map[string]any{"name": f[1], "steps": b.Steps(), "cycles": b.Cycles(), "state_equal": true, "raw_sha256": fmtHash(raw), "memory_sha256": fmtHash(mem), "gold": b.Word(oracle.Far(0x01b1, 0x86)), "dialogue_rendered": drawn, "dialogue_key": key, "label_keys": keys, "inside_changed": inside, "outside_changed": outside, "held_keys": b.HeldKeys()})
		default:
			t.Fatalf("未知命令 %s", line)
		}
	}
	if len(receipts) != 9 || len(machines[1].HeldKeys()) != 0 {
		t.Fatal("正常商店路徑未完整完成")
	}
	data, err := json.MarshalIndent(map[string]any{"commands_sha256": fmtHash(commands), "checkpoints": receipts, "dialogue_rejections": d.Rejections, "scope": "normal new game, reward, purchase, farewell, inventory; dynamic quotation remains original"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	fmt.Println("正常武器店交易中文配對通過，原版金額800；動態報價尚保留原文")
}

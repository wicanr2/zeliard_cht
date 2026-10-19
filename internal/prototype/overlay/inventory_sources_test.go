package overlay

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func TestSharedInventoryCatalogGuards(t *testing.T) {
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
	defer o.Close()
	l, err := NewLabels(o, filepath.Join(root, "ZELRES2.SAR"), "../../../text/labels-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	for key, width := range map[string]int{"zelres2.file.004cb4": 8, "zelres2.file.004bf0": 8, "zelres2.file.004cd9": 8, "zelres2.file.004c20": 8, "zelres2.file.004ac0": 8, "zelres2.file.004ac6": 8, "zelres2.file.004d03": 5, "zelres2.file.004d91": 5} {
		s := l.sources[key]
		if s == nil || s.width != width || s.inventoryWhite != (width == 8) {
			t.Fatal("物品白字與裝備小字分類不符", key)
		}
	}
	for _, bad := range []struct{ name, key, text string }{
		{"interior", "zelres2.file.004cb5", "石"},
		{"empty-second-line", "zelres2.file.004cc0", "石"},
		{"overflow", "zelres2.file.004ac6", "經驗值過長"},
		{"missing-glyph", "zelres2.file.004cb4", string(rune(0x10ffff))},
	} {
		t.Run(bad.name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "catalog.tsv")
			if err := os.WriteFile(p, []byte("key\tzh-TW\n"+bad.key+"\t"+bad.text+"\n"), 0644); err != nil {
				t.Fatal(err)
			}
			if _, err := NewLabels(o, filepath.Join(root, "ZELRES2.SAR"), p, font); err == nil {
				t.Fatal("不相容來源或譯文未拒絕")
			}
		})
	}
}

func TestMagiaInventoryNormalPathReadOnly(t *testing.T) {
	root, fontPath, savePath, out := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_ITEM_SAVE"), os.Getenv("ZELIARD_SHARED_OUTPUT")
	if root == "" || fontPath == "" || savePath == "" || out == "" {
		t.Skip("須掛載原版、字型、自建存檔及收據目錄")
	}
	if err := os.Mkdir(out, 0755); err != nil {
		t.Fatal(err)
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	save, err := os.ReadFile(savePath)
	if err != nil || fmtHash(save) != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
		t.Fatal("自建存檔身份不符", err)
	}
	info, err := os.Stat(savePath)
	if err != nil {
		t.Fatal(err)
	}
	var machines [2]*oracle.Oracle
	var scratches [2]string
	for i := range machines {
		machines[i], err = oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if err != nil {
			t.Fatal(err)
		}
		defer machines[i].Close()
		machines[i].SetDOSBoxCycles(3000)
		machines[i].TraceFiles()
		scratches[i] = filepath.Join(out, fmt.Sprintf("scratch-%d", i))
		if err = os.Mkdir(scratches[i], 0755); err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(scratches[i], "SAVETEST.usr")
		if err = os.WriteFile(p, save, 0644); err != nil {
			t.Fatal(err)
		}
		if err = os.Chtimes(p, info.ModTime(), info.ModTime()); err != nil {
			t.Fatal(err)
		}
		machines[i].SetScratch(scratches[i])
	}
	a, b := machines[0], machines[1]
	l, err := NewLabels(b, filepath.Join(root, "ZELRES2.SAR"), "../../../text/labels-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	commands, err := os.ReadFile("../../../tools/item-shop-magia.commands")
	if err != nil {
		t.Fatal(err)
	}
	presentations := 0
	run := func(n uint64) {
		for n > 0 {
			step := min(n, uint64(50000))
			for _, o := range machines {
				if err := o.Run(step); err != nil {
					t.Fatal(err)
				}
			}
			l.ObserveFrame(b.Indexed(), b.Palette())
			l.Draw(inventoryOriginalCanvas(b), b.Indexed(), b.Palette(), image.Point{})
			presentations++
			n -= step
		}
	}
	number := func(s string) uint64 {
		v, e := strconv.ParseUint(s, 0, 64)
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	var checkpoints []map[string]any
	var fallbacks int
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
			name := f[1]
			raw, pal, mem := b.Indexed(), b.Palette(), b.Bytes(oracle.Far(0, 0), 1<<20)
			if !bytes.Equal(mem, a.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(raw, a.Indexed()) || pal != a.Palette() || a.Regs() != b.Regs() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) || !reflect.DeepEqual(a.FileOps(), b.FileOps()) {
				t.Fatal("共用物品中文改寫原版狀態", name)
			}
			dst := inventoryOriginalCanvas(b)
			before := append([]byte(nil), dst.Pix...)
			l.ObserveFrame(raw, pal)
			drawn := l.Draw(dst, raw, pal, image.Point{})
			var stamps []map[string]any
			keys := map[string]int{}
			for _, s := range l.stamps {
				key := ""
				for k, v := range l.sources {
					if v == s.source {
						key = k
						break
					}
				}
				keys[key]++
				masks, known := l.paint.masks(s.paintSerial, raw)
				if !known {
					t.Fatal("未知繪圖歷史")
				}
				stamps = append(stamps, map[string]any{"key": key, "rows": s.rows, "box": s.box, "foreground": s.foreground, "shadow": s.shadow, "no_use_icon": s.noUseIcon, "masks": masks})
			}
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := dst.PixOffset(x, y)
					if bytes.Equal(before[i:i+4], dst.Pix[i:i+4]) {
						continue
					}
					allowed := false
					for _, s := range l.stamps {
						for _, row := range s.rows {
							allowed = allowed || image.Pt(x/3, y/3).In(row)
						}
					}
					if !allowed {
						t.Fatal("中文寫到來源區外", name, x, y)
					}
				}
			}
			wants := map[string]string{"cave-shared-selected": "zelres2.file.004cb4", "cave-shared-used": "zelres2.file.004bf0", "cave-inventory-empty": "zelres2.file.004ab0"}
			if key, ok := wants[name]; ok {
				if keys[key] == 0 {
					t.Fatal("共用物品來源沒有完成中文", name, key, keys)
				}
				if name == "cave-shared-selected" && (l.sources[key].width != 8 || !l.sources[key].inventoryWhite) {
					t.Fatal("Magia Stone錯走小字入口")
				}
			}
			if name == "cave-shared-used" {
				if keys["zelres2.file.004aca"] != 1 {
					t.Fatal("使用訊息缺少前句")
				}
				for _, s := range l.stamps {
					if s.source != l.sources["zelres2.file.004bf0"] {
						continue
					}
					if s.foreground != 9 || s.shadow != 0 || s.box != image.Rect(72, 86, 256, 94) {
						t.Fatal("使用句子幾何或顏色不符")
					}
					for _, bad := range []string{"pixel", "palette"} {
						copyLabels := *l
						copyLabels.stamps = []labelStamp{s}
						unknown := append([]byte(nil), raw...)
						badPal := pal
						if bad == "pixel" {
							unknown[86*320+72] ^= 128
						} else {
							badPal[9][0] ^= 1
						}
						canvas := image.NewRGBA(dst.Bounds())
						old := append([]byte(nil), canvas.Pix...)
						if copyLabels.Draw(canvas, unknown, badPal, image.Point{}) != 0 || !bytes.Equal(old, canvas.Pix) {
							t.Fatal("失配未原子退回", bad)
						}
						fallbacks++
					}
				}
			}
			if name == "cave-inventory-empty" && keys["zelres2.file.004bf0"] != 0 {
				t.Fatal("使用訊息殘留")
			}
			if strings.Contains(name, "closed") || name == "use-return-town" {
				if !bytes.Equal(dst.Pix, before) {
					t.Fatal("關閉後殘留物品中文", name, keys)
				}
			}
			if name == "item-inventory" || name == "cave-shared-selected" {
				if !bytes.Equal(b.Bytes(oracle.Far(0x1b1, 0xa6), 5), []byte{5, 0, 0, 0, 0}) {
					t.Fatal("Magia未正常取得")
				}
			}
			if name == "cave-shared-used" || name == "cave-inventory-empty" {
				if !bytes.Equal(b.Bytes(oracle.Far(0x1b1, 0xa6), 5), make([]byte, 5)) {
					t.Fatal("使用後未由原版移除")
				}
			}
			if name == "use-return-town" && b.Word(oracle.Far(0x1b1, 0x86)) != 0 {
				t.Fatal("原版金額不符")
			}
			if !bytes.Equal(mem, b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(raw, b.Indexed()) || pal != b.Palette() {
				t.Fatal("繪圖改寫原版")
			}
			fp, e := os.Create(filepath.Join(out, name+".png"))
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(fp, dst)
			fp.Close()
			if e != nil {
				t.Fatal(e)
			}
			if err = os.WriteFile(filepath.Join(out, name+".indexed"), raw, 0644); err != nil {
				t.Fatal(err)
			}
			var pb []byte
			for _, c := range pal {
				pb = append(pb, c[:]...)
			}
			if err = os.WriteFile(filepath.Join(out, name+".palette"), pb, 0644); err != nil {
				t.Fatal(err)
			}
			checkpoints = append(checkpoints, map[string]any{"name": name, "full_state_equal": true, "steps": b.Steps(), "cycles": b.Cycles(), "raw_sha256": fmtHash(raw), "memory_sha256": fmtHash(mem), "gold": b.Word(oracle.Far(0x1b1, 0x86)), "labels": drawn, "stamps": stamps})
		default:
			t.Fatal("未知命令", line)
		}
	}
	if len(checkpoints) != 49 || fallbacks != 2 || len(b.HeldKeys()) != 0 {
		t.Fatal("正常物品路徑不完整", len(checkpoints), fallbacks)
	}
	for _, scratch := range scratches {
		p := filepath.Join(scratch, "SAVETEST.usr")
		got, e := os.ReadFile(p)
		if e != nil || !bytes.Equal(got, save) {
			t.Fatal("存檔被改寫")
		}
		gotInfo, e := os.Stat(p)
		if e != nil || !gotInfo.ModTime().Equal(info.ModTime()) {
			t.Fatal("存檔mtime改變")
		}
	}
	data, err := json.MarshalIndent(map[string]any{"status": "PASS_MAGIA_INVENTORY_NORMAL_PAIR", "checkpoints": checkpoints, "presentations": presentations, "fallbacks": fallbacks, "commands_sha256": fmtHash(commands), "key_reads_equal": true, "file_ops_equal": true, "save_unchanged": true}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, 10), 0644); err != nil {
		t.Fatal(err)
	}
}

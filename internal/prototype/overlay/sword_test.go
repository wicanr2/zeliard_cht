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

func TestSwordSecondLineSourceRegions(t *testing.T) {
	root, fontPath := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT")
	if root == "" || fontPath == "" {
		t.Skip("須掛載原版與字型，skip不計驗收")
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
	for key, pad := range map[string]int{"004d0c": 5, "004d22": 6, "004d35": 4, "004d48": 4, "004d5f": 7, "004d78": 7} {
		s := l.sources["zelres2.file."+key]
		if s == nil || s.leftPadding != pad || s.width != 5 || s.inventory || len(s.raw) != pad+6 || !bytes.Equal(s.raw[:pad], bytes.Repeat([]byte{' '}, pad)) {
			t.Fatal("刀劍第二行未以字母區分類", key)
		}
	}
	p := filepath.Join(t.TempDir(), "wide.tsv")
	if err = os.WriteFile(p, []byte("key\tzh-TW\nzelres2.file.004d22\t過長刀劍名稱\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err = NewLabels(o, filepath.Join(root, "ZELRES2.SAR"), p, font); err == nil {
		t.Fatal("中文借用原版數字欄寬度")
	}
}

func TestSabreSourceUseNormalPathReadOnly(t *testing.T) {
	root, fontPath, out := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_SABRE_OUTPUT")
	if root == "" || fontPath == "" || out == "" {
		t.Skip("須掛載原版、字型與收據目錄，skip不計驗收")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	savePath := filepath.Join(root, "SAVEM1.USR")
	save, err := os.ReadFile(savePath)
	if err != nil || fmtHash(save) != "03df9abcf62b1712d6b46dd1fc5b23335a4b8711b3dc84875a1498260d37735c" {
		t.Fatal("來源存檔身份", err)
	}
	info, err := os.Stat(savePath)
	if err != nil {
		t.Fatal(err)
	}
	var machines [2]*oracle.Oracle
	var saves [2]string
	for i := range machines {
		machines[i], err = oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if err != nil {
			t.Fatal(err)
		}
		defer machines[i].Close()
		machines[i].SetDOSBoxCycles(3000)
		machines[i].TraceFiles()
		scratch := t.TempDir()
		saves[i] = filepath.Join(scratch, "SAVEM1.USR")
		if err = os.WriteFile(saves[i], save, 0644); err != nil {
			t.Fatal(err)
		}
		if err = os.Chtimes(saves[i], info.ModTime(), info.ModTime()); err != nil {
			t.Fatal(err)
		}
		machines[i].SetScratch(scratch)
	}
	a, b := machines[0], machines[1]
	l, err := NewLabels(b, filepath.Join(root, "ZELRES2.SAR"), "../../../text/labels-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	commands, err := os.ReadFile("../../../tools/sabre-source-use.commands")
	if err != nil {
		t.Fatal(err)
	}
	var points []map[string]any
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
	used := false
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
			if !bytes.Equal(mem, a.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(raw, a.Indexed()) || pal != a.Palette() || a.Regs() != b.Regs() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) {
				t.Fatal("中文改動原版狀態", name)
			}
			canvas := inventoryOriginalCanvas(b)
			before := append([]byte(nil), canvas.Pix...)
			l.ObserveFrame(raw, pal)
			drawn := l.Draw(canvas, raw, pal, image.Point{})
			var stamps []map[string]any
			var target *labelStamp
			keys := map[string]int{}
			for i := range l.stamps {
				s := &l.stamps[i]
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
					t.Fatal("未知清框歷史", name)
				}
				stamps = append(stamps, map[string]any{"key": key, "box": s.box, "rows": s.rows, "foreground": s.foreground, "shadow": s.shadow, "no_use_icon": s.noUseIcon, "masks": masks})
				if key == "zelres2.file.004d22" {
					target = s
				}
			}
			if name == "sabre-used" {
				used = true
				if keys["zelres2.file.004aca"] != 1 || keys["zelres2.file.004c20"] != 1 {
					t.Fatal("使用訊息來源缺失", keys)
				}
			}
			if name == "select-fifth" && keys["zelres2.file.004cd9"] != 1 {
				t.Fatal("刀劍油選取來源缺失", keys)
			}
			wantSlots := []byte{5, 1, 1, 1, 7}
			if used {
				wantSlots[4] = 0
			}
			if !bytes.Equal(b.Bytes(oracle.Far(0x1b1, 0xa6), 5), wantSlots) {
				t.Fatal("原版物品消耗不符", name)
			}
			if name == "equipment" || name == "cave-equipment" || name == "after-use-reopened" {
				if target == nil || target.box != image.Rect(238, 86, 263, 94) {
					t.Fatal("刀劍字母安全區不符", name)
				}
				changed := 0
				for y := 86 * 3; y < 94*3; y++ {
					for x := 204 * 3; x < 263*3; x++ {
						i := canvas.PixOffset(x, y)
						if !bytes.Equal(canvas.Pix[i:i+4], before[i:i+4]) {
							if x < 238*3 {
								t.Fatal("中文改原版數字", name, x, y)
							}
							changed++
						}
					}
				}
				if changed == 0 {
					t.Fatal("刀劍中文未畫出", name)
				}
				for _, bad := range []string{"word", "palette"} {
					copyLabels := *l
					copyLabels.stamps = []labelStamp{*target}
					unknown := append([]byte(nil), raw...)
					badPal := pal
					if bad == "word" {
						unknown[86*320+238] ^= 128
					} else {
						badPal[9][0] ^= 1
					}
					dst := image.NewRGBA(canvas.Bounds())
					old := append([]byte(nil), dst.Pix...)
					if copyLabels.Draw(dst, unknown, badPal, image.Point{}) != 0 || !bytes.Equal(old, dst.Pix) {
						t.Fatal("失配未原子退回", bad)
					}
				}
			}
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					i := canvas.PixOffset(x, y)
					if bytes.Equal(canvas.Pix[i:i+4], before[i:i+4]) {
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
			if !bytes.Equal(mem, b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(raw, b.Indexed()) || pal != b.Palette() {
				t.Fatal("繪圖回寫原版")
			}
			fp, e := os.Create(filepath.Join(out, name+".png"))
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(fp, canvas)
			fp.Close()
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(filepath.Join(out, name+".indexed"), raw, 0644); e != nil {
				t.Fatal(e)
			}
			var pb []byte
			for _, c := range pal {
				pb = append(pb, c[:]...)
			}
			if e = os.WriteFile(filepath.Join(out, name+".palette"), pb, 0644); e != nil {
				t.Fatal(e)
			}
			points = append(points, map[string]any{"name": name, "steps": b.Steps(), "cycles": b.Cycles(), "full_state_equal": true, "raw_sha256": fmtHash(raw), "memory_sha256": fmtHash(mem), "labels": drawn, "stamps": stamps, "used": used, "original_number_pixels_preserved": true})
		default:
			t.Fatal("未知命令", line)
		}
	}
	for _, p := range saves {
		got, e := os.ReadFile(p)
		if e != nil || !bytes.Equal(got, save) {
			t.Fatal("存檔改變", e)
		}
		after, e := os.Stat(p)
		if e != nil || !after.ModTime().Equal(info.ModTime()) {
			t.Fatal("存檔mtime改變", e)
		}
	}
	if len(points) != 15 || !used || len(b.HeldKeys()) != 0 || len(b.Unimplemented()) != 0 || len(b.Wrote()) != 0 {
		t.Fatal("正常使用流程未完成", len(points))
	}
	j, err := json.MarshalIndent(map[string]any{"status": "PASS_SABRE_SOURCE_SAVE_NORMAL_PAIR", "checkpoints": points, "presentations": presentations, "source_acquisition": "unknown", "scope": "來源SAVEM1正常F7、刀劍油選取／消耗／重開及刀劍數字區保留；不是取得或效果公式驗收"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "receipt.json"), append(j, 10), 0644); err != nil {
		t.Fatal(err)
	}
}

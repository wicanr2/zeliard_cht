package overlay

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

func TestShieldSecondLineSourceRegions(t *testing.T) {
	root, fontPath := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT")
	if root == "" || fontPath == "" {
		t.Skip("須掛載原版與字型，skip不計驗收")
	}
	font, e := bitmap.Load(fontPath)
	if e != nil {
		t.Fatal(e)
	}
	o, e := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
	if e != nil {
		t.Fatal(e)
	}
	defer o.Close()
	l, e := NewLabels(o, filepath.Join(root, "ZELRES2.SAR"), "../../../text/labels-prototype.tsv", font)
	if e != nil {
		t.Fatal(e)
	}
	wants := map[string]int{"004d96": 5, "004dad": 6, "004dc0": 5, "004dd2": 5, "004de4": 5, "004df9": 6}
	swords := map[string]bool{"004d0c": true, "004d22": true, "004d35": true, "004d48": true, "004d5f": true, "004d78": true}
	count := 0
	for key, s := range l.sources {
		pad, ok := wants[strings.TrimPrefix(key, "zelres2.file.")]
		if !ok {
			if s.leftPadding != 0 && !swords[strings.TrimPrefix(key, "zelres2.file.")] {
				t.Fatal("無關來源被裁空白", key)
			}
			continue
		}
		count++
		if s.leftPadding != pad || s.width != 5 || s.inventory || len(s.raw) != pad+7 || !bytes.Equal(s.raw[:pad], bytes.Repeat([]byte{' '}, pad)) {
			t.Fatal("六個第二行來源分類不符", key)
		}
	}
	if count != 6 {
		t.Fatal("來源數量不符", count)
	}
	// 此譯文能放進舊含空白矩形，放不進真正字母區，必須在載入時拒絕。
	p := filepath.Join(t.TempDir(), "wide.tsv")
	if e = os.WriteFile(p, []byte("key\tzh-TW\nzelres2.file.004dad\t過長盾牌名稱\n"), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e = NewLabels(o, filepath.Join(root, "ZELRES2.SAR"), p, font); e == nil {
		t.Fatal("文字仍借用原版數字欄寬度")
	}
}

func TestShieldNormalSourceSavePages(t *testing.T) {
	root, fontPath, out := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_SHIELD_OUTPUT")
	if root == "" || fontPath == "" || out == "" {
		t.Skip("須掛載原版、字型與新收據目錄，skip不計驗收")
	}
	font, e := bitmap.Load(fontPath)
	if e != nil {
		t.Fatal(e)
	}
	var checkpoints []map[string]any
	for _, entry := range []struct{ name, sha string }{{"SAVEM3", "e6e37ae9e79c88bf506a731cc4db5c4c4fe16db3febc475760e8c0370a1e421c"}, {"SAVEM", "cd081139852ad3ffcf0acecc7a60db5bc1fdd0a3c9d2c049d78f3e368e6b3ddc"}} {
		t.Run(entry.name, func(t *testing.T) {
			save, e := os.ReadFile(filepath.Join(root, entry.name+".USR"))
			if e != nil || fmtHash(save) != entry.sha {
				t.Fatal("來源存檔身份", e)
			}
			info, e := os.Stat(filepath.Join(root, entry.name+".USR"))
			if e != nil {
				t.Fatal(e)
			}
			var machines [2]*oracle.Oracle
			var scratches [2]string
			for i := range machines {
				machines[i], e = oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
				if e != nil {
					t.Fatal(e)
				}
				defer machines[i].Close()
				machines[i].SetDOSBoxCycles(3000)
				machines[i].TraceFiles()
				scratches[i] = t.TempDir()
				p := filepath.Join(scratches[i], entry.name+".USR")
				if e = os.WriteFile(p, save, 0644); e != nil {
					t.Fatal(e)
				}
				if e = os.Chtimes(p, info.ModTime(), info.ModTime()); e != nil {
					t.Fatal(e)
				}
				machines[i].SetScratch(scratches[i])
			}
			a, b := machines[0], machines[1]
			l, e := NewLabels(b, filepath.Join(root, "ZELRES2.SAR"), "../../../text/labels-prototype.tsv", font)
			if e != nil {
				t.Fatal(e)
			}
			run := func(n uint64) {
				for n > 0 {
					step := min(n, uint64(50000))
					for _, o := range machines {
						if e := o.Run(step); e != nil {
							t.Fatal(e)
						}
					}
					if b.Steps() > 85000000 {
						l.ObserveFrame(b.Indexed(), b.Palette())
						l.Draw(inventoryOriginalCanvas(b), b.Indexed(), b.Palette(), image.Point{})
					}
					n -= step
				}
			}
			key := func(scan uint8, held, release uint64) {
				for _, o := range machines {
					o.KeyDown(scan)
				}
				run(held)
				for _, o := range machines {
					o.KeyUp(scan)
				}
				run(release)
			}
			check := func(phase string, want bool) {
				raw, pal, mem := b.Indexed(), b.Palette(), b.Bytes(oracle.Far(0, 0), 1<<20)
				if !bytes.Equal(mem, a.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(raw, a.Indexed()) || pal != a.Palette() || a.Regs() != b.Regs() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) {
					t.Fatal("中文改動原版狀態", phase)
				}
				canvas := inventoryOriginalCanvas(b)
				before := append([]byte(nil), canvas.Pix...)
				l.ObserveFrame(raw, pal)
				drawn := l.Draw(canvas, raw, pal, image.Point{})
				var stamps []map[string]any
				var target *labelStamp
				for i := range l.stamps {
					s := &l.stamps[i]
					key := ""
					for k, p := range l.sources {
						if p == s.source {
							key = k
							break
						}
					}
					masks, known := l.paint.masks(s.paintSerial, raw)
					if !known {
						t.Fatal("未知原版清框歷史")
					}
					stamps = append(stamps, map[string]any{"key": key, "box": s.box, "rows": s.rows, "foreground": s.foreground, "shadow": s.shadow, "no_use_icon": s.noUseIcon, "masks": masks})
					if key == "zelres2.file.004dad" {
						target = s
					}
				}
				if (target != nil) != want {
					t.Fatal("盾牌文字顯示或撤除錯誤", phase, stamps)
				}
				if want {
					if target.box != image.Rect(238, 105, 268, 113) || len(target.rows) != 1 || target.rows[0] != target.box {
						t.Fatal("未使用原版字母安全區", target.box)
					}
					changed := 0
					for y := 105 * 3; y < 113*3; y++ {
						for x := 204 * 3; x < 268*3; x++ {
							i := canvas.PixOffset(x, y)
							same := bytes.Equal(canvas.Pix[i:i+4], before[i:i+4])
							if x < 238*3 && !same {
								t.Fatal("中文覆蓋原版數字", x, y)
							}
							if x >= 238*3 && !same {
								changed++
							}
						}
					}
					if changed == 0 {
						t.Fatal("缺少實際中文像素")
					}
					for _, bad := range []string{"word", "palette"} {
						copyLabels := *l
						copyLabels.stamps = []labelStamp{*target}
						unknown := append([]byte(nil), raw...)
						badPal := pal
						if bad == "word" {
							unknown[105*320+238] ^= 128
						} else {
							badPal[9][0] ^= 1
						}
						dst := image.NewRGBA(canvas.Bounds())
						old := append([]byte(nil), dst.Pix...)
						if copyLabels.Draw(dst, unknown, badPal, image.Point{}) != 0 || !bytes.Equal(dst.Pix, old) {
							t.Fatal("字母或色盤失配未原子退回", bad)
						}
					}
					// 合成邊界測試，不宣稱另一個原版數值：欄外像素變化不能清掉中文。
					copyLabels := *l
					copyLabels.stamps = []labelStamp{*target}
					unknown := append([]byte(nil), raw...)
					unknown[105*320+208] ^= 128
					dst := image.NewRGBA(canvas.Bounds())
					old := append([]byte(nil), dst.Pix...)
					if copyLabels.Draw(dst, unknown, pal, image.Point{}) != 1 {
						t.Fatal("數字欄仍污染文字守衛")
					}
					for y := 0; y < 600; y++ {
						for x := 0; x < 960; x++ {
							if image.Pt(x/3, y/3).In(target.box) {
								continue
							}
							i := dst.PixOffset(x, y)
							if !bytes.Equal(dst.Pix[i:i+4], old[i:i+4]) {
								t.Fatal("寫出文字區")
							}
						}
					}
				}
				if !bytes.Equal(mem, b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(raw, b.Indexed()) || pal != b.Palette() {
					t.Fatal("繪圖回寫原版")
				}
				name := entry.name + "-" + phase
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
				checkpoints = append(checkpoints, map[string]any{"name": name, "source_save": entry.name, "source_save_sha256": entry.sha, "phase": phase, "full_state_equal": true, "steps": b.Steps(), "cycles": b.Cycles(), "raw_sha256": fmtHash(raw), "memory_sha256": fmtHash(mem), "labels": drawn, "stamps": stamps, "shield_visible": want, "original_number_pixels_preserved": true})
			}
			run(12000000)
			for i := 0; i < 6; i++ {
				key(0x39, 50000, 7950000)
			}
			key(0x41, 500000, 2000000)
			key(0x15, 150000, 1500000)
			for i := 0; i < 8; i++ {
				key(0x0e, 100000, 200000)
			}
			scans := map[rune]uint8{'s': 0x1f, 'a': 0x1e, 'v': 0x2f, 'e': 0x12, 'm': 0x32, '3': 0x04}
			for _, ch := range strings.ToLower(entry.name) {
				key(scans[ch], 100000, 200000)
			}
			key(0x1c, 100000, 20000000)
			check("restored", false)
			key(0x1c, 500000, 2000000)
			check("opened", true)
			key(0x1c, 500000, 2000000)
			check("closed", false)
			key(0x1c, 500000, 2000000)
			check("reopened", true)
			for _, scratch := range scratches {
				p := filepath.Join(scratch, entry.name+".USR")
				got, e := os.ReadFile(p)
				if e != nil || !bytes.Equal(got, save) {
					t.Fatal("來源副本存檔改變", e)
				}
				after, e := os.Stat(p)
				if e != nil || !after.ModTime().Equal(info.ModTime()) {
					t.Fatal("來源副本mtime改變", e)
				}
			}
			if len(b.HeldKeys()) != 0 || len(b.Unimplemented()) != 0 {
				t.Fatal("原版操作或能力不完整")
			}
		})
	}
	if len(checkpoints) != 8 {
		t.Fatal("八個正常檢查點未齊", len(checkpoints))
	}
	b, e := json.MarshalIndent(map[string]any{"status": "PASS_SHIELD_SOURCE_SAVE_NORMAL_PAIR", "checkpoints": checkpoints, "source_acquisition": "來源存檔取得過程未知，未注入或還原機器狀態", "scope": "兩原版冷啟動F7後的開關／重開、完整狀態與數字保留；其他盾牌種類未正常驗證"}, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(b, 10), 0644); e != nil {
		t.Fatal(e)
	}
}

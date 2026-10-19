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

func TestHUDObservedPrefixGuards(t *testing.T) {
	root, fontPath := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT")
	if root == "" || fontPath == "" {
		t.Skip("須掛載原版與字型，skip不計驗收")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	h, err := NewHUD(nil, root, "../../../text/hud-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	s := h.sources["zelres3.file.00c31a"]
	if s == nil || s.prefixHash != "3f5edbc8b51aa017211fb0ca808f807a1ade51d95e14b49dadf593e90ada24eb" {
		t.Fatal("原始前綴身份不符")
	}
	alt := append([]byte(nil), s.prefix...)
	alt[0x13], alt[0x14] = 255, 255
	if fmtHash(alt) != "6c853b5dab1da5682a959dd804b0f23e5acdc3aa67099596335e1293694f03b2" || !hudPrefixMatches(s, alt) {
		t.Fatal("已觀察的完整替代前綴未接受")
	}
	for _, source := range h.sources {
		if !hudPrefixMatches(source, source.prefix) {
			t.Fatal("既有來源不再接受", source.key)
		}
		if source != s && !bytes.Equal(alt, source.prefix) && hudPrefixMatches(source, alt) {
			t.Fatal("替代前綴錯套其他來源", source.key)
		}
	}
	// 合成拒絕條件；不宣稱這些未知值曾在原版出現。
	for _, prefix := range [][]byte{s.prefix, alt} {
		for i := range prefix {
			bad := append([]byte(nil), prefix...)
			bad[i] ^= 1
			if hudPrefixMatches(s, bad) {
				t.Fatal("任意byte變化未拒絕", i)
			}
		}
	}
	if hudPrefixMatches(s, alt[:63]) || hudPrefixMatches(s, append(append([]byte(nil), alt...), 0)) {
		t.Fatal("前綴長度失配未拒絕")
	}
}

func TestCaveHUDObservedPrefixesNormalPaths(t *testing.T) {
	root, fontPath, ownSave, out, sourceDir := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_ITEM_SAVE"), os.Getenv("ZELIARD_HUD_PREFIX_OUTPUT"), os.Getenv("ZELIARD_HUD_PREFIX_SOURCE")
	if root == "" || fontPath == "" || ownSave == "" || out == "" || sourceDir == "" {
		t.Skip("須掛載原版、字型、自建存檔、來源收據與輸出，skip不計驗收")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	var all []map[string]any
	for _, entry := range []struct{ name, path, sha, commands, lifecycle string }{
		{"SAVEM1", filepath.Join(root, "SAVEM1.USR"), "03df9abcf62b1712d6b46dd1fc5b23335a4b8711b3dc84875a1498260d37735c", "../../../tools/sabre-source-use.commands", "../../../tools/cave-hud-source-save.commands"},
		{"SAVETEST", ownSave, "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb", "../../../tools/combat.commands", "../../../tools/cave-hud-prefix.commands"},
	} {
		t.Run(entry.name, func(t *testing.T) {
			dir := filepath.Join(out, entry.name)
			if e := os.Mkdir(dir, 0755); e != nil {
				t.Fatal(e)
			}
			save, e := os.ReadFile(entry.path)
			if e != nil || fmtHash(save) != entry.sha {
				t.Fatal("存檔來源身份不符", e)
			}
			info, e := os.Stat(entry.path)
			if e != nil {
				t.Fatal(e)
			}
			var machines [2]*oracle.Oracle
			var paths [2]string
			for i := range machines {
				machines[i], e = oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
				if e != nil {
					t.Fatal(e)
				}
				defer machines[i].Close()
				machines[i].SetDOSBoxCycles(3000)
				machines[i].TraceFiles()
				scratch := t.TempDir()
				paths[i] = filepath.Join(scratch, entry.name+".USR")
				if e = os.WriteFile(paths[i], save, 0644); e != nil {
					t.Fatal(e)
				}
				if e = os.Chtimes(paths[i], info.ModTime(), info.ModTime()); e != nil {
					t.Fatal(e)
				}
				machines[i].SetScratch(scratch)
			}
			a, b := machines[0], machines[1]
			h, e := NewHUD(b, root, "../../../text/hud-prototype.tsv", font)
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
					h.Draw(inventoryOriginalCanvas(b), b.Indexed(), b.Palette(), image.Point{})
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
			check := func(name string) {
				memory, raw, pal := b.Bytes(oracle.Far(0, 0), 1<<20), b.Indexed(), b.Palette()
				if !bytes.Equal(memory, a.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(raw, a.Indexed()) || pal != a.Palette() || b.Regs() != a.Regs() || b.Steps() != a.Steps() || b.Cycles() != a.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) {
					t.Fatal("HUD改動完整原版狀態", name)
				}
				canvas := inventoryOriginalCanvas(b)
				changed, keys := assertHUDCheckpoint(t, name, h, b, canvas)
				want := ""
				switch name {
				case "cave-entered", "cave-inventory", "cave-closed", "cave-reentered":
					want = "zelres3.file.00c31a"
				case "returned-town", "reentry-not-aligned", "reentry-still-town", "reentry-aligned":
					want = "zelres2.file.0399b8"
				}
				if want != "" && (len(keys) != 5 || !containsHUDKey(keys, want)) {
					t.Fatal("正常路徑未顯示五欄與正確名稱", name, keys)
				}
				if name == "cave-entered" {
					expected, e := os.ReadFile(filepath.Join(sourceDir, entry.name, "point-cave-entered.memory"))
					if e != nil || !bytes.Equal(memory, expected) {
						t.Fatal("與純原版讀取觀察不同", e)
					}
				}
				if want == "zelres3.file.00c31a" {
					isolated := *h
					isolated.fields = map[string]*hudSource{"scene": h.fields["scene"]}
					for _, bad := range []string{"pixels", "palette"} {
						r, p := append([]byte(nil), raw...), pal
						if bad == "pixels" {
							r[175*320+88] ^= 1
						} else {
							p[9][0] ^= 1
						}
						dst := image.NewRGBA(canvas.Bounds())
						before := append([]byte(nil), dst.Pix...)
						if isolated.Draw(dst, r, p, image.Point{}) != 0 || !bytes.Equal(before, dst.Pix) {
							t.Fatal("字形或色盤失配未原子退回", bad)
						}
					}
				}
				if !bytes.Equal(memory, b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(raw, b.Indexed()) || pal != b.Palette() {
					t.Fatal("繪製回寫原版")
				}
				var fields []map[string]any
				for _, key := range keys {
					s := h.sources[key]
					fields = append(fields, map[string]any{"key": key, "box": s.box, "foreground": s.foreground, "shadow": s.shadow})
				}
				f, e := os.Create(filepath.Join(dir, name+".png"))
				if e != nil {
					t.Fatal(e)
				}
				e = png.Encode(f, canvas)
				f.Close()
				if e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(filepath.Join(dir, name+".indexed"), raw, 0644); e != nil {
					t.Fatal(e)
				}
				var pb []byte
				for _, c := range pal {
					pb = append(pb, c[:]...)
				}
				if e = os.WriteFile(filepath.Join(dir, name+".palette"), pb, 0644); e != nil {
					t.Fatal(e)
				}
				all = append(all, map[string]any{"source_save": entry.name, "name": name, "steps": b.Steps(), "cycles": b.Cycles(), "memory_sha256": fmtHash(memory), "indexed_sha256": fmtHash(raw), "full_state_equal": true, "hud_keys": keys, "hud_fields": fields, "changed_pixels": changed, "outside_or_number_changes": 0})
			}
			execute := func(commands []byte, stopAtCave bool) {
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
						check(f[1])
						if stopAtCave && f[1] == "cave-entered" {
							return
						}
					default:
						t.Fatal("未知鍵序", line)
					}
				}
			}
			startup, e := os.ReadFile(entry.commands)
			if e != nil {
				t.Fatal(e)
			}
			execute(startup, true)
			lifecycle, e := os.ReadFile(entry.lifecycle)
			if e != nil {
				t.Fatal(e)
			}
			execute(lifecycle, false)
			for _, p := range paths {
				got, e := os.ReadFile(p)
				if e != nil || !bytes.Equal(got, save) {
					t.Fatal("存檔改變", e)
				}
				after, e := os.Stat(p)
				if e != nil || !after.ModTime().Equal(info.ModTime()) {
					t.Fatal("存檔mtime改變", e)
				}
			}
			if len(b.HeldKeys()) != 0 || len(b.Unimplemented()) != 0 || len(b.Wrote()) != 0 {
				t.Fatal("原版操作未完整結束")
			}
		})
	}
	if t.Failed() {
		return
	}
	if len(all) != 23 {
		t.Fatal("正常檢查點未齊", len(all))
	}
	data, err := json.MarshalIndent(map[string]any{"status": "PASS_CAVE_HUD_OBSERVED_PREFIX_NORMAL_PAIR", "checkpoints": all, "scope": "兩份正常F7、入洞、裝備頁開關、回鎮與重入的唯讀五欄HUD；原始欄位與其餘場景未知"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, 10), 0644); err != nil {
		t.Fatal(err)
	}
}

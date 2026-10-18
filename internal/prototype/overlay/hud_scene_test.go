package overlay

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

func TestHUDSceneNormalPathsReadOnly(t *testing.T) {
	root, fontPath, out, commandDir := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_SCENE_OUTPUT"), os.Getenv("ZELIARD_SCENE_COMMANDS")
	if root == "" || fontPath == "" || out == "" || commandDir == "" {
		t.Skip("須掛載原版、字型與正常場景鍵序")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, route := range []string{"weapon", "sage"} {
		t.Run(route, func(t *testing.T) {
			commands, err := os.ReadFile(filepath.Join(commandDir, route+"-pair.commands"))
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
			hud, err := NewHUD(machines[1], root, "../../../text/hud-prototype.tsv", font)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(out, route)
			if err = os.MkdirAll(path, 0755); err != nil {
				t.Fatal(err)
			}
			seen := map[string]bool{}
			var receipts []map[string]any
			run := func(n uint64) {
				for _, o := range machines {
					if e := o.Run(n); e != nil {
						t.Fatal(e)
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
				case "down":
					for _, o := range machines {
						o.KeyDown(uint8(number(f[1])))
					}
				case "up":
					for _, o := range machines {
						o.KeyUp(uint8(number(f[1])))
					}
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
					memory := a.Bytes(oracle.Far(0, 0), 1<<20)
					if !bytes.Equal(memory, b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.IP() != b.IP() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() {
						t.Fatal("場景覆繪改動原版狀態", route, f[1])
					}
					canvas := inventoryOriginalCanvas(b)
					changed, keys := assertHUDCheckpoint(t, f[1], hud, b, canvas)
					for _, key := range keys {
						seen[key] = true
					}
					want := map[string]string{"king-scene": "zelres2.file.017238", "greeting": "zelres2.file.0184c9", "hut-scene": "zelres2.file.017809", "sage-menu": "zelres2.file.01c997"}[f[1]]
					if want != "" && !containsHUDKey(keys, want) {
						t.Fatal("正常場景名稱中文缺漏", route, f[1], keys)
					}
					if !bytes.Equal(memory, b.Bytes(oracle.Far(0, 0), 1<<20)) {
						t.Fatal("場景繪製寫入原版")
					}
					fp, e := os.Create(filepath.Join(path, f[1]+".png"))
					if e != nil {
						t.Fatal(e)
					}
					e = png.Encode(fp, canvas)
					fp.Close()
					if e != nil {
						t.Fatal(e)
					}
					receipts = append(receipts, map[string]any{"name": f[1], "steps": b.Steps(), "cycles": b.Cycles(), "state_equal": true, "memory_sha256": fmtHash(memory), "indexed_sha256": fmtHash(b.Indexed()), "hud_keys": keys, "hud_pixels_changed": changed, "outside_changed": 0, "gold": b.Word(oracle.Far(0x1b1, 0x86)), "held_keys": b.HeldKeys()})
				default:
					t.Fatal("未知場景鍵序", line)
				}
			}
			expected := 10
			if route == "sage" {
				expected = 3
			}
			if len(receipts) != expected || len(machines[1].HeldKeys()) != 0 {
				t.Fatal("場景路徑未完整結束", route, len(receipts))
			}
			data, e := json.MarshalIndent(map[string]any{"commands_sha256": fmtHash(commands), "checkpoints": receipts, "seen_keys": seen, "scope": "正常國王／武器或長老／小屋路徑；完整1MiB與原版相同，未走過名稱不計正常驗收"}, "", "  ")
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(filepath.Join(path, "receipt.json"), append(data, '\n'), 0644); e != nil {
				t.Fatal(e)
			}
		})
	}
}

func containsHUDKey(keys []string, key string) bool {
	for _, v := range keys {
		if v == key {
			return true
		}
	}
	return false
}

// 合成入口只驗拒絕條件，不冒充正常場景可達證據。
func TestHUDSceneAdmissionRejectsUnknownSource(t *testing.T) {
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
	h, err := NewHUD(nil, root, "../../../text/hud-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	h.begin(o, 0x22cd)
	if h.pending != nil {
		t.Fatal("冷啟動未知入口不應承認來源")
	}
	canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
	before := append([]byte(nil), canvas.Pix...)
	if h.Draw(canvas, o.Indexed(), o.Palette(), image.Point{}) != 0 || !bytes.Equal(before, canvas.Pix) {
		t.Fatal("未完成入口不應顯示候選名稱")
	}
}

func TestHUDSceneAdmissionGuards(t *testing.T) {
	root, fontPath, fixture := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_HUD_FONT_SEGMENT")
	if root == "" || fontPath == "" || fixture == "" {
		t.Skip("須掛載原版、字型及已核對字形段")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	baseline, err := os.ReadFile(fixture)
	if err != nil || len(baseline) != 65536 {
		t.Fatal("字形段不符", err)
	}
	o, err := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
	if err != nil {
		t.Fatal(err)
	}
	defer o.Close()
	o.SetScratch(t.TempDir())
	if err = o.Run(12000000); err != nil {
		t.Fatal(err)
	}
	if o.VideoMode() != 0x13 {
		t.Fatal("未進入原版MCGA模式")
	}
	h, err := NewHUD(nil, root, "../../../text/hud-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, s := range h.sources {
		if s.base != 0 {
			continue
		}
		for _, base := range []uint16{0x8000, 0xa000} {
			header := base + uint16(s.body-4-s.module)
			memory := append([]byte(nil), baseline...)
			copy(memory[int(base):], s.prefix)
			copy(memory[int(header):], s.raw)
			o.SetBytes(oracle.Far(0x1b1, 0), memory)
			o.SetRegs(oracle.CallRegs{DS: 0x1b1, SI: header, SetDS: true, SetSI: true})
			h.begin(o, 0x22cd)
			if h.pending == nil || h.pending.key != s.key || h.pending.base != base || h.pending.header != header {
				t.Fatal("候選未由當次SI唯一解析來源", s.key, base)
			}
			for _, bad := range []string{"prefix", "source", "code", "font", "font-pointer", "ds", "si", "entry", "ambiguous"} {
				o.SetBytes(oracle.Far(0x1b1, 0), memory)
				o.SetRegs(oracle.CallRegs{DS: 0x1b1, SI: header, SetDS: true, SetSI: true})
				entry := uint16(0x22cd)
				switch bad {
				case "prefix":
					o.SetByte(oracle.Far(0x1b1, base), memory[base]^1)
				case "source":
					o.SetByte(oracle.Far(0x1b1, header+4), memory[header+4]^1)
				case "code":
					o.SetByte(oracle.Far(0x1b1, 0x22bf), memory[0x22bf]^1)
				case "font":
					o.SetByte(oracle.Far(0x1b1, 0xf856), memory[0xf856]^1)
				case "font-pointer":
					o.SetWord(oracle.Far(0x1b1, 0xf504), 0xf857)
				case "ds":
					o.SetRegs(oracle.CallRegs{DS: 0x1b2, SetDS: true})
				case "si":
					o.SetRegs(oracle.CallRegs{SI: header + 1, SetSI: true})
				case "entry":
					entry = 0x22bf
				case "ambiguous":
					duplicate := *s
					duplicate.key = "test.ambiguous"
					h.sources[duplicate.key] = &duplicate
				}
				h.begin(o, entry)
				delete(h.sources, "test.ambiguous")
				if h.pending != nil {
					t.Fatal("不明來源未拒絕", s.key, base, bad)
				}
			}
			checked++
		}
	}
	if checked != 88 {
		t.Fatal("44候選的兩種載入基址未驗完整", checked)
	}
}

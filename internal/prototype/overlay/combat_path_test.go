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

// 僅驗正常原版交戰與既有中文消費端；不重作敵人、傷害或掉落規則。
func TestCombatNormalPathReadOnly(t *testing.T) {
	root, fontPath, out, savePath := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_FONT"), os.Getenv("ZELIARD_COMBAT_OUTPUT"), os.Getenv("ZELIARD_COMBAT_SAVE")
	if root == "" || fontPath == "" || out == "" || savePath == "" {
		t.Skip("須掛載原版、字型、自己正常建立的存檔與輸出")
	}
	font, err := bitmap.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	save, err := os.ReadFile(savePath)
	if err != nil || fmtHash(save) != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
		t.Fatal("自建存檔不符", err)
	}
	info, err := os.Stat(savePath)
	if err != nil {
		t.Fatal(err)
	}
	var machines [2]*oracle.Oracle
	for i := range machines {
		o, err := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if err != nil {
			t.Fatal(err)
		}
		machines[i] = o
		defer o.Close()
		o.SetDOSBoxCycles(3000)
		scratch := t.TempDir()
		p := filepath.Join(scratch, "SAVETEST.usr")
		if err := os.WriteFile(p, save, 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, info.ModTime(), info.ModTime()); err != nil {
			t.Fatal(err)
		}
		o.SetScratch(scratch)
	}
	a, b := machines[0], machines[1]
	equal := func(name string) {
		if !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) || a.Regs() != b.Regs() || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) {
			t.Fatal("中文改動完整原版狀態", name)
		}
	}
	equal("before-execution")
	initialMemory := fmtHash(a.Bytes(oracle.Far(0, 0), 1<<20))
	dialogue, err := NewDialogue(b, filepath.Join(root, "ZELRES2.SAR"), "../../../text/dialogue-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	labels, err := NewLabels(b, filepath.Join(root, "ZELRES2.SAR"), "../../../text/labels-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	dialogue.UseLabels(labels)
	hud, err := NewHUD(b, root, "../../../text/hud-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	commands, err := os.ReadFile("../../../tools/combat.commands")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	number := func(s string) uint64 {
		n, err := strconv.ParseUint(s, 0, 64)
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	run := func(n uint64) {
		for _, o := range machines {
			if err := o.Run(n); err != nil {
				t.Fatal(err)
			}
		}
	}
	var points []map[string]any
	caveCount := 0
	for _, line := range strings.Split(string(commands), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
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
		case "down", "up":
			for _, o := range machines {
				if f[0] == "down" {
					o.KeyDown(uint8(number(f[1])))
				} else {
					o.KeyUp(uint8(number(f[1])))
				}
			}
		case "shot":
			equal(f[1])
			memory := b.Bytes(oracle.Far(0, 0), 1<<20)
			raw, pal := b.Indexed(), b.Palette()
			canvas := inventoryOriginalCanvas(b)
			before := append([]byte(nil), canvas.Pix...)
			drawn := dialogue.Draw(canvas, raw, pal, image.Point{})
			labels.ObserveFrame(raw, pal)
			labelCount := labels.Draw(canvas, raw, pal, image.Point{})
			changed, keys := assertHUDCheckpoint(t, f[1], hud, b, canvas)
			cave := fmtHash(b.Bytes(oracle.Far(0x1b1, 0xc000), 64)) == "3f5edbc8b51aa017211fb0ca808f807a1ade51d95e14b49dadf593e90ada24eb"
			if cave {
				caveCount++
				// Draw回傳保留stamp數，含完全被原版新畫面遮住的stamp。
				// 是否有殘字以底下每個實際像素判斷，不以內部計數推定。
				// SAR1生命／寶石與SAR2金幣／地點是本次實際消費端來源。
				for _, want := range []string{"zelres1.file.00db4c", "zelres1.file.00db54", "zelres2.file.000d34", "zelres2.file.000d3c", "zelres3.file.00c31a"} {
					if !containsHUDKey(keys, want) {
						t.Fatal("洞窟狀態列缺中文", f[1], want, keys)
					}
				}
				for y := 0; y < 600; y++ {
					for x := 0; x < 960; x++ {
						if image.Pt(x/3, y/3).In(image.Rect(48, 162, 272, 199)) {
							continue
						}
						i := canvas.PixOffset(x, y)
						if !bytes.Equal(before[i:i+4], canvas.Pix[i:i+4]) {
							t.Fatal("戰鬥畫面或框外被覆繪", f[1], x, y)
						}
					}
				}
			}
			if !bytes.Equal(memory, b.Bytes(oracle.Far(0, 0), 1<<20)) {
				t.Fatal("繪製改原版RAM")
			}
			file, err := os.Create(filepath.Join(out, f[1]+".png"))
			if err != nil {
				t.Fatal(err)
			}
			if err := png.Encode(file, canvas); err != nil {
				file.Close()
				t.Fatal(err)
			}
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
			var retainedLabels []map[string]any
			for _, stamp := range labels.stamps {
				var sourceKeys []string
				for key, source := range labels.sources {
					if stamp.source == source {
						sourceKeys = append(sourceKeys, key)
					}
				}
				masks, _ := labels.paint.masks(stamp.paintSerial, raw)
				visible := 0
				for _, row := range stamp.rows {
					for y := row.Min.Y; y < row.Max.Y; y++ {
						for x := row.Min.X; x < row.Max.X; x++ {
							if !occluded(image.Pt(x, y), masks) {
								visible++
							}
						}
					}
				}
				retainedLabels = append(retainedLabels, map[string]any{"keys": sourceKeys, "box": stamp.box, "visible_native_pixels": visible, "text": string(stamp.source.text)})
			}
			points = append(points, map[string]any{"name": f[1], "steps": b.Steps(), "cycles": b.Cycles(), "state_equal": true, "memory_sha256": fmtHash(memory), "indexed_sha256": fmtHash(raw), "cave": cave, "HUD_keys": keys, "HUD_changed_pixels": changed, "dialogue_stamp_active": drawn, "label_stamp_count": labelCount, "retained_labels": retainedLabels, "battle_and_outside_HUD_differences": 0, "held": b.HeldKeys()})
		default:
			t.Fatal("未知正常鍵序", line)
		}
	}
	equal("final")
	if len(points) != 25 || caveCount != 19 || len(b.HeldKeys()) != 0 || len(b.Unimplemented()) != 0 || len(b.Wrote()) != 0 {
		t.Fatal("戰鬥路徑不完整", len(points), caveCount, b.HeldKeys(), b.Unimplemented(), b.Wrote())
	}
	data, err := json.MarshalIndent(map[string]any{"commands_sha256": fmtHash(commands), "initial_memory_sha256": initialMemory, "checkpoints": points, "cave_checkpoints": caveCount, "held": b.HeldKeys(), "wrote": b.Wrote(), "seed_method": "兩原版執行器執行前完整初始RAM/Regs相同；固定冷啟動、cycles與正常輸入，未替換或重擲亂數；遊戲種子欄位未知，不宣稱隨機規則對拍", "scope": "正常F7自建存檔到Malicia出劍、交戰、第一寶石、跳躍與蹲下；既有對話/標籤/HUD唯讀，25點完整原版狀態相同；非HD或GUI驗收"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "receipt.json"), append(data, 10), 0644); err != nil {
		t.Fatal(err)
	}
}

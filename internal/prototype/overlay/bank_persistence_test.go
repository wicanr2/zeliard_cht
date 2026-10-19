package overlay

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
)

// 各自交易、另存、退出，再冷開機讀自己的存檔；不複製兩側的中途狀態。
func TestBankLedgerSaveLoadReadOnly(t *testing.T) {
	root, font := townInputs(t)
	savePath := os.Getenv("ZELIARD_TOWN_SAVE")
	if savePath == "" {
		t.Skip("須提供正常玩家路徑建立的SAVETEST；skip不計驗收")
	}
	sourceSave, err := os.ReadFile(savePath)
	if err != nil || fmtHash(sourceSave) != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
		t.Fatal("自建USR來源不符", err)
	}
	info, err := os.Stat(savePath)
	if err != nil {
		t.Fatal(err)
	}
	commands, err := os.ReadFile("../../../tools/bank-persistence.commands")
	if err != nil {
		t.Fatal(err)
	}
	scratch := [2]string{t.TempDir(), t.TempDir()}
	for _, dir := range scratch {
		p := filepath.Join(dir, "SAVETEST.usr")
		if err = os.WriteFile(p, sourceSave, 0644); err != nil {
			t.Fatal(err)
		}
		if err = os.Chtimes(p, info.ModTime(), info.ModTime()); err != nil {
			t.Fatal(err)
		}
	}
	var machines [2]*oracle.Oracle
	var audio [2]*oracle.Audio
	var d *Dialogue
	var l *Labels
	boots := 0
	exited := true
	load := func() {
		if !exited {
			t.Fatal("尚未正常退出，不可冷開機")
		}
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
			o.TraceFiles()
			machines[i] = o
			audio[i] = o.NewAudio(44100)
		}
		d, err = NewDialogue(machines[1], filepath.Join(root, "ZELRES2.SAR"), "../../../text/dialogue-prototype.tsv", font)
		if err != nil {
			t.Fatal(err)
		}
		l, err = NewLabels(machines[1], filepath.Join(root, "ZELRES2.SAR"), "../../../text/labels-prototype.tsv", font)
		if err != nil {
			t.Fatal(err)
		}
		d.UseLabels(l)
		boots++
		exited = false
	}
	load()
	defer func() {
		for _, o := range machines {
			if o != nil {
				o.Close()
			}
		}
	}()
	number := func(s string) uint64 {
		v, e := strconv.ParseUint(s, 0, 64)
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	run := func(n uint64) {
		for n > 0 && !exited {
			step := uint64(50000)
			if n < step {
				step = n
			}
			var ends [2]bool
			for i, o := range machines {
				e := o.Run(step)
				var end *oracle.ExitError
				if e != nil && !errors.As(e, &end) {
					t.Fatal(e)
				}
				ends[i] = end != nil
				_ = audio[i].Render()
			}
			if ends[0] != ends[1] || machines[0].Regs() != machines[1].Regs() || machines[0].Steps() != machines[1].Steps() || machines[0].Cycles() != machines[1].Cycles() {
				t.Fatal("每次原版執行的暫存器、時間或退出不同", boots, machines[0].Steps())
			}
			exited = ends[0]
			n -= step
		}
	}
	sourceDir := os.Getenv("ZELIARD_BANK_PERSISTENCE_SOURCE")
	canonical := func(v any) []byte {
		b, e := json.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		var x any
		if e = json.Unmarshal(b, &x); e != nil {
			t.Fatal(e)
		}
		b, e = json.Marshal(x)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	var checkpoints []map[string]any
	seen := map[string]bool{}
	for _, line := range strings.Split(string(commands), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "run":
			run(number(f[1]))
		case "new":
			load()
		case "down", "up":
			for _, o := range machines {
				if f[0] == "down" {
					o.KeyDown(uint8(number(f[1])))
				} else {
					o.KeyUp(uint8(number(f[1])))
				}
			}
		case "shot":
			name := f[1]
			a, b := machines[0], machines[1]
			memory := a.Bytes(oracle.Far(0, 0), 1<<20)
			if !bytes.Equal(memory, b.Bytes(oracle.Far(0, 0), 1<<20)) || a.Regs() != b.Regs() || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) {
				t.Fatal("存讀流程原版與中文完整狀態不同", name)
			}
			canvas := inventoryOriginalCanvas(b)
			drawn := d.Draw(canvas, b.Indexed(), b.Palette(), image.Point{})
			l.ObserveFrame(b.Indexed(), b.Palette())
			labels := l.Draw(canvas, b.Indexed(), b.Palette(), image.Point{})
			if !strings.HasPrefix(name, "ledger-") {
				assertBankCheckpoint(t, name, d, l, b, canvas, drawn)
			}
			if name == "ledger-restored-balance" {
				bank := d.bank
				if !drawn || bank == nil || !bank.active || bank.current != nil || len(bank.parts) != 1 || bank.parts[0].source.key != "zelres2.file.01a073" || string(bank.parts[0].number.raw) != "50\xff" || b.Word(oracle.Far(0x1b1, 0x86)) != 950 {
					t.Fatal("冷讀後餘額來源、原版數字或身上金幣不符", d.BankState())
				}
				// 沿既有數字合成驗證，金額仍由原版字形提供。
				assertBankCheckpoint(t, "final-balance", d, l, b, canvas, drawn)
			}
			if name == "ledger-save-confirmed" && (!drawn || d.current < 0 || d.paragraphs[d.current].key != "zelres2.file.01cc66" || d.visible != len(d.paragraphs[d.current].text)) {
				t.Fatal("正常存檔確認未完整中文顯示")
			}
			if name == "ledger-dos-exited" && !exited {
				t.Fatal("未證明正常DOS退出")
			}
			if name == "ledger-restored" && b.Word(oracle.Far(0x1b1, 0x86)) != 950 {
				t.Fatal("讀回身上金幣不符")
			}
			if name == "ledger-final-outside" && (drawn || d.Active() || d.bank.active || b.Word(oracle.Far(0x1b1, 0x86)) != 950) {
				t.Fatal("離店未撤除或原版金幣改動")
			}
			if !bytes.Equal(memory, b.Bytes(oracle.Far(0, 0), 1<<20)) {
				t.Fatal("顯示改寫原版記憶體", name)
			}
			if sourceDir != "" {
				data, e := os.ReadFile(filepath.Join(sourceDir, name+".json"))
				if e != nil {
					t.Fatal(e)
				}
				var original map[string]any
				if e = json.Unmarshal(data, &original); e != nil {
					t.Fatal(e)
				}
				var palette []byte
				for _, color := range b.Palette() {
					palette = append(palette, color[:]...)
				}
				for key, value := range map[string]any{"boot": boots, "steps": b.Steps(), "cycles": b.Cycles(), "regs": b.Regs(), "memory_sha256": fmtHash(memory), "indexed_sha256": fmtHash(b.Indexed()), "palette_sha256": fmtHash(palette), "bank": d.BankState(), "dialogue_drawn": drawn, "labels": labels, "held": b.HeldKeys(), "file_ops": b.FileOps()} {
					if !bytes.Equal(canonical(value), canonical(original[key])) {
						t.Fatal("配對與實際探針來源不同", name, key)
					}
				}
				fp, e := os.Open(filepath.Join(sourceDir, name+".png"))
				if e != nil {
					t.Fatal(e)
				}
				im, e := png.Decode(fp)
				fp.Close()
				if e != nil {
					t.Fatal(e)
				}
				expected := image.NewRGBA(im.Bounds())
				draw.Draw(expected, expected.Bounds(), im, im.Bounds().Min, draw.Src)
				if !bytes.Equal(expected.Pix, canvas.Pix) {
					t.Fatal("來源畫面與配對覆繪不同", name)
				}
			}
			checkpoints = append(checkpoints, map[string]any{"name": name, "boot": boots, "steps": b.Steps(), "cycles": b.Cycles(), "memory_sha256": fmtHash(memory), "state_equal": true, "drawn": drawn, "labels": labels, "bank": d.BankState(), "gold": b.Word(oracle.Far(0x1b1, 0x86))})
			seen[name] = true
		default:
			t.Fatal("未知正常鍵序", line)
		}
	}
	if boots != 2 || len(checkpoints) != 56 || len(machines[1].HeldKeys()) != 0 || !seen["ledger-save-confirmed"] || !seen["ledger-dos-exited"] || !seen["ledger-restored-balance"] || !seen["ledger-final-outside"] {
		t.Fatal("銀行持續性正常路徑未完整走完")
	}
	var saves [2][]byte
	for i, dir := range scratch {
		saves[i], err = os.ReadFile(filepath.Join(dir, "LEDGER.usr"))
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(saves[0]) != 256 || !bytes.Equal(saves[0], saves[1]) || fmtHash(saves[0]) != "c3b21bcfe48c5b68fe2051bc9e818c054d7ea9e604ea2df6742b0186fb9444ba" {
		t.Fatal("兩側各自產生的新存檔不符原版觀察")
	}
	readNew := false
	for _, op := range machines[1].FileOps() {
		if op.Op == "read" && strings.EqualFold(op.Name, "LEDGER.USR") && op.Len == 256 && !op.Failed {
			readNew = true
		}
	}
	if !readNew {
		t.Fatal("沒有從正式F7路徑讀取本次新存檔")
	}
	after, e := os.ReadFile(savePath)
	if e != nil || !bytes.Equal(after, sourceSave) {
		t.Fatal("原始自建USR改動", e)
	}
	finalInfo, e := os.Stat(savePath)
	if e != nil || !finalInfo.ModTime().Equal(info.ModTime()) {
		t.Fatal("原始自建USR時間戳改動", e)
	}
	if out := os.Getenv("ZELIARD_BANK_PERSISTENCE_OUTPUT"); out != "" {
		data, e := json.MarshalIndent(map[string]any{"status": "PASS_BANK_LEDGER_NORMAL_SAVE_LOAD_PAIR", "checkpoints": checkpoints, "full_original_comparisons": len(checkpoints), "cold_boots": boots, "generated_save_sha256": fmtHash(saves[0]), "generated_save_bytes": len(saves[0]), "new_save_read": readNew, "commands_sha256": fmtHash(commands), "source_save_unchanged": true, "observed_source_compared": sourceDir != "", "scope": "normal keyboard-only transaction, new save, DOS exit and cold F7; own native saves on each side, full original state and source canvas. Actual GUI and independent typography remain separate."}, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
}

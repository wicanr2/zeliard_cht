package overlay

import (
	"bytes"
	"encoding/json"
	"image"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
)

// 正常冷讀自建USR、左右走路、完整交談與關閉；語意與原版輸入不依賴譯文。
func TestTownSevenMurallaSourcesReadOnly(t *testing.T) {
	root, font := townInputs(t)
	savePath := os.Getenv("ZELIARD_TOWN_SAVE")
	if savePath == "" {
		t.Skip("須提供已由正常玩家路徑建立的SAVETEST；skip不計驗收")
	}
	save, err := os.ReadFile(savePath)
	if err != nil || fmtHash(save) != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
		t.Fatal("自建USR來源不符", err)
	}
	info, err := os.Stat(savePath)
	if err != nil {
		t.Fatal(err)
	}
	var machines [2]*oracle.Oracle
	var audio [2]*oracle.Audio
	for i := range machines {
		o, e := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if e != nil {
			t.Fatal(e)
		}
		defer o.Close()
		o.SetDOSBoxCycles(3000)
		o.TraceFiles()
		scratch := t.TempDir()
		p := filepath.Join(scratch, "SAVETEST.usr")
		if e = os.WriteFile(p, save, 0644); e != nil {
			t.Fatal(e)
		}
		if e = os.Chtimes(p, info.ModTime(), info.ModTime()); e != nil {
			t.Fatal(e)
		}
		o.SetScratch(scratch)
		machines[i], audio[i] = o, o.NewAudio(44100)
	}
	a, b := machines[0], machines[1]
	d, err := NewTownDialogue(b, filepath.Join(root, "ZELRES2.SAR"), "../../../text/town-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	commands, err := os.ReadFile("../../../tools/residents-muralla.commands")
	if err != nil {
		t.Fatal(err)
	}
	run := func(n uint64) {
		for n > 0 {
			step := uint64(50000)
			if n < step {
				step = n
			}
			for i, o := range machines {
				if e := o.Run(step); e != nil {
					t.Fatal(e)
				}
				_ = audio[i].Render()
			}
			n -= step
		}
	}
	number := func(s string) uint64 {
		n, e := strconv.ParseUint(s, 0, 64)
		if e != nil {
			t.Fatal(e)
		}
		return n
	}
	want := map[string]bool{"zelres2.file.039fff": true, "zelres2.file.03a0f8": true, "zelres2.file.039f30": true, "zelres2.file.039a8e": true, "zelres2.file.039b90": true, "zelres2.file.0399f5": true, "zelres2.file.039e6c": true}
	seen := map[string]int{}
	checkpoints, pages := 0, 0
	var receipt []map[string]any
	for _, line := range strings.Split(string(commands), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "run":
			run(number(f[1]))
		case "key":
			scan := uint8(number(f[1]))
			for _, o := range machines {
				o.KeyDown(scan)
			}
			run(number(f[2]))
			for _, o := range machines {
				o.KeyUp(scan)
			}
		case "shot":
			name := f[1]
			checkpoints++
			memory := a.Bytes(oracle.Far(0, 0), 1<<20)
			if !bytes.Equal(memory, b.Bytes(oracle.Far(0, 0), 1<<20)) || a.Regs() != b.Regs() || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) {
				t.Fatal("原版與中文完整狀態不同", name)
			}
			canvas := inventoryOriginalCanvas(b)
			before := append([]byte(nil), canvas.Pix...)
			drawn := d.Draw(canvas, b.Indexed(), b.Palette(), image.Point{})
			isPage := strings.Contains(name, "-page")
			if drawn != isPage || d.Rejected != 0 {
				t.Fatal("正常頁面／關閉狀態不符", name, d.State())
			}
			changed := 0
			if isPage {
				i := strings.Index(name, "zelres2.file.")
				expected := name[i : i+len("zelres2.file.")+6]
				if d.current == nil || d.current.key != expected || !want[expected] {
					t.Fatal("原版來源鍵不符", name, d.State())
				}
				seen[expected]++
				pages++
			}
			anchor, size := b.Word(oracle.Far(0x1b1, 0x7c4e)), b.Word(oracle.Far(0x1b1, 0x7c5a))
			x, y := int(anchor>>8)*8, int(anchor&255)
			box := image.Rect(x, y, x+int(size>>8)*4, y+int(size&255))
			safe := box.Inset(4)
			marker := image.Rect(x+84, y+74, x+92, y+82)
			for py := 0; py < 600; py++ {
				for px := 0; px < 960; px++ {
					at := canvas.PixOffset(px, py)
					if bytes.Equal(before[at:at+4], canvas.Pix[at:at+4]) {
						continue
					}
					changed++
					p := image.Pt(px/3, py/3)
					if !isPage || !p.In(safe) || d.prompt && p.In(marker) {
						t.Fatal("中文超出安全區、改動標記或未撤除", name, p)
					}
				}
			}
			if isPage && changed == 0 || !isPage && changed != 0 || !bytes.Equal(memory, b.Bytes(oracle.Far(0, 0), 1<<20)) {
				t.Fatal("中文覆繪或只讀契約失敗", name)
			}
			receipt = append(receipt, map[string]any{"name": name, "steps": b.Steps(), "cycles": b.Cycles(), "memory_sha256": fmtHash(memory), "state_equal": true, "drawn": drawn, "changed_pixels": changed, "outside_and_marker_changed": 0, "town": d.State()})
		default:
			t.Fatal("未知輸入語法", line)
		}
	}
	if checkpoints != 19 || pages != 11 || len(seen) != 7 || d.current != nil || d.Rejected != 0 || len(a.HeldKeys()) != 0 || len(a.Wrote()) != 0 || len(b.Wrote()) != 0 {
		t.Fatal("七來源完整路徑未完成", checkpoints, pages, seen, d.State())
	}
	after, err := os.ReadFile(savePath)
	if err != nil || !bytes.Equal(save, after) {
		t.Fatal("原始USR改動", err)
	}
	if out := os.Getenv("ZELIARD_RESIDENTS_OUTPUT"); out != "" {
		data, e := json.MarshalIndent(map[string]any{"status": "PASS_SEVEN_RESIDENT_INPUT_CORPUS", "full_original_comparisons": checkpoints, "pages": pages, "sources": seen, "checkpoints": receipt, "commands_sha256": fmtHash(commands), "save_unchanged": true, "scope": "normal cold load and original keyboard actions; read-only Chinese, outside pixels and continuation marker. Actual GUI and independent full typography have separate receipts."}, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
}

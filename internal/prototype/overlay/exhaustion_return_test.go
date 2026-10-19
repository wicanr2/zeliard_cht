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
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
	"github.com/wicanr2/zeliard_cht/internal/prototype/hd"
)

func exhaustionInputs(t *testing.T) (string, string, string, bitmap.Font) {
	t.Helper()
	root, project, study := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_PROJECT"), os.Getenv("ZELIARD_CAVE_STUDY")
	if root == "" || project == "" || study == "" || os.Getenv("ZELIARD_FONT") == "" {
		t.Skip("須掛載原版、專案、字型及已驗送返來源")
	}
	font, e := bitmap.Load(os.Getenv("ZELIARD_FONT"))
	if e != nil {
		t.Fatal(e)
	}
	return root, project, study, font
}

func exhaustionSage(t *testing.T, project, study string) *hd.Sage {
	t.Helper()
	s, e := hd.NewSage(filepath.Join(project, "workplace/save-path-verified/sage-menu.indexed"), filepath.Join(project, "workplace/save-path-verified/save-confirm-text.indexed"), filepath.Join(study, "exploration-return-source-v1/10-revival-message-full.palette"), filepath.Join(project, "workplace/hd-sage-prototype.png"))
	if e != nil {
		t.Fatal(e)
	}
	return s
}

func TestExhaustionReturnCatalogAndSequence(t *testing.T) {
	root, project, study, font := exhaustionInputs(t)
	rows, e := os.ReadFile(filepath.Join(study, "exploration-return-translation-v1.tsv"))
	if e != nil {
		t.Fatal(e)
	}
	load := func(text string) (*Dialogue, error) {
		p := filepath.Join(t.TempDir(), "dialogue.tsv")
		if e := os.WriteFile(p, []byte(text), 0644); e != nil {
			t.Fatal(e)
		}
		o, e := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if e != nil {
			t.Fatal(e)
		}
		t.Cleanup(o.Close)
		return NewDialogue(o, filepath.Join(root, "ZELRES2.SAR"), p, font)
	}
	d, e := load(string(rows))
	if e != nil || len(d.paragraphs) != 4 {
		t.Fatal("四字段載入不符", e)
	}
	for _, bad := range []string{"missing-first", "missing-last", "combined-overflow"} {
		t.Run(bad, func(t *testing.T) {
			lines := strings.Split(strings.TrimSpace(string(rows)), "\n")
			switch bad {
			case "missing-first":
				lines = append(lines[:1], lines[2:]...)
			case "missing-last":
				lines = lines[:len(lines)-1]
			case "combined-overflow":
				for i := 1; i < len(lines); i++ {
					lines[i] = strings.SplitN(lines[i], "\t", 2)[0] + "\t" + strings.Repeat("神", 20)
				}
			}
			if got, e := load(strings.Join(lines, "\n") + "\n"); e == nil || got != nil {
				t.Fatal("不完整或累積溢出未拒絕整組")
			}
		})
	}
	find := func(key string) int {
		for i, p := range d.paragraphs {
			if p.key == key {
				return i
			}
		}
		t.Fatal("缺來源", key)
		return -1
	}
	var lead []rune
	for _, key := range exhaustionReturnKeys {
		next := find(key)
		got, ok := d.retainedDialogueLead(next)
		if !ok || string(got) != string(lead) {
			t.Fatal("正常順序沒有保留已完成前文", key, string(got))
		}
		d.current, d.visible, d.active, d.lead = next, len(d.paragraphs[next].text), true, got
		lead = append(lead, d.paragraphs[next].text...)
	}
	d.reset()
	if _, ok := d.retainedDialogueLead(find(exhaustionReturnKeys[1])); ok {
		t.Fatal("缺第一段仍接受第二段")
	}
	d.current, d.visible, d.active = find(exhaustionReturnKeys[0]), 0, true
	if _, ok := d.retainedDialogueLead(find(exhaustionReturnKeys[1])); ok {
		t.Fatal("前段未完成仍借用全文")
	}
	d.visible = len(d.paragraphs[d.current].text)
	if _, ok := d.retainedDialogueLead(find(exhaustionReturnKeys[2])); ok {
		t.Fatal("失序仍借用前文")
	}
	d.reset()
	if d.State()["active"] != false || d.State()["key"] != "" || d.State()["retained"] != 0 {
		t.Fatal("reset後殘留")
	}
	// 舊售後一句保留契約不因新四段而改變。
	o, e := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
	if e != nil {
		t.Fatal(e)
	}
	defer o.Close()
	old, e := NewDialogue(o, filepath.Join(root, "ZELRES2.SAR"), filepath.Join(project, "text/dialogue-prototype.tsv"), font)
	if e != nil {
		t.Fatal(e)
	}
	a, b := -1, -1
	for i, p := range old.paragraphs {
		if p.key == "zelres2.file.01af36" {
			a = i
		}
		if p.key == "zelres2.file.01aeee" {
			b = i
		}
	}
	if a < 0 || b < 0 {
		t.Fatal("缺舊售後來源")
	}
	old.current, old.lead = a, []rune("舊前文")
	if got, ok := old.retainedDialogueLead(b); !ok || string(got) != string(old.paragraphs[a].text) {
		t.Fatal("改變舊售後保留契約")
	}
}

func TestExhaustionReturnNormalPairIndependentPixels(t *testing.T) {
	root, project, study, font := exhaustionInputs(t)
	out := os.Getenv("ZELIARD_RETURN_OUTPUT")
	savePath := os.Getenv("ZELIARD_CAVE_SAVE")
	if out == "" || savePath == "" {
		t.Skip("須指定新輸出及正常自建存檔")
	}
	read := func(path string) []byte {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	save := read(savePath)
	if fmtHash(save) != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
		t.Fatal("自建存檔不符")
	}
	info, e := os.Stat(savePath)
	if e != nil {
		t.Fatal(e)
	}
	var machines [2]*oracle.Oracle
	var saves []string
	for i := range machines {
		o, e := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if e != nil {
			t.Fatal(e)
		}
		machines[i] = o
		t.Cleanup(o.Close)
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
		saves = append(saves, p)
	}
	a, b := machines[0], machines[1]
	equal := func(name string) {
		if !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) || a.Regs() != b.Regs() || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) {
			t.Fatal("覆繪改變原版", name)
		}
	}
	equal("initial")
	d, e := NewDialogue(b, filepath.Join(root, "ZELRES2.SAR"), filepath.Join(project, "text/dialogue-prototype.tsv"), font)
	if e != nil {
		t.Fatal(e)
	}
	sage := exhaustionSage(t, project, study)
	var proof struct {
		Records []struct {
			Name, Lead, Text string
			Active, Prompt   bool
			Visible          int
			Key              string `json:"current_key"`
			PNG              string `json:"PNG_sha256"`
		}
	}
	if e = json.Unmarshal(read(filepath.Join(study, "exploration-return-preview-v1/receipt.json")), &proof); e != nil || len(proof.Records) != 9 {
		t.Fatal("獨立九圖不齊", e)
	}
	wants := map[string]int{}
	for i, p := range proof.Records {
		wants[p.Name] = i
	}
	if e = os.Mkdir(filepath.Join(out, "pixels"), 0755); e != nil {
		t.Fatal(e)
	}
	run := func(n uint64) {
		for n > 0 {
			chunk := n
			if a.Steps() >= 97000000 && chunk > 20000 {
				chunk = 20000
			}
			for _, o := range machines {
				if e := o.Run(chunk); e != nil {
					t.Fatal(e)
				}
			}
			n -= chunk
		}
	}
	number := func(s string) uint64 {
		n, e := strconv.ParseUint(s, 0, 64)
		if e != nil || n > 800000000 {
			t.Fatal("指令預算", s, e)
		}
		return n
	}
	commands := read(filepath.Join(study, "exploration-return-v1.commands"))
	var points []map[string]any
	var withdrawals []map[string]any
	pixelFrames := 0
	for _, line := range strings.Split(string(commands), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		switch f[0] {
		case "run":
			run(number(f[1]))
		case "key", "keys":
			var scans []uint8
			for _, s := range strings.Split(f[1], ",") {
				n := number(s)
				if n > 255 {
					t.Fatal("掃描碼")
				}
				scans = append(scans, uint8(n))
			}
			for _, o := range machines {
				for _, s := range scans {
					o.KeyDown(s)
				}
			}
			run(number(f[2]))
			for _, o := range machines {
				for _, s := range scans {
					o.KeyUp(s)
				}
			}
		case "down", "up":
			n := number(f[1])
			if n > 255 {
				t.Fatal("掃描碼")
			}
			for _, o := range machines {
				if f[0] == "down" {
					o.KeyDown(uint8(n))
				} else {
					o.KeyUp(uint8(n))
				}
			}
		case "shot":
			equal(f[1])
			name := fmt.Sprintf("%02d-%s", len(points), f[1])
			raw, pal := b.Indexed(), b.Palette()
			memory := b.Bytes(oracle.Far(0, 0), 1<<20)
			if !bytes.Equal(memory, read(filepath.Join(study, "exploration-return-source-v1", name+".memory"))) || !bytes.Equal(raw, read(filepath.Join(study, "exploration-return-source-v1", name+".indexed"))) {
				t.Fatal("偏離原探索同狀態", name)
			}
			canvas := inventoryOriginalCanvas(b)
			before := bytes.Clone(canvas.Pix)
			_, _, rgb := b.ScreenRGB()
			hdDrawn, pose := sage.Draw(canvas, raw, rgb, image.Point{})
			priorRejections := len(d.Rejections)
			drawn := d.Draw(canvas, raw, pal, image.Point{})
			state := d.State()
			if len(d.Rejections) != priorRejections {
				if name != "15-return-after-4" || len(d.Rejections) != priorRejections+1 || state["active"] != false || d.Glyphs != 157 || d.Paragraphs != 4 {
					t.Fatal("非預期的對話拒絕", name, d.Rejections)
				}
				rejection := d.Rejections[len(d.Rejections)-1]
				if rejection["kind"] != 2 || rejection["first_difference"] != 0 || rejection["visible"] != len([]rune("神靈正守護著你。")) {
					t.Fatal("正常離場撤除紀錄不符", rejection)
				}
				withdrawals = append(withdrawals, map[string]any{"point": name, "steps": b.Steps(), "rejection": rejection, "original_clear_and_return_verified": true})
			}
			pixels := 0
			if i, ok := wants[name]; ok {
				p := proof.Records[i]
				if drawn != p.Active || hdDrawn != p.Active || state["active"] != p.Active {
					t.Fatal("顯示或撤除狀態不符", name, drawn, hdDrawn, state)
				}
				if p.Active && (state["key"] != p.Key || d.visible != p.Visible || string(d.lead) != p.Lead || d.prompt != p.Prompt) {
					t.Fatal("逐字／保留／待續不同", name, state, p.Visible, string(d.lead), p.Lead)
				}
				data := read(filepath.Join(study, "exploration-return-preview-v1", name+".png"))
				if fmtHash(data) != p.PNG {
					t.Fatal("獨立圖身份不符")
				}
				want, e := png.Decode(bytes.NewReader(data))
				if e != nil {
					t.Fatal(e)
				}
				for y := 0; y < 600; y++ {
					for x := 0; x < 960; x++ {
						if canvas.RGBAAt(x, y) != color.RGBAModel.Convert(want.At(x, y)).(color.RGBA) {
							t.Fatalf("獨立整幅不符 %s (%d,%d)", name, x, y)
						}
					}
				}
				pixels = 960 * 600
				pixelFrames++
			} else if drawn || hdDrawn || !bytes.Equal(before, canvas.Pix) {
				t.Fatal("送返前誤蓋洞窟", name)
			}
			if !bytes.Equal(memory, b.Bytes(oracle.Far(0, 0), 1<<20)) {
				t.Fatal("繪製寫原版RAM")
			}
			equal("after-draw")
			file, e := os.Create(filepath.Join(out, "pixels", name+".png"))
			if e != nil {
				t.Fatal(e)
			}
			if e = png.Encode(file, canvas); e != nil {
				file.Close()
				t.Fatal(e)
			}
			file.Close()
			points = append(points, map[string]any{"name": name, "steps": b.Steps(), "cycles": b.Cycles(), "memory_sha256": fmtHash(memory), "indexed_sha256": fmtHash(raw), "state_equal": true, "dialogue": state, "dialogue_drawn": drawn, "sage_HD": hdDrawn, "sage_pose": pose, "independent_pixels": pixels, "pixel_differences": 0})
		case "quit":
		default:
			t.Fatal("未知命令", line)
		}
		equal(line)
	}
	if len(points) != 17 || pixelFrames != 9 || d.Glyphs != 157 || d.Scrolls != 10 || d.Paragraphs != 4 || d.Active() || len(d.Rejections) != 1 || len(withdrawals) != 1 || len(b.HeldKeys()) != 0 || len(b.Unimplemented()) != 0 || len(b.Wrote()) != 0 {
		t.Fatal("正常送返鏈未閉合", len(points), pixelFrames, d.Glyphs, d.Scrolls, d.Paragraphs, d.State(), d.Rejections)
	}
	for _, p := range saves {
		stat, e := os.Stat(p)
		if e != nil || !bytes.Equal(read(p), save) || !stat.ModTime().Equal(info.ModTime()) {
			t.Fatal("改變存檔或mtime", e)
		}
	}
	data, e := json.MarshalIndent(map[string]any{"status": "PASS_EXHAUSTION_RETURN_NORMAL_PAIR_AND_INDEPENDENT_PIXELS", "points": points, "commands_sha256": fmtHash(commands), "full_original_state_equal": true, "save_unchanged": true, "save_mtime_ns": info.ModTime().UnixNano(), "glyphs": d.Glyphs, "scrolls": d.Scrolls, "paragraphs": d.Paragraphs, "independent_frames": pixelFrames, "independent_pixels": pixelFrames * 960 * 600, "expected_original_withdrawals": withdrawals, "unexpected_rejections": 0, "fileops": b.FileOps(), "keyreads": b.KeyReads(), "wrote": b.Wrote(), "scope": "正常冷啟動、原版昏迷送返、四字段、待續、捲動及回鎮；未注入或重擲，不宣稱隨機規則／GUI或全遊戲完成"}, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, 10), 0644); e != nil {
		t.Fatal(e)
	}
}

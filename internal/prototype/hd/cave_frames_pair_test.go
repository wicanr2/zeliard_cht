package hd

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
	"github.com/wicanr2/zeliard_cht/internal/prototype/overlay"
)

func TestCaveFramesNormalPathsReadOnly(t *testing.T) {
	paths, out := os.Getenv("ZELIARD_CAVE_COMMANDS"), os.Getenv("ZELIARD_CAVE_OUTPUT")
	if paths == "" || out == "" {
		t.Skip("須掛載正常鍵序及輸出；skip不算驗收")
	}
	for _, path := range filepath.SplitList(paths) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			runCaveFramesPair(t, path, filepath.Join(out, strings.TrimSuffix(filepath.Base(path), ".commands")))
		})
	}
}

// 重播既有失敗GUI的同一份原生IRQ／RunCycles日誌，無狀態注入。
func TestCaveFramesRecordedGUIReadOnly(t *testing.T) {
	root, project, study, out := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_PROJECT"), os.Getenv("ZELIARD_CAVE_STUDY"), os.Getenv("ZELIARD_CAVE_JOURNAL_OUTPUT")
	if root == "" || project == "" || study == "" || out == "" {
		t.Skip("須掛載固定GUI日誌與輸出；skip不算驗收")
	}
	read := func(path string) []byte {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	guiDir := os.Getenv("ZELIARD_CAVE_GUI_REPLAY")
	customGUI := guiDir != ""
	if !customGUI {
		guiDir = filepath.Join(study, "gui-current-combat-v6")
	}
	data := read(filepath.Join(guiDir, "guest-inputs.jsonl"))
	type journalRecord struct {
		Kind          string      `json:"kind"`
		CyclesPerMS   uint64      `json:"cycles_per_ms"`
		SampleRate    int         `json:"audio_sample_rate"`
		InitialMemory string      `json:"initial_memory_sha256"`
		InitialRegs   oracle.Regs `json:"initial_regs"`
		Scan          uint8       `json:"scan"`
		Pressed       bool        `json:"pressed"`
		Budget        uint64      `json:"cycle_budget"`
		BeforeSteps   uint64      `json:"before_steps"`
		BeforeCycles  uint64      `json:"before_cycles"`
		AfterSteps    uint64      `json:"after_steps"`
		AfterCycles   uint64      `json:"after_cycles"`
		AfterRegs     oracle.Regs `json:"after_regs"`
		Audio         bool        `json:"audio_render"`
		Error         string      `json:"error"`
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var header journalRecord
	if e := json.Unmarshal([]byte(lines[0]), &header); e != nil || header.Kind != "header" {
		t.Fatal("journal header", e)
	}
	savePath := os.Getenv("ZELIARD_CAVE_SAVE")
	save := read(savePath)
	info, e := os.Stat(savePath)
	if e != nil {
		t.Fatal(e)
	}
	if townBackgroundHash(save) != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
		t.Fatal("save identity")
	}
	var machines [2]*oracle.Oracle
	var audio [2]*oracle.Audio
	var scratchSaves []string
	for i := range machines {
		o, e := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if e != nil {
			t.Fatal(e)
		}
		defer o.Close()
		machines[i] = o
		scratch := t.TempDir()
		p := filepath.Join(scratch, "SAVETEST.usr")
		if e = os.WriteFile(p, save, 0644); e != nil {
			t.Fatal(e)
		}
		if e = os.Chtimes(p, info.ModTime(), info.ModTime()); e != nil {
			t.Fatal(e)
		}
		o.SetScratch(scratch)
		scratchSaves = append(scratchSaves, p)
		o.SetDOSBoxCycles(header.CyclesPerMS)
		o.TraceFiles()
		audio[i] = o.NewAudio(header.SampleRate)
		if townBackgroundHash(o.Bytes(oracle.Far(0, 0), 1<<20)) != header.InitialMemory || o.Regs() != header.InitialRegs {
			t.Fatal("cold source differs")
		}
	}
	a, b := machines[0], machines[1]
	frames := NewCaveFrames(b)
	art, e := NewCaveTerrain(os.Getenv("ZELIARD_CAVE_ART"))
	if e != nil {
		t.Fatal(e)
	}
	if dir := os.Getenv("ZELIARD_CAVE_ICE_ART"); dir != "" {
		if e = art.LoadIce(dir); e != nil {
			t.Fatal(e)
		}
	}
	var player *CavePlayer
	if dir := os.Getenv("ZELIARD_CAVE_MPP3_ART"); dir != "" {
		if e = art.LoadMPP3(dir); e != nil {
			t.Fatal(e)
		}
	}
	if dir := os.Getenv("ZELIARD_CAVE_MPP5_ART"); dir != "" {
		if e = art.LoadMPP5(dir); e != nil {
			t.Fatal(e)
		}
	}
	if dir := os.Getenv("ZELIARD_CAVE_MPP6_ART"); dir != "" {
		if e = art.LoadMPP6(dir); e != nil {
			t.Fatal(e)
		}
	}
	if dir := os.Getenv("ZELIARD_CAVE_MPP7_ART"); dir != "" {
		if e = art.LoadMPP7(dir); e != nil {
			t.Fatal(e)
		}
	}
	loadRemainingNormalTestArt(t, art)
	if dir := os.Getenv("ZELIARD_CAVE_PELIGRO_ART"); dir != "" {
		if e = art.LoadPeligro(dir); e != nil {
			t.Fatal(e)
		}
	}
	if playerDir := os.Getenv("ZELIARD_CAVE_PLAYER_ART"); playerDir != "" {
		player, e = NewCavePlayer(playerDir)
		if e != nil {
			t.Fatal(e)
		}
		art.SetPlayer(player)
	}
	var weapon *CaveWeapon
	if dir := os.Getenv("ZELIARD_CAVE_WEAPON_ART"); dir != "" {
		weapon, e = NewCaveWeapon(dir)
		if e != nil {
			t.Fatal(e)
		}
		loadCyanNormalArt(t, weapon)
		art.SetWeapon(weapon)
	}
	weaponCoverage := newCaveWeaponCoverage(weapon)
	var objects *CaveObjects
	if dir := os.Getenv("ZELIARD_CAVE_OBJECT_ART"); dir != "" {
		objects, e = NewCaveObjects(dir)
		if e != nil {
			t.Fatal(e)
		}
		art.SetObjects(objects)
	}
	objectCoverage := newCaveObjectCoverage(objects)
	configureCavePortal(t, art)
	if e = os.Mkdir(out, 0755); e != nil {
		t.Fatal(e)
	}
	var missingSources []map[string]any
	seenSources := map[uint64]bool{}
	seenMissing := map[string]bool{}
	draws, paintFrames := 0, 0
	for index, line := range lines[1:] {
		var r journalRecord
		if e = json.Unmarshal([]byte(line), &r); e != nil {
			t.Fatal(e)
		}
		if a.Steps() != r.BeforeSteps || a.Cycles() != r.BeforeCycles {
			t.Fatal("record before", index+2)
		}
		for i, o := range machines {
			switch r.Kind {
			case "key":
				if r.Pressed {
					o.KeyDown(r.Scan)
				} else {
					o.KeyUp(r.Scan)
				}
			case "run":
				if r.Budget > 5000000 || o.Steps() > 200000000 {
					t.Fatal("journal budget")
				}
				err := o.RunCycles(r.Budget)
				message := ""
				if err != nil {
					message = err.Error()
				}
				if r.Audio {
					audio[i].Render()
				}
				if message != r.Error || o.Regs() != r.AfterRegs {
					t.Fatal("record registers/error", index+2, i)
				}
			default:
				t.Fatal("unknown record", r.Kind)
			}
			if o.Steps() != r.AfterSteps || o.Cycles() != r.AfterCycles {
				t.Fatal("record after", index+2, i)
			}
		}
		if r.Kind == "run" {
			_, _, rgb := b.ScreenRGB()
			selected := frames.Present(b.Indexed(), rgb, false)
			if selected != nil {
				weaponCoverage.observe(t, out, selected)
				objectCoverage.observe(t, out, selected)
				if player != nil && !seenSources[selected.Steps] {
					seenSources[selected.Steps] = true
					key, pose, hidden := player.selectPose(selected.Decoded)
					if pose == nil && !hidden && !seenMissing[key] {
						seenMissing[key] = true
						stem := fmt.Sprintf("player-missing-%02d", len(missingSources))
						for ext, value := range map[string][]byte{"memory": selected.Memory, "body.memory": selected.BodyMemory, "indexed": selected.SourceIndexed, "palette": selected.Palette} {
							if e = os.WriteFile(filepath.Join(out, stem+"."+ext), value, 0644); e != nil {
								t.Fatal(e)
							}
						}
						value, err := json.Marshal(selected.Paints)
						if err != nil {
							t.Fatal(err)
						}
						if e = os.WriteFile(filepath.Join(out, stem+".paints.json"), value, 0644); e != nil {
							t.Fatal(e)
						}
						missingSources = append(missingSources, map[string]any{"key": key, "file": stem, "steps": selected.Steps, "body_steps": selected.BodySteps, "visible_actor_pixels": len(selected.Decoded.Actor), "memory_sha256": townBackgroundHash(selected.Memory), "body_memory_sha256": townBackgroundHash(selected.BodyMemory), "indexed_sha256": townBackgroundHash(selected.SourceIndexed)})
					}
				}
				draws++
				if len(selected.Paints) > 0 {
					paintFrames++
				}
				if !art.Draw(caveOriginalCanvas(selected), selected, image.Point{}) {
					t.Fatal("recorded frame HD refused", index+2)
				}
			}
		}
		if !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Regs() != b.Regs() || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) {
			t.Fatal("adapter changed original", index+2)
		}
	}
	if frames.refused != 0 || (!customGUI && (frames.captured != 267 || paintFrames == 0)) || draws == 0 {
		t.Fatal("recorded path incomplete", frames.captured, frames.refused, draws, paintFrames, frames.lastRefusal)
	}
	if !bytes.Equal(b.Bytes(oracle.Far(0, 0), 1<<20), read(filepath.Join(guiDir, "town-final.png.memory"))) || !bytes.Equal(b.Indexed(), read(filepath.Join(guiDir, "town-final.png.indexed"))) {
		t.Fatal("final GUI original differs")
	}
	if !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) || len(b.Wrote()) != 0 || len(b.HeldKeys()) != 0 {
		t.Fatal("file/input parity")
	}
	for _, p := range scratchSaves {
		current, e := os.Stat(p)
		if e != nil || !current.ModTime().Equal(info.ModTime()) || !bytes.Equal(read(p), save) {
			t.Fatal("GUI重播改動存檔或mtime", e)
		}
	}
	for _, step := range weaponCoverage.captureOrder {
		if weaponCoverage.captureSteps[step] {
			t.Fatalf("未取得指定GUI完成來源：%d", step)
		}
	}
	for _, step := range objectCoverage.order {
		if objectCoverage.capture[step] {
			t.Fatalf("未取得指定GUI物件完成來源：%d", step)
		}
	}
	receipt := map[string]any{"status": "ORIGINAL_GUI_JOURNAL_PAIR_EQUAL", "records": len(lines), "completed_frames": frames.captured, "refused": frames.refused, "HD_draws": draws, "native_paint_draws": paintFrames, "full_RAM_and_regs_each_record_equal": true, "original_GUI_final_RAM_and_indexed_equal": true, "input_journal_sha256": townBackgroundHash(data), "final": frames.State(), "state_injection": false}
	receipt["player_sources"] = len(seenSources)
	receipt["player_missing_sources"] = missingSources
	receipt["weapon_coverage"] = weaponCoverage.summary()
	receipt["object_coverage"] = objectCoverage.summary()
	receipt["fileops"] = len(b.FileOps())
	receipt["save_unchanged"] = true
	encoded, e := json.MarshalIndent(receipt, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(encoded, 10), 0644); e != nil {
		t.Fatal(e)
	}
}

func runCaveFramesPair(t *testing.T, path, out string) {
	root, project := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_PROJECT")
	savePath := os.Getenv("ZELIARD_CAVE_SAVE")
	save, e := os.ReadFile(savePath)
	saveHash := townBackgroundHash(save)
	maxSteps := uint64(160000000)
	if saveHash == "03df9abcf62b1712d6b46dd1fc5b23335a4b8711b3dc84875a1498260d37735c" {
		// 043固定包內SAVEM1的正常平台往返；前史未知，不當新局全流程。
		maxSteps = 240000000
	} else if e != nil || saveHash != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
		t.Fatal("正常來源存檔不符", e)
	}
	info, e := os.Stat(savePath)
	if e != nil {
		t.Fatal(e)
	}
	var machines [2]*oracle.Oracle
	var scratchSaves [2]string
	for i := range machines {
		o, e := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
		if e != nil {
			t.Fatal(e)
		}
		machines[i] = o
		o.TraceFiles()
		defer o.Close()
		o.SetDOSBoxCycles(3000)
		scratch := t.TempDir()
		p := filepath.Join(scratch, "SAVETEST.usr")
		scratchSaves[i] = p
		if e = os.WriteFile(p, save, 0644); e != nil {
			t.Fatal(e)
		}
		if e = os.Chtimes(p, info.ModTime(), info.ModTime()); e != nil {
			t.Fatal(e)
		}
		o.SetScratch(scratch)
	}
	a, b := machines[0], machines[1]
	font, e := bitmap.Load(os.Getenv("ZELIARD_FONT"))
	if e != nil {
		t.Fatal(e)
	}
	labels, e := overlay.NewLabels(b, filepath.Join(root, "ZELRES2.SAR"), filepath.Join(project, "text/labels-prototype.tsv"), font)
	if e != nil {
		t.Fatal(e)
	}
	dialogue, e := overlay.NewDialogue(b, filepath.Join(root, "ZELRES2.SAR"), filepath.Join(project, "text/dialogue-prototype.tsv"), font)
	if e != nil {
		t.Fatal(e)
	}
	dialogue.UseLabels(labels)
	hud, e := overlay.NewHUD(b, root, filepath.Join(project, "text/hud-prototype.tsv"), font)
	if e != nil {
		t.Fatal(e)
	}
	frames := NewCaveFrames(b)
	var art *CaveTerrain
	if dir := os.Getenv("ZELIARD_CAVE_ART"); dir != "" {
		art, e = NewCaveTerrain(dir)
		if e != nil {
			t.Fatal(e)
		}
	}
	if dir := os.Getenv("ZELIARD_CAVE_ICE_ART"); dir != "" {
		if art == nil {
			t.Fatal("冰窟HD須搭配洞窟呈現器")
		}
		if e = art.LoadIce(dir); e != nil {
			t.Fatal(e)
		}
	}
	if dir := os.Getenv("ZELIARD_CAVE_PELIGRO_ART"); dir != "" {
		if art == nil {
			t.Fatal("MPP2須搭配洞窟呈現器")
		}
		if e = art.LoadPeligro(dir); e != nil {
			t.Fatal(e)
		}
	}
	if dir := os.Getenv("ZELIARD_CAVE_MPP3_ART"); dir != "" {
		if art == nil {
			t.Fatal("MPP3須搭配洞窟呈現器")
		}
		if e = art.LoadMPP3(dir); e != nil {
			t.Fatal(e)
		}
	}
	if dir := os.Getenv("ZELIARD_CAVE_MPP5_ART"); dir != "" {
		if art == nil {
			t.Fatal("MPP5須搭配洞窟呈現器")
		}
		if e = art.LoadMPP5(dir); e != nil {
			t.Fatal(e)
		}
	}
	if dir := os.Getenv("ZELIARD_CAVE_MPP6_ART"); dir != "" {
		if art == nil {
			t.Fatal("MPP6須搭配洞窟呈現器")
		}
		if e = art.LoadMPP6(dir); e != nil {
			t.Fatal(e)
		}
	}
	if dir := os.Getenv("ZELIARD_CAVE_MPP7_ART"); dir != "" {
		if art == nil {
			t.Fatal("MPP7須搭配洞窟呈現器")
		}
		if e = art.LoadMPP7(dir); e != nil {
			t.Fatal(e)
		}
	}
	loadRemainingNormalTestArt(t, art)
	if dir := os.Getenv("ZELIARD_CAVE_PLAYER_ART"); dir != "" {
		if art == nil {
			t.Fatal("主角HD須搭配已驗地形")
		}
		player, err := NewCavePlayer(dir)
		if err != nil {
			t.Fatal(err)
		}
		art.SetPlayer(player)
	}
	var weapon *CaveWeapon
	if dir := os.Getenv("ZELIARD_CAVE_WEAPON_ART"); dir != "" {
		if art == nil {
			t.Fatal("武器HD須搭配地形")
		}
		weapon, e = NewCaveWeapon(dir)
		if e != nil {
			t.Fatal(e)
		}
		loadCyanNormalArt(t, weapon)
		art.SetWeapon(weapon)
	}
	weaponCoverage := newCaveWeaponCoverage(weapon)
	var objects *CaveObjects
	if dir := os.Getenv("ZELIARD_CAVE_OBJECT_ART"); dir != "" {
		if art == nil {
			t.Fatal("物件HD須搭配地形")
		}
		objects, e = NewCaveObjects(dir)
		if e != nil {
			t.Fatal(e)
		}
		art.SetObjects(objects)
	}
	objectCoverage := newCaveObjectCoverage(objects)
	configureCavePortal(t, art)
	seenPlayerFrames := map[uint64]bool{}
	seenPlayerKeys := map[string]int{}
	var missingPlayerSources []map[string]any
	missingPlayerKeys := map[string]bool{}
	playerKnown, playerHidden, playerMissing := 0, 0, 0
	hdDraws := 0
	equal := func(name string) {
		if !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) || a.Regs() != b.Regs() || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) {
			t.Fatal("適配層改動完整原版", name)
		}
	}
	equal("initial")
	var selected *CavePresentation
	var observations []map[string]any
	draws, used, held, cleared := 0, 0, 0, 0
	present := func() {
		draws++
		raw := b.Indexed()
		_, _, rgb := b.ScreenRGB()
		selected = frames.Present(raw, rgb, false)
		if selected == nil {
			if frames.captured > 0 {
				cleared++
			}
			return
		}
		used++
		weaponCoverage.observe(t, out, selected)
		objectCoverage.observe(t, out, selected)
		if art != nil && art.player != nil && !seenPlayerFrames[selected.Steps] {
			seenPlayerFrames[selected.Steps] = true
			key, pose, hidden := art.player.selectPose(selected.Decoded)
			seenPlayerKeys[key]++
			switch {
			case hidden:
				playerHidden++
			case pose != nil:
				playerKnown++
			default:
				playerMissing++
				if !missingPlayerKeys[key] {
					missingPlayerKeys[key] = true
					stem := fmt.Sprintf("player-missing-%02d", len(missingPlayerSources))
					for suffix, data := range map[string][]byte{"memory": selected.Memory, "body.memory": selected.BodyMemory, "indexed": selected.SourceIndexed, "palette": selected.Palette} {
						if e := os.WriteFile(filepath.Join(out, stem+"."+suffix), data, 0644); e != nil {
							t.Fatal(e)
						}
					}
					encoded, e := json.Marshal(selected.Paints)
					if e != nil {
						t.Fatal(e)
					}
					if e = os.WriteFile(filepath.Join(out, stem+".paints.json"), encoded, 0644); e != nil {
						t.Fatal(e)
					}
					missingPlayerSources = append(missingPlayerSources, map[string]any{"key": key, "file": stem, "steps": selected.Steps, "cycles": selected.Cycles, "body_steps": selected.BodySteps, "body_cycles": selected.BodyCycles, "visible_actor_pixels": len(selected.Decoded.Actor), "memory_sha256": townBackgroundHash(selected.Memory), "body_memory_sha256": townBackgroundHash(selected.BodyMemory), "indexed_sha256": townBackgroundHash(selected.SourceIndexed), "native_paints": selected.Paints})
				}
			}
		}
		if !caveFrameEqual(raw, selected.SourceIndexed) {
			held++
		}
		for y := 0; y < 200; y++ {
			for x := 0; x < 320; x++ {
				if x >= 48 && x < 272 && y >= 14 && y < 158 {
					continue
				}
				i := y*320 + x
				if raw[i] != selected.Indexed[i] || !bytes.Equal(rgb[i*3:i*3+3], selected.RGB[i*3:i*3+3]) {
					t.Fatal("呈現改動遊戲區外", b.Steps(), x, y)
				}
			}
		}
	}
	run := func(n uint64) {
		if b.Steps() > maxSteps || n > maxSteps-b.Steps() {
			t.Fatal("固定指令預算超出")
		}
		for n > 0 {
			chunk := n
			if b.Steps() >= 97000000 && chunk > 20000 {
				chunk = 20000
			}
			for _, o := range machines {
				if e := o.Run(chunk); e != nil {
					t.Fatal(e)
				}
			}
			n -= chunk
			present()
		}
	}
	num := func(s string) uint64 {
		n, e := strconv.ParseUint(s, 0, 64)
		if e != nil {
			t.Fatal(e)
		}
		return n
	}
	commands, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.MkdirAll(out, 0755); e != nil {
		t.Fatal(e)
	}
	for _, line := range strings.Split(string(commands), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		switch f[0] {
		case "run":
			run(num(f[1]))
		case "key", "keys":
			var scans []uint8
			for _, s := range strings.Split(f[1], ",") {
				if num(s) > 255 {
					t.Fatal("掃描碼越界")
				}
				scans = append(scans, uint8(num(s)))
			}
			for _, o := range machines {
				for _, s := range scans {
					o.KeyDown(s)
				}
			}
			run(num(f[2]))
			for _, o := range machines {
				for _, s := range scans {
					o.KeyUp(s)
				}
			}
		case "down", "up":
			scan := num(f[1])
			if scan > 255 {
				t.Fatal("掃描碼越界")
			}
			for _, o := range machines {
				if f[0] == "down" {
					o.KeyDown(uint8(scan))
				} else {
					o.KeyUp(uint8(scan))
				}
			}
		case "shot":
			equal(f[1])
			if (strings.HasPrefix(f[1], "platform-") || os.Getenv("ZELIARD_CAVE_REQUIRE_PRESENTATION") == "1") && selected == nil {
				t.Fatal("平台正常點缺少可呈現的完成來源", f[1], frames.State())
			}
			before := b.Bytes(oracle.Far(0, 0), 1<<20)
			point := map[string]any{"name": f[1], "steps": b.Steps(), "cycles": b.Cycles(), "memory_sha256": townBackgroundHash(before), "indexed_sha256": townBackgroundHash(b.Indexed()), "state_equal": true, "presentation": frames.State()}
			if selected != nil {
				if _, e := decodeCaveFrameWithPaints(selected.Memory, selected.BodyMemory, selected.SourceIndexed, selected.Paints); e != nil {
					t.Fatal("選出來源無法重建", e)
				}
				for suffix, data := range map[string][]byte{"memory": selected.Memory, "body.memory": selected.BodyMemory, "indexed": selected.SourceIndexed, "rgb": selected.RGB, "palette": selected.Palette} {
					if e = os.WriteFile(filepath.Join(out, f[1]+".presented."+suffix), data, 0644); e != nil {
						t.Fatal(e)
					}
				}
				point["source_steps"] = selected.Steps
				point["body_steps"] = selected.BodySteps
			}
			if art != nil {
				if selected != nil {
					canvas := caveOriginalCanvas(selected)
					if !art.Draw(canvas, selected, image.Point{}) {
						t.Fatal("正常來源HD被拒絕", f[1])
					}
					hdDraws++
					beforeText := cloneTownBytes(canvas.Pix)
					raw, pal := b.Indexed(), b.Palette()
					dialogue.Draw(canvas, raw, pal, image.Point{})
					labels.ObserveFrame(raw, pal)
					labels.Draw(canvas, raw, pal, image.Point{})
					hud.Draw(canvas, raw, pal, image.Point{})
					for y := 42; y < 474; y++ {
						for x := 144; x < 816; x++ {
							i := canvas.PixOffset(x, y)
							if !bytes.Equal(beforeText[i:i+4], canvas.Pix[i:i+4]) {
								t.Fatal("中文覆寫HD遊戲區", f[1], x, y)
							}
						}
					}
					file, err := os.Create(filepath.Join(out, f[1]+".png"))
					if err != nil {
						t.Fatal(err)
					}
					if err = png.Encode(file, canvas); err != nil {
						file.Close()
						t.Fatal(err)
					}
					if err = file.Close(); err != nil {
						t.Fatal(err)
					}
					point["terrain_hd_drawn"] = true
					if art.player != nil {
						point["player_hd"] = art.PlayerState()
					}
					if art.weapon != nil {
						point["weapon_hd"] = art.WeaponState()
					}
					if art.objects != nil {
						point["objects_hd"] = art.ObjectState()
					}
					point["Chinese_overwrites_game_pixels"] = 0
					point["portal_hd"] = art.PortalState()
					point["common_hd"] = art.CommonState()
					point["magia_hd"] = art.MagiaState()
				} else {
					canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
					old := cloneTownBytes(canvas.Pix)
					if art.Draw(canvas, nil, image.Point{}) || !bytes.Equal(old, canvas.Pix) {
						t.Fatal("洞窟HD蓋住非洞窟來源", f[1])
					}
				}
			}
			if strings.HasPrefix(f[1], "return-town") && selected != nil {
				t.Fatal("回城未撤除洞窟來源")
			}
			if !bytes.Equal(before, b.Bytes(oracle.Far(0, 0), 1<<20)) {
				t.Fatal("呈現改寫原版")
			}
			observations = append(observations, point)
		default:
			t.Fatal("未知鍵序", line)
		}
	}
	equal("final")
	if frames.captured == 0 || used == 0 || len(b.Unimplemented()) != 0 || len(b.HeldKeys()) != 0 || len(b.Wrote()) != 0 {
		t.Fatal("正常路徑不完整", frames.captured, used, b.Unimplemented(), b.HeldKeys(), b.Wrote())
	}
	for _, p := range scratchSaves {
		data, err := os.ReadFile(p)
		if err != nil || !bytes.Equal(data, save) {
			t.Fatal("主角覆繪路徑改寫存檔", err)
		}
		fi, err := os.Stat(p)
		if err != nil || !fi.ModTime().Equal(info.ModTime()) {
			t.Fatal("主角覆繪路徑改寫存檔mtime", err)
		}
	}
	playerCoverage := map[string]any{"source_frames": len(seenPlayerFrames), "known_frames": playerKnown, "fully_occluded_frames": playerHidden, "missing_visible_frames": playerMissing, "keys": seenPlayerKeys, "missing_sources": missingPlayerSources, "scope": "正常完成來源盤點；缺素材保留原貌，未稱全動畫HD"}
	data, e := json.MarshalIndent(map[string]any{"points": observations, "hd_draws": hdDraws, "player_coverage": playerCoverage, "weapon_coverage": weaponCoverage.summary(), "object_coverage": objectCoverage.summary(), "save_unchanged": true, "source_save_sha256": saveHash, "max_steps": maxSteps, "commands_sha256": townBackgroundHash(commands), "captured": frames.captured, "refused": frames.refused, "draws": draws, "used": used, "held": held, "cleared_samples": cleared, "state_equal": true, "final": frames.State(), "fileops": b.FileOps(), "keyreads": b.KeyReads(), "wrote": b.Wrote(), "held_keys": b.HeldKeys(), "unimplemented": b.Unimplemented(), "limits": "正常原版與適配層配對；可掛本地地形／主角／武器HD及既有中文；非GUI。沒有注入狀態或固定未定位之亂數種子，不作隨機規則對拍"}, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, '\n'), 0644); e != nil {
		t.Fatal(e)
	}
}

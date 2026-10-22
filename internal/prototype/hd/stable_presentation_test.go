package hd

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
	"github.com/wicanr2/zeliard_cht/internal/prototype/overlay"
)

func TestTownDialogueBackdropAndUnknownOutside(t *testing.T) {
	s, source := townFrameFixture(t)
	box := image.Rect(56, 28, 232, 114)
	for _, outside := range []bool{false, true} {
		f := &TownFrames{scene: s}
		f.finish(source)
		if f.complete == nil {
			t.Fatal("既有正常來源未完成")
		}
		in := copyTownInput(source)
		in.steps++
		in.cycles++
		for y := box.Min.Y; y < box.Max.Y; y++ {
			for x := box.Min.X; x < box.Max.X; x++ {
				i := y*320 + x
				in.indexed[i] = 0
				clear(in.rgb[i*3 : i*3+3])
			}
		}
		before := copyTownInput(in)
		if outside {
			in.indexed[140*320+50] ^= 1
		}
		got := f.selectDialogueFrame(in, box)
		if outside {
			if got != nil {
				t.Fatal("框外未知變動仍沿舊圖")
			}
			continue
		}
		if got == nil || got.DialogueBox != box {
			t.Fatal("已知對話框未保留HD來源")
		}
		if !townFrameEqual(got.Indexed, source.indexed) {
			t.Fatal("未以完整原版背景交給HD層")
		}
		if !bytes.Equal(in.indexed, before.indexed) || !bytes.Equal(in.rgb, before.rgb) {
			t.Fatal("回寫原版對話框")
		}
		canvas := image.NewRGBA(image.Rect(0, 0, 960, 600))
		if !s.Draw(canvas, 0x13, got.Primary, got.Resources, got.Extra, got.Indexed, got.RGB, image.Point{}, true) {
			t.Fatal("對話背景不能通過HD守衛")
		}
	}
}

// 冷F7與既有居民鍵序，覆蓋開框、進字及續頁的正常來源。
func TestNormalTownDialogueBackdropContinuity(t *testing.T) {
	root, save := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_TOWN_SAVE")
	if root == "" || save == "" {
		t.Skip("須提供固定原版與正常存檔")
	}
	scene, _ := townFrameFixture(t)
	o, err := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
	if err != nil {
		t.Fatal(err)
	}
	defer o.Close()
	o.SetDOSBoxCycles(3000)
	scratch := t.TempDir()
	data, err := os.ReadFile(save)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(save)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(scratch, "SAVETEST.usr")
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}
	o.SetScratch(scratch)
	font, err := bitmap.Load(os.Getenv("ZELIARD_FONT"))
	if err != nil {
		t.Fatal(err)
	}
	d, err := overlay.NewTownDialogue(o, filepath.Join(root, "ZELRES2.SAR"), "../../../text/town-prototype.tsv", font)
	if err != nil {
		t.Fatal(err)
	}
	f := NewTownFrames(o, scene)
	commands, err := os.ReadFile("../../../tools/residents-muralla.commands")
	if err != nil {
		t.Fatal(err)
	}
	shown := 0
	run := func(n uint64) {
		for n > 0 {
			step := uint64(25000)
			if n < step {
				step = n
			}
			if err := o.Run(step); err != nil {
				t.Fatal(err)
			}
			n -= step
			raw := o.Indexed()
			_, _, rgb := o.ScreenRGB()
			if d.State()["key"] != "" {
				prior := f.complete
				got := f.PresentDialogue(raw, rgb, d.BackdropBox())
				if !d.BackdropBox().Empty() && got == nil {
					t.Fatalf("正常開框撤除HD: steps=%d prior=%v source=%v box=%v state=%v last=%v", o.Steps(), prior != nil, scene.frameSource(f.read(raw, rgb)), d.BackdropBox(), f.State(), f.lastRefusal)
				}
				if got != nil {
					shown++
				}
			} else {
				f.Present(raw, rgb, false)
			}
		}
	}
	for _, line := range strings.Split(string(commands), "\n") {
		a := strings.Fields(line)
		if len(a) == 0 {
			continue
		}
		number := func(s string) uint64 {
			n, e := strconv.ParseUint(s, 0, 64)
			if e != nil {
				t.Fatal(e)
			}
			return n
		}
		switch a[0] {
		case "run":
			run(number(a[1]))
		case "key":
			scan := uint8(number(a[1]))
			o.KeyDown(scan)
			run(number(a[2]))
			o.KeyUp(scan)
		case "shot":
			if strings.Contains(a[1], "039fff-page1") {
				if shown == 0 {
					t.Fatal("未進入居民對話")
				}
				t.Log("正常居民HD樣本", shown)
				return
			}
		}
	}
	t.Fatal("未到正常居民頁面")
}

func TestCaveFreshBodyKeepsCompletedPresentation(t *testing.T) {
	root, points := caveFixtures(t)
	source, body := caveFixture(t, root, points[20])
	f := &CaveFrames{body: &body}
	f.finish(source)
	f.HoldFreshBody(true)
	if f.complete == nil {
		t.Fatal("既有正常來源被拒絕")
	}
	live := cloneCaveInput(source)
	live.steps += 2
	live.cycles += 2
	fresh := cloneCaveInput(body)
	fresh.steps = source.steps + 1
	fresh.cycles = source.cycles + 1
	f.body = &fresh
	live.indexed[100*320+100] ^= 1
	live.rgb[(100*320+100)*3] ^= 1
	got := f.selectFrame(live, false)
	if got == nil || f.complete == nil || f.State()["pending_body"] != true {
		t.Fatal("可信新主體批次間隙閃回原貌")
	}
	if !caveFrameEqual(got.Indexed, source.indexed) {
		t.Fatal("呈現了未完成viewport")
	}
	if f.selectFrame(live, true) != nil || f.complete != nil {
		t.Fatal("新選項蓋住模態畫面")
	}
}

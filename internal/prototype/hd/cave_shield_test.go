package hd

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
)

func caveShieldRead(t *testing.T, path string) []byte {
	t.Helper()
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	return b
}

func caveShieldStudy(t *testing.T) string {
	t.Helper()
	p := os.Getenv("ZELIARD_SHIELD_STUDY")
	if p == "" {
		t.Skip("須提供027正常原版來源與獨立收據；skip不算驗收")
	}
	return p
}

func caveShieldFixture(t *testing.T, study, name string) *CavePresentation {
	t.Helper()
	base := filepath.Join(study, "cave-shield-source-v1", name)
	var meta struct {
		MemorySHA     string `json:"memory_sha256"`
		BodySHA       string `json:"body_memory_sha256"`
		RawSHA        string `json:"indexed_sha256"`
		PaletteSHA    string `json:"palette_sha256"`
		Steps, Cycles uint64
		BodySteps     uint64 `json:"body_steps"`
		BodyCycles    uint64 `json:"body_cycles"`
		Paints        []CavePaint
	}
	if e := json.Unmarshal(caveShieldRead(t, base+".json"), &meta); e != nil {
		t.Fatal(e)
	}
	m, b, raw, pal := caveShieldRead(t, base+".memory"), caveShieldRead(t, base+".body.memory"), caveShieldRead(t, base+".indexed"), caveShieldRead(t, base+".palette")
	if townBackgroundHash(m) != meta.MemorySHA || townBackgroundHash(b) != meta.BodySHA || townBackgroundHash(raw) != meta.RawSHA || townBackgroundHash(pal) != meta.PaletteSHA {
		t.Fatal("027原版來源身份不符", name)
	}
	rgb := make([]byte, 192000)
	for i, v := range raw {
		copy(rgb[i*3:], pal[int(v)*3:int(v)*3+3])
	}
	d, e := decodeCaveFrameWithPaints(m, b, raw, meta.Paints)
	if e != nil {
		t.Fatal(name, e)
	}
	return &CavePresentation{Memory: m, BodyMemory: b, SourceIndexed: raw, Indexed: cloneTownBytes(raw), RGB: rgb, Palette: pal, Decoded: d, Paints: meta.Paints, Steps: meta.Steps, Cycles: meta.Cycles, BodySteps: meta.BodySteps, BodyCycles: meta.BodyCycles}
}

func caveShieldArt(t *testing.T) *CaveTerrain {
	t.Helper()
	dir := os.Getenv("ZELIARD_CAVE_STUDY")
	art, e := NewCaveTerrain(dir)
	if e != nil {
		t.Fatal(e)
	}
	player, e := NewCavePlayer(dir)
	if e != nil {
		t.Fatal(e)
	}
	art.SetPlayer(player)
	weapon, e := NewCaveWeapon(dir)
	if e != nil {
		t.Fatal(e)
	}
	art.SetWeapon(weapon)
	objects, e := NewCaveObjects(dir)
	if e != nil {
		t.Fatal(e)
	}
	art.SetObjects(objects)
	portal, e := NewCavePortal(filepath.Join(dir, "portal-handpaint-v1.png"), cavePortalFont(t), "Muralla鎮")
	if e != nil {
		t.Fatal(e)
	}
	art.SetPortal(portal)
	return art
}

// 比較53份真實完成候選與事先產生的獨立Python合成，不稱冷重播配對。
func TestCaveShieldIndependentNativeAndHDPixels(t *testing.T) {
	study := caveShieldStudy(t)
	var receipt struct {
		Records []struct {
			Source string
			Key    string `json:"actor_key"`
			Pose   int    `json:"actor_pose"`
		}
		Differences int
	}
	if e := json.Unmarshal(caveShieldRead(t, filepath.Join(study, "cave-shield-inventory-v1/receipt.json")), &receipt); e != nil {
		t.Fatal(e)
	}
	if len(receipt.Records) != 53 || receipt.Differences != 0 {
		t.Fatal("027獨立來源數不符")
	}
	art := caveShieldArt(t)
	seen := map[string]bool{}
	for _, r := range receipt.Records {
		t.Run(r.Source, func(t *testing.T) {
			f := caveShieldFixture(t, study, r.Source)
			if f.Decoded.Differences != 0 || townBackgroundHash(f.Decoded.ActorRaster[:]) != r.Key {
				t.Fatal("獨立原版角色鍵不符")
			}
			wantRaw := caveShieldRead(t, filepath.Join(study, "cave-shield-inventory-v1", fmt.Sprintf("pose-%02d.indexed", r.Pose)))
			if !bytes.Equal(wantRaw, f.Decoded.ActorRaster[:]) {
				t.Fatal("原角色完整色號不符")
			}
			beforeM, beforeB, beforeR := cloneTownBytes(f.Memory), cloneTownBytes(f.BodyMemory), cloneTownBytes(f.SourceIndexed)
			canvas := caveOriginalCanvas(f)
			if !art.Draw(canvas, f, image.Point{}) || art.PlayerState()["matched"] != true || art.PlayerState()["pose"] != 37+r.Pose {
				t.Fatal("持盾HD未匹配", art.PlayerState())
			}
			fp, e := os.Open(filepath.Join(study, "cave-shield-composition-v2", r.Source+"-player-HD.png"))
			if e != nil {
				t.Fatal(e)
			}
			expected, e := png.Decode(fp)
			fp.Close()
			if e != nil {
				t.Fatal(e)
			}
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					if canvas.RGBAAt(x, y) != color.RGBAModel.Convert(expected.At(x, y)).(color.RGBA) {
						t.Fatalf("獨立高清整幅不同 %d,%d", x, y)
					}
				}
			}
			if !bytes.Equal(beforeM, f.Memory) || !bytes.Equal(beforeB, f.BodyMemory) || !bytes.Equal(beforeR, f.SourceIndexed) {
				t.Fatal("HD回寫原版")
			}
			seen[r.Key] = true
		})
	}
	if len(seen) != 17 {
		t.Fatal("17個持盾姿態未齊")
	}
}

// 純負向守衛測試；修改副本不作原版可達狀態證據。
func TestCaveShieldRejectsUnknownProfilesAndPrefixes(t *testing.T) {
	study := caveShieldStudy(t)
	f := caveShieldFixture(t, study, "source-000")
	if !caveFrameSource(caveFrameInput{mode: 0x13, memory: f.Memory, palette: f.Palette}) {
		t.Fatal("已驗完整前綴被拒絕")
	}
	for n := 0; n < 64; n++ {
		m := cloneTownBytes(f.Memory)
		m[0x1b10+0xc000+n] ^= 1
		if caveFrameSource(caveFrameInput{mode: 0x13, memory: m, palette: f.Palette}) {
			t.Fatal("未驗前綴被接納", n)
		}
	}
	for _, off := range []int{0x93, 0xe8, 0xff39, 0xff3a, 0xff41} {
		for _, v := range []byte{1, 3, 4, 255} {
			b := cloneTownBytes(f.BodyMemory)
			b[0x1b10+off] = v
			if _, e := decodeCaveFrameWithPaints(f.Memory, b, f.SourceIndexed, f.Paints); e == nil {
				t.Fatal("未知裝備或組合被接納", off, v)
			}
		}
	}
	art := caveShieldArt(t)
	key := townBackgroundHash(f.Decoded.ActorRaster[:])
	delete(art.player.poses, key)
	canvas := caveOriginalCanvas(f)
	if !art.Draw(canvas, f, image.Point{}) || art.PlayerState()["matched"] != false {
		t.Fatal("缺少姿態未保留原貌")
	}
	bad := *f
	bad.SourceIndexed = cloneTownBytes(f.SourceIndexed)
	bad.SourceIndexed[100*320+100] ^= 128
	canvas = caveOriginalCanvas(f)
	before := cloneTownBytes(canvas.Pix)
	if art.Draw(canvas, &bad, image.Point{}) || !bytes.Equal(canvas.Pix, before) {
		t.Fatal("原圖失配未原子退回")
	}
}

func TestCaveShieldNormalF7FullOriginalPair(t *testing.T) {
	study := caveShieldStudy(t)
	root := os.Getenv("ZELIARD_ORIG")
	project := os.Getenv("ZELIARD_PROJECT")
	if root == "" || project == "" {
		t.Skip("須提供原件與專案；skip不算驗收")
	}
	save := caveShieldRead(t, filepath.Join(root, "SAVEM1.USR"))
	if townBackgroundHash(save) != "03df9abcf62b1712d6b46dd1fc5b23335a4b8711b3dc84875a1498260d37735c" {
		t.Fatal("來源存檔不符")
	}
	info, e := os.Stat(filepath.Join(root, "SAVEM1.USR"))
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
		defer o.Close()
		machines[i] = o
		o.SetDOSBoxCycles(3000)
		o.TraceFiles()
		scratch := t.TempDir()
		dst := filepath.Join(scratch, "SAVEM1.USR")
		if e = os.WriteFile(dst, save, 0644); e != nil {
			t.Fatal(e)
		}
		if e = os.Chtimes(dst, info.ModTime(), info.ModTime()); e != nil {
			t.Fatal(e)
		}
		o.SetScratch(scratch)
		saves = append(saves, dst)
	}
	a, b := machines[0], machines[1]
	frames := NewCaveFrames(b)
	art := caveShieldArt(t)
	checks, draws := 0, 0
	equal := func() {
		t.Helper()
		checks++
		if !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) || a.Regs() != b.Regs() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) {
			t.Fatal("適配層改動完整原版", checks)
		}
	}
	var native struct {
		Points []struct {
			Name          string
			Steps, Cycles uint64
			Memory        string `json:"memory_sha256"`
			Raw           string `json:"indexed_sha256"`
		}
	}
	if e = json.Unmarshal(caveShieldRead(t, filepath.Join(study, "cave-shield-source-v1/receipt.json")), &native); e != nil {
		t.Fatal(e)
	}
	points := map[string]int{}
	for n, p := range native.Points {
		points[p.Name] = n
	}
	number := func(s string) uint64 {
		v, e := strconv.ParseUint(s, 0, 64)
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	run := func(n uint64) {
		if n > 250000000 || a.Steps()+n > 250000000 {
			t.Fatal("原版指令預算")
		}
		for _, o := range machines {
			if e := o.Run(n); e != nil {
				t.Fatal(e)
			}
		}
		equal()
	}
	shots := 0
	execute := func(commands string, startup bool) {
		for _, line := range strings.Split(commands, "\n") {
			f := strings.Fields(line)
			if len(f) == 0 {
				continue
			}
			switch f[0] {
			case "run":
				run(number(f[1]))
			case "key":
				k := byte(number(f[1]))
				for _, o := range machines {
					o.KeyDown(k)
				}
				run(number(f[2]))
				for _, o := range machines {
					o.KeyUp(k)
				}
				equal()
			case "down", "up":
				for _, o := range machines {
					if f[0] == "down" {
						o.KeyDown(byte(number(f[1])))
					} else {
						o.KeyUp(byte(number(f[1])))
					}
				}
				equal()
			case "phase":
			case "shot":
				equal()
				at, ok := points[f[1]]
				if !ok {
					t.Fatal("未知原版觀察點", f[1])
				}
				p := native.Points[at]
				if a.Steps() != p.Steps || a.Cycles() != p.Cycles || townBackgroundHash(a.Bytes(oracle.Far(0, 0), 1<<20)) != p.Memory || townBackgroundHash(a.Indexed()) != p.Raw {
					t.Fatal("純原版觀察不同", f[1])
				}
				shots++
				_, _, rgb := b.ScreenRGB()
				if selected := frames.Present(b.Indexed(), rgb, false); selected != nil {
					if !art.Draw(caveOriginalCanvas(selected), selected, image.Point{}) {
						t.Fatal("正常完成來源HD被拒絕", f[1])
					}
					draws++
				}
				if startup && f[1] == "cave-entered" {
					return
				}
			default:
				t.Fatal("未驗命令", f[0])
			}
		}
	}
	execute(string(caveShieldRead(t, filepath.Join(project, "tools/sabre-source-use.commands"))), true)
	execute(string(caveShieldRead(t, filepath.Join(study, "cave_shield_actions_v1.commands"))), false)
	if shots != 15 || draws == 0 || frames.captured != 172 || frames.refused != 0 {
		t.Fatal("正常持盾路徑未齊", shots, draws, frames.captured, frames.refused, frames.lastRefusal)
	}
	for _, o := range machines {
		if len(o.HeldKeys()) != 0 || len(o.Wrote()) != 0 || len(o.Unimplemented()) != 0 {
			t.Fatal("原版執行或放鍵未齊")
		}
	}
	for _, dst := range saves {
		after, e := os.Stat(dst)
		if e != nil || !after.ModTime().Equal(info.ModTime()) || !bytes.Equal(caveShieldRead(t, dst), save) {
			t.Fatal("來源存檔或mtime改變")
		}
	}
	t.Logf("027正常冷F7：%d完整原版配對、%d觀察點、%d呈現、%d完成畫格、零拒絕；來源取得未知", checks, shots, draws, frames.captured)
}

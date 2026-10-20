package hd

import (
	"bytes"
	"encoding/json"
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

type merchantSource struct {
	Key    string   `json:"key"`
	Pose   int      `json:"pose"`
	Normal []string `json:"normal_sources"`
}
type merchantBinding struct {
	Folder string `json:"folder"`
	File   string `json:"file"`
}
type merchantLayout struct {
	Frames []merchantSource  `json:"frames"`
	Normal []merchantBinding `json:"normal_bindings"`
	Mixed  []merchantBinding `json:"normal_mixed_rejections"`
	Guards []merchantBinding `json:"normal_source_guard_rejections"`
}

func merchantFramesInputs(t *testing.T) (*MerchantFrames, string, merchantLayout) {
	t.Helper()
	project := os.Getenv("ZELIARD_PROJECT")
	if project == "" || os.Getenv("ZELIARD_ORIG") == "" {
		t.Skip("須掛載武器店原版及完整圖組，skip不算驗收")
	}
	s, e := NewMerchantFrames(os.Getenv("ZELIARD_ORIG"), filepath.Join(project, "workplace/full-hd-v1/merchant-frames-complete-v2"))
	if e != nil {
		t.Fatal(e)
	}
	var layout merchantLayout
	if e = json.Unmarshal(interiorRead(t, filepath.Join(project, "workplace/full-hd-v1/merchant-frames-layout-v1/receipt.json")), &layout); e != nil {
		t.Fatal(e)
	}
	if len(layout.Frames) != 10 {
		t.Fatal("來源分母錯誤")
	}
	return s, project, layout
}

func merchantFrameSample(t *testing.T, project, relative string) ([]byte, []byte, []byte, []byte) {
	t.Helper()
	base := filepath.Join(project, "workplace/full-hd-v1", relative)
	mem := interiorRead(t, base+".memory")
	raw := interiorRead(t, base+".indexed")
	palette := interiorRead(t, base+".palette")
	rgb, _ := skyOriginal(raw, palette)
	return mem, raw, rgb, palette
}

func merchantFramesReceipt(t *testing.T, name string, value any) {
	t.Helper()
	if out := os.Getenv("ZELIARD_MERCHANT_OUTPUT"); out != "" {
		b, e := json.MarshalIndent(value, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(out, name), append(b, 10), 0644); e != nil {
			t.Fatal(e)
		}
	}
}

func TestMerchantReturnTownHDMaskFootprint(t *testing.T) {
	scene, base := sceneryTestInputs(t)
	_, project, background := backgroundTestInputs(t)
	root := filepath.Join(project, "workplace/full-hd-v1/merchant-frames-normal-v6")
	memory := interiorRead(t, filepath.Join(root, "point-015.memory"))
	raw := interiorRead(t, filepath.Join(root, "point-015.indexed"))
	palette := interiorRead(t, filepath.Join(root, "point-015.palette"))
	before := bytes.Clone(memory)
	p, r, extra := memory[0x1b10:0x11b10], memory[0x11b10:0x21b10], memory[0x21b10:0x31b10]
	if r[0xd794] != 41 || townBackgroundHash(r[0xd000:0xd780]) != "433de06e8f07c80b1caeea08f55059acda00c6126971bff1745e03801bd872fa" {
		t.Fatal("正常離店來源不符")
	}
	rgb, canvas := skyOriginal(raw, palette)
	ground, err := NewTownGround(scene, filepath.Join(base, "normal-v1/restored.indexed"), filepath.Join(base, "ground-handpaint-v1.png"))
	if err != nil {
		t.Fatal(err)
	}
	player, err := NewTownPlayer(scene, base)
	if err != nil {
		t.Fatal(err)
	}
	check := func() {
		t.Helper()
		_, canvas = skyOriginal(raw, palette)
		if !background.Draw(canvas, 0x13, p, r, raw, rgb, image.Point{}) ||
			!scene.Draw(canvas, 0x13, p, r, extra, raw, rgb, image.Point{}, true) ||
			!ground.Draw(canvas, 0x13, p, r, extra, raw, rgb, palette, image.Point{}) ||
			!player.Draw(canvas, 0x13, p, r, extra, raw, rgb, palette, image.Point{}) {
			t.Fatal("正常離店後四層HD未恢復")
		}
		frames := &TownFrames{scene: scene}
		frames.finish(townFrameInput{0x13, p, r, extra, raw, rgb, palette, 1, 1})
		if frames.complete == nil {
			t.Fatal("正常離店完整來源被拒絕")
		}
	}
	check()
	if !bytes.Equal(memory, before) {
		t.Fatal("覆繪改寫原版")
	}
	art := bytes.Clone(canvas.Pix)
	for i := 0xd780; i < 0xd800; i++ {
		r[i] ^= 0xa5
	}
	check()
	if !bytes.Equal(art, canvas.Pix) {
		t.Fatal("未使用遮罩尾端改變圖面")
	}
	r[0xd77f] ^= 1
	if background.Draw(canvas, 0x13, p, r, raw, rgb, image.Point{}) ||
		scene.Draw(canvas, 0x13, p, r, extra, raw, rgb, image.Point{}, true) || !bytes.Equal(art, canvas.Pix) {
		t.Fatal("有效遮罩最後位元組未原子拒絕")
	}
	merchantFramesReceipt(t, "return-town.json", map[string]any{
		"status": "PASS_NORMAL_MERCHANT_RETURN_TOWN_FOUR_HD_LAYERS", "memory_sha256": townBackgroundHash(before),
		"indexed_sha256": townBackgroundHash(raw), "effective_mask_sha256": townBackgroundHash(before[0x11b10+0xd000 : 0x11b10+0xd780]),
		"outside_mask_bytes": 128, "inputs_unchanged": true, "tail_variation_pixels_equal": true, "live_mask_boundary_rejected": true,
	})
}

func TestMerchantFramesAllSourcesAndPixels(t *testing.T) {
	s, project, layout := merchantFramesInputs(t)
	var rows []map[string]any
	prior, e := NewMerchant(filepath.Join(project, "workplace/inventory-source-v1/greeting.indexed"), filepath.Join(project, "workplace/inventory-source-v1/farewell.indexed"), filepath.Join(project, "workplace/merchant-path-v1/no-purchase/no-purchase-farewell.indexed"), filepath.Join(project, "workplace/music-release-v1/release-build/data/assets/town/frame/palette.bin"), filepath.Join(project, "workplace/merchant-hd-v1/merchant-prototype.png"), filepath.Join(project, "workplace/merchant-hd-v1/merchant-farewell-prototype.png"), filepath.Join(project, "workplace/merchant-hd-v1/merchant-no-purchase-prototype.png"))
	if e != nil {
		t.Fatal(e)
	}
	legacy := 0
	for _, row := range layout.Frames {
		t.Run(row.Key, func(t *testing.T) {
			if len(row.Normal) == 0 {
				t.Fatal("缺少正常來源", row.Key)
			}
			mem, raw, rgb, pal := merchantFrameSample(t, project, row.Normal[0])
			beforeMem, beforeRaw, beforeRGB, beforePal := bytes.Clone(mem), bytes.Clone(raw), bytes.Clone(rgb), bytes.Clone(pal)
			dst := interiorCanvas(t, rgb)
			before := bytes.Clone(dst.Pix)
			if !s.Draw(dst, 0x13, mem[0x1b10:0x11b10], mem[0x11b10:0x21b10], raw, rgb, pal, image.Pt(0, 48)) || s.State()["key"] != row.Key || s.State()["pose"] != row.Pose {
				t.Fatal("正常來源或姿態失配", row.Key, s.State())
			}
			art, e := png.Decode(bytes.NewReader(interiorRead(t, filepath.Join(project, "workplace/full-hd-v1/merchant-frames-complete-v2", row.Key+"-hd.png"))))
			if e != nil {
				t.Fatal(e)
			}
			pixels := 0
			box := image.Rect(171, 120, 453, 306)
			for y := 0; y < 680; y++ {
				for x := 0; x < 960; x++ {
					at := dst.PixOffset(x, y)
					if !image.Pt(x, y).In(box) {
						if !bytes.Equal(before[at:at+4], dst.Pix[at:at+4]) {
							t.Fatal("改寫區外或原版外框", x, y)
						}
						continue
					}
					if dst.RGBAAt(x, y) != color.RGBAModel.Convert(art.At(x-171, y-120)).(color.RGBA) {
						t.Fatal("直接3倍圖面失配", row.Key, x, y)
					}
					pixels++
				}
			}
			if !bytes.Equal(mem, beforeMem) || !bytes.Equal(raw, beforeRaw) || !bytes.Equal(rgb, beforeRGB) || !bytes.Equal(pal, beforePal) {
				t.Fatal("改寫原版來源")
			}
			if row.Key == "body-0" || row.Key == "body-2" {
				old := interiorCanvas(t, rgb)
				if ok, _ := prior.Draw(old, raw, rgb, image.Pt(0, 48)); !ok || !bytes.Equal(dst.Pix, old.Pix) {
					t.Fatal("已驗兩基底HD圖面漂移", row.Key)
				}
				legacy++
			}
			if out := os.Getenv("ZELIARD_MERCHANT_OUTPUT"); out != "" {
				f, e := os.Create(filepath.Join(out, row.Key+"-hd.png"))
				if e != nil {
					t.Fatal(e)
				}
				if e = png.Encode(f, dst); e != nil {
					t.Fatal(e)
				}
				f.Close()
			}
			rows = append(rows, map[string]any{"key": row.Key, "pose": row.Pose, "drawn_pixels": pixels, "normal_source": row.Normal[0], "input_unchanged": true})
		})
	}
	if legacy != 2 {
		t.Fatal("兩個舊基底未驗")
	}
	merchantFramesReceipt(t, "render.json", map[string]any{"status": "PASS_MERCHANT_NORMAL_SOURCES_AND_DIRECT_PIXELS", "frames": rows, "legacy_bases_preserved": legacy})
}

func TestMerchantFramesAllNativeRejections(t *testing.T) {
	s, project, layout := merchantFramesInputs(t)
	accepted, rejected := 0, 0
	for _, group := range []struct {
		rows []merchantBinding
		want bool
	}{{layout.Normal, true}, {layout.Mixed, false}, {layout.Guards, false}} {
		for _, row := range group.rows {
			mem, raw, rgb, pal := merchantFrameSample(t, project, row.Folder+"/"+row.File)
			dst := image.NewRGBA(image.Rect(0, 0, 960, 680))
			before := bytes.Clone(dst.Pix)
			got := s.Draw(dst, 0x13, mem[0x1b10:0x11b10], mem[0x11b10:0x21b10], raw, rgb, pal, image.Pt(0, 48))
			if got != group.want {
				t.Fatal("正常或混合來源判定不符", row, got)
			}
			if !got && !bytes.Equal(dst.Pix, before) {
				t.Fatal("來源失配沒有原子拒絕", row)
			}
			if got {
				accepted++
			} else {
				rejected++
			}
		}
	}
	if accepted != 21 || rejected != 122 {
		t.Fatal("正常來源分母不符", accepted, rejected)
	}
	for _, bad := range []string{"mode", "length", "module", "graphics", "palette", "indexed", "rgb", "canvas", "art"} {
		t.Run(bad, func(t *testing.T) {
			mem, raw, rgb, pal := merchantFrameSample(t, project, layout.Frames[0].Normal[0])
			p, r := mem[0x1b10:0x11b10], mem[0x11b10:0x21b10]
			dst := image.NewRGBA(image.Rect(0, 0, 960, 680))
			mode := byte(0x13)
			switch bad {
			case "mode":
				mode = 3
			case "length":
				raw = raw[:32]
			case "module":
				p[0xa000] ^= 1
			case "graphics":
				r[0x8000] ^= 1
			case "palette":
				pal[0] ^= 1
			case "indexed":
				raw[24*320+57] ^= 1
			case "rgb":
				rgb[(24*320+57)*3] ^= 1
			case "canvas":
				dst = image.NewRGBA(image.Rect(0, 0, 100, 100))
			case "art":
				old := s.art["body-0"]
				s.art["body-0"] = nil
				defer func() { s.art["body-0"] = old }()
			}
			before := bytes.Clone(dst.Pix)
			if s.Draw(dst, mode, p, r, raw, rgb, pal, image.Pt(0, 48)) || !bytes.Equal(dst.Pix, before) {
				t.Fatal("失配未原子拒絕", bad)
			}
		})
	}
	mem, raw, rgb, pal := merchantFrameSample(t, project, layout.Frames[0].Normal[0])
	for at := 0xbbfd; at < 0xbc47; at++ {
		mem[0x1b10+at] ^= 0x55
	}
	if !s.Draw(image.NewRGBA(image.Rect(0, 0, 960, 680)), 0x13, mem[0x1b10:0x11b10], mem[0x11b10:0x21b10], raw, rgb, pal, image.Pt(0, 48)) {
		t.Fatal("可寫商品資料阻擋完整來源")
	}
	merchantFramesReceipt(t, "native-sources.json", map[string]any{"status": "PASS_MERCHANT_NATIVE_MATCH_AND_ATOMIC_REJECTIONS", "accepted": accepted, "mixed_rejected": len(layout.Mixed), "source_guard_rejected": len(layout.Guards)})
}

func TestMerchantFramesMissingAssetGroup(t *testing.T) {
	_, project, _ := merchantFramesInputs(t)
	dir := t.TempDir()
	source := filepath.Join(project, "workplace/full-hd-v1/merchant-frames-complete-v2")
	entries, e := os.ReadDir(source)
	if e != nil {
		t.Fatal(e)
	}
	for _, file := range entries {
		if file.Name() == "manifest.json" || strings.HasPrefix(file.Name(), "body-") {
			if e = os.WriteFile(filepath.Join(dir, file.Name()), interiorRead(t, filepath.Join(source, file.Name())), 0644); e != nil {
				t.Fatal(e)
			}
		}
	}
	if e = os.WriteFile(filepath.Join(dir, "body-7-hd.png"), []byte("corrupt"), 0644); e != nil {
		t.Fatal(e)
	}
	if s, e := NewMerchantFrames(os.Getenv("ZELIARD_ORIG"), dir); e == nil || s != nil {
		t.Fatal("缺損圖組未整組拒絕")
	}
}

func TestMerchantFramesNormalBranchesReadOnly(t *testing.T) {
	s, project, _ := merchantFramesInputs(t)
	savePath := filepath.Join(project, "workplace/prose-v1/npc-normal-v2/scratch/SAVETEST.usr")
	save := interiorRead(t, savePath)
	info, e := os.Stat(savePath)
	if e != nil {
		t.Fatal(e)
	}
	if townBackgroundHash(save) != "78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb" {
		t.Fatal("自建USR身份")
	}
	var results []map[string]any
	allPoses := map[int]bool{}
	for _, branch := range []struct {
		key, file string
		gold      uint16
	}{{"purchase", "merchant-hd-purchase.commands", 800}, {"departure", "merchant-hd-departure.commands", 1000}} {
		t.Run(branch.key, func(t *testing.T) {
			var machines [2]*oracle.Oracle
			var scratch [2]string
			for i := range machines {
				o, e := oracle.Load(filepath.Join(os.Getenv("ZELIARD_ORIG"), "ZELIARD.EXE"), os.Getenv("ZELIARD_ORIG"))
				if e != nil {
					t.Fatal(e)
				}
				defer o.Close()
				o.SetDOSBoxCycles(3000)
				scratch[i] = t.TempDir()
				path := filepath.Join(scratch[i], "SAVETEST.usr")
				if e = os.WriteFile(path, save, 0644); e != nil {
					t.Fatal(e)
				}
				if e = os.Chtimes(path, info.ModTime(), info.ModTime()); e != nil {
					t.Fatal(e)
				}
				o.SetScratch(scratch[i])
				machines[i] = o
			}
			a, b := machines[0], machines[1]
			initial := townBackgroundHash(a.Bytes(oracle.Far(0, 0), 1<<20))
			if initial != townBackgroundHash(b.Bytes(oracle.Far(0, 0), 1<<20)) {
				t.Fatal("原版初始狀態不同")
			}
			canvas := image.NewRGBA(image.Rect(0, 0, 960, 680))
			seen := map[string]bool{}
			poses := map[int]bool{}
			var points []map[string]any
			sample := func(name string, force bool) {
				p, r := b.Bytes(oracle.Far(0x1b1, 0), 65536), b.Bytes(oracle.Far(0x11b1, 0), 65536)
				raw := b.Indexed()
				_, _, rgb := b.ScreenRGB()
				pal := []byte{}
				for _, c := range b.Palette() {
					pal = append(pal, c[:]...)
				}
				matched := s.Draw(canvas, b.VideoMode(), p, r, raw, rgb, pal, image.Pt(0, 48))
				key := ""
				if matched {
					key = s.State()["key"].(string)
					poses[s.State()["pose"].(int)] = true
					allPoses[s.State()["pose"].(int)] = true
				}
				if !force && (!matched || seen[key]) {
					return
				}
				seen[key] = true
				mem := b.Bytes(oracle.Far(0, 0), 1<<20)
				if !bytes.Equal(mem, a.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || a.Regs() != b.Regs() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) {
					t.Fatal("原版完整狀態不同", name, key)
				}
				points = append(points, map[string]any{"name": name, "key": key, "matched": matched, "state_equal": true, "steps": b.Steps(), "cycles": b.Cycles(), "memory_sha256": townBackgroundHash(mem)})
			}
			run := func(n uint64) {
				for n > 0 {
					step := uint64(25000)
					if n < step {
						step = n
					}
					for _, o := range machines {
						if e := o.Run(step); e != nil {
							t.Fatal(e)
						}
					}
					n -= step
					sample("animation", false)
				}
			}
			num := func(s string) uint64 {
				n, e := strconv.ParseUint(s, 0, 64)
				if e != nil {
					t.Fatal(e)
				}
				return n
			}
			commands := interiorRead(t, filepath.Join(project, "tools", branch.file))
			for _, line := range strings.Split(string(commands), "\n") {
				f := strings.Fields(line)
				if len(f) == 0 || strings.HasPrefix(f[0], "#") {
					continue
				}
				switch f[0] {
				case "run":
					run(num(f[1]))
				case "key":
					for _, o := range machines {
						o.KeyDown(uint8(num(f[1])))
					}
					run(num(f[2]))
					for _, o := range machines {
						o.KeyUp(uint8(num(f[1])))
					}
				case "shot":
					sample(f[1], true)
				default:
					t.Fatal("未知鍵序", line)
				}
			}
			if len(b.HeldKeys()) != 0 || b.Word(oracle.Far(0x1b1, 0x86)) != branch.gold || s.State()["matched"] != false {
				t.Fatal("正常金額、離店或持鍵不符", branch.key, s.State())
			}
			for _, folder := range scratch {
				path := filepath.Join(folder, "SAVETEST.usr")
				current, e := os.Stat(path)
				if e != nil || current.ModTime() != info.ModTime() || !bytes.Equal(save, interiorRead(t, path)) {
					t.Fatal("存檔或mtime改變", e)
				}
			}
			if !bytes.Equal(save, interiorRead(t, savePath)) {
				t.Fatal("來源USR改變")
			}
			results = append(results, map[string]any{"branch": branch.key, "points": points, "poses": poses, "gold": branch.gold, "save_and_mtime_unchanged": true, "commands_sha256": townBackgroundHash(commands), "initial_memory_sha256": initial, "randomness": "不重擲；獨立冷開機初始完整RAM相同，原版亂數及虛擬時間保持原樣"})
		})
	}
	if len(allPoses) != 8 {
		t.Fatal("正常姿態抽樣不足", allPoses)
	}
	merchantFramesReceipt(t, "normal.json", map[string]any{"status": "PASS_MERCHANT_TWO_NORMAL_BRANCHES_READ_ONLY", "branches": results, "poses": allPoses, "source_save_sha256": townBackgroundHash(save), "scope": "正常冷F7、購買與未購物離店；非GUI或其他城鎮劇情驗收"})
}

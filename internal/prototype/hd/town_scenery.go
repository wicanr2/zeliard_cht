// Muralla前景覆繪的來源、失配與文字契約見docs/spec/007-town-scenery-hd.md。
package hd

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

type townSceneSign struct {
	worldX, width int
	text          string
}

// Muralla完整canonical地圖及原生讀取足跡見規格022。
const townSceneTerrainTiles = 240

type TownScenery struct {
	tiles         [215][8]byte
	palette       []byte
	art           *image.RGBA
	font          bitmap.Font
	signs         []townSceneSign
	castle        *castleProfile
	castleVariant *TownScenery
}

func townSceneWord(b []byte, p int) int { return int(binary.LittleEndian.Uint16(b[p:])) }

func NewTownScenery(root, memoryRef, palettePath, artDir, signsPath string, font bitmap.Font) (*TownScenery, error) {
	for name, digest := range map[string]string{
		"ZELIARD.EXE": "fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341",
		"ZELRES1.SAR": "56a73b59e952dab47fb0171a3fb1cdb7d641baceafa97e2826057f877a590eb0",
		"ZELRES2.SAR": "bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1",
	} {
		if _, err := checked(filepath.Join(root, name), digest, 1<<20); err != nil {
			return nil, err
		}
	}
	memory, err := checked(memoryRef, "9dd3657107f25c097debf5f6be68137d055d98bda2e22a59f6b854d88be5fdd9", 1<<20)
	if err != nil || len(memory) != 1<<20 {
		return nil, fmt.Errorf("前景參考記憶體不符：%v", err)
	}
	palette, err := checked(palettePath, "6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c", 768)
	if err != nil {
		return nil, err
	}
	s := &TownScenery{palette: palette, font: font}
	p := memory[0x1b10:0x11b10]
	for x := range s.tiles {
		for y := range s.tiles[x] {
			v := p[0xc017+x*8+y]
			if v == 253 {
				found := false
				for pos := townSceneWord(p, 0xc00f); pos+8 <= len(p); pos += 8 {
					if townSceneWord(p, pos) == 65535 {
						break
					}
					if townSceneWord(p, pos) == x {
						v = p[pos+3]
						found = true
						break
					}
				}
				if !found {
					return nil, fmt.Errorf("參考底圖缺項")
				}
			}
			if v >= townSceneTerrainTiles {
				return nil, fmt.Errorf("參考底圖未支援")
			}
			s.tiles[x][y] = v
		}
	}
	panels, err := townScenePanels(artDir)
	if err != nil {
		return nil, err
	}
	s.art = image.NewRGBA(image.Rect(0, 0, 5160, 192))
	for _, p := range panels {
		for y := 0; y < 192; y++ {
			for x := p.start * 3; x < p.end*3; x++ {
				s.art.Set(x, y, p.sample(x, y))
			}
		}
	}
	if signsPath == "" {
		return s, nil
	}
	file, err := os.Open(signsPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) != 4 || strings.Join(records[0], "/") != "key/world_x/source/zh_TW" {
		return nil, fmt.Errorf("招牌TSV欄位不符")
	}
	expected := []struct {
		key, source string
		x           int
	}{{"weapon", "WEAPON", 321}, {"magic", "MAGIC", 897}, {"bank", "BANK", 1113}}
	for i, want := range expected {
		r := records[i+1]
		if len(r) != 4 {
			return nil, fmt.Errorf("招牌欄數不符")
		}
		x, e := strconv.Atoi(r[1])
		if e != nil || x != want.x || r[0] != want.key || r[2] != want.source {
			return nil, fmt.Errorf("招牌來源不符")
		}
		width := 0
		if !utf8.ValidString(r[3]) || strings.TrimSpace(r[3]) != r[3] {
			return nil, fmt.Errorf("招牌須使用完整UTF-8單行文字")
		}
		for _, c := range r[3] {
			if unicode.IsControl(c) {
				return nil, fmt.Errorf("招牌不能含控制字元")
			}
			glyph := font[c]
			if len(glyph) != 16 && len(glyph) != 32 {
				return nil, fmt.Errorf("招牌字缺：%U", c)
			}
			width += len(glyph) / 16 * 8
		}
		if width == 0 || width > 88 {
			return nil, fmt.Errorf("招牌文字超出安全框")
		}
		s.signs = append(s.signs, townSceneSign{x, width, r[3]})
	}
	return s, nil
}

// Draw先驗原版完整畫格。失配時不寫畫布；不修改dosgolem或輸入。
func (s *TownScenery) Draw(dst *image.RGBA, mode uint8, primary, resources, extra, indexed, rgb []byte, origin image.Point, chinese bool) bool {
	box := image.Rect(144, 234, 816, 426).Add(origin)
	if !box.In(dst.Bounds()) {
		return false
	}
	f, err := s.decode(mode, primary, resources, extra, indexed, rgb)
	if err != nil {
		return false
	}
	owner := f.scene
	var flags []castleFlagPaint
	if owner.castle != nil {
		var ok bool
		flags, ok = owner.castle.prepareFlags(dst, mode, primary, resources, indexed, rgb, f, origin)
		if !ok {
			return false
		}
		owner.castle.paintBackground(dst, flags)
	}
	for y := 0; y < 192; y++ {
		for x := 0; x < 672; x++ {
			i := (y/3)*224 + x/3
			if !f.foreground[i] || f.opaque[i] || len(f.flags) != 0 && f.flags[i] != 0 {
				continue
			}
			worldX := (f.camera+4)*24 + x
			c := owner.art.RGBAAt(worldX, y)
			if c.A == 0 {
				continue
			}
			px, py := box.Min.X+x, box.Min.Y+y
			old := dst.RGBAAt(px, py)
			inv := uint32(255 - c.A)
			dst.SetRGBA(px, py, color.RGBA{uint8(uint32(c.R) + uint32(old.R)*inv/255), uint8(uint32(c.G) + uint32(old.G)*inv/255), uint8(uint32(c.B) + uint32(old.B)*inv/255), 255})
		}
	}
	for _, flag := range flags {
		draw.Draw(dst, flag.box, flag.art, image.Point{}, draw.Over)
	}
	if chinese {
		for _, sign := range owner.signs {
			cx := box.Min.X + (sign.worldX-(f.camera+4)*8)*3
			safe := image.Rect(cx-48, box.Min.Y+102, cx+48, box.Min.Y+120)
			if !safe.In(box) {
				continue
			}
			draw.Draw(dst, safe, image.NewUniform(color.RGBA{47, 30, 18, 255}), image.Point{}, draw.Src)
			text := image.Rect(cx-sign.width/2, safe.Min.Y, cx-sign.width/2+sign.width, safe.Max.Y)
			owner.font.Draw(dst, text, sign.text, color.RGBA{255, 231, 160, 255})
		}
	}
	return true
}

type townSceneActor struct {
	x, group, phase, under int
	direction              bool
}
type townSceneFrame struct {
	camera             int
	foreground, opaque []bool
	base, decoded      []byte
	flags              []byte
	scene              *TownScenery
}

func townSceneTile(b []byte, off int) [64]byte {
	var out [64]byte
	for i := 0; i < 16; i++ {
		a, c, d := b[off+i*3], b[off+i*3+1], b[off+i*3+2]
		copy(out[i*4:], []byte{c >> 2, ((c & 3) << 4) | (a >> 4), ((a & 15) << 2) | (d >> 6), d & 63})
	}
	return out
}

// 只排除已驗人物及地形不讀取的保留區，封住未知group與ID。
// 正常重讀見規格007；Muralla地形0..239的完整足跡見規格022。
func townSceneSources(p, r, e []byte) ([]townSceneActor, error) {
	if len(p) != 65536 || len(r) != 65536 || len(e) != 65536 {
		return nil, fmt.Errorf("來源段長度不符")
	}
	guards := []struct {
		b []byte
		h string
	}{
		{p[0x3000:0x3628], "73f650db23e2c2fe32c6a218ff5ab57d26f7c2200faab027003827e0575f67c7"},
		{p[0x6975:0x6a77], "ca30aab0ce543dd922095b669255fe1a76dbcc5256992e1ee5b437fb91925182"},
		{p[0x6950:0x6975], "8a7c0b5eaab587ccec33a8cf20c84aa0219e72da794ac9286993b134694090f5"},
		{r[0x4000:0x5d80], "396a3b47dd240301fae669a494230194944987e2c46e32eaf35baa8e50d36825"},
		{r[0x6000:0x68a0], "045d4246ae0be72df3ac0201bc9ac8e5e1b50829783d6c79727fce7af0ef66c7"},
		{r[0x8000:0xae00], "6cddbaa678370307422a6a60798c0d75a53c46deacc1a9f299a716c40d234b45"},
		{r[0xd000:0xd800], "0bdb78b2702490f3309f72ea3046fda2417bd088a9a6eec1c3f179fcad4cef71"},
		{e[0x7000:0x74c0], "90de4921fb15fa26b9214b33e5c89daaa65106984db0f7417b3bd09f384392ce"},
		{e[0x8000:0x8800], "2c7708d6978dbd5661612bf36d921b12facb3f5aebe283012f2baedc7603e431"},
		{p[0xc000:0xc040], "025bf2982660c942cbefed303ccdd0523ff824756a0fe0bb2587b99e0931e7ec"},
	}
	for i, g := range guards {
		if townBackgroundHash(g.b) != g.h {
			return nil, fmt.Errorf("source guard %d", i)
		}
	}
	return townSceneActorRecords(p)
}

func townSceneActorRecords(p []byte) ([]townSceneActor, error) {
	if len(p) != 65536 {
		return nil, fmt.Errorf("居民記錄段長度不符")
	}
	actors := []townSceneActor{}
	for pos := townSceneWord(p, 0xc00f); ; pos += 8 {
		if pos < 0 || pos+8 > len(p) || len(actors) > 256 {
			return nil, fmt.Errorf("townSceneActor table bounds")
		}
		if townSceneWord(p, pos) == 65535 {
			break
		}
		if p[pos+2]&127 >= 5 {
			return nil, fmt.Errorf("unknown townSceneActor group")
		}
		actors = append(actors, townSceneActor{townSceneWord(p, pos), int(p[pos+2] & 127), int(p[pos+4] & 3), int(p[pos+3]), p[pos+2]&128 != 0})
	}
	return actors, nil
}

// 規格038：正常進門phase4的居民標記均已還原，記錄只供底圖核對。
// 其他phase4配置未驗，不能由居民表仍存在就補畫人物。
func townScenePaintActors(p []byte, actors []townSceneActor) ([]townSceneActor, error) {
	if len(p) != 65536 {
		return nil, fmt.Errorf("居民呈現來源長度不符")
	}
	if p[0xe7] != 4 {
		return actors, nil
	}
	for _, actor := range actors {
		pos := 0xc01c + actor.x*8
		if pos < 0xc01c || pos >= len(p) || actor.under >= 253 || int(p[pos]) != actor.under {
			return nil, fmt.Errorf("未驗背向居民地圖配置")
		}
	}
	return nil, nil
}

func (s *TownScenery) decode(mode uint8, p, r, e, raw, rgb []byte) (*townSceneFrame, error) {
	f, err := s.decodeOwn(mode, p, r, e, raw, rgb)
	if err != nil && s.castleVariant != nil {
		if _, sourceError := s.ownSources(p, r, e); sourceError != nil {
			return s.castleVariant.decodeOwn(mode, p, r, e, raw, rgb)
		}
	}
	return f, err
}

func (s *TownScenery) decodeOwn(mode uint8, p, r, e, raw, rgb []byte) (*townSceneFrame, error) {
	if mode != 0x13 || len(raw) != 64000 || len(rgb) != 192000 {
		return nil, fmt.Errorf("顯示輸入不符")
	}
	actors, err := s.ownSources(p, r, e)
	if err != nil {
		return nil, err
	}
	camera, mp := townSceneWord(p, 0x80), townSceneWord(p, 0xff2a)
	width, terrain := 215, townSceneTerrainTiles
	if s.castle != nil {
		width, terrain = 114, 157
	}
	if townSceneWord(p, 0xff2c) != 0x11b1 || townSceneWord(r, 0x8000) != 0x8006 || mp-camera*8 != 0xc017 || townSceneWord(p, 0xc002) != width || camera+31 >= width {
		return nil, fmt.Errorf("map geometry")
	}
	under := func(wx int) (int, error) {
		for _, a := range actors {
			if a.x == wx {
				if a.under >= terrain {
					return 0, fmt.Errorf("unverified nested townSceneActor underlay")
				}
				return a.under, nil
			}
		}
		return 0, fmt.Errorf("missing townSceneActor underlay")
	}
	f := &townSceneFrame{camera: camera, foreground: make([]bool, 14336), opaque: make([]bool, 14336), base: make([]byte, 14336), scene: s}
	if s.castle != nil {
		f.flags = make([]byte, 14336)
	}
	for col := 0; col < 28; col++ {
		for row := 0; row < 8; row++ {
			id := int(p[0xe000+col*8+row])
			if id >= 253 {
				id = int(p[mp+32+col*8+row])
				if id == 253 {
					var err error
					id, err = under(camera + 4 + col)
					if err != nil {
						return nil, err
					}
				}
			}
			canonical := id
			if s.castle != nil && id >= 2 && id <= 5 {
				canonical = 2
			}
			if id >= terrain || canonical != int(s.tiles[camera+4+col][row]) {
				return nil, fmt.Errorf("townSceneTile source mismatch %d %d", col, row)
			}
			pixels := townSceneTile(r, 0x8100+id*48)
			for dy := 0; dy < 8; dy++ {
				for dx := 0; dx < 8; dx++ {
					x, y := col*8+dx, row*8+dy
					i := y*224 + x
					if s.castle != nil && row < 3 && id >= 2 && id <= 5 {
						f.flags[i] = byte(id)
					}
					bg := row < 3 && r[0x8006+id] != 0 && r[0xd000+id*8+dy]&(0x80>>dx) != 0
					f.foreground[i] = !bg
					if bg {
						f.base[i] = p[0xa000+(x/8)*192+y*8+x%8]
					} else {
						f.base[i] = pixels[dy*8+dx]
					}
				}
			}
		}
	}
	f.decoded = append([]byte(nil), f.base...)
	sprite := func(x int, ids []byte, atlas, masks int) error {
		for col := 0; col < 2; col++ {
			for row := 0; row < 3; row++ {
				id := int(ids[col*3+row])
				pixels := townSceneTile(r, atlas+id*48)
				for dy := 0; dy < 8; dy++ {
					for dx := 0; dx < 8; dx++ {
						px, py := x+col*8+dx, 40+row*8+dy
						if px < 0 || px >= 224 {
							continue
						}
						i := py*224 + px
						keep := e[masks+id*8+dy]&(0x80>>dx) != 0
						v := pixels[dy*8+dx]
						if keep && v != 0 {
							return fmt.Errorf("unverified OR blend")
						}
						if !keep {
							f.decoded[i] = v
							f.opaque[i] = true
						}
					}
				}
			}
		}
		return nil
	}
	paintActors, err := townScenePaintActors(p, actors)
	if err != nil {
		return nil, err
	}
	for _, a := range paintActors {
		x := (a.x - camera - 4) * 8
		if x <= -16 || x >= 224 {
			continue
		}
		phase := a.phase
		if !a.direction {
			phase += 4
		}
		off := 0x4000 + a.group*48 + phase*6
		ids := append([]byte(nil), r[off:off+6]...)
		for i := range ids {
			ids[i]--
			if ids[i] >= 152 {
				return nil, fmt.Errorf("unknown townSceneActor tile")
			}
		}
		if err := sprite(x, ids, 0x4100, 0x7000); err != nil {
			return nil, err
		}
	}
	phase := int(p[0xe7])
	if phase >= 5 {
		return nil, fmt.Errorf("unverified player phase")
	}
	table := 0x6a59
	if p[0xc2]&1 != 0 {
		table = 0x6a3b
	}
	table += phase * 6
	if p[0x83] < 27 {
		if err := sprite(int(p[0x83])*8, p[table:table+6], 0x6000, 0x8000); err != nil {
			return nil, err
		}
	}
	for y := 0; y < 64; y++ {
		for x := 0; x < 224; x++ {
			i := (y+78)*320 + x + 48
			v := int(f.decoded[y*224+x])
			if f.decoded[y*224+x] != raw[i] || !bytes.Equal(rgb[i*3:i*3+3], s.palette[v*3:v*3+3]) {
				return nil, fmt.Errorf("original pixel mismatch %d %d", x, y)
			}
		}
	}
	return f, nil
}

type townScenePanel struct {
	start, end, top, bottom int
	knots, source2          []int
	art                     image.Image
}

func townScenePanels(artDir string) ([]townScenePanel, error) {
	panels := []townScenePanel{
		{0, 576, 221, 499, []int{0, 321, 480, 576}, []int{0, 2400, 3554, 4344}, nil},
		{576, 1152, 224, 470, []int{0, 321, 537, 576}, []int{0, 2396, 4019, 4344}, nil},
		{1152, 1720, 234, 491, []int{0, 240, 496, 568}, []int{0, 1795, 3826, 4344}, nil},
	}
	names := []string{"left", "middle", "right"}
	expected := []string{"fbd1bc1146036a115b1e618b7df9c168515458126f5fc44a5f3225d89624e362", "0f8d7d3509293fb70fc400c12ce0b5e69e13b0ae96752928f44d9f3e35f21c1c", "948a9342bcbebb8a9addd2de8616da8389dced0122918593d6f258e35c04985f"}
	for i := range panels {
		b, err := checked(filepath.Join(artDir, "town-"+names[i]+"-art-v2.png"), expected[i], 16<<20)
		if err != nil {
			return nil, err
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		if im.Bounds() != image.Rect(0, 0, 2172, 724) {
			return nil, fmt.Errorf("前景圖尺寸不符")
		}
		panels[i].art = im
	}
	return panels, nil
}
func (p townScenePanel) sample(worldHD, yHD int) color.Color {
	x := worldHD - p.start*3
	k := 0
	for k+1 < len(p.knots)-1 && x >= p.knots[k+1]*3 {
		k++
	}
	x2 := p.source2[k] + (x-p.knots[k]*3)*(p.source2[k+1]-p.source2[k])/((p.knots[k+1]-p.knots[k])*3)
	return p.art.At(x2/2, p.top+yHD*(p.bottom-p.top)/192)
}

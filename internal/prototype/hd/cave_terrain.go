// 洞窟靜態地形HD。來源、包絡、取樣與退回規則見規格011及037。
package hd

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"path/filepath"
)

type CaveTerrain struct {
	tiles       [128][]color.NRGBA
	material    []color.NRGBA
	mpp1        *caveTerrainArt
	ice         *caveTerrainArt
	peligro     *caveTerrainArt
	mpp3        *caveTerrainArt
	mpp5        *caveTerrainArt
	mpp6        *caveTerrainArt
	mpp7        *caveTerrainArt
	mpp8        *caveTerrainArt
	mpp9        *caveTerrainArt
	mppa        *caveTerrainArt
	mppb        *caveTerrainArt
	common      *caveCommonArt
	commonState map[string]any
	player      *CavePlayer
	playerState map[string]any
	weapon      *CaveWeapon
	weaponState map[string]any
	objects     *CaveObjects
	objectState map[string]any
	portal      *CavePortal
	portalState map[string]any
}

func (s *CaveTerrain) SetPortal(portal *CavePortal) { s.portal = portal; s.portalState = nil }

func (s *CaveTerrain) PortalState() map[string]any {
	if s == nil || s.portalState == nil {
		return nil
	}
	state := make(map[string]any, len(s.portalState))
	for key, value := range s.portalState {
		if points, ok := value.([]image.Point); ok {
			value = append([]image.Point{}, points...)
		}
		state[key] = value
	}
	return state
}

func (s *CaveTerrain) SetPlayer(player *CavePlayer) { s.player = player; s.playerState = nil }

func (s *CaveTerrain) SetWeapon(weapon *CaveWeapon) { s.weapon = weapon; s.weaponState = nil }

func (s *CaveTerrain) SetObjects(objects *CaveObjects) { s.objects = objects; s.objectState = nil }

func (s *CaveTerrain) ObjectState() map[string]any {
	if s == nil || s.objectState == nil {
		return nil
	}
	state := make(map[string]any, len(s.objectState))
	for key, value := range s.objectState {
		if keys, ok := value.([]string); ok {
			value = append([]string{}, keys...)
		}
		state[key] = value
	}
	return state
}

func (s *CaveTerrain) WeaponState() map[string]any {
	if s == nil || s.weaponState == nil {
		return nil
	}
	state := make(map[string]any, len(s.weaponState))
	for key, value := range s.weaponState {
		state[key] = value
	}
	return state
}

func (s *CaveTerrain) PlayerState() map[string]any {
	if s == nil || s.playerState == nil {
		return nil
	}
	state := make(map[string]any, len(s.playerState))
	for key, value := range s.playerState {
		state[key] = value
	}
	return state
}

func caveArtwork(path, sha string) (image.Image, error) {
	b, e := checked(path, sha, 16<<20)
	if e != nil {
		return nil, e
	}
	c, e := png.DecodeConfig(bytes.NewReader(b))
	if e != nil || c.Width != 1254 || c.Height != 1254 {
		return nil, fmt.Errorf("洞窟圖面尺寸不符")
	}
	return png.Decode(bytes.NewReader(b))
}

// caveArea直接從來源解析度取樣，使用8-bit直通色與預乘RGBA整數面積平均。
func caveArea(src image.Image, x, y, w, h int) color.NRGBA {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	left, right := x*sw, (x+1)*sw
	top, bottom := y*sh, (y+1)*sh
	var sum [4]uint64
	for yy := top / h; yy < (bottom+h-1)/h; yy++ {
		wy := min(bottom, (yy+1)*h) - max(top, yy*h)
		for xx := left / w; xx < (right+w-1)/w; xx++ {
			wx := min(right, (xx+1)*w) - max(left, xx*w)
			weight := uint64(wx * wy)
			c := color.NRGBAModel.Convert(src.At(b.Min.X+xx, b.Min.Y+yy)).(color.NRGBA)
			a := uint64(c.A) * weight
			sum[0] += uint64(c.R) * a
			sum[1] += uint64(c.G) * a
			sum[2] += uint64(c.B) * a
			sum[3] += a
		}
	}
	if sum[3] == 0 {
		return color.NRGBA{}
	}
	area := uint64(sw * sh)
	return color.NRGBA{uint8((sum[0] + sum[3]/2) / sum[3]), uint8((sum[1] + sum[3]/2) / sum[3]), uint8((sum[2] + sum[3]/2) / sum[3]), uint8((sum[3] + area/2) / area)}
}

func NewCaveTerrain(dir string) (*CaveTerrain, error) {
	atlas, e := caveArtwork(filepath.Join(dir, "terrain-handpaint-v2.png"), "07004b059f8ed0ecf50c9aaf3a8b35c965050725fc652325d5ef248842a0b1bc")
	if e != nil {
		return nil, e
	}
	material, e := caveArtwork(filepath.Join(dir, "rock-material-v1.png"), "09d331a411515a441790479646484467ca529517adc3f3cc235936921e57bf8f")
	if e != nil {
		return nil, e
	}
	s := &CaveTerrain{}
	for slot := 1; slot < 15; slot++ {
		for y := 0; y < 24; y++ {
			for x := 0; x < 24; x++ {
				s.tiles[slot] = append(s.tiles[slot], caveArea(atlas, slot%4*24+x, slot/4*24+y, 96, 96))
			}
		}
	}
	s.material = make([]color.NRGBA, 432*768)
	for y := 0; y < 768; y++ {
		for x := 0; x < 432; x++ {
			s.material[y*432+x] = caveArea(material, x, y, 432, 768)
		}
	}
	return s, nil
}

// Draw的守衛全部先完成，再交換地形。其餘圖層依原版遮罩回畫。
func (s *CaveTerrain) Draw(dst *image.RGBA, f *CavePresentation, origin image.Point) bool {
	if s != nil {
		s.playerState = nil
		s.weaponState = nil
		s.objectState = nil
		s.portalState = nil
		s.commonState = nil
		if s.common != nil {
			s.commonState = map[string]any{"loaded": true, "matched": false}
		}
		if s.player != nil {
			s.playerState = map[string]any{"loaded": true, "matched": false, "occluded": false}
		}
		if s.weapon != nil {
			s.weaponState = map[string]any{"loaded": true, "matched": false, "empty": false}
		}
		if s.objects != nil {
			s.objectState = map[string]any{"loaded": true, "matched": false}
		}
	}
	box := image.Rect(144, 42, 816, 474).Add(origin)
	if s == nil || dst == nil || f == nil || !box.In(dst.Bounds()) || len(s.material) != 432*768 || len(f.Indexed) != 64000 || len(f.RGB) != 192000 || !caveFrameSource(caveFrameInput{mode: 0x13, memory: f.Memory, palette: f.Palette}) {
		return false
	}
	if !caveFrameSource(caveFrameInput{mode: 0x13, memory: f.BodyMemory, palette: f.Palette}) {
		return false
	}
	if caveCatalogMapsDiffer(f.Memory, f.BodyMemory) {
		return false
	}
	r := f.Memory[0x11b10:0x21b10]
	tiles, material, limit, extraArt := s.tiles, s.material, 15, false
	var materialIDs [128]bool
	var secondaryIDs, textIDs [128]bool
	var secondaryMaterial []color.NRGBA
	for id := 3; id <= 8; id++ {
		materialIDs[id] = true
	}
	textFirst, textLast, legacyThrough := 0, 0, 0
	if s.mpp1 != nil || townBackgroundHash(r[0x8030:0x8030+14*48]) != "69f3a97cfe0356b544b67e462c5c59835a72c66099086e68f2f81b47f6cfa74b" {
		entry, terrain := caveTerrainIdentity(f.Memory)
		bodyEntry, _ := caveTerrainIdentity(f.BodyMemory)
		if entry == nil || bodyEntry == nil || entry.record != bodyEntry.record {
			return false
		}
		var art *caveTerrainArt
		switch terrain.key {
		case "MPP1":
			art = s.mpp1
			legacyThrough = 15
		case "MPP4":
			art = s.ice
		case "MPP2":
			art = s.peligro
		case "MPP3":
			art = s.mpp3
		case "MPP5":
			art = s.mpp5
		case "MPP6":
			art = s.mpp6
		case "MPP7":
			art = s.mpp7
		case "MPP8":
			art = s.mpp8
		case "MPP9":
			art = s.mpp9
		case "MPPA":
			art = s.mppa
		case "MPPB":
			art = s.mppb
		default:
			return false
		}
		if art == nil || art.count != terrain.tiles {
			return false
		}
		tiles, material, extraArt = art.tiles, art.material, true
		limit, materialIDs, textFirst, textLast = art.count+1, art.materialIDs, art.textFirst, art.textLast
		secondaryMaterial, secondaryIDs, textIDs = art.secondaryMaterial, art.secondaryIDs, art.textIDs
	}
	if !cavePaintTimeline(f.Paints, f.BodySteps, f.BodyCycles, f.Steps, f.Cycles) {
		return false
	}
	d, e := decodeCaveFrameWithPaints(f.Memory, f.BodyMemory, f.SourceIndexed, f.Paints)
	if e != nil {
		return false
	}
	for y := 14; y < 158; y++ {
		for x := 48; x < 272; x++ {
			i := y*320 + x
			v := f.SourceIndexed[i]
			if f.Indexed[i] != v || !bytes.Equal(f.RGB[i*3:i*3+3], f.Palette[int(v)*3:int(v)*3+3]) {
				return false
			}
		}
	}
	var grid [18][28]byte
	var decoded [128][]byte
	if len(material) != 432*768 {
		return false
	}
	for id := 1; id < limit; id++ {
		if secondaryIDs[id] && (!materialIDs[id] || len(secondaryMaterial) != 432*768) {
			return false
		}
		if len(tiles[id]) != 576 {
			return false
		}
		decoded[id], e = caveStatic(r, byte(id))
		if e != nil {
			return false
		}
	}
	var nativePortals []image.Point
	portalChecked := false
	commonPresent := false
	for _, c := range d.Cells {
		if c.ID >= 64 && c.ID <= 66 {
			commonPresent = true
			break
		}
	}
	var commonPositions []image.Point
	if commonPresent {
		if s.common == nil {
			return false
		}
		var commonOK bool
		commonPositions, commonOK = caveCommonPlacements(f.Memory, f.BodyMemory, d.Cells)
		if !commonOK || len(commonPositions) == 0 {
			return false
		}
		for i := range s.common.tiles {
			if len(s.common.tiles[i]) != 576 {
				return false
			}
			decoded[64+i], e = caveStatic(r, byte(64+i))
			if e != nil {
				return false
			}
		}
	}
	for _, c := range d.Cells {
		if c.ID >= 64 && c.ID <= 66 && commonPresent {
			grid[(c.Y-14)/8][(c.X-48)/8] = c.ID
			continue
		}
		if extraArt && int(c.ID) >= limit {
			// MPP1仍含規格020已驗的共用入口格，僅保留完整矩陣內的原版下層。
			if legacyThrough != 15 {
				return false
			}
			if !portalChecked {
				var sourceOK bool
				nativePortals, sourceOK = cavePortalLocations(f.Memory, d.Cells)
				_, bodyOK := cavePortalLocations(f.BodyMemory, nil)
				if !sourceOK || !bodyOK {
					return false
				}
				portalChecked = true
			}
			allowed := false
			for _, anchor := range nativePortals {
				x, y := (c.X-anchor.X)/8, (c.Y-anchor.Y)/8
				if c.X >= anchor.X && c.Y >= anchor.Y && x < 5 && y < 5 && cavePortalMatrix[y][x] == c.ID {
					allowed = true
					break
				}
			}
			if !allowed {
				return false
			}
		}
		grid[(c.Y-14)/8][(c.X-48)/8] = c.ID
	}
	rock := make([]bool, 224*144)
	for column := 0; column < 28; column++ {
		for row := 0; row < 18; {
			if !materialIDs[grid[row][column]] {
				row++
				continue
			}
			first := row
			for row < 18 && materialIDs[grid[row][column]] {
				row++
			}
			last := row
			for x := 0; x < 8; x++ {
				low, high := 144, -1
				for y := first * 8; y < last*8; y++ {
					if decoded[grid[y/8][column]][(y%8)*8+x] != 0 {
						low = min(low, y)
						high = max(high, y)
					}
				}
				for y := low; y <= high; y++ {
					rock[y*224+column*8+x] = true
				}
			}
		}
	}
	var playerPose *cavePlayerPose
	if s.player != nil {
		key, pose, hidden := s.player.selectPose(d)
		playerPose = pose
		s.playerState = map[string]any{"loaded": true, "matched": pose != nil, "occluded": hidden, "source_key": key, "source_steps": f.Steps, "body_steps": f.BodySteps}
		if pose != nil {
			s.playerState["pose"] = pose.number
		}
	}
	var weaponPose *caveWeaponPose
	if s.weapon != nil {
		key, pose, empty := s.weapon.selectPose(d)
		weaponPose = pose
		s.weaponState = map[string]any{"loaded": true, "matched": pose != nil, "empty": empty, "source_key": key, "source_steps": f.Steps, "body_steps": f.BodySteps}
		if pose != nil {
			s.weaponState["pose"] = pose.number
		}
	}
	if playerPose != nil || weaponPose != nil {
		// 移除原版角色及武器，還原當次下層。原貌角色與配色也由後續原圖層重畫。
		for y := 14; y < 158; y++ {
			for x := 48; x < 272; x++ {
				v := int(d.Background[y*320+x]) * 3
				c := color.RGBA{f.Palette[v], f.Palette[v+1], f.Palette[v+2], 255}
				for dy := 0; dy < 3; dy++ {
					for dx := 0; dx < 3; dx++ {
						dst.SetRGBA(x*3+dx+origin.X, y*3+dy+origin.Y, c)
					}
				}
			}
		}
	}
	for _, cell := range d.Cells {
		id := int(cell.ID)
		isCommon := commonPresent && id >= 64 && id <= 66
		if id == 0 || id >= limit && !isCommon {
			continue
		}
		tilePixels := tiles[id]
		if isCommon {
			tilePixels = s.common.tiles[id-64]
		}
		gx, gy := cell.X-48, cell.Y-14
		for y := 0; y < 24; y++ {
			for x := 0; x < 24; x++ {
				if materialIDs[id] {
					if !rock[(gy+y/3)*224+gx+x/3] {
						continue
					}
				} else if !(extraArt && (textIDs[id] || id >= textFirst && id <= textLast)) && decoded[id][(y/3)*8+x/3] == 0 {
					continue
				}
				c := tilePixels[y*24+x]
				if materialIDs[id] {
					ring := cell.Address - 0xe000
					tx, ty := ring%36*24+x, ring/36*24+y
					tx = min(tx, 863-tx)
					ty = min(ty, 1535-ty)
					c = material[ty*432+tx]
					if secondaryIDs[id] {
						c = secondaryMaterial[ty*432+tx]
					}
				}
				px, py := cell.X*3+x+origin.X, cell.Y*3+y+origin.Y
				old := dst.RGBAAt(px, py)
				if extraArt && id > legacyThrough {
					// 新圖面取代原生地形色；動態、主角與武器稍後依原順序回畫。
					old = color.RGBA{A: 255}
				}
				a, inv := uint32(c.A), uint32(255-c.A)
				dst.SetRGBA(px, py, color.RGBA{uint8((uint32(c.R)*a + uint32(old.R)*inv + 127) / 255), uint8((uint32(c.G)*a + uint32(old.G)*inv + 127) / 255), uint8((uint32(c.B)*a + uint32(old.B)*inv + 127) / 255), 255})
			}
		}
	}
	paintNative := func(layer []cavePixel) {
		for _, v := range layer {
			c := color.RGBA{f.Palette[int(v.Color)*3], f.Palette[int(v.Color)*3+1], f.Palette[int(v.Color)*3+2], 255}
			for y := 0; y < 3; y++ {
				for x := 0; x < 3; x++ {
					dst.SetRGBA(v.X*3+x+origin.X, v.Y*3+y+origin.Y, c)
				}
			}
		}
	}
	if s.portal != nil {
		s.portalState = s.portal.draw(dst, f, d, origin)
	}
	if s.objects == nil {
		paintNative(d.Dynamic)
	} else {
		s.objectState = s.objects.draw(dst, d, f.Palette, origin)
		s.objectState["source_steps"] = f.Steps
		s.objectState["body_steps"] = f.BodySteps
	}
	if playerPose != nil {
		playerPose.draw(dst, d, f.Paints, origin)
	} else {
		paintNative(d.Actor)
	}
	if weaponPose != nil {
		weaponPose.draw(dst, d, origin)
	} else {
		paintNative(d.Weapon)
	}
	if s.common != nil {
		s.commonState = map[string]any{"loaded": true, "matched": commonPresent, "empty": !commonPresent, "placements": commonPositions, "source_steps": f.Steps, "body_steps": f.BodySteps}
	}
	return true
}

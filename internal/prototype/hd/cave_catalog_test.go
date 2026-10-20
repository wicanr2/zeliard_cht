package hd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func caveCatalogFixtures(t *testing.T) (string, []byte, []byte) {
	t.Helper()
	project := os.Getenv("ZELIARD_PROJECT")
	if project == "" {
		t.Skip("須掛載本地原版來源；skip不算驗收")
	}
	read := func(p string) []byte {
		b, e := os.ReadFile(filepath.Join(project, p))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	return filepath.Join(project, "workplace/full-hd-v1/cave-catalog-fixtures-v1"), read("workplace/full-hd-v1/cave-save-survey-v2/SAVEM1/map-load-00.memory"), read("workplace/prose-v1/progression-v1/dungeon-entry.palette")
}

func TestCaveCatalogRealSourceKeys(t *testing.T) {
	dir, reference, palette := caveCatalogFixtures(t)
	read := func(name string) []byte {
		b, e := os.ReadFile(filepath.Join(dir, name))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	var manifest struct {
		Records []struct {
			Record   int    `json:"record"`
			Name     string `json:"name"`
			Terrain  string `json:"terrain"`
			Header   string `json:"header_file"`
			Resource string `json:"resource_file"`
		} `json:"records"`
	}
	if e := json.Unmarshal(read("manifest.json"), &manifest); e != nil {
		t.Fatal(e)
	}
	if len(manifest.Records) != 31 || len(caveMapEntries) != 31 || len(caveTerrainEntries) != 11 {
		t.Fatal("全量來源數量不符")
	}
	for _, row := range manifest.Records {
		t.Run(row.Name, func(t *testing.T) {
			// 合成資料只測來源鍵；不執行VM，不當正常玩家路徑或完整畫格。
			m := cloneTownBytes(reference)
			copy(m[0x1b10+0xc000:], read(row.Header))
			copy(m[0x11b10+0x8000:], read(row.Resource))
			before := cloneTownBytes(m)
			entry, terrain := caveTerrainIdentity(m)
			if entry == nil || terrain == nil || entry.record != row.Record || entry.name != row.Name || terrain.key != row.Terrain {
				t.Fatal("來源鍵錯配", entry, terrain)
			}
			if !caveFrameSource(caveFrameInput{mode: 0x13, memory: m, palette: palette}) || !bytes.Equal(m, before) {
				t.Fatal("來源拒絕或資料被寫入")
			}
			for _, offset := range []int{0x1b10 + 0xc000, 0x1b10 + 0xc012, 0x11b10 + 0x8000, 0x11b10 + 0x8030, 0x11b10 + 0x8030 + terrain.tiles*48 - 1} {
				m[offset] ^= 1
				if bad, _ := caveTerrainIdentity(m); bad != nil {
					t.Fatal("錯誤使用區被接受", offset)
				}
				m[offset] ^= 1
			}
		})
	}
}

func TestCaveCatalogUnknownAndCrossMap(t *testing.T) {
	dir, base, palette := caveCatalogFixtures(t)
	read := func(n string) []byte {
		b, e := os.ReadFile(filepath.Join(dir, n))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	left, right := cloneTownBytes(base), cloneTownBytes(base)
	copy(left[0x1b10+0xc000:], read("MP40.header"))
	copy(right[0x1b10+0xc000:], read("MP41.header"))
	for _, m := range [][]byte{left, right} {
		copy(m[0x11b10+0x8000:], read("MPP4.resource"))
	}
	if !caveCatalogMapsDiffer(left, right) || caveCatalogMapsDiffer(left, left) {
		t.Fatal("同家族不同地圖未區分")
	}
	// 刻意沿用相同假畫面，確認撤除依地圖身份，不靠畫面差異碰巧拒絕。
	in := caveFrameInput{mode: 0x13, memory: right, palette: palette, indexed: make([]byte, 64000), rgb: make([]byte, 192000), steps: 200, cycles: 200}
	frames := &CaveFrames{drawing: true, complete: &CavePresentation{Memory: left, Steps: 100, Cycles: 100}}
	if frames.selectFrame(in, false) != nil || frames.complete != nil {
		t.Fatal("同家族跨地圖仍沿用舊畫格")
	}
	body := caveFrameInput{mode: 0x13, memory: left, palette: palette, steps: 100, cycles: 100}
	frames.body = &body
	frames.finish(in)
	if frames.complete != nil || frames.lastRefusal["reason"] != "body and complete source belong to different maps" {
		t.Fatal("跨地圖的主體與完成來源未拒絕", frames.lastRefusal)
	}
	copy(right[0x11b10+0x8000:], read("MPP2.resource"))
	if entry, _ := caveTerrainIdentity(right); entry != nil {
		t.Fatal("地圖與圖集錯配被接受")
	}
	if entry, _ := caveTerrainIdentity(base[:1000]); entry != nil {
		t.Fatal("短來源被接受")
	}
}

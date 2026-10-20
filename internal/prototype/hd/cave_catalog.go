// 唯讀洞窟地圖／圖集身份。來源與READY範圍見規格037。
package hd

type caveTerrainEntry struct {
	key                    string
	tiles                  int
	headerHash, pixelsHash string
}

type caveMapEntry struct {
	record           int
	name, prefixHash string
	terrain          int
}

func caveTerrainIdentity(memory []byte) (*caveMapEntry, *caveTerrainEntry) {
	if len(memory) != 1<<20 {
		return nil, nil
	}
	p, r := memory[0x1b10:0x11b10], memory[0x11b10:0x21b10]
	prefix := townBackgroundHash(p[0xc000:0xc013])
	for i := range caveMapEntries {
		entry := &caveMapEntries[i]
		if entry.prefixHash != prefix || entry.terrain < 0 || entry.terrain >= len(caveTerrainEntries) {
			continue
		}
		terrain := &caveTerrainEntries[entry.terrain]
		if terrain.tiles < 1 || terrain.tiles > 127 || townBackgroundHash(r[0x8000:0x8030]) != terrain.headerHash || townBackgroundHash(r[0x8030:0x8030+terrain.tiles*48]) != terrain.pixelsHash {
			return nil, nil
		}
		return entry, terrain
	}
	return nil, nil
}

func caveCatalogMapsDiffer(a, b []byte) bool {
	left, _ := caveTerrainIdentity(a)
	right, _ := caveTerrainIdentity(b)
	return left != nil && right != nil && left.record != right.record
}

package hd

import "github.com/wicanr2/dosgolem/oracle"

// 場景判斷只決定讀圖時機；實際呈現仍走原有完整來源守衛。
func CaveRuntimeModule(o *oracle.Oracle) bool {
	return o != nil && o.VideoMode() == 0x13 && o.Word(oracle.Far(0x1b1, 0xff2c)) == 0x11b1 && townBackgroundHash(o.Bytes(oracle.Far(0x1b1, 0x6f9b), 0x7183-0x6f9b)) == "26dedabe338c9c2f6ebaeb8e26f5095486ce3c8863a02316420e72ed3e44aac5"
}

func CurrentCaveTerrainFamily(o *oracle.Oracle) string {
	if !CaveRuntimeModule(o) {
		return ""
	}
	_, terrain := caveTerrainIdentity(o.Bytes(oracle.Far(0, 0), 1<<20))
	if terrain == nil {
		return ""
	}
	return terrain.key
}

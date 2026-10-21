// 十二個已存主角來源，含六種可見色彩變體；來源及拒絕契約見046。
package hd

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"
)

var caveSavedPlayerPoses = []struct {
	key, hash string
	bounds    image.Rectangle
}{
	{"adfbaa6513a572294facc874d0309e3fc6df8e939d5cbbd705b93c89817e1d0f", "e2528785987e334e16062d8790c2b9f3c391eee07a963e4f0cb8e396ed148c4b", image.Rect(2, 0, 23, 24)},
	{"973cf8441e83bcdf05100e3d54444a33c52c1b3af053fa604f2cfe59db090603", "10773ffe44559187f9a9b6761942ea940bc0a28938f6e2d7c2ddf071d089345b", image.Rect(2, 0, 23, 22)},
	{"34d58d3bc59ab55120466506ac2623993524a6450fd7cb538d54c2b41e279354", "eb965872dd44731d5b1ad5f5c6764bd033d1dd4a08e6d9336f7fd926255942ad", image.Rect(2, 8, 23, 24)},
	{"a18640699ea0b46bc9e354991ed01c8adda0f868432cb77e7b501ca6c0960fa3", "dc0d7ba36a0a7de480655860148334ba0e8e16532612a023557de118ee6b9b5a", image.Rect(1, 0, 22, 24)},
	{"70fa8fadd8a00b8f2fe56f051114840975540bec19bfa16b0b396246c73cb271", "1461233ea14d373cfaf11b8db335222185a3c5e0019e3a051b0ccf1eee41ce3b", image.Rect(1, 0, 22, 22)},
	{"b1993a3edc02297283041a6609be7854943d753bd6cec5509f53975d8e0bed54", "b4012e592f05499bd17107eb50fb8df27a35549a1efc4dcef1ea5d3aa7a50ad3", image.Rect(0, 1, 22, 24)},
	{"913414fdb55e104184d56d5b069a58f4236e3a10edb1957cfd0514b89b0e42b0", "5295f392a2284fa7a76046568b022a4960896fc17bfb98ecf9a801882250a862", image.Rect(2, 8, 23, 24)},
	{"9d0ee7964b900939af26eef2a2b9733563098640349e8fa90fad06c77d474b97", "f2868f5a2d2658aada7a02db60a78b6badeaf2580aa2841a2a5ffcc57b8cf367", image.Rect(2, 0, 23, 24)},
	{"a1fb0dde2740bc2abe22b2464aa0a27a36bae08b81b6034c103841c12d0975c8", "19b7268518d7dc9cda97c49c32498dbef695a4ac6a59154b80bcbe27ea1c40e3", image.Rect(2, 0, 23, 24)},
	{"f67066ae97358b84f9d865fa936fdfaaaf494eff6147b0d9b80b914541f129ff", "ea37890aa63cc7e68317c84a12c7ca8be29071a1d6f0b664385072d52b810ecc", image.Rect(2, 1, 23, 24)},
	{"3540aa56a0e2a1102c2ff0ccdf65f8e5736026a2a82bef13021db451db7130b8", "afad24ede8fb9c4be70c81d456ccd5eb1ceba8e1d2305f58842b339a79c05227", image.Rect(2, 0, 21, 24)},
	{"85d437088ac916ca16dc82be5742d7ac55c0ea41ec46ccfb88f3543643fcfc8a", "2688f7362fecf4332aceb74cd80c7fdda62071d0955006398f35ebd85f9ebc7d", image.Rect(2, 0, 21, 24)},
}

func (p *CavePlayer) LoadExtra(dir string) error {
	if p == nil || len(p.poses) < len(cavePlayerRegistrations) {
		return fmt.Errorf("須先載入洞窟主角圖組")
	}
	poses := make(map[string]*cavePlayerPose, len(p.poses)+len(caveSavedPlayerPoses))
	for key, pose := range p.poses {
		poses[key] = pose
	}
	for n, r := range caveSavedPlayerPoses {
		im, err := checkedTerrainImage(filepath.Join(dir, "player-"+r.key[:12]+"-72x72.png"), r.hash, 72, 72)
		if err != nil {
			return err
		}
		pose := &cavePlayerPose{number: len(cavePlayerRegistrations) + n, bounds: image.Rectangle{Min: r.bounds.Min.Mul(3), Max: r.bounds.Max.Mul(3)}}
		for y := 0; y < 72; y++ {
			for x := 0; x < 72; x++ {
				pose.pixels = append(pose.pixels, color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA))
			}
		}
		poses[r.key] = pose
	}
	p.poses = poses
	return nil
}

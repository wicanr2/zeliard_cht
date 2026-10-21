// 五個正常組合主角來源；來源、圖面及原子載入契約見048。
package hd

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"
)

var caveProfilePlayerPoses = []struct {
	key, hash string
	bounds    image.Rectangle
}{
	{"d8cb8cfd414efb65ffe876a27827197c9dd8c2a4a66409028257d78c36c32517", "0d983e4728c2700ef0dd67c4fe8040219929f387d86371aff8f72f40bd2ae551", image.Rect(5, 1, 19, 22)},
	{"1f9eaad0ab7b2b4b5285f513a380b6e820c62e72cdc4a33b0280cb1b773765fe", "d30ed09f04a8c5dacc41dfd407c25d55a7037849e857580099bea5c48030dd9c", image.Rect(5, 0, 19, 24)},
	{"a70246c25a07c41f5c086178557bbb19ad800ff557420f954f35c33be6ebd829", "c2742135794c2b34d0bc797b8a046797ce3ce71c100b72d61133ec1124b651b2", image.Rect(5, 0, 18, 24)},
	{"76add99156752b3810fed75ed11fe8b38be168d03d212a888c1c26392155fde6", "51ded12acff6ae339979c8d9b3dca9d4f70b7554c026570451f33b93313ba1eb", image.Rect(5, 0, 18, 24)},
	{"c37032a1de85376a46fedcc23d2178b20e35612eec219c55a06fd72c521674ef", "7fa6f3dcb42b5bb33731f9e6ba739e35c7be56811883136a15d9abd510359b43", image.Rect(5, 0, 21, 24)},
}

func (p *CavePlayer) LoadProfiles(dir string) error {
	if p == nil || len(p.poses) < len(cavePlayerRegistrations) {
		return fmt.Errorf("須先載入洞窟主角圖組")
	}
	poses := make(map[string]*cavePlayerPose, len(p.poses)+len(caveProfilePlayerPoses))
	for key, pose := range p.poses {
		poses[key] = pose
	}
	for n, r := range caveProfilePlayerPoses {
		im, err := checkedTerrainImage(filepath.Join(dir, "player-"+r.key[:12]+"-72x72.png"), r.hash, 72, 72)
		if err != nil {
			return err
		}
		pose := &cavePlayerPose{number: len(cavePlayerRegistrations) + len(caveSavedPlayerPoses) + n, bounds: image.Rectangle{Min: r.bounds.Min.Mul(3), Max: r.bounds.Max.Mul(3)}}
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

// 青藍刀光的十個已保存來源；只增圖面，原版狀態與繪製時序不變，見046。
package hd

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"
)

var caveCyanWeapons = []struct {
	key, hash string
	bounds    image.Rectangle
}{
	{"24bb8522736fa4a94644686615ead3718986a1135066b878b445f012d6e05132", "e9239f083037bf8af42fd35ab1858ae82a5fcc05bd43cf28555dceed8a166cac", image.Rect(9, 8, 20, 24)},
	{"3cb8af06938ffbc66e36c6b6e596b1609b0cdcd996b5e8c3bee8bf4afda45a72", "355ac5c82a2573e15cc0d50188b06212c7e0a1ffa49a078e235704735be8a8bd", image.Rect(8, 12, 23, 26)},
	{"ce8b066c89a58e7a20ebd139c50848ea678d8604ae18dbd9f391ee81412c82ca", "3df6b7b0c3755e625a8646cd261da388083a42355a4a7b6cc0880f6cbb0eff12", image.Rect(3, 24, 24, 28)},
	{"b91e2ecd8b77f605b0a1b4bfb160479ccb4d0e9907e45a951ced1bbedc5d534a", "f0d42888f782be67077a91ffa78aa0ae3060ccbc83d68a69fe6c250b5d3810d4", image.Rect(12, 24, 31, 29)},
	{"b3b9802b33d378374f444d305c94f6aef0444756a3377b9514e2f645b3e8f4b0", "09479b8e247d0d37392123a616313aa52aaec4553ecff38f7a1b71d1f621f3e2", image.Rect(12, 24, 27, 26)},
	{"90b95093c51184692759ff9a4fad9e190ff3e0fdcc08b865bdb23482bc2cc74e", "dff006a801b3319994f1f31dadb1578566e778d1acd44d15e34d97539988b953", image.Rect(12, 8, 23, 24)},
	{"8c19e605b0dcd20d8817d1d98577acdab84a73c0742bbadcdd1a87338c908a4f", "9d9ea00a84e2475f9fa59ba78b451986edcf7f2eecb25164d1e50092ae9af6fc", image.Rect(9, 12, 25, 26)},
	{"204793599e76ef48dda0971e5fbdf30b593fd4100003587f5b07bfc3ec05b711", "4b2c4a1c319a16427adf39437e615b15a6e9be26fd98f04fd84b3bb167c1ff64", image.Rect(8, 24, 29, 28)},
	{"27e77e15981ed249ea259444772532dd48b7d77ebc594b184999c49ed9babdf1", "f5f7662e3850e06efb40639d5d34fb7953271976d1f49bcb986a86350ecc2c8c", image.Rect(1, 24, 20, 29)},
	{"4377d0c0f442e2ea728bf4935d061828f75aa8c98f9f80e2c2055e7b4bdc6e7f", "19a8e56dfc3df0eee6f6dc2ff9eb930803b2dc2748d7e1686e56fad577a63755", image.Rect(5, 24, 20, 26)},
}

func (w *CaveWeapon) LoadCyan(dir string) error {
	if w == nil || len(w.poses) < len(caveWeaponRegistrations) {
		return fmt.Errorf("須先載入洞窟武器圖組")
	}
	poses := make(map[string]*caveWeaponPose, len(w.poses)+len(caveCyanWeapons))
	for key, pose := range w.poses {
		poses[key] = pose
	}
	for n, r := range caveCyanWeapons {
		im, err := checkedTerrainImage(filepath.Join(dir, "weapon-"+r.key[:12]+"-96x96.png"), r.hash, 96, 96)
		if err != nil {
			return err
		}
		pose := &caveWeaponPose{number: len(caveWeaponRegistrations) + n, bounds: image.Rectangle{Min: r.bounds.Min.Mul(3), Max: r.bounds.Max.Mul(3)}}
		for y := 0; y < 96; y++ {
			for x := 0; x < 96; x++ {
				pose.pixels = append(pose.pixels, color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA))
			}
		}
		poses[r.key] = pose
	}
	w.poses = poses
	return nil
}

// 洞窟主角55個已驗姿態。來源、透明登錄與遮擋契約見規格017、023、027及028。
package hd

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"path/filepath"
)

type cavePlayerRegistration struct {
	key          string
	crop, native image.Rectangle
	pixelsSHA    string
}

var cavePlayerRegistrations = []cavePlayerRegistration{
	{"134def85cdcf57d10f214fb5d9542e7e0f00919178c76e35f644d15136097176", image.Rect(52, 10, 247, 313), image.Rect(3, 0, 18, 24), "9fb8869fc9cc513b9bf53bee1692f8444e4e4365f28c331b7c43d007f2e231f7"},
	{"011099c1bd944e332c3a00b0cf59822ea8266d670f6c6b7209602814b20961ac", image.Rect(364, 10, 601, 313), image.Rect(3, 0, 22, 24), "fcfd8329142fe39758e94e0896dc196ec8a2b7a8d74bdaf9d42c83d17dd71a87"},
	{"64cf33f733ee603eb7c8ec06a3359f61d44bb12719a28efbbb0d8765a1164953", image.Rect(693, 13, 853, 312), image.Rect(1, 0, 16, 24), "fd0396398f8dfe2bc7398c1350c13882e1898ab77089514c5c0f9bfc404494c2"},
	{"d1bbd95568b7533ba1cb3bec1b2b56bd2adfe74abc4dc522e1884503d3b2e2be", image.Rect(1017, 10, 1238, 310), image.Rect(1, 0, 22, 24), "c6958a953ed319c57a9625ca752e3b0d4e7497e0840272ceb727d6049bb8531d"},
	{"434052a09bec05467d7cc7b8d6bbafac63990037af7fedc3af450b27d573037b", image.Rect(56, 330, 250, 631), image.Rect(5, 0, 21, 24), "e482b21828ef6971940d6c7706d81b636cb866f512a06badcecf2c6007f460a7"},
	{"85cf38445986e393a1a973db01713a8217d5bda0a89200e7fe07102becba0568", image.Rect(367, 330, 580, 581), image.Rect(2, 0, 23, 22), "8e1ca7233ba89402179b032f3d138697cf7394f13bcc6a6ca785dab9c489eb58"},
	{"41c4423e60d19e66f64e172a002ef0925e435fb0c44ca4f239791f55b1f346c7", image.Rect(701, 331, 909, 605), image.Rect(4, 0, 23, 24), "3d58cf8200dd7991c14794673d24b8a119a9f87e38a08f74fbd962d1302c9810"},
	{"35ebe757119e4021fb418d7093a3a9f9bb4242a6b8bdd423cb26f2ec9bce014f", image.Rect(1022, 417, 1232, 628), image.Rect(5, 8, 23, 24), "33046cbe1aa4de14e8f66655547b05d5e39ed364c6ce577b6d0244199b79012e"},
	{"d3eca2289acabd86221755f7aea1b4d810a32a6e404ce803a174c3adeb28246e", image.Rect(102, 642, 251, 936), image.Rect(8, 0, 23, 24), "23c7fb0b6aef82a2da0c3b95909b012f603aa0574412ff05ab04ce0e4e4f3488"},
	{"2864b4ac02de96793b4fec28d09b4a343991adddeaa5d974a274ec6bae3db12f", image.Rect(346, 637, 591, 935), image.Rect(2, 0, 21, 24), "35bd1f4d611396cce132ce0d5188ebc992250cb6a0e88b21d90d5570324766cf"},
	{"0da2e9076321af3ada61d400e6a5dc4eaf8aad23ffe2a5679e1313cdb2f6107c", image.Rect(690, 640, 896, 934), image.Rect(5, 0, 21, 24), "c9274970003ac8fcd4a408af0c554860535953747bacab72d5e827d814f43e32"},
	{"a0a4f9501680ebb64f8976bce50e60926dea456321e34da2b18bddb4ec01ec68", image.Rect(1021, 642, 1187, 933), image.Rect(5, 1, 19, 22), "03822e9f63e1f8c051e658d06bd00362480c93e013b15f1496511f7fdaf0009d"},
	{"7ed7885c928a308dab869198a9639d0473e4d5c20fb01decc3854cd47e509fee", image.Rect(44, 950, 246, 1231), image.Rect(4, 2, 20, 24), "13a6d4dfb125ab3fd18f9b583bd2e4a1fb59b67297eb6897bafcdd305bac76e9"},
	{"1569b6fdba9b1b59ae1faeda5d1b614fcfff6083450bdddebc3952155790a19d", image.Rect(378, 943, 554, 1232), image.Rect(4, 0, 20, 24), "6d64e5551b071144fd42f0c4a5f981782e76e1495d5800f604f179f5abf1e7ce"},
	{"630449c1e3f8263a06d33803a2bb32b67df0abd893db5fc53a3096e2b2f62276", image.Rect(82, 15, 363, 432), image.Rect(3, 0, 18, 24), "b28beaa51cb803bd378313c652828bd359b62ff04193c1084d00022e1e3a02de"},
	{"33a11b8370831735ed528944786045a930a24e48ac557e8efa0d3271cfd02952", image.Rect(462, 22, 804, 431), image.Rect(0, 1, 19, 24), "127934199187cdba44bdb15b11dbdf4e16b19d18e867eb5be5a828d27ae3b142"},
	{"f9d93abd7371a6a8c7f8ae3c64d538562b6575b4241b309d0f75622ae6bea3a5", image.Rect(901, 18, 1175, 432), image.Rect(1, 0, 16, 24), "5ed4f0968e7274d271e9a21b96e708580f6df72d5d79b2f1d3e4fafbdae2b6af"},
	{"3dbe2ff88b97f7d4a64552f0dcae771a9c492be713e72f19bbbb700b787338d7", image.Rect(49, 435, 377, 837), image.Rect(0, 1, 19, 24), "9e69c4655c999cca161f2bc544254629d6ebd436c48d9b0ad4f26242ee7ca774"},
	{"038aa2096389c6a65939add3bb9c3ecaace979ca02d2f2596e2a1d5cd175f466", image.Rect(512, 436, 732, 834), image.Rect(1, 0, 16, 24), "0abc2d7d812f04b87eb2785fa74b737d520fb7ebcbe499031820af96aed2ec5f"},
	{"86c4e094e544cbb6fc5d6c700f8c9b80f73ba9d7da83f41c603fbf3feae8c21d", image.Rect(867, 436, 1232, 839), image.Rect(0, 1, 22, 24), "2499a045f73e461f953cc57eb91e2ab998407cf498765edb0b757403c1e33a3d"},
	{"91057118a51181b67e52b3155dc2e97b3f53308088306c9425d62f8a76713a35", image.Rect(93, 838, 368, 1228), image.Rect(8, 0, 23, 24), "7df5302ca572403ae2931039af4cc2612ddff501a2946c8135d89ac410eaaae2"},
	{"9596b0eccc8a9d547dae95bfa6b63e02ea0e186869c1231fe7fbb3d209d132f7", image.Rect(437, 833, 770, 1230), image.Rect(5, 1, 23, 24), "4c3ae48930fdaf365c4a20b41d8ee22b5b81e4877aadb4970c3fe97b99b131f9"},
	{"7ff59279f3035aab430548e09df44e704c0d721ad533b869ca80d4e1e62ad953", image.Rect(955, 834, 1160, 1193), image.Rect(4, 0, 23, 24), "fd6152a0b494d82b7dfa7487b5635d0e03c619d2db16b0db758737c23ebc7d2d"},
	{"8e1b514c2102b95bae36bd588078de0ddc1a8d6d40f52c1c8a23e584fee51556", image.Rect(80, 34, 561, 633), image.Rect(0, 1, 19, 24), "86699677da7aaee42a5c60851fe0fb02f56343b2cf45d1c7f70c2df2a59a32eb"},
	{"766bb52b18b0029d0014ed3ee9c5aebcacc062b2442382f0aabe8f826f6c9891", image.Rect(711, 39, 1231, 632), image.Rect(0, 1, 22, 24), "0ce6bfc40415d0fe41fc1b122f084be197eb6889f841285a266d0b8335ffcfc7"},
	{"c5f340d50151f37c2d7e6e4807f51d245f1fa835aaf032efa99da8256ea6f6e6", image.Rect(138, 43, 550, 602), image.Rect(8, 0, 23, 24), "e970902c8ce1d0f2885859d6b380433a6e800bc8e9eeab2223205c6b4866e267"},
	{"eaca017393e6e31a45a5f0fdcaca75aa7fdc6063d32c1426868608b0a6b802ad", image.Rect(650, 44, 1189, 604), image.Rect(2, 1, 23, 24), "caec7aca3634e9de886d03d348cf11ec6a4da793ea911820cd0ce1cfaa1d4c25"},
	{"abe90bbdc37844ae2a933e8a6b80ef96937c2b94de847b5005f6a1bb3a01b2ec", image.Rect(101, 650, 599, 1192), image.Rect(2, 0, 23, 24), "c816b12dd08174bc8bf73fe6144b210ec857ce5a1b9e31ca72beca29164d07e0"},
	{"889c704b4c20d24790ba26947fdf3aa243e89d3d6492ee9cf324df8a37cf84e0", image.Rect(111, 56, 577, 631), image.Rect(5, 1, 24, 24), "09ac6a3c8af69ebcf09d3f6402539a13b7d8a246d904351372d97df86ea0c416"},
	{"533e0dbc6391f4eaef8ff7c27789b2aabd6e4e61b5f793b3036e655c6a48bcd8", image.Rect(714, 58, 1226, 628), image.Rect(1, 0, 22, 24), "a8f19e7b255319103adec0dd8126329a360204ce21d55a44b56814e353b63f82"},
	{"c03571efe91aea6dc91514cf7b89a7f22752b34ecd6335cc1a9bef9aa687ade0", image.Rect(48, 25, 414, 448), image.Rect(2, 0, 23, 24), "029feb5a2a44cb4b567f6c73a73517b24d6991f0a333c52baaa4539a2b77e5a3"},
	{"acde2d3ef27ff5aca995429c975fc8452a4fc29b8601cc32c3e2553aa8a7c7a0", image.Rect(471, 156, 778, 450), image.Rect(1, 8, 19, 24), "fbb827921a3eec11404684330fa8d689d28cbe8a8b5a9ca5e778525a1ebed500"},
	{"f3c8eae54a8414476c042f57a5e390e93596da2d10954aec8a27357e4e7c25f2", image.Rect(994, 27, 1246, 453), image.Rect(1, 0, 20, 24), "6669a57cc06c72dc903ede9837fda3d5334da9f1c9acb132851ad3f90cfa7f1a"},
	{"e7ca0dc3722c9b053e578791f62f04f904d2a168aef8f82dd5ba4767a68c4685", image.Rect(124, 461, 326, 917), image.Rect(5, 0, 19, 24), "4a984f2c633477ac6117b4625fa996bc17836f5d8ae2df79d09a8c1af16f848a"},
	{"19fabad81a43edd6108a46f0dd091137bfcda2a9e04497277714af52ca002201", image.Rect(538, 515, 783, 895), image.Rect(4, 2, 20, 24), "f323c806506b263fac2d09ccb6a6ad91775b264fa41199389045a68ea12f8fd3"},
	{"55adeb9c9dd4bcdf75ae503ca91f564b325fd4331202f73fea8e224efed4987a", image.Rect(918, 617, 1238, 910), image.Rect(2, 8, 20, 24), "d976d2b4e8b2d6a4fb802fad3ac50bf64033542156041a48b69d2f6378fcd402"},
	{"eb7307bb6a7270d82ca058afc46655c942c3d00dce2a3bd8bb16f61a8c100b6a", image.Rect(32, 1079, 434, 1219), image.Rect(0, 16, 20, 24), "3f148dfa6019717f0eb981f0b5a15b5ecf3cedfdbfd9062982a1084ec0efced0"},
	{"0f679017ecc1056f351ffe5f593aebc8ab387104ad2bb48ef49153d57a728245", image.Rect(44, 88, 225, 348), image.Rect(3, 0, 22, 24), "02cec1ddac6029635b187c51d051bb3b76a0f3c672dcd04f4c5e754c3e9e3496"},
	{"fca3d24d4f7337fb31cd034a1ed667824aaf482cf03b41a3a463450eefe78a4c", image.Rect(302, 90, 485, 348), image.Rect(2, 0, 21, 24), "4e435b01dbf2dbf60632b89d446cb7f189ab1e79f082775c6ded74a51aeaeb79"},
	{"e9df75a21e30b86a2f74ebced52143a6fce79c84b5ab26e81f0feb1c50e9ec8c", image.Rect(549, 89, 753, 344), image.Rect(2, 0, 23, 24), "86397b7e2656c94325c8b8b11aa9f33af1e2536d8f3eb6ec556bb2295ec8931c"},
	{"a8cc961baa767a0b752281b020aeaa7651dcc9caaf268437c97ec3a0ef4fc31c", image.Rect(802, 95, 1018, 348), image.Rect(2, 1, 23, 24), "8e36ad6c0622f876fcf99f3bb83c39cf022fbe4ec8fd29a7f824c5d7ee27fa09"},
	{"f797754c9a7d69b389d3aef87b5879714ff66a587bc5c9d1db9140d8ddd38613", image.Rect(1072, 90, 1260, 341), image.Rect(2, 0, 23, 24), "e41e5bfe47df8d977f193217b86933fc5b882559f63a3e3da0dcb5573178644e"},
	{"dcc3d3643461ca18df1fe4ff582a92d8fcb94d67ee35552eb42d3b15785636c2", image.Rect(1311, 93, 1523, 343), image.Rect(2, 1, 24, 24), "290aa49e87bc93ab2aa1df0975ecd015b890500524ec0a354d324be144fbfead"},
	{"a1316100e601802e07dfd5b0b9e5b32bfc36358bc11e929afe6a7a219cb98b07", image.Rect(39, 385, 235, 648), image.Rect(2, 0, 21, 24), "245f37586d1563115dad73eb5bc2150824e84bbadce24d581e1128cd44fca0c1"},
	{"fca5774041878ef416c3d06a97e584b6398d0d0a4b6538cc722b4145440cc059", image.Rect(268, 378, 497, 643), image.Rect(2, 0, 21, 24), "05fe2e5f306b6787f14c8d6335260a2a2ccc7bba4c3c5f82e5961b5bb64998b6"},
	{"05080329443b32a15140f7df712fb93d6489f5889a6f85238654f6b4b4509c83", image.Rect(542, 383, 738, 646), image.Rect(1, 0, 22, 24), "5e2becb4736757eb36e10a8e1498459e001fa547b204f170809d3cffb7406d56"},
	{"a485b6781040d25678e79a8ff67f5e75ef35b074e17b1b25e5fa1ca5c0c4b241", image.Rect(766, 385, 1002, 643), image.Rect(0, 1, 22, 24), "d3c287bd2c417994b70a6e5a6983f998ddf5f5fb47aaa63ddd92c155d501532d"},
	{"a2fc5ff7f3d64f78c0334eca8c5f0b9fb470c5e899627df1175496d4a5f67c4e", image.Rect(1056, 383, 1250, 648), image.Rect(1, 0, 22, 24), "67e3c80510d76509716c1b1025eb0689e6c3faff8c954815c537d892a6e800d4"},
	{"4097a2d8e7a6cf82e3cfcede02240cf64f0f2ae173aba2fe7493f735e2d1e251", image.Rect(1306, 383, 1506, 648), image.Rect(3, 0, 22, 24), "0b931ffc5edc8632f5ed9a419fe9db1363d229f32948e30fb1de114725980a04"},
	{"03992760e6db0eec6b44cbf28dccb7276ff27f550e4e7946f89578df85bf2a0b", image.Rect(35, 691, 252, 949), image.Rect(3, 0, 22, 24), "b9010c85ea83c60ecc62e07877a2b6e122d4e70709dc16c83ab800e0117bd8a9"},
	{"2a055f848d8a173ec43e52ac48c2fedac1b6767674b3b1a17e979326f33aaa38", image.Rect(304, 679, 483, 918), image.Rect(1, 0, 22, 22), "f533c33e810cda5b0c829a8c89cedd2954715f72956f4bafe3c68aa53fe3258c"},
	{"58a554ead34d56d8d686d01d0418f8e09e94ef2dc72445216d172372fb9e923d", image.Rect(556, 683, 751, 933), image.Rect(1, 0, 22, 24), "f0768403ea3d64398f0bc21fbf825cce2253d4b31a49c4aafd251cda4dfa9a87"},
	{"3deabdb1bd9da91cc083ed10b806f3e403c10bdfd7a16c63872bbe04a98ff56c", image.Rect(808, 686, 1002, 945), image.Rect(1, 0, 22, 24), "09609caa8233801ccf13d298d29616bb384406bd911b15fdf6c5a55ec3ce6216"},
	{"0624a1da236e5645ab9dba5b0876f79bcc8c8c3656de96cef901d500262ccbde", image.Rect(1035, 759, 1277, 947), image.Rect(1, 8, 22, 24), "4b3a2dc5e3e746e1d0a9f792521bc944739d860be5ee65c16b01c1fb2846425b"},
	{"a119ce2b9025164280afcca26ca61fb31cf53bfc5760c46c5ca570b1a8cfac94", image.Rect(298, 79, 928, 1193), image.Rect(4, 0, 20, 24), "0a0a29b044de4c1a5f187d6443691570fe77d6648d31c2b1fc4282138a09e942"},
}

type cavePlayerPose struct {
	number int
	bounds image.Rectangle
	pixels []color.NRGBA
}
type CavePlayer struct {
	poses    map[string]*cavePlayerPose
	complete *CompleteArtwork
}

func (p *CavePlayer) SetCompleteArtwork(s *ManualArtwork) {
	if p != nil && s != nil {
		p.complete = s.complete
	}
}

type cavePlayerCrop struct {
	image.Image
	box image.Rectangle
}

func (c cavePlayerCrop) Bounds() image.Rectangle { return c.box }

// NewCavePlayer先驗八張固定圖面及全部55個取樣，再交出可用圖組。
func NewCavePlayer(dir string) (*CavePlayer, error) {
	atlas, err := caveArtwork(filepath.Join(dir, "player-atlas-handpaint-v1.png"), "a2ca6c6de0bdb07c7a97ae301ab66eab9c623b93da14ac5ec2cf7b2dc0930662")
	if err != nil {
		return nil, err
	}
	motion, err := caveArtwork(filepath.Join(dir, "player-motion-atlas-handpaint-v1.png"), "940e2e2b35dd6136824d74e725c40c66315f158669c7002638ac9db542da0c5b")
	if err != nil {
		return nil, err
	}
	gui, err := caveArtwork(filepath.Join(dir, "player-gui-atlas-handpaint-v1.png"), "dabf19f5149832968b1f1cd384874bd5f9f5a07de57e3e9a12532c9102907c17")
	if err != nil {
		return nil, err
	}
	left, err := caveArtwork(filepath.Join(dir, "player-left-atlas-handpaint-v2.png"), "bdaffc117f33e511b2e83ed394fc0095ca119cad0ce4ed82fa86b7fb79362cb3")
	if err != nil {
		return nil, err
	}
	extra, err := caveArtwork(filepath.Join(dir, "player-motion-extra-handpaint-v1.png"), "26f140767a0baf70930c0f96a112c23b2a407c7e943ec626b995dac8fd89ea67")
	if err != nil {
		return nil, err
	}
	expedition, err := caveArtwork(filepath.Join(dir, "player-expedition-atlas-handpaint-v1.png"), "7691d0862c5a98e49112ef2022ec5ac47529afc16a07221dc44e3a661925a0c3")
	if err != nil {
		return nil, err
	}
	player := &CavePlayer{poses: map[string]*cavePlayerPose{}}
	shieldData, err := checked(filepath.Join(dir, "player-shield-atlas-handpaint-v2.png"), "ea679274d3f144bef9841e660a26b0ec4a860957cb19989d98102c87f7f9633a", 16<<20)
	if err != nil {
		return nil, err
	}
	config, err := png.DecodeConfig(bytes.NewReader(shieldData))
	if err != nil || config.Width != 1536 || config.Height != 1024 {
		return nil, fmt.Errorf("持盾主角圖面尺寸不符")
	}
	shield, err := png.Decode(bytes.NewReader(shieldData))
	if err != nil {
		return nil, err
	}
	transition, err := caveArtwork(filepath.Join(dir, "player-shield-transition-handpaint-v1.png"), "3a7fa827fee261f0c9bcfd611fd8f2c19a0ed347218a976a52ccd9b08c771acb")
	if err != nil {
		return nil, err
	}
	for n, r := range cavePlayerRegistrations {
		source := atlas
		if n >= 14 {
			source = motion
		}
		if n >= 23 {
			source = gui
		}
		if n >= 25 {
			source = left
		}
		if n >= 28 {
			source = extra
		}
		if n >= 30 {
			source = expedition
		}
		if n >= 37 {
			source = shield
		}
		if n >= 54 {
			source = transition
		}
		if r.crop.Empty() || !r.crop.In(source.Bounds()) || r.native.Empty() || !r.native.In(image.Rect(0, 0, 24, 24)) || player.poses[r.key] != nil {
			return nil, fmt.Errorf("主角美術登錄不符")
		}
		box := image.Rectangle{Min: r.native.Min.Mul(3), Max: r.native.Max.Mul(3)}
		pose := &cavePlayerPose{number: n, bounds: box, pixels: make([]color.NRGBA, 72*72)}
		crop := cavePlayerCrop{Image: source, box: r.crop}
		for y := 0; y < box.Dy(); y++ {
			for x := 0; x < box.Dx(); x++ {
				pose.pixels[(box.Min.Y+y)*72+box.Min.X+x] = caveArea(crop, x, y, box.Dx(), box.Dy())
			}
		}
		raw := make([]byte, 0, 72*72*4)
		for _, c := range pose.pixels {
			raw = append(raw, c.R, c.G, c.B, c.A)
		}
		if townBackgroundHash(raw) != r.pixelsSHA {
			return nil, fmt.Errorf("主角直接高清取樣不符：%d", n)
		}
		player.poses[r.key] = pose
	}
	return player, nil
}

func (p *CavePlayer) selectPose(d *caveDecoded) (string, *cavePlayerPose, bool) {
	if p == nil || d == nil {
		return "", nil, false
	}
	key := townBackgroundHash(d.ActorRaster[:])
	// 原版後畫已完全遮住主角時，不因新美術的輪廓差異讓角色重現。
	if len(d.Actor) == 0 {
		return key, nil, true
	}
	pose := p.poses[key]
	if pose == nil && p.complete != nil {
		pixels, bounds := p.complete.compose("fman-all", d.ActorRaster[:], 24)
		if pixels != nil && !bounds.Empty() {
			pose = &cavePlayerPose{number: 20000, bounds: bounds, pixels: pixels}
			if len(p.poses) < 8192 {
				p.poses[key] = pose
			}
		}
	}
	if pose == nil || len(pose.pixels) != 72*72 || pose.bounds.Empty() || !pose.bounds.In(image.Rect(0, 0, 72, 72)) {
		return key, nil, false
	}
	native := image.Rect(d.ActorX*3, d.ActorY*3, d.ActorX*3+72, d.ActorY*3+72)
	game := image.Rect(144, 42, 816, 474)
	if !pose.bounds.Add(native.Min).In(game) {
		return key, nil, false
	}
	return key, pose, false
}

func cavePlayerPixelBlocked(x, y int, paints []CavePaint) bool {
	for _, p := range paints {
		dx, dy := x-p.X, y-p.Y
		if dx >= 0 && dx < 8 && dy >= 0 && dy < 8 && p.Mask&(uint64(1)<<uint(dy*8+dx)) != 0 {
			return true
		}
	}
	return false
}

func (p *cavePlayerPose) draw(dst *image.RGBA, d *caveDecoded, paints []CavePaint, origin image.Point) {
	for y := p.bounds.Min.Y; y < p.bounds.Max.Y; y++ {
		for x := p.bounds.Min.X; x < p.bounds.Max.X; x++ {
			c := p.pixels[y*72+x]
			if c.A == 0 || cavePlayerPixelBlocked(d.ActorX+x/3, d.ActorY+y/3, paints) {
				continue
			}
			px, py := d.ActorX*3+x+origin.X, d.ActorY*3+y+origin.Y
			old := dst.RGBAAt(px, py)
			a, inv := uint32(c.A), uint32(255-c.A)
			dst.SetRGBA(px, py, color.RGBA{uint8((uint32(c.R)*a + uint32(old.R)*inv + 127) / 255), uint8((uint32(c.G)*a + uint32(old.G)*inv + 127) / 255), uint8((uint32(c.B)*a + uint32(old.B)*inv + 127) / 255), 255})
		}
	}
}

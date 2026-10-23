// Package glyphpatch 產生字形目的段可改向的 GMMCGA.BIN，並在執行期把覆繪層已認領的
// 原版字形改寫到不顯示的 B000 段。原版目錄不動；patch 版只寫進使用者可寫層。
// 規格見 docs/spec/066-glyph-patch-stable-presentation.md。
package glyphpatch

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/dosgolem/oracle"
)

const (
	// DriverName 是原版 MCGA 驅動檔名；原版以小寫開檔，dosgolem 不分大小寫。
	DriverName     = "GMMCGA.BIN"
	OriginalSHA256 = "95aafd24d1300e8cb7fd66fc651839a968da1bdc7616c0ddf71d1337056b7579"
	driverBase     = 0x2000 // 載入於 01B1:2000
	driverSize     = 3273
	// VarOffset 是延長兩 bytes 的目的段變數，01B1 段內偏移；檔尾之後的區段在正常鍵序中零讀寫。
	VarOffset = 0x2cc9
	Visible   = 0xa000
	// Hidden 在真 VGA 的 mode 13h 不解碼，dosgolem 把它當一般 RAM，寫入不會顯示。
	Hidden = 0xb000
)

type site struct {
	at       uint16
	old, new []byte
}

// sites 是三處寫死 A000 的目的段載入。只把 mov ax,0A000h 換成 mov ax,[2CC9]：
// 指令數、DOSBox 週期與 AX 都與原版相同。2832 位於 27E9 的 push cs／pop ds 之後；
// 22FE、2333 的小字常式本身靠 DS 讀 F504 字型表與 2CBD 色號，原版正常運作時 DS 必為 01B1。
var sites = []site{
	{0x2832, []byte{0xb8, 0x00, 0xa0}, []byte{0xa1, 0xc9, 0x2c}},
	{0x22fe, []byte{0xb8, 0x00, 0xa0}, []byte{0xa1, 0xc9, 0x2c}},
	{0x2333, []byte{0xb8, 0x00, 0xa0}, []byte{0xa1, 0xc9, 0x2c}},
}

// Patched 由固定原版驅動產生 patch 版；SHA、長度或任一原 bytes 不符即拒絕。
func Patched(original []byte) ([]byte, error) {
	if fmt.Sprintf("%x", sha256.Sum256(original)) != OriginalSHA256 || len(original) != driverSize {
		return nil, fmt.Errorf("%s 版本不符", DriverName)
	}
	out := append([]byte(nil), original...)
	for _, s := range sites {
		at := int(s.at) - driverBase
		if !bytes.Equal(out[at:at+len(s.old)], s.old) {
			return nil, fmt.Errorf("%s %04X 原 bytes 不符", DriverName, s.at)
		}
		copy(out[at:], s.new)
	}
	return append(out, byte(Visible&0xff), byte(Visible>>8)), nil
}

// Install 在可寫層放入 patch 版驅動，回傳其 SHA-256。可寫層已有其他內容時拒絕覆寫。
func Install(originalDir, scratchDir string) (string, error) {
	original, err := os.ReadFile(filepath.Join(originalDir, DriverName))
	if err != nil {
		return "", err
	}
	patched, err := Patched(original)
	if err != nil {
		return "", err
	}
	sum := fmt.Sprintf("%x", sha256.Sum256(patched))
	target := filepath.Join(scratchDir, DriverName)
	if existing, err := os.ReadFile(target); err == nil {
		if bytes.Equal(existing, patched) {
			return sum, nil
		}
		return "", fmt.Errorf("可寫層已有其他 %s，不覆寫", DriverName)
	}
	tmp := target + ".tmp"
	if err = os.WriteFile(tmp, patched, 0600); err != nil {
		return "", err
	}
	return sum, os.Rename(tmp, target)
}

// Remove 只刪除本套件寫入的 patch 版，讓原版驅動恢復生效。
func Remove(originalDir, scratchDir string) error {
	original, err := os.ReadFile(filepath.Join(originalDir, DriverName))
	if err != nil {
		return err
	}
	patched, err := Patched(original)
	if err != nil {
		return err
	}
	target := filepath.Join(scratchDir, DriverName)
	existing, err := os.ReadFile(target)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !bytes.Equal(existing, patched) {
		return fmt.Errorf("可寫層的 %s 不是本工具產生，不刪除", DriverName)
	}
	return os.Remove(target)
}

// CodeView 讀原版程式碼；patch 位址目前是 patch 版內容時換回原 bytes，
// 讓既有的程式碼雜湊守衛照常比對原版。
func CodeView(o *oracle.Oracle, a oracle.Addr, n int) []byte {
	data := o.Bytes(a, n)
	if a.Seg != 0x1b1 {
		return data
	}
	for _, s := range sites {
		start, end := int(s.at)-int(a.Off), int(s.at)-int(a.Off)+len(s.new)
		if end <= 0 || start >= n || !bytes.Equal(o.Bytes(oracle.Far(0x1b1, s.at), len(s.new)), s.new) {
			continue
		}
		for i := max(start, 0); i < min(end, n); i++ {
			data[i] = s.old[i-start]
		}
	}
	return data
}

func patchedAt(o *oracle.Oracle, at uint16) bool {
	for _, s := range sites {
		if s.at == at {
			return bytes.Equal(o.Bytes(oracle.Far(0x1b1, at), len(s.new)), s.new)
		}
	}
	return false
}

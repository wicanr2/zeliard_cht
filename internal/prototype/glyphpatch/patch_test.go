package glyphpatch

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func originalDriver(t *testing.T) (string, []byte) {
	t.Helper()
	root := os.Getenv("ZELIARD_ORIG")
	if root == "" {
		t.Skip("須掛載原版")
	}
	data, err := os.ReadFile(filepath.Join(root, DriverName))
	if err != nil {
		t.Fatal(err)
	}
	return root, data
}

// TestPatchedOnlyChangesSites 核對 patch 版只改三處目的段載入並在檔尾附加 A000 變數。
func TestPatchedOnlyChangesSites(t *testing.T) {
	_, original := originalDriver(t)
	patched, err := Patched(original)
	if err != nil {
		t.Fatal(err)
	}
	if len(patched) != len(original)+2 || patched[len(original)] != 0x00 || patched[len(original)+1] != 0xa0 {
		t.Fatal("檔尾變數不是 A000")
	}
	changed := map[int]bool{}
	for _, s := range sites {
		at := int(s.at) - driverBase
		if !bytes.Equal(patched[at:at+len(s.new)], s.new) {
			t.Fatalf("%04X 未改成 mov ax,[2CC9]", s.at)
		}
		for i := range s.new {
			changed[at+i] = true
		}
	}
	for i := range original {
		if !changed[i] && patched[i] != original[i] {
			t.Fatalf("%04X 被改動", i+driverBase)
		}
	}
	broken := append([]byte(nil), original...)
	broken[0] ^= 1
	if _, err := Patched(broken); err == nil {
		t.Fatal("不同版本的驅動也被 patch")
	}
}

// TestInstallRemove 核對只寫可寫層、重複安裝不變、不覆寫或刪除他人檔案。
func TestInstallRemove(t *testing.T) {
	root, _ := originalDriver(t)
	scratch := t.TempDir()
	sum, err := Install(root, scratch)
	if err != nil || sum == "" {
		t.Fatal(err)
	}
	if again, err := Install(root, scratch); err != nil || again != sum {
		t.Fatal("重複安裝結果不同", err)
	}
	if err := Remove(root, scratch); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(scratch, DriverName)); !os.IsNotExist(err) {
		t.Fatal("移除後仍在")
	}
	foreign := filepath.Join(scratch, DriverName)
	if err := os.WriteFile(foreign, []byte("other"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(root, scratch); err == nil {
		t.Fatal("覆寫了可寫層既有的其他驅動")
	}
	if err := Remove(root, scratch); err == nil {
		t.Fatal("刪除了不是本工具產生的驅動")
	}
	if data, _ := os.ReadFile(foreign); string(data) != "other" {
		t.Fatal("他人檔案被改動")
	}
}

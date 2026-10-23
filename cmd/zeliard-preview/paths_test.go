package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInsideDirByIdentity(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "original")
	if err := os.MkdirAll(filepath.Join(root, "inner"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(base, "user"), 0700); err != nil {
		t.Fatal(err)
	}
	cases := map[string]bool{
		root:                                    true,
		filepath.Join(root, "inner"):            true,
		filepath.Join(root, "inner", "missing"): true,
		filepath.Join(base, "user"):             false,
		filepath.Join(base, "user", "missing"):  false,
		base:                                    false,
		filepath.Join(base, "original-other"):   false,
	}
	if runtime.GOOS == "windows" {
		// \\?\ 前綴與大小寫只改字串，不改檔案身份。
		cases[`\\?\`+filepath.Join(root, "inner")] = true
		cases[`\\?\`+filepath.Join(base, "user")] = false
		cases[strings.ToUpper(filepath.Join(root, "inner"))] = true
		if _, err := os.Stat(`Z:\`); err == nil && !strings.EqualFold(filepath.VolumeName(base), "Z:") {
			cases[`Z:\`] = false
		}
	} else {
		link := filepath.Join(base, "link")
		if err := os.Symlink(filepath.Join(root, "inner"), link); err != nil {
			t.Fatal(err)
		}
		cases[filepath.Join(link, "missing")] = true
	}
	for path, want := range cases {
		if got := insideDir(root, path); got != want {
			t.Errorf("insideDir(%q, %q)=%v，應為%v", root, path, got, want)
		}
	}
	if runtime.GOOS == "windows" {
		if !insideDir(`\\?\`+root, filepath.Join(root, "inner")) || insideDir(`\\?\`+root, filepath.Join(base, "user")) {
			t.Error(`\\?\ 原版目錄的身份比對錯誤`)
		}
	}
}

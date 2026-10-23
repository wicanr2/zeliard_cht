package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// insideDir 判斷 path 是否等於 root 或位於其下。逐層以檔案身份比對祖先目錄，
// 不受 Windows 磁碟區名稱、\\?\ 前綴、大小寫或短檔名的字串差異影響；
// 不同磁碟區不會有相同身份，因此視為不在原版目錄內。
func insideDir(root, path string) bool {
	rootInfo, err := os.Stat(root)
	if err != nil {
		rel, e := filepath.Rel(root, path)
		return e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
	}
	starts := []string{filepath.Clean(path)}
	// 符號連結或 junction 要解析最深的既有祖先，才看得出真正位置。
	for p := starts[0]; ; p = filepath.Dir(p) {
		if real, e := filepath.EvalSymlinks(p); e == nil {
			starts = append(starts, real)
			break
		}
		if filepath.Dir(p) == p {
			break
		}
	}
	for _, p := range starts {
		for {
			if info, e := os.Stat(p); e == nil && os.SameFile(rootInfo, info) {
				return true
			}
			parent := filepath.Dir(p)
			if parent == p {
				break
			}
			p = parent
		}
	}
	return false
}

// catalogIdentity 回傳譯表的絕對路徑與 SHA-256 前 16 碼，供日誌核對實際載入的版本。
func catalogIdentity(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return abs + " 無法讀取：" + err.Error()
	}
	return fmt.Sprintf("%s sha256=%x", abs, sha256.Sum256(data))[:len(abs)+24]
}

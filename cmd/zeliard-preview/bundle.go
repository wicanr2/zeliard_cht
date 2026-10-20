package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// bundleArguments 只在未指定原版目錄時啟用，讓三平台封包可直接啟動。
func bundleArguments() error {
	for _, arg := range os.Args[1:] {
		if arg == "-version" || arg == "--version" || arg == "-orig" || strings.HasPrefix(arg, "-orig=") {
			return nil
		}
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	base := filepath.Dir(exe)
	for _, candidate := range []string{base, filepath.Join(base, "../Resources/game")} {
		if _, e := os.Stat(filepath.Join(candidate, "runtime-args.json")); e == nil {
			base = candidate
			break
		}
	}
	raw, err := os.ReadFile(filepath.Join(base, "runtime-args.json"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var entries [][2]string
	if err = json.Unmarshal(raw, &entries); err != nil {
		return err
	}
	user := os.Getenv("ZELIARD_USER_DIR")
	if user == "" {
		root, e := os.UserConfigDir()
		if e != nil {
			return e
		}
		user = filepath.Join(root, "Zeliard-CHT")
	}
	if err = os.MkdirAll(filepath.Join(user, "game"), 0700); err != nil {
		return err
	}
	target := filepath.Join(user, "game/SAVETEST.usr")
	if _, e := os.Stat(target); os.IsNotExist(e) {
		source := filepath.Join(base, "starter/SAVETEST.usr")
		data, e := os.ReadFile(source)
		if e != nil {
			return e
		}
		f, e := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if e == nil {
			_, e = f.Write(data)
			closeErr := f.Close()
			if e != nil {
				return e
			}
			if closeErr != nil {
				return closeErr
			}
			if st, e := os.Stat(source); e == nil {
				_ = os.Chtimes(target, st.ModTime(), st.ModTime())
			}
		} else if !os.IsExist(e) {
			return e
		}
	}
	args := []string{"-orig", filepath.Join(base, "original"), "-font", filepath.Join(base, "font.hex.gz"), "-user-dir", user}
	for _, e := range entries {
		path := filepath.Join(base, filepath.FromSlash(e[1]))
		if _, err = os.Stat(path); err != nil {
			return fmt.Errorf("封包缺少 %s：%w", e[1], err)
		}
		args = append(args, e[0], path)
	}
	if raw, e := os.ReadFile(filepath.Join(base, "bundle-options.json")); e == nil {
		var options []string
		if e = json.Unmarshal(raw, &options); e != nil {
			return e
		}
		args = append(args, options...)
	}
	// 使用者命令列在最後，仍可覆寫封包預設。
	os.Args = append(append([]string{os.Args[0]}, args...), os.Args[1:]...)
	return nil
}

//go:build linux || darwin

package fm

import "github.com/ebitengine/purego"

func openLibrary(path string) (uintptr, error) {
	return purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_LOCAL)
}
func closeLibrary(handle uintptr) { _ = purego.Dlclose(handle) }

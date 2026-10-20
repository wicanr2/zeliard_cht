package fm

import "syscall"

func openLibrary(path string) (uintptr, error) {
	h, e := syscall.LoadLibrary(path)
	return uintptr(h), e
}
func closeLibrary(handle uintptr) { _ = syscall.FreeLibrary(syscall.Handle(handle)) }

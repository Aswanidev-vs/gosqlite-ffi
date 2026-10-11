//go:build linux || darwin || freebsd || openbsd || netbsd

package sqliteffi

import (
	"errors"
	"unsafe"

	"github.com/ebitengine/purego"
)

func openLibrary(path string) (uintptr, error) {
	return purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_LOCAL)
}

func closeLibrary(handle uintptr) error {
	return purego.Dlclose(handle)
}

func registerLibVersion(l *Library, handle uintptr) error {
	symbol, err := purego.Dlsym(handle, "sqlite3_libversion")
	if err != nil {
		return err
	}
	if symbol == 0 {
		return errors.New("symbol address is zero")
	}
	purego.RegisterFunc(&l.version, symbol)
	return nil
}

func cString(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	var b []byte
	for i := uintptr(0); ; i++ {
		v := *(*byte)(unsafe.Pointer(ptr + i))
		if v == 0 {
			break
		}
		b = append(b, v)
	}
	return string(b)
}

//go:build windows

package sqliteffi

import (
	"errors"
	"syscall"
	"unsafe"

	"github.com/ebitengine/purego"
)

func openLibrary(path string) (uintptr, error) {
	h, err := syscall.LoadLibrary(path)
	return uintptr(h), err
}

func closeLibrary(handle uintptr) error {
	return syscall.FreeLibrary(syscall.Handle(handle))
}

func registerLibVersion(l *Library, handle uintptr) error {
	symbol, err := syscall.GetProcAddress(syscall.Handle(handle), "sqlite3_libversion")
	if err != nil {
		return err
	}
	if symbol == 0 {
		return errors.New("symbol address is zero")
	}
	purego.RegisterFunc(&l.version, uintptr(symbol))
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

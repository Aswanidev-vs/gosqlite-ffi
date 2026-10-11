//go:build windows

package sqliteffi

import (
	"errors"
	"fmt"
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
	versionSymbol, err := syscall.GetProcAddress(syscall.Handle(handle), "sqlite3_libversion")
	if err != nil {
		return fmt.Errorf("sqlite3_libversion: %w", err)
	}
	if versionSymbol == 0 {
		return errors.New("sqlite3_libversion symbol address is zero")
	}

	versionNumberSymbol, err := syscall.GetProcAddress(syscall.Handle(handle), "sqlite3_libversion_number")
	if err != nil {
		return fmt.Errorf("sqlite3_libversion_number: %w", err)
	}
	if versionNumberSymbol == 0 {
		return errors.New("sqlite3_libversion_number symbol address is zero")
	}

	purego.RegisterFunc(&l.version, uintptr(versionSymbol))
	purego.RegisterFunc(&l.versionNumber, uintptr(versionNumberSymbol))
	return registerDatabaseAPI(l, func(name string) (uintptr, error) { p, err := syscall.GetProcAddress(syscall.Handle(handle), name); return uintptr(p), err })
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

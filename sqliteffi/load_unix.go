//go:build linux || darwin || freebsd || openbsd || netbsd

package sqliteffi

import (
	"errors"
	"fmt"
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
	versionSymbol, err := purego.Dlsym(handle, "sqlite3_libversion")
	if err != nil {
		return fmt.Errorf("sqlite3_libversion: %w", err)
	}
	if versionSymbol == 0 {
		return errors.New("sqlite3_libversion symbol address is zero")
	}

	versionNumberSymbol, err := purego.Dlsym(handle, "sqlite3_libversion_number")
	if err != nil {
		return fmt.Errorf("sqlite3_libversion_number: %w", err)
	}
	if versionNumberSymbol == 0 {
		return errors.New("sqlite3_libversion_number symbol address is zero")
	}

	purego.RegisterFunc(&l.version, versionSymbol)
	purego.RegisterFunc(&l.versionNumber, versionNumberSymbol)
	return registerDatabaseAPI(l, func(name string) (uintptr, error) { return purego.Dlsym(handle, name) })
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

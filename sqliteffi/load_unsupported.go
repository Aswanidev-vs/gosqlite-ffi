//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd && !windows

package sqliteffi

import "errors"

func openLibrary(string) (uintptr, error) {
	return 0, errors.New("dynamic library loading is not implemented for this OS")
}
func closeLibrary(uintptr) error { return nil }
func registerLibVersion(*Library, uintptr) error {
	return errors.New("dynamic library loading is not implemented for this OS")
}
func cString(uintptr) string { return "" }

// Package sqliteffi provides a small experimental CGO-free binding to the
// SQLite C library. The initial API exposes version diagnostics; database
// operations are not yet implemented.
package sqliteffi

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

)

// ErrClosed is returned when a method is called after the native library is closed.
var ErrClosed = errors.New("sqliteffi: library is closed")

// LibraryName returns the platform-default candidate names used by LoadDefault.
// The list is deliberately small and predictable. Set an explicit path with
// Load when an application needs to control which SQLite build is loaded.
func LibraryName() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"sqlite3.dll", "winsqlite3.dll"}
	case "darwin":
		return []string{"libsqlite3.dylib", "/usr/lib/libsqlite3.dylib"}
	default:
		return []string{"libsqlite3.so.0", "libsqlite3.so", "libsqlite3.so.3"}
	}
}

// Library represents a loaded SQLite shared library. It does not represent a
// database connection. Close prevents further calls through this wrapper and
// releases the dynamic-library handle according to the operating system.
// Calls through this wrapper are serialized with Close.
type Library struct {
	mu            sync.Mutex
	handle        uintptr
	version       func() uintptr
	versionNumber func() int32
	closed        bool
}

// Load opens an explicit native library path and resolves SQLite version
// diagnostics. The caller should pass a trusted, absolute path when the
// environment may be influenced by untrusted users.
func Load(path string) (*Library, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("sqliteffi: library path must not be empty")
	}
	return loadLibrary(path)
}

// LoadDefault tries the standard candidate library names for the current OS.
// For deterministic deployments, prefer Load with an absolute path to a known
// SQLite build.
func LoadDefault() (*Library, error) {
	var failures []string
	for _, name := range LibraryName() {
		lib, err := loadLibrary(name)
		if err == nil {
			return lib, nil
		}
		failures = append(failures, fmt.Sprintf("%s: %v", name, err))
	}
	return nil, fmt.Errorf("sqliteffi: unable to load SQLite library (%s)", strings.Join(failures, "; "))
}

func loadLibrary(path string) (*Library, error) {
	handle, err := openLibrary(path)
	if err != nil {
		return nil, fmt.Errorf("sqliteffi: load %q: %w", path, err)
	}
	lib := &Library{handle: handle}
	if err := registerLibVersion(lib, handle); err != nil {
		_ = closeLibrary(handle)
		return nil, fmt.Errorf("sqliteffi: resolve SQLite version diagnostics in %q: %w", path, err)
	}
	return lib, nil
}

// Version returns the runtime SQLite version string, such as "3.45.1".
// It is safe to call from multiple goroutines and returns ErrClosed after Close.
func (l *Library) Version() (string, error) {
	if l == nil {
		return "", ErrClosed
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed || l.version == nil {
		return "", ErrClosed
	}
	ptr := l.version()
	if ptr == 0 {
		return "", errors.New("sqliteffi: sqlite3_libversion returned a null pointer")
	}
	return cString(ptr), nil
}

// VersionNumber returns SQLite's numeric version, such as 3045001 for 3.45.1.
// The number follows SQLite's sqlite3_libversion_number convention:
// major*1000000 + minor*1000 + patch. It returns ErrClosed after Close.
func (l *Library) VersionNumber() (int, error) {
	if l == nil {
		return 0, ErrClosed
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed || l.versionNumber == nil {
		return 0, ErrClosed
	}
	return int(l.versionNumber()), nil
}

// Close releases the native-library handle. It is idempotent.
func (l *Library) Close() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	l.closed = true
	l.version = nil
	l.versionNumber = nil
	if l.handle == 0 {
		return nil
	}
	err := closeLibrary(l.handle)
	l.handle = 0
	if err != nil {
		return fmt.Errorf("sqliteffi: close library: %w", err)
	}
	return nil
}

// Platform loading and native calls are defined in files selected by Go build tags.

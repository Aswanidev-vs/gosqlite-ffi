// Package sqliteffi provides a small experimental CGO-free binding to the
// SQLite C library. The initial API deliberately exposes only a version
// diagnostic; database operations are not yet implemented.
package sqliteffi

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/ebitengine/purego"
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
// Callers must not race Close with Version; the type serializes both operations.
type Library struct {
	mu      sync.Mutex
	handle  uintptr
	version func() uintptr
	closed  bool
}

// Load opens an explicit native library path and resolves sqlite3_libversion.
// The caller should pass a trusted, absolute path when the environment may be
// influenced by untrusted users.
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
		return nil, fmt.Errorf("sqliteffi: resolve sqlite3_libversion in %q: %w", path, err)
	}
	return lib, nil
}

// Version returns the runtime SQLite version string. It is safe to call from
// multiple goroutines and returns ErrClosed after Close.
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

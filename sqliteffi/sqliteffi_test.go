package sqliteffi

import (
	"errors"
	"strings"
	"testing"
)

func TestLoadRejectsBlankPath(t *testing.T) {
	for _, path := range []string{"", "  ", "\n\t"} {
		if _, err := Load(path); err == nil {
			t.Fatalf("Load(%q) expected an error", path)
		}
	}
}

func TestNilLibraryReturnsClosed(t *testing.T) {
	var lib *Library
	if _, err := lib.Version(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Version error = %v, want ErrClosed", err)
	}
	if _, err := lib.VersionNumber(); !errors.Is(err, ErrClosed) {
		t.Fatalf("VersionNumber error = %v, want ErrClosed", err)
	}
	if err := lib.Close(); err != nil {
		t.Fatalf("Close(nil) error = %v", err)
	}
}

func TestLibraryNameCandidates(t *testing.T) {
	names := LibraryName()
	if len(names) == 0 {
		t.Fatal("expected at least one platform candidate")
	}
	for _, n := range names {
		if strings.TrimSpace(n) == "" {
			t.Fatal("library candidate must not be blank")
		}
	}
}

func TestVersionNumber(t *testing.T) {
	lib := &Library{versionNumber: func() int32 { return 3045001 }}
	got, err := lib.VersionNumber()
	if err != nil {
		t.Fatalf("VersionNumber() error = %v", err)
	}
	if got != 3045001 {
		t.Fatalf("VersionNumber() = %d, want 3045001", got)
	}
}

func TestVersionNumberClosedAfterClose(t *testing.T) {
	lib := &Library{versionNumber: func() int32 { return 3045001 }}
	if err := lib.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := lib.VersionNumber(); !errors.Is(err, ErrClosed) {
		t.Fatalf("VersionNumber error = %v, want ErrClosed", err)
	}
}

func TestCloseIdempotent(t *testing.T) {
	lib := &Library{closed: false}
	if err := lib.Close(); err != nil {
		t.Fatal(err)
	}
	if err := lib.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := lib.Version(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Version error = %v, want ErrClosed", err)
	}
}

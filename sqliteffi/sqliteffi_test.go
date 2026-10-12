package sqliteffi

import (
	"errors"
	"strings"
	"testing"
	"unsafe"
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

func TestBindEmptyBlobIsDistinctFromNull(t *testing.T) {
	var gotBlobPointer unsafe.Pointer
	var gotBlobLength int32
	nullBinds := 0
	lib := &Library{}
	lib.api.bindNull = func(unsafe.Pointer, int32) int32 {
		nullBinds++
		return sqliteOK
	}
	lib.api.bindBlob = func(_ unsafe.Pointer, _ int32, p unsafe.Pointer, n int32, destructor uintptr) int32 {
		gotBlobPointer = p
		gotBlobLength = n
		if destructor != sqliteTransient {
			t.Errorf("blob destructor = %v, want SQLITE_TRANSIENT", destructor)
		}
		return sqliteOK
	}
	db := &DB{lib: lib}
	var stmt int

	if err := db.bind(unsafe.Pointer(&stmt), 1, []byte{}); err != nil {
		t.Fatalf("bind empty blob: %v", err)
	}
	if gotBlobPointer == nil {
		t.Fatal("empty non-nil []byte was bound with a nil pointer, which SQLite interprets as SQL NULL")
	}
	if gotBlobLength != 0 {
		t.Fatalf("empty blob length = %d, want 0", gotBlobLength)
	}
	if nullBinds != 0 {
		t.Fatalf("empty blob invoked bindNull %d times", nullBinds)
	}

	if err := db.bind(unsafe.Pointer(&stmt), 1, []byte(nil)); err != nil {
		t.Fatalf("bind nil byte slice: %v", err)
	}
	if nullBinds != 1 {
		t.Fatalf("nil []byte bindNull calls = %d, want 1", nullBinds)
	}
}


func TestValidateScanDestinationsRejectsNilValues(t *testing.T) {
 type unsupported struct{ value int }

 var nilInt *int64
 var nilString *string
 var nilBytes *[]byte
 var nilAny *any
 var nilUnsupported *unsupported

 tests := []struct {
  name string
  dest any
 }{
  {name: "nil interface", dest: nil},
  {name: "nil int64 pointer", dest: nilInt},
  {name: "nil string pointer", dest: nilString},
  {name: "nil byte-slice pointer", dest: nilBytes},
  {name: "nil any pointer", dest: nilAny},
  {name: "nil unsupported pointer", dest: nilUnsupported},
 }
 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
   err := validateScanDestinations([]any{tt.dest})
   if err == nil {
    t.Fatal("expected nil destination to be rejected")
   }
  })
 }
}

func TestScanRejectsNilDestinationBeforeAccessingRows(t *testing.T) {
 var rows *Rows
 var destination *int64

 err := rows.Scan(destination)
 if err == nil {
  t.Fatal("expected nil destination error")
 }
 if !strings.Contains(err.Error(), "destination 0") {
  t.Fatalf("error = %q, want destination index", err)
 }
}

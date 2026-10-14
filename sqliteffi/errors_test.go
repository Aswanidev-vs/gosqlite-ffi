package sqliteffi

import (
	"errors"
	"fmt"
	"testing"
)

func TestSQLiteErrorSupportsErrorsIs(t *testing.T) {
	err := &SQLiteError{
		Code:       CodeConstraint,
		ResultCode: int(CodeConstraint) | (8 << 8),
		Message:    "UNIQUE constraint failed: item.id",
	}

	if !errors.Is(err, CodeConstraint) {
		t.Fatal("errors.Is(error, CodeConstraint) = false, want true")
	}
	if errors.Is(err, CodeBusy) {
		t.Fatal("errors.Is(error, CodeBusy) = true, want false")
	}
	if got, want := err.Error(), fmt.Sprintf("sqliteffi: SQLite error %d: %s", err.ResultCode, err.Message); got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestSQLiteErrorUnknownCode(t *testing.T) {
	const code SQLiteCode = 999
	if got, want := code.Error(), "sqliteffi: SQLite result code 999"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestSQLiteErrorMatchesPrimaryCodeForExtendedResult(t *testing.T) {
	err := &SQLiteError{
		Code:       CodeConstraint,
		ResultCode: int(CodeConstraint) | (8 << 8),
		Message:    "unique constraint",
	}
	if err.Code != SQLiteCode(err.ResultCode&0xff) {
		t.Fatalf("primary code = %d, result code = %d", err.Code, err.ResultCode)
	}
	if !errors.Is(err, CodeConstraint) {
		t.Fatal("extended constraint error did not match primary CodeConstraint")
	}
}

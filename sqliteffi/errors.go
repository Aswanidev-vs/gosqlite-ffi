package sqliteffi

import "fmt"

// SQLite result codes returned by SQLite's C API.
const (
	CodeError      SQLiteCode = 1
	CodeInternal   SQLiteCode = 2
	CodePerm       SQLiteCode = 3
	CodeAbort      SQLiteCode = 4
	CodeBusy       SQLiteCode = 5
	CodeLocked     SQLiteCode = 6
	CodeNoMem      SQLiteCode = 7
	CodeReadOnly   SQLiteCode = 8
	CodeInterrupt  SQLiteCode = 9
	CodeIOErr      SQLiteCode = 10
	CodeCorrupt    SQLiteCode = 11
	CodeNotFound   SQLiteCode = 12
	CodeFull       SQLiteCode = 13
	CodeCantOpen   SQLiteCode = 14
	CodeProtocol   SQLiteCode = 15
	CodeEmpty      SQLiteCode = 16
	CodeSchema     SQLiteCode = 17
	CodeTooBig     SQLiteCode = 18
	CodeConstraint SQLiteCode = 19
	CodeMismatch   SQLiteCode = 20
	CodeMisuse     SQLiteCode = 21
	CodeNoLFS      SQLiteCode = 22
	CodeAuth       SQLiteCode = 23
	CodeFormat     SQLiteCode = 24
	CodeRange      SQLiteCode = 25
	CodeNotADB     SQLiteCode = 26
	CodeNotice     SQLiteCode = 27
	CodeWarning    SQLiteCode = 28
)

// SQLiteCode identifies a primary SQLite result code. It implements error so
// callers can use errors.Is(err, sqliteffi.CodeBusy), for example.
type SQLiteCode int

// Error returns a stable name for common SQLite result codes.
func (c SQLiteCode) Error() string {
	if name, ok := sqliteCodeNames[c]; ok {
		return "sqliteffi: " + name
	}
	return fmt.Sprintf("sqliteffi: SQLite result code %d", c)
}

var sqliteCodeNames = map[SQLiteCode]string{
	CodeError: "SQLITE_ERROR", CodeInternal: "SQLITE_INTERNAL",
	CodePerm: "SQLITE_PERM", CodeAbort: "SQLITE_ABORT", CodeBusy: "SQLITE_BUSY",
	CodeLocked: "SQLITE_LOCKED", CodeNoMem: "SQLITE_NOMEM", CodeReadOnly: "SQLITE_READONLY",
	CodeInterrupt: "SQLITE_INTERRUPT", CodeIOErr: "SQLITE_IOERR", CodeCorrupt: "SQLITE_CORRUPT",
	CodeNotFound: "SQLITE_NOTFOUND", CodeFull: "SQLITE_FULL", CodeCantOpen: "SQLITE_CANTOPEN",
	CodeProtocol: "SQLITE_PROTOCOL", CodeEmpty: "SQLITE_EMPTY", CodeSchema: "SQLITE_SCHEMA",
	CodeTooBig: "SQLITE_TOOBIG", CodeConstraint: "SQLITE_CONSTRAINT", CodeMismatch: "SQLITE_MISMATCH",
	CodeMisuse: "SQLITE_MISUSE", CodeNoLFS: "SQLITE_NOLFS", CodeAuth: "SQLITE_AUTH",
	CodeFormat: "SQLITE_FORMAT", CodeRange: "SQLITE_RANGE", CodeNotADB: "SQLITE_NOTADB",
	CodeNotice: "SQLITE_NOTICE", CodeWarning: "SQLITE_WARNING",
}

// SQLiteError describes a failed SQLite operation. Code is the primary result
// code (the low eight bits); ResultCode preserves the exact code returned by
// the SQLite API call, which may include extended-code bits when enabled.
// Message is SQLite's diagnostic message for the connection.
type SQLiteError struct {
	Code       SQLiteCode
	ResultCode int
	Message    string
}

// Error formats the SQLite result code and native diagnostic message.
func (e *SQLiteError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("sqliteffi: SQLite error %d: %s", e.ResultCode, e.Message)
}

// Is lets errors.Is match a SQLiteCode against the primary code. This means
// errors.Is(err, sqliteffi.CodeConstraint) also matches extended constraint codes.
func (e *SQLiteError) Is(target error) bool {
	code, ok := target.(SQLiteCode)
	return ok && e != nil && e.Code == code
}

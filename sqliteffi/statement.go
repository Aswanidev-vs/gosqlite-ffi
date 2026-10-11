package sqliteffi

import (
	"errors"
	"fmt"
)

// ErrStatementBusy indicates that a statement still has an open result set.
var ErrStatementBusy = errors.New("sqliteffi: statement has active rows")

// Stmt is a reusable native SQLite prepared statement. A statement is not safe
// for concurrent use; calls are serialized by its owning DB.
type Stmt struct {
	db *DB
	handle unsafe.Pointer
	active bool
	closed bool
}

// Prepare compiles SQL into a reusable native SQLite statement.
func (db *DB) Prepare(sql string) (*Stmt, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if err := db.checkOpen(); err != nil { return nil, err }
	handle, err := db.prepare(sql, nil)
	if err != nil { return nil, err }
	return &Stmt{db: db, handle: handle}, nil
}

// Close finalizes the native statement. Close its Rows before closing it.
func (s *Stmt) Close() error {
	if s == nil || s.db == nil { return nil }
	s.db.mu.Lock()
	defer s.db.mu.Unlock()
	if s.closed { return nil }
	if s.active { return ErrStatementBusy }
	if err := s.db.checkOpen(); err != nil { return err }
	rc := s.db.lib.api.finalize(s.handle)
	s.handle = nil
	s.closed = true
	if rc != sqliteOK { return s.db.error(rc) }
	return nil
}

func (s *Stmt) bindArgs(args []any) error {
	a := s.db.lib.api
	if rc := a.reset(s.handle); rc != sqliteOK { return s.db.error(rc) }
	if rc := a.clearBindings(s.handle); rc != sqliteOK { return s.db.error(rc) }
	for i, arg := range args {
		if err := s.db.bind(s.handle, int32(i+1), arg); err != nil {
			return fmt.Errorf("sqliteffi: bind parameter %d: %w", i+1, err)
		}
	}
	return nil
}

// Exec executes the prepared statement and resets it for reuse.
func (s *Stmt) Exec(args ...any) (Result, error) {
	if s == nil || s.db == nil { return Result{}, ErrDBClosed }
	db := s.db
	db.mu.Lock()
	defer db.mu.Unlock()
	if err := db.checkOpen(); err != nil { return Result{}, err }
	if s.closed { return Result{}, errors.New("sqliteffi: statement is closed") }
	if s.active { return Result{}, ErrStatementBusy }
	if err := s.bindArgs(args); err != nil { return Result{}, err }
	for {
		rc := db.lib.api.step(s.handle)
		switch rc {
		case sqliteRow:
			continue
		case sqliteDone:
			result := Result{LastInsertID: db.lib.api.lastInsertRowID(db.handle), RowsAffected: int64(db.lib.api.changes(db.handle))}
			_ = db.lib.api.reset(s.handle)
			return result, nil
		default:
			err := db.error(rc)
			_ = db.lib.api.reset(s.handle)
			return Result{}, err
		}
	}
}

// Query executes the prepared statement and returns rows. Close Rows before
// reusing or closing the statement.
func (s *Stmt) Query(args ...any) (*Rows, error) {
	if s == nil || s.db == nil { return nil, ErrDBClosed }
	db := s.db
	db.mu.Lock()
	defer db.mu.Unlock()
	if err := db.checkOpen(); err != nil { return nil, err }
	if s.closed { return nil, errors.New("sqliteffi: statement is closed") }
	if s.active { return nil, ErrStatementBusy }
	if err := s.bindArgs(args); err != nil { return nil, err }
	rows := &Rows{db: db, stmt: s.handle, owner: s}
	count := db.lib.api.columnCount(s.handle)
	rows.columns = make([]string, int(count))
	for i := range rows.columns { rows.columns[i] = cString(db.lib.api.columnName(s.handle, int32(i))) }
	s.active = true
	db.activeRows++
	return rows, nil
}


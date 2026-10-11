package sqliteffi

import (
	"errors"
	"sync"
)

// ErrTxDone is returned when a transaction has already been committed or rolled back.
var ErrTxDone = errors.New("sqliteffi: transaction has already completed")

// Tx represents a transaction on a DB connection. Always call Commit or Rollback.
// As SQLite transactions are connection-scoped, do not use the DB concurrently
// from unrelated goroutines while a transaction is active.
type Tx struct {
	db *DB
	mu sync.Mutex
	done bool
}

// Begin starts a deferred SQLite transaction.
func (db *DB) Begin() (*Tx, error) {
	if _, err := db.Exec("BEGIN"); err != nil { return nil, err }
	return &Tx{db: db}, nil
}

// Exec runs a statement within the transaction.
func (tx *Tx) Exec(sql string, args ...any) (Result, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.done { return Result{}, ErrTxDone }
	return tx.db.Exec(sql, args...)
}

// Query runs a query within the transaction.
func (tx *Tx) Query(sql string, args ...any) (*Rows, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.done { return nil, ErrTxDone }
	return tx.db.Query(sql, args...)
}

// Commit commits the transaction.
func (tx *Tx) Commit() error { return tx.finish("COMMIT") }

// Rollback rolls the transaction back.
func (tx *Tx) Rollback() error { return tx.finish("ROLLBACK") }

func (tx *Tx) finish(command string) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.done { return ErrTxDone }
	if _, err := tx.db.Exec(command); err != nil { return err }
	tx.done = true
	return nil
}

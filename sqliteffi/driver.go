package sqliteffi

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
)

// Driver implements database/sql/driver.Driver. The DSN is a SQLite database
// path (use ":memory:" for an in-memory database). Register it with sql.Register.
type Driver struct{}

// Open opens a database whose name is the SQLite path/DSN.
func (Driver) Open(name string) (driver.Conn, error) {
	db, err := Open(name)
	if err != nil { return nil, err }
	return &driverConn{db: db}, nil
}

type driverConn struct { db *DB; closed bool }

func (c *driverConn) Prepare(query string) (driver.Stmt, error) {
	if c.closed { return nil, ErrDBClosed }
	stmt, err := c.db.Prepare(query)
	if err != nil { return nil, err }
	return &driverStmt{stmt: stmt}, nil
}
func (c *driverConn) Close() error {
	if c.closed { return nil }
	c.closed = true
	return c.db.Close()
}
func (c *driverConn) Begin() (driver.Tx, error) {
	if c.closed { return nil, ErrDBClosed }
	return c.db.Begin()
}
func (c *driverConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if err := ctx.Err(); err != nil { return nil, err }
	if opts.ReadOnly { return nil, errors.New("sqliteffi: read-only transactions are not supported") }
	if opts.Isolation != driver.IsolationLevel(0) && opts.Isolation != driver.IsolationLevel(sql.LevelDefault) {
		return nil, fmt.Errorf("sqliteffi: unsupported isolation level %d", opts.Isolation)
	}
	return c.Begin()
}
func (c *driverConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if err := ctx.Err(); err != nil { return nil, err }
	values := make([]any, len(args))
	for i, a := range args { values[i] = a.Value }
	result, err := c.db.Exec(query, values...)
	if err != nil { return nil, err }
	return driverResult{result: result}, nil
}
func (c *driverConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if err := ctx.Err(); err != nil { return nil, err }
	values := make([]any, len(args))
	for i, a := range args { values[i] = a.Value }
	rows, err := c.db.Query(query, values...)
	if err != nil { return nil, err }
	return &driverRows{rows: rows}, nil
}

type driverStmt struct { stmt *Stmt }
func (s *driverStmt) Close() error { return s.stmt.Close() }
func (s *driverStmt) NumInput() int { return -1 }
func (s *driverStmt) Exec(args []driver.Value) (driver.Result, error) {
	values := make([]any, len(args)); for i, a := range args { values[i] = a }
	r, err := s.stmt.Exec(values...); if err != nil { return nil, err }; return driverResult{result: r}, nil
}
func (s *driverStmt) Query(args []driver.Value) (driver.Rows, error) {
	values := make([]any, len(args)); for i, a := range args { values[i] = a }
	r, err := s.stmt.Query(values...); if err != nil { return nil, err }; return &driverRows{rows:r}, nil
}

type driverResult struct { result Result }
func (r driverResult) LastInsertId() (int64, error) { return r.result.LastInsertID, nil }
func (r driverResult) RowsAffected() (int64, error) { return r.result.RowsAffected, nil }

type driverRows struct { rows *Rows }
func (r *driverRows) Columns() []string { return r.rows.Columns() }
func (r *driverRows) Close() error { return r.rows.Close() }
func (r *driverRows) Next(dest []driver.Value) error {
	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil { return err }
		return io.EOF
	}
	values, err := r.rows.Values(); if err != nil { return err }
	if len(dest) != len(values) { return fmt.Errorf("sqliteffi: driver row has %d columns, destination has %d", len(values), len(dest)) }
	for i, v := range values {
		switch x := v.(type) {
		case nil: dest[i] = nil
		case int64: dest[i] = x
		case float64: dest[i] = x
		case string: dest[i] = x
		case []byte: dest[i] = append([]byte(nil), x...)
		default: return fmt.Errorf("sqliteffi: unsupported driver value %T", v)
		}
	}
	return nil
}

var _ driver.Driver = Driver{}
var _ driver.Conn = (*driverConn)(nil)
var _ driver.ConnBeginTx = (*driverConn)(nil)
var _ driver.ExecerContext = (*driverConn)(nil)
var _ driver.QueryerContext = (*driverConn)(nil)
var _ driver.Stmt = (*driverStmt)(nil)
var _ driver.Result = driverResult{}
var _ driver.Rows = (*driverRows)(nil)

package sqliteffi

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestPreparedStatementReuse(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "stmt.db"))
	if err != nil { t.Skipf("native SQLite library unavailable: %v", err) }
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE item (id INTEGER PRIMARY KEY, name TEXT)"); err != nil { t.Fatal(err) }
	stmt, err := db.Prepare("INSERT INTO item(name) VALUES (?)")
	if err != nil { t.Fatal(err) }
	defer stmt.Close()
	for _, name := range []string{"one", "two"} {
		if _, err := stmt.Exec(name); err != nil { t.Fatal(err) }
	}
	query, err := db.Prepare("SELECT name FROM item ORDER BY id")
	if err != nil { t.Fatal(err) }
	rows, err := query.Query()
	if err != nil { t.Fatal(err) }
	var got []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil { t.Fatal(err) }
		got = append(got, name)
	}
	if err := rows.Err(); err != nil { t.Fatal(err) }
	if err := rows.Close(); err != nil { t.Fatal(err) }
	if err := query.Close(); err != nil { t.Fatal(err) }
	if len(got) != 2 || got[0] != "one" || got[1] != "two" { t.Fatalf("unexpected rows: %#v", got) }
}

func TestTransactionCommitAndRollback(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil { t.Skipf("native SQLite library unavailable: %v", err) }
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE item (value INTEGER)"); err != nil { t.Fatal(err) }
	tx, err := db.Begin()
	if err != nil { t.Fatal(err) }
	if _, err := tx.Exec("INSERT INTO item VALUES (?)", 1); err != nil { t.Fatal(err) }
	if err := tx.Rollback(); err != nil { t.Fatal(err) }
	rows, err := db.Query("SELECT count(*) FROM item")
	if err != nil { t.Fatal(err) }
	defer rows.Close()
	if !rows.Next() { t.Fatalf("missing count row: %v", rows.Err()) }
	var count int64
	if err := rows.Scan(&count); err != nil { t.Fatal(err) }
	if count != 0 { t.Fatalf("count after rollback = %d, want 0", count) }
}

func TestDatabaseSQLDriver(t *testing.T) {
	const driverName = "sqliteffi_test"
	sql.Register(driverName, Driver{})
	db, err := sql.Open(driverName, ":memory:")
	if err != nil { t.Fatal(err) }
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE item (value TEXT)"); err != nil { t.Skipf("native SQLite unavailable: %v", err) }
	if _, err := db.Exec("INSERT INTO item(value) VALUES (?)", "driver"); err != nil { t.Fatal(err) }
	var value string
	if err := db.QueryRow("SELECT value FROM item").Scan(&value); err != nil { t.Fatal(err) }
	if value != "driver" { t.Fatalf("value = %q, want driver", value) }
	tx, err := db.Begin()
	if err != nil { t.Fatal(err) }
	if _, err := tx.Exec("INSERT INTO item(value) VALUES (?)", "rollback"); err != nil { t.Fatal(err) }
	if err := tx.Rollback(); err != nil { t.Fatal(err) }
	var count int
	if err := db.QueryRow("SELECT count(*) FROM item").Scan(&count); err != nil { t.Fatal(err) }
	if count != 1 { t.Fatalf("count = %d, want 1", count) }
}

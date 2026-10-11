package suite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

type OpenFunc func(string) (*sql.DB, error)

// Benchmarks runs the same workload definitions against a database/sql driver.
// Each top-level benchmark creates its own database to keep setup out of timing.
func Benchmarks(b *testing.B, open OpenFunc) {
	b.Helper()
	b.Run("OpenClose", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db, err := open(":memory:")
			if err != nil { b.Fatal(err) }
			if err := db.Close(); err != nil { b.Fatal(err) }
		}
	})
	b.Run("CRUD", func(b *testing.B) {
		db := openDB(b, open)
		defer db.Close()
		_, err := db.Exec("CREATE TABLE item (id INTEGER PRIMARY KEY, value TEXT NOT NULL)")
		if err != nil { b.Fatal(err) }
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			res, err := db.Exec("INSERT INTO item(value) VALUES (?)", fmt.Sprintf("v-%d", i))
			if err != nil { b.Fatal(err) }
			var value string
			if err := db.QueryRow("SELECT value FROM item WHERE id = ?", lastID(res)).Scan(&value); err != nil { b.Fatal(err) }
			if _, err := db.Exec("DELETE FROM item WHERE id = ?", lastID(res)); err != nil { b.Fatal(err) }
		}
	})
	b.Run("PreparedInsert", func(b *testing.B) {
		db := openDB(b, open)
		defer db.Close()
		if _, err := db.Exec("CREATE TABLE item (id INTEGER PRIMARY KEY, value TEXT NOT NULL)"); err != nil { b.Fatal(err) }
		stmt, err := db.Prepare("INSERT INTO item(value) VALUES (?)")
		if err != nil { b.Fatal(err) }
		defer stmt.Close()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := stmt.Exec(fmt.Sprintf("v-%d", i)); err != nil { b.Fatal(err) }
		}
	})
	b.Run("TransactionBulk100", func(b *testing.B) {
		db := openDB(b, open)
		defer db.Close()
		if _, err := db.Exec("CREATE TABLE item (id INTEGER PRIMARY KEY, value TEXT NOT NULL)"); err != nil { b.Fatal(err) }
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tx, err := db.Begin()
			if err != nil { b.Fatal(err) }
			for j := 0; j < 100; j++ {
				if _, err := tx.Exec("INSERT INTO item(value) VALUES (?)", fmt.Sprintf("%d-%d", i, j)); err != nil {
					_ = tx.Rollback()
					b.Fatal(err)
				}
			}
			if err := tx.Commit(); err != nil { b.Fatal(err) }
		}
	})
	b.Run("ReadScan1000", func(b *testing.B) {
		db := openDB(b, open)
		defer db.Close()
		if _, err := db.Exec("CREATE TABLE item (id INTEGER PRIMARY KEY, value TEXT NOT NULL)"); err != nil { b.Fatal(err) }
		tx, err := db.Begin()
		if err != nil { b.Fatal(err) }
		stmt, err := tx.Prepare("INSERT INTO item(value) VALUES (?)")
		if err != nil { b.Fatal(err) }
		for i := 0; i < 1000; i++ {
			if _, err := stmt.Exec(fmt.Sprintf("value-%d", i)); err != nil { b.Fatal(err) }
		}
		if err := stmt.Close(); err != nil { b.Fatal(err) }
		if err := tx.Commit(); err != nil { b.Fatal(err) }
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rows, err := db.Query("SELECT id, value FROM item ORDER BY id")
			if err != nil { b.Fatal(err) }
			var id int64
			var value string
			for rows.Next() {
				if err := rows.Scan(&id, &value); err != nil { b.Fatal(err) }
			}
			if err := rows.Err(); err != nil { b.Fatal(err) }
			if err := rows.Close(); err != nil { b.Fatal(err) }
		}
	})
	b.Run("ParallelReadWrite", func(b *testing.B) {
		db := openDB(b, open)
		defer db.Close()
		if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil { b.Fatal(err) }
		if _, err := db.Exec("CREATE TABLE item (id INTEGER PRIMARY KEY, value INTEGER NOT NULL)"); err != nil { b.Fatal(err) }
		db.SetMaxOpenConns(8)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			conn, err := db.Conn(b.Context())
			if err != nil { b.Error(err); return }
			defer conn.Close()
			if _, err := conn.ExecContext(b.Context(), "PRAGMA busy_timeout=5000"); err != nil { b.Error(err); return }
			for pb.Next() {
				if _, err := conn.ExecContext(b.Context(), "INSERT INTO item(value) VALUES (1)"); err != nil { b.Error(err); return }
				var count int64
				if err := conn.QueryRowContext(b.Context(), "SELECT count(*) FROM item").Scan(&count); err != nil { b.Error(err); return }
			}
		})
	})
}

// Stress exercises concurrent pooled connections against a file-backed database.
func Stress(t *testing.T, open OpenFunc) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "stress.db")
	db, err := open(path)
	if err != nil { t.Fatalf("open: %v", err) }
	defer db.Close()
	db.SetMaxOpenConns(8)
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil { t.Fatal(err) }
	if _, err := db.Exec("CREATE TABLE item (id INTEGER PRIMARY KEY, worker INTEGER NOT NULL, n INTEGER NOT NULL)"); err != nil { t.Fatal(err) }
	const workers, iterations = 8, 100
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			conn, err := db.Conn(t.Context())
			if err != nil { errs <- err; return }
			defer conn.Close()
			if _, err := conn.ExecContext(t.Context(), "PRAGMA busy_timeout=5000"); err != nil { errs <- err; return }
			for n := 0; n < iterations; n++ {
				if _, err := conn.ExecContext(t.Context(), "INSERT INTO item(worker,n) VALUES (?,?)", worker, n); err != nil { errs <- err; return }
				var got int
				if err := conn.QueryRowContext(t.Context(), "SELECT count(*) FROM item WHERE worker=?", worker).Scan(&got); err != nil { errs <- err; return }
			}
		}(worker)
	}
	wg.Wait()
	close(errs)
	for err := range errs { t.Error(err) }
	var count int
	if err := db.QueryRow("SELECT count(*) FROM item").Scan(&count); err != nil { t.Fatal(err) }
	if want := workers * iterations; count != want { t.Fatalf("rows = %d, want %d", count, want) }
}

func openDB(b *testing.B, open OpenFunc) *sql.DB {
	b.Helper()
	path := filepath.Join(b.TempDir(), "bench.db")
	db, err := open(path)
	if err != nil { b.Fatal(err) }
	db.SetMaxOpenConns(1)
	return db
}

func lastID(res sql.Result) int64 {
	id, _ := res.LastInsertId()
	return id
}

// Keep os referenced so platforms with different temp directory policies can
// report the path if setup failures are diagnosed in the future.
var _ = os.TempDir

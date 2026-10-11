package sqliteffi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// This stress test checks deterministic release of GoFFI-owned native handles.
// goleak's TestMain separately checks goroutine leaks. SQLite's native allocator
// is not necessarily introspectable in system builds, so use ASan/Valgrind with
// an instrumented SQLite build for definitive native-heap leak reports.
func TestResourceLifecycleStress(t *testing.T) {
	lib, err := LoadDefault()
	if err != nil {
		t.Skipf("native SQLite library unavailable: %v", err)
	}
	defer func() { require.NoError(t, lib.Close()) }()

	for iteration := 0; iteration < 40; iteration++ {
		db, err := lib.Open(":memory:")
		require.NoError(t, err, "iteration %d: open database", iteration)

		_, err = db.Exec("CREATE TABLE values_table (value INTEGER)")
		require.NoError(t, err, "iteration %d: create table", iteration)

		insert, err := db.Prepare("INSERT INTO values_table(value) VALUES (?)")
		require.NoError(t, err, "iteration %d: prepare insert", iteration)
		_, err = insert.Exec(iteration)
		require.NoError(t, err, "iteration %d: insert", iteration)
		require.NoError(t, insert.Close(), "iteration %d: finalize insert", iteration)

		query, err := db.Prepare("SELECT value FROM values_table")
		require.NoError(t, err, "iteration %d: prepare query", iteration)
		rows, err := query.Query()
		require.NoError(t, err, "iteration %d: query", iteration)
		require.True(t, rows.Next(), "iteration %d: expected one row", iteration)
		var value int64
		require.NoError(t, rows.Scan(&value), "iteration %d: scan", iteration)
		require.Equal(t, int64(iteration), value)
		require.False(t, rows.Next(), "iteration %d: unexpected second row", iteration)
		require.NoError(t, rows.Err(), "iteration %d: rows error", iteration)
		require.NoError(t, rows.Close(), "iteration %d: close rows", iteration)
		require.NoError(t, query.Close(), "iteration %d: finalize query", iteration)

		require.Zero(t, db.activeRows, "iteration %d: active rows must be released", iteration)
		require.NoError(t, db.Close(), "iteration %d: close database", iteration)
		require.Zero(t, lib.openDatabases, fmt.Sprintf("iteration %d: database handle count", iteration))
	}
}

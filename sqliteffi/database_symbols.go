package sqliteffi

import (
	"fmt"

	"github.com/ebitengine/purego"
)

// registerDatabaseAPI resolves and registers each SQLite entry point directly
// into the Library-owned function table. RegisterFunc installs Go closures in
// the supplied function variables; registering into the final table avoids
// registering closures against a temporary table and then copying it.
func registerDatabaseAPI(l *Library, lookup func(string) (uintptr, error)) error {
	symbols := []struct {
		name   string
		target any
	}{
		{"sqlite3_open_v2", &l.api.openV2},
		{"sqlite3_close_v2", &l.api.closeV2},
		{"sqlite3_errmsg", &l.api.errmsg},
		{"sqlite3_prepare_v2", &l.api.prepareV2},
		{"sqlite3_step", &l.api.step},
		{"sqlite3_finalize", &l.api.finalize},
		{"sqlite3_bind_null", &l.api.bindNull},
		{"sqlite3_bind_int64", &l.api.bindInt64},
		{"sqlite3_bind_double", &l.api.bindDouble},
		{"sqlite3_bind_text", &l.api.bindText},
		{"sqlite3_bind_blob", &l.api.bindBlob},
		{"sqlite3_column_count", &l.api.columnCount},
		{"sqlite3_column_name", &l.api.columnName},
		{"sqlite3_column_type", &l.api.columnType},
		{"sqlite3_column_int64", &l.api.columnInt64},
		{"sqlite3_column_double", &l.api.columnDouble},
		{"sqlite3_column_text", &l.api.columnText},
		{"sqlite3_column_blob", &l.api.columnBlob},
		{"sqlite3_column_bytes", &l.api.columnBytes},
		{"sqlite3_last_insert_rowid", &l.api.lastInsertRowID},
		{"sqlite3_changes", &l.api.changes},
		{"sqlite3_busy_timeout", &l.api.busyTimeout},
	}

	for _, symbol := range symbols {
		address, err := lookup(symbol.name)
		if err != nil {
			return fmt.Errorf("sqliteffi: resolve %s: %w", symbol.name, err)
		}
		if address == 0 {
			return fmt.Errorf("sqliteffi: resolve %s: symbol address is zero", symbol.name)
		}
		purego.RegisterFunc(symbol.target, address)
	}
	return nil
}

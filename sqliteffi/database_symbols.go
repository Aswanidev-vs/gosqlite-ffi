package sqliteffi

import (
 "fmt"

 "github.com/ebitengine/purego"
)

func registerDatabaseAPI(l *Library, lookup func(string) (uintptr,error)) error {
 var a databaseAPI
 symbols := []struct{name string; target any}{
  {"sqlite3_open_v2",&a.openV2},{"sqlite3_close_v2",&a.closeV2},{"sqlite3_errmsg",&a.errmsg},
  {"sqlite3_prepare_v2",&a.prepareV2},{"sqlite3_step",&a.step},{"sqlite3_finalize",&a.finalize},
  {"sqlite3_bind_null",&a.bindNull},{"sqlite3_bind_int64",&a.bindInt64},{"sqlite3_bind_double",&a.bindDouble},
  {"sqlite3_bind_text",&a.bindText},{"sqlite3_bind_blob",&a.bindBlob},
  {"sqlite3_column_count",&a.columnCount},{"sqlite3_column_name",&a.columnName},{"sqlite3_column_type",&a.columnType},
  {"sqlite3_column_int64",&a.columnInt64},{"sqlite3_column_double",&a.columnDouble},{"sqlite3_column_text",&a.columnText},
  {"sqlite3_column_blob",&a.columnBlob},{"sqlite3_column_bytes",&a.columnBytes},
  {"sqlite3_last_insert_rowid",&a.lastInsertRowID},{"sqlite3_changes",&a.changes},{"sqlite3_busy_timeout",&a.busyTimeout},
 }
 for _,s:=range symbols {
  address,err:=lookup(s.name); if err!=nil || address==0 { if err==nil { err=fmt.Errorf("symbol address is zero") }; return fmt.Errorf("sqliteffi: resolve %s: %w",s.name,err) }
  purego.RegisterFunc(s.target,address)
 }
 // Keep the typed function table alive as part of Library.
 l.api = a
 return nil
}

package sqliteffi

import (
 "fmt"
 "testing"
)

func benchmarkDB(b *testing.B) *DB {
 b.Helper()
 db,err:=Open(":memory:")
 if err!=nil { b.Skipf("native SQLite library unavailable: %v",err) }
 if _,err=db.Exec("CREATE TABLE items (id INTEGER PRIMARY KEY, value TEXT NOT NULL)"); err!=nil { _=db.Close(); b.Fatal(err) }
 return db
}

func BenchmarkExecInsert(b *testing.B) {
 db:=benchmarkDB(b); defer db.Close()
 b.ReportAllocs(); b.ResetTimer()
 for i:=0;i<b.N;i++ {
  if _,err:=db.Exec("INSERT INTO items(value) VALUES (?)",fmt.Sprintf("value-%d",i)); err!=nil { b.Fatal(err) }
 }
}

func BenchmarkQuerySingleRow(b *testing.B) {
 db:=benchmarkDB(b); defer db.Close()
 if _,err:=db.Exec("INSERT INTO items(value) VALUES (?)","benchmark"); err!=nil { b.Fatal(err) }
 b.ReportAllocs(); b.ResetTimer()
 for i:=0;i<b.N;i++ {
  rows,err:=db.Query("SELECT value FROM items WHERE id = ?",1); if err!=nil { b.Fatal(err) }
  if !rows.Next() { _=rows.Close(); b.Fatal("expected one row") }
  values,err:=rows.Values(); if err!=nil { _=rows.Close(); b.Fatal(err) }
  if len(values)!=1 { _=rows.Close(); b.Fatalf("got %d columns",len(values)) }
  if err:=rows.Close(); err!=nil { b.Fatal(err) }
 }
}

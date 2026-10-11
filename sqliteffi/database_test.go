package sqliteffi

import (
 "path/filepath"
 "testing"
)

func TestDatabaseRoundTrip(t *testing.T) {
 path:=filepath.Join(t.TempDir(),"example.db")
 db,err:=Open(path); if err!=nil { t.Skipf("native SQLite library unavailable: %v",err) }
 defer func(){ if err:=db.Close(); err!=nil { t.Error(err) } }()
 if _,err:=db.Exec("CREATE TABLE people (id INTEGER PRIMARY KEY, name TEXT NOT NULL, age INTEGER, score REAL, note TEXT, payload BLOB)"); err!=nil { t.Fatal(err) }
 res,err:=db.Exec("INSERT INTO people(name, age, score, note, payload) VALUES (?, ?, ?, ?, ?)", "Ada", 37, 9.5, nil, []byte{1,2,3})
 if err!=nil { t.Fatal(err) }
 if res.LastInsertID!=1 || res.RowsAffected!=1 { t.Fatalf("unexpected result: %+v",res) }
 rows,err:=db.Query("SELECT name, age, score, note, payload FROM people WHERE id = ?",res.LastInsertID); if err!=nil { t.Fatal(err) }
 if !rows.Next() { t.Fatalf("expected row, iteration error: %v",rows.Err()) }
 var name string; var age int64; var score float64; var note any; var payload []byte
 if err:=rows.Scan(&name,&age,&score,&note,&payload); err!=nil { t.Fatal(err) }
 if name!="Ada" || age!=37 || score!=9.5 || note!=nil || len(payload)!=3 || payload[2]!=3 { t.Fatalf("unexpected row: %q %d %f %#v %#v",name,age,score,note,payload) }
 if rows.Next() { t.Fatal("unexpected extra row") }
 if err:=rows.Err(); err!=nil { t.Fatal(err) }
 if err:=rows.Close(); err!=nil { t.Fatal(err) }
}

func TestDatabaseRejectsCloseWithActiveRows(t *testing.T) {
 db,err:=Open(":memory:"); if err!=nil { t.Skipf("native SQLite library unavailable: %v",err) }
 if _,err:=db.Exec("SELECT 1"); err!=nil { t.Fatal(err) }
 rows,err:=db.Query("SELECT 1"); if err!=nil { t.Fatal(err) }
 if err:=db.Close(); err!=ErrRowsOpen { t.Fatalf("Close error = %v, want ErrRowsOpen",err) }
 if err:=rows.Close(); err!=nil { t.Fatal(err) }
 if err:=db.Close(); err!=nil { t.Fatal(err) }
}

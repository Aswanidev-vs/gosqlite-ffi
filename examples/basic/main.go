package main

import (
 "fmt"
 "log"

 "github.com/Aswanidev-vs/gosqlite-ffi/sqliteffi"
)

func main() {
 db,err:=sqliteffi.Open("app.db")
 if err!=nil { log.Fatal(err) }
 defer db.Close()

 if _,err:=db.Exec(`CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  age INTEGER NOT NULL
 )`); err!=nil { log.Fatal(err) }

 result,err:=db.Exec("INSERT INTO users(name, age) VALUES (?, ?)", "Aswani", 18)
 if err!=nil { log.Fatal(err) }
 fmt.Printf("inserted user id=%d rows=%d\n",result.LastInsertID,result.RowsAffected)

 rows,err:=db.Query("SELECT id, name, age FROM users WHERE id = ?",result.LastInsertID)
 if err!=nil { log.Fatal(err) }
 defer rows.Close()

 for rows.Next() {
  var id int64
  var name string
  var age int64
  if err:=rows.Scan(&id,&name,&age); err!=nil { log.Fatal(err) }
  fmt.Printf("user: id=%d name=%s age=%d\n",id,name,age)
 }
 if err:=rows.Err(); err!=nil { log.Fatal(err) }
}

package sqliteffi

import (
 "errors"
 "fmt"
 "unsafe"
)

// Rows iterates a query result. Close it when iteration finishes.
type Rows struct {
 db *DB
 stmt unsafe.Pointer
 columns []string
 current bool
 done bool
 err error
 closed bool
 owner *Stmt
}

// Query prepares a SQL statement and returns rows. The caller must close Rows.
func (db *DB) Query(sql string, args ...any) (*Rows,error) {
 db.mu.Lock()
 defer db.mu.Unlock()
 if err:=db.checkOpen(); err!=nil { return nil,err }
 stmt,err:=db.prepare(sql,args); if err!=nil { return nil,err }
 rows:=&Rows{db:db,stmt:stmt}
 count:=db.lib.api.columnCount(stmt)
 rows.columns=make([]string,int(count))
 for i:=range rows.columns { rows.columns[i]=cString(db.lib.api.columnName(stmt,int32(i))) }
 db.activeRows++
 return rows,nil
}

// Columns returns a copy of the result column names.
func (r *Rows) Columns() []string { return append([]string(nil),r.columns...) }

// Next advances to the next row. It returns false at completion or on error.
func (r *Rows) Next() bool {
 if r==nil || r.closed || r.done { return false }
 r.db.mu.Lock()
 defer r.db.mu.Unlock()
 if err:=r.db.checkOpen(); err!=nil { r.err=err; r.done=true; return false }
 rc:=r.db.lib.api.step(r.stmt)
 switch rc {
 case sqliteRow: r.current=true; return true
 case sqliteDone: r.current=false; r.done=true; return false
 default: r.err=r.db.error(rc); r.done=true; return false
 }
}

// Values returns a copy of the current row. Integers are int64, floats float64,
// text strings, blobs []byte, and SQL NULL values nil.
func (r *Rows) Values() ([]any,error) {
 if r==nil || r.closed { return nil,ErrDBClosed }
 r.db.mu.Lock()
 defer r.db.mu.Unlock()
 if err:=r.db.checkOpen(); err!=nil { return nil,err }
 if !r.current { return nil,errors.New("sqliteffi: Next must return true before Values") }
 a:=r.db.lib.api
 out:=make([]any,len(r.columns))
 for i:=range out {
  col:=int32(i)
  switch a.columnType(r.stmt,col) {
  case sqliteInteger: out[i]=a.columnInt64(r.stmt,col)
  case sqliteFloat: out[i]=a.columnDouble(r.stmt,col)
  case sqliteText:
   n:=a.columnBytes(r.stmt,col); p:=a.columnText(r.stmt,col)
   if p==0 || n==0 { out[i]="" } else { out[i]=string(unsafe.Slice((*byte)(unsafe.Pointer(p)),int(n))) }
  case sqliteBlob:
   n:=a.columnBytes(r.stmt,col); p:=a.columnBlob(r.stmt,col)
   if n<=0 || p==0 { out[i]=[]byte{} } else { out[i]=append([]byte(nil),unsafe.Slice((*byte)(unsafe.Pointer(p)),int(n))...) }
  case sqliteNull: out[i]=nil
  default: out[i]=nil
  }
 }
 return out,nil
}

// Scan copies the current row into pointers to supported Go values.
func (r *Rows) Scan(dest ...any) error {
 values,err:=r.Values(); if err!=nil { return err }
 if len(dest)!=len(values) { return fmt.Errorf("sqliteffi: Scan has %d destinations for %d columns",len(dest),len(values)) }
 for i,v:=range values {
  switch d:=dest[i].(type) {
  case *any: *d=v
  case *int64:
   n,ok:=v.(int64); if !ok { return fmt.Errorf("sqliteffi: column %d has type %T, want int64",i,v) }; *d=n
  case *int:
   n,ok:=v.(int64); if !ok { return fmt.Errorf("sqliteffi: column %d has type %T, want int",i,v) }; *d=int(n)
  case *float64:
   switch n:=v.(type) { case float64:*d=n; case int64:*d=float64(n); default:return fmt.Errorf("sqliteffi: column %d has type %T, want number",i,v) }
  case *string:
   s,ok:=v.(string); if !ok { return fmt.Errorf("sqliteffi: column %d has type %T, want string",i,v) }; *d=s
  case *[]byte:
   if v==nil { *d=nil } else { b,ok:=v.([]byte); if !ok { return fmt.Errorf("sqliteffi: column %d has type %T, want []byte",i,v) }; *d=append((*d)[:0],b...) }
  case *bool:
   n,ok:=v.(int64); if !ok { return fmt.Errorf("sqliteffi: column %d has type %T, want integer bool",i,v) }; *d=n!=0
  default:return fmt.Errorf("sqliteffi: unsupported Scan destination %T",dest[i])
  }
 }
 return nil
}

// Err reports a query iteration error.
func (r *Rows) Err() error { if r==nil { return nil }; return r.err }

// Close finalizes the statement and releases its database connection.
func (r *Rows) Close() error {
 if r==nil || r.closed { return nil }
 r.db.mu.Lock()
 defer r.db.mu.Unlock()
 if r.closed { return nil }
 var rc int32
 if r.owner != nil {
  rc = r.db.lib.api.reset(r.stmt)
  _ = r.db.lib.api.clearBindings(r.stmt)
  r.owner.active = false
 } else {
  rc = r.db.lib.api.finalize(r.stmt)
 }
 r.stmt=nil; r.closed=true; r.current=false; r.db.activeRows--
 if rc!=sqliteOK && r.err==nil { r.err=r.db.error(rc) }
 return r.err
}

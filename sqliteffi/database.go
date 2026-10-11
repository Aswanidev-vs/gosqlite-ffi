package sqliteffi

import (
 "errors"
 "fmt"
 "math"
 "runtime"
 "sync"
 "unsafe"
)

const (
 sqliteOK int32 = 0
 sqliteRow int32 = 100
 sqliteDone int32 = 101
 sqliteOpenReadWrite int32 = 2
 sqliteOpenCreate int32 = 4
 sqliteInteger int32 = 1
 sqliteFloat int32 = 2
 sqliteText int32 = 3
 sqliteBlob int32 = 4
 sqliteNull int32 = 5
 sqliteTransient uintptr = ^uintptr(0)
)

var (
 ErrDatabaseOpen = errors.New("sqliteffi: close databases before closing their shared library")
 ErrDBClosed = errors.New("sqliteffi: database is closed")
 ErrRowsOpen = errors.New("sqliteffi: close active rows before closing the database")
)

type databaseAPI struct {
 openV2 func(*byte, *unsafe.Pointer, int32, *byte) int32
 closeV2 func(unsafe.Pointer) int32
 errmsg func(unsafe.Pointer) uintptr
 prepareV2 func(unsafe.Pointer, *byte, int32, *unsafe.Pointer, *uintptr) int32
 step func(unsafe.Pointer) int32
 finalize func(unsafe.Pointer) int32
 bindNull func(unsafe.Pointer, int32) int32
 bindInt64 func(unsafe.Pointer, int32, int64) int32
 bindDouble func(unsafe.Pointer, int32, float64) int32
 bindText func(unsafe.Pointer, int32, *byte, int32, uintptr) int32
 bindBlob func(unsafe.Pointer, int32, unsafe.Pointer, int32, uintptr) int32
 columnCount func(unsafe.Pointer) int32
 columnName func(unsafe.Pointer, int32) uintptr
 columnType func(unsafe.Pointer, int32) int32
 columnInt64 func(unsafe.Pointer, int32) int64
 columnDouble func(unsafe.Pointer, int32) float64
 columnText func(unsafe.Pointer, int32) uintptr
 columnBlob func(unsafe.Pointer, int32) uintptr
 columnBytes func(unsafe.Pointer, int32) int32
 lastInsertRowID func(unsafe.Pointer) int64
 changes func(unsafe.Pointer) int32
 busyTimeout func(unsafe.Pointer, int32) int32
}

// DB is a serialized SQLite connection. Open creates the database if absent.
type DB struct {
 mu sync.Mutex
 lib *Library
 handle unsafe.Pointer
 activeRows int
 ownsLibrary bool
 closed bool
}

// Open loads the default native SQLite library and opens or creates a database.
func Open(path string) (*DB, error) {
 lib, err := LoadDefault()
 if err != nil { return nil, err }
 db, err := lib.Open(path)
 if err != nil { _ = lib.Close(); return nil, err }
 db.ownsLibrary = true
 return db, nil
}

// Open opens or creates a database using an already-loaded SQLite library.
func (l *Library) Open(path string) (*DB, error) {
 if l == nil { return nil, ErrClosed }
 if path == "" { return nil, errors.New("sqliteffi: database path must not be empty") }
 l.mu.Lock()
 defer l.mu.Unlock()
 if l.closed || l.handle == 0 { return nil, ErrClosed }
 if l.api.openV2 == nil { return nil, errors.New("sqliteffi: database APIs were not initialized") }
 cpath := append([]byte(path), 0)
 var handle unsafe.Pointer
 rc := l.api.openV2(&cpath[0], &handle, sqliteOpenReadWrite|sqliteOpenCreate, nil)
 if rc != sqliteOK {
  msg := fmt.Sprintf("sqliteffi: open database: SQLite error %d", rc)
  if handle != nil { msg = l.dbError(handle, rc); _ = l.api.closeV2(handle) }
  return nil, errors.New(msg)
 }
 l.openDatabases++
 return &DB{lib:l, handle:handle}, nil
}

func (db *DB) checkOpen() error {
 if db == nil || db.closed || db.handle == nil { return ErrDBClosed }
 return nil
}
func (db *DB) error(rc int32) error { return errors.New(db.lib.dbError(db.handle, rc)) }
func (l *Library) dbError(handle unsafe.Pointer, rc int32) string {
 msg := "unknown SQLite error"
 if l.api.errmsg != nil { if p:=l.api.errmsg(handle); p!=0 { msg=cString(p) } }
 return fmt.Sprintf("sqliteffi: SQLite error %d: %s",rc,msg)
}

// SetBusyTimeout configures SQLite's busy timeout in milliseconds.
func (db *DB) SetBusyTimeout(milliseconds int) error {
 if milliseconds < 0 || milliseconds > math.MaxInt32 { return errors.New("sqliteffi: invalid busy timeout") }
 db.mu.Lock(); defer db.mu.Unlock()
 if err:=db.checkOpen(); err!=nil { return err }
 rc:=db.lib.api.busyTimeout(db.handle,int32(milliseconds))
 if rc!=sqliteOK { return db.error(rc) }; return nil
}

func (db *DB) prepare(sql string, args []any) (unsafe.Pointer,error) {
 if sql=="" { return nil, errors.New("sqliteffi: SQL must not be empty") }
 csql:=append([]byte(sql),0)
 var stmt unsafe.Pointer
 rc:=db.lib.api.prepareV2(db.handle,&csql[0],-1,&stmt,nil)
 if rc!=sqliteOK { return nil,db.error(rc) }
 if stmt==nil { return nil,errors.New("sqliteffi: SQL did not contain a statement") }
 for i,arg:=range args {
  if err:=db.bind(stmt,int32(i+1),arg); err!=nil { _=db.lib.api.finalize(stmt); return nil,fmt.Errorf("sqliteffi: bind parameter %d: %w",i+1,err) }
 }
 return stmt,nil
}
func (db *DB) bind(stmt unsafe.Pointer, index int32, value any) error {
 a:=db.lib.api
 var rc int32
 switch v:=value.(type) {
 case nil: rc=a.bindNull(stmt,index)
 case bool: if v { rc=a.bindInt64(stmt,index,1) } else { rc=a.bindInt64(stmt,index,0) }
 case int: rc=a.bindInt64(stmt,index,int64(v))
 case int8: rc=a.bindInt64(stmt,index,int64(v))
 case int16: rc=a.bindInt64(stmt,index,int64(v))
 case int32: rc=a.bindInt64(stmt,index,int64(v))
 case int64: rc=a.bindInt64(stmt,index,v)
 case uint: if uint64(v)>math.MaxInt64 { return errors.New("unsigned integer exceeds SQLite int64 range") }; rc=a.bindInt64(stmt,index,int64(v))
 case uint8: rc=a.bindInt64(stmt,index,int64(v))
 case uint16: rc=a.bindInt64(stmt,index,int64(v))
 case uint32: rc=a.bindInt64(stmt,index,int64(v))
 case uint64: if v>math.MaxInt64 { return errors.New("unsigned integer exceeds SQLite int64 range") }; rc=a.bindInt64(stmt,index,int64(v))
 case float32: rc=a.bindDouble(stmt,index,float64(v))
 case float64: rc=a.bindDouble(stmt,index,v)
 case string:
  b:=append([]byte(v),0); rc=a.bindText(stmt,index,&b[0],int32(len(b)-1),sqliteTransient); runtime.KeepAlive(b)
 case []byte:
  if v==nil { rc=a.bindNull(stmt,index) } else if len(v)==0 { rc=a.bindBlob(stmt,index,nil,0,sqliteTransient) } else { rc=a.bindBlob(stmt,index,unsafe.Pointer(&v[0]),int32(len(v)),sqliteTransient); runtime.KeepAlive(v) }
 default: return fmt.Errorf("unsupported parameter type %T",value)
 }
 if rc!=sqliteOK { return db.error(rc) }; return nil
}

// Result reports the last inserted row ID and affected row count.
type Result struct { LastInsertID int64; RowsAffected int64 }

// Exec runs one SQL statement with optional positional parameters.
func (db *DB) Exec(sql string, args ...any) (Result,error) {
 db.mu.Lock(); defer db.mu.Unlock()
 if err:=db.checkOpen(); err!=nil { return Result{},err }
 stmt,err:=db.prepare(sql,args); if err!=nil { return Result{},err }
 defer db.lib.api.finalize(stmt)
 for { rc:=db.lib.api.step(stmt); switch rc {
 case sqliteDone: return Result{db.lib.api.lastInsertRowID(db.handle),int64(db.lib.api.changes(db.handle))},nil
 case sqliteRow: continue
 default: return Result{},db.error(rc)
 } }
}

// Close closes the connection. Active query rows must be closed first.
func (db *DB) Close() error {
 if db==nil { return nil }
 db.mu.Lock(); defer db.mu.Unlock()
 if db.closed { return nil }
 if db.activeRows>0 { return ErrRowsOpen }
 rc:=db.lib.api.closeV2(db.handle); if rc!=sqliteOK { return db.error(rc) }
 db.handle=nil; db.closed=true
 db.lib.mu.Lock(); db.lib.openDatabases--; db.lib.mu.Unlock()
 if db.ownsLibrary { return db.lib.Close() }
 return nil
}

# Getting Started

## 1. Install a SQLite runtime

GoFFI SQLite calls an existing native SQLite shared library. It currently does not ship SQLite binaries.

- **Linux:** install your distribution's SQLite runtime package. Common names include `libsqlite3-0` on Debian/Ubuntu and `sqlite-libs` on Alpine.
- **macOS:** SQLite is commonly available as `/usr/lib/libsqlite3.dylib`; runtime behavior can vary by OS release.
- **Windows:** place a trusted `sqlite3.dll` in a location covered by the Windows loader rules or pass its absolute path to `Load`.

Prefer a controlled, known library path in production.

## 2. Add the package

```sh
go get github.com/Aswanidev-vs/gosqlite-ffi
```

## 3. Read SQLite version information

```go
package main

import (
    "fmt"
    "log"

    "github.com/Aswanidev-vs/gosqlite-ffi/sqliteffi"
)

func main() {
    lib, err := sqliteffi.LoadDefault()
    if err != nil {
        log.Fatal(err)
    }
    defer lib.Close()

    version, err := lib.Version()
    if err != nil {
        log.Fatal(err)
    }
    versionNumber, err := lib.VersionNumber()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("SQLite %s (numeric version %d)\n", version, versionNumber)
}
```

`VersionNumber` uses SQLite's documented numeric encoding (`major*1000000 + minor*1000 + patch`), allowing numeric comparisons without parsing the version string.

For a deterministic deployment, call `sqliteffi.Load` with a trusted absolute path.

## What is not implemented yet?

This milestone only loads the library and calls SQLite's version diagnostic functions. It does not open databases, run SQL, bind parameters, or support `database/sql`. Those features need to be implemented and tested before this library is suitable for application data.

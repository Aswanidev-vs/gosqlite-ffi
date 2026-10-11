# Architecture

```text
Go application
      |
      v
sqliteffi Go API
      |
      v
purego dynamic symbol invocation
      |
      v
SQLite shared library (sqlite3.dll / libsqlite3.so / libsqlite3.dylib)
```

## Goals

- Avoid compiling Go source through CGO.
- Reuse upstream SQLite rather than reimplementing SQL semantics.
- Keep the native boundary small and explicit.
- Make library lifetime and platform assumptions visible.

## Current implementation

`Load` or `LoadDefault` obtains a dynamic-library handle, resolves `sqlite3_libversion`, and registers a Go-callable function through `purego`. `Version` calls the function and copies the returned NUL-terminated C string into Go-owned memory. `Close` serializes against calls and releases the loaded library handle.

## Safety notes

- A native shared library is executable code. Only load a trusted path.
- Never close/unload a library while unregistered native callbacks or other native work might still use it. The current version wrapper serializes its own `Version` and `Close` operations, but future handles will need explicit lifetime ownership.
- SQLite calls and values will require exact ABI signatures, explicit ownership, and careful error conversion.
- Statement APIs should use bound parameters for values, not string concatenation.

## Planned modules

1. `sqliteffi`: low-level ABI bindings.
2. `sqlite`: ergonomic connection, statement, rows and transaction APIs.
3. `driver`: optional `database/sql/driver` integration.
4. Build/release automation: known SQLite versions and OS/architecture smoke tests.

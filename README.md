# GoFFI SQLite

**An experimental CGO-free SQLite database library for Go using `purego`.**

[Project website](https://aswanidev-vs.github.io/gosqlite-ffi/) · [Getting Started](docs/getting-started.md) · [Architecture](docs/architecture.md) · [Security Policy](SECURITY.md)

> **Status: experimental early implementation.** The library now opens/creates databases, executes single SQL statements with positional parameters, and queries typed rows. Transactions and `database/sql` integration are not implemented. Review the limitations and run your own tests before using it with important data.

## Why GoFFI SQLite?

- Explore a CGO-free Go build path with runtime calls through [`ebitengine/purego`](https://github.com/ebitengine/purego).
- Use upstream SQLite rather than maintaining a database engine.
- Keep the native API surface small, explicit, and testable.
- Include clear documentation, CI, vulnerability checks, Dependabot, and a static project website.

## Important trade-off

**CGO-free does not mean self-contained.** This project depends on a compatible SQLite shared library installed on the machine or shipped with your application. Library lookup behavior differs across operating systems. For deterministic deployments, control the SQLite build and load a trusted absolute path.

## Requirements

- Go 1.22 or newer
- A compatible native SQLite shared library
- An OS/architecture combination verified by CI and your own smoke tests

## Quick start

Install a compatible native SQLite shared library for your OS, then add the Go module:

```sh
go get github.com/Aswanidev-vs/gosqlite-ffi
```

Create `main.go`:

```go
package main

import (
    "fmt"
    "log"

    "github.com/Aswanidev-vs/gosqlite-ffi/sqliteffi"
)

func main() {
    db, err := sqliteffi.Open("app.db")
    if err != nil { log.Fatal(err) }
    defer db.Close()

    if _, err := db.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)"); err != nil { log.Fatal(err) }
    result, err := db.Exec("INSERT INTO users(name) VALUES (?)", "Ada")
    if err != nil { log.Fatal(err) }
    fmt.Println("inserted ID:", result.LastInsertID)

    rows, err := db.Query("SELECT id, name FROM users WHERE id = ?", result.LastInsertID)
    if err != nil { log.Fatal(err) }
    defer rows.Close()
    for rows.Next() {
        var id int64
        var name string
        if err := rows.Scan(&id, &name); err != nil { log.Fatal(err) }
        fmt.Println(id, name)
    }
    if err := rows.Err(); err != nil { log.Fatal(err) }
}
```

Run with `go run .`. The database file is created in the current working directory if it does not exist. SQL parameters support null, booleans, integer types, floating-point types, strings, and byte slices. Use placeholders for values rather than interpolating untrusted input into SQL.

See [examples/basic](examples/basic) for a complete runnable example.

## Benchmarks

Run the insert and single-row query benchmarks against the native SQLite library installed on your machine:

```sh
go test ./sqliteffi -run '^$' -bench 'Benchmark(ExecInsert|QuerySingleRow)$' -benchmem
```

Results depend on the SQLite build, OS, storage, Go version, and workload. These microbenchmarks currently measure GoFFI's API only; they are not a comparative performance claim against other drivers. Compare drivers using the same SQLite version, schema, data, transaction policy, and workload before drawing conclusions.

## Development

```sh
gofmt -w sqliteffi/*.go
go test ./...
go vet ./...
```

The initial implementation uses the `github.com/ebitengine/purego` dependency. Network access is needed when fetching modules for the first time. See [Development Guide](docs/development.md).

## Roadmap

- [x] Initial scaffold and docs
- [x] CGO-free shared library loading through purego
- [x] `sqlite3_libversion()` and `sqlite3_libversion_number()` diagnostics
- [x] Website, CI workflows, security policy and Dependabot
- [x] Database open/close handle APIs
- [x] Prepared statements and bound parameters
- [x] Row iteration and typed values
- [ ] Transactions, interruption and busy handling
- [ ] `database/sql/driver` integration
- [ ] Broader verified OS/architecture matrix and release artifacts

## Security

Please read [SECURITY.md](SECURITY.md). Never load a native library from an untrusted path. Use parameter binding for values; transactions and database/sql integration are not implemented yet.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Please include tests and document native ABI assumptions with every new binding.

## License

MIT. See [LICENSE](LICENSE).

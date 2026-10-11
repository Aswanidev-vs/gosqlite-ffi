# GoFFI SQLite

**An experimental CGO-free SQLite database library for Go using `purego`.**

[Project website](https://aswanidev-vs.github.io/gosqlite-ffi/) · [Getting Started](docs/getting-started.md) · [Architecture](docs/architecture.md) · [Security Policy](SECURITY.md)

> **Status: experimental early implementation.** The library supports database lifecycle management, parameterized SQL, reusable native prepared statements, typed row iteration, explicit transactions, and an experimental `database/sql` driver. Review the limitations and run your own tests before using it with important data.

## Why GoFFI SQLite?

- Explore a CGO-free Go build path with runtime calls through [`ebitengine/purego`](https://github.com/ebitengine/purego).
- Use upstream SQLite rather than maintaining a database engine.
- Keep the native API surface small, explicit, and testable.
- Include clear documentation, CI, vulnerability checks, Dependabot, and a static project website.

## Important trade-off

**CGO-free does not mean self-contained.** This project depends on a compatible SQLite shared library installed on the machine or shipped with your application. Library lookup behavior differs across operating systems. For deterministic deployments, control the SQLite build and load a trusted absolute path.

## Requirements

- Go 1.25 or newer
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

### Reusable prepared statements

```go
stmt, err := db.Prepare("INSERT INTO users(name) VALUES (?)")
if err != nil { log.Fatal(err) }
defer stmt.Close()
if _, err := stmt.Exec("Ada"); err != nil { log.Fatal(err) }
```

### Transactions

```go
tx, err := db.Begin()
if err != nil { log.Fatal(err) }
if _, err := tx.Exec("INSERT INTO users(name) VALUES (?)", "Grace"); err != nil {
    _ = tx.Rollback()
    log.Fatal(err)
}
if err := tx.Commit(); err != nil { log.Fatal(err) }
```

### `database/sql` driver

```go
sql.Register("sqliteffi", sqliteffi.Driver{})
sqldb, err := sql.Open("sqliteffi", "app.db")
if err != nil { log.Fatal(err) }
defer sqldb.Close()
```

Import `database/sql` to use this example. The DSN is the SQLite file path (or `:memory:`).

## Benchmarks

Run the insert and single-row query benchmarks against the native SQLite library installed on your machine:

```sh
go test ./sqliteffi -run '^$' -bench 'Benchmark(ExecInsert|QuerySingleRow)$' -benchmem
```

Results depend on the SQLite build, OS, storage, Go version, and workload. These microbenchmarks currently measure GoFFI's API only; they are not a comparative performance claim against other drivers. Compare drivers using the same SQLite version, schema, data, transaction policy, and workload before drawing conclusions.

## Comparative benchmarks and stress tests

A separate benchmark module compares GoFFI SQLite with:

- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) (CGO)
- [modernc.org/sqlite](https://modernc.org/sqlite) (pure Go, transpiled SQLite)
- [ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3) (pure Go, wasm2go)

The benchmark dependencies are isolated under `benchmarks/go.mod`, so they do not become production library dependencies. All implementations use the same `database/sql`-based workload harness.

Workloads include connection open/close, a CRUD round trip, reused prepared inserts, 100-row transaction batches, scans of 1,000 rows, and parallel reads/writes against a file-backed WAL database. Each driver also runs a concurrent stress test with eight workers and 100 insert/read cycles per worker, then verifies that all 800 rows were committed.

Run locally on Linux with a compatible native SQLite shared library and CGO enabled for mattn:

```sh
cd benchmarks
go mod tidy
go test ./... -run '^TestConcurrentStress
```sh
gofmt -w sqliteffi/*.go
go test ./...
go test -race ./...
go test ./sqliteffi -run TestResourceLifecycleStress -count=10
go vet -unsafeptr=false ./...
```

The initial implementation uses the `github.com/ebitengine/purego` dependency. Network access is needed when fetching modules for the first time. See [Development Guide](docs/development.md).

## Roadmap

- [x] Initial scaffold, documentation, and CI
- [x] CGO-free native library loading and version diagnostics
- [x] Database open/close, parameterized execution, and typed row iteration
- [x] Reusable prepared statements
- [x] Explicit transactions
- [x] Experimental `database/sql` driver
- [ ] Context cancellation, deadlines, and SQLite interruption
- [ ] Concurrency/transaction ownership guarantees and race-tested semantics
- [ ] Resource and native-memory leak verification
- [ ] Extended SQLite errors and value-binding edge cases
- [ ] Backup/restore and operational configuration APIs
- [ ] Fuzzing, robust failure-path tests, and performance baselines
- [ ] Verified OS/architecture matrix and versioned release artifacts

## Future implementation plan

The following items are planned; they are not promises of current support. Prioritize correctness and lifecycle safety before adding more surface area.

1. **Context cancellation and interruption** — context-aware driver calls, query deadlines, safe `sqlite3_interrupt` behavior, and tests for cancellation while a query is running.
2. **Concurrency and transaction safety** — define connection ownership rules, serialize transaction use, and validate behavior with the Go race detector.
3. **Resource and memory-leak verification** — package-wide goroutine leak detection with `goleak`, repeated open/prepare/query/close stress tests, native handle accounting, and ASan/Valgrind runs against an instrumented SQLite build. `goleak` detects goroutine leaks, not native heap leaks.
4. **Error and value correctness** — extended SQLite result codes, consistent error types, and more tests for NULL, empty text/blob, integer boundaries, and invalid bindings.
5. **SQLite backup and operational features** — online backup/restore, explicit journal/WAL configuration, and clearer busy-timeout behavior.
6. **Fuzzing and resilience** — fuzz parameter conversion, row scanning, malformed SQL handling, and repeated failure/cleanup paths.
7. **Benchmarks and compatibility** — reproducible workloads against other drivers, memory/allocation profiles, and documented trade-offs.
8. **Platform matrix and releases** — verify supported OS/architecture combinations and native SQLite distributions, publish versioned releases, and maintain compatibility notes.

### Test and leak checks

```sh
go test ./...
go test -race ./...
go test ./sqliteffi -run TestResourceLifecycleStress -count=10
go test ./sqliteffi -run '^$' -bench . -benchmem
```

The automated suite uses `testify` assertions and `goleak` to detect goroutines left behind by tests. Stress tests check that GoFFI-owned database/row handles are released. These checks do **not** prove that a native SQLite build has no heap leaks; definitive native-allocation checks require an instrumented SQLite library and tools such as AddressSanitizer or Valgrind on a compatible platform.
## Security

Please read [SECURITY.md](SECURITY.md). Never load a native library from an untrusted path. Use parameter binding for values. This driver is experimental; validate the target OS/library combination and serialize manual transaction use on a shared `DB` connection.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Please include tests and document native ABI assumptions with every new binding.

## License

MIT. See [LICENSE](LICENSE).
 -count=3 -timeout=10m
go test ./... -run '^
```sh
gofmt -w sqliteffi/*.go
go test ./...
go test -race ./...
go test ./sqliteffi -run TestResourceLifecycleStress -count=10
go vet -unsafeptr=false ./...
```

The initial implementation uses the `github.com/ebitengine/purego` dependency. Network access is needed when fetching modules for the first time. See [Development Guide](docs/development.md).

## Roadmap

- [x] Initial scaffold, documentation, and CI
- [x] CGO-free native library loading and version diagnostics
- [x] Database open/close, parameterized execution, and typed row iteration
- [x] Reusable prepared statements
- [x] Explicit transactions
- [x] Experimental `database/sql` driver
- [ ] Context cancellation, deadlines, and SQLite interruption
- [ ] Concurrency/transaction ownership guarantees and race-tested semantics
- [ ] Resource and native-memory leak verification
- [ ] Extended SQLite errors and value-binding edge cases
- [ ] Backup/restore and operational configuration APIs
- [ ] Fuzzing, robust failure-path tests, and performance baselines
- [ ] Verified OS/architecture matrix and versioned release artifacts

## Future implementation plan

The following items are planned; they are not promises of current support. Prioritize correctness and lifecycle safety before adding more surface area.

1. **Context cancellation and interruption** — context-aware driver calls, query deadlines, safe `sqlite3_interrupt` behavior, and tests for cancellation while a query is running.
2. **Concurrency and transaction safety** — define connection ownership rules, serialize transaction use, and validate behavior with the Go race detector.
3. **Resource and memory-leak verification** — package-wide goroutine leak detection with `goleak`, repeated open/prepare/query/close stress tests, native handle accounting, and ASan/Valgrind runs against an instrumented SQLite build. `goleak` detects goroutine leaks, not native heap leaks.
4. **Error and value correctness** — extended SQLite result codes, consistent error types, and more tests for NULL, empty text/blob, integer boundaries, and invalid bindings.
5. **SQLite backup and operational features** — online backup/restore, explicit journal/WAL configuration, and clearer busy-timeout behavior.
6. **Fuzzing and resilience** — fuzz parameter conversion, row scanning, malformed SQL handling, and repeated failure/cleanup paths.
7. **Benchmarks and compatibility** — reproducible workloads against other drivers, memory/allocation profiles, and documented trade-offs.
8. **Platform matrix and releases** — verify supported OS/architecture combinations and native SQLite distributions, publish versioned releases, and maintain compatibility notes.

### Test and leak checks

```sh
go test ./...
go test -race ./...
go test ./sqliteffi -run TestResourceLifecycleStress -count=10
go test ./sqliteffi -run '^$' -bench . -benchmem
```

The automated suite uses `testify` assertions and `goleak` to detect goroutines left behind by tests. Stress tests check that GoFFI-owned database/row handles are released. These checks do **not** prove that a native SQLite build has no heap leaks; definitive native-allocation checks require an instrumented SQLite library and tools such as AddressSanitizer or Valgrind on a compatible platform.
## Security

Please read [SECURITY.md](SECURITY.md). Never load a native library from an untrusted path. Use parameter binding for values. This driver is experimental; validate the target OS/library combination and serialize manual transaction use on a shared `DB` connection.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Please include tests and document native ABI assumptions with every new binding.

## License

MIT. See [LICENSE](LICENSE).
 -bench '^BenchmarkDriver/' -benchmem -benchtime=2s -count=5 -timeout=30m
```

The [comparative benchmark workflow](.github/workflows/benchmarks.yml) records Go version, OS/architecture, CGO setting, CPU information, runs stress tests, and uploads raw benchmark output. Compare results only when the machine, Go version, driver versions, SQLite version/build, journal mode, schema, and workload are accounted for. These measurements are workload-specific and are not a general performance ranking.

**Results status:** the benchmark harness has been added, but measured results are not yet published here. Results will be appended after the workflow completes successfully and its output has been reviewed. No performance claim is made before that validation.

## Development

```sh
gofmt -w sqliteffi/*.go
go test ./...
go test -race ./...
go test ./sqliteffi -run TestResourceLifecycleStress -count=10
go vet -unsafeptr=false ./...
```

The initial implementation uses the `github.com/ebitengine/purego` dependency. Network access is needed when fetching modules for the first time. See [Development Guide](docs/development.md).

## Roadmap

- [x] Initial scaffold, documentation, and CI
- [x] CGO-free native library loading and version diagnostics
- [x] Database open/close, parameterized execution, and typed row iteration
- [x] Reusable prepared statements
- [x] Explicit transactions
- [x] Experimental `database/sql` driver
- [ ] Context cancellation, deadlines, and SQLite interruption
- [ ] Concurrency/transaction ownership guarantees and race-tested semantics
- [ ] Resource and native-memory leak verification
- [ ] Extended SQLite errors and value-binding edge cases
- [ ] Backup/restore and operational configuration APIs
- [ ] Fuzzing, robust failure-path tests, and performance baselines
- [ ] Verified OS/architecture matrix and versioned release artifacts

## Future implementation plan

The following items are planned; they are not promises of current support. Prioritize correctness and lifecycle safety before adding more surface area.

1. **Context cancellation and interruption** — context-aware driver calls, query deadlines, safe `sqlite3_interrupt` behavior, and tests for cancellation while a query is running.
2. **Concurrency and transaction safety** — define connection ownership rules, serialize transaction use, and validate behavior with the Go race detector.
3. **Resource and memory-leak verification** — package-wide goroutine leak detection with `goleak`, repeated open/prepare/query/close stress tests, native handle accounting, and ASan/Valgrind runs against an instrumented SQLite build. `goleak` detects goroutine leaks, not native heap leaks.
4. **Error and value correctness** — extended SQLite result codes, consistent error types, and more tests for NULL, empty text/blob, integer boundaries, and invalid bindings.
5. **SQLite backup and operational features** — online backup/restore, explicit journal/WAL configuration, and clearer busy-timeout behavior.
6. **Fuzzing and resilience** — fuzz parameter conversion, row scanning, malformed SQL handling, and repeated failure/cleanup paths.
7. **Benchmarks and compatibility** — reproducible workloads against other drivers, memory/allocation profiles, and documented trade-offs.
8. **Platform matrix and releases** — verify supported OS/architecture combinations and native SQLite distributions, publish versioned releases, and maintain compatibility notes.

### Test and leak checks

```sh
go test ./...
go test -race ./...
go test ./sqliteffi -run TestResourceLifecycleStress -count=10
go test ./sqliteffi -run '^$' -bench . -benchmem
```

The automated suite uses `testify` assertions and `goleak` to detect goroutines left behind by tests. Stress tests check that GoFFI-owned database/row handles are released. These checks do **not** prove that a native SQLite build has no heap leaks; definitive native-allocation checks require an instrumented SQLite library and tools such as AddressSanitizer or Valgrind on a compatible platform.
## Security

Please read [SECURITY.md](SECURITY.md). Never load a native library from an untrusted path. Use parameter binding for values. This driver is experimental; validate the target OS/library combination and serialize manual transaction use on a shared `DB` connection.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Please include tests and document native ABI assumptions with every new binding.

## License

MIT. See [LICENSE](LICENSE).

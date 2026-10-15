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

### Typed SQLite errors

SQLite operation failures expose a `*sqliteffi.SQLiteError` with the primary result code, the raw result code returned by SQLite, and SQLite's diagnostic message. Use `errors.As` to inspect details and `errors.Is` to match common primary codes:

```go
var sqliteErr *sqliteffi.SQLiteError
if errors.As(err, &sqliteErr) {
    fmt.Printf("SQLite code %d (raw %d): %s\\n", sqliteErr.Code, sqliteErr.ResultCode, sqliteErr.Message)
}
if errors.Is(err, sqliteffi.CodeConstraint) {
    // Handle a constraint violation.
}
```

Import the standard-library `errors` package for this pattern. The raw result code preserves the code returned by the failing API call; this does not yet enable or query SQLite's extended-error-code mode.

### `database/sql` driver

```go
sql.Register("sqliteffi", sqliteffi.Driver{})
sqldb, err := sql.Open("sqliteffi", "app.db")
if err != nil { log.Fatal(err) }
defer sqldb.Close()
```

Import `database/sql` to use this example. The DSN is the SQLite file path (or `:memory:`).

The driver also implements `driver.Pinger` and `driver.Validator`. `sql.DB.PingContext` verifies the connection by executing `SELECT 1`; a context canceled before the call is rejected. Because native SQLite calls are synchronous in this version, cancellation does not interrupt a query already executing. The validator lets `database/sql` avoid reusing a closed connection.

## Benchmarks

Run the insert and single-row query benchmarks against the native SQLite library installed on your machine:

```sh
go test ./sqliteffi -run '^$' -bench 'Benchmark(ExecInsert|QuerySingleRow)$' -benchmem
```

Results depend on the SQLite build, OS, storage, Go version, and workload. These microbenchmarks currently measure GoFFI's API only; they are not a comparative performance claim against other drivers. Compare drivers using the same SQLite version, schema, data, transaction policy, and workload before drawing conclusions.

## Comparative benchmarks and stress tests

The comparison harness targets four Go SQLite options:

| Driver | Implementation | CGO |
|---|---|---|
| GoFFI SQLite | Native SQLite shared library through purego | No |
| [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) | Native SQLite | Yes |
| [modernc.org/sqlite](https://modernc.org/sqlite) | Transpiled SQLite | No |
| [ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3) | SQLite compiled to WebAssembly and translated to Go | No |

### Measured results

**Matched-runtime results measured on 2026-10-15** by GitHub Actions on Linux/amd64 using an Intel Xeon Platinum 8370C CPU @ 2.80 GHz. The workflow verified SQLite **3.53.4 for all four drivers**, ran the three-run concurrent stress tests successfully, and completed five benchmark repetitions per workload. Values below are medians and include time, bytes allocated, and allocation count. [Workflow and raw artifact](https://github.com/Aswanidev-vs/gosqlite-ffi/actions/runs/38676334220).

| Workload | GoFFI | mattn | modernc | ncruces |
|---|---:|---:|---:|---:|
| Open / close | 805.0 ns/op; 528 B/op; 7 allocs/op | 835.6 ns/op; 528 B/op; 7 allocs/op | 844.8 ns/op; 528 B/op; 7 allocs/op | 864.0 ns/op; 592 B/op; 7 allocs/op |
| CRUD round trip | 1,193,900 ns/op; 2,168 B/op; 86 allocs/op | 1,059,443 ns/op; 1,111 B/op; 35 allocs/op | 1,176,639 ns/op; 847 B/op; 31 allocs/op | 1,009,451 ns/op; 4,424 B/op; 86 allocs/op |
| Reused prepared inserts | 578,206 ns/op; 487 B/op; 25 allocs/op | 515,092 ns/op; 159 B/op; 7 allocs/op | 568,503 ns/op; 159 B/op; 8 allocs/op | 485,467 ns/op; 1,848 B/op; 32 allocs/op |
| 100-row transaction batch | 1,410,152 ns/op; 61,676 B/op; 2,618 allocs/op | 899,476 ns/op; 27,614 B/op; 1,103 allocs/op | 1,067,771 ns/op; 18,712 B/op; 995 allocs/op | 1,165,152 ns/op; 30,824 B/op; 1,019 allocs/op |
| Scan 1,000 rows | 3,005,918 ns/op; 358,112 B/op; 19,776 allocs/op | 462,749 ns/op; 37,768 B/op; 2,764 allocs/op | 556,451 ns/op; 37,680 B/op; 2,762 allocs/op | 466,713 ns/op; 16,008 B/op; 1,019 allocs/op |
| Parallel WAL read/write | 166,361 ns/op; 1,381 B/op; 50 allocs/op | 49,405 ns/op; 1,257 B/op; 30 allocs/op | 146,302 ns/op; 1,137 B/op; 28 allocs/op | 67,584 ns/op; 800 B/op; 17 allocs/op |

**Interpretation:** These are workload-specific microbenchmarks, not a universal ranking. GoFFI has competitive open/close results in this run, while other drivers are faster and/or allocate less in several query, insert, transaction, scan, and parallel workloads. In particular, the GoFFI 1,000-row scan now reports 19,776 allocations and 358,112 B/op, down from the historical 22,521 allocations and about 412 KB/op; the old run used a different CPU and mismatched SQLite versions, so treat that as directional context rather than a controlled before/after comparison. Review the raw artifact before drawing conclusions.

### Run the comparison

```sh
cd benchmarks
go mod tidy
go test ./... -run '^TestConcurrentStress$' -count=3 -timeout=10m
go test -v ./... -run '^$' -bench '^BenchmarkDriver/' -benchmem -benchtime=2s -count=5 -timeout=30m
```

The suite covers connection open/close, CRUD, prepared inserts, transaction batches, 1,000-row scans, parallel read/write operations, and an eight-worker stress test verifying 800 inserts. Record Go version, OS/architecture, CGO setting, CPU, driver versions, and SQLite runtime versions with results.

See [the benchmark-specific README](benchmarks/README.md) and [the benchmark workflow](.github/workflows/benchmarks.yml). Results depend on the runner, SQLite build, and workload, so do not generalize one run into a universal performance ranking.

## Development

```sh
gofmt -w sqliteffi/*.go
go test ./...
go test -race ./...
go test ./sqliteffi -run TestResourceLifecycleStress -count=10
go vet -unsafeptr=false ./...
```

The native call layer uses [`github.com/ebitengine/purego`](https://github.com/ebitengine/purego). Tests also use `testify` assertions and `goleak` for goroutine-leak detection. Network access is needed when fetching Go modules for the first time. See [Development Guide](docs/development.md).

## Roadmap

- [x] Initial scaffold, documentation, and CI
- [x] CGO-free native library loading and version diagnostics
- [x] Database open/close, parameterized execution, and typed row iteration
- [x] Reusable prepared statements
- [x] Explicit transactions
- [x] Experimental `database/sql` driver
- [ ] Context cancellation, deadlines, and SQLite interruption
- [ ] Concurrency/transaction ownership guarantees and race-tested semantics
- [x] Test assertion helpers and goroutine-leak checks (`testify`, `goleak`)
- [ ] Native-memory leak verification with an instrumented SQLite build
- [ ] Extended SQLite errors and value-binding edge cases
- [ ] Backup/restore and operational configuration APIs
- [ ] Fuzzing and expanded failure-path tests
- [ ] Refreshed matched-version comparative performance baseline
- [ ] Verified OS/architecture matrix and versioned release artifacts

## Future implementation plan

The following items are planned; they are not promises of current support. Prioritize correctness and lifecycle safety before adding more surface area.

1. **Context cancellation and interruption** — context-aware driver calls, query deadlines, safe `sqlite3_interrupt` behavior, and tests for cancellation while a query is running.
2. **Concurrency and transaction safety** — define connection ownership rules, serialize transaction use, and validate behavior with the Go race detector.
3. **Resource and memory-leak verification** — `testify` assertions and `goleak` checks are already included in the test dependencies and suite. Continue repeated open/prepare/query/close stress tests and native handle accounting, then add ASan/Valgrind runs against an instrumented SQLite build. `goleak` detects goroutine leaks, not native heap leaks.
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

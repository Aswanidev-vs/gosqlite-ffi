# Comparative SQLite benchmarks

This folder compares GoFFI SQLite with mattn/go-sqlite3, modernc.org/sqlite, and ncruces/go-sqlite3 through a shared `database/sql` workload harness.

## Status: measurements pending

**There are no verified benchmark numbers yet.** Publish measured values only after the CI benchmark job completes, the stress tests pass, and raw output has been reviewed. Until then, result cells remain pending.

| Workload | GoFFI | mattn | modernc | ncruces |
|---|---:|---:|---:|---:|
| Open / close | Pending | Pending | Pending | Pending |
| CRUD round trip | Pending | Pending | Pending | Pending |
| Reused prepared inserts | Pending | Pending | Pending | Pending |
| 100-row transaction batch | Pending | Pending | Pending | Pending |
| Scan 1,000 rows | Pending | Pending | Pending | Pending |
| Parallel WAL read/write | Pending | Pending | Pending | Pending |
| Allocations / bytes per operation | Pending | Pending | Pending | Pending |
| Concurrent stress | Pending | Pending | Pending | Pending |

## Run locally

```sh
go mod tidy
go test ./... -run '^TestConcurrentStress$' -count=3 -timeout=10m
go test -v ./... -run '^$' -bench '^BenchmarkDriver/' -benchmem -benchtime=2s -count=5 -timeout=30m
```

The workflow runs on GitHub Actions Linux and uploads raw output as an artifact. It records Go version, OS/architecture, CGO mode, CPU details, and SQLite runtime version. GoFFI requires a compatible native SQLite shared library; mattn/go-sqlite3 requires CGO and a C toolchain.

The stress workload uses eight workers with 100 insert/read cycles each and verifies 800 rows. This is a basic concurrency check, not proof of absence of all races or native-memory leaks.

## Interpretation

- Report five-run output, including `ns/op`, `B/op`, and `allocs/op`.
- Record driver versions and SQLite runtime/build versions.
- Do not generalize from one workload or operating system.
- Link raw CI artifacts alongside published tables.

Main documentation: [README](../README.md).
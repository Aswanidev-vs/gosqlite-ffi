# Comparative SQLite benchmarks

This folder compares GoFFI SQLite with mattn/go-sqlite3, modernc.org/sqlite, and ncruces/go-sqlite3 using a shared database/sql workload harness.

## Result status

**No verified benchmark numbers are available yet.** This page explicitly tracks pending work; it does not claim any driver is faster. Publish measurements only after CI succeeds, stress tests pass, and raw benchmark output is reviewed.

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

Workloads include open/close, CRUD, prepared inserts, transaction batches, scans, parallel reads/writes, and an eight-worker stress test verifying 800 inserts. Record Go version, OS/architecture, CGO mode, CPU, driver versions, and SQLite runtime versions.

GoFFI requires a compatible native SQLite shared library. mattn/go-sqlite3 requires CGO and a C toolchain. Report `ns/op`, `B/op`, and `allocs/op` from repeated runs, link raw CI artifacts, and avoid comparing measurements from incompatible environments.

Main documentation: [README](../README.md). Workflow: [comparative benchmarks](../.github/workflows/benchmarks.yml).
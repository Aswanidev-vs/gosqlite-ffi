# Comparative SQLite benchmarks

This folder compares GoFFI SQLite with mattn/go-sqlite3, modernc.org/sqlite, and ncruces/go-sqlite3 using a shared database/sql workload harness.

## Measured results

**Measured on 2026-10-11 in GitHub Actions** (Linux/amd64, AMD EPYC 7763, Go stable; five repetitions per benchmark, median reported). Each cell reports median `ns/op`, `B/op`, and `allocs/op`. The concurrent stress-test step passed all three runs. [Raw results artifact and workflow run](https://github.com/Aswanidev-vs/gosqlite-ffi/actions/runs/38136727123).

| Workload | GoFFI | mattn | modernc | ncruces |
|---|---:|---:|---:|---:|
| Open / close | 689.7 ns/op; 528 B/op; 7 allocs/op | 630.3 ns/op; 528 B/op; 7 allocs/op | 621.0 ns/op; 528 B/op; 7 allocs/op | 753.3 ns/op; 592 B/op; 7 allocs/op |
| CRUD round trip | 1,319,820 ns/op; 2,200 B/op; 88 allocs/op | 1,361,630 ns/op; 1,111 B/op; 35 allocs/op | 1,571,157 ns/op; 847 B/op; 31 allocs/op | 1,337,785 ns/op; 4,424 B/op; 86 allocs/op |
| Reused prepared inserts | 621,937 ns/op; 487 B/op; 25 allocs/op | 673,814 ns/op; 159 B/op; 7 allocs/op | 752,524 ns/op; 159 B/op; 8 allocs/op | 638,038 ns/op; 1,847 B/op; 32 allocs/op |
| 100-row transaction batch | 1,326,782 ns/op; 61,682 B/op; 2,619 allocs/op | 1,088,848 ns/op; 27,591 B/op; 1,101 allocs/op | 1,295,299 ns/op; 18,690 B/op; 993 allocs/op | 1,353,173 ns/op; 30,810 B/op; 1,017 allocs/op |
| Scan 1,000 rows | 3,016,697 ns/op; 412,089 B/op; 22,521 allocs/op | 490,171 ns/op; 37,768 B/op; 2,764 allocs/op | 442,472 ns/op; 37,680 B/op; 2,762 allocs/op | 475,526 ns/op; 16,008 B/op; 1,019 allocs/op |
| Parallel WAL read/write | 186,242 ns/op; 1,395 B/op; 52 allocs/op | 88,551 ns/op; 1,270 B/op; 30 allocs/op | 213,668 ns/op; 1,137 B/op; 28 allocs/op | 77,671 ns/op; 788 B/op; 17 allocs/op |
| Concurrent stress test (3 runs) | Passed | Passed | Passed | Passed |

**Comparability caveat:** GoFFI reported SQLite 3.45.1, while mattn, modernc, and ncruces reported SQLite 3.53.4. These are real measurements from the same runner and harness, but not a fully controlled SQLite-version comparison. Treat them as an initial baseline, not a definitive ranking; rerun with matched SQLite runtime versions before making strict driver-level conclusions.

## Run locally

```sh
go mod tidy
go test ./... -run '^TestConcurrentStress$' -count=3 -timeout=10m
go test -v ./... -run '^$' -bench '^BenchmarkDriver/' -benchmem -benchtime=2s -count=5 -timeout=30m
```

Workloads include open/close, CRUD, prepared inserts, transaction batches, scans, parallel reads/writes, and an eight-worker stress test verifying 800 inserts. Record Go version, OS/architecture, CGO mode, CPU, driver versions, and SQLite runtime versions.

GoFFI requires a compatible native SQLite shared library. mattn/go-sqlite3 requires CGO and a C toolchain. Report `ns/op`, `B/op`, and `allocs/op` from repeated runs, link raw CI artifacts, and avoid comparing measurements from incompatible environments.

Main documentation: [README](../README.md). Workflow: [comparative benchmarks](../.github/workflows/benchmarks.yml).
# Comparative SQLite benchmarks

This folder compares GoFFI SQLite with mattn/go-sqlite3, modernc.org/sqlite, and ncruces/go-sqlite3 using a shared database/sql workload harness.

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

## Run locally

```sh
go mod tidy
go test ./... -run '^TestConcurrentStress$' -count=3 -timeout=10m
go test -v ./... -run '^$' -bench '^BenchmarkDriver/' -benchmem -benchtime=2s -count=5 -timeout=30m
```

Workloads include open/close, CRUD, prepared inserts, transaction batches, scans, parallel reads/writes, and an eight-worker stress test verifying 800 inserts. Record Go version, OS/architecture, CGO mode, CPU, driver versions, and SQLite runtime versions.

GoFFI requires a compatible native SQLite shared library. mattn/go-sqlite3 requires CGO and a C toolchain. Report `ns/op`, `B/op`, and `allocs/op` from repeated runs, link raw CI artifacts, and avoid comparing measurements from incompatible environments.

Main documentation: [README](../README.md). Workflow: [comparative benchmarks](../.github/workflows/benchmarks.yml).
# Performance Roadmap

## P1: Reduce `database/sql` row-scan allocations

Evidence: the matched-runtime benchmark run on 2026-10-12 reported GoFFI scanning 1,000 rows at a median of 3.254 ms/op, 412,080 B/op, and 22,521 allocs/op on Linux/amd64 with Go 1.27.2 and SQLite 3.53.4. This is a workload-specific result, not a general driver ranking.

Current implementation path: `driverRows.Next` calls `Rows.Values`, creating an intermediate `[]any`, and then copies BLOB values a second time. `Rows.Values` already copies SQLite-owned BLOB bytes into Go-owned memory.

Acceptance criteria:
- Remove the redundant BLOB copy without reusing memory that callers may retain.
- Avoid the intermediate row slice if it can be done without duplicating or weakening column conversion semantics.
- Preserve NULL, empty BLOB, text, integer, and floating-point behavior.
- Add regression tests covering driver scans for scalar values, NULL, empty BLOB, and binary BLOB data.
- Run gofmt, `go test ./...`, `go test -race ./...`, and `go vet -unsafeptr=false ./...`.
- Re-run the comparative benchmark with the same Go, SQLite, OS, architecture, and workload settings; report medians, bytes/op, allocations/op, and raw workflow artifacts.

## P2: Context cancellation

Add context-aware execution and query interruption only after cancellation semantics and connection ownership are specified and tested. Verify that cancellation cannot interrupt an unrelated statement on a shared connection.

## P2: Platform support matrix

Document only OS/architecture combinations actually tested by CI, including native SQLite library discovery and ABI requirements. Add architecture-specific runners when available.

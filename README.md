# GoFFI SQLite

**An experimental CGO-free Go binding prototype for SQLite's C API using `purego`.**

[Project website](https://aswanidev-vs.github.io/gosqlite-ffi/) · [Getting Started](docs/getting-started.md) · [Architecture](docs/architecture.md) · [Security Policy](SECURITY.md)

> **Status: experimental scaffold.** The current milestone loads a native SQLite shared library and exposes `sqlite3_libversion()` and `sqlite3_libversion_number()` diagnostics. Database connections, SQL execution, prepared statements, transactions, and `database/sql` integration are not implemented. Do not use this prototype for production data.

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

`VersionNumber` follows SQLite's numeric convention: `major*1000000 + minor*1000 + patch`. It can be useful for deterministic version comparisons without parsing a string.

For deterministic deployments, use `sqliteffi.Load` with a trusted absolute library path. See [Getting Started](docs/getting-started.md) for platform notes.

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
- [ ] Database open/close handle APIs
- [ ] Prepared statements and bound parameters
- [ ] Row iteration and typed values
- [ ] Transactions, interruption and busy handling
- [ ] `database/sql/driver` integration
- [ ] Broader verified OS/architecture matrix and release artifacts

## Security

Please read [SECURITY.md](SECURITY.md). Never load a native library from an untrusted path. SQL execution and parameter binding are not implemented yet.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Please include tests and document native ABI assumptions with every new binding.

## License

MIT. See [LICENSE](LICENSE).

# Development Guide

## Local workflow

```sh
gofmt -w sqliteffi/*.go
go test ./...
go vet ./...
```

The project depends on `github.com/ebitengine/purego`. Run dependency commands with network access and commit `go.sum` so CI uses verified module checksums.

## Contribution checklist

1. Open an issue for substantial changes or platform support work.
2. Add tests for normal behavior, error cases and resource lifetimes.
3. Run tests and `go vet` on your platform.
4. Document any OS/architecture constraints and library lookup assumptions.
5. Open a pull request with a focused change and its test evidence.

## Native ABI review

Before adding a function binding, verify its exact C signature in the SQLite API documentation, pointer ownership and lifetime, integer widths, nullability, and the behavior of the selected `purego` release on every supported platform. Never guess a C function signature.

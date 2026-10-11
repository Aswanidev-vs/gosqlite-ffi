# Security Policy

## Project status

GoFFI SQLite is experimental. The first version only binds a SQLite version diagnostic and does not implement a database connection or SQL execution API. Do not rely on it for production data.

## Supported versions

Until the first tagged release, only the default branch is maintained. There is no production support guarantee.

## Reporting a vulnerability

Please do not publish exploitable vulnerabilities as public issues. Use GitHub's private vulnerability reporting for this repository if it is enabled. Otherwise, contact the maintainer privately through the email listed on the GitHub profile. Include affected commit or version, reproduction steps, impact, and a proposed mitigation when possible.

Please allow maintainers reasonable time to investigate before public disclosure.

## Security practices

- Keep `purego` and Go toolchains updated.
- Avoid loading shared libraries from writable or attacker-controlled directories.
- Ship a known, trusted SQLite build when deterministic versioning is important.
- Do not concatenate untrusted values into SQL. Parameter binding is planned but not yet implemented.
- Run `go test ./...`, `go vet ./...`, and review dependency alerts before releases.

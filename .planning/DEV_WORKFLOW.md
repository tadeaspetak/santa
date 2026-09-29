# Dev workflow

Requires Go 1.23 (CI pins `1.23`). Install tooling with `make tools` (staticcheck + `go mod tidy`).

| Command | Does |
| --- | --- |
| `make run PARAM="send -d"` | `go run` with version ldflag; pass subcommand via `PARAM` |
| `make test` | `go test ./...` (tests use temp dirs; also deletes stray root `santa-batch-*`) |
| `make vet` / `make fmt` / `make lint` | chained: lint → fmt → vet (`staticcheck ./...`) |
| `make build` | lint, then cross-compile `bin/santa-mac` (darwin/amd64) and `bin/santa-win.exe` |
| `make clean` | remove `bin/` and batch files |

- Git hooks: `git config core.hooksPath hooks` → `hooks/pre-commit` runs `make lint && make test`.
- CI (`.github/workflows/go.yml`, on push/PR to `main`): `make fmt`, `make test`, staticcheck.
- Tests: `cmd` (validation), `internal/app` (pairing, sending, PDF), `internal/data` (load/save),
  root `schema_test.go`. `mailer/*` is untested (thin wrappers over external services).
- **staticcheck is pinned to `2024.1.1`** in both CI and `make tools`: `@latest` (v0.8+) needs
  Go ≥ 1.26. Bump it together with the Go version.

## Versioning & release

- Version string lives in `Makefile` (`VERSION := v0.2.0`) and is injected into
  `cmd/version.Version` via `-ldflags -X`. `go run .` without the Makefile prints an empty version.
- Tags `v0.0.2` … `v0.2.0` exist. Releases are **manual**: bump `VERSION`, `make build`, tag,
  upload `bin/*` to a GitHub Release (no release workflow). No linux or darwin/arm64 build target.

## Safe manual testing

```sh
cp data.example.json /tmp/santa-test.json   # add an smtp or mailgun block (dummy values are fine for -d)
go run . send -d -p --data /tmp/santa-test.json
```
Run from a scratch dir or clean up the `santa-batch-*` files afterwards. `send` still shows the
interactive confirmation prompt even with `-d`.

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
- Tests (all offline, ~5s, `-race` clean):
  - `cmd/e2e_test.go`: whole commands (`init`, `check`, `preview`, `send` incl. prompt answers,
    `--to`, `--dry-run`, wrong credentials, failures halfway, Mailgun) against fake servers.
  - `internal/app`: pairing properties/distribution/errors, `Send`, PDF, `Render`.
  - `internal/app/mailer`: both mailers against the fake servers.
  - `internal/data`, `cmd` helpers, root `schema_test.go`.
- When changing behaviour, break it on purpose once to check a test catches it.
- **staticcheck is pinned to `2024.1.1`** in both CI and `make tools`: `@latest` (v0.8+) needs
  Go ≥ 1.26. Bump it together with the Go version.

## Versioning & release

- Version string lives in `Makefile` (`VERSION := v1.0.0`) and is injected into
  `cmd/version.Version` via `-ldflags -X`. `go run .` without the Makefile prints an empty version.
- Tags `v0.0.2` … `v0.2.0` exist (`v1.0.0` bumped in the Makefile, not tagged yet). Releases are **manual**: bump `VERSION`, `make build`, tag,
  upload `bin/*` to a GitHub Release (no release workflow). No linux or darwin/arm64 build target.

## Safe manual testing

Run from a scratch dir (the output folders land in the CWD):

```sh
go build -o /tmp/santa-test/santa . && cd /tmp/santa-test
./santa init && ./santa preview && ./santa send -d -p    # no mailer needed for any of these
```

Automated e2e tests (`cmd/e2e_test.go`) cover the flows below; manual runs are for eyeballing.

**Manually with real SMTP, safely:** point `smtp` at a local sink, e.g.
[Mailpit](https://mailpit.axllent.org) (`smtp` port 1025, UI on 8025) or a tiny Python socket
server. `user`/`pass` must be non-empty to pass validation but are ignored when the server
doesn't advertise `AUTH`. Then `check`, `preview -t`, `send -t`, `send -y` all work against it.
Pipe answers into the prompt: `echo y | ./santa send -t me@x.com` (Enter / anything but y = No).
Ports < 1024 (and possibly 587) can't be bound inside the Claude Code sandbox; use e.g. 2525.

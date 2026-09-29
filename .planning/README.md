# .planning — agent context for `santa`

Start here. These files are a map of the project for future agents/contributors, written 2026-09-29
(updated the same day after the bug-fix/test pass, the CLI workflow rework and the end-to-end
test suite; `Makefile` version is `v1.0.0`, last pushed release tag is `v0.2.0`). Verify against the code if something looks stale.

| File | Read it when you need… |
| --- | --- |
| [OVERVIEW.md](OVERVIEW.md) | what the app does, the user flow, tech stack, history |
| [ARCHITECTURE.md](ARCHITECTURE.md) | package layout, call graph, where each concern lives |
| [DATA_MODEL.md](DATA_MODEL.md) | the `data.json` config, Go structs, validation, JSON schema |
| [PAIRING.md](PAIRING.md) | how the randomised pairing (+ extra recipients) works |
| [DEV_WORKFLOW.md](DEV_WORKFLOW.md) | build, test, lint, CI, hooks, versioning/release |
| [CONCERNS.md](CONCERNS.md) | known bugs, gaps, gotchas, and ideas for future work |

## 30-second summary

- Go 1.23 CLI (cobra) that draws Secret Santa pairings from a JSON file and emails each giver
  their recipient(s) via **SMTP** (gomail) or **Mailgun** (EU API).
- Commands: `init` (sample `data.json`), `check` (data + pairing feasible + mailer login, sends
  nothing), `preview [-t me@x]` (one sample email with a made-up recipient, printed or sent),
  `send`, `version`. Global flag `--data` (default `data.json`).
- `send` flags: `--dry-run/-d` (no emails, no mailer config needed), `--to/-t <email>` (all emails
  to one address), `--pdf/-p` (printable PDF), `--yes/-y` (skip the confirmation, which defaults
  to **No**). Old names (`--debug`, `--alwaysSendTo`/`-a`, `--printPdf`) were removed on purpose.
- Every `send` writes a `santa-batch-<datetime>[-N]/` folder with `<giverEmail>.txt` per giver
  (+ `emails.pdf`) as an audit log.
- Core logic is in `internal/app` (`pair.go`, `extras.go`, `send.go`); tests live next to it.
  Every package has tests; `cmd/e2e_test.go` runs whole commands against fake SMTP/Mailgun servers
  from `internal/testutil`, so real sending is covered without any network.

## Hard rules for agents

- **Never read, print, or commit** the untracked `data*.json` files in the repo root (other than
  `data.example.json`). They hold real family emails and SMTP/Mailgun credentials.
  `.gitignore` ignores `*.json` on purpose (with exceptions for `data.example.json`, `schema.json`).
- **Never run `send` without `-d`** unless the user explicitly asks — it emails real people.
  `check` and `preview` without `-t` are always safe (`check` only logs in to the mailer).
  For end-to-end behaviour, prefer adding a case to `cmd/e2e_test.go` over manual runs.
- Adding a new JSON file to the repo? It will be ignored by git unless you add a `!` exception.

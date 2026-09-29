# .planning — agent context for `santa`

Start here. These files are a map of the project for future agents/contributors, written 2026-09-29
(updated after the bug-fix/test pass the same day, on top of `e5cc5a7` / tag `v0.2.0`). Verify against the code if something looks stale.

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
- Commands: `init` (writes a sample `data.json`), `send` (validate → confirm → pair → email),
  `version`. Global flag `--data` (default `data.json`).
- `send` flags: `-d` debug (no emails), `-a <email>` send everything to one address,
  `-p` also produce a printable PDF.
- Every run writes `santa-batch-<datetime>-<giverEmail>.txt` files to the CWD as an audit log.
- Core logic is in `internal/app` (`pair.go`, `extras.go`, `send.go`); tests live next to it.
  Every package with logic has tests (`cmd`, `internal/app`, `internal/data`, root `schema_test.go`).

## Hard rules for agents

- **Never read, print, or commit** the untracked `data*.json` files in the repo root (other than
  `data.example.json`). They hold real family emails and SMTP/Mailgun credentials.
  `.gitignore` ignores `*.json` on purpose (with exceptions for `data.example.json`, `schema.json`).
- **Never run `send` without `-d`** unless the user explicitly asks — it emails real people.
- Adding a new JSON file to the repo? It will be ignored by git unless you add a `!` exception.

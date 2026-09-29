# Known issues, gotchas & ideas

Found while mapping the code on 2026-09-29. Verify before acting.

## Fixed (2026-09-29)

- Extras "new round" didn't reset (`:=` shadowing in `extras.go`) → extras piled up on random
  givers. Regression: `TestPairExtrasBalanced`.
- `pairExtras` sorted the caller's slice in place. Regression: `TestPairDoesNotMutateInput`.
- Pairing used 10 random restarts and could fail on solvable inputs → now backtracking.
  Regression: `TestPairInvariants` ("multiple predestined").
- Self-predestined / duplicate predestined recipients produced invalid pairings → now errors.
- Duplicate participant emails weren't validated → `unique=Email` validator tag.
- `schema.json`: root `required` was inside `properties`; `minLength` → `minItems`; added
  `anyOf` smtp/mailgun. Regression: `schema_test.go`.
- `log.Fatalf` inside `internal/app` → returned errors (PDF, pairing, extras).
- Unused `var Validate` removed from `cmd/send.go`.
- Unknown emails in `excludedRecipients` / `excludedGivers` and a predestined recipient excluded
  by the same giver were silently accepted → `validatePreferences` reports them.
- CLI workflow rework: `check` / `preview` commands; `send` validates, pairs and verifies the
  mailer login before its prompt; prompt defaults to No and is skipped on dry runs; dry runs need
  no mailer config; per-run output folder; `--dry-run/--to/--pdf/--yes` (old flag names removed —
  **breaking** for anyone with scripts, mention it in the next release notes); fixed nameless
  `Usage:` line and log timestamps in errors; noisy Mailgun response output removed.
- End-to-end tests: `cmd` refactored to `NewRootCmd`/`RunE`; fake SMTP & Mailgun servers in
  `internal/testutil`; `promptui` dropped for a plain `[y/N]` prompt; mailers now tested.
- Version bumped to `v1.0.0` in the Makefile.
- SMTP port configurable (`smtp.port`, default 587).
- README "(TODO)" in Setup resolved.
- staticcheck pinned to `2024.1.1` in CI and `make tools` (`latest` needs Go ≥ 1.26).

## Open

- `Send` aborts on the first mail error leaving a partially-sent batch; the batch `.txt` files are
  the only recovery path (documented in README).
- Mailgun is EU-only (`mailer.MailgunEUAPIBase` passed by `cmd`) — US-region accounts won't work;
  `check` reports it as a login failure up front. Accepted by the owner for now (2026-09-29).
- PDF strips all Unicode symbols (emojis) because fpdf core fonts can't render them.
- Batch `.txt` files contain the pairing in plain text in the CWD — privacy relies on the
  organiser not opening them.
- `data.example.json` has no mailer block: fine for dry runs/preview, `check` flags it.
- `v1.0.0` not tagged/released yet (manual: build, tag, upload binaries). Release notes should
  call out the renamed `send` flags.
- Fake servers only speak plain SMTP + AUTH PLAIN; STARTTLS/implicit TLS (465) paths of gomail
  aren't exercised by tests.

## Ideas

- Deterministic seed flag (`--seed`) for reproducible debugging/tests.
- Upfront validation pass with friendly messages for all reference errors.
- Configurable SMTP port / Mailgun region; linux + darwin/arm64 builds; release GitHub Action.
- More template variables (e.g. `%{giverSalutation}`).

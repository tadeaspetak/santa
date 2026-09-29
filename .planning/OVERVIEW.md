# Overview

## Purpose

`santa` makes running a family Secret Santa easy: define participants and an email template in a
JSON file, let the app draw random giver→recipient pairings honouring preferences, and email each
giver privately. Nobody (including the organiser) needs to see the result.

Repo: https://github.com/tadeaspetak/santa — binaries distributed via GitHub Releases
(`santa-mac`, `santa-win.exe`). Module path: `github.com/tadeaspetak/santa`.

## Features

- Preferences per participant: `excludedRecipients`, `predestinedRecipient`.
- `extraRecipients`: people who **receive** but don't give (e.g. toddlers), with `excludedGivers`.
  A giver can therefore get multiple recipients; salutations are joined with
  `template.recipientsSeparator` (default `" and "`).
- Email template with a single placeholder: `%{recipientSalutation}` (in `subject` and `body`).
  `body` is HTML and gets wrapped in `<html><body>…</body></html>`.
- Mailers: SMTP (`port` optional, default 587) or Mailgun (EU region hardcoded). Mailgun wins if
  both set. Both implement `Verify` (log in without sending), used by `check` and before `send`.
- Safety/rehearsal: `check`, `preview` (sample recipient "Rudolph", + "Blitzen" if there are
  extras, so no pairing is revealed), `send -d` dry run, `send -t` redirect all mail to one address
  (real giver in `Reply-To`), per-run output folder with per-giver `.txt` copies, optional
  printable PDF (`-p`) with blank first/last pages so pairings can be folded/hidden. See
  `data.example.pdf`.

## User flow

1. Download the binary, run `./santa-mac init` → creates `data.json` pre-filled with sample data
   and a `$schema` pointing to `schema.json` on GitHub `main` (for editor autocompletion).
2. Edit `data.json` (mailer creds, template, participants, extras).
3. `check` → `preview` / `preview -t me@x` → `send -d [-p]` → `send -t me@x` (all rehearsals).
4. `send` for real: validates, pairs and verifies the mailer login **before** asking
   "Send N emails to the participants? This can't be undone [y/N]" (Enter cancels).

The user-facing documentation is the root `README.md` — keep it in sync with behaviour changes.

## Tech stack

| Concern | Library |
| --- | --- |
| CLI | `github.com/spf13/cobra` |
| Validation | `github.com/go-playground/validator/v10` (struct tags) |
| SMTP | `gopkg.in/gomail.v2` |
| Mailgun | `github.com/mailgun/mailgun-go/v4` |
| PDF | `github.com/go-pdf/fpdf` + `github.com/k3a/html2text` |
| Lint | `staticcheck` |

## History highlights

- Nov 2024: initial build (Mailgun only), PDF printing, extra recipients, CI.
- Dec 2024 (`1b72f81`, v0.2.0): **removed the interactive CLI** for editing participants/template
  (`cmd/participants/*`, `internal/prompt`, `internal/validation`) in favour of hand-editing JSON
  guided by `schema.json`; added SMTP support.
- Jul 2025: README/text fixes. Project is small and low-churn; mostly used seasonally.
- Sep 2026: bug-fix + test pass (backtracking pairing, preference validation); CLI workflow
  rework (`check`, `preview`, renamed `send` flags, output folders, default-No confirmation);
  end-to-end tests with fake mail servers; `promptui` replaced by a plain `[y/N]` line read;
  version bumped to v1.0.0.

# Architecture

```
main.go                     → cmd.Execute()
cmd/
  root.go                   NewRootCmd() builds a fresh tree (+ workflow in Long help), --data
                            flag, loadData(), newMailer(), `mailgunAPIBase` var (tests override);
                            Execute() prints "Error: …" to stderr and exits 1
  validate.go               validate(dat, requireMailer) + "translations" map; validateEmail()
  init.go                   `init`: refuses if file exists; writes sample data; prints next step
  check.go                  `check`: validate → app.Pair (discarded) → mailer.Verify;
                            returns errChecksFailed on any ✗
  preview.go                `preview`: app.Render with sampleRecipients(); print or send (-t)
  send.go                   `send`: flags → validate → Pair → Verify → confirm() → app.Send;
                            confirm() reads a line ([y/N]); newOutputDir() (unique folder)
  e2e_test.go               end-to-end: run(t, stdin, args...) on a temp CWD with example data
                            + fake servers; asserts emails as received form a valid pairing
  validate_test.go          table tests for validate() incl. dry-run mode and the example data
  send_test.go              newOutputDir(), confirm()
  version.go                `version`: prints version.Version
  version/version.go        `Version` var, injected via -ldflags in Makefile
internal/
  data/data.go              Data/Template/Mailgun/Smtp/Person/Participant/Extra structs;
                            LoadData / SaveData (SaveData un-escapes \uXXXX so emojis stay readable)
  data/data_test.go         save/load round trip, load errors, example file
  app/
    pair.go                 Pair() + pairParticipants() — randomised backtracking pairing
    preferences.go          validatePreferences() — cross-reference checks of the preferences
    extras.go               pairExtras() — assigns extra recipients to givers
    send.go                 Send(), Render(), HTMLToText(), Mailer interface, SendOpts, PDF
    utils.go                getRandomIndexInArray, shuffle (math/rand), getParticipantsByEmail
    mailer/smtp.go          smtpMailer (gomail)      — implements app.Mailer
    mailer/mailgun.go       mailgunMailer            — implements app.Mailer; API base injected
                            (cmd passes MailgunEUAPIBase)
    mailer/mailer_test.go   Verify/Send for both mailers against the fake servers
    pair_test.go            invariant (property) tests, deterministic, distribution, errors,
                            extras balance, no input mutation
    send_test.go            Send()/Render() with a mock Mailer; testOpts(t) → temp OutputDir
  testutil/                 test-only helpers (never imported by non-test code):
    smtp.go                 NewSMTPServer(t, user, pass): real SMTP on 127.0.0.1:<random>,
                            AUTH PLAIN, records Messages(), FailFor(addr) rejects a recipient
    mailgun.go              NewMailgunServer(t, domain, key): httptest fake of GET domain +
                            POST messages, records Messages(); use .APIBase
    testutil.go             Message type, Chdir(t, dir) (Go 1.23 has no t.Chdir)
schema_test.go              structural regression test for schema.json (package main)
```

## `send` data flow

```
loadData ──► validate(dat, requireMailer: !dryRun) ──► app.Pair → []giverWithRecipients ──►
(not dry run) newMailer + Verify ──► confirm (skipped for -d / -y; default No) ──►
app.Send(mailer, pairs, template, SendOpts{To, IsDryRun, ShouldPrintPdf, OutputDir, Out})
   mkdir OutputDir; for each pair: Render → write <giver>.txt → (PDF page) → (skip if dry run)
   → mailer.Send → progress line to Out
   finally: <OutputDir>/emails.pdf if -p
```
Everything that can fail (data, pairing, credentials) fails **before** the prompt. Nothing is
written if the user cancels.

## Key types / seams

- `app.Mailer` interface: `Send(sender, subject, body, recipient, replyTo string) error` and
  `Verify(ctx) error` (log in without sending: SMTP `Dial`+`Close`, Mailgun `GetDomain`).
  Add a new transport by implementing it in `internal/app/mailer` and wiring it in
  `cmd/root.go` `newMailer` (plus struct in `data.go`, validation tags, `schema.json`, README).
- `app.giverWithRecipients{giver data.Participant; recipients []data.Person}` — unexported;
  `Pair` returns it and `Send` consumes it (tests in the same package construct it directly).
- `data.Person` is embedded in both `Participant` and `Extra`, so JSON `salutation` sits flat on
  each object.

## Conventions

- Nothing calls `log.Fatalf`/`os.Exit` except `cmd.Execute`. `internal/app` returns errors
  (sentinels like `errNoValidPairing`, wrapped with `%w`, so tests use `errors.Is`); commands use
  `RunE` and write via `cmd.OutOrStdout()` / read via `cmd.InOrStdin()` so tests can drive them.
- New commands: add a `newXCmd()` constructor, register it in `NewRootCmd`, add e2e cases.
- `app.Send` writes into `SendOpts.OutputDir` and prints to `SendOpts.Out`; tests use
  `testOpts(t, …)` (temp dir + `io.Discard`). The CLI picks the dir via `newOutputDir` in the CWD.
- User-facing CLI output: `✓`/`✗` status lines, and every message ends by suggesting the next
  command (`init` → `check` → `preview` → `send --dry-run`). Keep that chain intact.
- Formatting: gofmt (tabs in Go); `.editorconfig` says 2 spaces for other files, 100 col limit.
- Comments are sparse, explanatory, occasionally chatty; commit messages use `feat:`/`fix:`/`chore:`.

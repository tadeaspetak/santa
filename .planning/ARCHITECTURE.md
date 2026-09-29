# Architecture

```
main.go                     → cmd.Execute()
cmd/
  root.go                   RootCmd, --data persistent flag, getDataPath(), registers subcommands
  init.go                   `init`: refuses if file exists; writes sample data via data.SaveData
  send.go                   `send`: load → validate → pick mailer → confirm → app.Pair → app.Send
                            also: validator error "translations" map, ask() promptui helper
  send_test.go              table tests for validate() incl. the example data
  version.go                `version`: prints version.Version
  version/version.go        `Version` var, injected via -ldflags in Makefile
internal/
  data/data.go              Data/Template/Mailgun/Smtp/Person/Participant/Extra structs;
                            LoadData / SaveData (SaveData un-escapes \uXXXX so emojis stay readable)
  data/data_test.go         save/load round trip, load errors, example file
  app/
    pair.go                 Pair() + pairParticipants() — randomised backtracking pairing
    extras.go               pairExtras() — assigns extra recipients to givers
    send.go                 Send(), Mailer interface, SendOpts, PDF generation, batch files
    utils.go                getRandomIndexInArray, shuffle (math/rand), getParticipantsByEmail
    mailer/smtp.go          smtpMailer (gomail)      — implements app.Mailer
    mailer/mailgun.go       mailgunMailer (EU API)   — implements app.Mailer
    pair_test.go            invariant (property) tests, deterministic, distribution, errors,
                            extras balance, no input mutation
    send_test.go            Send() with a mock Mailer; runs in a temp CWD (inTempDir helper)
schema_test.go              structural regression test for schema.json (package main)
```

## `send` data flow

```
data.LoadData(path) ──► validate(dat) ──► choose mailer ──► ask("…ready") ──►
app.Pair(participants, extras) → []giverWithRecipients ──►
app.Send(mailer, pairs, template, opts)
   for each pair: build subject/body → write batch .txt → (PDF page) → (skip if debug) → mailer.Send
   finally: write santa-batch-<datetime>.pdf if -p
```

## Key types / seams

- `app.Mailer` interface: `Send(sender, subject, body, recipient, replyTo string) error`.
  Add a new transport by implementing it in `internal/app/mailer` and wiring it in `cmd/send.go`
  (plus struct in `data.go`, validation tags, `schema.json`, README).
- `app.giverWithRecipients{giver data.Participant; recipients []data.Person}` — unexported;
  `Pair` returns it and `Send` consumes it (tests in the same package construct it directly).
- `data.Person` is embedded in both `Participant` and `Extra`, so JSON `salutation` sits flat on
  each object.

## Conventions

- `internal/app` returns errors (sentinels like `errNoValidPairing`, wrapped with `%w`, so tests
  use `errors.Is`); only the `cmd` layer calls `log.Fatalf`.
- Output files are written to the **current working directory**; tests that call `Send` must use
  the `inTempDir(t)` helper so they don't litter the repo.
- Formatting: gofmt (tabs in Go); `.editorconfig` says 2 spaces for other files, 100 col limit.
- Comments are sparse, explanatory, occasionally chatty; commit messages use `feat:`/`fix:`/`chore:`.

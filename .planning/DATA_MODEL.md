# Data model (`data.json`)

Canonical example: `data.example.json`. Go source of truth: `internal/data/data.go`.
Editor help: `schema.json` (referenced via `$schema`, fetched from GitHub `main`, so schema
changes only reach users after merging to `main`).

```jsonc
{
  "$schema": "https://raw.githubusercontent.com/tadeaspetak/santa/refs/heads/main/schema.json",
  "smtp":    { "host": "...", "port": 587, "user": "...", "pass": "..." }, // either smtp (port optional)…
  "mailgun": { "domain": "mg.example.com", "apiKey": "..." },    // …or mailgun (mailgun wins if both)
  "template": {
    "subject": "Find a gift for %{recipientSalutation}",        // required
    "body": "<p>… %{recipientSalutation} …</p>",                 // required, HTML
    "sender": "santa@mg.example.com",                            // required, must be an email
    "recipientsSeparator": " and "                               // optional, default " and "
  },
  "participants": [                                              // required, min 2
    { "email": "a@x.com", "salutation": "A",
      "excludedRecipients": ["b@x.com"],                         // optional
      "predestinedRecipient": "c@x.com" }                        // optional
  ],
  "extraRecipients": [                                           // optional
    { "salutation": "Baby", "excludedGivers": ["a@x.com"] }
  ]
}
```

## Go validation (`validator` tags, run in `cmd/send.go` `validate()`)

- `template` required; `subject`, `body` required; `sender` required + email format.
- `mailgun` `required_without=Smtp`, `smtp` `required_without=Mailgun`; inner fields required
  (`smtp.port` optional, 1–65535). `validate(dat, requireMailer=false)` drops the
  `required_without` errors, used by `send --dry-run` and `preview` without `--to`; `check`
  reports a missing mailer as its own ✗ line.
- `participants` required, `min=2`, `unique=Email`, each: `email` required+email, `salutation` required.
- `extraRecipients` each: `salutation` required.
- Custom messages only exist for the smtp/mailgun `required_without` case (`translations` map);
  everything else prints validator's default error string.

### Not caught by the validator

Cross-references between entries are checked by `app.validatePreferences`
(`internal/app/preferences.go`), run at the start of `app.Pair`. It reports **all** problems at
once (`errors.Join`, each with the offending email + who referenced it):

- `excludedRecipients` / `excludedGivers` email that isn't a participant;
- `predestinedRecipient` that isn't a participant, is the giver themselves, is also in the same
  giver's `excludedRecipients`, or is predestined by several givers.

Excluding yourself in `excludedRecipients` is allowed (harmless no-op).

## JSON schema (`schema.json`)

Mirrors the structs loosely: root `required` (participants, template), `anyOf` smtp/mailgun,
`participants.minItems: 2`. Guarded by `schema_test.go`.
When changing the data model, update **all four**: `data.go` structs/tags, `schema.json`,
`data.example.json` + `cmd/init.go` sample, and README.

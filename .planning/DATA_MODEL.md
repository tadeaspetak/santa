# Data model (`data.json`)

Canonical example: `data.example.json`. Go source of truth: `internal/data/data.go`.
Editor help: `schema.json` (referenced via `$schema`, fetched from GitHub `main`, so schema
changes only reach users after merging to `main`).

```jsonc
{
  "$schema": "https://raw.githubusercontent.com/tadeaspetak/santa/refs/heads/main/schema.json",
  "smtp":    { "host": "...", "user": "...", "pass": "..." },   // either smtp…
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
- `mailgun` `required_without=Smtp`, `smtp` `required_without=Mailgun`; inner fields required.
- `participants` required, `min=2`, `unique=Email`, each: `email` required+email, `salutation` required.
- `extraRecipients` each: `salutation` required.
- Custom messages only exist for the smtp/mailgun `required_without` case (`translations` map);
  everything else prints validator's default error string.

### Not caught by the validator

- That `excludedRecipients` / `excludedGivers` reference existing participants (typos are
  silently ignored).
- `predestinedRecipient` problems are caught by `app.Pair` (not the validator) and returned as
  errors: unknown email, giver predestined to themselves, one recipient predestined by two givers.
  A predestined recipient that's also in the giver's `excludedRecipients` is not flagged.

## JSON schema (`schema.json`)

Mirrors the structs loosely: root `required` (participants, template), `anyOf` smtp/mailgun,
`participants.minItems: 2`. Guarded by `schema_test.go`.
When changing the data model, update **all four**: `data.go` structs/tags, `schema.json`,
`data.example.json` + `cmd/init.go` sample, and README.

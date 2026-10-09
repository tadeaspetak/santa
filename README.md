# Secret Santa 🎅

Santa is a simple app written in Go that makes drawing your Secret Santa pairings a breeze.

> 💡 Secret Santa is a gift-exchange tradition where participants anonymously give gifts to a randomly assigned person. I have come to love this way of Christmas due to its promoting quality instead of quantity.

1. Using a simple JSON file, define your participants and the template for constructing emails.
2. Supported preferences include
   - excluded recipients
   - predestined recipients
   - extra recipients with excluded givers
3. Check your setup, preview the email and rehearse the whole thing without spoiling any surprises.
4. Once ready, fire off your stealthy emails using a simple SMTP config or your Mailgun account.

```sh
santa init                        # create data.json, then edit it
santa check                       # is everything ready? (sends nothing)
santa preview --to you@email.com  # send yourself one sample email
santa send --dry-run              # rehearse without sending anything
santa send --to you@email.com     # rehearse by sending all emails to yourself
santa send                        # the real thing 🎅
```

## Setup 🎄

First of, download the app from [https://github.com/tadeaspetak/santa/releases/latest](https://github.com/tadeaspetak/santa/releases/latest).

> ⚠️ If you're on Mac, make the file executable with `chmod +x santa-mac`. When running, you'll have to override the security check and allow the app to run (click the ❔).

1. Run the `init` command from your terminal, like `./santa-mac init` (the examples below just use `santa`).
2. This command generates a `data.json` file with a sample family, email template and SMTP config. Use `--data other.json` with any command to use a different file.
3. Use your favourite editor to edit the `data.json` file. The attached JSON schema will help you with the config.
4. Run `santa check` any time to see whether your data and mailer are ready (more in [Before sending](#before-sending)).

### Mailer

You can send emails using either a simple SMTP server or a Mailgun account.

#### SMTP

You can use whatever SMTP you have at your disposal. Since `gmail` accounts are so ubiquitous, let's look at how to use your `gmail` account for sending:

1. `Enable IMAP` in your settings at https://mail.google.com/mail/u/0/#settings/fwdandpop.
2. Create an app password at https://myaccount.google.com/apppasswords.

Once done, enter your `smtp` config as follows:

```json
"smtp": {
  "host": "smtp.gmail.com",
  "user": "your.email@gmail.com",
  "pass": "xxx yyyy zzz"
},
```

The port defaults to `587`; set `"port"` if your server uses a different one. `santa check` tells you whether the login works without sending anything.

### Mailgun

If you'd like to use Mailgun, you will need a Mailgun sending domain and an API key to send emails to your participants.

1. If you don't have an account, create one at https://signup.mailgun.com/.
2. If you don't have a verified sending domain, create one such as `mg.your-domain.com`. Verify the sending domain by setting the correct DNS records as per Mailgun instructions.
3. Create a new API key in `Settings` -> `API Security`.

Once done, use the `mailgun` prop in the following way:

```json
"mailgun": {
  "domain": "mg.your-domain.com",
  "apiKey": "xxx"
}
```

### Email template

Each participant will receive an email constructed according to your template. The email should inform them of who they shall be picking a gift for this year.

This is where the `template` section enters the scene:

1. In `subject` and `body`, you have the `%{recipientSalutation}` variable at your disposal. This will be replaced by the value of the `salutation` from the recipient(s).
2. `body` is automatically wrapped in `<html><body>`.
3. When using Mailgun, `sender` must be on the domain from the [Mailgun section](#mailgun).
   `senderName` (optional) is the name mailboxes show instead of the `sender` address.
4. The `recipientsSeparator` separates multiple gift-recipients. This is only important in cases where [extra recipients](#extra-recipients) have been supplied.

Consider the following as a reasonable starting point:

```json
"template": {
  "subject": "🎄 Find a gift for %{recipientSalutation}",
  "body": "<p>Hi</p><p>Come up with something lovely for %{recipientSalutation}.</p><p>Happy hunting,<br/>Your Santa 🎅</p>",
  "sender": "santa@mg.your-domain.com",
  "senderName": "Santa",
  "recipientsSeparator": " and "
}
```

### Participants

The `participants` array defines the participants in your Secret Santa. Each participant has the following attributes:

- `email` (required): participant email address (must be unique)
- `salutation` (required): a salutation used when replacing the `%{recipientSalutation}` variable in the [email template](#email-template)
- `excludedRecipients` (optional): a list of emails of the recipients that should **not** be considered valid for this gift-giver
- `predestinedRecipient` (optional): an email of the recipient this gift-giver should be assigned (can't be excluded by the same gift-giver, and each participant can be predestined for one gift-giver at most)

All emails in `excludedRecipients` and `predestinedRecipient` must belong to other participants. Any typos are reported before anything is sent.

In the example below, `Mom` and `Dad` won't ever need to give a gift to each other; they like it that way. `Emily` explicitly asked to give a gift to `Auntie` since she has the perfect idea, so we've made `Auntie` her predestined recipient.

```json
"participants": [
    {
      "email": "mom@family.com",
      "salutation": "Mom",
      "excludedRecipients": ["dad@family.com"]
    },
    {
      "email": "dad@family.com",
      "salutation": "Dad",
      "excludedRecipients": ["mom@family.com"]
    },
    {
      "email": "auntie@family.com",
      "salutation": "Auntie",
    },
    {
      "email": "emily@family.com",
      "salutation": "Emily",
      "predestinedRecipient": "auntie@family.com"
    },
    {
      "email": "jake@family.com",
      "salutation": "Jake",
    }
  ]
```

#### Extra recipients

Say that you have babies, toddlers or small kids in your family. You might want them to **receive** presents, but they wouldn't be giving anything to anyone quite yet.

That's where the `extraRecipients` property comes in. You define those who will be assigned to the gift-givers (`participants`) as gift-recipients, but they will **not** be giving presents to anyone.

You can use the `excludedGivers` property on each of the `extraRecipients` to make sure certain people will **not** get that particular extra recipient assigned. Just like above, these must be emails of existing participants.

```json
"extraRecipients": [
  {
    "salutation": "Auntie's toddler",
    "excludedGivers": ["auntie@family.com"]
  },
  {
    "salutation": "Jake's toddler",
    "excludedGivers": ["jake@family.com", "emily@family.com"]
  }
]
```

Note that you might also want to adjust the `recipientsSeparator` in the `template`.

## <a name="sending">Sending</a>

Alright, so you've set up your mailer, template and participants, and now you're ready to send it all out 🦉 Before bombarding your loved ones with a barrage of potentially flawed emails, `santa`'s got your back.

### Before sending

Work your way through these steps. None of them reveal who got whom (unless you peek into the output folder).

1. **`santa check`** makes sure that
   - your data file is valid,
   - your preferences (predestined & excluded recipients, excluded givers) make a valid pairing possible,
   - your mailer accepts your credentials. It only logs in, nothing is sent.
2. **`santa preview`** prints a sample email in your terminal, using a made-up recipient ("Rudolph") instead of a real one. **`santa preview --to you@email.com`** sends that sample to your mailbox so you can see what it looks like in an actual email client.
3. **`santa send --dry-run`** (`-d`) does everything except sending: it draws the pairings and writes each email into the [output folder](#output-folder). No mailer config is needed, so you can try this with the example data straight away.
4. **`santa send --to you@email.com`** (`-t`) is a full rehearsal: the emails are generated as usual and actually sent, but all of them go to you. The `Reply-To` header of each shows the participant it was meant for. Each run draws new pairings, so there's nothing to spoil.

### Let's roll already

Run `santa send` without any flags. It tells you how many emails are about to go out and waits for your confirmation. Just pressing Enter cancels, so type `y` when you're sure. Use `--yes` (`-y`) to skip the question, e.g. in scripts.

Good luck 🛷

### Output folder

Every `send` creates a `santa-batch-<datetime>` folder with a copy of each email, named after the gift-giver it's for (e.g. `mom@family.com.txt`):

1. If you e.g. discover a typo in someone's email, you don't have to scrape the whole batch. You'd just look at who that particular gift-giver was supposed to get as a recipient and tell them. Not perfect, but heaps better than nothing.
2. If sending fails halfway through (e.g. your mail provider hits a limit), the error tells you how many emails went out and the folder has the rest.
3. It's excellent for debugging.

Steer away from opening those files on real runs to keep yourself in the dark together with everyone else.

### Printable PDF

If you'd like a printable PDF too, add the `--pdf` (`-p`) flag. It's saved as `emails.pdf` in the output folder. The construction of the PDF is simple:

1. The first and the last pages are empty so that if you open / print this file, you won't see anything by mistake.
2. All other pages have the gift-giver (=email recipient) at the top half of the page and the message (subject + body) at the bottom half. That way, you can always e.g. fold the pages in such a way that you won't see anything.

See the [printed PDF for the example data](https://github.com/tadeaspetak/santa/blob/main/data.example.pdf). If you're only interested in the PDF, run `santa send --dry-run --pdf`.

### All commands & flags

| Command | Flags | |
| --- | --- | --- |
| `santa init` | | create a sample data file |
| `santa check` | | validate data, pairing & mailer login, sends nothing |
| `santa preview` | `--to`, `-t` | print a sample email, or send it to the given address |
| `santa send` | `--dry-run`, `-d` | don't send anything |
| | `--to`, `-t` | send all emails to the given address instead |
| | `--pdf`, `-p` | also generate a printable PDF |
| | `--yes`, `-y` | don't ask for confirmation |
| `santa version` | | print the version |

All commands accept `--data <path>` (default `data.json`) and `--help`.

## Dev notes

Run the tests with `make test`. Besides unit tests, `cmd/e2e_test.go` runs whole `santa` commands against fake SMTP and Mailgun servers (see `internal/testutil`), so the full flow, prompts included, is tested without sending a single real email.

To use the git hooks in the `hooks` directory, run `git config core.hooksPath hooks`.

To rehearse the whole flow without any real emails going out, point the `smtp` config at a local mail catcher such as [Mailpit](https://mailpit.axllent.org) (`"host": "localhost", "port": 1025`, any `user` and `pass`).

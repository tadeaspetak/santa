# Secret Santa 🎅

Santa is a simple CLI app written in Go that makes drawing your Secret Santa pairings a breeze.

> 💡 Secret Santa is a gift-exchange tradition where participants anonymously give gifts to a randomly assigned person. I have come to love this way of Christmas due to its promoting quality instead of quantity.

The app enables you to define your participants and the email(s) they will be receiving. Preferences include excluded or predestined recipients, and extra recipients with potentially excluded givers. Santa makes it simple to debug and even print the pairings before randomising the gift-givers and recipients (taking everything into account) and firing off your stealthy emails using your Mailgun account.

1. Start by [setting up the app](#lets-get-you-going-).
1. Once set up, [test and send your stealthy emails](#sending).

## Let's get you going 🎄

First of, download the app from [https://github.com/tadeaspetak/santa/releases/latest](https://github.com/tadeaspetak/santa/releases/latest).

> ⚠️ If you're on Mac, make the file executable with `chmod +x santa-mac`. When running, you'll have to override the security check and allow the app to run (click the ❔).

Run the `init` command from your terminal, like `./santa-mac init`. This command will help you with the basic setup:

1. [mailgun](#mailgun) config necessary for sending emails to your participants
2. [email template](#template) used for constructing those emails
3. [participant definitions](#participants) for the app to know who's who and what their Santa-related preferences are

All settings are saved in a `data.json` file, also easily readable & editable by a human. You can override the data-file path using the `-data` flag.

### Mailgun

You need a Mailgun sending domain and an API key to send emails to your participants.

1. If you don't have an account, create one at https://signup.mailgun.com/.
2. If you don't have a verified sending domain, create one such as `mg.your-domain.com`. Verify the sending domain by setting the correct DNS records as per Mailgun instructions.
3. Create a new API key in `Settings` -> `API Security`.

Once done, use the `mailgun` command to enter the domain and your API key.

### Email template

Each participant will receive an email constructed based on your template. The email should inform them of who they shall be picking a gift for this year. Use the `template` command to adjust your template.

1. In `subject` and `body`, you have the `%{recipientSalutation}` variable at your disposal. This will be replaced by the value of the `salutation` from the recipient.
2. `body` is automatically wrapped in `<html><body>`.
3. `sender` must be on the domain from the [Mailgun section](#mailgun).
4. The `recipientsSeparator` separates multiple gift-recipients. This is only important in cases where [extra recipients](#extra-recipients) have been supplied.

Consider the following as a reasonable starting point:

```json
{
  "subject": "🎄 Find a gift for %{recipientSalutation}",
  "body": "<p>Hi</p><p>Come up with something lovely for %{recipientSalutation}.</p><p>Happy hunting,<br/>Your Santa 🎅</p>",
  "sender": "santa@mg.your-domain.com",
  "recipientsSeparator": " and "
}
```

### Participants

The `participants` array defines the participants in your Secret Santa. To manage your participants, use the `participants` command folder (`add`, `delete` and `edit`).

Each participant has the following attributes:

- `email` (required): participant email address (must be unique)
- `salutation` (required): a salutation used when replacing the `%{recipientSalutation}` variable in the [email template](#email-template)
- `excludedRecipients` (optional): a list of emails of the recipients that should **not** be considered valid for this gift-giver
- `predestinedRecipient` (optional): an email of the recipient this gift-giver should be assigned

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

You can use the `excludedGivers` property on each of the `extraRecipients` to make sure certain people will **not** get that particular extra recipient assigned.

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

> ❗ You'll have to edit these manully in your `data.json` file, they don't have any CLI support (yet).

## <a name="sending">Sending</a>

Alright, so you've set up your Mailgun, template and participants, and now you're ready to send it all out 🦉 This is where the `send` command finally enters the game. Before using it, read on.

### Logs

Every (about to be) sent email is saved in a `santa-batch-%datetime-%giftgiver.txt` file:

1. If you e.g. discover a typo in someone's email, you don't have to scrape the whole batch. You'd just look at who that particular gift-giver was supposed to get as a recipient and tell them. Not perfect, but heaps better than nothing.
2. It's excellent for debugging.

Steer away from opening those files on real runs to keep yourself in the dark together with everyone else.

### Printable PDF

If you'd like to generate a printable PDF when sending a batch, use the `-p` flag. The construction of the PDF is simple:

1. The first and the last pages are empty so that if you open / print this file, you won't see anything by mistake.
2. All other pages have the gift-giver (=email recipient) at the top half of the page and the message (subject + body) at the bottom half. That way, you can always e.g. fold the pages in such a way that you won't see anything.

See the [printed PDF for the example data](https://github.com/tadeaspetak/santa/blob/main/data.example.pdf). If you're only interested in generating such a PDF, run `send  -p -d` (more details on the `-d` flag below).

### Debugging & testing

Before bombarding your loved ones with a barrage of potentially flawed emails with a potentially flawed setup, `santa`'s got your back:

1. Use the `-d` flag with the `send` command for debugging purposes. That way, everything runs as per usual, but no emails are sent. Check the batch files to make sure your emails are correct and the results are what you expect them to be.
2. When you're ready, you can do a test run with the `-a always.send.to@email.com` flag. The `-a` flag generates emails as per usual and actually sends them (unless `-d` is also used). But, it always sends them to the hardcoded email address, presumably you. You can find the gift-giver this would have normally been sent to in the `reply-to` header.

### Let's roll already

When you're ready to run the `send` command without any flags, good luck 🛷

## Dev notes

To use the git hooks in the `hooks` directory, run `git config core.hooksPath hooks`.

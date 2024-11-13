# Secret Santa 🎅

Santa is a simple CLI app that makes drawing your Secret Santa pairings a breeze.

> 💡 Secret Santa is a gift-exchange tradition where participants anonymously give gifts to a randomly assigned person. I have personally come to love this due to its promoting quality instead of quantity.

## Let's get you going 🎄

First of, download the app from TODO.

Once downloaded, run the `init` from your terminal like `./secret-santa init` (TODO the `init` command)(TODO: do you need to change permissions to executable?). This command will walk you through the basic setup:

1. [mailgun](#mailgun) config necessary for sending emails to your participants
2. [email template](#email-template) used for constructing those emails
3. [participant definitions](#participants) for the app to know who's who and what their Santa-related preferences are

### <a name="mailgun">Mailgun</a>

You need a Mailgun sending domain and an API key to send emails to your participants.

1. If you don't have an account, create one at https://signup.mailgun.com/.
2. If you don't have a verified sending domain, create one such as `mg.your-domain.com`. Verify the sending domain by setting the correct DNS records as per Mailgun instructions.
3. Create a new API key in `Settings` -> `API Security`.

Once done, use the `mailgun` command to enter the domain and your API key.

### <a name="email-template">Email template</a>

Each participant will receive an email constructed based on your template. The email should inform them of who they shall be picking a gift for this year.

1. In `subject` and `body`, you have the `%{recipientSalutation}` variable at your disposal. This will be replaced by the value of the `salutation` from the recipient.
2. `body` is automatically wrapped in `<html><body>`.
3. `sender` must be on the domain from the [Mailgun section](#mailgun).

Consider the following as a reasonable starting point:

```json
{
  "subject": "🎄 Find a gift for %{recipientSalutation}",
  "body": "<p>Hi</p><p>Come up with something lovely for %{recipientSalutation}.</p><p>Happy hunting,<br/>Your 🎅</p>",
  "sender": "santa@mg.your-domain.com"
}
```

### <a name="participant">Participants</a>

The `participants` array defines the participants in your Secret Santa. You have the following options:

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
      "predestinedRecipient": "aunt@family.com"
    },
    {
      "email": "jake@family.com",
      "salutation": "Jake",
    }
  ]
```

### Dev notes

To use the git hooks in the `hooks` directory, run `git config core.hooksPath hooks`.

TODO:

- figure out releases

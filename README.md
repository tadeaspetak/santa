# Secret Santa 🎅

Secret Santa is a simple CLI app that makes drawing your Secret Santa pairings a breeze.

TODO: what's secret santa?

## Let's get you going 🎄

First of, download the app from TODO.

Once downloaded, run the `init` from your terminal like `./secret-santa init` (TODO the `init` command)(TODO: do you need to change permissions to executable?). This command will walk you through the basic setup:

1. [mailgun](#mailgun) config necessary for sending emails to your participants
2. [email template](#template) used for constructing those emails
3. [participant definitions](#participants) for the app to know who's who and what their Santa-related preferences are

### <a name="mailgun">Mailgun</a>

You need a Mailgun sending domain and an API key to send emails to your participants.

1. If you don't have an account, create one at https://signup.mailgun.com/.
2. If you don't have a verified sending domain, create one such as `mg.your-domain.com`. Verify the sending domain by setting the correct DNS records as per Mailgun instructions.
3. Create a new API key in `Settings` -> `API Security`.

Once done, use the `mailgun` command to enter the domain and your API key.

### <a name="template">Email template</a>

TODO

### <a name="participant">Participants</a>

TODO

### Dev notes

Use hooks in the `hooks` directory: `git config core.hooksPath hooks`.

TODO:

- finish readme
- figure out releases

# TODO

- [ ] Create
- [ ] Finish the readme
- [ ] Update the `data.example.json`
- [ ] Make sure linting, formatting etc. is part of the build; use https://staticcheck.dev/docs/running-staticcheck/ instead of the deprecated golint.
- [ ] Tests
  - how to run all tests in a mod
- [ ] validation
      err = validation.Validate.Struct(data)
      if err != nil {
      log.Fatal(data, err)
      }

### Notes

If set, the `email.recipient` prop overrides the actual email recipients to allow you to easilly test what you are about to send.

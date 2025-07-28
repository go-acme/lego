# E2E tests

- Install [Pebble](https://github.com/letsencrypt/pebble):
```bash
go install github.com/letsencrypt/pebble/v2/cmd/pebble@main
go install github.com/letsencrypt/pebble/v2/cmd/pebble-challtestsrv@main
```

- Launch tests:
```bash
make e2e
```

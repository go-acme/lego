# E2E tests

- Install [Pebble](https://github.com/letsencrypt/pebble):
```bash
go install github.com/letsencrypt/pebble/v2/cmd/pebble@v2.9.0
go install github.com/letsencrypt/pebble/v2/cmd/pebble-challtestsrv@v2.9.0
```

- Launch tests:
```bash
make e2e
```

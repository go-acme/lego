# E2E tests

- Install [Pebble](https://github.com/letsencrypt/pebble):
```bash
go install github.com/letsencrypt/pebble/v2/cmd/pebble@v2.10.1
go install github.com/letsencrypt/pebble/v2/cmd/pebble-challtestsrv@v2.10.1
```

- Launch tests:
```bash
make e2e
```

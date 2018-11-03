# E2E tests

How to run:

- Add the following entry to your `/etc/hosts`:
```
127.0.0.1	acme.wtf
```

- Install [Pebble](https://github.com/letsencrypt/pebble):
```bash
go get -u github.com/letsencrypt/pebble/...
```

- Launch tests:
```bash
make e2e
```

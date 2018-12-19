# E2E tests

How to run:

- Add the following entries to your `/etc/hosts`:
```
127.0.0.1	acme.wtf
127.0.0.1	lego.wtf
127.0.0.1	acme.lego.wtf
127.0.0.1	légô.wtf
127.0.0.1	xn--lg-bja9b.wtf
```

- Install [Pebble](https://github.com/letsencrypt/pebble):
```bash
go get -u github.com/letsencrypt/pebble/...
```

- Launch tests:
```bash
make e2e
```

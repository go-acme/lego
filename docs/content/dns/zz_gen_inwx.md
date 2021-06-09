---
title: "INWX"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: inwx
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/inwx/inwx.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v2.0.0

Configuration for [INWX](https://www.inwx.de/en).


<!--more-->

- Code: `inwx`

Here is an example bash command using the INWX provider:

```bash
INWX_USERNAME=xxxxxxxxxx \
INWX_PASSWORD=yyyyyyyyyy \
lego --email myemail@example.com --dns inwx --domains my.example.org run

# 2FA
INWX_USERNAME=xxxxxxxxxx \
INWX_PASSWORD=yyyyyyyyyy \
INWX_SHARED_SECRET=zzzzzzzzzz \
lego --email myemail@example.com --dns inwx --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `INWX_PASSWORD` | Password |
| `INWX_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `INWX_POLLING_INTERVAL` | Time between DNS propagation check |
| `INWX_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation (default 360s) |
| `INWX_SANDBOX` | Activate the sandbox (boolean) |
| `INWX_SHARED_SECRET` | shared secret related to 2FA |
| `INWX_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://www.inwx.de/en/help/apidoc)
- [Go client](https://github.com/nrdcg/goinwx)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/inwx/inwx.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

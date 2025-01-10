---
title: "INWX"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: inwx
dnsprovider:
  since:    "v2.0.0"
  code:     "inwx"
  url:      "https://www.inwx.de/en"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/inwx/inwx.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [INWX](https://www.inwx.de/en).


<!--more-->

- Code: `inwx`
- Since: v2.0.0


Here is an example bash command using the INWX provider:

```bash
INWX_USERNAME=xxxxxxxxxx \
INWX_PASSWORD=yyyyyyyyyy \
lego --email you@example.com --dns inwx -d '*.example.com' -d example.com run

# 2FA
INWX_USERNAME=xxxxxxxxxx \
INWX_PASSWORD=yyyyyyyyyy \
INWX_SHARED_SECRET=zzzzzzzzzz \
lego --email you@example.com --dns inwx -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `INWX_PASSWORD` | Password |
| `INWX_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `INWX_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `INWX_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 360) |
| `INWX_SANDBOX` | Activate the sandbox (boolean) |
| `INWX_SHARED_SECRET` | shared secret related to 2FA |
| `INWX_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.inwx.de/en/help/apidoc)
- [Go client](https://github.com/nrdcg/goinwx)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/inwx/inwx.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

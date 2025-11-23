---
title: "Gigahost.no"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: gigahostno
dnsprovider:
  since:    "v4.29.0"
  code:     "gigahostno"
  url:      "https://gigahost.no/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gigahostno/gigahostno.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Gigahost.no](https://gigahost.no/).


<!--more-->

- Code: `gigahostno`
- Since: v4.29.0


Here is an example bash command using the Gigahost.no provider:

```bash
GIGAHOSTNO_USERNAME="xxxxxxxxxxxxxxxxxxxxx" \
GIGAHOSTNO_PASSWORD="yyyyyyyyyyyyyyyyyyyyy" \
lego --email you@example.com --dns gigahostno -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `GIGAHOSTNO_PASSWORD` | Password |
| `GIGAHOSTNO_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GIGAHOSTNO_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `GIGAHOSTNO_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `GIGAHOSTNO_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `GIGAHOSTNO_SECRET` | TOTP secret |
| `GIGAHOSTNO_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://gigahost.no/api-dokumentasjon)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gigahostno/gigahostno.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

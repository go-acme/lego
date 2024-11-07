---
title: "dnsHome.de"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dnshomede
dnsprovider:
  since:    "v4.10.0"
  code:     "dnshomede"
  url:      "https://www.dnshome.de"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnshomede/dnshomede.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [dnsHome.de](https://www.dnshome.de).


<!--more-->

- Code: `dnshomede`
- Since: v4.10.0


Here is an example bash command using the dnsHome.de provider:

```bash
DNSHOMEDE_CREDENTIALS=example.org:password \
lego --email you@example.com --dns dnshomede -d '*.example.com' -d example.com run

DNSHOMEDE_CREDENTIALS=my.example.org:password1,demo.example.org:password2 \
lego --email you@example.com --dns dnshomede -d my.example.org -d demo.example.org
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DNSHOMEDE_CREDENTIALS` | Comma-separated list of domain:password credential pairs |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DNSHOMEDE_HTTP_TIMEOUT` | API request timeout |
| `DNSHOMEDE_POLLING_INTERVAL` | Time between DNS propagation checks |
| `DNSHOMEDE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation; defaults to 300s (5 minutes) |
| `DNSHOMEDE_SEQUENCE_INTERVAL` | Time between sequential requests |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).





<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnshomede/dnshomede.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

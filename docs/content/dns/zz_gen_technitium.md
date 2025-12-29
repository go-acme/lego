---
title: "Technitium"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: technitium
dnsprovider:
  since:    "v4.20.0"
  code:     "technitium"
  url:      "https://technitium.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/technitium/technitium.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Technitium](https://technitium.com/).


<!--more-->

- Code: `technitium`
- Since: v4.20.0


Here is an example bash command using the Technitium provider:

```bash
TECHNITIUM_SERVER_BASE_URL="https://localhost:5380" \
TECHNITIUM_API_TOKEN="xxxxxxxxxxxxxxxxxxxxx" \
lego --dns technitium -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `TECHNITIUM_API_TOKEN` | API token |
| `TECHNITIUM_SERVER_BASE_URL` | Server base URL |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `TECHNITIUM_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `TECHNITIUM_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `TECHNITIUM_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `TECHNITIUM_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

Technitium DNS Server supports Dynamic Updates (RFC2136) for primary zones,
so you can also use the [RFC2136 provider](https://go-acme.github.io/lego/dns/rfc2136/index.html).

[RFC2136 provider](https://go-acme.github.io/lego/dns/rfc2136/index.html) is much better compared to the HTTP API option from security perspective.
Technitium recommends to use it in production over the HTTP API.



## More information

- [API documentation](https://github.com/TechnitiumSoftware/DnsServer/blob/0f83d23e605956b66ac76921199e241d9cc061bd/APIDOCS.md)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/technitium/technitium.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

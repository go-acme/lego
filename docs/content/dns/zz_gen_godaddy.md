---
title: "Go Daddy"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: godaddy
dnsprovider:
  since:    "v0.5.0"
  code:     "godaddy"
  url:      "https://godaddy.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/godaddy/godaddy.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Go Daddy](https://godaddy.com).


<!--more-->

- Code: `godaddy`
- Since: v0.5.0


Here is an example bash command using the Go Daddy provider:

```bash
GODADDY_API_KEY=xxxxxxxx \
GODADDY_API_SECRET=yyyyyyyy \
lego --email you@example.com --dns godaddy -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `GODADDY_API_KEY` | API key |
| `GODADDY_API_SECRET` | API secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GODADDY_HTTP_TIMEOUT` | API request timeout |
| `GODADDY_POLLING_INTERVAL` | Time between DNS propagation check |
| `GODADDY_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `GODADDY_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

GoDaddy has recently (2024-04) updated the account requirements to access parts of their production Domains API:

- Availability API: Limited to accounts with 50 or more domains.
- Management and DNS APIs: Limited to accounts with 10 or more domains and/or an active Discount Domain Club plan.

https://community.letsencrypt.org/t/getting-unauthorized-url-error-while-trying-to-get-cert-for-subdomains/217329/12



## More information

- [API documentation](https://developer.godaddy.com/doc/endpoint/domains)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/godaddy/godaddy.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

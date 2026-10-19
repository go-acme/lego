---
title: "Go Daddy (v3)"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: godaddyv3
dnsprovider:
  since:    "v5.6.0"
  code:     "godaddyv3"
  url:      "https://godaddy.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/godaddyv3/godaddyv3.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Go Daddy (v3)](https://godaddy.com).


<!--more-->

- Code: `godaddyv3`
- Since: v5.6.0


Here is an example bash command using the Go Daddy (v3) provider:

```bash
GODADDYV3_PERSONAL_ACCESS_TOKEN="xxxxxxxxxxxxxxxxxxxxx" \
lego run --dns godaddyv3 -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `GODADDYV3_PERSONAL_ACCESS_TOKEN` | Personal Access Token (PAT) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GODADDYV3_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `GODADDYV3_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `GODADDYV3_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `GODADDYV3_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

GoDaddy has recently (2024-04) updated the account requirements to access parts of their production Domains API:

- Availability API: Limited to accounts with 50 or more domains.
- Management and DNS APIs: Limited to accounts with 10 or more domains and/or an active Discount Domain Club plan.

https://community.letsencrypt.org/t/getting-unauthorized-url-error-while-trying-to-get-cert-for-subdomains/217329/12



## More information

- [API documentation](https://developer.godaddy.com/en/docs/references/rest/domains/v3)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/godaddyv3/godaddyv3.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

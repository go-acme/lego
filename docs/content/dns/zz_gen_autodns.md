---
title: "Autodns"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: autodns
dnsprovider:
  since:    "v3.2.0"
  code:     "autodns"
  url:      "https://www.internetx.com/domains/autodns/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/autodns/autodns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Autodns](https://www.internetx.com/domains/autodns/).


<!--more-->

- Code: `autodns`
- Since: v3.2.0


Here is an example bash command using the Autodns provider:

```bash
AUTODNS_API_USER=username \
AUTODNS_API_PASSWORD=supersecretpassword \
lego --email you@example.com --dns autodns -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `AUTODNS_API_PASSWORD` | User Password |
| `AUTODNS_API_USER` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `AUTODNS_CONTEXT` | API context (4 for production, 1 for testing. Defaults to 4) |
| `AUTODNS_ENDPOINT` | API endpoint URL, defaults to https://api.autodns.com/v1/ |
| `AUTODNS_HTTP_TIMEOUT` | API request timeout, defaults to 30 seconds |
| `AUTODNS_POLLING_INTERVAL` | Time between DNS propagation check |
| `AUTODNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `AUTODNS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://help.internetx.com/display/APIJSONEN)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/autodns/autodns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

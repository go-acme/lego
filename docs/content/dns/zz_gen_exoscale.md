---
title: "Exoscale"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: exoscale
dnsprovider:
  since:    "v0.4.0"
  code:     "exoscale"
  url:      "https://www.exoscale.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/exoscale/exoscale.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Exoscale](https://www.exoscale.com/).


<!--more-->

- Code: `exoscale`
- Since: v0.4.0


Here is an example bash command using the Exoscale provider:

```bash
EXOSCALE_API_KEY=abcdefghijklmnopqrstuvwx \
EXOSCALE_API_SECRET=xxxxxxx \
lego --dns exoscale -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `EXOSCALE_API_KEY` | API key |
| `EXOSCALE_API_SECRET` | API secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `EXOSCALE_ENDPOINT` | API endpoint URL |
| `EXOSCALE_HTTP_TIMEOUT` | API request timeout in seconds (Default: 60) |
| `EXOSCALE_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `EXOSCALE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `EXOSCALE_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://openapi-v2.exoscale.com/#endpoint-dns)
- [Go client](https://github.com/exoscale/egoscale)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/exoscale/exoscale.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

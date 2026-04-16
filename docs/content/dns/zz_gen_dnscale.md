---
title: "DNScale"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dnscale
dnsprovider:
  since:    "v4.22.0"
  code:     "dnscale"
  url:      "https://dnscale.eu"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnscale/dnscale.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [DNScale](https://dnscale.eu).


<!--more-->

- Code: `dnscale`
- Since: v4.22.0


Here is an example bash command using the DNScale provider:

```bash
DNSCALE_API_TOKEN=xxxxxxxx \
lego --dns dnscale -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DNSCALE_API_TOKEN` | API token (create at app.dnscale.eu with records:read and records:write scopes) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DNSCALE_API_URL` | API URL (Default: https://api.dnscale.eu) |
| `DNSCALE_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `DNSCALE_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 5) |
| `DNSCALE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `DNSCALE_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.dnscale.eu)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnscale/dnscale.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

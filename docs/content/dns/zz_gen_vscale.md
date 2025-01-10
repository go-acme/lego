---
title: "Vscale"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: vscale
dnsprovider:
  since:    "v2.0.0"
  code:     "vscale"
  url:      "https://vscale.io/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/vscale/vscale.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Vscale](https://vscale.io/).


<!--more-->

- Code: `vscale`
- Since: v2.0.0


Here is an example bash command using the Vscale provider:

```bash
VSCALE_API_TOKEN=xxxxx \
lego --email you@example.com --dns vscale -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `VSCALE_API_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `VSCALE_BASE_URL` | API endpoint URL |
| `VSCALE_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `VSCALE_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `VSCALE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `VSCALE_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://developers.vscale.io/documentation/api/v1/#api-Domains_Records)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/vscale/vscale.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

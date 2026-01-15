---
title: "DDnss (DynDNS Service)"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: ddnss
dnsprovider:
  since:    "v4.32.0"
  code:     "ddnss"
  url:      "https://ddnss.de/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ddnss/ddnss.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [DDnss (DynDNS Service)](https://ddnss.de/).


<!--more-->

- Code: `ddnss`
- Since: v4.32.0


Here is an example bash command using the DDnss (DynDNS Service) provider:

```bash
DDNSS_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego --dns ddnss -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DDNSS_KEY` | Update key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DDNSS_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `DDNSS_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `DDNSS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `DDNSS_SEQUENCE_INTERVAL` | Time between sequential requests in seconds (Default: 60) |
| `DDNSS_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://ddnss.de/info.php)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ddnss/ddnss.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

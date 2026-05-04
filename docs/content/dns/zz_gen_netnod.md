---
title: "Netnod"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: netnod
dnsprovider:
  since:    "v4.34.0"
  code:     "netnod"
  url:      "https://www.netnod.se/dns/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/netnod/netnod.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Netnod](https://www.netnod.se/dns/).


<!--more-->

- Code: `netnod`
- Since: v4.34.0


Here is an example bash command using the Netnod provider:

```bash
NETNOD_TOKEN="xxx" \
lego run --dns netnod -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NETNOD_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NETNOD_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `NETNOD_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `NETNOD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `NETNOD_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.google.com/document/d/1GfpCZjhdniWzM3fjTSS73D0uCUnK478IzcXBIOYOemQ/edit?usp=sharing)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/netnod/netnod.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

---
title: "Octenium"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: octenium
dnsprovider:
  since:    "v4.27.0"
  code:     "octenium"
  url:      "https://octenium.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/octenium/octenium.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Octenium](https://octenium.com/).


<!--more-->

- Code: `octenium`
- Since: v4.27.0


Here is an example bash command using the Octenium provider:

```bash
OCTENIUM_API_KEY="xxx" \
lego --dns octenium -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `OCTENIUM_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `OCTENIUM_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `OCTENIUM_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `OCTENIUM_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `OCTENIUM_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://octenium.com/api#tag/Domains-DNS)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/octenium/octenium.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

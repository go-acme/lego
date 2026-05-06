---
title: "Fornex"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: fornex
dnsprovider:
  since:    "v5.0.0"
  code:     "fornex"
  url:      "https://fornex.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/fornex/fornex.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Fornex](https://fornex.com/).


<!--more-->

- Code: `fornex`
- Since: v5.0.0


Here is an example bash command using the Fornex provider:

```bash
FORNEX_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego run --dns fornex -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `FORNEX_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `FORNEX_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `FORNEX_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `FORNEX_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `FORNEX_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://fornex.com/api/#tag/Domain:-Entry)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/fornex/fornex.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

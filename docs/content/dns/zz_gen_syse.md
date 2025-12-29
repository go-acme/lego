---
title: "Syse"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: syse
dnsprovider:
  since:    "v4.30.0"
  code:     "syse"
  url:      "https://www.syse.no/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/syse/syse.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Syse](https://www.syse.no/).


<!--more-->

- Code: `syse`
- Since: v4.30.0


Here is an example bash command using the Syse provider:

```bash
SYSE_CREDENTIALS=example.com:password \
lego --dns syse -d '*.example.com' -d example.com run

SYSE_CREDENTIALS=example.org:password1,example.com:password2 \
lego --dns syse -d '*.example.org' -d example.org -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SYSE_CREDENTIALS` | Comma-separated list of `zone:password` credential pairs |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SYSE_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `SYSE_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `SYSE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 1200) |
| `SYSE_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.syse.no/api/dns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/syse/syse.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

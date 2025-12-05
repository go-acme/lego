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
SYSE_PASSWORD="xxxxxxxxxxxxxxxxxxxxx" \
lego --email you@example.com --dns syse -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SYSE_PASSWORD` | Example |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SYSE_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `SYSE_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `SYSE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `SYSE_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.syse.no/api/dns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/syse/syse.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

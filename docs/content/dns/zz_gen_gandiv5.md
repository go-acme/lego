---
title: "Gandi Live DNS (v5)"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: gandiv5
dnsprovider:
  since:    "v0.5.0"
  code:     "gandiv5"
  url:      "https://www.gandi.net"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gandiv5/gandiv5.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Gandi Live DNS (v5)](https://www.gandi.net).


<!--more-->

- Code: `gandiv5`
- Since: v0.5.0


Here is an example bash command using the Gandi Live DNS (v5) provider:

```bash
GANDIV5_PERSONAL_ACCESS_TOKEN=abcdefghijklmnopqrstuvwx \
lego --email you@example.com --dns gandiv5 -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `GANDIV5_API_KEY` | API key (Deprecated) |
| `GANDIV5_PERSONAL_ACCESS_TOKEN` | Personal Access Token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GANDIV5_HTTP_TIMEOUT` | API request timeout in seconds (Default: 10) |
| `GANDIV5_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 20) |
| `GANDIV5_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 1200) |
| `GANDIV5_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.gandi.net/docs/livedns/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gandiv5/gandiv5.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

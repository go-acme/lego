---
title: "Gandi"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: gandi
dnsprovider:
  since:    "v0.3.0"
  code:     "gandi"
  url:      "https://www.gandi.net"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gandi/gandi.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Gandi](https://www.gandi.net).


<!--more-->

- Code: `gandi`
- Since: v0.3.0


Here is an example bash command using the Gandi provider:

```bash
GANDI_API_KEY=abcdefghijklmnopqrstuvwx \
lego --email you@example.com --dns gandi -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `GANDI_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GANDI_HTTP_TIMEOUT` | API request timeout in seconds (Default: 60) |
| `GANDI_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 60) |
| `GANDI_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 2400) |
| `GANDI_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://doc.rpc.gandi.net/index.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gandi/gandi.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

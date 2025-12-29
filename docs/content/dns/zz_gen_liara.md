---
title: "Liara"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: liara
dnsprovider:
  since:    "v4.10.0"
  code:     "liara"
  url:      "https://liara.ir"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/liara/liara.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Liara](https://liara.ir).


<!--more-->

- Code: `liara`
- Since: v4.10.0


Here is an example bash command using the Liara provider:

```bash
LIARA_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego --dns liara -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `LIARA_API_KEY` | The API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `LIARA_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `LIARA_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `LIARA_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `LIARA_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 3600) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://openapi.liara.ir/?urls.primaryName=DNS)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/liara/liara.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

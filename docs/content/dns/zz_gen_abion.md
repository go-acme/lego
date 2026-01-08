---
title: "Abion"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: abion
dnsprovider:
  since:    "v4.32.0"
  code:     "abion"
  url:      "https://abion.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/abion/abion.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Abion](https://abion.com).


<!--more-->

- Code: `abion`
- Since: v4.32.0


Here is an example bash command using the Abion provider:

```bash
ABION_API_KEY="xxxxxxxxxxxx" \
lego --dns abion -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ABION_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ABION_HTTP_TIMEOUT` | API request timeout in seconds (Default: 10) |
| `ABION_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `ABION_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `ABION_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://demo.abion.com/pmapi-doc/openapi-ui/index.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/abion/abion.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

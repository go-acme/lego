---
title: "Wannafind"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: wannafind
dnsprovider:
  since:    "v4.32.0"
  code:     "wannafind"
  url:      "https://www.wannafind.dk/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/wannafind/wannafind.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Wannafind](https://www.wannafind.dk/).


<!--more-->

- Code: `wannafind`
- Since: v4.32.0


Here is an example bash command using the Wannafind provider:

```bash
WANNAFIND_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego --dns wannafind -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `WANNAFIND_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `WANNAFIND_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `WANNAFIND_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `WANNAFIND_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `WANNAFIND_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.wannafind.dk/dns/swagger/index.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/wannafind/wannafind.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

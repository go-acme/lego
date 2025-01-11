---
title: "Vultr"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: vultr
dnsprovider:
  since:    "v0.3.1"
  code:     "vultr"
  url:      "https://www.vultr.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/vultr/vultr.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Vultr](https://www.vultr.com/).


<!--more-->

- Code: `vultr`
- Since: v0.3.1


Here is an example bash command using the Vultr provider:

```bash
VULTR_API_KEY=xxxxx \
lego --email you@example.com --dns vultr -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `VULTR_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `VULTR_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `VULTR_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `VULTR_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `VULTR_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.vultr.com/api/#dns)
- [Go client](https://github.com/vultr/govultr)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/vultr/vultr.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

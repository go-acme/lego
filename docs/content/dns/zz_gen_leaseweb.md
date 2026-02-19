---
title: "Leaseweb"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: leaseweb
dnsprovider:
  since:    "v4.32.0"
  code:     "leaseweb"
  url:      "https://www.leaseweb.com/en/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/leaseweb/leaseweb.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Leaseweb](https://www.leaseweb.com/en/).


<!--more-->

- Code: `leaseweb`
- Since: v4.32.0


Here is an example bash command using the Leaseweb provider:

```bash
LEASEWEB_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego --dns leaseweb -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `LEASEWEB_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `LEASEWEB_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `LEASEWEB_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `LEASEWEB_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `LEASEWEB_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://developer.leaseweb.com/docs/#tag/DNS)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/leaseweb/leaseweb.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

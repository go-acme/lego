---
title: "PrimeDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: primedns
dnsprovider:
  since:    "v5.6.0"
  code:     "primedns"
  url:      "https://primedns.net/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/primedns/primedns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [PrimeDNS](https://primedns.net/).


<!--more-->

- Code: `primedns`
- Since: v5.6.0


Here is an example bash command using the PrimeDNS provider:

```bash
PRIMEDNS_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego run --dns primedns -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `PRIMEDNS_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `PRIMEDNS_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `PRIMEDNS_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `PRIMEDNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 180) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.primedns.net/v1/docs)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/primedns/primedns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

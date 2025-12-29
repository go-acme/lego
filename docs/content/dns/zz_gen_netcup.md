---
title: "Netcup"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: netcup
dnsprovider:
  since:    "v1.1.0"
  code:     "netcup"
  url:      "https://www.netcup.eu/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/netcup/netcup.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Netcup](https://www.netcup.eu/).


<!--more-->

- Code: `netcup`
- Since: v1.1.0


Here is an example bash command using the Netcup provider:

```bash
NETCUP_CUSTOMER_NUMBER=xxxx \
NETCUP_API_KEY=yyyy \
NETCUP_API_PASSWORD=zzzz \
lego --dns netcup -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NETCUP_API_KEY` | API key |
| `NETCUP_API_PASSWORD` | API password |
| `NETCUP_CUSTOMER_NUMBER` | Customer number |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NETCUP_HTTP_TIMEOUT` | API request timeout in seconds (Default: 10) |
| `NETCUP_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 30) |
| `NETCUP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 900) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.netcup-wiki.de/wiki/DNS_API)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/netcup/netcup.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

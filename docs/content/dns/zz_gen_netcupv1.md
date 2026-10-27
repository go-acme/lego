---
title: "Netcup v1"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: netcupv1
dnsprovider:
  since:    "v5.6.0"
  code:     "netcupv1"
  url:      "https://www.netcup.eu/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/netcupv1/netcupv1.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Netcup v1](https://www.netcup.eu/).


<!--more-->

- Code: `netcupv1`
- Since: v5.6.0


Here is an example bash command using the Netcup v1 provider:

```bash
NETCUPV1_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego run --dns netcupv1 -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NETCUPV1_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NETCUPV1_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `NETCUPV1_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 30) |
| `NETCUPV1_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 900) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.netcup.com/en/helpcenter/documentation/netcup-api#/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/netcupv1/netcupv1.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

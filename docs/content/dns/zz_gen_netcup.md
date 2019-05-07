---
title: "Netcup"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: netcup
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/netcup/netcup.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v1.1.0

Configuration for [Netcup](https://www.netcup.eu/).


<!--more-->

- Code: `netcup`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NETCUP_API_KEY` | API key |
| `NETCUP_API_PASSWORD` | API password |
| `NETCUP_CUSTOMER_NUMBER` | Customer number |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NETCUP_HTTP_TIMEOUT` | API request timeout |
| `NETCUP_POLLING_INTERVAL` | Time between DNS propagation check |
| `NETCUP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `NETCUP_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://www.netcup-wiki.de/wiki/DNS_API)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/netcup/netcup.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

---
title: "Exoscale"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: exoscale
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/exoscale/exoscale.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.4.0

Configuration for [Exoscale](https://www.exoscale.com/).


<!--more-->

- Code: `exoscale`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `EXOSCALE_API_KEY` | API key |
| `EXOSCALE_API_SECRET` | API secret |
| `EXOSCALE_ENDPOINT` | API endpoint URL |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `EXOSCALE_HTTP_TIMEOUT` | API request timeout |
| `EXOSCALE_POLLING_INTERVAL` | Time between DNS propagation check |
| `EXOSCALE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `EXOSCALE_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://community.exoscale.com/documentation/dns/api/)
- [Go client](https://github.com/exoscale/egoscale)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/exoscale/exoscale.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

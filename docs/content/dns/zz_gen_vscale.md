---
title: "Vscale"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: vscale
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/vscale/vscale.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v2.0.0

Configuration for [Vscale](https://vscale.io/).


<!--more-->

- Code: `vscale`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `VSCALE_API_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `VSCALE_BASE_URL` | API enddpoint URL |
| `VSCALE_HTTP_TIMEOUT` | API request timeout |
| `VSCALE_POLLING_INTERVAL` | Time between DNS propagation check |
| `VSCALE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `VSCALE_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://developers.vscale.io/documentation/api/v1/#api-Domains_Records)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/vscale/vscale.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

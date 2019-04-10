---
title: "Glesys"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: glesys
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/glesys/glesys.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.5.0

Configuration for [Glesys](https://glesys.com/).


<!--more-->

- Code: `glesys`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `GLESYS_API_KEY` | API key |
| `GLESYS_API_USER` | API user |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GLESYS_HTTP_TIMEOUT` | API request timeout |
| `GLESYS_POLLING_INTERVAL` | Time between DNS propagation check |
| `GLESYS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `GLESYS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://github.com/GleSYS/API/wiki/API-Documentation)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/glesys/glesys.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

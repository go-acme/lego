---
title: "Gandi Live DNS (v5)"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: gandiv5
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gandiv5/gandiv5.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.5.0

Configuration for [Gandi Live DNS (v5)](https://www.gandi.net).


<!--more-->

- Code: `gandiv5`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `GANDIV5_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GANDIV5_HTTP_TIMEOUT` | API request timeout |
| `GANDIV5_POLLING_INTERVAL` | Time between DNS propagation check |
| `GANDIV5_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `GANDIV5_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](http://doc.livedns.gandi.net)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gandiv5/gandiv5.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

---
title: "Domain Offensive (do.de)"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dode
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dode/dode.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v2.4.0

Configuration for [Domain Offensive (do.de)](https://www.do.de/).


<!--more-->

- Code: `dode`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DODE_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DODE_HTTP_TIMEOUT` | API request timeout |
| `DODE_POLLING_INTERVAL` | Time between DNS propagation check |
| `DODE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DODE_SEQUENCE_INTERVAL` | Interval between iteration |
| `DODE_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://www.do.de/wiki/LetsEncrypt_-_Entwickler)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dode/dode.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

---
title: "Duck DNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: duckdns
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/duckdns/duckdns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.5.0

Configuration for [Duck DNS](https://www.duckdns.org/).


<!--more-->

- Code: `duckdns`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DUCKDNS_TOKEN` | Account token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DUCKDNS_HTTP_TIMEOUT` | API request timeout |
| `DUCKDNS_POLLING_INTERVAL` | Time between DNS propagation check |
| `DUCKDNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DUCKDNS_SEQUENCE_INTERVAL` | Interval between iteration |
| `DUCKDNS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://www.duckdns.org/spec.jsp)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/duckdns/duckdns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

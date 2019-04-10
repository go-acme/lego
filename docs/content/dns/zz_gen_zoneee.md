---
title: "Zone.ee"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: zoneee
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/zoneee/zoneee.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v2.1.0

Configuration for [Zone.ee](https://www.zone.ee/).


<!--more-->

- Code: `zoneee`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ZONEEE_API_KEY` | API key |
| `ZONEEE_API_USER` | API user |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ZONEEE_ENDPOINT` | API endpoint URL |
| `ZONEEE_HTTP_TIMEOUT` | API request timeout |
| `ZONEEE_POLLING_INTERVAL` | Time between DNS propagation check |
| `ZONEEE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `ZONEEE_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://api.zone.eu/v2)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/zoneee/zoneee.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

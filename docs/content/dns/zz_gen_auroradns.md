---
title: "Aurora DNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: auroradns
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/auroradns/auroradns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.4.0

Configuration for [Aurora DNS](https://www.pcextreme.com/aurora/dns).


<!--more-->

- Code: `auroradns`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `AURORA_ENDPOINT` | API endpoint URL |
| `AURORA_KEY` | User API key |
| `AURORA_USER_ID` | User ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `AURORA_POLLING_INTERVAL` | Time between DNS propagation check |
| `AURORA_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `AURORA_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://libcloud.readthedocs.io/en/latest/dns/drivers/auroradns.html#api-docs)
- [Go client](https://github.com/nrdcg/auroradns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/auroradns/auroradns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

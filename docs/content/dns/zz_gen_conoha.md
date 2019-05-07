---
title: "ConoHa"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: conoha
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/conoha/conoha.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v1.2.0

Configuration for [ConoHa](https://www.conoha.jp/).


<!--more-->

- Code: `conoha`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `CONOHA_API_PASSWORD` | The API password |
| `CONOHA_API_USERNAME` | The API username |
| `CONOHA_TENANT_ID` | Tenant ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `CONOHA_HTTP_TIMEOUT` | API request timeout |
| `CONOHA_POLLING_INTERVAL` | Time between DNS propagation check |
| `CONOHA_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `CONOHA_REGION` | The region |
| `CONOHA_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://www.conoha.jp/docs/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/conoha/conoha.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

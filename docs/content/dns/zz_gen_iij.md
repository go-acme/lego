---
title: "Internet Initiative Japan"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: iij
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/iij/iij.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v1.1.0

Configuration for [Internet Initiative Japan](https://www.iij.ad.jp/en/).


<!--more-->

- Code: `iij`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `IIJ_API_ACCESS_KEY` | API access key |
| `IIJ_API_SECRET_KEY` | API secret key |
| `IIJ_DO_SERVICE_CODE` | DO service code |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `IIJ_POLLING_INTERVAL` | Time between DNS propagation check |
| `IIJ_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `IIJ_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](http://manual.iij.jp/p2/pubapi/http://manual.iij.jp/p2/pubapi/)
- [Go client](https://github.com/iij/doapi)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/iij/iij.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

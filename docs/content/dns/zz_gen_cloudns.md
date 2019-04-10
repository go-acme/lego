---
title: "ClouDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: cloudns
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/cloudns/cloudns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v2.3.0

Configuration for [ClouDNS](https://www.cloudns.net).


<!--more-->

- Code: `cloudns`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `CLOUDNS_AUTH_ID` | The API user ID |
| `CLOUDNS_AUTH_PASSWORD` | The password for API user ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `CLOUDNS_HTTP_TIMEOUT` | API request timeout |
| `CLOUDNS_POLLING_INTERVAL` | Time between DNS propagation check |
| `CLOUDNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `CLOUDNS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://www.cloudns.net/wiki/article/42/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/cloudns/cloudns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

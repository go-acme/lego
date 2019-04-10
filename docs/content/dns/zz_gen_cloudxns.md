---
title: "CloudXNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: cloudxns
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/cloudxns/cloudxns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.5.0

Configuration for [CloudXNS](https://www.cloudxns.net/).


<!--more-->

- Code: `cloudxns`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `CLOUDXNS_API_KEY` | The API key |
| `CLOUDXNS_SECRET_KEY` | THe API secret key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `CLOUDXNS_HTTP_TIMEOUT` | API request timeout |
| `CLOUDXNS_POLLING_INTERVAL` | Time between DNS propagation check |
| `CLOUDXNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `CLOUDXNS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://www.cloudxns.net/Public/Doc/CloudXNS_api2.0_doc_zh-cn.zip)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/cloudxns/cloudxns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
